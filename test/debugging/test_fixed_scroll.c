// 修正版スクロールイベント生成テスト
// コンパイル: clang test_fixed_scroll.c -framework ApplicationServices -framework CoreFoundation -o test_fixed_scroll

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <unistd.h>
#include <mach/mach_time.h>

// CGEventの非公開フィールド定数
const int kCGEventFieldNSEventType = 55;
const int kCGEventFieldScrollWheelPhase = 99;
const int kCGEventFieldMomentumScrollPhase = 123;
const int kCGEventFieldContinuous = 88;
const int kCGEventFieldGesturePhase = 132;
const int kCGEventFieldGestureDeltaX = 116;
const int kCGEventFieldGestureDeltaY = 119;
const int kCGEventFieldIOHIDEventSubtype = 110;

const int NSEventTypeScrollWheel = 22;
const int NSEventTypeGesture = 29;
const int kIOHIDEventTypeScroll = 6;

// 修正版のスクロールイベント送信
void postScrollEventFixed(CGEventSourceRef source, double deltaX, double deltaY, int phase) {
    // 1. 標準的な方法でスクロールイベントを作成
    CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(
        source,
        kCGScrollEventUnitPixel,
        2,  // 2軸
        (int32_t)deltaY,
        (int32_t)deltaX
    );
    
    if (!scrollEvent) {
        printf("Failed to create scroll event\n");
        return;
    }
    
    // 2. 必要なフィールドを設定
    CGEventSetIntegerValueField(scrollEvent, kCGEventFieldContinuous, 1);
    CGEventSetIntegerValueField(scrollEvent, kCGEventFieldScrollWheelPhase, phase);
    CGEventSetIntegerValueField(scrollEvent, kCGEventFieldMomentumScrollPhase, 0);
    
    // 3. 現在のマウス位置を設定
    CGEventRef posEvent = CGEventCreate(NULL);
    if (posEvent) {
        CGPoint currentPos = CGEventGetLocation(posEvent);
        CGEventSetLocation(scrollEvent, currentPos);
        CFRelease(posEvent);
    }
    
    // 4. タイムスタンプを設定
    CGEventSetTimestamp(scrollEvent, mach_absolute_time());
    
    // 5. イベントを送信
    CGEventPost(kCGHIDEventTap, scrollEvent);
    
    // 6. ジェスチャーイベントも送信（オプション）
    CGEventRef gestureEvent = CGEventCreate(source);
    if (gestureEvent) {
        CGEventSetType(gestureEvent, (CGEventType)NSEventTypeGesture);
        CGEventSetIntegerValueField(gestureEvent, kCGEventFieldNSEventType, NSEventTypeGesture);
        CGEventSetIntegerValueField(gestureEvent, kCGEventFieldIOHIDEventSubtype, kIOHIDEventTypeScroll);
        CGEventSetDoubleValueField(gestureEvent, kCGEventFieldGestureDeltaX, -deltaX);
        CGEventSetDoubleValueField(gestureEvent, kCGEventFieldGestureDeltaY, -deltaY);
        CGEventSetIntegerValueField(gestureEvent, kCGEventFieldGesturePhase, phase);
        CGEventSetTimestamp(gestureEvent, mach_absolute_time());
        
        CGEventPost(kCGHIDEventTap, gestureEvent);
        CFRelease(gestureEvent);
    }
    
    CFRelease(scrollEvent);
}

int main() {
    printf("=== 修正版スクロールテスト ===\n");
    printf("3秒後に開始します...\n");
    sleep(3);
    
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!source) {
        fprintf(stderr, "Failed to create event source\n");
        return 1;
    }
    
    printf("2本指スクロールをシミュレート...\n");
    
    // Phase: MayBegin (128)
    printf("Phase: MayBegin\n");
    postScrollEventFixed(source, 0, 0, 128);
    usleep(50000);
    
    // Phase: Began (1)
    printf("Phase: Began\n");
    postScrollEventFixed(source, 0, 0, 1);
    usleep(50000);
    
    // Phase: Changed (2) - 実際のスクロール
    printf("Phase: Changed (scrolling up)\n");
    for (int i = 0; i < 20; i++) {
        postScrollEventFixed(source, 0, 10, 2);
        printf("  Scroll %d/20\n", i+1);
        usleep(30000);  // 30ms
    }
    
    // Phase: Ended (4)
    printf("Phase: Ended\n");
    postScrollEventFixed(source, 0, 0, 4);
    
    sleep(1);
    
    printf("\n下方向にもスクロール...\n");
    
    // 逆方向
    printf("Phase: Began\n");
    postScrollEventFixed(source, 0, 0, 1);
    usleep(50000);
    
    printf("Phase: Changed (scrolling down)\n");
    for (int i = 0; i < 20; i++) {
        postScrollEventFixed(source, 0, -10, 2);
        printf("  Scroll %d/20\n", i+1);
        usleep(30000);
    }
    
    printf("Phase: Ended\n");
    postScrollEventFixed(source, 0, 0, 4);
    
    CFRelease(source);
    
    printf("\n=== 完了 ===\n");
    printf("スクロールが発生しましたか？\n");
    
    return 0;
}