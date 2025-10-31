// Package tiles tests Phase 11.1 diagonal wall and multi-layer terrain rendering.
package tiles

import (
	"image/color"
	"testing"
)

func TestGenerateDiagonalWall_AllDirections(t *testing.T) {
	tests := []struct {
		name      string
		direction DiagonalDirection
		wantErr   bool
	}{
		{
			name:      "diagonal NE",
			direction: DiagonalNE,
			wantErr:   false,
		},
		{
			name:      "diagonal NW",
			direction: DiagonalNW,
			wantErr:   false,
		},
		{
			name:      "diagonal SE",
			direction: DiagonalSE,
			wantErr:   false,
		},
		{
			name:      "diagonal SW",
			direction: DiagonalSW,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			config := Config{
				Type:    TileWallNE, // Type doesn't matter for direct function call
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			}

			img, err := gen.Generate(Config{
				Type:    getTileTypeForDirection(tt.direction),
				Width:   config.Width,
				Height:  config.Height,
				GenreID: config.GenreID,
				Seed:    config.Seed,
				Variant: config.Variant,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("generateDiagonalWall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("generateDiagonalWall() returned nil image")
					return
				}

				bounds := img.Bounds()
				if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
					t.Errorf("image dimensions = %dx%d, want %dx%d",
						bounds.Dx(), bounds.Dy(), config.Width, config.Height)
				}

				// Verify diagonal pattern exists (should have mix of colors)
				hasVariation := false
				var firstColor color.Color
				for y := bounds.Min.Y; y < bounds.Max.Y && !hasVariation; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						c := img.At(x, y)
						if firstColor == nil {
							firstColor = c
						} else if c != firstColor {
							hasVariation = true
							break
						}
					}
				}

				if !hasVariation {
					t.Error("diagonal wall image has no color variation")
				}
			}
		})
	}
}

func getTileTypeForDirection(dir DiagonalDirection) TileType {
	switch dir {
	case DiagonalNE:
		return TileWallNE
	case DiagonalNW:
		return TileWallNW
	case DiagonalSE:
		return TileWallSE
	case DiagonalSW:
		return TileWallSW
	default:
		return TileWallNE
	}
}

func TestGeneratePlatform(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "basic platform",
			config: Config{
				Type:    TilePlatform,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
		{
			name: "large platform",
			config: Config{
				Type:    TilePlatform,
				Width:   64,
				Height:  64,
				GenreID: "scifi", // Fixed: should be "scifi" not "sci-fi"
				Seed:    54321,
				Variant: 0.8,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			img, err := gen.Generate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("generatePlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("generatePlatform() returned nil image")
					return
				}

				bounds := img.Bounds()
				if bounds.Dx() != tt.config.Width || bounds.Dy() != tt.config.Height {
					t.Errorf("image dimensions = %dx%d, want %dx%d",
						bounds.Dx(), bounds.Dy(), tt.config.Width, tt.config.Height)
				}

				// Verify edge effects (should have different colors at edges)
				topLeftColor := img.At(bounds.Min.X+1, bounds.Min.Y+1)
				bottomRightColor := img.At(bounds.Max.X-2, bounds.Max.Y-2)

				if topLeftColor == bottomRightColor {
					t.Error("platform lacks 3D edge effect (top-left and bottom-right should differ)")
				}
			}
		})
	}
}

func TestGenerateRamp(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "basic ramp",
			config: Config{
				Type:    TileRamp,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			img, err := gen.Generate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateRamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("generateRamp() returned nil image")
					return
				}

				bounds := img.Bounds()

				// Verify gradient (bottom should be lighter than top)
				topY := bounds.Min.Y + 5
				bottomY := bounds.Max.Y - 5
				centerX := bounds.Min.X + bounds.Dx()/2

				topColor := img.At(centerX, topY)
				bottomColor := img.At(centerX, bottomY)

				tr, tg, tb, _ := topColor.RGBA()
				br, bg, bb, _ := bottomColor.RGBA()

				// Bottom should be lighter (higher RGB values)
				// Use 16-bit values for comparison (RGBA returns 16-bit)
				if br <= tr || bg <= tg || bb <= tb {
					t.Errorf("ramp gradient incorrect: top(%d,%d,%d) bottom(%d,%d,%d) - bottom should be lighter",
						tr, tg, tb, br, bg, bb)
				}
			}
		})
	}
}

