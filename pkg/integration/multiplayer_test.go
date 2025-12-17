package integration

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestMultiplayerDeterministicGeneration verifies that all content generators
// produce identical output when given the same seed across multiple "clients".
// This is critical for multiplayer synchronization.
func TestMultiplayerDeterministicGeneration(t *testing.T) {
	const worldSeed = int64(987654321)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "fantasy",
	}

	// Simulate two separate clients generating the same world
	t.Run("TerrainDeterminism", func(t *testing.T) {
		terrainGen := terrain.NewBSPGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		terrainSeed1 := seedGen1.GetSeed("terrain", 0)
		terrainSeed2 := seedGen2.GetSeed("terrain", 0)

		if terrainSeed1 != terrainSeed2 {
			t.Fatalf("SeedGenerator not deterministic: %d vs %d", terrainSeed1, terrainSeed2)
		}

		terrainParams := params
		terrainParams.Custom = map[string]interface{}{
			"width":  80,
			"height": 60,
		}

		result1, err1 := terrainGen.Generate(terrainSeed1, terrainParams)
		result2, err2 := terrainGen.Generate(terrainSeed2, terrainParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Terrain generation failed: err1=%v, err2=%v", err1, err2)
		}

		terrain1 := result1.(*terrain.Terrain)
		terrain2 := result2.(*terrain.Terrain)

		// Verify dimensions match
		if terrain1.Width != terrain2.Width || terrain1.Height != terrain2.Height {
			t.Errorf("Terrain dimensions mismatch: (%d,%d) vs (%d,%d)",
				terrain1.Width, terrain1.Height, terrain2.Width, terrain2.Height)
		}

		// Verify tiles match
		mismatchCount := 0
		for y := 0; y < terrain1.Height; y++ {
			for x := 0; x < terrain1.Width; x++ {
				tile1 := terrain1.GetTile(x, y)
				tile2 := terrain2.GetTile(x, y)
				if tile1 != tile2 {
					mismatchCount++
				}
			}
		}

		if mismatchCount > 0 {
			t.Errorf("Terrain has %d tile mismatches", mismatchCount)
		}

		t.Logf("✓ Terrain generation is deterministic (%dx%d, 0 mismatches)", terrain1.Width, terrain1.Height)
	})

	t.Run("EntityDeterminism", func(t *testing.T) {
		entityGen := entity.NewEntityGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		entitySeed1 := seedGen1.GetSeed("entity", 0)
		entitySeed2 := seedGen2.GetSeed("entity", 0)

		entityParams := params
		entityParams.Custom = map[string]interface{}{"count": 10}

		result1, err1 := entityGen.Generate(entitySeed1, entityParams)
		result2, err2 := entityGen.Generate(entitySeed2, entityParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Entity generation failed: err1=%v, err2=%v", err1, err2)
		}

		entities1 := result1.([]*entity.Entity)
		entities2 := result2.([]*entity.Entity)

		if len(entities1) != len(entities2) {
			t.Fatalf("Entity count mismatch: %d vs %d", len(entities1), len(entities2))
		}

		for i := range entities1 {
			e1, e2 := entities1[i], entities2[i]
			if e1.Name != e2.Name || e1.Type != e2.Type || e1.Stats.Level != e2.Stats.Level {
				t.Errorf("Entity %d mismatch: (%s,%s,%d) vs (%s,%s,%d)",
					i, e1.Name, e1.Type, e1.Stats.Level, e2.Name, e2.Type, e2.Stats.Level)
			}
			if e1.Stats.Health != e2.Stats.Health || e1.Stats.Damage != e2.Stats.Damage {
				t.Errorf("Entity %d stats mismatch: health(%d vs %d) damage(%d vs %d)",
					i, e1.Stats.Health, e2.Stats.Health, e1.Stats.Damage, e2.Stats.Damage)
			}
		}

		t.Logf("✓ Entity generation is deterministic (%d entities)", len(entities1))
	})

	t.Run("ItemDeterminism", func(t *testing.T) {
		itemGen := item.NewItemGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		itemSeed1 := seedGen1.GetSeed("item", 0)
		itemSeed2 := seedGen2.GetSeed("item", 0)

		itemParams := params
		itemParams.Custom = map[string]interface{}{"count": 15}

		result1, err1 := itemGen.Generate(itemSeed1, itemParams)
		result2, err2 := itemGen.Generate(itemSeed2, itemParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Item generation failed: err1=%v, err2=%v", err1, err2)
		}

		items1 := result1.([]*item.Item)
		items2 := result2.([]*item.Item)

		if len(items1) != len(items2) {
			t.Fatalf("Item count mismatch: %d vs %d", len(items1), len(items2))
		}

		for i := range items1 {
			it1, it2 := items1[i], items2[i]
			if it1.Name != it2.Name || it1.Type != it2.Type || it1.Rarity != it2.Rarity {
				t.Errorf("Item %d mismatch: (%s,%s,%s) vs (%s,%s,%s)",
					i, it1.Name, it1.Type, it1.Rarity, it2.Name, it2.Type, it2.Rarity)
			}
			// Critical: descriptions must match (multiplayer sync requirement)
			if it1.Description != it2.Description {
				t.Errorf("Item %d description mismatch:\n  Client1: %s\n  Client2: %s",
					i, it1.Description, it2.Description)
			}
		}

		t.Logf("✓ Item generation is deterministic (%d items)", len(items1))
	})

	t.Run("MagicDeterminism", func(t *testing.T) {
		magicGen := magic.NewSpellGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		magicSeed1 := seedGen1.GetSeed("magic", 0)
		magicSeed2 := seedGen2.GetSeed("magic", 0)

		magicParams := params
		magicParams.Custom = map[string]interface{}{"count": 8}

		result1, err1 := magicGen.Generate(magicSeed1, magicParams)
		result2, err2 := magicGen.Generate(magicSeed2, magicParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Magic generation failed: err1=%v, err2=%v", err1, err2)
		}

		spells1 := result1.([]*magic.Spell)
		spells2 := result2.([]*magic.Spell)

		if len(spells1) != len(spells2) {
			t.Fatalf("Spell count mismatch: %d vs %d", len(spells1), len(spells2))
		}

		for i := range spells1 {
			s1, s2 := spells1[i], spells2[i]
			if s1.Name != s2.Name || s1.Element != s2.Element || s1.Stats.ManaCost != s2.Stats.ManaCost {
				t.Errorf("Spell %d mismatch: (%s,%s,%d) vs (%s,%s,%d)",
					i, s1.Name, s1.Element, s1.Stats.ManaCost, s2.Name, s2.Element, s2.Stats.ManaCost)
			}
		}

		t.Logf("✓ Magic generation is deterministic (%d spells)", len(spells1))
	})

	t.Run("SkillsDeterminism", func(t *testing.T) {
		skillsGen := skills.NewSkillTreeGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		skillsSeed1 := seedGen1.GetSeed("skills", 0)
		skillsSeed2 := seedGen2.GetSeed("skills", 0)

		result1, err1 := skillsGen.Generate(skillsSeed1, params)
		result2, err2 := skillsGen.Generate(skillsSeed2, params)

		if err1 != nil || err2 != nil {
			t.Fatalf("Skills generation failed: err1=%v, err2=%v", err1, err2)
		}

		trees1 := result1.([]*skills.SkillTree)
		trees2 := result2.([]*skills.SkillTree)

		if len(trees1) != len(trees2) {
			t.Fatalf("Skill tree count mismatch: %d vs %d", len(trees1), len(trees2))
		}

		// Compare first tree as representative
		if len(trees1) > 0 {
			tree1, tree2 := trees1[0], trees2[0]
			if len(tree1.Nodes) != len(tree2.Nodes) {
				t.Errorf("Skill node count mismatch: %d vs %d", len(tree1.Nodes), len(tree2.Nodes))
			}

			for i := 0; i < len(tree1.Nodes) && i < len(tree2.Nodes); i++ {
				sk1, sk2 := tree1.Nodes[i], tree2.Nodes[i]
				if sk1.Skill.Name != sk2.Skill.Name || sk1.Skill.Tier != sk2.Skill.Tier {
					t.Errorf("Skill %d mismatch: (%s,%d) vs (%s,%d)",
						i, sk1.Skill.Name, sk1.Skill.Tier, sk2.Skill.Name, sk2.Skill.Tier)
				}
			}
		}

		t.Logf("✓ Skills generation is deterministic (%d trees)", len(trees1))
	})

	t.Run("QuestDeterminism", func(t *testing.T) {
		questGen := quest.NewQuestGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		questSeed1 := seedGen1.GetSeed("quest", 0)
		questSeed2 := seedGen2.GetSeed("quest", 0)

		questParams := params
		questParams.Custom = map[string]interface{}{"count": 5}

		result1, err1 := questGen.Generate(questSeed1, questParams)
		result2, err2 := questGen.Generate(questSeed2, questParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Quest generation failed: err1=%v, err2=%v", err1, err2)
		}

		quests1 := result1.([]*quest.Quest)
		quests2 := result2.([]*quest.Quest)

		if len(quests1) != len(quests2) {
			t.Fatalf("Quest count mismatch: %d vs %d", len(quests1), len(quests2))
		}

		for i := range quests1 {
			q1, q2 := quests1[i], quests2[i]
			if q1.Name != q2.Name || q1.Type != q2.Type || q1.Difficulty != q2.Difficulty {
				t.Errorf("Quest %d mismatch: (%s,%s,%d) vs (%s,%s,%d)",
					i, q1.Name, q1.Type, q1.Difficulty, q2.Name, q2.Type, q2.Difficulty)
			}
		}

		t.Logf("✓ Quest generation is deterministic (%d quests)", len(quests1))
	})

	t.Run("StationDeterminism", func(t *testing.T) {
		stationGen := station.NewStationGenerator()
		seedGen1 := procgen.NewSeedGenerator(worldSeed)
		seedGen2 := procgen.NewSeedGenerator(worldSeed)

		stationSeed1 := seedGen1.GetSeed("station", 0)
		stationSeed2 := seedGen2.GetSeed("station", 0)

		stationParams := params
		stationParams.Custom = map[string]interface{}{"count": 3}

		result1, err1 := stationGen.Generate(stationSeed1, stationParams)
		result2, err2 := stationGen.Generate(stationSeed2, stationParams)

		if err1 != nil || err2 != nil {
			t.Fatalf("Station generation failed: err1=%v, err2=%v", err1, err2)
		}

		stations1 := result1.([]*station.StationData)
		stations2 := result2.([]*station.StationData)

		if len(stations1) != len(stations2) {
			t.Fatalf("Station count mismatch: %d vs %d", len(stations1), len(stations2))
		}

		for i := range stations1 {
			st1, st2 := stations1[i], stations2[i]
			if st1.Name != st2.Name || st1.StationType != st2.StationType || st1.Seed != st2.Seed {
				t.Errorf("Station %d mismatch: (%s,%v,%d) vs (%s,%v,%d)",
					i, st1.Name, st1.StationType, st1.Seed, st2.Name, st2.StationType, st2.Seed)
			}
		}

		t.Logf("✓ Station generation is deterministic (%d stations)", len(stations1))
	})
}

