package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestNewMeleeSwingArcSystem(t *testing.T) {
	tests := []struct {
		name  string
		world *World
		seed  int64
	}{
		{"with world", NewWorld(), 42},
		{"nil world", nil, 0},
		{"large seed", NewWorld(), 999999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewMeleeSwingArcSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("NewMeleeSwingArcSystem returned nil")
			}
			if sys.genreID != "fantasy" {
				t.Errorf("default genre = %q, want fantasy", sys.genreID)
			}
			if sys.rng == nil {
				t.Error("rng is nil")
			}
		})
	}
}

func TestMeleeSwingArcSystem_SetGenre(t *testing.T) {
	sys := NewMeleeSwingArcSystem(NewWorld(), 42)
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "scifi", "post-apocalyptic", "postapoc", "unknown"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("SetGenre(%q): genreID = %q", g, sys.genreID)
		}
	}
}

func TestMeleeSwingArcSystem_GenrePresets(t *testing.T) {
	sys := NewMeleeSwingArcSystem(NewWorld(), 42)
	tests := []struct {
		genre        string
		wantR, wantG float64
	}{
		{"horror", 0.8, 0.2},
		{"cyberpunk", 0.3, 1.0},
		{"fantasy", 1.0, 0.95},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			p := sys.getPreset(tt.genre)
			if math.Abs(p.R-tt.wantR) > 0.01 || math.Abs(p.G-tt.wantG) > 0.01 {
				t.Errorf("preset R,G = %.2f,%.2f; want %.2f,%.2f", p.R, p.G, tt.wantR, tt.wantG)
			}
			if p.IntensityScale <= 0 || p.IntensityScale > 2.0 {
				t.Errorf("IntensityScale = %.2f out of range", p.IntensityScale)
			}
			if p.ThicknessScale <= 0 || p.ThicknessScale > 2.0 {
				t.Errorf("ThicknessScale = %.2f out of range", p.ThicknessScale)
			}
		})
	}
}

func TestMeleeSwingArcComponent_Type(t *testing.T) {
	c := &MeleeSwingArcComponent{}
	if c.Type() != "melee_swing_arc" {
		t.Errorf("Type() = %q, want melee_swing_arc", c.Type())
	}
}

func TestMeleeSwingArcSystem_UpdateAttachesComponent(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval // force full scan

	entity := NewEntity(1)
	entity.AddComponent(&AnimationComponent{CurrentState: AnimationStateIdle, Playing: true})
	entity.AddComponent(&AttackComponent{Damage: 10, Range: 25, Cooldown: 1.0})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	comp, has := entity.GetComponent("melee_swing_arc")
	if !has || comp == nil {
		t.Fatal("MeleeSwingArcComponent not attached after full scan")
	}
}

func TestMeleeSwingArcSystem_UpdateSkipsWithoutComponents(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval

	entity := NewEntity(1)
	// No animation or attack component
	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	_, has := entity.GetComponent("melee_swing_arc")
	if has {
		t.Error("should not attach arc to entity without animation+attack")
	}
}

func TestMeleeSwingArcSystem_ActivatesOnAttackStart(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval

	entity := NewEntity(1)
	anim := &AnimationComponent{CurrentState: AnimationStateIdle, Playing: true}
	entity.AddComponent(anim)
	entity.AddComponent(&AttackComponent{Damage: 10, Range: 30, Cooldown: 0.5})

	entities := []*Entity{entity}

	// First update: idle state, attaches component
	sys.Update(entities, 0.016)
	comp, _ := entity.GetComponent("melee_swing_arc")
	arc := comp.(*MeleeSwingArcComponent)
	if arc.Active {
		t.Error("arc should not be active during idle")
	}

	// Transition to attack
	anim.CurrentState = AnimationStateAttack
	sys.Update(entities, 0.016)
	if !arc.Active {
		t.Error("arc should activate on attack start")
	}
	if arc.Phase < 0 || arc.Phase > 1.0 {
		t.Errorf("Phase = %.3f out of range", arc.Phase)
	}
	if arc.R <= 0 || arc.G <= 0 || arc.B <= 0 {
		t.Errorf("arc color not set: R=%.2f G=%.2f B=%.2f", arc.R, arc.G, arc.B)
	}
	if arc.ArcRadius <= 0 {
		t.Errorf("ArcRadius = %.2f, want > 0", arc.ArcRadius)
	}
}

