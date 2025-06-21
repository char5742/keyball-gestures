//go:build darwin
// +build darwin

package features

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework ApplicationServices

#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>
#import <pthread.h>

// マウスのデルタ値を保持
static int32_t mouseDeltaX = 0;
static int32_t mouseDeltaY = 0;
static CFMachPortRef mouseEventTap = NULL;
static CFRunLoopSourceRef mouseRunLoopSource = NULL;
static CFRunLoopRef mouseRunLoop = NULL;
static pthread_t mouseRunLoopThread;
static bool mouseGrabbed = false;
static bool mouseIsRunning = false;
static pthread_mutex_t mouseMutex = PTHREAD_MUTEX_INITIALIZER;

// マウスイベントコールバック
CGEventRef mouseEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // タイムアウトでタップが無効になった場合、再有効化
    if (type == kCGEventTapDisabledByTimeout) {
        CGEventTapEnable(mouseEventTap, true);
        return event;
    }
    
    if (type == kCGEventMouseMoved || 
        type == kCGEventLeftMouseDragged || 
        type == kCGEventRightMouseDragged || 
        type == kCGEventOtherMouseDragged) {
        
        // デルタ値を取得
        int64_t deltaX = CGEventGetIntegerValueField(event, kCGMouseEventDeltaX);
        int64_t deltaY = CGEventGetIntegerValueField(event, kCGMouseEventDeltaY);
        
        pthread_mutex_lock(&mouseMutex);
        // 累積
        mouseDeltaX += (int32_t)deltaX;
        mouseDeltaY += (int32_t)deltaY;
        pthread_mutex_unlock(&mouseMutex);
        
        // Grabモードの場合、イベントを消費
        if (mouseGrabbed) {
            fprintf(stderr, "[MouseDarwin] Mouse grabbed, consuming event: dx=%lld, dy=%lld\n", deltaX, deltaY);
            fflush(stderr);
            return NULL;  // イベントを消費
        }
    }
    
    // イベントを通過させる
    return event;
}

// RunLoopを実行するスレッド関数
void* mouseRunLoopThreadFunc(void* arg) {
    @autoreleasepool {
        // スレッドに名前を設定
        [[NSThread currentThread] setName:@"com.keyball.gestures.mouse-monitor"];
        
        mouseRunLoop = CFRunLoopGetCurrent();
        
        // イベントタップを作成
        CGEventMask eventMask = CGEventMaskBit(kCGEventMouseMoved) |
                               CGEventMaskBit(kCGEventLeftMouseDragged) |
                               CGEventMaskBit(kCGEventRightMouseDragged) |
                               CGEventMaskBit(kCGEventOtherMouseDragged);
        
        mouseEventTap = CGEventTapCreate(
            kCGHIDEventTap,           // 最低レベルでタップ
            kCGHeadInsertEventTap,     // 最初に処理
            kCGEventTapOptionDefault,  // Grabモード時はイベントを消費可能
            eventMask,
            mouseEventCallback,
            NULL
        );
        
        if (!mouseEventTap) {
            return NULL;
        }
        
        // RunLoopソースを作成
        mouseRunLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, mouseEventTap, 0);
        CFRunLoopAddSource(mouseRunLoop, mouseRunLoopSource, kCFRunLoopCommonModes);
        
        // イベントタップを有効化
        CGEventTapEnable(mouseEventTap, true);
        
        // RunLoopを実行
        CFRunLoopRun();
        
        // クリーンアップ（通常ここには到達しない）
        if (mouseRunLoopSource) {
            CFRunLoopRemoveSource(mouseRunLoop, mouseRunLoopSource, kCFRunLoopCommonModes);
            CFRelease(mouseRunLoopSource);
        }
        if (mouseEventTap) {
            CFRelease(mouseEventTap);
        }
    }
    return NULL;
}

