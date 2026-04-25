package engine

import (
	"testing"
)

func TestAffinityType_String(t *testing.T) {
	tests := []struct {
		name     string
		affinity AffinityType
		want     string
	}{
		{"none", AffinityNone, "None"},
		{"aggressor", AffinityAggressor, "Aggressor"},
		{"defender", AffinityDefender, "Defender"},
		{"caster", AffinityCaster, "Caster"},
		{"supportive", AffinitySupportive, "Supportive"},
		{"stealthy", AffinityStealthy, "Stealthy"},
		{"tactical", AffinityTactical, "Tactical"},
		{"burst damage", AffinityBurstDamage, "Burst Damage"},
		{"area damage", AffinityAreaDamage, "Area Damage"},
		{"drainer", AffinityDrainer, "Drainer"},
		{"summoner", AffinitySummoner, "Summoner"},
		{"unknown", AffinityType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.affinity.String(); got != tt.want {
				t.Errorf("AffinityType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAffinityType_Description(t *testing.T) {
	tests := []struct {
		name     string
		affinity AffinityType
		wantLen  bool // Just check it has content
	}{
		{"none", AffinityNone, true},
		{"aggressor", AffinityAggressor, true},
		{"defender", AffinityDefender, true},
		{"caster", AffinityCaster, true},
		{"unknown", AffinityType(99), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.affinity.Description()
			if tt.wantLen && len(desc) == 0 {
				t.Errorf("AffinityType.Description() returned empty string")
			}
		})
	}
}

func TestAffinityLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level AffinityLevel
		want  string
	}{
		{"none", AffinityLevelNone, "None"},
		{"novice", AffinityLevelNovice, "Novice"},
		{"apprentice", AffinityLevelApprentice, "Apprentice"},
		{"journeyman", AffinityLevelJourneyman, "Journeyman"},
		{"expert", AffinityLevelExpert, "Expert"},
		{"master", AffinityLevelMaster, "Master"},
		{"grandmaster", AffinityLevelGrandmaster, "Grandmaster"},
		{"unknown", AffinityLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("AffinityLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAffinityLevel_XPThreshold(t *testing.T) {
	tests := []struct {
		name  string
		level AffinityLevel
		want  int
	}{
		{"none", AffinityLevelNone, 0},
		{"novice", AffinityLevelNovice, 100},
		{"apprentice", AffinityLevelApprentice, 500},
		{"journeyman", AffinityLevelJourneyman, 1500},
		{"expert", AffinityLevelExpert, 4000},
		{"master", AffinityLevelMaster, 10000},
		{"grandmaster", AffinityLevelGrandmaster, 25000},
		{"unknown", AffinityLevel(99), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.XPThreshold(); got != tt.want {
				t.Errorf("AffinityLevel.XPThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewClassAffinityComponent(t *testing.T) {
	comp := NewClassAffinityComponent()

	if comp == nil {
		t.Fatal("NewClassAffinityComponent returned nil")
	}
	if comp.Affinities == nil {
		t.Error("Affinities map is nil")
	}
	if comp.BonusesApplied == nil {
		t.Error("BonusesApplied map is nil")
	}
	if comp.PrimaryAffinity != AffinityNone {
		t.Errorf("PrimaryAffinity = %v, want %v", comp.PrimaryAffinity, AffinityNone)
	}
	if comp.Type() != "class_affinity" {
		t.Errorf("Type() = %v, want class_affinity", comp.Type())
	}
}

func TestClassAffinityComponent_GetAffinity(t *testing.T) {
	comp := NewClassAffinityComponent()

	// Should create new affinity data if not exists
	data := comp.GetAffinity(AffinityAggressor)
	if data == nil {
		t.Fatal("GetAffinity returned nil")
	}
	if data.Level != AffinityLevelNone {
		t.Errorf("Initial level = %v, want %v", data.Level, AffinityLevelNone)
	}
	if data.XP != 0 {
		t.Errorf("Initial XP = %v, want 0", data.XP)
	}

	// Should return same data on second call
	data2 := comp.GetAffinity(AffinityAggressor)
	if data != data2 {
		t.Error("GetAffinity returned different instance")
	}
}

func TestClassAffinityComponent_GetAffinityLevel(t *testing.T) {
	comp := NewClassAffinityComponent()

	// Initial level should be None
	level := comp.GetAffinityLevel(AffinityCaster)
	if level != AffinityLevelNone {
		t.Errorf("GetAffinityLevel() = %v, want %v", level, AffinityLevelNone)
	}

	// Set some XP and check level
	data := comp.GetAffinity(AffinityCaster)
	data.XP = 600
	data.Level = AffinityLevelApprentice

	level = comp.GetAffinityLevel(AffinityCaster)
	if level != AffinityLevelApprentice {
		t.Errorf("GetAffinityLevel() = %v, want %v", level, AffinityLevelApprentice)
	}
}

func TestClassAffinityComponent_GetXPToNextLevel(t *testing.T) {
	comp := NewClassAffinityComponent()
	data := comp.GetAffinity(AffinityAggressor)

	tests := []struct {
		name      string
		xp        int
		level     AffinityLevel
		wantRange [2]int // [min, max] expected range
	}{
		{"at start", 0, AffinityLevelNone, [2]int{100, 100}},
		{"halfway to novice", 50, AffinityLevelNone, [2]int{50, 50}},
		{"novice to apprentice", 100, AffinityLevelNovice, [2]int{400, 400}},
		{"at grandmaster", 25000, AffinityLevelGrandmaster, [2]int{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data.XP = tt.xp
			data.Level = tt.level
			got := comp.GetXPToNextLevel(AffinityAggressor)
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("GetXPToNextLevel() = %v, want in range %v", got, tt.wantRange)
			}
		})
	}
}

func TestClassAffinityComponent_GetProgressToNextLevel(t *testing.T) {
	comp := NewClassAffinityComponent()
	data := comp.GetAffinity(AffinityDefender)

	tests := []struct {
		name  string
		xp    int
		level AffinityLevel
		want  float64
	}{
		{"at start", 0, AffinityLevelNone, 0.0},
		{"halfway to novice", 50, AffinityLevelNone, 0.5},
		{"at novice", 100, AffinityLevelNovice, 0.0},
		{"at grandmaster", 25000, AffinityLevelGrandmaster, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data.XP = tt.xp
			data.Level = tt.level
			got := comp.GetProgressToNextLevel(AffinityDefender)
			if got < tt.want-0.01 || got > tt.want+0.01 {
				t.Errorf("GetProgressToNextLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassAffinityComponent_RecalculatePrimaryAffinities(t *testing.T) {
	comp := NewClassAffinityComponent()

	// No affinities yet
	comp.RecalculatePrimaryAffinities()
	if comp.PrimaryAffinity != AffinityNone {
		t.Errorf("PrimaryAffinity = %v, want %v", comp.PrimaryAffinity, AffinityNone)
	}

	// Add some affinity XP
	comp.GetAffinity(AffinityAggressor).XP = 500
	comp.GetAffinity(AffinityCaster).XP = 300
	comp.GetAffinity(AffinityDefender).XP = 100

	comp.RecalculatePrimaryAffinities()

	if comp.PrimaryAffinity != AffinityAggressor {
		t.Errorf("PrimaryAffinity = %v, want %v", comp.PrimaryAffinity, AffinityAggressor)
	}
	if comp.SecondaryAffinity != AffinityCaster {
		t.Errorf("SecondaryAffinity = %v, want %v", comp.SecondaryAffinity, AffinityCaster)
	}
}

func TestGetAffinityBonuses(t *testing.T) {
	tests := []struct {
		name           string
		affinity       AffinityType
		level          AffinityLevel
		checkDamageMul bool
		minDamageMul   float64
	}{
		{"aggressor none", AffinityAggressor, AffinityLevelNone, true, 1.0},
		{"aggressor grandmaster", AffinityAggressor, AffinityLevelGrandmaster, true, 1.2},
		{"caster grandmaster", AffinityCaster, AffinityLevelGrandmaster, false, 0},
		{"defender grandmaster", AffinityDefender, AffinityLevelGrandmaster, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonuses := GetAffinityBonuses(tt.affinity, tt.level)
			if tt.checkDamageMul && bonuses.DamageMultiplier < tt.minDamageMul {
				t.Errorf("DamageMultiplier = %v, want >= %v", bonuses.DamageMultiplier, tt.minDamageMul)
			}
		})
	}
}

func TestGetAbilityAffinities(t *testing.T) {
	tests := []struct {
		name      string
		abilityID string
		wantLen   int
		wantFirst AffinityType
	}{
		{"power_strike", "power_strike", 2, AffinityAggressor},
		{"heal", "heal", 1, AffinitySupportive},
		{"fireball", "fireball", 2, AffinityCaster},
		{"stealth", "stealth", 1, AffinityStealthy},
		{"unknown", "unknown_ability", 0, AffinityNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			affinities := GetAbilityAffinities(tt.abilityID)
			if len(affinities) != tt.wantLen {
				t.Errorf("len(GetAbilityAffinities(%q)) = %v, want %v", tt.abilityID, len(affinities), tt.wantLen)
			}
			if tt.wantLen > 0 && affinities[0] != tt.wantFirst {
				t.Errorf("GetAbilityAffinities(%q)[0] = %v, want %v", tt.abilityID, affinities[0], tt.wantFirst)
			}
		})
	}
}

func TestClassAffinityComponent_SerializeDeserialize(t *testing.T) {
	comp := NewClassAffinityComponent()

	// Set up some data
	comp.GetAffinity(AffinityAggressor).XP = 500
	comp.GetAffinity(AffinityAggressor).Level = AffinityLevelApprentice
	comp.GetAffinity(AffinityAggressor).AbilitiesUsed = 50
	comp.GetAffinity(AffinityAggressor).DamageDealt = 10000.5
	comp.GetAffinity(AffinityAggressor).TimesTriggered = 10

	comp.GetAffinity(AffinityCaster).XP = 200
	comp.GetAffinity(AffinityCaster).Level = AffinityLevelNovice

	comp.PrimaryAffinity = AffinityAggressor
	comp.SecondaryAffinity = AffinityCaster
	comp.TotalAffinityXP = 700

	// Serialize
	data := comp.Serialize()
	if len(data) == 0 {
		t.Fatal("Serialize returned empty data")
	}

	// Deserialize into new component
	comp2 := NewClassAffinityComponent()
	err := comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data
	if comp2.TotalAffinityXP != 700 {
		t.Errorf("TotalAffinityXP = %v, want 700", comp2.TotalAffinityXP)
	}
	if comp2.PrimaryAffinity != AffinityAggressor {
		t.Errorf("PrimaryAffinity = %v, want %v", comp2.PrimaryAffinity, AffinityAggressor)
	}

	aggData := comp2.GetAffinity(AffinityAggressor)
	if aggData.XP != 500 {
		t.Errorf("Aggressor XP = %v, want 500", aggData.XP)
	}
	if aggData.AbilitiesUsed != 50 {
		t.Errorf("Aggressor AbilitiesUsed = %v, want 50", aggData.AbilitiesUsed)
	}
}

func TestClassAffinityComponent_DeserializeInvalidData(t *testing.T) {
	comp := NewClassAffinityComponent()

	// Too short data
	err := comp.Deserialize([]byte{1, 2})
	if err == nil {
		t.Error("Deserialize should fail with short data")
	}
}

func TestNewClassAffinitySystem(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	if system == nil {
		t.Fatal("NewClassAffinitySystem returned nil")
	}
	if system.world != world {
		t.Error("System world reference incorrect")
	}
	if system.xpPerAbilityUse <= 0 {
		t.Error("xpPerAbilityUse should be positive")
	}
}

func TestClassAffinitySystem_Update(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())
	entity.AddComponent(&StatsComponent{})

	entities := []*Entity{entity}

	// Set dirty flag
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.Dirty = true
	affinity.GetAffinity(AffinityAggressor).XP = 500
	affinity.GetAffinity(AffinityAggressor).Level = AffinityLevelApprentice

	// Force immediate update by setting frameCounter high
	system.frameCounter = system.updateInterval - 1
	system.Update(entities, 0.016)

	// After update, dirty should be false
	if affinity.Dirty {
		t.Error("Dirty flag should be cleared after update")
	}
}

func TestClassAffinitySystem_OnAbilityUsed(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Use an ability
	system.OnAbilityUsed(entity, "fireball", 500.0, 0)

	// Check that affinity XP was awarded
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)

	casterData := affinity.GetAffinity(AffinityCaster)
	if casterData.XP <= 0 {
		t.Error("Caster affinity should have received XP from fireball")
	}
	if casterData.AbilitiesUsed != 1 {
		t.Errorf("AbilitiesUsed = %v, want 1", casterData.AbilitiesUsed)
	}

	// Area damage affinity should also have received XP
	areaData := affinity.GetAffinity(AffinityAreaDamage)
	if areaData.XP <= 0 {
		t.Error("Area Damage affinity should have received XP from fireball")
	}
}

func TestClassAffinitySystem_OnAbilityUsed_WithStreaks(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Use same-affinity abilities multiple times
	for i := 0; i < 5; i++ {
		system.OnAbilityUsed(entity, "fireball", 100.0, 0)
	}

	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	casterData := affinity.GetAffinity(AffinityCaster)

	if casterData.CurrentStreak != 5 {
		t.Errorf("CurrentStreak = %v, want 5", casterData.CurrentStreak)
	}
	if casterData.PeakStreak != 5 {
		t.Errorf("PeakStreak = %v, want 5", casterData.PeakStreak)
	}
}

func TestClassAffinitySystem_OnKill(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Get initial XP
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)

	// Simulate kill with fireball
	system.OnKill(entity, "fireball")

	casterData := affinity.GetAffinity(AffinityCaster)
	if casterData.XP < system.xpPerKill {
		t.Errorf("XP after kill = %v, want >= %v", casterData.XP, system.xpPerKill)
	}
	if casterData.TimesTriggered != 1 {
		t.Errorf("TimesTriggered = %v, want 1", casterData.TimesTriggered)
	}
}

func TestClassAffinitySystem_GetDamageMultiplier(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// With no affinities, should return 1.0
	mult := system.GetDamageMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("GetDamageMultiplier() = %v, want 1.0", mult)
	}

	// Set up aggressor affinity
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinityAggressor).XP = 25000
	affinity.GetAffinity(AffinityAggressor).Level = AffinityLevelGrandmaster
	affinity.PrimaryAffinity = AffinityAggressor

	mult = system.GetDamageMultiplier(entity)
	if mult <= 1.0 {
		t.Errorf("GetDamageMultiplier() with Grandmaster Aggressor = %v, want > 1.0", mult)
	}
}

func TestClassAffinitySystem_GetSpellDamageMultiplier(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// With no affinities, should return 1.0
	mult := system.GetSpellDamageMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("GetSpellDamageMultiplier() = %v, want 1.0", mult)
	}

	// Set up caster affinity
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinityCaster).XP = 25000
	affinity.GetAffinity(AffinityCaster).Level = AffinityLevelGrandmaster
	affinity.PrimaryAffinity = AffinityCaster

	mult = system.GetSpellDamageMultiplier(entity)
	if mult <= 1.0 {
		t.Errorf("GetSpellDamageMultiplier() with Grandmaster Caster = %v, want > 1.0", mult)
	}
}

func TestClassAffinitySystem_GetHealingMultiplier(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Set up supportive affinity
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinitySupportive).XP = 25000
	affinity.GetAffinity(AffinitySupportive).Level = AffinityLevelGrandmaster
	affinity.PrimaryAffinity = AffinitySupportive

	mult := system.GetHealingMultiplier(entity)
	if mult <= 1.0 {
		t.Errorf("GetHealingMultiplier() with Grandmaster Supportive = %v, want > 1.0", mult)
	}
}

func TestClassAffinitySystem_GetCooldownReduction(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// With no affinities, should return 0
	reduction := system.GetCooldownReduction(entity)
	if reduction != 0.0 {
		t.Errorf("GetCooldownReduction() = %v, want 0.0", reduction)
	}

	// Set up tactical affinity (has high cooldown reduction)
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinityTactical).XP = 25000
	affinity.GetAffinity(AffinityTactical).Level = AffinityLevelGrandmaster
	affinity.PrimaryAffinity = AffinityTactical

	reduction = system.GetCooldownReduction(entity)
	if reduction <= 0 {
		t.Errorf("GetCooldownReduction() with Grandmaster Tactical = %v, want > 0", reduction)
	}
	if reduction > 0.40 {
		t.Errorf("GetCooldownReduction() should be capped at 0.40, got %v", reduction)
	}
}

