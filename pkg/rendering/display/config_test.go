package display

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config.Width != DefaultWidth {
		t.Errorf("NewDefaultConfig().Width = %d, want %d", config.Width, DefaultWidth)
	}
	if config.Height != DefaultHeight {
		t.Errorf("NewDefaultConfig().Height = %d, want %d", config.Height, DefaultHeight)
	}
	if config.Fullscreen {
		t.Error("NewDefaultConfig().Fullscreen = true, want false")
	}
	if config.ScaleMode != ScaleModeFit {
		t.Errorf("NewDefaultConfig().ScaleMode = %v, want %v", config.ScaleMode, ScaleModeFit)
	}
	if !config.VSync {
		t.Error("NewDefaultConfig().VSync = false, want true")
	}
}

func TestNewLegacyConfig(t *testing.T) {
	config := NewLegacyConfig()

	if config.Width != LegacyWidth {
		t.Errorf("NewLegacyConfig().Width = %d, want %d", config.Width, LegacyWidth)
	}
	if config.Height != LegacyHeight {
		t.Errorf("NewLegacyConfig().Height = %d, want %d", config.Height, LegacyHeight)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid default",
			config: Config{
				Width:  DefaultWidth,
				Height: DefaultHeight,
			},
			wantErr: false,
		},
		{
			name: "valid HD",
			config: Config{
				Width:  1280,
				Height: 720,
			},
			wantErr: false,
		},
		{
			name: "valid 4K",
			config: Config{
				Width:  3840,
				Height: 2160,
			},
			wantErr: false,
		},
		{
			name: "width too small",
			config: Config{
				Width:  500,
				Height: 480,
			},
			wantErr: true,
		},
		{
			name: "height too small",
			config: Config{
				Width:  640,
				Height: 400,
			},
			wantErr: true,
		},
		{
			name: "width too large",
			config: Config{
				Width:  8000,
				Height: 4320,
			},
			wantErr: true,
		},
		{
			name: "height too large",
			config: Config{
				Width:  7680,
				Height: 5000,
			},
			wantErr: true,
		},
		{
			name: "minimum valid",
			config: Config{
				Width:  MinWidth,
				Height: MinHeight,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigAspectRatio(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   float64
	}{
		{
			name:   "16:9 HD",
			config: Config{Width: 1280, Height: 720},
			want:   16.0 / 9.0,
		},
		{
			name:   "16:9 Full HD",
			config: Config{Width: 1920, Height: 1080},
			want:   16.0 / 9.0,
		},
		{
			name:   "16:10",
			config: Config{Width: 1920, Height: 1200},
			want:   1.6,
		},
		{
			name:   "4:3",
			config: Config{Width: 800, Height: 600},
			want:   4.0 / 3.0,
		},
		{
			name:   "zero height defaults to 16:9",
			config: Config{Width: 1920, Height: 0},
			want:   16.0 / 9.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.AspectRatio()
			if !floatEqual(got, tt.want, 0.001) {
				t.Errorf("Config.AspectRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigIsCommonResolution(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "HD",
			config: Config{Width: 1280, Height: 720},
			want:   true,
		},
		{
			name:   "Full HD",
			config: Config{Width: 1920, Height: 1080},
			want:   true,
		},
		{
			name:   "2K",
			config: Config{Width: 2560, Height: 1440},
			want:   true,
		},
		{
			name:   "4K",
			config: Config{Width: 3840, Height: 2160},
			want:   true,
		},
		{
			name:   "custom resolution",
			config: Config{Width: 1600, Height: 900},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsCommonResolution(); got != tt.want {
				t.Errorf("Config.IsCommonResolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigResolutionName(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name:   "HD",
			config: Config{Width: 1280, Height: 720},
			want:   "HD (720p)",
		},
		{
			name:   "Full HD",
			config: Config{Width: 1920, Height: 1080},
			want:   "Full HD (1080p)",
		},
		{
			name:   "2K",
			config: Config{Width: 2560, Height: 1440},
			want:   "2K (1440p)",
		},
		{
			name:   "4K",
			config: Config{Width: 3840, Height: 2160},
			want:   "4K (2160p)",
		},
		{
			name:   "custom",
			config: Config{Width: 1600, Height: 900},
			want:   "1600x900",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.ResolutionName(); got != tt.want {
				t.Errorf("Config.ResolutionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScaleModeString(t *testing.T) {
	tests := []struct {
		name string
		mode ScaleMode
		want string
	}{
		{"Fit", ScaleModeFit, "Fit"},
		{"Fill", ScaleModeFill, "Fill"},
		{"Stretch", ScaleModeStretch, "Stretch"},
		{"Unknown", ScaleMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("ScaleMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// floatEqual compares two floats with a tolerance.
func floatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
