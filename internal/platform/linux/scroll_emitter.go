//go:build linux
// +build linux

package linux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/consts"
	"github.com/char5742/keyball-gestures/internal/core"
	"github.com/char5742/keyball-gestures/internal/platform/common"
	"github.com/char5742/keyball-gestures/internal/types"
	"github.com/char5742/keyball-gestures/internal/utils"
)

// LinuxScrollEmitter はLinux用のスクロールエミッター実装
type LinuxScrollEmitter struct {
	deviceFile      *os.File
	cfg             *config.Config
	fingerPositions []core.FingerPosition
}

// NewLinuxScrollEmitter は新しいLinuxScrollEmitterを作成する
func NewLinuxScrollEmitter(cfg *config.Config) (common.ScrollEmitter, error) {
	fd, err := createTouchPad("/dev/uinput", []byte("VirtualTouchPad"),
		cfg.TouchPad.MinX, cfg.TouchPad.MaxX, cfg.TouchPad.MinY, cfg.TouchPad.MaxY)
	if err != nil {
		return nil, fmt.Errorf("仮想タッチパッドの作成に失敗しました: %v", err)
	}

	return &LinuxScrollEmitter{
		deviceFile:      fd,
		cfg:             cfg,
		fingerPositions: make([]core.FingerPosition, 4), // 最大4本指
	}, nil
}

// SendGesture はジェスチャーイベントを送信する
func (e *LinuxScrollEmitter) SendGesture(event *core.GestureEvent) error {
	if event == nil {
		return nil
	}

	switch event.State {
	case core.GestureStateBegan:
		// ジェスチャー開始時の処理
		return e.initFingers(event.FingerCount)

	case core.GestureStateChanged:
		// ジェスチャー継続中の処理
		return e.moveFingers(event.FingerCount, int32(event.DeltaX), int32(event.DeltaY))

	case core.GestureStateEnded:
		// ジェスチャー終了時の処理
		return e.liftAllFingers(event.FingerCount)
	}

	return nil
}

// Close はリソースをクリーンアップする
func (e *LinuxScrollEmitter) Close() error {
	_ = releaseDevice(e.deviceFile)
	return e.deviceFile.Close()
}

// initFingers は指の初期位置を設定する
func (e *LinuxScrollEmitter) initFingers(count int) error {
	offset := int32(20)
	centerX := e.cfg.TouchPad.MaxX / 2
	centerY := e.cfg.TouchPad.MaxY / 2
	startY := centerY - offset*(int32(count)-1)/2

	for i := 0; i < count; i++ {
		e.fingerPositions[i].X = centerX
		e.fingerPositions[i].Y = startY + offset*int32(i)

		if err := e.multiTouchDown(i, i, e.fingerPositions[i].X, e.fingerPositions[i].Y); err != nil {
			return err
		}
	}
	return nil
}

// moveFingers は指の位置を更新する
func (e *LinuxScrollEmitter) moveFingers(count int, dx, dy int32) error {
	for i := 0; i < count; i++ {
		e.fingerPositions[i].X += dx
		e.fingerPositions[i].Y += dy

		// 範囲制限
		e.fingerPositions[i].X = clamp(e.fingerPositions[i].X, e.cfg.TouchPad.MinX, e.cfg.TouchPad.MaxX)
		e.fingerPositions[i].Y = clamp(e.fingerPositions[i].Y, e.cfg.TouchPad.MinY, e.cfg.TouchPad.MaxY)

		if err := e.multiTouchMove(i, e.fingerPositions[i].X, e.fingerPositions[i].Y); err != nil {
			return err
		}
	}
	return nil
}

// liftAllFingers はすべての指を持ち上げる
func (e *LinuxScrollEmitter) liftAllFingers(count int) error {
	for i := 0; i < count; i++ {
		if err := e.multiTouchUp(i); err != nil {
			return err
		}
	}
	return nil
}

