//go:build !headless
// +build !headless

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
	player := s.findLivePlayerEntity(entities)
	if player == nil {
		return
	}

	slots := s.getSpellSlotsComponent(player)
	if slots == nil || slots.IsCasting() {
		return
	}

	input := s.getPlayerInput(player)
	if input == nil {
		return
	}

	slotIndex := s.detectSpellSlotInput(input)
	if slotIndex >= 0 {
		s.castingSystem.StartCast(player, slotIndex)
	}
}

// findLivePlayerEntity finds the player entity that is alive and has input.
func (s *PlayerSpellCastingSystem) findLivePlayerEntity(entities []*Entity) *Entity {
	for _, entity := range entities {
		if entity.HasComponent("input") && !entity.HasComponent("dead") {
			return entity
		}
	}
	return nil
}

// getSpellSlotsComponent retrieves and validates the spell slots component.
func (s *PlayerSpellCastingSystem) getSpellSlotsComponent(player *Entity) *SpellSlotComponent {
	if !player.HasComponent("spell_slots") {
		return nil
	}

	slotsComp, _ := player.GetComponent("spell_slots")
	slots, ok := slotsComp.(*SpellSlotComponent)
	if !ok {
		return nil
	}
	return slots
}

// getPlayerInput retrieves the InputProvider from the player entity.
func (s *PlayerSpellCastingSystem) getPlayerInput(player *Entity) InputProvider {
	inputComp, hasInput := player.GetComponent("input")
	if !hasInput {
		return nil
	}
	input, ok := inputComp.(InputProvider)
	if !ok {
		return nil
	}
	return input
}

// detectSpellSlotInput checks which spell slot key (1-5) is pressed.
func (s *PlayerSpellCastingSystem) detectSpellSlotInput(input InputProvider) int {
	if input.IsSpellPressed(1) {
		return 0
	}
	if input.IsSpellPressed(2) {
		return 1
	}
	if input.IsSpellPressed(3) {
		return 2
	}
	if input.IsSpellPressed(4) {
		return 3
	}
	if input.IsSpellPressed(5) {
		return 4
	}
	return -1
}
