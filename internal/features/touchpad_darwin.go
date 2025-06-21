//go:build darwin
// +build darwin

package features

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework ApplicationServices

#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>
#import <mach/mach_time.h>

// CGEventの非公開フィールド定数（mac-mouse-fixから）
const int kCGEventFieldNSEventType = 55;
const int kCGEventFieldIOHIDEventSubtype = 110;
const int kCGEventFieldScrollWheelDeltaAxis1 = 11;  // Line units
const int kCGEventFieldScrollWheelDeltaAxis2 = 12;
const int kCGEventFieldScrollWheelPointDeltaAxis1 = 96;  // Pixel units
const int kCGEventFieldScrollWheelPointDeltaAxis2 = 97;
const int kCGEventFieldScrollWheelFixedPtDeltaAxis1 = 93;  // Fixed point (16.16)
const int kCGEventFieldScrollWheelFixedPtDeltaAxis2 = 94;
const int kCGEventFieldScrollWheelPhase = 99;
const int kCGEventFieldMomentumScrollPhase = 123;
const int kCGEventFieldGesturePhase = 132;
const int kCGEventFieldGestureDeltaX = 116;
const int kCGEventFieldGestureDeltaY = 119;
const int kCGEventFieldContinuous = 88;

// NSEventType定数
const int NSEventTypeScrollWheel = 22;
const int NSEventTypeGesture = 29;
const int NSEventTypeGestureChange = 30;
const int NSEventTypeGestureEnd = 31;

// IOHIDEventタイプ定数
const int kIOHIDEventTypeScroll = 6;

// フェーズ定数
const int kIOHIDEventPhaseBegan = 1;
const int kIOHIDEventPhaseChanged = 2;
const int kIOHIDEventPhaseEnded = 4;
const int kIOHIDEventPhaseCancelled = 8;
const int kIOHIDEventPhaseMayBegin = 128;
const int kIOHIDEventPhaseUndefined = 0;

// モメンタムフェーズ定数はCoreGraphicsで定義済み
// kCGMomentumScrollPhaseNone = 0
// kCGMomentumScrollPhaseBegan = 1  
// kCGMomentumScrollPhaseChanged = 2
// kCGMomentumScrollPhaseEnded = 3

// グローバル変数
static CGEventSourceRef _eventSource = NULL;
static dispatch_queue_t _eventQueue = NULL;
static CFRunLoopRef _runLoop = NULL;
static pthread_t _runLoopThread;

// 固定小数点変換（16.16形式）
static int64_t toFixed16_16(double value) {
    return (int64_t)round(value * 65536.0);
}

// 初期化関数
int initializeTouchpad() {
    // アクセシビリティ権限チェック
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean accessibilityEnabled = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    if (!accessibilityEnabled) {
        return 0;
    }
    
    // イベントソースを作成
    _eventSource = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!_eventSource) {
        return 0;
    }
    
    // 高優先度のディスパッチキューを作成
    dispatch_queue_attr_t attr = dispatch_queue_attr_make_with_qos_class(
        DISPATCH_QUEUE_SERIAL, QOS_CLASS_USER_INTERACTIVE, -1);
    _eventQueue = dispatch_queue_create("com.keyball.gestures.event", attr);
    
    return 1;
}

