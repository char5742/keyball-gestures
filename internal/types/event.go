package types

import "syscall"

// Event は入力イベントを表す構造体
type Event struct {
	Time  syscall.Timeval // イベント発生時刻
	Type  uint16          // イベントタイプ
	Code  uint16          // イベントコード
	Value int32           // イベント値
}
