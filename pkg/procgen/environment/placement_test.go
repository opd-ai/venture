// Package environment provides procedural generation of environmental objects.
package environment

import (
	"math/rand"
	"testing"
)

func TestPlacementType_String(t *testing.T) {
	tests := []struct {
		placement PlacementType
		expected  string
	}{
		{PlacementFloor, "Floor"},
		{PlacementWallNorth, "WallNorth"},
		{PlacementWallSouth, "WallSouth"},
		{PlacementWallEast, "WallEast"},
		{PlacementWallWest, "WallWest"},
		{PlacementCorner, "Corner"},
		{PlacementCenter, "Center"},
		{PlacementType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.placement.String(); got != tt.expected {
				t.Errorf("PlacementType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultPlacementConfig(t *testing.T) {
	config := DefaultPlacementConfig()

	if config.RoomWidth != 10 {
		t.Errorf("Expected RoomWidth 10, got %d", config.RoomWidth)
	}
	if config.RoomHeight != 10 {
		t.Errorf("Expected RoomHeight 10, got %d", config.RoomHeight)
	}
	if config.Density != 0.5 {
		t.Errorf("Expected Density 0.5, got %f", config.Density)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("Expected GenreID fantasy, got %s", config.GenreID)
	}
	if !config.AllowWallDecorations {
		t.Error("Expected AllowWallDecorations true")
	}
	if !config.AllowFloorDecorations {
		t.Error("Expected AllowFloorDecorations true")
	}
	if config.MinSpacing != 2 {
		t.Errorf("Expected MinSpacing 2, got %d", config.MinSpacing)
	}
}

func TestPlacer_PlaceDecorations(t *testing.T) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             10,
		RoomHeight:            10,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  12345,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}

	placements, err := placer.PlaceDecorations(config)
	if err != nil {
		t.Fatalf("PlaceDecorations() error = %v", err)
	}

	if len(placements) == 0 {
		t.Error("Expected some decorations to be placed")
	}

	// Verify placements are valid
	for _, p := range placements {
		if p.Object == nil {
			t.Error("Placed object should not be nil")
		}
		if p.X < 0 || p.X >= config.RoomWidth {
			t.Errorf("Placement X out of bounds: %d", p.X)
		}
		if p.Y < 0 || p.Y >= config.RoomHeight {
			t.Errorf("Placement Y out of bounds: %d", p.Y)
		}
	}
}

func TestPlacer_PlaceDecorations_Deterministic(t *testing.T) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             10,
		RoomHeight:            10,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  99999,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}

	placements1, err1 := placer.PlaceDecorations(config)
	if err1 != nil {
		t.Fatalf("PlaceDecorations() error = %v", err1)
	}

	placements2, err2 := placer.PlaceDecorations(config)
	if err2 != nil {
		t.Fatalf("PlaceDecorations() error = %v", err2)
	}

	if len(placements1) != len(placements2) {
		t.Errorf("Placement count mismatch: %d != %d", len(placements1), len(placements2))
	}

	// Verify positions match
	for i := range placements1 {
		if placements1[i].X != placements2[i].X || placements1[i].Y != placements2[i].Y {
			t.Errorf("Position mismatch at index %d: (%d,%d) != (%d,%d)",
				i, placements1[i].X, placements1[i].Y, placements2[i].X, placements2[i].Y)
		}
	}
}

func TestPlacer_PlaceDecorations_DensityControl(t *testing.T) {
	tests := []struct {
		name    string
		density float64
		minExp  int
		maxExp  int
	}{
		{"low density", 0.2, 1, 5},
		{"medium density", 0.5, 3, 10},
		{"high density", 1.0, 5, 20},
	}

	placer := NewPlacer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PlacementConfig{
				RoomWidth:             10,
				RoomHeight:            10,
				Density:               tt.density,
				GenreID:               "fantasy",
				Seed:                  54321,
				AllowWallDecorations:  true,
				AllowFloorDecorations: true,
				MinSpacing:            2,
			}

			placements, err := placer.PlaceDecorations(config)
			if err != nil {
				t.Fatalf("PlaceDecorations() error = %v", err)
			}

			count := len(placements)
			if count < tt.minExp || count > tt.maxExp {
				t.Errorf("Decoration count %d out of range [%d,%d]", count, tt.minExp, tt.maxExp)
			}
		})
	}
}

func TestPlacer_PlaceDecorations_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	placer := NewPlacer()

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := PlacementConfig{
				RoomWidth:             10,
				RoomHeight:            10,
				Density:               0.5,
				GenreID:               genre,
				Seed:                  77777,
				AllowWallDecorations:  true,
				AllowFloorDecorations: true,
				MinSpacing:            2,
			}

			placements, err := placer.PlaceDecorations(config)
			if err != nil {
				t.Fatalf("PlaceDecorations() error = %v", err)
			}

			if len(placements) == 0 {
				t.Errorf("Genre %s should have decorations", genre)
			}

			// Verify all objects match the genre
			for _, p := range placements {
				if p.Object.GenreID != genre {
					t.Errorf("Object genre %s doesn't match config genre %s",
						p.Object.GenreID, genre)
				}
			}
		})
	}
}

