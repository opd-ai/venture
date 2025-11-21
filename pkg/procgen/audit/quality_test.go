package audit

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
)

// TestQualityThresholds_AllGenerators validates that all generators meet minimum quality standards
// This implements Phase 62.2: Quality Threshold Validation
func TestQualityThresholds_AllGenerators(t *testing.T) {
	const samples = 1000 // Test 1000 samples per generator

	generators := []struct {
		name      string
		generator procgen.Generator
		validator QualityValidator
	}{
		{"Terrain", terrain.NewBSPGenerator(), &TerrainQualityValidator{}},
		{"Entity", entity.NewEntityGenerator(), &EntityQualityValidator{}},
		{"Item", item.NewItemGenerator(), &ItemQualityValidator{}},
		{"Magic", magic.NewSpellGenerator(), &MagicQualityValidator{}},
		{"Quest", quest.NewQuestGenerator(), &QuestQualityValidator{}},
		{"Recipe", recipe.NewRecipeGenerator(), &RecipeQualityValidator{}},
		{"Station", station.NewStationGenerator(), &StationQualityValidator{}},
		{"Vehicle", vehicle.NewVehicleGenerator(), &VehicleQualityValidator{}},
		{"Companion", companion.NewGenerator(), &CompanionQualityValidator{}},
		{"Building", building.NewGenerator(), &BuildingQualityValidator{}},
		{"Furniture", furniture.NewGenerator(), &FurnitureQualityValidator{}},
		{"Legendary", legendary.NewLegendaryQuestGenerator(), &LegendaryQualityValidator{}},
		{"Skills", skills.NewSkillTreeGenerator(), &SkillsQualityValidator{}},
	}

	for _, tc := range generators {
		t.Run(tc.name, func(t *testing.T) {
			passCount := 0
			failCount := 0
			var firstFailure error

			for i := 0; i < samples; i++ {
				seed := int64(1000 + i)
				params := procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      5,
					GenreID:    "fantasy",
				}

				result, err := tc.generator.Generate(seed, params)
				if err != nil {
					t.Fatalf("Generation failed for %s (seed %d): %v", tc.name, seed, err)
				}

				// Run built-in validator
				if err := tc.generator.Validate(result); err != nil {
					failCount++
					if firstFailure == nil {
						firstFailure = fmt.Errorf("built-in validation failed (seed %d): %w", seed, err)
					}
					continue
				}

				// Run quality-specific validator
				if err := tc.validator.Validate(result, params); err != nil {
					failCount++
					if firstFailure == nil {
						firstFailure = fmt.Errorf("quality validation failed (seed %d): %w", seed, err)
					}
					continue
				}

				passCount++
			}

			passRate := float64(passCount) / float64(samples) * 100.0
			t.Logf("%s: %d/%d passed (%.1f%%)", tc.name, passCount, samples, passRate)

			// Acceptance: ≥99% pass rate
			if passRate < 99.0 {
				t.Errorf("%s quality validation failed: %.1f%% pass rate (expected ≥99%%)", tc.name, passRate)
				if firstFailure != nil {
					t.Errorf("First failure: %v", firstFailure)
				}
			}
		})
	}
}

// QualityValidator interface for generator-specific quality checks
type QualityValidator interface {
	Validate(result interface{}, params procgen.GenerationParams) error
}

// TerrainQualityValidator validates terrain generation quality
type TerrainQualityValidator struct{}

