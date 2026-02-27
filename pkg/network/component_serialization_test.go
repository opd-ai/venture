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

// TestComponentSerializer_Expression verifies expression serialization.
func TestComponentSerializer_Expression(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name           string
		expressionType uint8
		expressionTime float64
		cooldown       float64
	}{
		{"Wave", 0, 3.0, 0.0},
		{"Dance", 2, 5.5, 1.5},
		{"Sleep", 11, 999.0, 3.0},
		{"Zero values", 0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data := s.SerializeExpression(tt.expressionType, tt.expressionTime, tt.cooldown)
			if len(data) != 17 {
				t.Errorf("Expected 17 bytes, got %d", len(data))
			}

			// Deserialize
			expType, expTime, cd, err := s.DeserializeExpression(data)
			if err != nil {
				t.Errorf("Deserialize failed: %v", err)
			}

			// Verify
			if expType != tt.expressionType {
				t.Errorf("ExpressionType mismatch: got %d, want %d", expType, tt.expressionType)
			}
			if expTime != tt.expressionTime {
				t.Errorf("ExpressionTime mismatch: got %.2f, want %.2f", expTime, tt.expressionTime)
			}
			if cd != tt.cooldown {
				t.Errorf("Cooldown mismatch: got %.2f, want %.2f", cd, tt.cooldown)
			}
		})
	}
}

// TestComponentSerializer_Expression_InvalidData verifies error handling.
func TestComponentSerializer_Expression_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too short", []byte{1, 2, 3}},
		{"too long", make([]byte, 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := s.DeserializeExpression(tt.data)
			if err == nil {
				t.Error("Expected error for invalid data, got nil")
			}
		})
	}
}

// TestComponentSerializer_Expression_SizeRequirement verifies <50 byte requirement.
func TestComponentSerializer_Expression_SizeRequirement(t *testing.T) {
	s := NewComponentSerializer()

	data := s.SerializeExpression(5, 3.0, 1.0)

	if len(data) >= 50 {
		t.Errorf("Expression serialization = %d bytes, want < 50 bytes", len(data))
	}
}

// BenchmarkSerializeExpression measures expression serialization performance.
func BenchmarkSerializeExpression(b *testing.B) {
	s := NewComponentSerializer()
	for i := 0; i < b.N; i++ {
		_ = s.SerializeExpression(2, 3.0, 1.5)
	}
}

// BenchmarkDeserializeExpression measures expression deserialization performance.
func BenchmarkDeserializeExpression(b *testing.B) {
	s := NewComponentSerializer()
	data := s.SerializeExpression(2, 3.0, 1.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = s.DeserializeExpression(data)
	}
}

