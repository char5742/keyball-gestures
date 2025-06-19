package features

import "io"

// TouchPad 絶対座標入力デバイスを表現するインターフェース
type TouchPad interface {
	MultiTouchDown(slot int, trackingID int, x int32, y int32) error
	MultiTouchMove(slot int, x int32, y int32) error
	MultiTouchUp(slot int) error
	io.Closer
}

// TouchPadConfig タッチパッドの設定
type TouchPadConfig struct {
	Name                    string
	MinX                    int32
	MaxX                    int32
	MinY                    int32
	MaxY                    int32
	MotionSmoothingFactor   float64
	MotionWarmUpCount       int
	MouseDeltaFactor        float64
}