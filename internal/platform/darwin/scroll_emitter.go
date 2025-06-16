//go:build darwin
// +build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "cgevent_helper.h"
*/
import "C"
import (
	"fmt"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/core"
	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// DarwinScrollEmitter はmacOS用のスクロールエミッター実装
type DarwinScrollEmitter struct {
	cfg *config.Config
}

// NewDarwinScrollEmitter は新しいDarwinScrollEmitterを作成する
func NewDarwinScrollEmitter(cfg *config.Config) (common.ScrollEmitter, error) {
	return &DarwinScrollEmitter{
		cfg: cfg,
	}, nil
}

// SendGesture はジェスチャーイベントを送信する
func (e *DarwinScrollEmitter) SendGesture(event *core.GestureEvent) error {
	if event == nil {
		return nil
	}

	// CGEventのスクロールフェーズ定義
	const (
		scrollPhaseBegan     = 1
		scrollPhaseChanged   = 2
		scrollPhaseEnded     = 4
		scrollPhaseCancelled = 8
		scrollPhaseMayBegin  = 128
	)

	var phase C.int
	switch event.State {
	case core.GestureStateBegan:
		phase = scrollPhaseBegan
	case core.GestureStateChanged:
		phase = scrollPhaseChanged
	case core.GestureStateEnded:
		phase = scrollPhaseEnded
	default:
		return fmt.Errorf("unknown gesture state: %v", event.State)
	}

	// デバッグログを削除
	// fmt.Printf("[DarwinScrollEmitter] Sending gesture: fingers=%d, dx=%.2f, dy=%.2f, state=%v\n",
	// 	event.FingerCount, event.DeltaX, event.DeltaY, event.State)

	// スクロールイベントを送信
	// macOSでは4本指ジェスチャーの場合、デフォルトでMission ControlやApp Exposeが動作する
	// 2本指の場合は通常のスクロール
	C.sendScrollEvent(
		C.double(event.DeltaX),
		C.double(event.DeltaY),
		C.int(event.FingerCount),
		phase,
	)

	return nil
}

// Close はリソースをクリーンアップする
func (e *DarwinScrollEmitter) Close() error {
	// 特にクリーンアップするリソースはない
	return nil
}