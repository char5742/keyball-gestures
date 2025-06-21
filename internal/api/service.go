package api

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/core"
	"github.com/char5742/keyball-gestures/internal/platform"
	"github.com/char5742/keyball-gestures/internal/platform/common"
)

// GestureService は新しいアーキテクチャに対応したジェスチャー認識サービス
type GestureService struct {
	cfg                   *config.Config
	stopChan              chan struct{}
	running               bool
	statusMutex           sync.RWMutex
	updateConfig          chan *config.Config
	reconnectOnDisconnect bool

	// プラットフォーム共通インターフェース
	factory       common.PlatformFactory
	scrollEmitter common.ScrollEmitter
	mouseInput    common.MouseInput
	keyboardInput common.KeyboardInput
	deviceMonitor common.DeviceMonitor

	// コアコンポーネント
	gestureDetector *core.GestureDetector
	motionFilter    *core.MotionFilter
}

// NewGestureService は新しいジェスチャー認識サービスを作成する
func NewGestureService(cfg *config.Config) *GestureService {
	return &GestureService{
		cfg:                   cfg,
		stopChan:              make(chan struct{}),
		running:               false,
		updateConfig:          make(chan *config.Config, 1),
		reconnectOnDisconnect: true,
	}
}

// Start はジェスチャー認識サービスを開始する
func (s *GestureService) Start() error {
	log.Printf("GestureService: Start メソッドが呼ばれました (OS: %s)", runtime.GOOS)

	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if s.running {
		return fmt.Errorf("サービスは既に実行中です")
	}

	// プラットフォームファクトリーを作成
	s.factory = platform.NewPlatformFactory(s.cfg)

	// スクロールエミッターを作成
	log.Println("スクロールエミッターを作成します")
	scrollEmitter, err := s.factory.CreateScrollEmitter()
	if err != nil {
		return fmt.Errorf("スクロールエミッターの作成に失敗しました: %v", err)
	}
	s.scrollEmitter = scrollEmitter
	log.Println("スクロールエミッターの作成に成功しました")

	// 利用可能なデバイスを取得
	log.Println("デバイス一覧を取得します")
	devices, err := s.factory.GetAvailableDevices()
	if err != nil {
		s.scrollEmitter.Close()
		return fmt.Errorf("デバイス一覧の取得に失敗しました: %v", err)
	}
	log.Printf("検出されたデバイス数: %d", len(devices))

	// デバイスを選択
	mouseDevice, keyboardDevice := s.selectDevices(devices)
	if mouseDevice == nil {
		s.scrollEmitter.Close()
		return fmt.Errorf("マウスデバイスが見つかりませんでした")
	}
	if keyboardDevice == nil {
		s.scrollEmitter.Close()
		return fmt.Errorf("キーボードデバイスが見つかりませんでした")
	}

	log.Printf("使用するキーボード: %s", keyboardDevice.Name)
	log.Printf("使用するマウス: %s", mouseDevice.Name)

	// マウス入力を作成
	log.Printf("マウス入力デバイスを作成: %s", mouseDevice.Path)
	mouseInput, err := s.factory.CreateMouseInput(mouseDevice.Path)
	if err != nil {
		s.scrollEmitter.Close()
		return fmt.Errorf("マウス入力デバイスの作成に失敗しました: %v", err)
	}
	s.mouseInput = mouseInput
	log.Println("マウス入力デバイスの作成に成功しました")

	// キーボード入力を作成
	log.Printf("キーボード入力デバイスを作成: %s", keyboardDevice.Path)
	keyboardInput, err := s.factory.CreateKeyboardInput(keyboardDevice.Path)
	if err != nil {
		s.scrollEmitter.Close()
		s.mouseInput.Close()
		return fmt.Errorf("キーボード入力デバイスの作成に失敗しました: %v", err)
	}
	s.keyboardInput = keyboardInput
	log.Println("キーボード入力デバイスの作成に成功しました")

	// コアコンポーネントを初期化
	s.gestureDetector = core.NewGestureDetector(s.cfg)
	s.motionFilter = core.NewMotionFilter(s.cfg.Motion.FilterSmoothingFactor, s.cfg.Motion.FilterWarmUpCount)

	// デバイスモニターを非同期で初期化
	go func() {
		log.Println("非同期でデバイスモニターを初期化します")
		deviceMonitor, err := s.factory.CreateDeviceMonitor()
		if err != nil {
			log.Printf("警告: デバイスモニターの初期化に失敗しました: %v", err)
			return
		}

		s.statusMutex.Lock()
		s.deviceMonitor = deviceMonitor
		s.statusMutex.Unlock()

		// デバイスイベントのコールバックを登録
		deviceMonitor.RegisterCallback(func(event common.DeviceEvent) {
			log.Printf("デバイスイベント: タイプ=%v, パス=%s, デバイスタイプ=%v",
				event.Type, event.Path, event.DeviceType)

			if !s.reconnectOnDisconnect {
				return
			}

			// デバイスが削除された場合の処理
			if event.Type == common.DeviceRemoved {
				// TODO: 再接続処理
			}
		})
		deviceMonitor.Start()
		log.Println("デバイスモニターの初期化に成功しました")
	}()

	s.stopChan = make(chan struct{})
	s.running = true

	// ジェスチャー認識のメインループを開始
	log.Println("ジェスチャー認識のメインループを開始します")
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

	s.reconnectOnDisconnect = false
	close(s.stopChan)
	s.running = false

	return nil
}

