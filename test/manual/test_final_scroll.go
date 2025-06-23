package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 最終スクロールテスト ===")
	fmt.Println("trackpad_dumpで確認してください: cd test/debugging && ./trackpad_dump")
	fmt.Println("")
	
	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.3,  // より軽いスムージング
		MotionWarmUpCount:     2,
		ScrollScaleFactor:     1.0,   // 1:1スケール
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("3秒後に開始...")
	time.Sleep(3 * time.Second)

	// 2本指タッチダウン
	touchpad.MultiTouchDown(0, 1, 600, 600)
	time.Sleep(10 * time.Millisecond)
	touchpad.MultiTouchDown(1, 2, 650, 600)
	
	// 初期待機（MayBegin→Began）
	time.Sleep(50 * time.Millisecond)
	
	// ネイティブパターンでスクロール
	y0, y1 := int32(600), int32(600)
	
	// 小さな動きから始めて加速
	movements := []int32{-1, -1, -2, -3, -3, -2, -2, -1, -1}
	
	for i, dy := range movements {
		y0 += dy
		y1 += dy
		touchpad.MultiTouchMove(0, 600, y0)
		touchpad.MultiTouchMove(1, 650, y1)
		fmt.Printf("Move %d: Δ(0,%d)\n", i+1, dy)
		time.Sleep(17 * time.Millisecond)
	}
	
	// 終了
	time.Sleep(20 * time.Millisecond)
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	
	fmt.Println("\n完了！ログを確認してください")
}