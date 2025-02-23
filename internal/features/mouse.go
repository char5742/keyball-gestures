package features

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/char5742/keyball-gestures/internal/consts"
	"github.com/char5742/keyball-gestures/internal/types"
	"github.com/char5742/keyball-gestures/internal/utils"
)

// マウス入力を扱うインターフェース
type Mouse interface {
	HandleSignals()
	// マウスの移動量を取得する
	GetMouseDelta() (dx int32, dy int32)
	// マウス操作を専有する
	Grab() error
	// マウス操作の専有を解除する
	Release() error
	Close() error
}

type virtualMouse struct {
	file    *os.File
	grabbed bool
}

// 指定されたパスでマウスを作成する
func CreateMouse(path string) (Mouse, error) {
	f, err := os.OpenFile(path, syscall.O_RDWR|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, fmt.Errorf("failed to open device file: %w", err)
	}
	return &virtualMouse{file: f}, nil
}

func (m *virtualMouse) HandleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutting down...")
		os.Exit(0)
	}()
}

func (m *virtualMouse) GetMouseDelta() (dx int32, dy int32) {
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

func (m *virtualMouse) Grab() error {
	if m.grabbed {
		return nil
	}
	if err := utils.IOCtl(m.file, consts.EVIOCGRAB, 1); err != nil {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	m.grabbed = true
	return nil
}

func (m *virtualMouse) Release() error {
	if !m.grabbed {
		return nil
	}
	if err := utils.IOCtl(m.file, consts.EVIOCGRAB, 0); err != nil {
		return fmt.Errorf("failed to release device: %w", err)
	}
	m.grabbed = false
	return nil
}

func (m *virtualMouse) Close() error {
	_ = m.Release()
	return m.file.Close()
}
