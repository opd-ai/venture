// Package engine provides post office building spawning for cities.
// This file implements procedural post office placement with clerk NPCs.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// PostOfficeSpawner spawns post office buildings in city terrains
type PostOfficeSpawner struct {
	world         *World
	courierSystem *CourierSystem
}

// NewPostOfficeSpawner creates a new post office spawner
func NewPostOfficeSpawner(world *World, courierSystem *CourierSystem) *PostOfficeSpawner {
	return &PostOfficeSpawner{
		world:         world,
		courierSystem: courierSystem,
	}
}

// PostOfficeResult contains information about a spawned post office
type PostOfficeResult struct {
	BuildingID uint64
	ClerkID    uint64
	ClerkName  string
	X, Y       float64
}

// SpawnInCity spawns a post office building in a city terrain
func (s *PostOfficeSpawner) SpawnInCity(cityTerrain *terrain.Terrain, blocks []*terrain.CityBlock, genreID string, seed int64) (*PostOfficeResult, error) {
	if cityTerrain == nil {
		return nil, fmt.Errorf("terrain is nil")
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks provided")
	}

	rng := rand.New(rand.NewSource(seed))

	// Find suitable building blocks for post offices
	candidates := s.findSuitableBlocks(blocks, cityTerrain)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable blocks found for post office")
	}

	// Choose block closest to city center
	block := s.selectCentralBlock(candidates, cityTerrain)

	// Generate clerk name
	clerkName := s.generateClerkName(rng, genreID)

	// Spawn post office
	cx, cy := block.Rect.Center()
	buildingID, clerkID := s.courierSystem.SpawnPostOffice(float64(cx), float64(cy), clerkName)

	return &PostOfficeResult{
		BuildingID: buildingID,
		ClerkID:    clerkID,
		ClerkName:  clerkName,
		X:          float64(cx),
		Y:          float64(cy),
	}, nil
}

// findSuitableBlocks finds building blocks suitable for post offices
func (s *PostOfficeSpawner) findSuitableBlocks(blocks []*terrain.CityBlock, cityTerrain *terrain.Terrain) []*terrain.CityBlock {
	suitable := make([]*terrain.CityBlock, 0)

	for _, block := range blocks {
		// Post offices should be in large buildings (area >= 64)
		if block.BlockType != terrain.BlockBuilding {
			continue
		}

		area := block.Rect.Width * block.Rect.Height
		if area < 64 {
			continue
		}

		// Ensure block has street access
		if s.hasStreetAccess(block, cityTerrain) {
			suitable = append(suitable, block)
		}
	}

	return suitable
}

// hasStreetAccess checks if a block has adjacent street tiles
func (s *PostOfficeSpawner) hasStreetAccess(block *terrain.CityBlock, cityTerrain *terrain.Terrain) bool {
	rect := block.Rect
	return s.checkTopSide(rect, cityTerrain) ||
		s.checkBottomSide(rect, cityTerrain) ||
		s.checkLeftSide(rect, cityTerrain) ||
		s.checkRightSide(rect, cityTerrain)
}

// checkTopSide checks for street access on the top side of a block
func (s *PostOfficeSpawner) checkTopSide(rect terrain.Rect, cityTerrain *terrain.Terrain) bool {
	y := rect.Y - 1
	if y < 0 {
		return false
	}
	for x := rect.X; x < rect.X+rect.Width; x++ {
		if cityTerrain.GetTile(x, y) == terrain.TileCorridor {
			return true
		}
	}
	return false
}

// checkBottomSide checks for street access on the bottom side of a block
func (s *PostOfficeSpawner) checkBottomSide(rect terrain.Rect, cityTerrain *terrain.Terrain) bool {
	y := rect.Y + rect.Height
	if y >= cityTerrain.Height {
		return false
	}
	for x := rect.X; x < rect.X+rect.Width; x++ {
		if cityTerrain.GetTile(x, y) == terrain.TileCorridor {
			return true
		}
	}
	return false
}

// checkLeftSide checks for street access on the left side of a block
func (s *PostOfficeSpawner) checkLeftSide(rect terrain.Rect, cityTerrain *terrain.Terrain) bool {
	x := rect.X - 1
	if x < 0 {
		return false
	}
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		if cityTerrain.GetTile(x, y) == terrain.TileCorridor {
			return true
		}
	}
	return false
}

// checkRightSide checks for street access on the right side of a block
func (s *PostOfficeSpawner) checkRightSide(rect terrain.Rect, cityTerrain *terrain.Terrain) bool {
	x := rect.X + rect.Width
	if x >= cityTerrain.Width {
		return false
	}
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		if cityTerrain.GetTile(x, y) == terrain.TileCorridor {
			return true
		}
	}
	return false
}

