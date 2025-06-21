//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== Simple 4-finger Test ===")
	
	cfg := features.TouchPadConfig{
		Name:                  "simple-test",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		MotionSmoothingFactor: 0.3,
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		log.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	// 4本指を同時にタッチダウン
	fmt.Println("Testing 4-finger touch...")
	
	// slot 0
	err = touchpad.MultiTouchDown(0, 1001, 1000, 1000)
	if err != nil {
		fmt.Printf("Error slot 0: %v\n", err)
	} else {
		fmt.Println("Slot 0: OK")
	}
	
	// slot 1
	err = touchpad.MultiTouchDown(1, 1002, 1200, 1000)
	if err != nil {
		fmt.Printf("Error slot 1: %v\n", err)
	} else {
		fmt.Println("Slot 1: OK")
	}
	
	// slot 2
	err = touchpad.MultiTouchDown(2, 1003, 1400, 1000)
	if err != nil {
		fmt.Printf("Error slot 2: %v\n", err)
	} else {
		fmt.Println("Slot 2: OK")
	}
	
	// slot 3
	err = touchpad.MultiTouchDown(3, 1004, 1600, 1000)
	if err != nil {
		fmt.Printf("Error slot 3: %v\n", err)
	} else {
		fmt.Println("Slot 3: OK")
	}

	// 待機
	time.Sleep(200 * time.Millisecond)

	// 移動
	fmt.Println("\nMoving fingers...")
	for i := 0; i < 5; i++ {
		dx := int32(100)
		touchpad.MultiTouchMove(0, 1000+dx*(int32(i)+1), 1000)
		touchpad.MultiTouchMove(1, 1200+dx*(int32(i)+1), 1000)
		touchpad.MultiTouchMove(2, 1400+dx*(int32(i)+1), 1000)
		touchpad.MultiTouchMove(3, 1600+dx*(int32(i)+1), 1000)
		time.Sleep(50 * time.Millisecond)
	}

	// タッチアップ
	fmt.Println("\nReleasing fingers...")
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	touchpad.MultiTouchUp(2)
	touchpad.MultiTouchUp(3)

	fmt.Println("\nDone. Check trackpad_dump output.")
}