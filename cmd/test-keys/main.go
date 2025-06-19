package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework ApplicationServices

#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>
#import <pthread.h>
#import <stdio.h>

static int keyPressCount = 0;
static pthread_mutex_t countMutex = PTHREAD_MUTEX_INITIALIZER;

// すべてのキーイベントをカウントするコールバック
CGEventRef keyCountCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout) {
        CGEventTapEnable((CFMachPortRef)refcon, true);
        return event;
    }
    
    if (type == kCGEventKeyDown) {
        CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        pthread_mutex_lock(&countMutex);
        keyPressCount++;
        pthread_mutex_unlock(&countMutex);
        fprintf(stderr, "[TestKeys] KeyDown detected: keyCode=0x%x (%d)\n", keyCode, keyCode);
        fflush(stderr);
    }
    
    return event;
}

int getKeyPressCount() {
    pthread_mutex_lock(&countMutex);
    int count = keyPressCount;
    pthread_mutex_unlock(&countMutex);
    return count;
}

void testEventTap() {
    fprintf(stderr, "[TestKeys] Starting event tap test...\n");
    fflush(stderr);
    
    // アクセシビリティ権限をチェック
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    Boolean accessibilityEnabled = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    if (!accessibilityEnabled) {
        fprintf(stderr, "[TestKeys] Accessibility permission denied\n");
        fflush(stderr);
        return;
    }
    fprintf(stderr, "[TestKeys] Accessibility permission granted\n");
    fflush(stderr);
    
    CGEventMask eventMask = (1 << kCGEventKeyDown);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGHIDEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        eventMask,
        keyCountCallback,
        NULL
    );
    
    if (!eventTap) {
        fprintf(stderr, "[TestKeys] Failed to create event tap\n");
        fflush(stderr);
        return;
    }
    fprintf(stderr, "[TestKeys] Event tap created\n");
    fflush(stderr);
    
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    
    // Store eventTap reference for timeout re-enable
    keyCountCallback(NULL, kCGEventTapDisabledByTimeout, NULL, eventTap);
    
    fprintf(stderr, "[TestKeys] Event tap enabled, monitoring key presses...\n");
    fflush(stderr);
    
    // 10秒間実行
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 10 * NSEC_PER_SEC), dispatch_get_main_queue(), ^{
        CFRunLoopStop(CFRunLoopGetCurrent());
    });
    
    CFRunLoopRun();
    
    // クリーンアップ
    CGEventTapEnable(eventTap, false);
    CFRunLoopRemoveSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CFRelease(runLoopSource);
    CFRelease(eventTap);
    
    fprintf(stderr, "[TestKeys] Event tap test completed\n");
    fflush(stderr);
}
*/
import "C"

func main() {
	log.Println("Key event tap test")
	log.Println("This will monitor all key presses for 10 seconds")
	log.Println("Press any keys to test...")
	
	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// バックグラウンドでイベントタップをテスト
	go func() {
		C.testEventTap()
	}()
	
	// 定期的にカウントを表示
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(10 * time.Second)
	
	for {
		select {
		case <-sigCh:
			log.Println("Interrupted")
			return
		case <-ticker.C:
			count := int(C.getKeyPressCount())
			fmt.Printf("Total key presses detected: %d\n", count)
		case <-timeout:
			log.Println("Test completed")
			finalCount := int(C.getKeyPressCount())
			fmt.Printf("Final count: %d key presses detected\n", finalCount)
			if finalCount == 0 {
				log.Println("WARNING: No key presses detected. Check accessibility permissions.")
			}
			return
		}
	}
}