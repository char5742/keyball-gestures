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
	log.Println("Input monitoring test")
	log.Println("Press F13/F14 keys and move mouse. Press Ctrl+C to exit.")

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

	// メインループ
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	consecutiveNoInput := 0
	
	for {
		select {
		case <-sigCh:
			log.Println("Shutting down...")
			return
		case <-ticker.C:
			// キー入力チェック
			key := keyboard.GetKey()
			if key > 0 {
				fmt.Printf("Key pressed: %d (F13=%d, F14=%d)\n", key, 183, 184)
				consecutiveNoInput = 0
			}
			
			// マウス移動チェック
			dx, dy := mouse.GetMouseDelta()
			if dx != 0 || dy != 0 {
				fmt.Printf("Mouse moved: dx=%d, dy=%d\n", dx, dy)
				consecutiveNoInput = 0
			}
			
			// 入力がない場合
			if key <= 0 && dx == 0 && dy == 0 {
				consecutiveNoInput++
				if consecutiveNoInput == 40 { // 2秒ごと
					log.Println("Waiting for input...")
					consecutiveNoInput = 0
				}
			}
		}
	}
}