//go:build integration
// +build integration

package integration

import (
	"runtime"
	"testing"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

// TestKeyboardCreation tests keyboard device creation
func TestKeyboardCreation(t *testing.T) {
	// Skip on macOS due to RunLoop blocking issue in tests
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping keyboard creation test on macOS due to RunLoop blocking")
	}
	
	kb, err := features.CreateKeyboard("test-keyboard")
	if err != nil {
		t.Fatalf("Failed to create keyboard: %v", err)
	}
	defer kb.Close()

	// Test that keyboard is created successfully
	if kb == nil {
		t.Error("Expected keyboard to be created, but got nil")
	}
}

// TestMouseCreation tests mouse device creation
func TestMouseCreation(t *testing.T) {
	// Skip on macOS due to RunLoop blocking issue in tests
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping mouse creation test on macOS due to RunLoop blocking")
	}
	
	mouse, err := features.CreateMouse("test-mouse")
	if err != nil {
		t.Fatalf("Failed to create mouse: %v", err)
	}
	defer mouse.Close()

	// Test that mouse is created successfully
	if mouse == nil {
		t.Error("Expected mouse to be created, but got nil")
	}
}

// TestTouchPadCreation tests touchpad device creation
func TestTouchPadCreation(t *testing.T) {
	cfg := features.TouchPadConfig{
		Name:                  "test-touchpad",
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

	// Test that touchpad is created successfully
	if touchpad == nil {
		t.Error("Expected touchpad to be created, but got nil")
	}
}

// TestKeyboardInput tests keyboard input detection
func TestKeyboardInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping keyboard input test in short mode")
	}
	
	// Skip on macOS due to RunLoop blocking issue in tests
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping keyboard input test on macOS due to RunLoop blocking")
	}

	kb, err := features.CreateKeyboard("test-keyboard")
	if err != nil {
		t.Fatalf("Failed to create keyboard: %v", err)
	}
	defer kb.Close()

	// Test F13 key detection
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			// Timeout is expected in automated tests
			return
		default:
			key := kb.GetKey()
			// In automated tests, we don't expect actual key presses
			// This just verifies the API works
			if key != 0 {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// TestMouseMovement tests mouse movement detection
func TestMouseMovement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mouse movement test in short mode")
	}
	
	// Skip on macOS due to RunLoop blocking issue in tests
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping mouse movement test on macOS due to RunLoop blocking")
	}

	mouse, err := features.CreateMouse("test-mouse")
	if err != nil {
		t.Fatalf("Failed to create mouse: %v", err)
	}
	defer mouse.Close()

	// Test mouse delta detection
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			// Timeout is expected in automated tests
			return
		default:
			dx, dy := mouse.GetMouseDelta()
			// In automated tests, we don't expect actual mouse movement
			// This just verifies the API works
			_, _ = dx, dy
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// TestTouchPadScroll tests touchpad scroll functionality
func TestTouchPadScroll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping touchpad scroll test in short mode")
	}

	cfg := features.TouchPadConfig{
		Name:                  "test-touchpad",
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

	// Simulate 2-finger scroll
	t.Run("TwoFingerScroll", func(t *testing.T) {
		// Start 2-finger touch
		err := touchpad.MultiTouchDown(0, 1001, 500, 500)
		if err != nil {
			t.Errorf("Failed to start first finger touch: %v", err)
		}

		err = touchpad.MultiTouchDown(1, 1002, 600, 500)
		if err != nil {
			t.Errorf("Failed to start second finger touch: %v", err)
		}

		// Move fingers up (scroll)
		for i := 0; i < 5; i++ {
			err = touchpad.MultiTouchMove(0, 500, int32(500-i*50))
			if err != nil {
				t.Errorf("Failed to move first finger: %v", err)
			}

			err = touchpad.MultiTouchMove(1, 600, int32(500-i*50))
			if err != nil {
				t.Errorf("Failed to move second finger: %v", err)
			}

			time.Sleep(10 * time.Millisecond)
		}

		// End touches
		err = touchpad.MultiTouchUp(0)
		if err != nil {
			t.Errorf("Failed to end first finger touch: %v", err)
		}

		err = touchpad.MultiTouchUp(1)
		if err != nil {
			t.Errorf("Failed to end second finger touch: %v", err)
		}
	})
}

// TestTouchPadGestures tests touchpad gesture functionality
func TestTouchPadGestures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping touchpad gesture test in short mode")
	}

	cfg := features.TouchPadConfig{
		Name:                  "test-touchpad",
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

	// Test 4-finger swipe
	t.Run("FourFingerSwipe", func(t *testing.T) {
		// Start 4-finger touch
		fingers := []struct {
			id int
			x  int32
			y  int32
		}{
			{0, 400, 500},
			{1, 500, 500},
			{2, 600, 500},
			{3, 700, 500},
		}

		// Touch down all fingers
		for _, f := range fingers {
			err := touchpad.MultiTouchDown(f.id, 2000+f.id, f.x, f.y)
			if err != nil {
				t.Errorf("Failed to start finger %d touch: %v", f.id, err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Move fingers left (swipe)
		for i := 0; i < 5; i++ {
			for _, f := range fingers {
				err := touchpad.MultiTouchMove(f.id, f.x-int32(i*50), f.y)
				if err != nil {
					t.Errorf("Failed to move finger %d: %v", f.id, err)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}

		// End all touches
		for _, f := range fingers {
			err := touchpad.MultiTouchUp(f.id)
			if err != nil {
				t.Errorf("Failed to end finger %d touch: %v", f.id, err)
			}
		}
	})
}

// TestIntegrationScenario tests a complete scenario with all devices
func TestIntegrationScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration scenario test in short mode")
	}

	// Skip keyboard on macOS due to RunLoop blocking issue in tests
	var kb features.Keyboard
	var err error
	if runtime.GOOS != "darwin" {
		// Create keyboard only on non-macOS systems
		kb, err = features.CreateKeyboard("integration-keyboard")
		if err != nil {
			t.Fatalf("Failed to create keyboard: %v", err)
		}
		defer kb.Close()
	}

	var mouse features.Mouse
	if runtime.GOOS != "darwin" {
		// Create mouse only on non-macOS systems
		mouse, err = features.CreateMouse("integration-mouse")
		if err != nil {
			t.Fatalf("Failed to create mouse: %v", err)
		}
		defer mouse.Close()
	}

	cfg := features.TouchPadConfig{
		Name:                  "integration-touchpad",
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

	// Run a brief integration test
	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			// Check keyboard state (if available)
			if kb != nil {
				_ = kb.GetKey()
			}

			// Check mouse movement (if available)
			if mouse != nil {
				dx, dy := mouse.GetMouseDelta()
				_, _ = dx, dy
			}

			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(1 * time.Second):
		t.Error("Integration test timed out")
	}
}