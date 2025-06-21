//go:build integration && darwin
// +build integration,darwin

package integration

import (
	"fmt"
	"sync"
	"time"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>

// イベント記録用構造体
typedef struct {
    CGEventType type;
    double timestamp;
    int64_t deltaX;
    int64_t deltaY;
    CGPoint location;
} RecordedEvent;

// グローバル変数
static RecordedEvent events[1000];
static int eventCount = 0;
static pthread_mutex_t eventMutex = PTHREAD_MUTEX_INITIALIZER;
static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static CFRunLoopRef runLoop = NULL;
static pthread_t monitorThread;
static bool isMonitoring = false;

// イベントコールバック
static CGEventRef testEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    pthread_mutex_lock(&eventMutex);
    
    if (eventCount < 1000) {
        RecordedEvent *rec = &events[eventCount];
        rec->type = type;
        rec->timestamp = (double)CGEventGetTimestamp(event) / 1000000000.0;
        rec->location = CGEventGetLocation(event);
        
        // スクロールイベントの場合、デルタ値を記録
        if (type == kCGEventScrollWheel) {
            rec->deltaY = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
            rec->deltaX = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis2);
        } else {
            rec->deltaX = 0;
            rec->deltaY = 0;
        }
        
        eventCount++;
    }
    
    pthread_mutex_unlock(&eventMutex);
    return event;
}

// モニタースレッド
static void* monitorThreadFunc(void* arg) {
    runLoop = CFRunLoopGetCurrent();
    
    // イベントマスクを作成
    CGEventMask eventMask = 0;
    eventMask |= CGEventMaskBit(kCGEventLeftMouseDown);
    eventMask |= CGEventMaskBit(kCGEventRightMouseDown);
    eventMask |= CGEventMaskBit(kCGEventScrollWheel);
    eventMask |= CGEventMaskBit(29); // ジェスチャー開始
    eventMask |= CGEventMaskBit(30); // ジェスチャー変更
    eventMask |= CGEventMaskBit(31); // ジェスチャー終了
    
    // イベントタップを作成
    eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        eventMask,
        testEventCallback,
        NULL
    );
    
    if (!eventTap) {
        return NULL;
    }
    
    // RunLoopソースを作成
    runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
    
    // イベントタップを有効化
    CGEventTapEnable(eventTap, true);
    
    // RunLoopを実行
    CFRunLoopRun();
    
    return NULL;
}

// モニター開始
void startEventMonitor() {
    if (isMonitoring) return;
    
    pthread_mutex_lock(&eventMutex);
    eventCount = 0;
    pthread_mutex_unlock(&eventMutex);
    
    isMonitoring = true;
    pthread_create(&monitorThread, NULL, monitorThreadFunc, NULL);
    
    // モニターが起動するまで少し待つ
    usleep(100000); // 100ms
}

// モニター停止
void stopEventMonitor() {
    if (!isMonitoring) return;
    
    if (runLoop) {
        CFRunLoopStop(runLoop);
    }
    
    pthread_join(monitorThread, NULL);
    
    if (eventTap) {
        CGEventTapEnable(eventTap, false);
        CFRelease(eventTap);
        eventTap = NULL;
    }
    
    if (runLoopSource) {
        CFRelease(runLoopSource);
        runLoopSource = NULL;
    }
    
    isMonitoring = false;
}

// イベント数を取得
int getEventCount() {
    pthread_mutex_lock(&eventMutex);
    int count = eventCount;
    pthread_mutex_unlock(&eventMutex);
    return count;
}

// 特定のタイプのイベント数を取得
int getEventCountByType(CGEventType type) {
    pthread_mutex_lock(&eventMutex);
    int count = 0;
    for (int i = 0; i < eventCount; i++) {
        if (events[i].type == type) {
            count++;
        }
    }
    pthread_mutex_unlock(&eventMutex);
    return count;
}

// スクロールイベントがあるかチェック
int hasScrollEvents() {
    return getEventCountByType(kCGEventScrollWheel) > 0;
}

// ジェスチャーイベントがあるかチェック
int hasGestureEvents() {
    int count = 0;
    count += getEventCountByType(29);
    count += getEventCountByType(30);
    count += getEventCountByType(31);
    return count > 0;
}

// イベントをクリア
void clearEvents() {
    pthread_mutex_lock(&eventMutex);
    eventCount = 0;
    pthread_mutex_unlock(&eventMutex);
}

// デバッグ用：全イベントを出力
void printAllEvents() {
    pthread_mutex_lock(&eventMutex);
    for (int i = 0; i < eventCount; i++) {
        RecordedEvent *e = &events[i];
        printf("[%.2f] Type=%d, Location=(%.1f,%.1f), Delta=(%lld,%lld)\n",
               e->timestamp, e->type, e->location.x, e->location.y, e->deltaX, e->deltaY);
    }
    pthread_mutex_unlock(&eventMutex);
}
*/
import "C"

// EventMonitor はCGEventTapを使用してイベントを監視する
type EventMonitor struct {
	mu sync.Mutex
}

// NewEventMonitor creates a new event monitor
func NewEventMonitor() *EventMonitor {
	return &EventMonitor{}
}

// Start begins monitoring events
func (m *EventMonitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	C.startEventMonitor()
	return nil
}

// Stop stops monitoring events
func (m *EventMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	C.stopEventMonitor()
}

// GetEventCount returns the total number of recorded events
func (m *EventMonitor) GetEventCount() int {
	return int(C.getEventCount())
}

// HasScrollEvents checks if any scroll events were recorded
func (m *EventMonitor) HasScrollEvents() bool {
	return C.hasScrollEvents() > 0
}

// HasGestureEvents checks if any gesture events were recorded
func (m *EventMonitor) HasGestureEvents() bool {
	return C.hasGestureEvents() > 0
}

// Clear clears all recorded events
func (m *EventMonitor) Clear() {
	C.clearEvents()
}

// WaitForEvents waits for events to be recorded
func (m *EventMonitor) WaitForEvents(timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if m.GetEventCount() > 0 {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for events")
}

// PrintDebugInfo prints all recorded events for debugging
func (m *EventMonitor) PrintDebugInfo() {
	C.printAllEvents()
}