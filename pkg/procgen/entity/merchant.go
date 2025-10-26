// Package entity provides procedural merchant generation.
// This file implements merchant-specific entity generation, inventory stocking,
// and spawn location logic for both fixed and nomadic merchants.
package entity

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// MerchantType represents the behavior pattern of a merchant NPC.
type MerchantType int

const (
	// MerchantFixed represents stationary shopkeepers in settlements
	MerchantFixed MerchantType = iota
	// MerchantNomadic represents wandering merchants that spawn periodically
	MerchantNomadic
)

// String returns the string representation of a merchant type.
func (m MerchantType) String() string {
	switch m {
	case MerchantFixed:
		return "fixed"
	case MerchantNomadic:
		return "nomadic"
	default:
		return "unknown"
	}
}

// MerchantData holds merchant-specific generation data that will be
// converted to engine components at runtime.
type MerchantData struct {
	// Entity is the base NPC entity
	Entity *Entity

	// MerchantType determines spawn behavior
	MerchantType MerchantType

	// Inventory contains items available for purchase
	Inventory []*item.Item

	// PriceMultiplier affects buy/sell prices (1.5 = 50% markup)
	PriceMultiplier float64

	// BuyBackPercentage is the percentage paid when buying from player
	BuyBackPercentage float64

	// SpawnX and SpawnY are suggested spawn coordinates
	SpawnX, SpawnY float64
}

// MerchantNameTemplates provides genre-specific merchant names.
var MerchantNameTemplates = map[string][]string{
	"fantasy": {
		"Aldric the Trader", "Mirena's Goods", "Thorn's Supplies",
		"Eldwin the Merchant", "Garrick's Wares", "Lyra's Emporium",
	},
	"scifi": {
		"Tech Trader Station", "Quantum Supplies", "Nexus Exchange",
		"Orbital Merchant Hub", "Void Trader", "Cyber Market",
	},
	"horror": {
		"The Bone Trader", "Cursed Goods", "Shadow Market",
		"Whisper's Wares", "The Dead Merchant", "Haunted Shop",
	},
	"cyberpunk": {
		"Chrome Exchange", "Neural Market", "Augment Shop",
		"Black Market Node", "Tech Fence", "Data Broker",
	},
	"postapoc": {
		"Wasteland Trader", "Salvage Market", "Scrap Exchange",
		"Survivor's Supplies", "Junk Dealer", "Raider's Cache",
	},
}

// GenerateMerchant creates a merchant NPC with inventory.
// The merchantType parameter determines if the merchant is fixed (settlement)
// or nomadic (wandering). Inventory is generated using the item generator
// with genre-appropriate stock.
func (g *EntityGenerator) GenerateMerchant(seed int64, params procgen.GenerationParams, merchantType MerchantType) (*MerchantData, error) {
	if g.logger != nil {
		g.logger.WithFields(map[string]interface{}{
			"seed":         seed,
			"genreID":      params.GenreID,
			"merchantType": merchantType.String(),
		}).Debug("generating merchant")
	}

	rng := rand.New(rand.NewSource(seed))

	// Get NPC templates for this genre
	templates := g.templates[params.GenreID]
	if templates == nil {
		templates = g.templates[""] // fallback to default
	}

	// Find NPC template
	var npcTemplate *EntityTemplate
	for i := range templates {
		if templates[i].BaseType == TypeNPC {
			npcTemplate = &templates[i]
			break
		}
	}

	if npcTemplate == nil {
		return nil, fmt.Errorf("no NPC template found for genre %s", params.GenreID)
	}

	// Generate base NPC entity
	merchantEntity := &Entity{
		Type:   TypeNPC,
		Size:   SizeMedium,
		Seed:   seed,
		Rarity: RarityCommon, // Merchants are always common rarity
		Tags:   []string{"merchant", "friendly", "trader"},
	}

	// Generate merchant name
	merchantEntity.Name = g.generateMerchantName(params.GenreID, rng)

	// Merchants are always level 1 (non-combatants)
	merchantEntity.Stats.Level = 1

	// Generate defensive stats (merchants don't fight)
	merchantEntity.Stats = g.generateStats(*npcTemplate, 1, RarityCommon, rng)
	// Override damage to zero - merchants don't attack
	merchantEntity.Stats.Damage = 0

	// Determine price multiplier based on type
	priceMultiplier := 1.5 // base markup
	if merchantType == MerchantNomadic {
		// Nomadic merchants charge more (they travel)
		priceMultiplier = 1.8
	}

	// Generate inventory using item generator
	inventorySize := 15 + rng.Intn(10) // 15-24 items
	inventory, err := g.generateMerchantInventory(seed+1000, params, inventorySize, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate merchant inventory: %w", err)
	}

	merchant := &MerchantData{
		Entity:            merchantEntity,
		MerchantType:      merchantType,
		Inventory:         inventory,
		PriceMultiplier:   priceMultiplier,
		BuyBackPercentage: 0.5, // Buy at 50% of value
		SpawnX:            0,   // Will be set by terrain/placement system
		SpawnY:            0,
	}

	if g.logger != nil {
		g.logger.WithFields(map[string]interface{}{
			"name":          merchantEntity.Name,
			"inventorySize": len(inventory),
			"type":          merchantType.String(),
		}).Info("merchant generated")
	}

	return merchant, nil
}

