package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("=== 実際のスクロール動作テスト ===")
	
	// まず、test_fixed_scrollを実行して基本的なスクロールが動作するか確認
	log.Println("1. 基本的なCGEventCreateScrollWheelEventでのテスト")
	log.Println("3秒後に開始...")
	time.Sleep(3 * time.Second)
	
	cmd := exec.Command("./test/test_fixed_scroll")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("test_fixed_scroll実行エラー: %v", err)
	}
	log.Printf("結果:\n%s", output)
	
	log.Println("\n=== ブラウザでスクロールが発生しましたか？ ===")
	log.Println("発生した場合: CGEventCreateScrollWheelEventは正常に動作しています")
	log.Println("発生しなかった場合: アクセシビリティ権限を確認してください")
	time.Sleep(5 * time.Second)
	
	// 次に、修正したtouchpadでのテスト
	log.Println("\n2. 修正したTouchPadでのテスト")
	
	cfg := features.TouchPadConfig{
		Name:                  "ScrollTest",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		MotionSmoothingFactor: 0.0,
		MotionWarmUpCount:     0,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		log.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	log.Println("2本指スクロールシミュレーション開始...")
	
	// 2本指タッチダウン
	err = touchpad.MultiTouchDown(0, 1001, 16000, 16000)
	if err != nil {
		log.Printf("Error on first finger down: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	err = touchpad.MultiTouchDown(1, 1002, 17000, 16000)
	if err != nil {
		log.Printf("Error on second finger down: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 大きく上にスクロール
	log.Println("大きく上にスクロール...")
	for i := 0; i < 10; i++ {
		y := int32(16000 - i*1000) // より大きな移動量
		err = touchpad.MultiTouchMove(0, 16000, y)
		if err != nil {
			log.Printf("Error moving finger 0: %v", err)
		}
		err = touchpad.MultiTouchMove(1, 17000, y)
		if err != nil {
			log.Printf("Error moving finger 1: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		fmt.Printf(".")
	}
	fmt.Println()

	// タッチアップ
	err = touchpad.MultiTouchUp(0)
	if err != nil {
		log.Printf("Error on finger 0 up: %v", err)
	}
	err = touchpad.MultiTouchUp(1)
	if err != nil {
		log.Printf("Error on finger 1 up: %v", err)
	}

	log.Println("\n=== 結果確認 ===")
	log.Println("TouchPadでスクロールが発生しましたか？")
	log.Println("発生した場合: 修正は成功です！")
	log.Println("発生しなかった場合: さらなる調査が必要です")
}