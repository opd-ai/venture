package tiles

import (
	"image"
	"testing"
)

func TestTileType_String(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		expected string
	}{
		{"Floor", TileFloor, "floor"},
		{"Wall", TileWall, "wall"},
		{"Door", TileDoor, "door"},
		{"Corridor", TileCorridor, "corridor"},
		{"Water", TileWater, "water"},
		{"Lava", TileLava, "lava"},
		{"Trap", TileTrap, "trap"},
		{"Stairs", TileStairs, "stairs"},
		{"Unknown", TileType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tileType.String()
			if got != tt.expected {
				t.Errorf("TileType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != TileFloor {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, TileFloor)
	}
	if config.Width != 32 {
		t.Errorf("DefaultConfig Width = %v, want 32", config.Width)
	}
	if config.Height != 32 {
		t.Errorf("DefaultConfig Height = %v, want 32", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want fantasy", config.GenreID)
	}
	if config.Variant != 0.5 {
		t.Errorf("DefaultConfig Variant = %v, want 0.5", config.Variant)
	}
	if config.Custom == nil {
		t.Error("DefaultConfig Custom should not be nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Zero width",
			config: Config{
				Type:    TileFloor,
				Width:   0,
				Height:  32,
				GenreID: "fantasy",
				Variant: 0.5,
			},
			wantErr: true,
		},
		{
			name: "Zero height",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  0,
				GenreID: "fantasy",
				Variant: 0.5,
			},
			wantErr: true,
		},
		{
			name: "Negative width",
			config: Config{
				Type:    TileFloor,
				Width:   -10,
				Height:  32,
				GenreID: "fantasy",
				Variant: 0.5,
			},
			wantErr: true,
		},
		{
			name: "Empty GenreID",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "",
				Variant: 0.5,
			},
			wantErr: true,
		},
		{
			name: "Variant too low",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Variant: -0.1,
			},
			wantErr: true,
		},
		{
			name: "Variant too high",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Variant: 1.1,
			},
			wantErr: true,
		},
		{
			name: "Variant at bounds - 0",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Variant: 0.0,
			},
			wantErr: false,
		},
		{
			name: "Variant at bounds - 1",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Variant: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPattern_String(t *testing.T) {
	tests := []struct {
		name     string
		pattern  Pattern
		expected string
	}{
		{"Solid", PatternSolid, "solid"},
		{"Checkerboard", PatternCheckerboard, "checkerboard"},
		{"Dots", PatternDots, "dots"},
		{"Lines", PatternLines, "lines"},
		{"Brick", PatternBrick, "brick"},
		{"Grain", PatternGrain, "grain"},
		{"Unknown", Pattern(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pattern.String()
			if got != tt.expected {
				t.Errorf("Pattern.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Generate floor tile",
			config: Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate wall tile",
			config: Config{
				Type:    TileWall,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate door tile",
			config: Config{
				Type:    TileDoor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate corridor tile",
			config: Config{
				Type:    TileCorridor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate water tile",
			config: Config{
				Type:    TileWater,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate lava tile",
			config: Config{
				Type:    TileLava,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate trap tile",
			config: Config{
				Type:    TileTrap,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate stairs tile",
			config: Config{
				Type:    TileStairs,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Invalid config - zero width",
			config: Config{
				Type:    TileFloor,
				Width:   0,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if img == nil {
					t.Error("Generator.Generate() returned nil image")
					return
				}
				bounds := img.Bounds()
				if bounds.Dx() != tt.config.Width {
					t.Errorf("Generated image width = %v, want %v", bounds.Dx(), tt.config.Width)
				}
				if bounds.Dy() != tt.config.Height {
					t.Errorf("Generated image height = %v, want %v", bounds.Dy(), tt.config.Height)
				}
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	// Generate the same tile twice
	img1, err1 := gen.Generate(config)
	img2, err2 := gen.Generate(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating tiles: %v, %v", err1, err2)
	}

	// Compare pixel by pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("Pixel at (%d, %d) differs: %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func TestGenerator_DifferentSeeds(t *testing.T) {
	gen := NewGenerator()

	config1 := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	config2 := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    54321,
		Variant: 0.5,
	}

	img1, err1 := gen.Generate(config1)
	img2, err2 := gen.Generate(config2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating tiles: %v, %v", err1, err2)
	}

	// Images should be different
	bounds := img1.Bounds()
	different := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !different; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				different = true
				break
			}
		}
	}

	if !different {
		t.Error("Tiles generated with different seeds should be different")
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name:    "Valid image",
			result:  image.NewRGBA(image.Rect(0, 0, 32, 32)),
			wantErr: false,
		},
		{
			name:    "Nil image",
			result:  (*image.RGBA)(nil),
			wantErr: true,
		},
		{
			name:    "Wrong type",
			result:  "not an image",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_AllGenres(t *testing.T) {
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: genre,
				Seed:    12345,
				Variant: 0.5,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate tile for genre %s: %v", genre, err)
			}
			if img == nil {
				t.Errorf("Generated nil image for genre %s", genre)
			}
		})
	}
}

func TestGenerator_VariantRange(t *testing.T) {
	gen := NewGenerator()

	variants := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, variant := range variants {
		t.Run("", func(t *testing.T) {
			config := Config{
				Type:    TileFloor,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: variant,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate tile with variant %f: %v", variant, err)
			}
			if img == nil {
				t.Errorf("Generated nil image with variant %f", variant)
			}
		})
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    TileFloor,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
		Seed:    12345,
		Variant: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_GenerateAllTypes(b *testing.B) {
	gen := NewGenerator()
	tileTypes := []TileType{
		TileFloor, TileWall, TileDoor, TileCorridor,
		TileWater, TileLava, TileTrap, TileStairs,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tileType := range tileTypes {
			config := Config{
				Type:    tileType,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    int64(i),
				Variant: 0.5,
			}
			_, err := gen.Generate(config)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
