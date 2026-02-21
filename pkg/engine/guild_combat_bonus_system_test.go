package engine

import (
	"testing"
)

func TestNewGuildCombatBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("NewGuildCombatBonusSystem returned nil")
	}

	if system.world != world {
		t.Error("system.world not set correctly")
	}

	if system.rng == nil {
		t.Error("system.rng is nil")
	}

	if system.bonusRange != 200.0 {
		t.Errorf("default bonusRange = %v, want 200.0", system.bonusRange)
	}

	if len(system.genreMultipliers) == 0 {
		t.Error("genreMultipliers not initialized")
	}
}

func TestGuildCombatBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	system.SetGenre("cyberpunk")
	if system.genreID != "cyberpunk" {
		t.Errorf("genreID = %v, want cyberpunk", system.genreID)
	}

	// Test genre multiplier
	mult := system.GetGenreMultiplier()
	if mult != 1.2 {
		t.Errorf("cyberpunk multiplier = %v, want 1.2", mult)
	}
}

func TestGuildCombatBonusSystem_SetBonusRange(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	system.SetBonusRange(300.0)
	if system.bonusRange != 300.0 {
		t.Errorf("bonusRange = %v, want 300.0", system.bonusRange)
	}
}

func TestGuildCombatBonusSystem_Update_NoGuildMembers(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	// Create entity without guild
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Should not have bonus component
	_, ok := entity.GetComponent("guildcombatbonus")
	if ok {
		t.Error("entity without guild should not have bonus component")
	}
}

func TestGuildCombatBonusSystem_Update_SingleGuildMember(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	// Create single guild member
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Single member should have no bonus (needs nearby members)
	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		// Component may not exist, which is fine
		return
	}
	comp := compRaw.(*GuildCombatBonusComponent)
	if comp.AttackBonus != 0 {
		t.Errorf("single member attackBonus = %v, want 0", comp.AttackBonus)
	}
}

func TestGuildCombatBonusSystem_Update_TwoNearbyMembers(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 150, Y: 100}) // 50 pixels away
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// Both should have 1-member bonus (5% attack, 3% defense, 2% crit)
	comp1, ok := entity1.GetComponent("guildcombatbonus")
	if !ok {
		t.Fatal("entity1 missing guildcombatbonus component")
	}
	bonus1 := comp1.(*GuildCombatBonusComponent)

	if bonus1.NearbyGuildMemberCount != 1 {
		t.Errorf("entity1 nearbyCount = %v, want 1", bonus1.NearbyGuildMemberCount)
	}
	if bonus1.AttackBonus != 0.05 {
		t.Errorf("entity1 attackBonus = %v, want 0.05", bonus1.AttackBonus)
	}
	if bonus1.DefenseBonus != 0.03 {
		t.Errorf("entity1 defenseBonus = %v, want 0.03", bonus1.DefenseBonus)
	}
	if bonus1.CritBonus != 0.02 {
		t.Errorf("entity1 critBonus = %v, want 0.02", bonus1.CritBonus)
	}
}

func TestGuildCombatBonusSystem_Update_FourNearbyMembers_Capped(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create 5 nearby guild members (should cap at 4+ bonus)
	entities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(100 + i*20), Y: 100})
		entity.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})
		entities[i] = entity
	}

	system.Update(entities, 1.0)

	// First entity should have 4-member bonus (capped at 20% attack, 12% defense, 8% crit)
	comp, ok := entities[0].GetComponent("guildcombatbonus")
	if !ok {
		t.Fatal("entity missing guildcombatbonus component")
	}
	bonus := comp.(*GuildCombatBonusComponent)

	if bonus.NearbyGuildMemberCount != 4 {
		t.Errorf("nearbyCount = %v, want 4", bonus.NearbyGuildMemberCount)
	}
	if bonus.AttackBonus != 0.20 {
		t.Errorf("attackBonus = %v, want 0.20 (capped)", bonus.AttackBonus)
	}
	if bonus.DefenseBonus != 0.12 {
		t.Errorf("defenseBonus = %v, want 0.12 (capped)", bonus.DefenseBonus)
	}
	if bonus.CritBonus != 0.08 {
		t.Errorf("critBonus = %v, want 0.08 (capped)", bonus.CritBonus)
	}
}

func TestGuildCombatBonusSystem_Update_OutOfRange(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two distant guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 500, Y: 0}) // 500 pixels away (beyond 200 range)
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// Should have no bonus (out of range)
	comp, ok := entity1.GetComponent("guildcombatbonus")
	if ok {
		bonus := comp.(*GuildCombatBonusComponent)
		if bonus.AttackBonus != 0 {
			t.Errorf("out-of-range attackBonus = %v, want 0", bonus.AttackBonus)
		}
	}
}

