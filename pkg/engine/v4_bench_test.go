package engine

import (
	"testing"
)

// BenchmarkVehicleCombatSystem benchmarks vehicle combat processing.
func BenchmarkVehicleCombatSystem(b *testing.B) {
	world := NewWorld()
	system := NewVehicleCombatSystem(world)

	// Create 100 vehicles with combat components
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&VehicleComponent{
			VehicleType:  VehicleMech,
			Speed:        50.0,
			MaxSpeed:     100.0,
			Acceleration: 10.0,
		})
		entity.AddComponent(&VehicleCombatComponent{
			RammingDamage:          25.0,
			MinRammingSpeed:        30.0,
			WeaponMounted:          true,
			WeaponDamage:           50.0,
			WeaponRange:            100.0,
			WeaponCooldown:         2.0,
			CurrentWeaponCooldown:  0.0,
			CurrentRammingCooldown: 0.0,
			ArmorRating:            50.0,
		})
		entity.AddComponent(&HealthComponent{Current: 200, Max: 200})
	}

	world.Update(0) // Process pending entities

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entities := world.GetEntitiesWith("vehicle_combat")
		system.Update(entities, 0.016)
	}
}

// BenchmarkCompanionAISystem benchmarks companion AI behavior processing.
func BenchmarkCompanionAISystem(b *testing.B) {
	world := NewWorld()
	system := &CompanionAISystem{world: world}

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 500, Y: 500})

	// Create 50 companions
	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&VelocityComponent{})
		entity.AddComponent(&CompanionComponent{
			OwnerID:       player.ID,
			CompanionType: CompanionTypePet,
			Loyalty:       75.0,
			Experience:    100.0,
			Level:         5,
			Behavior:      BehaviorPassive, // Using valid constant
		})
		entity.AddComponent(&CompanionStatsComponent{
			Attack:  25.0,
			Defense: 20.0,
			Speed:   40.0,
			HP:      150.0,
			MaxHP:   150.0,
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}

// BenchmarkCompanionProgressionSystem benchmarks companion XP and leveling.
func BenchmarkCompanionProgressionSystem(b *testing.B) {
	world := NewWorld()
	system := &CompanionProgressionSystem{world: world}

	// Create 100 companions
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&CompanionComponent{
			CompanionType: CompanionTypePet,
			Loyalty:       50.0,
			Experience:    float64(i * 10),
			Level:         i%10 + 1,
		})
		entity.AddComponent(&CompanionStatsComponent{
			Attack:  20.0,
			Defense: 15.0,
			Speed:   30.0,
			HP:      100.0,
			MaxHP:   100.0,
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}

// BenchmarkMiniGameSystem benchmarks mini-game state processing.
func BenchmarkMiniGameSystem(b *testing.B) {
	world := NewWorld()
	system := &MiniGameSystem{world: world}

	// Create 20 active mini-games
	for i := 0; i < 20; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&MiniGameComponent{
			GameType:    MiniGameCard, // Using valid constant
			Active:      true,
			Difficulty:  0.5,
			TimeLimit:   600.0,
			TimeElapsed: float64(i * 10),
			Reward: &Reward{
				Gold: 100,
				XP:   50,
			},
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(nil, 0.016)
	}
}

// BenchmarkExpressionSystem benchmarks expression/emote processing.
func BenchmarkExpressionSystem(b *testing.B) {
	world := NewWorld()
	system := &ExpressionSystem{
		world: world,
		// audioManager can be nil for benchmarking
	}

	// Create 100 entities with active expressions
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ExpressionComponent{
			ActiveExpression: ExpressionType(i % 12),
			ExpressionTime:   2.0,
			Cooldown:         0.0,
		})
	}

	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}

// BenchmarkV4IntegratedScenario benchmarks multiple V4 systems together (realistic load).
func BenchmarkV4IntegratedScenario(b *testing.B) {
	world := NewWorld()

	// Initialize V4 systems
	vehicleCombat := NewVehicleCombatSystem(world)
	companionAI := &CompanionAISystem{world: world}
	companionProgression := &CompanionProgressionSystem{world: world}
	miniGame := &MiniGameSystem{world: world}
	expression := &ExpressionSystem{world: world}

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 500, Y: 500})

	// Create 10 vehicles
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 50), Y: float64(i * 50)})
		entity.AddComponent(&VehicleComponent{VehicleType: VehicleMech, Speed: 50, MaxSpeed: 100})
		entity.AddComponent(&VehicleCombatComponent{
			RammingDamage:          25,
			WeaponMounted:          true,
			WeaponDamage:           50,
			CurrentWeaponCooldown:  0.0,
			CurrentRammingCooldown: 0.0,
		})
		entity.AddComponent(&HealthComponent{Current: 200, Max: 200})
	}

	// Create 10 companions
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 40), Y: float64(i * 40)})
		entity.AddComponent(&VelocityComponent{})
		entity.AddComponent(&CompanionComponent{
			OwnerID:       player.ID,
			CompanionType: CompanionTypePet,
			Loyalty:       50.0,
			Level:         5,
			Behavior:      BehaviorPassive,
		})
		entity.AddComponent(&CompanionStatsComponent{
			Attack: 20, Defense: 15, Speed: 30, HP: 100, MaxHP: 100,
		})
	}

	// Create 5 active mini-games
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&MiniGameComponent{
			GameType:    MiniGameCard,
			Active:      true,
			Difficulty:  0.5,
			TimeLimit:   600,
			TimeElapsed: 0,
			Reward:      &Reward{Gold: 100, XP: 50},
		})
	}

	// Create 5 entities with expressions
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ExpressionComponent{
			ActiveExpression: ExpressionWave,
			ExpressionTime:   2.0,
		})
	}

	world.Update(0) // Process all pending entities

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Update all V4 systems (simulating one game frame)
		vehicleEntities := world.GetEntitiesWith("vehicle_combat")
		vehicleCombat.Update(vehicleEntities, 0.016)
		companionAI.Update(0.016)
		companionProgression.Update(0.016)
		miniGame.Update(nil, 0.016)
		expression.Update(0.016)
	}
}
