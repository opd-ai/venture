package engine

import (
	"testing"
)

func TestFactionTerritoryInfluenceComponent_Type(t *testing.T) {
	comp := NewFactionTerritoryInfluenceComponent("test_faction", 1, 2)
	if comp.Type() != "faction_territory_influence" {
		t.Errorf("expected Type() = 'faction_territory_influence', got '%s'", comp.Type())
	}
}

func TestFactionTerritoryInfluenceComponent_Defaults(t *testing.T) {
	tests := []struct {
		name            string
		factionID       string
		zoneX           int
		zoneZ           int
		wantRadius      float64
		wantStrength    float64
		wantDamageBonus float64
	}{
		{
			name:            "basic creation",
			factionID:       "faction_a",
			zoneX:           0,
			zoneZ:           0,
			wantRadius:      256.0,
			wantStrength:    1.0,
			wantDamageBonus: 0.15,
		},
		{
			name:            "different coordinates",
			factionID:       "faction_b",
			zoneX:           5,
			zoneZ:           -3,
			wantRadius:      256.0,
			wantStrength:    1.0,
			wantDamageBonus: 0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFactionTerritoryInfluenceComponent(tt.factionID, tt.zoneX, tt.zoneZ)

			if comp.FactionID != tt.factionID {
				t.Errorf("FactionID = %s, want %s", comp.FactionID, tt.factionID)
			}
			if comp.ZoneX != tt.zoneX {
				t.Errorf("ZoneX = %d, want %d", comp.ZoneX, tt.zoneX)
			}
			if comp.ZoneZ != tt.zoneZ {
				t.Errorf("ZoneZ = %d, want %d", comp.ZoneZ, tt.zoneZ)
			}
			if comp.ZoneRadius != tt.wantRadius {
				t.Errorf("ZoneRadius = %f, want %f", comp.ZoneRadius, tt.wantRadius)
			}
			if comp.InfluenceStrength != tt.wantStrength {
				t.Errorf("InfluenceStrength = %f, want %f", comp.InfluenceStrength, tt.wantStrength)
			}
			if comp.FriendlyDamageBonus != tt.wantDamageBonus {
				t.Errorf("FriendlyDamageBonus = %f, want %f", comp.FriendlyDamageBonus, tt.wantDamageBonus)
			}
		})
	}
}

func TestFactionTerritoryModifierComponent_Type(t *testing.T) {
	comp := NewFactionTerritoryModifierComponent()
	if comp.Type() != "faction_territory_modifier" {
		t.Errorf("expected Type() = 'faction_territory_modifier', got '%s'", comp.Type())
	}
}

func TestFactionTerritoryModifierComponent_Defaults(t *testing.T) {
	comp := NewFactionTerritoryModifierComponent()

	if comp.EffectiveDamageModifier != 1.0 {
		t.Errorf("EffectiveDamageModifier = %f, want 1.0", comp.EffectiveDamageModifier)
	}
	if comp.EffectiveDefenseModifier != 1.0 {
		t.Errorf("EffectiveDefenseModifier = %f, want 1.0", comp.EffectiveDefenseModifier)
	}
	if comp.EffectiveXPModifier != 1.0 {
		t.Errorf("EffectiveXPModifier = %f, want 1.0", comp.EffectiveXPModifier)
	}
	if comp.InFactionZone {
		t.Error("InFactionZone should be false by default")
	}
	if !comp.Dirty {
		t.Error("Dirty should be true by default")
	}
}

func TestFactionTerritoryModifierComponent_Reset(t *testing.T) {
	comp := NewFactionTerritoryModifierComponent()
	comp.ActiveFactionID = "some_faction"
	comp.ReputationLevel = 75
	comp.EffectiveDamageModifier = 1.5
	comp.InFactionZone = true
	comp.ZoneX = 3
	comp.ZoneZ = 5

	comp.Reset()

	if comp.ActiveFactionID != "" {
		t.Errorf("ActiveFactionID = %s, want empty", comp.ActiveFactionID)
	}
	if comp.ReputationLevel != 0 {
		t.Errorf("ReputationLevel = %d, want 0", comp.ReputationLevel)
	}
	if comp.EffectiveDamageModifier != 1.0 {
		t.Errorf("EffectiveDamageModifier = %f, want 1.0", comp.EffectiveDamageModifier)
	}
	if comp.InFactionZone {
		t.Error("InFactionZone should be false after Reset")
	}
	if comp.Dirty {
		t.Error("Dirty should be false after Reset")
	}
}

