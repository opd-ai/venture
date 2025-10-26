package entity

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestGenerateMerchant tests basic merchant generation.
func TestGenerateMerchant(t *testing.T) {
	tests := []struct {
		name         string
		seed         int64
		genreID      string
		merchantType MerchantType
		wantErr      bool
	}{
		{
			name:         "fixed fantasy merchant",
			seed:         12345,
			genreID:      "fantasy",
			merchantType: MerchantFixed,
			wantErr:      false,
		},
		{
			name:         "nomadic scifi merchant",
			seed:         67890,
			genreID:      "scifi",
			merchantType: MerchantNomadic,
			wantErr:      false,
		},
		{
			name:         "fixed horror merchant",
			seed:         11111,
			genreID:      "horror",
			merchantType: MerchantFixed,
			wantErr:      false,
		},
		{
			name:         "nomadic cyberpunk merchant",
			seed:         22222,
			genreID:      "cyberpunk",
			merchantType: MerchantNomadic,
			wantErr:      false,
		},
		{
			name:         "fixed postapoc merchant",
			seed:         33333,
			genreID:      "postapoc",
			merchantType: MerchantFixed,
			wantErr:      false,
		},
		{
			name:         "default genre merchant",
			seed:         44444,
			genreID:      "",
			merchantType: MerchantFixed,
			wantErr:      false,
		},
	}

	gen := NewEntityGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID:    tt.genreID,
				Difficulty: 0.5,
				Depth:      1,
			}

			merchant, err := gen.GenerateMerchant(tt.seed, params, tt.merchantType)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMerchant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Validate merchant data
			if merchant == nil {
				t.Fatal("expected merchant, got nil")
			}

			if merchant.Entity == nil {
				t.Fatal("expected entity, got nil")
			}

			if merchant.Entity.Type != TypeNPC {
				t.Errorf("expected TypeNPC, got %v", merchant.Entity.Type)
			}

			if merchant.Entity.Name == "" {
				t.Error("expected non-empty name")
			}

			if merchant.MerchantType != tt.merchantType {
				t.Errorf("expected merchant type %v, got %v", tt.merchantType, merchant.MerchantType)
			}

			if len(merchant.Inventory) == 0 {
				t.Error("expected non-empty inventory")
			}

			if merchant.PriceMultiplier <= 0 {
				t.Errorf("expected positive price multiplier, got %v", merchant.PriceMultiplier)
			}

			if merchant.BuyBackPercentage <= 0 || merchant.BuyBackPercentage > 1 {
				t.Errorf("expected buyback percentage between 0 and 1, got %v", merchant.BuyBackPercentage)
			}
		})
	}
}

// TestGenerateMerchantDeterminism verifies merchants generate deterministically.
func TestGenerateMerchantDeterminism(t *testing.T) {
	gen := NewEntityGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
	}

	// Generate twice with same seed
	merchant1, err1 := gen.GenerateMerchant(seed, params, MerchantFixed)
	merchant2, err2 := gen.GenerateMerchant(seed, params, MerchantFixed)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}

	// Compare results
	if merchant1.Entity.Name != merchant2.Entity.Name {
		t.Errorf("names differ: %s vs %s", merchant1.Entity.Name, merchant2.Entity.Name)
	}

	if len(merchant1.Inventory) != len(merchant2.Inventory) {
		t.Errorf("inventory sizes differ: %d vs %d", len(merchant1.Inventory), len(merchant2.Inventory))
	}

	if merchant1.PriceMultiplier != merchant2.PriceMultiplier {
		t.Errorf("price multipliers differ: %v vs %v", merchant1.PriceMultiplier, merchant2.PriceMultiplier)
	}
}

// TestGenerateMerchantPricing tests merchant pricing differs by type.
func TestGenerateMerchantPricing(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
	}

	fixed, err := gen.GenerateMerchant(12345, params, MerchantFixed)
	if err != nil {
		t.Fatalf("failed to generate fixed merchant: %v", err)
	}

	nomadic, err := gen.GenerateMerchant(12346, params, MerchantNomadic)
	if err != nil {
		t.Fatalf("failed to generate nomadic merchant: %v", err)
	}

	// Nomadic merchants should charge more
	if nomadic.PriceMultiplier <= fixed.PriceMultiplier {
		t.Errorf("expected nomadic merchant (%v) to charge more than fixed merchant (%v)",
			nomadic.PriceMultiplier, fixed.PriceMultiplier)
	}

	// Both should have reasonable multipliers
	if fixed.PriceMultiplier < 1.0 || fixed.PriceMultiplier > 3.0 {
		t.Errorf("fixed merchant price multiplier out of range: %v", fixed.PriceMultiplier)
	}

	if nomadic.PriceMultiplier < 1.0 || nomadic.PriceMultiplier > 3.0 {
		t.Errorf("nomadic merchant price multiplier out of range: %v", nomadic.PriceMultiplier)
	}
}

// TestGenerateMerchantInventorySize tests inventory generation.
func TestGenerateMerchantInventorySize(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
	}

	merchant, err := gen.GenerateMerchant(12345, params, MerchantFixed)
	if err != nil {
		t.Fatalf("failed to generate merchant: %v", err)
	}

	// Inventory should have 15-24 items
	if len(merchant.Inventory) < 10 || len(merchant.Inventory) > 30 {
		t.Errorf("inventory size out of expected range: %d", len(merchant.Inventory))
	}

	// All items should be valid
	for i, itm := range merchant.Inventory {
		if itm == nil {
			t.Errorf("inventory slot %d is nil", i)
			continue
		}

		if itm.Name == "" {
			t.Errorf("inventory slot %d has empty name", i)
		}

		if itm.Stats.Value <= 0 {
			t.Errorf("inventory slot %d (%s) has invalid value: %d", i, itm.Name, itm.Stats.Value)
		}
	}
}

