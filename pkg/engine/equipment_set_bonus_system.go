// Package engine provides the equipment set bonus system.
// This system detects equipped set pieces and calculates cumulative bonuses.
package engine

import (
	"hash/fnv"
	"math/rand"
	"sort"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// EquipmentSetBonusSystem detects equipped set pieces and applies tiered bonuses.
type EquipmentSetBonusSystem struct {
	// setRegistry maps SetID to set definitions
	setRegistry map[string]*EquipmentSetDefinition
	// logger for system events
	logger *logrus.Logger
}

// NewEquipmentSetBonusSystem creates a new equipment set bonus system.
func NewEquipmentSetBonusSystem(seed int64, genreID string, logger *logrus.Logger) *EquipmentSetBonusSystem {
	s := &EquipmentSetBonusSystem{
		setRegistry: make(map[string]*EquipmentSetDefinition),
		logger:      logger,
	}

	// Generate procedural set definitions based on seed and genre
	s.generateSetDefinitions(seed, genreID)

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"system_name": "EquipmentSetBonusSystem",
			"set_count":   len(s.setRegistry),
			"genre":       genreID,
		}).Debug("Equipment set bonus system initialized")
	}

	return s
}

// generateSetDefinitions creates genre-appropriate set definitions.
func (s *EquipmentSetBonusSystem) generateSetDefinitions(seed int64, genreID string) {
	rng := rand.New(rand.NewSource(seed))

	// Base sets available in all genres
	baseSets := []struct {
		suffix      string
		focus       string // "damage", "defense", "speed", "balanced"
		description string
	}{
		{"Guardian", "defense", "Armor forged for those who protect others"},
		{"Berserker", "damage", "Equipment stained with the blood of countless foes"},
		{"Swift", "speed", "Lightweight gear for those who strike like wind"},
		{"Warlord", "balanced", "Battle-tested equipment of legendary commanders"},
		{"Shadow", "speed", "Gear worn by those who move unseen"},
		{"Juggernaut", "defense", "Impossibly heavy armor for unstoppable warriors"},
	}

	// Genre-specific prefixes
	genrePrefixes := map[string][]string{
		"fantasy":   {"Dragon's", "Phoenix", "Ancient", "Enchanted", "Royal", "Mystic"},
		"scifi":     {"Quantum", "Plasma", "Nanotech", "Exo", "Pulse", "Void"},
		"horror":    {"Cursed", "Eldritch", "Nightmare", "Abyssal", "Tormented", "Wraith"},
		"cyberpunk": {"Chrome", "Neon", "Corp", "Street", "Synth", "Wire"},
		"postapoc":  {"Salvage", "Wastelander's", "Rad", "Scrap", "Survivor's", "Rust"},
	}

	prefixes := genrePrefixes[genreID]
	if len(prefixes) == 0 {
		prefixes = genrePrefixes["fantasy"] // default
	}

	// Shuffle prefixes deterministically
	rng.Shuffle(len(prefixes), func(i, j int) {
		prefixes[i], prefixes[j] = prefixes[j], prefixes[i]
	})

	for i, base := range baseSets {
		prefix := prefixes[i%len(prefixes)]
		setID := prefix + "_" + base.suffix
		setName := prefix + " " + base.suffix

		def := &EquipmentSetDefinition{
			SetID:       setID,
			SetName:     setName,
			Description: base.description,
			TotalPieces: 6,
			GenreID:     genreID,
			Tiers:       s.generateTiers(rng, base.focus),
		}

		s.setRegistry[setID] = def
	}
}