func TestFactionTerritoryModifierComponent_ReputationLevels(t *testing.T) {
	tests := []struct {
		name         string
		reputation   int
		wantFriendly bool
		wantHostile  bool
		wantNeutral  bool
	}{
		{"max friendly", 100, true, false, false},
		{"friendly threshold", 51, true, false, false},
		{"neutral high", 50, false, false, true},
		{"neutral zero", 0, false, false, true},
		{"neutral low", -49, false, false, true},
		{"hostile threshold", -50, false, true, false},
		{"max hostile", -100, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFactionTerritoryModifierComponent()
			comp.ReputationLevel = tt.reputation

			if got := comp.IsFriendly(); got != tt.wantFriendly {
				t.Errorf("IsFriendly() = %v, want %v", got, tt.wantFriendly)
			}
			if got := comp.IsHostile(); got != tt.wantHostile {
				t.Errorf("IsHostile() = %v, want %v", got, tt.wantHostile)
			}
			if got := comp.IsNeutral(); got != tt.wantNeutral {
				t.Errorf("IsNeutral() = %v, want %v", got, tt.wantNeutral)
			}
		})
	}
}

func TestFactionTerritoryInfluenceSystem_Creation(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)

	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	if system == nil {
		t.Fatal("NewFactionTerritoryInfluenceSystem returned nil")
	}
	if system.world != world {
		t.Error("System world reference mismatch")
	}
	if system.factionSystem != factionSystem {
		t.Error("System factionSystem reference mismatch")
	}
}

func TestFactionTerritoryInfluenceSystem_NoZones(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create entity with position but no zones exist
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&FactionComponent{FactionID: "player_faction", Reputation: 50, IsPlayerFaction: true})

	entities := []*Entity{entity}

	// Should not panic with no zones
	system.Update(entities, 0.5)

	// Entity should not have modifier added when no zones exist
	// (First update won't process because of interval)
	system.Update(entities, 0.5)
}

func TestFactionTerritoryInfluenceSystem_EntityEntersZone(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create zone at origin
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("test_faction", 0, 0)
	zoneComp.ZoneRadius = 100.0
	zone.AddComponent(zoneComp)

	// Create player inside zone with friendly reputation
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&FactionComponent{FactionID: "test_faction", Reputation: 75, IsPlayerFaction: true})

	entities := []*Entity{zone, player}

	// Force update by accumulating enough time
	system.Update(entities, 0.3)

	// Check modifier was applied
	modComp, ok := player.GetComponent("faction_territory_modifier")
	if !ok {
		t.Fatal("Expected faction_territory_modifier component to be added")
	}
	modifier := modComp.(*FactionTerritoryModifierComponent)

	if !modifier.InFactionZone {
		t.Error("Expected InFactionZone = true")
	}
	if modifier.ActiveFactionID != "test_faction" {
		t.Errorf("ActiveFactionID = %s, want test_faction", modifier.ActiveFactionID)
	}
}

func TestFactionTerritoryInfluenceSystem_FriendlyBonus(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create zone
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("friendly_faction", 0, 0)
	zoneComp.ZoneRadius = 200.0
	zoneComp.FriendlyDamageBonus = 0.20
	zone.AddComponent(zoneComp)

	// Create player with max friendly reputation (member of faction)
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&FactionComponent{FactionID: "friendly_faction", Reputation: 0, IsPlayerFaction: false})

	entities := []*Entity{zone, player}
	system.Update(entities, 0.3)

	modComp, _ := player.GetComponent("faction_territory_modifier")
	modifier := modComp.(*FactionTerritoryModifierComponent)

	// Faction member should get max friendly bonus
	if modifier.EffectiveDamageModifier <= 1.0 {
		t.Errorf("Expected damage modifier > 1.0 for friendly, got %f", modifier.EffectiveDamageModifier)
	}
}

