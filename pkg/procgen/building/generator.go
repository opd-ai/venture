package building

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// Generator implements procgen.Generator for building generation
type Generator struct{}

// NewGenerator creates a new building generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a procedural building
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Extract custom parameters
	buildingType := TypeHouse // Default
	if val, ok := params.Custom["buildingType"]; ok {
		if bt, ok := val.(BuildingType); ok {
			buildingType = bt
		} else if btInt, ok := val.(int); ok {
			buildingType = BuildingType(btInt)
		}
	}

	// Extract floor count for multi-floor buildings
	floors := 1
	if val, ok := params.Custom["floors"]; ok {
		if f, ok := val.(int); ok && f >= 1 && f <= 5 {
			floors = f
		}
	}

	// Determine architectural style based on genre and building type
	style := GetStyleForGenreAndType(params.GenreID, buildingType, rng)

	// Generate dimensions based on building type
	width, height := g.generateDimensions(buildingType, rng)

	// Create building
	building := &Building{
		Type:       buildingType,
		Style:      style,
		GenreID:    params.GenreID,
		Width:      width,
		Height:     height,
		Floors:     floors,
		Rooms:      []Room{},
		Doors:      []Door{},
		Windows:    []Window{},
		FloorRooms: make(map[int][]Room),
	}

	// Generate floor plan (potentially multi-floor)
	if err := g.generateFloorPlan(building, rng); err != nil {
		return nil, fmt.Errorf("floor plan generation failed: %w", err)
	}

	// Generate roof
	building.RoofType = g.generateRoof(buildingType, style, rng)

	// Place windows
	g.placeWindows(building, rng)

	return building, nil
}

// Validate checks if the generated building meets quality standards
func (g *Generator) Validate(result interface{}) error {
	building, ok := result.(*Building)
	if !ok {
		return fmt.Errorf("invalid result type: expected *Building, got %T", result)
	}

	return building.Validate()
}

// generateDimensions creates appropriate dimensions for a building type
func (g *Generator) generateDimensions(buildingType BuildingType, rng *rand.Rand) (int, int) {
	switch buildingType {
	case TypeHouse:
		// Small to medium houses
		return 8 + rng.Intn(8), 8 + rng.Intn(8) // 8-15 tiles
	case TypeWorkshop:
		// Wide, medium height
		return 12 + rng.Intn(8), 8 + rng.Intn(4) // 12-19 x 8-11 tiles
	case TypeStorage:
		// Large and boxy
		return 10 + rng.Intn(10), 10 + rng.Intn(10) // 10-19 tiles
	case TypeTower:
		// Tall and narrow
		return 6 + rng.Intn(4), 16 + rng.Intn(16) // 6-9 x 16-31 tiles
	case TypeManor:
		// Large and grand
		return 20 + rng.Intn(20), 16 + rng.Intn(16) // 20-39 x 16-31 tiles
	case TypeGuildHall:
		// Very large: 32-64 tiles square
		return 32 + rng.Intn(33), 32 + rng.Intn(33) // 32-64 tiles
	default:
		return 10, 10
	}
}

// generateFloorPlan creates rooms and connections
func (g *Generator) generateFloorPlan(building *Building, rng *rand.Rand) error {
	// Determine room count based on building type and size
	roomCount := g.calculateRoomCount(building, rng)

	// Generate rooms using recursive subdivision or grid placement
	switch building.Type {
	case TypeHouse:
		return g.generateHouseLayout(building, roomCount, rng)
	case TypeWorkshop:
		return g.generateWorkshopLayout(building, roomCount, rng)
	case TypeStorage:
		return g.generateStorageLayout(building, roomCount, rng)
	case TypeTower:
		return g.generateTowerLayout(building, roomCount, rng)
	case TypeManor:
		return g.generateManorLayout(building, roomCount, rng)
	case TypeGuildHall:
		return g.generateGuildHallLayout(building, roomCount, rng)
	default:
		return g.generateHouseLayout(building, roomCount, rng)
	}
}

