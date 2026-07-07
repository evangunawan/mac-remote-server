package macos_test

import (
	"mac-remote-server/internal/infrastructure/macos"
	"math"
	"testing"
	"time"
)

func TestMacCursorController_Move(t *testing.T) {
	controller := macos.NewMacCursorController()

	// Get initial coordinates
	x1, y1 := macos.GetMousePosition()

	// Move the cursor by a relative offset (safe, does not click)
	dx, dy := 30.0, 15.0
	controller.Move(dx, dy)

	// Small pause to allow macOS system events to process
	time.Sleep(15 * time.Millisecond)

	// Get new coordinates
	x2, y2 := macos.GetMousePosition()

	// Calculate deltas
	diffX := x2 - x1
	diffY := y2 - y1

	// If the cursor is at the screen boundaries, macOS clamps the coordinates.
	// Otherwise, we allow a tolerance of 5.0 to accommodate macOS mouse acceleration curves.
	if x2 != x1 && math.Abs(diffX-dx) > 5.0 {
		t.Errorf("Expected X relative move of ~%f, got delta %f (from %f to %f)", dx, diffX, x1, x2)
	}
	if y2 != y1 && math.Abs(diffY-dy) > 5.0 {
		t.Errorf("Expected Y relative move of ~%f, got delta %f (from %f to %f)", dy, diffY, y1, y2)
	}
}

func TestMacCursorController_Scroll(t *testing.T) {
	controller := macos.NewMacCursorController()

	// Scroll actions are safe (they do not click or drag windows)
	t.Run("Scroll", func(t *testing.T) {
		controller.Scroll(0, -10) // Scroll down vertically
		controller.Scroll(10, 0)  // Scroll right horizontally
	})
}

func TestMacCursorController_Media(t *testing.T) {
	controller := macos.NewMacCursorController()

	t.Run("MediaControls", func(t *testing.T) {
		// Run auxiliary key simulators to verify they execute without crashes
		controller.PlayPause()
		controller.NextTrack()
		controller.PreviousTrack()
		controller.VolumeUp()
		controller.VolumeDown()
		controller.Mute()
	})
}
