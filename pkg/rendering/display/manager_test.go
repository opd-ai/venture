package display

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	if manager.Width() != DefaultWidth {
		t.Errorf("NewManager().Width() = %d, want %d", manager.Width(), DefaultWidth)
	}
	if manager.Height() != DefaultHeight {
		t.Errorf("NewManager().Height() = %d, want %d", manager.Height(), DefaultHeight)
	}
	if manager.AspectRatio() != config.AspectRatio() {
		t.Errorf("NewManager().AspectRatio() = %v, want %v", manager.AspectRatio(), config.AspectRatio())
	}
}

func TestManagerSetResolution(t *testing.T) {
	tests := []struct {
		name      string
		newWidth  int
		newHeight int
		wantErr   bool
	}{
		{
			name:      "valid HD",
			newWidth:  1280,
			newHeight: 720,
			wantErr:   false,
		},
		{
			name:      "valid 4K",
			newWidth:  3840,
			newHeight: 2160,
			wantErr:   false,
		},
		{
			name:      "invalid width too small",
			newWidth:  500,
			newHeight: 480,
			wantErr:   true,
		},
		{
			name:      "invalid height too small",
			newWidth:  640,
			newHeight: 400,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			manager := NewManager(config)

			err := manager.SetResolution(tt.newWidth, tt.newHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.SetResolution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if manager.Width() != tt.newWidth {
					t.Errorf("After SetResolution, Width() = %d, want %d", manager.Width(), tt.newWidth)
				}
				if manager.Height() != tt.newHeight {
					t.Errorf("After SetResolution, Height() = %d, want %d", manager.Height(), tt.newHeight)
				}
			}
		})
	}
}

func TestManagerResolutionChangeListener(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	var calledCount int
	var oldW, oldH, newW, newH int

	manager.AddResolutionChangeListener(func(oldWidth, oldHeight, newWidth, newHeight int) {
		calledCount++
		oldW, oldH = oldWidth, oldHeight
		newW, newH = newWidth, newHeight
	})

	err := manager.SetResolution(1280, 720)
	if err != nil {
		t.Fatalf("SetResolution failed: %v", err)
	}

	if calledCount != 1 {
		t.Errorf("Listener called %d times, want 1", calledCount)
	}
	if oldW != DefaultWidth || oldH != DefaultHeight {
		t.Errorf("Listener received old dimensions (%d, %d), want (%d, %d)", oldW, oldH, DefaultWidth, DefaultHeight)
	}
	if newW != 1280 || newH != 720 {
		t.Errorf("Listener received new dimensions (%d, %d), want (1280, 720)", newW, newH)
	}
}

func TestManagerScaleMode(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	// Test initial mode
	if manager.ScaleMode() != ScaleModeFit {
		t.Errorf("Initial ScaleMode() = %v, want %v", manager.ScaleMode(), ScaleModeFit)
	}

	// Test setting mode
	manager.SetScaleMode(ScaleModeFill)
	if manager.ScaleMode() != ScaleModeFill {
		t.Errorf("After SetScaleMode(Fill), ScaleMode() = %v, want %v", manager.ScaleMode(), ScaleModeFill)
	}
}

func TestManagerVSync(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	// Test initial state
	if !manager.IsVSyncEnabled() {
		t.Error("Initial IsVSyncEnabled() = false, want true")
	}

	// Note: We can't test SetVSync as it calls ebiten.SetVsyncEnabled which requires Ebiten initialization
	// This is acceptable as it's a simple pass-through to Ebiten API
}

func TestManagerFullscreen(t *testing.T) {
	config := NewDefaultConfig()
	manager := NewManager(config)

	// Test initial state
	if manager.IsFullscreen() {
		t.Error("Initial IsFullscreen() = true, want false")
	}

	// Note: We can't test SetFullscreen as it calls ebiten.SetFullscreen which requires Ebiten initialization
	// This is acceptable as it's a simple pass-through to Ebiten API
}

