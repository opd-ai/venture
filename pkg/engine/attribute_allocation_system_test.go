// Package engine provides tests for the attribute allocation system.
package engine

import (
	"testing"
)

func TestCoreAttribute_String(t *testing.T) {
	tests := []struct {
		attr CoreAttribute
		want string
	}{
		{AttrStrength, "Strength"},
		{AttrAgility, "Agility"},
		{AttrIntelligence, "Intelligence"},
		{AttrVitality, "Vitality"},
		{AttrEndurance, "Endurance"},
		{AttrLuck, "Luck"},
		{CoreAttribute(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.attr.String(); got != tt.want {
				t.Errorf("CoreAttribute.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoreAttribute_ShortName(t *testing.T) {
	tests := []struct {
		attr CoreAttribute
		want string
	}{
		{AttrStrength, "STR"},
		{AttrAgility, "AGI"},
		{AttrIntelligence, "INT"},
		{AttrVitality, "VIT"},
		{AttrEndurance, "END"},
		{AttrLuck, "LCK"},
		{CoreAttribute(99), "UNK"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.attr.ShortName(); got != tt.want {
				t.Errorf("CoreAttribute.ShortName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttributeFromString(t *testing.T) {
	tests := []struct {
		input   string
		want    CoreAttribute
		wantErr bool
	}{
		{"Strength", AttrStrength, false},
		{"STR", AttrStrength, false},
		{"strength", AttrStrength, false},
		{"str", AttrStrength, false},
		{"Agility", AttrAgility, false},
		{"AGI", AttrAgility, false},
		{"Intelligence", AttrIntelligence, false},
		{"INT", AttrIntelligence, false},
		{"Vitality", AttrVitality, false},
		{"VIT", AttrVitality, false},
		{"Endurance", AttrEndurance, false},
		{"END", AttrEndurance, false},
		{"Luck", AttrLuck, false},
		{"LCK", AttrLuck, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := AttributeFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AttributeFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("AttributeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAttributeAllocationComponent(t *testing.T) {
	comp := NewAttributeAllocationComponent()

	if comp == nil {
		t.Fatal("NewAttributeAllocationComponent returned nil")
	}

	// Check base values are initialized to 10
	for i := 0; i < int(NumCoreAttributes); i++ {
		if comp.BaseAttributes[i] != 10 {
			t.Errorf("BaseAttributes[%d] = %d, want 10", i, comp.BaseAttributes[i])
		}
	}

	if comp.PointsPerLevel != 3 {
		t.Errorf("PointsPerLevel = %d, want 3", comp.PointsPerLevel)
	}

	if comp.AppliedBonuses == nil {
		t.Error("AppliedBonuses map not initialized")
	}
}

func TestAttributeAllocationComponent_Type(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	if got := comp.Type(); got != "attribute_allocation" {
		t.Errorf("Type() = %v, want attribute_allocation", got)
	}
}

func TestAttributeAllocationComponent_GetTotal(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.AllocatedPoints[AttrStrength] = 5
	comp.BonusPoints[AttrStrength] = 3

	// Base (10) + Allocated (5) + Bonus (3) = 18
	if got := comp.GetTotal(AttrStrength); got != 18 {
		t.Errorf("GetTotal(Strength) = %d, want 18", got)
	}

	// Test invalid attribute
	if got := comp.GetTotal(CoreAttribute(-1)); got != 0 {
		t.Errorf("GetTotal(invalid) = %d, want 0", got)
	}
}

func TestAttributeAllocationComponent_CanAllocate(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.UnspentPoints = 5

	// Can allocate within budget
	if !comp.CanAllocate(AttrStrength, 3) {
		t.Error("CanAllocate should return true for 3 points with 5 unspent")
	}

	// Cannot allocate more than available
	if comp.CanAllocate(AttrStrength, 10) {
		t.Error("CanAllocate should return false for 10 points with 5 unspent")
	}

	// Cannot allocate zero or negative
	if comp.CanAllocate(AttrStrength, 0) {
		t.Error("CanAllocate should return false for 0 points")
	}
	if comp.CanAllocate(AttrStrength, -1) {
		t.Error("CanAllocate should return false for negative points")
	}

	// Invalid attribute
	if comp.CanAllocate(CoreAttribute(-1), 1) {
		t.Error("CanAllocate should return false for invalid attribute")
	}
}

func TestAttributeAllocationComponent_TotalAllocatedPoints(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.AllocatedPoints[AttrStrength] = 5
	comp.AllocatedPoints[AttrAgility] = 3
	comp.AllocatedPoints[AttrIntelligence] = 7

	if got := comp.TotalAllocatedPoints(); got != 15 {
		t.Errorf("TotalAllocatedPoints() = %d, want 15", got)
	}
}

func TestAttributeAllocationComponent_Serialize(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.UnspentPoints = 10
	comp.AllocatedPoints[AttrStrength] = 5
	comp.RespecCount = 2

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	newComp := &AttributeAllocationComponent{}
	if err := newComp.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if newComp.UnspentPoints != 10 {
		t.Errorf("UnspentPoints after deserialize = %d, want 10", newComp.UnspentPoints)
	}
	if newComp.AllocatedPoints[AttrStrength] != 5 {
		t.Errorf("AllocatedPoints[Strength] after deserialize = %d, want 5", newComp.AllocatedPoints[AttrStrength])
	}
	if newComp.RespecCount != 2 {
		t.Errorf("RespecCount after deserialize = %d, want 2", newComp.RespecCount)
	}
	if !newComp.Dirty {
		t.Error("Dirty flag should be set after deserialize")
	}
}

func TestNewAttributeAllocationSystem(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	if system == nil {
		t.Fatal("NewAttributeAllocationSystem returned nil")
	}

	if system.effects == nil {
		t.Error("effects not initialized")
	}
}

func TestAttributeAllocationSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	tests := []struct {
		genre    string
		wantMult float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.9},
		{"horror", 1.1},
		{"cyberpunk", 0.95},
		{"postapoc", 1.05},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			if system.genreMultiplier != tt.wantMult {
				t.Errorf("genreMultiplier = %v, want %v", system.genreMultiplier, tt.wantMult)
			}
		})
	}
}

func TestAttributeAllocationSystem_AllocatePoints(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.UnspentPoints = 10
	entity.AddComponent(attrComp)
	entity.AddComponent(&StatsComponent{Attack: 100})

	// Allocate 5 points to strength
	err := system.AllocatePoints(entity, AttrStrength, 5)
	if err != nil {
		t.Fatalf("AllocatePoints() error = %v", err)
	}

	if attrComp.GetAllocated(AttrStrength) != 5 {
		t.Errorf("GetAllocated(Strength) = %d, want 5", attrComp.GetAllocated(AttrStrength))
	}
	if attrComp.UnspentPoints != 5 {
		t.Errorf("UnspentPoints = %d, want 5", attrComp.UnspentPoints)
	}
	if !attrComp.Dirty {
		t.Error("Dirty flag should be set after allocation")
	}
}

func TestAttributeAllocationSystem_AllocatePoints_InsufficientPoints(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.UnspentPoints = 3
	entity.AddComponent(attrComp)

	err := system.AllocatePoints(entity, AttrStrength, 10)
	if err == nil {
		t.Error("AllocatePoints should fail with insufficient points")
	}
}

func TestAttributeAllocationSystem_AwardPoints(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	entity.AddComponent(attrComp)

	err := system.AwardPoints(entity, 5)
	if err != nil {
		t.Fatalf("AwardPoints() error = %v", err)
	}

	if attrComp.UnspentPoints != 5 {
		t.Errorf("UnspentPoints = %d, want 5", attrComp.UnspentPoints)
	}
	if attrComp.TotalPointsEarned != 5 {
		t.Errorf("TotalPointsEarned = %d, want 5", attrComp.TotalPointsEarned)
	}
}

func TestAttributeAllocationSystem_Respec(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.AllocatedPoints[AttrStrength] = 10
	attrComp.AllocatedPoints[AttrAgility] = 5
	attrComp.UnspentPoints = 0
	entity.AddComponent(attrComp)
	entity.AddComponent(&InventoryComponent{Gold: 1000})
	entity.AddComponent(&StatsComponent{Attack: 100})

	// Respec with gold cost
	err := system.Respec(entity, 500)
	if err != nil {
		t.Fatalf("Respec() error = %v", err)
	}

	// All points should be returned
	if attrComp.UnspentPoints != 15 {
		t.Errorf("UnspentPoints after respec = %d, want 15", attrComp.UnspentPoints)
	}
	if attrComp.GetAllocated(AttrStrength) != 0 {
		t.Errorf("Allocated strength after respec = %d, want 0", attrComp.GetAllocated(AttrStrength))
	}
	if attrComp.RespecCount != 1 {
		t.Errorf("RespecCount = %d, want 1", attrComp.RespecCount)
	}

	// Gold should be deducted
	inv := entity.GetComponentFast(&InventoryComponent{}).(*InventoryComponent)
	if inv.Gold != 500 {
		t.Errorf("Gold after respec = %d, want 500", inv.Gold)
	}
}

func TestAttributeAllocationSystem_Respec_InsufficientGold(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.AllocatedPoints[AttrStrength] = 10
	entity.AddComponent(attrComp)
	entity.AddComponent(&InventoryComponent{Gold: 100})

	err := system.Respec(entity, 500)
	if err == nil {
		t.Error("Respec should fail with insufficient gold")
	}
}

func TestAttributeAllocationSystem_GetRespecCost(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	entity.AddComponent(attrComp)

	// First respec: 500 * 2^0 = 500
	if cost := system.GetRespecCost(entity); cost != 500 {
		t.Errorf("First respec cost = %d, want 500", cost)
	}

	attrComp.RespecCount = 1
	// Second respec: 500 * 2^1 = 1000
	if cost := system.GetRespecCost(entity); cost != 1000 {
		t.Errorf("Second respec cost = %d, want 1000", cost)
	}

	attrComp.RespecCount = 4
	// Fifth respec: 500 * 2^4 = 8000
	if cost := system.GetRespecCost(entity); cost != 8000 {
		t.Errorf("Fifth respec cost = %d, want 8000", cost)
	}

	// Capped at 16x
	attrComp.RespecCount = 10
	if cost := system.GetRespecCost(entity); cost != 8000 {
		t.Errorf("Capped respec cost = %d, want 8000", cost)
	}
}

func TestAttributeAllocationSystem_Update_AppliesBonuses(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.AllocatedPoints[AttrStrength] = 10 // 10 extra STR
	attrComp.Dirty = true
	entity.AddComponent(attrComp)

	baseAttack := 100.0
	stats := &StatsComponent{
		Attack:      baseAttack,
		Defense:     50,
		MagicPower:  20,
		Evasion:     0.0,
		BlockChance: 0.0,
		CritChance:  0.0,
	}
	entity.AddComponent(stats)
	entity.AddComponent(&HealthComponent{Current: 500, Max: 500})
	entity.AddComponent(&ManaComponent{Current: 100, Max: 100})

	// Force update interval
	system.timeSinceUpdate = system.updateInterval
	system.Update([]*Entity{entity}, 0.1)

	// STR: base 10 + allocated 10 = 20 total
	// Attack bonus: 20 * 2.0 = 40
	expectedAttack := baseAttack + 40.0
	if stats.Attack != expectedAttack {
		t.Errorf("Attack after bonus = %v, want %v", stats.Attack, expectedAttack)
	}

	// Dirty flag should be cleared
	if attrComp.Dirty {
		t.Error("Dirty flag should be cleared after update")
	}
}

func TestAttributeAllocationSystem_Update_SkipsClean(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	attrComp.AllocatedPoints[AttrStrength] = 10
	attrComp.Dirty = false // Not dirty
	entity.AddComponent(attrComp)

	stats := &StatsComponent{Attack: 100}
	entity.AddComponent(stats)

	system.timeSinceUpdate = system.updateInterval
	system.Update([]*Entity{entity}, 0.1)

	// Stats should be unchanged since component wasn't dirty
	if stats.Attack != 100 {
		t.Errorf("Attack should be unchanged = %v, want 100", stats.Attack)
	}
}

func TestAttributeAllocationSystem_SetBonusPoints(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	attrComp := NewAttributeAllocationComponent()
	entity.AddComponent(attrComp)

	err := system.SetBonusPoints(entity, AttrStrength, 5)
	if err != nil {
		t.Fatalf("SetBonusPoints() error = %v", err)
	}

	if attrComp.GetBonus(AttrStrength) != 5 {
		t.Errorf("GetBonus(Strength) = %d, want 5", attrComp.GetBonus(AttrStrength))
	}
	if !attrComp.Dirty {
		t.Error("Dirty flag should be set after bonus change")
	}
}

func TestAttributeAllocationComponent_GetAttributeSummary(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.UnspentPoints = 7
	comp.AllocatedPoints[AttrStrength] = 5

	summary := comp.GetAttributeSummary()
	// Base 10 + allocated 5 for STR
	expected := "STR:15 AGI:10 INT:10 VIT:10 END:10 LCK:10 (Unspent:7)"
	if summary != expected {
		t.Errorf("GetAttributeSummary() = %v, want %v", summary, expected)
	}
}

func TestAttributeAllocationComponent_Copy(t *testing.T) {
	comp := NewAttributeAllocationComponent()
	comp.UnspentPoints = 10
	comp.AllocatedPoints[AttrStrength] = 5
	comp.RespecCount = 2

	copyComp := comp.Copy()

	if copyComp.UnspentPoints != 10 {
		t.Errorf("Copy UnspentPoints = %d, want 10", copyComp.UnspentPoints)
	}
	if copyComp.AllocatedPoints[AttrStrength] != 5 {
		t.Errorf("Copy AllocatedPoints[Strength] = %d, want 5", copyComp.AllocatedPoints[AttrStrength])
	}

	// Modify original, copy should be unaffected
	comp.AllocatedPoints[AttrStrength] = 99
	if copyComp.AllocatedPoints[AttrStrength] != 5 {
		t.Error("Copy should be independent of original")
	}
}

func TestDefaultAttributeEffects(t *testing.T) {
	effects := DefaultAttributeEffects()

	if effects == nil {
		t.Fatal("DefaultAttributeEffects returned nil")
	}

	// Verify some key values
	if effects.AttackBonusPerStr != 2.0 {
		t.Errorf("AttackBonusPerStr = %v, want 2.0", effects.AttackBonusPerStr)
	}
	if effects.MaxHealthPerVit != 15.0 {
		t.Errorf("MaxHealthPerVit = %v, want 15.0", effects.MaxHealthPerVit)
	}
	if effects.CritChancePerLuck != 0.4 {
		t.Errorf("CritChancePerLuck = %v, want 0.4", effects.CritChancePerLuck)
	}
}

func TestAttributeAllocationSystem_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	entity := NewEntity()
	// No attribute_allocation component

	err := system.AllocatePoints(entity, AttrStrength, 1)
	if err == nil {
		t.Error("AllocatePoints should fail without component")
	}

	err = system.AwardPoints(entity, 1)
	if err == nil {
		t.Error("AwardPoints should fail without component")
	}

	if got := system.GetUnspentPoints(entity); got != 0 {
		t.Errorf("GetUnspentPoints without component = %d, want 0", got)
	}
}

func BenchmarkAttributeAllocationSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewAttributeAllocationSystem(world, 12345)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity()
		attrComp := NewAttributeAllocationComponent()
		attrComp.Dirty = true
		attrComp.AllocatedPoints[AttrStrength] = 10
		attrComp.AllocatedPoints[AttrVitality] = 10
		entity.AddComponent(attrComp)
		entity.AddComponent(&StatsComponent{
			Attack:     100,
			Defense:    50,
			MagicPower: 20,
		})
		entity.AddComponent(&HealthComponent{Current: 500, Max: 500})
		entity.AddComponent(&ManaComponent{Current: 100, Max: 100})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark all dirty
		for _, e := range entities {
			comp, _ := e.GetComponent("attribute_allocation")
			comp.(*AttributeAllocationComponent).Dirty = true
		}
		system.timeSinceUpdate = system.updateInterval
		system.Update(entities, 0.016)
	}
}
