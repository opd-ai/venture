// Package engine provides the inventory management system.
// This file implements InventorySystem which handles item management,
// equipment, and inventory operations for entities.
package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// Scroll spell effect constants
const (
	// DefaultScrollEffectDuration is the default duration for scroll spell effects in seconds.
	DefaultScrollEffectDuration = 5.0
	// DefaultScrollEffectRadius is the default radius for area-targeting scrolls.
	DefaultScrollEffectRadius = 64.0
)

// InventorySystem manages inventory and equipment operations.
type InventorySystem struct {
	world             *World
	logger            *logrus.Entry
	spellEffectSystem *SpellEffectSystem
}

// NewInventorySystem creates a new inventory system.
func NewInventorySystem(world *World) *InventorySystem {
	return NewInventorySystemWithLogger(world, nil)
}

// NewInventorySystemWithLogger creates a new inventory system with a logger.
func NewInventorySystemWithLogger(world *World, logger *logrus.Logger) *InventorySystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "inventory")
	}
	return &InventorySystem{
		world:  world,
		logger: logEntry,
	}
}

// SetSpellEffectSystem sets the spell effect system for consumable spell triggers.
// Gap A2: Consumable Spell Effect Activation - scrolls trigger spell effects when used.
func (s *InventorySystem) SetSpellEffectSystem(spellSystem *SpellEffectSystem) {
	s.spellEffectSystem = spellSystem
}

// AddItemToInventory adds an item to an entity's inventory.
// Returns true if successful, false if inventory is full.
func (s *InventorySystem) AddItemToInventory(entityID uint64, itm *item.Item) (bool, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return false, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return false, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	success := invComp.AddItem(itm)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID": entityID,
			"itemName": itm.Name,
			"itemType": itm.Type.String(),
			"success":  success,
		}).Debug("adding item to inventory")
	}

	return success, nil
}

// RemoveItemFromInventory removes an item from inventory by index.
func (s *InventorySystem) RemoveItemFromInventory(entityID uint64, index int) (*item.Item, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return nil, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	itm := invComp.RemoveItem(index)
	if itm == nil {
		return nil, fmt.Errorf("invalid item index %d", index)
	}

	return itm, nil
}

// EquipItem equips an item from inventory to the appropriate slot.
// The item is removed from inventory and placed in equipment.
func (s *InventorySystem) EquipItem(entityID uint64, inventoryIndex int) error {
	invComp, equipComp, err := s.getInventoryAndEquipment(entityID)
	if err != nil {
		return err
	}

	itm, err := s.validateInventoryIndex(invComp, inventoryIndex)
	if err != nil {
		return err
	}

	entity, _ := s.world.GetEntity(entityID)
	if err := s.validateClassRestrictions(entity, itm); err != nil {
		return err
	}

	slot, canEquip := equipComp.GetSlotForItem(itm)
	if !canEquip {
		return fmt.Errorf("item %s cannot be equipped", itm.Name)
	}

	previousItem := equipComp.Equip(itm, slot)
	invComp.RemoveItem(inventoryIndex)

	if err := s.handlePreviousItem(invComp, equipComp, previousItem, itm, slot, inventoryIndex); err != nil {
		return err
	}

	s.applyEquipmentStats(entityID)
	s.logEquipAction(entityID, itm, slot, previousItem)
	return nil
}

// getInventoryAndEquipment retrieves inventory and equipment components for an entity.
func (s *InventorySystem) getInventoryAndEquipment(entityID uint64) (*InventoryComponent, *EquipmentComponent, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return nil, nil, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil, nil, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return nil, nil, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	comp2, ok := entity.GetComponent("equipment")
	if !ok {
		return nil, nil, fmt.Errorf("entity %d does not have equipment component", entityID)
	}
	equipComp, ok := comp2.(*EquipmentComponent)
	if !ok {
		return nil, nil, fmt.Errorf("entity %d equipment component has wrong type", entityID)
	}

	return invComp, equipComp, nil
}

// validateInventoryIndex checks if inventory index is valid and returns the item.
func (s *InventorySystem) validateInventoryIndex(invComp *InventoryComponent, index int) (*item.Item, error) {
	if index < 0 || index >= len(invComp.Items) {
		return nil, fmt.Errorf("invalid inventory index %d", index)
	}
	return invComp.Items[index], nil
}