// TestMultiplayerCrossGenreDeterminism verifies deterministic generation
// works across all supported genres. Different clients might use different
// genres but must still generate deterministically.
func TestMultiplayerCrossGenreDeterminism(t *testing.T) {
	const worldSeed = int64(111222333)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run("Genre_"+genreID, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    genreID,
				Custom: map[string]interface{}{
					"width":  60,
					"height": 40,
				},
			}

			// Generate terrain twice with same seed
			terrainGen := terrain.NewBSPGenerator()
			seedGen := procgen.NewSeedGenerator(worldSeed)
			terrainSeed := seedGen.GetSeed("terrain", 0)

			result1, err1 := terrainGen.Generate(terrainSeed, params)
			result2, err2 := terrainGen.Generate(terrainSeed, params)

			if err1 != nil || err2 != nil {
				t.Fatalf("Terrain generation failed for %s: err1=%v, err2=%v", genreID, err1, err2)
			}

			terrain1 := result1.(*terrain.Terrain)
			terrain2 := result2.(*terrain.Terrain)

			// Verify identical terrain
			if terrain1.Width != terrain2.Width || terrain1.Height != terrain2.Height {
				t.Errorf("Genre %s: terrain dimension mismatch", genreID)
			}

			mismatchCount := 0
			for y := 0; y < terrain1.Height; y++ {
				for x := 0; x < terrain1.Width; x++ {
					if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
						mismatchCount++
					}
				}
			}

			if mismatchCount > 0 {
				t.Errorf("Genre %s: %d tile mismatches", genreID, mismatchCount)
			}

			t.Logf("✓ Genre %s: terrain deterministic", genreID)
		})
	}
}

