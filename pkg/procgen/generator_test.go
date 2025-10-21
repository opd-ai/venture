package procgen

import "testing"

func TestSeedGenerator(t *testing.T) {
	sg := NewSeedGenerator(12345)
	
	// Test deterministic seed generation
	seed1 := sg.GetSeed("terrain", 0)
	seed2 := sg.GetSeed("terrain", 0)
	
	if seed1 != seed2 {
		t.Error("Expected deterministic seed generation")
	}
	
	// Test different categories produce different seeds
	terrainSeed := sg.GetSeed("terrain", 0)
	entitySeed := sg.GetSeed("entity", 0)
	
	if terrainSeed == entitySeed {
		t.Error("Expected different seeds for different categories")
	}
	
	// Test different indices produce different seeds
	index0 := sg.GetSeed("terrain", 0)
	index1 := sg.GetSeed("terrain", 1)
	
	if index0 == index1 {
		t.Error("Expected different seeds for different indices")
	}
}

func TestGenerationParams(t *testing.T) {
	params := GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}
	
	params.Custom["test"] = "value"
	
	if params.Custom["test"] != "value" {
		t.Error("Custom parameters not working")
	}
}
