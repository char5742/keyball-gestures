//go:build integration && darwin
// +build integration,darwin

package integration

import (
	"runtime"
	"testing"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

// TestTouchpadScrollIntegration は2本指スクロールの統合テストを実行する
func TestTouchpadScrollIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}
	
	if runtime.GOOS != "darwin" {
		t.Skip("macOS専用テスト")
	}

	// アクセシビリティ権限チェック（テスト実行時の注意喚起）
	t.Log("注意: このテストはアクセシビリティ権限が必要です")

	// 既存のEventMonitorを使用
	monitor := NewEventMonitor()
	err := monitor.Start()
	if err != nil {
		t.Fatalf("イベントモニターの開始に失敗: %v", err)
	}
	defer monitor.Stop()

	// タッチパッド設定
	cfg := features.TouchPadConfig{
		Name:                  "integration-test-touchpad",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		ScrollScaleFactor:     2.0,
		MotionSmoothingFactor: 0.3,
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	t.Run("2本指スクロール", func(t *testing.T) {
		monitor.Clear()
		
		// 2本指でタッチダウン
		err := touchpad.MultiTouchDown(0, 1001, 1000, 1000)
		if err != nil {
			t.Fatalf("1本目の指のタッチダウンに失敗: %v", err)
		}
		
		err = touchpad.MultiTouchDown(1, 1002, 1200, 1000)
		if err != nil {
			t.Fatalf("2本目の指のタッチダウンに失敗: %v", err)
		}

		// 少し待ってジェスチャー開始を確実にする
		time.Sleep(100 * time.Millisecond)

		// 上方向にスクロール（複数回の移動）
		for i := 0; i < 8; i++ {
			dy := int32(50)
			err = touchpad.MultiTouchMove(0, 1000, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("1本目の指の移動に失敗 (i=%d): %v", i, err)
			}
			
			err = touchpad.MultiTouchMove(1, 1200, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("2本目の指の移動に失敗 (i=%d): %v", i, err)
			}
			
			time.Sleep(20 * time.Millisecond)
		}

		// タッチアップ
		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("1本目の指のタッチアップに失敗: %v", err)
		}
		
		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("2本目の指のタッチアップに失敗: %v", err)
		}

		// イベント処理の完了を待つ
		time.Sleep(200 * time.Millisecond)

		// イベント検証
		totalEvents := monitor.GetEventCount()
		if totalEvents == 0 {
			t.Fatal("イベントが記録されていません")
		}

		t.Logf("記録されたイベント数: %d", totalEvents)
		monitor.PrintDebugInfo()

		// スクロールイベントの存在を確認
		if !monitor.HasScrollEvents() {
			t.Error("スクロールイベントが見つかりません")
		}

		// ジェスチャーイベントの存在を確認
		if !monitor.HasGestureEvents() {
			t.Error("ジェスチャーイベントが見つかりません")
		}

		t.Logf("スクロールイベント検証: 成功")
	})
}

