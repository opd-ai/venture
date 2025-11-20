package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/modding"
)

func main() {
	// Command line flags
	modsDir := flag.String("mods-dir", "mods", "Directory containing mod files")
	modID := flag.String("mod", "", "Specific mod ID to load (empty = load all)")
	listOnly := flag.Bool("list", false, "List all mods without loading")
	enableMod := flag.String("enable", "", "Enable a specific mod by ID")
	disableMod := flag.String("disable", "", "Disable a specific mod by ID")
	showStats := flag.Bool("stats", false, "Show mod manager statistics")
	showRules := flag.Bool("rules", false, "Show active rules after applying mods")

	flag.Parse()

	// Create loader
	config := modding.DefaultConfig()
	config.ModsDirectory = *modsDir
	loader := modding.NewLoaderWithConfig(config)

	// List mods
	if *listOnly {
		mods, err := loader.LoadAll()
		if err != nil {
			log.Fatalf("Failed to load mods: %v", err)
		}

		fmt.Printf("Found %d mods in %s:\n\n", len(mods), *modsDir)
		for _, mod := range mods {
			printMod(mod)
		}
		return
	}

	// Create manager
	manager := modding.NewManager()

	// Load mods
	var mods []*modding.Mod
	var err error

	if *modID != "" {
		// Load specific mod
		modPath := loader.GetModPath(*modID)
		mod, err := loader.LoadFromFile(modPath)
		if err != nil {
			log.Fatalf("Failed to load mod %s: %v", *modID, err)
		}
		mods = []*modding.Mod{mod}
	} else {
		// Load all mods
		mods, err = loader.LoadAll()
		if err != nil {
			log.Fatalf("Failed to load mods: %v", err)
		}
	}

	// Add mods to manager
	fmt.Printf("Loading %d mod(s)...\n", len(mods))
	for _, mod := range mods {
		if err := manager.AddMod(mod); err != nil {
			log.Printf("Warning: Failed to add mod %s: %v", mod.ID, err)
			continue
		}
		fmt.Printf("✓ Loaded: %s (v%s) by %s\n", mod.Name, mod.Version, mod.Author)
	}

	// Enable/disable mod if requested
	if *enableMod != "" {
		if err := manager.EnableMod(*enableMod); err != nil {
			log.Fatalf("Failed to enable mod %s: %v", *enableMod, err)
		}
		fmt.Printf("✓ Enabled mod: %s\n", *enableMod)
	}

	if *disableMod != "" {
		if err := manager.DisableMod(*disableMod); err != nil {
			log.Fatalf("Failed to disable mod %s: %v", *disableMod, err)
		}
		fmt.Printf("✓ Disabled mod: %s\n", *disableMod)
	}

	// Apply rules
	fmt.Println("\nApplying mod rules...")
	if err := manager.ApplyRules(); err != nil {
		log.Fatalf("Failed to apply rules: %v", err)
	}
	fmt.Println("✓ Rules applied successfully")

	// Show active rules
	if *showRules {
		fmt.Println("\nActive Rules:")
		for _, mod := range manager.ListMods() {
			if !mod.Enabled {
				continue
			}

			fmt.Printf("\n  From %s:\n", mod.Name)
			for ruleName, value := range mod.Rules {
				fmt.Printf("    %s = %v\n", ruleName, value)
			}
		}

		// Show example rule queries
		fmt.Println("\nExample Rule Queries:")
		fmt.Printf("  difficulty_multiplier = %.2f\n", manager.GetRuleFloat64("difficulty_multiplier", 1.0))
		fmt.Printf("  permadeath_enabled = %v\n", manager.GetRuleBool("permadeath_enabled", false))
		fmt.Printf("  spawn_rate_multiplier = %.2f\n", manager.GetRuleFloat64("spawn_rate_multiplier", 1.0))
	}

	// Show statistics
	if *showStats {
		stats := manager.GetStats()
		fmt.Println("\nMod Manager Statistics:")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// Show loaded mods summary
	fmt.Println("\nLoaded Mods Summary:")
	for _, mod := range manager.ListMods() {
		status := "disabled"
		if mod.Enabled {
			status = "enabled"
		}
		fmt.Printf("  [%s] %s - %s\n", status, mod.ID, mod.Description)
	}
}

func printMod(mod *modding.Mod) {
	fmt.Printf("ID: %s\n", mod.ID)
	fmt.Printf("Name: %s\n", mod.Name)
	fmt.Printf("Version: %s\n", mod.Version)
	fmt.Printf("Author: %s\n", mod.Author)
	fmt.Printf("Description: %s\n", mod.Description)
	fmt.Printf("Type: %s\n", mod.Type)
	fmt.Printf("Enabled: %v\n", mod.Enabled)

	if len(mod.Dependencies) > 0 {
		fmt.Printf("Dependencies: %v\n", mod.Dependencies)
	}

	if len(mod.Rules) > 0 {
		fmt.Println("Rules:")
		for name, value := range mod.Rules {
			fmt.Printf("  %s = %v\n", name, value)
		}
	}

	if len(mod.GeneratorParams) > 0 {
		fmt.Println("Generator Params:")
		for name, value := range mod.GeneratorParams {
			fmt.Printf("  %s = %v\n", name, value)
		}
	}

	fmt.Println()
}
