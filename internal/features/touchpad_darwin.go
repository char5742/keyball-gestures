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
const int kCGEventFieldIOHIDEventFlags = 115;
const int kCGEventFieldGestureMask = 134;

// NSEventType定数
const int NSEventTypeScrollWheel = 22;
const int NSEventTypeGesture = 29;
const int NSEventTypeGestureChange = 30;
const int NSEventTypeGestureEnd = 31;

// IOHIDEventタイプ定数
const int kIOHIDEventTypeScroll = 6;
const int kIOHIDEventTypeSwipe = 15;  // 旧macOS用のスワイプタイプ
const int kIOHIDEventSubtypeSwipeNew = 23;  // macOS 14+ 用のスワイプサブタイプ

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

// 現在のマウス位置を取得
CGPoint getCurrentMousePosition() {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint pos = CGPointMake(0, 0);
    if (event) {
        pos = CGEventGetLocation(event);
        CFRelease(event);
    }
    return pos;
}

// スクロールイベントを送信（修正版）
void postScrollEventPair(double deltaX, double deltaY, int phase, int momentumPhase, CGPoint pos) {
    // カーソル位置を固定するため、現在位置を保存
    CGPoint currentPos = getCurrentMousePosition();
    
    // 標準的な方法でスクロールイベントを作成
    CGEventRef e22 = CGEventCreateScrollWheelEvent(
        _eventSource,
        kCGScrollEventUnitPixel,
        2,  // 2軸（Y, X）
        (int32_t)deltaY,
        (int32_t)deltaX
    );
    if (!e22) return;
    
    // トラックパッドからのイベントであることを示す
    CGEventSetIntegerValueField(e22, kCGEventFieldContinuous, 1);
    
    // フェーズを設定
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelPhase, phase);
    CGEventSetIntegerValueField(e22, kCGEventFieldMomentumScrollPhase, momentumPhase);
    
    // Fixed-pointデルタも設定（Safari 15以降で優先される）
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelFixedPtDeltaAxis1, toFixed16_16(deltaX));
    CGEventSetIntegerValueField(e22, kCGEventFieldScrollWheelFixedPtDeltaAxis2, toFixed16_16(deltaY));
    
    // スクロールイベントはジェスチャー開始位置で発生させる
    CGEventSetLocation(e22, pos);
    
    // タイムスタンプを設定（必ず新しい値を取得）
    CGEventSetTimestamp(e22, mach_absolute_time());
    
    // Type 29イベント（Gesture）を作成
    CGEventRef e29 = CGEventCreate(_eventSource);
    if (!e29) {
        CFRelease(e22);
        return;
    }
    
    CGEventSetType(e29, (CGEventType)NSEventTypeGesture);
    CGEventSetIntegerValueField(e29, kCGEventFieldNSEventType, NSEventTypeGesture);
    CGEventSetIntegerValueField(e29, kCGEventFieldIOHIDEventSubtype, kIOHIDEventTypeScroll);
    
    // ジェスチャーデルタを設定
    CGEventSetDoubleValueField(e29, kCGEventFieldGestureDeltaX, -deltaX);  // 反転
    CGEventSetDoubleValueField(e29, kCGEventFieldGestureDeltaY, -deltaY);  // 反転
    
    // ジェスチャーフェーズを設定
    CGEventSetIntegerValueField(e29, kCGEventFieldGesturePhase, phase);
    
    // タイムスタンプを再取得（必ず新しい値を取得）
    CGEventSetTimestamp(e29, mach_absolute_time());
    
    // 両方のイベントを送信（trackpad_dump.cで検出可能にするためkCGSessionEventTapに統一）
    CGEventPost(kCGSessionEventTap, e22);
    CGEventPost(kCGSessionEventTap, e29);
    
    // カーソル位置を元に戻す（ジェスチャー開始位置に固定）
    CGDisplayMoveCursorToPoint(CGMainDisplayID(), pos);
    
    CFRelease(e22);
    CFRelease(e29);
}