func (v *TerrainQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	t, ok := result.(*terrain.Terrain)
	if !ok {
		return fmt.Errorf("expected *terrain.Terrain, got %T", result)
	}

	// Check minimum walkable tiles (≥25% for dungeons - they're meant to be mostly walls)
	totalTiles := t.Width * t.Height
	walkableCount := 0
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if t.Tiles[y][x].IsWalkableTile() {
				walkableCount++
			}
		}
	}
	walkablePercent := float64(walkableCount) / float64(totalTiles) * 100.0
	if walkablePercent < 25.0 {
		return fmt.Errorf("insufficient walkable tiles: %.1f%% (expected ≥25%%)", walkablePercent)
	}

	// Check minimum rooms (≥3)
	if len(t.Rooms) < 3 {
		return fmt.Errorf("insufficient rooms: %d (expected ≥3)", len(t.Rooms))
	}

	// Room connectivity check: all rooms should be connected via corridors/doors
	// For BSP generator, rooms are implicitly connected via the recursive split structure
	// We'll validate by checking that corridors exist
	corridorCount := 0
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if t.Tiles[y][x] == terrain.TileCorridor || t.Tiles[y][x] == terrain.TileDoor {
				corridorCount++
			}
		}
	}
	if corridorCount == 0 && len(t.Rooms) > 1 {
		return fmt.Errorf("multiple rooms but no corridors/doors found")
	}

	return nil
}

// EntityQualityValidator validates entity generation quality
type EntityQualityValidator struct{}

func (v *EntityQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	// Entity generator may return a single entity or a slice
	var e *entity.Entity

	switch v := result.(type) {
	case *entity.Entity:
		e = v
	case []*entity.Entity:
		if len(v) == 0 {
			return fmt.Errorf("no entities generated")
		}
		e = v[0] // Validate first entity as a sample
	default:
		return fmt.Errorf("expected *entity.Entity or []*entity.Entity, got %T", result)
	}

	// Check stats are within expected range (0.8-1.2x for depth/rarity)
	expectedHealth := 100 + (params.Depth * 20) // Base formula
	rarityMultiplier := getRarityMultiplier(e.Rarity)
	expectedMin := float64(expectedHealth) * rarityMultiplier * 0.8
	expectedMax := float64(expectedHealth) * rarityMultiplier * 1.2

	if float64(e.Stats.Health) < expectedMin || float64(e.Stats.Health) > expectedMax {
		return fmt.Errorf("health out of range: %d (expected %.0f-%.0f)", e.Stats.Health, expectedMin, expectedMax)
	}

	// Check no negative values
	if e.Stats.Health < 0 || e.Stats.Damage < 0 || e.Stats.Defense < 0 || e.Stats.Speed < 0 {
		return fmt.Errorf("negative stat values detected: H=%d D=%d Def=%d S=%.1f",
			e.Stats.Health, e.Stats.Damage, e.Stats.Defense, e.Stats.Speed)
	}

	// Check name is non-empty
	if e.Name == "" {
		return fmt.Errorf("empty entity name")
	}

	return nil
}

// getRarityMultiplier returns the stat multiplier for a rarity level
func getRarityMultiplier(r entity.Rarity) float64 {
	switch r {
	case entity.RarityCommon:
		return 1.0
	case entity.RarityUncommon:
		return 1.2
	case entity.RarityRare:
		return 1.5
	case entity.RarityEpic:
		return 2.0
	case entity.RarityLegendary:
		return 3.0
	default:
		return 1.0
	}
}

// ItemQualityValidator validates item generation quality
type ItemQualityValidator struct{}

func (v *ItemQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	var i *item.Item

	switch v := result.(type) {
	case *item.Item:
		i = v
	case []*item.Item:
		if len(v) == 0 {
			return fmt.Errorf("no items generated")
		}
		i = v[0] // Validate first item as a sample
	default:
		return fmt.Errorf("expected *item.Item or []*item.Item, got %T", result)
	}

	// Check stat balance (damage roughly equals defense * 0.8)
	if i.Stats.Damage > 0 && i.Stats.Defense > 0 {
		ratio := float64(i.Stats.Damage) / float64(i.Stats.Defense)
		// Allow 0.5-1.5 ratio (generous tolerance)
		if ratio < 0.5 || ratio > 1.5 {
			return fmt.Errorf("stat imbalance: damage/defense ratio %.2f (expected 0.5-1.5)", ratio)
		}
	}

	// Check no negative values
	if i.Stats.Damage < 0 || i.Stats.Defense < 0 || i.Stats.Value < 0 {
		return fmt.Errorf("negative stat values: damage=%d defense=%d value=%d",
			i.Stats.Damage, i.Stats.Defense, i.Stats.Value)
	}

	// Check name is non-empty
	if i.Name == "" {
		return fmt.Errorf("empty item name")
	}

	return nil
}

