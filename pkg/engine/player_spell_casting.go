package engine

import "github.com/hajimehoshi/ebiten/v2"

// PlayerSpellCastingSystem handles player spell casting from input.
type PlayerSpellCastingSystem struct {
	castingSystem *SpellCastingSystem
	world         *World

	// Key bindings for spell slots
	KeySpell1 ebiten.Key
	KeySpell2 ebiten.Key
	KeySpell3 ebiten.Key
	KeySpell4 ebiten.Key
	KeySpell5 ebiten.Key
}

// NewPlayerSpellCastingSystem creates a player spell casting system.
func NewPlayerSpellCastingSystem(castingSystem *SpellCastingSystem, world *World) *PlayerSpellCastingSystem {
	return &PlayerSpellCastingSystem{
		castingSystem: castingSystem,
		world:         world,
		KeySpell1:     ebiten.Key1,
		KeySpell2:     ebiten.Key2,
		KeySpell3:     ebiten.Key3,
		KeySpell4:     ebiten.Key4,
		KeySpell5:     ebiten.Key5,
	}
}

// Update processes spell casting input for the player.
func (s *PlayerSpellCastingSystem) Update(entities []*Entity, deltaTime float64) {
	// Find player entity
	var player *Entity
	for _, entity := range entities {
		if entity.HasComponent("input") {
			// Skip dead entities - they cannot cast spells (Category 1.1)
			if entity.HasComponent("dead") {
				continue
			}
			player = entity
			break
		}
	}

	if player == nil {
		return
	}

	// Check if player has spell slots
	if !player.HasComponent("spell_slots") {
		return
	}

	// Get spell slots
	slotsComp, _ := player.GetComponent("spell_slots")
	slots := slotsComp.(*SpellSlotComponent)

	// If currently casting, don't start new cast
	if slots.IsCasting() {
		return
	}

	// GAP-002 REPAIR: Read spell input flags from InputProvider
	inputComp, hasInput := player.GetComponent("input")
	if !hasInput {
		return
	}
	input, ok := inputComp.(InputProvider)
	if !ok {
		return // Not an InputProvider
	}

	// Check spell slot input flags (keys 1-5)
	slotIndex := -1
	if input.IsSpellPressed(1) {
		slotIndex = 0
	} else if input.IsSpellPressed(2) {
		slotIndex = 1
	} else if input.IsSpellPressed(3) {
		slotIndex = 2
	} else if input.IsSpellPressed(4) {
		slotIndex = 3
	} else if input.IsSpellPressed(5) {
		slotIndex = 4
	}

	// Attempt to cast spell
	if slotIndex >= 0 {
		s.castingSystem.StartCast(player, slotIndex)
	}
}