func TestGeneratePit(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "basic pit",
			config: Config{
				Type:    TilePit,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			img, err := gen.Generate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("generatePit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if img == nil {
					t.Error("generatePit() returned nil image")
					return
				}

				bounds := img.Bounds()

				// Verify pit is dark (center should be darker than edges)
				centerColor := img.At(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)
				edgeColor := img.At(bounds.Min.X+1, bounds.Min.Y+1)

				cr, cg, cb, _ := centerColor.RGBA()
				er, eg, eb, _ := edgeColor.RGBA()

				// Center should be darker (lower RGB values)
				if cr > er || cg > eg || cb > eb {
					t.Error("pit vignette incorrect (center should be darker than edges)")
				}
			}
		})
	}
}

func TestDeterministicRendering_Phase11(t *testing.T) {
	// Test that Phase 11.1 tiles render deterministically with same seed
	gen := NewGenerator()

	tileTypes := []TileType{
		TileWallNE,
		TileWallNW,
		TileWallSE,
		TileWallSW,
		TilePlatform,
		TileRamp,
		TilePit,
	}

	for _, tileType := range tileTypes {
		t.Run(tileType.String(), func(t *testing.T) {
			config := Config{
				Type:    tileType,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
				Variant: 0.5,
			}

			// Generate twice with same seed
			img1, err1 := gen.Generate(config)
			if err1 != nil {
				t.Fatalf("first generation failed: %v", err1)
			}

			img2, err2 := gen.Generate(config)
			if err2 != nil {
				t.Fatalf("second generation failed: %v", err2)
			}

			// Compare images pixel by pixel
			bounds := img1.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					c1 := img1.At(x, y)
					c2 := img2.At(x, y)

					if c1 != c2 {
						t.Errorf("non-deterministic rendering at (%d,%d): %v != %v", x, y, c1, c2)
						return
					}
				}
			}
		})
	}
}

func TestIsInsideTriangle(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name     string
		px, py   int
		x1, y1   int
		x2, y2   int
		x3, y3   int
		expected bool
	}{
		{
			name: "point inside triangle",
			px:   5, py: 5,
			x1: 0, y1: 0,
			x2: 10, y2: 0,
			x3: 5, y3: 10,
			expected: true,
		},
		{
			name: "point outside triangle",
			px:   15, py: 15,
			x1: 0, y1: 0,
			x2: 10, y2: 0,
			x3: 5, y3: 10,
			expected: false,
		},
		{
			name: "point on vertex",
			px:   0, py: 0,
			x1: 0, y1: 0,
			x2: 10, y2: 0,
			x3: 5, y3: 10,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.isInsideTriangle(tt.px, tt.py, tt.x1, tt.y1, tt.x2, tt.y2, tt.x3, tt.y3)
			if result != tt.expected {
				t.Errorf("isInsideTriangle() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenreVariation_Phase11(t *testing.T) {
	// Test that different genres produce different visual styles
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	config := Config{
		Type:    TilePlatform,
		Width:   32,
		Height:  32,
		Seed:    12345,
		Variant: 0.5,
	}

	images := make(map[string]color.Color)

	for _, genre := range genres {
		config.GenreID = genre
		img, err := gen.Generate(config)
		if err != nil {
			t.Fatalf("failed to generate for genre %s: %v", genre, err)
		}

		// Sample center pixel
		centerColor := img.At(16, 16)
		images[genre] = centerColor
	}

	// Verify that at least some genres produce different colors
	allSame := true
	var firstColor color.Color
	for _, c := range images {
		if firstColor == nil {
			firstColor = c
		} else if c != firstColor {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("all genres produced identical colors (expected variation)")
	}
}