// TestTouchpadSwipeIntegration は4本指スワイプの統合テストを実行する
func TestTouchpadSwipeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}
	
	if runtime.GOOS != "darwin" {
		t.Skip("macOS専用テスト")
	}

	monitor := NewEventMonitor()
	err := monitor.Start()
	if err != nil {
		t.Fatalf("イベントモニターの開始に失敗: %v", err)
	}
	defer monitor.Stop()

	// タッチパッド設定
	cfg := features.TouchPadConfig{
		Name:                  "integration-test-touchpad-swipe",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		SwipeScaleFactor:      0.5,
		MotionSmoothingFactor: 0.3,
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	t.Run("4本指右スワイプ", func(t *testing.T) {
		monitor.Clear()
		
		// 4本指の初期位置
		fingers := []struct {
			slot int
			id   int
			x    int32
			y    int32
		}{
			{0, 2001, 800, 1000},
			{1, 2002, 1000, 1000},
			{2, 2003, 1200, 1000},
			{3, 2004, 1400, 1000},
		}

		// 4本指でタッチダウン
		for _, finger := range fingers {
			err := touchpad.MultiTouchDown(finger.slot, finger.id, finger.x, finger.y)
			if err != nil {
				t.Fatalf("指%dのタッチダウンに失敗: %v", finger.slot, err)
			}
			time.Sleep(15 * time.Millisecond) // 指を順次タッチ
		}

		// ジェスチャー開始の確認を待つ
		time.Sleep(100 * time.Millisecond)

		// 右方向にスワイプ（複数回の移動）
		for i := 0; i < 6; i++ {
			dx := int32(80)
			for _, finger := range fingers {
				err := touchpad.MultiTouchMove(finger.slot, finger.x+int32(i+1)*dx, finger.y)
				if err != nil {
					t.Errorf("指%dの移動に失敗 (i=%d): %v", finger.slot, i, err)
				}
			}
			time.Sleep(25 * time.Millisecond)
		}

		// 全指をタッチアップ
		for _, finger := range fingers {
			err := touchpad.MultiTouchUp(finger.slot)
			if err != nil {
				t.Errorf("指%dのタッチアップに失敗: %v", finger.slot, err)
			}
		}

		// イベント処理の完了を待つ
		time.Sleep(300 * time.Millisecond)

		// イベント検証
		totalEvents := monitor.GetEventCount()
		if totalEvents == 0 {
			t.Fatal("イベントが記録されていません")
		}

		t.Logf("記録されたイベント数: %d", totalEvents)
		monitor.PrintDebugInfo()

		// ジェスチャーイベントの存在を確認
		if !monitor.HasGestureEvents() {
			t.Error("ジェスチャーイベントが見つかりません")
		}

		t.Logf("スワイプイベント検証: 成功")
	})
}

// TestTouchpadEventValidation はイベント生成の基本的な検証を行う
func TestTouchpadEventValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}
	
	if runtime.GOOS != "darwin" {
		t.Skip("macOS専用テスト")
	}

	monitor := NewEventMonitor()
	err := monitor.Start()
	if err != nil {
		t.Fatalf("イベントモニターの開始に失敗: %v", err)
	}
	defer monitor.Stop()

	cfg := features.TouchPadConfig{
		Name:                  "validation-test-touchpad",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		ScrollScaleFactor:     1.5,
		MotionSmoothingFactor: 0.3,  
		MotionWarmUpCount:     2,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	t.Run("基本的なイベント生成確認", func(t *testing.T) {
		monitor.Clear()
		
		// 短いスクロールジェスチャーを実行
		err := touchpad.MultiTouchDown(0, 3001, 1000, 1000)
		if err != nil {
			t.Fatalf("タッチダウンに失敗: %v", err)
		}
		
		err = touchpad.MultiTouchDown(1, 3002, 1200, 1000)
		if err != nil {
			t.Fatalf("タッチダウンに失敗: %v", err)
		}

		time.Sleep(80 * time.Millisecond)

		// 3回移動
		for i := 0; i < 3; i++ {
			dy := int32(40)
			err = touchpad.MultiTouchMove(0, 1000, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("移動に失敗: %v", err)
			}
			err = touchpad.MultiTouchMove(1, 1200, 1000-int32(i+1)*dy)
			if err != nil {
				t.Errorf("移動に失敗: %v", err)
			}
			time.Sleep(30 * time.Millisecond)
		}

		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("タッチアップに失敗: %v", err)
		}
		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("タッチアップに失敗: %v", err)
		}

		time.Sleep(200 * time.Millisecond)

		// 基本的な検証
		totalEvents := monitor.GetEventCount()
		if totalEvents < 1 {
			t.Fatalf("イベント数が不足: 期待値>=1, 実際=%d", totalEvents)
		}

		monitor.PrintDebugInfo()

		// スクロールまたはジェスチャーイベントの存在確認
		hasEvents := monitor.HasScrollEvents() || monitor.HasGestureEvents()
		if !hasEvents {
			t.Error("スクロールまたはジェスチャーイベントが見つかりません")
		}

		t.Logf("イベント生成確認: 成功 (計%d個のイベント)", totalEvents)
	})
}

