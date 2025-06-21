package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 4本指スワイプテスト ===")
	fmt.Println("別のターミナルで以下を実行してください：")
	fmt.Println("./test/debugging/trackpad_dump")
	fmt.Println("\n5秒後に開始します...")
	time.Sleep(5 * time.Second)

	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.01,
		ScrollScaleFactor:     2.0,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n開始: 4本指タッチダウン")
	
	// 4本指タッチダウン（ゆっくりと）
	for i := 0; i < 4; i++ {
		fmt.Printf("  指 %d をタッチダウン\n", i+1)
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\nすべての指がタッチダウンしました")
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n右スワイプを開始...")
	
	// 右スワイプ（ゆっくりと）
	for j := 1; j <= 10; j++ {
		fmt.Printf("  移動 %d/10\n", j)
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*20)+int32(j*30), 500)
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\nスワイプ完了")
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\nタッチアップ...")
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		fmt.Printf("  指 %d をタッチアップ\n", i+1)
		touchpad.MultiTouchUp(i)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\nテスト完了！")
	fmt.Println("trackpad_dumpの出力を確認してください")
}