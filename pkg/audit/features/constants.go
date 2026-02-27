// Package features provides feature category constants.
// This file defines feature categories for feature completeness validation.
// Code relocated from: feature_completeness.go
package features

// FeatureCategory represents a category of features
type FeatureCategory string

const (
	// CategoryCore represents core gameplay features including movement, combat, inventory, and player progression
	CategoryCore FeatureCategory = "Core Gameplay"
	// CategoryAdvanced represents advanced systems including multiclassing, prestige, and skill trees
	CategoryAdvanced FeatureCategory = "Advanced Systems"
	// CategoryVehicles represents vehicle-related features including mounting, fleet management, and vehicle physics
	CategoryVehicles FeatureCategory = "Vehicles"
	// CategorySocial represents social features including chat, mail, guilds, and player trading
	CategorySocial FeatureCategory = "Social"
	// CategoryHousing represents housing features including blueprints, furniture placement, and guild halls
	CategoryHousing FeatureCategory = "Housing"
	// CategoryGuilds represents guild-specific features including permissions, banks, and cross-server guilds
	CategoryGuilds FeatureCategory = "Guilds"
	// CategoryCombat represents combat features including spells, status effects, and PvP systems
	CategoryCombat FeatureCategory = "Combat"
	// CategoryEconomy represents economic features including marketplace, crafting, and trade routes
	CategoryEconomy FeatureCategory = "Economy"
	// CategoryContent represents procedurally generated content including terrain, quests, items, and NPCs
	CategoryContent FeatureCategory = "Content"
	// CategoryMeta represents meta-game features including achievements, leaderboards, and world events
	CategoryMeta FeatureCategory = "Meta-Game"
)
