#ifndef CGEVENT_HELPER_H
#define CGEVENT_HELPER_H

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <math.h>

// スクロールイベントを作成して送信
void sendScrollEvent(double deltaX, double deltaY, int fingerCount, int phase);

// Event Tapのコールバック関数の型定義
CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon);

// Event Tapを作成
CFMachPortRef createEventTap();

// Event Tapを開始
void startEventTap(CFMachPortRef eventTap);

// Event Tapを停止
void stopEventTap(CFMachPortRef eventTap);

// マウスデルタを取得するための構造体
typedef struct {
    int32_t dx;
    int32_t dy;
} MouseDelta;

// グローバル変数として最新のマウスデルタを保持
extern MouseDelta lastMouseDelta;

// キー押下状態を保持
extern int32_t pressedKey;

#endif // CGEVENT_HELPER_H