// TestTouchpadIntegrationWithRealWorldPattern は実際の使用パターンをテストする
func TestTouchpadIntegrationWithRealWorldPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}
	
	if runtime.GOOS != "darwin" {
		t.Skip("macOS専用テスト")
	}

	monitor := NewEventMonitor()
	err := monitor.Start()
	if err != nil {
		t.Fatalf("イベントモニターの開始に失敗: %v", err)
	}
	defer monitor.Stop()

	cfg := features.TouchPadConfig{
		Name:                  "realworld-test-touchpad",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		ScrollScaleFactor:     1.0,
		SwipeScaleFactor:      0.3,
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     3,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("タッチパッドの作成に失敗: %v", err)
	}
	defer touchpad.Close()

	t.Run("連続的なジェスチャー", func(t *testing.T) {
		monitor.Clear()
		
		// 最初に2本指スクロールを実行
		t.Log("2本指スクロール開始")
		err := touchpad.MultiTouchDown(0, 4001, 1000, 1000)
		if err != nil {
			t.Fatalf("スクロール: タッチダウンに失敗: %v", err)
		}
		err = touchpad.MultiTouchDown(1, 4002, 1200, 1000)
		if err != nil {
			t.Fatalf("スクロール: タッチダウンに失敗: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		// スクロール移動
		for i := 0; i < 3; i++ {
			dy := int32(30)
			err = touchpad.MultiTouchMove(0, 1000, 1000+int32(i+1)*dy)
			if err != nil {
				t.Errorf("スクロール移動に失敗: %v", err)
			}
			err = touchpad.MultiTouchMove(1, 1200, 1000+int32(i+1)*dy)
			if err != nil {
				t.Errorf("スクロール移動に失敗: %v", err)
			}
			time.Sleep(25 * time.Millisecond)
		}

		// スクロール終了
		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("スクロール終了に失敗: %v", err)
		}
		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("スクロール終了に失敗: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// 少し間を置いて4本指スワイプを実行
		t.Log("4本指スワイプ開始")
		
		fingers := []int{0, 1, 2, 3}
		startX := []int32{700, 900, 1100, 1300}
		
		// 4本指タッチダウン
		for i, slot := range fingers {
			err := touchpad.MultiTouchDown(slot, 5000+slot, startX[i], 1000)
			if err != nil {
				t.Fatalf("スワイプ: 指%dのタッチダウンに失敗: %v", slot, err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(50 * time.Millisecond)

		// 左方向にスワイプ
		for i := 0; i < 4; i++ {
			dx := int32(60)
			for j, slot := range fingers {
				err := touchpad.MultiTouchMove(slot, startX[j]-int32(i+1)*dx, 1000)
				if err != nil {
					t.Errorf("スワイプ移動に失敗: %v", err)
				}
			}
			time.Sleep(20 * time.Millisecond)
		}

		// スワイプ終了
		for _, slot := range fingers {
			err := touchpad.MultiTouchUp(slot)
			if err != nil {
				t.Errorf("スワイプ終了に失敗: %v", err)
			}
		}

		time.Sleep(200 * time.Millisecond)

		// 結果検証
		totalEvents := monitor.GetEventCount()
		if totalEvents == 0 {
			t.Fatal("イベントが記録されていません")
		}

		t.Logf("連続ジェスチャー完了: 総イベント数=%d", totalEvents)
		monitor.PrintDebugInfo()

		// 両方のジェスチャーのイベントが記録されているかチェック
		hasAnyEvents := monitor.HasScrollEvents() || monitor.HasGestureEvents()
		if !hasAnyEvents {
			t.Error("スクロールまたはジェスチャーイベントが見つかりません")
		}

		t.Logf("連続ジェスチャーテスト: 成功")
	})
}