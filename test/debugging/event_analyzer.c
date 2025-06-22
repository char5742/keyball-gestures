#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>

// trackpad_dump.cと同じフィールド定義
const int kCGEventFieldNSEventType = 55;
const int kCGEventFieldIOHIDEventSubtype = 110;
const int kCGEventFieldGesturePhase = 132;
const int kCGEventFieldGestureMask = 134;
const int kCGEventFieldIOHIDEventFlags = 115;
const int kCGEventFieldGestureDeltaX = 116;
const int kCGEventFieldGestureDeltaY = 119;

// イベントシーケンスを記録
typedef struct {
    CGEventType type;
    int64_t subtype;
    int64_t phase;
    int64_t mask;
    int64_t flags;
    double deltaX;
    double deltaY;
} EventInfo;

EventInfo events[100];
int eventCount = 0;
int swipeEventCount = 0;

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    if (eventCount < 100) {
        EventInfo *info = &events[eventCount];
        info->type = type;
        info->subtype = CGEventGetIntegerValueField(event, kCGEventFieldIOHIDEventSubtype);
        info->phase = CGEventGetIntegerValueField(event, kCGEventFieldGesturePhase);
        info->mask = CGEventGetIntegerValueField(event, kCGEventFieldGestureMask);
        info->flags = CGEventGetIntegerValueField(event, kCGEventFieldIOHIDEventFlags);
        info->deltaX = CGEventGetDoubleValueField(event, kCGEventFieldGestureDeltaX);
        info->deltaY = CGEventGetDoubleValueField(event, kCGEventFieldGestureDeltaY);
        
        // スワイプイベントをカウント
        if (info->subtype == 23) {
            swipeEventCount++;
        }
        
        eventCount++;
    }
    
    return event;
}

void analyzeSequence() {
    printf("\n=== イベントシーケンス解析 ===\n");
    printf("総イベント数: %d\n", eventCount);
    printf("スワイプイベント数: %d\n\n", swipeEventCount);
    
    int type30Count = 0;
    int type29Count = 0;
    int type29AfterType30 = 0;
    int firstChangedFlags = -1;
    
    for (int i = 0; i < eventCount; i++) {
        EventInfo *e = &events[i];
        
        printf("[%d] type=%ld", i, (long)e->type);
        
        if (e->type == 30) {
            type30Count++;
            printf("(Change) sub=%lld(Swipe) phase=%lld mask=0x%llX flags=0x%llX Δ(%.1f,%.1f)",
                   e->subtype, e->phase, e->mask, e->flags, e->deltaX, e->deltaY);
            
            // 最初のChangedのflagsを記録
            if (e->phase == 2 && firstChangedFlags == -1) {
                firstChangedFlags = (int)e->flags;
            }
        } else if (e->type == 29) {
            type29Count++;
            printf("(Begin) sub=%lld", e->subtype);
            
            // type=30の直後にtype=29が来ているかチェック
            if (i > 0 && events[i-1].type == 30) {
                type29AfterType30++;
            }
        }
        
        printf("\n");
    }
    
    printf("\n=== 解析結果 ===\n");
    printf("type=30の数: %d\n", type30Count);
    printf("type=29の数: %d\n", type29Count);
    printf("type=30の直後のtype=29: %d\n", type29AfterType30);
    printf("最初のChangedのflags: 0x%X\n", firstChangedFlags);
    
    // ネイティブとの比較
    printf("\n=== ネイティブとの比較 ===\n");
    if (type30Count == swipeEventCount && type29Count == swipeEventCount) {
        printf("✅ type=30とtype=29の数が一致\n");
    } else {
        printf("❌ type=30(%d)とtype=29(%d)の数が不一致\n", type30Count, type29Count);
    }
    
    if (type29AfterType30 == type30Count) {
        printf("✅ すべてのtype=30の後にtype=29\n");
    } else {
        printf("❌ type=30の後にtype=29が来ない場合がある\n");
    }
    
    if (firstChangedFlags == 0x8) {
        printf("✅ 最初のChangedのflagsは0x8\n");
    } else {
        printf("❌ 最初のChangedのflagsが0x8ではない（実際: 0x%X）\n", firstChangedFlags);
    }
}

int main() {
    printf("=== イベントアナライザー ===\n");
    printf("4本指スワイプジェスチャーを検出します...\n");
    printf("10秒後に解析結果を表示します\n\n");
    
    // イベントタップを作成
    CGEventMask mask = CGEventMaskBit(29) | CGEventMaskBit(30) | CGEventMaskBit(31);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        mask,
        eventCallback,
        NULL
    );
    
    if (!eventTap) {
        fprintf(stderr, "アクセシビリティ権限が必要です\n");
        return 1;
    }
    
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(NULL, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    
    // 10秒間実行
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 10.0, false);
    
    // 解析
    analyzeSequence();
    
    return 0;
}