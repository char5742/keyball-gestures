package api

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/features"
)

// GestureService はジェスチャー認識サービスを管理する構造体
type GestureService struct {
	cfg          *config.Config
	stopChan     chan struct{}
	running      bool
	statusMutex  sync.RWMutex
	touchPad     features.TouchPad
	keyboard     features.Keyboard
	mouse        features.Mouse
	updateConfig chan *config.Config
}

// NewGestureService は新しいジェスチャー認識サービスを作成する
func NewGestureService(cfg *config.Config) *GestureService {
	return &GestureService{
		cfg:          cfg,
		stopChan:     make(chan struct{}),
		running:      false,
		updateConfig: make(chan *config.Config, 1),
	}
}

// Start はジェスチャー認識サービスを開始する
func (s *GestureService) Start() error {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if s.running {
		return fmt.Errorf("サービスは既に実行中です")
	}

	// 仮想タッチパッドデバイスの作成
	padDevice, err := features.CreateTouchPad("/dev/uinput", []byte("VirtualTouchPad"),
		s.cfg.TouchPad.MinX, s.cfg.TouchPad.MaxX, s.cfg.TouchPad.MinY, s.cfg.TouchPad.MaxY)
	if err != nil {
		return fmt.Errorf("仮想タッチパッドの作成に失敗しました: %v", err)
	}
	s.touchPad = padDevice

	// デバイス一覧の取得
	devices, err := features.GetDevices()
	if err != nil {
		s.touchPad.Close()
		return fmt.Errorf("デバイス一覧の取得に失敗しました: %v", err)
	}

	// 設定ファイルで指定された優先デバイスまたは最初のマウスとキーボードを使用
	var mouseDevice *features.Device
	var keyboardDevice *features.Device

	// 優先デバイスの名前
	preferredKeyboard := s.cfg.DevicePrefs.PreferredKeyboardDevice
	preferredMouse := s.cfg.DevicePrefs.PreferredMouseDevice

	// 最初に見つかったデバイスを初期値として保存
	var firstMouseDevice *features.Device
	var firstKeyboardDevice *features.Device

	for _, device := range devices {
		if device.Type == features.DeviceTypeMouse {
			// 最初のマウスを記録
			if firstMouseDevice == nil {
				firstMouseDevice = &device
			}
			// 優先マウスが指定されており、名前が一致するか確認
			if preferredMouse != "" && device.Name == preferredMouse {
				mouseDevice = &device
			}
		} else if device.Type == features.DeviceTypeKeyboard {
			// 最初のキーボードを記録
			if firstKeyboardDevice == nil {
				firstKeyboardDevice = &device
			}
			// 優先キーボードが指定されており、名前が一致するか確認
			if preferredKeyboard != "" && device.Name == preferredKeyboard {
				keyboardDevice = &device
			}
		}
	}

	// 優先デバイスが見つからなかった場合は最初のデバイスを使用
	if mouseDevice == nil {
		mouseDevice = firstMouseDevice
	}
	if keyboardDevice == nil {
		keyboardDevice = firstKeyboardDevice
	}

	if mouseDevice == nil {
		s.touchPad.Close()
		return fmt.Errorf("マウスデバイスが見つかりませんでした")
	}
	if keyboardDevice == nil {
		s.touchPad.Close()
		return fmt.Errorf("キーボードデバイスが見つかりませんでした")
	}

	log.Printf("使用するキーボード: %s", keyboardDevice.Name)
	log.Printf("使用するマウス: %s", mouseDevice.Name)

	// マウスとキーボードデバイスを開く
	mouse, err := features.CreateMouse(mouseDevice.Path)
	if err != nil {
		s.touchPad.Close()
		return fmt.Errorf("マウスデバイスのオープンに失敗しました[path=%s]: %v", mouseDevice.Path, err)
	}
	s.mouse = mouse

	keyboard, err := features.CreateKeyboard(keyboardDevice.Path)
	if err != nil {
		s.touchPad.Close()
		s.mouse.Close()
		return fmt.Errorf("キーボードデバイスのオープンに失敗しました: %v", err)
	}
	s.keyboard = keyboard

	s.stopChan = make(chan struct{})
	s.running = true

	// ジェスチャー認識のメインループを開始
	go s.runGestureLoop()

	return nil
}

// Stop はジェスチャー認識サービスを停止する
func (s *GestureService) Stop() error {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if !s.running {
		return fmt.Errorf("サービスは実行されていません")
	}

	close(s.stopChan)
	s.running = false

	// デバイスのクローズは runGestureLoop 内で行われる

	return nil
}

// UpdateConfig は設定を更新する
func (s *GestureService) UpdateConfig(cfg *config.Config) {
	select {
	case s.updateConfig <- cfg:
		// 設定更新チャネルに送信成功
	default:
		// チャネルがブロックされている場合は古い設定を破棄して新しい設定を送信
		select {
		case <-s.updateConfig:
		default:
		}
		s.updateConfig <- cfg
	}
}

