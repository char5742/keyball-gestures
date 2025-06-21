//go:build linux
// +build linux

package linux

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"

	"github.com/char5742/keyball-gestures/internal/consts"
	"github.com/char5742/keyball-gestures/internal/platform/common"
	"github.com/char5742/keyball-gestures/internal/types"
)

// LinuxKeyboardInput はLinux用のキーボード入力実装
type LinuxKeyboardInput struct {
	file *os.File
}

// NewLinuxKeyboardInput は新しいLinuxKeyboardInputを作成する
func NewLinuxKeyboardInput(devicePath string) (common.KeyboardInput, error) {
	f, err := os.OpenFile(devicePath, syscall.O_RDONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, fmt.Errorf("failed to open device file: %w", err)
	}
	return &LinuxKeyboardInput{file: f}, nil
}

// GetKey は押されているキーを取得する
func (k *LinuxKeyboardInput) GetKey() int32 {
	var e types.Event
	size := binary.Size(e)
	buf := make([]byte, size)

	_, err := k.file.Read(buf)
	if err != nil {
		return 0
	}

	e.Time.Sec = int64(binary.LittleEndian.Uint64(buf[0:8]))
	e.Time.Usec = int64(binary.LittleEndian.Uint64(buf[8:16]))
	e.Type = binary.LittleEndian.Uint16(buf[16:18])
	e.Code = binary.LittleEndian.Uint16(buf[18:20])
	e.Value = int32(binary.LittleEndian.Uint32(buf[20:24]))

	if e.Type == consts.Key && e.Value > 0 {
		return int32(e.Code)
	}

	return 0
}

// Close はリソースをクリーンアップする
func (k *LinuxKeyboardInput) Close() error {
	return k.file.Close()
}