// multiTouchDown タッチイベントを開始する
func (e *LinuxScrollEmitter) multiTouchDown(slot int, trackingID int, x int32, y int32) error {
	events := []types.Event{
		{Type: consts.Abs, Code: consts.AbsMtSlot, Value: int32(slot)},
		{Type: consts.Abs, Code: consts.AbsMtTrackingId, Value: int32(trackingID)},
		{Type: consts.Abs, Code: consts.AbsMtPositionX, Value: x},
		{Type: consts.Abs, Code: consts.AbsMtPositionY, Value: y},
		{Type: consts.Abs, Code: consts.AbsMtTouchMajor, Value: 50},
		{Type: consts.Abs, Code: consts.AbsMtPressure, Value: 30},
		{Type: consts.Key, Code: consts.BtnTouch, Value: 1},
		{Type: consts.Syn, Code: consts.SynReport, Value: 0},
	}

	return writeEvents(e.deviceFile, events)
}

// multiTouchMove タッチ位置を更新する
func (e *LinuxScrollEmitter) multiTouchMove(slot int, x int32, y int32) error {
	events := []types.Event{
		{Type: consts.Abs, Code: consts.AbsMtSlot, Value: int32(slot)},
		{Type: consts.Abs, Code: consts.AbsMtPositionX, Value: x},
		{Type: consts.Abs, Code: consts.AbsMtPositionY, Value: y},
		{Type: consts.Abs, Code: consts.AbsMtTouchMajor, Value: 50},
		{Type: consts.Syn, Code: consts.SynReport, Value: 0},
	}

	return writeEvents(e.deviceFile, events)
}

// multiTouchUp タッチイベントを終了する
func (e *LinuxScrollEmitter) multiTouchUp(slot int) error {
	events := []types.Event{
		{Type: consts.Abs, Code: consts.AbsMtSlot, Value: int32(slot)},
		{Type: consts.Abs, Code: consts.AbsMtTrackingId, Value: -1},
		{Type: consts.Abs, Code: consts.AbsMtTouchMajor, Value: 0},
		{Type: consts.Key, Code: consts.BtnTouch, Value: 0},
		{Type: consts.Syn, Code: consts.SynReport, Value: 0},
	}

	return writeEvents(e.deviceFile, events)
}

// createTouchPad 新しいタッチパッドデバイスを作成する
func createTouchPad(path string, name []byte, minX int32, maxX int32, minY int32, maxY int32) (*os.File, error) {
	deviceFile, err := createDeviceFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not create absolute axis input device: %v", err)
	}

	// キー入力イベント(EV_KEY)を登録する
	err = registerDevice(deviceFile, uintptr(consts.Key))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("キー入力イベント(EV_KEY)の登録に失敗しました: %v", err)
	}

	// キー入力の種類（マウスボタン、タッチ検出など）を登録する
	for _, ev := range []int{
		consts.MouseBtnLeft,
		consts.MouseBtnRight,
		consts.BtnTouch,
		consts.BtnToolFinger,
	} {
		if err = utils.IOCtl(deviceFile, consts.SetKeyBit, uintptr(ev)); err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("キー入力種別の登録に失敗しました %v: %v", ev, err)
		}
	}

	// 絶対座標入力イベント(EV_ABS)を登録する
	err = registerDevice(deviceFile, uintptr(consts.Abs))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("絶対座標入力イベント(EV_ABS)の登録に失敗しました: %v", err)
	}

	// タッチパッドのプロパティを設定する
	if err := utils.IOCtl(deviceFile, consts.SetPropBit, uintptr(consts.PropPointer)); err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("ポインターデバイスプロパティの設定に失敗しました: %v", err)
	}
	if err := utils.IOCtl(deviceFile, consts.SetPropBit, uintptr(consts.PropButtonpad)); err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("ボタンパッドプロパティの設定に失敗しました: %v", err)
	}

	// X軸とY軸の座標を登録する
	for _, ev := range []int{consts.AbsX, consts.AbsY} {
		if err = utils.IOCtl(deviceFile, consts.SetAbsBit, uintptr(ev)); err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("座標軸の登録に失敗しました %v: %v", ev, err)
		}
	}

	// マルチタッチイベントを登録する
	for _, ev := range []int{
		consts.AbsMtSlot,
		consts.AbsMtPositionX,
		consts.AbsMtPositionY,
		consts.AbsMtTrackingId,
		consts.AbsMtTouchMajor,
		consts.AbsMtPressure,
	} {
		if err = utils.IOCtl(deviceFile, consts.SetAbsBit, uintptr(ev)); err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("マルチタッチイベントの登録に失敗しました %v: %v", ev, err)
		}
	}

	var absMin [consts.AbsSize]int32
	var absMax [consts.AbsSize]int32

	absMin[consts.AbsX] = minX
	absMax[consts.AbsX] = maxX
	absMin[consts.AbsY] = minY
	absMax[consts.AbsY] = maxY

	absMin[consts.AbsMtSlot] = 0
	absMax[consts.AbsMtSlot] = 9

	absMin[consts.AbsMtPositionX] = minX
	absMax[consts.AbsMtPositionX] = maxX
	absMin[consts.AbsMtPositionY] = minY
	absMax[consts.AbsMtPositionY] = maxY

	absMin[consts.AbsMtTouchMajor] = 0
	absMax[consts.AbsMtTouchMajor] = 255

	absMin[consts.AbsMtPressure] = 0
	absMax[consts.AbsMtPressure] = 255

	userDev := types.UserDev{
		Name: toUinputName(name),
		ID: types.InputID{
			Bustype: consts.BusUsb,
			Vendor:  0x4711,
			Product: 0x0817,
			Version: 1,
		},
		Absmin: absMin,
		Absmax: absMax,
	}

	fd, err := createUsbDevice(deviceFile, userDev)
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("USBデバイスの作成に失敗しました: %v", err)
	}

	return fd, nil
}

