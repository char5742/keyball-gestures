// 統合テスト用のイベント監視コンポーネント
// trackpad_cgevent_simple.cを基に、イベントをJSONで記録する機能を追加
// コンパイル: clang event_monitor.c -framework ApplicationServices -framework CoreFoundation -o event_monitor
// 実行: ./event_monitor <出力ファイルパス> <監視時間（秒）>

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>
#include <string.h>
#include <time.h>

// グローバル変数
static FILE *outputFile = NULL;
static int eventCount = 0;
static int shouldStop = 0;
static time_t startTime;
static int monitorDuration = 10; // デフォルト10秒

// イベント情報を保持する構造体
typedef struct {
    double timestamp;
    int type;
    const char *typeName;
    double x;
    double y;
    double deltaX;
    double deltaY;
    int isGesture;
    int fingerCount;
} EventInfo;

// イベントタイプ名を取得
const char* getEventTypeName(CGEventType type) {
    switch (type) {
        case kCGEventLeftMouseDown: return "LeftMouseDown";
        case kCGEventRightMouseDown: return "RightMouseDown";
        case kCGEventScrollWheel: return "ScrollWheel";
        case 29: return "GestureBegin";
        case 30: return "GestureChange";
        case 31: return "GestureEnd";
        default: return "Unknown";
    }
}

// 終了処理の前方宣言
void finishRecording();

// イベント情報をJSONで記録
void recordEvent(EventInfo *info) {
    if (!outputFile) return;
    
    // 最初のイベントでない場合はカンマを追加
    if (eventCount > 0) {
        fprintf(outputFile, ",\n");
    }
    
    fprintf(outputFile, "    {\n");
    fprintf(outputFile, "      \"timestamp\": %.3f,\n", info->timestamp);
    fprintf(outputFile, "      \"type\": %d,\n", info->type);
    fprintf(outputFile, "      \"typeName\": \"%s\",\n", info->typeName);
    fprintf(outputFile, "      \"x\": %.1f,\n", info->x);
    fprintf(outputFile, "      \"y\": %.1f,\n", info->y);
    fprintf(outputFile, "      \"deltaX\": %.1f,\n", info->deltaX);
    fprintf(outputFile, "      \"deltaY\": %.1f,\n", info->deltaY);
    fprintf(outputFile, "      \"isGesture\": %s,\n", info->isGesture ? "true" : "false");
    fprintf(outputFile, "      \"fingerCount\": %d\n", info->fingerCount);
    fprintf(outputFile, "    }");
    
    fflush(outputFile);
    eventCount++;
}

// イベントコールバック
static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    // 監視時間をチェック
    time_t currentTime = time(NULL);
    if (currentTime - startTime >= monitorDuration) {
        finishRecording();
        CFRunLoopStop(CFRunLoopGetCurrent());
        return event;
    }
    
    // イベント情報を取得
    EventInfo info = {0};
    CGPoint location = CGEventGetLocation(event);
    info.timestamp = (double)CGEventGetTimestamp(event) / 1000000000.0;
    info.type = type;
    info.typeName = getEventTypeName(type);
    info.x = location.x;
    info.y = location.y;
    
    // イベントタイプに応じて詳細情報を取得
    switch (type) {
        case kCGEventScrollWheel:
            {
                int64_t isContinuous = CGEventGetIntegerValueField(event, kCGScrollWheelEventIsContinuous);
                if (isContinuous) {
                    info.deltaY = (double)CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
                    info.deltaX = (double)CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis2);
                    info.isGesture = 1;
                    info.fingerCount = 2; // 2本指スクロール
                    
                    printf("[%.2f] 2本指スクロール検出: X=%.1f, Y=%.1f\n", 
                           info.timestamp, info.deltaX, info.deltaY);
                    recordEvent(&info);
                }
            }
            break;
            
        case 29: // GestureBegin
        case 30: // GestureChange  
        case 31: // GestureEnd
            {
                // IOHIDEventSubtypeを確認して、ジェスチャーの種類を判定
                int64_t subtype = CGEventGetIntegerValueField(event, 110); // kCGEventFieldIOHIDEventSubtype
                
                // デバッグ: 各フィールドの値を確認
                printf("[DEBUG] Event type: %d, subtype: %lld\n", type, subtype);
                
                // より多くのフィールドを試す
                if (subtype == 2) {  // 4本指スワイプの場合のみ詳細デバッグ
                    printf("[DEBUG] === 4本指スワイプ詳細デバッグ ===\n");
                    for (int field = 0; field < 200; field++) {
                        double dval = CGEventGetDoubleValueField(event, field);
                        int64_t ival = CGEventGetIntegerValueField(event, field);
                        if (dval != 0.0 || ival != 0) {
                            printf("[DEBUG] Field %d: double=%f, int=%lld\n", field, dval, ival);
                        }
                    }
                    printf("[DEBUG] === デバッグ終了 ===\n");
                }
                
                // 既知のフィールドを試してみる
                double gestureX = CGEventGetDoubleValueField(event, 116); // kCGEventFieldGestureDeltaX
                double gestureY = CGEventGetDoubleValueField(event, 119); // kCGEventFieldGestureDeltaY
                double scrollX = CGEventGetDoubleValueField(event, 96);  // kCGEventFieldScrollWheelPointDeltaAxis2
                double scrollY = CGEventGetDoubleValueField(event, 97);  // kCGEventFieldScrollWheelPointDeltaAxis1
                int64_t scrollIntX = CGEventGetIntegerValueField(event, 11); // kCGEventFieldScrollWheelDeltaAxis2
                int64_t scrollIntY = CGEventGetIntegerValueField(event, 12); // kCGEventFieldScrollWheelDeltaAxis1
                
                printf("[DEBUG] GestureDelta: X=%f, Y=%f\n", gestureX, gestureY);
                printf("[DEBUG] ScrollPoint: X=%f, Y=%f\n", scrollX, scrollY);
                printf("[DEBUG] ScrollInt: X=%lld, Y=%lld\n", scrollIntX, scrollIntY);
                
                // ジェスチャーイベント用のデルタフィールドを使用
                info.deltaX = gestureX;
                info.deltaY = gestureY;
                info.isGesture = 1;
                
                // サブタイプで判定: 6 = スクロール (2本指), 2 = スワイプ (4本指)
                if (subtype == 6) {
                    // スクロールジェスチャーはスキップ（ScrollWheelイベントで既に記録）
                    break;
                } else if (subtype == 2) {
                    info.fingerCount = 4; // 4本指スワイプ
                    int64_t phase = CGEventGetIntegerValueField(event, 132); // kCGEventFieldGesturePhase
                    printf("[%.2f] 4本指スワイプ検出 (タイプ: %d, フェーズ: %lld): X=%.1f, Y=%.1f\n", 
                           info.timestamp, type, phase, info.deltaX, info.deltaY);
                    recordEvent(&info);
                } else {
                    // その他のジェスチャー
                    info.fingerCount = 0; // 不明
                    printf("[%.2f] その他のジェスチャー検出 (タイプ: %d, サブタイプ: %lld): X=%.1f, Y=%.1f\n", 
                           info.timestamp, type, subtype, info.deltaX, info.deltaY);
                    recordEvent(&info);
                }
            }
            break;
            
        case kCGEventLeftMouseDown:
        case kCGEventRightMouseDown:
            printf("[%.2f] %s: x=%.1f, y=%.1f\n", 
                   info.timestamp, info.typeName, info.x, info.y);
            recordEvent(&info);
            break;
    }
    
    return event;
}

