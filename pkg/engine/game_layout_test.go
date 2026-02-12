package engine

import "testing"

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
