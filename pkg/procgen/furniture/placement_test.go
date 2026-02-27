package furniture

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestPlacementValidatorBasic(t *testing.T) {
	validator := NewPlacementValidator(10.0, 8.0)

	if validator.RoomWidth != 10.0 || validator.RoomHeight != 8.0 {
		t.Errorf("NewPlacementValidator() incorrect dimensions: %.1f×%.1f, want 10×8", validator.RoomWidth, validator.RoomHeight)
	}

	if validator.GetPlacementCount() != 0 {
		t.Errorf("New validator has %d placements, want 0", validator.GetPlacementCount())
	}
}

func TestValidatePlacementInBounds(t *testing.T) {
	validator := NewPlacementValidator(10.0, 8.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	result, _ := gen.Generate(50000, params)
	furniture := result.(*Furniture)

	// Valid placement
	err := validator.ValidatePlacement(furniture, 2.0, 2.0, DirNorth)
	if err != nil {
		t.Errorf("ValidatePlacement() valid position error = %v", err)
	}

	// Negative position
	err = validator.ValidatePlacement(furniture, -1.0, 2.0, DirNorth)
	if err == nil {
		t.Error("ValidatePlacement() should fail for negative X")
	}

	// Out of bounds
	err = validator.ValidatePlacement(furniture, 9.5, 2.0, DirNorth)
	if err == nil {
		t.Error("ValidatePlacement() should fail for position too far right")
	}

	err = validator.ValidatePlacement(furniture, 2.0, 7.5, DirNorth)
	if err == nil {
		t.Error("ValidatePlacement() should fail for position too far down")
	}
}

func TestPlaceFurniture(t *testing.T) {
	validator := NewPlacementValidator(10.0, 8.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Table"},
	}

	result, _ := gen.Generate(51000, params)
	furniture := result.(*Furniture)

	// Place furniture
	err := validator.PlaceFurniture(furniture, 2.0, 2.0, DirNorth)
	if err != nil {
		t.Fatalf("PlaceFurniture() error = %v", err)
	}

	if validator.GetPlacementCount() != 1 {
		t.Errorf("PlacementCount = %d, want 1", validator.GetPlacementCount())
	}

	// Try to place overlapping furniture
	result2, _ := gen.Generate(51001, params)
	furniture2 := result2.(*Furniture)

	err = validator.PlaceFurniture(furniture2, 2.0, 2.0, DirNorth)
	if err == nil {
		t.Error("PlaceFurniture() should fail for overlapping placement")
	}

	// Should still have only 1 placement
	if validator.GetPlacementCount() != 1 {
		t.Errorf("PlacementCount = %d after failed placement, want 1", validator.GetPlacementCount())
	}
}

func TestCollisionDetection(t *testing.T) {
	validator := NewPlacementValidator(20.0, 16.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Place first furniture
	params.Custom = map[string]interface{}{"SubType": "Table"}
	result1, _ := gen.Generate(52000, params)
	furniture1 := result1.(*Furniture)
	validator.PlaceFurniture(furniture1, 5.0, 5.0, DirNorth)

	// Try to place too close
	params.Custom = map[string]interface{}{"SubType": "Chair"}
	result2, _ := gen.Generate(52001, params)
	furniture2 := result2.(*Furniture)

	err := validator.PlaceFurniture(furniture2, 5.0, 5.0, DirNorth)
	if err == nil {
		t.Error("PlaceFurniture() should detect collision at same position")
	}

	// Place far enough away
	err = validator.PlaceFurniture(furniture2, 10.0, 10.0, DirNorth)
	if err != nil {
		t.Errorf("PlaceFurniture() failed for non-overlapping position: %v", err)
	}

	if validator.GetPlacementCount() != 2 {
		t.Errorf("PlacementCount = %d, want 2", validator.GetPlacementCount())
	}
}

func TestRemoveFurniture(t *testing.T) {
	validator := NewPlacementValidator(10.0, 8.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	// Place 3 items
	for i := 0; i < 3; i++ {
		result, _ := gen.Generate(int64(53000+i), params)
		furniture := result.(*Furniture)
		validator.PlaceFurniture(furniture, float64(i)*2.0, 2.0, DirNorth)
	}

	if validator.GetPlacementCount() != 3 {
		t.Fatalf("PlacementCount = %d, want 3", validator.GetPlacementCount())
	}

	// Remove middle item
	err := validator.RemoveFurniture(1)
	if err != nil {
		t.Fatalf("RemoveFurniture(1) error = %v", err)
	}

	if validator.GetPlacementCount() != 2 {
		t.Errorf("PlacementCount = %d after removal, want 2", validator.GetPlacementCount())
	}

	// Remove invalid index
	err = validator.RemoveFurniture(10)
	if err == nil {
		t.Error("RemoveFurniture() should fail for invalid index")
	}
}

func TestClear(t *testing.T) {
	validator := NewPlacementValidator(10.0, 8.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	// Place items
	for i := 0; i < 5; i++ {
		result, _ := gen.Generate(int64(54000+i), params)
		furniture := result.(*Furniture)
		validator.PlaceFurniture(furniture, float64(i)*1.5, 2.0, DirNorth)
	}

	if validator.GetPlacementCount() != 5 {
		t.Fatalf("PlacementCount = %d, want 5", validator.GetPlacementCount())
	}

	validator.Clear()

	if validator.GetPlacementCount() != 0 {
		t.Errorf("PlacementCount = %d after Clear(), want 0", validator.GetPlacementCount())
	}
}

func TestGetOccupancyPercent(t *testing.T) {
	validator := NewPlacementValidator(10.0, 10.0) // 100 tiles

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Table"},
	}

	result, _ := gen.Generate(55000, params)
	furniture := result.(*Furniture)

	// Initial occupancy should be 0
	occupancy := validator.GetOccupancyPercent()
	if occupancy != 0.0 {
		t.Errorf("Initial occupancy = %.1f%%, want 0%%", occupancy)
	}

	// Place furniture
	validator.PlaceFurniture(furniture, 2.0, 2.0, DirNorth)

	occupancy = validator.GetOccupancyPercent()
	if occupancy <= 0.0 || occupancy > 100.0 {
		t.Errorf("Occupancy = %.1f%%, want >0 and <=100", occupancy)
	}
}

func TestRotatedDimensions(t *testing.T) {
	validator := NewPlacementValidator(10.0, 10.0)

	furniture := &Furniture{
		CollisionWidth: 3.0,
		CollisionDepth: 2.0,
	}

	// North/South: width × depth
	w, d := validator.getRotatedDimensions(furniture, DirNorth)
	if w != 3.0 || d != 2.0 {
		t.Errorf("North rotation = %.1f×%.1f, want 3×2", w, d)
	}

	// East/West: depth × width (swapped)
	w, d = validator.getRotatedDimensions(furniture, DirEast)
	if w != 2.0 || d != 3.0 {
		t.Errorf("East rotation = %.1f×%.1f, want 2×3", w, d)
	}

	// Diagonal: should be larger (diagonal distance)
	w, d = validator.getRotatedDimensions(furniture, DirNorthEast)
	diagonal := 3.605 // sqrt(3^2 + 2^2) ≈ 3.605
	if w < diagonal-0.1 || w > diagonal+0.1 {
		t.Errorf("NorthEast rotation width = %.3f, want ~%.3f", w, diagonal)
	}
}

func TestFindValidPlacement(t *testing.T) {
	validator := NewPlacementValidator(10.0, 10.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	result, _ := gen.Generate(56000, params)
	furniture := result.(*Furniture)

	x, y, dir, ok := validator.FindValidPlacement(furniture, DirNorth)
	if !ok {
		t.Fatal("FindValidPlacement() failed to find valid position in empty room")
	}

	// Position should be within room
	if x < 0 || y < 0 {
		t.Errorf("FindValidPlacement() returned negative position: (%.1f, %.1f)", x, y)
	}

	// Should be able to actually place there
	err := validator.PlaceFurniture(furniture, x, y, dir)
	if err != nil {
		t.Errorf("Cannot place at FindValidPlacement() result: %v", err)
	}
}

func TestGetFurnitureAt(t *testing.T) {
	validator := NewPlacementValidator(10.0, 10.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Table"},
	}

	result, _ := gen.Generate(57000, params)
	furniture := result.(*Furniture)

	validator.PlaceFurniture(furniture, 3.0, 4.0, DirNorth)

	// Check inside furniture bounds
	placed := validator.GetFurnitureAt(3.5, 4.5)
	if placed == nil {
		t.Error("GetFurnitureAt() returned nil for position inside furniture")
	}
	if placed != nil && placed.Furniture.ID != furniture.ID {
		t.Error("GetFurnitureAt() returned wrong furniture")
	}

	// Check outside furniture bounds
	placed = validator.GetFurnitureAt(8.0, 8.0)
	if placed != nil {
		t.Errorf("GetFurnitureAt() returned furniture for empty position: %v", placed)
	}
}

func TestRotateFurniture(t *testing.T) {
	validator := NewPlacementValidator(20.0, 20.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Table"},
	}

	result, _ := gen.Generate(58000, params)
	furniture := result.(*Furniture)

	validator.PlaceFurniture(furniture, 5.0, 5.0, DirNorth)

	original := validator.Placements[0].Direction

	// Rotate
	err := validator.RotateFurniture(0)
	if err != nil {
		t.Errorf("RotateFurniture(0) error = %v", err)
	}

	rotated := validator.Placements[0].Direction
	if rotated == original {
		t.Error("RotateFurniture() did not change direction")
	}

	// Rotate invalid index
	err = validator.RotateFurniture(10)
	if err == nil {
		t.Error("RotateFurniture() should fail for invalid index")
	}
}

func BenchmarkValidatePlacement(b *testing.B) {
	validator := NewPlacementValidator(20.0, 16.0)

	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	result, _ := gen.Generate(59000, params)
	furniture := result.(*Furniture)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidatePlacement(furniture, 5.0, 5.0, DirNorth)
	}
}

func BenchmarkFindValidPlacement(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"SubType": "Chair"},
	}

	result, _ := gen.Generate(60000, params)
	furniture := result.(*Furniture)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator := NewPlacementValidator(10.0, 10.0)
		validator.FindValidPlacement(furniture, DirNorth)
	}
}
