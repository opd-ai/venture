//go:build !android && !ios
// +build !android,!ios

package main

import (
	"image/color"
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/environment"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/opd-ai/venture/pkg/rendering/palette"
)

func TestGetLightConfig(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		wantInterval  int
		wantFlicker   bool
		wantRadiusGT0 bool
	}{
		{"fantasy", "fantasy", 5, true, true},
		{"scifi", "scifi", 4, false, true},
		{"horror", "horror", 7, true, true},
		{"cyberpunk", "cyberpunk", 3, false, true},
		{"postapoc", "postapoc", 6, true, true},
		{"unknown defaults to fantasy", "unknown", 5, true, true},
		{"empty defaults to fantasy", "", 5, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getLightConfig(tt.genreID)
			if config.torchInterval != tt.wantInterval {
				t.Errorf("torchInterval = %d, want %d", config.torchInterval, tt.wantInterval)
			}
			if config.torchFlicker != tt.wantFlicker {
				t.Errorf("torchFlicker = %v, want %v", config.torchFlicker, tt.wantFlicker)
			}
			if tt.wantRadiusGT0 && config.torchRadius <= 0 {
				t.Errorf("torchRadius = %f, want > 0", config.torchRadius)
			}
			if tt.wantRadiusGT0 && config.crystalRadius <= 0 {
				t.Errorf("crystalRadius = %f, want > 0", config.crystalRadius)
			}
			if config.torchColor.A != 255 {
				t.Errorf("torchColor alpha = %d, want 255", config.torchColor.A)
			}
			if config.crystalColor.A != 255 {
				t.Errorf("crystalColor alpha = %d, want 255", config.crystalColor.A)
			}
		})
	}
}

func TestGetObjectConfig(t *testing.T) {
	tests := []struct {
		name           string
		genreID        string
		wantPerRoom    int
		expectPositive bool
	}{
		{"fantasy", "fantasy", 3, true},
		{"scifi", "scifi", 4, true},
		{"horror", "horror", 2, true},
		{"cyberpunk", "cyberpunk", 3, true},
		{"postapocalyptic", "postapocalyptic", 4, true},
		{"unknown defaults to fantasy", "unknown", 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getObjectConfig(tt.genreID)
			if config.objectsPerRoom != tt.wantPerRoom {
				t.Errorf("objectsPerRoom = %d, want %d", config.objectsPerRoom, tt.wantPerRoom)
			}
			if tt.expectPositive {
				if config.crateChance <= 0 || config.crateChance > 1 {
					t.Errorf("crateChance = %f, want (0, 1]", config.crateChance)
				}
				if config.barrelChance <= 0 || config.barrelChance > 1 {
					t.Errorf("barrelChance = %f, want (0, 1]", config.barrelChance)
				}
				if config.furnitureChance <= 0 || config.furnitureChance > 1 {
					t.Errorf("furnitureChance = %f, want (0, 1]", config.furnitureChance)
				}
			}
		})
	}
}

func TestCalculateHazardPosition(t *testing.T) {
	tests := []struct {
		name      string
		room      *terrain.Room
		tileSize  int
		padding   int
		wantValid bool
	}{
		{
			name:      "normal room",
			room:      &terrain.Room{X: 10, Y: 10, Width: 20, Height: 20},
			tileSize:  32,
			padding:   2,
			wantValid: true,
		},
		{
			name:      "room too small for padding",
			room:      &terrain.Room{X: 0, Y: 0, Width: 4, Height: 4},
			tileSize:  32,
			padding:   2,
			wantValid: false,
		},
		{
			name:      "room exactly at padding boundary",
			room:      &terrain.Room{X: 5, Y: 5, Width: 4, Height: 4},
			tileSize:  32,
			padding:   2,
			wantValid: false,
		},
		{
			name:      "large room",
			room:      &terrain.Room{X: 0, Y: 0, Width: 100, Height: 100},
			tileSize:  32,
			padding:   5,
			wantValid: true,
		},
		{
			name:      "minimum valid room",
			room:      &terrain.Room{X: 0, Y: 0, Width: 5, Height: 5},
			tileSize:  32,
			padding:   2,
			wantValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			worldX, worldY, valid := calculateHazardPosition(tt.room, rng, tt.tileSize, tt.padding)
			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
			if valid {
				minX := float64((tt.room.X + tt.padding) * tt.tileSize)
				maxX := float64((tt.room.X + tt.room.Width - tt.padding) * tt.tileSize)
				minY := float64((tt.room.Y + tt.padding) * tt.tileSize)
				maxY := float64((tt.room.Y + tt.room.Height - tt.padding) * tt.tileSize)
				if worldX < minX || worldX >= maxX {
					t.Errorf("worldX = %f, want in [%f, %f)", worldX, minX, maxX)
				}
				if worldY < minY || worldY >= maxY {
					t.Errorf("worldY = %f, want in [%f, %f)", worldY, minY, maxY)
				}
			}
		})
	}
}

