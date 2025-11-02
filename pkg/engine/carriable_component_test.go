package engine

import (
	"math"
	"testing"
)

func TestNewCarriableComponent(t *testing.T) {
	tests := []struct {
		name                   string
		weight                 float64
		wantWeight             float64
		wantThrowVelMultiplier float64
		wantImpactDamage       float64
	}{
		{"light_object", 0.2, 0.2, 5.0, 2.0},
		{"medium_object", 0.5, 0.5, 2.0, 5.0},
		{"heavy_object", 1.0, 1.0, 1.0, 10.0},
		{"too_light_clamped", 0.05, 0.1, 10.0, 1.0},
		{"too_heavy_clamped", 1.5, 1.0, 1.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewCarriableComponent(tt.weight)

			if comp.Weight != tt.wantWeight {
				t.Errorf("Weight = %v, want %v", comp.Weight, tt.wantWeight)
			}
			if comp.ThrowVelocityMultiplier != tt.wantThrowVelMultiplier {
				t.Errorf("ThrowVelocityMultiplier = %v, want %v", comp.ThrowVelocityMultiplier, tt.wantThrowVelMultiplier)
			}
			if comp.ImpactDamage != tt.wantImpactDamage {
				t.Errorf("ImpactDamage = %v, want %v", comp.ImpactDamage, tt.wantImpactDamage)
			}
			if comp.IsCarried {
				t.Error("IsCarried should be false initially")
			}
			if comp.CarriedBy != 0 {
				t.Error("CarriedBy should be 0 initially")
			}
			if !comp.CanPickUp {
				t.Error("CanPickUp should be true initially")
			}
		})
	}
}

func TestCarriableComponent_Type(t *testing.T) {
	comp := NewCarriableComponent(0.5)
	if comp.Type() != "carriable" {
		t.Errorf("Type() = %v, want 'carriable'", comp.Type())
	}
}

func TestCarriableComponent_Pickup(t *testing.T) {
	comp := NewCarriableComponent(0.5)
	carrierID := uint64(123)

	comp.Pickup(carrierID)

	if !comp.IsCarried {
		t.Error("IsCarried should be true after Pickup()")
	}
	if comp.CarriedBy != carrierID {
		t.Errorf("CarriedBy = %v, want %v", comp.CarriedBy, carrierID)
	}
}

func TestCarriableComponent_Drop(t *testing.T) {
	comp := NewCarriableComponent(0.5)
	comp.Pickup(123)

	comp.Drop()

	if comp.IsCarried {
		t.Error("IsCarried should be false after Drop()")
	}
	if comp.CarriedBy != 0 {
		t.Errorf("CarriedBy = %v, want 0", comp.CarriedBy)
	}
}

