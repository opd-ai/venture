// Package engine provides equipment visual system for updating equipment layers on sprites.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// EquipmentVisualSystem updates equipment visual components and regenerates equipment layers.
type EquipmentVisualSystem struct {
	spriteGenerator *sprites.Generator
}

// NewEquipmentVisualSystem creates a new equipment visual system.
func NewEquipmentVisualSystem(spriteGenerator *sprites.Generator) *EquipmentVisualSystem {
	return &EquipmentVisualSystem{
		spriteGenerator: spriteGenerator,
	}
}

// Update processes all entities with equipment visual components.
func (s *EquipmentVisualSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// First, sync equipment visual component with equipment component changes
		s.syncEquipmentChanges(entity)

		equipComp := s.getEquipmentVisualComponent(entity)
		if equipComp == nil {
			continue
		}

		// Skip if not dirty
		if !equipComp.Dirty {
			continue
		}

		// Get sprite component for base configuration
		spriteComp := s.getSpriteComponent(entity)
		if spriteComp == nil {
			continue
		}

		// Regenerate equipment layers
		if err := s.regenerateEquipmentLayers(entity, equipComp, spriteComp); err != nil {
			// Log error but continue processing other entities
			// Note: This system doesn't have a logger field, so we silently continue
			// In production, errors are rare (usually programming errors) and should be caught in testing
			continue
		}

		// Mark as clean
		equipComp.MarkClean()
	}
}

// regenerateEquipmentLayers creates new equipment visual layers.
func (s *EquipmentVisualSystem) regenerateEquipmentLayers(entity *Entity, equipComp *EquipmentVisualComponent, spriteComp *EbitenSprite) error {
	// Build composite config
	compositeConfig := s.buildCompositeConfig(entity, equipComp, spriteComp)

	// Generate composite sprite
	compositeImg, err := s.spriteGenerator.GenerateComposite(compositeConfig)
	if err != nil {
		return fmt.Errorf("failed to generate composite: %w", err)
	}

	// Update sprite component with composite image
	spriteComp.Image = compositeImg

	return nil
}

// buildCompositeConfig creates a composite configuration from components.
func (s *EquipmentVisualSystem) buildCompositeConfig(entity *Entity, equipComp *EquipmentVisualComponent, spriteComp *EbitenSprite) sprites.CompositeConfig {
	// Get or create palette
	entitySeed := s.getEntitySeed(entity)
	genreID := s.getGenreID(entity)

	// Generate palette for base config
	pal, err := s.spriteGenerator.GetPaletteGenerator().Generate(genreID, entitySeed)
	if err != nil {
		// Fallback to default palette if generation fails
		pal = nil
	}

	// Base sprite config
	baseConfig := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(spriteComp.Width),
		Height:     int(spriteComp.Height),
		Seed:       entitySeed,
		Complexity: 0.5,
		GenreID:    genreID,
		Palette:    pal,
	}

	// Build layers (always include body)
	layers := []sprites.LayerConfig{
		{
			Type:      sprites.LayerBody,
			ZIndex:    10,
			OffsetX:   0,
			OffsetY:   0,
			Scale:     1.0,
			Visible:   true,
			Seed:      baseConfig.Seed,
			ShapeType: 0, // Circle
		},
		{
			Type:      sprites.LayerHead,
			ZIndex:    20,
			OffsetX:   0,
			OffsetY:   -8,
			Scale:     1.0,
			Visible:   true,
			Seed:      baseConfig.Seed + 1,
			ShapeType: 0, // Circle
		},
	}

	// Build equipment visuals
	equipment := make([]sprites.EquipmentVisual, 0)

	if equipComp.HasWeapon() && equipComp.ShowWeapon {
		equipment = append(equipment, sprites.EquipmentVisual{
			Slot:   "weapon",
			ItemID: equipComp.WeaponID,
			Seed:   equipComp.WeaponSeed,
			Layer:  sprites.LayerWeapon,
			Params: make(map[string]interface{}),
		})
	}

	if equipComp.HasArmor() && equipComp.ShowArmor {
		equipment = append(equipment, sprites.EquipmentVisual{
			Slot:   "armor",
			ItemID: equipComp.ArmorID,
			Seed:   equipComp.ArmorSeed,
			Layer:  sprites.LayerArmor,
			Params: make(map[string]interface{}),
		})
	}

	if equipComp.HasAccessories() && equipComp.ShowAccessories {
		for i, accessoryID := range equipComp.AccessoryIDs {
			equipment = append(equipment, sprites.EquipmentVisual{
				Slot:   "accessory",
				ItemID: accessoryID,
				Seed:   equipComp.AccessorySeeds[i],
				Layer:  sprites.LayerAccessory,
				Params: make(map[string]interface{}),
			})
		}
	}

	// Get status effects if available
	statusEffects := s.getStatusEffects(entity)

	return sprites.CompositeConfig{
		BaseConfig:    baseConfig,
		Layers:        layers,
		Equipment:     equipment,
		StatusEffects: statusEffects,
	}
}