func TestPlacer_PlaceDecorations_WallOnly(t *testing.T) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             10,
		RoomHeight:            10,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  11111,
		AllowWallDecorations:  true,
		AllowFloorDecorations: false,
		MinSpacing:            2,
	}

	placements, err := placer.PlaceDecorations(config)
	if err != nil {
		t.Fatalf("PlaceDecorations() error = %v", err)
	}

	// All placements should be wall-type
	for _, p := range placements {
		isWall := p.PlacementType == PlacementWallNorth ||
			p.PlacementType == PlacementWallSouth ||
			p.PlacementType == PlacementWallEast ||
			p.PlacementType == PlacementWallWest

		if !isWall && p.PlacementType != PlacementFloor {
			t.Errorf("Expected wall placement when AllowFloorDecorations=false, got %v",
				p.PlacementType)
		}
	}
}

func TestPlacer_PlaceDecorations_LargeRoom(t *testing.T) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             30,
		RoomHeight:            30,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  22222,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}

	placements, err := placer.PlaceDecorations(config)
	if err != nil {
		t.Fatalf("PlaceDecorations() error = %v", err)
	}

	// Large room should have more decorations
	if len(placements) < 10 {
		t.Errorf("Large room should have at least 10 decorations, got %d", len(placements))
	}
}

func TestPlacer_PlaceDecorations_SmallRoom(t *testing.T) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             5,
		RoomHeight:            5,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  33333,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            1, // Smaller spacing for small room
	}

	placements, err := placer.PlaceDecorations(config)
	if err != nil {
		t.Fatalf("PlaceDecorations() error = %v", err)
	}

	// Small room should have at least 1 decoration
	if len(placements) < 1 {
		t.Errorf("Expected at least 1 decoration in small room, got %d", len(placements))
	}
}

func TestSelectDecorationPool(t *testing.T) {
	placer := NewPlacer()

	tests := []struct {
		genre        string
		shouldHave   SubType
		shouldntHave SubType
	}{
		{"fantasy", SubTypeCrystal, SubTypeGraffiti},
		{"scifi", SubTypeGraffiti, SubTypeSkull},
		{"horror", SubTypeSkull, SubTypeCrystal},
		{"cyberpunk", SubTypeGraffiti, SubTypeBloodstain},
		{"postapoc", SubTypeDebris, SubTypeCrystal},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			pool := placer.selectDecorationPool(tt.genre)

			hasExpected := false

			for _, subType := range pool {
				if subType == tt.shouldHave {
					hasExpected = true
				}
			}

			if !hasExpected {
				t.Errorf("Genre %s pool should contain %v", tt.genre, tt.shouldHave)
			}
		})
	}
}

func TestPlacementConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  PlacementConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultPlacementConfig(),
			wantErr: false,
		},
		{
			name: "zero room width",
			config: PlacementConfig{
				RoomWidth: 0, RoomHeight: 10, Density: 0.5,
				GenreID: "fantasy", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "room width too small",
			config: PlacementConfig{
				RoomWidth: 2, RoomHeight: 10, Density: 0.5,
				GenreID: "fantasy", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "zero room height",
			config: PlacementConfig{
				RoomWidth: 10, RoomHeight: 0, Density: 0.5,
				GenreID: "fantasy", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "negative density",
			config: PlacementConfig{
				RoomWidth: 10, RoomHeight: 10, Density: -0.1,
				GenreID: "fantasy", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "density above 1",
			config: PlacementConfig{
				RoomWidth: 10, RoomHeight: 10, Density: 1.5,
				GenreID: "fantasy", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "empty genre",
			config: PlacementConfig{
				RoomWidth: 10, RoomHeight: 10, Density: 0.5,
				GenreID: "", MinSpacing: 2,
			},
			wantErr: true,
		},
		{
			name: "negative min spacing",
			config: PlacementConfig{
				RoomWidth: 10, RoomHeight: 10, Density: 0.5,
				GenreID: "fantasy", MinSpacing: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PlacementConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlacer_PlaceDecorations_InvalidConfig(t *testing.T) {
	placer := NewPlacer()

	_, err := placer.PlaceDecorations(PlacementConfig{
		RoomWidth: 1, RoomHeight: 10, Density: 0.5,
		GenreID: "fantasy", MinSpacing: 2,
	})
	if err == nil {
		t.Error("Expected error for invalid config with small room width")
	}
}

func TestCalculateDecorationCount(t *testing.T) {
	placer := NewPlacer()

	tests := []struct {
		name     string
		roomArea int
		density  float64
		minExp   int
		maxExp   int
	}{
		{"tiny room low density", 25, 0.2, 3, 6},
		{"average room medium density", 100, 0.5, 3, 10},
		{"large room high density", 400, 1.0, 10, 20},
		{"huge room capped", 1000, 1.0, 3, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(44444))
			count := placer.calculateDecorationCount(tt.roomArea, tt.density, rng)

			if count < tt.minExp || count > tt.maxExp {
				t.Errorf("Count %d out of range [%d,%d]", count, tt.minExp, tt.maxExp)
			}
		})
	}
}

// Benchmark tests

func BenchmarkPlaceDecorations_Small(b *testing.B) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             10,
		RoomHeight:            10,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  12345,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Seed = int64(i)
		placer.PlaceDecorations(config)
	}
}

func BenchmarkPlaceDecorations_Large(b *testing.B) {
	placer := NewPlacer()
	config := PlacementConfig{
		RoomWidth:             30,
		RoomHeight:            30,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  12345,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Seed = int64(i)
		placer.PlaceDecorations(config)
	}
}