// calculateRoomCount determines how many rooms to generate
func (g *Generator) calculateRoomCount(building *Building, rng *rand.Rand) int {
	area := building.Width * building.Height

	switch building.Type {
	case TypeHouse:
		// 2-4 rooms for houses
		return 2 + rng.Intn(3)
	case TypeWorkshop:
		// 2-3 rooms (main workshop + storage)
		return 2 + rng.Intn(2)
	case TypeStorage:
		// 1-2 large rooms
		return 1 + rng.Intn(2)
	case TypeTower:
		// Vertical rooms based on height
		floors := building.Height / 6
		if floors < 2 {
			floors = 2
		}
		if floors > 8 {
			floors = 8
		}
		return floors
	case TypeManor:
		// 4-8 rooms for manors
		baseRooms := 4
		bonusRooms := (area - 320) / 100 // Add room per 100 tiles above minimum
		if bonusRooms < 0 {
			bonusRooms = 0
		}
		if bonusRooms > 4 {
			bonusRooms = 4
		}
		return baseRooms + bonusRooms
	case TypeGuildHall:
		// 10-20 rooms per floor for guild halls
		baseRooms := 10 * building.Floors
		bonusRooms := (area - 1024) / 200 // Add room per 200 tiles above minimum
		if bonusRooms < 0 {
			bonusRooms = 0
		}
		maxBonus := 10 * building.Floors
		if bonusRooms > maxBonus {
			bonusRooms = maxBonus
		}
		return baseRooms + bonusRooms
	default:
		return 1 + rng.Intn(3)
	}
}

// generateHouseLayout creates a simple house floor plan
func (g *Generator) generateHouseLayout(building *Building, roomCount int, rng *rand.Rand) error {
	// Simple horizontal split for houses
	remaining := building.Width
	x := 0

	for i := 0; i < roomCount && remaining > 0; i++ {
		roomWidth := remaining / (roomCount - i)
		if roomWidth < 4 {
			roomWidth = 4
		}
		if roomWidth > remaining {
			roomWidth = remaining
		}

		roomType := RoomLiving
		if i == 0 {
			roomType = RoomEntrance
		} else if i == roomCount-1 {
			roomType = RoomBedroom
		}

		room := Room{
			X:      x,
			Y:      0,
			Width:  roomWidth,
			Height: building.Height,
			Type:   roomType,
		}
		building.Rooms = append(building.Rooms, room)

		// Add door to next room (except last room)
		if i < roomCount-1 {
			door := Door{
				X:    x + roomWidth,
				Y:    building.Height / 2,
				Type: DoorWooden,
			}
			building.Doors = append(building.Doors, door)
		}

		x += roomWidth
		remaining -= roomWidth
	}

	return nil
}

// generateWorkshopLayout creates a workshop floor plan
func (g *Generator) generateWorkshopLayout(building *Building, roomCount int, rng *rand.Rand) error {
	// Main workshop area + smaller storage/entrance
	entranceWidth := building.Width / 4
	if entranceWidth < 4 {
		entranceWidth = 4
	}

	// Entrance takes full height on left side
	entrance := Room{
		X:      0,
		Y:      0,
		Width:  entranceWidth,
		Height: building.Height,
		Type:   RoomEntrance,
	}
	building.Rooms = append(building.Rooms, entrance)

	// Main workshop
	workshop := Room{
		X:      entranceWidth,
		Y:      0,
		Width:  building.Width - entranceWidth,
		Height: building.Height,
		Type:   RoomWorkshop,
	}
	building.Rooms = append(building.Rooms, workshop)

	// Connect rooms - door at shared vertical wall
	door := Door{
		X:    entranceWidth,
		Y:    building.Height / 2,
		Type: DoorWooden,
	}
	building.Doors = append(building.Doors, door)

	return nil
}

// generateStorageLayout creates a storage building floor plan
func (g *Generator) generateStorageLayout(building *Building, roomCount int, rng *rand.Rand) error {
	if roomCount == 1 {
		// Single large storage room
		room := Room{
			X:      0,
			Y:      0,
			Width:  building.Width,
			Height: building.Height,
			Type:   RoomStorage,
		}
		building.Rooms = append(building.Rooms, room)
	} else {
		// Split into two storage areas
		splitX := building.Width / 2

		room1 := Room{
			X:      0,
			Y:      0,
			Width:  splitX,
			Height: building.Height,
			Type:   RoomEntrance, // Entrance counts as one storage area
		}
		building.Rooms = append(building.Rooms, room1)

		room2 := Room{
			X:      splitX,
			Y:      0,
			Width:  building.Width - splitX,
			Height: building.Height,
			Type:   RoomStorage,
		}
		building.Rooms = append(building.Rooms, room2)

		// Connect rooms
		door := Door{
			X:    splitX,
			Y:    building.Height / 2,
			Type: DoorMetal, // Sturdy door for storage
		}
		building.Doors = append(building.Doors, door)
	}

	return nil
}

