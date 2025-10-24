//go:build test
// +build test

package item

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestItemDescriptionDeterminism verifies that item descriptions are deterministic.
// This test ensures the fix for non-deterministic rand.Intn() usage in generateDescription.
func TestItemDescriptionDeterminism(t *testing.T) {
	gen := NewItemGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
			"type":  "weapon",
		},
	}

	// Generate items twice with same seed
	result1, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	result2, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	items1 := result1.([]*Item)
	items2 := result2.([]*Item)

	// Verify same number of items
	if len(items1) != len(items2) {
		t.Fatalf("Different number of items: %d vs %d", len(items1), len(items2))
	}

	// Verify all properties match, including descriptions
	for i := range items1 {
		item1 := items1[i]
		item2 := items2[i]

		if item1.Name != item2.Name {
			t.Errorf("Item %d name mismatch: %s vs %s", i, item1.Name, item2.Name)
		}

		if item1.Type != item2.Type {
			t.Errorf("Item %d type mismatch: %s vs %s", i, item1.Type, item2.Type)
		}

		if item1.Rarity != item2.Rarity {
			t.Errorf("Item %d rarity mismatch: %s vs %s", i, item1.Rarity, item2.Rarity)
		}

		// CRITICAL: Descriptions must match (this was the bug)
		if item1.Description != item2.Description {
			t.Errorf("Item %d description mismatch:\n  First:  %s\n  Second: %s",
				i, item1.Description, item2.Description)
		}

		// Verify stats match
		if item1.Stats.Damage != item2.Stats.Damage {
			t.Errorf("Item %d damage mismatch: %d vs %d", i, item1.Stats.Damage, item2.Stats.Damage)
		}

		if item1.Stats.Defense != item2.Stats.Defense {
			t.Errorf("Item %d defense mismatch: %d vs %d", i, item1.Stats.Defense, item2.Stats.Defense)
		}

		if item1.Stats.Value != item2.Stats.Value {
			t.Errorf("Item %d value mismatch: %d vs %d", i, item1.Stats.Value, item2.Stats.Value)
		}
	}

	t.Logf("✓ Successfully verified determinism for %d items", len(items1))
}

// TestItemDescriptionDeterminismAcrossGenres ensures descriptions are deterministic
// across different genres with same seed.
func TestItemDescriptionDeterminismAcrossGenres(t *testing.T) {
	gen := NewItemGenerator()
	seed := int64(99999)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		params := procgen.GenerationParams{
			Difficulty: 0.7,
			Depth:      10,
			GenreID:    genre,
			Custom: map[string]interface{}{
				"count": 5,
			},
		}

		// Generate twice with same seed for this genre
		result1, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Generation failed for genre %s: %v", genre, err)
		}

		result2, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Second generation failed for genre %s: %v", genre, err)
		}

		items1 := result1.([]*Item)
		items2 := result2.([]*Item)

		// Verify descriptions match for this genre
		for i := range items1 {
			if items1[i].Description != items2[i].Description {
				t.Errorf("Genre %s, Item %d: Description mismatch:\n  First:  %s\n  Second: %s",
					genre, i, items1[i].Description, items2[i].Description)
			}
		}

		t.Logf("✓ Genre %s: Descriptions are deterministic", genre)
	}
}

// TestItemDescriptionVariety ensures different seeds produce different descriptions
// (not stuck on same description).
func TestItemDescriptionVariety(t *testing.T) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "weapon",
			"rarity": RarityLegendary, // High rarity to get rarity-specific descriptions
		},
	}

	descriptionCounts := make(map[string]int)

	// Generate items with different seeds
	for seed := int64(1); seed <= 50; seed++ {
		result, err := gen.Generate(seed, params)
		if err != nil {
			t.Fatalf("Generation failed for seed %d: %v", seed, err)
		}

		items := result.([]*Item)
		if len(items) > 0 {
			description := items[0].Description
			descriptionCounts[description]++
		}
	}

	// Should see variety in descriptions (at least 2 different ones)
	uniqueDescriptions := len(descriptionCounts)
	if uniqueDescriptions < 2 {
		t.Errorf("Not enough variety in descriptions. Only %d unique descriptions found", uniqueDescriptions)
		for desc, count := range descriptionCounts {
			t.Logf("  %dx: %s", count, desc)
		}
	}

	t.Logf("✓ Found %d unique descriptions across 50 seeds", uniqueDescriptions)
	for desc, count := range descriptionCounts {
		t.Logf("  %dx: %s", count, desc)
	}
}
