//go:build !android && !ios
// +build !android,!ios

package main

import (
	"encoding/json"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/qol"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/saveload"
)

// serializePlayerState extracts all player state for saving.
func serializePlayerState(player *engine.Entity, game *engine.EbitenGame) *saveload.PlayerState {
	playerState := &saveload.PlayerState{EntityID: player.ID}

	serializePosition(player, playerState)
	serializeHealth(player, playerState)
	serializeStats(player, playerState)
	serializeExperience(player, playerState)
	serializeInventory(player, playerState)
	serializeEquipment(player, playerState)
	serializeManaAndSpells(player, playerState)
	serializeTutorialState(game, playerState)
	serializeOnboardingState(game, playerState)
	serializeContextTutorialState(game, playerState)
	serializeQoLState(player, playerState)

	playerState.Speed = 1.0
	return playerState
}

// serializePosition extracts player position to state.
func serializePosition(player *engine.Entity, state *saveload.PlayerState) {
	if posComp, ok := player.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		state.X, state.Y = pos.X, pos.Y
	}
}

// serializeHealth extracts player health to state.
func serializeHealth(player *engine.Entity, state *saveload.PlayerState) {
	if healthComp, ok := player.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		state.CurrentHealth, state.MaxHealth = health.Current, health.Max
	}
}

// serializeStats extracts player stats to state.
func serializeStats(player *engine.Entity, state *saveload.PlayerState) {
	if statsComp, ok := player.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		state.Attack = stats.Attack
		state.Defense = stats.Defense
		state.MagicPower = stats.MagicPower
	}
}

// serializeExperience extracts player level and XP to state.
func serializeExperience(player *engine.Entity, state *saveload.PlayerState) {
	if expComp, ok := player.GetComponent("experience"); ok {
		exp, ok := expComp.(*engine.ExperienceComponent)
		if !ok {
			return
		}
		state.Level = exp.Level
		state.Experience = exp.CurrentXP
	}
}

// serializeInventory extracts inventory data to state.
func serializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv, ok := invComp.(*engine.InventoryComponent)
		if !ok {
			return
		}
		state.Gold = inv.Gold
		state.Items = make([]saveload.ItemData, 0, len(inv.Items))
		for _, itm := range inv.Items {
			state.Items = append(state.Items, saveload.ItemToData(itm))
		}
	}
}

// serializeEquipment extracts equipped items to state.
func serializeEquipment(player *engine.Entity, state *saveload.PlayerState) {
	if equip, hasEquip := player.GetComponent("equipment"); hasEquip {
		equipment, ok := equip.(*engine.EquipmentComponent)
		if !ok {
			return
		}
		if weapon := equipment.Slots[engine.SlotMainHand]; weapon != nil {
			weaponData := saveload.ItemToData(weapon)
			state.EquippedItems.Weapon = &weaponData
		}
		if armor := equipment.Slots[engine.SlotChest]; armor != nil {
			armorData := saveload.ItemToData(armor)
			state.EquippedItems.Armor = &armorData
		}
		if accessory := equipment.Slots[engine.SlotAccessory1]; accessory != nil {
			accessoryData := saveload.ItemToData(accessory)
			state.EquippedItems.Accessory = &accessoryData
		}
	}
}

// serializeManaAndSpells extracts mana and spell data to state.
func serializeManaAndSpells(player *engine.Entity, state *saveload.PlayerState) {
	if manaComp, hasMana := player.GetComponent("mana"); hasMana {
		mana, ok := manaComp.(*engine.ManaComponent)
		if !ok {
			return
		}
		state.CurrentMana = mana.Current
		state.MaxMana = mana.Max
	}

	if slotsComp, hasSlots := player.GetComponent("spell_slots"); hasSlots {
		slots, ok := slotsComp.(*engine.SpellSlotComponent)
		if !ok {
			return
		}
		state.Spells = make([]saveload.SpellData, 0, 5)
		for i := 0; i < 5; i++ {
			if spell := slots.GetSlot(i); spell != nil {
				state.Spells = append(state.Spells, saveload.SpellToData(spell))
			}
		}
	}
}

// serializeTutorialState extracts tutorial state to state.
func serializeTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.TutorialSystem != nil {
		enabled, showUI, currentStep, completed := game.TutorialSystem.ExportState()
		state.TutorialState = &saveload.TutorialStateData{
			Enabled:        enabled,
			ShowUI:         showUI,
			CurrentStepIdx: currentStep,
			CompletedSteps: completed,
		}
	}
}

// serializeOnboardingState extracts onboarding manager state to state.
// Phase 4.5: Enables persistence of onboarding flow progress.
func serializeOnboardingState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.OnboardingManager != nil {
		data := game.OnboardingManager.ExportState()
		state.OnboardingState = &saveload.OnboardingStateData{
			CurrentState: data.CurrentState,
			Enabled:      data.Enabled,
			Skipped:      data.Skipped,
			PlayerClass:  data.PlayerClass,
		}
	}
}

