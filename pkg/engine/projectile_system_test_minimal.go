package engine

import (
"fmt"
"testing"
)

func TestProjectileSystem_CollisionMinimal(t *testing.T) {
w := NewWorld()
sys := NewProjectileSystem(w)

// Create target at (115, 100)
target := w.CreateEntity()
target.AddComponent(&PositionComponent{X: 115.0, Y: 100.0})
target.AddComponent(&HealthComponent{Current: 100.0, Max: 100.0})
fmt.Printf("Created target ID=%d at (115, 100), health=100\n", target.ID)

// Spawn projectile at (100, 100) moving right at 100 units/sec
projComp := NewProjectileComponent(25.0, 100.0, 10.0, "arrow", 999)
proj := sys.SpawnProjectile(100.0, 100.0, 100.0, 0.0, projComp)
fmt.Printf("Created projectile ID=%d at (100, 100), velocity=(100, 0), ownerID=999\n", proj.ID)

// Process pending
w.Update(0.0)

// Check initial state
entities := w.GetEntitiesWith("health", "position")
fmt.Printf("Entities with health and position: %d\n", len(entities))
for _, e := range entities {
if pos, ok := e.GetComponent("position"); ok {
p := pos.(*PositionComponent)
fmt.Printf("  Entity ID=%d at (%.1f, %.1f)\n", e.ID, p.X, p.Y)
}
}

// Update for 0.15 seconds - projectile should be at (115, 100), exactly on target
fmt.Println("\nUpdating for 0.15s...")
sys.Update(nil, 0.15)
w.Update(0.0)

// Check projectile position
if projPos, ok := proj.GetComponent("position"); ok {
p := projPos.(*PositionComponent)
fmt.Printf("Projectile now at (%.1f, %.1f)\n", p.X, p.Y)
} else {
fmt.Println("Projectile despawned (collision occurred?)")
}

// Check target health
if healthComp, ok := target.GetComponent("health"); ok {
health := healthComp.(*HealthComponent)
fmt.Printf("Target health: %.1f\n", health.Current)
if health.Current == 100.0 {
t.Error("Target should have taken damage")
}
} else {
t.Fatal("Target health component missing")
}
}
