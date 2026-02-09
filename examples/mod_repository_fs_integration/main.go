// Example: Integrating FileSystemModRepository with client mod browser
//
// This example shows how to set up the FileSystemModRepository in a real
// client application and integrate it with the ModBrowserSystem.

package main

import (
	"encoding/json"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/modding"
	"github.com/sirupsen/logrus"
)

func setupModBrowserWithFilesystem() {
	// 1. Create ECS world
	world := engine.NewWorld()

	// 2. Create modding manager for actual mod loading
	moddingManager := modding.NewManager()

	// 3. Create filesystem mod repository
	repo := engine.NewFileSystemModRepository("mods")

	// Optional: Enable debug logging for troubleshooting
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	repo.SetLogger(logger)

	// 4. Create mod browser system
	browserSystem := engine.NewModBrowserSystem(world)
	browserSystem.SetRepository(repo)

	// 5. Set install callback to integrate with modding manager
	browserSystem.SetInstallCallback(func(modID string, modData []byte) error {
		log.Printf("Installing mod: %s", modID)

		// Parse mod data
		var mod modding.Mod
		if err := json.Unmarshal(modData, &mod); err != nil {
			return err
		}

		// Add to modding manager
		if err := moddingManager.AddMod(&mod); err != nil {
			return err
		}

		// Enable the mod
		if err := moddingManager.EnableMod(mod.ID); err != nil {
			return err
		}

		// Apply rules if it's a rule mod
		if mod.Type == modding.ModTypeRule {
			if err := moddingManager.ApplyRules(); err != nil {
				return err
			}
		}

		log.Printf("Successfully installed mod: %s", mod.Name)
		return nil
	})

	// 6. Set uninstall callback
	browserSystem.SetUninstallCallback(func(modID string) error {
		log.Printf("Uninstalling mod: %s", modID)

		if err := moddingManager.RemoveMod(modID); err != nil {
			return err
		}

		log.Printf("Successfully uninstalled mod: %s", modID)
		return nil
	})

	// 7. Create mod browser entity
	entity := world.CreateEntity()
	browserComp := engine.NewModBrowserComponent()
	entity.AddComponent(browserComp)

	// 8. Trigger initial mod fetch
	browserComp.RefreshPending = true

	// 9. Update system to load mods
	entities := []*engine.Entity{entity}
	browserSystem.Update(entities, 0.016) // One frame

	// 10. Access loaded mods
	availableMods := browserComp.GetFilteredMods()
	log.Printf("Loaded %d mods from filesystem", len(availableMods))

	for _, mod := range availableMods {
		log.Printf("  - %s v%s by %s", mod.Name, mod.Version, mod.Author)
		if browserComp.IsInstalled(mod.ID) {
			log.Printf("    Status: INSTALLED")
		} else {
			log.Printf("    Status: Available for download")
		}
	}

	// Example: Install a mod programmatically
	if len(availableMods) > 0 {
		modToInstall := availableMods[0]
		log.Printf("Installing mod: %s", modToInstall.Name)

		// Download and install
		data, err := repo.DownloadMod(modToInstall.ID, func(downloaded, total int64) {
			progress := float64(downloaded) / float64(total) * 100
			log.Printf("Download progress: %.1f%%", progress)
		})
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}

		// Trigger install callback
		var mod modding.Mod
		json.Unmarshal(data, &mod)
		moddingManager.AddMod(&mod)
		moddingManager.EnableMod(mod.ID)

		browserComp.SetInstalled(modToInstall.ID, true)
		log.Printf("Mod installed successfully!")
	}
}

// This function would be called during client initialization, likely in
// cmd/client/handlers.go alongside other system initialization.
func main() {
	setupModBrowserWithFilesystem()
}