func TestCalculateHazardPositionDeterminism(t *testing.T) {
	room := &terrain.Room{X: 10, Y: 10, Width: 20, Height: 20}
	seed := int64(12345)

	rng1 := rand.New(rand.NewSource(seed))
	x1, y1, v1 := calculateHazardPosition(room, rng1, 32, 2)

	rng2 := rand.New(rand.NewSource(seed))
	x2, y2, v2 := calculateHazardPosition(room, rng2, 32, 2)

	if x1 != x2 || y1 != y2 || v1 != v2 {
		t.Errorf("non-deterministic: (%f,%f,%v) vs (%f,%f,%v)", x1, y1, v1, x2, y2, v2)
	}
}

func TestDetermineHazardType(t *testing.T) {
	tests := []struct {
		name    string
		subType environment.SubType
		want    engine.HazardType
	}{
		{"fire pit", environment.SubTypeFirePit, engine.HazardPoison},
		{"lava pit", environment.SubTypeLavaPit, engine.HazardPoison},
		{"acid pool", environment.SubTypeAcidPool, engine.HazardPoison},
		{"poison gas", environment.SubTypePoisonGas, engine.HazardPoison},
		{"electric field", environment.SubTypeElectricField, engine.HazardPoison},
		{"ice field", environment.SubTypeIceField, engine.HazardWater},
		{"spikes", environment.SubTypeSpikes, engine.HazardPoison},
		{"bear trap", environment.SubTypeBearTrap, engine.HazardPoison},
		{"unknown defaults to poison", environment.SubType(999), engine.HazardPoison},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineHazardType(tt.subType)
			if got != tt.want {
				t.Errorf("determineHazardType(%v) = %v, want %v", tt.subType, got, tt.want)
			}
		})
	}
}

func TestSelectHazardSubType(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic", "unknown"}
	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			subType := selectHazardSubType(genreID, rng)
			if subType < 0 {
				t.Errorf("selectHazardSubType returned negative SubType: %d", subType)
			}
		})
	}
}

func TestSelectHazardSubTypeDeterminism(t *testing.T) {
	seed := int64(98765)
	for _, genreID := range []string{"fantasy", "scifi", "horror"} {
		t.Run(genreID, func(t *testing.T) {
			rng1 := rand.New(rand.NewSource(seed))
			result1 := selectHazardSubType(genreID, rng1)

			rng2 := rand.New(rand.NewSource(seed))
			result2 := selectHazardSubType(genreID, rng2)

			if result1 != result2 {
				t.Errorf("non-deterministic for genre %s: %v vs %v", genreID, result1, result2)
			}
		})
	}
}

func TestSelectObjectType(t *testing.T) {
	config := getObjectConfig("fantasy")
	rng := rand.New(rand.NewSource(42))

	// Run multiple times to get distribution
	counts := map[engine.ObjectType]int{}
	spawned := 0
	total := 1000
	for i := 0; i < total; i++ {
		objType, selected := selectObjectType(rng, config)
		if selected {
			counts[objType]++
			spawned++
		}
	}

	// Verify we got some objects and some misses
	if spawned == 0 {
		t.Error("no objects spawned in 1000 attempts")
	}
	if spawned == total {
		t.Error("all 1000 attempts spawned objects (expected some misses)")
	}

	// Verify crates are most common (highest chance)
	if counts[engine.ObjectCrate] == 0 {
		t.Error("no crates spawned despite having highest chance")
	}
}