func TestClassAffinitySystem_SetOnAffinityLevelUp(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	callbackCalled := false
	var callbackAffinity AffinityType
	var callbackLevel AffinityLevel

	system.SetOnAffinityLevelUp(func(entity *Entity, affinity AffinityType, newLevel AffinityLevel) {
		callbackCalled = true
		callbackAffinity = affinity
		callbackLevel = newLevel
	})

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Award enough XP for level up
	for i := 0; i < 20; i++ {
		system.OnAbilityUsed(entity, "fireball", 500.0, 0)
	}

	if !callbackCalled {
		t.Error("Level up callback was not called")
	}
	if callbackAffinity != AffinityCaster && callbackAffinity != AffinityAreaDamage {
		t.Errorf("Callback affinity = %v, expected Caster or AreaDamage", callbackAffinity)
	}
	if callbackLevel < AffinityLevelNovice {
		t.Errorf("Callback level = %v, want >= Novice", callbackLevel)
	}
}

func TestEnsureClassAffinityComponent(t *testing.T) {
	// Nil entity
	comp := EnsureClassAffinityComponent(nil)
	if comp != nil {
		t.Error("Should return nil for nil entity")
	}

	// Entity without component
	entity := NewEntity(1)
	comp = EnsureClassAffinityComponent(entity)
	if comp == nil {
		t.Fatal("Should create component for entity")
	}

	// Entity with existing component
	comp2 := EnsureClassAffinityComponent(entity)
	if comp2 != comp {
		t.Error("Should return existing component")
	}
}

