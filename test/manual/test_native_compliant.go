package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== ネイティブ完全準拠テスト ===")
	fmt.Println("期待される出力（ネイティブログと同じ）：")
	fmt.Println("  Gesture type=30 sub=23 phase=Began mask=0x1 flags=0x0")
	fmt.Println("          Δ(0.0,0.0)")
	fmt.Println("  Gesture type=29 sub=0")
	fmt.Println("  Gesture type=30 sub=23 phase=Changed mask=0x2 flags=0x8  ← 最初のChanged")
	fmt.Println("          Δ(0.0,0.0)")
	fmt.Println("  Gesture type=29 sub=0")
	fmt.Println("  Gesture type=30 sub=23 phase=Changed mask=0x2 flags=0x0  ← 2回目以降")
	fmt.Println("          Δ(0.0,0.0)")
	
	time.Sleep(2 * time.Second)

	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n=== 開始 ===")
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
	}

	time.Sleep(100 * time.Millisecond)

	// 3回移動
	for j := 1; j <= 3; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*20)+int32(j*50), 500)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}

	fmt.Println("\n完了！")
}