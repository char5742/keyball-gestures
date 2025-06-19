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

// Keyball39からのすべてのHIDイベントを表示するコールバック
void keyballHIDCallback(void *context, IOReturn result, void *sender, IOHIDValueRef value) {
    if (result != kIOReturnSuccess) {
        return;
    }
    
    IOHIDElementRef element = IOHIDValueGetElement(value);
    uint32_t usagePage = IOHIDElementGetUsagePage(element);
    uint32_t usage = IOHIDElementGetUsage(element);
    CFIndex intValue = IOHIDValueGetIntegerValue(value);
    IOHIDElementCookie cookie = IOHIDElementGetCookie(element);
    
    // デバイス情報を取得
    IOHIDDeviceRef device = IOHIDElementGetDevice(element);
    CFStringRef product = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDProductKey));
    char deviceName[256] = "Unknown";
    if (product) {
        CFStringGetCString(product, deviceName, sizeof(deviceName), kCFStringEncodingUTF8);
    }
    
    // デバイスのUsagePageとUsageを取得
    CFNumberRef deviceUsagePageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsagePageKey));
    CFNumberRef deviceUsageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsageKey));
    int deviceUsagePage = 0, deviceUsage = 0;
    if (deviceUsagePageRef) CFNumberGetValue(deviceUsagePageRef, kCFNumberIntType, &deviceUsagePage);
    if (deviceUsageRef) CFNumberGetValue(deviceUsageRef, kCFNumberIntType, &deviceUsage);
    
    // すべてのイベントを表示（値が変化したもののみ）
    if (intValue != 0 || usagePage == kHIDPage_KeyboardOrKeypad) {
        fprintf(stderr, "[Keyball] Device: %s (0x%04x:0x%04x), Element: Page=0x%04x Usage=0x%04x Cookie=%d Value=%ld\n", 
                deviceName, deviceUsagePage, deviceUsage, usagePage, usage, (int)cookie, (long)intValue);
        
        // 特定のイベントタイプの詳細情報
        if (usagePage == kHIDPage_KeyboardOrKeypad) {
            fprintf(stderr, "         ^ Keyboard event - Key code: 0x%02x (decimal: %d)\n", usage, usage);
            if (usage == 0x68) fprintf(stderr, "         ^ This is F13!\n");
            if (usage == 0x69) fprintf(stderr, "         ^ This is F14!\n");
        } else if (usagePage == 0xff60) {
            fprintf(stderr, "         ^ Custom HID event from Keyball vendor page\n");
        } else if (usagePage == kHIDPage_Button) {
            fprintf(stderr, "         ^ Button event\n");
        } else if (usagePage == kHIDPage_GenericDesktop) {
            if (usage == kHIDUsage_GD_X || usage == kHIDUsage_GD_Y) {
                fprintf(stderr, "         ^ Mouse movement\n");
            }
        }
        
        fflush(stderr);
    }
}

void testKeyballHID() {
    fprintf(stderr, "[Keyball] Creating HID Manager for Keyball39 devices...\n");
    fflush(stderr);
    
    IOHIDManagerRef hidManager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
    if (!hidManager) {
        fprintf(stderr, "[Keyball] Failed to create HID Manager\n");
        fflush(stderr);
        return;
    }
    
    // Keyball39デバイスのみをフィルタリング
    CFMutableArrayRef matchingArray = CFArrayCreateMutable(kCFAllocatorDefault, 0, &kCFTypeArrayCallBacks);
    
    // Keyball39の名前でマッチング
    CFMutableDictionaryRef matching = CFDictionaryCreateMutable(
        kCFAllocatorDefault,
        0,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );
    
    CFStringRef productString = CFSTR("Keyball39");
    CFDictionarySetValue(matching, CFSTR(kIOHIDProductKey), productString);
    CFArrayAppendValue(matchingArray, matching);
    CFRelease(matching);
    
    IOHIDManagerSetDeviceMatchingMultiple(hidManager, matchingArray);
    CFRelease(matchingArray);
    
    // コールバックを登録
    IOHIDManagerRegisterInputValueCallback(hidManager, keyballHIDCallback, NULL);
    
    // RunLoopに登録
    IOHIDManagerScheduleWithRunLoop(hidManager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
    
    // デバイスを開く
    IOReturn ret = IOHIDManagerOpen(hidManager, kIOHIDOptionsTypeNone);
    if (ret != kIOReturnSuccess) {
        fprintf(stderr, "[Keyball] Failed to open HID Manager: 0x%x\n", ret);
        fflush(stderr);
        CFRelease(hidManager);
        return;
    }
    
    // 接続されたKeyball39デバイスを表示
    CFSetRef devices = IOHIDManagerCopyDevices(hidManager);
    if (devices) {
        CFIndex count = CFSetGetCount(devices);
        fprintf(stderr, "[Keyball] Found %ld Keyball39 devices:\n", count);
        
        CFTypeRef values[count];
        CFSetGetValues(devices, values);
        for (CFIndex i = 0; i < count; i++) {
            IOHIDDeviceRef device = (IOHIDDeviceRef)values[i];
            
            // Usage PageとUsageを取得
            CFNumberRef usagePageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsagePageKey));
            CFNumberRef usageRef = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsageKey));
            
            int usagePage = 0, usage = 0;
            if (usagePageRef) CFNumberGetValue(usagePageRef, kCFNumberIntType, &usagePage);
            if (usageRef) CFNumberGetValue(usageRef, kCFNumberIntType, &usage);
            
            const char* deviceType = "Unknown";
            if (usagePage == 0x0001 && usage == 0x0006) {
                deviceType = "Keyboard";
            } else if (usagePage == 0x0001 && usage == 0x0002) {
                deviceType = "Mouse";
            } else if (usagePage == 0xff60) {
                deviceType = "Custom/Vendor";
            }
            
            fprintf(stderr, "  Device %ld: Keyball39 [%s] (UsagePage: 0x%04x, Usage: 0x%04x)\n", 
                    i, deviceType, usagePage, usage);
            
            // エレメントの数を表示
            CFArrayRef elements = IOHIDDeviceCopyMatchingElements(device, NULL, kIOHIDOptionsTypeNone);
            if (elements) {
                CFIndex elementCount = CFArrayGetCount(elements);
                fprintf(stderr, "           Elements: %ld\n", elementCount);
                CFRelease(elements);
            }
        }
        CFRelease(devices);
    }
    fflush(stderr);
    
    fprintf(stderr, "\n[Keyball] Monitoring Keyball39 HID events.\n");
    fprintf(stderr, "[Keyball] Press F13/F14 or any other keys on your Keyball39...\n");
    fprintf(stderr, "[Keyball] Press Ctrl+C to stop.\n\n");
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
	log.Println("Keyball39 HID Event Monitor")
	log.Println("This will show ALL HID events from Keyball39 devices only")
	
	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigCh
		fmt.Println("\nStopping...")
		os.Exit(0)
	}()
	
	// Keyball HIDテストを実行
	C.testKeyballHID()
}