func TestGuildCombatBonusSystem_Update_DifferentGuilds(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby members of different guilds
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild2", Rank: "Member"}) // Different guild

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// Should have no bonus (different guilds)
	comp, ok := entity1.GetComponent("guildcombatbonus")
	if ok {
		bonus := comp.(*GuildCombatBonusComponent)
		if bonus.AttackBonus != 0 {
			t.Errorf("different-guild attackBonus = %v, want 0", bonus.AttackBonus)
		}
	}
}

func TestGuildCombatBonusSystem_RankBonus(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create member and leader nearby
	member := world.CreateEntity()
	member.AddComponent(&PositionComponent{X: 100, Y: 100})
	member.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	leader := world.CreateEntity()
	leader.AddComponent(&PositionComponent{X: 120, Y: 100})
	leader.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Leader"})

	entities := []*Entity{member, leader}
	system.Update(entities, 1.0)

	// Member should have rank bonus from leader nearby
	comp, ok := member.GetComponent("guildcombatbonus")
	if !ok {
		t.Fatal("member missing guildcombatbonus component")
	}
	bonus := comp.(*GuildCombatBonusComponent)

	expectedRankBonus := 0.05 // Leader rank bonus
	if bonus.RankBonus != expectedRankBonus {
		t.Errorf("rankBonus = %v, want %v", bonus.RankBonus, expectedRankBonus)
	}
}

func TestGuildCombatBonusSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre      string
		multiplier float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.1},
		{"horror", 0.7},
		{"cyberpunk", 1.2},
		{"postapoc", 0.85},
		{"unknown", 1.0}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewGuildCombatBonusSystem(world, 12345)
			system.SetGenre(tt.genre)

			mult := system.GetGenreMultiplier()
			if mult != tt.multiplier {
				t.Errorf("genre %s multiplier = %v, want %v", tt.genre, mult, tt.multiplier)
			}
		})
	}
}

func TestGuildCombatBonusSystem_GetDamageMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// GetDamageMultiplier should return 1.0 + attack bonus
	mult := system.GetDamageMultiplier(entity1)
	// Adjust for rank bonus calculation: Member has 0.01 rank bonus
	// RankBonus is 0.01, so additional attack is 0.01*0.5 = 0.005
	// Total: 1.0 + 0.05 + 0.005 = 1.055
	if mult < 1.04 || mult > 1.06 {
		t.Errorf("damageMultiplier = %v, want ~1.05", mult)
	}
}

func TestGuildCombatBonusSystem_GetDefenseMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// GetDefenseMultiplier should return 1.0 + defense bonus
	mult := system.GetDefenseMultiplier(entity1)
	if mult < 1.02 || mult > 1.04 {
		t.Errorf("defenseMultiplier = %v, want ~1.03", mult)
	}
}

func TestGuildCombatBonusSystem_GetCritBonus(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// GetCritBonus should return crit bonus value
	crit := system.GetCritBonus(entity1)
	if crit < 0.02 || crit > 0.03 {
		t.Errorf("critBonus = %v, want ~0.02", crit)
	}
}

func TestGuildCombatBonusSystem_GetHealingBonus(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create two nearby guild members
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity1, entity2}
	system.Update(entities, 1.0)

	// GetHealingBonus should return passive healing rate
	heal := system.GetHealingBonus(entity1)
	if heal < 0.4 || heal > 0.6 {
		t.Errorf("healingBonus = %v, want ~0.5", heal)
	}
}

func TestGuildCombatBonusSystem_HasGuildBonus(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Entity without bonus
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 0, Y: 0})

	if system.HasGuildBonus(entity1) {
		t.Error("entity without guild should not have bonus")
	}

	// Entity with nearby guild member
	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity2.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 120, Y: 100})
	entity3.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})

	entities := []*Entity{entity2, entity3}
	system.Update(entities, 1.0)

	if !system.HasGuildBonus(entity2) {
		t.Error("entity with nearby guild member should have bonus")
	}
}

func TestGuildCombatBonusSystem_GetNearbyGuildMemberCount(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create 3 nearby guild members
	entities := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(100 + i*20), Y: 100})
		entity.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})
		entities[i] = entity
	}

	system.Update(entities, 1.0)

	count := system.GetNearbyGuildMemberCount(entities[0])
	if count != 2 {
		t.Errorf("nearbyGuildMemberCount = %v, want 2", count)
	}
}

