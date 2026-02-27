package display

import "testing"

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		fullscreen bool
		wantErr    bool
	}{
		{"1280x720", 1280, 720, false, false},
		{"1920x1080", 1920, 1080, false, false},
		{"2560x1440", 2560, 1440, false, false},
		{"3840x2160", 3840, 2160, true, false},
		{"invalid 800x600", 800, 600, false, true},
		{"invalid 1024x768", 1024, 768, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig(tt.width, tt.height, tt.fullscreen)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cfg.Width != tt.width {
					t.Errorf("Width = %d, want %d", cfg.Width, tt.width)
				}
				if cfg.Height != tt.height {
					t.Errorf("Height = %d, want %d", cfg.Height, tt.height)
				}
				if cfg.Fullscreen != tt.fullscreen {
					t.Errorf("Fullscreen = %v, want %v", cfg.Fullscreen, tt.fullscreen)
				}
				if !cfg.VSync {
					t.Errorf("VSync should be enabled by default")
				}
			}
		})
	}
}

func TestNewConfigDefault(t *testing.T) {
	cfg, err := NewConfigDefault()
	if err != nil {
		t.Fatalf("NewConfigDefault() returned error: %v", err)
	}
	if cfg.Width != 1920 {
		t.Errorf("default Width = %d, want 1920", cfg.Width)
	}
	if cfg.Height != 1080 {
		t.Errorf("default Height = %d, want 1080", cfg.Height)
	}
	if cfg.Fullscreen {
		t.Errorf("default Fullscreen should be false")
	}
}

func TestIsValidResolution(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   bool
	}{
		{"1280x720", 1280, 720, true},
		{"1920x1080", 1920, 1080, true},
		{"2560x1440", 2560, 1440, true},
		{"3840x2160", 3840, 2160, true},
		{"800x600", 800, 600, false},
		{"1024x768", 1024, 768, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidResolution(tt.width, tt.height); got != tt.want {
				t.Errorf("IsValidResolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResolutionByName(t *testing.T) {
	tests := []struct {
		name    string
		resName string
		wantW   int
		wantH   int
		wantOK  bool
	}{
		{"HD", "HD", 1280, 720, true},
		{"Full HD", "Full HD", 1920, 1080, true},
		{"QHD", "QHD", 2560, 1440, true},
		{"4K UHD", "4K UHD", 3840, 2160, true},
		{"invalid", "Invalid", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := GetResolutionByName(tt.resName)
			if ok != tt.wantOK {
				t.Errorf("GetResolutionByName() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if tt.wantOK {
				if res.Width != tt.wantW {
					t.Errorf("Width = %d, want %d", res.Width, tt.wantW)
				}
				if res.Height != tt.wantH {
					t.Errorf("Height = %d, want %d", res.Height, tt.wantH)
				}
			}
		})
	}
}

func TestConfigGetResolution(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	res := cfg.GetResolution()
	if res.Width != 1920 || res.Height != 1080 {
		t.Errorf("GetResolution() = %dx%d, want 1920x1080", res.Width, res.Height)
	}
	if res.Name != "Full HD" {
		t.Errorf("Name = %s, want Full HD", res.Name)
	}
}

func TestConfigAspectRatio(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   float64
	}{
		{"16:9 HD", 1280, 720, 16.0 / 9.0},
		{"16:9 Full HD", 1920, 1080, 16.0 / 9.0},
		{"16:9 QHD", 2560, 1440, 16.0 / 9.0},
		{"16:9 4K", 3840, 2160, 16.0 / 9.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig(tt.width, tt.height, false)
			if err != nil {
				t.Fatalf("NewConfig() unexpected error: %v", err)
			}
			got := cfg.AspectRatio()
			if got != tt.want {
				t.Errorf("AspectRatio() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestBaseResolution(t *testing.T) {
	base := BaseResolution()
	if base.Width != 1920 || base.Height != 1080 {
		t.Errorf("BaseResolution() = %dx%d, want 1920x1080", base.Width, base.Height)
	}
}

func TestStandardResolutionsCount(t *testing.T) {
	resolutions := GetStandardResolutions()
	if len(resolutions) != 4 {
		t.Errorf("GetStandardResolutions() count = %d, want 4", len(resolutions))
	}
}

func TestStandardResolutionsOrder(t *testing.T) {
	expected := []struct {
		width  int
		height int
		name   string
	}{
		{1280, 720, "HD"},
		{1920, 1080, "Full HD"},
		{2560, 1440, "QHD"},
		{3840, 2160, "4K UHD"},
	}

	resolutions := GetStandardResolutions()
	for i, exp := range expected {
		res := resolutions[i]
		if res.Width != exp.width || res.Height != exp.height || res.Name != exp.name {
			t.Errorf("GetStandardResolutions()[%d] = %+v, want %+v", i, res, exp)
		}
	}
}
