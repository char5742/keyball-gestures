#include "cgevent_helper.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

// グローバル変数
MouseDelta lastMouseDelta = {0, 0};
int32_t pressedKey = 0;
pthread_mutex_t mouseDeltaMutex = PTHREAD_MUTEX_INITIALIZER;
pthread_mutex_t keyMutex = PTHREAD_MUTEX_INITIALIZER;
int gestureTriggered = 0;  // 4本指ジェスチャーのトリガー状態

// マウスカーソル固定用
CGPoint lockedPosition = {0, 0};
int cursorLocked = 0;
pthread_mutex_t cursorMutex = PTHREAD_MUTEX_INITIALIZER;

// スクロールフェーズの定義
typedef enum {
    kScrollPhaseBegan = 1,
    kScrollPhaseChanged = 2,
    kScrollPhaseEnded = 4,
    kScrollPhaseCancelled = 8,
    kScrollPhaseMayBegin = 128
} ScrollPhase;

// スクロールイベントを作成して送信
void sendScrollEvent(double deltaX, double deltaY, int fingerCount, int phase) {
    static double accumulatedX = 0;
    static double accumulatedY = 0;
    
    // 4本指ジェスチャーの場合は累積する
    if (fingerCount == 4) {
        if (phase == kScrollPhaseBegan) {
            accumulatedX = 0;
            accumulatedY = 0;
        }
        accumulatedX += deltaX;
        accumulatedY += deltaY;
        
        // 累積値が闾値を超えたときのみログ
        if (fabs(accumulatedX) > 50.0 || fabs(accumulatedY) > 50.0) {
            fprintf(stderr, "[CGEvent] 4-finger gesture accumulated: dx=%.2f, dy=%.2f, phase=%d\n", 
                    accumulatedX, accumulatedY, phase);
        }
    } else {
        // 2本指スクロールのログを削除
        // fprintf(stderr, "[CGEvent] sendScrollEvent: dx=%.2f, dy=%.2f, fingers=%d, phase=%d\n", 
        //         deltaX, deltaY, fingerCount, phase);
    }
    
    // スクロールの方向を反転（Natural Scrolling）
    deltaY = -deltaY;
    deltaX = -deltaX;
    
    // 値が小さすぎる場合はスキップ
    if (fabs(deltaX) < 0.1 && fabs(deltaY) < 0.1) {
        return;
    }
    
    if (fingerCount == 2) {
        // 2本指スクロール - ピクセル単位でスムーズに
        CGEventRef event = CGEventCreateScrollWheelEvent(
            NULL,                    // イベントソース
            kCGScrollEventUnitPixel, // ピクセル単位でスムーズに
            2,                       // 軸の数
            (int32_t)deltaY,        // Y軸（垂直）
            (int32_t)deltaX         // X軸（水平）
        );
        
        if (event != NULL) {
            // イベントを投稿
            CGEventPost(kCGHIDEventTap, event);
            CFRelease(event);
            
            fprintf(stderr, "[CGEvent] 2-finger scroll event posted (pixel units)\n");
        }
    } else if (fingerCount == 4) {
        // 4本指ジェスチャー - システムのアクセシビリティAPIを使用
        // fprintf(stderr, "[CGEvent] 4-finger gesture detected, dx=%.2f, dy=%.2f, phase=%d\n", deltaX, deltaY, phase);
        
        // 累積値で判定し、一度だけアクションを実行
        double threshold = 50.0;
        
        // phase: 1=began, 2=changed, 4=ended
        if (phase == 1) {  // kScrollPhaseBegan
            gestureTriggered = 0;
            accumulatedX = 0;
            accumulatedY = 0;
            fprintf(stderr, "[CGEvent] 4-finger gesture began, resetting\n");
        } else if (phase == 4) {  // kScrollPhaseEnded
            gestureTriggered = 0;
            accumulatedX = 0;
            accumulatedY = 0;
            fprintf(stderr, "[CGEvent] 4-finger gesture ended, resetting\n");
        }
        
        // phase == 2 は kScrollPhaseChanged
        if (!gestureTriggered && phase == 2) {
            if (fabs(accumulatedY) > fabs(accumulatedX)) {
                // 垂直方向のジェスチャー
                if (accumulatedY < -threshold) {
                    // 上スワイプ - Mission Control
                    fprintf(stderr, "[CGEvent] 4-finger swipe UP - Mission Control\n");
                    
                    // AppleScriptでMission Controlを直接実行
                    system("open -a 'Mission Control'");
                    
                    gestureTriggered = 1;
                } else if (accumulatedY > threshold) {
                    // 下スワイプ - App Expose
                    fprintf(stderr, "[CGEvent] 4-finger swipe DOWN - Show Desktop\n");
                    
                    // デスクトップを表示（F11相当）
                    system("osascript -e 'tell application \"System Events\" to key code 103'");
                    
                    gestureTriggered = 1;
                }
            } else {
                // 水平方向のジェスチャー
                if (accumulatedX < -threshold) {
                    // 左スワイプ - 次のデスクトップ
                    fprintf(stderr, "[CGEvent] 4-finger swipe LEFT - Next Desktop\n");
                    
                    // 次のスペースに移動
                    system("osascript -e 'tell application \"System Events\" to key code 124 using {control down}'");
                    
                    gestureTriggered = 1;
                } else if (accumulatedX > threshold) {
                    // 右スワイプ - 前のデスクトップ
                    fprintf(stderr, "[CGEvent] 4-finger swipe RIGHT - Previous Desktop\n");
                    
                    // 前のスペースに移動
                    system("osascript -e 'tell application \"System Events\" to key code 123 using {control down}'");
                    
                    gestureTriggered = 1;
                }
            }
        }
    }
}

