package guild_vehicle

import (
	"testing"
	"time"
)

func TestFormationType_String(t *testing.T) {
	tests := []struct {
		name      string
		formation FormationType
		want      string
	}{
		{"None", FormationNone, "None"},
		{"Line", FormationLine, "Line"},
		{"Wedge", FormationWedge, "Wedge"},
		{"Column", FormationColumn, "Column"},
		{"Circle", FormationCircle, "Circle"},
		{"Unknown", FormationType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.formation.String(); got != tt.want {
				t.Errorf("FormationType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSiegeEngineType_String(t *testing.T) {
	tests := []struct {
		name  string
		siege SiegeEngineType
		want  string
	}{
		{"None", SiegeNone, "None"},
		{"BatteringRam", SiegeBatteringRam, "BatteringRam"},
		{"Catapult", SiegeCatapult, "Catapult"},
		{"SiegeTower", SiegeTower, "SiegeTower"},
		{"BallistaBattery", SiegeBallistaBattery, "BallistaBattery"},
		{"Unknown", SiegeEngineType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.siege.String(); got != tt.want {
				t.Errorf("SiegeEngineType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSiegeEngineType_GetSiegeDamageMultiplier(t *testing.T) {
	tests := []struct {
		name  string
		siege SiegeEngineType
		want  float64
	}{
		{"None", SiegeNone, 1.0},
		{"BatteringRam", SiegeBatteringRam, 3.0},
		{"Catapult", SiegeCatapult, 5.0},
		{"SiegeTower", SiegeTower, 2.0},
		{"BallistaBattery", SiegeBallistaBattery, 4.0},
		{"Unknown", SiegeEngineType(99), 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.siege.GetSiegeDamageMultiplier(); got != tt.want {
				t.Errorf("SiegeEngineType.GetSiegeDamageMultiplier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFormationBonus(t *testing.T) {
	tests := []struct {
		name        string
		formation   FormationType
		wantDamage  float64
		wantDefense float64
	}{
		{"None", FormationNone, 1.0, 1.0},
		{"Line", FormationLine, 1.05, 1.0},
		{"Wedge", FormationWedge, 1.07, 1.0},
		{"Column", FormationColumn, 1.0, 1.10},
		{"Circle", FormationCircle, 1.0, 1.08},
		{"Unknown", FormationType(99), 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := GetFormationBonus(tt.formation)
			if bonus.DamageMultiplier != tt.wantDamage {
				t.Errorf("GetFormationBonus().DamageMultiplier = %v, want %v", bonus.DamageMultiplier, tt.wantDamage)
			}
			if bonus.DefenseMultiplier != tt.wantDefense {
				t.Errorf("GetFormationBonus().DefenseMultiplier = %v, want %v", bonus.DefenseMultiplier, tt.wantDefense)
			}
			if bonus.Formation != tt.formation && tt.formation != FormationType(99) {
				t.Errorf("GetFormationBonus().Formation = %v, want %v", bonus.Formation, tt.formation)
			}
		})
	}
}

func TestFleet_GetVehicleCount(t *testing.T) {
	fleet := &Fleet{
		FleetID:  "test-fleet",
		GuildID:  "test-guild",
		Vehicles: make(map[uint64]*GuildVehicle),
	}

	if got := fleet.GetVehicleCount(); got != 0 {
		t.Errorf("empty fleet GetVehicleCount() = %v, want 0", got)
	}

	fleet.Vehicles[1] = &GuildVehicle{VehicleID: 1}
	fleet.Vehicles[2] = &GuildVehicle{VehicleID: 2}
	fleet.Vehicles[3] = &GuildVehicle{VehicleID: 3}

	if got := fleet.GetVehicleCount(); got != 3 {
		t.Errorf("fleet with 3 vehicles GetVehicleCount() = %v, want 3", got)
	}
}

func TestFleet_GetTotalMaintenanceCost(t *testing.T) {
	fleet := &Fleet{
		FleetID:  "test-fleet",
		GuildID:  "test-guild",
		Vehicles: make(map[uint64]*GuildVehicle),
	}

	if got := fleet.GetTotalMaintenanceCost(); got != 0 {
		t.Errorf("empty fleet GetTotalMaintenanceCost() = %v, want 0", got)
	}

	fleet.Vehicles[1] = &GuildVehicle{VehicleID: 1, MaintenanceCost: 100}
	fleet.Vehicles[2] = &GuildVehicle{VehicleID: 2, MaintenanceCost: 250}
	fleet.Vehicles[3] = &GuildVehicle{VehicleID: 3, MaintenanceCost: 500}

	if got := fleet.GetTotalMaintenanceCost(); got != 850 {
		t.Errorf("fleet with 850 total cost GetTotalMaintenanceCost() = %v, want 850", got)
	}
}

func TestFleet_GetSiegeEngineCount(t *testing.T) {
	fleet := &Fleet{
		FleetID:  "test-fleet",
		GuildID:  "test-guild",
		Vehicles: make(map[uint64]*GuildVehicle),
	}

	if got := fleet.GetSiegeEngineCount(); got != 0 {
		t.Errorf("empty fleet GetSiegeEngineCount() = %v, want 0", got)
	}

	fleet.Vehicles[1] = &GuildVehicle{VehicleID: 1, SiegeType: SiegeNone}
	fleet.Vehicles[2] = &GuildVehicle{VehicleID: 2, SiegeType: SiegeBatteringRam}
	fleet.Vehicles[3] = &GuildVehicle{VehicleID: 3, SiegeType: SiegeCatapult}
	fleet.Vehicles[4] = &GuildVehicle{VehicleID: 4, SiegeType: SiegeNone}

	if got := fleet.GetSiegeEngineCount(); got != 2 {
		t.Errorf("fleet with 2 siege engines GetSiegeEngineCount() = %v, want 2", got)
	}
}

func TestGuildVehicle_HasAccess(t *testing.T) {
	vehicle := &GuildVehicle{
		VehicleID:    1,
		SharedAccess: make(map[string]bool),
	}

	// Test with no access
	if vehicle.HasAccess("player1") {
		t.Error("expected no access for player1")
	}

	// Grant access
	vehicle.SharedAccess["player1"] = true
	if !vehicle.HasAccess("player1") {
		t.Error("expected access for player1")
	}

	// Test different player
	if vehicle.HasAccess("player2") {
		t.Error("expected no access for player2")
	}

	// Test nil SharedAccess map
	vehicle.SharedAccess = nil
	if vehicle.HasAccess("player1") {
		t.Error("expected no access when SharedAccess is nil")
	}
}

func TestGuildVehicleFleetComponent_Type(t *testing.T) {
	comp := &GuildVehicleFleetComponent{}
	if got := comp.Type(); got != "guild_vehicle_fleet" {
		t.Errorf("Type() = %v, want guild_vehicle_fleet", got)
	}
}

// Benchmarks

func BenchmarkGetFormationBonus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetFormationBonus(FormationWedge)
	}
}

func BenchmarkFleet_GetVehicleCount(b *testing.B) {
	fleet := &Fleet{
		Vehicles: make(map[uint64]*GuildVehicle),
	}
	for i := uint64(0); i < 20; i++ {
		fleet.Vehicles[i] = &GuildVehicle{VehicleID: i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fleet.GetVehicleCount()
	}
}

func BenchmarkFleet_GetTotalMaintenanceCost(b *testing.B) {
	fleet := &Fleet{
		Vehicles: make(map[uint64]*GuildVehicle),
	}
	for i := uint64(0); i < 20; i++ {
		fleet.Vehicles[i] = &GuildVehicle{VehicleID: i, MaintenanceCost: 100}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fleet.GetTotalMaintenanceCost()
	}
}

func BenchmarkFleet_GetSiegeEngineCount(b *testing.B) {
	fleet := &Fleet{
		Vehicles: make(map[uint64]*GuildVehicle),
	}
	for i := uint64(0); i < 20; i++ {
		siegeType := SiegeNone
		if i%5 == 0 {
			siegeType = SiegeCatapult
		}
		fleet.Vehicles[i] = &GuildVehicle{VehicleID: i, SiegeType: siegeType}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fleet.GetSiegeEngineCount()
	}
}

func BenchmarkGuildVehicle_HasAccess(b *testing.B) {
	vehicle := &GuildVehicle{
		VehicleID:    1,
		SharedAccess: make(map[string]bool),
	}
	vehicle.SharedAccess["player1"] = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vehicle.HasAccess("player1")
	}
}

func TestGuildVehicleFleetComponent_Serialize(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		comp    *GuildVehicleFleetComponent
		wantErr bool
	}{
		{
			name: "basic component",
			comp: &GuildVehicleFleetComponent{
				GuildID:             "guild-123",
				FleetID:             "fleet-456",
				SiegeType:           SiegeCatapult,
				FormationPosition:   2,
				LastFormationUpdate: now,
			},
			wantErr: false,
		},
		{
			name:    "empty component",
			comp:    &GuildVehicleFleetComponent{},
			wantErr: false,
		},
		{
			name: "siege engine leader",
			comp: &GuildVehicleFleetComponent{
				GuildID:             "siege-guild",
				FleetID:             "siege-fleet",
				SiegeType:           SiegeBatteringRam,
				FormationPosition:   0,
				LastFormationUpdate: now,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.comp.Serialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("Serialize() returned empty data")
			}
		})
	}
}

func TestGuildVehicleFleetComponent_Deserialize(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		comp    *GuildVehicleFleetComponent
		wantErr bool
	}{
		{
			name: "round trip basic",
			comp: &GuildVehicleFleetComponent{
				GuildID:             "guild-123",
				FleetID:             "fleet-456",
				SiegeType:           SiegeCatapult,
				FormationPosition:   2,
				LastFormationUpdate: now,
			},
			wantErr: false,
		},
		{
			name:    "round trip empty",
			comp:    &GuildVehicleFleetComponent{},
			wantErr: false,
		},
		{
			name: "round trip all siege types",
			comp: &GuildVehicleFleetComponent{
				GuildID:             "test-guild",
				FleetID:             "test-fleet",
				SiegeType:           SiegeBallistaBattery,
				FormationPosition:   5,
				LastFormationUpdate: now,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := tt.comp.Serialize()
			if err != nil {
				t.Fatalf("Serialize() failed: %v", err)
			}

			// Deserialize into new component
			restored := &GuildVehicleFleetComponent{}
			err = restored.Deserialize(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify fields match
			if restored.GuildID != tt.comp.GuildID {
				t.Errorf("GuildID = %v, want %v", restored.GuildID, tt.comp.GuildID)
			}
			if restored.FleetID != tt.comp.FleetID {
				t.Errorf("FleetID = %v, want %v", restored.FleetID, tt.comp.FleetID)
			}
			if restored.SiegeType != tt.comp.SiegeType {
				t.Errorf("SiegeType = %v, want %v", restored.SiegeType, tt.comp.SiegeType)
			}
			if restored.FormationPosition != tt.comp.FormationPosition {
				t.Errorf("FormationPosition = %v, want %v", restored.FormationPosition, tt.comp.FormationPosition)
			}
		})
	}
}

func TestGuildVehicleFleetComponent_Deserialize_InvalidData(t *testing.T) {
	comp := &GuildVehicleFleetComponent{}

	// Test with invalid JSON
	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}

	// Test with empty data
	err = comp.Deserialize([]byte{})
	if err == nil {
		t.Error("Deserialize() expected error for empty data")
	}
}

func BenchmarkGuildVehicleFleetComponent_Serialize(b *testing.B) {
	comp := &GuildVehicleFleetComponent{
		GuildID:             "benchmark-guild",
		FleetID:             "benchmark-fleet",
		SiegeType:           SiegeCatapult,
		FormationPosition:   3,
		LastFormationUpdate: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = comp.Serialize()
	}
}

func BenchmarkGuildVehicleFleetComponent_Deserialize(b *testing.B) {
	comp := &GuildVehicleFleetComponent{
		GuildID:             "benchmark-guild",
		FleetID:             "benchmark-fleet",
		SiegeType:           SiegeCatapult,
		FormationPosition:   3,
		LastFormationUpdate: time.Now(),
	}

	data, _ := comp.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		restored := &GuildVehicleFleetComponent{}
		_ = restored.Deserialize(data)
	}
}
