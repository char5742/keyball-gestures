//go:build integration && darwin
// +build integration,darwin

package integration

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>

// trackpad_dump.cと全く同じ実装
static const char* phaseName(int64_t p){
    switch(p){case 128:return "MayBegin";case 1:return "Began";
              case 2:return "Changed";case 4:return "Ended";
              case 8:return "Cancelled";default:return "?";}
}

static const char* subtypeName(int64_t s){
    switch(s){
        case 2:  return "Pinch";          // 2-finger pinch/rotate
        case 6:  return "Scroll";         // 2-finger scroll
        case 15: return "Swipe(legacy)";  // Mojave 以前
        case 23: return "Swipe";          // 3/4-finger, Sonoma+
        default: return "?";
    }
}

// イベント記録用構造体
typedef struct {
    CGEventType type;
    double timestamp;
    int64_t subtype;
    int64_t phase;
    int64_t mask;
    int64_t flags;
    double gestureX;
    double gestureY;
    double scrollX;
    double scrollY;
    int isContinuous;
    char description[256];
} CapturedEvent;

// グローバル変数（trackpad_dump.cと同じ設計）
static CapturedEvent capturedEvents[1000];
static int capturedEventCount = 0;
static pthread_mutex_t captureMutex = PTHREAD_MUTEX_INITIALIZER;
static CFMachPortRef captureTap = NULL;
static CFRunLoopSourceRef captureRunLoopSource = NULL;
static CFRunLoopRef captureRunLoop = NULL;
static pthread_t captureMonitorThread;
static bool isCaptureMonitoring = false;

// trackpad_dump.cと全く同じコールバック実装
static CGEventRef trackpadDumpCallback(CGEventTapProxy proxy, CGEventType t, CGEventRef e, void* info) {
    pthread_mutex_lock(&captureMutex);
    
    if (capturedEventCount < 1000) {
        CapturedEvent *capture = &capturedEvents[capturedEventCount];
        
        double ts = CGEventGetTimestamp(e)/1e9;
        int64_t sub = CGEventGetIntegerValueField(e,110);
        int64_t ph  = CGEventGetIntegerValueField(e,132);
        int64_t msk = CGEventGetIntegerValueField(e,134);
        int64_t flg = CGEventGetIntegerValueField(e,115);
        
        capture->type = t;
        capture->timestamp = ts;
        capture->subtype = sub;
        capture->phase = ph;
        capture->mask = msk;
        capture->flags = flg;
        
        // trackpad_dump.cと同じ判定ロジック
        if(t==kCGEventScrollWheel){
            double dx = CGEventGetDoubleValueField(e,96);
            double dy = CGEventGetDoubleValueField(e,97);
            capture->scrollX = dx;
            capture->scrollY = dy;
            capture->gestureX = 0;
            capture->gestureY = 0;
            capture->isContinuous = (int)CGEventGetIntegerValueField(e,kCGScrollWheelEventIsContinuous);
            
            if(!capture->isContinuous) {
                pthread_mutex_unlock(&captureMutex);
                return e; // trackpad_dump.cと同じ：非Continuousは無視
            }
            
            snprintf(capture->description, sizeof(capture->description),
                    "[%.3f] Scroll %s Δ(%.1f,%.1f)", ts, phaseName(ph), dx, dy);
        } else if(t==29||t==30||t==31){
            double gdx = CGEventGetDoubleValueField(e,116);
            double gdy = CGEventGetDoubleValueField(e,119);
            capture->gestureX = gdx;
            capture->gestureY = gdy;
            capture->scrollX = 0;
            capture->scrollY = 0;
            capture->isContinuous = (int)CGEventGetIntegerValueField(e,88);
            
            // trackpad_dump.cと同じフォーマット
            char maskStr[64] = {0};
            if(!msk) {
                snprintf(maskStr, sizeof(maskStr), "mask=0 ");
            } else {
                snprintf(maskStr, sizeof(maskStr), "mask=0x%llX [", msk);
                if(msk&1) strcat(maskStr, "Left ↤ ");
                if(msk&2) strcat(maskStr, "Right ↦ ");
                if(msk&4) strcat(maskStr, "Up ⇡ ");
                if(msk&8) strcat(maskStr, "Down ⇣ ");
                strcat(maskStr, "] ");
            }
            
            snprintf(capture->description, sizeof(capture->description),
                    "[%.3f] Gesture type=%d(%s) sub=%lld(%s) phase=%s %sflags=0x%llX%s%s",
                    ts, t, (t==29?"Begin":t==30?"Change":"End"),
                    sub, subtypeName(sub), phaseName(ph), maskStr, flg,
                    (gdx||gdy) ? "\n          " : "",
                    (gdx||gdy) ? "" : "");
            
            if(gdx||gdy) {
                char deltaStr[64];
                snprintf(deltaStr, sizeof(deltaStr), "Δ(%.1f,%.1f)", gdx, gdy);
                strcat(capture->description, deltaStr);
            }
        } else {
            capture->gestureX = 0;
            capture->gestureY = 0;
            capture->scrollX = 0;
            capture->scrollY = 0;
            capture->isContinuous = 0;
            snprintf(capture->description, sizeof(capture->description),
                    "[%.3f] Other type=%d", ts, t);
        }
        
        capturedEventCount++;
    }
    
    pthread_mutex_unlock(&captureMutex);
    return e;
}