// generateTowerLayout creates a tower floor plan with stacked floors
func (g *Generator) generateTowerLayout(building *Building, roomCount int, rng *rand.Rand) error {
	floorHeight := building.Height / roomCount
	if floorHeight < 5 {
		floorHeight = 5
		roomCount = building.Height / floorHeight
	}

	for i := 0; i < roomCount; i++ {
		roomType := RoomTower
		if i == 0 {
			roomType = RoomEntrance
		}

		room := Room{
			X:      0,
			Y:      i * floorHeight,
			Width:  building.Width,
			Height: floorHeight,
			Type:   roomType,
		}
		building.Rooms = append(building.Rooms, room)

		// Add door/stairs to next floor
		if i < roomCount-1 {
			door := Door{
				X:    building.Width / 2,
				Y:    (i + 1) * floorHeight,
				Type: DoorWooden,
			}
			building.Doors = append(building.Doors, door)
		}
	}

	return nil
}

// generateManorLayout creates a complex manor floor plan
func (g *Generator) generateManorLayout(building *Building, roomCount int, rng *rand.Rand) error {
	// Grid-based layout for manor
	cols := 2 + rng.Intn(2) // 2-3 columns
	rows := (roomCount + cols - 1) / cols

	roomWidth := building.Width / cols
	roomHeight := building.Height / rows

	roomIdx := 0
	for row := 0; row < rows && roomIdx < roomCount; row++ {
		for col := 0; col < cols && roomIdx < roomCount; col++ {
			roomType := RoomLiving
			if roomIdx == 0 {
				roomType = RoomEntrance
			} else {
				// Vary room types
				switch roomIdx % 3 {
				case 0:
					roomType = RoomBedroom
				case 1:
					roomType = RoomKitchen
				case 2:
					roomType = RoomLiving
				}
			}

			room := Room{
				X:      col * roomWidth,
				Y:      row * roomHeight,
				Width:  roomWidth,
				Height: roomHeight,
				Type:   roomType,
			}
			building.Rooms = append(building.Rooms, room)

			// Add horizontal door
			if col < cols-1 && roomIdx+1 < roomCount {
				door := Door{
					X:    (col + 1) * roomWidth,
					Y:    row*roomHeight + roomHeight/2,
					Type: DoorWooden,
				}
				building.Doors = append(building.Doors, door)
			}

			// Add vertical door
			if row < rows-1 && roomIdx+cols < roomCount {
				door := Door{
					X:    col*roomWidth + roomWidth/2,
					Y:    (row + 1) * roomHeight,
					Type: DoorWooden,
				}
				building.Doors = append(building.Doors, door)
			}

			roomIdx++
		}
	}

	return nil
}

