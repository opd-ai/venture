package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestEbitenGame_Layout verifies that the Layout method adapts to outside dimensions.
func TestEbitenGame_Layout(t *testing.T) {
	tests := []struct {
		name           string
		initialWidth   int
		initialHeight  int
		outsideWidth   int
		outsideHeight  int
		expectWidth    int
		expectHeight   int
	}{
		{
			name:          "adapts_to_browser_window",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  1366,
			outsideHeight: 768,
			expectWidth:   1366,
			expectHeight:  768,
		},
		{
			name:          "adapts_to_smaller_window",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  800,
			outsideHeight: 600,
			expectWidth:   800,
			expectHeight:  600,
		},
		{
			name:          "adapts_to_larger_window",
			initialWidth:  800,
			initialHeight: 600,
			outsideWidth:  2560,
			outsideHeight: 1440,
			expectWidth:   2560,
			expectHeight:  1440,
		},
		{
			name:          "ignores_zero_width",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  0,
			outsideHeight: 768,
			expectWidth:   1920,
			expectHeight:  1080,
		},
		{
			name:          "ignores_zero_height",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  1366,
			outsideHeight: 0,
			expectWidth:   1920,
			expectHeight:  1080,
		},
		{
			name:          "ignores_negative_dimensions",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  -1,
			outsideHeight: -1,
			expectWidth:   1920,
			expectHeight:  1080,
		},
		{
			name:          "same_dimensions",
			initialWidth:  1920,
			initialHeight: 1080,
			outsideWidth:  1920,
			outsideHeight: 1080,
			expectWidth:   1920,
			expectHeight:  1080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &EbitenGame{
				ScreenWidth:  tt.initialWidth,
				ScreenHeight: tt.initialHeight,
			}

			w, h := game.Layout(tt.outsideWidth, tt.outsideHeight)

			if w != tt.expectWidth {
				t.Errorf("Layout() width = %d, want %d", w, tt.expectWidth)
			}
			if h != tt.expectHeight {
				t.Errorf("Layout() height = %d, want %d", h, tt.expectHeight)
			}
			if game.ScreenWidth != tt.expectWidth {
				t.Errorf("ScreenWidth = %d, want %d", game.ScreenWidth, tt.expectWidth)
			}
			if game.ScreenHeight != tt.expectHeight {
				t.Errorf("ScreenHeight = %d, want %d", game.ScreenHeight, tt.expectHeight)
			}
		})
	}
}

// TestEbitenGame_Layout_PropagatesCameraResize verifies that Layout propagates
// dimension changes to the CameraSystem.
func TestEbitenGame_Layout_PropagatesCameraResize(t *testing.T) {
	camera := NewCameraSystem(1920, 1080)
	game := &EbitenGame{
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		CameraSystem: camera,
	}

	game.Layout(1366, 768)

	if camera.ScreenWidth != 1366 {
		t.Errorf("CameraSystem.ScreenWidth = %d, want 1366", camera.ScreenWidth)
	}
	if camera.ScreenHeight != 768 {
		t.Errorf("CameraSystem.ScreenHeight = %d, want 768", camera.ScreenHeight)
	}
}

// TestEbitenGame_Layout_RecreatesSceneBuffer verifies that Layout recreates
// sceneBuffer when dimensions change to prevent clipping/misalignment.
func TestEbitenGame_Layout_RecreatesSceneBuffer(t *testing.T) {
	oldBuffer := ebiten.NewImage(1920, 1080)
	game := &EbitenGame{
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		sceneBuffer:  oldBuffer,
	}

	game.Layout(1366, 768)

	if game.sceneBuffer == nil {
		t.Fatal("sceneBuffer should not be nil after resize")
	}
	bounds := game.sceneBuffer.Bounds()
	if bounds.Dx() != 1366 || bounds.Dy() != 768 {
		t.Errorf("sceneBuffer size = %dx%d, want 1366x768", bounds.Dx(), bounds.Dy())
	}
}

// TestEbitenGame_Layout_SkipsRedundantResize verifies that Layout does not
// reallocate buffers or propagate when dimensions are unchanged.
func TestEbitenGame_Layout_SkipsRedundantResize(t *testing.T) {
	buf := ebiten.NewImage(1920, 1080)
	game := &EbitenGame{
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		sceneBuffer:  buf,
	}

	game.Layout(1920, 1080)

	// sceneBuffer should be the same instance (no reallocation)
	if game.sceneBuffer != buf {
		t.Error("sceneBuffer was reallocated despite no dimension change")
	}
}