// TestComponentSerializer_Territory verifies territory serialization.
func TestComponentSerializer_Territory(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name    string
		data    TerritoryData
		wantErr bool
	}{
		{
			name: "neutral_territory",
			data: TerritoryData{
				ID:              "territory-001",
				ChunkX:          10,
				ChunkZ:          20,
				OwnerGuildID:    "",
				Status:          0, // StatusNeutral
				CaptureProgress: 0.0,
				CapturingGuild:  "",
				LastUpdateUnix:  1234567890,
				ResourceBonus:   0.10,
				XPBonus:         0.05,
			},
			wantErr: false,
		},
		{
			name: "owned_territory",
			data: TerritoryData{
				ID:              "territory-002",
				ChunkX:          -5,
				ChunkZ:          15,
				OwnerGuildID:    "guild-123",
				Status:          1, // StatusOwned
				CaptureProgress: 0.0,
				CapturingGuild:  "",
				LastUpdateUnix:  1234567891,
				ResourceBonus:   0.15,
				XPBonus:         0.08,
			},
			wantErr: false,
		},
		{
			name: "contested_territory",
			data: TerritoryData{
				ID:              "territory-003",
				ChunkX:          0,
				ChunkZ:          0,
				OwnerGuildID:    "guild-456",
				Status:          2, // StatusContested
				CaptureProgress: 0.67,
				CapturingGuild:  "guild-789",
				LastUpdateUnix:  1234567892,
				ResourceBonus:   0.20,
				XPBonus:         0.10,
			},
			wantErr: false,
		},
		{
			name: "edge_case_coords",
			data: TerritoryData{
				ID:              "territory-edge",
				ChunkX:          -999999,
				ChunkZ:          999999,
				OwnerGuildID:    "",
				Status:          0,
				CaptureProgress: 1.0,
				CapturingGuild:  "guild-max",
				LastUpdateUnix:  9999999999,
				ResourceBonus:   0.50,
				XPBonus:         0.25,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized, err := s.SerializeTerritory(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeTerritory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(serialized) == 0 {
				t.Error("SerializeTerritory() returned empty bytes")
				return
			}

			// Deserialize
			deserialized, err := s.DeserializeTerritory(serialized)
			if err != nil {
				t.Errorf("DeserializeTerritory() error = %v", err)
				return
			}

			// Verify all fields
			if deserialized.ID != tt.data.ID {
				t.Errorf("ID mismatch: got %v, want %v", deserialized.ID, tt.data.ID)
			}
			if deserialized.ChunkX != tt.data.ChunkX {
				t.Errorf("ChunkX mismatch: got %v, want %v", deserialized.ChunkX, tt.data.ChunkX)
			}
			if deserialized.ChunkZ != tt.data.ChunkZ {
				t.Errorf("ChunkZ mismatch: got %v, want %v", deserialized.ChunkZ, tt.data.ChunkZ)
			}
			if deserialized.OwnerGuildID != tt.data.OwnerGuildID {
				t.Errorf("OwnerGuildID mismatch: got %v, want %v", deserialized.OwnerGuildID, tt.data.OwnerGuildID)
			}
			if deserialized.Status != tt.data.Status {
				t.Errorf("Status mismatch: got %v, want %v", deserialized.Status, tt.data.Status)
			}
			if deserialized.CaptureProgress != tt.data.CaptureProgress {
				t.Errorf("CaptureProgress mismatch: got %v, want %v", deserialized.CaptureProgress, tt.data.CaptureProgress)
			}
			if deserialized.CapturingGuild != tt.data.CapturingGuild {
				t.Errorf("CapturingGuild mismatch: got %v, want %v", deserialized.CapturingGuild, tt.data.CapturingGuild)
			}
			if deserialized.LastUpdateUnix != tt.data.LastUpdateUnix {
				t.Errorf("LastUpdateUnix mismatch: got %v, want %v", deserialized.LastUpdateUnix, tt.data.LastUpdateUnix)
			}
			if deserialized.ResourceBonus != tt.data.ResourceBonus {
				t.Errorf("ResourceBonus mismatch: got %v, want %v", deserialized.ResourceBonus, tt.data.ResourceBonus)
			}
			if deserialized.XPBonus != tt.data.XPBonus {
				t.Errorf("XPBonus mismatch: got %v, want %v", deserialized.XPBonus, tt.data.XPBonus)
			}
		})
	}
}

// TestComponentSerializer_Territory_InvalidData verifies error handling.
func TestComponentSerializer_Territory_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"corrupted", []byte{1, 2, 3, 4, 5}},
		{"truncated", []byte{0x0e, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.DeserializeTerritory(tt.data)
			if err == nil {
				t.Error("DeserializeTerritory() expected error, got nil")
			}
		})
	}
}

