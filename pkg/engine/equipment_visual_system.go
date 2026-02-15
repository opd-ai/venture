// Package engine provides equipment visual system for updating equipment layers on sprites.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/palette"
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

// regenerateEquipmentLayers regenerates the entity sprite using the template
// pipeline with equipment overlay data, producing a high-quality aerial sprite
// that includes visible equipment.
func (s *EquipmentVisualSystem) regenerateEquipmentLayers(entity *Entity, equipComp *EquipmentVisualComponent, spriteComp *EbitenSprite) error {
	entitySeed := s.getEntitySeed(entity)
	genreID := s.getGenreID(entity)

	config := sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(spriteComp.Width),
		Height:     int(spriteComp.Height),
		Seed:       entitySeed,
		Complexity: 0.7,
		GenreID:    genreID,
		Custom:     make(map[string]interface{}),
	}

	config.Custom["useAerial"] = true
	config.Custom["entityType"] = s.resolveEntityType(entity)
	config.Custom["facing"] = s.resolveEntityFacing(entity)

	// Build equipment visuals for template pipeline overlay
	equipment := s.buildEquipmentLayers(entity, equipComp, genreID)
	if len(equipment) > 0 {
		config.Custom["equipmentVisuals"] = equipment
		config.Custom["hasWeapon"] = equipComp.HasWeapon()
		config.Custom["hasShield"] = false
	}

	img, err := s.spriteGenerator.Generate(config)
	if err != nil {
		return fmt.Errorf("failed to generate template sprite with equipment: %w", err)
	}

	spriteComp.Image = img
	return nil
}

// resolveEntityType determines the entity type string for template selection.
func (s *EquipmentVisualSystem) resolveEntityType(entity *Entity) string {
	if entity.HasComponent("input") {
		return "humanoid"
	}
	if cvComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok && cv.Form != FormHumanoid {
			return string(cv.Form)
		}
	}
	if roleComp, ok := entity.GetComponent("npc_role_visual"); ok {
		if npcRole, ok := roleComp.(*NpcRoleVisualComponent); ok && npcRole.Role != "" {
			return npcRole.Role
		}
	}
	return "humanoid"
}

// resolveEntityFacing returns the current facing direction for an entity.
func (s *EquipmentVisualSystem) resolveEntityFacing(entity *Entity) string {
	if animComp, ok := entity.GetComponent("animation"); ok {
		if anim, ok := animComp.(*AnimationComponent); ok && anim.LastFacing != "" {
			return anim.LastFacing
		}
	}
	return "down"
}

// buildCompositeConfig creates a composite configuration from components.
func (s *EquipmentVisualSystem) buildCompositeConfig(entity *Entity, equipComp *EquipmentVisualComponent, spriteComp *EbitenSprite) sprites.CompositeConfig {
	entitySeed := s.getEntitySeed(entity)
	genreID := s.getGenreID(entity)
	pal := s.generatePalette(genreID, entitySeed)

	baseConfig := s.createBaseConfig(spriteComp, entitySeed, genreID, pal)
	layers := s.createBaseLayers(baseConfig.Seed)
	equipment := s.buildEquipmentLayers(entity, equipComp, genreID)
	statusEffects := s.getStatusEffects(entity)

	return sprites.CompositeConfig{
		BaseConfig:    baseConfig,
		Layers:        layers,
		Equipment:     equipment,
		StatusEffects: statusEffects,
	}
}

// generatePalette generates a color palette for the entity.
func (s *EquipmentVisualSystem) generatePalette(genreID string, seed int64) *palette.Palette {
	pal, err := s.spriteGenerator.GetPaletteGenerator().Generate(genreID, seed)
	if err != nil {
		return nil
	}
	return pal
}

// createBaseConfig creates the base sprite configuration.
func (s *EquipmentVisualSystem) createBaseConfig(spriteComp *EbitenSprite, seed int64, genreID string, pal *palette.Palette) sprites.Config {
	return sprites.Config{
		Type:       sprites.SpriteEntity,
		Width:      int(spriteComp.Width),
		Height:     int(spriteComp.Height),
		Seed:       seed,
		Complexity: 0.5,
		GenreID:    genreID,
		Palette:    pal,
	}
}