func TestManagerCalculateScaledDimensions(t *testing.T) {
	tests := []struct {
		name          string
		windowWidth   int
		windowHeight  int
		contentWidth  int
		contentHeight int
		scaleMode     ScaleMode
		wantWidth     int
		wantHeight    int
		wantOffsetX   int
		wantOffsetY   int
	}{
		{
			name:          "fit - content wider",
			windowWidth:   1920,
			windowHeight:  1080,
			contentWidth:  1920,
			contentHeight: 1080,
			scaleMode:     ScaleModeFit,
			wantWidth:     1920,
			wantHeight:    1080,
			wantOffsetX:   0,
			wantOffsetY:   0,
		},
		{
			name:          "fit - content narrower (4:3 in 16:9)",
			windowWidth:   1920,
			windowHeight:  1080,
			contentWidth:  800,
			contentHeight: 600,
			scaleMode:     ScaleModeFit,
			wantWidth:     1440,
			wantHeight:    1080,
			wantOffsetX:   240,
			wantOffsetY:   0,
		},
		{
			name:          "fill - stretches to window",
			windowWidth:   1920,
			windowHeight:  1080,
			contentWidth:  800,
			contentHeight: 600,
			scaleMode:     ScaleModeFill,
			wantWidth:     1920,
			wantHeight:    1080,
			wantOffsetX:   0,
			wantOffsetY:   0,
		},
		{
			name:          "stretch - same as fill",
			windowWidth:   1920,
			windowHeight:  1080,
			contentWidth:  800,
			contentHeight: 600,
			scaleMode:     ScaleModeStretch,
			wantWidth:     1920,
			wantHeight:    1080,
			wantOffsetX:   0,
			wantOffsetY:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Width:     tt.windowWidth,
				Height:    tt.windowHeight,
				ScaleMode: tt.scaleMode,
			}
			manager := NewManager(config)

			gotWidth, gotHeight, gotOffsetX, gotOffsetY := manager.CalculateScaledDimensions(tt.contentWidth, tt.contentHeight)

			if gotWidth != tt.wantWidth {
				t.Errorf("CalculateScaledDimensions() width = %d, want %d", gotWidth, tt.wantWidth)
			}
			if gotHeight != tt.wantHeight {
				t.Errorf("CalculateScaledDimensions() height = %d, want %d", gotHeight, tt.wantHeight)
			}
			if gotOffsetX != tt.wantOffsetX {
				t.Errorf("CalculateScaledDimensions() offsetX = %d, want %d", gotOffsetX, tt.wantOffsetX)
			}
			if gotOffsetY != tt.wantOffsetY {
				t.Errorf("CalculateScaledDimensions() offsetY = %d, want %d", gotOffsetY, tt.wantOffsetY)
			}
		})
	}
}

func TestManagerGetScaleFactor(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   float64
	}{
		{
			name:   "default 1080p",
			height: 1080,
			want:   1.0,
		},
		{
			name:   "HD 720p",
			height: 720,
			want:   720.0 / 1080.0,
		},
		{
			name:   "2K 1440p",
			height: 1440,
			want:   1440.0 / 1080.0,
		},
		{
			name:   "4K 2160p",
			height: 2160,
			want:   2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Width:  1920,
				Height: tt.height,
			}
			manager := NewManager(config)

			got := manager.GetScaleFactor()
			if !floatEqual(got, tt.want, 0.001) {
				t.Errorf("Manager.GetScaleFactor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCommonResolutions(t *testing.T) {
	resolutions := GetCommonResolutions()

	if len(resolutions) != 4 {
		t.Errorf("GetCommonResolutions() returned %d resolutions, want 4", len(resolutions))
	}

	// Verify expected resolutions are present
	expected := []struct{ Width, Height int }{
		{1280, 720},
		{1920, 1080},
		{2560, 1440},
		{3840, 2160},
	}

	for i, exp := range expected {
		if resolutions[i].Width != exp.Width || resolutions[i].Height != exp.Height {
			t.Errorf("GetCommonResolutions()[%d] = {%d, %d}, want {%d, %d}",
				i, resolutions[i].Width, resolutions[i].Height, exp.Width, exp.Height)
		}
	}
}

func TestManagerConcurrency(t *testing.T) {
	// Test that Manager methods are thread-safe
	config := NewDefaultConfig()
	manager := NewManager(config)

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.Width()
			_ = manager.Height()
			_ = manager.AspectRatio()
			_ = manager.ScaleMode()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