// MagicQualityValidator validates magic spell generation quality
type MagicQualityValidator struct{}

func (v *MagicQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	var s *magic.Spell

	switch v := result.(type) {
	case *magic.Spell:
		s = v
	case []*magic.Spell:
		if len(v) == 0 {
			return fmt.Errorf("no spells generated")
		}
		s = v[0]
	default:
		return fmt.Errorf("expected *magic.Spell or []*magic.Spell, got %T", result)
	}

	// Check mana cost scales with damage (cost ~= damage * 10, approximately)
	if s.Stats.Damage > 0 {
		expectedCost := s.Stats.Damage * 10
		tolerance := 20 // Allow ±20 mana
		if s.Stats.ManaCost < expectedCost-tolerance || s.Stats.ManaCost > expectedCost+tolerance {
			return fmt.Errorf("mana cost mismatch: %d (expected ~%d for damage %d)",
				s.Stats.ManaCost, expectedCost, s.Stats.Damage)
		}
	}

	// Check cooldown ≥ cast time
	if s.Stats.Cooldown < s.Stats.CastTime {
		return fmt.Errorf("cooldown < cast time: %.2fs < %.2fs", s.Stats.Cooldown, s.Stats.CastTime)
	}

	// Check no negative values
	if s.Stats.Damage < 0 || s.Stats.ManaCost < 0 || s.Stats.Healing < 0 {
		return fmt.Errorf("negative values: damage=%d mana=%d healing=%d",
			s.Stats.Damage, s.Stats.ManaCost, s.Stats.Healing)
	}

	// Check name is non-empty
	if s.Name == "" {
		return fmt.Errorf("empty spell name")
	}

	return nil
}

// QuestQualityValidator validates quest generation quality
type QuestQualityValidator struct{}

func (v *QuestQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	var q *quest.Quest

	switch v := result.(type) {
	case *quest.Quest:
		q = v
	case []*quest.Quest:
		if len(v) == 0 {
			return fmt.Errorf("no quests generated")
		}
		q = v[0]
	default:
		return fmt.Errorf("expected *quest.Quest or []*quest.Quest, got %T", result)
	}

	// Check minimum objectives (≥3)
	if len(q.Objectives) < 3 {
		return fmt.Errorf("insufficient objectives: %d (expected ≥3)", len(q.Objectives))
	}

	// Check rewards scale with difficulty
	expectedRewardMin := int(params.Difficulty * 100)
	if q.Reward.Gold < expectedRewardMin {
		return fmt.Errorf("insufficient reward: %d gold (expected ≥%d for difficulty %.1f)",
			q.Reward.Gold, expectedRewardMin, params.Difficulty)
	}

	// Check name and description are non-empty
	if q.Name == "" || q.Description == "" {
		return fmt.Errorf("empty quest name or description")
	}

	return nil
}

// RecipeQualityValidator validates crafting recipe quality
type RecipeQualityValidator struct{}

func (v *RecipeQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	// Recipe generator returns []*engine.Recipe (a slice)
	recipes, ok := result.([]*engine.Recipe)
	if !ok {
		return fmt.Errorf("expected []*engine.Recipe, got %T", result)
	}

	if len(recipes) == 0 {
		return fmt.Errorf("no recipes generated")
	}

	// Validate first recipe as a sample
	r := recipes[0]

	// Check minimum inputs (≥1)
	if len(r.Materials) < 1 {
		return fmt.Errorf("no input materials (expected ≥1)")
	}

	// Check name is non-empty
	if r.Name == "" {
		return fmt.Errorf("empty recipe name")
	}

	// Check success chance is valid (0-1)
	if r.BaseSuccessChance < 0.0 || r.BaseSuccessChance > 1.0 {
		return fmt.Errorf("invalid success chance: %.2f (expected 0.0-1.0)", r.BaseSuccessChance)
	}

	return nil
}

