package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== Mission Control動作テスト ===")
	fmt.Println("システム環境設定 > トラックパッド > その他のジェスチャ で")
	fmt.Println("「Mission Control」が「4本指で上にスワイプ」に設定されていることを確認してください")
	fmt.Println("\n5秒後に上スワイプを実行します...")
	
	time.Sleep(5 * time.Second)

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

	// 画面中央付近から開始
	centerX := int32(600)
	centerY := int32(400)
	
	fmt.Println("4本指で上スワイプ開始...")
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, centerX+int32(i*50), centerY)
	}
	
	// 少し待機（ジェスチャー認識のため）
	time.Sleep(50 * time.Millisecond)
	
	// 上方向に大きくスワイプ（Y座標を減少）
	steps := 15
	for j := 1; j <= steps; j++ {
		newY := centerY - int32(j*20) // より大きな移動量
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, centerX+int32(i*50), newY)
		}
		time.Sleep(15 * time.Millisecond) // より自然なタイミング
	}
	
	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	fmt.Println("\n完了！")
	fmt.Println("Mission Controlが起動しましたか？")
	fmt.Println("\n起動しない場合は:")
	fmt.Println("1. システム環境設定でジェスチャーが有効か確認")
	fmt.Println("2. アクセシビリティ権限を確認")
	fmt.Println("3. 他のアプリケーションがジェスチャーを奪っていないか確認")
}