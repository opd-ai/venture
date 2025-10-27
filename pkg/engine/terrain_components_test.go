package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/world"
)

// TestMaterialType_String tests String() method for all material types.
func TestMaterialType_String(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		want     string
	}{
		{"stone", MaterialStone, "stone"},
		{"wood", MaterialWood, "wood"},
		{"earth", MaterialEarth, "earth"},
		{"metal", MaterialMetal, "metal"},
		{"glass", MaterialGlass, "glass"},
		{"ice", MaterialIce, "ice"},
		{"unknown", MaterialType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.material.String()
			if got != tt.want {
				t.Errorf("MaterialType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMaterialType_IsFlammable tests flammability check for materials.
func TestMaterialType_IsFlammable(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		want     bool
	}{
		{"stone not flammable", MaterialStone, false},
		{"wood flammable", MaterialWood, true},
		{"earth not flammable", MaterialEarth, false},
		{"metal not flammable", MaterialMetal, false},
		{"glass not flammable", MaterialGlass, false},
		{"ice not flammable", MaterialIce, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.material.IsFlammable()
			if got != tt.want {
				t.Errorf("MaterialType.IsFlammable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMaterialType_BaseDurability tests durability values for materials.
func TestMaterialType_BaseDurability(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		want     float64
	}{
		{"stone durability", MaterialStone, 100.0},
		{"wood durability", MaterialWood, 50.0},
		{"earth durability", MaterialEarth, 30.0},
		{"metal durability", MaterialMetal, 200.0},
		{"glass durability", MaterialGlass, 20.0},
		{"ice durability", MaterialIce, 40.0},
		{"unknown defaults to 50", MaterialType(999), 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.material.BaseDurability()
			if got != tt.want {
				t.Errorf("MaterialType.BaseDurability() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewDestructibleComponent tests component creation with defaults.
func TestNewDestructibleComponent(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		tileX    int
		tileY    int
	}{
		{"stone tile", MaterialStone, 10, 20},
		{"wood tile", MaterialWood, 5, 15},
		{"metal tile", MaterialMetal, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewDestructibleComponent(tt.material, tt.tileX, tt.tileY)

			if comp.Material != tt.material {
				t.Errorf("Material = %v, want %v", comp.Material, tt.material)
			}
			if comp.TileX != tt.tileX {
				t.Errorf("TileX = %v, want %v", comp.TileX, tt.tileX)
			}
			if comp.TileY != tt.tileY {
				t.Errorf("TileY = %v, want %v", comp.TileY, tt.tileY)
			}

			expectedMaxHealth := tt.material.BaseDurability()
			if comp.MaxHealth != expectedMaxHealth {
				t.Errorf("MaxHealth = %v, want %v", comp.MaxHealth, expectedMaxHealth)
			}
			if comp.Health != expectedMaxHealth {
				t.Errorf("Health = %v, want %v (should start at max)", comp.Health, expectedMaxHealth)
			}
			if comp.IsDestroyed {
				t.Error("IsDestroyed should be false initially")
			}
		})
	}
}

// TestDestructibleComponent_Type tests Type() method.
func TestDestructibleComponent_Type(t *testing.T) {
	comp := NewDestructibleComponent(MaterialStone, 0, 0)
	want := "destructible"
	got := comp.Type()
	if got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
}

// TestDestructibleComponent_TakeDamage tests damage application and destruction.
func TestDestructibleComponent_TakeDamage(t *testing.T) {
	tests := []struct {
		name          string
		material      MaterialType
		damage        float64
		wantHealth    float64
		wantDestroyed bool
	}{
		{"partial damage to wood", MaterialWood, 20.0, 30.0, false},
		{"destroy wood completely", MaterialWood, 50.0, 0.0, true},
		{"overkill damage", MaterialWood, 100.0, 0.0, true},
		{"small damage to metal", MaterialMetal, 10.0, 190.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewDestructibleComponent(tt.material, 0, 0)
			destroyed := comp.TakeDamage(tt.damage)

			if comp.Health != tt.wantHealth {
				t.Errorf("Health = %v, want %v", comp.Health, tt.wantHealth)
			}
			if destroyed != tt.wantDestroyed {
				t.Errorf("TakeDamage() returned %v, want %v", destroyed, tt.wantDestroyed)
			}
			if comp.IsDestroyed != tt.wantDestroyed {
				t.Errorf("IsDestroyed = %v, want %v", comp.IsDestroyed, tt.wantDestroyed)
			}
		})
	}
}

// TestDestructibleComponent_HealthPercent tests health percentage calculation.
func TestDestructibleComponent_HealthPercent(t *testing.T) {
	tests := []struct {
		name        string
		damage      float64
		wantPercent float64
	}{
		{"full health", 0, 1.0},
		{"half health", 25.0, 0.5},
		{"quarter health", 37.5, 0.25},
		{"zero health", 50.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewDestructibleComponent(MaterialWood, 0, 0)
			if tt.damage > 0 {
				comp.TakeDamage(tt.damage)
			}

			got := comp.HealthPercent()
			if got != tt.wantPercent {
				t.Errorf("HealthPercent() = %v, want %v", got, tt.wantPercent)
			}
		})
	}
}

// TestDestructibleComponent_HealthPercentZeroMax tests division by zero protection.
func TestDestructibleComponent_HealthPercentZeroMax(t *testing.T) {
	comp := NewDestructibleComponent(MaterialWood, 0, 0)
	comp.MaxHealth = 0 // Invalid state

	got := comp.HealthPercent()
	if got != 0.0 {
		t.Errorf("HealthPercent() with MaxHealth=0 = %v, want 0.0", got)
	}
}

// TestNewFireComponent tests fire component creation with validation.
func TestNewFireComponent(t *testing.T) {
	tests := []struct {
		name          string
		intensity     float64
		tileX         int
		tileY         int
		maxDuration   float64
		wantIntensity float64
		wantDuration  float64
	}{
		{"normal fire", 0.8, 10, 20, 15.0, 0.8, 15.0},
		{"low intensity", 0.2, 5, 5, 10.0, 0.2, 10.0},
		{"clamped high intensity", 1.5, 0, 0, 12.0, 1.0, 12.0},
		{"clamped negative intensity", -0.5, 0, 0, 12.0, 0.0, 12.0},
		{"default duration", 0.5, 3, 7, 0, 0.5, 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFireComponent(tt.intensity, tt.tileX, tt.tileY, tt.maxDuration)

			if comp.Intensity != tt.wantIntensity {
				t.Errorf("Intensity = %v, want %v", comp.Intensity, tt.wantIntensity)
			}
			if comp.TileX != tt.tileX {
				t.Errorf("TileX = %v, want %v", comp.TileX, tt.tileX)
			}
			if comp.TileY != tt.tileY {
				t.Errorf("TileY = %v, want %v", comp.TileY, tt.tileY)
			}
			if comp.MaxDuration != tt.wantDuration {
				t.Errorf("MaxDuration = %v, want %v", comp.MaxDuration, tt.wantDuration)
			}
			if comp.Duration != 0 {
				t.Errorf("Duration = %v, want 0 (should start at 0)", comp.Duration)
			}
			if comp.IsExtinguished {
				t.Error("IsExtinguished should be false initially")
			}

			// Verify derived values
			expectedSpread := 0.3 * tt.wantIntensity
			if comp.SpreadChance != expectedSpread {
				t.Errorf("SpreadChance = %v, want %v", comp.SpreadChance, expectedSpread)
			}
			expectedDamage := 5.0 * tt.wantIntensity
			if comp.DamagePerSecond != expectedDamage {
				t.Errorf("DamagePerSecond = %v, want %v", comp.DamagePerSecond, expectedDamage)
			}
		})
	}
}

// TestFireComponent_Type tests Type() method.
func TestFireComponent_Type(t *testing.T) {
	comp := NewFireComponent(0.5, 0, 0, 10.0)
	want := "fire"
	got := comp.Type()
	if got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
}

// TestFireComponent_Update tests fire duration and extinguishing.
func TestFireComponent_Update(t *testing.T) {
	tests := []struct {
		name             string
		maxDuration      float64
		updates          []float64 // sequence of deltaTime values
		wantDuration     float64
		wantExtinguished bool
	}{
		{"short burn", 10.0, []float64{3.0, 3.0}, 6.0, false},
		{"exactly max duration", 10.0, []float64{5.0, 5.0}, 10.0, true},
		{"exceeds max duration", 10.0, []float64{6.0, 6.0}, 12.0, true},
		{"multiple small updates", 5.0, []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}, 6.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFireComponent(0.5, 0, 0, tt.maxDuration)

			for _, dt := range tt.updates {
				comp.Update(dt)
			}

			if comp.Duration != tt.wantDuration {
				t.Errorf("Duration = %v, want %v", comp.Duration, tt.wantDuration)
			}
			if comp.IsExtinguished != tt.wantExtinguished {
				t.Errorf("IsExtinguished = %v, want %v", comp.IsExtinguished, tt.wantExtinguished)
			}
		})
	}
}

// TestFireComponent_RemainingTime tests remaining time calculation.
func TestFireComponent_RemainingTime(t *testing.T) {
	tests := []struct {
		name     string
		elapsed  float64
		wantTime float64
	}{
		{"just started", 0, 10.0},
		{"half burned", 5.0, 5.0},
		{"almost done", 9.5, 0.5},
		{"extinguished", 10.0, 0.0},
		{"over time", 12.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewFireComponent(0.5, 0, 0, 10.0)
			comp.Duration = tt.elapsed

			got := comp.RemainingTime()
			if got != tt.wantTime {
				t.Errorf("RemainingTime() = %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// TestNewBuildableComponent tests buildable component creation.
func TestNewBuildableComponent(t *testing.T) {
	tests := []struct {
		name             string
		tileX            int
		tileY            int
		resultType       world.TileType
		constructionTime float64
		wantTime         float64
	}{
		{"wall construction", 10, 20, world.TileWall, 5.0, 5.0},
		{"door construction", 5, 15, world.TileDoor, 3.0, 3.0},
		{"default time", 0, 0, world.TileWall, 0, 3.0},
		{"negative time defaults", 3, 7, world.TileWall, -1.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewBuildableComponent(tt.tileX, tt.tileY, tt.resultType, tt.constructionTime)

			if comp.TileX != tt.tileX {
				t.Errorf("TileX = %v, want %v", comp.TileX, tt.tileX)
			}
			if comp.TileY != tt.tileY {
				t.Errorf("TileY = %v, want %v", comp.TileY, tt.tileY)
			}
			if comp.ResultTileType != tt.resultType {
				t.Errorf("ResultTileType = %v, want %v", comp.ResultTileType, tt.resultType)
			}
			if comp.ConstructionTime != tt.wantTime {
				t.Errorf("ConstructionTime = %v, want %v", comp.ConstructionTime, tt.wantTime)
			}
			if comp.ElapsedTime != 0 {
				t.Errorf("ElapsedTime = %v, want 0", comp.ElapsedTime)
			}
			if comp.IsComplete {
				t.Error("IsComplete should be false initially")
			}

			// Verify default materials (10 stone)
			if stoneQty, ok := comp.RequiredMaterials[MaterialStone]; !ok || stoneQty != 10 {
				t.Errorf("RequiredMaterials[MaterialStone] = %v, want 10", stoneQty)
			}
		})
	}
}

// TestBuildableComponent_Type tests Type() method.
func TestBuildableComponent_Type(t *testing.T) {
	comp := NewBuildableComponent(0, 0, world.TileWall, 3.0)
	want := "buildable"
	got := comp.Type()
	if got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
}

// TestBuildableComponent_Update tests construction progress.
func TestBuildableComponent_Update(t *testing.T) {
	tests := []struct {
		name         string
		buildTime    float64
		updates      []float64 // sequence of deltaTime values
		wantElapsed  float64
		wantComplete bool
	}{
		{"partial progress", 5.0, []float64{2.0}, 2.0, false},
		{"complete build", 3.0, []float64{1.5, 1.5}, 3.0, true},
		{"exceed time", 3.0, []float64{5.0}, 5.0, true},
		{"multiple updates", 4.0, []float64{1.0, 1.0, 1.0, 1.0}, 4.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewBuildableComponent(0, 0, world.TileWall, tt.buildTime)

			for _, dt := range tt.updates {
				comp.Update(dt)
			}

			if comp.ElapsedTime != tt.wantElapsed {
				t.Errorf("ElapsedTime = %v, want %v", comp.ElapsedTime, tt.wantElapsed)
			}
			if comp.IsComplete != tt.wantComplete {
				t.Errorf("IsComplete = %v, want %v", comp.IsComplete, tt.wantComplete)
			}
		})
	}
}

// TestBuildableComponent_Progress tests progress percentage calculation.
func TestBuildableComponent_Progress(t *testing.T) {
	tests := []struct {
		name        string
		buildTime   float64
		elapsed     float64
		wantPercent float64
	}{
		{"not started", 5.0, 0, 0.0},
		{"half done", 4.0, 2.0, 0.5},
		{"three quarters", 8.0, 6.0, 0.75},
		{"complete", 3.0, 3.0, 1.0},
		{"over time clamped", 2.0, 5.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewBuildableComponent(0, 0, world.TileWall, tt.buildTime)
			comp.ElapsedTime = tt.elapsed

			got := comp.Progress()
			if got != tt.wantPercent {
				t.Errorf("Progress() = %v, want %v", got, tt.wantPercent)
			}
		})
	}
}

// TestBuildableComponent_ProgressZeroTime tests division by zero protection.
func TestBuildableComponent_ProgressZeroTime(t *testing.T) {
	comp := NewBuildableComponent(0, 0, world.TileWall, 3.0)
	comp.ConstructionTime = 0 // Invalid state

	got := comp.Progress()
	if got != 1.0 {
		t.Errorf("Progress() with ConstructionTime=0 = %v, want 1.0", got)
	}
}

// TestBuildableComponent_UpdateAfterComplete tests that updates after completion do nothing.
func TestBuildableComponent_UpdateAfterComplete(t *testing.T) {
	comp := NewBuildableComponent(0, 0, world.TileWall, 3.0)
	comp.IsComplete = true
	comp.ElapsedTime = 3.0

	comp.Update(5.0) // Should not change elapsed time

	if comp.ElapsedTime != 3.0 {
		t.Errorf("ElapsedTime = %v, want 3.0 (should not change after complete)", comp.ElapsedTime)
	}
}

// TestDestructibleComponent_LastDamageTime tests that damage updates time.
func TestDestructibleComponent_LastDamageTime(t *testing.T) {
	comp := NewDestructibleComponent(MaterialWood, 0, 0)
	initialTime := comp.LastDamageTime

	time.Sleep(10 * time.Millisecond) // Small delay

	comp.TakeDamage(10.0)

	if !comp.LastDamageTime.After(initialTime) {
		t.Error("LastDamageTime should be updated after TakeDamage()")
	}
}

// BenchmarkDestructibleComponent_TakeDamage benchmarks damage application.
func BenchmarkDestructibleComponent_TakeDamage(b *testing.B) {
	comp := NewDestructibleComponent(MaterialStone, 0, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.Health = 100.0 // Reset health
		comp.TakeDamage(10.0)
	}
}

// BenchmarkFireComponent_Update benchmarks fire updates.
func BenchmarkFireComponent_Update(b *testing.B) {
	comp := NewFireComponent(0.8, 0, 0, 15.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.Duration = 0  // Reset
		comp.Update(0.016) // ~60 FPS frame time
	}
}

// BenchmarkBuildableComponent_Update benchmarks construction updates.
func BenchmarkBuildableComponent_Update(b *testing.B) {
	comp := NewBuildableComponent(0, 0, world.TileWall, 3.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comp.ElapsedTime = 0 // Reset
		comp.IsComplete = false
		comp.Update(0.016) // ~60 FPS frame time
	}
}