// TestComponentSerializer_Siege verifies siege serialization.
func TestComponentSerializer_Siege(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name    string
		data    SiegeData
		wantErr bool
	}{
		{
			name: "preparation_phase",
			data: SiegeData{
				ID:                    "siege-001",
				TerritoryID:           "territory-123",
				AttackerGuildID:       "guild-attack",
				DefenderGuildID:       "guild-defend",
				Phase:                 0, // PhasePreparation
				StartTimeUnix:         1234567890,
				PhaseStartTimeUnix:    1234567890,
				EndTimeUnix:           0,
				VictoryCondition:      0,
				WinnerGuildID:         "",
				Attackers:             []string{"player-1", "player-2", "player-3"},
				Defenders:             []string{"player-4", "player-5"},
				ControlPointsCaptured: 0,
				TotalControlPoints:    5,
				GuildHallHP:           1000.0,
				GuildHallMaxHP:        1000.0,
				DefenderTreasury:      50000,
				LootPercentage:        0.30,
				LootDistributed:       false,
			},
			wantErr: false,
		},
		{
			name: "assault_phase",
			data: SiegeData{
				ID:                    "siege-002",
				TerritoryID:           "territory-456",
				AttackerGuildID:       "guild-red",
				DefenderGuildID:       "guild-blue",
				Phase:                 1, // PhaseAssault
				StartTimeUnix:         1234567890,
				PhaseStartTimeUnix:    1234571490, // 1 hour later
				EndTimeUnix:           0,
				VictoryCondition:      0,
				WinnerGuildID:         "",
				Attackers:             []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"},
				Defenders:             []string{"p11", "p12", "p13", "p14", "p15"},
				ControlPointsCaptured: 3,
				TotalControlPoints:    5,
				GuildHallHP:           650.5,
				GuildHallMaxHP:        1000.0,
				DefenderTreasury:      75000,
				LootPercentage:        0.30,
				LootDistributed:       false,
			},
			wantErr: false,
		},
		{
			name: "ended_victory",
			data: SiegeData{
				ID:                    "siege-003",
				TerritoryID:           "territory-789",
				AttackerGuildID:       "guild-alpha",
				DefenderGuildID:       "guild-beta",
				Phase:                 3, // PhaseEnded
				StartTimeUnix:         1234567890,
				PhaseStartTimeUnix:    1234578690,
				EndTimeUnix:           1234585890,
				VictoryCondition:      0, // VictoryCapturePoints
				WinnerGuildID:         "guild-alpha",
				Attackers:             []string{"player-a1", "player-a2"},
				Defenders:             []string{"player-d1"},
				ControlPointsCaptured: 5,
				TotalControlPoints:    5,
				GuildHallHP:           1000.0,
				GuildHallMaxHP:        1000.0,
				DefenderTreasury:      100000,
				LootPercentage:        0.30,
				LootDistributed:       true,
			},
			wantErr: false,
		},
		{
			name: "empty_participants",
			data: SiegeData{
				ID:                    "siege-empty",
				TerritoryID:           "territory-000",
				AttackerGuildID:       "guild-x",
				DefenderGuildID:       "guild-y",
				Phase:                 0,
				StartTimeUnix:         1000000000,
				PhaseStartTimeUnix:    1000000000,
				EndTimeUnix:           0,
				VictoryCondition:      0,
				WinnerGuildID:         "",
				Attackers:             []string{},
				Defenders:             []string{},
				ControlPointsCaptured: 0,
				TotalControlPoints:    5,
				GuildHallHP:           1000.0,
				GuildHallMaxHP:        1000.0,
				DefenderTreasury:      0,
				LootPercentage:        0.30,
				LootDistributed:       false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized, err := s.SerializeSiege(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeSiege() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(serialized) == 0 {
				t.Error("SerializeSiege() returned empty bytes")
				return
			}

			// Deserialize
			deserialized, err := s.DeserializeSiege(serialized)
			if err != nil {
				t.Errorf("DeserializeSiege() error = %v", err)
				return
			}

			// Verify all fields
			if deserialized.ID != tt.data.ID {
				t.Errorf("ID mismatch: got %v, want %v", deserialized.ID, tt.data.ID)
			}
			if deserialized.TerritoryID != tt.data.TerritoryID {
				t.Errorf("TerritoryID mismatch: got %v, want %v", deserialized.TerritoryID, tt.data.TerritoryID)
			}
			if deserialized.AttackerGuildID != tt.data.AttackerGuildID {
				t.Errorf("AttackerGuildID mismatch: got %v, want %v", deserialized.AttackerGuildID, tt.data.AttackerGuildID)
			}
			if deserialized.DefenderGuildID != tt.data.DefenderGuildID {
				t.Errorf("DefenderGuildID mismatch: got %v, want %v", deserialized.DefenderGuildID, tt.data.DefenderGuildID)
			}
			if deserialized.Phase != tt.data.Phase {
				t.Errorf("Phase mismatch: got %v, want %v", deserialized.Phase, tt.data.Phase)
			}
			if deserialized.StartTimeUnix != tt.data.StartTimeUnix {
				t.Errorf("StartTimeUnix mismatch: got %v, want %v", deserialized.StartTimeUnix, tt.data.StartTimeUnix)
			}
			if deserialized.PhaseStartTimeUnix != tt.data.PhaseStartTimeUnix {
				t.Errorf("PhaseStartTimeUnix mismatch: got %v, want %v", deserialized.PhaseStartTimeUnix, tt.data.PhaseStartTimeUnix)
			}
			if deserialized.VictoryCondition != tt.data.VictoryCondition {
				t.Errorf("VictoryCondition mismatch: got %v, want %v", deserialized.VictoryCondition, tt.data.VictoryCondition)
			}
			if deserialized.WinnerGuildID != tt.data.WinnerGuildID {
				t.Errorf("WinnerGuildID mismatch: got %v, want %v", deserialized.WinnerGuildID, tt.data.WinnerGuildID)
			}
			if len(deserialized.Attackers) != len(tt.data.Attackers) {
				t.Errorf("Attackers length mismatch: got %v, want %v", len(deserialized.Attackers), len(tt.data.Attackers))
			}
			if len(deserialized.Defenders) != len(tt.data.Defenders) {
				t.Errorf("Defenders length mismatch: got %v, want %v", len(deserialized.Defenders), len(tt.data.Defenders))
			}
			if deserialized.ControlPointsCaptured != tt.data.ControlPointsCaptured {
				t.Errorf("ControlPointsCaptured mismatch: got %v, want %v", deserialized.ControlPointsCaptured, tt.data.ControlPointsCaptured)
			}
			if deserialized.TotalControlPoints != tt.data.TotalControlPoints {
				t.Errorf("TotalControlPoints mismatch: got %v, want %v", deserialized.TotalControlPoints, tt.data.TotalControlPoints)
			}
			if deserialized.GuildHallHP != tt.data.GuildHallHP {
				t.Errorf("GuildHallHP mismatch: got %v, want %v", deserialized.GuildHallHP, tt.data.GuildHallHP)
			}
			if deserialized.GuildHallMaxHP != tt.data.GuildHallMaxHP {
				t.Errorf("GuildHallMaxHP mismatch: got %v, want %v", deserialized.GuildHallMaxHP, tt.data.GuildHallMaxHP)
			}
			if deserialized.DefenderTreasury != tt.data.DefenderTreasury {
				t.Errorf("DefenderTreasury mismatch: got %v, want %v", deserialized.DefenderTreasury, tt.data.DefenderTreasury)
			}
			if deserialized.LootPercentage != tt.data.LootPercentage {
				t.Errorf("LootPercentage mismatch: got %v, want %v", deserialized.LootPercentage, tt.data.LootPercentage)
			}
			if deserialized.LootDistributed != tt.data.LootDistributed {
				t.Errorf("LootDistributed mismatch: got %v, want %v", deserialized.LootDistributed, tt.data.LootDistributed)
			}
		})
	}
}