// TestGenerateMerchantStats tests merchant entity stats.
func TestGenerateMerchantStats(t *testing.T) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
	}

	merchant, err := gen.GenerateMerchant(12345, params, MerchantFixed)
	if err != nil {
		t.Fatalf("failed to generate merchant: %v", err)
	}

	entity := merchant.Entity

	// Merchants should be level 1
	if entity.Stats.Level != 1 {
		t.Errorf("expected level 1, got %d", entity.Stats.Level)
	}

	// Merchants should have zero damage (non-combatant)
	if entity.Stats.Damage != 0 {
		t.Errorf("expected zero damage, got %d", entity.Stats.Damage)
	}

	// Merchants should have health
	if entity.Stats.MaxHealth <= 0 {
		t.Errorf("expected positive health, got %d", entity.Stats.MaxHealth)
	}

	// Merchants should have defense
	if entity.Stats.Defense < 0 {
		t.Errorf("expected non-negative defense, got %d", entity.Stats.Defense)
	}

	// Merchants should have speed
	if entity.Stats.Speed <= 0 {
		t.Errorf("expected positive speed, got %v", entity.Stats.Speed)
	}
}

// TestGenerateMerchantSpawnPoints tests spawn point generation.
func TestGenerateMerchantSpawnPoints(t *testing.T) {
	tests := []struct {
		name         string
		worldSeed    int64
		width        int
		height       int
		merchantType MerchantType
		count        int
	}{
		{
			name:         "fixed merchants in small world",
			worldSeed:    12345,
			width:        800,
			height:       600,
			merchantType: MerchantFixed,
			count:        3,
		},
		{
			name:         "nomadic merchants in large world",
			worldSeed:    67890,
			width:        1600,
			height:       1200,
			merchantType: MerchantNomadic,
			count:        5,
		},
		{
			name:         "single fixed merchant",
			worldSeed:    11111,
			width:        1024,
			height:       768,
			merchantType: MerchantFixed,
			count:        1,
		},
		{
			name:         "many nomadic merchants",
			worldSeed:    22222,
			width:        2000,
			height:       2000,
			merchantType: MerchantNomadic,
			count:        10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := GenerateMerchantSpawnPoints(tt.worldSeed, tt.width, tt.height, tt.merchantType, tt.count)

			if len(points) != tt.count {
				t.Errorf("expected %d points, got %d", tt.count, len(points))
			}

			// Validate points are within bounds (with some margin for offsets)
			for i, pt := range points {
				if pt.X < -200 || pt.X > float64(tt.width)+200 {
					t.Errorf("point %d X coordinate out of bounds: %v", i, pt.X)
				}
				if pt.Y < -200 || pt.Y > float64(tt.height)+200 {
					t.Errorf("point %d Y coordinate out of bounds: %v", i, pt.Y)
				}
			}
		})
	}
}

// TestGenerateMerchantSpawnPointsDeterminism tests deterministic spawn points.
func TestGenerateMerchantSpawnPointsDeterminism(t *testing.T) {
	seed := int64(99999)
	width, height := 1000, 800
	count := 5

	// Generate twice with same seed
	points1 := GenerateMerchantSpawnPoints(seed, width, height, MerchantFixed, count)
	points2 := GenerateMerchantSpawnPoints(seed, width, height, MerchantFixed, count)

	if len(points1) != len(points2) {
		t.Fatalf("point counts differ: %d vs %d", len(points1), len(points2))
	}

	// Compare coordinates
	for i := range points1 {
		if points1[i].X != points2[i].X || points1[i].Y != points2[i].Y {
			t.Errorf("point %d differs: (%v, %v) vs (%v, %v)",
				i, points1[i].X, points1[i].Y, points2[i].X, points2[i].Y)
		}
	}
}

// TestMerchantTypeString tests String() method.
func TestMerchantTypeString(t *testing.T) {
	tests := []struct {
		merchantType MerchantType
		want         string
	}{
		{MerchantFixed, "fixed"},
		{MerchantNomadic, "nomadic"},
		{MerchantType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.merchantType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateMerchantGenreVariety tests all genres produce valid merchants.
func TestGenerateMerchantGenreVariety(t *testing.T) {
	gen := NewEntityGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID:    genre,
				Difficulty: 0.5,
				Depth:      1,
			}

			merchant, err := gen.GenerateMerchant(12345, params, MerchantFixed)
			if err != nil {
				t.Fatalf("failed to generate %s merchant: %v", genre, err)
			}

			// Verify merchant has genre-appropriate name
			if merchant.Entity.Name == "" {
				t.Errorf("%s merchant has empty name", genre)
			}

			// Verify inventory exists
			if len(merchant.Inventory) == 0 {
				t.Errorf("%s merchant has empty inventory", genre)
			}

			// Log for manual inspection
			t.Logf("%s merchant: %s with %d items (price multiplier: %.2f)",
				genre, merchant.Entity.Name, len(merchant.Inventory), merchant.PriceMultiplier)
		})
	}
}

// BenchmarkGenerateMerchant benchmarks merchant generation.
func BenchmarkGenerateMerchant(b *testing.B) {
	gen := NewEntityGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Depth:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateMerchant(int64(i), params, MerchantFixed)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}

// BenchmarkGenerateMerchantSpawnPoints benchmarks spawn point generation.
func BenchmarkGenerateMerchantSpawnPoints(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateMerchantSpawnPoints(int64(i), 1000, 800, MerchantFixed, 5)
	}
}
