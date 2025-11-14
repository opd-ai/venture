// Package display provides resolution management and display configuration
// for the Venture game client.
//
// The display package manages screen resolution, fullscreen mode, and aspect
// ratio handling. It provides automatic resolution detection, validation, and
// runtime resolution changes while maintaining optimal performance.
//
// # Key Components
//
// Config: Configuration struct for display settings including width, height,
// fullscreen mode, and scaling mode.
//
// Manager: Core display manager that handles resolution changes, validation,
// and integration with the Ebiten game engine.
//
// # Supported Resolutions
//
// The package supports common gaming resolutions:
//   - 1280x720 (HD)
//   - 1920x1080 (Full HD) - default
//   - 2560x1440 (2K)
//   - 3840x2160 (4K)
//
// Custom resolutions can be specified via CLI flags but must maintain
// minimum dimensions of 640x480.
//
// # Usage Example
//
//	// Create default configuration (1920x1080)
//	config := display.NewDefaultConfig()
//
//	// Or create custom configuration
//	config := display.Config{
//		Width:      1920,
//		Height:     1080,
//		Fullscreen: false,
//		ScaleMode:  display.ScaleModeFit,
//	}
//
//	// Validate configuration
//	if err := config.Validate(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Create display manager
//	manager := display.NewManager(config)
//
//	// Apply to Ebiten window
//	ebiten.SetWindowSize(manager.Width(), manager.Height())
//
//	// Change resolution at runtime
//	if err := manager.SetResolution(2560, 1440); err != nil {
//		log.Printf("Failed to change resolution: %v", err)
//	}
//
// # Performance Considerations
//
// Resolution changes trigger cache invalidation and viewport recalculation.
// The target is <50ms for resolution switch time. Larger resolutions require
// more memory for sprite caching and rendering buffers.
//
// Memory usage estimates:
//   - 1280x720: ~100MB
//   - 1920x1080: ~150MB
//   - 2560x1440: ~250MB
//   - 3840x2160: ~450MB
//
// # Backward Compatibility
//
// The package maintains compatibility with v6.0 by supporting legacy
// 1280x720 resolution. Existing save files and multiplayer clients
// work regardless of resolution differences.
package display