// IsRunning はサービスが実行中かどうかを返す
func (s *GestureService) IsRunning() bool {
	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()
	return s.running
}

// runGestureLoop はジェスチャー認識のメインループ
func (s *GestureService) runGestureLoop() {
	defer func() {
		// サービス終了時にデバイスをクローズ
		if s.touchPad != nil {
			s.touchPad.Close()
		}
		if s.mouse != nil {
			s.mouse.Close()
		}
		if s.keyboard != nil {
			s.keyboard.Close()
		}
		log.Println("ジェスチャー認識サービスを停止しました")
	}()

	const maxFingers = 4
	var (
		fingerCount     int
		fingerPositions [maxFingers]struct{ x, y int32 }
		prevKey         int32
		grabbed         bool
		lastScrollTime  time.Time
	)

	// 設定値を取得するための関数（設定更新に対応）
	getCfg := func() *config.Config {
		select {
		case newCfg := <-s.updateConfig:
			log.Println("設定を更新しました")
			s.cfg = newCfg
		default:
		}
		return s.cfg
	}

	cfg := getCfg()
	motionFilter := features.NewMotionFilter(cfg.Motion.FilterSmoothingFactor, cfg.Motion.FilterWarmUpCount)

	log.Println("ジェスチャー認識を開始しました...")

	for {
		select {
		case <-s.stopChan:
			return
		default:
			cfg = getCfg()

			pressedKey := s.keyboard.GetKey()
			dxRaw, dyRaw := s.mouse.GetMouseDelta()
			dx, dy := motionFilter.Filter(dxRaw*int32(cfg.Motion.MouseDeltaFactor), dyRaw*int32(cfg.Motion.MouseDeltaFactor))

			now := time.Now()

			// 何も動いていない場合、最後のスクロールから閾値を超えていればリセット
			// これにより、タッチパッドの範囲内で無限にスクロールが可能
			if now.Sub(lastScrollTime) > cfg.Gesture.ResetThreshold && fingerCount > 0 {
				liftAllFingers(s.touchPad, fingerCount)
				motionFilter.Reset()
				initFingers(s.touchPad, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
			}
			lastScrollTime = now

			switch {
			case pressedKey == int32(cfg.Input.TwoFingerKey) && fingerCount == 0:
				if !grabbed {
					s.mouse.Grab()
					grabbed = true
				}
				log.Println("2本指ジェスチャー開始")
				fingerCount = 2
				initFingers(s.touchPad, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
				prevKey = pressedKey

			case pressedKey == int32(cfg.Input.FourFingerKey) && fingerCount == 0:
				if !grabbed {
					s.mouse.Grab()
					grabbed = true
				}
				log.Println("4本指ジェスチャー開始")
				fingerCount = 4
				initFingers(s.touchPad, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
				prevKey = pressedKey

			case (pressedKey == int32(cfg.Input.FourFingerKey) || pressedKey == int32(cfg.Input.TwoFingerKey)) && fingerCount > 0:
				if pressedKey == prevKey {
					for i := 0; i < fingerCount; i++ {
						fingerPositions[i].x += dx
						fingerPositions[i].y += dy

						fingerPositions[i].x = clamp(fingerPositions[i].x, cfg.TouchPad.MinX, cfg.TouchPad.MaxX)
						fingerPositions[i].y = clamp(fingerPositions[i].y, cfg.TouchPad.MinY, cfg.TouchPad.MaxY)

						_ = s.touchPad.MultiTouchMove(i, fingerPositions[i].x, fingerPositions[i].y)
					}
				} else {
					liftAllFingers(s.touchPad, fingerCount)
					motionFilter.Reset()
					fingerCount = 0
				}
				prevKey = pressedKey

			default:
				if grabbed {
					s.mouse.Release()
					grabbed = false
				}
				if fingerCount > 0 {
					liftAllFingers(s.touchPad, fingerCount)
					log.Println("ジェスチャー終了")
					fingerCount = 0
				}
				if pressedKey != 0 {
					prevKey = pressedKey
				} else {
					prevKey = 0
				}
			}

			time.Sleep(100 * time.Microsecond)
		}
	}
}

// initFingers は指の初期位置を設定する
func initFingers(padDevice features.TouchPad, fingerPositions *[maxFingers]struct{ x, y int32 }, count int, centerX, centerY int32) {
	offset := int32(20)
	startY := centerY - offset*(int32(count)-1)/2

	for i := 0; i < count; i++ {
		fingerPositions[i].x = centerX
		fingerPositions[i].y = startY + offset*int32(i)

		_ = padDevice.MultiTouchDown(i, i, fingerPositions[i].x, fingerPositions[i].y)
	}
}

// liftAllFingers はすべての指を持ち上げる
func liftAllFingers(padDevice features.TouchPad, count int) {
	for i := 0; i < count; i++ {
		_ = padDevice.MultiTouchUp(i)
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

const maxFingers = 4
