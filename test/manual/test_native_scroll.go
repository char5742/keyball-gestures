package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== ネイティブ準拠スクロールテスト ===")
	fmt.Println("2本指スクロールをネイティブトラックパッドと同じパターンで実行します")
	fmt.Println("")
	
	// タッチパッドを初期化（デフォルトのスケールファクターを使用）
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     2,
		ScrollScaleFactor:     0,  // デフォルト値（0.05）を使用
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("trackpad_dumpを別ターミナルで実行してイベントを監視してください")
	fmt.Println("cd test/debugging && ./trackpad_dump")
	fmt.Println("")
	fmt.Println("3秒後に2本指スクロールを開始します...")
	time.Sleep(3 * time.Second)

	// 2本指タッチダウン
	fmt.Println("\n=== 2本指タッチダウン ===")
	touchpad.MultiTouchDown(0, 1, 600, 600)
	touchpad.MultiTouchDown(1, 2, 650, 600)
	
	// ジェスチャー開始の待機
	time.Sleep(60 * time.Millisecond)
	
	// ネイティブログを再現する動きのパターン
	// 実際の指の動きはもっと小さい（1ピクセルずつ）
	fmt.Println("\n=== スクロール移動開始（ネイティブパターン） ===")
	movements := []struct {
		dx, dy int32
		delay  time.Duration
	}{
		// MayBegin後の最初の動き
		{0, -1, 54 * time.Millisecond},   // 最初は小さく
		{0, -1, 12 * time.Millisecond},   // 徐々に加速
		{0, -2, 16 * time.Millisecond},   
		{0, -3, 17 * time.Millisecond},   
		{0, -3, 17 * time.Millisecond},   
		{0, -2, 17 * time.Millisecond},   // 減速開始
		{0, -2, 17 * time.Millisecond},
		{0, -2, 17 * time.Millisecond},
		{0, -1, 17 * time.Millisecond},   
		{0, -1, 8 * time.Millisecond},    // 停止前
		{0, 0, 9 * time.Millisecond},     // 停止
		// 数回の空イベント
		{0, 0, 8 * time.Millisecond},
		{0, 0, 8 * time.Millisecond},
		{0, 0, 8 * time.Millisecond},
		{0, 0, 8 * time.Millisecond},
		{0, 0, 8 * time.Millisecond},
		{0, 0, 16 * time.Millisecond},
	}
	
	baseX0, baseY0 := int32(600), int32(600)
	baseX1, baseY1 := int32(650), int32(600)
	
	for i, move := range movements {
		if move.dx != 0 || move.dy != 0 {
			baseY0 += move.dy
			baseY1 += move.dy
			baseX0 += move.dx
			baseX1 += move.dx
			
			touchpad.MultiTouchMove(0, baseX0, baseY0)
			touchpad.MultiTouchMove(1, baseX1, baseY1)
			
			fmt.Printf("移動 %d: Δ(%d,%d) - 待機 %dms\n", i+1, move.dx, move.dy, move.delay/time.Millisecond)
		}
		time.Sleep(move.delay)
	}
	
	// タッチアップ（Endedイベント）
	fmt.Println("\n=== タッチアップ（Ended） ===")
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	
	fmt.Println("\n=== 期待される結果 ===")
	fmt.Println("- Scrollイベントのデルタ値: -1.0〜-3.0程度")
	fmt.Println("- Gestureイベントのデルタ値: 同じく小さな値")
	fmt.Println("- フラグ: 0x80000000（Began/Ended）, 0x40000000/0x3F800000（Changed）")
	fmt.Println("- 各イベント前にtype=29 sub=0のダミーイベント")
	fmt.Println("- イベント間隔: 12-17ms（ネイティブと同等）")
}