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

// createModInstallCallback creates a callback function for installing mods.
func createModInstallCallback(moddingManager *modding.Manager) func(string, []byte) error {
	return func(modID string, modData []byte) error {
		log.Printf("Installing mod: %s", modID)

		var mod modding.Mod
		if err := json.Unmarshal(modData, &mod); err != nil {
			return err
		}

		if err := moddingManager.AddMod(&mod); err != nil {
			return err
		}

		if err := moddingManager.EnableMod(mod.ID); err != nil {
			return err
		}

		if mod.Type == modding.ModTypeRule {
			if err := moddingManager.ApplyRules(); err != nil {
				return err
			}
		}

		log.Printf("Successfully installed mod: %s", mod.Name)
		return nil
	}
}

// createModUninstallCallback creates a callback function for uninstalling mods.
func createModUninstallCallback(moddingManager *modding.Manager) func(string) error {
	return func(modID string) error {
		log.Printf("Uninstalling mod: %s", modID)

		if err := moddingManager.RemoveMod(modID); err != nil {
			return err
		}

		log.Printf("Successfully uninstalled mod: %s", modID)
		return nil
	}
}

// initializeModRepository creates and configures a filesystem mod repository with debug logging.
func initializeModRepository() *engine.FileSystemModRepository {
	repo := engine.NewFileSystemModRepository("mods")
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	repo.SetLogger(logger)
	return repo
}

// setupBrowserSystem creates and configures a mod browser system with callbacks.
func setupBrowserSystem(world *engine.World, repo *engine.FileSystemModRepository, moddingManager *modding.Manager) *engine.ModBrowserSystem {
	browserSystem := engine.NewModBrowserSystem(world)
	browserSystem.SetRepository(repo)
	browserSystem.SetInstallCallback(createModInstallCallback(moddingManager))
	browserSystem.SetUninstallCallback(createModUninstallCallback(moddingManager))
	return browserSystem
}

// loadAvailableMods initializes browser entity and loads available mods.
func loadAvailableMods(world *engine.World, browserSystem *engine.ModBrowserSystem) (*engine.Entity, *engine.ModBrowserComponent) {
	entity := world.CreateEntity()
	browserComp := engine.NewModBrowserComponent()
	entity.AddComponent(browserComp)
	browserComp.RefreshPending = true

	entities := []*engine.Entity{entity}
	browserSystem.Update(entities, 0.016)

	return entity, browserComp
}

// displayAvailableMods logs all available mods with their status.
func displayAvailableMods(browserComp *engine.ModBrowserComponent) {
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
}

// installFirstAvailableMod downloads and installs the first available mod as an example.
func installFirstAvailableMod(repo *engine.FileSystemModRepository, browserComp *engine.ModBrowserComponent, moddingManager *modding.Manager) {
	availableMods := browserComp.GetFilteredMods()
	if len(availableMods) == 0 {
		return
	}

	modToInstall := availableMods[0]
	log.Printf("Installing mod: %s", modToInstall.Name)

	data, err := repo.DownloadMod(modToInstall.ID, func(downloaded, total int64) {
		progress := float64(downloaded) / float64(total) * 100
		log.Printf("Download progress: %.1f%%", progress)
	})
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	var mod modding.Mod
	if err := json.Unmarshal(data, &mod); err != nil {
		log.Printf("Failed to unmarshal mod data: %v", err)
		return
	}

	moddingManager.AddMod(&mod)
	moddingManager.EnableMod(mod.ID)
	browserComp.SetInstalled(modToInstall.ID, true)
	log.Printf("Mod installed successfully!")
}

func setupModBrowserWithFilesystem() {
	world := engine.NewWorld()
	moddingManager := modding.NewManager()
	repo := initializeModRepository()
	browserSystem := setupBrowserSystem(world, repo, moddingManager)
	_, browserComp := loadAvailableMods(world, browserSystem)
	displayAvailableMods(browserComp)
	installFirstAvailableMod(repo, browserComp, moddingManager)
}

// This function would be called during client initialization, likely in
// cmd/client/handlers.go alongside other system initialization.
func main() {
	setupModBrowserWithFilesystem()
}
