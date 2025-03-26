package gui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/char5742/keyball-gestures/internal/features"
)

// ServiceStatus はサービスの状態を表す型
type ServiceStatus int

const (
	StatusStopped ServiceStatus = iota
	StatusRunning
	StatusError
)

// ServiceManager はジェスチャーサービスを管理する構造体
type ServiceManager struct {
	app             *App
	deviceManager   *DeviceManager
	configManager   *ConfigManager
	status          ServiceStatus
	statusLabel     *widget.Label
	startStopButton *widget.Button
	logOutput       *widget.Entry
	padDevice       features.TouchPad
	mouseDevice     features.Mouse
	keyboardDevice  features.Keyboard
	stopChan        chan struct{}
	motionFilter    *features.MotionFilter
	monitor         *GestureMonitor
	mutex           sync.Mutex
}

// NewServiceManager は新しいサービスマネージャを作成する
func NewServiceManager(app *App, deviceManager *DeviceManager, configManager *ConfigManager) *ServiceManager {
	sm := &ServiceManager{
		app:           app,
		deviceManager: deviceManager,
		configManager: configManager,
		status:        StatusStopped,
		stopChan:      make(chan struct{}),
	}

	return sm
}

// CreateServicePanel はサービス管理パネルを作成する
func (sm *ServiceManager) CreateServicePanel() fyne.CanvasObject {
	sm.statusLabel = widget.NewLabel("ステータス: 停止中")

	sm.startStopButton = widget.NewButton("サービス開始", func() {
		if sm.status == StatusRunning {
			sm.StopService()
		} else {
			sm.StartService()
		}
	})

	sm.logOutput = widget.NewMultiLineEntry()
	sm.logOutput.SetPlaceHolder("ログ出力")
	sm.logOutput.Disable()

	// ジェスチャーモニターを作成
	sm.monitor = NewGestureMonitor()
	monitorView := sm.monitor.CreateMonitorView()

	return container.NewVBox(
		container.NewHBox(sm.startStopButton, sm.statusLabel),
		widget.NewSeparator(),
		widget.NewLabel("ジェスチャーモニター:"),
		monitorView,
		widget.NewSeparator(),
		widget.NewLabel("ログ:"),
		container.NewScroll(sm.logOutput),
	)
}

// StartService はサービスを開始する
func (sm *ServiceManager) StartService() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.status == StatusRunning {
		sm.Log("サービスは既に実行中です")
		return
	}

	// ステータス更新
	sm.status = StatusRunning
	sm.startStopButton.SetText("サービス停止")
	sm.statusLabel.SetText("ステータス: 実行中")

	go func() {
		err := sm.runServiceLoop()
		if err != nil {
			sm.mutex.Lock()
			sm.status = StatusError
			sm.startStopButton.SetText("サービス開始")
			sm.statusLabel.SetText("ステータス: エラー")
			sm.mutex.Unlock()
			sm.Log("エラー: " + err.Error())
		}
	}()

	sm.Log("サービスを開始しました")
}

// StopService はサービスを停止する
func (sm *ServiceManager) StopService() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.status != StatusRunning {
		sm.Log("サービスは実行中ではありません")
		return
	}

	// 停止シグナル送信
	close(sm.stopChan)
	sm.stopChan = make(chan struct{})

	// ステータス更新
	sm.status = StatusStopped
	sm.startStopButton.SetText("サービス開始")
	sm.statusLabel.SetText("ステータス: 停止中")

	// デバイスをクローズ
	if sm.padDevice != nil {
		sm.padDevice.Close()
		sm.padDevice = nil
	}
	if sm.mouseDevice != nil {
		sm.mouseDevice.Close()
		sm.mouseDevice = nil
	}
	if sm.keyboardDevice != nil {
		sm.keyboardDevice.Close()
		sm.keyboardDevice = nil
	}

	sm.Log("サービスを停止しました")
}

// Log はログメッセージを追加する
func (sm *ServiceManager) Log(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// UIスレッドで実行
	if sm.logOutput != nil {
		sm.logOutput.SetText(sm.logOutput.Text + logMessage)
		// 自動スクロール (簡易的な実装)
		sm.logOutput.CursorRow = len(sm.logOutput.Text)
	}
}

