package core

import (
	"time"

	"github.com/char5742/keyball-gestures/internal/config"
)

// GestureType はジェスチャーの種類を表す
type GestureType int

const (
	GestureNone GestureType = iota
	GestureTwoFinger
	GestureFourFinger
)

// GestureEvent はジェスチャーイベントを表す
type GestureEvent struct {
	Type        GestureType
	FingerCount int
	DeltaX      float64
	DeltaY      float64
	Timestamp   time.Time
	State       GestureState
}

// GestureState はジェスチャーの状態を表す
type GestureState int

const (
	GestureStateBegan GestureState = iota
	GestureStateChanged
	GestureStateEnded
)

// FingerPosition は指の位置を表す
type FingerPosition struct {
	X int32
	Y int32
}

// GestureDetector はジェスチャーを検出する
type GestureDetector struct {
	cfg                *config.Config
	fingerCount        int
	fingerPositions    []FingerPosition
	lastScrollTime     time.Time
	currentGestureType GestureType
	gestureState       GestureState
}

// NewGestureDetector は新しいGestureDetectorを作成する
func NewGestureDetector(cfg *config.Config) *GestureDetector {
	return &GestureDetector{
		cfg:             cfg,
		fingerPositions: make([]FingerPosition, 4), // 最大4本指
	}
}

// UpdateConfig は設定を更新する
func (gd *GestureDetector) UpdateConfig(cfg *config.Config) {
	gd.cfg = cfg
}

// ProcessInput は入力を処理してジェスチャーイベントを生成する
func (gd *GestureDetector) ProcessInput(pressedKey int32, dx, dy int32) *GestureEvent {
	now := time.Now()
	
	// デバッグ：設定値を確認（削除）
	// if pressedKey > 0 {
	// 	fmt.Printf("[GestureDetector] Key pressed: %d, TwoFingerKey: %d, FourFingerKey: %d\n", 
	// 		pressedKey, gd.cfg.Input.TwoFingerKey, gd.cfg.Input.FourFingerKey)
	// }

	// リセット閾値チェック
	if now.Sub(gd.lastScrollTime) > gd.cfg.Gesture.ResetThreshold && gd.fingerCount > 0 {
		// ジェスチャーをリセット
		event := &GestureEvent{
			Type:        gd.currentGestureType,
			FingerCount: gd.fingerCount,
			DeltaX:      0,
			DeltaY:      0,
			Timestamp:   now,
			State:       GestureStateEnded,
		}
		gd.fingerCount = 0
		gd.currentGestureType = GestureNone
		gd.gestureState = GestureStateEnded
		return event
	}

	if dx != 0 || dy != 0 {
		gd.lastScrollTime = now
	}

	// ジェスチャー判定
	switch {
	case pressedKey == int32(gd.cfg.Input.TwoFingerKey) && gd.fingerCount == 0:
		// 2本指ジェスチャー開始
		gd.fingerCount = 2
		gd.currentGestureType = GestureTwoFinger
		gd.gestureState = GestureStateBegan
		gd.initFingerPositions(gd.fingerCount, gd.cfg.TouchPad.MaxX/2, gd.cfg.TouchPad.MaxY/2)
		return &GestureEvent{
			Type:        GestureTwoFinger,
			FingerCount: 2,
			DeltaX:      0,
			DeltaY:      0,
			Timestamp:   now,
			State:       GestureStateBegan,
		}

	case pressedKey == int32(gd.cfg.Input.FourFingerKey) && gd.fingerCount == 0:
		// 4本指ジェスチャー開始
		gd.fingerCount = 4
		gd.currentGestureType = GestureFourFinger
		gd.gestureState = GestureStateBegan
		gd.initFingerPositions(gd.fingerCount, gd.cfg.TouchPad.MaxX/2, gd.cfg.TouchPad.MaxY/2)
		return &GestureEvent{
			Type:        GestureFourFinger,
			FingerCount: 4,
			DeltaX:      0,
			DeltaY:      0,
			Timestamp:   now,
			State:       GestureStateBegan,
		}

	case (pressedKey == int32(gd.cfg.Input.FourFingerKey) || pressedKey == int32(gd.cfg.Input.TwoFingerKey)) && gd.fingerCount > 0:
		// ジェスチャー継続中
		if (pressedKey == int32(gd.cfg.Input.TwoFingerKey) && gd.fingerCount == 2) ||
			(pressedKey == int32(gd.cfg.Input.FourFingerKey) && gd.fingerCount == 4) {
			// 指の位置を更新
			gd.updateFingerPositions(dx, dy)
			gd.gestureState = GestureStateChanged
			return &GestureEvent{
				Type:        gd.currentGestureType,
				FingerCount: gd.fingerCount,
				DeltaX:      float64(dx),
				DeltaY:      float64(dy),
				Timestamp:   now,
				State:       GestureStateChanged,
			}
		} else {
			// キーが変わった場合はジェスチャー終了
			event := &GestureEvent{
				Type:        gd.currentGestureType,
				FingerCount: gd.fingerCount,
				DeltaX:      0,
				DeltaY:      0,
				Timestamp:   now,
				State:       GestureStateEnded,
			}
			gd.fingerCount = 0
			gd.currentGestureType = GestureNone
			gd.gestureState = GestureStateEnded
			return event
		}

	default:
		// ジェスチャー終了またはなし
		if gd.fingerCount > 0 {
			event := &GestureEvent{
				Type:        gd.currentGestureType,
				FingerCount: gd.fingerCount,
				DeltaX:      0,
				DeltaY:      0,
				Timestamp:   now,
				State:       GestureStateEnded,
			}
			gd.fingerCount = 0
			gd.currentGestureType = GestureNone
			gd.gestureState = GestureStateEnded
			return event
		}
		return nil
	}
}

// GetFingerPositions は現在の指の位置を返す
func (gd *GestureDetector) GetFingerPositions() []FingerPosition {
	return gd.fingerPositions[:gd.fingerCount]
}

// initFingerPositions は指の初期位置を設定する
func (gd *GestureDetector) initFingerPositions(count int, centerX, centerY int32) {
	offset := int32(20)
	startY := centerY - offset*(int32(count)-1)/2

	for i := 0; i < count; i++ {
		gd.fingerPositions[i].X = centerX
		gd.fingerPositions[i].Y = startY + offset*int32(i)
	}
}

// updateFingerPositions は指の位置を更新する
func (gd *GestureDetector) updateFingerPositions(dx, dy int32) {
	for i := 0; i < gd.fingerCount; i++ {
		gd.fingerPositions[i].X += dx
		gd.fingerPositions[i].Y += dy

		// 範囲制限
		gd.fingerPositions[i].X = clamp(gd.fingerPositions[i].X, gd.cfg.TouchPad.MinX, gd.cfg.TouchPad.MaxX)
		gd.fingerPositions[i].Y = clamp(gd.fingerPositions[i].Y, gd.cfg.TouchPad.MinY, gd.cfg.TouchPad.MaxY)
	}
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