func TestClassAffinitySystem_GetAffinityStatusForUI(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Set up some affinity data
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinityAggressor).XP = 600
	affinity.GetAffinity(AffinityAggressor).Level = AffinityLevelApprentice
	affinity.GetAffinity(AffinityAggressor).AbilitiesUsed = 30
	affinity.GetAffinity(AffinityAggressor).DamageDealt = 5000.0
	affinity.PrimaryAffinity = AffinityAggressor

	status := system.GetAffinityStatusForUI(entity, AffinityAggressor)
	if status == nil {
		t.Fatal("GetAffinityStatusForUI returned nil")
	}

	if status.Name != "Aggressor" {
		t.Errorf("Name = %v, want Aggressor", status.Name)
	}
	if status.LevelName != "Apprentice" {
		t.Errorf("LevelName = %v, want Apprentice", status.LevelName)
	}
	if status.CurrentXP != 600 {
		t.Errorf("CurrentXP = %v, want 600", status.CurrentXP)
	}
	if status.AbilitiesUsed != 30 {
		t.Errorf("AbilitiesUsed = %v, want 30", status.AbilitiesUsed)
	}
	if !status.IsPrimary {
		t.Error("IsPrimary should be true")
	}
}

func TestClassAffinitySystem_GetAllAffinities(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Initially empty
	all := system.GetAllAffinities(entity)
	if len(all) != 0 {
		t.Errorf("GetAllAffinities() = %v items, want 0", len(all))
	}

	// Add some affinity data
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.GetAffinity(AffinityAggressor).XP = 100
	affinity.GetAffinity(AffinityCaster).XP = 50

	all = system.GetAllAffinities(entity)
	if len(all) != 2 {
		t.Errorf("GetAllAffinities() = %v items, want 2", len(all))
	}
}