// getEntitySeed gets a deterministic seed for the entity.
func (s *EquipmentVisualSystem) getEntitySeed(entity *Entity) int64 {
	// Use entity ID as base seed
	return int64(entity.ID)
}

// getGenreID gets the genre ID for the entity.
func (s *EquipmentVisualSystem) getGenreID(entity *Entity) string {
	// Try to get genre from entity component
	if genreComp, hasGenre := entity.GetComponent("genre"); hasGenre {
		if genre, ok := genreComp.(*GenreComponent); ok {
			return genre.GetPrimaryGenre()
		}
	}

	// Default to fantasy genre if not specified
	return "fantasy"
}

// getStatusEffects extracts status effects from entity components.
func (s *EquipmentVisualSystem) getStatusEffects(entity *Entity) []sprites.StatusEffect {
	effects := make([]sprites.StatusEffect, 0)

	// Get all components and check for status effects
	for name, comp := range entity.Components {
		if name == "status_effect" {
			if effectComp, ok := comp.(*StatusEffectComponent); ok {
				// Convert engine status effect to sprite status effect
				effect := sprites.StatusEffect{
					Type:          effectComp.EffectType,
					Intensity:     effectComp.Magnitude,
					Color:         s.getEffectColor(effectComp.EffectType),
					AnimSpeed:     1.0,
					ParticleCount: int(effectComp.Magnitude * 10), // Scale particles by magnitude
				}
				effects = append(effects, effect)
			}
		}
	}

	return effects
}

// getEffectColor returns the visual color for a status effect type.
func (s *EquipmentVisualSystem) getEffectColor(effectType string) string {
	switch effectType {
	case "poison":
		return "green"
	case "burning", "fire":
		return "red"
	case "frozen", "ice":
		return "cyan"
	case "stunned":
		return "yellow"
	case "bleeding":
		return "darkred"
	case "blessed", "heal":
		return "gold"
	case "cursed":
		return "purple"
	default:
		return "white"
	}
}

// syncEquipmentChanges updates the equipment visual component based on changes in the equipment component.
func (s *EquipmentVisualSystem) syncEquipmentChanges(entity *Entity) {
	equipVisualComp := s.getEquipmentVisualComponent(entity)
	if equipVisualComp == nil {
		return
	}

	// Get equipment component to check for changes
	comp, ok := entity.GetComponent("equipment")
	if !ok {
		return
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return
	}

	// Check each equipment slot for changes and update visual component
	mainHand := equipComp.GetEquipped(SlotMainHand)
	if mainHand != nil {
		// Use item ID as unique identifier and item seed for generation
		itemID := mainHand.ID
		itemSeed := mainHand.Seed
		if equipVisualComp.WeaponID != itemID {
			equipVisualComp.SetWeapon(itemID, itemSeed)
		}
	} else if equipVisualComp.HasWeapon() {
		equipVisualComp.ClearWeapon()
	}

	// Check armor (chest slot is primary armor visual)
	chest := equipComp.GetEquipped(SlotChest)
	if chest != nil {
		itemID := chest.ID
		itemSeed := chest.Seed
		if equipVisualComp.ArmorID != itemID {
			equipVisualComp.SetArmor(itemID, itemSeed)
		}
	} else if equipVisualComp.HasArmor() {
		equipVisualComp.ClearArmor()
	}

	// TODO: Add accessory syncing when more equipment slots are used
}

// Helper methods

func (s *EquipmentVisualSystem) getEquipmentVisualComponent(entity *Entity) *EquipmentVisualComponent {
	comp, ok := entity.GetComponent("equipment_visual")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentVisualComponent)
	if !ok {
		return nil
	}
	return equipComp
}

func (s *EquipmentVisualSystem) getSpriteComponent(entity *Entity) *EbitenSprite {
	comp, ok := entity.GetComponent("sprite")
	if !ok || comp == nil {
		return nil
	}
	spriteComp, ok := comp.(*EbitenSprite)
	if !ok {
		return nil
	}
	return spriteComp
}

// EquipItem updates equipment visuals when an item is equipped.
func (s *EquipmentVisualSystem) EquipItem(entity *Entity, slot, itemID string, seed int64) {
	equipComp := s.getEquipmentVisualComponent(entity)
	if equipComp == nil {
		// Create component if it doesn't exist
		equipComp = NewEquipmentVisualComponent()
		entity.AddComponent(equipComp)
	}

	switch slot {
	case "weapon":
		equipComp.SetWeapon(itemID, seed)
	case "armor":
		equipComp.SetArmor(itemID, seed)
	case "accessory":
		equipComp.AddAccessory(itemID, seed)
	}
}

// UnequipItem removes equipment visuals when an item is unequipped.
func (s *EquipmentVisualSystem) UnequipItem(entity *Entity, slot string) {
	equipComp := s.getEquipmentVisualComponent(entity)
	if equipComp == nil {
		return
	}

	switch slot {
	case "weapon":
		equipComp.ClearWeapon()
	case "armor":
		equipComp.ClearArmor()
	case "accessories":
		equipComp.ClearAccessories()
	}
}
