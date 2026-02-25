# Save/Load System

Package `saveload` provides persistent game state management through file-based serialization for the Venture procedural action-RPG.

## Overview

The save/load system allows players to save their progress and resume gameplay later. It uses JSON format for human-readable save files.

## Features

- **Player State Persistence**: Position, health, stats, level, experience, inventory, and equipment
- **World State Persistence**: Terrain seed, genre, dimensions, game time, difficulty, depth
- **Game Settings**: Screen resolution, fullscreen mode, vsync, audio volumes, key bindings
- **Save File Management**: Create, read, update, delete save files
- **Metadata Support**: Browse saves without loading full state (quick save list)
- **Version Tracking**: Save format versioning with migration support for versions 0.9.0-1.0.0
- **Security**: Save name validation prevents path traversal attacks
- **Error Handling**: Comprehensive validation and error messages
- **Corruption Recovery**: Automatic backup creation and checksum validation (NEW)
- **Data Integrity**: SHA256 checksums detect corrupted save files (NEW)
- **Automatic Backups**: Previous save version preserved before overwriting (NEW)
- **Smart Recovery**: Automatic restoration from backup on corruption detection (NEW)

## Usage

### Creating a Save Manager

```go
import (
    "github.com/opd-ai/venture/pkg/saveload"
    "github.com/sirupsen/logrus"
)

// Create manager with save directory
manager, err := saveload.NewSaveManager("./saves")
if err != nil {
    logrus.WithError(err).Fatal("failed to create save manager")
}
```

### Saving Game State

```go
// Create save data
save := saveload.NewGameSave()

// Populate player state
save.PlayerState.EntityID = playerEntity.ID
save.PlayerState.X = playerPos.X
save.PlayerState.Y = playerPos.Y
save.PlayerState.Level = playerLevel
save.PlayerState.Experience = playerXP
save.PlayerState.CurrentHealth = playerHealth
save.PlayerState.MaxHealth = playerMaxHealth

// Populate world state
save.WorldState.Seed = worldSeed
save.WorldState.GenreID = "fantasy"
save.WorldState.Width = terrainWidth
save.WorldState.Height = terrainHeight
save.WorldState.GameTime = elapsedTime

// Populate settings
save.Settings.ScreenWidth = 1920
save.Settings.ScreenHeight = 1080
save.Settings.MasterVolume = 0.8

// Save to file
err = manager.SaveGame("quicksave", save)
if err != nil {
    logrus.WithError(err).Fatal("failed to save game")
}
```

### Loading Game State

```go
// Load save file
save, err := manager.LoadGame("quicksave")
if err != nil {
    logrus.WithError(err).Fatal("failed to load game")
}

// Restore player state
playerEntity.ID = save.PlayerState.EntityID
playerPos.X = save.PlayerState.X
playerPos.Y = save.PlayerState.Y
playerLevel = save.PlayerState.Level
playerXP = save.PlayerState.Experience

// Restore world state (regenerate using saved seed)
terrainGen := terrain.NewBSPGenerator()
params := procgen.GenerationParams{
    Difficulty: save.WorldState.Difficulty,
    Depth:      save.WorldState.Depth,
    GenreID:    save.WorldState.GenreID,
    Custom: map[string]interface{}{
        "width":  save.WorldState.Width,
        "height": save.WorldState.Height,
    },
}
regeneratedTerrain, _ := terrainGen.Generate(save.WorldState.Seed, params)
```

### Listing Saves

```go
// Get all save files
saves, err := manager.ListSaves()
if err != nil {
    logrus.WithError(err).Fatal("failed to list saves")
}

// Display saves (sorted by timestamp, newest first)
for _, save := range saves {
    logrus.WithFields(logrus.Fields{
        "name":     save.Name,
        "level":    save.PlayerLevel,
        "genre":    save.GenreID,
        "hours":    save.GameTime / 3600,
        "created":  save.Timestamp.Format("2006-01-02 15:04"),
    }).Info("save file")
}
```

### Checking If Save Exists

```go
if manager.SaveExists("autosave") {
    fmt.Println("Autosave found!")
}
```

### Deleting a Save

```go
err = manager.DeleteSave("old-save")
if err != nil {
    logrus.WithError(err).Fatal("failed to delete save")
}
```

### Getting Save Metadata

```go
// Get metadata without loading entire save
metadata, err := manager.GetSaveMetadata("save1")
if err != nil {
    logrus.WithError(err).Fatal("failed to get metadata")
}

logrus.WithFields(logrus.Fields{
    "level":     metadata.PlayerLevel,
    "file_size": metadata.FileSize,
}).Info("save metadata")
```

## Corruption Recovery (Production Ready)

### Saving with Automatic Backup and Checksum

