package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 修正版 4本指スワイプテスト ===")
	fmt.Println("期待されるイベントシーケンス：")
	fmt.Println("  30 sub=23 mask=0x1 phase=Began")
	fmt.Println("  29 sub=0")
	fmt.Println("  30 sub=23 mask=0x2 phase=Changed (複数回)")
	fmt.Println("  29 sub=0")
	fmt.Println("  30 sub=23 mask=0x4 phase=Ended")
	fmt.Println("  29 sub=0")
	fmt.Println("\n3秒後に開始します...")
	time.Sleep(3 * time.Second)

	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.01,  // 実際には使用されない
		ScrollScaleFactor:     2.0,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n=== 4本指タッチダウン ===")
	
	// 4本指タッチダウン（同時に）
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== スワイプ開始（右方向） ===")
	
	// 移動を5回実行
	for j := 1; j <= 5; j++ {
		fmt.Printf("  移動 %d/5\n", j)
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*20)+int32(j*50), 500)
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\n=== タッチアップ ===")
	
	// タッチアップ（同時に）
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}

	time.Sleep(500 * time.Millisecond)
	
	fmt.Println("\n完了！")
}