// createBaseLayers creates the base body and head layers.
func (s *EquipmentVisualSystem) createBaseLayers(seed int64) []sprites.LayerConfig {
	return []sprites.LayerConfig{
		{
			Type:      sprites.LayerBody,
			ZIndex:    sprites.ZIndexBody,
			OffsetX:   0,
			OffsetY:   0,
			Scale:     1.0,
			Visible:   true,
			Seed:      seed,
			ShapeType: 0,
		},
		{
			Type:      sprites.LayerHead,
			ZIndex:    sprites.ZIndexHead,
			OffsetX:   0,
			OffsetY:   -8,
			Scale:     1.0,
			Visible:   true,
			Seed:      seed + 1,
			ShapeType: 0,
		},
	}
}

// buildEquipmentLayers creates equipment visual layers.
func (s *EquipmentVisualSystem) buildEquipmentLayers(entity *Entity, equipComp *EquipmentVisualComponent, genreID string) []sprites.EquipmentVisual {
	equipment := make([]sprites.EquipmentVisual, 0)
	equipmentComp := s.getEquipmentComponent(entity)

	s.addWeaponLayer(&equipment, equipComp, equipmentComp, genreID)
	s.addArmorLayer(&equipment, equipComp, equipmentComp, genreID)
	s.addAccessoryLayers(&equipment, equipComp, equipmentComp, genreID)

	return equipment
}

// addWeaponLayer adds weapon equipment visual if present.
func (s *EquipmentVisualSystem) addWeaponLayer(equipment *[]sprites.EquipmentVisual, equipComp *EquipmentVisualComponent, equipmentComp *EquipmentComponent, genreID string) {
	if equipComp.HasWeapon() && equipComp.ShowWeapon {
		weaponItem := s.getEquippedItem(equipmentComp, SlotMainHand)
		*equipment = append(*equipment, s.buildEquipmentVisual(
			"weapon",
			equipComp.WeaponID,
			equipComp.WeaponSeed,
			sprites.LayerWeapon,
			weaponItem,
			genreID,
		))
	}
}

// addArmorLayer adds armor equipment visual if present.
func (s *EquipmentVisualSystem) addArmorLayer(equipment *[]sprites.EquipmentVisual, equipComp *EquipmentVisualComponent, equipmentComp *EquipmentComponent, genreID string) {
	if equipComp.HasArmor() && equipComp.ShowArmor {
		armorItem := s.getEquippedItem(equipmentComp, SlotChest)
		*equipment = append(*equipment, s.buildEquipmentVisual(
			"armor",
			equipComp.ArmorID,
			equipComp.ArmorSeed,
			sprites.LayerArmor,
			armorItem,
			genreID,
		))
	}
}