func TestClassAffinitySystem_GetPrimaryAffinityName(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	// Initially "None"
	name := system.GetPrimaryAffinityName(entity)
	if name != "None" {
		t.Errorf("GetPrimaryAffinityName() = %v, want None", name)
	}

	// Set primary affinity
	comp, _ := entity.GetComponent("class_affinity")
	affinity := comp.(*ClassAffinityComponent)
	affinity.PrimaryAffinity = AffinityStealthy

	name = system.GetPrimaryAffinityName(entity)
	if name != "Stealthy" {
		t.Errorf("GetPrimaryAffinityName() = %v, want Stealthy", name)
	}
}

func TestClassAffinitySystem_calculateAffinityLevel(t *testing.T) {
	system := &ClassAffinitySystem{}

	tests := []struct {
		name string
		xp   int
		want AffinityLevel
	}{
		{"zero", 0, AffinityLevelNone},
		{"below novice", 50, AffinityLevelNone},
		{"at novice", 100, AffinityLevelNovice},
		{"at apprentice", 500, AffinityLevelApprentice},
		{"at journeyman", 1500, AffinityLevelJourneyman},
		{"at expert", 4000, AffinityLevelExpert},
		{"at master", 10000, AffinityLevelMaster},
		{"at grandmaster", 25000, AffinityLevelGrandmaster},
		{"above grandmaster", 50000, AffinityLevelGrandmaster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := system.calculateAffinityLevel(tt.xp); got != tt.want {
				t.Errorf("calculateAffinityLevel(%v) = %v, want %v", tt.xp, got, tt.want)
			}
		})
	}
}

