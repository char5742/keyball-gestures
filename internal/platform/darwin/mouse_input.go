//go:build darwin
// +build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "cgevent_helper.h"
#include <pthread.h>
#include <CoreFoundation/CoreFoundation.h>

extern MouseDelta lastMouseDelta;
extern pthread_mutex_t mouseDeltaMutex;

// マウスデルタを取得する関数
void getMouseDelta(int32_t *dx, int32_t *dy) {
    pthread_mutex_lock(&mouseDeltaMutex);
    *dx = lastMouseDelta.dx;
    *dy = lastMouseDelta.dy;
    // 読み取り後はリセット
    lastMouseDelta.dx = 0;
    lastMouseDelta.dy = 0;
    pthread_mutex_unlock(&mouseDeltaMutex);
    
    if (*dx != 0 || *dy != 0) {
        fprintf(stderr, "[CGEvent] getMouseDelta returning: dx=%d, dy=%d\n", *dx, *dy);
    }
}

#include <stdio.h>
void debugLog(const char* msg) {
    fprintf(stderr, "[CGEvent Debug] %s\n", msg);
}
*/
import "C"
import (
	"fmt"

	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// DarwinMouseInput はmacOS用のマウス入力実装
type DarwinMouseInput struct {
	eventTap C.CFMachPortRef
}

// NewDarwinMouseInput は新しいDarwinMouseInputを作成する
func NewDarwinMouseInput() (common.MouseInput, error) {
	// Event Tapを作成
	eventTap := C.createEventTap()
	if eventTap == 0 {
		return nil, fmt.Errorf("failed to create event tap - アクセシビリティ権限を確認してください")
	}

	// Event Tapを別のゴルーチンで開始
	go func() {
		C.startEventTap(eventTap)
		C.CFRunLoopRun()
	}()

	return &DarwinMouseInput{
		eventTap: eventTap,
	}, nil
}

// GetMouseDelta はマウスの移動量を取得する
func (m *DarwinMouseInput) GetMouseDelta() (dx int32, dy int32) {
	var cdx, cdy C.int32_t
	C.getMouseDelta(&cdx, &cdy)
	return int32(cdx), int32(cdy)
}

// Grab はマウス入力を専有する（マウスカーソルの動きを無効化）
func (m *DarwinMouseInput) Grab() error {
	// マウスカーソルの動きを無効化
	C.lockMouseCursor()
	return nil
}

// Release はマウス入力の専有を解除する（マウスカーソルの動きを再有効化）
func (m *DarwinMouseInput) Release() error {
	// マウスカーソルの動きを再有効化
	C.unlockMouseCursor()
	return nil
}

// Close はリソースをクリーンアップする
func (m *DarwinMouseInput) Close() error {
	if m.eventTap != 0 {
		C.stopEventTap(m.eventTap)
		m.eventTap = 0
	}
	return nil
}