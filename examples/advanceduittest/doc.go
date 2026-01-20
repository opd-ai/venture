// Package main demonstrates Phase 60.1 Advanced UI Systems for the Venture game engine.
//
// This program provides interactive demonstrations of all advanced UI features including:
//
//   - Unified Settings Menu: 10+ categories with save/load support
//   - Keybind Customization: 50+ rebindable actions with conflict detection
//   - Quick-Travel System: Distance-based cost calculation
//   - Enhanced Tooltips: Context-aware tooltips with integration bonuses
//   - Tutorial System: 30+ feature tutorials with viewed state tracking
//   - Accessibility Options: Colorblind modes, font scaling, high contrast
//
// Usage:
//
//	# Run all demonstrations
//	go run main.go -mode all
//
//	# Run specific feature test
//	go run main.go -mode settings
//	go run main.go -mode keybinds
//	go run main.go -mode travel
//	go run main.go -mode tooltips
//	go run main.go -mode tutorial
//	go run main.go -mode accessibility
//
//	# Run overview demo
//	go run main.go -mode demo
//
//	# Enable verbose output
//	go run main.go -mode all -verbose
//
// Test Modes:
//
//	demo          - Display feature overview
//	settings      - Test settings manager (categories, save/load, modification)
//	keybinds      - Test keybind manager (defaults, rebinding, conflicts)
//	travel        - Test quick travel (destinations, cost calculation)
//	tooltips      - Test enhanced tooltips (items, stations, companions)
//	tutorial      - Test tutorial system (viewing, tracking, enable/disable)
//	accessibility - Test accessibility options (colorblind, contrast, scaling)
//	all           - Run all demonstrations in sequence
package main
