package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	// trackpad_dumpを起動
	fmt.Println("trackpad_dumpを起動しています...")
	cmd := exec.Command("./test/debugging/trackpad_dump")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		log.Fatalf("trackpad_dumpの起動に失敗: %v", err)
	}
	
	// 少し待機
	time.Sleep(2 * time.Second)
	
	// タッチパッドを初期化
	config := features.TouchPadConfig{
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     4,
		SwipeScaleFactor:      0.01,
		ScrollScaleFactor:     2.0,
	}

	touchpad, err := features.CreateTouchPad(config)
	if err != nil {
		log.Fatalf("タッチパッドの初期化に失敗: %v", err)
	}

	fmt.Println("\n=== 4本指スワイプを生成 ===")
	
	// 4本指タッチダウン
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*10), 500)
		time.Sleep(5 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	// 右スワイプ（小刻みに動かす）
	for j := 0; j < 10; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*10)+int32(j*20), 500)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// タッチアップ
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
		time.Sleep(5 * time.Millisecond)
	}

	touchpad.Close()
	
	// trackpad_dumpを終了
	time.Sleep(2 * time.Second)
	cmd.Process.Kill()
	
	fmt.Println("\n完了")
}