package event

import "syscall"

// イベントタイプの定数（input-event-codes.hより）
const (
	Syn      = 0x00 // 同期イベント
	Key      = 0x01 // キーイベント
	Rel      = 0x02 // 相対座標イベント
	Abs      = 0x03 // 絶対座標イベント
	RelX     = 0x0  // X軸の相対移動
	RelY     = 0x1  // Y軸の相対移動
	RelWheel = 0x8  // ホイールの相対移動

	AbsX            = 0x00 // X軸の絶対座標
	AbsY            = 0x01 // Y軸の絶対座標
	AbsMtSlot       = 0x2f // マルチタッチスロット
	AbsMtTouchMajor = 0x30 // タッチ領域の長径
	AbsMtPositionX  = 0x35 // マルチタッチのX座標
	AbsMtPositionY  = 0x36 // マルチタッチのY座標
	AbsMtTrackingId = 0x39 // タッチ追跡用ID
	AbsMtPressure   = 0x3a // タッチ圧力

	SynReport     = 0     // イベント報告の同期
	MouseBtnLeft  = 0x110 // マウス左ボタン
	MouseBtnRight = 0x111 // マウス右ボタン
	BtnTouch      = 0x14a // タッチイベント
	BtnToolFinger = 0x145 // 指によるタッチ
)

// Event は入力イベントを表す構造体
type Event struct {
	Time  syscall.Timeval // イベント発生時刻
	Type  uint16          // イベントタイプ
	Code  uint16          // イベントコード
	Value int32           // イベント値
}
