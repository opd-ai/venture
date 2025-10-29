// Package engine provides tests for aiming functionality.
package engine

import (
	"math"
	"testing"
)

// TestAimComponent_Type verifies component type identifier
func TestAimComponent_Type(t *testing.T) {
	comp := NewAimComponent(0)
	if got := comp.Type(); got != "aim" {
		t.Errorf("Type() = %q, want %q", got, "aim")
	}
}

// TestNewAimComponent tests component creation with defaults
func TestNewAimComponent(t *testing.T) {
	tests := []struct {
		name         string
		initialAngle float64
		wantAngle    float64
	}{
		{"zero angle", 0, 0},
		{"quarter turn", math.Pi / 2, math.Pi / 2},
		{"negative angle", -math.Pi / 4, 7 * math.Pi / 4},
		{"large angle", 3 * math.Pi, math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(tt.initialAngle)

			if !floatEqual(comp.AimAngle, tt.wantAngle, 0.0001) {
				t.Errorf("AimAngle = %v, want %v", comp.AimAngle, tt.wantAngle)
			}
			if comp.HasTarget {
				t.Error("HasTarget should default to false")
			}
			if comp.AutoAim {
				t.Error("AutoAim should default to false")
			}
			if comp.SnapRadius != 100.0 {
				t.Errorf("SnapRadius = %v, want 100.0", comp.SnapRadius)
			}
			if comp.AutoAimStrength != 0.3 {
				t.Errorf("AutoAimStrength = %v, want 0.3", comp.AutoAimStrength)
			}
		})
	}
}

// TestAimComponent_SetAimAngle tests direct angle setting
func TestAimComponent_SetAimAngle(t *testing.T) {
	comp := NewAimComponent(0)
	comp.HasTarget = true // Should be cleared

	comp.SetAimAngle(math.Pi / 2)

	if !floatEqual(comp.AimAngle, math.Pi/2, 0.0001) {
		t.Errorf("AimAngle = %v, want %v", comp.AimAngle, math.Pi/2)
	}
	if comp.HasTarget {
		t.Error("HasTarget should be false after SetAimAngle")
	}
}

// TestAimComponent_SetAimTarget tests target-based aiming
func TestAimComponent_SetAimTarget(t *testing.T) {
	comp := NewAimComponent(0)

	comp.SetAimTarget(100, 200)

	if comp.AimTarget.X != 100 {
		t.Errorf("AimTarget.X = %v, want 100", comp.AimTarget.X)
	}
	if comp.AimTarget.Y != 200 {
		t.Errorf("AimTarget.Y = %v, want 200", comp.AimTarget.Y)
	}
	if !comp.HasTarget {
		t.Error("HasTarget should be true after SetAimTarget")
	}
}

// TestAimComponent_ClearAimTarget tests target clearing
func TestAimComponent_ClearAimTarget(t *testing.T) {
	comp := NewAimComponent(0)
	comp.SetAimTarget(100, 200)

	comp.ClearAimTarget()

	if comp.HasTarget {
		t.Error("HasTarget should be false after ClearAimTarget")
	}
}

