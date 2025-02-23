package features

import (
	"os"
	"strings"
)

type Device struct {
	Name string
	Path string
	Type DeviceType
}

// デバイスタイプを表す列挙型
type DeviceType int

const (
	DeviceTypeKeyboard DeviceType = iota
	DeviceTypeMouse
)

// 現在接続されているデバイスを取得する関数
func GetDevices() ([]Device, error) {
	entries, err := os.ReadDir("/dev/input/by-id")
	if err != nil {
		return nil, err
	}
	var devices []Device
	for _, entry := range entries {
		// eventが含まれない場合はスキップ
		if !strings.Contains(entry.Name(), "event") {
			continue
		}
		fullPath := "/dev/input/by-id/" + entry.Name()
		realPath, err := os.Readlink(fullPath)
		absPath := "/dev/input/by-id/" + realPath
		if err != nil {
			return nil, err
		}
		if strings.Contains(entry.Name(), "kbd") {
			devices = append(devices, Device{Name: entry.Name(), Path: absPath, Type: DeviceTypeKeyboard})
		}
		if strings.Contains(entry.Name(), "mouse") {
			devices = append(devices, Device{Name: entry.Name(), Path: absPath, Type: DeviceTypeMouse})
		}
	}

	return devices, nil
}
