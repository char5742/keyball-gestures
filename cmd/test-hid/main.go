package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

// 全てのHIDイベントを表示するコールバック
void allHIDCallback(void *context, IOReturn result, void *sender, IOHIDValueRef value) {
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
    
    if (intValue != 0) {  // 値が0でない場合のみ表示
        fprintf(stderr, "[HID] Device: %s, UsagePage: 0x%04x, Usage: 0x%04x, Value: %ld\n", 
                deviceName, usagePage, usage, (long)intValue);
        
        // キーボードイベントの場合、追加情報を表示
        if (usagePage == kHIDPage_KeyboardOrKeypad) {
            fprintf(stderr, "      ^ Keyboard key: 0x%x (decimal: %d)\n", usage, usage);
            if (usage == 0x68) fprintf(stderr, "      ^ This is F13!\n");
            if (usage == 0x69) fprintf(stderr, "      ^ This is F14!\n");
        }
        
        fflush(stderr);
    }
}

void testAllHID() {
    fprintf(stderr, "[HID] Creating HID Manager for all devices...\n");
    fflush(stderr);
    
    IOHIDManagerRef hidManager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
    if (!hidManager) {
        fprintf(stderr, "[HID] Failed to create HID Manager\n");
        fflush(stderr);
        return;
    }
    
    // すべてのHIDデバイスを監視（フィルタなし）
    IOHIDManagerSetDeviceMatching(hidManager, NULL);
    
    // コールバックを登録
    IOHIDManagerRegisterInputValueCallback(hidManager, allHIDCallback, NULL);
    
    // RunLoopに登録
    IOHIDManagerScheduleWithRunLoop(hidManager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
    
    // デバイスを開く
    IOReturn ret = IOHIDManagerOpen(hidManager, kIOHIDOptionsTypeNone);
    if (ret != kIOReturnSuccess) {
        fprintf(stderr, "[HID] Failed to open HID Manager: 0x%x\n", ret);
        fflush(stderr);
        CFRelease(hidManager);
        return;
    }
    
    // 接続されたデバイスを表示
    CFSetRef devices = IOHIDManagerCopyDevices(hidManager);
    if (devices) {
        CFIndex count = CFSetGetCount(devices);
        fprintf(stderr, "[HID] Found %ld HID devices\n", count);
        
        CFTypeRef values[count];
        CFSetGetValues(devices, values);
        for (CFIndex i = 0; i < count; i++) {
            IOHIDDeviceRef device = (IOHIDDeviceRef)values[i];
            
            // 製品名
            CFStringRef product = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDProductKey));
            char productName[256] = "Unknown";
            if (product) {
                CFStringGetCString(product, productName, sizeof(productName), kCFStringEncodingUTF8);
            }
            
            // Usage PageとUsageを取得
            CFNumberRef usagePageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsagePageKey));
            CFNumberRef usageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsageKey));
            
            int usagePage = 0, usage = 0;
            if (usagePageRef) CFNumberGetValue(usagePageRef, kCFNumberIntType, &usagePage);
            if (usageRef) CFNumberGetValue(usageRef, kCFNumberIntType, &usage);
            
            fprintf(stderr, "[HID] Device %ld: %s (UsagePage: 0x%04x, Usage: 0x%04x)\n", 
                    i, productName, usagePage, usage);
        }
        CFRelease(devices);
    }
    fflush(stderr);
    
    fprintf(stderr, "[HID] Monitoring all HID events. Press Ctrl+C to stop.\n");
    fprintf(stderr, "[HID] Try pressing F13/F14 keys on your Keyball39...\n");
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
	log.Println("HID Event Monitor")
	log.Println("This will show ALL HID events from ALL devices")
	log.Println("Press F13/F14 on your Keyball39 to test...")
	
	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigCh
		fmt.Println("\nStopping...")
		os.Exit(0)
	}()
	
	// HIDテストを実行
	C.testAllHID()
}