// serializeContextTutorialState extracts context-sensitive tutorial state to state.
// Phase 4.5: Enables persistence of viewed tutorial topics.
func serializeContextTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.ContextualTutorial != nil {
		data := game.ContextualTutorial.ExportState()
		state.ContextTutorialState = &saveload.ContextTutorialStateData{
			Enabled:      data.Enabled,
			ViewedTopics: data.ViewedTopics,
		}
	}
}

// deserializePlayerState restores all player state from a save.
func deserializePlayerState(player *engine.Entity, playerState *saveload.PlayerState, game *engine.EbitenGame) {
	deserializePosition(player, playerState)
	deserializeHealth(player, playerState)
	deserializeStats(player, playerState)
	deserializeExperience(player, playerState)
	deserializeInventory(player, playerState)
	deserializeEquipment(player, playerState)
	deserializeManaAndSpells(player, playerState)
	deserializeTutorialState(game, playerState)
	deserializeOnboardingState(game, playerState)
	deserializeContextTutorialState(game, playerState)
	deserializeQoLState(player, playerState)
}

// deserializePosition restores player position from state.
func deserializePosition(player *engine.Entity, state *saveload.PlayerState) {
	if posComp, ok := player.GetComponent("position"); ok {
		pos, ok := posComp.(*engine.PositionComponent)
		if !ok {
			return
		}
		pos.X, pos.Y = state.X, state.Y
	}
}

// deserializeHealth restores player health from state.
func deserializeHealth(player *engine.Entity, state *saveload.PlayerState) {
	if healthComp, ok := player.GetComponent("health"); ok {
		health, ok := healthComp.(*engine.HealthComponent)
		if !ok {
			return
		}
		health.Current, health.Max = state.CurrentHealth, state.MaxHealth
	}
}

// deserializeStats restores player stats from state.
func deserializeStats(player *engine.Entity, state *saveload.PlayerState) {
	if statsComp, ok := player.GetComponent("stats"); ok {
		stats, ok := statsComp.(*engine.StatsComponent)
		if !ok {
			return
		}
		stats.Attack = state.Attack
		stats.Defense = state.Defense
		stats.MagicPower = state.MagicPower
	}
}

// deserializeExperience restores player level and XP from state.
func deserializeExperience(player *engine.Entity, state *saveload.PlayerState) {
	if expComp, ok := player.GetComponent("experience"); ok {
		exp, ok := expComp.(*engine.ExperienceComponent)
		if !ok {
			return
		}
		exp.Level = state.Level
		exp.CurrentXP = state.Experience
	}
}

// deserializeInventory restores inventory from state.
func deserializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv, ok := invComp.(*engine.InventoryComponent)
		if !ok {
			return
		}
		inv.Items = make([]*item.Item, 0, len(state.Items))
		for _, itemData := range state.Items {
			restoredItem := saveload.DataToItem(itemData)
			inv.Items = append(inv.Items, restoredItem)
		}
		inv.Gold = state.Gold
	}
}

// deserializeEquipment restores equipped items from state.
func deserializeEquipment(player *engine.Entity, state *saveload.PlayerState) {
	if equipComp, ok := player.GetComponent("equipment"); ok {
		equipment, ok := equipComp.(*engine.EquipmentComponent)
		if !ok {
			return
		}
		equipment.Slots = make(map[engine.EquipmentSlot]*item.Item)

		if state.EquippedItems.Weapon != nil {
			weapon := saveload.DataToItem(*state.EquippedItems.Weapon)
			equipment.Slots[engine.SlotMainHand] = weapon
		}
		if state.EquippedItems.Armor != nil {
			armor := saveload.DataToItem(*state.EquippedItems.Armor)
			equipment.Slots[engine.SlotChest] = armor
		}
		if state.EquippedItems.Accessory != nil {
			accessory := saveload.DataToItem(*state.EquippedItems.Accessory)
			equipment.Slots[engine.SlotAccessory1] = accessory
		}
		equipment.StatsDirty = true
	}
}

// deserializeManaAndSpells restores mana and spells from state.
func deserializeManaAndSpells(player *engine.Entity, state *saveload.PlayerState) {
	if manaComp, ok := player.GetComponent("mana"); ok {
		mana, ok := manaComp.(*engine.ManaComponent)
		if !ok {
			return
		}
		mana.Current, mana.Max = state.CurrentMana, state.MaxMana
	}

	if slotsComp, ok := player.GetComponent("spell_slots"); ok {
		slots, ok := slotsComp.(*engine.SpellSlotComponent)
		if !ok {
			return
		}
		for i := 0; i < 5; i++ {
			slots.Slots[i] = nil
		}
		for i, spellData := range state.Spells {
			if i < 5 {
				restoredSpell := saveload.DataToSpell(spellData)
				slots.SetSlot(i, restoredSpell)
			}
		}
	}
}

// deserializeTutorialState restores tutorial state from state.
func deserializeTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.TutorialSystem != nil && state.TutorialState != nil {
		tutState := state.TutorialState
		game.TutorialSystem.ImportState(
			tutState.Enabled,
			tutState.ShowUI,
			tutState.CurrentStepIdx,
			tutState.CompletedSteps,
		)
	}
}

