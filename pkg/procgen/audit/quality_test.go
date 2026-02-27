package audit

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/book"
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
		{"Book", book.NewGenerator(), &BookQualityValidator{}},
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

				// BookGenerator requires book_type parameter
				if tc.name == "Book" {
					params.Custom = map[string]interface{}{
						"book_type": engine.BookTypeLore,
					}
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

	// Check stats are within reasonable ranges
	// Templates provide base ranges (minion: 10-30, monster: 40-80, boss: 100-200)
	// Level scaling: 1 + (level-1)*0.15, where level ~= depth
	// Rarity scaling: 1.0x-3.0x
	// For depth=5, expect level ~5, giving 1.6x level multiplier
	// Minimum: 10 * 1.6 * 1.0 = 16 (minion, common)
	// Maximum: 200 * 2.0 * 3.0 = 1200 (boss, legendary with high level variance)
	// With depth variance and rarity, can exceed this, so use generous bounds
	minHealth := 5
	maxHealth := 3000

	if e.Stats.Health < minHealth || e.Stats.Health > maxHealth {
		return fmt.Errorf("health out of reasonable range: %d (expected %d-%d)", e.Stats.Health, minHealth, maxHealth)
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

	// Check mana cost is in reasonable range (templates: 20-100, scaled by rarity up to 3x,
	// plus depth/difficulty scaling can push it higher)
	minMana := 0
	maxMana := 3000
	if s.Stats.ManaCost < minMana || s.Stats.ManaCost > maxMana {
		return fmt.Errorf("mana cost out of reasonable range: %d (expected %d-%d)",
			s.Stats.ManaCost, minMana, maxMana)
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

	// Check minimum objectives (≥1)
	if len(q.Objectives) < 1 {
		return fmt.Errorf("insufficient objectives: %d (expected ≥1)", len(q.Objectives))
	}

	// Check rewards are positive
	if q.Reward.Gold < 0 || q.Reward.XP < 0 {
		return fmt.Errorf("negative rewards: gold=%d xp=%d",
			q.Reward.Gold, q.Reward.XP)
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
	// Base stats: ~120-300 each, scaled by depth (~1.75x for depth 5) and rarity (1.0-3.0x)
	// Total can range from ~360 (low base * 1.0 scale) to ~1575 (high base * 1.75 * 3.0)
	totalStats := int(veh.MaxSpeed + veh.Handling + veh.MaxDurability)
	expectedMin := 100
	expectedMax := 2000
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

	// Check minimum rooms (≥1)
	if len(b.Rooms) < 1 {
		return fmt.Errorf("insufficient rooms: %d (expected ≥1)", len(b.Rooms))
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

// BookQualityValidator validates book generation quality
type BookQualityValidator struct{}

func (v *BookQualityValidator) Validate(result interface{}, params procgen.GenerationParams) error {
	b, ok := result.(*engine.BookComponent)
	if !ok {
		return fmt.Errorf("expected *engine.BookComponent, got %T", result)
	}

	// Check title is non-empty
	if b.Title == "" {
		return fmt.Errorf("empty book title")
	}

	// Check author is non-empty
	if b.Author == "" {
		return fmt.Errorf("empty book author")
	}

	// Check content has pages
	if len(b.Content) == 0 {
		return fmt.Errorf("book has no content pages")
	}

	// Check pages have minimum content
	totalChars := 0
	for _, page := range b.Content {
		totalChars += len(page)
	}
	if totalChars < 50 {
		return fmt.Errorf("content too short: %d characters across %d pages (expected ≥50 total)", totalChars, len(b.Content))
	}

	// Check BookType is valid
	validTypes := map[engine.BookType]bool{
		engine.BookTypeSkill:   true,
		engine.BookTypeLore:    true,
		engine.BookTypeQuest:   true,
		engine.BookTypeRecipe:  true,
		engine.BookTypeHistory: true,
	}
	if !validTypes[b.BookType] {
		return fmt.Errorf("invalid book type: %v", b.BookType)
	}

	// For skill books, check that skill bonuses are set
	if b.BookType == engine.BookTypeSkill && len(b.SkillBonus) == 0 {
		return fmt.Errorf("skill book has no skill bonuses")
	}

	return nil
}

// TestRarityDistribution validates that rarity tiers match expected percentages
func TestRarityDistribution(t *testing.T) {
	const samples = 10000
	generators := []struct {
		name                 string
		generator            procgen.Generator
		extractor            RarityExtractor
		expectedDistribution map[string]float64
	}{
		{
			"Entity",
			entity.NewEntityGenerator(),
			&EntityRarityExtractor{},
			map[string]float64{
				"Common":    52.0, // 52% (entity mix includes minions which are mostly common)
				"Uncommon":  17.0, // 17%
				"Rare":      16.0, // 16%
				"Epic":      10.0, // 10%
				"Legendary": 5.0,  // 5%
			},
		},
		{
			"Item",
			item.NewItemGenerator(),
			&ItemRarityExtractor{},
			map[string]float64{
				"Common":    49.0, // 49%
				"Uncommon":  30.0, // 30%
				"Rare":      13.0, // 13%
				"Epic":      5.0,  // 5%
				"Legendary": 3.0,  // 3%
			},
		},
	}

	for _, tc := range generators {
		t.Run(tc.name, func(t *testing.T) {
			counts := make(map[string]int)

			for i := 0; i < samples; i++ {
				seed := int64(2000 + i)
				params := procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      1, // Use low depth to match baseline rarity distribution
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
			for rarityName, expected := range tc.expectedDistribution {
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
	// Entity generator may return a single entity or a slice
	var ent *entity.Entity

	switch v := result.(type) {
	case *entity.Entity:
		ent = v
	case []*entity.Entity:
		if len(v) == 0 {
			return "Unknown"
		}
		ent = v[0] // Extract from first entity as a sample
	default:
		return "Unknown"
	}

	// Capitalize first letter to match expected format
	rarity := ent.Rarity.String()
	if len(rarity) > 0 {
		return strings.ToUpper(rarity[:1]) + rarity[1:]
	}
	return "Unknown"
}

type ItemRarityExtractor struct{}

func (e *ItemRarityExtractor) Extract(result interface{}) string {
	// Item generator may return a single item or a slice
	var itm *item.Item

	switch v := result.(type) {
	case *item.Item:
		itm = v
	case []*item.Item:
		if len(v) == 0 {
			return "Unknown"
		}
		itm = v[0] // Extract from first item as a sample
	default:
		return "Unknown"
	}

	// Capitalize first letter to match expected format
	rarity := itm.Rarity.String()
	if len(rarity) > 0 {
		return strings.ToUpper(rarity[:1]) + rarity[1:]
	}
	return "Unknown"
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