// Event Tapコールバック関数
CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    static int callback_count = 0;
    
    // Event Tapが無効になった場合
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        fprintf(stderr, "[CGEvent] Event tap disabled\n");
        return event;
    }
    
    // コールバックのデバッグログは削除
    
    // マウス移動イベントの処理
    if (type == kCGEventMouseMoved || 
        type == kCGEventLeftMouseDragged || 
        type == kCGEventRightMouseDragged ||
        type == kCGEventOtherMouseDragged) {
        
        // マウスボタンが押されている状態でMouseMovedが来た場合は特別に処理
        int shouldProcess = 0;
        if (type == kCGEventMouseMoved && pressedKey > 1000) {
            // Mouse 3/4が押されている間のMouseMovedイベントを処理
            shouldProcess = 1;
            // fprintf(stderr, "[CGEvent] Processing MouseMoved while button %d is pressed\n", pressedKey - 1000);
        } else if (type != kCGEventMouseMoved) {
            // 通常のドラッグイベント
            shouldProcess = 1;
        }
        
        if (shouldProcess) {
            // デルタ値を取得（異なるフィールドも試す）
            int64_t dx = CGEventGetIntegerValueField(event, kCGMouseEventDeltaX);
            int64_t dy = CGEventGetIntegerValueField(event, kCGMouseEventDeltaY);
            
            // より高精度なデルタ値も試す
            double dx_double = CGEventGetDoubleValueField(event, kCGMouseEventDeltaX);
            double dy_double = CGEventGetDoubleValueField(event, kCGMouseEventDeltaY);
            
            // 少しでも動きがあれば処理
            if (fabs(dx_double) > 0.01 || fabs(dy_double) > 0.01 || dx != 0 || dy != 0) {
                // グローバル変数に保存（スレッドセーフ）
                pthread_mutex_lock(&mouseDeltaMutex);
                // double値の方が精度が高い場合はそちらを使用
                if (fabs(dx_double) > fabs((double)dx) || fabs(dy_double) > fabs((double)dy)) {
                    lastMouseDelta.dx += (int32_t)(dx_double + 0.5);  // 四捨五入して累積
                    lastMouseDelta.dy += (int32_t)(dy_double + 0.5);  // 四捨五入して累積
                } else {
                    lastMouseDelta.dx += (int32_t)dx;  // 累積する
                    lastMouseDelta.dy += (int32_t)dy;  // 累積する
                }
                pthread_mutex_unlock(&mouseDeltaMutex);
                
                // デバッグ出力（Mouse 3/4が押されている時のみ）
                // マウス移動量のログを削除
                // if (pressedKey == 1002 || pressedKey == 1003) {
                //     const char* eventTypeName = "Unknown";
                //     if (type == kCGEventMouseMoved) eventTypeName = "MouseMoved";
                //     else if (type == kCGEventLeftMouseDragged) eventTypeName = "LeftDragged";
                //     else if (type == kCGEventRightMouseDragged) eventTypeName = "RightDragged";
                //     else if (type == kCGEventOtherMouseDragged) eventTypeName = "OtherDragged";
                //     
                //     fprintf(stderr, "[CGEvent] %s: dx=%lld(%.2f), dy=%lld(%.2f) (accumulated: %d, %d)\n", 
                //             eventTypeName, dx, dx_double, dy, dy_double, lastMouseDelta.dx, lastMouseDelta.dy);
                // }
            }
        }
    }
    
    // キーボードイベントの処理
    if (type == kCGEventKeyDown) {
        int64_t keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        pthread_mutex_lock(&keyMutex);
        pressedKey = (int32_t)keyCode;
        pthread_mutex_unlock(&keyMutex);
    } else if (type == kCGEventKeyUp) {
        pthread_mutex_lock(&keyMutex);
        pressedKey = 0;
        pthread_mutex_unlock(&keyMutex);
    }
    
    // マウスボタンイベントの処理（Mouse 3, Mouse 4など）
    if (type == kCGEventOtherMouseDown) {
        int64_t button = CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
        
        // Mouse 3 (button 2) または Mouse 4 (button 3) の場合
        if (button == 2 || button == 3) {
            pthread_mutex_lock(&keyMutex);
            // マウスボタンを特別なキーコードとして扱う（1000 + ボタン番号）
            pressedKey = 1000 + (int32_t)button;
            pthread_mutex_unlock(&keyMutex);
            
            // 現在のカーソル位置を記録
            pthread_mutex_lock(&cursorMutex);
            CGEventRef currentEvent = CGEventCreate(NULL);
            if (currentEvent != NULL) {
                lockedPosition = CGEventGetLocation(currentEvent);
                cursorLocked = 1;
                CFRelease(currentEvent);
                fprintf(stderr, "[CGEvent] Mouse button %lld pressed - locking cursor at (%.0f, %.0f)\n", 
                        button, lockedPosition.x, lockedPosition.y);
            }
            pthread_mutex_unlock(&cursorMutex);
            
            // イベントを消費して、他のアプリケーションに伝播しないようにする
            return NULL;
        }
    } else if (type == kCGEventOtherMouseUp) {
        int64_t button = CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
        
        // Mouse 3 (button 2) または Mouse 4 (button 3) の場合
        if (button == 2 || button == 3) {
            pthread_mutex_lock(&keyMutex);
            pressedKey = 0;
            pthread_mutex_unlock(&keyMutex);
            
            // カーソルロックを解除
            pthread_mutex_lock(&cursorMutex);
            cursorLocked = 0;
            pthread_mutex_unlock(&cursorMutex);
            
            // 4本指ジェスチャーのリセット
            if (button == 2) {  // Mouse 3 = 4本指ジェスチャー
                gestureTriggered = 0;
                fprintf(stderr, "[CGEvent] Mouse button %lld released - unlocking cursor\n", button);
            } else {
                fprintf(stderr, "[CGEvent] Mouse button %lld released - unlocking cursor\n", button);
            }
            
            // イベントを消費して、他のアプリケーションに伝播しないようにする
            return NULL;
        }
    }
    
    // 特定のマウスボタンの状態を確認
    if (type == kCGEventMouseMoved) {
        // 現在のイベントからボタンの状態を取得
        CGMouseButton button3State = CGEventSourceButtonState(kCGEventSourceStateHIDSystemState, 2); // Mouse 3
        CGMouseButton button4State = CGEventSourceButtonState(kCGEventSourceStateHIDSystemState, 3); // Mouse 4
        
        if (button3State || button4State) {
            fprintf(stderr, "[CGEvent] MouseMoved with button state - Button3: %d, Button4: %d\n", 
                    button3State, button4State);
        }
    }
    
    // カーソルがロックされている場合は元の位置に戻す
    pthread_mutex_lock(&cursorMutex);
    if (cursorLocked && (type == kCGEventMouseMoved || type == kCGEventLeftMouseDragged || 
                         type == kCGEventRightMouseDragged || type == kCGEventOtherMouseDragged)) {
        // カーソルを固定位置に戻す
        CGWarpMouseCursorPosition(lockedPosition);
        pthread_mutex_unlock(&cursorMutex);
        
        // ジェスチャー中はマウス移動イベントを消費
        if (pressedKey == 1002 || pressedKey == 1003) {
            return NULL;
        }
    } else {
        pthread_mutex_unlock(&cursorMutex);
    }
    
    // その他のイベントはそのまま通過させる
    return event;
}