// TestAimComponent_UpdateAimAngle tests angle calculation from target
func TestAimComponent_UpdateAimAngle(t *testing.T) {
	tests := []struct {
		name      string
		entityX   float64
		entityY   float64
		targetX   float64
		targetY   float64
		wantAngle float64
	}{
		{
			name:      "aim right",
			entityX:   0,
			entityY:   0,
			targetX:   100,
			targetY:   0,
			wantAngle: 0,
		},
		{
			name:      "aim down",
			entityX:   0,
			entityY:   0,
			targetX:   0,
			targetY:   100,
			wantAngle: math.Pi / 2,
		},
		{
			name:      "aim left",
			entityX:   0,
			entityY:   0,
			targetX:   -100,
			targetY:   0,
			wantAngle: math.Pi,
		},
		{
			name:      "aim up",
			entityX:   0,
			entityY:   0,
			targetX:   0,
			targetY:   -100,
			wantAngle: 3 * math.Pi / 2,
		},
		{
			name:      "aim diagonal",
			entityX:   100,
			entityY:   100,
			targetX:   200,
			targetY:   200,
			wantAngle: math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(0)
			comp.SetAimTarget(tt.targetX, tt.targetY)

			angle := comp.UpdateAimAngle(tt.entityX, tt.entityY)

			if !floatEqual(angle, tt.wantAngle, 0.01) {
				t.Errorf("UpdateAimAngle() = %v, want %v", angle, tt.wantAngle)
			}
			if !floatEqual(comp.AimAngle, tt.wantAngle, 0.01) {
				t.Errorf("AimAngle = %v, want %v", comp.AimAngle, tt.wantAngle)
			}
		})
	}
}

// TestAimComponent_UpdateAimAngleNoTarget tests behavior without target
func TestAimComponent_UpdateAimAngleNoTarget(t *testing.T) {
	comp := NewAimComponent(math.Pi / 2)

	angle := comp.UpdateAimAngle(0, 0)

	if !floatEqual(angle, math.Pi/2, 0.0001) {
		t.Errorf("UpdateAimAngle() = %v, want %v (should keep current angle)", angle, math.Pi/2)
	}
}

// TestAimComponent_GetAimDirection tests direction vector calculation
func TestAimComponent_GetAimDirection(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		wantX float64
		wantY float64
	}{
		{"right", 0, 1.0, 0.0},
		{"down", math.Pi / 2, 0.0, 1.0},
		{"left", math.Pi, -1.0, 0.0},
		{"up", 3 * math.Pi / 2, 0.0, -1.0},
		{"down-right", math.Pi / 4, 0.707, 0.707},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(tt.angle)
			x, y := comp.GetAimDirection()

			if !floatEqual(x, tt.wantX, 0.01) {
				t.Errorf("x = %v, want %v", x, tt.wantX)
			}
			if !floatEqual(y, tt.wantY, 0.01) {
				t.Errorf("y = %v, want %v", y, tt.wantY)
			}
		})
	}
}

// TestAimComponent_GetAttackOrigin tests attack origin calculation
func TestAimComponent_GetAttackOrigin(t *testing.T) {
	tests := []struct {
		name         string
		angle        float64
		entityX      float64
		entityY      float64
		weaponOffset float64
		wantX        float64
		wantY        float64
	}{
		{
			name:         "aim right",
			angle:        0,
			entityX:      100,
			entityY:      100,
			weaponOffset: 20,
			wantX:        120,
			wantY:        100,
		},
		{
			name:         "aim down",
			angle:        math.Pi / 2,
			entityX:      100,
			entityY:      100,
			weaponOffset: 20,
			wantX:        100,
			wantY:        120,
		},
		{
			name:         "aim left",
			angle:        math.Pi,
			entityX:      100,
			entityY:      100,
			weaponOffset: 20,
			wantX:        80,
			wantY:        100,
		},
		{
			name:         "aim diagonal",
			angle:        math.Pi / 4,
			entityX:      0,
			entityY:      0,
			weaponOffset: 10,
			wantX:        7.07,
			wantY:        7.07,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(tt.angle)
			x, y := comp.GetAttackOrigin(tt.entityX, tt.entityY, tt.weaponOffset)

			if !floatEqual(x, tt.wantX, 0.1) {
				t.Errorf("x = %v, want %v", x, tt.wantX)
			}
			if !floatEqual(y, tt.wantY, 0.1) {
				t.Errorf("y = %v, want %v", y, tt.wantY)
			}
		})
	}
}