// スクロールイベントを送信（mac-mouse-fix方式）
void postScrollEventPair(double deltaX, double deltaY, int phase, int momentumPhase) {
    // Type 22イベント（ScrollWheel）を作成
    CGEventRef e22 = CGEventCreate(_eventSource);
    if (!e22) return;
    
    CGEventSetIntegerValueField(e22, kCGEventFieldNSEventType, NSEventTypeScrollWheel);
    CGEventSetIntegerValueField(e22, kCGEventFieldContinuous, 1);  // Continuous scrolling
    
    // デルタ値を設定（3種類すべて）
    int lineY = (int)round(deltaY);
    int lineX = (int)round(deltaX);
    
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelDeltaAxis1, lineY);
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelDeltaAxis2, lineX);
    
    CGEventSetDoubleValueField(e22, kCGEventFieldScrollWheelPointDeltaAxis1, deltaY);
    CGEventSetDoubleValueField(e22, kCGEventFieldScrollWheelPointDeltaAxis2, deltaX);
    
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelFixedPtDeltaAxis1, toFixed16_16(deltaY));
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelFixedPtDeltaAxis2, toFixed16_16(deltaX));
    
    // フェーズを設定
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelPhase, phase);
    CGEventSetIntegerValueField(e22, kCGEventFieldMomentumScrollPhase, momentumPhase);
    
    // Type 29イベント（Gesture）を作成
    CGEventRef e29 = CGEventCreate(_eventSource);
    if (!e29) {
        CFRelease(e22);
        return;
    }
    
    CGEventSetIntegerValueField(e29, kCGEventFieldNSEventType, NSEventTypeGesture);
    CGEventSetIntegerValueField(e29, kCGEventFieldIOHIDEventSubtype, kIOHIDEventTypeScroll);
    
    // ジェスチャーデルタを設定
    CGEventSetDoubleValueField(e29, kCGEventFieldGestureDeltaX, -deltaX);  // 反転
    CGEventSetDoubleValueField(e29, kCGEventFieldGestureDeltaY, -deltaY);  // 反転
    
    // ジェスチャーフェーズを設定
    CGEventSetIntegerValueField(e29, kCGEventFieldGesturePhase, phase);
    
    // 両方のイベントを送信
    CGEventPost(kCGHIDEventTap, e22);
    CGEventPost(kCGHIDEventTap, e29);
    
    CFRelease(e22);
    CFRelease(e29);
}

// 4本指スワイプイベントを送信
void postSwipeGesture(double deltaX, double deltaY, int phase) {
    // デバッグ出力
    NSLog(@"postSwipeGesture: deltaX=%f, deltaY=%f, phase=%d", deltaX, deltaY, phase);
    
    // 空のイベントを作成（純粋なジェスチャーイベント）
    CGEventRef gesture = CGEventCreate(_eventSource);
    if (!gesture) {
        NSLog(@"Failed to create event");
        return;
    }
    
    // フェーズに応じてNSEventTypeを設定
    int nsEventType = NSEventTypeGesture;  // デフォルトはBegin
    if (phase == kIOHIDEventPhaseChanged) {
        nsEventType = NSEventTypeGestureChange;
    } else if (phase == kIOHIDEventPhaseEnded) {
        nsEventType = NSEventTypeGestureEnd;
    }
    
    // ジェスチャーイベントとして必要なフィールドをすべて設定
    CGEventSetIntegerValueField(gesture, kCGEventFieldNSEventType, nsEventType);
    CGEventSetIntegerValueField(gesture, kCGEventFieldIOHIDEventSubtype, 2);  // スワイプ
    
    // ジェスチャーのdelta値を設定（これが最も重要）
    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaX, deltaX);
    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaY, deltaY);
    
    // ジェスチャーフェーズを設定
    CGEventSetIntegerValueField(gesture, kCGEventFieldGesturePhase, phase);
    
    // タイムスタンプを設定（必須）
    CGEventSetTimestamp(gesture, mach_absolute_time());
    
    // イベントタイプを設定（重要）
    CGEventSetType(gesture, (CGEventType)nsEventType);
    
    // デバッグ: 設定した値を確認
    double checkX = CGEventGetDoubleValueField(gesture, kCGEventFieldGestureDeltaX);
    double checkY = CGEventGetDoubleValueField(gesture, kCGEventFieldGestureDeltaY);
    NSLog(@"After setting - GestureX=%f, GestureY=%f, EventType=%d", checkX, checkY, nsEventType);
    
    // イベントを送信
    CGEventPost(kCGHIDEventTap, gesture);
    
    CFRelease(gesture);
}