// trackpad_dump.cと同じモニタースレッド
static void* trackpadDumpMonitorThreadFunc(void* arg) {
    captureRunLoop = CFRunLoopGetCurrent();
    
    // trackpad_dump.cと全く同じ設定
    CGEventMask m = CGEventMaskBit(kCGEventScrollWheel)
                  | CGEventMaskBit(29)|CGEventMaskBit(30)|CGEventMaskBit(31);
    
    captureTap = CGEventTapCreate(kCGSessionEventTap,        // trackpad_dump.cと同じ
                                  kCGHeadInsertEventTap,     // trackpad_dump.cと同じ
                                  kCGEventTapOptionListenOnly, // trackpad_dump.cと同じ
                                  m, trackpadDumpCallback, NULL);
    
    if (!captureTap) {
        printf("trackpadDumpMonitor: Need Accessibility permission\n");
        return NULL;
    }
    
    captureRunLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, captureTap, 0);
    CFRunLoopAddSource(captureRunLoop, captureRunLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(captureTap, true);
    
    printf("trackpadDumpMonitor: Listening started\n");
    CFRunLoopRun();
    
    return NULL;
}

// trackpad_dump.c監視開始
void startTrackpadDumpMonitor() {
    if (isCaptureMonitoring) return;
    
    pthread_mutex_lock(&captureMutex);
    capturedEventCount = 0;
    pthread_mutex_unlock(&captureMutex);
    
    isCaptureMonitoring = true;
    pthread_create(&captureMonitorThread, NULL, trackpadDumpMonitorThreadFunc, NULL);
    
    // 起動を待つ
    usleep(200000); // 200ms
}

// trackpad_dump.c監視停止
void stopTrackpadDumpMonitor() {
    if (!isCaptureMonitoring) return;
    
    if (captureRunLoop) {
        CFRunLoopStop(captureRunLoop);
    }
    
    pthread_join(captureMonitorThread, NULL);
    
    if (captureTap) {
        CGEventTapEnable(captureTap, false);
        CFRelease(captureTap);
        captureTap = NULL;
    }
    
    if (captureRunLoopSource) {
        CFRelease(captureRunLoopSource);
        captureRunLoopSource = NULL;
    }
    
    isCaptureMonitoring = false;
}

// キャプチャされたイベント数を取得
int getCapturedEventCount() {
    pthread_mutex_lock(&captureMutex);
    int count = capturedEventCount;
    pthread_mutex_unlock(&captureMutex);
    return count;
}

// 特定タイプのイベント数を取得
int getCapturedEventCountByType(CGEventType type) {
    pthread_mutex_lock(&captureMutex);
    int count = 0;
    for (int i = 0; i < capturedEventCount; i++) {
        if (capturedEvents[i].type == type) {
            count++;
        }
    }
    pthread_mutex_unlock(&captureMutex);
    return count;
}

// サブタイプ別イベント数を取得
int getCapturedEventCountBySubtype(int64_t subtype) {
    pthread_mutex_lock(&captureMutex);
    int count = 0;
    for (int i = 0; i < capturedEventCount; i++) {
        if (capturedEvents[i].subtype == subtype) {
            count++;
        }
    }
    pthread_mutex_unlock(&captureMutex);
    return count;
}

// ジェスチャーイベント数を取得
int getCapturedGestureEventCount() {
    pthread_mutex_lock(&captureMutex);
    int count = 0;
    for (int i = 0; i < capturedEventCount; i++) {
        if (capturedEvents[i].type == 29 || capturedEvents[i].type == 30 || capturedEvents[i].type == 31) {
            count++;
        }
    }
    pthread_mutex_unlock(&captureMutex);
    return count;
}

