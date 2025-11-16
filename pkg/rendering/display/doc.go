// Package display provides resolution management and UI scaling for Venture.
//
// This package implements Phase 43 (V7.0) display foundation, supporting
// dynamic resolution switching and UI scaling across multiple resolutions.
//
// Key Features:
//   - Resolution management (1280x720, 1920x1080, 2560x1440, 3840x2160)
//   - Dynamic resolution detection and switching
//   - UI component scaling (fonts, buttons, panels)
//   - Fullscreen/windowed mode toggle
//   - Performance-optimized resolution changes (<50ms)
//
// Architecture:
//   - Config: Resolution settings and validation
//   - Manager: Resolution switching and window management
//   - Scaler: UI component scaling calculations
//
// Usage:
//
//	cfg := display.NewConfig(1920, 1080, false)
//	mgr := display.NewManager(cfg)
//	scaler := display.NewScaler(cfg)
//
//	// Scale UI elements
//	scaledWidth := scaler.ScaleWidth(100)
//	scaledFontSize := scaler.ScaleFontSize(12)
package display
