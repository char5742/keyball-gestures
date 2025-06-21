//go:build darwin
// +build darwin

package darwin

import (
	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// Factory はmacOS用のPlatformFactory実装
type Factory struct {
	cfg *config.Config
}

// NewFactory は新しいmacOS用のFactoryを作成する
func NewFactory(cfg *config.Config) common.PlatformFactory {
	return &Factory{cfg: cfg}
}

// CreateScrollEmitter はmacOS用のスクロールエミッターを作成する
func (f *Factory) CreateScrollEmitter() (common.ScrollEmitter, error) {
	return NewDarwinScrollEmitter(f.cfg)
}

// CreateMouseInput はmacOS用のマウス入力デバイスを作成する
func (f *Factory) CreateMouseInput(devicePath string) (common.MouseInput, error) {
	// macOSではdevicePathは使用しない（Event Tapで全体を監視）
	return NewDarwinMouseInput()
}

// CreateKeyboardInput はmacOS用のキーボード入力デバイスを作成する
func (f *Factory) CreateKeyboardInput(devicePath string) (common.KeyboardInput, error) {
	// macOSではdevicePathは使用しない（Event Tapで全体を監視）
	return NewDarwinKeyboardInput()
}

// CreateDeviceMonitor はmacOS用のデバイスモニターを作成する
func (f *Factory) CreateDeviceMonitor() (common.DeviceMonitor, error) {
	// TODO: IOKitを使用したデバイスモニターの実装
	// 現時点では簡易実装
	return &darwinDeviceMonitor{}, nil
}

// GetAvailableDevices は利用可能なデバイスを取得する
func (f *Factory) GetAvailableDevices() ([]common.Device, error) {
	// macOSではEvent Tapで全体を監視するため、仮想デバイスを返す
	devices := []common.Device{
		{
			Name: "System Mouse",
			Path: "system",
			Type: common.DeviceTypeMouse,
		},
		{
			Name: "System Keyboard",
			Path: "system",
			Type: common.DeviceTypeKeyboard,
		},
	}
	return devices, nil
}

// darwinDeviceMonitor はmacOS用のデバイスモニター実装（簡易版）
type darwinDeviceMonitor struct {
	callback func(event common.DeviceEvent)
}

func (m *darwinDeviceMonitor) RegisterCallback(callback func(event common.DeviceEvent)) {
	m.callback = callback
}

func (m *darwinDeviceMonitor) Start() error {
	// TODO: IOKitを使用した実装
	return nil
}

func (m *darwinDeviceMonitor) Stop() error {
	return nil
}

func (m *darwinDeviceMonitor) Close() error {
	return nil
}