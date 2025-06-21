//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"log"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== Slot Debug Test ===")
	
	cfg := features.TouchPadConfig{
		Name:              "slot-test",
		MinX:              0,
		MaxX:              32767,
		MinY:              0,
		MaxY:              32767,
		MotionWarmUpCount: 2,  // 重要: これを確認
	}

	fmt.Printf("MotionWarmUpCount: %d\n", cfg.MotionWarmUpCount)
	
	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	defer touchpad.Close()

	// スロットの組み合わせをテスト
	tests := []struct {
		name  string
		slots []int
	}{
		{"Sequential 0-3", []int{0, 1, 2, 3}},
		{"Sequential 1-4", []int{1, 2, 3, 4}},
		{"Non-sequential", []int{0, 2, 4, 6}},
		{"Only 2 slots", []int{0, 1}},
	}

	for _, test := range tests {
		fmt.Printf("\n--- Test: %s ---\n", test.name)
		
		// タッチダウン
		for i, slot := range test.slots {
			err = touchpad.MultiTouchDown(slot, 1000+slot, int32(1000+i*200), 1000)
			if err != nil {
				fmt.Printf("Slot %d: ERROR - %v\n", slot, err)
			} else {
				fmt.Printf("Slot %d: DOWN\n", slot)
			}
		}
		
		// タッチアップ
		for _, slot := range test.slots {
			err = touchpad.MultiTouchUp(slot)
			if err != nil {
				fmt.Printf("Slot %d: UP ERROR\n", slot)
			}
		}
		
		fmt.Println("Check trackpad_dump for this test")
	}
}