// validateClassRestrictions checks if entity's class can use the item.
func (s *InventorySystem) validateClassRestrictions(entity *Entity, itm *item.Item) error {
	comp, ok := entity.GetComponent("class_progression")
	if !ok {
		return nil
	}

	classComp, ok := comp.(*ClassProgressionComponent)
	if !ok {
		return nil
	}

	primaryClassName := classComp.Class.LowerName()
	if itm.CanBeUsedByClass(primaryClassName) {
		return nil
	}

	if classComp.SecondaryClass != nil {
		secondaryClassName := classComp.SecondaryClass.LowerName()
		if itm.CanBeUsedByClass(secondaryClassName) {
			return nil
		}
	}

	return fmt.Errorf("item %s cannot be equipped by %s", itm.Name, classComp.Class.String())
}

// handlePreviousItem manages the previously equipped item when swapping equipment.
func (s *InventorySystem) handlePreviousItem(invComp *InventoryComponent, equipComp *EquipmentComponent, previousItem, newItem *item.Item, slot EquipmentSlot, inventoryIndex int) error {
	if previousItem == nil {
		return nil
	}

	if invComp.AddItem(previousItem) {
		return nil
	}

	equipComp.Equip(previousItem, slot)
	invComp.Items = append(invComp.Items[:inventoryIndex],
		append([]*item.Item{newItem}, invComp.Items[inventoryIndex:]...)...)
	return fmt.Errorf("cannot equip: inventory full for swapped item")
}

// logEquipAction logs the equipment action if logger is configured.
func (s *InventorySystem) logEquipAction(entityID uint64, itm *item.Item, slot EquipmentSlot, previousItem *item.Item) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":     entityID,
			"itemName":     itm.Name,
			"slot":         slot.String(),
			"previousItem": previousItem != nil,
		}).Info("item equipped")
	}
}

// UnequipItem removes an item from an equipment slot and adds it to inventory.
func (s *InventorySystem) UnequipItem(entityID uint64, slot EquipmentSlot) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	comp2, ok := entity.GetComponent("equipment")
	if !ok {
		return fmt.Errorf("entity %d does not have equipment component", entityID)
	}
	equipComp, ok := comp2.(*EquipmentComponent)
	if !ok {
		return fmt.Errorf("entity %d equipment component has wrong type", entityID)
	}

	// Unequip the item
	itm := equipComp.Unequip(slot)
	if itm == nil {
		return fmt.Errorf("no item equipped in slot %s", slot.String())
	}

	// Add to inventory
	if !invComp.AddItem(itm) {
		// Inventory is full, re-equip the item
		equipComp.Equip(itm, slot)
		return fmt.Errorf("cannot unequip: inventory full")
	}

	// Update entity stats
	s.applyEquipmentStats(entityID)

	return nil
}

// UseConsumable uses a consumable item from inventory.
// The item is removed from inventory after use.
func (s *InventorySystem) UseConsumable(entityID uint64, inventoryIndex int) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Get item from inventory
	if inventoryIndex < 0 || inventoryIndex >= len(invComp.Items) {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}
	itm := invComp.Items[inventoryIndex]

	// Check if item is consumable
	if !itm.IsConsumable() {
		return fmt.Errorf("item %s is not consumable", itm.Name)
	}

	// Apply consumable effects
	if err := s.applyConsumableEffects(entityID, itm); err != nil {
		return fmt.Errorf("failed to use consumable: %w", err)
	}

	// Remove from inventory
	invComp.RemoveItem(inventoryIndex)

	return nil
}

// applyConsumableEffects applies the effects of a consumable item to an entity.
func (s *InventorySystem) applyConsumableEffects(entityID uint64, itm *item.Item) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	// Get health component if it exists
	comp, hasHealth := entity.GetComponent("health")
	var healthComp *HealthComponent
	if hasHealth {
		// Type assert with safety check
		if hc, ok := comp.(*HealthComponent); ok {
			healthComp = hc
		}
	}

	// Apply effects based on consumable type
	switch itm.ConsumableType {
	case item.ConsumablePotion:
		// Health potions restore health
		if healthComp != nil {
			// Restore health based on item value (simple implementation)
			healAmount := float64(itm.Stats.Value) / 10.0
			healthComp.Heal(healAmount)
		}

	case item.ConsumableScroll:
		// Gap A2 RESOLVED: Consumable Spell Effect Activation
		// Scrolls trigger spell effects when used based on SpellEffectID
		s.applyScrollSpellEffect(entity, itm)

	case item.ConsumableFood:
		// Food restores health over time
		if healthComp != nil {
			healAmount := float64(itm.Stats.Value) / 20.0
			healthComp.Heal(healAmount)
		}

	case item.ConsumableBomb:
		// Bombs would deal area damage
		// This would require position and collision detection
		// Placeholder for now
	}

	return nil
}