func TestMeleeSwingArcSystem_ArcFadesOverTime(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval

	entity := NewEntity(1)
	anim := &AnimationComponent{CurrentState: AnimationStateAttack, Playing: true}
	entity.AddComponent(anim)
	entity.AddComponent(&AttackComponent{Damage: 10, Range: 20, Cooldown: 0.5})

	entities := []*Entity{entity}

	// First frame: attach + activate
	sys.Update(entities, 0.016)
	comp, _ := entity.GetComponent("melee_swing_arc")
	arc := comp.(*MeleeSwingArcComponent)
	initialA := arc.A

	// Advance several frames
	for i := 0; i < 5; i++ {
		sys.Update(entities, 0.05)
	}
	if arc.A >= initialA {
		t.Errorf("opacity should decrease: initial=%.3f current=%.3f", initialA, arc.A)
	}

	// Advance enough to complete
	for i := 0; i < 20; i++ {
		sys.Update(entities, 0.05)
	}
	if arc.Active {
		t.Error("arc should deactivate after phase completes")
	}
}

func TestMeleeSwingArcSystem_FacingToArcAngles(t *testing.T) {
	sys := NewMeleeSwingArcSystem(NewWorld(), 42)
	dirs := []Direction{DirUp, DirDown, DirLeft, DirRight}
	for _, d := range dirs {
		start, end := sys.facingToArcAngles(d)
		if start == end {
			t.Errorf("direction %v: start == end", d)
		}
		if math.IsNaN(start) || math.IsNaN(end) {
			t.Errorf("direction %v: NaN angles", d)
		}
	}
}

func TestMeleeSwingArcSystem_ApplyRarityBoost(t *testing.T) {
	sys := NewMeleeSwingArcSystem(NewWorld(), 42)
	base := swingArcMaterialColor{R: 0.8, G: 0.7, B: 0.6}
	tests := []struct {
		rarity   item.Rarity
		wantMore bool
	}{
		{item.RarityCommon, false},
		{item.RarityUncommon, true},
		{item.RarityRare, true},
		{item.RarityEpic, true},
		{item.RarityLegendary, true},
	}
	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			result := sys.applyRarityBoost(base, tt.rarity)
			boosted := result.R > base.R || result.G > base.G || result.B > base.B
			if boosted != tt.wantMore {
				t.Errorf("rarity %v: boosted=%v want=%v", tt.rarity, boosted, tt.wantMore)
			}
			if result.R > 1.0 || result.G > 1.0 || result.B > 1.0 {
				t.Error("color exceeds 1.0")
			}
		})
	}
}

func TestMeleeSwingArcSystem_ResolveWeaponColorNoEquipment(t *testing.T) {
	sys := NewMeleeSwingArcSystem(NewWorld(), 42)
	entity := NewEntity(1)
	color := sys.resolveWeaponColor(entity)
	if color.R != 0.9 || color.G != 0.9 || color.B != 0.9 {
		t.Errorf("default color = %.2f,%.2f,%.2f; want 0.9,0.9,0.9", color.R, color.G, color.B)
	}
}

func TestMeleeSwingArcSystem_DoesNotReactivateWhileAttacking(t *testing.T) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval

	entity := NewEntity(1)
	anim := &AnimationComponent{CurrentState: AnimationStateAttack, Playing: true}
	entity.AddComponent(anim)
	entity.AddComponent(&AttackComponent{Damage: 10, Range: 20, Cooldown: 0.5})

	entities := []*Entity{entity}

	// First frame: activate
	sys.Update(entities, 0.016)
	comp, _ := entity.GetComponent("melee_swing_arc")
	arc := comp.(*MeleeSwingArcComponent)

	// Advance but stay in attack state
	sys.Update(entities, 0.05)
	phase := arc.Phase

	sys.Update(entities, 0.05)
	// Phase should advance, not reset (no re-activation)
	if arc.Phase <= phase {
		t.Error("phase should advance, not reset during sustained attack")
	}
}

func BenchmarkMeleeSwingArcSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewMeleeSwingArcSystem(world, 42)
	sys.timeSinceCheck = sys.updateInterval

	entities := make([]*Entity, 200)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&AnimationComponent{CurrentState: AnimationStateIdle, Playing: true})
		e.AddComponent(&AttackComponent{Damage: 10, Range: 20, Cooldown: 1.0})
		e.AddComponent(&MeleeSwingArcComponent{})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