// deserializeOnboardingState restores onboarding manager state from state.
// Phase 4.5: Restores onboarding flow progress after load.
func deserializeOnboardingState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.OnboardingManager != nil && state.OnboardingState != nil {
		data := engine.OnboardingStateData{
			CurrentState: state.OnboardingState.CurrentState,
			Enabled:      state.OnboardingState.Enabled,
			Skipped:      state.OnboardingState.Skipped,
			PlayerClass:  state.OnboardingState.PlayerClass,
		}
		game.OnboardingManager.ImportState(data)
	}
}

// deserializeContextTutorialState restores context-sensitive tutorial state from state.
// Phase 4.5: Restores viewed tutorial topics after load.
func deserializeContextTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.ContextualTutorial != nil && state.ContextTutorialState != nil {
		data := saveload.ContextTutorialStateData{
			Enabled:      state.ContextTutorialState.Enabled,
			ViewedTopics: state.ContextTutorialState.ViewedTopics,
		}
		game.ContextualTutorial.ImportState(data)
	}
}

// serializeQoLState extracts QoL preferences to state.
// AUDIT.md fix: Allows QoL settings to persist across saves/loads.
func serializeQoLState(player *engine.Entity, state *saveload.PlayerState) {
	qolComp, ok := player.GetComponent("qol")
	if !ok {
		return
	}
	q, ok := qolComp.(*qol.QoLComponent)
	if !ok {
		return
	}

	// Serialize craft queue to JSON
	var craftQueueJSON []byte
	if len(q.CraftQueue) > 0 {
		craftQueueJSON, _ = json.Marshal(q.CraftQueue)
	}

	state.QoLData = &saveload.QoLStateData{
		PlayerID:        q.PlayerID,
		AutoLootEnabled: q.AutoLootEnabled,
		AutoLootRadius:  q.AutoLootRadius,
		CraftQueueJSON:  craftQueueJSON,
		SortPreset:      q.SortPreset,
		MountWhistle:    q.MountWhistle,
		RecipeTracking:  q.RecipeTracking,
	}
}

// deserializeQoLState restores QoL preferences from state.
// AUDIT.md fix: Restores QoL settings after load.
func deserializeQoLState(player *engine.Entity, state *saveload.PlayerState) {
	if state.QoLData == nil {
		return
	}

	// Get or create QoL component
	qolComp, ok := player.GetComponent("qol")
	var q *qol.QoLComponent
	if ok {
		q, ok = qolComp.(*qol.QoLComponent)
		if !ok {
			return
		}
	} else {
		// Create new QoL component if it doesn't exist
		q = &qol.QoLComponent{
			PlayerID: player.ID,
		}
		player.AddComponent(q)
	}

	// Restore QoL settings
	q.PlayerID = state.QoLData.PlayerID
	q.AutoLootEnabled = state.QoLData.AutoLootEnabled
	q.AutoLootRadius = state.QoLData.AutoLootRadius
	q.SortPreset = state.QoLData.SortPreset
	q.MountWhistle = state.QoLData.MountWhistle
	q.RecipeTracking = state.QoLData.RecipeTracking

	// Deserialize craft queue from JSON
	if len(state.QoLData.CraftQueueJSON) > 0 {
		var craftQueue []*qol.CraftQueueEntry
		if err := json.Unmarshal(state.QoLData.CraftQueueJSON, &craftQueue); err == nil {
			q.CraftQueue = craftQueue
		}
	}
}

// createGameSave creates a complete game save from current game state.
func createGameSave(player *engine.Entity, game *engine.EbitenGame, generatedTerrain *terrain.Terrain) *saveload.GameSave {
	playerState := serializePlayerState(player, game)

	// Get fog of war data
	var fogOfWar [][]bool
	if game.MapUI != nil {
		fogOfWar = game.MapUI.GetFogOfWar()
	}

	return &saveload.GameSave{
		Version:     saveload.SaveVersion,
		PlayerState: playerState,
		WorldState: &saveload.WorldState{
			Seed:       *seed,
			GenreID:    *genreID,
			Width:      generatedTerrain.Width,
			Height:     generatedTerrain.Height,
			Difficulty: 0.5,
			Depth:      1,
			FogOfWar:   fogOfWar,
		},
		Settings: &saveload.GameSettings{
			ScreenWidth:  *width,
			ScreenHeight: *height,
			Fullscreen:   false,
			VSync:        true,
			MasterVolume: 1.0,
			MusicVolume:  0.7,
			SFXVolume:    0.8,
			KeyBindings:  make(map[string]string),
		},
	}
}

// loadGameSave restores game state from a saved game.
func loadGameSave(player *engine.Entity, gameSave *saveload.GameSave, game *engine.EbitenGame) {
	deserializePlayerState(player, gameSave.PlayerState, game)

	// Restore fog of war
	if game.MapUI != nil && gameSave.WorldState != nil && gameSave.WorldState.FogOfWar != nil {
		game.MapUI.SetFogOfWar(gameSave.WorldState.FogOfWar)
	}
}
