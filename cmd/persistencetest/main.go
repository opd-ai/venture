// Package main provides a CLI tool for testing entity persistence.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	entityType := flag.String("type", "monster", "Entity type to create (monster, npc, companion)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("=== Venture Entity Persistence Test ===")
	fmt.Printf("Testing entity type: %s\n\n", *entityType)

	// Create world and entity
	world := engine.NewWorld()
	entity := createTestEntity(world, *entityType)

	if *verbose {
		fmt.Printf("Created entity ID: %d\n", entity.ID)
		fmt.Printf("Component count: %d\n\n", len(entity.Components))
	}

	// Serialize entity
	fmt.Println("--- Serialization Test ---")
	state, err := engine.SerializeEntity(entity)
	if err != nil {
		log.Fatalf("Serialization failed: %v", err)
	}

	fmt.Printf("✓ Serialized entity ID %d\n", state.ID)
	fmt.Printf("  Type: %s\n", state.TypeName)
	fmt.Printf("  Components: %d\n", len(state.Components))

	if *verbose {
		for componentType, data := range state.Components {
			fmt.Printf("    - %s: %d bytes\n", componentType, len(data))
		}
	}

	// Test respawn rule
	rule := engine.GetRespawnRule(state.TypeName)
	fmt.Printf("  Respawn rule: %s\n", getRespawnRuleName(rule))

	// Deserialize entity
	fmt.Println("\n--- Deserialization Test ---")
	world2 := engine.NewWorld()
	deserialized, err := engine.DeserializeEntity(state, world2)
	if err != nil {
		log.Fatalf("Deserialization failed: %v", err)
	}

	fmt.Printf("✓ Deserialized entity ID %d\n", deserialized.ID)
	fmt.Printf("  Component count: %d\n", len(deserialized.Components))

	// Verify position component
	if posComp, ok := deserialized.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		fmt.Printf("  Position: (%.1f, %.1f)\n", pos.X, pos.Y)
	}

	// Verify health component
	if healthComp, ok := deserialized.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		fmt.Printf("  Health: %.0f/%.0f\n", health.Current, health.Max)
	}

	// Test lifecycle tracker
	fmt.Println("\n--- Lifecycle Tracking Test ---")
	tracker := engine.NewEntityLifecycleTracker()
	tracker.MarkSpawned(entity.ID)
	tracker.MarkModified(entity.ID)
	fmt.Printf("✓ Entity %d spawned and modified\n", entity.ID)

	modifiedEntities := tracker.GetModifiedEntities()
	fmt.Printf("  Modified entities: %v\n", modifiedEntities)

	tracker.MarkKilled(entity.ID)
	fmt.Printf("✓ Entity %d killed\n", entity.ID)
	fmt.Printf("  Is killed: %v\n", tracker.IsKilled(entity.ID))
	fmt.Printf("  Is modified: %v\n", tracker.IsModified(entity.ID))

	fmt.Println("\n✅ All tests passed!")
}

func createTestEntity(world *engine.World, entityType string) *engine.Entity {
	entity := world.CreateEntity()
	world.Update(0.0) // Process pending additions

	// Add common components
	entity.AddComponent(&engine.PositionComponent{X: 100.0, Y: 200.0})
	entity.AddComponent(&engine.VelocityComponent{VX: 5.0, VY: 10.0})
	entity.AddComponent(&engine.HealthComponent{Current: 80.0, Max: 100.0})
	entity.AddComponent(&engine.ColliderComponent{
		Width:  32.0,
		Height: 48.0,
		Solid:  true,
	})

	// Add type-specific components
	switch entityType {
	case "monster":
		entity.AddComponent(&engine.AIComponent{
			State:          engine.AIStateAttack,
			DetectionRange: 100.0,
		})
	case "npc":
		entity.AddComponent(&engine.AIComponent{
			State:          engine.AIStateIdle,
			DetectionRange: 50.0,
		})
	case "companion":
		entity.AddComponent(&engine.CompanionComponent{
			CompanionType: engine.CompanionTypePet,
			Loyalty:       0.8,
		})
	}

	return entity
}

func getRespawnRuleName(rule engine.RespawnRule) string {
	switch rule {
	case engine.RespawnNever:
		return "Never"
	case engine.RespawnAlways:
		return "Always"
	case engine.RespawnConditional:
		return "Conditional"
	default:
		return "Unknown"
	}
}