// trackpad_dump.cと同じフォーマットで全イベントを出力
void printCapturedEvents() {
    pthread_mutex_lock(&captureMutex);
    printf("\n=== Captured Events (trackpad_dump.c format) ===\n");
    for (int i = 0; i < capturedEventCount; i++) {
        printf("%s\n", capturedEvents[i].description);
    }
    printf("=== End of Captured Events (total: %d) ===\n\n", capturedEventCount);
    pthread_mutex_unlock(&captureMutex);
}

// イベントをクリア
void clearCapturedEvents() {
    pthread_mutex_lock(&captureMutex);
    capturedEventCount = 0;
    pthread_mutex_unlock(&captureMutex);
}
*/
import "C"

// TrackpadDumpMonitor はtrackpad_dump.cと全く同じ方法でイベントを監視する
type TrackpadDumpMonitor struct {
	mu sync.Mutex
}

// NewTrackpadDumpMonitor creates a new trackpad dump monitor
func NewTrackpadDumpMonitor() *TrackpadDumpMonitor {
	return &TrackpadDumpMonitor{}
}

// Start begins monitoring with trackpad_dump.c equivalent settings
func (m *TrackpadDumpMonitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	C.startTrackpadDumpMonitor()
	return nil
}

// Stop stops monitoring
func (m *TrackpadDumpMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	C.stopTrackpadDumpMonitor()
}

// GetEventCount returns the total number of captured events
func (m *TrackpadDumpMonitor) GetEventCount() int {
	return int(C.getCapturedEventCount())
}

// GetEventCountByType returns the number of events of a specific type
func (m *TrackpadDumpMonitor) GetEventCountByType(eventType int) int {
	return int(C.getCapturedEventCountByType(C.CGEventType(eventType)))
}

// GetEventCountBySubtype returns the number of events of a specific subtype
func (m *TrackpadDumpMonitor) GetEventCountBySubtype(subtype int64) int {
	return int(C.getCapturedEventCountBySubtype(C.int64_t(subtype)))
}

// GetGestureEventCount returns the number of gesture events (type 29, 30, 31)
func (m *TrackpadDumpMonitor) GetGestureEventCount() int {
	return int(C.getCapturedGestureEventCount())
}

// Clear clears all captured events
func (m *TrackpadDumpMonitor) Clear() {
	C.clearCapturedEvents()
}

// PrintEvents prints all captured events in trackpad_dump.c format
func (m *TrackpadDumpMonitor) PrintEvents() {
	C.printCapturedEvents()
}

