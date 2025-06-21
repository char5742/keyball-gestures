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
	fmt.Println("=== Debug Swipe Test ===")
	fmt.Println("NSLogの出力を確認してください（Console.appまたはlog stream）")
	
	cfg := features.TouchPadConfig{
		Name:                  "debug-swipe",
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

	fmt.Println("\n上方向スワイプをテスト（mask=0x4が表示される問題の再現）")
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		err = touchpad.MultiTouchDown(i, 1000+i, int32(1000+i*200), 1000)
		if err != nil {
			fmt.Printf("Error slot %d: %v\n", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// 上方向に移動（Y座標を減少）
	fmt.Println("上方向に移動中...")
	for j := 0; j < 10; j++ {
		for i := 0; i < 4; i++ {
			// Y座標を減少させて上方向に移動
			err = touchpad.MultiTouchMove(i, int32(1000+i*200), int32(1000-50*(j+1)))
			if err != nil {
				fmt.Printf("Move error: %v\n", err)
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\nテスト完了。trackpad_dumpとNSLogの出力を確認してください。")
	fmt.Println("NSLogを見るには: log stream --predicate 'subsystem contains \"keyball\"'")
}