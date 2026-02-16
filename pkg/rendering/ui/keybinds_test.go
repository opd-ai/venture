package ui

import (
	"os"
	"testing"
)

func TestKeybindManager_Defaults(t *testing.T) {
	km := NewKeybindManager()

	// Test movement bindings
	kb, err := km.GetBinding(ActionMoveUp)
	if err != nil {
		t.Fatalf("failed to get move up binding: %v", err)
	}
	if kb.PrimaryKey != KeyW {
		t.Errorf("expected W for move up, got %s", kb.PrimaryKey)
	}
	if kb.SecondaryKey != KeyUp {
		t.Errorf("expected Up for move up secondary, got %s", kb.SecondaryKey)
	}

	// Test combat bindings
	kb, err = km.GetBinding(ActionAttack)
	if err != nil {
		t.Fatalf("failed to get attack binding: %v", err)
	}
	if kb.PrimaryKey != MouseLeft {
		t.Errorf("expected MouseLeft for attack, got %s", kb.PrimaryKey)
	}
}

func TestKeybindManager_SetBinding(t *testing.T) {
	km := NewKeybindManager()

	// Change attack to a unique key not used by defaults
	if err := km.SetBinding(ActionAttack, KeyX, ""); err != nil {
		t.Errorf("failed to set binding: %v", err)
	}

	kb, _ := km.GetBinding(ActionAttack)
	if kb.PrimaryKey != KeyX {
		t.Errorf("expected X for attack, got %s", kb.PrimaryKey)
	}

	// Verify reverse lookup
	if action := km.GetActionForKey(KeyX); action != ActionAttack {
		t.Errorf("expected attack for X, got %s", action)
	}
}

func TestKeybindManager_ConflictDetection(t *testing.T) {
	km := NewKeybindManager()

	// Try to bind Space to two different actions
	km.SetBinding(ActionDodge, KeySpace, "")

	// This should fail due to conflict (Space already bound to Dodge)
	if err := km.SetBinding(ActionAttack, KeySpace, ""); err == nil {
		t.Error("expected error for conflicting keybind")
	}
}

func TestKeybindManager_SecondaryBinding(t *testing.T) {
	km := NewKeybindManager()

	// Set primary and secondary using unique keys
	if err := km.SetBinding(ActionInventory, KeyI, KeyQ); err != nil {
		t.Errorf("failed to set binding with secondary: %v", err)
	}

	// Both should map to inventory
	if action := km.GetActionForKey(KeyI); action != ActionInventory {
		t.Errorf("primary key I should map to inventory")
	}
	if action := km.GetActionForKey(KeyQ); action != ActionInventory {
		t.Errorf("secondary key Q should map to inventory")
	}
}

func TestKeybindManager_DetectConflicts(t *testing.T) {
	km := NewKeybindManager()

	// Default bindings have some intentional duplicates (mount/dismount both use V)
	// Test should allow these known duplicates but detect real conflicts
	conflicts := km.DetectConflicts()
	// We expect mount/dismount to both use V, which is intentional
	if len(conflicts) > 0 {
		// Just log them but don't fail - some duplicates are intentional
		t.Logf("Detected %d potential conflicts (some may be intentional): %v", len(conflicts), conflicts)
	}
}

func TestKeybindManager_ListAllBindings(t *testing.T) {
	km := NewKeybindManager()

	bindings := km.ListAllBindings()
	if len(bindings) < 40 {
		t.Errorf("expected at least 40 default bindings, got %d", len(bindings))
	}
}

func TestKeybindManager_ResetToDefaults(t *testing.T) {
	km := NewKeybindManager()

	// Modify bindings
	km.SetBinding(ActionAttack, KeySpace, "")
	km.SetBinding(ActionInventory, KeyJ, "")

	// Reset
	km.ResetToDefaults()

	// Verify defaults restored
	kb, _ := km.GetBinding(ActionAttack)
	if kb.PrimaryKey != MouseLeft {
		t.Error("expected attack reset to MouseLeft")
	}
}

func TestKeybindManager_SaveLoad(t *testing.T) {
	km := NewKeybindManager()
	filename := "test_keybinds.json"
	defer os.Remove(filename)

	// Modify bindings using unique keys that don't conflict with defaults
	if err := km.SetBinding(ActionAttack, KeyX, Mouse4); err != nil {
		t.Fatalf("failed to set attack binding: %v", err)
	}
	if err := km.SetBinding(ActionInventory, KeyI, KeyQ); err != nil {
		t.Fatalf("failed to set inventory binding: %v", err)
	}

	// Save
	if err := km.Save(filename); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load into new manager
	km2 := NewKeybindManager()
	if err := km2.Load(filename); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify loaded bindings
	kb, _ := km2.GetBinding(ActionAttack)
	if kb.PrimaryKey != KeyX {
		t.Errorf("expected attack X after load, got %s", kb.PrimaryKey)
	}
	if kb.SecondaryKey != Mouse4 {
		t.Errorf("expected attack secondary Mouse4 after load, got %s", kb.SecondaryKey)
	}
}