// デバイスファイルを作成する
func createDeviceFile(path string) (fd *os.File, err error) {
	deviceFile, err := os.OpenFile(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, errors.New("デバイスファイルを開くのに失敗しました")
	}
	return deviceFile, err
}

// デバイスを解放する
func releaseDevice(deviceFile *os.File) error {
	return utils.IOCtl(deviceFile, consts.DevDestroy, uintptr(0))
}

// デバイスを登録する
func registerDevice(deviceFile *os.File, evType uintptr) error {
	err := utils.IOCtl(deviceFile, consts.SetEvBit, evType)
	if err != nil {
		defer deviceFile.Close()
		err = releaseDevice(deviceFile)
		if err != nil {
			return fmt.Errorf("デバイスを解放するのに失敗しました: %v", err)
		}
		return fmt.Errorf("無効なファイルハンドルがutils.IOCtlから返されました: %v", err)
	}
	return nil
}

// USBデバイスを作成する
func createUsbDevice(deviceFile *os.File, dev types.UserDev) (fd *os.File, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, dev)
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("ユーザーデバイスバッファの書き込みに失敗しました: %v", err)
	}
	_, err = deviceFile.Write(buf.Bytes())
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("デバイス構造体をデバイスファイルに書き込むのに失敗しました: %v", err)
	}

	err = utils.IOCtl(deviceFile, consts.DevCreate, uintptr(0))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("デバイスの作成に失敗しました: %v", err)
	}

	return deviceFile, err
}

// イベントを書き込む
func writeEvents(deviceFile *os.File, events []types.Event) error {
	for _, ev := range events {
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, ev); err != nil {
			return fmt.Errorf("イベントをバッファに書き込むのに失敗しました: %v", err)
		}
		if _, err := deviceFile.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("イベントの書き込みに失敗しました: %v", err)
		}
	}
	return nil
}

// 名前をuinput用の固定長配列に変換する
func toUinputName(name []byte) (uinputName [consts.MaxNameSize]byte) {
	var fixedSizeName [consts.MaxNameSize]byte
	copy(fixedSizeName[:], name)
	return fixedSizeName
}

// clamp は値を最小値と最大値の間に制限する
func clamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}