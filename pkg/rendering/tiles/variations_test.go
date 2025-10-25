package tiles

import (
	"testing"
)

func TestGenerateVariations(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  Config
		count   int
		wantErr bool
	}{
		{
			name: "5 floor variations",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			},
			count:   5,
			wantErr: false,
		},
		{
			name: "8 wall variations",
			config: Config{
				Type:    TileWall,
				Width:   32,
				Height:  32,
				GenreID: "scifi",
				Seed:    54321,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			},
			count:   8,
			wantErr: false,
		},
		{
			name: "invalid count",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			},
			count:   0,
			wantErr: true,
		},
		{
			name: "invalid config",
			config: Config{
				Type:    TileFloor,
				Width:   -1,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			},
			count:   5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varSet, err := gen.GenerateVariations(tt.config, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateVariations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if varSet == nil {
				t.Fatal("GenerateVariations() returned nil without error")
			}

			if varSet.Count != tt.count {
				t.Errorf("Count = %d, want %d", varSet.Count, tt.count)
			}

			if len(varSet.Variations) != tt.count {
				t.Errorf("len(Variations) = %d, want %d", len(varSet.Variations), tt.count)
			}

			if varSet.Type != tt.config.Type {
				t.Errorf("Type = %v, want %v", varSet.Type, tt.config.Type)
			}

			// Verify all variations are valid
			for i, img := range varSet.Variations {
				if img == nil {
					t.Errorf("Variation %d is nil", i)
					continue
				}

				bounds := img.Bounds()
				if bounds.Dx() != tt.config.Width || bounds.Dy() != tt.config.Height {
					t.Errorf("Variation %d size = %dx%d, want %dx%d",
						i, bounds.Dx(), bounds.Dy(), tt.config.Width, tt.config.Height)
				}
			}
		})
	}
}

func TestVariationSet_GetVariation(t *testing.T) {
	gen := NewGenerator()

	config := DefaultConfig()
	config.Seed = 12345
	varSet, err := gen.GenerateVariations(config, 5)
	if err != nil {
		t.Fatalf("GenerateVariations() error = %v", err)
	}

	tests := []struct {
		name    string
		index   int
		wantErr bool
	}{
		{"first variation", 0, false},
		{"middle variation", 2, false},
		{"last variation", 4, false},
		{"negative index", -1, true},
		{"index too high", 5, true},
		{"index way too high", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := varSet.GetVariation(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVariation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("GetVariation() returned nil without error")
			}
		})
	}
}

func TestVariationSet_GetVariationBySeed(t *testing.T) {
	gen := NewGenerator()

	config := DefaultConfig()
	config.Seed = 12345
	varSet, err := gen.GenerateVariations(config, 5)
	if err != nil {
		t.Fatalf("GenerateVariations() error = %v", err)
	}

	tests := []struct {
		name string
		seed int64
	}{
		{"seed 0", 0},
		{"seed 1", 1},
		{"seed 100", 100},
		{"seed 12345", 12345},
		{"negative seed", -42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := varSet.GetVariationBySeed(tt.seed)
			if img == nil {
				t.Error("GetVariationBySeed() returned nil")
			}

			// Verify determinism: same seed returns same variation
			img2 := varSet.GetVariationBySeed(tt.seed)
			if img != img2 {
				t.Error("GetVariationBySeed() not deterministic")
			}
		})
	}
}

