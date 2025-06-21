//go:build darwin
// +build darwin

package features

import (
	"fmt"
	"log"
	"sync"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework ApplicationServices

#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>
#import <pthread.h>
#import <stdio.h>

// キーコードの定義（macOSのキーコード）
#define kVK_F1 0x7A
#define kVK_F2 0x78
#define kVK_F13 0x69
#define kVK_F14 0x6B

// グローバル変数
static int32_t currentKey = -1;
static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static CFRunLoopRef runLoop = NULL;
static pthread_t runLoopThread;
static bool isRunning = false;
static bool gestureActive = false;  // ジェスチャーがアクティブかどうか
static pthread_mutex_t keyMutex = PTHREAD_MUTEX_INITIALIZER;

// イベントコールバック
CGEventRef keyEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // タイムアウトでタップが無効になった場合、再有効化
    if (type == kCGEventTapDisabledByTimeout) {
        CGEventTapEnable(eventTap, true);
        return event;
    }
    
    if (type == kCGEventKeyDown) {
        CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        
        // F13/F14キーの場合、記録してイベントを消費
        if (keyCode == kVK_F13 || keyCode == kVK_F14) {
            pthread_mutex_lock(&keyMutex);
            currentKey = (int32_t)keyCode;
            gestureActive = true;
            pthread_mutex_unlock(&keyMutex);
            
            fprintf(stderr, "[KeyboardDarwin] F13/F14 KeyDown detected and consumed: keyCode=0x%x\n", keyCode);
            fflush(stderr);
            
            return NULL;  // イベントを消費（システムに伝わらない）
        }
        
        // F1/F2キーの場合は記録のみ
        if (keyCode == kVK_F1 || keyCode == kVK_F2) {
            pthread_mutex_lock(&keyMutex);
            currentKey = (int32_t)keyCode;
            pthread_mutex_unlock(&keyMutex);
            fprintf(stderr, "[KeyboardDarwin] Function key detected: keyCode=0x%x\n", keyCode);
            fflush(stderr);
        }
    } else if (type == kCGEventKeyUp) {
        CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        
        pthread_mutex_lock(&keyMutex);
        bool wasOurKey = (currentKey == (int32_t)keyCode);
        bool wasActive = gestureActive;
        
        if (wasOurKey) {
            currentKey = -1;
            if (keyCode == kVK_F13 || keyCode == kVK_F14) {
                gestureActive = false;
            }
            fprintf(stderr, "[KeyboardDarwin] Function key released: keyCode=0x%x\n", keyCode);
            fflush(stderr);
        }
        pthread_mutex_unlock(&keyMutex);
        
        // F13/F14キーのリリースも消費
        if ((keyCode == kVK_F13 || keyCode == kVK_F14) && wasOurKey && wasActive) {
            fprintf(stderr, "[KeyboardDarwin] F13/F14 KeyUp consumed\n");
            fflush(stderr);
            return NULL;
        }
    }
    
    // その他のイベントは通過させる
    return event;
}

// RunLoopを実行するスレッド関数
void* runLoopThreadFunc(void* arg) {
    @autoreleasepool {
        // スレッドに名前を設定
        [[NSThread currentThread] setName:@"com.keyball.gestures.keyboard-monitor"];
        
        runLoop = CFRunLoopGetCurrent();
        
        // イベントタップを作成
        CGEventMask eventMask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp);
        fprintf(stderr, "[KeyboardDarwin] Creating event tap...\n");
        fflush(stderr);
        eventTap = CGEventTapCreate(
            kCGHIDEventTap,             // 最低レベルでタップ
            kCGHeadInsertEventTap,       // 最初に処理
            kCGEventTapOptionDefault,    // イベントを消費可能
            eventMask,
            keyEventCallback,
            NULL
        );
        
        if (!eventTap) {
            fprintf(stderr, "[KeyboardDarwin] Failed to create event tap\n");
            fflush(stderr);
            return NULL;
        }
        fprintf(stderr, "[KeyboardDarwin] Event tap created successfully\n");
        fflush(stderr);
        
        // RunLoopソースを作成
        runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
        CFRunLoopAddSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
        
        // イベントタップを有効化
        CGEventTapEnable(eventTap, true);
        
        // RunLoopを実行
        fprintf(stderr, "[KeyboardDarwin] Starting RunLoop...\n");
        fflush(stderr);
        CFRunLoopRun();
        fprintf(stderr, "[KeyboardDarwin] RunLoop stopped\n");
        fflush(stderr);
        
        // クリーンアップ（通常ここには到達しない）
        if (runLoopSource) {
            CFRunLoopRemoveSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
            CFRelease(runLoopSource);
        }
        if (eventTap) {
            CFRelease(eventTap);
        }
    }
    return NULL;
}