// Event Tapを作成
CFMachPortRef createEventTap() {
    // 監視するイベントタイプのマスク
    CGEventMask eventMask = (
        CGEventMaskBit(kCGEventMouseMoved) |
        CGEventMaskBit(kCGEventLeftMouseDragged) |
        CGEventMaskBit(kCGEventRightMouseDragged) |
        CGEventMaskBit(kCGEventOtherMouseDragged) |
        CGEventMaskBit(kCGEventKeyDown) |
        CGEventMaskBit(kCGEventKeyUp) |
        CGEventMaskBit(kCGEventOtherMouseDown) |
        CGEventMaskBit(kCGEventOtherMouseUp)
    );
    
    // Event Tapを作成（HIDレベルで試す）
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGHIDEventTap,         // HIDレベルのタップに変更
        kCGHeadInsertEventTap,  // タップの位置
        kCGEventTapOptionDefault,    // イベントを処理（必要に応じて消費）
        eventMask,              // 監視するイベント
        eventTapCallback,       // コールバック関数
        NULL                    // ユーザーデータ
    );
    
    if (eventTap == NULL) {
        fprintf(stderr, "[CGEvent] Failed to create event tap. Check accessibility permissions.\n");
    } else {
        fprintf(stderr, "[CGEvent] Event tap created successfully.\n");
    }
    
    return eventTap;
}

