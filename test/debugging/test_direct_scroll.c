// 最小限のスクロールイベント生成テスト
// コンパイル: clang test_direct_scroll.c -framework ApplicationServices -framework CoreFoundation -o test_direct_scroll
// 実行: ./test_direct_scroll

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <unistd.h>

int main() {
    printf("=== 直接スクロールイベントテスト ===\n");
    printf("3秒後にスクロールイベントを送信します...\n");
    sleep(3);
    
    // イベントソースを作成
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!source) {
        fprintf(stderr, "Failed to create event source\n");
        return 1;
    }
    
    // マウスの現在位置を取得
    CGEventRef event = CGEventCreate(source);
    CGPoint currentPos = CGEventGetLocation(event);
    CFRelease(event);
    
    printf("現在のマウス位置: x=%.1f, y=%.1f\n", currentPos.x, currentPos.y);
    
    // スクロールイベントを作成（標準的な方法）
    printf("\n=== 方法1: CGEventCreateScrollWheelEvent ===\n");
    for (int i = 0; i < 5; i++) {
        // 上にスクロール（deltaY > 0）
        CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(
            source,
            kCGScrollEventUnitLine,
            1,  // axisCount
            10  // delta1 (Y軸)
        );
        
        if (scrollEvent) {
            // イベントの場所を設定
            CGEventSetLocation(scrollEvent, currentPos);
            
            // イベントを送信
            CGEventPost(kCGHIDEventTap, scrollEvent);
            CFRelease(scrollEvent);
            
            printf("スクロールイベント送信 %d/5\n", i+1);
            usleep(100000);  // 100ms待機
        }
    }
    
    sleep(1);
    
    // 別の方法：CGEventCreateScrollWheelEvent2
    printf("\n=== 方法2: CGEventCreateScrollWheelEvent2 ===\n");
    for (int i = 0; i < 5; i++) {
        // 下にスクロール（deltaY < 0）
        CGEventRef scrollEvent = CGEventCreateScrollWheelEvent2(
            source,
            kCGScrollEventUnitPixel,
            2,      // wheelCount
            -50,    // wheel1 (Y軸)
            0,      // wheel2 (X軸)
            0       // wheel3 (Z軸)
        );
        
        if (scrollEvent) {
            CGEventSetLocation(scrollEvent, currentPos);
            CGEventPost(kCGHIDEventTap, scrollEvent);
            CFRelease(scrollEvent);
            
            printf("ピクセルスクロールイベント送信 %d/5\n", i+1);
            usleep(100000);
        }
    }
    
    CFRelease(source);
    
    printf("\n=== 完了 ===\n");
    printf("ブラウザやテキストエディタでスクロールが発生しましたか？\n");
    
    return 0;
}