func TestQuickTravelManager_RegisterDestination(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "home1",
		Name:     "My House",
		X:        100,
		Y:        200,
		Cost:     0,
		Unlocked: false,
		Category: "House",
	}

	qtm.RegisterDestination(dest)

	// Should not be in unlocked list yet
	unlocked := qtm.ListUnlocked()
	if len(unlocked) > 0 {
		t.Error("expected no unlocked destinations")
	}
}

func TestQuickTravelManager_UnlockDestination(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "home1",
		Name:     "My House",
		X:        100,
		Y:        200,
		Unlocked: false,
		Category: "House",
	}
	qtm.RegisterDestination(dest)

	// Unlock
	if err := qtm.UnlockDestination("home1"); err != nil {
		t.Errorf("failed to unlock: %v", err)
	}

	// Should be in unlocked list
	unlocked := qtm.ListUnlocked()
	if len(unlocked) != 1 {
		t.Errorf("expected 1 unlocked destination, got %d", len(unlocked))
	}
}

func TestQuickTravelManager_CalculateCost(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "far_city",
		Name:     "Far City",
		X:        1000,
		Y:        1000,
		Unlocked: true,
		Category: "City",
	}
	qtm.RegisterDestination(dest)

	// Calculate cost from origin
	cost, err := qtm.CalculateCost(0, 0, "far_city")
	if err != nil {
		t.Errorf("failed to calculate cost: %v", err)
	}

	// Should be between 100 and 1000 gold
	if cost < 100 || cost > 1000 {
		t.Errorf("expected cost 100-1000, got %d", cost)
	}

	// Closer destination should be cheaper
	closeDest := &TravelDestination{
		ID:       "nearby",
		Name:     "Nearby",
		X:        50,
		Y:        50,
		Unlocked: true,
	}
	qtm.RegisterDestination(closeDest)

	closeCost, _ := qtm.CalculateCost(0, 0, "nearby")
	if closeCost >= cost {
		t.Error("expected nearby destination to be cheaper")
	}
}

func TestQuickTravelManager_Travel(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "city",
		Name:     "City",
		X:        500,
		Y:        500,
		Unlocked: true,
		Category: "City",
	}
	qtm.RegisterDestination(dest)

	playerGold := 1000

	// Attempt travel
	resultDest, cost, err := qtm.Travel(0, 0, "city", &playerGold)
	if err != nil {
		t.Errorf("travel failed: %v", err)
	}

	if resultDest.ID != "city" {
		t.Errorf("expected city destination, got %s", resultDest.ID)
	}

	if playerGold >= 1000 {
		t.Error("expected gold to be deducted")
	}

	expectedGold := 1000 - cost
	if playerGold != expectedGold {
		t.Errorf("expected gold %d, got %d", expectedGold, playerGold)
	}
}

func TestQuickTravelManager_InsufficientGold(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "city",
		Name:     "City",
		X:        500,
		Y:        500,
		Unlocked: true,
	}
	qtm.RegisterDestination(dest)

	playerGold := 50 // Not enough

	_, _, err := qtm.Travel(0, 0, "city", &playerGold)
	if err == nil {
		t.Error("expected error for insufficient gold")
	}
}

func TestQuickTravelManager_ListByCategory(t *testing.T) {
	qtm := NewQuickTravelManager()

	qtm.RegisterDestination(&TravelDestination{ID: "h1", Category: "House", Unlocked: true})
	qtm.RegisterDestination(&TravelDestination{ID: "h2", Category: "House", Unlocked: true})
	qtm.RegisterDestination(&TravelDestination{ID: "c1", Category: "City", Unlocked: true})

	houses := qtm.ListByCategory("House")
	if len(houses) != 2 {
		t.Errorf("expected 2 houses, got %d", len(houses))
	}

	cities := qtm.ListByCategory("City")
	if len(cities) != 1 {
		t.Errorf("expected 1 city, got %d", len(cities))
	}
}

