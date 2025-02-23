package types

import "github.com/char5742/keyball-gestures/internal/consts"

// InputID はデバイス識別子を表す構造体
type InputID struct {
	Bustype uint16 // バスタイプ
	Vendor  uint16 // ベンダーID
	Product uint16 // 製品ID
	Version uint16 // バージョン
}

// UserDev はuinputユーザーデバイスの設定を表す構造体
type UserDev struct {
	Name       [consts.MaxNameSize]byte // デバイス名
	ID         InputID                  // デバイス識別子
	EffectsMax uint32                   // 最大エフェクト数
	Absmax     [consts.AbsSize]int32    // 絶対座標の最大値
	Absmin     [consts.AbsSize]int32    // 絶対座標の最小値
	Absfuzz    [consts.AbsSize]int32    // 絶対座標のファジー値
	Absflat    [consts.AbsSize]int32    // 絶対座標のフラット値
}