// TestComponentSerializer_Siege_InvalidData verifies error handling.
func TestComponentSerializer_Siege_InvalidData(t *testing.T) {
	s := NewComponentSerializer()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"corrupted", []byte{1, 2, 3, 4, 5}},
		{"truncated", []byte{0x0e, 0x00, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.DeserializeSiege(tt.data)
			if err == nil {
				t.Error("DeserializeSiege() expected error, got nil")
			}
		})
	}
}

// BenchmarkSerializeTerritory measures territory serialization performance.
func BenchmarkSerializeTerritory(b *testing.B) {
	s := NewComponentSerializer()
	data := TerritoryData{
		ID:              "territory-bench",
		ChunkX:          100,
		ChunkZ:          200,
		OwnerGuildID:    "guild-benchmark",
		Status:          1,
		CaptureProgress: 0.5,
		CapturingGuild:  "guild-attacker",
		LastUpdateUnix:  1234567890,
		ResourceBonus:   0.15,
		XPBonus:         0.08,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.SerializeTerritory(data)
	}
}

// BenchmarkDeserializeTerritory measures territory deserialization performance.
func BenchmarkDeserializeTerritory(b *testing.B) {
	s := NewComponentSerializer()
	data := TerritoryData{
		ID:              "territory-bench",
		ChunkX:          100,
		ChunkZ:          200,
		OwnerGuildID:    "guild-benchmark",
		Status:          1,
		CaptureProgress: 0.5,
		CapturingGuild:  "guild-attacker",
		LastUpdateUnix:  1234567890,
		ResourceBonus:   0.15,
		XPBonus:         0.08,
	}
	serialized, _ := s.SerializeTerritory(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DeserializeTerritory(serialized)
	}
}

// BenchmarkSerializeSiege measures siege serialization performance.
func BenchmarkSerializeSiege(b *testing.B) {
	s := NewComponentSerializer()
	data := SiegeData{
		ID:                    "siege-bench",
		TerritoryID:           "territory-123",
		AttackerGuildID:       "guild-attack",
		DefenderGuildID:       "guild-defend",
		Phase:                 1,
		StartTimeUnix:         1234567890,
		PhaseStartTimeUnix:    1234571490,
		EndTimeUnix:           0,
		VictoryCondition:      0,
		WinnerGuildID:         "",
		Attackers:             []string{"p1", "p2", "p3", "p4", "p5"},
		Defenders:             []string{"p6", "p7", "p8"},
		ControlPointsCaptured: 2,
		TotalControlPoints:    5,
		GuildHallHP:           750.0,
		GuildHallMaxHP:        1000.0,
		DefenderTreasury:      60000,
		LootPercentage:        0.30,
		LootDistributed:       false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.SerializeSiege(data)
	}
}

// BenchmarkDeserializeSiege measures siege deserialization performance.
func BenchmarkDeserializeSiege(b *testing.B) {
	s := NewComponentSerializer()
	data := SiegeData{
		ID:                    "siege-bench",
		TerritoryID:           "territory-123",
		AttackerGuildID:       "guild-attack",
		DefenderGuildID:       "guild-defend",
		Phase:                 1,
		StartTimeUnix:         1234567890,
		PhaseStartTimeUnix:    1234571490,
		EndTimeUnix:           0,
		VictoryCondition:      0,
		WinnerGuildID:         "",
		Attackers:             []string{"p1", "p2", "p3", "p4", "p5"},
		Defenders:             []string{"p6", "p7", "p8"},
		ControlPointsCaptured: 2,
		TotalControlPoints:    5,
		GuildHallHP:           750.0,
		GuildHallMaxHP:        1000.0,
		DefenderTreasury:      60000,
		LootPercentage:        0.30,
		LootDistributed:       false,
	}
	serialized, _ := s.SerializeSiege(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DeserializeSiege(serialized)
	}
}
