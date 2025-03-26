package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/features"
	"github.com/char5742/keyball-gestures/internal/gui"
)

const maxFingers = 4

func main() {
	// コマンドライン引数の解析
	useGui := flag.Bool("gui", false, "GUIモードで起動します")
	configPath := flag.String("config", "", "設定ファイルのパス (指定しない場合はデフォルトパスを使用)")
	flag.Parse()

	// デフォルト設定ファイルパスの設定
	defaultConfigPath := ""
	configDir, err := os.UserConfigDir()
	if err == nil {
		defaultConfigPath = filepath.Join(configDir, "keyball-gestures", "config.toml")
	}

	// 設定ファイルパスの決定
	cfgPath := defaultConfigPath
	if *configPath != "" {
		cfgPath = *configPath
	}

	// 設定ファイルの読み込み
	var cfg *config.Config
	if cfgPath != "" {
		cfg, err = config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("設定ファイルの読み込みに失敗しました: %v\nデフォルト設定を使用します\n", err)
			cfg = config.DefaultConfig()
		} else {
			fmt.Printf("設定ファイルを読み込みました: %s\n", cfgPath)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// シグナルハンドラの設定
	handleSignals()

	// GUIモードかCLIモードかを判断
	if *useGui {
		// GUIモードで実行
		fmt.Println("GUIモードで起動します...")
		gui.RunGUI()
	} else {
		// CLIモードで実行
		fmt.Println("CLIモードで起動します...")
		runCLI(cfg)
	}
}

// CLIモードでの実行
func runCLI(cfg *config.Config) {
	// 仮想タッチパッドデバイスの作成
	padDevice, err := features.CreateTouchPad("/dev/uinput", []byte("VirtualTouchPad"),
		cfg.TouchPad.MinX, cfg.TouchPad.MaxX, cfg.TouchPad.MinY, cfg.TouchPad.MaxY)
	if err != nil {
		fmt.Printf("仮想タッチパッドの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	defer padDevice.Close()

	// デバイス一覧の取得
	devices, err := features.GetDevices()
	if err != nil {
		fmt.Printf("デバイス一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 設定ファイルで指定された優先デバイスまたは最初のマウスとキーボードを使用
	var mouseDevice *features.Device
	var keyboardDevice *features.Device

	// 優先デバイスの名前
	preferredKeyboard := cfg.DevicePrefs.PreferredKeyboardDevice
	preferredMouse := cfg.DevicePrefs.PreferredMouseDevice

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
		fmt.Println("エラー: マウスデバイスが見つかりませんでした")
		os.Exit(1)
	}
	if keyboardDevice == nil {
		fmt.Println("エラー: キーボードデバイスが見つかりませんでした")
		os.Exit(1)
	}

	fmt.Printf("使用するキーボード: %s\n", keyboardDevice.Name)
	fmt.Printf("使用するマウス: %s\n", mouseDevice.Name)

	// マウスとキーボードデバイスを開く
	mouse, err := features.CreateMouse(mouseDevice.Path)
	if err != nil {
		fmt.Printf("マウスデバイスのオープンに失敗しました[path=%s]: %v\n", mouseDevice.Path, err)
		os.Exit(1)
	}
	defer mouse.Close()

	keyboard, err := features.CreateKeyboard(keyboardDevice.Path)
	if err != nil {
		fmt.Printf("キーボードデバイスのオープンに失敗しました: %v\n", err)
		os.Exit(1)
	}
	defer keyboard.Close()

	// ジェスチャー認識のメインループ
	runGestureLoop(padDevice, keyboard, mouse, cfg)
}

// ジェスチャー認識のメインループ
func runGestureLoop(padDevice features.TouchPad, keyboard features.Keyboard, mouse features.Mouse, cfg *config.Config) {
	var (
		fingerCount     int
		fingerPositions [maxFingers]struct{ x, y int32 }
		prevKey         int32
		grabbed         bool
		lastScrollTime  time.Time
	)

	motionFilter := features.NewMotionFilter(cfg.Motion.FilterSmoothingFactor, cfg.Motion.FilterWarmUpCount)

	fmt.Println("ジェスチャー認識を開始しました...")

	for {
		pressedKey := keyboard.GetKey()
		dxRaw, dyRaw := mouse.GetMouseDelta()
		dx, dy := motionFilter.Filter(dxRaw*int32(cfg.Motion.MouseDeltaFactor), dyRaw*int32(cfg.Motion.MouseDeltaFactor))

		now := time.Now()

		// 何も動いていない場合、最後のスクロールから閾値を超えていればリセット
		// これにより、タッチパッドの範囲内で無限にスクロールが可能
		if now.Sub(lastScrollTime) > cfg.Gesture.ResetThreshold && fingerCount > 0 {
			liftAllFingers(padDevice, fingerCount)
			motionFilter.Reset()
			initFingers(padDevice, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
		}
		lastScrollTime = now

		switch {
		case pressedKey == int32(cfg.Input.TwoFingerKey) && fingerCount == 0:
			if !grabbed {
				mouse.Grab()
				grabbed = true
			}
			fmt.Println("2本指ジェスチャー開始")
			fingerCount = 2
			initFingers(padDevice, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
			prevKey = pressedKey

		case pressedKey == int32(cfg.Input.FourFingerKey) && fingerCount == 0:
			if !grabbed {
				mouse.Grab()
				grabbed = true
			}
			fmt.Println("4本指ジェスチャー開始")
			fingerCount = 4
			initFingers(padDevice, &fingerPositions, fingerCount, cfg.TouchPad.MaxX/2, cfg.TouchPad.MaxY/2)
			prevKey = pressedKey

		case (pressedKey == int32(cfg.Input.FourFingerKey) || pressedKey == int32(cfg.Input.TwoFingerKey)) && fingerCount > 0:
			if pressedKey == prevKey {
				for i := 0; i < fingerCount; i++ {
					fingerPositions[i].x += dx
					fingerPositions[i].y += dy

					fingerPositions[i].x = clamp(fingerPositions[i].x, cfg.TouchPad.MinX, cfg.TouchPad.MaxX)
					fingerPositions[i].y = clamp(fingerPositions[i].y, cfg.TouchPad.MinY, cfg.TouchPad.MaxY)

					_ = padDevice.MultiTouchMove(i, fingerPositions[i].x, fingerPositions[i].y)
				}
			} else {
				liftAllFingers(padDevice, fingerCount)
				motionFilter.Reset()
				fingerCount = 0
			}
			prevKey = pressedKey

		default:
			if grabbed {
				mouse.Release()
				grabbed = false
			}
			if fingerCount > 0 {
				liftAllFingers(padDevice, fingerCount)
				fmt.Println("ジェスチャー終了")
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

func initFingers(padDevice features.TouchPad, fingerPositions *[maxFingers]struct{ x, y int32 }, count int, centerX, centerY int32) {
	offset := int32(20)
	startY := centerY - offset*(int32(count)-1)/2

	for i := 0; i < count; i++ {
		fingerPositions[i].x = centerX
		fingerPositions[i].y = startY + offset*int32(i)

		_ = padDevice.MultiTouchDown(i, i, fingerPositions[i].x, fingerPositions[i].y)
	}
}

func liftAllFingers(padDevice features.TouchPad, count int) {
	for i := 0; i < count; i++ {
		_ = padDevice.MultiTouchUp(i)
	}
}

func handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("シャットダウンします...")
		os.Exit(0)
	}()
}

func clamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
