package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== デルタ値ありバージョンのテスト ===")
	fmt.Println("postSwipeGestureを一時的に修正して、デルタ値を送信します")
	
	// タッチパッドを初期化（スケールファクターを設定）
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      1.0,  // デルタ値を有効にする
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}
	defer touchpad.Close()

	fmt.Println("\n注意: このテストでは、内部的にデルタ値が計算されますが、")
	fmt.Println("現在のpostSwipeGesture関数はデルタを0に固定しています。")
	fmt.Println("もしMission Controlを動作させるにはデルタ値が必要なら、")
	fmt.Println("postSwipeGesture内の以下の行を修正する必要があります：")
	fmt.Println("")
	fmt.Println("変更前:")
	fmt.Println("    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaX, 0.0);")
	fmt.Println("    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaY, 0.0);")
	fmt.Println("")
	fmt.Println("変更後:")
	fmt.Println("    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaX, deltaX);")
	fmt.Println("    CGEventSetDoubleValueField(gesture, kCGEventFieldGestureDeltaY, deltaY);")
	
	fmt.Println("\n5秒後にテスト開始...")
	time.Sleep(5 * time.Second)
	
	// 上スワイプ（Mission Control）
	fmt.Println("上スワイプ実行...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 600+int32(i*50), 600)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	// 上に大きくスワイプ
	for j := 1; j <= 10; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 600+int32(i*50), 600-int32(j*40))
		}
		time.Sleep(20 * time.Millisecond)
	}
	
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\n完了！")
	fmt.Println("\nもしMission Controlが起動しない場合、")
	fmt.Println("touchpad_darwin.goのpostSwipeGesture関数を上記のように修正してください。")
}