// キーボード監視を開始
int startKeyboardMonitoring() {
    if (isRunning) {
        return 1;
    }
    
    // アクセシビリティ権限をチェック
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean accessibilityEnabled = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    if (!accessibilityEnabled) {
        return 0;
    }
    
    // RunLoopスレッドを開始
    int result = pthread_create(&runLoopThread, NULL, runLoopThreadFunc, NULL);
    if (result != 0) {
        return 0;
    }
    
    isRunning = true;
    return 1;
}

// キーボード監視を停止
void stopKeyboardMonitoring() {
    if (!isRunning) {
        return;
    }
    
    isRunning = false;
    
    // RunLoopを停止
    if (runLoop) {
        CFRunLoopStop(runLoop);
    }
    
    // スレッドの終了を待つ
    pthread_join(runLoopThread, NULL);
    
    pthread_mutex_lock(&keyMutex);
    currentKey = -1;
    pthread_mutex_unlock(&keyMutex);
    
    eventTap = NULL;
    runLoopSource = NULL;
    runLoop = NULL;
}

// 現在押されているキーを取得
int32_t getCurrentKey() {
    pthread_mutex_lock(&keyMutex);
    int32_t key = currentKey;
    pthread_mutex_unlock(&keyMutex);
    return key;
}

// F1キーのキーコード (macOS)
int32_t getF1KeyCode() {
    return kVK_F1;
}

// F2キーのキーコード (macOS)
int32_t getF2KeyCode() {
    return kVK_F2;
}

// F13キーのキーコード (macOS)
int32_t getF13KeyCode() {
    return kVK_F13;
}

// F14キーのキーコード (macOS)
int32_t getF14KeyCode() {
    return kVK_F14;
}
*/
import "C"

type darwinKeyboard struct {
	mu       sync.Mutex
	deviceID string
	running  bool
}

// CreateKeyboard macOS実装：CGEventTapを使用してキーボードイベントを監視
func CreateKeyboard(path string) (Keyboard, error) {
	// macOSではパスは使用されない（deviceIDとして扱う）
	kb := &darwinKeyboard{
		deviceID: path,
	}
	
	// CGEventTapを開始
	if C.startKeyboardMonitoring() == 0 {
		return nil, fmt.Errorf("キーボード監視の開始に失敗しました。アクセシビリティ権限を確認してください")
	}
	
	kb.running = true
	log.Println("macOSキーボード監視を開始しました")
	return kb, nil
}

// GetKey 現在押されているキーを取得
func (kb *darwinKeyboard) GetKey() int32 {
	if !kb.running {
		return -1
	}
	
	// Cコードから現在のキーを取得
	currentKey := int32(C.getCurrentKey())
	
	// macOSのキーコードをLinux互換のキーコードに変換
	// F1 (macOS: 0x7A) -> Linux: 122
	// F2 (macOS: 0x78) -> Linux: 120
	// F13 (macOS: 0x69) -> Linux: 183
	// F14 (macOS: 0x6B) -> Linux: 184
	if currentKey == int32(C.getF1KeyCode()) {
		return 122  // Linux F1
	} else if currentKey == int32(C.getF2KeyCode()) {
		return 120  // Linux F2
	} else if currentKey == int32(C.getF13KeyCode()) {
		return 183  // Linux F13
	} else if currentKey == int32(C.getF14KeyCode()) {
		return 184  // Linux F14
	}
	
	return -1
}

// Close キーボード監視を終了
func (kb *darwinKeyboard) Close() error {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	
	if kb.running {
		C.stopKeyboardMonitoring()
		kb.running = false
		log.Println("macOSキーボード監視を停止しました")
	}
	
	return nil
}