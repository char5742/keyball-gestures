// トラックパッドイベントの問題を特定するテスト
// コンパイル: clang touchpad_fix_test.c -framework ApplicationServices -framework CoreFoundation -o touchpad_fix_test

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <unistd.h>

// CGEventの非公開フィールド定数
const int kCGEventFieldScrollWheelPhase = 99;
const int kCGEventFieldContinuous = 88;

int main() {
    printf("=== トラックパッドスクロールテスト ===\n");
    printf("3秒後に開始します...\n");
    sleep(3);
    
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!source) {
        fprintf(stderr, "Failed to create event source\n");
        return 1;
    }
    
    // マウスの現在位置を取得
    CGEventRef event = CGEventCreate(source);
    CGPoint currentPos = CGEventGetLocation(event);
    CFRelease(event);
    
    printf("マウス位置: x=%.1f, y=%.1f\n", currentPos.x, currentPos.y);
    
    // トラックパッドの2本指スクロールをシミュレート
    printf("\n=== 2本指スクロール（トラックパッド風）===\n");
    
    // フェーズ: MayBegin (128)
    printf("Phase: MayBegin\n");
    CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, 0, 0);
    if (scrollEvent) {
        CGEventSetLocation(scrollEvent, currentPos);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldContinuous, 1);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldScrollWheelPhase, 128); // MayBegin
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
    }
    usleep(50000);
    
    // フェーズ: Began (1)
    printf("Phase: Began\n");
    scrollEvent = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, 0, 0);
    if (scrollEvent) {
        CGEventSetLocation(scrollEvent, currentPos);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldContinuous, 1);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldScrollWheelPhase, 1); // Began
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
    }
    usleep(50000);
    
    // フェーズ: Changed (2) - 実際のスクロール
    printf("Phase: Changed (scrolling)\n");
    for (int i = 0; i < 10; i++) {
        scrollEvent = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, 10, 0);
        if (scrollEvent) {
            CGEventSetLocation(scrollEvent, currentPos);
            CGEventSetIntegerValueField(scrollEvent, kCGEventFieldContinuous, 1);
            CGEventSetIntegerValueField(scrollEvent, kCGEventFieldScrollWheelPhase, 2); // Changed
            CGEventPost(kCGHIDEventTap, scrollEvent);
            CFRelease(scrollEvent);
            printf("  Scroll %d/10\n", i+1);
        }
        usleep(50000);
    }
    
    // フェーズ: Ended (4)
    printf("Phase: Ended\n");
    scrollEvent = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, 0, 0);
    if (scrollEvent) {
        CGEventSetLocation(scrollEvent, currentPos);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldContinuous, 1);
        CGEventSetIntegerValueField(scrollEvent, kCGEventFieldScrollWheelPhase, 4); // Ended
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
    }
    
    CFRelease(source);
    
    printf("\n=== 完了 ===\n");
    printf("スクロールが発生しましたか？\n");
    
    return 0;
}