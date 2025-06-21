package common

import (
	"io"

	"github.com/char5742/keyball-gestures/internal/core"
)

// ScrollEmitter はプラットフォーム固有のスクロールイベントを送信するインターフェース
type ScrollEmitter interface {
	// SendGesture はジェスチャーイベントを送信する
	SendGesture(event *core.GestureEvent) error
	// Close はリソースをクリーンアップする
	io.Closer
}

// MouseInput はマウス入力を取得するインターフェース
type MouseInput interface {
	// GetMouseDelta はマウスの移動量を取得する
	GetMouseDelta() (dx int32, dy int32)
	// Grab はマウス入力を専有する
	Grab() error
	// Release はマウス入力の専有を解除する
	Release() error
	// Close はリソースをクリーンアップする
	io.Closer
}

// KeyboardInput はキーボード入力を取得するインターフェース
type KeyboardInput interface {
	// GetKey は押されているキーを取得する
	GetKey() int32
	// Close はリソースをクリーンアップする
	io.Closer
}

// DeviceEvent はデバイスイベントを表す
type DeviceEvent struct {
	Type       DeviceEventType
	DeviceType DeviceType
	Path       string
	Name       string
}

// DeviceEventType はデバイスイベントの種類
type DeviceEventType int

const (
	DeviceAdded DeviceEventType = iota
	DeviceRemoved
)

// DeviceType はデバイスの種類
type DeviceType int

const (
	DeviceTypeMouse DeviceType = iota
	DeviceTypeKeyboard
	DeviceTypeOther
)

// DeviceMonitor はデバイスの接続・切断を監視するインターフェース
type DeviceMonitor interface {
	// RegisterCallback はデバイスイベントのコールバックを登録する
	RegisterCallback(callback func(event DeviceEvent))
	// Start は監視を開始する
	Start() error
	// Stop は監視を停止する
	Stop() error
	// Close はリソースをクリーンアップする
	io.Closer
}

// PlatformFactory はプラットフォーム固有の実装を作成するファクトリインターフェース
type PlatformFactory interface {
	// CreateScrollEmitter はスクロールエミッターを作成する
	CreateScrollEmitter() (ScrollEmitter, error)
	// CreateMouseInput はマウス入力デバイスを作成する
	CreateMouseInput(devicePath string) (MouseInput, error)
	// CreateKeyboardInput はキーボード入力デバイスを作成する
	CreateKeyboardInput(devicePath string) (KeyboardInput, error)
	// CreateDeviceMonitor はデバイスモニターを作成する
	CreateDeviceMonitor() (DeviceMonitor, error)
	// GetAvailableDevices は利用可能なデバイスを取得する
	GetAvailableDevices() ([]Device, error)
}

// Device はデバイス情報を表す
type Device struct {
	Name string
	Path string
	Type DeviceType
}