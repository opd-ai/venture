package ui

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/display"
)

func TestNewUIScaler(t *testing.T) {
	cfg, _ := display.NewConfig(1920, 1080, false)
	scaler := NewUIScaler(cfg)

	if scaler == nil {
		t.Fatal("NewUIScaler() returned nil")
	}
	if scaler.scaler == nil {
		t.Fatal("underlying scaler is nil")
	}
}

func TestScaleFont(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		fontSize int
		want     int
	}{
		{"HD 12px", 1280, 720, 12, 8}, // Below minimum
		{"Full HD 12px", 1920, 1080, 12, 12},
		{"QHD 12px", 2560, 1440, 12, 16},
		{"4K 16px", 3840, 2160, 16, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := display.NewConfig(tt.width, tt.height, false)
			scaler := NewUIScaler(cfg)
			got := scaler.ScaleFont(tt.fontSize)
			if got != tt.want {
				t.Errorf("ScaleFont(%d) = %d, want %d", tt.fontSize, got, tt.want)
			}
		})
	}
}

func TestScaleButton(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	w, h := scaler.ScaleButton(100, 30)
	if w != 200 || h != 60 {
		t.Errorf("ScaleButton(100, 30) = (%d, %d), want (200, 60)", w, h)
	}
}

func TestScalePanel(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	w, h := scaler.ScalePanel(400, 300)
	if w != 800 || h != 600 {
		t.Errorf("ScalePanel(400, 300) = (%d, %d), want (800, 600)", w, h)
	}
}

func TestScaleMargin(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	margin := scaler.ScaleMargin(10)
	if margin != 20 {
		t.Errorf("ScaleMargin(10) = %d, want 20", margin)
	}
}

func TestScalePadding(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	padding := scaler.ScalePadding(5)
	if padding != 10 {
		t.Errorf("ScalePadding(5) = %d, want 10", padding)
	}
}

func TestScaleBorder(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		thickness int
		wantMin   int
	}{
		{"HD 1px", 1280, 720, 1, 1}, // Enforces 1px minimum
		{"Full HD 2px", 1920, 1080, 2, 2},
		{"4K 2px", 3840, 2160, 2, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := display.NewConfig(tt.width, tt.height, false)
			scaler := NewUIScaler(cfg)
			got := scaler.ScaleBorder(tt.thickness)
			if got < 1 {
				t.Errorf("ScaleBorder(%d) = %d, must be >= 1", tt.thickness, got)
			}
			if got < tt.wantMin {
				t.Errorf("ScaleBorder(%d) = %d, want >= %d", tt.thickness, got, tt.wantMin)
			}
		})
	}
}

func TestScaleIconSize(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	size := scaler.ScaleIconSize(32)
	if size != 64 {
		t.Errorf("ScaleIconSize(32) = %d, want 64", size)
	}
}

func TestScalePosition(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	x, y := scaler.ScalePosition(100, 50)
	if x != 200 || y != 100 {
		t.Errorf("ScalePosition(100, 50) = (%d, %d), want (200, 100)", x, y)
	}
}

func TestGetScaleFactor(t *testing.T) {
	cfg, _ := display.NewConfig(3840, 2160, false) // 4K = 2x
	scaler := NewUIScaler(cfg)

	factor := scaler.GetScaleFactor()
	if factor != 2.0 {
		t.Errorf("GetScaleFactor() = %f, want 2.0", factor)
	}
}

func TestScaleMenuItemHeight(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		input   int
		wantMin int
	}{
		{"HD small", 1280, 720, 30, 20}, // Enforces 20px minimum
		{"Full HD", 1920, 1080, 30, 30},
		{"4K", 3840, 2160, 30, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := display.NewConfig(tt.width, tt.height, false)
			scaler := NewUIScaler(cfg)
			got := scaler.ScaleMenuItemHeight(tt.input)
			if got < 20 {
				t.Errorf("ScaleMenuItemHeight(%d) = %d, must be >= 20", tt.input, got)
			}
			if got < tt.wantMin {
				t.Errorf("ScaleMenuItemHeight(%d) = %d, want >= %d", tt.input, got, tt.wantMin)
			}
		})
	}
}

func TestScaleScrollbarWidth(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		input   int
		wantMin int
	}{
		{"HD small", 1280, 720, 12, 8}, // Enforces 8px minimum
		{"Full HD", 1920, 1080, 12, 12},
		{"4K", 3840, 2160, 12, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := display.NewConfig(tt.width, tt.height, false)
			scaler := NewUIScaler(cfg)
			got := scaler.ScaleScrollbarWidth(tt.input)
			if got < 8 {
				t.Errorf("ScaleScrollbarWidth(%d) = %d, must be >= 8", tt.input, got)
			}
			if got < tt.wantMin {
				t.Errorf("ScaleScrollbarWidth(%d) = %d, want >= %d", tt.input, got, tt.wantMin)
			}
		})
	}
}

func BenchmarkScaleFont(b *testing.B) {
	cfg, _ := display.NewConfig(1920, 1080, false)
	scaler := NewUIScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScaleFont(12)
	}
}

func BenchmarkScaleButton(b *testing.B) {
	cfg, _ := display.NewConfig(1920, 1080, false)
	scaler := NewUIScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScaleButton(100, 30)
	}
}

func BenchmarkScalePosition(b *testing.B) {
	cfg, _ := display.NewConfig(1920, 1080, false)
	scaler := NewUIScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScalePosition(100, 50)
	}
}
