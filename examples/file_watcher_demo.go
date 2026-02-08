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
	// Create temporary directory for demo
	tmpDir, err := os.MkdirTemp("", "filewatcher-demo-")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Demo directory: %s\n\n", tmpDir)

	// Create a sample mod file
	modID := "demo-mod"
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
		log.Fatalf("Failed to write mod file: %v", err)
	}

	// Create FileSystemFileWatcher
	watcher := engine.NewFileSystemFileWatcher(tmpDir)
	fmt.Println("✓ Created FileSystemFileWatcher")

	// Get initial hash
	hash1, err := watcher.GetFileHash(modID)
	if err != nil {
		log.Fatalf("GetFileHash failed: %v", err)
	}
	fmt.Printf("✓ Initial hash: %s\n", hash1[:16]+"...")

	// Get version
	version, err := watcher.GetModVersion(modID)
	if err != nil {
		log.Fatalf("GetModVersion failed: %v", err)
	}
	fmt.Printf("✓ Version: %s\n", version)

	// Get mod data
	data, err := watcher.GetModData(modID)
	if err != nil {
		log.Fatalf("GetModData failed: %v", err)
	}
	fmt.Printf("✓ Mod data size: %d bytes\n", len(data))

	// Simulate file change
	fmt.Println("\nSimulating mod file update...")
	modJSON2 := []byte(`{
  "id": "demo-mod",
  "name": "Demo Mod",
  "version": "2.0.0",
  "author": "Demo",
  "type": "rule",
  "rules": {
    "difficulty": 3.0
  }
}`)
	if err := os.WriteFile(modFile, modJSON2, 0o644); err != nil {
		log.Fatalf("Failed to update mod file: %v", err)
	}

	// Invalidate cache to detect change
	watcher.InvalidateCache(modID)
	fmt.Println("✓ Cache invalidated")

	// Get new hash
	hash2, err := watcher.GetFileHash(modID)
	if err != nil {
		log.Fatalf("GetFileHash (updated) failed: %v", err)
	}
	fmt.Printf("✓ Updated hash: %s\n", hash2[:16]+"...")

	// Verify hash changed
	if hash1 == hash2 {
		log.Fatal("ERROR: Hash should have changed!")
	}
	fmt.Println("✓ Hash changed successfully (hot reload detected)")

	// Get updated version
	version2, err := watcher.GetModVersion(modID)
	if err != nil {
		log.Fatalf("GetModVersion (updated) failed: %v", err)
	}
	fmt.Printf("✓ Updated version: %s\n", version2)

	fmt.Println("\n✓ FileSystemFileWatcher demo complete!")
	fmt.Println("\nUsage with HotReloadSystem:")
	fmt.Println("  watcher := engine.NewFileSystemFileWatcher(\"mods\")")
	fmt.Println("  hotReload := engine.NewHotReloadSystem(world, modManager)")
	fmt.Println("  hotReload.SetFileWatcher(watcher)")
	fmt.Println("  // Hot reload system will now detect file changes")
}
