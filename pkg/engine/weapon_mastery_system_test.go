package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestWeaponMasteryComponent_Type(t *testing.T) {
	comp := NewWeaponMasteryComponent()
	if comp.Type() != "weapon_mastery" {
		t.Errorf("expected type 'weapon_mastery', got %q", comp.Type())
	}
}

func TestWeaponMasteryComponent_GetMastery(t *testing.T) {
	comp := NewWeaponMasteryComponent()

	// First call should create new entry
	data := comp.GetMastery("sword")
	if data == nil {
		t.Fatal("expected non-nil mastery data")
	}
	if data.XP != 0 {
		t.Errorf("expected XP 0, got %d", data.XP)
	}
	if data.Level != MasteryNovice {
		t.Errorf("expected MasteryNovice, got %v", data.Level)
	}

	// Second call should return same entry
	data2 := comp.GetMastery("sword")
	if data != data2 {
		t.Error("expected same pointer for repeated GetMastery call")
	}

	// Different weapon should have independent data
	data3 := comp.GetMastery("bow")
	if data == data3 {
		t.Error("different weapon types should have different data")
	}
}

func TestWeaponMasteryLevel_XPThreshold(t *testing.T) {
	tests := []struct {
		level    WeaponMasteryLevel
		expected int
	}{
		{MasteryNovice, 0},
		{MasteryApprentice, 100},
		{MasteryJourneyman, 300},
		{MasteryExpert, 600},
		{MasteryMaster, 1000},
		{MasteryGrandmaster, 1500},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := tt.level.XPThreshold(); got != tt.expected {
				t.Errorf("XPThreshold() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestWeaponMasteryLevel_String(t *testing.T) {
	tests := []struct {
		level    WeaponMasteryLevel
		expected string
	}{
		{MasteryNovice, "Novice"},
		{MasteryApprentice, "Apprentice"},
		{MasteryJourneyman, "Journeyman"},
		{MasteryExpert, "Expert"},
		{MasteryMaster, "Master"},
		{MasteryGrandmaster, "Grandmaster"},
		{WeaponMasteryLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWeaponMasteryComponent_GetXPToNextLevel(t *testing.T) {
	comp := NewWeaponMasteryComponent()

	// Novice with 0 XP needs 100 for Apprentice
	xpNeeded := comp.GetXPToNextLevel("sword")
	if xpNeeded != 100 {
		t.Errorf("expected 100 XP to next level, got %d", xpNeeded)
	}

	// Set to 50 XP, needs 50 more
	comp.Mastery["sword"].XP = 50
	xpNeeded = comp.GetXPToNextLevel("sword")
	if xpNeeded != 50 {
		t.Errorf("expected 50 XP to next level, got %d", xpNeeded)
	}

	// At Grandmaster, needs 0
	comp.Mastery["sword"].XP = 1500
	comp.Mastery["sword"].Level = MasteryGrandmaster
	xpNeeded = comp.GetXPToNextLevel("sword")
	if xpNeeded != 0 {
		t.Errorf("expected 0 XP at max level, got %d", xpNeeded)
	}
}

func TestWeaponMasteryComponent_GetProgressToNextLevel(t *testing.T) {
	comp := NewWeaponMasteryComponent()

	// Novice 0/100 = 0%
	progress := comp.GetProgressToNextLevel("sword")
	if progress != 0.0 {
		t.Errorf("expected 0.0 progress, got %f", progress)
	}

	// Novice 50/100 = 50%
	comp.Mastery["sword"].XP = 50
	progress = comp.GetProgressToNextLevel("sword")
	if progress != 0.5 {
		t.Errorf("expected 0.5 progress, got %f", progress)
	}

	// Grandmaster = 100%
	comp.Mastery["sword"].Level = MasteryGrandmaster
	comp.Mastery["sword"].XP = 1500
	progress = comp.GetProgressToNextLevel("sword")
	if progress != 1.0 {
		t.Errorf("expected 1.0 progress at max, got %f", progress)
	}
}

func TestWeaponMasteryComponent_Serialize(t *testing.T) {
	comp := NewWeaponMasteryComponent()
	comp.Mastery["sword"] = &WeaponMasteryData{
		XP:           150,
		Level:        MasteryApprentice,
		TotalKills:   10,
		TotalDamage:  500.5,
		CriticalHits: 3,
	}
	comp.TotalMasteryXP = 150

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty serialized data")
	}

	// Deserialize into new component
	comp2 := &WeaponMasteryComponent{}
	err = comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if comp2.TotalMasteryXP != 150 {
		t.Errorf("expected TotalMasteryXP 150, got %d", comp2.TotalMasteryXP)
	}
	if comp2.Mastery["sword"].XP != 150 {
		t.Errorf("expected sword XP 150, got %d", comp2.Mastery["sword"].XP)
	}
	if comp2.Mastery["sword"].Level != MasteryApprentice {
		t.Errorf("expected MasteryApprentice, got %v", comp2.Mastery["sword"].Level)
	}
	if !comp2.Dirty {
		t.Error("expected Dirty=true after deserialize")
	}
}

func TestWeaponMasteryComponent_GetHighestMasteryLevel(t *testing.T) {
	comp := NewWeaponMasteryComponent()

	// Empty = Novice
	if comp.GetHighestMasteryLevel() != MasteryNovice {
		t.Error("expected Novice for empty component")
	}

	// Add some masteries
	comp.Mastery["sword"] = &WeaponMasteryData{Level: MasteryExpert}
	comp.Mastery["bow"] = &WeaponMasteryData{Level: MasteryJourneyman}
	comp.Mastery["axe"] = &WeaponMasteryData{Level: MasteryMaster}

	if comp.GetHighestMasteryLevel() != MasteryMaster {
		t.Errorf("expected MasteryMaster, got %v", comp.GetHighestMasteryLevel())
	}
}

func TestWeaponMasteryComponent_GetMasteredWeaponCount(t *testing.T) {
	comp := NewWeaponMasteryComponent()
	comp.Mastery["sword"] = &WeaponMasteryData{Level: MasteryExpert}
	comp.Mastery["bow"] = &WeaponMasteryData{Level: MasteryJourneyman}
	comp.Mastery["axe"] = &WeaponMasteryData{Level: MasteryNovice}

	// Count at Journeyman or above
	count := comp.GetMasteredWeaponCount(MasteryJourneyman)
	if count != 2 {
		t.Errorf("expected 2 weapons at Journeyman+, got %d", count)
	}

	// Count at Expert or above
	count = comp.GetMasteredWeaponCount(MasteryExpert)
	if count != 1 {
		t.Errorf("expected 1 weapon at Expert+, got %d", count)
	}
}

func TestGetMasteryBonuses(t *testing.T) {
	tests := []struct {
		level        WeaponMasteryLevel
		expectedDmg  float64
		expectedCrit float64
	}{
		{MasteryNovice, 1.00, 0.00},
		{MasteryApprentice, 1.05, 0.02},
		{MasteryJourneyman, 1.10, 0.04},
		{MasteryExpert, 1.18, 0.06},
		{MasteryMaster, 1.28, 0.08},
		{MasteryGrandmaster, 1.40, 0.10},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			bonuses := GetMasteryBonuses(tt.level)
			if bonuses.DamageMultiplier != tt.expectedDmg {
				t.Errorf("DamageMultiplier = %f, want %f", bonuses.DamageMultiplier, tt.expectedDmg)
			}
			if bonuses.CritChanceBonus != tt.expectedCrit {
				t.Errorf("CritChanceBonus = %f, want %f", bonuses.CritChanceBonus, tt.expectedCrit)
			}
		})
	}
}

func TestWeaponMasterySystem_AwardMasteryXP(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	// Award XP
	system.AwardMasteryXP(entity, "sword", 50)

	data := masteryComp.GetMastery("sword")
	if data.XP != 50 {
		t.Errorf("expected XP 50, got %d", data.XP)
	}
	if masteryComp.TotalMasteryXP != 50 {
		t.Errorf("expected TotalMasteryXP 50, got %d", masteryComp.TotalMasteryXP)
	}
}

func TestWeaponMasterySystem_LevelUp(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	levelUpCalled := false
	var levelUpWeapon string
	var levelUpLevel WeaponMasteryLevel

	system.SetOnMasteryLevelUp(func(e *Entity, wt string, level WeaponMasteryLevel) {
		levelUpCalled = true
		levelUpWeapon = wt
		levelUpLevel = level
	})

	// Award enough XP to reach Apprentice (100)
	system.AwardMasteryXP(entity, "sword", 100)

	if !levelUpCalled {
		t.Error("expected level up callback to be called")
	}
	if levelUpWeapon != "sword" {
		t.Errorf("expected weapon 'sword', got %q", levelUpWeapon)
	}
	if levelUpLevel != MasteryApprentice {
		t.Errorf("expected MasteryApprentice, got %v", levelUpLevel)
	}

	data := masteryComp.GetMastery("sword")
	if data.Level != MasteryApprentice {
		t.Errorf("expected MasteryApprentice, got %v", data.Level)
	}
	if !masteryComp.Dirty {
		t.Error("expected dirty flag to be set after level up")
	}
}

func TestWeaponMasterySystem_OnHit(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	system.OnHit(entity, "sword", 100.0)

	data := masteryComp.GetMastery("sword")
	if data.XP != system.xpPerHit {
		t.Errorf("expected XP %d, got %d", system.xpPerHit, data.XP)
	}
	if data.TotalDamage != 100.0 {
		t.Errorf("expected TotalDamage 100.0, got %f", data.TotalDamage)
	}
}

func TestWeaponMasterySystem_OnKill(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	system.OnKill(entity, "bow")

	data := masteryComp.GetMastery("bow")
	if data.XP != system.xpPerKill {
		t.Errorf("expected XP %d, got %d", system.xpPerKill, data.XP)
	}
	if data.TotalKills != 1 {
		t.Errorf("expected TotalKills 1, got %d", data.TotalKills)
	}
}

func TestWeaponMasterySystem_OnCriticalHit(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	system.OnCriticalHit(entity, "dagger")

	data := masteryComp.GetMastery("dagger")
	if data.XP != system.xpPerCritical {
		t.Errorf("expected XP %d, got %d", system.xpPerCritical, data.XP)
	}
	if data.CriticalHits != 1 {
		t.Errorf("expected CriticalHits 1, got %d", data.CriticalHits)
	}
}

func TestWeaponMasterySystem_GetDamageMultiplier(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)

	// No mastery component - returns 1.0
	mult := system.GetDamageMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("expected 1.0 without mastery component, got %f", mult)
	}

	// Add mastery component but no weapon
	masteryComp := NewWeaponMasteryComponent()
	entity.AddComponent(masteryComp)

	mult = system.GetDamageMultiplier(entity)
	if mult != 1.0 {
		t.Errorf("expected 1.0 without equipped weapon, got %f", mult)
	}

	// Add equipment with weapon
	equipComp := NewEquipmentComponent()
	weapon := &item.Item{
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
	}
	equipComp.Equip(weapon, SlotMainHand)
	entity.AddComponent(equipComp)

	// Set sword mastery to Expert
	masteryComp.Mastery["sword"] = &WeaponMasteryData{
		Level: MasteryExpert,
	}

	mult = system.GetDamageMultiplier(entity)
	expected := GetMasteryBonuses(MasteryExpert).DamageMultiplier
	if mult != expected {
		t.Errorf("expected %f damage multiplier, got %f", expected, mult)
	}
}

func TestWeaponMasterySystem_Update(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)

	// Add components
	masteryComp := NewWeaponMasteryComponent()
	masteryComp.Mastery["sword"] = &WeaponMasteryData{
		Level: MasteryJourneyman,
		XP:    300,
	}
	masteryComp.Dirty = true
	entity.AddComponent(masteryComp)

	equipComp := NewEquipmentComponent()
	equipComp.Equip(&item.Item{
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
	}, SlotMainHand)
	entity.AddComponent(equipComp)

	statsComp := &StatsComponent{
		CritChance: 0.1,
		CritDamage: 1.5,
	}
	entity.AddComponent(statsComp)

	attackComp := &AttackComponent{
		Cooldown: 1.0,
	}
	entity.AddComponent(attackComp)

	entities := []*Entity{entity}

	// Run update cycles until dirty flag processed
	for i := 0; i < system.updateInterval+1; i++ {
		system.Update(entities, 0.016)
	}

	// Check dirty flag cleared
	if masteryComp.Dirty {
		t.Error("expected dirty flag to be cleared after update")
	}

	// Check bonuses were applied
	expectedBonuses := GetMasteryBonuses(MasteryJourneyman)
	if statsComp.CritChance != 0.1+expectedBonuses.CritChanceBonus {
		t.Errorf("expected CritChance %f, got %f",
			0.1+expectedBonuses.CritChanceBonus, statsComp.CritChance)
	}

	// Check bonus tracking
	if masteryComp.BonusesApplied["sword"] != MasteryJourneyman {
		t.Errorf("expected BonusesApplied[sword] = MasteryJourneyman, got %v",
			masteryComp.BonusesApplied["sword"])
	}
}