// selectCentralBlock chooses the block closest to the city center
func (s *PostOfficeSpawner) selectCentralBlock(blocks []*terrain.CityBlock, cityTerrain *terrain.Terrain) *terrain.CityBlock {
	centerX := cityTerrain.Width / 2
	centerY := cityTerrain.Height / 2

	var best *terrain.CityBlock
	minDist := int(1e9)

	for _, block := range blocks {
		cx, cy := block.Rect.Center()
		dx := cx - centerX
		dy := cy - centerY
		dist := dx*dx + dy*dy

		if dist < minDist {
			minDist = dist
			best = block
		}
	}

	return best
}

// SpawnInTerrain spawns a post office in the largest suitable room.
// This is a generic fallback for terrains without city blocks. It picks
// the largest room that meets a minimum area threshold and places the
// post office at the room center.
func (s *PostOfficeSpawner) SpawnInTerrain(t *terrain.Terrain, genreID string, seed int64) (*PostOfficeResult, error) {
	if t == nil {
		return nil, fmt.Errorf("terrain is nil")
	}

	if len(t.Rooms) == 0 {
		return nil, fmt.Errorf("terrain has no rooms")
	}

	const minRoomArea = 36 // 6x6 minimum for a post office

	// Find the largest room meeting the threshold
	var best *terrain.Room
	bestArea := 0
	for _, room := range t.Rooms {
		area := room.Width * room.Height
		if area >= minRoomArea && area > bestArea {
			best = room
			bestArea = area
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no room large enough for post office (need %d area)", minRoomArea)
	}

	rng := rand.New(rand.NewSource(seed ^ 0x504F5354)) // "POST"
	clerkName := s.generateClerkName(rng, genreID)

	cx, cy := best.Center()
	buildingID, clerkID := s.courierSystem.SpawnPostOffice(float64(cx), float64(cy), clerkName)

	return &PostOfficeResult{
		BuildingID: buildingID,
		ClerkID:    clerkID,
		ClerkName:  clerkName,
		X:          float64(cx),
		Y:          float64(cy),
	}, nil
}

// generateClerkName generates a procedural name for a post office clerk
func (s *PostOfficeSpawner) generateClerkName(rng *rand.Rand, genreID string) string {
	var firstNames []string
	var lastNames []string

	switch genreID {
	case "fantasy":
		firstNames = []string{"Aldric", "Bram", "Cedric", "Dorian", "Eldric", "Fenwick", "Gareth", "Hadrian", "Ivor", "Jasper"}
		lastNames = []string{"Postmaster", "Mailkeeper", "Courier", "Lettersmith", "Parcelman", "Swift", "Rider", "Messenger", "Scribe", "Keeper"}
	case "scifi":
		firstNames = []string{"Zeta", "Nexus", "Orion", "Pulsar", "Quantum", "Relay", "Sigma", "Theta", "Vector", "Vortex"}
		lastNames = []string{"Station", "Terminal", "Hub", "Node", "Port", "Relay", "Center", "Exchange", "Depot", "Gateway"}
	case "horror":
		firstNames = []string{"Silas", "Mortimer", "Ebenezer", "Cornelius", "Bartholomew", "Lazarus", "Ignatius", "Malachi", "Thaddeus", "Zachariah"}
		lastNames = []string{"Graves", "Crow", "Raven", "Ash", "Shadow", "Grimm", "Dusk", "Hollow", "Mourning", "Nightshade"}
	case "cyberpunk":
		firstNames = []string{"Neo", "Cipher", "Ghost", "Razor", "Spike", "Blade", "Chrome", "Jax", "Zero", "Echo"}
		lastNames = []string{"Datalink", "Netrunner", "Courier", "Runner", "Packet", "Transfer", "Stream", "Protocol", "Handler", "Router"}
	case "postapoc":
		firstNames = []string{"Rust", "Ash", "Scrap", "Dust", "Grit", "Flint", "Steel", "Iron", "Gravel", "Stone"}
		lastNames = []string{"Mailrun", "Roadrunner", "Wastelander", "Survivor", "Runner", "Wanderer", "Drifter", "Tracker", "Scout", "Rider"}
	default:
		firstNames = []string{"Alan", "Ben", "Carl", "Dan", "Eric", "Frank", "George", "Henry", "Ian", "Jack"}
		lastNames = []string{"Postmaster", "Clerk", "Courier", "Mail", "Letter", "Parcel", "Swift", "Quick", "Fast", "Express"}
	}

	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]

	// 70% chance to use both names, 30% single name
	if rng.Float64() < 0.7 {
		return fmt.Sprintf("%s %s", firstName, lastName)
	}
	return firstName
}
