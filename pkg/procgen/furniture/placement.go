package furniture

import (
	"fmt"
	"math"
)

// This file contains furniture placement validation logic including:
// - PlacementValidator for room-based placement validation
// - PlacedFurniture for tracking positioned furniture
// - Collision detection with AABB (Axis-Aligned Bounding Box)
// - Automatic placement finding
// - Rotation support (4-way and 8-way)

// PlacementValidator validates furniture placement in rooms
type PlacementValidator struct {
	// Room dimensions in tiles
	RoomWidth  float64
	RoomHeight float64

	// Existing furniture placements
	Placements []*PlacedFurniture
}

// PlacedFurniture represents furniture with position and rotation
type PlacedFurniture struct {
	Furniture *Furniture
	X         float64 // Position in tiles
	Y         float64
	Direction Direction
}

// NewPlacementValidator creates a validator for a room
func NewPlacementValidator(roomWidth, roomHeight float64) *PlacementValidator {
	return &PlacementValidator{
		RoomWidth:  roomWidth,
		RoomHeight: roomHeight,
		Placements: make([]*PlacedFurniture, 0),
	}
}

// ValidatePlacement checks if furniture can be placed at the given position
func (pv *PlacementValidator) ValidatePlacement(furniture *Furniture, x, y float64, direction Direction) error {
	// Get rotated dimensions
	width, depth := pv.getRotatedDimensions(furniture, direction)

	// Check room boundaries
	if x < 0 || y < 0 {
		return fmt.Errorf("position out of bounds: (%.2f, %.2f)", x, y)
	}
	if x+width > pv.RoomWidth || y+depth > pv.RoomHeight {
		return fmt.Errorf("furniture extends beyond room: position (%.2f, %.2f), size (%.2f, %.2f), room (%.2f, %.2f)",
			x, y, width, depth, pv.RoomWidth, pv.RoomHeight)
	}

	// Check collision with existing furniture
	for _, existing := range pv.Placements {
		if pv.checkCollision(furniture, x, y, direction, existing) {
			return fmt.Errorf("collision with existing %s at (%.2f, %.2f)", existing.Furniture.SubType, existing.X, existing.Y)
		}
	}

	return nil
}

// PlaceFurniture validates and places furniture, returning error if invalid
func (pv *PlacementValidator) PlaceFurniture(furniture *Furniture, x, y float64, direction Direction) error {
	if err := pv.ValidatePlacement(furniture, x, y, direction); err != nil {
		return err
	}

	placed := &PlacedFurniture{
		Furniture: furniture,
		X:         x,
		Y:         y,
		Direction: direction,
	}

	pv.Placements = append(pv.Placements, placed)
	return nil
}

// RemoveFurniture removes furniture at the given index
func (pv *PlacementValidator) RemoveFurniture(index int) error {
	if index < 0 || index >= len(pv.Placements) {
		return fmt.Errorf("invalid index: %d", index)
	}

	pv.Placements = append(pv.Placements[:index], pv.Placements[index+1:]...)
	return nil
}

// Clear removes all furniture placements
func (pv *PlacementValidator) Clear() {
	pv.Placements = make([]*PlacedFurniture, 0)
}

// GetPlacementCount returns the number of placed furniture items
func (pv *PlacementValidator) GetPlacementCount() int {
	return len(pv.Placements)
}

// GetOccupancyPercent returns the percentage of room floor space occupied (0-100)
func (pv *PlacementValidator) GetOccupancyPercent() float64 {
	totalArea := pv.RoomWidth * pv.RoomHeight
	if totalArea == 0 {
		return 0
	}

	occupiedArea := 0.0
	for _, placed := range pv.Placements {
		width, depth := pv.getRotatedDimensions(placed.Furniture, placed.Direction)
		occupiedArea += width * depth
	}

	return (occupiedArea / totalArea) * 100.0
}