func TestTooltipBuilder(t *testing.T) {
	tb := NewTooltipBuilder("Legendary Sword")
	tooltip := tb.
		SetRarity("Legendary").
		AddDescription("A powerful weapon").
		AddStat("Damage", 150).
		AddStat("Speed", 2.5).
		AddBonus("+50% critical damage").
		AddRequirement("Level 50").
		SetCost(10000).
		Build()

	if tooltip.Title != "Legendary Sword" {
		t.Error("incorrect title")
	}
	if tooltip.Rarity != "Legendary" {
		t.Error("incorrect rarity")
	}
	if len(tooltip.Description) != 1 {
		t.Error("missing description")
	}
	if len(tooltip.Stats) != 2 {
		t.Error("expected 2 stats")
	}
	if len(tooltip.Bonuses) != 1 {
		t.Error("expected 1 bonus")
	}
	if tooltip.Cost != 10000 {
		t.Error("incorrect cost")
	}
}

func TestCreateItemTooltip(t *testing.T) {
	tooltip := CreateItemTooltip("Steel Sword", "Rare", 100, 50, 1.5)

	if tooltip.Rarity != "Rare" {
		t.Error("incorrect rarity")
	}

	if tooltip.Stats["Damage"] != 100 {
		t.Error("incorrect damage stat")
	}
	if tooltip.Stats["Defense"] != 50 {
		t.Error("incorrect defense stat")
	}

	// Should have crafting bonus
	if len(tooltip.Bonuses) == 0 {
		t.Error("expected crafting bonus")
	}
}

func TestCreateStationTooltip(t *testing.T) {
	tooltip := CreateStationTooltip("Forge", 4, 40)

	if tooltip.Stats["Quality"] != "Master" {
		t.Error("expected Master quality")
	}
	if tooltip.Stats["Unlocked Recipes"] != 40 {
		t.Error("expected 40 recipes")
	}
	if len(tooltip.Bonuses) == 0 {
		t.Error("expected crafting bonus")
	}
}

func TestKeybindManager_NoDefaultConflicts(t *testing.T) {
	km := NewKeybindManager()

	// After fixing VehicleMount/VehicleDismount, there should be no conflicts
	conflicts := km.DetectConflicts()
	if len(conflicts) > 0 {
		t.Errorf("default bindings should have no conflicts, got: %v", conflicts)
	}
}

func TestKeybindManager_SaveLoadPreservesDescription(t *testing.T) {
	km := NewKeybindManager()
	filename := "test_keybinds_desc.json"
	defer os.Remove(filename)

	// Verify description exists before save
	kb, err := km.GetBinding(ActionAttack)
	if err != nil {
		t.Fatalf("failed to get binding: %v", err)
	}
	if kb.Description == "" {
		t.Fatal("expected description before save")
	}

	// Save and load
	if err := km.Save(filename); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	km2 := NewKeybindManager()
	if err := km2.Load(filename); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify description preserved after load
	kb2, err := km2.GetBinding(ActionAttack)
	if err != nil {
		t.Fatalf("failed to get binding after load: %v", err)
	}
	if kb2.Description != kb.Description {
		t.Errorf("description lost after save/load: expected %q, got %q", kb.Description, kb2.Description)
	}
}

func TestQuickTravelManager_LockedDestination(t *testing.T) {
	qtm := NewQuickTravelManager()

	dest := &TravelDestination{
		ID:       "locked",
		Name:     "Locked City",
		X:        500,
		Y:        500,
		Unlocked: false,
	}
	qtm.RegisterDestination(dest)

	gold := 1000
	_, err := qtm.CanTravel(0, 0, "locked", gold)
	if err == nil {
		t.Error("expected error for locked destination")
	}
}

func TestQuickTravelManager_NotFoundDestination(t *testing.T) {
	qtm := NewQuickTravelManager()

	_, err := qtm.CalculateCost(0, 0, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent destination")
	}

	err = qtm.UnlockDestination("nonexistent")
	if err == nil {
		t.Error("expected error for unlocking nonexistent destination")
	}
}

func BenchmarkKeybindManager_GetBinding(b *testing.B) {
	km := NewKeybindManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		km.GetBinding(ActionAttack)
	}
}

func BenchmarkKeybindManager_GetActionForKey(b *testing.B) {
	km := NewKeybindManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		km.GetActionForKey(KeyW)
	}
}

func BenchmarkQuickTravelManager_CalculateCost(b *testing.B) {
	qtm := NewQuickTravelManager()
	qtm.RegisterDestination(&TravelDestination{
		ID: "dest", X: 1000, Y: 1000, Unlocked: true,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qtm.CalculateCost(0, 0, "dest")
	}
}
