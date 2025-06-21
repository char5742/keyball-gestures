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
	fmt.Println("=== Timing Test ===")
	fmt.Println("gestureDetectionWindow(50ms)経過後にジェスチャーが開始されるかテスト")
	
	cfg := features.TouchPadConfig{
		Name:                  "timing-test",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MotionWarmUpCount:     2,
		MotionSmoothingFactor: 0.3,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n1. 4本指タッチダウン（高速）")
	for i := 0; i < 4; i++ {
		err = touchpad.MultiTouchDown(i, 2000+i, int32(1000+i*200), 1000)
		if err != nil {
			fmt.Printf("Error slot %d: %v\n", i, err)
		}
		time.Sleep(5 * time.Millisecond) // 合計20ms < 50ms
	}
	
	fmt.Println("\n2. 60ms待機（gestureDetectionWindow超過）")
	time.Sleep(60 * time.Millisecond)
	
	fmt.Println("\n3. 追加のタッチダウンでジェスチャー判定をトリガー")
	err = touchpad.MultiTouchDown(4, 2004, 1800, 1000)
	if err == nil {
		fmt.Println("5本目のタッチダウン成功")
		touchpad.MultiTouchUp(4)
	}
	
	fmt.Println("\n4. 移動開始")
	for j := 0; j < 3; j++ {
		for i := 0; i < 4; i++ {
			err = touchpad.MultiTouchMove(i, int32(1000+i*200+100*(j+1)), 1000)
			if err != nil {
				fmt.Printf("Move error slot %d: %v\n", i, err)
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	
	fmt.Println("\n5. タッチアップ")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\nテスト完了。trackpad_dumpでジェスチャーが検出されたか確認してください。")
}