// StationQualityValidator validates crafting station quality
type StationQualityValidator struct{}

func (v *StationQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	// Station generator returns []*station.StationData (a slice)
	stations, ok := result.([]*station.StationData)
	if !ok {
		return fmt.Errorf("expected []*station.StationData, got %T", result)
	}

	if len(stations) == 0 {
		return fmt.Errorf("no stations generated")
	}

	// Validate first station as a sample
	s := stations[0]

	// Check name is non-empty
	if s.Name == "" {
		return fmt.Errorf("empty station name")
	}

	// Check station type is valid (0-2 for the 3 station types)
	if s.StationType < station.StationAlchemyTable || s.StationType > station.StationWorkbench {
		return fmt.Errorf("invalid station type: %d", s.StationType)
	}

	return nil
}

// VehicleQualityValidator validates vehicle generation quality
type VehicleQualityValidator struct{}

func (v *VehicleQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	var veh *vehicle.Vehicle

	switch v := result.(type) {
	case *vehicle.Vehicle:
		veh = v
	case []*vehicle.Vehicle:
		if len(v) == 0 {
			return fmt.Errorf("no vehicles generated")
		}
		veh = v[0]
	default:
		return fmt.Errorf("expected *vehicle.Vehicle or []*vehicle.Vehicle, got %T", result)
	}

	// Check speed/handling/durability trade-offs are balanced
	// Total stat points should be roughly consistent
	totalStats := int(veh.MaxSpeed + veh.Handling + veh.MaxDurability)
	expectedMin := 150 // Adjust based on actual implementation
	expectedMax := 400
	if totalStats < expectedMin || totalStats > expectedMax {
		return fmt.Errorf("unbalanced stats: total=%d (expected %d-%d)", totalStats, expectedMin, expectedMax)
	}

	// Check no negative values
	if veh.MaxSpeed < 0 || veh.Handling < 0 || veh.MaxDurability < 0 {
		return fmt.Errorf("negative stats: maxSpeed=%.1f handling=%.1f maxDurability=%.1f",
			veh.MaxSpeed, veh.Handling, veh.MaxDurability)
	}

	// Check name is non-empty
	if veh.Name == "" {
		return fmt.Errorf("empty vehicle name")
	}

	return nil
}

// CompanionQualityValidator validates companion generation quality
type CompanionQualityValidator struct{}

func (v *CompanionQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	c, ok := result.(*companion.Companion)
	if !ok {
		return fmt.Errorf("expected *companion.Companion, got %T", result)
	}

	// Check stats are reasonable
	if c.HP <= 0 || c.Attack < 0 || c.Defense < 0 {
		return fmt.Errorf("invalid stats: hp=%.1f attack=%.1f defense=%.1f", c.HP, c.Attack, c.Defense)
	}

	// Check loyalty is in valid range [0.0, 100.0] (0-100 scale, not 0-1)
	if c.Loyalty < 0.0 || c.Loyalty > 100.0 {
		return fmt.Errorf("loyalty out of range: %.2f (expected 0.0-100.0)", c.Loyalty)
	}

	// Check name is non-empty
	if c.Name == "" {
		return fmt.Errorf("empty companion name")
	}

	return nil
}

// BuildingQualityValidator validates building generation quality
type BuildingQualityValidator struct{}