// addAccessoryLayers adds accessory equipment visuals if present.
func (s *EquipmentVisualSystem) addAccessoryLayers(equipment *[]sprites.EquipmentVisual, equipComp *EquipmentVisualComponent, equipmentComp *EquipmentComponent, genreID string) {
	if !equipComp.HasAccessories() || !equipComp.ShowAccessories {
		return
	}

	accessorySlots := []EquipmentSlot{SlotAccessory1, SlotAccessory2, SlotAccessory3}
	for i, accessoryID := range equipComp.AccessoryIDs {
		var accessoryItem *item.Item
		if equipmentComp != nil && i < len(accessorySlots) {
			accessoryItem = s.getEquippedItem(equipmentComp, accessorySlots[i])
		}
		*equipment = append(*equipment, s.buildEquipmentVisual(
			"accessory",
			accessoryID,
			equipComp.AccessorySeeds[i],
			sprites.LayerAccessory,
			accessoryItem,
			genreID,
		))
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

	equipComp := s.getEquipmentComponentTyped(entity)
	if equipComp == nil {
		return
	}

	s.syncMainHandSlot(equipComp, equipVisualComp)
	s.syncChestSlot(equipComp, equipVisualComp)
	s.syncAccessories(equipComp, equipVisualComp)
}

// getEquipmentComponentTyped retrieves and type-asserts the equipment component.
func (s *EquipmentVisualSystem) getEquipmentComponentTyped(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// syncMainHandSlot synchronizes the main hand weapon slot.
func (s *EquipmentVisualSystem) syncMainHandSlot(equipComp *EquipmentComponent, equipVisualComp *EquipmentVisualComponent) {
	mainHand := equipComp.GetEquipped(SlotMainHand)
	if mainHand != nil {
		if equipVisualComp.WeaponID != mainHand.ID {
			equipVisualComp.SetWeapon(mainHand.ID, mainHand.Seed)
		}
	} else if equipVisualComp.HasWeapon() {
		equipVisualComp.ClearWeapon()
	}
}

// syncChestSlot synchronizes the chest armor slot.
func (s *EquipmentVisualSystem) syncChestSlot(equipComp *EquipmentComponent, equipVisualComp *EquipmentVisualComponent) {
	chest := equipComp.GetEquipped(SlotChest)
	if chest != nil {
		if equipVisualComp.ArmorID != chest.ID {
			equipVisualComp.SetArmor(chest.ID, chest.Seed)
		}
	} else if equipVisualComp.HasArmor() {
		equipVisualComp.ClearArmor()
	}
}

// syncAccessories synchronizes accessory slots with visual component.
func (s *EquipmentVisualSystem) syncAccessories(equipComp *EquipmentComponent, equipVisualComp *EquipmentVisualComponent) {
	// Get currently equipped accessories
	equippedAccessories := make([]*item.Item, 0, 3)
	accessorySlots := []EquipmentSlot{SlotAccessory1, SlotAccessory2, SlotAccessory3}

	for _, slot := range accessorySlots {
		acc := equipComp.GetEquipped(slot)
		if acc != nil {
			equippedAccessories = append(equippedAccessories, acc)
		}
	}

	// Check if accessories have changed
	changed := len(equippedAccessories) != len(equipVisualComp.AccessoryIDs)
	if !changed {
		// Check if any individual accessory has changed
		for i, acc := range equippedAccessories {
			if i >= len(equipVisualComp.AccessoryIDs) || equipVisualComp.AccessoryIDs[i] != acc.ID {
				changed = true
				break
			}
		}
	}

	// If changed, rebuild accessory list
	if changed {
		equipVisualComp.ClearAccessories()
		for _, acc := range equippedAccessories {
			equipVisualComp.AddAccessory(acc.ID, acc.Seed)
		}
	}
}

// buildEquipmentVisual creates an EquipmentVisual with all Phase 15.3 properties.
func (s *EquipmentVisualSystem) buildEquipmentVisual(slot, itemID string, seed int64, layer sprites.LayerType, itm *item.Item, genreID string) sprites.EquipmentVisual {
	visual := sprites.EquipmentVisual{
		Slot:        slot,
		ItemID:      itemID,
		Seed:        seed,
		Layer:       layer,
		Material:    sprites.MaterialMetal,
		DamageState: sprites.DamageStatePristine,
		Enchantment: sprites.EnchantmentGlow{Active: false},
		DetailLevel: 0.5,
		Params:      make(map[string]interface{}),
	}

	// If we have item data, populate enhanced properties
	if itm != nil {
		// Determine material type
		visual.Material = s.getMaterialType(itm, slot, genreID)

		// Calculate damage state from durability
		visual.DamageState = sprites.GetDamageStateFromDurability(itm.Stats.Durability, itm.Stats.DurabilityMax)

		// Add enchantment glow based on rarity
		visual.Enchantment = sprites.GetEnchantmentFromRarity(itm.Rarity.String())

		// Set detail level based on rarity
		visual.DetailLevel = sprites.GetDetailLevelFromRarity(itm.Rarity.String())
	}

	return visual
}

// getMaterialType determines the material type for an item.
func (s *EquipmentVisualSystem) getMaterialType(itm *item.Item, slot, genreID string) sprites.MaterialType {
	// First check tags for explicit material
	material := sprites.GetMaterialTypeFromTags(itm.Tags, genreID)

	// If no tag match, use item type-based defaults
	if material == sprites.MaterialMetal && len(itm.Tags) == 0 {
		switch itm.Type {
		case item.TypeWeapon:
			material = sprites.GetMaterialTypeFromWeaponType(itm.WeaponType.String(), genreID)
		case item.TypeArmor:
			material = sprites.GetMaterialTypeFromArmorType(itm.ArmorType.String(), genreID)
		}
	}

	return material
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *EquipmentVisualSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// getEquippedItem gets an item from a specific equipment slot.
func (s *EquipmentVisualSystem) getEquippedItem(equipComp *EquipmentComponent, slot EquipmentSlot) *item.Item {
	if equipComp == nil {
		return nil
	}
	return equipComp.GetEquipped(slot)
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