func TestWeaponMasterySystem_GetMasteryStatusForUI(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)

	// No weapon - returns nil
	status := system.GetMasteryStatusForUI(entity)
	if status != nil {
		t.Error("expected nil status without weapon")
	}

	// Add components
	masteryComp := NewWeaponMasteryComponent()
	masteryComp.Mastery["sword"] = &WeaponMasteryData{
		Level:        MasteryExpert,
		XP:           750,
		TotalKills:   50,
		TotalDamage:  10000.0,
		CriticalHits: 25,
	}
	masteryComp.TotalMasteryXP = 750
	entity.AddComponent(masteryComp)

	equipComp := NewEquipmentComponent()
	equipComp.Equip(&item.Item{
		Type:       item.TypeWeapon,
		WeaponType: item.WeaponSword,
	}, SlotMainHand)
	entity.AddComponent(equipComp)

	status = system.GetMasteryStatusForUI(entity)
	if status == nil {
		t.Fatal("expected non-nil status")
	}

	if status.WeaponType != "sword" {
		t.Errorf("expected WeaponType 'sword', got %q", status.WeaponType)
	}
	if status.Level != MasteryExpert {
		t.Errorf("expected MasteryExpert, got %v", status.Level)
	}
	if status.CurrentXP != 750 {
		t.Errorf("expected CurrentXP 750, got %d", status.CurrentXP)
	}
	if status.TotalKills != 50 {
		t.Errorf("expected TotalKills 50, got %d", status.TotalKills)
	}
	// Use tolerance for float comparison
	if status.DamageBonus < 17.9 || status.DamageBonus > 18.1 { // (1.18 - 1.0) * 100
		t.Errorf("expected DamageBonus ~18.0%%, got %f%%", status.DamageBonus)
	}
}

