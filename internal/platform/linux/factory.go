//go:build linux
// +build linux

package linux

import (
	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/features"
	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// Factory はLinux用のPlatformFactory実装
type Factory struct {
	cfg *config.Config
}

// NewFactory は新しいLinux用のFactoryを作成する
func NewFactory(cfg *config.Config) common.PlatformFactory {
	return &Factory{cfg: cfg}
}

// CreateScrollEmitter はLinux用のスクロールエミッターを作成する
func (f *Factory) CreateScrollEmitter() (common.ScrollEmitter, error) {
	return NewLinuxScrollEmitter(f.cfg)
}

// CreateMouseInput はLinux用のマウス入力デバイスを作成する
func (f *Factory) CreateMouseInput(devicePath string) (common.MouseInput, error) {
	return NewLinuxMouseInput(devicePath)
}

// CreateKeyboardInput はLinux用のキーボード入力デバイスを作成する
func (f *Factory) CreateKeyboardInput(devicePath string) (common.KeyboardInput, error) {
	return NewLinuxKeyboardInput(devicePath)
}

// CreateDeviceMonitor はLinux用のデバイスモニターを作成する
func (f *Factory) CreateDeviceMonitor() (common.DeviceMonitor, error) {
	// 既存のDeviceMonitorをラップする
	monitor, err := features.GetDeviceMonitor()
	if err != nil {
		return nil, err
	}
	return &linuxDeviceMonitorAdapter{monitor: monitor}, nil
}

// GetAvailableDevices は利用可能なデバイスを取得する
func (f *Factory) GetAvailableDevices() ([]common.Device, error) {
	// 既存のGetDevices関数を使用
	devices, err := features.GetDevices()
	if err != nil {
		return nil, err
	}

	// featuresのDeviceからcommon.Deviceに変換
	result := make([]common.Device, len(devices))
	for i, d := range devices {
		deviceType := common.DeviceTypeOther
		switch d.Type {
		case features.DeviceTypeMouse:
			deviceType = common.DeviceTypeMouse
		case features.DeviceTypeKeyboard:
			deviceType = common.DeviceTypeKeyboard
		}

		result[i] = common.Device{
			Name: d.Name,
			Path: d.Path,
			Type: deviceType,
		}
	}

	return result, nil
}

// linuxDeviceMonitorAdapter は既存のDeviceMonitorをcommon.DeviceMonitorに適合させる
type linuxDeviceMonitorAdapter struct {
	monitor *features.DeviceMonitor
}

func (a *linuxDeviceMonitorAdapter) RegisterCallback(callback func(event common.DeviceEvent)) {
	a.monitor.RegisterCallback(func(event features.DeviceEvent) {
		// featuresのイベントタイプをcommonのイベントタイプに変換
		eventType := common.DeviceAdded
		if event.Type == features.DeviceRemoved {
			eventType = common.DeviceRemoved
		}

		deviceType := common.DeviceTypeOther
		switch event.Device.Type {
		case features.DeviceTypeMouse:
			deviceType = common.DeviceTypeMouse
		case features.DeviceTypeKeyboard:
			deviceType = common.DeviceTypeKeyboard
		}

		callback(common.DeviceEvent{
			Type:       eventType,
			DeviceType: deviceType,
			Path:       event.Path,
			Name:       event.Device.Name,
		})
	})
}

func (a *linuxDeviceMonitorAdapter) Start() error {
	// DeviceMonitorは自動的に開始される
	return nil
}

func (a *linuxDeviceMonitorAdapter) Stop() error {
	// 現在の実装では停止機能はない
	return nil
}

func (a *linuxDeviceMonitorAdapter) Close() error {
	// 現在の実装ではクローズ機能はない
	return nil
}