func TestClassAffinitySystem_NoComponentEntity(t *testing.T) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1) // No affinity component

	// These should not panic and return safe defaults
	mult := system.GetDamageMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("GetDamageMultiplier() without component = %v, want 1.0", mult)
	}

	spellMult := system.GetSpellDamageMultiplier(entity)
	if spellMult != 1.0 {
		t.Errorf("GetSpellDamageMultiplier() without component = %v, want 1.0", spellMult)
	}

	healMult := system.GetHealingMultiplier(entity)
	if healMult != 1.0 {
		t.Errorf("GetHealingMultiplier() without component = %v, want 1.0", healMult)
	}

	cooldown := system.GetCooldownReduction(entity)
	if cooldown != 0.0 {
		t.Errorf("GetCooldownReduction() without component = %v, want 0.0", cooldown)
	}

	name := system.GetPrimaryAffinityName(entity)
	if name != "None" {
		t.Errorf("GetPrimaryAffinityName() without component = %v, want None", name)
	}
}

func BenchmarkClassAffinitySystem_OnAbilityUsed(b *testing.B) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	entity.AddComponent(NewClassAffinityComponent())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnAbilityUsed(entity, "fireball", 100.0, 0)
	}
}

func BenchmarkClassAffinitySystem_GetDamageMultiplier(b *testing.B) {
	world := &World{}
	system := NewClassAffinitySystem(world)

	entity := NewEntity(1)
	comp := NewClassAffinityComponent()
	comp.GetAffinity(AffinityAggressor).Level = AffinityLevelGrandmaster
	comp.PrimaryAffinity = AffinityAggressor
	entity.AddComponent(comp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetDamageMultiplier(entity)
	}
}

