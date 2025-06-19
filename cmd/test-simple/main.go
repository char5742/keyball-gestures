package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("Simple macOS test")

	// キーボードテスト
	log.Println("Testing keyboard...")
	keyboard, err := features.CreateKeyboard("Keyball39")
	if err != nil {
		log.Fatalf("Failed to create keyboard: %v", err)
	}
	defer keyboard.Close()

	log.Println("Keyboard created. Press F13 or F14...")
	
	// マウステスト
	log.Println("Testing mouse...")
	mouse, err := features.CreateMouse("Keyball39")
	if err != nil {
		log.Fatalf("Failed to create mouse: %v", err)
	}
	defer mouse.Close()

	log.Println("Mouse created.")

	// 10秒間監視
	for i := 0; i < 100; i++ {
		key := keyboard.GetKey()
		if key > 0 {
			fmt.Printf("Key pressed: %d\n", key)
		}
		
		dx, dy := mouse.GetMouseDelta()
		if dx != 0 || dy != 0 {
			fmt.Printf("Mouse moved: dx=%d, dy=%d\n", dx, dy)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	log.Println("Test completed")
}