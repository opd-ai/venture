// Package engine provides tests for fishing system determinism.
// This test file verifies that the FishingSystem uses deterministic random
// generation as required by the project's code assistance guidelines.

package engine

import (
	"testing"
)

// TestFishingSystem_Determinism verifies that fish selection is deterministic
// when using the same seed.
func TestFishingSystem_Determinism(t *testing.T) {
	seed := int64(42)

	// Create two fishing systems with the same seed
	world1 := NewWorld()
	fs1 := NewFishingSystem(world1, seed)

	world2 := NewWorld()
	fs2 := NewFishingSystem(world2, seed)

	// Create identical fishing spots
	spot1 := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
	fs1.SpotAddFishType(spot1, "bass", 10.0)
	fs1.SpotAddFishType(spot1, "trout", 8.0)
	fs1.SpotAddFishType(spot1, "pike", 5.0)

	spot2 := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
	fs2.SpotAddFishType(spot2, "bass", 10.0)
	fs2.SpotAddFishType(spot2, "trout", 8.0)
	fs2.SpotAddFishType(spot2, "pike", 5.0)

	// Create identical fishing components
	fishComp1 := NewFishingComponent()
	fishComp1.FishingSkill = 50
	fishComp1.CastDistance = 0.7

	fishComp2 := NewFishingComponent()
	fishComp2.FishingSkill = 50
	fishComp2.CastDistance = 0.7

	// Select fish multiple times - should get identical results
	for i := 0; i < 10; i++ {
		fish1, _ := fs1.selectFish(spot1, fishComp1)
		fish2, _ := fs2.selectFish(spot2, fishComp2)

		if fish1 == nil && fish2 == nil {
			continue // Both nil is acceptable
		}

		if fish1 == nil || fish2 == nil {
			t.Fatalf("iteration %d: one fish is nil, other is not (fish1=%v, fish2=%v)", i, fish1, fish2)
		}

		if fish1.ID != fish2.ID {
			t.Errorf("iteration %d: fish mismatch (fish1=%s, fish2=%s)", i, fish1.ID, fish2.ID)
		}
	}
}

// TestFishingSystem_WeightDeterminism verifies that fish weight calculation is deterministic.
func TestFishingSystem_WeightDeterminism(t *testing.T) {
	seed := int64(123)

	world1 := NewWorld()
	fs1 := NewFishingSystem(world1, seed)

	world2 := NewWorld()
	fs2 := NewFishingSystem(world2, seed)

	// Get a fish type
	fish := fs1.GetFishType("bass")
	if fish == nil {
		t.Fatal("bass fish type not found")
	}

	skillLevel := 25

	// Calculate weight multiple times - should be identical
	for i := 0; i < 5; i++ {
		weight1 := fs1.calculateFishWeight(fish, skillLevel)
		weight2 := fs2.calculateFishWeight(fish, skillLevel)

		if weight1 != weight2 {
			t.Errorf("iteration %d: weight mismatch (weight1=%f, weight2=%f)", i, weight1, weight2)
		}
	}
}

// TestFishingSystem_NonDeterministicWithDifferentSeeds verifies that different
// seeds produce different results (sanity check).
func TestFishingSystem_NonDeterministicWithDifferentSeeds(t *testing.T) {
	world1 := NewWorld()
	fs1 := NewFishingSystem(world1, 111)

	world2 := NewWorld()
	fs2 := NewFishingSystem(world2, 222)

	fish := fs1.GetFishType("bass")
	if fish == nil {
		t.Fatal("bass fish type not found")
	}

	skillLevel := 25

	// Calculate weights - should be different with high probability
	weight1 := fs1.calculateFishWeight(fish, skillLevel)
	weight2 := fs2.calculateFishWeight(fish, skillLevel)

	if weight1 == weight2 {
		t.Log("WARNING: Different seeds produced same weight (unlikely but possible)")
	}
}

// TestFishingSystem_SelectRandomFishDeterminism tests the weighted random fish selection.
func TestFishingSystem_SelectRandomFishDeterminism(t *testing.T) {
	seed := int64(999)

	world1 := NewWorld()
	fs1 := NewFishingSystem(world1, seed)

	world2 := NewWorld()
	fs2 := NewFishingSystem(world2, seed)

	// Create eligible fish lists
	eligible1 := eligibleFishList{
		fish: []weightedFish{
			{fish: fs1.GetFishType("bass"), weight: 10.0},
			{fish: fs1.GetFishType("trout"), weight: 5.0},
			{fish: fs1.GetFishType("pike"), weight: 2.0},
		},
		totalWeight: 17.0,
	}

	eligible2 := eligibleFishList{
		fish: []weightedFish{
			{fish: fs2.GetFishType("bass"), weight: 10.0},
			{fish: fs2.GetFishType("trout"), weight: 5.0},
			{fish: fs2.GetFishType("pike"), weight: 2.0},
		},
		totalWeight: 17.0,
	}

	// Select fish multiple times
	for i := 0; i < 20; i++ {
		fish1 := fs1.selectRandomFish(eligible1)
		fish2 := fs2.selectRandomFish(eligible2)

		if fish1 == nil && fish2 == nil {
			continue
		}

		if fish1 == nil || fish2 == nil {
			t.Fatalf("iteration %d: one fish is nil, other is not", i)
		}

		if fish1.ID != fish2.ID {
			t.Errorf("iteration %d: selected fish mismatch (fish1=%s, fish2=%s)", i, fish1.ID, fish2.ID)
		}
	}
}
