package features

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/char5742/keyball-gestures/internal/consts"
	"github.com/char5742/keyball-gestures/internal/types"
	"github.com/char5742/keyball-gestures/internal/utils"
)

// 絶対座標入力デバイスを表現するインターフェース
type TouchPad interface {
	MultiTouchDown(slot int, trackingID int, x int32, y int32) error
	MultiTouchMove(slot int, x int32, y int32) error
	MultiTouchUp(slot int) error
	io.Closer
}

type virtualTouchPad struct {
	name       []byte
	deviceFile *os.File
}

// 新しいタッチパッドデバイスを作成する
func CreateTouchPad(path string, name []byte, minX int32, maxX int32, minY int32, maxY int32) (TouchPad, error) {
	fd, err := createTouchPad(path, name, minX, maxX, minY, maxY)
	if err != nil {
		return nil, err
	}

	return &virtualTouchPad{name: name, deviceFile: fd}, nil
}

func (vt *virtualTouchPad) Close() error {
	_ = releaseDevice(vt.deviceFile)
	return vt.deviceFile.Close()
}

func createTouchPad(path string, name []byte, minX int32, maxX int32, minY int32, maxY int32) (*os.File, error) {
	deviceFile, err := createDeviceFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not create absolute axis input device: %v", err)
	}

	// キー入力イベント(EV_KEY)を登録する
	// これによりマウスボタンやタッチ入力などの検出が可能になる
	err = registerDevice(deviceFile, uintptr(consts.Key))
	if err != nil {
		_ = deviceFile.Close()
		return nil, fmt.Errorf("キー入力イベント(EV_KEY)の登録に失敗しました: %v", err)
	}

	// キー入力の種類（マウスボタン、タッチ検出など）を登録する
	for _, ev := range []int{
		consts.MouseBtnLeft,  // マウス左ボタン
		consts.MouseBtnRight, // マウス右ボタン
		consts.BtnTouch,      // 画面タッチの検出
		consts.BtnToolFinger, // 指の接触検出
	} {
		if err = utils.IOCtl(deviceFile, consts.SetKeyBit, uintptr(ev)); err != nil {
			_ = deviceFile.Close()
			return nil, fmt.Errorf("キー入力種別の登録に失敗しました %v: %v", ev, err)
		}
	}

	// 絶対座標入力イベント(EV_ABS)を登録する
	// これによりタッチパッドの位置情報を取得可能になる
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
		consts.AbsMtSlot,       // スロット（指の識別子）
		consts.AbsMtPositionX,  // X座標
		consts.AbsMtPositionY,  // Y座標
		consts.AbsMtTrackingId, // タッチの追跡ID
		consts.AbsMtTouchMajor, // タッチ領域の主軸
		consts.AbsMtPressure,   // タッチ圧力
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

// タッチイベントを開始する
func (vt *virtualTouchPad) MultiTouchDown(slot int, trackingID int, x int32, y int32) error {
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

	return writeEvents(vt.deviceFile, events)
}

// タッチ位置を更新する
func (vt *virtualTouchPad) MultiTouchMove(slot int, x int32, y int32) error {
	events := []types.Event{
		{Type: consts.Abs, Code: consts.AbsMtSlot, Value: int32(slot)},
		{Type: consts.Abs, Code: consts.AbsMtPositionX, Value: x},
		{Type: consts.Abs, Code: consts.AbsMtPositionY, Value: y},
		{Type: consts.Abs, Code: consts.AbsMtTouchMajor, Value: 50},
		{Type: consts.Syn, Code: consts.SynReport, Value: 0},
	}

	return writeEvents(vt.deviceFile, events)
}

// タッチイベントを終了する
func (vt *virtualTouchPad) MultiTouchUp(slot int) error {
	events := []types.Event{
		{Type: consts.Abs, Code: consts.AbsMtSlot, Value: int32(slot)},
		{Type: consts.Abs, Code: consts.AbsMtTrackingId, Value: -1},
		{Type: consts.Abs, Code: consts.AbsMtTouchMajor, Value: 0},
		{Type: consts.Key, Code: consts.BtnTouch, Value: 0},
		{Type: consts.Syn, Code: consts.SynReport, Value: 0},
	}

	return writeEvents(vt.deviceFile, events)
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