// runServiceLoop はサービスのメインループを実行する
func (sm *ServiceManager) runServiceLoop() error {
	// 設定を取得
	config := sm.configManager.GetConfig()

	// 仮想タッチパッドの作成
	var err error
	sm.padDevice, err = features.CreateTouchPad("/dev/uinput", []byte("VirtualTouchPad"), 0, 32767, 0, 32767)
	if err != nil {
		return fmt.Errorf("仮想タッチパッドの作成に失敗しました: %w", err)
	}

	// キーボードとマウスデバイスを取得
	keyboard := sm.deviceManager.GetSelectedKeyboard()
	mouse := sm.deviceManager.GetSelectedMouse()

	if keyboard == nil {
		return fmt.Errorf("キーボードデバイスが選択されていません")
	}
	if mouse == nil {
		return fmt.Errorf("マウスデバイスが選択されていません")
	}

	sm.Log(fmt.Sprintf("キーボード: %s", keyboard.Name))
	sm.Log(fmt.Sprintf("マウス: %s", mouse.Name))

	// デバイスを開く
	sm.mouseDevice, err = features.CreateMouse(mouse.Path)
	if err != nil {
		return fmt.Errorf("マウスデバイスのオープンに失敗しました: %w", err)
	}

	sm.keyboardDevice, err = features.CreateKeyboard(keyboard.Path)
	if err != nil {
		return fmt.Errorf("キーボードデバイスのオープンに失敗しました: %w", err)
	}

	// モーションフィルタの作成
	sm.motionFilter = features.NewMotionFilter(
		config.MotionFilterFactor,
		config.MotionFilterWarmUpCount,
	)

	sm.Log("ジェスチャーサービスを初期化しました")
	sm.Log(fmt.Sprintf("2本指スワイプキー: %d, 4本指スワイプキー: %d",
		config.TwoFingerKey, config.FourFingerKey))

	// ジェスチャーループ
	return sm.gestureLoop(config)
}