func TestFactionTerritoryInfluenceSystem_HostilePenalty(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create enemy faction zone
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("enemy_faction", 0, 0)
	zoneComp.ZoneRadius = 200.0
	zoneComp.HostileDamagePenalty = 0.20
	zoneComp.HostileDetectionBonus = 0.30
	zone.AddComponent(zoneComp)

	// Add enemy faction with hostile relationship
	enemyFaction := &Faction{
		ID:            "enemy_faction",
		Name:          "Enemy",
		Relationships: map[string]int{"player_faction": -75},
	}
	factionSystem.AddFaction(enemyFaction)

	// Create player with hostile reputation
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&FactionComponent{FactionID: "player_faction", Reputation: -75, IsPlayerFaction: false})

	entities := []*Entity{zone, player}
	system.Update(entities, 0.3)

	modComp, _ := player.GetComponent("faction_territory_modifier")
	modifier := modComp.(*FactionTerritoryModifierComponent)

	// Hostile entity should get damage penalty
	if modifier.EffectiveDamageModifier >= 1.0 {
		t.Errorf("Expected damage modifier < 1.0 for hostile, got %f", modifier.EffectiveDamageModifier)
	}
	// Hostile entity should have increased detection
	if modifier.EffectiveDetectionModifier <= 1.0 {
		t.Errorf("Expected detection modifier > 1.0 for hostile, got %f", modifier.EffectiveDetectionModifier)
	}
}

func TestFactionTerritoryInfluenceSystem_ContestedZone(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create contested zone
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("contested_faction", 0, 0)
	zoneComp.ZoneRadius = 200.0
	zoneComp.IsContested = true
	zoneComp.ContestProgress = 0.5
	zoneComp.ContestingFactionID = "attacker_faction"
	zone.AddComponent(zoneComp)

	// Create member entity
	member := world.CreateEntity()
	member.AddComponent(&PositionComponent{X: 0, Y: 0})
	member.AddComponent(&FactionComponent{FactionID: "contested_faction", Reputation: 0, IsPlayerFaction: false})

	entities := []*Entity{zone, member}
	system.Update(entities, 0.3)

	modComp, _ := member.GetComponent("faction_territory_modifier")
	modifier := modComp.(*FactionTerritoryModifierComponent)

	// Contested zone should reduce bonuses
	// Normal friendly damage would be 1.15, contested should be less
	if modifier.EffectiveDamageModifier >= 1.15 {
		t.Errorf("Expected reduced damage modifier in contested zone, got %f", modifier.EffectiveDamageModifier)
	}
}

func TestFactionTerritoryInfluenceSystem_EntityLeavesZone(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create zone at origin
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("test_faction", 0, 0)
	zoneComp.ZoneRadius = 100.0
	zone.AddComponent(zoneComp)

	// Create player inside zone
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&FactionComponent{FactionID: "test_faction", Reputation: 75, IsPlayerFaction: true})

	entities := []*Entity{zone, player}

	// Enter zone
	system.Update(entities, 0.3)

	modComp, _ := player.GetComponent("faction_territory_modifier")
	modifier := modComp.(*FactionTerritoryModifierComponent)

	if !modifier.InFactionZone {
		t.Fatal("Expected entity to be in zone after first update")
	}

	// Move player outside zone
	posComp, _ := player.GetComponent("position")
	pos := posComp.(*PositionComponent)
	pos.X = 500
	pos.Y = 500

	// Leave zone
	system.Update(entities, 0.3)

	if modifier.InFactionZone {
		t.Error("Expected entity to be outside zone after moving")
	}
	if modifier.EffectiveDamageModifier != 1.0 {
		t.Errorf("Expected damage modifier reset to 1.0, got %f", modifier.EffectiveDamageModifier)
	}
}

func TestFactionTerritoryInfluenceSystem_CreateFactionZone(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	zone := system.CreateFactionZone("new_faction", 2, 3, 150.0)

	if zone == nil {
		t.Fatal("CreateFactionZone returned nil")
	}

	zoneComp, ok := zone.GetComponent("faction_territory_influence")
	if !ok {
		t.Fatal("Created zone missing faction_territory_influence component")
	}
	influence := zoneComp.(*FactionTerritoryInfluenceComponent)

	if influence.FactionID != "new_faction" {
		t.Errorf("FactionID = %s, want new_faction", influence.FactionID)
	}
	if influence.ZoneX != 2 {
		t.Errorf("ZoneX = %d, want 2", influence.ZoneX)
	}
	if influence.ZoneZ != 3 {
		t.Errorf("ZoneZ = %d, want 3", influence.ZoneZ)
	}
	if influence.ZoneRadius != 150.0 {
		t.Errorf("ZoneRadius = %f, want 150.0", influence.ZoneRadius)
	}
}

