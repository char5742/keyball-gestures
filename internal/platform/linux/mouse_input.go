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
	"github.com/char5742/keyball-gestures/internal/utils"
)

// LinuxMouseInput はLinux用のマウス入力実装
type LinuxMouseInput struct {
	file    *os.File
	grabbed bool
}

// NewLinuxMouseInput は新しいLinuxMouseInputを作成する
func NewLinuxMouseInput(devicePath string) (common.MouseInput, error) {
	f, err := os.OpenFile(devicePath, syscall.O_RDWR|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, fmt.Errorf("failed to open device file: %w", err)
	}
	return &LinuxMouseInput{file: f}, nil
}

// GetMouseDelta はマウスの移動量を取得する
func (m *LinuxMouseInput) GetMouseDelta() (dx int32, dy int32) {
	var e types.Event
	size := binary.Size(e)
	buf := make([]byte, size)

	_, err := m.file.Read(buf)
	if err != nil {
		return 0, 0
	}

	e.Time.Sec = int64(binary.LittleEndian.Uint64(buf[0:8]))
	e.Time.Usec = int64(binary.LittleEndian.Uint64(buf[8:16]))
	e.Type = binary.LittleEndian.Uint16(buf[16:18])
	e.Code = binary.LittleEndian.Uint16(buf[18:20])
	e.Value = int32(binary.LittleEndian.Uint32(buf[20:24]))

	if e.Type == consts.Rel {
		switch e.Code {
		case consts.RelX:
			dx += e.Value
		case consts.RelY:
			dy += e.Value
		}
	}

	return dx, dy
}

// Grab はマウス入力を専有する
func (m *LinuxMouseInput) Grab() error {
	if m.grabbed {
		return nil
	}
	if err := utils.IOCtl(m.file, consts.EVIOCGRAB, 1); err != nil {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	m.grabbed = true
	return nil
}

// Release はマウス入力の専有を解除する
func (m *LinuxMouseInput) Release() error {
	if !m.grabbed {
		return nil
	}
	if err := utils.IOCtl(m.file, consts.EVIOCGRAB, 0); err != nil {
		return fmt.Errorf("failed to release device: %w", err)
	}
	m.grabbed = false
	return nil
}

// Close はリソースをクリーンアップする
func (m *LinuxMouseInput) Close() error {
	_ = m.Release()
	return m.file.Close()
}