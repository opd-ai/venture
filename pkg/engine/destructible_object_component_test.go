package engine

import (
	"testing"
)

func TestObjectType_String(t *testing.T) {
	tests := []struct {
		name     string
		objType  ObjectType
		expected string
	}{
		{"crate", ObjectCrate, "crate"},
		{"barrel", ObjectBarrel, "barrel"},
		{"furniture", ObjectFurniture, "furniture"},
		{"weak_wall", ObjectWeakWall, "weak_wall"},
		{"poison_container", ObjectPoisonContainer, "poison_container"},
		{"explosive_barrel", ObjectExplosiveBarrel, "explosive_barrel"},
		{"unknown", ObjectType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.objType.String()
			if got != tt.expected {
				t.Errorf("ObjectType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewDestructibleObjectComponent(t *testing.T) {
	tests := []struct {
		name                string
		objType             ObjectType
		wantHealth          float64
		wantMaxHealth       float64
		wantExplosionRadius float64
		wantExplosionDamage float64
		wantPoisonDuration  float64
		wantDebrisCount     int
		wantLootTable       string
	}{
		{
			name:            "crate",
			objType:         ObjectCrate,
			wantHealth:      20.0,
			wantMaxHealth:   20.0,
			wantDebrisCount: 3,
			wantLootTable:   "common_items",
		},
		{
			name:            "barrel",
			objType:         ObjectBarrel,
			wantHealth:      30.0,
			wantMaxHealth:   30.0,
			wantDebrisCount: 3,
		},
		{
			name:            "furniture",
			objType:         ObjectFurniture,
			wantHealth:      15.0,
			wantMaxHealth:   15.0,
			wantDebrisCount: 3,
		},
		{
			name:            "weak_wall",
			objType:         ObjectWeakWall,
			wantHealth:      40.0,
			wantMaxHealth:   40.0,
			wantDebrisCount: 5,
		},
		{
			name:               "poison_container",
			objType:            ObjectPoisonContainer,
			wantHealth:         15.0,
			wantMaxHealth:      15.0,
			wantPoisonDuration: 10.0,
			wantDebrisCount:    3,
		},
		{
			name:                "explosive_barrel",
			objType:             ObjectExplosiveBarrel,
			wantHealth:          25.0,
			wantMaxHealth:       25.0,
			wantExplosionRadius: 96.0,
			wantExplosionDamage: 50.0,
			wantDebrisCount:     8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewDestructibleObjectComponent(tt.objType)

			if comp.ObjectType != tt.objType {
				t.Errorf("ObjectType = %v, want %v", comp.ObjectType, tt.objType)
			}
			if comp.Health != tt.wantHealth {
				t.Errorf("Health = %v, want %v", comp.Health, tt.wantHealth)
			}
			if comp.MaxHealth != tt.wantMaxHealth {
				t.Errorf("MaxHealth = %v, want %v", comp.MaxHealth, tt.wantMaxHealth)
			}
			if comp.ExplosionRadius != tt.wantExplosionRadius {
				t.Errorf("ExplosionRadius = %v, want %v", comp.ExplosionRadius, tt.wantExplosionRadius)
			}
			if comp.ExplosionDamage != tt.wantExplosionDamage {
				t.Errorf("ExplosionDamage = %v, want %v", comp.ExplosionDamage, tt.wantExplosionDamage)
			}
			if comp.PoisonDuration != tt.wantPoisonDuration {
				t.Errorf("PoisonDuration = %v, want %v", comp.PoisonDuration, tt.wantPoisonDuration)
			}
			if comp.DebrisCount != tt.wantDebrisCount {
				t.Errorf("DebrisCount = %v, want %v", comp.DebrisCount, tt.wantDebrisCount)
			}
			if comp.LootTable != tt.wantLootTable {
				t.Errorf("LootTable = %v, want %v", comp.LootTable, tt.wantLootTable)
			}
			if comp.IsDestroyed {
				t.Error("IsDestroyed should be false initially")
			}
		})
	}
}

func TestDestructibleObjectComponent_Type(t *testing.T) {
	comp := NewDestructibleObjectComponent(ObjectCrate)
	if comp.Type() != "destructibleObject" {
		t.Errorf("Type() = %v, want 'destructibleObject'", comp.Type())
	}
}

func TestDestructibleObjectComponent_TakeDamage(t *testing.T) {
	tests := []struct {
		name          string
		initialHealth float64
		damage        float64
		wantDestroyed bool
		wantHealth    float64
	}{
		{"partial_damage", 30.0, 10.0, false, 20.0},
		{"exact_destroy", 30.0, 30.0, true, 0.0},
		{"overkill_damage", 30.0, 50.0, true, 0.0},
		{"no_damage", 30.0, 0.0, false, 30.0},
		{"small_damage", 30.0, 1.0, false, 29.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DestructibleObjectComponent{
				Health:    tt.initialHealth,
				MaxHealth: 30.0,
			}

			destroyed := comp.TakeDamage(tt.damage)

			if destroyed != tt.wantDestroyed {
				t.Errorf("TakeDamage() returned %v, want %v", destroyed, tt.wantDestroyed)
			}
			if comp.Health != tt.wantHealth {
				t.Errorf("Health = %v, want %v", comp.Health, tt.wantHealth)
			}
			if comp.IsDestroyed != tt.wantDestroyed {
				t.Errorf("IsDestroyed = %v, want %v", comp.IsDestroyed, tt.wantDestroyed)
			}

			// Note: LastDamageTime is no longer updated by the deprecated TakeDamage method.
			// Use DestructibleObjectSystem.ApplyDamageToComponent() for proper time tracking.
		})
	}
}

func TestDestructibleObjectComponent_HealthPercent(t *testing.T) {
	tests := []struct {
		name        string
		health      float64
		maxHealth   float64
		wantPercent float64
	}{
		{"full_health", 30.0, 30.0, 1.0},
		{"half_health", 15.0, 30.0, 0.5},
		{"low_health", 3.0, 30.0, 0.1},
		{"zero_health", 0.0, 30.0, 0.0},
		{"zero_max_health", 10.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DestructibleObjectComponent{
				Health:    tt.health,
				MaxHealth: tt.maxHealth,
			}

			got := comp.HealthPercent()
			if got != tt.wantPercent {
				t.Errorf("HealthPercent() = %v, want %v", got, tt.wantPercent)
			}
		})
	}
}

func TestDestructibleObjectComponent_IsExplosive(t *testing.T) {
	tests := []struct {
		name            string
		explosionRadius float64
		want            bool
	}{
		{"explosive", 96.0, true},
		{"not_explosive", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DestructibleObjectComponent{
				ExplosionRadius: tt.explosionRadius,
			}

			if got := comp.IsExplosive(); got != tt.want {
				t.Errorf("IsExplosive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDestructibleObjectComponent_EmitsPoison(t *testing.T) {
	tests := []struct {
		name           string
		poisonDuration float64
		want           bool
	}{
		{"emits_poison", 10.0, true},
		{"no_poison", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DestructibleObjectComponent{
				PoisonDuration: tt.poisonDuration,
			}

			if got := comp.EmitsPoison(); got != tt.want {
				t.Errorf("EmitsPoison() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDebrisComponent(t *testing.T) {
	tests := []struct {
		name         string
		sourceType   ObjectType
		lifetime     float64
		angularVel   float64
		wantLifetime float64
	}{
		{"normal_debris", ObjectCrate, 5.0, 1.5, 5.0},
		{"zero_lifetime_default", ObjectBarrel, 0.0, 1.0, 5.0},
		{"negative_lifetime_default", ObjectFurniture, -1.0, 0.5, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewDebrisComponent(tt.sourceType, tt.lifetime, tt.angularVel)

			if comp.SourceObjectType != tt.sourceType {
				t.Errorf("SourceObjectType = %v, want %v", comp.SourceObjectType, tt.sourceType)
			}
			if comp.Lifetime != tt.wantLifetime {
				t.Errorf("Lifetime = %v, want %v", comp.Lifetime, tt.wantLifetime)
			}
			if comp.MaxLifetime != tt.wantLifetime {
				t.Errorf("MaxLifetime = %v, want %v", comp.MaxLifetime, tt.wantLifetime)
			}
			if comp.AngularVelocity != tt.angularVel {
				t.Errorf("AngularVelocity = %v, want %v", comp.AngularVelocity, tt.angularVel)
			}
			if comp.IsStationary {
				t.Error("IsStationary should be false initially")
			}
		})
	}
}

func TestDebrisComponent_Type(t *testing.T) {
	comp := NewDebrisComponent(ObjectCrate, 5.0, 1.0)
	if comp.Type() != "debris" {
		t.Errorf("Type() = %v, want 'debris'", comp.Type())
	}
}

func TestDebrisComponent_Update(t *testing.T) {
	comp := NewDebrisComponent(ObjectCrate, 5.0, 1.0)

	comp.Update(1.0)
	if comp.Lifetime != 4.0 {
		t.Errorf("After 1s, Lifetime = %v, want 4.0", comp.Lifetime)
	}

	comp.Update(3.5)
	if comp.Lifetime != 0.5 {
		t.Errorf("After 4.5s total, Lifetime = %v, want 0.5", comp.Lifetime)
	}

	comp.Update(1.0)
	if comp.Lifetime != -0.5 {
		t.Errorf("After 5.5s total, Lifetime = %v, want -0.5", comp.Lifetime)
	}
}

func TestDebrisComponent_ShouldDespawn(t *testing.T) {
	tests := []struct {
		name     string
		lifetime float64
		want     bool
	}{
		{"still_alive", 2.0, false},
		{"just_expired", 0.0, false},
		{"expired", -0.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DebrisComponent{Lifetime: tt.lifetime}
			if got := comp.ShouldDespawn(); got != tt.want {
				t.Errorf("ShouldDespawn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDebrisComponent_RemainingLifetimePercent(t *testing.T) {
	tests := []struct {
		name        string
		lifetime    float64
		maxLifetime float64
		want        float64
	}{
		{"full_lifetime", 5.0, 5.0, 1.0},
		{"half_lifetime", 2.5, 5.0, 0.5},
		{"expired", 0.0, 5.0, 0.0},
		{"negative_expired", -1.0, 5.0, 0.0},
		{"over_max", 6.0, 5.0, 1.0},
		{"zero_max", 3.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DebrisComponent{
				Lifetime:    tt.lifetime,
				MaxLifetime: tt.maxLifetime,
			}
			got := comp.RemainingLifetimePercent()
			if got != tt.want {
				t.Errorf("RemainingLifetimePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDestructibleObjectSystem_ApplyDamageToComponent tests the ECS-compliant damage method.
func TestDestructibleObjectSystem_ApplyDamageToComponent(t *testing.T) {
	tests := []struct {
		name          string
		initialHealth float64
		damage        float64
		gameTime      float64
		wantDestroyed bool
		wantHealth    float64
	}{
		{"partial_damage", 30.0, 10.0, 5.0, false, 20.0},
		{"exact_destroy", 30.0, 30.0, 10.0, true, 0.0},
		{"overkill_damage", 30.0, 50.0, 15.0, true, 0.0},
		{"no_damage", 30.0, 0.0, 1.0, false, 30.0},
		{"small_damage", 30.0, 1.0, 2.5, false, 29.0},
	}

	sys := NewDestructibleObjectSystem(32, 12345)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &DestructibleObjectComponent{
				Health:    tt.initialHealth,
				MaxHealth: 30.0,
			}

			destroyed := sys.ApplyDamageToComponent(comp, tt.damage, tt.gameTime)

			if destroyed != tt.wantDestroyed {
				t.Errorf("ApplyDamageToComponent() returned %v, want %v", destroyed, tt.wantDestroyed)
			}
			if comp.Health != tt.wantHealth {
				t.Errorf("Health = %v, want %v", comp.Health, tt.wantHealth)
			}
			if comp.IsDestroyed != tt.wantDestroyed {
				t.Errorf("IsDestroyed = %v, want %v", comp.IsDestroyed, tt.wantDestroyed)
			}
			if comp.LastDamageTime != tt.gameTime {
				t.Errorf("LastDamageTime = %v, want %v", comp.LastDamageTime, tt.gameTime)
			}
		})
	}
}

// TestDestructibleObjectSystem_ApplyDamageToComponent_NilAndDestroyed tests edge cases.
func TestDestructibleObjectSystem_ApplyDamageToComponent_NilAndDestroyed(t *testing.T) {
	sys := NewDestructibleObjectSystem(32, 12345)

	// Test nil component
	if sys.ApplyDamageToComponent(nil, 10.0, 1.0) {
		t.Error("ApplyDamageToComponent(nil) should return false")
	}

	// Test already destroyed component
	comp := &DestructibleObjectComponent{
		Health:      0,
		MaxHealth:   30.0,
		IsDestroyed: true,
	}
	if sys.ApplyDamageToComponent(comp, 10.0, 1.0) {
		t.Error("ApplyDamageToComponent(destroyed) should return false")
	}
}

// TestDestructibleObjectSystem_GameTime tests game time accumulation.
func TestDestructibleObjectSystem_GameTime(t *testing.T) {
	sys := NewDestructibleObjectSystem(32, 12345)

	// Initial game time should be 0
	if sys.GetGameTime() != 0 {
		t.Errorf("Initial GetGameTime() = %v, want 0", sys.GetGameTime())
	}

	// Update should accumulate game time
	sys.Update([]*Entity{}, 1.5)
	if sys.GetGameTime() != 1.5 {
		t.Errorf("After 1.5s, GetGameTime() = %v, want 1.5", sys.GetGameTime())
	}

	sys.Update([]*Entity{}, 2.0)
	if sys.GetGameTime() != 3.5 {
		t.Errorf("After 3.5s total, GetGameTime() = %v, want 3.5", sys.GetGameTime())
	}
}
