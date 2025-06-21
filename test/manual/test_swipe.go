package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	// 設定
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.01,
		ScrollScaleFactor:     2.0,
	}

	// タッチパッドを初期化
	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("4本指スワイプテストを開始します...")
	fmt.Println("trackpad_dumpでイベントを監視してください")
	time.Sleep(2 * time.Second)

	// 4本指タッチダウン
	fmt.Println("\n=== 4本指タッチダウン ===")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 100+int32(i*10), 100)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	// 右スワイプ
	fmt.Println("\n=== 右スワイプ開始 ===")
	for j := 0; j < 20; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 100+int32(i*10)+int32(j*15), 100)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// タッチアップ
	fmt.Println("\n=== タッチアップ ===")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\nテスト完了")
}