//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== Touchpad Manual Test ===")
	fmt.Println("このテストプログラムは修正されたtouchpad_darwin.goをテストします")
	fmt.Println("")
	fmt.Println("実行前に以下を準備してください：")
	fmt.Println("1. 別のターミナルで trackpad_dump を実行")
	fmt.Println("   cd test/debugging && ./trackpad_dump")
	fmt.Println("2. アクセシビリティ権限が設定されていることを確認")
	fmt.Println("")
	fmt.Print("準備が完了したらEnterキーを押してください...")
	fmt.Scanln()

	// タッチパッド設定
	cfg := features.TouchPadConfig{
		Name:                  "manual-test-touchpad",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		ScrollScaleFactor:     1.5,
		SwipeScaleFactor:      0.5,
		MotionSmoothingFactor: 0.3,
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		log.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n=== 4本指スワイプテスト開始 ===")
	fmt.Println("trackpad_dumpで以下のようなログが表示されることを確認してください：")
	fmt.Println("[time] Gesture type=29(Begin) sub=23(Swipe) phase=Began ...")
	fmt.Println("[time] Gesture type=30(Change) sub=23(Swipe) phase=Changed ...")
	fmt.Println("[time] Gesture type=31(End) sub=23(Swipe) phase=Ended ...")
	fmt.Println("")

	// 4本指スワイプを実行
	fingers := []struct {
		slot int
		id   int
		x    int32
		y    int32
	}{
		{0, 7001, 800, 1000},
		{1, 7002, 1000, 1000},
		{2, 7003, 1200, 1000},
		{3, 7004, 1400, 1000},
	}

	fmt.Println("1. 4本指タッチダウン開始...")
	for i, finger := range fingers {
		err := touchpad.MultiTouchDown(finger.slot, finger.id, finger.x, finger.y)
		if err != nil {
			log.Printf("指%dのタッチダウンに失敗: %v", finger.slot, err)
		} else {
			fmt.Printf("   指%d タッチダウン完了 (slot=%d, id=%d)\n", i+1, finger.slot, finger.id)
		}
		time.Sleep(10 * time.Millisecond) // 合計40ms < 50ms (gestureDetectionWindow内)
	}

	fmt.Println("\n2. ジェスチャー開始を待機中...")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n3. 右方向スワイプ実行中...")
	for i := 0; i < 10; i++ {
		dx := int32(80)
		for _, finger := range fingers {
			err := touchpad.MultiTouchMove(finger.slot, finger.x+int32(i+1)*dx, finger.y)
			if err != nil {
				log.Printf("指%dの移動に失敗: %v", finger.slot, err)
			}
		}
		fmt.Printf("   移動 %d/10 完了\n", i+1)
		time.Sleep(35 * time.Millisecond)
	}

	fmt.Println("\n4. タッチアップ実行中...")
	for i, finger := range fingers {
		err := touchpad.MultiTouchUp(finger.slot)
		if err != nil {
			log.Printf("指%dのタッチアップに失敗: %v", finger.slot, err)
		}
		fmt.Printf("   指%d タッチアップ完了\n", i+1)
	}

	fmt.Println("\n=== 4本指スワイプテスト完了 ===")
	fmt.Println("trackpad_dumpでジェスチャーイベント（type=29,30,31）が表示されていれば成功です！")
	
	fmt.Println("\n=== 2本指スクロールテスト開始 ===")
	time.Sleep(1 * time.Second)

	fmt.Println("5. 2本指スクロール実行中...")
	err = touchpad.MultiTouchDown(0, 8001, 1000, 1000)
	if err != nil {
		log.Printf("スクロール: 1本目タッチダウンに失敗: %v", err)
	}
	
	err = touchpad.MultiTouchDown(1, 8002, 1200, 1000)
	if err != nil {
		log.Printf("スクロール: 2本目タッチダウンに失敗: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 8; i++ {
		dy := int32(50)
		err = touchpad.MultiTouchMove(0, 1000, 1000-int32(i+1)*dy)
		if err != nil {
			log.Printf("スクロール移動に失敗: %v", err)
		}
		err = touchpad.MultiTouchMove(1, 1200, 1000-int32(i+1)*dy)
		if err != nil {
			log.Printf("スクロール移動に失敗: %v", err)
		}
		fmt.Printf("   スクロール %d/8 完了\n", i+1)
		time.Sleep(30 * time.Millisecond)
	}

	err = touchpad.MultiTouchUp(0)
	if err != nil {
		log.Printf("スクロール終了に失敗: %v", err)
	}
	err = touchpad.MultiTouchUp(1)
	if err != nil {
		log.Printf("スクロール終了に失敗: %v", err)
	}

	fmt.Println("\n=== 全テスト完了 ===")
	fmt.Println("trackpad_dumpで以下が確認できていれば修正が成功しています：")
	fmt.Println("✅ スクロールイベント（type=22）")
	fmt.Println("✅ ジェスチャーイベント（type=29,30,31, sub=23）")
	fmt.Println("")
	fmt.Println("結果をtrackpad_dumpのログで確認してください。")
}