// generateTiers creates bonus tiers based on set focus.
func (s *EquipmentSetBonusSystem) generateTiers(rng *rand.Rand, focus string) []SetBonusTier {
	tiers := make([]SetBonusTier, 3)

	// Base values with some randomization
	baseMultiplier := 0.8 + rng.Float64()*0.4 // 0.8-1.2x

	switch focus {
	case "damage":
		tiers[0] = SetBonusTier{
			PiecesRequired:      2,
			DamageBonus:         int(5 * baseMultiplier),
			CriticalChanceBonus: 0.03,
			SpecialEffect:       "+5% damage to low health targets",
		}
		tiers[1] = SetBonusTier{
			PiecesRequired:      4,
			DamageBonus:         int(15 * baseMultiplier),
			AttackSpeedBonus:    0.1,
			CriticalChanceBonus: 0.05,
			SpecialEffect:       "Critical hits deal 25% more damage",
		}
		tiers[2] = SetBonusTier{
			PiecesRequired:      6,
			DamageBonus:         int(30 * baseMultiplier),
			AttackSpeedBonus:    0.15,
			CriticalChanceBonus: 0.1,
			SpecialEffect:       "Kills restore 5% health",
		}

	case "defense":
		tiers[0] = SetBonusTier{
			PiecesRequired: 2,
			DefenseBonus:   int(10 * baseMultiplier),
			HealthBonus:    int(25 * baseMultiplier),
			SpecialEffect:  "Reduce damage from critical hits by 15%",
		}
		tiers[1] = SetBonusTier{
			PiecesRequired: 4,
			DefenseBonus:   int(25 * baseMultiplier),
			HealthBonus:    int(75 * baseMultiplier),
			ManaRegenBonus: 0.1,
			SpecialEffect:  "Block chance +10%",
		}
		tiers[2] = SetBonusTier{
			PiecesRequired: 6,
			DefenseBonus:   int(50 * baseMultiplier),
			HealthBonus:    int(150 * baseMultiplier),
			ManaRegenBonus: 0.2,
			SpecialEffect:  "Damage taken reduced by 15% when below 30% health",
		}

	case "speed":
		tiers[0] = SetBonusTier{
			PiecesRequired:      2,
			MovementSpeedBonus:  0.08,
			AttackSpeedBonus:    0.05,
			CriticalChanceBonus: 0.02,
			SpecialEffect:       "Dodge attacks grant brief invulnerability",
		}
		tiers[1] = SetBonusTier{
			PiecesRequired:     4,
			MovementSpeedBonus: 0.15,
			AttackSpeedBonus:   0.12,
			DamageBonus:        int(8 * baseMultiplier),
			SpecialEffect:      "Moving increases crit chance (max +10%)",
		}
		tiers[2] = SetBonusTier{
			PiecesRequired:     6,
			MovementSpeedBonus: 0.25,
			AttackSpeedBonus:   0.2,
			DamageBonus:        int(15 * baseMultiplier),
			SpecialEffect:      "First hit after moving deals 50% more damage",
		}

	default: // balanced
		tiers[0] = SetBonusTier{
			PiecesRequired:      2,
			DamageBonus:         int(3 * baseMultiplier),
			DefenseBonus:        int(5 * baseMultiplier),
			HealthBonus:         int(15 * baseMultiplier),
			CriticalChanceBonus: 0.02,
			SpecialEffect:       "Well-rounded combat bonuses",
		}
		tiers[1] = SetBonusTier{
			PiecesRequired:     4,
			DamageBonus:        int(8 * baseMultiplier),
			DefenseBonus:       int(12 * baseMultiplier),
			HealthBonus:        int(40 * baseMultiplier),
			AttackSpeedBonus:   0.08,
			MovementSpeedBonus: 0.05,
			SpecialEffect:      "Gain experience 5% faster",
		}
		tiers[2] = SetBonusTier{
			PiecesRequired:      6,
			DamageBonus:         int(18 * baseMultiplier),
			DefenseBonus:        int(25 * baseMultiplier),
			HealthBonus:         int(80 * baseMultiplier),
			AttackSpeedBonus:    0.15,
			MovementSpeedBonus:  0.1,
			CriticalChanceBonus: 0.08,
			ManaRegenBonus:      0.15,
			SpecialEffect:       "All skills cooldown 10% faster",
		}
	}

	return tiers
}

// Update processes all entities and updates their set bonuses.
func (s *EquipmentSetBonusSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		s.updateEntity(entity)
	}
}

// updateEntity updates set bonuses for a single entity.
func (s *EquipmentSetBonusSystem) updateEntity(entity *Entity) {
	// Get required components
	equipComp, hasEquip := entity.GetComponent("equipment")
	if !hasEquip || equipComp == nil {
		return
	}
	equipment, ok := equipComp.(*EquipmentComponent)
	if !ok {
		return
	}

	// Get or create set bonus component
	setBonusComp, hasSetBonus := entity.GetComponent("equipment_set_bonus")
	var setBonus *EquipmentSetBonusComponent
	if !hasSetBonus || setBonusComp == nil {
		setBonus = NewEquipmentSetBonusComponent()
		entity.AddComponent(setBonus)
	} else {
		setBonus = setBonusComp.(*EquipmentSetBonusComponent)
	}

	// Calculate equipment hash to detect changes
	currentHash := s.calculateEquipmentHash(equipment)
	if currentHash == setBonus.LastEquipmentHash && !setBonus.Dirty {
		return // No changes
	}

	// Count pieces per set
	setPieceCounts := make(map[string]int)
	for _, itm := range equipment.Slots {
		if itm != nil && itm.SetID != "" {
			setPieceCounts[itm.SetID]++
		}
	}

	// Update active sets
	setBonus.ActiveSets = make(map[string]*ActiveSetBonus)

	for setID, pieceCount := range setPieceCounts {
		def, exists := s.setRegistry[setID]
		if !exists {
			continue
		}

		activeSet := &ActiveSetBonus{
			SetID:          setID,
			SetName:        def.SetName,
			PiecesEquipped: pieceCount,
			TotalPieces:    def.TotalPieces,
			ActiveTiers:    make([]int, 0),
			CombinedBonus:  SetBonusTier{},
		}

		// Activate applicable tiers
		for i, tier := range def.Tiers {
			if pieceCount >= tier.PiecesRequired {
				activeSet.ActiveTiers = append(activeSet.ActiveTiers, i)
				// Combine bonuses
				activeSet.CombinedBonus.DamageBonus += tier.DamageBonus
				activeSet.CombinedBonus.DefenseBonus += tier.DefenseBonus
				activeSet.CombinedBonus.AttackSpeedBonus += tier.AttackSpeedBonus
				activeSet.CombinedBonus.CriticalChanceBonus += tier.CriticalChanceBonus
				activeSet.CombinedBonus.MovementSpeedBonus += tier.MovementSpeedBonus
				activeSet.CombinedBonus.HealthBonus += tier.HealthBonus
				activeSet.CombinedBonus.ManaRegenBonus += tier.ManaRegenBonus
			}
		}

		if len(activeSet.ActiveTiers) > 0 {
			setBonus.ActiveSets[setID] = activeSet
		}
	}

	setBonus.LastEquipmentHash = currentHash
	setBonus.Dirty = false

	if s.logger != nil && len(setBonus.ActiveSets) > 0 {
		s.logger.WithFields(logrus.Fields{
			"system_name":  "EquipmentSetBonusSystem",
			"entity_id":    entity.ID,
			"active_sets":  len(setBonus.ActiveSets),
			"total_damage": setBonus.GetTotalDamageBonus(),
		}).Debug("Updated entity set bonuses")
	}
}

