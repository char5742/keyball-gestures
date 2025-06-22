package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 実装とネイティブの差異分析 ===")
	fmt.Println("\nネイティブログの特徴:")
	fmt.Println("1. type=30の後に必ずtype=29が来る")
	fmt.Println("2. 最初のChangedのflagsは0x8")
	fmt.Println("3. Δ(0.0,0.0)と常に表示される")
	fmt.Println("4. sub=0のtype=29も含まれる")
	
	fmt.Println("\n現在の実装の問題の可能性:")
	fmt.Println("1. イベントの送信タイミング？")
	fmt.Println("2. 追加のフィールドが必要？")
	fmt.Println("3. イベントソースの設定？")
	
	fmt.Println("\n=== デバッグ実行 ===")
	
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

	// 最小限のスワイプを実行
	fmt.Println("\n最小限の4本指スワイプを実行...")
	
	// タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// 1回だけ大きく移動
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchMove(i, 800+int32(i*20), 500)  // 300ピクセル移動
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\n完了")
	
	fmt.Println("\n=== 考えられる原因 ===")
	fmt.Println("1. CGEventSourceの設定が不適切")
	fmt.Println("2. イベントの送信先（kCGSessionEventTap）が間違っている")
	fmt.Println("3. 必須フィールドが欠けている")
	fmt.Println("4. macOSのセキュリティ機能が影響している")
}