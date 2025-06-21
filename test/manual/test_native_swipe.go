package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== ネイティブ準拠テスト ===")
	fmt.Println("このテストでは、デルタ値を0にしてネイティブと完全に同じ動作を確認します")
	fmt.Println("\n3秒後に開始...")
	time.Sleep(3 * time.Second)

	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.0,  // デルタを0にする
		ScrollScaleFactor:     2.0,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n=== 4本指タッチダウン ===")
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== 右スワイプ ===")
	
	// 移動を3回実行
	for j := 1; j <= 3; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*20)+int32(j*100), 500)
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\n=== タッチアップ ===")
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}

	fmt.Println("\n完了！")
}