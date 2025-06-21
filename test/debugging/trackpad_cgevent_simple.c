// CGEventTapを使用したシンプルなマウス/トラックパッドイベント監視
// コンパイル: clang trackpad_cgevent_simple.c -framework ApplicationServices -framework CoreFoundation -o trackpad_cgevent_simple
// 実行: ./trackpad_cgevent_simple（アクセシビリティ許可が必要）

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <stdlib.h>

// イベントコールバック
static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo)
{
    static int eventCount = 0;
    eventCount++;
    
    // イベント情報を取得
    CGPoint location = CGEventGetLocation(event);
    double timestamp = (double)CGEventGetTimestamp(event) / 1000000000.0; // ナノ秒から秒に変換
    
    // 主要なイベントのみ表示
    switch (type) {
        case kCGEventLeftMouseDown:
            printf("[%.2f] 左クリック: x=%.1f, y=%.1f\n", timestamp, location.x, location.y);
            break;
            
        case kCGEventRightMouseDown:
            printf("[%.2f] 右クリック: x=%.1f, y=%.1f\n", timestamp, location.x, location.y);
            break;
            
        case kCGEventScrollWheel:
            {
                // スクロールイベントがトラックパッドからのものかチェック
                int64_t isContinuous = CGEventGetIntegerValueField(event, kCGScrollWheelEventIsContinuous);
                
                // トラックパッドからのジェスチャーでない場合（マウスホイールなど）はスキップ
                if (!isContinuous) {
                    break;
                }
                
                // マルチタッチジェスチャーの場合、指の数に関する情報を取得しようとする
                // 注: 正確な指の数の検出にはプライベートAPIが必要な場合がある
                // ここでは、スクロール量の大きさとパターンから推測する
                
                int64_t deltaY = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
                int64_t deltaX = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis2);
                
                // 2本指スワイプは通常のスクロール動作として検出される
                printf("[%.2f] 2本指スワイプ: X=%lld, Y=%lld\n", timestamp, deltaX, deltaY);
                
            }
            break;
            
        default:
            // その他のイベント（29-31がジェスチャー関連）
            if (type >= 30 && type <= 31) {
                // ジェスチャーイベントの詳細情報を取得
                // タイプ29: ジェスチャー開始
                // タイプ30: ジェスチャー変更
                // タイプ31: ジェスチャー終了
                
                // 4本指ジェスチャーは通常これらのイベントとして検出される
                printf("[%.2f] 4本指ジェスチャーイベント (タイプ: %d)\n", timestamp, type);
                
                // ジェスチャーの方向や移動量を取得
                int64_t deltaY = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
                int64_t deltaX = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis2);
                if (deltaX != 0 || deltaY != 0) {
                    printf("  → 移動量: X=%lld, Y=%lld\n", deltaX, deltaY);
                }
            }
            break;
    }
    
    // イベントを通過させる
    return event;
}

int main(void)
{
    printf("=== マウス/トラックパッドイベント監視 ===\n\n");
    
    // 監視するイベントマスクを作成
    CGEventMask eventMask = 0;
    eventMask |= CGEventMaskBit(kCGEventLeftMouseDown);
    eventMask |= CGEventMaskBit(kCGEventRightMouseDown);
    eventMask |= CGEventMaskBit(kCGEventScrollWheel);
    
    // ジェスチャー関連のイベント（29-31）も監視
    eventMask |= CGEventMaskBit(29);
    eventMask |= CGEventMaskBit(30);
    eventMask |= CGEventMaskBit(31);
    
    // イベントタップを作成
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        eventMask,
        eventCallback,
        NULL
    );
    
    if (!eventTap) {
        fprintf(stderr, "エラー: イベントタップの作成に失敗しました\n");
        fprintf(stderr, "アクセシビリティ権限を確認してください:\n");
        fprintf(stderr, "システム設定 > プライバシーとセキュリティ > アクセシビリティ\n");
        return 1;
    }
    
    // RunLoopソースを作成して追加
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault, eventTap, 0
    );
    CFRunLoopAddSource(
        CFRunLoopGetCurrent(),
        runLoopSource,
        kCFRunLoopCommonModes
    );
    
    // イベントタップを有効化
    CGEventTapEnable(eventTap, true);
    
    printf("イベント監視を開始しました\n");
    printf("- 左クリック、右クリック\n");
    printf("- 2本指スワイプ\n");
    printf("- 4本指ジェスチャー\n");
    printf("Ctrl+C で終了します\n\n");
    
    // イベントループ実行
    CFRunLoopRun();
    
    return 0;
}