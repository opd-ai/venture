package building

import (
	"fmt"
	"math/rand"
)

// BuildingType represents different building types
type BuildingType int

const (
	TypeHouse BuildingType = iota
	TypeWorkshop
	TypeStorage
	TypeTower
	TypeManor
)

// String returns the string representation of BuildingType
func (t BuildingType) String() string {
	switch t {
	case TypeHouse:
		return "House"
	case TypeWorkshop:
		return "Workshop"
	case TypeStorage:
		return "Storage"
	case TypeTower:
		return "Tower"
	case TypeManor:
		return "Manor"
	default:
		return "Unknown"
	}
}

// ArchitecturalStyle represents different building styles per genre
type ArchitecturalStyle int

const (
	// Fantasy styles
	StyleMedieval ArchitecturalStyle = iota
	StyleElven
	StyleDwarven
	StyleWizardTower
	StyleVillage

	// Sci-Fi styles
	StyleModular
	StyleBrutalist
	StyleOrganic
	StyleGeometric
	StyleCrystalline

	// Horror styles
	StyleGothic
	StyleDecayed
	StyleAsylum
	StyleMansion
	StyleCrypt

	// Cyberpunk styles
	StyleNeon
	StyleIndustrial
	StyleCorporate
	StyleUnderground
	StyleMegastructure

	// Post-Apocalyptic styles
	StyleSalvage
	StyleBunker
	StyleRuins
	StyleFortified
	StyleScrapyard
)

// String returns the string representation of ArchitecturalStyle
func (s ArchitecturalStyle) String() string {
	styles := map[ArchitecturalStyle]string{
		StyleMedieval:      "Medieval",
		StyleElven:         "Elven",
		StyleDwarven:       "Dwarven",
		StyleWizardTower:   "WizardTower",
		StyleVillage:       "Village",
		StyleModular:       "Modular",
		StyleBrutalist:     "Brutalist",
		StyleOrganic:       "Organic",
		StyleGeometric:     "Geometric",
		StyleCrystalline:   "Crystalline",
		StyleGothic:        "Gothic",
		StyleDecayed:       "Decayed",
		StyleAsylum:        "Asylum",
		StyleMansion:       "Mansion",
		StyleCrypt:         "Crypt",
		StyleNeon:          "Neon",
		StyleIndustrial:    "Industrial",
		StyleCorporate:     "Corporate",
		StyleUnderground:   "Underground",
		StyleMegastructure: "Megastructure",
		StyleSalvage:       "Salvage",
		StyleBunker:        "Bunker",
		StyleRuins:         "Ruins",
		StyleFortified:     "Fortified",
		StyleScrapyard:     "Scrapyard",
	}
	if name, ok := styles[s]; ok {
		return name
	}
	return "Unknown"
}

// GetGenreStyles returns the 5 architectural styles for a genre
func GetGenreStyles(genreID string) []ArchitecturalStyle {
	styles := map[string][]ArchitecturalStyle{
		"fantasy":   {StyleMedieval, StyleElven, StyleDwarven, StyleWizardTower, StyleVillage},
		"scifi":     {StyleModular, StyleBrutalist, StyleOrganic, StyleGeometric, StyleCrystalline},
		"horror":    {StyleGothic, StyleDecayed, StyleAsylum, StyleMansion, StyleCrypt},
		"cyberpunk": {StyleNeon, StyleIndustrial, StyleCorporate, StyleUnderground, StyleMegastructure},
		"postapoc":  {StyleSalvage, StyleBunker, StyleRuins, StyleFortified, StyleScrapyard},
	}
	if s, ok := styles[genreID]; ok {
		return s
	}
	return styles["fantasy"] // Default fallback
}

// Room represents a room in a building floor plan
type Room struct {
	X      int
	Y      int
	Width  int
	Height int
	Type   RoomType
}

// RoomType represents different room types
type RoomType int

const (
	RoomEntrance RoomType = iota
	RoomLiving
	RoomBedroom
	RoomStorage
	RoomWorkshop
	RoomKitchen
	RoomHallway
	RoomTower
)

// String returns the string representation of RoomType
func (r RoomType) String() string {
	switch r {
	case RoomEntrance:
		return "Entrance"
	case RoomLiving:
		return "Living"
	case RoomBedroom:
		return "Bedroom"
	case RoomStorage:
		return "Storage"
	case RoomWorkshop:
		return "Workshop"
	case RoomKitchen:
		return "Kitchen"
	case RoomHallway:
		return "Hallway"
	case RoomTower:
		return "Tower"
	default:
		return "Unknown"
	}
}

// Door represents a connection between rooms
type Door struct {
	X    int
	Y    int
	Type DoorType
}

// DoorType represents different door types
type DoorType int

const (
	DoorWooden DoorType = iota
	DoorMetal
	DoorGlass
	DoorSecret
)

// String returns the string representation of DoorType
func (d DoorType) String() string {
	switch d {
	case DoorWooden:
		return "Wooden"
	case DoorMetal:
		return "Metal"
	case DoorGlass:
		return "Glass"
	case DoorSecret:
		return "Secret"
	default:
		return "Unknown"
	}
}

// Window represents a window in a building
type Window struct {
	X    int
	Y    int
	Type WindowType
}

// WindowType represents different window types
type WindowType int

const (
	WindowSmall WindowType = iota
	WindowLarge
	WindowStained
	WindowBroken
)

// String returns the string representation of WindowType
func (w WindowType) String() string {
	switch w {
	case WindowSmall:
		return "Small"
	case WindowLarge:
		return "Large"
	case WindowStained:
		return "Stained"
	case WindowBroken:
		return "Broken"
	default:
		return "Unknown"
	}
}

