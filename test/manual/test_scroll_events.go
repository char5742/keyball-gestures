package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== スクロールイベント検証テスト ===")
	fmt.Println("2本指スクロールのイベントパターンを確認します")
	fmt.Println("")
	
	// タッチパッドを初期化（デフォルトのスケールファクターを使用）
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     2,
		ScrollScaleFactor:     0,  // デフォルト値（0.3）を使用
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("trackpad_dumpを別ターミナルで実行してイベントを監視してください")
	fmt.Println("3秒後に2本指スクロールを開始します...")
	time.Sleep(3 * time.Second)

	// 2本指タッチダウン
	fmt.Println("\n=== 2本指タッチダウン ===")
	touchpad.MultiTouchDown(0, 1, 500, 500)
	touchpad.MultiTouchDown(1, 2, 550, 500)
	
	time.Sleep(100 * time.Millisecond)
	
	// 小さな増分で移動（ネイティブと同じパターン）
	fmt.Println("\n=== スクロール移動開始 ===")
	movements := []struct {
		dx, dy int32
		delay  time.Duration
	}{
		{0, -1, 20 * time.Millisecond},   // 小さな開始
		{0, -3, 20 * time.Millisecond},   // 徐々に加速
		{1, -9, 20 * time.Millisecond},   // より大きな動き
		{1, -9, 20 * time.Millisecond},   // 継続
		{1, -7, 20 * time.Millisecond},   // 減速開始
		{1, -6, 20 * time.Millisecond},
		{1, -5, 20 * time.Millisecond},
		{1, -3, 20 * time.Millisecond},
		{1, -1, 20 * time.Millisecond},   // 停止に向けて
		{0, -1, 20 * time.Millisecond},   // 最後の小さな動き
	}
	
	baseX0, baseY0 := int32(500), int32(500)
	baseX1, baseY1 := int32(550), int32(500)
	
	for i, move := range movements {
		baseY0 += move.dy
		baseY1 += move.dy
		baseX0 += move.dx
		baseX1 += move.dx
		
		touchpad.MultiTouchMove(0, baseX0, baseY0)
		touchpad.MultiTouchMove(1, baseX1, baseY1)
		
		fmt.Printf("移動 %d: Δ(%d,%d)\n", i+1, move.dx, move.dy)
		time.Sleep(move.delay)
	}
	
	// タッチアップ
	fmt.Println("\n=== タッチアップ ===")
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	
	fmt.Println("\n=== 結果確認 ===")
	fmt.Println("期待される結果:")
	fmt.Println("- デルタ値が小さな値（-1〜-9程度）")
	fmt.Println("- フラグが0x80000000, 0x40000000, 0x3F800000のいずれか")
	fmt.Println("- type=29とScrollイベントが交互に出現")
	fmt.Println("- Gesture sub=6(Scroll)と正しいフェーズ")
}