// マウス監視を開始
int startMouseMonitoring() {
    if (mouseIsRunning) {
        return 1;
    }
    
    // アクセシビリティ権限をチェック
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean accessibilityEnabled = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    if (!accessibilityEnabled) {
        return 0;
    }
    
    // RunLoopスレッドを開始
    int result = pthread_create(&mouseRunLoopThread, NULL, mouseRunLoopThreadFunc, NULL);
    if (result != 0) {
        return 0;
    }
    
    mouseIsRunning = true;
    return 1;
}

// マウス監視を停止
void stopMouseMonitoring() {
    if (!mouseIsRunning) {
        return;
    }
    
    mouseIsRunning = false;
    
    // RunLoopを停止
    if (mouseRunLoop) {
        CFRunLoopStop(mouseRunLoop);
    }
    
    // スレッドの終了を待つ
    pthread_join(mouseRunLoopThread, NULL);
    
    pthread_mutex_lock(&mouseMutex);
    mouseDeltaX = 0;
    mouseDeltaY = 0;
    mouseGrabbed = false;
    pthread_mutex_unlock(&mouseMutex);
    
    mouseEventTap = NULL;
    mouseRunLoopSource = NULL;
    mouseRunLoop = NULL;
}

// マウスデルタを取得してリセット
void getAndResetMouseDelta(int32_t* dx, int32_t* dy) {
    pthread_mutex_lock(&mouseMutex);
    *dx = mouseDeltaX;
    *dy = mouseDeltaY;
    mouseDeltaX = 0;
    mouseDeltaY = 0;
    pthread_mutex_unlock(&mouseMutex);
}

// マウスをグラブ
void setMouseGrabbed(bool grabbed) {
    pthread_mutex_lock(&mouseMutex);
    mouseGrabbed = grabbed;
    pthread_mutex_unlock(&mouseMutex);
    fprintf(stderr, "[MouseDarwin] Mouse grabbed state changed: %s\n", grabbed ? "true" : "false");
    fflush(stderr);
}

// マウスグラブ状態を取得
bool isMouseGrabbed() {
    pthread_mutex_lock(&mouseMutex);
    bool grabbed = mouseGrabbed;
    pthread_mutex_unlock(&mouseMutex);
    return grabbed;
}
*/
import "C"

type darwinMouse struct {
	mu       sync.Mutex
	deviceID string
	running  bool
}

// CreateMouse macOS実装：CGEventTapを使用してマウスイベントを監視
func CreateMouse(path string) (Mouse, error) {
	// macOSではパスは使用されない（deviceIDとして扱う）
	m := &darwinMouse{
		deviceID: path,
	}
	
	// CGEventTapを開始
	if C.startMouseMonitoring() == 0 {
		return nil, fmt.Errorf("マウス監視の開始に失敗しました。アクセシビリティ権限を確認してください")
	}
	
	m.running = true
	log.Println("macOSマウス監視を開始しました")
	return m, nil
}

// HandleSignals シグナルハンドリング
func (m *darwinMouse) HandleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutting down...")
		m.Close()
		os.Exit(0)
	}()
}

// GetMouseDelta マウスの移動量を取得
func (m *darwinMouse) GetMouseDelta() (dx int32, dy int32) {
	if !m.running {
		return 0, 0
	}
	
	var cDx, cDy C.int32_t
	C.getAndResetMouseDelta(&cDx, &cDy)
	
	return int32(cDx), int32(cDy)
}

// Grab マウス操作を専有
func (m *darwinMouse) Grab() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return fmt.Errorf("マウス監視が実行されていません")
	}
	
	C.setMouseGrabbed(true)
	log.Println("マウスをグラブしました")
	return nil
}

// Release マウス操作の専有を解除
func (m *darwinMouse) Release() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return fmt.Errorf("マウス監視が実行されていません")
	}
	
	C.setMouseGrabbed(false)
	log.Println("マウスのグラブを解除しました")
	return nil
}

// Close マウス監視を終了
func (m *darwinMouse) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		C.stopMouseMonitoring()
		m.running = false
		log.Println("macOSマウス監視を停止しました")
	}
	
	return nil
}