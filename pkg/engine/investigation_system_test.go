package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/story"
)

// TestNewInvestigationSystem tests system creation.
func TestNewInvestigationSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewInvestigationSystem(world, seed)

	if system.world != world {
		t.Error("System world not set correctly")
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.fragmentHidden == nil {
		t.Error("fragmentHidden map not initialized")
	}
}

// TestInvestigationSystem_StartInvestigation tests starting an investigation.
func TestInvestigationSystem_StartInvestigation(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create entity with investigation component
	entity := world.CreateEntity()
	entity.AddComponent(NewInvestigationComponent())

	// Start investigation
	success := system.StartInvestigation(entity)

	if !success {
		t.Error("StartInvestigation should succeed with valid component")
	}

	invComp, _ := entity.GetComponent("investigation")
	inv := invComp.(*InvestigationComponent)

	if !inv.IsInvestigating {
		t.Error("Investigation should be active after starting")
	}
}

// TestInvestigationSystem_StartInvestigation_NoComponent tests starting without component.
func TestInvestigationSystem_StartInvestigation_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create entity without investigation component
	entity := world.CreateEntity()

	// Start investigation (should fail)
	success := system.StartInvestigation(entity)

	if success {
		t.Error("StartInvestigation should fail without component")
	}
}

// TestInvestigationSystem_StartInvestigation_Cooldown tests cooldown check.
func TestInvestigationSystem_StartInvestigation_Cooldown(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.CooldownElapsed = 0.5 // On cooldown
	entity.AddComponent(invComp)

	// Should fail due to cooldown
	success := system.StartInvestigation(entity)

	if success {
		t.Error("StartInvestigation should fail during cooldown")
	}
}

// TestInvestigationSystem_IsInvestigating tests investigation status check.
func TestInvestigationSystem_IsInvestigating(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewInvestigationComponent())

	// Initially not investigating
	if system.IsInvestigating(entity) {
		t.Error("Should not be investigating initially")
	}

	// Start investigating
	system.StartInvestigation(entity)

	if !system.IsInvestigating(entity) {
		t.Error("Should be investigating after start")
	}
}

// TestInvestigationSystem_SetFragmentHidden tests hiding fragments.
func TestInvestigationSystem_SetFragmentHidden(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	fragID := uint64(12345)

	// Initially not hidden
	if system.IsFragmentHidden(fragID) {
		t.Error("Fragment should not be hidden initially")
	}

	// Hide fragment
	system.SetFragmentHidden(fragID, true)

	if !system.IsFragmentHidden(fragID) {
		t.Error("Fragment should be hidden after setting")
	}

	// Unhide fragment
	system.SetFragmentHidden(fragID, false)

	if system.IsFragmentHidden(fragID) {
		t.Error("Fragment should not be hidden after unsetting")
	}
}

// TestInvestigationSystem_GetInvestigationProgress tests progress retrieval.
func TestInvestigationSystem_GetInvestigationProgress(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.TotalInvestigations = 5
	invComp.MarkFragmentRevealed(111)
	invComp.MarkFragmentRevealed(222)
	invComp.MarkFragmentRevealed(333)
	entity.AddComponent(invComp)

	total, revealed, err := system.GetInvestigationProgress(entity)

	if err != nil {
		t.Fatalf("GetInvestigationProgress failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected 5 total investigations, got %d", total)
	}

	if revealed != 3 {
		t.Errorf("Expected 3 revealed fragments, got %d", revealed)
	}
}

// TestInvestigationSystem_GetInvestigationProgress_NoComponent tests error handling.
func TestInvestigationSystem_GetInvestigationProgress_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()

	_, _, err := system.GetInvestigationProgress(entity)

	if err == nil {
		t.Error("Expected error for entity without investigation component")
	}
}

// TestInvestigationSystem_HideRandomFragments tests random hiding.
func TestInvestigationSystem_HideRandomFragments(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create some fragment entities
	for i := 0; i < 10; i++ {
		frag := world.CreateEntity()
		frag.AddComponent(&StoryFragmentComponent{
			Fragment: story.StoryFragment{
				Type: story.FragmentNote,
			},
			SeriesID:    "test",
			SequenceNum: i,
		})
	}

	world.Update(0.0) // Commit entities

	// Verify fragments were created
	fragments := world.GetEntitiesWith("storyfragment")
	if len(fragments) != 10 {
		t.Fatalf("Expected 10 fragments, got %d", len(fragments))
	}

	// Hide 50% of fragments
	system.HideRandomFragments(0.5)

	// Count hidden fragments
	hiddenCount := 0
	for _, isHidden := range system.fragmentHidden {
		if isHidden {
			hiddenCount++
		}
	}

	// Should be roughly half (with some RNG variance)
	if hiddenCount < 2 || hiddenCount > 8 {
		t.Errorf("Expected ~5 hidden fragments with 50%% rate, got %d", hiddenCount)
	}
}

