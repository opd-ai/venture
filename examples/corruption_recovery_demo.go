// Package main demonstrates the save file corruption recovery features.
// This example shows how to use SaveGameWithBackup and LoadGameWithRecovery
// for production-ready save file management with automatic corruption detection.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logger for demonstration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create save directory for demo
	saveDir := "./demo_saves"
	defer os.RemoveAll(saveDir) // Cleanup demo directory

	// Create save manager with logger
	manager, err := saveload.NewSaveManagerWithLogger(saveDir, logger)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Save File Corruption Recovery Demo ===\n")

	// Step 1: Create and save initial game state
	fmt.Println("Step 1: Creating initial save...")
	save := saveload.NewGameSave()
	save.PlayerState.Level = 10
	save.PlayerState.Experience = 5000
	save.PlayerState.X = 100.0
	save.PlayerState.Y = 200.0
	save.WorldState.Seed = 12345
	save.WorldState.GenreID = "fantasy"

	err = manager.SaveGameWithBackup("demo", save)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Saved: Level %d, XP %d\n", save.PlayerState.Level, save.PlayerState.Experience)
	fmt.Printf("  Files created: demo.sav, demo.sav.sha256\n\n")

	// Step 2: Update the save (creates backup)
	fmt.Println("Step 2: Updating save (creates backup)...")
	save.PlayerState.Level = 20
	save.PlayerState.Experience = 15000

	err = manager.SaveGameWithBackup("demo", save)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Updated: Level %d, XP %d\n", save.PlayerState.Level, save.PlayerState.Experience)
	fmt.Printf("  Backup created: demo.sav.bak (contains Level 10 version)\n\n")

	// Step 3: Verify backup exists
	fmt.Println("Step 3: Checking backup...")
	if manager.BackupExists("demo") {
		fmt.Printf("✓ Backup exists at: %s\n\n", manager.GetBackupPath("demo"))
	}

	// Step 4: Simulate corruption
	fmt.Println("Step 4: Simulating save file corruption...")
	savePath := saveDir + "/demo.sav"
	err = os.WriteFile(savePath, []byte("CORRUPTED DATA {{{"), 0o644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✗ Corrupted save file!\n\n")

	// Step 5: Attempt recovery
	fmt.Println("Step 5: Loading with automatic recovery...")
	recovered, err := manager.LoadGameWithRecovery("demo")
	if err != nil {
		log.Fatalf("Recovery failed: %v", err)
	}

	fmt.Printf("✓ Recovery successful!\n")
	fmt.Printf("  Restored from backup: Level %d, XP %d\n", 
		recovered.PlayerState.Level, recovered.PlayerState.Experience)
	fmt.Printf("  Note: Backup contained the previous version (Level 10)\n\n")

	// Step 6: List all backups
	fmt.Println("Step 6: Listing all backups...")
	backups, err := manager.ListBackups()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Found %d saves with backups:\n", len(backups))
	for _, name := range backups {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()

	// Step 7: Cleanup backups (optional)
	fmt.Println("Step 7: Cleaning up backups...")
	err = manager.CleanupBackups("demo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Removed backup and checksum files\n")
	fmt.Printf("  Kept: demo.sav (main save file)\n\n")

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("1. Use SaveGameWithBackup() for production saves")
	fmt.Println("2. Use LoadGameWithRecovery() for production loads")
	fmt.Println("3. Backups are created automatically before overwriting")
	fmt.Println("4. Checksums detect corruption during load")
	fmt.Println("5. Recovery is automatic and logged")
	fmt.Println("6. Clean up old backups periodically to save space")
}