func TestFactionTerritoryInfluenceSystem_GetEntityZoneModifier(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Entity without modifier
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 0, Y: 0})

	if modifier := system.GetEntityZoneModifier(entity1); modifier != nil {
		t.Error("Expected nil modifier for entity without component")
	}

	// Entity with modifier
	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity2.AddComponent(NewFactionTerritoryModifierComponent())

	modifier := system.GetEntityZoneModifier(entity2)
	if modifier == nil {
		t.Error("Expected non-nil modifier for entity with component")
	}
}

func TestFactionTerritoryInfluenceSystem_MultipleZones(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create two overlapping zones
	zone1 := world.CreateEntity()
	zone1Comp := NewFactionTerritoryInfluenceComponent("faction_a", 0, 0)
	zone1Comp.ZoneRadius = 200.0
	zone1Comp.InfluenceStrength = 0.5
	zone1.AddComponent(zone1Comp)

	zone2 := world.CreateEntity()
	zone2Comp := NewFactionTerritoryInfluenceComponent("faction_b", 0, 0)
	zone2Comp.ZoneRadius = 200.0
	zone2Comp.InfluenceStrength = 1.0 // Stronger
	zone2.AddComponent(zone2Comp)

	// Create player at origin (in both zones)
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&FactionComponent{FactionID: "faction_b", Reputation: 75, IsPlayerFaction: true})

	entities := []*Entity{zone1, zone2, player}
	system.Update(entities, 0.3)

	modComp, _ := player.GetComponent("faction_territory_modifier")
	modifier := modComp.(*FactionTerritoryModifierComponent)

	// Stronger zone should win
	if modifier.ActiveFactionID != "faction_b" {
		t.Errorf("Expected stronger faction_b zone to apply, got %s", modifier.ActiveFactionID)
	}
}

func TestFactionTerritoryInfluenceSystem_NoPositionEntity(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create zone
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("test_faction", 0, 0)
	zoneComp.ZoneRadius = 200.0
	zone.AddComponent(zoneComp)

	// Create entity without position
	entity := world.CreateEntity()
	entity.AddComponent(&FactionComponent{FactionID: "test_faction", Reputation: 75, IsPlayerFaction: true})

	entities := []*Entity{zone, entity}

	// Should not panic
	system.Update(entities, 0.3)

	// Should not add modifier
	_, ok := entity.GetComponent("faction_territory_modifier")
	if ok {
		t.Error("Entity without position should not get modifier")
	}
}

func TestFactionTerritoryInfluenceSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create zone
	zone := world.CreateEntity()
	zoneComp := NewFactionTerritoryInfluenceComponent("test_faction", 0, 0)
	zoneComp.ZoneRadius = 200.0
	zone.AddComponent(zoneComp)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&FactionComponent{FactionID: "test_faction", Reputation: 75, IsPlayerFaction: true})

	entities := []*Entity{zone, player}

	// Small update should not process (0.1 < 0.25 interval)
	system.Update(entities, 0.1)

	_, ok := player.GetComponent("faction_territory_modifier")
	if ok {
		t.Error("Expected no modifier after small delta time")
	}

	// Second update should trigger (0.1 + 0.2 > 0.25)
	system.Update(entities, 0.2)

	_, ok = player.GetComponent("faction_territory_modifier")
	if !ok {
		t.Error("Expected modifier after accumulated time exceeds interval")
	}
}

func BenchmarkFactionTerritoryInfluenceSystem_Update(b *testing.B) {
	world := NewWorld()
	factionSystem := NewFactionSystem(world, nil)
	system := NewFactionTerritoryInfluenceSystem(world, factionSystem)

	// Create 10 zones
	for i := 0; i < 10; i++ {
		zone := world.CreateEntity()
		zoneComp := NewFactionTerritoryInfluenceComponent("faction_"+string(rune('A'+i)), i, i)
		zoneComp.ZoneRadius = 300.0
		zone.AddComponent(zoneComp)
	}

	// Create 100 entities
	entities := make([]*Entity, 110)
	for i := 0; i < 10; i++ {
		entities[i] = world.GetEntity(uint64(i + 1))
	}
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&FactionComponent{FactionID: "player_faction", Reputation: 50, IsPlayerFaction: true})
		entities[10+i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.3)
	}
}