// クリーンアップ
void cleanupTouchpad() {
    if (_eventSource) {
        CFRelease(_eventSource);
        _eventSource = NULL;
    }
}
*/
import "C"

// macOSタッチパッド実装
type darwinTouchPad struct {
	config          TouchPadConfig
	mu              sync.Mutex
	touchSlots      map[int]*touchSlot
	scrollStarted   bool
	swipeStarted    bool  // 4本指スワイプフラグ
	lastEventTime   time.Time
	motionFilter    *MotionFilter
	initialized     bool
	currentScrollPhase int
	currentSwipePhase  int  // 4本指スワイプのフェーズ
	pendingGesture    bool  // ジェスチャー開始を保留中
	firstTouchTime    time.Time  // 最初のタッチの時刻
	gestureTimer     *time.Timer  // ジェスチャー開始のタイマー
}

// タッチスロット情報
type touchSlot struct {
	trackingID int
	lastX      int32
	lastY      int32
	startX     int32
	startY     int32
	isActive   bool
}

// CreateTouchPad macOS実装：CoreGraphicsを使用してタッチパッドイベントをシミュレート
func CreateTouchPad(config TouchPadConfig) (TouchPad, error) {
	// 初期化
	if C.initializeTouchpad() == 0 {
		return nil, fmt.Errorf("タッチパッドの初期化に失敗しました。アクセシビリティ権限を確認してください")
	}

	dt := &darwinTouchPad{
		config:       config,
		touchSlots:   make(map[int]*touchSlot),
		motionFilter: NewMotionFilter(config.MotionSmoothingFactor, config.MotionWarmUpCount),
		initialized:  true,
	}

	log.Println("macOSタッチパッドを初期化しました")
	return dt, nil
}

// MultiTouchDown タッチ開始イベント
func (dt *darwinTouchPad) MultiTouchDown(slot int, trackingID int, x int32, y int32) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.initialized {
		return fmt.Errorf("タッチパッドが初期化されていません")
	}

	// スロット情報を保存
	dt.touchSlots[slot] = &touchSlot{
		trackingID: trackingID,
		lastX:      x,
		lastY:      y,
		startX:     x,
		startY:     y,
		isActive:   true,
	}

	// タッチ開始前のアクティブなタッチ数を取得
	previousCount := dt.getActiveTouchCount() - 1  // 今追加したタッチを除く
	
	// 最初のタッチの時刻を記録
	if previousCount == 0 {
		dt.firstTouchTime = time.Now()
		dt.pendingGesture = true
		log.Printf("最初のタッチダウン: slot=%d", slot)
	}
	
	// すべてのタッチが追加された後のアクティブなタッチ数を取得
	activeCount := dt.getActiveTouchCount()
	
	// ジェスチャー判定を遅延させる（50ms以内に追加されたタッチは同時とみなす）
	if dt.pendingGesture && time.Since(dt.firstTouchTime) < 50*time.Millisecond {
		// まだジェスチャーを開始しない
		log.Printf("タッチ追加中: activeCount=%d", activeCount)
		return nil
	}
	
	// ジェスチャーの種類を判定して開始
	if dt.pendingGesture && !dt.scrollStarted && !dt.swipeStarted {
		dt.pendingGesture = false
		
		if activeCount == 2 {
			// 2本指スクロール開始
			dt.scrollStarted = true
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseMayBegin)
			
			// MayBeginフェーズを送信
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseMayBegin, C.kCGMomentumScrollPhaseNone)
			
			// 少し遅延を入れてからBeganフェーズを送信
			time.Sleep(1 * time.Millisecond)
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseBegan)
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseBegan, C.kCGMomentumScrollPhaseNone)
			
			log.Printf("2本指スクロール開始: activeCount=%d", activeCount)
		} else if activeCount == 4 {
			// 4本指スワイプ開始
			dt.swipeStarted = true
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseMayBegin)
			
			// MayBeginフェーズを送信
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseMayBegin)
			log.Printf("4本指スワイプ MayBegin送信")
			
			// 少し遅延を入れてからBeganフェーズを送信
			time.Sleep(1 * time.Millisecond)
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseBegan)
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseBegan)
			
			log.Printf("4本指スワイプ開始: activeCount=%d, phase=Began", activeCount)
		}
	}

	dt.lastEventTime = time.Now()
	return nil
}