// calculateEquipmentHash generates a hash of current equipment for change detection.
func (s *EquipmentSetBonusSystem) calculateEquipmentHash(equipment *EquipmentComponent) uint64 {
	h := fnv.New64a()

	// Sort slots for deterministic hashing
	slots := make([]int, 0, len(equipment.Slots))
	for slot := range equipment.Slots {
		slots = append(slots, int(slot))
	}
	sort.Ints(slots)

	for _, slot := range slots {
		itm := equipment.Slots[EquipmentSlot(slot)]
		if itm != nil {
			h.Write([]byte(itm.ID))
			h.Write([]byte(itm.SetID))
		}
	}

	return h.Sum64()
}

// GetSetDefinition returns the definition for a set ID.
func (s *EquipmentSetBonusSystem) GetSetDefinition(setID string) *EquipmentSetDefinition {
	return s.setRegistry[setID]
}

// GetAllSetDefinitions returns all registered set definitions.
func (s *EquipmentSetBonusSystem) GetAllSetDefinitions() []*EquipmentSetDefinition {
	defs := make([]*EquipmentSetDefinition, 0, len(s.setRegistry))
	for _, def := range s.setRegistry {
		defs = append(defs, def)
	}
	return defs
}

// RegisterSet adds a custom set definition to the registry.
func (s *EquipmentSetBonusSystem) RegisterSet(def *EquipmentSetDefinition) {
	if def != nil && def.SetID != "" {
		s.setRegistry[def.SetID] = def
	}
}

// GetRandomSetID returns a random set ID for item generation.
func (s *EquipmentSetBonusSystem) GetRandomSetID(rng *rand.Rand) string {
	if len(s.setRegistry) == 0 {
		return ""
	}

	// Collect all set IDs
	setIDs := make([]string, 0, len(s.setRegistry))
	for id := range s.setRegistry {
		setIDs = append(setIDs, id)
	}

	return setIDs[rng.Intn(len(setIDs))]
}

// AssignSetToItem assigns a set ID to an item based on item properties.
// Returns true if a set was assigned.
func (s *EquipmentSetBonusSystem) AssignSetToItem(itm *item.Item, rng *rand.Rand) bool {
	if itm == nil || !itm.IsEquippable() {
		return false
	}

	// Only rare+ items can be part of sets
	if itm.Rarity < item.RarityRare {
		return false
	}

	// Higher rarity = higher chance of being a set piece
	setChance := 0.0
	switch itm.Rarity {
	case item.RarityRare:
		setChance = 0.15
	case item.RarityEpic:
		setChance = 0.35
	case item.RarityLegendary:
		setChance = 0.60
	}

	if rng.Float64() > setChance {
		return false
	}

	// Assign to a random set
	itm.SetID = s.GetRandomSetID(rng)
	if itm.SetID == "" {
		return false
	}

	// Assign piece index based on armor type
	switch itm.Type {
	case item.TypeWeapon:
		itm.SetPieceIndex = 0
	case item.TypeArmor:
		switch itm.ArmorType {
		case item.ArmorHelmet:
			itm.SetPieceIndex = 1
		case item.ArmorChest:
			itm.SetPieceIndex = 2
		case item.ArmorLegs:
			itm.SetPieceIndex = 3
		case item.ArmorBoots:
			itm.SetPieceIndex = 4
		case item.ArmorGloves:
			itm.SetPieceIndex = 5
		}
	case item.TypeAccessory:
		itm.SetPieceIndex = rng.Intn(3) + 6 // 6, 7, or 8 for accessories
	}

	return true
}
