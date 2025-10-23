//go:build test
// +build test

package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestBSPGenerator_RoomTypes(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Verify terrain has rooms
	if len(terrain.Rooms) == 0 {
		t.Fatal("Expected rooms to be generated")
	}

	// First room should be spawn
	if terrain.Rooms[0].Type != RoomSpawn {
		t.Errorf("Expected first room to be spawn, got %v", terrain.Rooms[0].Type)
	}

	// Last room should be exit
	if len(terrain.Rooms) > 1 {
		lastIdx := len(terrain.Rooms) - 1
		if terrain.Rooms[lastIdx].Type != RoomExit {
			t.Errorf("Expected last room to be exit, got %v", terrain.Rooms[lastIdx].Type)
		}
	}

	// Should have at least one boss room if 3+ rooms
	if len(terrain.Rooms) >= 3 {
		hasBoss := false
		for _, room := range terrain.Rooms {
			if room.Type == RoomBoss {
				hasBoss = true
				break
			}
		}
		if !hasBoss {
			t.Error("Expected at least one boss room in dungeon with 3+ rooms")
		}
	}

	// Count room types
	typeCounts := make(map[RoomType]int)
	for _, room := range terrain.Rooms {
		typeCounts[room.Type]++
	}

	t.Logf("Generated %d rooms:", len(terrain.Rooms))
	for roomType, count := range typeCounts {
		t.Logf("  %s: %d", roomType, count)
	}

	// Verify we have variety
	if len(terrain.Rooms) >= 5 {
		if len(typeCounts) < 3 {
			t.Errorf("Expected at least 3 different room types in dungeon with %d rooms, got %d", len(terrain.Rooms), len(typeCounts))
		}
	}
}

func TestBSPGenerator_RoomTypesDeterministic(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  50,
			"height": 30,
		},
	}

	seed := int64(54321)

	// Generate twice with same seed
	result1, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	result2, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Verify same number of rooms
	if len(terrain1.Rooms) != len(terrain2.Rooms) {
		t.Errorf("Room counts differ: %d vs %d", len(terrain1.Rooms), len(terrain2.Rooms))
	}

	// Verify room types match
	for i := 0; i < len(terrain1.Rooms) && i < len(terrain2.Rooms); i++ {
		if terrain1.Rooms[i].Type != terrain2.Rooms[i].Type {
			t.Errorf("Room %d type differs: %v vs %v", i, terrain1.Rooms[i].Type, terrain2.Rooms[i].Type)
		}
	}
}

func TestRoomType_String(t *testing.T) {
	tests := []struct {
		roomType RoomType
		expected string
	}{
		{RoomNormal, "normal"},
		{RoomTreasure, "treasure"},
		{RoomBoss, "boss"},
		{RoomTrap, "trap"},
		{RoomSpawn, "spawn"},
		{RoomExit, "exit"},
		{RoomType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.roomType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBSPGenerator_SmallDungeonRoomTypes(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  30,
			"height": 20,
		},
	}

	result, err := gen.Generate(99999, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Even small dungeons should have spawn/exit
	if len(terrain.Rooms) == 0 {
		t.Fatal("Expected at least one room")
	}

	// First room is spawn
	if terrain.Rooms[0].Type != RoomSpawn {
		t.Errorf("Expected spawn room, got %v", terrain.Rooms[0].Type)
	}

	// If only one room, it's both spawn and exit
	if len(terrain.Rooms) == 1 {
		t.Log("Single room dungeon - spawn only")
	} else {
		// Multiple rooms should have exit
		hasExit := false
		for _, room := range terrain.Rooms {
			if room.Type == RoomExit {
				hasExit = true
				break
			}
		}
		if !hasExit {
			t.Error("Expected exit room in multi-room dungeon")
		}
	}
}