// 終了処理
void finishRecording() {
    if (outputFile && !shouldStop) {
        // JSONの終了を書き込む
        fprintf(outputFile, "\n  ],\n");
        fprintf(outputFile, "  \"endTime\": %ld,\n", time(NULL));
        fprintf(outputFile, "  \"totalEvents\": %d\n", eventCount);
        fprintf(outputFile, "}\n");
        fclose(outputFile);
        outputFile = NULL;
    }
    shouldStop = 1;
}

// シグナルハンドラ
void signalHandler(int sig) {
    finishRecording();
    CFRunLoopStop(CFRunLoopGetCurrent());
}

int main(int argc, char *argv[]) {
    // 引数の処理
    if (argc < 2) {
        fprintf(stderr, "使用方法: %s <出力ファイルパス> [監視時間（秒）]\n", argv[0]);
        return 1;
    }
    
    const char *outputPath = argv[1];
    if (argc >= 3) {
        monitorDuration = atoi(argv[2]);
        if (monitorDuration <= 0) {
            fprintf(stderr, "エラー: 監視時間は正の整数である必要があります\n");
            return 1;
        }
    }
    
    // 出力ファイルを開く
    outputFile = fopen(outputPath, "w");
    if (!outputFile) {
        fprintf(stderr, "エラー: 出力ファイル '%s' を開けません\n", outputPath);
        return 1;
    }
    
    // JSONの開始を書き込む
    fprintf(outputFile, "{\n");
    fprintf(outputFile, "  \"startTime\": %ld,\n", time(NULL));
    fprintf(outputFile, "  \"duration\": %d,\n", monitorDuration);
    fprintf(outputFile, "  \"events\": [\n");
    
    printf("=== 統合テスト用イベント監視 ===\n");
    printf("出力ファイル: %s\n", outputPath);
    printf("監視時間: %d秒\n\n", monitorDuration);
    
    // 監視するイベントマスクを作成
    CGEventMask eventMask = 0;
    eventMask |= CGEventMaskBit(kCGEventLeftMouseDown);
    eventMask |= CGEventMaskBit(kCGEventRightMouseDown);
    eventMask |= CGEventMaskBit(kCGEventScrollWheel);
    eventMask |= CGEventMaskBit(29); // GestureBegin
    eventMask |= CGEventMaskBit(30); // GestureChange
    eventMask |= CGEventMaskBit(31); // GestureEnd
    
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
        fclose(outputFile);
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
    
    // シグナルハンドラを設定
    signal(SIGINT, signalHandler);
    signal(SIGTERM, signalHandler);
    
    printf("イベント監視を開始しました\n");
    printf("監視中のイベント:\n");
    printf("- 左クリック、右クリック\n");
    printf("- 2本指スクロール\n");
    printf("- 4本指ジェスチャー\n");
    printf("Ctrl+C または %d秒後に自動終了します\n\n", monitorDuration);
    
    // 開始時刻を記録
    startTime = time(NULL);
    
    // イベントループ実行
    CFRunLoopRun();
    
    // 終了処理
    finishRecording();
    
    // クリーンアップ
    CFRelease(eventTap);
    CFRelease(runLoopSource);
    
    printf("\n監視を終了しました\n");
    printf("総イベント数: %d\n", eventCount);
    printf("結果は '%s' に保存されました\n", outputPath);
    
    return 0;
}