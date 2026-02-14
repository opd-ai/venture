package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestNewTimeOfDayCompanionBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTimeOfDayCompanionBonusSystem returned nil")
	}

	if system.world != world {
		t.Error("System world not set correctly")
	}

	if len(system.timeBonuses) == 0 {
		t.Error("Time bonuses not initialized")
	}

	if len(system.genreMultipliers) == 0 {
		t.Error("Genre multipliers not initialized")
	}
}

func TestTimeOfDayCompanionBonusComponent_Type(t *testing.T) {
	comp := &TimeOfDayCompanionBonusComponent{
		AttackBonus:  1.2,
		DefenseBonus: 1.1,
		SpeedBonus:   1.15,
	}

	if comp.Type() != "timeofday_companion_bonus" {
		t.Errorf("Expected type 'timeofday_companion_bonus', got '%s'", comp.Type())
	}
}

func TestTimeOfDayCompanionBonusSystem_SetLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 67890)

	system.SetLightingSystem(lightingSystem)

	if system.lightingSystem != lightingSystem {
		t.Error("Lighting system not set correctly")
	}
}

func TestTimeOfDayCompanionBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genre != genre {
			t.Errorf("Genre not set correctly: expected %s, got %s", genre, system.genre)
		}
	}
}

func TestTimeOfDayCompanionBonusSystem_CalculateTimeBonus(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	tests := []struct {
		name         string
		compType     CompanionType
		timeOfDay    palette.TimeOfDay
		expectBonus  bool
		expectAttack float64 // Expected attack multiplier (approximate)
	}{
		{"undead_night", CompanionTypeUndead, palette.TimeOfDayNight, true, 1.35},
		{"undead_day", CompanionTypeUndead, palette.TimeOfDayDay, true, 0.75},
		{"pet_day", CompanionTypePet, palette.TimeOfDayDay, true, 1.15},
		{"pet_night", CompanionTypePet, palette.TimeOfDayNight, true, 0.85},
		{"spirit_dusk", CompanionTypeSpirit, palette.TimeOfDayDusk, true, 1.25},
		{"robot_any", CompanionTypeRobot, palette.TimeOfDayDay, true, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := system.calculateTimeBonus(tt.compType, tt.timeOfDay)

			if tt.expectBonus && bonus == nil {
				t.Error("Expected bonus but got nil")
				return
			}

			if !tt.expectBonus && bonus != nil {
				t.Error("Expected no bonus but got one")
				return
			}

			if bonus != nil {
				// Allow small tolerance for floating point
				if bonus.AttackBonus < tt.expectAttack-0.01 || bonus.AttackBonus > tt.expectAttack+0.01 {
					t.Errorf("Expected attack bonus ~%.2f, got %.2f", tt.expectAttack, bonus.AttackBonus)
				}
			}
		})
	}
}

func TestTimeOfDayCompanionBonusSystem_GenreMultiplier(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		genre            string
		expectedStrength float64 // Relative strength (1.0 = normal)
	}{
		{"fantasy", 1.0},
		{"scifi", 0.5},
		{"horror", 1.4},
		{"cyberpunk", 0.4},
		{"postapoc", 1.2},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system := NewTimeOfDayCompanionBonusSystem(world, 12345)
			system.SetGenre(tt.genre)

			// Undead at night has base +35% attack in fantasy
			// Other genres scale this bonus
			bonus := system.calculateTimeBonus(CompanionTypeUndead, palette.TimeOfDayNight)
			if bonus == nil {
				t.Fatal("Expected bonus, got nil")
			}

			// Calculate expected attack based on genre multiplier
			// Base undead night: 1.35 → deviation from 1.0 is 0.35
			// Scaled: 1.0 + 0.35 * genreMultiplier
			expectedAttack := 1.0 + 0.35*tt.expectedStrength
			tolerance := 0.02

			if bonus.AttackBonus < expectedAttack-tolerance || bonus.AttackBonus > expectedAttack+tolerance {
				t.Errorf("Genre %s: expected attack ~%.2f, got %.2f",
					tt.genre, expectedAttack, bonus.AttackBonus)
			}
		})
	}
}

