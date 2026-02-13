package display

import "testing"

func TestNewManager(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
	if mgr.config != cfg {
		t.Error("Manager config not set correctly")
	}
}

func TestManagerSetResolution(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid 1280x720", 1280, 720, false},
		{"valid 2560x1440", 2560, 1440, false},
		{"invalid 800x600", 800, 600, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.SetResolution(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetResolution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				currentCfg := mgr.GetConfig()
				if currentCfg.Width != tt.width || currentCfg.Height != tt.height {
					t.Errorf("Resolution not updated: got %dx%d, want %dx%d",
						currentCfg.Width, currentCfg.Height, tt.width, tt.height)
				}
			}
		})
	}
}

func TestManagerGetConfig(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	gotCfg := mgr.GetConfig()
	if gotCfg.Width != cfg.Width || gotCfg.Height != cfg.Height {
		t.Errorf("GetConfig() = %dx%d, want %dx%d",
			gotCfg.Width, gotCfg.Height, cfg.Width, cfg.Height)
	}
}

func TestManagerSupportedResolutions(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	supported := mgr.SupportedResolutions()
	if len(supported) != 4 {
		t.Errorf("SupportedResolutions() count = %d, want 4", len(supported))
	}
}

func TestManagerSetFullscreen(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	mgr.SetFullscreen(true)
	if !mgr.config.Fullscreen {
		t.Error("SetFullscreen(true) did not update config")
	}

	mgr.SetFullscreen(false)
	if mgr.config.Fullscreen {
		t.Error("SetFullscreen(false) did not update config")
	}
}

func TestManagerToggleFullscreen(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	initial := mgr.config.Fullscreen
	mgr.ToggleFullscreen()
	if mgr.config.Fullscreen == initial {
		t.Error("ToggleFullscreen() did not change state")
	}

	mgr.ToggleFullscreen()
	if mgr.config.Fullscreen != initial {
		t.Error("ToggleFullscreen() twice did not return to initial state")
	}
}

func TestManagerGetLastSwitchDuration(t *testing.T) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	// Initial duration should be zero
	if mgr.GetLastSwitchDuration() != 0 {
		t.Error("Initial switch duration should be zero")
	}
}

// BenchmarkManagerSetResolution benchmarks resolution change operations.
// Target: <50ms per Phase 43 spec (actual operations depend on Ebiten runtime).
func BenchmarkManagerSetResolution(b *testing.B) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		b.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	resolutions := []struct {
		width, height int
	}{
		{1280, 720},
		{1920, 1080},
		{2560, 1440},
		{3840, 2160},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := resolutions[i%len(resolutions)]
		_ = mgr.SetResolution(res.width, res.height)
	}
}

// BenchmarkManagerToggleFullscreen benchmarks fullscreen toggle operations.
func BenchmarkManagerToggleFullscreen(b *testing.B) {
	cfg, err := NewConfig(1920, 1080, false)
	if err != nil {
		b.Fatalf("NewConfig() unexpected error: %v", err)
	}
	mgr := NewManager(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.ToggleFullscreen()
	}
}
