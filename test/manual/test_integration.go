package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

func main() {
	fmt.Println("=== 統合テスト: trackpad_dumpによる検証 ===")
	
	// trackpad_dumpを起動
	dumpPath := filepath.Join("test", "debugging", "trackpad_dump")
	cmd := exec.Command(dumpPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		log.Printf("trackpad_dumpの起動に失敗（スキップ）: %v", err)
		// trackpad_dumpなしでも継続
	} else {
		fmt.Println("trackpad_dumpを起動しました")
		defer func() {
			cmd.Process.Kill()
			cmd.Wait()
		}()
	}
	
	// 少し待機
	time.Sleep(1 * time.Second)
	
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
	defer touchpad.Close()

	fmt.Println("\n=== 4本指スワイプテスト ===")
	fmt.Println("期待されるイベント:")
	fmt.Println("  Gesture type=30(Change) sub=23(Swipe) phase=Began mask=0x1 flags=0x0")
	fmt.Println("  Gesture type=29(Begin) sub=0")
	fmt.Println("  Gesture type=30(Change) sub=23(Swipe) phase=Changed mask=0x2 flags=0x8")
	fmt.Println("  Gesture type=29(Begin) sub=0")
	fmt.Println("  ...")
	fmt.Println("")
	
	// 4本指タッチダウン
	fmt.Println("4本指タッチダウン...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchDown(i, i+1, 500+int32(i*20), 500)
		time.Sleep(10 * time.Millisecond) // 素早く追加
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// 右スワイプ
	fmt.Println("右スワイプ中...")
	for j := 1; j <= 10; j++ {
		for i := 0; i < 4; i++ {
			touchpad.MultiTouchMove(i, 500+int32(i*20)+int32(j*30), 500)
		}
		time.Sleep(20 * time.Millisecond)
	}
	
	// タッチアップ
	fmt.Println("タッチアップ...")
	for i := 0; i < 4; i++ {
		touchpad.MultiTouchUp(i)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	fmt.Println("\n=== 2本指スクロールテスト ===")
	fmt.Println("期待されるイベント:")
	fmt.Println("  Scroll ? Δ(X,Y)")
	fmt.Println("  Gesture type=29(Begin) sub=6(Scroll)")
	fmt.Println("")
	
	// 2本指タッチダウン
	fmt.Println("2本指タッチダウン...")
	touchpad.MultiTouchDown(0, 11, 600, 600)
	time.Sleep(10 * time.Millisecond)
	touchpad.MultiTouchDown(1, 12, 700, 600)
	
	time.Sleep(100 * time.Millisecond)
	
	// 下スクロール
	fmt.Println("下スクロール中...")
	for j := 1; j <= 5; j++ {
		touchpad.MultiTouchMove(0, 600, 600+int32(j*50))
		touchpad.MultiTouchMove(1, 700, 600+int32(j*50))
		time.Sleep(30 * time.Millisecond)
	}
	
	// タッチアップ
	fmt.Println("タッチアップ...")
	touchpad.MultiTouchUp(0)
	touchpad.MultiTouchUp(1)
	
	time.Sleep(1 * time.Second)
	
	fmt.Println("\n=== テスト完了 ===")
	fmt.Println("上記のtrackpad_dump出力を確認してください：")
	fmt.Println("1. 4本指スワイプ: type=30, sub=23, Δ(0.0,0.0)")
	fmt.Println("2. 2本指スクロール: type=22(Scroll), Δ(X,Y)が0以外")
}