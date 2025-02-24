package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

const (
	// 仮想タッチパッドの範囲

	minX = 0
	maxX = 32767
	minY = 0
	maxY = 32767

	// マウスの移動量を調整する係数
	mouseDeltaFactor = 15

	// ジェスチャーをトリガーするためのキーコード
	// これらのキーを押しながらマウスを動かすと仮想タッチパッドが動きます

	twoFingerKey  = 184 // F14
	fourFingerKey = 183 // F13

	// モーションフィルター

	// 左右の移動を滑らかにするための係数
	// 0.0-1.0の範囲。1.0に近いほど滑らかになりますが、遅延が大きくなります
	motionFilterSmoothingFactor = 0.85
	// フィルターが動作し始めるまでのカウント
	// この値が大きいとフィルターが動作し始めるまでの時間が長くなります
	motionFilterWarmUpCount = 10

	// スクロールをリセットするまでの時間
	resetThreshold = 50 * time.Millisecond
)

const maxFingers = 4

func main() {
	handleSignals()

	// Create virtual touchpad device
	padDevice, err := features.CreateTouchPad("/dev/uinput", []byte("VirtualTouchPad"), minX, maxX, minY, maxY)
	if err != nil {
		panic(fmt.Errorf("仮想タッチパッドの作成に失敗しました: %w", err))
	}
	defer padDevice.Close()

	devices, err := features.GetDevices()

	if err != nil {
		panic(fmt.Errorf("デバイス一覧の取得に失敗しました: %w", err))
	}

	// 理想はデバイス一覧からターゲットを選択させたいが、
	// とりあえず最初のマウスとキーボードを使う
	var mouseDevice *features.Device
	var keyboardDevice *features.Device
	for _, device := range devices {
		if device.Type == features.DeviceTypeMouse {
			mouseDevice = &device

		} else if device.Type == features.DeviceTypeKeyboard {
			keyboardDevice = &device
		}
	}

	mouse, err := features.CreateMouse(mouseDevice.Path)
	if err != nil {
		panic(fmt.Errorf("対象のマウスに対してオブジェクトの生成に失敗しました[path=%s]: %w", mouseDevice.Path, err))
	}
	defer mouse.Close()

	keyboard, err := features.CreateKeyboard(keyboardDevice.Path)
	if err != nil {
		panic(fmt.Errorf("failed to create keyboard device: %w", err))
	}
	defer keyboard.Close()

	var (
		fingerCount     int
		fingerPositions [maxFingers]struct{ x, y int32 }
		prevKey         int32
		grabbed         bool
		lastScrollTime  time.Time
	)

	motionFilter := features.NewMotionFilter(motionFilterSmoothingFactor, motionFilterWarmUpCount)

	for {
		pressedKey := keyboard.GetKey()
		dxRaw, dyRaw := mouse.GetMouseDelta()
		dx, dy := motionFilter.Filter(dxRaw*mouseDeltaFactor, dyRaw*mouseDeltaFactor)
		// print finger positions 1
		fmt.Printf("Positions: x: %v, y: %v\n", fingerPositions[0].x, fingerPositions[0].y)

		now := time.Now()

		// 何も動いていない場合、最後のスクロールから閾値を超えていればリセット
		// これにより、タッチパッドの範囲内で無限にスクロールが可能
		if now.Sub(lastScrollTime) > resetThreshold && fingerCount > 0 {
			liftAllFingers(padDevice, fingerCount)
			motionFilter.Reset()
			initFingers(padDevice, &fingerPositions, fingerCount, maxX/2, maxY/2)
		}
		lastScrollTime = now

		switch {
		case pressedKey == twoFingerKey && fingerCount == 0:
			if !grabbed {
				mouse.Grab()
				grabbed = true
			}
			fmt.Println("2-finger operation triggered")
			fingerCount = 2
			initFingers(padDevice, &fingerPositions, fingerCount, maxX/2, maxY/2)
			prevKey = pressedKey

		case pressedKey == fourFingerKey && fingerCount == 0:
			if !grabbed {
				mouse.Grab()
				grabbed = true
			}
			fmt.Println("4-finger operation triggered")
			fingerCount = 4
			initFingers(padDevice, &fingerPositions, fingerCount, maxX/2, maxY/2)
			prevKey = pressedKey

		case (pressedKey == fourFingerKey || pressedKey == twoFingerKey) && fingerCount > 0:
			if pressedKey == prevKey {
				for i := 0; i < fingerCount; i++ {
					fingerPositions[i].x += dx
					fingerPositions[i].y += dy

					fingerPositions[i].x = clamp(fingerPositions[i].x, minX, maxX)
					fingerPositions[i].y = clamp(fingerPositions[i].y, minY, maxY)

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
		fmt.Println("Shutting down...")
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