// gestureLoop はジェスチャー認識のメインループ
func (sm *ServiceManager) gestureLoop(config Config) error {
	const maxFingers = 4

	var (
		fingerCount     int
		fingerPositions [maxFingers]struct{ x, y int32 }
		prevKey         int32
		grabbed         bool
		lastScrollTime  time.Time
	)

	resetThreshold := time.Duration(config.ResetThresholdMillisec) * time.Millisecond

	// Monitor用
	lastDx, lastDy := int32(0), int32(0)

	for {
		select {
		case <-sm.stopChan:
			return nil
		default:
			// キー状態とマウス移動を取得
			pressedKey := sm.keyboardDevice.GetKey()
			dxRaw, dyRaw := sm.mouseDevice.GetMouseDelta()
			dx, dy := sm.motionFilter.Filter(dxRaw*int32(config.MouseDeltaFactor), dyRaw*int32(config.MouseDeltaFactor))

			// モニタ更新（有意な変化があった場合のみ）
			if dx != 0 || dy != 0 {
				lastDx, lastDy = dx, dy
				sm.monitor.UpdateGestureDisplay(fingerCount, lastDx, lastDy)
			}

			now := time.Now()

			// リセット処理
			if now.Sub(lastScrollTime) > resetThreshold && fingerCount > 0 {
				liftAllFingers(sm.padDevice, fingerCount)
				sm.motionFilter.Reset()
				initFingers(sm.padDevice, &fingerPositions, fingerCount, 32767/2, 32767/2)
				sm.Log("フィンガー位置をリセットしました")
			}

			if pressedKey != 0 || dx != 0 || dy != 0 {
				lastScrollTime = now
			}

			// ジェスチャー開始/更新/終了処理
			switch {
			case pressedKey == int32(config.TwoFingerKey) && fingerCount == 0:
				if !grabbed {
					sm.mouseDevice.Grab()
					grabbed = true
				}
				fingerCount = 2
				initFingers(sm.padDevice, &fingerPositions, fingerCount, 32767/2, 32767/2)
				prevKey = pressedKey
				sm.Log("2本指ジェスチャー開始")

			case pressedKey == int32(config.FourFingerKey) && fingerCount == 0:
				if !grabbed {
					sm.mouseDevice.Grab()
					grabbed = true
				}
				fingerCount = 4
				initFingers(sm.padDevice, &fingerPositions, fingerCount, 32767/2, 32767/2)
				prevKey = pressedKey
				sm.Log("4本指ジェスチャー開始")

			case (pressedKey == int32(config.TwoFingerKey) || pressedKey == int32(config.FourFingerKey)) && fingerCount > 0:
				if pressedKey == prevKey {
					// フィンガー移動
					for i := 0; i < fingerCount; i++ {
						fingerPositions[i].x += dx
						fingerPositions[i].y += dy

						// 範囲内に制限
						fingerPositions[i].x = clamp(fingerPositions[i].x, 0, 32767)
						fingerPositions[i].y = clamp(fingerPositions[i].y, 0, 32767)

						_ = sm.padDevice.MultiTouchMove(i, fingerPositions[i].x, fingerPositions[i].y)
					}
				} else {
					// キーが変わった場合はリセット
					liftAllFingers(sm.padDevice, fingerCount)
					sm.motionFilter.Reset()
					fingerCount = 0
				}
				prevKey = pressedKey

			default:
				// ジェスチャー終了
				if grabbed {
					sm.mouseDevice.Release()
					grabbed = false
				}
				if fingerCount > 0 {
					liftAllFingers(sm.padDevice, fingerCount)
					sm.Log("ジェスチャー終了")
					fingerCount = 0
				}
				if pressedKey != 0 {
					prevKey = pressedKey
				} else {
					prevKey = 0
				}
			}
		}

		// スリープで負荷軽減
		time.Sleep(100 * time.Microsecond)
	}
}

// GestureMonitor はジェスチャーの視覚的表示を管理する
type GestureMonitor struct {
	display *widget.Label
}

// NewGestureMonitor は新しいジェスチャーモニターを作成する
func NewGestureMonitor() *GestureMonitor {
	return &GestureMonitor{
		display: widget.NewLabel("ジェスチャーなし"),
	}
}

// CreateMonitorView はモニターのビューを作成する
func (gm *GestureMonitor) CreateMonitorView() fyne.CanvasObject {
	return gm.display
}

// UpdateGestureDisplay はジェスチャーの表示を更新する
func (gm *GestureMonitor) UpdateGestureDisplay(fingerCount int, dx, dy int32) {
	var direction string

	// 方向を判定
	if abs(dx) > abs(dy) {
		// 水平方向
		if dx > 0 {
			direction = "→ 右"
		} else {
			direction = "← 左"
		}
	} else {
		// 垂直方向
		if dy > 0 {
			direction = "↓ 下"
		} else {
			direction = "↑ 上"
		}
	}

	gesture := fmt.Sprintf("%d本指スワイプ: %s (dx=%d, dy=%d)",
		fingerCount, direction, dx, dy)

	gm.display.SetText(gesture)
}

// ユーティリティ関数

// initFingers は複数の指を初期化する
func initFingers(padDevice features.TouchPad, fingerPositions *[4]struct{ x, y int32 }, count int, centerX, centerY int32) {
	offset := int32(20)
	startY := centerY - offset*(int32(count)-1)/2

	for i := 0; i < count; i++ {
		fingerPositions[i].x = centerX
		fingerPositions[i].y = startY + offset*int32(i)

		_ = padDevice.MultiTouchDown(i, i, fingerPositions[i].x, fingerPositions[i].y)
	}
}

// liftAllFingers は全ての指を持ち上げる
func liftAllFingers(padDevice features.TouchPad, count int) {
	for i := 0; i < count; i++ {
		_ = padDevice.MultiTouchUp(i)
	}
}

// clamp は値を範囲内に制限する
func clamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// abs は絶対値を返す
func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}
