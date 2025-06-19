package main

import (
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("macOS scroll test")
	
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

	log.Println("Touchpad created. Starting scroll in 3 seconds...")
	time.Sleep(3 * time.Second)

	// 2本指でタッチダウン
	log.Println("Touch down: 2 fingers")
	err = touchpad.MultiTouchDown(0, 1000, 5000, 5000)
	if err != nil {
		log.Printf("Error touch down 0: %v", err)
	}
	
	err = touchpad.MultiTouchDown(1, 1001, 5100, 5000)
	if err != nil {
		log.Printf("Error touch down 1: %v", err)
	}

	// 上方向にスクロール
	log.Println("Scrolling up...")
	for i := 0; i < 30; i++ {
		y := int32(5000 - i*100) // 上方向に移動
		
		err = touchpad.MultiTouchMove(0, 5000, y)
		if err != nil {
			log.Printf("Error move 0: %v", err)
		}
		
		err = touchpad.MultiTouchMove(1, 5100, y)
		if err != nil {
			log.Printf("Error move 1: %v", err)
		}
		
		time.Sleep(30 * time.Millisecond)
	}

	// タッチアップ
	log.Println("Touch up")
	err = touchpad.MultiTouchUp(0)
	if err != nil {
		log.Printf("Error touch up 0: %v", err)
	}
	
	err = touchpad.MultiTouchUp(1)
	if err != nil {
		log.Printf("Error touch up 1: %v", err)
	}

	log.Println("Scroll test completed")
}