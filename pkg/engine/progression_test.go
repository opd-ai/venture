package engine

import (
	"testing"
)

// TestExperienceComponent tests the ExperienceComponent functionality.
func TestExperienceComponent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *ExperienceComponent
		xpToAdd   int
		wantLevel bool
	}{
		{
			name: "new component starts at level 1",
			setup: func() *ExperienceComponent {
				return NewExperienceComponent()
			},
			xpToAdd:   0,
			wantLevel: false,
		},
		{
			name: "adding XP below threshold doesn't level up",
			setup: func() *ExperienceComponent {
				return NewExperienceComponent()
			},
			xpToAdd:   50,
			wantLevel: false,
		},
		{
			name: "adding XP at threshold triggers level up",
			setup: func() *ExperienceComponent {
				return NewExperienceComponent()
			},
			xpToAdd:   100,
			wantLevel: true,
		},
		{
			name: "adding XP above threshold triggers level up",
			setup: func() *ExperienceComponent {
				return NewExperienceComponent()
			},
			xpToAdd:   150,
			wantLevel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := tt.setup()

			// Check initial state
			if exp.Level != 1 {
				t.Errorf("initial level = %d, want 1", exp.Level)
			}

			// Add XP
			gotLevel := exp.AddXP(tt.xpToAdd)

			if gotLevel != tt.wantLevel {
				t.Errorf("AddXP() = %v, want %v", gotLevel, tt.wantLevel)
			}

			// Check current XP
			if exp.CurrentXP != tt.xpToAdd {
				t.Errorf("CurrentXP = %d, want %d", exp.CurrentXP, tt.xpToAdd)
			}

			// Check total XP
			if exp.TotalXP != tt.xpToAdd {
				t.Errorf("TotalXP = %d, want %d", exp.TotalXP, tt.xpToAdd)
			}
		})
	}
}