// applyEquipmentStats updates an entity's stats based on equipped items.
func (s *InventorySystem) applyEquipmentStats(entityID uint64) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return
	}

	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	s.applyDefenseStats(entity, equipComp)
	s.applyAttackStats(entity, equipComp)
}

// getEquipmentComponent retrieves and validates the equipment component.
func (s *InventorySystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok || equipComp == nil {
		return nil
	}
	return equipComp
}

// applyDefenseStats updates the entity's defense stat based on equipment.
func (s *InventorySystem) applyDefenseStats(entity *Entity, equipComp *EquipmentComponent) {
	comp, ok := entity.GetComponent("stats")
	if !ok {
		return
	}
	statsComp, ok := comp.(*StatsComponent)
	if !ok {
		return
	}
	equipStats := equipComp.GetStats()
	statsComp.Defense = float64(equipStats.Defense)
}

// applyAttackStats updates the entity's attack component based on equipped weapon.
func (s *InventorySystem) applyAttackStats(entity *Entity, equipComp *EquipmentComponent) {
	comp, ok := entity.GetComponent("attack")
	if !ok {
		return
	}
	attackComp, ok := comp.(*AttackComponent)
	if !ok {
		return
	}

	s.applyWeaponDamage(attackComp, equipComp)
	s.applyWeaponSpeed(attackComp, equipComp)
	s.applyWeaponDamageType(attackComp, equipComp)
}

// applyWeaponDamage sets the attack damage from the equipped weapon.
func (s *InventorySystem) applyWeaponDamage(attackComp *AttackComponent, equipComp *EquipmentComponent) {
	weaponDamage := equipComp.GetWeaponDamage()
	if weaponDamage > 0 {
		attackComp.Damage = float64(weaponDamage)
	}
}

// applyWeaponSpeed sets the attack cooldown from the equipped weapon speed.
func (s *InventorySystem) applyWeaponSpeed(attackComp *AttackComponent, equipComp *EquipmentComponent) {
	weaponSpeed := equipComp.GetWeaponSpeed()
	if weaponSpeed > 0 {
		attackComp.Cooldown = 1.0 / weaponSpeed
	}
}

// applyWeaponDamageType sets the damage type based on the equipped weapon.
func (s *InventorySystem) applyWeaponDamageType(attackComp *AttackComponent, equipComp *EquipmentComponent) {
	mainHand := equipComp.GetEquipped(SlotMainHand)
	if mainHand == nil {
		return
	}
	switch mainHand.WeaponType {
	case item.WeaponStaff:
		attackComp.DamageType = combat.DamageMagical
	default:
		attackComp.DamageType = combat.DamagePhysical
	}
}

// DropItem removes an item from inventory and places it in the world.
// The item is spawned as a physical entity at the entity's current position
// that can be picked up by players.
func (s *InventorySystem) DropItem(entityID uint64, inventoryIndex int) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Check inventory index is valid before removing
	if inventoryIndex < 0 || inventoryIndex >= len(invComp.Items) || invComp.Items[inventoryIndex] == nil {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}

	// Get entity position to spawn dropped item BEFORE removing from inventory
	posComp, ok := entity.GetComponent("position")
	if !ok {
		// Entity has no position - can't drop item in world
		return fmt.Errorf("entity %d has no position component, cannot drop item", entityID)
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return fmt.Errorf("entity %d has invalid position component type", entityID)
	}

	// Remove item from inventory (only after we know we can drop it)
	itm := invComp.RemoveItem(inventoryIndex)
	if itm == nil {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}

	// Spawn item entity in the world at entity's position
	// The SpawnItemInWorld function creates a collectable item entity
	// with collision detection for automatic pickup
	SpawnItemInWorld(s.world, itm, pos.X, pos.Y)

	return nil
}