func TestTimeOfDayCompanionBonusSystem_Update_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	// Create a companion entity
	entity := world.CreateEntity()
	entity.AddComponent(&CompanionComponent{
		CompanionType: CompanionTypeUndead,
	})
	entity.AddComponent(&CompanionStatsComponent{
		Attack:  10.0,
		Defense: 10.0,
		Speed:   10.0,
	})

	entities := []*Entity{entity}

	// Should not panic without lighting system
	system.Update(entities, 1.0)

	// Stats should be unchanged
	stats := entity.GetComponentCached("companionstats").(*CompanionStatsComponent)
	if stats.Attack != 10.0 {
		t.Error("Stats changed without lighting system")
	}
}

func TestTimeOfDayCompanionBonusSystem_ApplyBonusToStats(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 100.0,
		Speed:   100.0,
	})

	bonus := &TimeOfDayCompanionBonusComponent{
		AttackBonus:  1.20,
		DefenseBonus: 1.10,
		SpeedBonus:   1.15,
	}

	system.applyBonusToStats(entity, bonus)

	stats := entity.GetComponentCached("companionstats").(*CompanionStatsComponent)
	if stats.Attack != 120.0 {
		t.Errorf("Expected attack 120.0, got %.2f", stats.Attack)
	}
	if stats.Defense != 110.0 {
		t.Errorf("Expected defense 110.0, got %.2f", stats.Defense)
	}
	if stats.Speed != 115.0 {
		t.Errorf("Expected speed 115.0, got %.2f", stats.Speed)
	}
}

func TestTimeOfDayCompanionBonusSystem_ReverseStatBonus(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CompanionStatsComponent{
		Attack:  120.0,
		Defense: 110.0,
		Speed:   115.0,
	})

	bonus := &TimeOfDayCompanionBonusComponent{
		AttackBonus:  1.20,
		DefenseBonus: 1.10,
		SpeedBonus:   1.15,
	}

	system.reverseStatBonus(entity, bonus)

	stats := entity.GetComponentCached("companionstats").(*CompanionStatsComponent)
	if stats.Attack < 99.9 || stats.Attack > 100.1 {
		t.Errorf("Expected attack ~100.0, got %.2f", stats.Attack)
	}
	if stats.Defense < 99.9 || stats.Defense > 100.1 {
		t.Errorf("Expected defense ~100.0, got %.2f", stats.Defense)
	}
	if stats.Speed < 99.9 || stats.Speed > 100.1 {
		t.Errorf("Expected speed ~100.0, got %.2f", stats.Speed)
	}
}

func TestTimeOfDayCompanionBonusSystem_CompanionTypeName(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	tests := []struct {
		compType CompanionType
		expected string
	}{
		{CompanionTypePet, "Pet"},
		{CompanionTypeSummon, "Summon"},
		{CompanionTypeHireling, "Hireling"},
		{CompanionTypeElemental, "Elemental"},
		{CompanionTypeUndead, "Undead"},
		{CompanionTypeRobot, "Robot"},
		{CompanionTypeSpirit, "Spirit"},
		{CompanionTypeInsect, "Insect"},
		{CompanionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := system.companionTypeName(tt.compType)
			if name != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, name)
			}
		})
	}
}

func TestTimeOfDayCompanionBonusSystem_HasActiveBonus(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	if system.HasActiveBonus(123) {
		t.Error("Expected no bonus for non-existent entity")
	}

	system.bonusCache[123] = &TimeOfDayCompanionBonusComponent{}
	if !system.HasActiveBonus(123) {
		t.Error("Expected bonus for cached entity")
	}
}

