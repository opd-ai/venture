package engine

import (
	"fmt"
	"testing"
)

func TestProjectileSystem_ExplosiveProjectile_Debug(t *testing.T) {
	w := NewWorld()
	sys := NewProjectileSystem(w)

	// Create multiple targets
	target1 := w.CreateEntity()
	target1.AddComponent(&PositionComponent{X: 175.0, Y: 75.0})
	target1.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	target2 := w.CreateEntity()
	target2.AddComponent(&PositionComponent{X: 210.0, Y: 75.0})
	target2.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	target3 := w.CreateEntity()
	target3.AddComponent(&PositionComponent{X: 300.0, Y: 100.0})
	target3.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})

	// Spawn explosive projectile
	projComp := NewExplosiveProjectile(50.0, 400.0, 5.0, 50.0, "grenade", 1)
	sys.SpawnProjectile(100.0, 100.0, 400.0, 0.0, projComp)

	// Process pending entities
	w.Update(0.0)

	// Get projectile position before update
	entities := w.GetEntities()
	for _, e := range entities {
		if _, ok := e.GetComponent("projectile"); ok {
			if pos, ok := e.GetComponent("position"); ok {
				p := pos.(*PositionComponent)
				fmt.Printf("Projectile start: (%.2f, %.2f)\n", p.X, p.Y)
			}
		}
	}

	// Update
	sys.Update(entities, 0.2)
	w.Update(0.0)

	// Check projectile status
	entities = w.GetEntities()
	projectileExists := false
	for _, e := range entities {
		if _, ok := e.GetComponent("projectile"); ok {
			projectileExists = true
			if pos, ok := e.GetComponent("position"); ok {
				p := pos.(*PositionComponent)
				fmt.Printf("Projectile after update: (%.2f, %.2f)\n", p.X, p.Y)
			}
		}
	}
	if !projectileExists {
		fmt.Println("Projectile was despawned (collision occurred)")
	}

	// Check health
	healthComp1, _ := target1.GetComponent("health")
	health1 := healthComp1.(*HealthComponent)
	fmt.Printf("Target 1 health: %.2f\n", health1.Current)

	healthComp2, _ := target2.GetComponent("health")
	health2 := healthComp2.(*HealthComponent)
	fmt.Printf("Target 2 health: %.2f\n", health2.Current)

	healthComp3, _ := target3.GetComponent("health")
	health3 := healthComp3.(*HealthComponent)
	fmt.Printf("Target 3 health: %.2f\n", health3.Current)
}