func (v *BuildingQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	b, ok := result.(*building.Building)
	if !ok {
		return fmt.Errorf("expected *building.Building, got %T", result)
	}

	// Check minimum rooms (≥3)
	if len(b.Rooms) < 3 {
		return fmt.Errorf("insufficient rooms: %d (expected ≥3)", len(b.Rooms))
	}

	// Check all rooms are connected (simplified check)
	// A more robust implementation would use BFS connectivity checks
	if len(b.Doors) == 0 && len(b.Rooms) > 1 {
		return fmt.Errorf("multiple rooms but no doors found")
	}

	// Check valid floor plan (width/height > 0)
	if b.Width <= 0 || b.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", b.Width, b.Height)
	}

	// Check floors count is reasonable (1-5)
	if b.Floors < 1 || b.Floors > 5 {
		return fmt.Errorf("invalid floor count: %d (expected 1-5)", b.Floors)
	}

	return nil
}

// FurnitureQualityValidator validates furniture generation quality
type FurnitureQualityValidator struct{}

func (v *FurnitureQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	f, ok := result.(*furniture.Furniture)
	if !ok {
		return fmt.Errorf("expected *furniture.Furniture, got %T", result)
	}

	// Check dimensions are positive
	if f.Width <= 0 || f.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %.1fx%.1f", f.Width, f.Height)
	}

	// Check name is non-empty
	if f.Name == "" {
		return fmt.Errorf("empty furniture name")
	}

	// Check color is valid (non-zero alpha for visibility)
	if f.PrimaryColor.A == 0 {
		return fmt.Errorf("invisible furniture (primary color alpha=0)")
	}

	return nil
}

// LegendaryQualityValidator validates legendary quest quality
type LegendaryQualityValidator struct{}

func (v *LegendaryQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	lq, ok := result.(*legendary.LegendaryQuest)
	if !ok {
		return fmt.Errorf("expected *legendary.LegendaryQuest, got %T", result)
	}

	// Check minimum phases (≥5)
	if len(lq.Phases) < 5 {
		return fmt.Errorf("insufficient phases: %d (expected ≥5)", len(lq.Phases))
	}

	// Check duration is reasonable (10-20 hours)
	if lq.EstimatedHours < 10 || lq.EstimatedHours > 20 {
		return fmt.Errorf("duration out of range: %.1f hours (expected 10-20)", lq.EstimatedHours)
	}

	// Check name and description are non-empty
	if lq.Name == "" || lq.Description == "" {
		return fmt.Errorf("empty name or description")
	}

	// Check rewards exist
	if lq.Rewards == nil {
		return fmt.Errorf("no rewards defined")
	}

	return nil
}

// SkillsQualityValidator validates skill tree generation quality
type SkillsQualityValidator struct{}

func (v *SkillsQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	var st *skills.SkillTree

	switch v := result.(type) {
	case *skills.SkillTree:
		st = v
	case []*skills.SkillTree:
		if len(v) == 0 {
			return fmt.Errorf("no skill trees generated")
		}
		st = v[0]
	default:
		return fmt.Errorf("expected *skills.SkillTree or []*skills.SkillTree, got %T", result)
	}

	// Check minimum skills (≥10)
	if len(st.Nodes) < 10 {
		return fmt.Errorf("insufficient skills: %d (expected ≥10)", len(st.Nodes))
	}

	// Check prerequisites are valid (reference existing skills)
	nodeIDs := make(map[string]bool)
	for _, node := range st.Nodes {
		if node.Skill != nil {
			nodeIDs[node.Skill.ID] = true
		}
	}

	for _, node := range st.Nodes {
		if node.Skill != nil && node.Skill.Requirements.PrerequisiteIDs != nil {
			for _, prereqID := range node.Skill.Requirements.PrerequisiteIDs {
				if !nodeIDs[prereqID] {
					return fmt.Errorf("skill %s references invalid prerequisite %s", node.Skill.ID, prereqID)
				}
			}
		}
	}

	// Check name is non-empty
	if st.Name == "" {
		return fmt.Errorf("empty skill tree name")
	}

	return nil
}