func TestTimeOfDayCompanionBonusSystem_GetBonusCount(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	if system.GetBonusCount() != 0 {
		t.Error("Expected 0 bonuses initially")
	}

	system.bonusCache[1] = &TimeOfDayCompanionBonusComponent{}
	system.bonusCache[2] = &TimeOfDayCompanionBonusComponent{}

	if system.GetBonusCount() != 2 {
		t.Errorf("Expected 2 bonuses, got %d", system.GetBonusCount())
	}
}

func TestTimeOfDayCompanionBonusSystem_RemoveTimeBonus(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&CompanionStatsComponent{
		Attack:  120.0,
		Defense: 110.0,
		Speed:   115.0,
	})

	bonus := &TimeOfDayCompanionBonusComponent{
		AttackBonus:  1.20,
		DefenseBonus: 1.10,
		SpeedBonus:   1.15,
	}
	entity.AddComponent(bonus)
	system.bonusCache[entity.ID] = bonus

	system.removeTimeBonus(entity)

	if _, ok := entity.GetComponent("timeofday_companion_bonus"); ok {
		t.Error("Bonus component not removed")
	}

	if system.HasActiveBonus(entity.ID) {
		t.Error("Bonus not removed from cache")
	}

	// Stats should be reversed
	stats := entity.GetComponentCached("companionstats").(*CompanionStatsComponent)
	if stats.Attack < 99.9 || stats.Attack > 100.1 {
		t.Errorf("Expected attack ~100.0 after removal, got %.2f", stats.Attack)
	}
}

func TestTimeOfDayCompanionBonusSystem_AllTimePeriodsHaveBonuses(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	timePeriods := []palette.TimeOfDay{
		palette.TimeOfDayDawn,
		palette.TimeOfDayDay,
		palette.TimeOfDayDusk,
		palette.TimeOfDayNight,
	}

	companionTypes := []CompanionType{
		CompanionTypePet,
		CompanionTypeElemental,
		CompanionTypeUndead,
		CompanionTypeRobot,
		CompanionTypeSpirit,
		CompanionTypeInsect,
	}

	for _, timeOfDay := range timePeriods {
		for _, compType := range companionTypes {
			bonus := system.calculateTimeBonus(compType, timeOfDay)
			if bonus == nil {
				t.Errorf("Missing bonus for %s at %s",
					system.companionTypeName(compType), timeOfDay.String())
			}
		}
	}
}

func TestTimeOfDayCompanionBonusSystem_GetCurrentTimeOfDay_NoLightingSystem(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	// Without lighting system, should return default (Day)
	timeOfDay := system.GetCurrentTimeOfDay()
	if timeOfDay != palette.TimeOfDayDay {
		t.Errorf("Expected Day without lighting system, got %s", timeOfDay.String())
	}
}

func TestTimeOfDayCompanionBonusSystem_GetLastTimeOfDay(t *testing.T) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)

	// Default should be Day
	if system.GetLastTimeOfDay() != palette.TimeOfDayDay {
		t.Errorf("Expected Day as default last time, got %s", system.GetLastTimeOfDay().String())
	}
}

func BenchmarkTimeOfDayCompanionBonusSystem_CalculateBonus(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)
	system.SetGenre("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.calculateTimeBonus(CompanionTypeUndead, palette.TimeOfDayNight)
	}
}

func BenchmarkTimeOfDayCompanionBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTimeOfDayCompanionBonusSystem(world, 12345)
	lightingSystem := NewTimeOfDayLightingSystem(world, 67890)
	system.SetLightingSystem(lightingSystem)
	system.SetGenre("fantasy")

	// Create 50 companion entities
	entities := make([]*Entity, 50)
	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&CompanionComponent{
			CompanionType: CompanionType(i % 6),
		})
		entity.AddComponent(&CompanionStatsComponent{
			Attack:  10.0,
			Defense: 10.0,
			Speed:   10.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceCheck = 2.0 // Force update
		system.lastTimeOfDay = palette.TimeOfDayDawn
		system.Update(entities, 0.016)
	}
}