// TransferItem moves an item from one entity's inventory to another's.
func (s *InventorySystem) TransferItem(fromEntityID, toEntityID uint64, inventoryIndex int) error {
	// Get source entity
	fromEntity, ok := s.world.GetEntity(fromEntityID)
	if !ok {
		return fmt.Errorf("source entity %d not found", fromEntityID)
	}

	comp, ok := fromEntity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("source entity %d does not have inventory component", fromEntityID)
	}
	fromInv, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("source entity %d inventory component has wrong type", fromEntityID)
	}

	// Get destination entity
	toEntity, ok := s.world.GetEntity(toEntityID)
	if !ok {
		return fmt.Errorf("destination entity %d not found", toEntityID)
	}

	comp2, ok := toEntity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("destination entity %d does not have inventory component", toEntityID)
	}
	toInv, ok := comp2.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("destination entity %d inventory component has wrong type", toEntityID)
	}

	// Get item from source inventory
	if inventoryIndex < 0 || inventoryIndex >= len(fromInv.Items) {
		return fmt.Errorf("invalid inventory index %d", inventoryIndex)
	}
	itm := fromInv.Items[inventoryIndex]

	// Check if destination can accept the item
	if !toInv.CanAddItem(itm) {
		return fmt.Errorf("destination inventory cannot accept item")
	}

	// Transfer the item
	fromInv.RemoveItem(inventoryIndex)
	toInv.AddItem(itm)

	return nil
}

// GetInventoryValue returns the total value of all items in an entity's inventory.
func (s *InventorySystem) GetInventoryValue(entityID uint64) (int, error) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return 0, fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return 0, fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return 0, fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	totalValue := invComp.Gold
	for _, itm := range invComp.Items {
		totalValue += itm.GetValue()
	}

	return totalValue, nil
}

// SortInventoryByValue sorts inventory items by value (descending).
func (s *InventorySystem) SortInventoryByValue(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort (good enough for small inventories)
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].GetValue() < items[j+1].GetValue() {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// SortInventoryByWeight sorts inventory items by weight (ascending).
func (s *InventorySystem) SortInventoryByWeight(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].Stats.Weight > items[j+1].Stats.Weight {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// SortInventoryByType sorts inventory items by type.
func (s *InventorySystem) SortInventoryByType(entityID uint64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return fmt.Errorf("entity %d not found", entityID)
	}

	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return fmt.Errorf("entity %d does not have inventory component", entityID)
	}
	invComp, ok := comp.(*InventoryComponent)
	if !ok {
		return fmt.Errorf("entity %d inventory component has wrong type", entityID)
	}

	// Simple bubble sort by type
	items := invComp.Items
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if items[j].Type > items[j+1].Type {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}

	return nil
}

// Update implements the System interface.
// InventorySystem doesn't need per-frame updates, so this is a no-op.
func (s *InventorySystem) Update(entities []*Entity, deltaTime float64) {
	// GAP-006 REPAIR: Ensure equipment stats are recalculated when dirty
	for _, entity := range entities {
		// Only process entities with equipment component
		equipComp, hasEquip := entity.GetComponent("equipment")
		if !hasEquip {
			continue
		}

		equipment, ok := equipComp.(*EquipmentComponent)
		if !ok {
			continue
		}

		// Recalculate equipment stats if dirty
		// This ensures CachedStats is up-to-date for CharacterUI display
		// and combat calculations
		if equipment.StatsDirty {
			equipment.RecalculateStats()
		}

		// Note: Equipment stats are now available via equipment.CachedStats
		// CharacterUI reads equipment.CachedStats to show bonuses
		// Combat system uses GetWeaponDamage() and GetTotalDefense()
		// No need to modify base StatsComponent - equipment provides bonuses separately
	}
}

