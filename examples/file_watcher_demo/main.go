package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/engine"
)

// FileWatcherDemo demonstrates FileSystemFileWatcher usage
func main() {
	tmpDir, err := setupDemoEnvironment()
	if err != nil {
		log.Fatalf("Failed to setup demo: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Demo directory: %s\n\n", tmpDir)

	modID := "demo-mod"
	if err := createInitialModFile(tmpDir, modID); err != nil {
		log.Fatalf("Failed to create mod file: %v", err)
	}

	watcher := engine.NewFileSystemFileWatcher(tmpDir)
	fmt.Println("✓ Created FileSystemFileWatcher")

	if err := verifyInitialModState(watcher, modID); err != nil {
		log.Fatalf("Initial verification failed: %v", err)
	}

	if err := simulateModUpdate(tmpDir, watcher, modID); err != nil {
		log.Fatalf("Update simulation failed: %v", err)
	}

	displayUsageInstructions()
}

// setupDemoEnvironment creates a temporary directory for the demo.
func setupDemoEnvironment() (string, error) {
	tmpDir, err := os.MkdirTemp("", "filewatcher-demo-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	return tmpDir, nil
}

// createInitialModFile creates the initial demo mod file.
func createInitialModFile(tmpDir, modID string) error {
	modJSON := []byte(`{
  "id": "demo-mod",
  "name": "Demo Mod",
  "version": "1.0.0",
  "author": "Demo",
  "type": "rule",
  "rules": {
    "difficulty": 2.0
  }
}`)
	modFile := filepath.Join(tmpDir, modID+".json")
	if err := os.WriteFile(modFile, modJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write mod file: %w", err)
	}
	return nil
}

// verifyInitialModState checks hash, version, and data of the initial mod.
func verifyInitialModState(watcher *engine.FileSystemFileWatcher, modID string) error {
	hash, err := watcher.GetFileHash(modID)
	if err != nil {
		return fmt.Errorf("GetFileHash failed: %w", err)
	}
	fmt.Printf("✓ Initial hash: %s\n", hash[:16]+"...")

	version, err := watcher.GetModVersion(modID)
	if err != nil {
		return fmt.Errorf("GetModVersion failed: %w", err)
	}
	fmt.Printf("✓ Version: %s\n", version)

	data, err := watcher.GetModData(modID)
	if err != nil {
		return fmt.Errorf("GetModData failed: %w", err)
	}
	fmt.Printf("✓ Mod data size: %d bytes\n", len(data))

	return nil
}

// simulateModUpdate modifies the mod file and verifies hot reload detection.
func simulateModUpdate(tmpDir string, watcher *engine.FileSystemFileWatcher, modID string) error {
	fmt.Println("\nSimulating mod file update...")

	hash1, err := watcher.GetFileHash(modID)
	if err != nil {
		return fmt.Errorf("failed to get initial hash: %w", err)
	}

	modJSON := []byte(`{
  "id": "demo-mod",
  "name": "Demo Mod",
  "version": "2.0.0",
  "author": "Demo",
  "type": "rule",
  "rules": {
    "difficulty": 3.0
  }
}`)
	modFile := filepath.Join(tmpDir, modID+".json")
	if err := os.WriteFile(modFile, modJSON, 0o644); err != nil {
		return fmt.Errorf("failed to update mod file: %w", err)
	}

	watcher.InvalidateCache(modID)
	fmt.Println("✓ Cache invalidated")

	hash2, err := watcher.GetFileHash(modID)
	if err != nil {
		return fmt.Errorf("GetFileHash (updated) failed: %w", err)
	}
	fmt.Printf("✓ Updated hash: %s\n", hash2[:16]+"...")

	if hash1 == hash2 {
		return fmt.Errorf("hash should have changed")
	}
	fmt.Println("✓ Hash changed successfully (hot reload detected)")

	version, err := watcher.GetModVersion(modID)
	if err != nil {
		return fmt.Errorf("GetModVersion (updated) failed: %w", err)
	}
	fmt.Printf("✓ Updated version: %s\n", version)

	fmt.Println("\n✓ FileSystemFileWatcher demo complete!")
	return nil
}

// displayUsageInstructions shows how to integrate the file watcher.
func displayUsageInstructions() {
	fmt.Println("\nUsage with HotReloadSystem:")
	fmt.Println("  watcher := engine.NewFileSystemFileWatcher(\"mods\")")
	fmt.Println("  hotReload := engine.NewHotReloadSystem(world, modManager)")
	fmt.Println("  hotReload.SetFileWatcher(watcher)")
	fmt.Println("  // Hot reload system will now detect file changes")
}