func TestContextActionType_String(t *testing.T) {
	tests := []struct {
		name       string
		actionType ContextActionType
		expected   string
	}{
		{"open", ActionOpen, "Open"},
		{"close", ActionClose, "Close"},
		{"push", ActionPush, "Push"},
		{"pull", ActionPull, "Pull"},
		{"activate", ActionActivate, "Activate"},
		{"talk", ActionTalk, "Talk"},
		{"pickup", ActionPickup, "Pickup"},
		{"read", ActionRead, "Read"},
		{"unknown", ContextActionType(999), "Interact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.actionType.String()
			if got != tt.expected {
				t.Errorf("ContextActionType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewContextActionComponent(t *testing.T) {
	tests := []struct {
		name             string
		actionType       ContextActionType
		actionText       string
		wantText         string
		wantType         ContextActionType
		wantRange        float64
		wantCooldownTime float64
	}{
		{"door_open", ActionOpen, "Open Door", "Open Door", ActionOpen, 48.0, 0.5},
		{"chest_open", ActionOpen, "Open Chest", "Open Chest", ActionOpen, 48.0, 0.5},
		{"lever_activate", ActionActivate, "Pull Lever", "Pull Lever", ActionActivate, 48.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewContextActionComponent(tt.actionType, tt.actionText)

			if comp.ActionText != tt.wantText {
				t.Errorf("ActionText = %v, want %v", comp.ActionText, tt.wantText)
			}
			if comp.ActionType != tt.wantType {
				t.Errorf("ActionType = %v, want %v", comp.ActionType, tt.wantType)
			}
			if comp.InteractionRange != tt.wantRange {
				t.Errorf("InteractionRange = %v, want %v", comp.InteractionRange, tt.wantRange)
			}
			if !comp.IsAvailable {
				t.Error("IsAvailable should be true initially")
			}
			if comp.RequiresKey {
				t.Error("RequiresKey should be false initially")
			}
			if comp.CooldownTime != tt.wantCooldownTime {
				t.Errorf("CooldownTime = %v, want %v", comp.CooldownTime, tt.wantCooldownTime)
			}
			if comp.CooldownElapsed != 0 {
				t.Error("CooldownElapsed should be 0 initially")
			}
		})
	}
}

func TestContextActionComponent_Type(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen, "Open")
	if comp.Type() != "contextAction" {
		t.Errorf("Type() = %v, want 'contextAction'", comp.Type())
	}
}

func TestContextActionComponent_Update(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen, "Open")
	comp.CooldownElapsed = 1.0

	comp.Update(0.3)
	if comp.CooldownElapsed != 0.7 {
		t.Errorf("After 0.3s, CooldownElapsed = %v, want 0.7", comp.CooldownElapsed)
	}

	comp.Update(0.5)
	// Use approximate comparison for floating point
	if math.Abs(comp.CooldownElapsed-0.2) > 0.001 {
		t.Errorf("After 0.8s total, CooldownElapsed = %v, want 0.2", comp.CooldownElapsed)
	}

	comp.Update(0.5)
	if comp.CooldownElapsed != 0.0 {
		t.Errorf("After 1.3s total, CooldownElapsed = %v, want 0.0 (clamped)", comp.CooldownElapsed)
	}
}

func TestContextActionComponent_CanInteract(t *testing.T) {
	tests := []struct {
		name            string
		isAvailable     bool
		cooldownElapsed float64
		want            bool
	}{
		{"available_no_cooldown", true, 0.0, true},
		{"available_with_cooldown", true, 0.5, false},
		{"unavailable_no_cooldown", false, 0.0, false},
		{"unavailable_with_cooldown", false, 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &ContextActionComponent{
				IsAvailable:     tt.isAvailable,
				CooldownElapsed: tt.cooldownElapsed,
			}
			if got := comp.CanInteract(); got != tt.want {
				t.Errorf("CanInteract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextActionComponent_Activate(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen, "Open")
	comp.CooldownTime = 1.0
	comp.CooldownElapsed = 0.0

	comp.Activate()

	if comp.CooldownElapsed != comp.CooldownTime {
		t.Errorf("CooldownElapsed = %v, want %v", comp.CooldownElapsed, comp.CooldownTime)
	}
}

func TestHazardType_String(t *testing.T) {
	tests := []struct {
		name       string
		hazardType HazardType
		expected   string
	}{
		{"poison", HazardPoison, "poison"},
		{"oil", HazardOil, "oil"},
		{"water", HazardWater, "water"},
		{"smoke", HazardSmoke, "smoke"},
		{"unknown", HazardType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.hazardType.String()
			if got != tt.expected {
				t.Errorf("HazardType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewHazardComponent(t *testing.T) {
	tests := []struct {
		name             string
		hazardType       HazardType
		duration         float64
		radius           float64
		wantDamagePerSec float64
		wantMovementMult float64
	}{
		{"poison", HazardPoison, 10.0, 32.0, 5.0, 1.0},
		{"oil", HazardOil, 15.0, 48.0, 0.0, 0.7},
		{"water", HazardWater, 20.0, 64.0, 0.0, 0.8},
		{"smoke", HazardSmoke, 8.0, 96.0, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewHazardComponent(tt.hazardType, tt.duration, tt.radius)

			if comp.HazardType != tt.hazardType {
				t.Errorf("HazardType = %v, want %v", comp.HazardType, tt.hazardType)
			}
			if comp.Duration != tt.duration {
				t.Errorf("Duration = %v, want %v", comp.Duration, tt.duration)
			}
			if comp.Radius != tt.radius {
				t.Errorf("Radius = %v, want %v", comp.Radius, tt.radius)
			}
			if comp.DamagePerSecond != tt.wantDamagePerSec {
				t.Errorf("DamagePerSecond = %v, want %v", comp.DamagePerSecond, tt.wantDamagePerSec)
			}
			if comp.MovementMultiplier != tt.wantMovementMult {
				t.Errorf("MovementMultiplier = %v, want %v", comp.MovementMultiplier, tt.wantMovementMult)
			}
			if !comp.IsLingering {
				t.Error("IsLingering should be true initially")
			}
		})
	}
}

func TestHazardComponent_Type(t *testing.T) {
	comp := NewHazardComponent(HazardPoison, 10.0, 32.0)
	if comp.Type() != "hazard" {
		t.Errorf("Type() = %v, want 'hazard'", comp.Type())
	}
}

func TestHazardComponent_Update(t *testing.T) {
	comp := NewHazardComponent(HazardPoison, 10.0, 32.0)

	comp.Update(3.0)
	if comp.Duration != 7.0 {
		t.Errorf("After 3s, Duration = %v, want 7.0", comp.Duration)
	}

	comp.Update(5.0)
	if comp.Duration != 2.0 {
		t.Errorf("After 8s total, Duration = %v, want 2.0", comp.Duration)
	}

	comp.Update(3.0)
	if comp.Duration != -1.0 {
		t.Errorf("After 11s total, Duration = %v, want -1.0", comp.Duration)
	}
}

func TestHazardComponent_ShouldRemove(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		want     bool
	}{
		{"still_active", 5.0, false},
		{"just_expired", 0.0, false},
		{"expired", -0.1, true},
		{"permanent", -1.0, false}, // -1 = permanent
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &HazardComponent{Duration: tt.duration}
			if got := comp.ShouldRemove(); got != tt.want {
				t.Errorf("ShouldRemove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHazardComponent_IsDamaging(t *testing.T) {
	tests := []struct {
		name            string
		damagePerSecond float64
		want            bool
	}{
		{"damaging", 5.0, true},
		{"not_damaging", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &HazardComponent{DamagePerSecond: tt.damagePerSecond}
			if got := comp.IsDamaging(); got != tt.want {
				t.Errorf("IsDamaging() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHazardComponent_AffectsMovement(t *testing.T) {
	tests := []struct {
		name               string
		movementMultiplier float64
		want               bool
	}{
		{"slows_movement", 0.7, true},
		{"normal_movement", 1.0, false},
		{"speeds_movement", 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &HazardComponent{MovementMultiplier: tt.movementMultiplier}
			if got := comp.AffectsMovement(); got != tt.want {
				t.Errorf("AffectsMovement() = %v, want %v", got, tt.want)
			}
		})
	}
}