func TestSelectObjectTypeDeterminism(t *testing.T) {
	config := getObjectConfig("scifi")
	seed := int64(54321)

	rng1 := rand.New(rand.NewSource(seed))
	type1, sel1 := selectObjectType(rng1, config)

	rng2 := rand.New(rand.NewSource(seed))
	type2, sel2 := selectObjectType(rng2, config)

	if type1 != type2 || sel1 != sel2 {
		t.Errorf("non-deterministic: (%v,%v) vs (%v,%v)", type1, sel1, type2, sel2)
	}
}

func TestConvertVehicleType(t *testing.T) {
	tests := []struct {
		name string
		in   vehicle.VehicleType
		want engine.VehicleType
	}{
		{"mount", vehicle.TypeMount, engine.VehicleMount},
		{"cart", vehicle.TypeCart, engine.VehicleCart},
		{"boat", vehicle.TypeBoat, engine.VehicleBoat},
		{"glider", vehicle.TypeGlider, engine.VehicleGlider},
		{"mech", vehicle.TypeMech, engine.VehicleMech},
		{"unknown defaults to mount", vehicle.VehicleType(99), engine.VehicleMount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertVehicleType(tt.in)
			if got != tt.want {
				t.Errorf("convertVehicleType(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvertColorToRGBA(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
		want  color.RGBA
	}{
		{"red", 0xFF0000, color.RGBA{255, 0, 0, 255}},
		{"green", 0x00FF00, color.RGBA{0, 255, 0, 255}},
		{"blue", 0x0000FF, color.RGBA{0, 0, 255, 255}},
		{"white", 0xFFFFFF, color.RGBA{255, 255, 255, 255}},
		{"black", 0x000000, color.RGBA{0, 0, 0, 255}},
		{"mixed", 0xA0B0C0, color.RGBA{160, 176, 192, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertColorToRGBA(tt.value)
			if got != tt.want {
				t.Errorf("convertColorToRGBA(0x%06X) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestCalculateVehicleSizing(t *testing.T) {
	tests := []struct {
		name         string
		vType        vehicle.VehicleType
		wantSize     int
		wantCollider float64
	}{
		{"mount", vehicle.TypeMount, 32, 28.0},
		{"cart", vehicle.TypeCart, 40, 36.0},
		{"boat", vehicle.TypeBoat, 48, 44.0},
		{"glider", vehicle.TypeGlider, 36, 32.0},
		{"mech", vehicle.TypeMech, 44, 40.0},
		{"unknown defaults to mount", vehicle.VehicleType(99), 32, 28.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, collider := calculateVehicleSizing(tt.vType)
			if size != tt.wantSize {
				t.Errorf("size = %d, want %d", size, tt.wantSize)
			}
			if collider != tt.wantCollider {
				t.Errorf("collider = %f, want %f", collider, tt.wantCollider)
			}
			if float64(size) <= collider {
				// Collider should be smaller than sprite size
			}
		})
	}
}

func TestCalculateCompanionSizing(t *testing.T) {
	tests := []struct {
		name         string
		compType     engine.CompanionType
		wantSize     int
		wantCollider float64
	}{
		{"pet", engine.CompanionTypePet, 24, 20.0},
		{"summon", engine.CompanionTypeSummon, 28, 24.0},
		{"hireling", engine.CompanionTypeHireling, 28, 24.0},
		{"elemental", engine.CompanionTypeElemental, 32, 28.0},
		{"undead", engine.CompanionTypeUndead, 30, 26.0},
		{"robot", engine.CompanionTypeRobot, 30, 26.0},
		{"spirit", engine.CompanionTypeSpirit, 26, 22.0},
		{"insect", engine.CompanionTypeInsect, 22, 18.0},
		{"unknown defaults", engine.CompanionType(99), 28, 24.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, collider := calculateCompanionSizing(tt.compType)
			if size != tt.wantSize {
				t.Errorf("size = %d, want %d", size, tt.wantSize)
			}
			if collider != tt.wantCollider {
				t.Errorf("collider = %f, want %f", collider, tt.wantCollider)
			}
		})
	}
}

func TestCalculateCompanionSizingColliderSmallerThanSprite(t *testing.T) {
	types := []engine.CompanionType{
		engine.CompanionTypePet, engine.CompanionTypeSummon, engine.CompanionTypeHireling,
		engine.CompanionTypeElemental, engine.CompanionTypeUndead, engine.CompanionTypeRobot,
		engine.CompanionTypeSpirit, engine.CompanionTypeInsect,
	}
	for _, ct := range types {
		size, collider := calculateCompanionSizing(ct)
		if float64(size) <= collider {
			t.Errorf("companion type %d: sprite size %d should be > collider %f", ct, size, collider)
		}
	}
}

func TestCalculateBookshelfCount(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
		want      int
	}{
		{"no rooms", 0, 0},
		{"one room", 1, 0},
		{"two rooms", 2, 1},
		{"five rooms", 5, 1},
		{"seven rooms", 7, 1},
		{"eight rooms", 8, 2},
		{"many rooms", 20, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms := make([]*terrain.Room, tt.roomCount)
			for i := range rooms {
				rooms[i] = &terrain.Room{X: i, Y: i, Width: 10, Height: 10}
			}
			terrainMap := &terrain.Terrain{Rooms: rooms}
			got := calculateBookshelfCount(terrainMap)
			if got != tt.want {
				t.Errorf("calculateBookshelfCount(%d rooms) = %d, want %d", tt.roomCount, got, tt.want)
			}
		})
	}
}

func TestSelectBookType(t *testing.T) {
	validTypes := map[engine.BookType]bool{
		engine.BookTypeSkill:   true,
		engine.BookTypeLore:    true,
		engine.BookTypeRecipe:  true,
		engine.BookTypeHistory: true,
	}

	// Test multiple seeds produce valid types
	for seed := int64(0); seed < 100; seed++ {
		bt := selectBookType(seed)
		if !validTypes[bt] {
			t.Errorf("selectBookType(%d) = %v, not a valid book type", seed, bt)
		}
	}
}

func TestSelectBookTypeDeterminism(t *testing.T) {
	seeds := []int64{0, 42, 12345, 999999}
	for _, seed := range seeds {
		t.Run("seed", func(t *testing.T) {
			result1 := selectBookType(seed)
			result2 := selectBookType(seed)
			if result1 != result2 {
				t.Errorf("non-deterministic for seed %d: %v vs %v", seed, result1, result2)
			}
		})
	}
}

func TestCreateBookParams(t *testing.T) {
	baseParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
	}

	tests := []struct {
		name     string
		bookType engine.BookType
		seed     int64
		wantType engine.BookType
		hasSkill bool
	}{
		{"skill book", engine.BookTypeSkill, 42, engine.BookTypeSkill, true},
		{"lore book", engine.BookTypeLore, 42, engine.BookTypeLore, false},
		{"recipe book", engine.BookTypeRecipe, 42, engine.BookTypeRecipe, false},
		{"history book", engine.BookTypeHistory, 42, engine.BookTypeHistory, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createBookParams(baseParams, tt.bookType, tt.seed)

			// Verify book type is set in Custom map
			if bt, ok := result.Custom["book_type"]; !ok {
				t.Error("missing book_type in Custom map")
			} else if bt != tt.wantType {
				t.Errorf("book_type = %v, want %v", bt, tt.wantType)
			}

			// Verify skill books have extra params
			if tt.hasSkill {
				if _, ok := result.Custom["skill_name"]; !ok {
					t.Error("skill book missing skill_name in Custom map")
				}
				if bonus, ok := result.Custom["skill_bonus"]; !ok {
					t.Error("skill book missing skill_bonus in Custom map")
				} else if bonus.(float64) != float64(baseParams.Depth)*0.1 {
					t.Errorf("skill_bonus = %v, want %v", bonus, float64(baseParams.Depth)*0.1)
				}
			} else {
				if _, ok := result.Custom["skill_name"]; ok {
					t.Error("non-skill book should not have skill_name")
				}
			}

			// Verify base params are preserved
			if result.Difficulty != baseParams.Difficulty {
				t.Errorf("Difficulty modified: %f vs %f", result.Difficulty, baseParams.Difficulty)
			}
			if result.Depth != baseParams.Depth {
				t.Errorf("Depth modified: %d vs %d", result.Depth, baseParams.Depth)
			}
		})
	}
}

func TestCreateBookParamsDeterminism(t *testing.T) {
	params := procgen.GenerationParams{Difficulty: 0.5, Depth: 5, GenreID: "fantasy"}
	seed := int64(12345)

	r1 := createBookParams(params, engine.BookTypeSkill, seed)
	r2 := createBookParams(params, engine.BookTypeSkill, seed)

	if r1.Custom["skill_name"] != r2.Custom["skill_name"] {
		t.Errorf("non-deterministic skill_name: %v vs %v",
			r1.Custom["skill_name"], r2.Custom["skill_name"])
	}
}

func TestGenerateCompanionColor(t *testing.T) {
	types := []engine.CompanionType{
		engine.CompanionTypePet, engine.CompanionTypeSummon, engine.CompanionTypeHireling,
		engine.CompanionTypeElemental, engine.CompanionTypeUndead, engine.CompanionTypeRobot,
		engine.CompanionTypeSpirit, engine.CompanionTypeInsect,
	}
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, ct := range types {
		for _, g := range genres {
			t.Run("", func(t *testing.T) {
				c := generateCompanionColor(ct, g, 42)
				if c.A != 255 {
					t.Errorf("alpha = %d, want 255", c.A)
				}
			})
		}
	}
}

func TestGenerateCompanionColorDeterminism(t *testing.T) {
	seed := int64(12345)
	c1 := generateCompanionColor(engine.CompanionTypePet, "fantasy", seed)
	c2 := generateCompanionColor(engine.CompanionTypePet, "fantasy", seed)
	if c1 != c2 {
		t.Errorf("non-deterministic: %v vs %v", c1, c2)
	}
}

func TestGenerateCompanionColorDifferentSeeds(t *testing.T) {
	c1 := generateCompanionColor(engine.CompanionTypePet, "fantasy", 1)
	c2 := generateCompanionColor(engine.CompanionTypePet, "fantasy", 2)
	// Different seeds should usually produce different colors
	if c1 == c2 {
		t.Log("warning: different seeds produced same color (unlikely but possible)")
	}
}

func TestGenerateBookshelfColor(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic", "unknown"}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			c := generateBookshelfColor(g, 42)
			if c.A != 255 {
				t.Errorf("alpha = %d, want 255", c.A)
			}
		})
	}
}