// TestInvestigationSystem_HideRandomFragments_InvalidPercentage tests boundary handling.
func TestInvestigationSystem_HideRandomFragments_InvalidPercentage(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create fragment
	frag := world.CreateEntity()
	frag.AddComponent(&StoryFragmentComponent{})

	// Try invalid percentages (should not crash)
	system.HideRandomFragments(-0.5)
	system.HideRandomFragments(1.5)

	// No hidden fragments should exist (invalid percentages ignored)
	hiddenCount := 0
	for _, isHidden := range system.fragmentHidden {
		if isHidden {
			hiddenCount++
		}
	}

	if hiddenCount != 0 {
		t.Errorf("Expected 0 hidden fragments with invalid percentage, got %d", hiddenCount)
	}
}

// TestInvestigationSystem_Update_ProcessesInvestigations tests update processing.
func TestInvestigationSystem_Update_ProcessesInvestigations(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create investigator with position
	investigator := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.InvestigationDuration = 0.05 // Very short duration
	investigator.AddComponent(invComp)
	investigator.AddComponent(&PositionComponent{X: 100, Y: 100})

	world.Update(0.0) // Commit entity

	// Start investigation
	if !system.StartInvestigation(investigator) {
		t.Fatal("Failed to start investigation")
	}

	// Sleep to allow time to pass
	time.Sleep(60 * time.Millisecond)

	// Update system
	system.Update(0.1)

	// Need to commit changes
	world.Update(0.0)

	// Investigation should have stopped after completion
	if system.IsInvestigating(investigator) {
		t.Error("Investigation should have stopped after completion")
	}
}

// TestInvestigationSystem_Update_UpdatesCooldown tests cooldown updating.
func TestInvestigationSystem_Update_UpdatesCooldown(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.CooldownElapsed = 0.3
	entity.AddComponent(invComp)
	entity.AddComponent(&PositionComponent{})

	world.Update(0.0) // Commit entity

	system.Update(0.5)

	inv, _ := entity.GetComponent("investigation")
	invComponent := inv.(*InvestigationComponent)

	if invComponent.CooldownElapsed != 0.8 {
		t.Errorf("Expected CooldownElapsed 0.8, got %f", invComponent.CooldownElapsed)
	}
}

// TestInvestigationSystem_FragmentReveal tests fragment revelation.
func TestInvestigationSystem_FragmentReveal(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 54321) // Different seed for determinism

	// Create investigator
	investigator := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.InvestigationSkillBonus = 1.0 // 100% detection chance
	invComp.InvestigationDuration = 0.05  // Short duration
	investigator.AddComponent(invComp)
	investigator.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create hidden fragment nearby
	fragment := world.CreateEntity()
	fragment.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:        story.FragmentNote,
			Content:     "Test fragment",
			DiscoveryXP: 10,
		},
		SeriesID:    "test_series",
		SequenceNum: 1,
	})
	fragment.AddComponent(&PositionComponent{X: 110, Y: 110}) // Close to investigator

	system.SetFragmentHidden(fragment.ID, true)

	world.Update(0.0) // Commit entities

	// Start investigation
	if !system.StartInvestigation(investigator) {
		t.Fatal("Failed to start investigation")
	}

	// Wait for investigation to complete
	time.Sleep(60 * time.Millisecond)

	// Process investigation
	system.Update(0.1)

	// Fragment should be revealed
	if system.IsFragmentHidden(fragment.ID) {
		t.Error("Fragment should have been revealed")
	}

	// Check investigator's revealed list
	inv, _ := investigator.GetComponent("investigation")
	invComponent := inv.(*InvestigationComponent)

	if !invComponent.HasRevealedFragment(fragment.ID) {
		t.Error("Fragment should be in investigator's revealed list")
	}
}

// TestInvestigationSystem_FragmentReveal_OutOfRange tests distance check.
func TestInvestigationSystem_FragmentReveal_OutOfRange(t *testing.T) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Create investigator
	investigator := world.CreateEntity()
	invComp := NewInvestigationComponent()
	invComp.InvestigationSkillBonus = 1.0
	investigator.AddComponent(invComp)
	investigator.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create hidden fragment far away
	fragment := world.CreateEntity()
	fragment.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{Type: story.FragmentNote},
	})
	fragment.AddComponent(&PositionComponent{X: 500, Y: 500}) // Far away

	system.SetFragmentHidden(fragment.ID, true)

	world.Update(0.0) // Commit entities

	// Start and complete investigation
	system.StartInvestigation(investigator)
	invComp.InvestigationStartTime = invComp.InvestigationStartTime.Add(-3)
	system.Update(0.1)

	// Fragment should still be hidden (out of range)
	if !system.IsFragmentHidden(fragment.ID) {
		t.Error("Fragment should still be hidden (out of range)")
	}
}

// Benchmark investigation system operations
func BenchmarkInvestigationSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	// Setup entities
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewInvestigationComponent())
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
	}

	world.Update(0.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}

func BenchmarkInvestigationSystem_StartInvestigation(b *testing.B) {
	world := NewWorld()
	system := NewInvestigationSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(NewInvestigationComponent())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invComp, _ := entity.GetComponent("investigation")
		inv := invComp.(*InvestigationComponent)
		inv.CooldownElapsed = 1.0 // Reset cooldown
		system.StartInvestigation(entity)
	}
}
