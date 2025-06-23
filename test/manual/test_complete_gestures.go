package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 完全なジェスチャー検証テスト ===")
	fmt.Println("2本指スクロールと4本指スワイプの両方をテストします")
	fmt.Println("")
	
	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     2,
		ScrollScaleFactor:     0,    // デフォルト値（0.3）
		SwipeScaleFactor:      0.5,  // 4本指スワイプ用
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("trackpad_dumpを別ターミナルで実行してイベントを監視してください")
	fmt.Println("cd test/debugging && ./trackpad_dump")
	fmt.Println("")

	// 2本指スクロールテスト
	fmt.Println("=== 2本指スクロールテスト ===")
	fmt.Println("3秒後に開始...")
	time.Sleep(3 * time.Second)
	
	// 2本指タッチダウン
	touchpad.MultiTouchDown(0, 1, 600, 600)
	touchpad.MultiTouchDown(1, 2, 650, 600)
	
	time.Sleep(100 * time.Millisecond)
	
	// 上方向にスクロール（小さな増分）
	for i := 1; i <= 10; i++ {
		touchpad.MultiTouchMove(0, 600, 600-int32(i*3))
		touchpad.MultiTouchMove(1, 650, 600-int32(i*3))
		time.Sleep(20 * time.Millisecond)
	}
	
	// タッチアップ
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	
	fmt.Println("2本指スクロール完了")
	time.Sleep(2 * time.Second)
	
	// 4本指スワイプテスト
	fmt.Println("\n=== 4本指スワイプテスト（上方向 - Mission Control） ===")
	fmt.Println("3秒後に開始...")
	time.Sleep(3 * time.Second)
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 600+int32(i*50), 600)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// 上方向に大きくスワイプ
	for j := 1; j <= 10; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 600+int32(i*50), 600-int32(j*30))
		}
		time.Sleep(20 * time.Millisecond)
	}
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("4本指スワイプ完了")
	time.Sleep(2 * time.Second)
	
	fmt.Println("\n=== テスト完了 ===")
	fmt.Println("ログを確認してください:")
	fmt.Println("- 2本指スクロール: 小さなデルタ値、適切なフラグ")
	fmt.Println("- 4本指スワイプ: type=30、sub=23、適切なデルタ値")
}