// MultiTouchMove タッチ移動イベント
func (dt *darwinTouchPad) MultiTouchMove(slot int, x int32, y int32) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.initialized {
		return fmt.Errorf("タッチパッドが初期化されていません")
	}

	touch, exists := dt.touchSlots[slot]
	if !exists || !touch.isActive {
		return fmt.Errorf("スロット %d が見つかりません", slot)
	}

	// デルタを計算
	deltaX := x - touch.lastX
	deltaY := y - touch.lastY

	// ジェスチャー判定中の場合
	if dt.pendingGesture && time.Since(dt.firstTouchTime) >= 50*time.Millisecond {
		dt.pendingGesture = false
		activeCount := dt.getActiveTouchCount()
		
		if activeCount == 2 && !dt.scrollStarted && !dt.swipeStarted {
			// 2本指スクロール開始
			dt.scrollStarted = true
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseMayBegin)
			
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseMayBegin, C.kCGMomentumScrollPhaseNone)
			time.Sleep(1 * time.Millisecond)
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseBegan)
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseBegan, C.kCGMomentumScrollPhaseNone)
			
			log.Printf("2本指スクロール開始（移動時）: activeCount=%d", activeCount)
		} else if activeCount == 4 && !dt.scrollStarted && !dt.swipeStarted {
			// 4本指スワイプ開始
			dt.swipeStarted = true
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseMayBegin)
			
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseMayBegin)
			log.Printf("4本指スワイプ MayBegin送信（移動時）")
			time.Sleep(1 * time.Millisecond)
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseBegan)
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseBegan)
			
			log.Printf("4本指スワイプ開始（移動時）: activeCount=%d, phase=Began", activeCount)
		}
	}

	// 2本指スクロール中の場合
	if dt.scrollStarted && dt.getActiveTouchCount() == 2 {
		// モーションフィルターを適用
		filteredDeltaX, filteredDeltaY := dt.motionFilter.Filter(deltaX, deltaY)
		
		// スケーリング（タッチパッド座標系からピクセルへ）
		// mac-mouse-fixの実装に基づいて調整
		scaleFactor := dt.config.MouseDeltaFactor * 0.1
		scaledDeltaX := float64(filteredDeltaX) * scaleFactor
		scaledDeltaY := float64(filteredDeltaY) * scaleFactor
		
		// 最小閾値を設定（小さすぎる動きは無視）
		if abs(scaledDeltaX) > 0.1 || abs(scaledDeltaY) > 0.1 {
			// 現在のフェーズがBeganまたはMayBeginの場合、Changedに移行
			if dt.currentScrollPhase == int(C.kIOHIDEventPhaseBegan) ||
			   dt.currentScrollPhase == int(C.kIOHIDEventPhaseMayBegin) {
				dt.currentScrollPhase = int(C.kIOHIDEventPhaseChanged)
			}
			
			// スクロールイベントペアを送信（mac-mouse-fix方式）
			C.postScrollEventPair(
				C.double(scaledDeltaX),
				C.double(scaledDeltaY),
				C.int(dt.currentScrollPhase),
				C.kCGMomentumScrollPhaseNone,
			)
			
			log.Printf("スクロール移動: dx=%.2f, dy=%.2f, phase=%d", scaledDeltaX, scaledDeltaY, dt.currentScrollPhase)
		}
	}
	
	// 4本指スワイプ中の場合
	if dt.swipeStarted && dt.getActiveTouchCount() == 4 {
		// モーションフィルターを適用
		filteredDeltaX, filteredDeltaY := dt.motionFilter.Filter(deltaX, deltaY)
		
		// スケーリング（4本指スワイプはより大きな動きが必要）
		scaleFactor := dt.config.MouseDeltaFactor * 0.2
		scaledDeltaX := float64(filteredDeltaX) * scaleFactor
		scaledDeltaY := float64(filteredDeltaY) * scaleFactor
		
		// 最小閾値を設定（小さすぎる動きは無視）
		if abs(scaledDeltaX) > 0.5 || abs(scaledDeltaY) > 0.5 {
			// 現在のフェーズがBeganまたはMayBeginの場合、Changedに移行
			if dt.currentSwipePhase == int(C.kIOHIDEventPhaseBegan) ||
			   dt.currentSwipePhase == int(C.kIOHIDEventPhaseMayBegin) {
				dt.currentSwipePhase = int(C.kIOHIDEventPhaseChanged)
			}
			
			// スワイプイベントを送信
			C.postSwipeGesture(
				C.double(scaledDeltaX),
				C.double(scaledDeltaY),
				C.int(dt.currentSwipePhase),
			)
			
			log.Printf("4本指スワイプ移動: dx=%.2f, dy=%.2f, phase=%d (type=%d)", scaledDeltaX, scaledDeltaY, dt.currentSwipePhase, 
				func() int {
					if dt.currentSwipePhase == int(C.kIOHIDEventPhaseChanged) {
						return 30 // GestureChange
					} else if dt.currentSwipePhase == int(C.kIOHIDEventPhaseEnded) {
						return 31 // GestureEnd
					}
					return 29 // GestureBegin
				}())
		}
	}

	// 位置を更新
	touch.lastX = x
	touch.lastY = y
	dt.lastEventTime = time.Now()

	return nil
}