func TestEnsureWeaponMasteryComponent(t *testing.T) {
	// nil entity returns nil
	if EnsureWeaponMasteryComponent(nil) != nil {
		t.Error("expected nil for nil entity")
	}

	entity := NewEntity(1)

	// First call creates component
	comp1 := EnsureWeaponMasteryComponent(entity)
	if comp1 == nil {
		t.Fatal("expected non-nil component")
	}

	// Second call returns same component
	comp2 := EnsureWeaponMasteryComponent(entity)
	if comp1 != comp2 {
		t.Error("expected same component on repeated call")
	}
}

func TestWeaponMasterySystem_calculateMasteryLevel(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	tests := []struct {
		xp       int
		expected WeaponMasteryLevel
	}{
		{0, MasteryNovice},
		{50, MasteryNovice},
		{99, MasteryNovice},
		{100, MasteryApprentice},
		{299, MasteryApprentice},
		{300, MasteryJourneyman},
		{599, MasteryJourneyman},
		{600, MasteryExpert},
		{999, MasteryExpert},
		{1000, MasteryMaster},
		{1499, MasteryMaster},
		{1500, MasteryGrandmaster},
		{9999, MasteryGrandmaster},
	}

	for _, tt := range tests {
		t.Run(tt.expected.String(), func(t *testing.T) {
			result := system.calculateMasteryLevel(tt.xp)
			if result != tt.expected {
				t.Errorf("calculateMasteryLevel(%d) = %v, want %v", tt.xp, result, tt.expected)
			}
		})
	}
}

func TestWeaponMasterySystem_GetAllMasteryLevels(t *testing.T) {
	world := &World{}
	system := NewWeaponMasterySystem(world)

	entity := NewEntity(1)
	masteryComp := NewWeaponMasteryComponent()
	masteryComp.Mastery["sword"] = &WeaponMasteryData{Level: MasteryExpert, XP: 700}
	masteryComp.Mastery["bow"] = &WeaponMasteryData{Level: MasteryNovice, XP: 50}
	entity.AddComponent(masteryComp)

	allLevels := system.GetAllMasteryLevels(entity)
	if allLevels == nil {
		t.Fatal("expected non-nil map")
	}

	if len(allLevels) != 2 {
		t.Errorf("expected 2 entries, got %d", len(allLevels))
	}

	if allLevels["sword"].Level != MasteryExpert {
		t.Errorf("expected sword at MasteryExpert, got %v", allLevels["sword"].Level)
	}
	if allLevels["bow"].Level != MasteryNovice {
		t.Errorf("expected bow at MasteryNovice, got %v", allLevels["bow"].Level)
	}
}