// TestRarityDistribution validates that rarity tiers match expected percentages
func TestRarityDistribution(t *testing.T) {
	const samples = 10000
	generators := []struct {
		name      string
		generator procgen.Generator
		extractor RarityExtractor
	}{
		{"Entity", entity.NewEntityGenerator(), &EntityRarityExtractor{}},
		{"Item", item.NewItemGenerator(), &ItemRarityExtractor{}},
	}

	expectedDistribution := map[string]float64{
		"Common":    50.0, // 50%
		"Uncommon":  30.0, // 30%
		"Rare":      15.0, // 15%
		"Epic":      4.0,  // 4%
		"Legendary": 1.0,  // 1%
	}

	for _, tc := range generators {
		t.Run(tc.name, func(t *testing.T) {
			counts := make(map[string]int)

			for i := 0; i < samples; i++ {
				seed := int64(2000 + i)
				params := procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      10, // Higher depth for better rarity chances
					GenreID:    "fantasy",
				}

				result, err := tc.generator.Generate(seed, params)
				if err != nil {
					t.Fatalf("Generation failed (seed %d): %v", seed, err)
				}

				rarity := tc.extractor.Extract(result)
				counts[rarity]++
			}

			// Check distribution matches expected ±5%
			for rarityName, expected := range expectedDistribution {
				actual := float64(counts[rarityName]) / float64(samples) * 100.0
				diff := math.Abs(actual - expected)
				if diff > 5.0 {
					t.Errorf("%s rarity distribution off: %.1f%% (expected %.1f%% ±5%%)", rarityName, actual, expected)
				}
				t.Logf("%s: %.1f%% (expected %.1f%%)", rarityName, actual, expected)
			}
		})
	}
}

// RarityExtractor interface for extracting rarity from generation results
type RarityExtractor interface {
	Extract(result interface{}) string
}

type EntityRarityExtractor struct{}

func (e *EntityRarityExtractor) Extract(result interface{}) string {
	ent := result.(*entity.Entity)
	return ent.Rarity.String()
}

type ItemRarityExtractor struct{}

func (e *ItemRarityExtractor) Extract(result interface{}) string {
	i := result.(*item.Item)
	return i.Rarity.String()
}

// TestGenreDistinctiveness validates that different genres produce distinct content
func TestGenreDistinctiveness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	generator := entity.NewEntityGenerator()

	// Generate samples for each genre
	genreResults := make(map[string][]interface{})
	for _, genre := range genres {
		samples := make([]interface{}, 0, 100)
		for i := 0; i < 100; i++ {
			seed := int64(3000 + i)
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    genre,
			}

			result, err := generator.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generation failed for %s (seed %d): %v", genre, seed, err)
			}
			samples = append(samples, result)
		}
		genreResults[genre] = samples
	}

	// Compare genres pairwise - check that they produce different content
	for i, genre1 := range genres {
		for _, genre2 := range genres[i+1:] {
			// Calculate difference between genre outputs
			// For this test, we'll serialize and compare JSON
			diff := calculateGenreDifference(genreResults[genre1], genreResults[genre2])
			t.Logf("%s vs %s: %.1f%% different", genre1, genre2, diff)

			// Expect at least 30% difference between genres
			if diff < 30.0 {
				t.Errorf("Genres too similar: %s vs %s (%.1f%% different, expected ≥30%%)", genre1, genre2, diff)
			}
		}
	}
}

// Helper functions

func calculateGenreDifference(samples1, samples2 []interface{}) float64 {
	if len(samples1) == 0 || len(samples2) == 0 {
		return 0.0
	}

	// Compare first 10 samples from each genre
	compareCount := 10
	if len(samples1) < compareCount {
		compareCount = len(samples1)
	}
	if len(samples2) < compareCount {
		compareCount = len(samples2)
	}

	differentCount := 0
	for i := 0; i < compareCount; i++ {
		json1, _ := json.Marshal(samples1[i])
		json2, _ := json.Marshal(samples2[i])

		// Simple byte-wise comparison
		if string(json1) != string(json2) {
			differentCount++
		}
	}

	return float64(differentCount) / float64(compareCount) * 100.0
}
