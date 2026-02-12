package engine

import "testing"

// TestFishingSystem_MultiplayerSync demonstrates that FishingSystem now produces
// identical results across server and client with the same seed, fixing the
// multiplayer desync issue described in AUDIT.md Gap #1.
func TestFishingSystem_MultiplayerSync(t *testing.T) {
	// Simulate server and client with same world seed
	serverSeed := int64(987654321)

	// Server-side fishing system
	serverWorld := NewWorld()
	serverFS := NewFishingSystem(serverWorld, serverSeed)

	// Client-side fishing system (same seed)
	clientWorld := NewWorld()
	clientFS := NewFishingSystem(clientWorld, serverSeed)

	// Both create identical fishing spots
	serverSpot := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
	serverFS.SpotAddFishType(serverSpot, "bass", 10.0)
	serverFS.SpotAddFishType(serverSpot, "trout", 8.0)
	serverFS.SpotAddFishType(serverSpot, "pike", 5.0)

	clientSpot := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "lake")
	clientFS.SpotAddFishType(clientSpot, "bass", 10.0)
	clientFS.SpotAddFishType(clientSpot, "trout", 8.0)
	clientFS.SpotAddFishType(clientSpot, "pike", 5.0)

	// Both players have identical fishing skills
	serverFisher := NewFishingComponent()
	serverFisher.FishingSkill = 30
	serverFisher.CastDistance = 0.6

	clientFisher := NewFishingComponent()
	clientFisher.FishingSkill = 30
	clientFisher.CastDistance = 0.6

	// Perform multiple fishing attempts - results should be identical
	for attempt := 0; attempt < 20; attempt++ {
		serverFish, serverWeight := serverFS.selectFish(serverSpot, serverFisher)
		clientFish, clientWeight := clientFS.selectFish(clientSpot, clientFisher)

		// Verify fish selection is synchronized
		if serverFish == nil && clientFish == nil {
			continue // Both nil is acceptable
		}

		if serverFish == nil || clientFish == nil {
			t.Fatalf("attempt %d: server/client desync - one caught fish, other didn't (server=%v, client=%v)",
				attempt, serverFish, clientFish)
		}

		if serverFish.ID != clientFish.ID {
			t.Errorf("attempt %d: caught different fish (server=%s, client=%s)",
				attempt, serverFish.ID, clientFish.ID)
		}

		// Verify weight is synchronized (critical for consistent game state)
		if serverWeight != clientWeight {
			t.Errorf("attempt %d: weight mismatch (server=%f, client=%f)",
				attempt, serverWeight, clientWeight)
		}
	}
}

// TestFishingSystem_SaveLoadReproducibility demonstrates that fishing results
// are reproducible after save/load when using the same seed.
func TestFishingSystem_SaveLoadReproducibility(t *testing.T) {
	seed := int64(1234567890)

	// Original game session
	world1 := NewWorld()
	fs1 := NewFishingSystem(world1, seed)

	spot1 := NewFishingSpotComponent(WaterTypeSaltwater, DepthDeep, "ocean")
	fs1.SpotAddFishType(spot1, "tuna", 10.0)
	fs1.SpotAddFishType(spot1, "mackerel", 15.0)
	fs1.SpotAddFishType(spot1, "marlin", 3.0)

	fisher1 := NewFishingComponent()
	fisher1.FishingSkill = 50
	fisher1.CastDistance = 0.8

	// Catch some fish
	var originalCatches []string
	var originalWeights []float64
	for i := 0; i < 10; i++ {
		fish, weight := fs1.selectFish(spot1, fisher1)
		if fish != nil {
			originalCatches = append(originalCatches, fish.ID)
			originalWeights = append(originalWeights, weight)
		}
	}

	// Simulate save/load by creating new system with same seed
	world2 := NewWorld()
	fs2 := NewFishingSystem(world2, seed)

	spot2 := NewFishingSpotComponent(WaterTypeSaltwater, DepthDeep, "ocean")
	fs2.SpotAddFishType(spot2, "tuna", 10.0)
	fs2.SpotAddFishType(spot2, "mackerel", 15.0)
	fs2.SpotAddFishType(spot2, "marlin", 3.0)

	fisher2 := NewFishingComponent()
	fisher2.FishingSkill = 50
	fisher2.CastDistance = 0.8

	// Replay should produce identical results
	for i := 0; i < len(originalCatches); i++ {
		fish, weight := fs2.selectFish(spot2, fisher2)

		if fish == nil {
			t.Fatalf("catch %d: got nil fish after reload, expected %s",
				i, originalCatches[i])
		}

		if fish.ID != originalCatches[i] {
			t.Errorf("catch %d: fish mismatch after reload (original=%s, reloaded=%s)",
				i, originalCatches[i], fish.ID)
		}

		if weight != originalWeights[i] {
			t.Errorf("catch %d: weight mismatch after reload (original=%f, reloaded=%f)",
				i, originalWeights[i], weight)
		}
	}
}
