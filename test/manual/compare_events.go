package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== ネイティブとの詳細比較 ===")
	fmt.Println("\nネイティブログ（test/debugging/nativelog）より:")
	fmt.Println("[39144.087] Gesture type=30(Change) sub=23(Swipe) phase=Began mask=0x1 [Began ] flags=0x0")
	fmt.Println("          Δ(0.0,0.0)")
	fmt.Println("[39144.105] Gesture type=29(Begin) sub=0(?) phase=? mask=0 flags=0x0")
	fmt.Println("[39144.105] Gesture type=30(Change) sub=23(Swipe) phase=Changed mask=0x2 [Changed ] flags=0x8")
	fmt.Println("          Δ(0.0,0.0)")
	fmt.Println("[39144.118] Gesture type=29(Begin) sub=0(?) phase=? mask=0 flags=0x0")
	fmt.Println("[39144.118] Gesture type=30(Change) sub=23(Swipe) phase=Changed mask=0x2 [Changed ] flags=0x0")
	fmt.Println("          Δ(0.0,0.0)")
	
	fmt.Println("\n重要なポイント:")
	fmt.Println("1. type=30の後に必ずtype=29が来る")
	fmt.Println("2. 最初のChangedはflags=0x8")
	fmt.Println("3. デルタは常に0.0")
	fmt.Println("4. フェーズはBegan(1)→Changed(2)→Ended(4)")
	
	fmt.Println("\n\n別のターミナルでtrackpad_dumpを実行してください:")
	fmt.Println("cd test/debugging && ./trackpad_dump")
	fmt.Println("\nEnterキーを押すとテストを開始します...")
	fmt.Scanln()

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

	fmt.Println("\n=== 最小限のテスト ===")
	
	// 4本指タッチダウン（一度に）
	fmt.Println("4本指タッチダウン...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
	}
	
	time.Sleep(200 * time.Millisecond)
	
	// 1回だけ移動
	fmt.Println("1回移動...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchMove(i, 550+int32(i*20), 500)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// もう1回移動
	fmt.Println("もう1回移動...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchMove(i, 600+int32(i*20), 500)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// タッチアップ
	fmt.Println("タッチアップ...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\n完了！trackpad_dumpの出力を確認してください")
	fmt.Println("\n期待される出力:")
	fmt.Println("1. Gesture type=30(Change) sub=23(Swipe) phase=Began")
	fmt.Println("2. Gesture type=29(Begin) sub=0  ← 重要：type=30の直後")
	fmt.Println("3. Gesture type=30(Change) sub=23(Swipe) phase=Changed flags=0x8")
	fmt.Println("4. Gesture type=29(Begin) sub=0")
	fmt.Println("5. Gesture type=30(Change) sub=23(Swipe) phase=Changed flags=0x0")
	fmt.Println("6. Gesture type=29(Begin) sub=0")
	fmt.Println("7. Gesture type=30(Change) sub=23(Swipe) phase=Ended")
	fmt.Println("8. Gesture type=29(Begin) sub=0")
}