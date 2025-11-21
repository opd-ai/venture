package raids

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestInstanceManager_CreateInstance(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "test-group",
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	raid := result.(*RaidDungeon)

	t.Run("valid instance", func(t *testing.T) {
		instance, err := im.CreateInstance(raid, "group-1", []string{"p1", "p2", "p3", "p4", "p5"})
		if err != nil {
			t.Fatalf("CreateInstance() error = %v", err)
		}

		if instance.GroupID != "group-1" {
			t.Errorf("GroupID = %q, want %q", instance.GroupID, "group-1")
		}

		if len(instance.PlayerIDs) != 5 {
			t.Errorf("PlayerIDs count = %d, want 5", len(instance.PlayerIDs))
		}
	})

	t.Run("insufficient players", func(t *testing.T) {
		_, err := im.CreateInstance(raid, "group-2", []string{"p1", "p2"})
		if err == nil {
			t.Error("CreateInstance() should fail with insufficient players")
		}
	})

	t.Run("too many players", func(t *testing.T) {
		players := make([]string, 20)
		for i := range players {
			players[i] = "player"
		}
		_, err := im.CreateInstance(raid, "group-3", players)
		if err == nil {
			t.Error("CreateInstance() should fail with too many players")
		}
	})
}

func TestInstanceManager_GetInstance(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      12,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"tier":     TierHeroic,
			"group_id": "get-test",
		},
	}

	result, _ := gen.Generate(54321, params)
	raid := result.(*RaidDungeon)

	instance, _ := im.CreateInstance(raid, "get-group", []string{"p1", "p2", "p3", "p4", "p5", "p6"})

	t.Run("existing instance", func(t *testing.T) {
		retrieved, exists := im.GetInstance(instance.InstanceID)
		if !exists {
			t.Fatal("GetInstance() should find existing instance")
		}

		if retrieved.InstanceID != instance.InstanceID {
			t.Errorf("InstanceID = %q, want %q", retrieved.InstanceID, instance.InstanceID)
		}
	})

	t.Run("non-existent instance", func(t *testing.T) {
		_, exists := im.GetInstance("nonexistent")
		if exists {
			t.Error("GetInstance() should not find non-existent instance")
		}
	})
}

func TestInstanceManager_GetGroupInstance(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "horror",
		Custom: map[string]interface{}{
			"tier":     TierMythic,
			"group_id": "group-test",
		},
	}

	result, _ := gen.Generate(99999, params)
	raid := result.(*RaidDungeon)

	groupID := "test-group-99"
	im.CreateInstance(raid, groupID, []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"})

	t.Run("find group instance", func(t *testing.T) {
		instance, exists := im.GetGroupInstance(groupID, TierMythic)
		if !exists {
			t.Fatal("GetGroupInstance() should find group's instance")
		}

		if instance.GroupID != groupID {
			t.Errorf("GroupID = %q, want %q", instance.GroupID, groupID)
		}
	})

	t.Run("wrong tier", func(t *testing.T) {
		_, exists := im.GetGroupInstance(groupID, TierNormal)
		if exists {
			t.Error("GetGroupInstance() should not find instance with wrong tier")
		}
	})
}

func TestInstanceManager_CompleteInstance(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "cyberpunk",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "complete-test",
		},
	}

	result, _ := gen.Generate(11111, params)
	raid := result.(*RaidDungeon)

	instance, _ := im.CreateInstance(raid, "complete-group", []string{"p1", "p2", "p3", "p4", "p5"})

	t.Run("complete valid instance", func(t *testing.T) {
		err := im.CompleteInstance(instance.InstanceID)
		if err != nil {
			t.Errorf("CompleteInstance() error = %v", err)
		}

		retrieved, _ := im.GetInstance(instance.InstanceID)
		if !retrieved.Completed {
			t.Error("Instance should be marked as completed")
		}
	})

	t.Run("complete non-existent", func(t *testing.T) {
		err := im.CompleteInstance("nonexistent")
		if err == nil {
			t.Error("CompleteInstance() should error on non-existent instance")
		}
	})

	t.Run("already completed", func(t *testing.T) {
		err := im.CompleteInstance(instance.InstanceID)
		if err == nil {
			t.Error("CompleteInstance() should error on already completed instance")
		}
	})
}

func TestInstanceManager_RemoveInstance(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      12,
		GenreID:    "postapoc",
		Custom: map[string]interface{}{
			"tier":     TierLegendary,
			"group_id": "remove-test",
		},
	}

	result, _ := gen.Generate(22222, params)
	raid := result.(*RaidDungeon)

	instance, _ := im.CreateInstance(raid, "remove-group", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"})

	im.RemoveInstance(instance.InstanceID)

	_, exists := im.GetInstance(instance.InstanceID)
	if exists {
		t.Error("Instance should be removed")
	}
}

func TestInstanceManager_CleanupExpired(t *testing.T) {
	im := NewInstanceManagerWithTimeout(10 * time.Millisecond)
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "cleanup-test",
		},
	}

	result, _ := gen.Generate(33333, params)
	raid := result.(*RaidDungeon)

	im.CreateInstance(raid, "cleanup-1", []string{"p1", "p2", "p3", "p4", "p5"})
	im.CreateInstance(raid, "cleanup-2", []string{"p1", "p2", "p3", "p4", "p5"})

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	removed := im.CleanupExpired()
	if removed != 2 {
		t.Errorf("CleanupExpired() removed %d, want 2", removed)
	}
}

func TestInstanceManager_GetActiveInstanceCount(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "count-test",
		},
	}

	result, _ := gen.Generate(44444, params)
	raid := result.(*RaidDungeon)

	if count := im.GetActiveInstanceCount(); count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	im.CreateInstance(raid, "count-1", []string{"p1", "p2", "p3", "p4", "p5"})
	im.CreateInstance(raid, "count-2", []string{"p1", "p2", "p3", "p4", "p5"})

	if count := im.GetActiveInstanceCount(); count != 2 {
		t.Errorf("Count after creates = %d, want 2", count)
	}
}

func TestInstanceManager_GetGroupInstances(t *testing.T) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "multi-test",
		},
	}

	result, _ := gen.Generate(55555, params)
	raid := result.(*RaidDungeon)

	groupID := "multi-group"
	im.CreateInstance(raid, groupID, []string{"p1", "p2", "p3", "p4", "p5"})

	instances := im.GetGroupInstances(groupID)
	if len(instances) != 1 {
		t.Errorf("GetGroupInstances() returned %d, want 1", len(instances))
	}

	if instances[0].GroupID != groupID {
		t.Errorf("GroupID = %q, want %q", instances[0].GroupID, groupID)
	}
}

func BenchmarkInstanceManager_CreateInstance(b *testing.B) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "bench",
		},
	}

	result, _ := gen.Generate(12345, params)
	raid := result.(*RaidDungeon)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		im.CreateInstance(raid, "bench-group", []string{"p1", "p2", "p3", "p4", "p5"})
	}
}

func BenchmarkInstanceManager_GetInstance(b *testing.B) {
	im := NewInstanceManager()
	gen := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "bench",
		},
	}

	result, _ := gen.Generate(12345, params)
	raid := result.(*RaidDungeon)

	instance, _ := im.CreateInstance(raid, "bench-group", []string{"p1", "p2", "p3", "p4", "p5"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		im.GetInstance(instance.InstanceID)
	}
}
