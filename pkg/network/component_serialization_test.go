package network

import (
	"testing"
)

// TestComponentSerializer_Position verifies position serialization.
func TestComponentSerializer_Position(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name string
		x, y float64
	}{
		{"zero", 0, 0},
		{"positive", 100.5, 200.75},
		{"negative", -50.25, -150.5},
		{"large", 999999.999, 888888.888},
		{"small", 0.001, 0.002},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data := s.SerializePosition(tt.x, tt.y)
			if len(data) != 16 {
				t.Errorf("Expected 16 bytes, got %d", len(data))
			}

			// Deserialize
			x, y, err := s.DeserializePosition(data)
			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}

			// Verify
			if x != tt.x || y != tt.y {
				t.Errorf("Position mismatch: got (%.2f, %.2f), want (%.2f, %.2f)", x, y, tt.x, tt.y)
			}
		})
	}
}

// TestComponentSerializer_Position_InvalidData verifies error handling.
func TestComponentSerializer_Position_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too_short", []byte{1, 2, 3}},
		{"too_long", make([]byte, 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := s.DeserializePosition(tt.data)
			if err == nil {
				t.Error("Expected error for invalid data")
			}
		})
	}
}

// TestComponentSerializer_Velocity verifies velocity serialization.
func TestComponentSerializer_Velocity(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name   string
		vx, vy float64
	}{
		{"zero", 0, 0},
		{"positive", 50.0, 75.0},
		{"negative", -25.0, -30.0},
		{"mixed", 100.0, -50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeVelocity(tt.vx, tt.vy)
			vx, vy, err := s.DeserializeVelocity(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if vx != tt.vx || vy != tt.vy {
				t.Errorf("Velocity mismatch: got (%.2f, %.2f), want (%.2f, %.2f)", vx, vy, tt.vx, tt.vy)
			}
		})
	}
}

