package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== macOSシステムジェスチャーテスト ===")
	fmt.Println("")
	fmt.Println("システム環境設定 > トラックパッド > その他のジェスチャ で以下を確認:")
	fmt.Println("- Mission Control: 4本指で上にスワイプ")
	fmt.Println("- アプリケーションExposé: 4本指で下にスワイプ")
	fmt.Println("- フルスクリーンアプリケーション間をスワイプ: 4本指で左右にスワイプ")
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

	// ジェスチャー実行関数
	performGesture := func(name string, dx, dy int32) {
		fmt.Printf("\n=== %s ===\n", name)
		fmt.Println("3秒後に実行...")
		time.Sleep(3 * time.Second)
		
		// 4本指タッチダウン
		centerX := int32(600)
		centerY := int32(400)
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchDown(i, i+1, centerX+int32(i*40), centerY)
		}
		
		time.Sleep(50 * time.Millisecond)
		
		// ジェスチャー実行（大きめの移動量）
		steps := 10
		for j := 1; j <= steps; j++ {
			for i := 0; i < 4; i++ {
				newX := centerX + int32(i*40) + dx*int32(j)*3
				newY := centerY + dy*int32(j)*3
				touchpad.MultiTouchMove(i, newX, newY)
			}
			time.Sleep(20 * time.Millisecond)
		}
		
		// タッチアップ
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchUp(i)
		}
		
		fmt.Println("完了！動作を確認してください")
		time.Sleep(2 * time.Second)
	}

	// 各ジェスチャーをテスト
	performGesture("Mission Control（上スワイプ）", 0, -20)
	performGesture("アプリケーションExposé（下スワイプ）", 0, 20)
	performGesture("左のスペースへ移動（右スワイプ）", 20, 0)
	performGesture("右のスペースへ移動（左スワイプ）", -20, 0)

	fmt.Println("\n=== テスト完了 ===")
	fmt.Println("すべてのジェスチャーが正しく動作しましたか？")
}