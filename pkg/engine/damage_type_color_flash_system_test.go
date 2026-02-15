package engine

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
)

func TestNewDamageTypeFlashComponent(t *testing.T) {
	comp := NewDamageTypeFlashComponent()
	if comp == nil {
		t.Fatal("expected non-nil component")
	}
	if comp.Type() != "damage_type_flash" {
		t.Errorf("expected type 'damage_type_flash', got %q", comp.Type())
	}
	if comp.FlashDuration != 0.3 {
		t.Errorf("expected default duration 0.3, got %f", comp.FlashDuration)
	}
	if comp.FlashTimer != 0 {
		t.Errorf("expected zero flash timer, got %f", comp.FlashTimer)
	}
}

func TestNewDamageTypeColorFlashSystem(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
	if sys.flashDuration != 0.3 {
		t.Errorf("expected flash duration 0.3, got %f", sys.flashDuration)
	}
}

func TestDamageTypeColorFlashSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"scifi", "scifi"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewDamageTypeColorFlashSystem(world, 42)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("expected genre %q, got %q", tt.genreID, sys.genreID)
			}
		})
	}
}

func TestDamageTypeColorFlashSystem_ColorForDamageType(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	tests := []struct {
		name       string
		damageType combat.DamageType
		wantColor  color.RGBA
	}{
		{"physical", combat.DamagePhysical, sys.preset.Physical},
		{"magical", combat.DamageMagical, sys.preset.Magical},
		{"fire", combat.DamageFire, sys.preset.Fire},
		{"ice", combat.DamageIce, sys.preset.Ice},
		{"lightning", combat.DamageLightning, sys.preset.Lightning},
		{"poison", combat.DamagePoison, sys.preset.Poison},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sys.colorForDamageType(tt.damageType)
			if c != tt.wantColor {
				t.Errorf("expected %v, got %v", tt.wantColor, c)
			}
		})
	}
}

func TestDamageTypeColorFlashSystem_OnDamageDealtNilTarget(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)
	// Should not panic
	sys.OnDamageDealt(nil, nil, 10.0)
}

func TestDamageTypeColorFlashSystem_OnDamageDealtNoComponent(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)
	target := NewEntity(1)
	// Should not panic when target has no DamageTypeFlashComponent
	sys.OnDamageDealt(nil, target, 10.0)
}

func TestDamageTypeColorFlashSystem_OnDamageDealtSetsFlash(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	attacker := NewEntity(1)
	attacker.AddComponent(&AttackComponent{
		Damage:     20.0,
		DamageType: combat.DamageFire,
	})

	target := NewEntity(2)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)

	sys.OnDamageDealt(attacker, target, 25.0)

	if flash.FlashTimer <= 0 {
		t.Error("expected flash timer > 0 after damage")
	}
	if flash.LastDamageType != combat.DamageFire {
		t.Errorf("expected fire damage type, got %v", flash.LastDamageType)
	}
	if flash.FlashColor != sys.preset.Fire {
		t.Errorf("expected fire color, got %v", flash.FlashColor)
	}
	if flash.FlashIntensity <= 0 {
		t.Error("expected positive flash intensity")
	}
}

func TestDamageTypeColorFlashSystem_OnDamageDealtNilAttacker(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(2)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)

	sys.OnDamageDealt(nil, target, 10.0)

	// Should default to physical damage type
	if flash.FlashColor != sys.preset.Physical {
		t.Errorf("expected physical color for nil attacker, got %v", flash.FlashColor)
	}
}

func TestDamageTypeColorFlashSystem_IntensityClamping(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)

	// Very high damage should cap at maxIntensity
	sys.OnDamageDealt(nil, target, 999.0)
	if flash.FlashIntensity > sys.maxIntensity {
		t.Errorf("intensity %f exceeds max %f", flash.FlashIntensity, sys.maxIntensity)
	}

	// Very low damage should have minimum 0.2 intensity
	sys.OnDamageDealt(nil, target, 0.1)
	if flash.FlashIntensity < 0.2 {
		t.Errorf("intensity %f below minimum 0.2", flash.FlashIntensity)
	}
}

func TestDamageTypeColorFlashSystem_UpdateDecaysTimer(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)
	target.AddComponent(NewVisualFeedbackComponent())

	// Trigger a flash
	sys.OnDamageDealt(nil, target, 30.0)
	initialTimer := flash.FlashTimer

	entities := []*Entity{target}
	sys.Update(entities, 0.1)

	if flash.FlashTimer >= initialTimer {
		t.Error("expected timer to decay after update")
	}
}

func TestDamageTypeColorFlashSystem_UpdateAppliesTint(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)
	feedback := NewVisualFeedbackComponent()
	target.AddComponent(feedback)

	attacker := NewEntity(2)
	attacker.AddComponent(&AttackComponent{DamageType: combat.DamagePoison})

	sys.OnDamageDealt(attacker, target, 40.0)
	sys.Update([]*Entity{target}, 0.05)

	// Tint should be shifted from 1.0 toward the poison green channel
	if feedback.TintG >= 1.0 {
		t.Error("expected green tint to shift below 1.0 or stay if green > 255")
	}
}

func TestDamageTypeColorFlashSystem_UpdateNoFlashNoEffect(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)
	feedback := NewVisualFeedbackComponent()
	target.AddComponent(feedback)

	// No damage dealt, timer is 0
	sys.Update([]*Entity{target}, 0.1)

	// Tint should remain default
	if feedback.TintR != 1.0 || feedback.TintG != 1.0 || feedback.TintB != 1.0 {
		t.Error("expected default tint when no flash active")
	}
}

func TestDamageTypeColorFlashSystem_UpdateTimerExpires(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)
	feedback := NewVisualFeedbackComponent()
	target.AddComponent(feedback)

	sys.OnDamageDealt(nil, target, 20.0)

	// Fast-forward past flash duration
	sys.Update([]*Entity{target}, 1.0)

	if flash.FlashTimer != 0 {
		t.Errorf("expected timer to be 0, got %f", flash.FlashTimer)
	}
}

func TestDamageTypeColorFlashSystem_GenrePresetsDiffer(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	fantasy := sys.getGenrePreset("fantasy")
	horror := sys.getGenrePreset("horror")

	if fantasy.Fire == horror.Fire {
		t.Error("expected different fire colors for fantasy vs horror")
	}
	if fantasy.Physical == horror.Physical {
		t.Error("expected different physical colors for fantasy vs horror")
	}
}

func TestDamageTypeColorFlashSystem_UpdateSkipsNoComponent(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	// Entity without the component
	target := NewEntity(1)
	// Should not panic
	sys.Update([]*Entity{target}, 0.1)
}

func TestDamageTypeColorFlashSystem_UpdateSkipsNoFeedback(t *testing.T) {
	world := NewWorld()
	sys := NewDamageTypeColorFlashSystem(world, 42)

	target := NewEntity(1)
	flash := NewDamageTypeFlashComponent()
	target.AddComponent(flash)
	// No VisualFeedbackComponent

	sys.OnDamageDealt(nil, target, 20.0)
	// Should not panic, just skip tint application
	sys.Update([]*Entity{target}, 0.1)

	// Timer should still decay
	if flash.FlashTimer >= 0.3 {
		t.Error("expected timer to decay even without feedback component")
	}
}
