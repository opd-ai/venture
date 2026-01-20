// Package qol - recipetracker.go
// This file contains the RecipeTracker implementation for tracking recipe availability and materials.
// Code relocated from: manager.go

package qol

import (
	"sync"
)

// RecipeTracker tracks recipe availability and missing materials
type RecipeTracker struct {
	tracking map[uint64]map[string]*RecipeTrackingInfo // playerID -> recipeID -> tracking info
	mu       sync.RWMutex
}

// NewRecipeTracker creates a new recipe tracker
func NewRecipeTracker() *RecipeTracker {
	return &RecipeTracker{
		tracking: make(map[uint64]map[string]*RecipeTrackingInfo),
	}
}

// TrackRecipe adds a recipe to tracking for a player
func (r *RecipeTracker) TrackRecipe(playerID uint64, info *RecipeTrackingInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tracking[playerID]; !exists {
		r.tracking[playerID] = make(map[string]*RecipeTrackingInfo)
	}

	info.MissingMats = make(map[string]int)
	info.CanCraft = true
	info.MaxCraftable = int(^uint(0) >> 1) // Max int

	for matID, required := range info.RequiredMats {
		available := info.AvailableMats[matID]
		if available < required {
			info.MissingMats[matID] = required - available
			info.CanCraft = false
			info.MaxCraftable = 0
		} else {
			maxFromThisMat := available / required
			if maxFromThisMat < info.MaxCraftable {
				info.MaxCraftable = maxFromThisMat
			}
		}
	}

	if !info.CanCraft {
		info.MaxCraftable = 0
	}

	r.tracking[playerID][info.RecipeID] = info
}

// GetTrackedRecipes retrieves all tracked recipes for a player
func (r *RecipeTracker) GetTrackedRecipes(playerID uint64) []*RecipeTrackingInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	recipes, exists := r.tracking[playerID]
	if !exists {
		return make([]*RecipeTrackingInfo, 0)
	}

	result := make([]*RecipeTrackingInfo, 0, len(recipes))
	for _, info := range recipes {
		result = append(result, info)
	}

	return result
}

// UntrackRecipe removes a recipe from tracking
func (r *RecipeTracker) UntrackRecipe(playerID uint64, recipeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if recipes, exists := r.tracking[playerID]; exists {
		delete(recipes, recipeID)
	}
}

// UpdateMaterialAvailability updates available materials and recalculates craftability
func (r *RecipeTracker) UpdateMaterialAvailability(playerID uint64, recipeID string, availableMats map[string]int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	recipes, exists := r.tracking[playerID]
	if !exists {
		return
	}

	info, exists := recipes[recipeID]
	if !exists {
		return
	}

	info.AvailableMats = availableMats
	info.MissingMats = make(map[string]int)
	info.CanCraft = true
	info.MaxCraftable = int(^uint(0) >> 1)

	for matID, required := range info.RequiredMats {
		available := availableMats[matID]
		if available < required {
			info.MissingMats[matID] = required - available
			info.CanCraft = false
			info.MaxCraftable = 0
		} else {
			maxFromThisMat := available / required
			if maxFromThisMat < info.MaxCraftable {
				info.MaxCraftable = maxFromThisMat
			}
		}
	}

	if !info.CanCraft {
		info.MaxCraftable = 0
	}
}