// WaitForEvents waits for a specific number of events to be captured
func (m *TrackpadDumpMonitor) WaitForEvents(expectedCount int, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if m.GetEventCount() >= expectedCount {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %d events, got %d", expectedCount, m.GetEventCount())
}

// TestTouchpadWithTrackpadDumpMonitor はtrackpad_dump.cと同じ方法でイベントを検証する
func TestTouchpadWithTrackpadDumpMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}
	
	if runtime.GOOS != "darwin" {
		t.Skip("macOS専用テスト")
	}

	// trackpad_dump.cと同じ監視方法を使用
	monitor := NewTrackpadDumpMonitor()
	err := monitor.Start()
	if err != nil {
		t.Fatalf("trackpad_dump監視の開始に失敗: %v", err)
	}
	defer monitor.Stop()

	cfg := features.TouchPadConfig{
		Name:                  "trackpad-dump-test",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		ScrollScaleFactor:     1.5,
		SwipeScaleFactor:      0.5,
		MotionSmoothingFactor: 0.3,
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	t.Run("4本指スワイプ検証_trackpad_dump形式", func(t *testing.T) {
		monitor.Clear()
		
		t.Log("【重要】4本指スワイプをtrackpad_dump.c形式で監視開始")
		
		// 4本指でタッチダウン
		fingers := []struct {
			slot int
			id   int
			x    int32
			y    int32
		}{
			{0, 5001, 800, 1000},
			{1, 5002, 1000, 1000},
			{2, 5003, 1200, 1000},
			{3, 5004, 1400, 1000},
		}

		// 順次タッチダウン
		for _, finger := range fingers {
			err := touchpad.MultiTouchDown(finger.slot, finger.id, finger.x, finger.y)
			if err != nil {
				t.Fatalf("指%dのタッチダウンに失敗: %v", finger.slot, err)
			}
			time.Sleep(20 * time.Millisecond)
		}

		// ジェスチャー開始を待つ
		time.Sleep(150 * time.Millisecond)

		// 右方向にスワイプ
		for i := 0; i < 8; i++ {
			dx := int32(100)
			for _, finger := range fingers {
				err := touchpad.MultiTouchMove(finger.slot, finger.x+int32(i+1)*dx, finger.y)
				if err != nil {
					t.Errorf("指%dの移動に失敗: %v", finger.slot, err)
				}
			}
			time.Sleep(30 * time.Millisecond)
		}

		// タッチアップ
		for _, finger := range fingers {
			err := touchpad.MultiTouchUp(finger.slot)
			if err != nil {
				t.Errorf("指%dのタッチアップに失敗: %v", finger.slot, err)
			}
		}

		// 十分な処理時間を確保
		time.Sleep(500 * time.Millisecond)

		// trackpad_dump.c形式で結果を出力
		totalEvents := monitor.GetEventCount()
		t.Logf("trackpad_dump.c検出イベント数: %d", totalEvents)
		
		monitor.PrintEvents()

		// 重要な検証
		gestureEvents := monitor.GetGestureEventCount()
		scrollEvents := monitor.GetEventCountByType(22) // kCGEventScrollWheel
		swipeSubtypeEvents := monitor.GetEventCountBySubtype(23) // Swipe subtype

		t.Logf("検証結果:")
		t.Logf("  - 総イベント数: %d", totalEvents)
		t.Logf("  - ジェスチャーイベント数 (type 29,30,31): %d", gestureEvents)
		t.Logf("  - スクロールイベント数 (type 22): %d", scrollEvents)
		t.Logf("  - スワイプサブタイプイベント数 (sub 23): %d", swipeSubtypeEvents)

		// 実際のtrackpad_dump.cログのパターンと比較
		if totalEvents == 0 {
			t.Fatal("❌ trackpad_dump.c形式の監視でイベントが全く検出されませんでした")
		}

		if gestureEvents == 0 {
			t.Error("❌ ジェスチャーイベント (type 29,30,31) が検出されませんでした")
		} else {
			t.Log("✅ ジェスチャーイベントが検出されました")
		}

		if swipeSubtypeEvents == 0 {
			t.Error("❌ スワイプサブタイプ (sub=23) が検出されませんでした")
		} else {
			t.Log("✅ スワイプサブタイプが検出されました")
		}

		// 実際のログと比較して期待値を設定
		if gestureEvents > 0 && swipeSubtypeEvents > 0 {
			t.Log("🎉 4本指スワイプがtrackpad_dump.c形式で正常に検出されました")
		} else {
			t.Error("💥 4本指スワイプの検出に問題があります。touchpad_darwin.goの修正が必要です")
		}
	})

	t.Run("2本指スクロール検証_trackpad_dump形式", func(t *testing.T) {
		monitor.Clear()
		
		t.Log("【参考】2本指スクロールをtrackpad_dump.c形式で監視")
		
		// 2本指スクロール
		err := touchpad.MultiTouchDown(0, 6001, 1000, 1000)
		if err != nil {
			t.Fatalf("スクロール: タッチダウンに失敗: %v", err)
		}
		
		err = touchpad.MultiTouchDown(1, 6002, 1200, 1000)
		if err != nil {
			t.Fatalf("スクロール: タッチダウンに失敗: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// スクロール移動
		for i := 0; i < 5; i++ {
			dy := int32(50)
			err = touchpad.MultiTouchMove(0, 1000, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("スクロール移動に失敗: %v", err)
			}
			err = touchpad.MultiTouchMove(1, 1200, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("スクロール移動に失敗: %v", err)
			}
			time.Sleep(25 * time.Millisecond)
		}

		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("スクロール終了に失敗: %v", err)
		}
		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("スクロール終了に失敗: %v", err)
		}

		time.Sleep(300 * time.Millisecond)

		// 結果出力
		totalEvents := monitor.GetEventCount()
		t.Logf("スクロール検出イベント数: %d", totalEvents)
		
		if totalEvents > 0 {
			monitor.PrintEvents()
			t.Log("✅ 2本指スクロールはtrackpad_dump.c形式で検出されました")
		} else {
			t.Log("⚠️ 2本指スクロールがtrackpad_dump.c形式で検出されませんでした")
		}
	})
}