// RoofType represents different roof types
type RoofType int

const (
	RoofFlat RoofType = iota
	RoofGabled
	RoofHipped
	RoofDomed
	RoofSpire
)

// String returns the string representation of RoofType
func (r RoofType) String() string {
	switch r {
	case RoofFlat:
		return "Flat"
	case RoofGabled:
		return "Gabled"
	case RoofHipped:
		return "Hipped"
	case RoofDomed:
		return "Domed"
	case RoofSpire:
		return "Spire"
	default:
		return "Unknown"
	}
}

// Building represents a complete building with floor plan
type Building struct {
	Type     BuildingType
	Style    ArchitecturalStyle
	GenreID  string
	Width    int
	Height   int
	Rooms    []Room
	Doors    []Door
	Windows  []Window
	RoofType RoofType
}

// GetRoomCount returns the number of rooms
func (b *Building) GetRoomCount() int {
	return len(b.Rooms)
}

// IsNavigable checks if all rooms are accessible from the entrance
func (b *Building) IsNavigable() bool {
	if len(b.Rooms) == 0 {
		return false
	}

	// Find entrance room
	entranceIdx := -1
	for i, room := range b.Rooms {
		if room.Type == RoomEntrance {
			entranceIdx = i
			break
		}
	}
	if entranceIdx == -1 {
		return false
	}

	// BFS to check connectivity
	visited := make(map[int]bool)
	queue := []int{entranceIdx}
	visited[entranceIdx] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Find connected rooms via doors
		for i, room := range b.Rooms {
			if visited[i] {
				continue
			}
			if b.areRoomsConnected(b.Rooms[current], room) {
				visited[i] = true
				queue = append(queue, i)
			}
		}
	}

	// All rooms should be reachable
	return len(visited) == len(b.Rooms)
}

// areRoomsConnected checks if two rooms share a door
func (b *Building) areRoomsConnected(r1, r2 Room) bool {
	// Rooms are connected if they share a wall and have a door on that wall
	// Check if rooms share a wall
	shareWall := false
	var doorX, doorY int

	// Check horizontal adjacency
	if r1.Y == r2.Y && r1.Height == r2.Height {
		if r1.X+r1.Width == r2.X {
			shareWall = true
			doorX = r1.X + r1.Width
			doorY = r1.Y + r1.Height/2
		} else if r2.X+r2.Width == r1.X {
			shareWall = true
			doorX = r2.X + r2.Width
			doorY = r2.Y + r2.Height/2
		}
	}

	// Check vertical adjacency
	if r1.X == r2.X && r1.Width == r2.Width {
		if r1.Y+r1.Height == r2.Y {
			shareWall = true
			doorX = r1.X + r1.Width/2
			doorY = r1.Y + r1.Height
		} else if r2.Y+r2.Height == r1.Y {
			shareWall = true
			doorX = r2.X + r2.Width/2
			doorY = r2.Y + r2.Height
		}
	}

	if !shareWall {
		return false
	}

	// Check if there's a door at the shared wall
	for _, door := range b.Doors {
		if door.X == doorX && door.Y == doorY {
			return true
		}
	}

	return false
}

// GetStyleForGenreAndType returns an appropriate style for a building type and genre
func GetStyleForGenreAndType(genreID string, buildingType BuildingType, rng *rand.Rand) ArchitecturalStyle {
	styles := GetGenreStyles(genreID)
	if len(styles) == 0 {
		return StyleMedieval // Fallback
	}

	// Some building types prefer certain styles
	switch buildingType {
	case TypeTower:
		// Towers prefer vertical/imposing styles
		if genreID == "fantasy" {
			return StyleWizardTower
		}
		if genreID == "scifi" {
			return StyleBrutalist
		}
	case TypeManor:
		// Manors prefer elaborate styles
		if genreID == "fantasy" {
			return StyleMedieval
		}
		if genreID == "horror" {
			return StyleMansion
		}
	}

	// Default: random from genre styles
	return styles[rng.Intn(len(styles))]
}

// Validate checks if the building meets quality standards
func (b *Building) Validate() error {
	// Check basic dimensions
	if b.Width < 4 || b.Height < 4 {
		return fmt.Errorf("building too small: %dx%d (minimum 4x4)", b.Width, b.Height)
	}
	if b.Width > 64 || b.Height > 64 {
		return fmt.Errorf("building too large: %dx%d (maximum 64x64)", b.Width, b.Height)
	}

	// Check room count
	if len(b.Rooms) < 1 {
		return fmt.Errorf("building has no rooms")
	}
	if len(b.Rooms) > 8 {
		return fmt.Errorf("building has too many rooms: %d (maximum 8)", len(b.Rooms))
	}

	// Check for entrance room
	hasEntrance := false
	for _, room := range b.Rooms {
		if room.Type == RoomEntrance {
			hasEntrance = true
			break
		}
	}
	if !hasEntrance {
		return fmt.Errorf("building has no entrance room")
	}

	// Check navigability (all rooms accessible from entrance)
	if !b.IsNavigable() {
		return fmt.Errorf("building floor plan is not navigable")
	}

	// Check for room overlap
	for i, r1 := range b.Rooms {
		for j, r2 := range b.Rooms {
			if i >= j {
				continue
			}
			if roomsOverlap(r1, r2) {
				return fmt.Errorf("rooms %d and %d overlap", i, j)
			}
		}
	}

	return nil
}

// roomsOverlap checks if two rooms overlap
func roomsOverlap(r1, r2 Room) bool {
	return !(r1.X+r1.Width <= r2.X ||
		r2.X+r2.Width <= r1.X ||
		r1.Y+r1.Height <= r2.Y ||
		r2.Y+r2.Height <= r1.Y)
}