For production use, it's recommended to use `SaveGameWithBackup` instead of `SaveGame`. This method:
- Creates a backup of the existing save before overwriting (`.bak` file)
- Generates a SHA256 checksum for integrity verification (`.sha256` file)
- Automatically restores the backup if the save operation fails

```go
// Save with automatic backup and checksum (RECOMMENDED)
err = manager.SaveGameWithBackup("savegame", save)
if err != nil {
    logrus.WithError(err).Fatal("failed to save game with backup")
}
// Files created:
// - savegame.sav (current save)
// - savegame.sav.bak (previous version)
// - savegame.sav.sha256 (checksum)
```

### Loading with Automatic Recovery

Use `LoadGameWithRecovery` to enable automatic corruption detection and recovery:

```go
// Load with automatic recovery (RECOMMENDED)
save, err := manager.LoadGameWithRecovery("savegame")
if err != nil {
    logrus.WithError(err).Fatal("failed to load game with recovery")
}
// Recovery process:
// 1. Validates checksum
// 2. If corrupted, automatically restores from .bak file
// 3. If backup is also corrupted, returns error
// 4. Logs all recovery attempts
```

### Manual Backup Management

```go
// Check if backup exists
if manager.BackupExists("savegame") {
    fmt.Println("Backup available")
}

// Get backup file path
backupPath := manager.GetBackupPath("savegame")
fmt.Println("Backup location:", backupPath)

// List all saves with backups
backups, err := manager.ListBackups()
for _, name := range backups {
    fmt.Println("Backup for:", name)
}

// Cleanup old backups and checksums (optional)
err = manager.CleanupBackups("savegame")
// Removes .bak and .sha256 files, keeps .sav
```

### Recovery Workflow

When corruption is detected:

1. **Checksum Mismatch**: LoadGameWithRecovery detects corrupted file via checksum
2. **Automatic Recovery**: Attempts to restore from `.bak` file
3. **Validation**: Verifies backup is valid before restoring
4. **Logging**: All recovery attempts logged with structured fields
5. **Fallback**: If recovery fails, returns detailed error

```go
// Example with explicit error handling
save, err := manager.LoadGameWithRecovery("savegame")
if err != nil {
    if strings.Contains(err.Error(), "no valid backup") {
        // Both save and backup are corrupted
        logrus.Warn("cannot recover - save file lost")
        // Offer user to start new game
    } else if strings.Contains(err.Error(), "not found") {
        // Save doesn't exist
        logrus.Warn("no save file found")
    } else {
        // Other errors
        logrus.WithError(err).Error("load error")
    }
    return
}
// Save loaded successfully (possibly after recovery)
```

## Save File Format

Save files use JSON format with `.sav` extension:

```json
{
  "version": "1.0.0",
  "timestamp": "2025-10-22T17:30:00Z",
  "player": {
    "entity_id": 12345,
    "x": 100.5,
    "y": 200.7,
    "current_health": 85.0,
    "max_health": 100.0,
    "level": 10,
    "experience": 5000,
    "attack": 50.0,
    "defense": 30.0,
    "magic_power": 40.0,
    "speed": 100.0,
    "items": [
      {
        "id": "sword_001",
        "name": "Iron Sword",
        "type": "weapon",
        "rarity": "common",
        "seed": 12345,
        "damage": 10,
        "value": 50,
        "weight": 2.5
      }
    ],
    "equipped_items": {
      "weapon": {
        "id": "sword_001",
        "name": "Iron Sword",
        "type": "weapon",
        "rarity": "common",
        "seed": 12345,
        "damage": 10,
        "value": 50,
        "weight": 2.5
      },
      "armor": null,
      "accessory": null
    }
  },
  "world": {
    "seed": 67890,
    "genre_id": "fantasy",
    "width": 100,
    "height": 80,
    "game_time": 3600.5,
    "difficulty": 0.5,
    "depth": 5,
    "modified_entities": [
      {
        "entity_id": 3001,
        "x": 50.0,
        "y": 60.0,
        "health": 0.0,
        "is_alive": false
      }
    ]
  },
  "settings": {
    "screen_width": 1920,
    "screen_height": 1080,
    "fullscreen": false,
    "vsync": true,
    "master_volume": 1.0,
    "music_volume": 0.7,
    "sfx_volume": 0.8,
    "key_bindings": {
      "move_up": "w",
      "attack": "space"
    }
  }
}
```

## Deterministic World Regeneration

Venture uses procedural generation, so most world content doesn't need to be saved. The save system stores:

1. **World Seed**: Regenerates identical terrain/entities
2. **Modified Entities**: Only entities changed from procedural state
   - Killed enemies (health=0, is_alive=false)
   - Picked up items (is_picked=true)
   - NPCs that moved from spawn position

When loading, the game:
1. Regenerates world from seed (same terrain, monsters, items)
2. Applies modifications (remove killed enemies, picked items)
3. Keeps save files small (KB instead of MB)

