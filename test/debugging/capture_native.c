// ネイティブトラックパッドのイベントをキャプチャして詳細解析
#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>

// すべてのイベントフィールドを記録
static CGEventRef captureCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    static int count = 0;
    
    if (type == 30 && CGEventGetIntegerValueField(event, 110) == 23) {  // type=30, sub=23のみ
        printf("\n=== Event #%d ===\n", ++count);
        printf("Type: %d\n", type);
        
        // 重要なフィールドをすべて記録
        for (int i = 0; i < 200; i++) {
            int64_t val = CGEventGetIntegerValueField(event, i);
            if (val != 0) {
                printf("Field[%d]: %lld", i, val);
                
                // 既知のフィールドに名前を付ける
                switch(i) {
                    case 55: printf(" (NSEventType)"); break;
                    case 88: printf(" (Continuous)"); break;
                    case 99: printf(" (ScrollWheelPhase)"); break;
                    case 110: printf(" (IOHIDEventSubtype)"); break;
                    case 115: printf(" (IOHIDEventFlags)"); break;
                    case 123: printf(" (MomentumScrollPhase)"); break;
                    case 132: printf(" (GesturePhase)"); break;
                    case 134: printf(" (GestureMask)"); break;
                }
                printf("\n");
            }
        }
        
        // ダブル値のフィールドもチェック
        double dx = CGEventGetDoubleValueField(event, 116);
        double dy = CGEventGetDoubleValueField(event, 119);
        if (dx != 0 || dy != 0) {
            printf("GestureDelta: (%.2f, %.2f)\n", dx, dy);
        }
        
        // 位置情報
        CGPoint loc = CGEventGetLocation(event);
        printf("Location: (%.1f, %.1f)\n", loc.x, loc.y);
        
        // タイムスタンプ
        printf("Timestamp: %llu\n", CGEventGetTimestamp(event));
    }
    
    return event;
}

int main() {
    printf("=== ネイティブトラックパッドイベントキャプチャ ===\n");
    printf("実際のトラックパッドで4本指スワイプを実行してください\n");
    printf("10秒間記録します...\n\n");
    
    CGEventMask mask = CGEventMaskBit(29) | CGEventMaskBit(30) | CGEventMaskBit(31);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        mask,
        captureCallback,
        NULL
    );
    
    if (!eventTap) {
        fprintf(stderr, "アクセシビリティ権限が必要です\n");
        return 1;
    }
    
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(NULL, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 10.0, false);
    
    printf("\n\n記録完了\n");
    return 0;
}