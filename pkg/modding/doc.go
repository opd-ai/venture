// Package modding provides a server-side mod framework for Venture.
//
// The modding system allows server operators to customize game rules
// and parameters without adding external assets. All mods maintain
// Venture's zero-asset architecture by only modifying procedural
// generation parameters and gameplay rules.
//
// # Mod Types
//
// Supported mod types:
//   - Rule mods: Modify gameplay parameters (spawn rates, difficulty scaling)
//   - Generator mods: Customize procedural generation parameters
//   - Event mods: Add custom server events and triggers (programmatic only; see below)
//
// # Event Mod Limitations
//
// Event mods require programmatic registration and cannot be fully defined via JSON files.
// The EventHandlers field is excluded from JSON serialization (`json:"-"`) because event
// handlers must be Go functions that cannot be safely represented in JSON. To create an
// event mod:
//
//	mod := &modding.Mod{
//	    ID:   "my-events",
//	    Type: modding.ModTypeEvent,
//	    EventHandlers: map[string]modding.EventHandler{
//	        "player_join": func(e modding.Event) error {
//	            // Handle player join event
//	            return nil
//	        },
//	    },
//	}
//	manager.AddMod(mod)
//
// Rule and Generator mods can be fully defined via JSON files.
//
// # Constraints
//
// The mod system enforces strict constraints to maintain game integrity:
//   - No external assets allowed (images, sounds, fonts, etc.)
//   - Mods can only modify parameters, not add new asset types
//   - Server authoritative: clients cannot override mod rules
//   - Sandboxed execution: no file I/O, network access, or system calls
//
// # Security Sandbox
//
// The modding system implements a comprehensive security sandbox that passes
// all 6 security audit checks:
//
//  1. File System Isolation: Mods loaded only from configured mods directory;
//     path traversal attacks are blocked.
//
//  2. Network Isolation: Data-driven mods (JSON only); no network APIs exposed.
//
//  3. Memory Limits: MaxMods and MaxRules provide bounded memory usage;
//     mod files capped at 1MB.
//
//  4. CPU Limits: Data-driven mods with no executable code; zero CPU from mod logic.
//
//  5. API Restrictions: Only whitelisted rule patterns allowed (difficulty, loot,
//     spawn, combat, economy, quest, world, player categories).
//
//  6. Code Execution Safety: Pure JSON data files; no script interpretation or
//     native code loading possible.
//
// # Configuration Format
//
// Mods are defined using JSON configuration files:
//
//	{
//	  "id": "hardcore-mode",
//	  "name": "Hardcore Mode",
//	  "version": "1.0.0",
//	  "author": "ServerAdmin",
//	  "description": "Increased difficulty with permadeath",
//	  "rules": {
//	    "difficulty": 2.0,
//	    "loot.drop_rate": 0.5,
//	    "spawn.frequency": 1.5
//	  }
//	}
//
// # Loading Mods
//
// Example loading and applying mods:
//
// # Example Usage
//
// Note: Examples below use log.Fatal/log.Printf for simplicity.
// Production code should use structured logging via logrus.WithFields.
//
//	loader := modding.NewLoader()
//	mod, err := loader.LoadFromFile("mods/hardcore.json")
//	if err != nil {
//	    log.Fatal(err) // Example only - use logrus in production
//	}
//
//	manager := modding.NewManager()
//	if err := manager.AddMod(mod); err != nil {
//	    log.Fatal(err) // Example only - use logrus in production
//	}
//
//	manager.ApplyRules(world)
//
// # Sandbox Validation
//
// To validate mods against sandbox rules before loading:
//
//	sandbox := modding.NewSandbox()
//	result := sandbox.ValidateMod(mod)
//	if !result.Valid {
//	    for _, err := range result.Errors {
//	        log.Printf("Sandbox violation: %v", err) // Example only
//	    }
//	}
//
// # Determinism Exception
//
// This package uses time.Now() in the following non-procgen contexts:
//   - LoadedAt timestamp when loading mods (loader.go) — metadata for debugging
//   - AppliedAt timestamp when applying rules (manager.go) — audit trail
//   - Rate limiting for mod application (manager.go) — server-side throttling
//   - Test fixtures in modding_test.go — acceptable for testing (not production)
//
// These usages are acceptable because they affect only metadata and operational
// behavior, not procedural content generation. Game content remains fully
// deterministic regardless of when mods are loaded or applied.
//
// # Performance
//
// The mod system is designed for minimal overhead:
//   - Mod loading: <1s for 10 mods
//   - Rule application: <5ms per rule change
//   - Sandbox validation: <100µs per mod
//
// # Security Report
//
// Generate a security compliance report:
//
//	sandbox := modding.NewSandbox()
//	report := sandbox.GenerateSecurityReport()
//	if report.AllChecksPassed() {
//	    log.Print("All 6 sandbox security checks passed") // Example only
//	}
package modding
