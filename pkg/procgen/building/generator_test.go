package building

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestBuildingTypeString(t *testing.T) {
	tests := []struct {
		name string
		bt   BuildingType
		want string
	}{
		{"House", TypeHouse, "House"},
		{"Workshop", TypeWorkshop, "Workshop"},
		{"Storage", TypeStorage, "Storage"},
		{"Tower", TypeTower, "Tower"},
		{"Manor", TypeManor, "Manor"},
		{"GuildHall", TypeGuildHall, "GuildHall"},
		{"Unknown", BuildingType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bt.String(); got != tt.want {
				t.Errorf("BuildingType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchitecturalStyleString(t *testing.T) {
	tests := []struct {
		name  string
		style ArchitecturalStyle
		want  string
	}{
		{"Medieval", StyleMedieval, "Medieval"},
		{"Elven", StyleElven, "Elven"},
		{"Gothic", StyleGothic, "Gothic"},
		{"Neon", StyleNeon, "Neon"},
		{"Salvage", StyleSalvage, "Salvage"},
		{"Unknown", ArchitecturalStyle(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.String(); got != tt.want {
				t.Errorf("ArchitecturalStyle.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGenreStyles(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		want    int
	}{
		{"Fantasy", "fantasy", 5},
		{"SciFi", "scifi", 5},
		{"Horror", "horror", 5},
		{"Cyberpunk", "cyberpunk", 5},
		{"PostApoc", "postapoc", 5},
		{"Unknown", "unknown", 5}, // Fallback to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := GetGenreStyles(tt.genreID)
			if len(styles) != tt.want {
				t.Errorf("GetGenreStyles() returned %d styles, want %d", len(styles), tt.want)
			}
		})
	}
}

func TestRoomTypeString(t *testing.T) {
	tests := []struct {
		name string
		rt   RoomType
		want string
	}{
		{"Entrance", RoomEntrance, "Entrance"},
		{"Living", RoomLiving, "Living"},
		{"Bedroom", RoomBedroom, "Bedroom"},
		{"Storage", RoomStorage, "Storage"},
		{"Workshop", RoomWorkshop, "Workshop"},
		{"Kitchen", RoomKitchen, "Kitchen"},
		{"Hallway", RoomHallway, "Hallway"},
		{"Tower", RoomTower, "Tower"},
		{"Unknown", RoomType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rt.String(); got != tt.want {
				t.Errorf("RoomType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoorTypeString(t *testing.T) {
	tests := []struct {
		name string
		dt   DoorType
		want string
	}{
		{"Wooden", DoorWooden, "Wooden"},
		{"Metal", DoorMetal, "Metal"},
		{"Glass", DoorGlass, "Glass"},
		{"Secret", DoorSecret, "Secret"},
		{"Unknown", DoorType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.want {
				t.Errorf("DoorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindowTypeString(t *testing.T) {
	tests := []struct {
		name string
		wt   WindowType
		want string
	}{
		{"Small", WindowSmall, "Small"},
		{"Large", WindowLarge, "Large"},
		{"Stained", WindowStained, "Stained"},
		{"Broken", WindowBroken, "Broken"},
		{"Unknown", WindowType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wt.String(); got != tt.want {
				t.Errorf("WindowType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoofTypeString(t *testing.T) {
	tests := []struct {
		name string
		rt   RoofType
		want string
	}{
		{"Flat", RoofFlat, "Flat"},
		{"Gabled", RoofGabled, "Gabled"},
		{"Hipped", RoofHipped, "Hipped"},
		{"Domed", RoofDomed, "Domed"},
		{"Spire", RoofSpire, "Spire"},
		{"Unknown", RoofType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rt.String(); got != tt.want {
				t.Errorf("RoofType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildingGetRoomCount(t *testing.T) {
	building := &Building{
		Rooms: []Room{
			{Type: RoomEntrance},
			{Type: RoomLiving},
			{Type: RoomBedroom},
		},
	}

	if got := building.GetRoomCount(); got != 3 {
		t.Errorf("GetRoomCount() = %d, want 3", got)
	}
}

func TestBuildingValidate(t *testing.T) {
	tests := []struct {
		name     string
		building *Building
		wantErr  bool
	}{
		{
			name: "Valid house",
			building: &Building{
				Width:  10,
				Height: 10,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 5, Height: 10, Type: RoomEntrance},
					{X: 5, Y: 0, Width: 5, Height: 10, Type: RoomLiving},
				},
				Doors: []Door{
					{X: 5, Y: 5},
				},
			},
			wantErr: false,
		},
		{
			name: "Too small",
			building: &Building{
				Width:  3,
				Height: 3,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 3, Height: 3, Type: RoomEntrance},
				},
			},
			wantErr: true,
		},
		{
			name: "Too large",
			building: &Building{
				Width:  100,
				Height: 100,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 100, Height: 100, Type: RoomEntrance},
				},
			},
			wantErr: true,
		},
		{
			name: "No rooms",
			building: &Building{
				Width:  10,
				Height: 10,
				Rooms:  []Room{},
			},
			wantErr: true,
		},
		{
			name: "Too many rooms",
			building: &Building{
				Width:  10,
				Height: 10,
				Rooms:  make([]Room, 10),
			},
			wantErr: true,
		},
		{
			name: "No entrance",
			building: &Building{
				Width:  10,
				Height: 10,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 10, Height: 10, Type: RoomLiving},
				},
			},
			wantErr: true,
		},
		{
			name: "Overlapping rooms",
			building: &Building{
				Width:  10,
				Height: 10,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 6, Height: 10, Type: RoomEntrance},
					{X: 4, Y: 0, Width: 6, Height: 10, Type: RoomLiving},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.building.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Building.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildingIsNavigable(t *testing.T) {
	tests := []struct {
		name     string
		building *Building
		want     bool
	}{
		{
			name: "Navigable house",
			building: &Building{
				Rooms: []Room{
					{X: 0, Y: 0, Width: 5, Height: 10, Type: RoomEntrance},
					{X: 5, Y: 0, Width: 5, Height: 10, Type: RoomLiving},
				},
				Doors: []Door{
					{X: 5, Y: 5},
				},
			},
			want: true,
		},
		{
			name: "No entrance",
			building: &Building{
				Rooms: []Room{
					{X: 0, Y: 0, Width: 10, Height: 10, Type: RoomLiving},
				},
			},
			want: false,
		},
		{
			name: "Disconnected rooms",
			building: &Building{
				Rooms: []Room{
					{X: 0, Y: 0, Width: 5, Height: 10, Type: RoomEntrance},
					{X: 6, Y: 0, Width: 5, Height: 10, Type: RoomLiving},
				},
				Doors: []Door{},
			},
			want: false,
		},
		{
			name: "Empty building",
			building: &Building{
				Rooms: []Room{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.building.IsNavigable(); got != tt.want {
				t.Errorf("Building.IsNavigable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratorGenerate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name         string
		seed         int64
		buildingType BuildingType
		genreID      string
		wantRooms    int // Minimum expected rooms
	}{
		{"House fantasy", 12345, TypeHouse, "fantasy", 2},
		{"Workshop scifi", 67890, TypeWorkshop, "scifi", 2},
		{"Storage horror", 11111, TypeStorage, "horror", 1},
		{"Tower cyberpunk", 22222, TypeTower, "cyberpunk", 2},
		{"Manor postapoc", 33333, TypeManor, "postapoc", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID: tt.genreID,
				Custom: map[string]interface{}{
					"buildingType": tt.buildingType,
				},
			}

			result, err := gen.Generate(tt.seed, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			building, ok := result.(*Building)
			if !ok {
				t.Fatalf("Generate() returned wrong type: %T", result)
			}

			if building.Type != tt.buildingType {
				t.Errorf("Building type = %v, want %v", building.Type, tt.buildingType)
			}

			if building.GenreID != tt.genreID {
				t.Errorf("Building genre = %v, want %v", building.GenreID, tt.genreID)
			}

			if len(building.Rooms) < tt.wantRooms {
				t.Errorf("Room count = %d, want at least %d", len(building.Rooms), tt.wantRooms)
			}

			// Validate the building
			if err := gen.Validate(building); err != nil {
				t.Errorf("Generated building failed validation: %v", err)
			}
		})
	}
}

func TestGeneratorDeterminism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(42)
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeManor,
		},
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	result2, err2 := gen.Generate(seed, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("Generate() errors: %v, %v", err1, err2)
	}

	building1 := result1.(*Building)
	building2 := result2.(*Building)

	// Check key properties match
	if building1.Width != building2.Width {
		t.Errorf("Width mismatch: %d vs %d", building1.Width, building2.Width)
	}
	if building1.Height != building2.Height {
		t.Errorf("Height mismatch: %d vs %d", building1.Height, building2.Height)
	}
	if len(building1.Rooms) != len(building2.Rooms) {
		t.Errorf("Room count mismatch: %d vs %d", len(building1.Rooms), len(building2.Rooms))
	}
	if len(building1.Doors) != len(building2.Doors) {
		t.Errorf("Door count mismatch: %d vs %d", len(building1.Doors), len(building2.Doors))
	}
}

func TestGeneratorValidate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name: "Valid building",
			result: &Building{
				Width:  10,
				Height: 10,
				Rooms: []Room{
					{X: 0, Y: 0, Width: 5, Height: 10, Type: RoomEntrance},
					{X: 5, Y: 0, Width: 5, Height: 10, Type: RoomLiving},
				},
				Doors: []Door{
					{X: 5, Y: 5},
				},
			},
			wantErr: false,
		},
		{
			name:    "Wrong type",
			result:  "not a building",
			wantErr: true,
		},
		{
			name: "Invalid building",
			result: &Building{
				Width:  2,
				Height: 2,
				Rooms:  []Room{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerationTime(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeManor,
		},
	}

	// Test 100 generations
	for i := 0; i < 100; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			t.Fatalf("Generation %d failed: %v", i, err)
		}
	}
	// If we get here, all generations completed (implicitly under timeout)
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeHouse,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeHouse,
		},
	}

	result, _ := gen.Generate(12345, params)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Validate(result)
	}
}

func BenchmarkIsNavigable(b *testing.B) {
	building := &Building{
		Rooms: []Room{
			{X: 0, Y: 0, Width: 5, Height: 10, Type: RoomEntrance},
			{X: 5, Y: 0, Width: 5, Height: 10, Type: RoomLiving},
			{X: 10, Y: 0, Width: 5, Height: 10, Type: RoomBedroom},
		},
		Doors: []Door{
			{X: 5, Y: 5},
			{X: 10, Y: 5},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		building.IsNavigable()
	}
}

func TestGenerateGuildHall(t *testing.T) {
	gen := NewGenerator()
	tests := []struct {
		name    string
		seed    int64
		genreID string
		floors  int
		wantErr bool
	}{
		{"fantasy 1 floor", 12345, "fantasy", 1, false},
		{"scifi 3 floors", 67890, "scifi", 3, false},
		{"horror 5 floors", 11111, "horror", 5, false},
		{"cyberpunk 2 floors", 22222, "cyberpunk", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID: tt.genreID,
				Custom: map[string]interface{}{
					"buildingType": TypeGuildHall,
					"floors":       tt.floors,
				},
			}

			result, err := gen.Generate(tt.seed, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			building, ok := result.(*Building)
			if !ok {
				t.Fatalf("result is not *Building")
			}

			if building.Type != TypeGuildHall {
				t.Errorf("Type = %v, want %v", building.Type, TypeGuildHall)
			}

			if building.Floors != tt.floors {
				t.Errorf("Floors = %d, want %d", building.Floors, tt.floors)
			}

			if building.Width < 32 || building.Width > 64 {
				t.Errorf("Width = %d, want 32-64", building.Width)
			}

			if building.Height < 32 || building.Height > 64 {
				t.Errorf("Height = %d, want 32-64", building.Height)
			}

			// Verify multi-floor rooms
			if len(building.FloorRooms) != tt.floors {
				t.Errorf("FloorRooms count = %d, want %d", len(building.FloorRooms), tt.floors)
			}

			// Verify each floor has rooms
			for floor := 0; floor < tt.floors; floor++ {
				rooms, ok := building.FloorRooms[floor]
				if !ok {
					t.Errorf("Floor %d not found in FloorRooms", floor)
					continue
				}
				if len(rooms) == 0 {
					t.Errorf("Floor %d has no rooms", floor)
				}
			}

			// Guild halls should have many rooms
			// Main Rooms list contains ground floor (for validation)
			// FloorRooms contains all floors
			if len(building.Rooms) < 2 {
				t.Errorf("Ground floor (Rooms) count = %d, want >=2", len(building.Rooms))
			}

			// Total rooms across all floors should be many
			totalRooms := 0
			for _, floorRooms := range building.FloorRooms {
				totalRooms += len(floorRooms)
			}
			if totalRooms < 10 {
				t.Errorf("Total rooms across all floors = %d, want >=10", totalRooms)
			}
		})
	}
}

func TestGuildHallDeterminism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(42424)
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeGuildHall,
			"floors":       3,
		},
	}

	// Generate twice with same seed
	result1, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("first generation failed: %v", err)
	}

	result2, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("second generation failed: %v", err)
	}

	b1 := result1.(*Building)
	b2 := result2.(*Building)

	// Verify identical results
	if b1.Width != b2.Width || b1.Height != b2.Height {
		t.Errorf("Dimensions differ: (%d,%d) vs (%d,%d)", b1.Width, b1.Height, b2.Width, b2.Height)
	}

	if len(b1.Rooms) != len(b2.Rooms) {
		t.Errorf("Room count differs: %d vs %d", len(b1.Rooms), len(b2.Rooms))
	}

	if len(b1.Doors) != len(b2.Doors) {
		t.Errorf("Door count differs: %d vs %d", len(b1.Doors), len(b2.Doors))
	}

	if b1.Floors != b2.Floors {
		t.Errorf("Floor count differs: %d vs %d", b1.Floors, b2.Floors)
	}
}

func TestGuildHallValidation(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeGuildHall,
			"floors":       4,
		},
	}

	result, err := gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Validate should pass
	err = gen.Validate(result)
	if err != nil {
		t.Errorf("Validate() failed: %v", err)
	}
}

func BenchmarkGenerateGuildHall(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"buildingType": TypeGuildHall,
			"floors":       3,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}