func TestGenerateTileSet(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name           string
		genreID        string
		seed           int64
		tileSize       int
		variationCount int
		wantErr        bool
	}{
		{
			name:           "fantasy 5 variations",
			genreID:        "fantasy",
			seed:           12345,
			tileSize:       32,
			variationCount: 5,
			wantErr:        false,
		},
		{
			name:           "scifi 8 variations",
			genreID:        "scifi",
			seed:           54321,
			tileSize:       32,
			variationCount: 8,
			wantErr:        false,
		},
		{
			name:           "horror 6 variations",
			genreID:        "horror",
			seed:           11111,
			tileSize:       32,
			variationCount: 6,
			wantErr:        false,
		},
		{
			name:           "invalid tile size",
			genreID:        "fantasy",
			seed:           12345,
			tileSize:       0,
			variationCount: 5,
			wantErr:        true,
		},
		{
			name:           "invalid variation count",
			genreID:        "fantasy",
			seed:           12345,
			tileSize:       32,
			variationCount: 0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tileSet, err := gen.GenerateTileSet(tt.genreID, tt.seed, tt.tileSize, tt.variationCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTileSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tileSet == nil {
				t.Fatal("GenerateTileSet() returned nil without error")
			}

			if tileSet.GenreID != tt.genreID {
				t.Errorf("GenreID = %s, want %s", tileSet.GenreID, tt.genreID)
			}

			if tileSet.Seed != tt.seed {
				t.Errorf("Seed = %d, want %d", tileSet.Seed, tt.seed)
			}

			if tileSet.TileSize != tt.tileSize {
				t.Errorf("TileSize = %d, want %d", tileSet.TileSize, tt.tileSize)
			}

			if tileSet.VariantCount != tt.variationCount {
				t.Errorf("VariantCount = %d, want %d", tileSet.VariantCount, tt.variationCount)
			}

			// Verify all tile types are present
			expectedTypes := []TileType{
				TileFloor, TileWall, TileDoor, TileCorridor,
				TileWater, TileLava, TileTrap, TileStairs,
			}

			for _, tileType := range expectedTypes {
				varSet, ok := tileSet.Variations[tileType]
				if !ok {
					t.Errorf("Missing variations for tile type %s", tileType)
					continue
				}

				if varSet.Count != tt.variationCount {
					t.Errorf("Tile type %s has %d variations, want %d",
						tileType, varSet.Count, tt.variationCount)
				}
			}
		})
	}
}

func TestTileSet_GetTile(t *testing.T) {
	gen := NewGenerator()

	tileSet, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		t.Fatalf("GenerateTileSet() error = %v", err)
	}

	tests := []struct {
		name     string
		tileType TileType
		x, y     int
		wantErr  bool
	}{
		{"floor at 0,0", TileFloor, 0, 0, false},
		{"wall at 5,3", TileWall, 5, 3, false},
		{"door at 10,10", TileDoor, 10, 10, false},
		{"water at 100,200", TileWater, 100, 200, false},
		{"invalid type", TileType(999), 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := tileSet.GetTile(tt.tileType, tt.x, tt.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("GetTile() returned nil without error")
			}

			if !tt.wantErr {
				// Verify determinism: same position returns same tile
				img2, err := tileSet.GetTile(tt.tileType, tt.x, tt.y)
				if err != nil {
					t.Errorf("Second GetTile() error = %v", err)
				}
				if img != img2 {
					t.Error("GetTile() not deterministic for same position")
				}

				// Different positions should potentially return different variations
				img3, err := tileSet.GetTile(tt.tileType, tt.x+1, tt.y+1)
				if err != nil {
					t.Errorf("Third GetTile() error = %v", err)
				}
				if img3 == nil {
					t.Error("GetTile() at different position returned nil")
				}
			}
		})
	}
}

func TestTileSet_GetVariationSet(t *testing.T) {
	gen := NewGenerator()

	tileSet, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		t.Fatalf("GenerateTileSet() error = %v", err)
	}

	tests := []struct {
		name     string
		tileType TileType
		wantErr  bool
	}{
		{"floor variations", TileFloor, false},
		{"wall variations", TileWall, false},
		{"door variations", TileDoor, false},
		{"stairs variations", TileStairs, false},
		{"invalid type", TileType(999), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varSet, err := tileSet.GetVariationSet(tt.tileType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVariationSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && varSet == nil {
				t.Error("GetVariationSet() returned nil without error")
			}
		})
	}
}