// 4本指スワイプイベントを送信（macOS 14+仕様準拠）
void postSwipeGesture(double deltaX, double deltaY, int phase, CGPoint pos, double totalDX, double totalDY) {
    // フェーズマスクを設定（GestureMask = phase bit）
    // ネイティブ仕様：
    // mask=0x1: Beganフェーズ
    // mask=0x2: Changedフェーズ
    // mask=0x4: Endedフェーズ
    // mask=0x8: Cancelledフェーズ
    uint32_t phaseMask = 0;
    switch (phase) {
        case kIOHIDEventPhaseBegan:      phaseMask = 0x1; break;
        case kIOHIDEventPhaseChanged:    phaseMask = 0x2; break;
        case kIOHIDEventPhaseEnded:      phaseMask = 0x4; break;
        case kIOHIDEventPhaseCancelled:  phaseMask = 0x8; break;
        default: return;  // MayBeginは送らない
    }
    
    // デバッグ出力
    NSLog(@"postSwipeGesture: Δ(%.1f,%.1f), phase=%d, phaseMask=0x%x", deltaX, deltaY, phase, phaseMask);
    
    // 空のイベントを作成（純粋なジェスチャーイベント）
    CGEventRef gesture = CGEventCreate(_eventSource);
    if (!gesture) {
        NSLog(@"Failed to create event");
        return;
    }
    
    // macOS 14+: すべてのフェーズでtype=30(GestureChange)を使用
    CGEventSetType(gesture, (CGEventType)NSEventTypeGestureChange);
    CGEventSetIntegerValueField(gesture, kCGEventFieldNSEventType, NSEventTypeGestureChange);
    
    // subtype=23 (Swipe) 固定
    CGEventSetIntegerValueField(gesture, kCGEventFieldIOHIDEventSubtype, kIOHIDEventSubtypeSwipeNew);
    
    // delta値を設定（GestureDeltaフィールドは0だが、実際の動きを伝える）
    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaX, deltaX);
    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaY, deltaY);
    
    // ジェスチャーフェーズを設定
    CGEventSetIntegerValueField(gesture, kCGEventFieldGesturePhase, phase);
    
    // Continuousフラグを設定（トラックパッドからのジェスチャー）
    CGEventSetIntegerValueField(gesture, kCGEventFieldContinuous, 1);
    
    // フェーズマスクを設定（phase bit表現）
    CGEventSetIntegerValueField(gesture, kCGEventFieldGestureMask, phaseMask);
    
    // IOHIDEventFlags: Began=0x1, Cancel=0x8, それ以外0
    int flags = (phase == kIOHIDEventPhaseBegan) ? 0x1 :
                (phase == kIOHIDEventPhaseCancelled) ? 0x8 : 0;
    CGEventSetIntegerValueField(gesture, kCGEventFieldIOHIDEventFlags, flags);
    
    // ジェスチャー開始時の位置を設定（マウスカーソルを固定するため）
    CGEventSetLocation(gesture, pos);
    
    // タイムスタンプを設定（必須）
    CGEventSetTimestamp(gesture, mach_absolute_time());
    
    // イベントを送信
    CGEventPost(kCGSessionEventTap, gesture);
    CFRelease(gesture);
    
    // ネイティブ動作に合わせて、各type=30の後にdummy 29を送信
    CGEventRef dummy = CGEventCreate(_eventSource);
    if (dummy) {
        CGEventSetType(dummy, (CGEventType)NSEventTypeGesture);   // 29
        CGEventSetIntegerValueField(dummy, kCGEventFieldNSEventType, NSEventTypeGesture);
        CGEventSetIntegerValueField(dummy, kCGEventFieldIOHIDEventSubtype, 0);
        CGEventSetTimestamp(dummy, mach_absolute_time());
        CGEventPost(kCGSessionEventTap, dummy);
        CFRelease(dummy);
    }
    
    // カーソル位置を元に戻す（ジェスチャー開始位置に固定）
    CGDisplayMoveCursorToPoint(CGMainDisplayID(), pos);
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

// ジェスチャー関連の定数
const (
	// スワイプ関連
	swipeInterval  = 12 * time.Millisecond  // より高頻度のレート制限
	
	// タイミング関連
	gestureBeginDelay = 8 * time.Millisecond   // MayBegin→Beginの遅延
	gestureDetectionWindow = 50 * time.Millisecond  // ジェスチャー判定ウィンドウ
	
	// スクロール関連
	scrollInterval  = 12 * time.Millisecond  // スクロールのレート制限
	scrollThreshold = 0.5                     // スクロール認識の最小閾値
)

// ジェスチャー開始位置を保存するための構造体
type gesturePosition struct {
	x, y C.double
}