// MultiTouchUp タッチ終了イベント
func (dt *darwinTouchPad) MultiTouchUp(slot int) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.initialized {
		return fmt.Errorf("タッチパッドが初期化されていません")
	}

	touch, exists := dt.touchSlots[slot]
	if !exists {
		return fmt.Errorf("スロット %d が見つかりません", slot)
	}

	touch.isActive = false
	activeCount := dt.getActiveTouchCount()

	// すべてのタッチが終了した場合
	if activeCount == 0 {
		if dt.scrollStarted {
			// スクロール終了イベントを送信
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseEnded, C.kCGMomentumScrollPhaseNone)
			
			dt.scrollStarted = false
			dt.currentScrollPhase = 0
			dt.motionFilter.Reset()
			
			log.Println("スクロール終了")
			
			// 慣性スクロールはmacOSが自動的に処理するため、シミュレートしない
		}
		
		if dt.swipeStarted {
			// スワイプ終了イベントを送信
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseEnded)
			
			dt.swipeStarted = false
			dt.currentSwipePhase = 0
			dt.motionFilter.Reset()
			
			log.Println("4本指スワイプ終了 (type=31)")
		}
	}

	// スロットをクリーンアップ
	delete(dt.touchSlots, slot)
	dt.lastEventTime = time.Now()

	return nil
}

// Close リソースの解放
func (dt *darwinTouchPad) Close() error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.initialized {
		return nil
	}

	// アクティブなジェスチャーがある場合は終了
	if dt.scrollStarted {
		C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseEnded, C.kCGMomentumScrollPhaseNone)
	}
	if dt.swipeStarted {
		C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseEnded)
	}

	// Cリソースをクリーンアップ
	C.cleanupTouchpad()

	dt.touchSlots = nil
	dt.initialized = false
	
	log.Println("macOSタッチパッドをクローズしました")
	return nil
}

// getActiveTouchCount アクティブなタッチ数を取得
func (dt *darwinTouchPad) getActiveTouchCount() int {
	count := 0
	for _, touch := range dt.touchSlots {
		if touch.isActive {
			count++
		}
	}
	return count
}

// abs 絶対値を返す
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}