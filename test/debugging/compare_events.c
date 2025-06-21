// 実際のトラックパッドとシミュレーションのイベントを比較
// コンパイル: clang compare_events.c -framework ApplicationServices -framework CoreFoundation -o compare_events

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <unistd.h>

// CGEventの非公開フィールド定数
const int kCGEventFieldNSEventType = 55;
const int kCGEventFieldIOHIDEventSubtype = 110;
const int kCGEventFieldScrollWheelDeltaAxis1 = 11;
const int kCGEventFieldScrollWheelDeltaAxis2 = 12;
const int kCGEventFieldScrollWheelPointDeltaAxis1 = 96;
const int kCGEventFieldScrollWheelPointDeltaAxis2 = 97;
const int kCGEventFieldScrollWheelFixedPtDeltaAxis1 = 93;
const int kCGEventFieldScrollWheelFixedPtDeltaAxis2 = 94;
const int kCGEventFieldScrollWheelPhase = 99;
const int kCGEventFieldMomentumScrollPhase = 123;
const int kCGEventFieldContinuous = 88;

void printEventDetails(CGEventRef event, const char* label) {
    printf("\n=== %s ===\n", label);
    
    CGEventType type = CGEventGetType(event);
    printf("CGEventType: %d\n", type);
    
    // 各フィールドの値を表示
    printf("NSEventType: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldNSEventType));
    printf("IOHIDEventSubtype: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldIOHIDEventSubtype));
    printf("Continuous: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldContinuous));
    
    printf("ScrollWheelDeltaAxis1 (line): %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldScrollWheelDeltaAxis1));
    printf("ScrollWheelDeltaAxis2 (line): %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldScrollWheelDeltaAxis2));
    
    printf("ScrollWheelPointDeltaAxis1: %.2f\n", CGEventGetDoubleValueField(event, kCGEventFieldScrollWheelPointDeltaAxis1));
    printf("ScrollWheelPointDeltaAxis2: %.2f\n", CGEventGetDoubleValueField(event, kCGEventFieldScrollWheelPointDeltaAxis2));
    
    printf("ScrollWheelFixedPtDeltaAxis1: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldScrollWheelFixedPtDeltaAxis1));
    printf("ScrollWheelFixedPtDeltaAxis2: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldScrollWheelFixedPtDeltaAxis2));
    
    printf("ScrollWheelPhase: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldScrollWheelPhase));
    printf("MomentumScrollPhase: %lld\n", CGEventGetIntegerValueField(event, kCGEventFieldMomentumScrollPhase));
    
    CGPoint location = CGEventGetLocation(event);
    printf("Location: (%.1f, %.1f)\n", location.x, location.y);
}

CGEventRef capturedEvent = NULL;

// イベントキャプチャコールバック
CGEventRef captureCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    if (type == kCGEventScrollWheel && !capturedEvent) {
        int64_t isContinuous = CGEventGetIntegerValueField(event, kCGEventFieldContinuous);
        if (isContinuous) {
            // トラックパッドからのスクロールイベントをキャプチャ
            capturedEvent = CGEventCreateCopy(event);
            printf("\n*** トラックパッドイベントをキャプチャしました ***\n");
            printEventDetails(event, "キャプチャした実際のトラックパッドイベント");
        }
    }
    return event;
}

int main() {
    printf("=== イベント比較ツール ===\n");
    printf("1. まず実際のトラックパッドで2本指スクロールしてください\n");
    printf("2. その後、シミュレーションイベントと比較します\n\n");
    
    // イベントタップを作成してトラックパッドイベントをキャプチャ
    CGEventMask eventMask = CGEventMaskBit(kCGEventScrollWheel);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        eventMask,
        captureCallback,
        NULL
    );
    
    if (!eventTap) {
        fprintf(stderr, "Failed to create event tap\n");
        return 1;
    }
    
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault, eventTap, 0
    );
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    
    printf("トラックパッドでスクロールしてください...\n");
    
    // 5秒間イベントをキャプチャ
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 5.0, false);
    
    if (!capturedEvent) {
        printf("トラックパッドイベントをキャプチャできませんでした\n");
        return 1;
    }
    
    // イベントソースを作成
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    
    printf("\n\n=== シミュレーションイベントとの比較 ===\n");
    
    // 1. 標準的なCGEventCreateScrollWheelEvent
    CGEventRef standardEvent = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, 10, 0);
    if (standardEvent) {
        CGEventSetIntegerValueField(standardEvent, kCGEventFieldContinuous, 1);
        CGEventSetIntegerValueField(standardEvent, kCGEventFieldScrollWheelPhase, 2);
        printEventDetails(standardEvent, "標準的なCGEventCreateScrollWheelEvent");
        CFRelease(standardEvent);
    }
    
    // 2. キャプチャしたイベントのクローン（値を変更）
    CGEventRef clonedEvent = CGEventCreateCopy(capturedEvent);
    if (clonedEvent) {
        // デルタ値だけ変更
        CGEventSetIntegerValueField(clonedEvent, kCGEventFieldScrollWheelDeltaAxis1, 10);
        CGEventSetDoubleValueField(clonedEvent, kCGEventFieldScrollWheelPointDeltaAxis1, 10.0);
        CGEventSetIntegerValueField(clonedEvent, kCGEventFieldScrollWheelFixedPtDeltaAxis1, 10 * 65536);
        
        printEventDetails(clonedEvent, "キャプチャしたイベントのクローン（デルタ値変更）");
        
        printf("\n\n=== クローンイベントを送信してテスト ===\n");
        printf("3秒後にクローンイベントを送信します...\n");
        sleep(3);
        
        for (int i = 0; i < 5; i++) {
            CGEventPost(kCGHIDEventTap, clonedEvent);
            printf("送信 %d/5\n", i+1);
            usleep(100000);
        }
        
        CFRelease(clonedEvent);
    }
    
    // クリーンアップ
    CFRelease(capturedEvent);
    CFRelease(source);
    CFRunLoopRemoveSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CFRelease(runLoopSource);
    CFRelease(eventTap);
    
    printf("\n=== 完了 ===\n");
    printf("クローンイベントでスクロールが発生しましたか？\n");
    
    return 0;
}