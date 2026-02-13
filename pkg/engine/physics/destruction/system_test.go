package destruction

import (
	"testing"
)

func TestIntegrityState_String(t *testing.T) {
	tests := []struct {
		name  string
		state IntegrityState
		want  string
	}{
		{"pristine", IntegrityPristine, "Pristine"},
		{"damaged", IntegrityDamaged, "Damaged"},
		{"critical", IntegrityCritical, "Critical"},
		{"collapsed", IntegrityCollapsed, "Collapsed"},
		{"unknown", IntegrityState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("IntegrityState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportType_String(t *testing.T) {
	tests := []struct {
		name string
		typ  SupportType
		want string
	}{
		{"wall", SupportWall, "Wall"},
		{"column", SupportColumn, "Column"},
		{"beam", SupportBeam, "Beam"},
		{"foundation", SupportFoundation, "Foundation"},
		{"unknown", SupportType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("SupportType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaterialType_String(t *testing.T) {
	tests := []struct {
		name string
		mat  MaterialType
		want string
	}{
		{"wood", MaterialWood, "Wood"},
		{"stone", MaterialStone, "Stone"},
		{"metal", MaterialMetal, "Metal"},
		{"glass", MaterialGlass, "Glass"},
		{"concrete", MaterialConcrete, "Concrete"},
		{"brick", MaterialBrick, "Brick"},
		{"unknown", MaterialType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mat.String(); got != tt.want {
				t.Errorf("MaterialType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMaterialProperties(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		validate func(*testing.T, MaterialProperties)
	}{
		{
			name:     "wood_properties",
			material: MaterialWood,
			validate: func(t *testing.T, props MaterialProperties) {
				if props.Flammable != true {
					t.Error("Wood should be flammable")
				}
				if props.Durability >= props.Density {
					t.Error("Wood durability should be < density")
				}
			},
		},
		{
			name:     "stone_properties",
			material: MaterialStone,
			validate: func(t *testing.T, props MaterialProperties) {
				if props.Flammable != false {
					t.Error("Stone should not be flammable")
				}
				if props.Durability < 0.8 {
					t.Error("Stone should be very durable")
				}
			},
		},
		{
			name:     "glass_properties",
			material: MaterialGlass,
			validate: func(t *testing.T, props MaterialProperties) {
				if props.Durability > 0.2 {
					t.Error("Glass should be fragile")
				}
				if props.Friction > 0.2 {
					t.Error("Glass should have low friction")
				}
			},
		},
		{
			name:     "concrete_properties",
			material: MaterialConcrete,
			validate: func(t *testing.T, props MaterialProperties) {
				if props.Density != 1.0 {
					t.Error("Concrete should have highest density")
				}
				if props.Bounciness > 0.1 {
					t.Error("Concrete should have very low bounciness")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := GetMaterialProperties(tt.material)

			if props.Density < 0 || props.Density > 1 {
				t.Errorf("Density out of range: %v", props.Density)
			}
			if props.Bounciness < 0 || props.Bounciness > 1 {
				t.Errorf("Bounciness out of range: %v", props.Bounciness)
			}
			if props.Friction < 0 || props.Friction > 1 {
				t.Errorf("Friction out of range: %v", props.Friction)
			}
			if props.Durability < 0 || props.Durability > 1 {
				t.Errorf("Durability out of range: %v", props.Durability)
			}

			tt.validate(t, props)
		})
	}
}

func TestFallingObject_IsGrounded(t *testing.T) {
	tests := []struct {
		name string
		obj  FallingObject
		want bool
	}{
		{
			name: "on_ground_flag",
			obj:  FallingObject{OnGround: true, Z: 0, VelZ: 0},
			want: true,
		},
		{
			name: "near_ground_slow",
			obj:  FallingObject{OnGround: false, Z: 0.05, VelZ: 0.05},
			want: true,
		},
		{
			name: "above_ground",
			obj:  FallingObject{OnGround: false, Z: 10, VelZ: -5},
			want: false,
		},
		{
			name: "near_ground_fast",
			obj:  FallingObject{OnGround: false, Z: 0.05, VelZ: 5},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.obj.IsGrounded(); got != tt.want {
				t.Errorf("IsGrounded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.EnableIntegrityChecks != true {
		t.Error("Default config should enable integrity checks")
	}
	if cfg.EnableDebris != true {
		t.Error("Default config should enable debris")
	}
	if cfg.Gravity <= 0 {
		t.Error("Gravity should be positive")
	}
	if cfg.MaxDebrisParticles <= 0 {
		t.Error("MaxDebrisParticles should be positive")
	}
	if cfg.MaxFallingObjects <= 0 {
		t.Error("MaxFallingObjects should be positive")
	}
	if cfg.CollapseThreshold <= 0 || cfg.CollapseThreshold >= 1 {
		t.Error("CollapseThreshold should be between 0 and 1")
	}
	if cfg.UpdateFrequency <= 0 {
		t.Error("UpdateFrequency should be positive")
	}
}

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{"with_config", DefaultConfig()},
		{"nil_config", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.config)
			if sys == nil {
				t.Fatal("NewSystem() returned nil")
			}
			if sys.config == nil {
				t.Error("System config is nil")
			}
			if sys.buildings == nil {
				t.Error("System buildings map is nil")
			}
			if sys.debris == nil {
				t.Error("System debris slice is nil")
			}
			if sys.fallingObjects == nil {
				t.Error("System fallingObjects slice is nil")
			}
		})
	}
}

func TestSystem_RegisterBuilding(t *testing.T) {
	sys := NewSystem(nil)

	tests := []struct {
		name     string
		id       string
		width    int
		height   int
		floors   int
		material MaterialType
		wantErr  bool
	}{
		{"small_house", "house1", 8, 8, 1, MaterialWood, false},
		{"large_manor", "manor1", 24, 24, 3, MaterialStone, false},
		{"tower", "tower1", 6, 16, 5, MaterialBrick, false},
		{"empty_id", "", 8, 8, 1, MaterialWood, true},
		{"zero_width", "invalid1", 0, 8, 1, MaterialWood, true},
		{"negative_height", "invalid2", 8, -1, 1, MaterialWood, true},
		{"zero_floors", "invalid3", 8, 8, 0, MaterialWood, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.RegisterBuilding(tt.id, tt.width, tt.height, tt.floors, tt.material)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterBuilding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			integrity, err := sys.GetIntegrity(tt.id)
			if err != nil {
				t.Fatalf("GetIntegrity() error = %v", err)
			}

			if integrity.BuildingID != tt.id {
				t.Errorf("BuildingID = %v, want %v", integrity.BuildingID, tt.id)
			}
			if integrity.State != IntegrityPristine {
				t.Errorf("State = %v, want %v", integrity.State, IntegrityPristine)
			}
			if integrity.CurrentHealth != 1.0 {
				t.Errorf("CurrentHealth = %v, want 1.0", integrity.CurrentHealth)
			}
			if len(integrity.Supports) == 0 {
				t.Error("No supports generated")
			}

			foundFoundation := false
			for _, support := range integrity.Supports {
				if support.Type == SupportFoundation {
					foundFoundation = true
					break
				}
			}
			if !foundFoundation {
				t.Error("No foundation supports found")
			}
		})
	}
}

func TestSystem_ApplyDamage(t *testing.T) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("building1", 10, 10, 1, MaterialStone)

	tests := []struct {
		name         string
		buildingID   string
		x, y, floor  int
		amount       float64
		radius       float64
		wantErr      bool
		validateFunc func(*testing.T, *StructuralIntegrity)
	}{
		{
			name:       "valid_damage",
			buildingID: "building1",
			x:          5,
			y:          5,
			floor:      0,
			amount:     0.9,
			radius:     10.0,
			wantErr:    false,
			validateFunc: func(t *testing.T, integrity *StructuralIntegrity) {
				if len(integrity.DamagedAreas) == 0 {
					t.Error("No damaged areas recorded")
				}
				if integrity.CurrentHealth > 0.9 {
					t.Errorf("Health should decrease after damage, got %v", integrity.CurrentHealth)
				}
			},
		},
		{
			name:       "invalid_building",
			buildingID: "nonexistent",
			x:          5,
			y:          5,
			floor:      0,
			amount:     0.3,
			radius:     3.0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.ApplyDamage(tt.buildingID, tt.x, tt.y, tt.floor, tt.amount, tt.radius)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyDamage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFunc != nil {
				integrity, _ := sys.GetIntegrity(tt.buildingID)
				tt.validateFunc(t, integrity)
			}
		})
	}
}

func TestSystem_Update(t *testing.T) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("building1", 10, 10, 1, MaterialWood)

	_ = sys.ApplyDamage("building1", 5, 5, 0, 0.3, 5.0)

	initialHealth := 0.0
	if integrity, _ := sys.GetIntegrity("building1"); integrity != nil {
		initialHealth = integrity.CurrentHealth
	}

	for i := 0; i < 100; i++ {
		sys.Update(0.1)
	}

	integrity, _ := sys.GetIntegrity("building1")
	if integrity.CurrentHealth > initialHealth {
		t.Errorf("Health should not increase, initial=%v current=%v", initialHealth, integrity.CurrentHealth)
	}
}

func TestSystem_UpdateZeroDelta(t *testing.T) {
	sys := NewSystem(nil)
	sys.Update(0)
	sys.Update(-0.1)
}

func TestSystem_Collapse(t *testing.T) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("fragile", 8, 8, 1, MaterialGlass)

	_ = sys.ApplyDamage("fragile", 4, 4, 0, 0.95, 15.0)

	for i := 0; i < 100; i++ {
		sys.Update(0.1)
	}

	integrity, _ := sys.GetIntegrity("fragile")
	if integrity.State != IntegrityCollapsed && integrity.State != IntegrityCritical {
		t.Errorf("Building should be critical or collapsed, state = %v (health=%v)",
			integrity.State, integrity.CurrentHealth)
	}

	if integrity.State == IntegrityCollapsed && sys.config.EnableDebris && sys.GetDebrisCount() == 0 {
		t.Error("Collapsed building should generate debris")
	}
}

func TestSystem_RemoveBuilding(t *testing.T) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("temp", 8, 8, 1, MaterialWood)

	_, err := sys.GetIntegrity("temp")
	if err != nil {
		t.Fatal("Building should exist")
	}

	err = sys.RemoveBuilding("temp")
	if err != nil {
		t.Errorf("RemoveBuilding() unexpected error = %v", err)
	}

	_, err = sys.GetIntegrity("temp")
	if err == nil {
		t.Error("Building should not exist after removal")
	}

	// Test removing nonexistent building
	err = sys.RemoveBuilding("nonexistent")
	if err == nil {
		t.Error("RemoveBuilding() should return error for nonexistent building")
	}
}

func TestSystem_SpawnFallingObject(t *testing.T) {
	sys := NewSystem(nil)

	err := sys.SpawnFallingObject(100, 100, 200, MaterialStone, 16, 16)
	if err != nil {
		t.Errorf("SpawnFallingObject() unexpected error = %v", err)
	}

	if sys.GetFallingObjectCount() != 1 {
		t.Errorf("Falling object count = %v, want 1", sys.GetFallingObjectCount())
	}

	// Fill up to the limit
	for i := 1; i < sys.config.MaxFallingObjects; i++ {
		_ = sys.SpawnFallingObject(float64(i), float64(i), 100, MaterialWood, 8, 8)
	}

	if sys.GetFallingObjectCount() != sys.config.MaxFallingObjects {
		t.Errorf("Falling object count = %v, want %v", sys.GetFallingObjectCount(), sys.config.MaxFallingObjects)
	}

	// Try to exceed the limit - should return error
	err = sys.SpawnFallingObject(999, 999, 100, MaterialWood, 8, 8)
	if err == nil {
		t.Error("SpawnFallingObject() should return error when limit reached")
	}

	if sys.GetFallingObjectCount() > sys.config.MaxFallingObjects {
		t.Errorf("Falling object count exceeds max: %v > %v",
			sys.GetFallingObjectCount(), sys.config.MaxFallingObjects)
	}
}

func TestSystem_FallingObjectPhysics(t *testing.T) {
	sys := NewSystem(nil)
	_ = sys.SpawnFallingObject(100, 100, 200, MaterialStone, 16, 16)

	objs := sys.GetFallingObjects()
	if len(objs) == 0 {
		t.Fatal("No falling objects spawned")
	}

	obj := objs[0]
	initialZ := obj.Z

	for i := 0; i < 100; i++ {
		sys.Update(0.016)
	}

	objs = sys.GetFallingObjects()
	if len(objs) == 0 {
		return
	}

	obj = objs[0]
	if obj.Z >= initialZ {
		t.Error("Object should fall due to gravity")
	}

	if obj.IsGrounded() && obj.Bounces == 0 {
		t.Error("Object should bounce before settling")
	}
}

func TestSystem_DebrisPhysics(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DebrisLifetime = 1.0
	sys := NewSystem(cfg)

	_ = sys.RegisterBuilding("test", 8, 8, 1, MaterialWood)
	_ = sys.ApplyDamage("test", 4, 4, 0, 1.0, 10.0)

	for i := 0; i < 100; i++ {
		sys.Update(0.1)
	}

	integrity, _ := sys.GetIntegrity("test")
	if integrity.State == IntegrityCollapsed {
		initialDebris := sys.GetDebrisCount()

		for i := 0; i < 20; i++ {
			sys.Update(0.1)
		}

		finalDebris := sys.GetDebrisCount()
		if finalDebris >= initialDebris {
			t.Error("Debris should decay over time")
		}
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("bench1", 10, 10, 1, MaterialStone)
	_ = sys.RegisterBuilding("bench2", 16, 16, 2, MaterialWood)
	_ = sys.RegisterBuilding("bench3", 24, 24, 3, MaterialBrick)

	_ = sys.ApplyDamage("bench1", 5, 5, 0, 0.3, 3.0)
	_ = sys.ApplyDamage("bench2", 8, 8, 0, 0.4, 4.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(0.016)
	}
}

func BenchmarkSystem_ApplyDamage(b *testing.B) {
	sys := NewSystem(nil)
	_ = sys.RegisterBuilding("bench", 20, 20, 2, MaterialStone)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.ApplyDamage("bench", 10, 10, 0, 0.1, 2.0)
	}
}

func BenchmarkSystem_RegisterBuilding(b *testing.B) {
	sys := NewSystem(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.RegisterBuilding("bench", 16, 16, 2, MaterialStone)
	}
}

func TestGenerateCollapseDebris_Deterministic(t *testing.T) {
	// Test that debris generation is deterministic with the same seed and building ID
	tests := []struct {
		name       string
		seed       int64
		buildingID string
	}{
		{"seed_12345", 12345, "test_building_1"},
		{"seed_99999", 99999, "test_building_2"},
		{"same_seed_diff_building", 12345, "test_building_different"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use low collapse threshold to ensure collapse triggers
			config1 := &Config{
				EnableIntegrityChecks: true,
				DamagePropagationRate: 0.5,
				CollapseThreshold:     0.6, // Higher threshold to trigger collapse sooner
				EnableDebris:          true,
				MaxDebrisParticles:    500,
				DebrisLifetime:        10.0,
				Gravity:               980.0,
				AirResistance:         0.05,
				GroundFriction:        0.8,
				MaxFallingObjects:     100,
				UpdateFrequency:       30.0,
				Seed:                  tt.seed,
			}
			config2 := &Config{
				EnableIntegrityChecks: true,
				DamagePropagationRate: 0.5,
				CollapseThreshold:     0.6,
				EnableDebris:          true,
				MaxDebrisParticles:    500,
				DebrisLifetime:        10.0,
				Gravity:               980.0,
				AirResistance:         0.05,
				GroundFriction:        0.8,
				MaxFallingObjects:     100,
				UpdateFrequency:       30.0,
				Seed:                  tt.seed,
			}

			sys1 := NewSystem(config1)
			sys2 := NewSystem(config2)

			// Register buildings with same config
			_ = sys1.RegisterBuilding(tt.buildingID, 10, 10, 2, MaterialStone)
			_ = sys2.RegisterBuilding(tt.buildingID, 10, 10, 2, MaterialStone)

			// Apply heavy damage to all floors to trigger collapse
			// Apply damage multiple times across the building to destroy supports
			for floor := 0; floor < 2; floor++ {
				for x := 0; x < 10; x += 2 {
					_ = sys1.ApplyDamage(tt.buildingID, x, 5, floor, 1.0, 3.0)
					_ = sys2.ApplyDamage(tt.buildingID, x, 5, floor, 1.0, 3.0)
				}
			}

			// Run updates to trigger collapse and debris generation
			for i := 0; i < 100; i++ {
				sys1.Update(0.1)
				sys2.Update(0.1)
			}

			debris1 := sys1.GetDebris()
			debris2 := sys2.GetDebris()

			if len(debris1) != len(debris2) {
				t.Fatalf("Debris count mismatch: %d vs %d", len(debris1), len(debris2))
			}

			if len(debris1) == 0 {
				// Get integrity to check state
				integrity1, _ := sys1.GetIntegrity(tt.buildingID)
				t.Skipf("No debris generated (state: %v, health: %.2f)", integrity1.State, integrity1.CurrentHealth)
			}

			// Compare first few debris properties (debris may have moved due to physics)
			// Only compare initial properties that should be deterministic
			for i := 0; i < len(debris1) && i < 5; i++ {
				if debris1[i].Material != debris2[i].Material {
					t.Errorf("Debris[%d] Material mismatch: %v vs %v", i, debris1[i].Material, debris2[i].Material)
				}
				if debris1[i].Size != debris2[i].Size {
					t.Errorf("Debris[%d] Size mismatch: %v vs %v", i, debris1[i].Size, debris2[i].Size)
				}
			}
		})
	}
}

func TestHashBuildingID(t *testing.T) {
	// Test that hash is deterministic
	tests := []struct {
		name string
		id   string
	}{
		{"simple", "building1"},
		{"complex", "guild_hall_12345"},
		{"empty", ""},
		{"unicode", "建物"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashBuildingID(tt.id)
			hash2 := hashBuildingID(tt.id)
			if hash1 != hash2 {
				t.Errorf("Hash not deterministic: %d vs %d", hash1, hash2)
			}
		})
	}

	// Test that different IDs produce different hashes
	hash1 := hashBuildingID("building1")
	hash2 := hashBuildingID("building2")
	if hash1 == hash2 {
		t.Error("Different IDs should produce different hashes")
	}
}