// UpdateConfig は設定を更新する
func (s *GestureService) UpdateConfig(cfg *config.Config) {
	select {
	case s.updateConfig <- cfg:
		// 設定更新チャネルに送信成功
	default:
		// 古い設定を破棄して新しい設定を送信
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

// selectDevices は利用可能なデバイスから使用するデバイスを選択する
func (s *GestureService) selectDevices(devices []common.Device) (*common.Device, *common.Device) {
	var mouseDevice *common.Device
	var keyboardDevice *common.Device

	// 優先デバイスの名前
	preferredKeyboard := s.cfg.DevicePrefs.PreferredKeyboardDevice
	preferredMouse := s.cfg.DevicePrefs.PreferredMouseDevice

	// 最初に見つかったデバイスを初期値として保存
	var firstMouseDevice *common.Device
	var firstKeyboardDevice *common.Device

	for i := range devices {
		device := &devices[i]
		log.Printf("検出デバイス: 名前=%s, パス=%s, タイプ=%v", device.Name, device.Path, device.Type)

		if device.Type == common.DeviceTypeMouse {
			if firstMouseDevice == nil {
				firstMouseDevice = device
			}
			if preferredMouse != "" && strings.Contains(device.Name, preferredMouse) {
				mouseDevice = device
				log.Printf("優先マウスデバイスが見つかりました: %s", device.Name)
			}
		} else if device.Type == common.DeviceTypeKeyboard {
			if firstKeyboardDevice == nil {
				firstKeyboardDevice = device
			}
			if preferredKeyboard != "" && strings.Contains(device.Name, preferredKeyboard) {
				keyboardDevice = device
				log.Printf("優先キーボードデバイスが見つかりました: %s", device.Name)
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

	return mouseDevice, keyboardDevice
}

// runGestureLoop はジェスチャー認識のメインループ
func (s *GestureService) runGestureLoop() {
	defer func() {
		// サービス終了時にリソースをクリーンアップ
		if s.scrollEmitter != nil {
			s.scrollEmitter.Close()
		}
		if s.mouseInput != nil {
			s.mouseInput.Close()
		}
		if s.keyboardInput != nil {
			s.keyboardInput.Close()
		}
		if s.deviceMonitor != nil {
			s.deviceMonitor.Stop()
			s.deviceMonitor.Close()
		}
		log.Println("ジェスチャー認識サービスを停止しました")
	}()

	var grabbed bool

	// 設定値を取得するための関数
	getCfg := func() *config.Config {
		select {
		case newCfg := <-s.updateConfig:
			log.Println("設定を更新しました")
			s.cfg = newCfg
			s.gestureDetector.UpdateConfig(newCfg)
			s.motionFilter = core.NewMotionFilter(newCfg.Motion.FilterSmoothingFactor, newCfg.Motion.FilterWarmUpCount)
		default:
		}
		return s.cfg
	}

	cfg := getCfg()
	log.Println("ジェスチャー認識を開始しました...")

	for {
		select {
		case <-s.stopChan:
			return
		default:
			cfg = getCfg()

			// デバイス参照をsafeにアクセス
			s.statusMutex.RLock()
			keyboardAlive := s.keyboardInput != nil
			mouseAlive := s.mouseInput != nil

			var pressedKey int32 = -1
			var dxRaw, dyRaw int32 = 0, 0

			if keyboardAlive {
				pressedKey = s.keyboardInput.GetKey()
			}
			if mouseAlive {
				dxRaw, dyRaw = s.mouseInput.GetMouseDelta()
			}
			
			// マウス移動量のログを削除
			// if pressedKey > 0 && (dxRaw != 0 || dyRaw != 0) {
			// 	log.Printf("[Debug] Mouse movement detected: dx=%d, dy=%d while key %d pressed", dxRaw, dyRaw, pressedKey)
			// }
			s.statusMutex.RUnlock()

			if !keyboardAlive || !mouseAlive {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// モーションフィルターを適用
			dx, dy := s.motionFilter.Filter(
				dxRaw*int32(cfg.Motion.MouseDeltaFactor),
				dyRaw*int32(cfg.Motion.MouseDeltaFactor),
			)

			// ジェスチャー検出
			gestureEvent := s.gestureDetector.ProcessInput(pressedKey, dx, dy)
			if gestureEvent != nil {
				// ジェスチャーイベントが発生した場合
				switch gestureEvent.State {
				case core.GestureStateBegan:
					if !grabbed {
						s.statusMutex.RLock()
						if s.mouseInput != nil {
							s.mouseInput.Grab()
							grabbed = true
						}
						s.statusMutex.RUnlock()
					}
					// log.Printf("%d本指ジェスチャー開始", gestureEvent.FingerCount)

				case core.GestureStateChanged:
					// ジェスチャー継続中

				case core.GestureStateEnded:
					if grabbed {
						s.statusMutex.RLock()
						if s.mouseInput != nil {
							s.mouseInput.Release()
						}
						s.statusMutex.RUnlock()
						grabbed = false
					}
					// log.Printf("%d本指ジェスチャー終了", gestureEvent.FingerCount)
					s.motionFilter.Reset()
				}

				// スクロールイベントを送信
				if err := s.scrollEmitter.SendGesture(gestureEvent); err != nil {
					log.Printf("スクロールイベントの送信に失敗しました: %v", err)
				}
			}

			time.Sleep(100 * time.Microsecond)
		}
	}
}