func BenchmarkClassAffinityComponent_Serialize(b *testing.B) {
	comp := NewClassAffinityComponent()
	comp.GetAffinity(AffinityAggressor).XP = 1000
	comp.GetAffinity(AffinityCaster).XP = 500
	comp.GetAffinity(AffinityDefender).XP = 300

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Serialize()
	}
}

// TestG37_ClassAffinity_ManaRegenPrecision verifies that mana regen removal
// uses the stored absolute value, not a recomputed one, so that changes to
// mana.Max between apply and removal do not leave a residual regen bonus.
func TestG37_ClassAffinity_ManaRegenPrecision(t *testing.T) {
	world := NewWorld()
	system := NewClassAffinitySystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&StatsComponent{})
	mana := &ManaComponent{Current: 100, Max: 100, Regen: 0.0}
	entity.AddComponent(mana)

	affinityComp := NewClassAffinityComponent()
	// Use AffinityAggressor which has a mana regen bonus defined.
	affinityComp.Affinities[AffinityAggressor] = &AffinityData{
		Level: AffinityLevelNovice,
	}
	affinityComp.PrimaryAffinity = AffinityAggressor
	affinityComp.Dirty = true
	entity.AddComponent(affinityComp)

	entities := []*Entity{entity}

	// First batch of frames with mana.Max = 100
	for i := 0; i < 35; i++ { // >30 to pass updateInterval
		system.Update(entities, 1.0/60.0)
	}

	regenAfterFirstApply := mana.Regen

	// Simulate a mana.Max change (e.g. equipment upgrade).
	mana.Max = 200

	// Force re-apply by marking dirty (level change).
	affinityComp.Affinities[AffinityAggressor].Level = AffinityLevelApprentice
	affinityComp.Dirty = true

	for i := 0; i < 35; i++ {
		system.Update(entities, 1.0/60.0)
	}

	regenAfterSecondApply := mana.Regen

	// Now revert back to Novice and check there's no residual regen.
	affinityComp.Affinities[AffinityAggressor].Level = AffinityLevelNovice
	affinityComp.Dirty = true

	for i := 0; i < 35; i++ {
		system.Update(entities, 1.0/60.0)
	}

	// After reverting the level, mana.Regen must not be negative.
	// If G37 is NOT fixed, removal would recompute using mana.Max=200,
	// subtracting more than was added, yielding a negative regen.
	if mana.Regen < 0 {
		t.Errorf("G37: mana.Regen is negative (%.4f): removal used wrong Max", mana.Regen)
	}
	_ = regenAfterFirstApply
	_ = regenAfterSecondApply
}
