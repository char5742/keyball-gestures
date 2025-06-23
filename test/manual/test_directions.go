package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 4方向スワイプテスト ===")
	fmt.Println("各方向のデルタ値を確認します")
	fmt.Println("")

	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.5,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	// テスト用の関数
	testSwipe := func(direction string, dx, dy int32) {
		fmt.Printf("\n=== %sスワイプ ===\n", direction)
		
		// 4本指タッチダウン
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchDown(i, i+1, 500+int32(i*30), 500)
		}
		
		time.Sleep(100 * time.Millisecond)
		
		// 指定方向に移動
		for j := 1; j <= 5; j++ {
			for i := 0; i < 4; i++ {
				newX := 500 + int32(i*30) + dx*int32(j)
				newY := 500 + dy*int32(j)
				touchpad.MultiTouchMove(i, newX, newY)
			}
			time.Sleep(30 * time.Millisecond)
		}
		
		// タッチアップ
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchUp(i)
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("3秒後に開始します...")
	time.Sleep(3 * time.Second)

	// 右スワイプ（X+）
	testSwipe("右", 50, 0)
	
	// 左スワイプ（X-）
	testSwipe("左", -50, 0)
	
	// 上スワイプ（Y-）
	testSwipe("上", 0, -50)
	
	// 下スワイプ（Y+）
	testSwipe("下", 0, 50)

	fmt.Println("\n=== 結果確認 ===")
	fmt.Println("期待される出力:")
	fmt.Println("右: Δ(+X, 0)")
	fmt.Println("左: Δ(-X, 0)")
	fmt.Println("上: Δ(0, -Y)")
	fmt.Println("下: Δ(0, +Y)")
}