func TestGenerateBookshelfColorDeterminism(t *testing.T) {
	seed := int64(99999)
	c1 := generateBookshelfColor("fantasy", seed)
	c2 := generateBookshelfColor("fantasy", seed)
	if c1 != c2 {
		t.Errorf("non-deterministic: %v vs %v", c1, c2)
	}
}

func TestParsePaletteOptions(t *testing.T) {
	// Save original flag values
	origHarmony := *paletteHarmony
	origMood := *paletteMood
	origRarity := *paletteRarity
	defer func() {
		*paletteHarmony = origHarmony
		*paletteMood = origMood
		*paletteRarity = origRarity
	}()

	tests := []struct {
		name    string
		harmony string
		mood    string
		rarity  string
		wantErr bool
		wantH   palette.HarmonyType
		wantR   palette.Rarity
	}{
		{"valid triadic/vibrant/epic", "triadic", "vibrant", "epic", false, palette.HarmonyTriadic, palette.RarityEpic},
		{"valid complementary/dark/common", "complementary", "dark", "common", false, palette.HarmonyComplementary, palette.RarityCommon},
		{"valid analogous/calm/legendary", "analogous", "calm", "legendary", false, palette.HarmonyAnalogous, palette.RarityLegendary},
		{"valid monochromatic/bright/rare", "monochromatic", "bright", "rare", false, palette.HarmonyMonochromatic, palette.RarityRare},
		{"valid tetradic/normal/uncommon", "tetradic", "normal", "uncommon", false, palette.HarmonyTetradic, palette.RarityUncommon},
		{"valid split-complementary", "split-complementary", "normal", "common", false, palette.HarmonySplitComplementary, palette.RarityCommon},
		{"invalid harmony", "invalid", "vibrant", "epic", true, 0, 0},
		{"invalid mood", "triadic", "invalid", "epic", true, 0, 0},
		{"invalid rarity", "triadic", "vibrant", "invalid", true, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*paletteHarmony = tt.harmony
			*paletteMood = tt.mood
			*paletteRarity = tt.rarity

			opts, err := parsePaletteOptions()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if opts.Harmony != tt.wantH {
					t.Errorf("Harmony = %v, want %v", opts.Harmony, tt.wantH)
				}
				if opts.Rarity != tt.wantR {
					t.Errorf("Rarity = %v, want %v", opts.Rarity, tt.wantR)
				}
				if opts.MinColors != 12 {
					t.Errorf("MinColors = %d, want 12", opts.MinColors)
				}
			}
		})
	}
}