// applyScrollSpellEffect applies a spell effect from a scroll consumable.
// Gap A2: Consumable Spell Effect Activation - scrolls trigger spell effects when used.
func (s *InventorySystem) applyScrollSpellEffect(entity *Entity, itm *item.Item) {
	// Skip if no spell effect system is available
	if s.spellEffectSystem == nil {
		if s.logger != nil {
			s.logger.WithField("item", itm.Name).Debug("no spell effect system available for scroll")
		}
		return
	}

	// Skip if scroll has no spell effect ID
	if itm.SpellEffectID == "" {
		if s.logger != nil {
			s.logger.WithField("item", itm.Name).Debug("scroll has no spell effect ID")
		}
		return
	}

	// Get entity position for targeting
	var targetX, targetY float64
	if posComp, hasPos := entity.GetComponent("position"); hasPos {
		if pos, ok := posComp.(*PositionComponent); ok && pos != nil {
			targetX = pos.X
			targetY = pos.Y
		}
	}

	// Map spell effect ID to effect type and get default target type
	effectType, defaultTargetType := s.mapSpellEffectIDWithTarget(itm.SpellEffectID)

	// Calculate magnitude based on item value/rarity
	magnitude := s.calculateScrollMagnitude(itm)

	// Determine duration: use item's SpellDuration if set, otherwise use default
	duration := itm.SpellDuration
	if duration <= 0 {
		duration = DefaultScrollEffectDuration
	}

	// Determine target type: use item's SpellTargetType if set, otherwise use spell-based default
	targetType := s.parseTargetType(itm.SpellTargetType, defaultTargetType)

	// Determine radius: use item's SpellRadius if set, otherwise use default for area spells
	radius := itm.SpellRadius
	if radius <= 0 && targetType == TargetArea {
		radius = DefaultScrollEffectRadius
	}

	// Apply the spell effect to the entity
	s.spellEffectSystem.ApplySpellEffect(
		entity,
		effectType,
		magnitude,
		duration,
		targetType,
		entity.ID, // Caster is the user
		targetX,
		targetY,
		radius,
	)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"item":        itm.Name,
			"spell_id":    itm.SpellEffectID,
			"effect_type": effectType.String(),
			"target_type": targetType.String(),
			"magnitude":   magnitude,
			"duration":    duration,
			"radius":      radius,
		}).Debug("scroll spell effect applied")
	}
}

// mapSpellEffectIDWithTarget converts a spell effect ID string to an EffectType and default TargetType.
func (s *InventorySystem) mapSpellEffectIDWithTarget(spellID string) (EffectType, TargetType) {
	switch spellID {
	case "fireball", "lightning", "ice":
		return EffectElementalFusion, TargetArea // Offensive spells target area
	case "protection", "shield":
		return EffectMetamagic, TargetSelf // Defensive spells target self
	case "teleportation", "blink":
		return EffectTeleportation, TargetSelf // Movement spells target self
	case "haste", "slow":
		return EffectTimeManipulation, TargetSelf // Buff/debuff target self (or could be entity)
	case "levitation", "gravity":
		return EffectGravityControl, TargetSelf // Movement modifiers target self
	case "heal", "drain":
		return EffectLifeDrain, TargetSelf // Healing targets self
	case "summon":
		return EffectSummoning, TargetArea // Summons appear in an area
	case "invisibility", "decoy":
		return EffectIllusion, TargetSelf // Illusions affect self
	case "wall", "pit", "bridge":
		return EffectTerrainManipulation, TargetTerrain // Terrain spells target terrain
	case "transmute":
		return EffectTransmutation, TargetTerrain // Transmutation affects terrain/objects
	default:
		// Default to elemental fusion with area targeting for unknown spell IDs
		return EffectElementalFusion, TargetArea
	}
}

// parseTargetType converts a string target type to TargetType, with a fallback default.
func (s *InventorySystem) parseTargetType(targetTypeStr string, defaultType TargetType) TargetType {
	switch targetTypeStr {
	case "self":
		return TargetSelf
	case "entity":
		return TargetEntity
	case "area":
		return TargetArea
	case "terrain":
		return TargetTerrain
	default:
		return defaultType
	}
}



// calculateScrollMagnitude determines the spell effect magnitude based on item properties.
func (s *InventorySystem) calculateScrollMagnitude(itm *item.Item) float64 {
	// Base magnitude from item value
	baseMagnitude := float64(itm.Stats.Value) / 10.0

	// Increase magnitude based on rarity
	switch itm.Rarity {
	case item.RarityUncommon:
		baseMagnitude *= 1.2
	case item.RarityRare:
		baseMagnitude *= 1.5
	case item.RarityEpic:
		baseMagnitude *= 2.0
	case item.RarityLegendary:
		baseMagnitude *= 3.0
	}

	return baseMagnitude
}