## Error Handling

The package provides detailed error messages for common issues:

- **File Not Found**: `"save file not found: savename"`
- **Corrupted Save**: `"failed to parse save file: <json error>"`
- **Invalid Name**: `"save name cannot contain path separators"`
- **Missing Fields**: `"save file missing player state"`
- **Version Mismatch**: `"save file version X.Y.Z is not supported"`

## Security

### Path Traversal Prevention

Save names are validated to prevent directory traversal attacks:

```go
// ✅ Valid names
"quicksave"
"save_01"
"my-save-2025"

// ❌ Invalid names (rejected)
"../../../etc/passwd"
"saves\\..\\config"
"save:file"
```

### File Permissions

Save files are created with permissions `0644` (readable by all, writable by owner only).

## Version Tracking

The save format uses semantic versioning (currently `1.0.0`). The migration system in `migrator.go` supports automatic migration from versions 0.9.0, 0.9.1, 0.9.2, and 0.9.3 to the current version.

When loading a save file:
1. The version is checked against the current `SaveVersion`
2. If older, the `DefaultMigrator` applies necessary transformations
3. Save files are upgraded in-place after successful migration

Custom migrators can be provided via `NewSaveManagerWithMigrator()` for specialized migration needs.

## Performance

- **Save Time**: 5-10ms (JSON marshaling + file write)
- **Load Time**: 10-20ms (file read + JSON unmarshaling + validation)
- **File Size**: ~2-10KB per save (compact JSON)
- **Memory**: Minimal (saves loaded on-demand, not kept in memory)

## Integration with Game Systems

### Engine Integration

```go
// In game loop
if inputSystem.IsKeyPressed(KeyF5) {
    // Quick save
    save := createSaveFromGameState(game)
    saveManager.SaveGame("quicksave", save)
}

if inputSystem.IsKeyPressed(KeyF9) {
    // Quick load
    save, _ := saveManager.LoadGame("quicksave")
    loadSaveIntoGameState(game, save)
}
```

### Inventory System Integration

```go
// Save inventory items
save.PlayerState.InventoryItems = make([]uint64, 0)
for _, item := range inventory.GetAllItems() {
    save.PlayerState.InventoryItems = append(save.PlayerState.InventoryItems, item.ID)
}

// Load inventory items
inventory.Clear()
for _, itemID := range save.PlayerState.InventoryItems {
    // Regenerate item from ID (deterministic)
    item := itemGen.RegenerateItem(itemID, save.WorldState.Seed)
    inventory.AddItem(item)
}
```

### Progression System Integration

```go
// Save progression
save.PlayerState.Level = progressionSystem.GetLevel(playerID)
save.PlayerState.Experience = progressionSystem.GetExperience(playerID)

// Load progression
progressionSystem.SetLevel(playerID, save.PlayerState.Level)
progressionSystem.SetExperience(playerID, save.PlayerState.Experience)
```

## Testing

The package includes comprehensive tests covering:

- Save/load workflow (basic and complex data)
- Error handling (missing files, corrupted JSON, invalid names)
- Save listing and metadata
- Version validation
- File system operations
- Security (path traversal prevention)

Run tests:

```bash
go test -tags test ./pkg/saveload -v
go test -tags test ./pkg/saveload -cover
```

**Test Coverage**: 84.4% of statements

## Future Enhancements

Potential improvements for future versions:

1. **Compression**: Gzip save files to reduce disk usage
2. **Encryption**: Optional save file encryption for anti-cheat
3. **Cloud Saves**: Sync saves across devices via cloud storage
4. **Auto-save**: Periodic automatic saves every N minutes
5. **Save Slots**: Multiple named save slots per player
6. ~~**Backup**: Automatic backup of previous save before overwriting~~ ✅ IMPLEMENTED
7. **Statistics**: Track playtime, death count, achievements in save metadata

## File Structure

```
./saves/
├── quicksave.sav           # Quick save slot
├── quicksave.sav.bak       # Backup of previous version (NEW)
├── quicksave.sav.sha256    # Checksum for integrity verification (NEW)
├── autosave.sav            # Auto-save slot
├── autosave.sav.bak        # Auto-save backup (NEW)
├── autosave.sav.sha256     # Auto-save checksum (NEW)
├── save1.sav               # Manual save #1
├── save2.sav               # Manual save #2
└── checkpoint.sav          # Checkpoint save
```

## API Reference

### Types

- **`GameSave`**: Complete save file with version, timestamp, player, world, settings
- **`PlayerState`**: Player position, health, stats, inventory, equipment
- **`WorldState`**: World seed, genre, dimensions, time, modified entities
- **`GameSettings`**: Graphics, audio, control settings
- **`SaveMetadata`**: Summary info (name, version, timestamp, level, genre)
- **`ModifiedEntity`**: Entity that differs from procedural generation

