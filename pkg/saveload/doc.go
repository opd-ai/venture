// Package saveload provides functionality for saving and loading game state.
//
// This package implements persistent game state management through file-based
// serialization. It supports saving player progress, world state, and game
// settings to disk in JSON format for human readability and extensibility.
//
// # Key Features
//
//   - Player state serialization (position, stats, inventory, progression)
//   - World state serialization (terrain seed, genre, dimensions, time)
//   - Game settings persistence (graphics, audio, controls)
//   - Save file versioning and migration
//   - Comprehensive error handling for file I/O
//   - Deterministic loading (same save = same game state)
//
// # File Format
//
// Save files use JSON format with the following structure:
//
//	{
//	  "version": "1.0.0",
//	  "timestamp": "2025-10-22T17:30:00Z",
//	  "player": { ... },
//	  "world": { ... },
//	  "settings": { ... }
//	}
//
// # Usage Example
//
//	// Create a save manager
//	manager := saveload.NewSaveManager("./saves")
//
//	// Save game state
//	save := &saveload.GameSave{
//	    PlayerState: playerState,
//	    WorldState:  worldState,
//	    Settings:    settings,
//	}
//	err := manager.SaveGame("save1", save)
//
//	// Load game state
//	loaded, err := manager.LoadGame("save1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Version Compatibility
//
// The package supports save file versioning to handle format changes across
// game versions. Older save files are automatically migrated to the current
// format when loaded.
package saveload