// TestMultiplayerDifferentSeeds verifies that different seeds produce
// different content (not stuck on same output). Important for variety.
func TestMultiplayerDifferentSeeds(t *testing.T) {
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
	}

	itemGen := item.NewItemGenerator()

	// Generate items with different seeds
	result1, err1 := itemGen.Generate(12345, params)
	if err1 != nil {
		t.Fatalf("Item generation failed for seed 12345: %v", err1)
	}
	result2, err2 := itemGen.Generate(67890, params)
	if err2 != nil {
		t.Fatalf("Item generation failed for seed 67890: %v", err2)
	}

	items1 := result1.([]*item.Item)
	items2 := result2.([]*item.Item)

	// Should have different names/types
	allSame := true
	for i := range items1 {
		if i >= len(items2) {
			break
		}
		if items1[i].Name != items2[i].Name || items1[i].Type != items2[i].Type {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Different seeds produced identical items - no variety!")
	}

	t.Logf("✓ Different seeds produce different content (variety confirmed)")
}

// TestMultiplayerSeedIndependence verifies that different content types
// get different seeds even with same base seed. This prevents all content
// from being identical across types.
func TestMultiplayerSeedIndependence(t *testing.T) {
	const baseSeed = int64(555666777)
	seedGen := procgen.NewSeedGenerator(baseSeed)

	terrainSeed := seedGen.GetSeed("terrain", 0)
	entitySeed := seedGen.GetSeed("entity", 0)
	itemSeed := seedGen.GetSeed("item", 0)
	magicSeed := seedGen.GetSeed("magic", 0)

	// All seeds should be different
	seeds := []int64{terrainSeed, entitySeed, itemSeed, magicSeed}
	for i := 0; i < len(seeds); i++ {
		for j := i + 1; j < len(seeds); j++ {
			if seeds[i] == seeds[j] {
				t.Errorf("Seed collision: category %d and %d have same seed %d", i, j, seeds[i])
			}
		}
	}

	// Different indices should also produce different seeds
	seed0 := seedGen.GetSeed("terrain", 0)
	seed1 := seedGen.GetSeed("terrain", 1)
	seed2 := seedGen.GetSeed("terrain", 2)

	if seed0 == seed1 || seed0 == seed2 || seed1 == seed2 {
		t.Error("Same category with different indices produced identical seeds")
	}

	t.Logf("✓ Seed independence verified (no collisions)")
}