// TestComponentSerializer_Velocity_InvalidData verifies error handling.
func TestComponentSerializer_Velocity_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, _, err := s.DeserializeVelocity([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Health verifies health serialization.
func TestComponentSerializer_Health(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name         string
		current, max float64
	}{
		{"full", 100, 100},
		{"partial", 50, 100},
		{"low", 10, 100},
		{"high", 500, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeHealth(tt.current, tt.max)
			current, max, err := s.DeserializeHealth(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if current != tt.current || max != tt.max {
				t.Errorf("Health mismatch: got (%.0f/%.0f), want (%.0f/%.0f)", current, max, tt.current, tt.max)
			}
		})
	}
}

// TestComponentSerializer_Health_InvalidData verifies error handling.
func TestComponentSerializer_Health_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, _, err := s.DeserializeHealth([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Stats verifies stats serialization.
func TestComponentSerializer_Stats(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name                        string
		attack, defense, magicPower float64
	}{
		{"zero", 0, 0, 0},
		{"balanced", 50, 50, 50},
		{"warrior", 100, 75, 25},
		{"mage", 25, 25, 100},
		{"tank", 50, 150, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeStats(tt.attack, tt.defense, tt.magicPower)
			attack, defense, magicPower, err := s.DeserializeStats(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if attack != tt.attack || defense != tt.defense || magicPower != tt.magicPower {
				t.Errorf("Stats mismatch: got (%.0f/%.0f/%.0f), want (%.0f/%.0f/%.0f)",
					attack, defense, magicPower, tt.attack, tt.defense, tt.magicPower)
			}
		})
	}
}

// TestComponentSerializer_Stats_InvalidData verifies error handling.
func TestComponentSerializer_Stats_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, _, _, err := s.DeserializeStats([]byte{1, 2, 3, 4})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Team verifies team serialization.
func TestComponentSerializer_Team(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name   string
		teamID uint64
	}{
		{"neutral", 0},
		{"team1", 1},
		{"team2", 2},
		{"team_large", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeTeam(tt.teamID)
			teamID, err := s.DeserializeTeam(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if teamID != tt.teamID {
				t.Errorf("Team mismatch: got %d, want %d", teamID, tt.teamID)
			}
		})
	}
}

// TestComponentSerializer_Team_InvalidData verifies error handling.
func TestComponentSerializer_Team_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, err := s.DeserializeTeam([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Level verifies level serialization.
func TestComponentSerializer_Level(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name      string
		level, xp uint32
	}{
		{"level1", 1, 0},
		{"level5", 5, 1000},
		{"level10", 10, 5000},
		{"max", 100, 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeLevel(tt.level, tt.xp)
			level, xp, err := s.DeserializeLevel(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if level != tt.level || xp != tt.xp {
				t.Errorf("Level mismatch: got (%d, %d), want (%d, %d)", level, xp, tt.level, tt.xp)
			}
		})
	}
}

// TestComponentSerializer_Level_InvalidData verifies error handling.
func TestComponentSerializer_Level_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, _, err := s.DeserializeLevel([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Input verifies input serialization.
func TestComponentSerializer_Input(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name   string
		dx, dy int8
	}{
		{"none", 0, 0},
		{"right", 1, 0},
		{"left", -1, 0},
		{"up", 0, -1},
		{"down", 0, 1},
		{"diagonal", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeInput(tt.dx, tt.dy)
			dx, dy, err := s.DeserializeInput(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if dx != tt.dx || dy != tt.dy {
				t.Errorf("Input mismatch: got (%d, %d), want (%d, %d)", dx, dy, tt.dx, tt.dy)
			}
		})
	}
}

// TestComponentSerializer_Input_InvalidData verifies error handling.
func TestComponentSerializer_Input_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, _, err := s.DeserializeInput([]byte{1})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Attack verifies attack serialization.
func TestComponentSerializer_Attack(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name     string
		targetID uint64
	}{
		{"entity1", 1},
		{"entity100", 100},
		{"entity_large", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeAttack(tt.targetID)
			targetID, err := s.DeserializeAttack(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if targetID != tt.targetID {
				t.Errorf("Attack mismatch: got %d, want %d", targetID, tt.targetID)
			}
		})
	}
}

// TestComponentSerializer_Attack_InvalidData verifies error handling.
func TestComponentSerializer_Attack_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, err := s.DeserializeAttack([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_Item verifies item serialization.
func TestComponentSerializer_Item(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name   string
		itemID uint64
	}{
		{"item1", 1},
		{"item50", 50},
		{"item_large", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := s.SerializeItem(tt.itemID)
			itemID, err := s.DeserializeItem(data)

			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}
			if itemID != tt.itemID {
				t.Errorf("Item mismatch: got %d, want %d", itemID, tt.itemID)
			}
		})
	}
}

// TestComponentSerializer_Item_InvalidData verifies error handling.
func TestComponentSerializer_Item_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	_, err := s.DeserializeItem([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

// TestComponentSerializer_NewInstance verifies constructor.
func TestComponentSerializer_NewInstance(t *testing.T) {
	s := NewComponentSerializer()
	if s == nil {
		t.Error("Expected non-nil serializer")
	}
}

// BenchmarkSerializePosition measures position serialization performance.
func BenchmarkSerializePosition(b *testing.B) {
	s := NewComponentSerializer()
	for i := 0; i < b.N; i++ {
		_ = s.SerializePosition(123.45, 678.90)
	}
}

// BenchmarkDeserializePosition measures position deserialization performance.
func BenchmarkDeserializePosition(b *testing.B) {
	s := NewComponentSerializer()
	data := s.SerializePosition(123.45, 678.90)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = s.DeserializePosition(data)
	}
}

// BenchmarkSerializeHealth measures health serialization performance.
func BenchmarkSerializeHealth(b *testing.B) {
	s := NewComponentSerializer()
	for i := 0; i < b.N; i++ {
		_ = s.SerializeHealth(100, 150)
	}
}

// BenchmarkSerializeStats measures stats serialization performance.
func BenchmarkSerializeStats(b *testing.B) {
	s := NewComponentSerializer()
	for i := 0; i < b.N; i++ {
		_ = s.SerializeStats(50, 75, 100)
	}
}
