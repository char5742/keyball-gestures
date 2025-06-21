//go:build integration && darwin
// +build integration,darwin

package integration

import (
	"runtime"
	"testing"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

// TestTouchPadScrollWithValidation tests touchpad scroll functionality with event validation
func TestTouchPadScrollWithValidation(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Event monitoring is only available on macOS")
	}

	cfg := features.TouchPadConfig{
		Name:                  "test-touchpad-validation",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     3,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	// イベントモニターを作成
	monitor := NewEventMonitor()
	if err := monitor.Start(); err != nil {
		t.Fatalf("Failed to start event monitor: %v", err)
	}
	defer monitor.Stop()

	// モニターが準備できるまで待つ
	time.Sleep(100 * time.Millisecond)

	t.Run("TwoFingerScrollValidation", func(t *testing.T) {
		// イベントをクリア
		monitor.Clear()

		// 2本指タッチを開始
		err := touchpad.MultiTouchDown(0, 1001, 5000, 5000)
		if err != nil {
			t.Errorf("Failed to start first finger touch: %v", err)
		}

		err = touchpad.MultiTouchDown(1, 1002, 5100, 5000)
		if err != nil {
			t.Errorf("Failed to start second finger touch: %v", err)
		}

		// 少し待つ
		time.Sleep(50 * time.Millisecond)

		// 指を上に移動（スクロール）
		for i := 0; i < 10; i++ {
			err = touchpad.MultiTouchMove(0, 5000, int32(5000-i*100))
			if err != nil {
				t.Errorf("Failed to move first finger: %v", err)
			}

			err = touchpad.MultiTouchMove(1, 5100, int32(5000-i*100))
			if err != nil {
				t.Errorf("Failed to move second finger: %v", err)
			}

			time.Sleep(20 * time.Millisecond)
		}

		// タッチを終了
		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("Failed to end first finger touch: %v", err)
		}

		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("Failed to end second finger touch: %v", err)
		}

		// イベントが記録されるまで待つ
		time.Sleep(100 * time.Millisecond)

		// スクロールイベントが生成されたか確認
		if !monitor.HasScrollEvents() {
			t.Error("Expected scroll events to be generated, but none were detected")
			monitor.PrintDebugInfo()
		}

		eventCount := monitor.GetEventCount()
		if eventCount == 0 {
			t.Error("No events were recorded")
		} else {
			t.Logf("Recorded %d events", eventCount)
		}
	})
}

// TestTouchPadGesturesWithValidation tests touchpad gestures with event validation
func TestTouchPadGesturesWithValidation(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Event monitoring is only available on macOS")
	}

	cfg := features.TouchPadConfig{
		Name:                  "test-touchpad-gesture-validation",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     3,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		t.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	// イベントモニターを作成
	monitor := NewEventMonitor()
	if err := monitor.Start(); err != nil {
		t.Fatalf("Failed to start event monitor: %v", err)
	}
	defer monitor.Stop()

	// モニターが準備できるまで待つ
	time.Sleep(100 * time.Millisecond)

	t.Run("FourFingerSwipeValidation", func(t *testing.T) {
		// イベントをクリア
		monitor.Clear()

		// 4本指タッチを開始
		fingers := []struct {
			id int
			x  int32
			y  int32
		}{
			{0, 4000, 5000},
			{1, 4300, 5000},
			{2, 4600, 5000},
			{3, 4900, 5000},
		}

		// すべての指をタッチダウン
		for _, f := range fingers {
			err := touchpad.MultiTouchDown(f.id, 2000+f.id, f.x, f.y)
			if err != nil {
				t.Errorf("Failed to start finger %d touch: %v", f.id, err)
			}
			time.Sleep(20 * time.Millisecond)
		}

		// 少し待つ
		time.Sleep(50 * time.Millisecond)

		// 指を左に移動（スワイプ）
		for i := 0; i < 10; i++ {
			for _, f := range fingers {
				err := touchpad.MultiTouchMove(f.id, f.x-int32(i*100), f.y)
				if err != nil {
					t.Errorf("Failed to move finger %d: %v", f.id, err)
				}
			}
			time.Sleep(20 * time.Millisecond)
		}

		// すべての指を離す
		for _, f := range fingers {
			err := touchpad.MultiTouchUp(f.id)
			if err != nil {
				t.Errorf("Failed to end finger %d touch: %v", f.id, err)
			}
			time.Sleep(20 * time.Millisecond)
		}

		// イベントが記録されるまで待つ
		time.Sleep(100 * time.Millisecond)

		// ジェスチャーイベントが生成されたか確認
		if !monitor.HasGestureEvents() && !monitor.HasScrollEvents() {
			t.Error("Expected gesture or scroll events to be generated, but none were detected")
			monitor.PrintDebugInfo()
		}

		eventCount := monitor.GetEventCount()
		if eventCount == 0 {
			t.Error("No events were recorded")
		} else {
			t.Logf("Recorded %d events", eventCount)
		}
	})
}

// TestRealDeviceIntegration tests with actual device if available
func TestRealDeviceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real device test in short mode")
	}

	if runtime.GOOS != "darwin" {
		t.Skip("Real device test is only available on macOS")
	}

	// キーボードの作成を試みる
	kb, err := features.CreateKeyboard("Keyball39")
	if err != nil {
		t.Logf("Could not create keyboard (device may not be connected): %v", err)
		t.Skip("Skipping test - Keyball39 keyboard not available")
	}
	defer kb.Close()

	// マウスの作成を試みる
	mouse, err := features.CreateMouse("Keyball39")
	if err != nil {
		t.Logf("Could not create mouse (device may not be connected): %v", err)
		t.Skip("Skipping test - Keyball39 mouse not available")
	}
	defer mouse.Close()

	// イベントモニターを作成
	monitor := NewEventMonitor()
	if err := monitor.Start(); err != nil {
		t.Fatalf("Failed to start event monitor: %v", err)
	}
	defer monitor.Stop()

	t.Log("Real device test ready. Please press F13 or F14 key within 5 seconds...")

	// 5秒間、実際のキー入力を待つ
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	keyDetected := false
	for {
		select {
		case <-timeout:
			if !keyDetected {
				t.Log("No key press detected (this is expected in CI environment)")
			}
			return
		case <-ticker.C:
			key := kb.GetKey()
			if key > 0 {
				keyDetected = true
				t.Logf("Key detected: %d", key)
				if key == 183 {
					t.Log("F13 key pressed!")
				} else if key == 184 {
					t.Log("F14 key pressed!")
				}
			}
		}
	}
}