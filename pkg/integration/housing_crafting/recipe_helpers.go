package housing_crafting

import "fmt"

// Recipe generation helpers for StationManager.
// This file contains deterministic recipe generation methods that create
// recipe IDs based on station type and quality tier.
//
// Code relocated from: station_manager.go

// Helper methods for recipe generation (deterministic based on station type)

func (sm *StationManager) getBaseRecipes(stationType StationType) []string {
	recipes := make([]string, 0, 10)
	prefix := stationType.String()

	for i := 1; i <= 10; i++ {
		recipes = append(recipes, fmt.Sprintf("%s_basic_%d", prefix, i))
	}
	return recipes
}

func (sm *StationManager) getStandardRecipes(stationType StationType) []string {
	recipes := make([]string, 0, 10)
	prefix := stationType.String()

	for i := 1; i <= 10; i++ {
		recipes = append(recipes, fmt.Sprintf("%s_standard_%d", prefix, i))
	}
	return recipes
}

func (sm *StationManager) getAdvancedRecipes(stationType StationType) []string {
	recipes := make([]string, 0, 10)
	prefix := stationType.String()

	for i := 1; i <= 10; i++ {
		recipes = append(recipes, fmt.Sprintf("%s_advanced_%d", prefix, i))
	}
	return recipes
}

func (sm *StationManager) getMasterRecipes(stationType StationType) []string {
	recipes := make([]string, 0, 10)
	prefix := stationType.String()

	for i := 1; i <= 10; i++ {
		recipes = append(recipes, fmt.Sprintf("%s_master_%d", prefix, i))
	}
	return recipes
}
