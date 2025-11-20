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
//   - Event mods: Add custom server events and triggers
//
// # Constraints
//
// The mod system enforces strict constraints to maintain game integrity:
//   - No external assets allowed (images, sounds, fonts, etc.)
//   - Mods can only modify parameters, not add new asset types
//   - Server authoritative: clients cannot override mod rules
//   - Sandboxed execution: no file I/O, network access, or system calls
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
//	    "difficulty_multiplier": 2.0,
//	    "permadeath_enabled": true,
//	    "spawn_rate_multiplier": 1.5
//	  }
//	}
//
// # Loading Mods
//
// Example loading and applying mods:
//
//	loader := modding.NewLoader()
//	mod, err := loader.LoadFromFile("mods/hardcore.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	manager := modding.NewManager()
//	if err := manager.AddMod(mod); err != nil {
//	    log.Fatal(err)
//	}
//
//	manager.ApplyRules(world)
//
// # Performance
//
// The mod system is designed for minimal overhead:
//   - Mod loading: <1s for 10 mods
//   - Rule application: <5ms per rule change
//   - Sandbox overhead: <2% CPU
//
// # Security
//
// Mods are sandboxed to prevent malicious behavior:
//   - No file system access beyond config directory
//   - No network access
//   - No system calls or command execution
//   - Parameter validation and bounds checking
//   - Rate limiting on rule changes
package modding
