package api

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/features"
)

// GestureService はジェスチャー認識サービスを管理する構造体
type GestureService struct {
	cfg                   *config.Config
	stopChan              chan struct{}
	running               bool
	statusMutex           sync.RWMutex
	touchPad              features.TouchPad
	keyboard              features.Keyboard
	mouse                 features.Mouse
	keyboardDevice        *features.Device
	mouseDevice           *features.Device
	updateConfig          chan *config.Config
	deviceMonitor         *features.DeviceMonitor
	reconnectOnDisconnect bool
}

// NewGestureService は新しいジェスチャー認識サービスを作成する
func NewGestureService(cfg *config.Config) *GestureService {
	return &GestureService{
		cfg:                   cfg,
		stopChan:              make(chan struct{}),
		running:               false,
		updateConfig:          make(chan *config.Config, 1),
		reconnectOnDisconnect: true, // デフォルトで自動再接続を有効化
	}
}

// Start はジェスチャー認識サービスを開始する
func (s *GestureService) Start() error {
	log.Println("GestureService: Start メソッドが呼ばれました")

	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if s.running {
		return fmt.Errorf("サービスは既に実行中です")
	}

	// 仮想タッチパッドデバイスの作成
	log.Println("仮想タッチパッドデバイスを作成します")
	padDevice, err := features.CreateTouchPad("/dev/uinput", []byte("VirtualTouchPad"),
		s.cfg.TouchPad.MinX, s.cfg.TouchPad.MaxX, s.cfg.TouchPad.MinY, s.cfg.TouchPad.MaxY)
	if err != nil {
		return fmt.Errorf("仮想タッチパッドの作成に失敗しました: %v", err)
	}
	s.touchPad = padDevice
	log.Println("仮想タッチパッドデバイスの作成に成功しました")

	// デバイス一覧の取得（デバイスモニターを使用せずに直接取得）
	log.Println("デバイス一覧を直接取得します")
	devices, err := features.GetDevices()
	if err != nil {
		s.touchPad.Close()
		return fmt.Errorf("デバイス一覧の取得に失敗しました: %v", err)
	}
	log.Printf("検出されたデバイス数: %d", len(devices))

	// 設定ファイルで指定された優先デバイスまたは最初のマウスとキーボードを使用
	var mouseDevice *features.Device
	var keyboardDevice *features.Device

	// 優先デバイスの名前
	preferredKeyboard := s.cfg.DevicePrefs.PreferredKeyboardDevice
	preferredMouse := s.cfg.DevicePrefs.PreferredMouseDevice
	log.Printf("優先キーボード: %s, 優先マウス: %s", preferredKeyboard, preferredMouse)

	// 最初に見つかったデバイスを初期値として保存
	var firstMouseDevice *features.Device
	var firstKeyboardDevice *features.Device

	for i := range devices {
		device := &devices[i]
		log.Printf("検出デバイス: 名前=%s, パス=%s, タイプ=%v", device.Name, device.Path, device.Type)

		if device.Type == features.DeviceTypeMouse {
			// 最初のマウスを記録
			if firstMouseDevice == nil {
				firstMouseDevice = device
				log.Printf("最初のマウスデバイス: %s", device.Name)
			}
			// 優先マウスが指定されており、名前が一致するか確認
			if preferredMouse != "" && strings.Contains(device.Name, preferredMouse) {
				mouseDevice = device
				log.Printf("優先マウスデバイスが見つかりました: %s", device.Name)
			}
		} else if device.Type == features.DeviceTypeKeyboard {
			// 最初のキーボードを記録
			if firstKeyboardDevice == nil {
				firstKeyboardDevice = device
				log.Printf("最初のキーボードデバイス: %s", device.Name)
			}
			// 優先キーボードが指定されており、名前が一致するか確認
			if preferredKeyboard != "" && strings.Contains(device.Name, preferredKeyboard) {
				keyboardDevice = device
				log.Printf("優先キーボードデバイスが見つかりました: %s", device.Name)
			}
		}
	}

	// 優先デバイスが見つからなかった場合は最初のデバイスを使用
	if mouseDevice == nil {
		mouseDevice = firstMouseDevice
		if mouseDevice != nil {
			log.Printf("優先マウスが見つからないため、最初のマウスを使用: %s", mouseDevice.Name)
		}
	}
	if keyboardDevice == nil {
		keyboardDevice = firstKeyboardDevice
		if keyboardDevice != nil {
			log.Printf("優先キーボードが見つからないため、最初のキーボードを使用: %s", keyboardDevice.Name)
		}
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

	// デバイス情報を保存
	s.keyboardDevice = keyboardDevice
	s.mouseDevice = mouseDevice

	// マウスとキーボードデバイスを開く
	log.Printf("マウスデバイスをオープン: %s", mouseDevice.Path)
	mouse, err := features.CreateMouse(mouseDevice.Path)
	if err != nil {
		s.touchPad.Close()
		return fmt.Errorf("マウスデバイスのオープンに失敗しました[path=%s]: %v", mouseDevice.Path, err)
	}
	s.mouse = mouse
	log.Println("マウスデバイスのオープンに成功しました")

	log.Printf("キーボードデバイスをオープン: %s", keyboardDevice.Path)
	keyboard, err := features.CreateKeyboard(keyboardDevice.Path)
	if err != nil {
		s.touchPad.Close()
		s.mouse.Close()
		return fmt.Errorf("キーボードデバイスのオープンに失敗しました: %v", err)
	}
	s.keyboard = keyboard
	log.Println("キーボードデバイスのオープンに成功しました")

	// デバイスモニターを非同期で初期化
	go func() {
		log.Println("非同期でデバイスモニターを初期化します")
		deviceMonitor, err := features.GetDeviceMonitor()
		if err != nil {
			log.Printf("警告: デバイスモニターの初期化に失敗しました: %v", err)
			return
		}

		s.statusMutex.Lock()
		s.deviceMonitor = deviceMonitor
		s.statusMutex.Unlock()
		log.Println("デバイスモニターの初期化に成功しました")

		// デバイスの切断・再接続を処理するコールバックを登録
		deviceMonitor.RegisterCallback(func(event features.DeviceEvent) {
			log.Printf("デバイスイベントコールバック: タイプ=%v, パス=%s, デバイスタイプ=%v",
				event.Type, event.Path, event.Device.Type)

			if !s.reconnectOnDisconnect {
				log.Println("自動再接続は無効化されています")
				return
			}

			s.statusMutex.Lock()
			defer s.statusMutex.Unlock()

			if !s.running {
				log.Println("サービスが実行中ではないため、イベントを無視します")
				return
			}

			// デバイスが削除された場合
			if event.Type == features.DeviceRemoved {
				// キーボードデバイスの確認
				if s.keyboard != nil && s.keyboardDevice != nil {
					log.Printf("キーボードパスの比較: イベント=%s vs 現在=%s",
						event.Path, s.keyboardDevice.Path)
				}

				// マウスデバイスの確認
				if s.mouse != nil && s.mouseDevice != nil {
					log.Printf("マウスパスの比較: イベント=%s vs 現在=%s",
						event.Path, s.mouseDevice.Path)
				}

				// 監視しているデバイスが削除されたかチェック
				if s.keyboard != nil && s.keyboardDevice != nil && event.Path == s.keyboardDevice.Path {
					log.Printf("キーボードデバイスが切断されました: %s", event.Path)
					// 自動再接続を試みる
					go s.attemptReconnect()
				} else if s.mouse != nil && s.mouseDevice != nil && event.Path == s.mouseDevice.Path {
					log.Printf("マウスデバイスが切断されました: %s", event.Path)
					// 自動再接続を試みる
					go s.attemptReconnect()
				} else {
					log.Printf("このデバイス切断は無視します: %s", event.Path)
				}
			}
		})
		log.Println("デバイスモニターにコールバックを登録しました")
	}()

	s.stopChan = make(chan struct{})
	s.running = true

	// ジェスチャー認識のメインループを開始
	log.Println("ジェスチャー認識のメインループを開始します")
	go s.runGestureLoop()

	// デバイスの健全性チェックを定期的に実行
	go s.runDeviceHealthCheck()

	return nil
}

// 再接続を試みる新しいメソッド
func (s *GestureService) attemptReconnect() {
	// サービスが実行中でなければ何もしない
	s.statusMutex.Lock()
	if !s.running || !s.reconnectOnDisconnect {
		s.statusMutex.Unlock()
		return
	}
	s.statusMutex.Unlock()

	log.Println("デバイスの再接続を試みています...")

	// 既存のリソースをクリーンアップ
	s.statusMutex.Lock()
	if s.keyboard != nil {
		log.Println("既存のキーボードデバイスをクローズします")
		_ = s.keyboard.Close()
		s.keyboard = nil
	}
	if s.mouse != nil {
		log.Println("既存のマウスデバイスをクローズします")
		_ = s.mouse.Close()
		s.mouse = nil
	}
	s.statusMutex.Unlock()

	// 一度プロセスを止めてudevが新しいデバイスファイルを作成する時間を与える
	log.Println("デバイスの再認識のため3秒間待機します...")
	time.Sleep(3 * time.Second)

	// 最大5回まで再接続を試みる
	for attempt := 0; attempt < 5; attempt++ {
		log.Printf("再接続試行 %d 回目", attempt+1)

		s.statusMutex.Lock()
		if !s.running || !s.reconnectOnDisconnect {
			log.Println("サービスが停止しているため、再接続を中止します")
			s.statusMutex.Unlock()
			return
		}

		// デバイス一覧を再取得（キャッシュを使わず強制的に再スキャン）
		log.Println("デバイス一覧を強制的に再スキャンします")
		devices, err := features.ScanDevices() // 公開された関数を使用
		if err != nil {
			log.Printf("デバイス一覧の取得に失敗しました: %v", err)
			s.statusMutex.Unlock()

			// 少し待ってから再試行
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("検出されたデバイス数: %d", len(devices))
		for i, dev := range devices {
			log.Printf("  デバイス %d: 名前=%s, パス=%s, タイプ=%v", i, dev.Name, dev.Path, dev.Type)
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

		// 既存のデバイス検索ロジックと同じ
		for i := range devices {
			device := &devices[i]
			if device.Type == features.DeviceTypeMouse {
				if firstMouseDevice == nil {
					firstMouseDevice = device
				}
				if preferredMouse != "" && strings.Contains(device.Name, preferredMouse) {
					mouseDevice = device
					log.Printf("優先マウスデバイスが見つかりました: %s → %s", device.Name, device.Path)
				}
			} else if device.Type == features.DeviceTypeKeyboard {
				if firstKeyboardDevice == nil {
					firstKeyboardDevice = device
				}
				if preferredKeyboard != "" && strings.Contains(device.Name, preferredKeyboard) {
					keyboardDevice = device
					log.Printf("優先キーボードデバイスが見つかりました: %s → %s", device.Name, device.Path)
				}
			}
		}

		// 優先デバイスが見つからなかった場合は最初のデバイスを使用
		if mouseDevice == nil {
			mouseDevice = firstMouseDevice
			if mouseDevice != nil {
				log.Printf("優先マウスが見つからないため、最初のマウスを使用: %s → %s", mouseDevice.Name, mouseDevice.Path)
			}
		}
		if keyboardDevice == nil {
			keyboardDevice = firstKeyboardDevice
			if keyboardDevice != nil {
				log.Printf("優先キーボードが見つからないため、最初のキーボードを使用: %s → %s", keyboardDevice.Name, keyboardDevice.Path)
			}
		}

		// 必要なデバイスが見つからない場合は次の試行へ
		if mouseDevice == nil || keyboardDevice == nil {
			log.Println("必要なデバイスが見つかりませんでした。再試行します...")
			s.statusMutex.Unlock()
			time.Sleep(2 * time.Second)
			continue
		}

		// 新しいデバイスを開く
		log.Printf("キーボードデバイスをオープン: %s", keyboardDevice.Path)
		keyboard, err := features.CreateKeyboard(keyboardDevice.Path)
		if err != nil {
			log.Printf("キーボードデバイスのオープンに失敗しました: %v", err)
			s.statusMutex.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("マウスデバイスをオープン: %s", mouseDevice.Path)
		mouse, err := features.CreateMouse(mouseDevice.Path)
		if err != nil {
			_ = keyboard.Close()
			log.Printf("マウスデバイスのオープンに失敗しました: %v", err)
			s.statusMutex.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		// デバイスの参照を更新
		s.keyboard = keyboard
		s.mouse = mouse
		s.keyboardDevice = keyboardDevice
		s.mouseDevice = mouseDevice

		log.Printf("デバイスの再接続に成功しました - キーボード: %s, マウス: %s",
			keyboardDevice.Name, mouseDevice.Name)

		// デバイスモニターを更新
		if s.deviceMonitor != nil {
			go func() {
				// 新しい状態をデバイスモニターにも反映
				_, _ = features.RescanDevices()
			}()
		}

		s.statusMutex.Unlock()
		return
	}

	log.Println("デバイスの再接続に失敗しました。再接続を停止します。")
}

// Stop はジェスチャー認識サービスを停止する
func (s *GestureService) Stop() error {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if !s.running {
		return fmt.Errorf("サービスは実行されていません")
	}

	// 再接続フラグをオフに
	s.reconnectOnDisconnect = false

	close(s.stopChan)
	s.running = false

	// デバイスのクローズは runGestureLoop 内で行われる

	return nil
}

// SetAutoReconnect は自動再接続機能のオン/オフを切り替える
func (s *GestureService) SetAutoReconnect(enable bool) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	s.reconnectOnDisconnect = enable
	log.Printf("デバイス自動再接続: %v", enable)
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

			// デバイス参照をsafeにアクセスするためにロックを取得
			s.statusMutex.RLock()

			// キーボードとマウスの参照が有効かチェック
			keyboardAlive := s.keyboard != nil
			mouseAlive := s.mouse != nil

			var pressedKey int32 = -1
			var dxRaw, dyRaw int32 = 0, 0

			// キーボードがある場合のみ入力を取得
			if keyboardAlive {
				pressedKey = s.keyboard.GetKey()
			}

			// マウスがある場合のみ移動を取得
			if mouseAlive {
				dxRaw, dyRaw = s.mouse.GetMouseDelta()
			}

			// ロックを解放（後続の処理でデバイスを参照しない）
			s.statusMutex.RUnlock()

			// デバイスがない場合は処理をスキップ
			if !keyboardAlive || !mouseAlive {
				time.Sleep(500 * time.Millisecond)
				continue
			}

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
					s.statusMutex.RLock()
					if s.mouse != nil {
						s.mouse.Grab()
						grabbed = true
					}
					s.statusMutex.RUnlock()
				}
				log.Println("2本指ジェスチャー開始")
				fingerCount = 2
				initFingers(s.touchPad, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
				prevKey = pressedKey

			case pressedKey == int32(cfg.Input.FourFingerKey) && fingerCount == 0:
				if !grabbed {
					s.statusMutex.RLock()
					if s.mouse != nil {
						s.mouse.Grab()
						grabbed = true
					}
					s.statusMutex.RUnlock()
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
					s.statusMutex.RLock()
					if s.mouse != nil {
						s.mouse.Release()
					}
					s.statusMutex.RUnlock()
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

// runDeviceHealthCheck はデバイスの健全性チェックを定期的に実行する
func (s *GestureService) runDeviceHealthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("デバイス健全性監視を開始します")

	for {
		select {
		case <-s.stopChan:
			log.Println("デバイス健全性監視を終了します")
			return
		case <-ticker.C:
			s.statusMutex.Lock()
			if !s.running {
				s.statusMutex.Unlock()
				return
			}

			// デバイスの状態確認
			devicesFailed := false

			// キーボードデバイスのテスト
			if s.keyboard != nil {
				// キーボードデバイスへのアクセスを試みる
				key := s.keyboard.GetKey()
				// エラーかどうかは結果ではなくファイル状態で判断
				if !s.isKeyboardDeviceAlive() {
					log.Printf("キーボードデバイスが応答しません: GetKey=%d", key)
					devicesFailed = true
				}
			} else {
				log.Println("キーボードデバイスがnilです")
				devicesFailed = true
			}

			// マウスデバイスのテスト
			if s.mouse != nil {
				// マウスデバイスへのアクセスを試みる
				if !s.isMouseDeviceAlive() {
					log.Println("マウスデバイスが応答しません")
					devicesFailed = true
				}
			} else {
				log.Println("マウスデバイスがnilです")
				devicesFailed = true
			}

			if devicesFailed {
				log.Println("デバイスが正常に応答しないため、再接続を試みます...")
				s.statusMutex.Unlock()
				go s.attemptReconnect()
			} else {
				s.statusMutex.Unlock()
				log.Println("デバイス健全性チェック完了: すべてのデバイスが正常です")
			}
		}
	}
}

// isKeyboardDeviceAlive はキーボードデバイスへのアクセスをテストし、生きているかどうかを返す
func (s *GestureService) isKeyboardDeviceAlive() bool {
	if s.keyboardDevice == nil {
		return false
	}

	// デバイスファイルが存在するか確認
	if _, err := os.Stat(s.keyboardDevice.Path); os.IsNotExist(err) {
		log.Printf("キーボードデバイスファイルが存在しません: %s", s.keyboardDevice.Path)
		return false
	}

	return true
}

// isMouseDeviceAlive はマウスデバイスへのアクセスをテストし、生きているかどうかを返す
func (s *GestureService) isMouseDeviceAlive() bool {
	if s.mouseDevice == nil {
		return false
	}

	// デバイスファイルが存在するか確認
	if _, err := os.Stat(s.mouseDevice.Path); os.IsNotExist(err) {
		log.Printf("マウスデバイスファイルが存在しません: %s", s.mouseDevice.Path)
		return false
	}

	return true
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