### SaveManager Methods

**Basic Operations:**
- **`NewSaveManager(dir string) (*SaveManager, error)`**: Create manager
- **`SaveGame(name string, save *GameSave) error`**: Save game state (basic)
- **`LoadGame(name string) (*GameSave, error)`**: Load game state (basic)
- **`DeleteSave(name string) error`**: Delete save file
- **`ListSaves() ([]*SaveMetadata, error)`**: List all saves
- **`GetSaveMetadata(name string) (*SaveMetadata, error)`**: Get save info
- **`SaveExists(name string) bool`**: Check if save exists

**Production-Ready Operations (RECOMMENDED):**
- **`SaveGameWithBackup(name string, save *GameSave) error`**: Save with automatic backup and checksum
- **`LoadGameWithRecovery(name string) (*GameSave, error)`**: Load with corruption detection and recovery

**Backup Management:**
- **`BackupExists(name string) bool`**: Check if backup exists for save
- **`GetBackupPath(name string) string`**: Get path to backup file
- **`ListBackups() ([]string, error)`**: List all saves with backups
- **`CleanupBackups(name string) error`**: Remove backup and checksum files

### Helper Functions

- **`NewGameSave() *GameSave`**: Create new save with defaults

## Example: Complete Save/Load Workflow

```go
package main

import (
    "github.com/opd-ai/venture/pkg/saveload"
    "github.com/opd-ai/venture/pkg/engine"
    "github.com/sirupsen/logrus"
)

func saveGameState(game *engine.Game, saveName string) error {
    // Create save manager
    manager, err := saveload.NewSaveManager("./saves")
    if err != nil {
        return err
    }

    // Create save data
    save := saveload.NewGameSave()

    // Extract player state from game
    playerID := game.PlayerID
    if pos := game.World.GetComponent(playerID, "position"); pos != nil {
        posComp := pos.(*engine.PositionComponent)
        save.PlayerState.X = posComp.X
        save.PlayerState.Y = posComp.Y
    }

    if health := game.World.GetComponent(playerID, "health"); health != nil {
        healthComp := health.(*engine.HealthComponent)
        save.PlayerState.CurrentHealth = healthComp.Current
        save.PlayerState.MaxHealth = healthComp.Max
    }

    // Extract world state
    save.WorldState.Seed = game.WorldSeed
    save.WorldState.GenreID = game.GenreID
    save.WorldState.GameTime = game.GameTime

    // Extract settings
    save.Settings.ScreenWidth = game.ScreenWidth
    save.Settings.ScreenHeight = game.ScreenHeight

    // Save to file
    return manager.SaveGame(saveName, save)
}

func loadGameState(saveName string) (*saveload.GameSave, error) {
    manager, err := saveload.NewSaveManager("./saves")
    if err != nil {
        return nil, err
    }

    return manager.LoadGame(saveName)
}

func main() {
    // Save game
    if err := saveGameState(game, "quicksave"); err != nil {
        logrus.WithError(err).Error("save failed")
    } else {
        logrus.Info("game saved successfully")
    }

    // Load game
    save, err := loadGameState("quicksave")
    if err != nil {
        logrus.WithError(err).Error("load failed")
    } else {
        logrus.WithFields(logrus.Fields{
            "level": save.PlayerState.Level,
            "genre": save.WorldState.GenreID,
        }).Info("game loaded")
    }
}
```

## Troubleshooting

### Save File Won't Load

1. Check file exists: `manager.SaveExists("savename")`
2. Verify save name doesn't have path separators
3. Check for JSON syntax errors (open .sav file in text editor)
4. Verify version compatibility
5. Check file permissions (must be readable)

### Save File Corrupted

If a save file becomes corrupted:

**Automatic Recovery (Recommended):**
1. Use `LoadGameWithRecovery("name")` - automatically detects and fixes corruption
2. Check logs for recovery status and details
3. If recovery succeeds, save is restored from `.bak` file
4. If both save and backup are corrupted, error is returned

**Manual Recovery:**
1. Check if backup exists: `manager.BackupExists("name")`
2. Manually copy `.bak` file to `.sav` if needed
3. Try loading with detailed error: `save, err := manager.LoadGame("name")`
4. Examine error message for JSON parsing issues
5. Open `.sav` file in text editor to inspect JSON
6. If manually edited, validate JSON syntax

**Prevention:**
- Always use `SaveGameWithBackup()` in production
- Keep checksum files (`.sha256`) for integrity verification
- Don't manually edit save files unless necessary

### Large Save Files

If save files grow unexpectedly large:

1. Check `ModifiedEntities` array length
2. Reduce entities saved (only save meaningful modifications)
3. Consider compressing save files (future enhancement)
4. Clear old modified entities periodically

## License

Part of the Venture project. See root LICENSE file.