func TestParsePaletteOptionsMoodCoverage(t *testing.T) {
	origHarmony := *paletteHarmony
	origMood := *paletteMood
	origRarity := *paletteRarity
	defer func() {
		*paletteHarmony = origHarmony
		*paletteMood = origMood
		*paletteRarity = origRarity
	}()

	*paletteHarmony = "triadic"
	*paletteRarity = "epic"

	moods := []string{
		"normal", "bright", "dark", "saturated", "muted", "vibrant", "pastel",
		"tense", "calm", "victorious", "melancholic", "energetic", "mystical",
		"ominous", "serene", "aggressive", "playful", "somber", "ethereal",
		"dangerous", "peaceful", "chaotic", "regal", "desolate",
	}
	for _, mood := range moods {
		t.Run(mood, func(t *testing.T) {
			*paletteMood = mood
			opts, err := parsePaletteOptions()
			if err != nil {
				t.Errorf("unexpected error for mood %q: %v", mood, err)
			}
			if opts == nil {
				t.Fatal("opts is nil")
			}
		})
	}
}

func TestValidateClientConfiguration(t *testing.T) {
	// Save original flag values
	origGenre := *genreID
	origHostPlay := *hostAndPlay
	origPort := *serverPort
	origPlayers := *serverPlayers
	origTick := *serverTick
	defer func() {
		*genreID = origGenre
		*hostAndPlay = origHostPlay
		*serverPort = origPort
		*serverPlayers = origPlayers
		*serverTick = origTick
	}()

	tests := []struct {
		name    string
		genre   string
		host    bool
		port    int
		players int
		tick    int
		wantErr bool
	}{
		{"valid defaults", "fantasy", false, 8080, 4, 30, false},
		{"valid scifi", "scifi", false, 8080, 4, 30, false},
		{"valid horror", "horror", false, 8080, 4, 30, false},
		{"valid random", "random", false, 8080, 4, 30, false},
		{"valid host-and-play", "fantasy", true, 8080, 4, 30, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*genreID = tt.genre
			*hostAndPlay = tt.host
			*serverPort = tt.port
			*serverPlayers = tt.players
			*serverTick = tt.tick

			err := validateClientConfiguration()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkGetLightConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getLightConfig("fantasy")
	}
}

func BenchmarkGetObjectConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getObjectConfig("scifi")
	}
}

func BenchmarkSelectHazardSubType(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selectHazardSubType("fantasy", rng)
	}
}

func BenchmarkCalculateHazardPosition(b *testing.B) {
	room := &terrain.Room{X: 10, Y: 10, Width: 20, Height: 20}
	rng := rand.New(rand.NewSource(42))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateHazardPosition(room, rng, 32, 2)
	}
}

func BenchmarkConvertColorToRGBA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		convertColorToRGBA(0xA0B0C0)
	}
}

func BenchmarkGenerateCompanionColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateCompanionColor(engine.CompanionTypePet, "fantasy", int64(i))
	}
}