func TestGuildCombatBonusSystem_NilEntity(t *testing.T) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)

	// All methods should handle nil safely
	if system.GetDamageMultiplier(nil) != 1.0 {
		t.Error("GetDamageMultiplier(nil) should return 1.0")
	}
	if system.GetDefenseMultiplier(nil) != 1.0 {
		t.Error("GetDefenseMultiplier(nil) should return 1.0")
	}
	if system.GetCritBonus(nil) != 0.0 {
		t.Error("GetCritBonus(nil) should return 0.0")
	}
	if system.GetHealingBonus(nil) != 0.0 {
		t.Error("GetHealingBonus(nil) should return 0.0")
	}
	if system.GetNearbyGuildMemberCount(nil) != 0 {
		t.Error("GetNearbyGuildMemberCount(nil) should return 0")
	}
	if system.HasGuildBonus(nil) {
		t.Error("HasGuildBonus(nil) should return false")
	}
}

func TestGuildCombatBonusComponent_Type(t *testing.T) {
	comp := NewGuildCombatBonusComponent()
	if comp.Type() != "guildcombatbonus" {
		t.Errorf("Type() = %v, want guildcombatbonus", comp.Type())
	}
}

func TestGuildCombatBonusComponent_ClearBonuses(t *testing.T) {
	comp := &GuildCombatBonusComponent{
		NearbyGuildMemberCount: 3,
		AttackBonus:            0.15,
		DefenseBonus:           0.09,
		CritBonus:              0.06,
		HealingBonus:           1.5,
		RankBonus:              0.05,
	}

	comp.ClearBonuses()

	if comp.NearbyGuildMemberCount != 0 {
		t.Errorf("NearbyGuildMemberCount = %v, want 0", comp.NearbyGuildMemberCount)
	}
	if comp.AttackBonus != 0 {
		t.Errorf("AttackBonus = %v, want 0", comp.AttackBonus)
	}
	if comp.DefenseBonus != 0 {
		t.Errorf("DefenseBonus = %v, want 0", comp.DefenseBonus)
	}
	if comp.CritBonus != 0 {
		t.Errorf("CritBonus = %v, want 0", comp.CritBonus)
	}
	if comp.HealingBonus != 0 {
		t.Errorf("HealingBonus = %v, want 0", comp.HealingBonus)
	}
	if comp.RankBonus != 0 {
		t.Errorf("RankBonus = %v, want 0", comp.RankBonus)
	}
}

func TestGuildCombatBonusComponent_GetTotalMultipliers(t *testing.T) {
	comp := &GuildCombatBonusComponent{
		AttackBonus:  0.15,
		DefenseBonus: 0.09,
		CritBonus:    0.06,
		RankBonus:    0.05, // Leader bonus
	}

	// Attack: 1.0 + 0.15 + 0.05*0.5 = 1.175
	attackMult := comp.GetTotalAttackMultiplier()
	expected := 1.175
	if attackMult != expected {
		t.Errorf("GetTotalAttackMultiplier() = %v, want %v", attackMult, expected)
	}

	// Defense: 1.0 + 0.09 + 0.05*0.3 = 1.105
	defenseMult := comp.GetTotalDefenseMultiplier()
	expectedDef := 1.105
	if defenseMult != expectedDef {
		t.Errorf("GetTotalDefenseMultiplier() = %v, want %v", defenseMult, expectedDef)
	}

	// Crit: 0.06 + 0.05*0.05 = 0.0625
	critBonus := comp.GetTotalCritBonus()
	expectedCrit := 0.0625
	if critBonus != expectedCrit {
		t.Errorf("GetTotalCritBonus() = %v, want %v", critBonus, expectedCrit)
	}
}

func TestGuildCombatBonusComponent_HasSignificantBonus(t *testing.T) {
	tests := []struct {
		name     string
		comp     *GuildCombatBonusComponent
		expected bool
	}{
		{
			name:     "no bonus",
			comp:     &GuildCombatBonusComponent{},
			expected: false,
		},
		{
			name:     "tiny attack bonus",
			comp:     &GuildCombatBonusComponent{AttackBonus: 0.005},
			expected: false,
		},
		{
			name:     "significant attack bonus",
			comp:     &GuildCombatBonusComponent{AttackBonus: 0.05},
			expected: true,
		},
		{
			name:     "significant defense bonus",
			comp:     &GuildCombatBonusComponent{DefenseBonus: 0.03},
			expected: true,
		},
		{
			name:     "significant crit bonus",
			comp:     &GuildCombatBonusComponent{CritBonus: 0.02},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.comp.HasSignificantBonus(); got != tt.expected {
				t.Errorf("HasSignificantBonus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Benchmark proximity calculations
func BenchmarkGuildCombatBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewGuildCombatBonusSystem(world, 12345)
	system.SetGenre("fantasy")
	system.updateDelay = 0 // Force update every call

	// Create 20 guild members
	entities := make([]*Entity, 20)
	for i := 0; i < 20; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 30), Y: 100})
		entity.AddComponent(&GuildComponent{GuildID: "guild1", Rank: "Member"})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
