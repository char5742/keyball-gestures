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
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation

#import <IOKit/IOKitLib.h>
#import <IOKit/hid/IOHIDManager.h>
#import <IOKit/hid/IOHIDKeys.h>
#import <IOKit/hid/IOHIDUsageTables.h>
#import <CoreFoundation/CoreFoundation.h>
#import <stdio.h>
#import <string.h>

static int eventCount = 0;

// Keyball39からのHIDイベントを処理するコールバック
void keyballEventCallback(void *context, IOReturn result, void *sender, IOHIDValueRef value) {
    if (result != kIOReturnSuccess) {
        return;
    }
    
    IOHIDElementRef element = IOHIDValueGetElement(value);
    uint32_t usagePage = IOHIDElementGetUsagePage(element);
    uint32_t usage = IOHIDElementGetUsage(element);
    CFIndex intValue = IOHIDValueGetIntegerValue(value);
    
    // デバイス情報を取得
    IOHIDDeviceRef device = IOHIDElementGetDevice(element);
    CFStringRef product = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDProductKey));
    char deviceName[256] = "Unknown";
    if (product) {
        CFStringGetCString(product, deviceName, sizeof(deviceName), kCFStringEncodingUTF8);
    }
    
    // Keyball39からのイベントのみ処理
    if (strstr(deviceName, "Keyball39") == NULL) {
        return;
    }
    
    eventCount++;
    
    // 値が変化したイベントを表示
    if (intValue != 0 || usagePage == kHIDPage_KeyboardOrKeypad) {
        fprintf(stderr, "[Event #%d] Page=0x%04x Usage=0x%04x Value=%ld ", 
                eventCount, usagePage, usage, (long)intValue);
        
        // イベントタイプの解釈
        if (usagePage == kHIDPage_KeyboardOrKeypad) {
            fprintf(stderr, "KEYBOARD");
            if (usage >= 0x04 && usage <= 0x27) {  // A-Z, 0-9
                fprintf(stderr, " (Key: '%c')", (usage >= 0x04 && usage <= 0x1D) ? 'A' + (usage - 0x04) : '0' + (usage - 0x1E));
            } else if (usage == 0x68) {
                fprintf(stderr, " (F13)");
            } else if (usage == 0x69) {
                fprintf(stderr, " (F14)");
            } else if (usage >= 0x3A && usage <= 0x45) {  // F1-F12
                fprintf(stderr, " (F%d)", usage - 0x39);
            }
        } else if (usagePage == 0xff60) {
            fprintf(stderr, "CUSTOM/VENDOR");
        } else if (usagePage == kHIDPage_Button) {
            fprintf(stderr, "BUTTON");
        } else if (usagePage == kHIDPage_GenericDesktop) {
            if (usage == kHIDUsage_GD_X) {
                fprintf(stderr, "MOUSE_X");
            } else if (usage == kHIDUsage_GD_Y) {
                fprintf(stderr, "MOUSE_Y");
            } else if (usage == kHIDUsage_GD_Wheel) {
                fprintf(stderr, "WHEEL");
            }
        }
        
        fprintf(stderr, "\n");
        fflush(stderr);
    }
}

void monitorKeyball() {
    fprintf(stderr, "[Monitor] Starting Keyball39 monitor...\n");
    fflush(stderr);
    
    IOHIDManagerRef hidManager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
    if (!hidManager) {
        fprintf(stderr, "[Monitor] Failed to create HID Manager\n");
        fflush(stderr);
        return;
    }
    
    // すべてのデバイスを受け入れる（後でフィルタリング）
    IOHIDManagerSetDeviceMatching(hidManager, NULL);
    
    // コールバックを登録
    IOHIDManagerRegisterInputValueCallback(hidManager, keyballEventCallback, NULL);
    
    // RunLoopに登録
    IOHIDManagerScheduleWithRunLoop(hidManager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
    
    // デバイスを開く
    IOReturn ret = IOHIDManagerOpen(hidManager, kIOHIDOptionsTypeNone);
    if (ret != kIOReturnSuccess) {
        fprintf(stderr, "[Monitor] Failed to open HID Manager: 0x%x\n", ret);
        fflush(stderr);
        CFRelease(hidManager);
        return;
    }
    
    // Keyball39デバイスを表示
    CFSetRef devices = IOHIDManagerCopyDevices(hidManager);
    if (devices) {
        CFIndex totalCount = CFSetGetCount(devices);
        int keyballCount = 0;
        
        CFTypeRef values[totalCount];
        CFSetGetValues(devices, values);
        
        fprintf(stderr, "\n[Monitor] Checking %ld devices for Keyball39...\n", totalCount);
        
        for (CFIndex i = 0; i < totalCount; i++) {
            IOHIDDeviceRef device = (IOHIDDeviceRef)values[i];
            CFStringRef product = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDProductKey));
            
            if (product) {
                char productName[256];
                if (CFStringGetCString(product, productName, sizeof(productName), kCFStringEncodingUTF8)) {
                    if (strstr(productName, "Keyball39") != NULL) {
                        keyballCount++;
                        
                        CFNumberRef usagePageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsagePageKey));
                        CFNumberRef usageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsageKey));
                        
                        int usagePage = 0, usage = 0;
                        if (usagePageRef) CFNumberGetValue(usagePageRef, kCFNumberIntType, &usagePage);
                        if (usageRef) CFNumberGetValue(usageRef, kCFNumberIntType, &usage);
                        
                        fprintf(stderr, "  ✓ Found: %s (0x%04x:0x%04x)\n", productName, usagePage, usage);
                    }
                }
            }
        }
        
        fprintf(stderr, "\n[Monitor] Found %d Keyball39 device(s)\n", keyballCount);
        CFRelease(devices);
    }
    
    fprintf(stderr, "\n[Monitor] Ready. Press keys on Keyball39...\n");
    fprintf(stderr, "[Monitor] Looking for F13 (0x68) and F14 (0x69) keys\n");
    fprintf(stderr, "[Monitor] Press Ctrl+C to stop\n\n");
    fflush(stderr);
    
    // RunLoopを実行
    CFRunLoopRun();
    
    // クリーンアップ
    IOHIDManagerClose(hidManager, kIOHIDOptionsTypeNone);
    IOHIDManagerUnscheduleFromRunLoop(hidManager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
    CFRelease(hidManager);
}
*/
import "C"

func main() {
	log.Println("=== Keyball39 Specific Monitor ===")
	log.Println("This will monitor ONLY Keyball39 events")
	
	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// 定期的な状態表示
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	go func() {
		for {
			select {
			case <-sigCh:
				fmt.Println("\nStopping...")
				os.Exit(0)
			case <-ticker.C:
				// 5秒ごとに待機中メッセージ
				fmt.Println("[Monitor] Still monitoring... Press F13/F14 on Keyball39")
			}
		}
	}()
	
	// モニタリング開始
	C.monitorKeyball()
}