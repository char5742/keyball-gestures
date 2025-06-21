package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	log.Println("=== スクロールシミュレーションテスト ===")
	log.Println("trackpad_cgevent_simpleを別ターミナルで実行してください:")
	log.Println("  ./trackpad_cgevent_simple")
	log.Println("")
	log.Println("10秒後にスクロールシミュレーションを開始します...")
	time.Sleep(10 * time.Second)

	// タッチパッドを作成
	cfg := features.TouchPadConfig{
		Name:                  "TestTouchPad",
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

	fmt.Println("\n=== 開始: 2本指スクロール ===")

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

	// 上にスクロール（指を上に移動）
	fmt.Println("上にスクロール中...")
	for i := 0; i < 20; i++ {
		y := int32(16000 - i*500)
		err = touchpad.MultiTouchMove(0, 16000, y)
		if err != nil {
			log.Printf("Error moving finger 0: %v", err)
		}
		err = touchpad.MultiTouchMove(1, 17000, y)
		if err != nil {
			log.Printf("Error moving finger 1: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// タッチアップ
	err = touchpad.MultiTouchUp(0)
	if err != nil {
		log.Printf("Error on finger 0 up: %v", err)
	}
	err = touchpad.MultiTouchUp(1)
	if err != nil {
		log.Printf("Error on finger 1 up: %v", err)
	}

	fmt.Println("\n=== 完了 ===")
	fmt.Println("trackpad_cgevent_simpleでスクロールイベントが検出されましたか？")
	fmt.Println("ブラウザ等でスクロールが発生しましたか？")
}