// gesturePositionをC.CGPointに変換
func (p gesturePosition) toCGPoint() C.CGPoint {
	return C.CGPoint{x: p.x, y: p.y}
}

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
	gestureStartPos   gesturePosition  // ジェスチャー開始時のマウス位置
	
	// レート制限用
	lastSwipeSentAt  time.Time  // 最後にスワイプを送信した時刻
	lastScrollSentAt time.Time  // 最後にスクロールを送信した時刻
	primarySlot      int        // 代表スロット（最初のタッチ）
	
	// スワイプ制御用
	swipeFirstEvent  bool       // 最初のスワイプイベントかどうか
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
		dt.primarySlot = slot  // 代表スロットを記録
		log.Printf("最初のタッチダウン: slot=%d (代表スロット)", slot)
	}
	
	// すべてのタッチが追加された後のアクティブなタッチ数を取得
	activeCount := dt.getActiveTouchCount()
	log.Printf("MultiTouchDown後: slot=%d, trackingID=%d, 現在のアクティブタッチ数=%d", slot, trackingID, activeCount)
	
	// ジェスチャー判定を遅延させる（gestureDetectionWindow以内に追加されたタッチは同時とみなす）
	if dt.pendingGesture && time.Since(dt.firstTouchTime) < gestureDetectionWindow {
		// まだジェスチャーを開始しない
		log.Printf("タッチ追加中: activeCount=%d", activeCount)
		return nil
	}
	
	// gestureDetectionWindow経過後、またはMotionWarmUpCountに達したらジェスチャーを開始
	if dt.pendingGesture && !dt.scrollStarted && !dt.swipeStarted && 
	   (time.Since(dt.firstTouchTime) >= gestureDetectionWindow || activeCount >= dt.config.MotionWarmUpCount) {
		dt.pendingGesture = false
		log.Printf("ジェスチャー判定開始: activeCount=%d", activeCount)
		
		if activeCount == 2 {
			// 2本指スクロール開始
			dt.scrollStarted = true
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseMayBegin)
			
			// ジェスチャー開始時のマウス位置を保存
			cPos := C.getCurrentMousePosition()
			dt.gestureStartPos = gesturePosition{x: cPos.x, y: cPos.y}
			
			// MayBeginフェーズを送信
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseMayBegin, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
			
			// 少し遅延を入れてからBeganフェーズを送信（ネイティブ間隔）
			time.Sleep(gestureBeginDelay)
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseBegan)
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseBegan, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
			
			log.Printf("2本指スクロール開始: activeCount=%d, startPos=(%.1f,%.1f)", activeCount, float64(dt.gestureStartPos.x), float64(dt.gestureStartPos.y))
		} else if activeCount == 4 {
			// 4本指スワイプ開始（MayBeginは送らない）
			dt.swipeStarted = true
			dt.swipeFirstEvent = true
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseBegan)
			
			// ジェスチャー開始時のマウス位置を保存
			cPos := C.getCurrentMousePosition()
			dt.gestureStartPos = gesturePosition{x: cPos.x, y: cPos.y}
			
			// Beganフェーズを送信（MayBeginはスキップ）
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseBegan, dt.gestureStartPos.toCGPoint(), 0, 0)
			
			log.Printf("4本指スワイプ開始: activeCount=%d, phase=Began, startPos=(%.1f,%.1f)", activeCount, float64(dt.gestureStartPos.x), float64(dt.gestureStartPos.y))
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
			
			// ジェスチャー開始時のマウス位置を保存
			cPos := C.getCurrentMousePosition()
			dt.gestureStartPos = gesturePosition{x: cPos.x, y: cPos.y}
			
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseMayBegin, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
			time.Sleep(1 * time.Millisecond)
			dt.currentScrollPhase = int(C.kIOHIDEventPhaseBegan)
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseBegan, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
			
			log.Printf("2本指スクロール開始（移動時）: activeCount=%d, startPos=(%.1f,%.1f)", activeCount, float64(dt.gestureStartPos.x), float64(dt.gestureStartPos.y))
		} else if activeCount == 4 && !dt.scrollStarted && !dt.swipeStarted {
			// 4本指スワイプ開始（MayBeginは送らない）
			dt.swipeStarted = true
			dt.swipeFirstEvent = true
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseBegan)
			
			// ジェスチャー開始時のマウス位置を保存
			cPos := C.getCurrentMousePosition()
			dt.gestureStartPos = gesturePosition{x: cPos.x, y: cPos.y}
			
			// Beganフェーズを送信（MayBeginはスキップ）
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseBegan, dt.gestureStartPos.toCGPoint(), 0, 0)
			
			log.Printf("4本指スワイプ開始（移動時）: activeCount=%d, phase=Began, startPos=(%.1f,%.1f)", activeCount, float64(dt.gestureStartPos.x), float64(dt.gestureStartPos.y))
		}
	}

	// 2本指スクロール中の場合
	if dt.scrollStarted && dt.getActiveTouchCount() == 2 {
		// 代表スロット以外は無視
		if slot != dt.primarySlot {
			return nil
		}
		
		// モーションフィルターを適用
		filteredDeltaX, filteredDeltaY := dt.motionFilter.Filter(deltaX, deltaY)
		
		// スケーリング（タッチパッド座標系からピクセルへ）
		// ScrollScaleFactorが設定されていない場合はデフォルト値を使用
		scaleFactor := dt.config.ScrollScaleFactor
		if scaleFactor == 0 {
			scaleFactor = dt.config.MouseDeltaFactor * 0.05  // 後方互換性
		}
		scaledDeltaX := float64(filteredDeltaX) * scaleFactor
		scaledDeltaY := float64(filteredDeltaY) * scaleFactor
		
		// 最小閾値を設定（小さすぎる動きは無視）
		if abs(scaledDeltaX) < scrollThreshold && abs(scaledDeltaY) < scrollThreshold {
			// デルタが小さすぎる場合はイベントを送信しない
			return nil
		}
		
		// レート制限チェック
		if time.Since(dt.lastScrollSentAt) < scrollInterval {
			// まだ送信間隔に達していない場合はスキップ
			return nil
		}
		
		dt.lastScrollSentAt = time.Now()
		
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
			dt.gestureStartPos.toCGPoint(),
		)
		
		log.Printf("スクロール移動: dx=%.2f, dy=%.2f, phase=%d", scaledDeltaX, scaledDeltaY, dt.currentScrollPhase)
	}
	
	// 4本指スワイプ中の場合
	if dt.swipeStarted && dt.getActiveTouchCount() == 4 {
		// 代表スロット以外は無視
		if slot != dt.primarySlot {
			return nil
		}
		
		// レート制限チェック
		if time.Since(dt.lastSwipeSentAt) < swipeInterval {
			// まだ送信間隔に達していない場合はスキップ
			return nil
		}
		
		dt.lastSwipeSentAt = time.Now()
		
		// 現在のフェーズがBeganの場合、Changedに移行
		if dt.currentSwipePhase == int(C.kIOHIDEventPhaseBegan) {
			dt.currentSwipePhase = int(C.kIOHIDEventPhaseChanged)
		}
		
		// モーションフィルターを適用
		filteredDeltaX, filteredDeltaY := dt.motionFilter.Filter(deltaX, deltaY)
		
		// スケーリング（タッチパッド座標系からピクセルへ）
		scaleFactor := dt.config.SwipeScaleFactor
		if scaleFactor == 0 {
			scaleFactor = 0.3  // デフォルト値（Mission Control動作用）
		}
		scaledDeltaX := float64(filteredDeltaX) * scaleFactor
		scaledDeltaY := float64(filteredDeltaY) * scaleFactor
		
		// スワイプイベントを送信
		C.postSwipeGesture(
			C.double(scaledDeltaX),
			C.double(scaledDeltaY),
			C.int(dt.currentSwipePhase),
			dt.gestureStartPos.toCGPoint(),
			0, 0,  // 累積は使わない
		)
		
		log.Printf("4本指スワイプ移動: Δ(%.1f,%.1f), phase=%d", scaledDeltaX, scaledDeltaY, dt.currentSwipePhase)
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
			C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseEnded, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
			
			dt.scrollStarted = false
			dt.currentScrollPhase = 0
			dt.motionFilter.Reset()
			
			log.Println("スクロール終了")
			
			// 慣性スクロールはmacOSが自動的に処理するため、シミュレートしない
		}
		
		if dt.swipeStarted {
			// スワイプ終了イベントを送信（deltaと累積は0）
			C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseEnded, dt.gestureStartPos.toCGPoint(), 0, 0)
			
			dt.swipeStarted = false
			dt.currentSwipePhase = 0
			dt.swipeFirstEvent = false
			dt.motionFilter.Reset()
			
			log.Println("4本指スワイプ終了")
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
		C.postScrollEventPair(0, 0, C.kIOHIDEventPhaseEnded, C.kCGMomentumScrollPhaseNone, dt.gestureStartPos.toCGPoint())
	}
	if dt.swipeStarted {
		C.postSwipeGesture(0, 0, C.kIOHIDEventPhaseEnded, dt.gestureStartPos.toCGPoint(), 0, 0)
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