// Event Tapを開始
void startEventTap(CFMachPortRef eventTap) {
    if (eventTap == NULL) {
        fprintf(stderr, "[CGEvent] Cannot start NULL event tap\n");
        return;
    }
    
    // Run Loop Sourceを作成
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    if (runLoopSource == NULL) {
        fprintf(stderr, "[CGEvent] Failed to create run loop source\n");
        return;
    }
    
    // Current Run Loopに追加
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    
    // Event Tapを有効化
    CGEventTapEnable(eventTap, true);
    
    fprintf(stderr, "[CGEvent] Event tap started and enabled\n");
    
    // Run Loop Sourceをリリース
    CFRelease(runLoopSource);
}

// Event Tapを停止
void stopEventTap(CFMachPortRef eventTap) {
    if (eventTap == NULL) {
        return;
    }
    
    // Event Tapを無効化
    CGEventTapEnable(eventTap, false);
    
    // リリース
    CFRelease(eventTap);
}

// マウスカーソルを固定
void lockMouseCursor() {
    pthread_mutex_lock(&cursorMutex);
    
    // 現在のカーソル位置を記録
    CGEventRef currentEvent = CGEventCreate(NULL);
    if (currentEvent != NULL) {
        lockedPosition = CGEventGetLocation(currentEvent);
        cursorLocked = 1;
        CFRelease(currentEvent);
        
        // マウスの動きとカーソルの連動を切り離す
        CGAssociateMouseAndMouseCursorPosition(false);
        
        fprintf(stderr, "[CGEvent] Cursor locked at (%.0f, %.0f)\n", 
                lockedPosition.x, lockedPosition.y);
    }
    
    pthread_mutex_unlock(&cursorMutex);
}

// マウスカーソルの固定を解除
void unlockMouseCursor() {
    pthread_mutex_lock(&cursorMutex);
    
    cursorLocked = 0;
    
    // マウスの動きとカーソルの連動を再開
    CGAssociateMouseAndMouseCursorPosition(true);
    
    fprintf(stderr, "[CGEvent] Cursor unlocked\n");
    
    pthread_mutex_unlock(&cursorMutex);
}