// TestExperienceProgressToNextLevel tests the progress calculation.
func TestExperienceProgressToNextLevel(t *testing.T) {
	exp := NewExperienceComponent()
	exp.RequiredXP = 100

	tests := []struct {
		currentXP int
		want      float64
	}{
		{0, 0.0},
		{25, 0.25},
		{50, 0.50},
		{75, 0.75},
		{100, 1.0},
		{150, 1.0}, // Should cap at 1.0
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			exp.CurrentXP = tt.currentXP
			got := exp.ProgressToNextLevel()

			if got != tt.want {
				t.Errorf("ProgressToNextLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLevelScalingComponent tests the level scaling calculations.
func TestLevelScalingComponent(t *testing.T) {
	scaling := NewLevelScalingComponent()

	tests := []struct {
		name       string
		level      int
		wantHealth float64
		wantAttack float64
	}{
		{
			name:       "level 1",
			level:      1,
			wantHealth: 100.0, // Base health
			wantAttack: 10.0,  // Base attack
		},
		{
			name:       "level 2",
			level:      2,
			wantHealth: 110.0, // Base + 1 * perLevel
			wantAttack: 12.0,
		},
		{
			name:       "level 5",
			level:      5,
			wantHealth: 140.0, // Base + 4 * perLevel
			wantAttack: 18.0,
		},
		{
			name:       "level 10",
			level:      10,
			wantHealth: 190.0, // Base + 9 * perLevel
			wantAttack: 28.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHealth := scaling.CalculateHealthForLevel(tt.level)
			if gotHealth != tt.wantHealth {
				t.Errorf("CalculateHealthForLevel(%d) = %v, want %v",
					tt.level, gotHealth, tt.wantHealth)
			}

			gotAttack := scaling.CalculateAttackForLevel(tt.level)
			if gotAttack != tt.wantAttack {
				t.Errorf("CalculateAttackForLevel(%d) = %v, want %v",
					tt.level, gotAttack, tt.wantAttack)
			}
		})
	}
}

// TestProgressionSystemAwardXP tests awarding XP to entities.
func TestProgressionSystemAwardXP(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	// Create test entity
	entity := world.CreateEntity()
	exp := NewExperienceComponent()
	entity.AddComponent(exp)

	world.Update(0)

	// Award XP that doesn't level up
	err := ps.AwardXP(entity, 50)
	if err != nil {
		t.Errorf("AwardXP() error = %v", err)
	}

	if exp.CurrentXP != 50 {
		t.Errorf("CurrentXP = %d, want 50", exp.CurrentXP)
	}
	if exp.Level != 1 {
		t.Errorf("Level = %d, want 1", exp.Level)
	}

	// Award XP that causes level up
	err = ps.AwardXP(entity, 50)
	if err != nil {
		t.Errorf("AwardXP() error = %v", err)
	}

	if exp.Level != 2 {
		t.Errorf("Level = %d, want 2", exp.Level)
	}
	if exp.CurrentXP != 0 {
		t.Errorf("CurrentXP = %d, want 0 (rolled over)", exp.CurrentXP)
	}
}

// TestProgressionSystemAwardXPWithScaling tests XP award with stat scaling.
func TestProgressionSystemAwardXPWithScaling(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	// Create test entity with all components
	entity := world.CreateEntity()
	entity.AddComponent(NewExperienceComponent())
	entity.AddComponent(NewLevelScalingComponent())
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(NewStatsComponent())

	world.Update(0)

	// Get initial stats
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	statsComp, _ := entity.GetComponent("stats")
	stats := statsComp.(*StatsComponent)

	initialMaxHealth := health.Max
	initialAttack := stats.Attack

	// Award enough XP to level up
	err := ps.AwardXP(entity, 100)
	if err != nil {
		t.Errorf("AwardXP() error = %v", err)
	}

	expComp, _ := entity.GetComponent("experience")
	exp := expComp.(*ExperienceComponent)
	if exp.Level != 2 {
		t.Errorf("Level = %d, want 2", exp.Level)
	}

	// Check that stats increased
	if health.Max <= initialMaxHealth {
		t.Errorf("Max health didn't increase: %v -> %v", initialMaxHealth, health.Max)
	}
	if stats.Attack <= initialAttack {
		t.Errorf("Attack didn't increase: %v -> %v", initialAttack, stats.Attack)
	}

	// Check skill point was awarded
	if exp.SkillPoints != 1 {
		t.Errorf("SkillPoints = %d, want 1", exp.SkillPoints)
	}
}

// TestProgressionSystemLevelUpCallback tests level up callbacks.
func TestProgressionSystemLevelUpCallback(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	// Track callback invocations
	callbackCount := 0
	var callbackLevel int

	ps.AddLevelUpCallback(func(entity *Entity, newLevel int) {
		callbackCount++
		callbackLevel = newLevel
	})

	// Create test entity
	entity := world.CreateEntity()
	entity.AddComponent(NewExperienceComponent())
	world.Update(0)

	// Level up
	err := ps.AwardXP(entity, 100)
	if err != nil {
		t.Errorf("AwardXP() error = %v", err)
	}

	if callbackCount != 1 {
		t.Errorf("callback count = %d, want 1", callbackCount)
	}
	if callbackLevel != 2 {
		t.Errorf("callback level = %d, want 2", callbackLevel)
	}
}

// TestProgressionSystemMultipleLevelUps tests gaining multiple levels at once.
func TestProgressionSystemMultipleLevelUps(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	// Create test entity
	entity := world.CreateEntity()
	entity.AddComponent(NewExperienceComponent())
	world.Update(0)

	// Award massive XP (enough for multiple levels)
	// Level 1->2: 100 XP
	// Level 2->3: ~173 XP (using default curve)
	// Total: ~273 XP needed for level 3
	err := ps.AwardXP(entity, 500)
	if err != nil {
		t.Errorf("AwardXP() error = %v", err)
	}

	expComp, _ := entity.GetComponent("experience")
	exp := expComp.(*ExperienceComponent)
	if exp.Level < 3 {
		t.Errorf("Level = %d, want >= 3", exp.Level)
	}
}

// TestProgressionSystemXPCurves tests different XP curves.
func TestProgressionSystemXPCurves(t *testing.T) {
	tests := []struct {
		name  string
		curve XPCurveFunc
		level int
		want  int
	}{
		{"default level 1", DefaultXPCurve, 1, 100},
		{"default level 2", DefaultXPCurve, 2, 282},  // 100 * 2^1.5
		{"default level 5", DefaultXPCurve, 5, 1118}, // 100 * 5^1.5
		{"linear level 1", LinearXPCurve, 1, 100},
		{"linear level 5", LinearXPCurve, 5, 500},
		{"exponential level 1", ExponentialXPCurve, 1, 100},
		{"exponential level 5", ExponentialXPCurve, 5, 2500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.curve(tt.level)
			if got != tt.want {
				t.Errorf("curve(%d) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

// TestProgressionSystemCalculateXPReward tests XP reward calculation.
func TestProgressionSystemCalculateXPReward(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	tests := []struct {
		name      string
		level     int
		wantMinXP int
	}{
		{"level 1 enemy", 1, 10},
		{"level 5 enemy", 5, 50},
		{"level 10 enemy", 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create enemy at specific level
			enemy := world.CreateEntity()
			exp := NewExperienceComponent()
			exp.Level = tt.level
			enemy.AddComponent(exp)
			world.Update(0)

			got := ps.CalculateXPReward(enemy)
			if got < tt.wantMinXP {
				t.Errorf("CalculateXPReward() = %d, want >= %d", got, tt.wantMinXP)
			}
		})
	}
}

// TestProgressionSystemInitializeEntityAtLevel tests spawning entities at specific levels.
func TestProgressionSystemInitializeEntityAtLevel(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	tests := []struct {
		name  string
		level int
	}{
		{"level 1", 1},
		{"level 5", 5},
		{"level 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(NewLevelScalingComponent())
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
			entity.AddComponent(NewStatsComponent())
			world.Update(0)

			err := ps.InitializeEntityAtLevel(entity, tt.level)
			if err != nil {
				t.Errorf("InitializeEntityAtLevel() error = %v", err)
			}

			expComp, _ := entity.GetComponent("experience")
			exp := expComp.(*ExperienceComponent)
			if exp.Level != tt.level {
				t.Errorf("Level = %d, want %d", exp.Level, tt.level)
			}

			// Check that stats are scaled for level
			healthComp, _ := entity.GetComponent("health")
			health := healthComp.(*HealthComponent)
			if tt.level > 1 && health.Max <= 100 {
				t.Errorf("Health not scaled for level %d: %v", tt.level, health.Max)
			}
		})
	}
}

// TestProgressionSystemSpendSkillPoint tests spending skill points.
func TestProgressionSystemSpendSkillPoint(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	entity := world.CreateEntity()
	exp := NewExperienceComponent()
	exp.SkillPoints = 2
	entity.AddComponent(exp)
	world.Update(0)

	// Spend first point
	err := ps.SpendSkillPoint(entity)
	if err != nil {
		t.Errorf("SpendSkillPoint() error = %v", err)
	}
	if exp.SkillPoints != 1 {
		t.Errorf("SkillPoints = %d, want 1", exp.SkillPoints)
	}

	// Spend second point
	err = ps.SpendSkillPoint(entity)
	if err != nil {
		t.Errorf("SpendSkillPoint() error = %v", err)
	}
	if exp.SkillPoints != 0 {
		t.Errorf("SkillPoints = %d, want 0", exp.SkillPoints)
	}

	// Try to spend when no points available
	err = ps.SpendSkillPoint(entity)
	if err == nil {
		t.Error("SpendSkillPoint() expected error when no points available")
	}
}

// TestProgressionSystemErrorCases tests error handling.
func TestProgressionSystemErrorCases(t *testing.T) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	t.Run("award XP to nil entity", func(t *testing.T) {
		err := ps.AwardXP(nil, 100)
		if err == nil {
			t.Error("expected error for nil entity")
		}
	})

	t.Run("award negative XP", func(t *testing.T) {
		entity := world.CreateEntity()
		entity.AddComponent(NewExperienceComponent())
		world.Update(0)

		err := ps.AwardXP(entity, -10)
		if err == nil {
			t.Error("expected error for negative XP")
		}
	})

	t.Run("award XP to entity without experience component", func(t *testing.T) {
		entity := world.CreateEntity()
		world.Update(0)

		err := ps.AwardXP(entity, 100)
		if err == nil {
			t.Error("expected error for entity without experience component")
		}
	})
}

// BenchmarkProgressionSystemAwardXP benchmarks XP awarding.
func BenchmarkProgressionSystemAwardXP(b *testing.B) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(NewExperienceComponent())
	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.AwardXP(entity, 10)
	}
}

// BenchmarkProgressionSystemLevelUp benchmarks leveling up.
func BenchmarkProgressionSystemLevelUp(b *testing.B) {
	world := NewWorld()
	ps := NewProgressionSystem(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewExperienceComponent())
		entity.AddComponent(NewLevelScalingComponent())
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(NewStatsComponent())
		world.Update(0)

		ps.AwardXP(entity, 100)
	}
}
