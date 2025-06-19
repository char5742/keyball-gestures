package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("=== Gesture Test ===")
	log.Println("F1キー: 2本指ジェスチャー（スクロール）")
	log.Println("F2キー: 4本指ジェスチャー（スワイプ）")
	log.Println("")

	// タッチパッドを作成
	touchpad, err := features.CreateTouchPad(features.TouchPadConfig{
		Name:                  "TestTouchPad",
		MinX:                  0,
		MaxX:                  10000,
		MinY:                  0,
		MaxY:                  10000,
		MotionSmoothingFactor: 0.0,
		MotionWarmUpCount:     0,
		MouseDeltaFactor:      50,
	})
	if err != nil {
		log.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	// キーボードを作成
	keyboard, err := features.CreateKeyboard("Keyball39")
	if err != nil {
		log.Fatalf("Failed to create keyboard: %v", err)
	}
	defer keyboard.Close()

	// マウスを作成
	mouse, err := features.CreateMouse("Keyball39")
	if err != nil {
		log.Fatalf("Failed to create mouse: %v", err)
	}
	defer mouse.Close()

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// ジェスチャー状態
	var (
		fingerCount     int
		fingerPositions [4]struct{ x, y int32 }
		prevKey         int32
		grabbed         bool
	)

	// メインループ
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	log.Println("準備完了！キーを押してテストしてください")

	for {
		select {
		case <-sigCh:
			log.Println("Shutting down...")
			return
		case <-ticker.C:
			// キー入力チェック
			key := keyboard.GetKey()

			// マウス移動チェック
			dx, dy := mouse.GetMouseDelta()

			switch {
			case key == 122 && fingerCount == 0: // F1 (2本指)
				if !grabbed {
					mouse.Grab()
					grabbed = true
				}
				log.Println("2本指ジェスチャー開始")
				fingerCount = 2
				// 指の初期位置を設定
				for i := 0; i < fingerCount; i++ {
					fingerPositions[i].x = 5000
					fingerPositions[i].y = int32(5000 + i*20)
					touchpad.MultiTouchDown(i, i, fingerPositions[i].x, fingerPositions[i].y)
				}
				prevKey = key

			case key == 120 && fingerCount == 0: // F2 (4本指)
				if !grabbed {
					mouse.Grab()
					grabbed = true
				}
				log.Println("4本指ジェスチャー開始")
				fingerCount = 4
				// 指の初期位置を設定
				for i := 0; i < fingerCount; i++ {
					fingerPositions[i].x = 5000
					fingerPositions[i].y = int32(5000 + i*20)
					touchpad.MultiTouchDown(i, i, fingerPositions[i].x, fingerPositions[i].y)
				}
				prevKey = key

			case (key == 122 || key == 120) && fingerCount > 0:
				if key == prevKey && (dx != 0 || dy != 0) {
					// 指を移動
					for i := 0; i < fingerCount; i++ {
						fingerPositions[i].x += dx
						fingerPositions[i].y += dy
						touchpad.MultiTouchMove(i, fingerPositions[i].x, fingerPositions[i].y)
					}
					
					if fingerCount == 2 {
						fmt.Printf("2本指移動: dx=%d, dy=%d\n", dx, dy)
					} else if fingerCount == 4 {
						fmt.Printf("4本指移動: dx=%d, dy=%d\n", dx, dy)
					}
				}
				prevKey = key

			default:
				if grabbed {
					mouse.Release()
					grabbed = false
				}
				if fingerCount > 0 {
					// すべての指を持ち上げる
					for i := 0; i < fingerCount; i++ {
						touchpad.MultiTouchUp(i)
					}
					log.Printf("%d本指ジェスチャー終了\n", fingerCount)
					fingerCount = 0
				}
				prevKey = 0
			}
		}
	}
}