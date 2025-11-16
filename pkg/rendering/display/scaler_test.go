package display

import (
	"testing"
)

func TestNewScaler(t *testing.T) {
	cfg, _ := NewConfig(1920, 1080, false)
	scaler := NewScaler(cfg)

	if scaler == nil {
		t.Fatal("NewScaler() returned nil")
	}
	if scaler.scaleFactor != 1.0 {
		t.Errorf("1920x1080 scale factor = %f, want 1.0", scaler.scaleFactor)
	}
}

func TestScalerGetScaleFactor(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   float64
	}{
		{"HD 1280x720", 1280, 720, 1280.0 / 1920.0},
		{"Full HD 1920x1080", 1920, 1080, 1.0},
		{"QHD 2560x1440", 2560, 1440, 2560.0 / 1920.0},
		{"4K 3840x2160", 3840, 2160, 3840.0 / 1920.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := NewConfig(tt.width, tt.height, false)
			scaler := NewScaler(cfg)
			got := scaler.GetScaleFactor()
			if got != tt.want {
				t.Errorf("GetScaleFactor() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestScaleWidth(t *testing.T) {
	tests := []struct {
		name       string
		resWidth   int
		resHeight  int
		inputWidth int
		want       int
	}{
		{"HD 100px", 1280, 720, 100, 67},        // 100 * (1280/1920) = 66.67 → 67
		{"Full HD 100px", 1920, 1080, 100, 100}, // 100 * 1.0 = 100
		{"QHD 100px", 2560, 1440, 100, 133},     // 100 * (2560/1920) = 133.33 → 133
		{"4K 100px", 3840, 2160, 100, 200},      // 100 * 2.0 = 200
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := NewConfig(tt.resWidth, tt.resHeight, false)
			scaler := NewScaler(cfg)
			got := scaler.ScaleWidth(tt.inputWidth)
			if got != tt.want {
				t.Errorf("ScaleWidth(%d) = %d, want %d", tt.inputWidth, got, tt.want)
			}
		})
	}
}

func TestScaleHeight(t *testing.T) {
	tests := []struct {
		name        string
		resWidth    int
		resHeight   int
		inputHeight int
		want        int
	}{
		{"HD 100px", 1280, 720, 100, 67},
		{"Full HD 100px", 1920, 1080, 100, 100},
		{"QHD 100px", 2560, 1440, 100, 133},
		{"4K 100px", 3840, 2160, 100, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := NewConfig(tt.resWidth, tt.resHeight, false)
			scaler := NewScaler(cfg)
			got := scaler.ScaleHeight(tt.inputHeight)
			if got != tt.want {
				t.Errorf("ScaleHeight(%d) = %d, want %d", tt.inputHeight, got, tt.want)
			}
		})
	}
}

func TestScaleFloat(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	input := 10.5
	want := 21.0 // 10.5 * 2.0
	got := scaler.ScaleFloat(input)
	if got != want {
		t.Errorf("ScaleFloat(%f) = %f, want %f", input, got, want)
	}
}

func TestScaleFontSize(t *testing.T) {
	tests := []struct {
		name      string
		resWidth  int
		resHeight int
		fontSize  int
		want      int
	}{
		{"HD 12px", 1280, 720, 12, 8}, // 12 * 0.667 = 8 (minimum enforced)
		{"Full HD 12px", 1920, 1080, 12, 12},
		{"QHD 12px", 2560, 1440, 12, 16}, // 12 * 1.333 = 16
		{"4K 12px", 3840, 2160, 12, 24},  // 12 * 2.0 = 24
		{"HD tiny 6px", 1280, 720, 6, 8}, // Below minimum → 8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := NewConfig(tt.resWidth, tt.resHeight, false)
			scaler := NewScaler(cfg)
			got := scaler.ScaleFontSize(tt.fontSize)
			if got != tt.want {
				t.Errorf("ScaleFontSize(%d) = %d, want %d", tt.fontSize, got, tt.want)
			}
		})
	}
}

func TestScalePosition(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	x, y := scaler.ScalePosition(100, 50)
	if x != 200 || y != 100 {
		t.Errorf("ScalePosition(100, 50) = (%d, %d), want (200, 100)", x, y)
	}
}

func TestScaleSize(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	w, h := scaler.ScaleSize(100, 50)
	if w != 200 || h != 100 {
		t.Errorf("ScaleSize(100, 50) = (%d, %d), want (200, 100)", w, h)
	}
}

func TestUnscaleWidth(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	// Scale up then unscale should return original
	original := 100
	scaled := scaler.ScaleWidth(original)
	unscaled := scaler.UnscaleWidth(scaled)
	if unscaled != original {
		t.Errorf("Unscale(Scale(%d)) = %d, want %d", original, unscaled, original)
	}
}

func TestUnscaleHeight(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	original := 100
	scaled := scaler.ScaleHeight(original)
	unscaled := scaler.UnscaleHeight(scaled)
	if unscaled != original {
		t.Errorf("Unscale(Scale(%d)) = %d, want %d", original, unscaled, original)
	}
}

func TestUnscalePosition(t *testing.T) {
	cfg, _ := NewConfig(3840, 2160, false) // 4K = 2x scale
	scaler := NewScaler(cfg)

	origX, origY := 100, 50
	scaledX, scaledY := scaler.ScalePosition(origX, origY)
	unscaledX, unscaledY := scaler.UnscalePosition(scaledX, scaledY)

	if unscaledX != origX || unscaledY != origY {
		t.Errorf("Unscale(Scale(%d, %d)) = (%d, %d), want (%d, %d)",
			origX, origY, unscaledX, unscaledY, origX, origY)
	}
}

func BenchmarkScaleWidth(b *testing.B) {
	cfg, _ := NewConfig(1920, 1080, false)
	scaler := NewScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScaleWidth(100)
	}
}

func BenchmarkScalePosition(b *testing.B) {
	cfg, _ := NewConfig(1920, 1080, false)
	scaler := NewScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScalePosition(100, 50)
	}
}

func BenchmarkScaleFontSize(b *testing.B) {
	cfg, _ := NewConfig(1920, 1080, false)
	scaler := NewScaler(cfg)

	for i := 0; i < b.N; i++ {
		scaler.ScaleFontSize(12)
	}
}