// generateMerchantName creates a genre-appropriate merchant name.
func (g *EntityGenerator) generateMerchantName(genreID string, rng *rand.Rand) string {
	names := MerchantNameTemplates[genreID]
	if names == nil {
		names = MerchantNameTemplates["fantasy"] // fallback
	}

	return names[rng.Intn(len(names))]
}

// generateMerchantInventory creates merchant stock using the item generator.
// Inventory is weighted toward consumables and common items with some rares.
func (g *EntityGenerator) generateMerchantInventory(seed int64, params procgen.GenerationParams, count int, rng *rand.Rand) ([]*item.Item, error) {
	itemGen := item.NewItemGenerator()

	inventory := make([]*item.Item, 0, count)

	for i := 0; i < count; i++ {
		// Use different seed for each item
		itemSeed := seed + int64(i)*100

		// Generate item with merchant-appropriate parameters
		// 60% consumables, 30% equipment, 10% rare items
		roll := rng.Float64()

		itemParams := params
		if itemParams.Custom == nil {
			itemParams.Custom = make(map[string]interface{})
		}

		if roll < 0.6 {
			// Consumable (potion, food, scroll)
			itemParams.Custom["type"] = "consumable"
		} else if roll < 0.9 {
			// Equipment (weapon, armor)
			itemParams.Custom["type"] = "equipment"
		} else {
			// Rare item (any type, higher quality)
			itemParams.Difficulty = 0.8 // Higher quality
		}

		// Generate single item
		itemParams.Custom["count"] = 1

		result, err := itemGen.Generate(itemSeed, itemParams)
		if err != nil {
			if g.logger != nil {
				g.logger.WithError(err).Warn("failed to generate merchant item, skipping")
			}
			continue
		}

		items, ok := result.([]*item.Item)
		if !ok || len(items) == 0 {
			continue
		}

		inventory = append(inventory, items[0])
	}

	return inventory, nil
}

// GenerateMerchantSpawnPoints creates deterministic spawn locations for merchants.
// For fixed merchants, returns locations in settlement areas.
// For nomadic merchants, returns wandering path points based on time seed.
func GenerateMerchantSpawnPoints(worldSeed int64, worldWidth, worldHeight int, merchantType MerchantType, count int) []struct{ X, Y float64 } {
	rng := rand.New(rand.NewSource(worldSeed))
	points := make([]struct{ X, Y float64 }, count)

	if merchantType == MerchantFixed {
		// Fixed merchants spawn in settlement areas (safe zones)
		// Place near center and corners for accessibility
		safeZones := []struct{ X, Y float64 }{
			{float64(worldWidth) * 0.5, float64(worldHeight) * 0.5},   // center
			{float64(worldWidth) * 0.25, float64(worldHeight) * 0.25}, // top-left
			{float64(worldWidth) * 0.75, float64(worldHeight) * 0.25}, // top-right
			{float64(worldWidth) * 0.25, float64(worldHeight) * 0.75}, // bottom-left
			{float64(worldWidth) * 0.75, float64(worldHeight) * 0.75}, // bottom-right
		}

		for i := 0; i < count; i++ {
			zone := safeZones[i%len(safeZones)]
			// Add some randomness within the zone
			offsetX := (rng.Float64() - 0.5) * 100
			offsetY := (rng.Float64() - 0.5) * 100
			points[i] = struct{ X, Y float64 }{
				X: zone.X + offsetX,
				Y: zone.Y + offsetY,
			}
		}
	} else {
		// Nomadic merchants spawn at random locations
		for i := 0; i < count; i++ {
			points[i] = struct{ X, Y float64 }{
				X: float64(rng.Intn(worldWidth)),
				Y: float64(rng.Intn(worldHeight)),
			}
		}
	}

	return points
}
