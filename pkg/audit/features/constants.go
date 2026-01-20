// Package features provides feature category constants.
// This file defines feature categories for feature completeness validation.
// Code relocated from: feature_completeness.go
package features

// FeatureCategory represents a category of features
type FeatureCategory string

const (
	CategoryCore     FeatureCategory = "Core Gameplay"
	CategoryAdvanced FeatureCategory = "Advanced Systems"
	CategoryVehicles FeatureCategory = "Vehicles"
	CategorySocial   FeatureCategory = "Social"
	CategoryHousing  FeatureCategory = "Housing"
	CategoryGuilds   FeatureCategory = "Guilds"
	CategoryCombat   FeatureCategory = "Combat"
	CategoryEconomy  FeatureCategory = "Economy"
	CategoryContent  FeatureCategory = "Content"
	CategoryMeta     FeatureCategory = "Meta-Game"
)
