package territory

import "testing"

func TestTerritoryStatus_String(t *testing.T) {
	tests := []struct {
		status   TerritoryStatus
		expected string
	}{
		{StatusNeutral, "Neutral"},
		{StatusOwned, "Owned"},
		{StatusContested, "Contested"},
		{TerritoryStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStructureType_String(t *testing.T) {
	tests := []struct {
		structureType StructureType
		expected      string
	}{
		{StructureTypeWall, "Wall"},
		{StructureTypeTower, "Tower"},
		{StructureTypeGuard, "Guard"},
		{StructureType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.structureType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"TerritoryChunkSize", TerritoryChunkSize, 5},
		{"BaseCaptureTime", BaseCaptureTime, 60},
		{"DefenderTimeBonus", DefenderTimeBonus, 30},
		{"BaseResourceBonus", BaseResourceBonus, 0.10},
		{"BaseXPBonus", BaseXPBonus, 0.05},
		{"WarDeclarationCost", WarDeclarationCost, 1000},
		{"PeaceDeclarationCost", PeaceDeclarationCost, 500},
		{"WarDurationDays", WarDurationDays, 7},
		{"WallBaseHP", WallBaseHP, 1000.0},
		{"TowerBaseHP", TowerBaseHP, 500.0},
		{"GuardBaseHP", GuardBaseHP, 500.0},
		{"TowerDamage", TowerDamage, 100.0},
		{"GuardLevel", GuardLevel, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.value)
			}
		})
	}
}

func TestTerritoryCoords(t *testing.T) {
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	if coords.ChunkX != 10 {
		t.Errorf("expected ChunkX 10, got %d", coords.ChunkX)
	}
	if coords.ChunkZ != 20 {
		t.Errorf("expected ChunkZ 20, got %d", coords.ChunkZ)
	}
}
