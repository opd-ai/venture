package main

// This example demonstrates the combat system.
// Run with: go run -tags test ./examples/combat_demo.go

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	fmt.Println("=== Combat System Example ===")

	// Create world and systems
	world := engine.NewWorld()
	combatSystem := engine.NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Track events
	damageDealt := 0.0
	deathCount := 0

	combatSystem.SetDamageCallback(func(attacker, target *engine.Entity, damage float64) {
		damageDealt += damage
		attackerX, attackerY, _ := engine.GetPosition(attacker)
		targetX, targetY, _ := engine.GetPosition(target)
		fmt.Printf("💥 Entity %d at (%.0f,%.0f) dealt %.1f damage to Entity %d at (%.0f,%.0f)\n",
			attacker.ID, attackerX, attackerY, damage, target.ID, targetX, targetY)
	})

	combatSystem.SetDeathCallback(func(entity *engine.Entity) {
		deathCount++
		posX, posY, _ := engine.GetPosition(entity)
		fmt.Printf("💀 Entity %d at (%.0f,%.0f) has died!\n", entity.ID, posX, posY)
	})

	// Example 1: Basic melee combat
	fmt.Println("Example 1: Basic Melee Combat")
	fmt.Println("-------------------------------")

	warrior := world.CreateEntity()
	warrior.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	warrior.AddComponent(&engine.HealthComponent{Current: 150, Max: 150})
	warrior.AddComponent(&engine.AttackComponent{
		Damage:     25,
		DamageType: combat.DamagePhysical,
		Range:      10,
		Cooldown:   1.0,
	})
	warriorStats := engine.NewStatsComponent()
	warriorStats.Attack = 15
	warriorStats.Defense = 10
	warrior.AddComponent(warriorStats)
	warrior.AddComponent(&engine.TeamComponent{TeamID: 1})

	goblin := world.CreateEntity()
	goblin.AddComponent(&engine.PositionComponent{X: 8, Y: 0})
	goblin.AddComponent(&engine.HealthComponent{Current: 50, Max: 50})
	goblin.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      10,
		Cooldown:   1.5,
	})
	goblinStats := engine.NewStatsComponent()
	goblinStats.Attack = 5
	goblinStats.Defense = 2
	goblin.AddComponent(goblinStats)
	goblin.AddComponent(&engine.TeamComponent{TeamID: 2})

	world.Update(0) // Process additions

	fmt.Println("⚔️  Warrior (HP: 150, ATK: 40, DEF: 10) vs Goblin (HP: 50, ATK: 20, DEF: 2)")
	fmt.Println()

	// Simulate combat
	for i := 0; i < 5; i++ {
		// Update cooldowns
		world.Update(0.5)

		// Warrior attacks goblin
		if combatSystem.Attack(warrior, goblin) {
			goblinHealth, _ := goblin.GetComponent("health")
			fmt.Printf("   Goblin HP: %.0f/50\n", goblinHealth.(*engine.HealthComponent).Current)
		}

		// Goblin counterattacks if alive
		goblinHealth, _ := goblin.GetComponent("health")
		if !goblinHealth.(*engine.HealthComponent).IsDead() {
			world.Update(0.5)
			if combatSystem.Attack(goblin, warrior) {
				warriorHealth, _ := warrior.GetComponent("health")
				fmt.Printf("   Warrior HP: %.0f/150\n", warriorHealth.(*engine.HealthComponent).Current)
			}
		}

		fmt.Println()
	}

	// Example 2: Magic combat with resistances
	fmt.Println("\nExample 2: Magic Combat with Resistances")
	fmt.Println("-----------------------------------------")

	// Reset world
	world = engine.NewWorld()
	combatSystem = engine.NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	mage := world.CreateEntity()
	mage.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	mage.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	mage.AddComponent(&engine.AttackComponent{
		Damage:     30,
		DamageType: combat.DamageFire,
		Range:      100,
		Cooldown:   2.0,
	})
	mageStats := engine.NewStatsComponent()
	mageStats.MagicPower = 20
	mage.AddComponent(mageStats)

	// Fire elemental with fire resistance
	elemental := world.CreateEntity()
	elemental.AddComponent(&engine.PositionComponent{X: 50, Y: 0})
	elemental.AddComponent(&engine.HealthComponent{Current: 80, Max: 80})
	elementalStats := engine.NewStatsComponent()
	elementalStats.MagicDefense = 5
	elementalStats.Resistances[combat.DamageFire] = 0.75 // 75% fire resistance
	elemental.AddComponent(elementalStats)

	world.Update(0)

	fmt.Println("🔥 Mage (Fire damage: 30+20) vs Fire Elemental (75% fire resistance)")

	combatSystem.SetDamageCallback(func(attacker, target *engine.Entity, damage float64) {
		fmt.Printf("   Damage dealt: %.1f (reduced by resistance)\n", damage)
	})

	combatSystem.Attack(mage, elemental)
	elementalHealth, _ := elemental.GetComponent("health")
	fmt.Printf("   Elemental HP: %.0f/80\n\n", elementalHealth.(*engine.HealthComponent).Current)

	// Example 3: Status effects (poison)
	fmt.Println("Example 3: Status Effects")
	fmt.Println("-------------------------")

	world = engine.NewWorld()
	combatSystem = engine.NewCombatSystem(54321)
	world.AddSystem(combatSystem)

	target := world.CreateEntity()
	target.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	fmt.Println("🧪 Applying poison effect (10 damage per second for 5 seconds)")
	combatSystem.ApplyStatusEffect(target, "poison", 5.0, 10.0, 1.0)

	targetHealth, _ := target.GetComponent("health")
	fmt.Printf("Initial HP: %.0f\n", targetHealth.(*engine.HealthComponent).Current)

	// Simulate over time
	for i := 1; i <= 6; i++ {
		world.Update(1.0)
		targetHealth, _ = target.GetComponent("health")
		fmt.Printf("After %d second(s): HP = %.0f\n", i, targetHealth.(*engine.HealthComponent).Current)
	}

	// Check effect expired
	_, hasEffect := target.GetComponent("status_effect")
	if !hasEffect {
		fmt.Println("✓ Poison effect expired")
	}

	// Example 4: Critical hits
	fmt.Println("Example 4: Critical Hits")
	fmt.Println("------------------------")

	world = engine.NewWorld()
	combatSystem = engine.NewCombatSystem(99999)
	world.AddSystem(combatSystem)

	critHits := 0
	normalHits := 0

	combatSystem.SetDamageCallback(func(attacker, target *engine.Entity, damage float64) {
		if damage > 30 {
			critHits++
			fmt.Printf("⚡ CRITICAL HIT! %.0f damage\n", damage)
		} else {
			normalHits++
			fmt.Printf("   Normal hit: %.0f damage\n", damage)
		}
	})

	rogue := world.CreateEntity()
	rogue.AddComponent(&engine.AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      10,
		Cooldown:   0,
	})
	rogueStats := engine.NewStatsComponent()
	rogueStats.Attack = 10
	rogueStats.CritChance = 0.3 // 30% crit chance
	rogueStats.CritDamage = 2.5 // 250% damage on crit
	rogue.AddComponent(rogueStats)

	dummy := world.CreateEntity()
	dummy.AddComponent(&engine.HealthComponent{Current: 1000, Max: 1000})
	dummy.AddComponent(engine.NewStatsComponent())

	world.Update(0)

	fmt.Println("🗡️  Rogue attacking (30% crit chance, 2.5x crit damage):")
	fmt.Println()

	// Perform multiple attacks to see crits
	for i := 0; i < 20; i++ {
		combatSystem.Attack(rogue, dummy)
	}

	fmt.Printf("\nResults: %d critical hits, %d normal hits\n", critHits, normalHits)
	fmt.Printf("Crit rate: %.1f%%\n\n", float64(critHits)/20.0*100)

	// Example 5: Team-based combat
	fmt.Println("Example 5: Team-Based Combat")
	fmt.Println("-----------------------------")

	world = engine.NewWorld()
	combatSystem = engine.NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Create player team
	_ = createCombatEntity(world, 1, 0, 0, 100, 1)
	_ = createCombatEntity(world, 1, 20, 0, 100, 1)

	// Create enemy team
	_ = createCombatEntity(world, 2, 100, 0, 80, 2)
	_ = createCombatEntity(world, 2, 120, 0, 80, 2)
	_ = createCombatEntity(world, 2, 110, 20, 80, 2)

	world.Update(0)

	// Create a reference player for enemy finding
	player1 := world.CreateEntity()
	player1.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	player1.AddComponent(&engine.TeamComponent{TeamID: 1})
	world.Update(0)

	fmt.Println("👥 Team 1 (2 players) vs Team 2 (3 enemies)")
	fmt.Printf("   Player 1 at (0, 0)\n")
	fmt.Printf("   Player 2 at (20, 0)\n")
	fmt.Printf("   Enemy 1 at (100, 0)\n")
	fmt.Printf("   Enemy 2 at (120, 0)\n")
	fmt.Printf("   Enemy 3 at (110, 20)\n\n")

	// Find enemies in range for player 1
	enemies := engine.FindEnemiesInRange(world, player1, 150)
	fmt.Printf("Player 1 can see %d enemies within range 150\n", len(enemies))

	// Find nearest enemy
	nearest := engine.FindNearestEnemy(world, player1, 150)
	if nearest != nil {
		nearestX, nearestY, _ := engine.GetPosition(nearest)
		fmt.Printf("Nearest enemy to Player 1 is at (%.0f, %.0f)\n", nearestX, nearestY)
		distance := engine.GetDistance(player1, nearest)
		fmt.Printf("Distance: %.1f units\n", distance)
	}

	fmt.Println("\n=== Combat System Demo Complete ===")
	fmt.Printf("\nTotal damage dealt: %.1f\n", damageDealt)
	fmt.Printf("Total deaths: %d\n", deathCount)
}

// Helper function to create a combat entity
func createCombatEntity(world *engine.World, id uint64, x, y, hp float64, team int) *engine.Entity {
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
	entity.AddComponent(&engine.HealthComponent{Current: hp, Max: hp})
	entity.AddComponent(&engine.AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   1.0,
	})
	entity.AddComponent(engine.NewStatsComponent())
	entity.AddComponent(&engine.TeamComponent{TeamID: team})
	return entity
}