// getRotatedDimensions returns width and depth after rotation
func (pv *PlacementValidator) getRotatedDimensions(furniture *Furniture, direction Direction) (float64, float64) {
	width := furniture.CollisionWidth
	depth := furniture.CollisionDepth

	// Rotation affects orientation
	switch direction {
	case DirNorth, DirSouth:
		return width, depth
	case DirEast, DirWest:
		return depth, width // Swap width and depth
	case DirNorthEast, DirSouthEast, DirSouthWest, DirNorthWest:
		// Diagonal rotation: use diagonal distance
		diagonal := math.Sqrt(width*width + depth*depth)
		return diagonal, diagonal
	}

	return width, depth
}

// checkCollision tests if two furniture items collide
func (pv *PlacementValidator) checkCollision(furniture *Furniture, x, y float64, direction Direction, existing *PlacedFurniture) bool {
	// Get dimensions for new furniture
	newWidth, newDepth := pv.getRotatedDimensions(furniture, direction)

	// Get dimensions for existing furniture
	existWidth, existDepth := pv.getRotatedDimensions(existing.Furniture, existing.Direction)

	// AABB collision detection
	newLeft := x
	newRight := x + newWidth
	newTop := y
	newBottom := y + newDepth

	existLeft := existing.X
	existRight := existing.X + existWidth
	existTop := existing.Y
	existBottom := existing.Y + existDepth

	// Check overlap
	if newRight <= existLeft || newLeft >= existRight {
		return false // No horizontal overlap
	}
	if newBottom <= existTop || newTop >= existBottom {
		return false // No vertical overlap
	}

	return true // Collision detected
}

// FindValidPlacement attempts to find a valid position for furniture
// Returns x, y, direction and success flag
func (pv *PlacementValidator) FindValidPlacement(furniture *Furniture, preferredDirection Direction) (float64, float64, Direction, bool) {
	// Try grid positions with some spacing
	spacing := 0.5 // Half-tile spacing

	for y := 0.0; y < pv.RoomHeight; y += spacing {
		for x := 0.0; x < pv.RoomWidth; x += spacing {
			// Try preferred direction first
			if pv.ValidatePlacement(furniture, x, y, preferredDirection) == nil {
				return x, y, preferredDirection, true
			}

			// Try other 4-directional rotations
			directions := []Direction{DirNorth, DirEast, DirSouth, DirWest}
			for _, dir := range directions {
				if dir == preferredDirection {
					continue
				}
				if pv.ValidatePlacement(furniture, x, y, dir) == nil {
					return x, y, dir, true
				}
			}
		}
	}

	return 0, 0, DirNorth, false
}

// GetFurnitureAt returns furniture at the given position, or nil if none
func (pv *PlacementValidator) GetFurnitureAt(x, y float64) *PlacedFurniture {
	for _, placed := range pv.Placements {
		width, depth := pv.getRotatedDimensions(placed.Furniture, placed.Direction)

		if x >= placed.X && x < placed.X+width &&
			y >= placed.Y && y < placed.Y+depth {
			return placed
		}
	}

	return nil
}

// RotateFurniture rotates furniture to the next direction (4-way rotation)
func (pv *PlacementValidator) RotateFurniture(index int) error {
	if index < 0 || index >= len(pv.Placements) {
		return fmt.Errorf("invalid index: %d", index)
	}

	placed := pv.Placements[index]

	// Determine next direction (4-way rotation)
	var nextDir Direction
	switch placed.Direction {
	case DirNorth:
		nextDir = DirEast
	case DirEast:
		nextDir = DirSouth
	case DirSouth:
		nextDir = DirWest
	case DirWest:
		nextDir = DirNorth
	default:
		// 8-way rotation
		nextDir = Direction((int(placed.Direction) + 1) % 8)
	}

	// Temporarily remove furniture
	tempFurniture := placed.Furniture
	tempX := placed.X
	tempY := placed.Y
	pv.Placements = append(pv.Placements[:index], pv.Placements[index+1:]...)

	// Try to place with new rotation
	if err := pv.PlaceFurniture(tempFurniture, tempX, tempY, nextDir); err != nil {
		// Rotation failed, restore original
		pv.Placements = append(pv.Placements[:index], append([]*PlacedFurniture{placed}, pv.Placements[index:]...)...)
		return fmt.Errorf("cannot rotate: %w", err)
	}

	return nil
}
