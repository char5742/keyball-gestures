//go:build darwin
// +build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "cgevent_helper.h"
#include <pthread.h>

extern int32_t pressedKey;
extern pthread_mutex_t keyMutex;

// 押されているキーを取得する関数
int32_t getPressedKey() {
    pthread_mutex_lock(&keyMutex);
    int32_t key = pressedKey;
    pthread_mutex_unlock(&keyMutex);
    return key;
}
*/
import "C"
import (
	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// DarwinKeyboardInput はmacOS用のキーボード入力実装
type DarwinKeyboardInput struct {
	// Event Tapはマウス入力と共有される
}

// NewDarwinKeyboardInput は新しいDarwinKeyboardInputを作成する
func NewDarwinKeyboardInput() (common.KeyboardInput, error) {
	// Event TapはDarwinMouseInputで作成されるため、ここでは何もしない
	return &DarwinKeyboardInput{}, nil
}

// GetKey は押されているキーを取得する
func (k *DarwinKeyboardInput) GetKey() int32 {
	return int32(C.getPressedKey())
}

// Close はリソースをクリーンアップする
func (k *DarwinKeyboardInput) Close() error {
	// Event TapはDarwinMouseInputで管理されるため、ここでは何もしない
	return nil
}