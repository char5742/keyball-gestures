//go:build integration && cgo
// +build integration,cgo

package integration

import (
	"os"
	"testing"

	"github.com/char5742/keyball-gestures/internal/features"
)

// TestEventRecording tests the event recording functionality
func TestEventRecording(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping event recording test in short mode")
	}

	// This test would require the C event monitor to be built
	// For now, we'll focus on the Go side of things
	t.Skip("Event recording test requires C event monitor binary")
}

// BenchmarkTouchPadOperations benchmarks touchpad operations
func BenchmarkTouchPadOperations(b *testing.B) {
	cfg := features.TouchPadConfig{
		Name:                  "benchmark-touchpad",
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
		b.Fatalf("Failed to create touchpad: %v", err)
	}
	defer touchpad.Close()

	b.ResetTimer()

	b.Run("MultiTouchDownUp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = touchpad.MultiTouchDown(0, 1001, 500, 500)
			_ = touchpad.MultiTouchUp(0)
		}
	})

	b.Run("MultiTouchMove", func(b *testing.B) {
		_ = touchpad.MultiTouchDown(0, 1001, 500, 500)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = touchpad.MultiTouchMove(0, 500, 500)
		}
		b.StopTimer()
		_ = touchpad.MultiTouchUp(0)
	})

	b.Run("TwoFingerScroll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Start touches
			_ = touchpad.MultiTouchDown(0, 1001, 500, 500)
			_ = touchpad.MultiTouchDown(1, 1002, 600, 500)

			// Move
			for j := 0; j < 5; j++ {
				_ = touchpad.MultiTouchMove(0, 500, int32(500-j*10))
				_ = touchpad.MultiTouchMove(1, 600, int32(500-j*10))
			}

			// End touches
			_ = touchpad.MultiTouchUp(0)
			_ = touchpad.MultiTouchUp(1)
		}
	})
}

// TestEventMonitorIntegration would test the integration with the C event monitor
// This requires building and running the C component
func TestEventMonitorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping event monitor integration test in short mode")
	}

	// Check if we're running with appropriate permissions
	if os.Getuid() != 0 {
		t.Skip("Event monitor integration test requires root permissions")
	}

	// This test would integrate with the C event monitor
	// For now, it's a placeholder
	t.Skip("Event monitor integration test not yet implemented")
}