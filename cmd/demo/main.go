package main

import (
	"fmt"
	"log"
	"time"

	"github.com/char5742/keyball-gestures/internal/api"
	"github.com/char5742/keyball-gestures/internal/config"
)

func main() {
	log.Println("=== Keyball Gestures Demo ===")
	log.Println("このデモではF1/F2キーを使用してジェスチャーをテストします")
	log.Println("F1キー + マウス移動: 2本指スクロール")
	log.Println("F2キー + マウス移動: 4本指スワイプ")
	log.Println("")
	
	// テスト用の設定
	cfg := &config.Config{
		TouchPad: config.TouchPadConfig{
			MinX: 0,
			MaxX: 10000,
			MinY: 0,
			MaxY: 10000,
		},
		Input: config.InputConfig{
			TwoFingerKey:  122, // F1
			FourFingerKey: 120, // F2
		},
		Motion: config.MotionConfig{
			FilterSmoothingFactor: 0.0,
			FilterWarmUpCount:     0,
			MouseDeltaFactor:      50,
		},
		Gesture: config.GestureConfig{
			ResetThreshold: 1000 * time.Millisecond,
		},
		DevicePrefs: config.DevicePrefsConfig{
			PreferredKeyboardDevice: "Keyball39",
			PreferredMouseDevice:    "Keyball39",
		},
	}
	
	// サービスを作成
	service := api.NewGestureService(cfg)
	
	// サービスを開始
	log.Println("ジェスチャーサービスを開始します...")
	if err := service.Start(); err != nil {
		log.Fatalf("サービスの開始に失敗しました: %v", err)
	}
	
	log.Println("")
	log.Println("準備完了！以下を試してください：")
	log.Println("1. F1キーを押しながらマウスを上下に動かす → スクロール")
	log.Println("2. F2キーを押しながらマウスを左右に動かす → スワイプ")
	log.Println("")
	log.Println("Ctrl+Cで終了します")
	
	// 無限ループ
	for {
		time.Sleep(1 * time.Second)
		if service.IsRunning() {
			fmt.Print(".")
		} else {
			log.Println("\nサービスが停止しました")
			break
		}
	}
}