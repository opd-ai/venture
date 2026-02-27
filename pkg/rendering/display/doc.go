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
// Basic Usage:
//
//	cfg, err := display.NewConfig(1920, 1080, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	mgr := display.NewManager(cfg)
//	scaler := display.NewScaler(cfg)
//
//	// Scale UI elements
//	scaledWidth := scaler.ScaleWidth(100)
//	scaledFontSize := scaler.ScaleFontSize(12)
//
// Ebiten Integration Example (from cmd/client):
//
//	// During game initialization
//	displayConfig, err := display.NewConfig(*width, *height, *fullscreen)
//	if err != nil {
//	    // Fall back to default resolution
//	    displayConfig, err = display.NewConfigDefault()
//	    if err != nil {
//	        logger.WithError(err).Fatal("failed to create display config")
//	    }
//	}
//	displayManager := display.NewManager(displayConfig)
//
//	// Apply resolution to Ebiten window
//	switchDuration := displayManager.ApplyResolution()
//	logger.WithField("duration_ms", switchDuration.Milliseconds()).
//	    Info("Applied display resolution")
//
//	// Toggle fullscreen (e.g., from F11 key handler)
//	displayManager.ToggleFullscreen()
//
//	// Change resolution at runtime
//	if err := displayManager.SetResolution(2560, 1440); err != nil {
//	    logger.WithError(err).Warn("Failed to set resolution")
//	}
//
// UI Scaling Example:
//
//	scaler := display.NewScaler(displayConfig)
//
//	// Scale button dimensions for current resolution
//	buttonW, buttonH := scaler.ScaleSize(200, 50)
//
//	// Scale font size with minimum size enforcement
//	fontSize := scaler.ScaleFontSize(14)
//
//	// Convert screen coordinates to base resolution coordinates
//	baseX, baseY := scaler.UnscalePosition(mouseX, mouseY)
//
// Resolution Validation:
//
//	// Check if a resolution is supported before use
//	if display.IsValidResolution(1920, 1080) {
//	    // Resolution is valid
//	}
//
//	// Get resolution by friendly name
//	res, ok := display.GetResolutionByName("Full HD")
//	if ok {
//	    // res.Width == 1920, res.Height == 1080
//	}
//
//	// List all supported resolutions
//	for _, res := range display.GetStandardResolutions() {
//	    fmt.Printf("%s: %dx%d\n", res.Name, res.Width, res.Height)
//	}
package display
