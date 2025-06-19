package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("macOS touchpad test program")

	// タッチパッドを作成
	touchpad, err := features.CreateTouchPad(features.TouchPadConfig{
		Name:                  "TestTouchPad",
		MinX:                  0,
		MaxX:                  10000,
		MinY:                  0,
		MaxY:                  10000,
		MotionSmoothingFactor: 0.85,
		MotionWarmUpCount:     10,
		MouseDeltaFactor:      10,
	})
	if err != nil {
		log.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	log.Println("Touchpad created successfully")

	// テストシナリオ: 2本指スクロール
	log.Println("Starting 2-finger scroll test in 3 seconds...")
	time.Sleep(3 * time.Second)

	// 2本の指をタッチダウン
	log.Println("Touch down: 2 fingers")
	touchpad.MultiTouchDown(0, 1000, 5000, 5000)
	touchpad.MultiTouchDown(1, 1001, 5100, 5000)

	// スクロール動作をシミュレート
	log.Println("Simulating scroll...")
	for i := 0; i < 20; i++ {
		y := int32(5000 - i*50) // 上方向にスクロール
		touchpad.MultiTouchMove(0, 5000, y)
		touchpad.MultiTouchMove(1, 5100, y)
		time.Sleep(20 * time.Millisecond)
	}

	// タッチアップ
	log.Println("Touch up")
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)

	log.Println("Test completed")
	
	// キーボードとマウスのテスト
	log.Println("\nTesting keyboard and mouse...")
	
	keyboard, err := features.CreateKeyboard("test")
	if err != nil {
		log.Printf("Failed to create keyboard: %v", err)
	} else {
		defer keyboard.Close()
		
		log.Println("Press F13 or F14 key...")
		for i := 0; i < 50; i++ {
			key := keyboard.GetKey()
			if key > 0 {
				fmt.Printf("Key pressed: %d\n", key)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	mouse, err := features.CreateMouse("test")
	if err != nil {
		log.Printf("Failed to create mouse: %v", err)
	} else {
		defer mouse.Close()
		
		log.Println("Move mouse to see delta values...")
		for i := 0; i < 50; i++ {
			dx, dy := mouse.GetMouseDelta()
			if dx != 0 || dy != 0 {
				fmt.Printf("Mouse delta: dx=%d, dy=%d\n", dx, dy)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}