// TestAimComponent_ApplyAutoAim tests aim assist functionality
func TestAimComponent_ApplyAutoAim(t *testing.T) {
	tests := []struct {
		name         string
		autoAim      bool
		strength     float64
		entityX      float64
		entityY      float64
		enemyX       float64
		enemyY       float64
		snapRadius   float64
		initialAngle float64
		wantApplied  bool
		wantAngle    float64 // Approximate expected angle
	}{
		{
			name:         "auto-aim disabled",
			autoAim:      false,
			strength:     0.3,
			entityX:      0,
			entityY:      0,
			enemyX:       50,
			enemyY:       0,
			snapRadius:   100,
			initialAngle: math.Pi / 2,
			wantApplied:  false,
			wantAngle:    math.Pi / 2,
		},
		{
			name:         "enemy out of range",
			autoAim:      true,
			strength:     0.3,
			entityX:      0,
			entityY:      0,
			enemyX:       200,
			enemyY:       0,
			snapRadius:   100,
			initialAngle: math.Pi / 2,
			wantApplied:  false,
			wantAngle:    math.Pi / 2,
		},
		{
			name:         "auto-aim applied",
			autoAim:      true,
			strength:     1.0, // Full snap
			entityX:      0,
			entityY:      0,
			enemyX:       50,
			enemyY:       0,
			snapRadius:   100,
			initialAngle: math.Pi / 2,
			wantApplied:  true,
			wantAngle:    0, // Should snap to 0 (right)
		},
		{
			name:         "partial auto-aim",
			autoAim:      true,
			strength:     0.5, // Partial snap
			entityX:      0,
			entityY:      0,
			enemyX:       50,
			enemyY:       0,
			snapRadius:   100,
			initialAngle: math.Pi / 2, // Aiming down
			wantApplied:  true,
			wantAngle:    math.Pi / 4, // Between down and right
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(tt.initialAngle)
			comp.AutoAim = tt.autoAim
			comp.AutoAimStrength = tt.strength
			comp.SnapRadius = tt.snapRadius

			applied := comp.ApplyAutoAim(tt.entityX, tt.entityY, tt.enemyX, tt.enemyY)

			if applied != tt.wantApplied {
				t.Errorf("ApplyAutoAim() = %v, want %v", applied, tt.wantApplied)
			}
			if !floatEqual(comp.AimAngle, tt.wantAngle, 0.3) {
				t.Errorf("AimAngle = %v, want ~%v", comp.AimAngle, tt.wantAngle)
			}
		})
	}
}

// TestAimComponent_IsAimingAt tests aim direction checking
func TestAimComponent_IsAimingAt(t *testing.T) {
	tests := []struct {
		name       string
		aimAngle   float64
		entityX    float64
		entityY    float64
		targetX    float64
		targetY    float64
		tolerance  float64
		wantAiming bool
	}{
		{
			name:       "directly aiming",
			aimAngle:   0,
			entityX:    0,
			entityY:    0,
			targetX:    100,
			targetY:    0,
			tolerance:  math.Pi / 16,
			wantAiming: true,
		},
		{
			name:       "within tolerance",
			aimAngle:   0.1, // Slightly off right
			entityX:    0,
			entityY:    0,
			targetX:    100,
			targetY:    0,
			tolerance:  math.Pi / 16,
			wantAiming: true,
		},
		{
			name:       "outside tolerance",
			aimAngle:   math.Pi / 4, // 45 degrees
			entityX:    0,
			entityY:    0,
			targetX:    100,
			targetY:    0,
			tolerance:  math.Pi / 16,
			wantAiming: false,
		},
		{
			name:       "opposite direction",
			aimAngle:   math.Pi,
			entityX:    0,
			entityY:    0,
			targetX:    100,
			targetY:    0,
			tolerance:  math.Pi / 16,
			wantAiming: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewAimComponent(tt.aimAngle)
			aiming := comp.IsAimingAt(tt.entityX, tt.entityY, tt.targetX, tt.targetY, tt.tolerance)

			if aiming != tt.wantAiming {
				t.Errorf("IsAimingAt() = %v, want %v", aiming, tt.wantAiming)
			}
		})
	}
}
