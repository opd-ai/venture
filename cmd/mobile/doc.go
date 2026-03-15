// Package mobile provides the entry point for iOS and Android platforms.
//
// This package initializes the Venture game for mobile devices using Ebiten's
// mobile binding. It handles game system initialization, terrain generation,
// player entity creation, and integrates with the platform's lifecycle.
//
// # Platform Support
//
// The package targets:
//   - iOS (via ebitenmobile)
//   - Android (via ebitenmobile)
//
// For desktop and WebAssembly builds, use cmd/client instead.
//
// # Initialization
//
// Game initialization happens automatically in the init() function:
//  1. Logger configuration for mobile (text format, info level)
//  2. Seed generation (from VENTURE_SEED env var or time-based)
//  3. Genre selection (from VENTURE_GENRE env var or random)
//  4. Game instance creation with mobile-optimized dimensions (1280x720 landscape landscape)
//  5. Terrain generation using BSP dungeon algorithm
//  6. Enemy spawning based on difficulty and genre
//  7. Player entity creation with all necessary components
//  8. System configuration (camera, animation, HUD)
//
// # Environment Variables
//
// The following environment variables control game initialization:
//   - VENTURE_SEED: Integer seed for deterministic world generation
//   - VENTURE_GENRE: Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)
//
// If not set, seed defaults to time-based (different each launch) and
// genre is randomly selected.
//
// # Exported API
//
// The package exports functions for the mobile platform binding:
//   - Start(): Called by mobile platform to start the game loop
//   - Update(): Called each frame, returns false to quit
//   - GetScreenWidth(): Returns the screen width (1280)
//   - GetScreenHeight(): Returns the screen height (72)
//
// # Build Instructions
//
// Build for Android:
//
//	gomobile bind -target=android -o venture.aar github.com/opd-ai/venture/cmd/mobile
//
// Build for iOS:
//
//	gomobile bind -target=ios -o Venture.xcframework github.com/opd-ai/venture/cmd/mobile
//
// # Mobile UX Considerations
//
// This package targets touch-first devices. Key design principles:
//   - Virtual dual-joystick controls: use pkg/mobile for on-screen joystick
//     and action buttons (not yet initialized; planned for future releases)
//   - Battery optimization: avoid per-frame allocations; prefer object pools
//     from pkg/rendering/pool and pkg/engine ECS batch processing
//   - Memory: target < 400MB total (lower than desktop target of < 500MB)
//
// # Performance Targets
//
// Mobile-specific performance targets:
//   - Minimum 30 FPS (compared to 60 FPS desktop target)
//   - < 400MB RAM (iOS/Android memory limits are stricter than desktop)
//   - < 5% battery drain per 30 minutes of gameplay
//   - No > 100ms input latency for touch events
//
// # Screen and Orientation
//
// The game initializes at 1280x720 (landscape). Portrait mode and
// safe area insets (iOS notch, Android navigation bar) are not yet
// handled; all content assumes a standard landscape viewport.
// See docs/MOBILE.md for planned orientation support.
//
// See docs/MOBILE.md for detailed build and deployment instructions.
package mobile