func TestValidateTileSet(t *testing.T) {
	gen := NewGenerator()

	validSet, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		t.Fatalf("GenerateTileSet() error = %v", err)
	}

	tests := []struct {
		name    string
		tileSet *TileSet
		wantErr bool
	}{
		{
			name:    "valid tile set",
			tileSet: validSet,
			wantErr: false,
		},
		{
			name:    "nil tile set",
			tileSet: nil,
			wantErr: true,
		},
		{
			name: "empty genreID",
			tileSet: &TileSet{
				GenreID:      "",
				Seed:         12345,
				TileSize:     32,
				VariantCount: 5,
				Variations:   make(map[TileType]*VariationSet),
			},
			wantErr: true,
		},
		{
			name: "invalid tile size",
			tileSet: &TileSet{
				GenreID:      "fantasy",
				Seed:         12345,
				TileSize:     0,
				VariantCount: 5,
				Variations:   make(map[TileType]*VariationSet),
			},
			wantErr: true,
		},
		{
			name: "invalid variant count",
			tileSet: &TileSet{
				GenreID:      "fantasy",
				Seed:         12345,
				TileSize:     32,
				VariantCount: 0,
				Variations:   make(map[TileType]*VariationSet),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTileSet(tt.tileSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTileSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTileVariationDeterminism(t *testing.T) {
	gen := NewGenerator()

	config := DefaultConfig()
	config.Seed = 12345

	// Generate twice with same parameters
	varSet1, err := gen.GenerateVariations(config, 5)
	if err != nil {
		t.Fatalf("First GenerateVariations() error = %v", err)
	}

	varSet2, err := gen.GenerateVariations(config, 5)
	if err != nil {
		t.Fatalf("Second GenerateVariations() error = %v", err)
	}

	if varSet1.Count != varSet2.Count {
		t.Errorf("Count differs: %d vs %d", varSet1.Count, varSet2.Count)
	}

	// Compare each variation
	for i := 0; i < varSet1.Count; i++ {
		img1 := varSet1.Variations[i]
		img2 := varSet2.Variations[i]

		if img1 == nil || img2 == nil {
			t.Errorf("Variation %d is nil", i)
			continue
		}

		// Compare pixel-by-pixel
		bounds1 := img1.Bounds()
		bounds2 := img2.Bounds()

		if bounds1 != bounds2 {
			t.Errorf("Variation %d bounds differ: %v vs %v", i, bounds1, bounds2)
			continue
		}

		pixelsDiffer := false
		for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
			for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
				r1, g1, b1, a1 := img1.At(x, y).RGBA()
				r2, g2, b2, a2 := img2.At(x, y).RGBA()

				if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
					pixelsDiffer = true
					break
				}
			}
			if pixelsDiffer {
				break
			}
		}

		if pixelsDiffer {
			t.Errorf("Variation %d pixels differ for same seed", i)
		}
	}
}

func TestTileSetDeterminism(t *testing.T) {
	gen := NewGenerator()

	// Generate twice with same parameters
	set1, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		t.Fatalf("First GenerateTileSet() error = %v", err)
	}

	set2, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		t.Fatalf("Second GenerateTileSet() error = %v", err)
	}

	// Verify same tile at same position returns identical image
	tile1, err := set1.GetTile(TileFloor, 5, 10)
	if err != nil {
		t.Fatalf("First GetTile() error = %v", err)
	}

	tile2, err := set2.GetTile(TileFloor, 5, 10)
	if err != nil {
		t.Fatalf("Second GetTile() error = %v", err)
	}

	// Compare tiles pixel-by-pixel
	bounds1 := tile1.Bounds()
	bounds2 := tile2.Bounds()

	if bounds1 != bounds2 {
		t.Errorf("Tile bounds differ: %v vs %v", bounds1, bounds2)
	}

	pixelsDiffer := false
	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := tile1.At(x, y).RGBA()
			r2, g2, b2, a2 := tile2.At(x, y).RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				pixelsDiffer = true
				break
			}
		}
		if pixelsDiffer {
			break
		}
	}

	if pixelsDiffer {
		t.Error("Tiles differ for same seed and position")
	}
}

// Benchmarks

func BenchmarkGenerateVariations(b *testing.B) {
	gen := NewGenerator()
	config := DefaultConfig()
	config.Seed = 12345

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateVariations(config, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateTileSet(b *testing.B) {
	gen := NewGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetTile(b *testing.B) {
	gen := NewGenerator()
	tileSet, err := gen.GenerateTileSet("fantasy", 12345, 32, 5)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tileSet.GetTile(TileFloor, i%100, i%100)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetVariationBySeed(b *testing.B) {
	gen := NewGenerator()
	config := DefaultConfig()
	varSet, err := gen.GenerateVariations(config, 5)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = varSet.GetVariationBySeed(int64(i))
	}
}