// generateGuildHallLayout creates a multi-floor layout for guild halls
func (g *Generator) generateGuildHallLayout(building *Building, roomCount int, rng *rand.Rand) error {
	// Guild halls have multiple floors (1-5)
	floorsCount := building.Floors
	if floorsCount < 1 {
		floorsCount = 1
	}

	// Distribute rooms across floors
	roomsPerFloor := roomCount / floorsCount
	if roomsPerFloor < 2 {
		roomsPerFloor = 2
	}

	// Each floor uses a grid layout similar to manor
	cols := 3 + rng.Intn(2) // 3-4 columns for guild halls
	rows := (roomsPerFloor + cols - 1) / cols

	roomWidth := building.Width / cols
	roomHeight := building.Height / rows

	// Generate rooms for each floor
	for floor := 0; floor < floorsCount; floor++ {
		floorRooms := []Room{}
		roomIdx := 0

		for row := 0; row < rows && roomIdx < roomsPerFloor; row++ {
			for col := 0; col < cols && roomIdx < roomsPerFloor; col++ {
				// Determine room type based on floor
				roomType := RoomLiving
				if floor == 0 && roomIdx == 0 {
					// Ground floor entrance
					roomType = RoomEntrance
				} else if floor == floorsCount-1 {
					// Top floor: meeting rooms and special rooms
					roomType = RoomLiving
				} else {
					// Middle floors: varied rooms
					switch roomIdx % 4 {
					case 0:
						roomType = RoomStorage
					case 1:
						roomType = RoomWorkshop
					case 2:
						roomType = RoomLiving
					case 3:
						roomType = RoomBedroom
					}
				}

				room := Room{
					X:      col * roomWidth,
					Y:      row * roomHeight,
					Width:  roomWidth,
					Height: roomHeight,
					Type:   roomType,
				}

				// For multi-floor buildings, only add ground floor to main Rooms list
				// All floors go into FloorRooms for completeness
				if floor == 0 {
					building.Rooms = append(building.Rooms, room)
				}
				floorRooms = append(floorRooms, room)

				// Add horizontal door (only to ground floor for validation)
				if floor == 0 && col < cols-1 && roomIdx+1 < roomsPerFloor {
					door := Door{
						X:    (col + 1) * roomWidth,
						Y:    row*roomHeight + roomHeight/2,
						Type: DoorMetal, // Guild halls use metal doors
					}
					building.Doors = append(building.Doors, door)
				}

				// Add vertical door (only to ground floor for validation)
				if floor == 0 && row < rows-1 && roomIdx+cols < roomsPerFloor {
					door := Door{
						X:    col*roomWidth + roomWidth/2,
						Y:    (row + 1) * roomHeight,
						Type: DoorMetal,
					}
					building.Doors = append(building.Doors, door)
				}

				roomIdx++
			}
		}

		// Store rooms per floor
		building.FloorRooms[floor] = floorRooms
	}

	return nil
}

// generateRoof selects an appropriate roof type
func (g *Generator) generateRoof(buildingType BuildingType, style ArchitecturalStyle, rng *rand.Rand) RoofType {
	switch buildingType {
	case TypeTower:
		return RoofSpire
	case TypeManor:
		if style == StyleMedieval || style == StyleGothic {
			return RoofGabled
		}
		return RoofHipped
	case TypeGuildHall:
		// Guild halls use imposing roofs
		if style == StyleMedieval || style == StyleGothic || style == StyleElven {
			return RoofGabled
		}
		if style == StyleCrystalline || style == StyleGeometric {
			return RoofDomed
		}
		return RoofHipped
	case TypeStorage:
		return RoofFlat
	default:
		// Random for houses and workshops
		roofTypes := []RoofType{RoofGabled, RoofHipped, RoofFlat}
		return roofTypes[rng.Intn(len(roofTypes))]
	}
}

// placeWindows adds windows to exterior walls
func (g *Generator) placeWindows(building *Building, rng *rand.Rand) {
	windowCount := (building.Width + building.Height) / 4
	if windowCount > 20 {
		windowCount = 20
	}

	for i := 0; i < windowCount; i++ {
		// Random exterior wall position
		var x, y int
		if rng.Intn(2) == 0 {
			// Horizontal wall
			x = rng.Intn(building.Width)
			if rng.Intn(2) == 0 {
				y = 0 // Top wall
			} else {
				y = building.Height - 1 // Bottom wall
			}
		} else {
			// Vertical wall
			y = rng.Intn(building.Height)
			if rng.Intn(2) == 0 {
				x = 0 // Left wall
			} else {
				x = building.Width - 1 // Right wall
			}
		}

		// Determine window type based on genre and style
		windowType := WindowSmall
		if building.Style == StyleGothic || building.Style == StyleMansion {
			windowType = WindowStained
		} else if building.Style == StyleDecayed || building.Style == StyleRuins {
			if rng.Float64() < 0.3 {
				windowType = WindowBroken
			}
		} else if rng.Float64() < 0.3 {
			windowType = WindowLarge
		}

		window := Window{
			X:    x,
			Y:    y,
			Type: windowType,
		}
		building.Windows = append(building.Windows, window)
	}
}
