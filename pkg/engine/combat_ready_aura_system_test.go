package engine

import (
	"math"
	"testing"
)

func TestCombatReadyAuraComponentType(t *testing.T) {
	comp := NewCombatReadyAuraComponent()
	if comp.Type() != "combat_ready_aura" {
		t.Errorf("expected type 'combat_ready_aura', got %q", comp.Type())
	}
}

func TestCombatReadyAuraComponentDefaults(t *testing.T) {
	comp := NewCombatReadyAuraComponent()
	if comp.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if comp.Opacity != 0 {
		t.Errorf("expected Opacity 0, got %f", comp.Opacity)
	}
	if comp.AuraRadius != 2.0 {
		t.Errorf("expected AuraRadius 2.0, got %f", comp.AuraRadius)
	}
}

func TestNewCombatReadyAuraSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestCombatReadyAuraSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)

	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("expected genre %q, got %q", g, sys.genreID)
		}
	}
}

func TestCombatReadyAuraSkipsWithoutAI(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})

	entities := []*Entity{entity}
	sys.Update(entities, 0.5)

	_, ok := entity.GetComponent("combat_ready_aura")
	if ok {
		t.Error("entity without AI should not get combat aura")
	}
}

func TestCombatReadyAuraSkipsPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(NewStubInput())

	entities := []*Entity{entity}
	// Force past throttle
	sys.timeSinceCheck = sys.stateCheckInterval + 0.01
	sys.Update(entities, 0.01)

	comp, ok := entity.GetComponent("combat_ready_aura")
	if ok {
		ac := comp.(*CombatReadyAuraComponent)
		if ac.Enabled {
			t.Error("player entities should not get combat aura")
		}
	}
}

func TestCombatReadyAuraHostileStates(t *testing.T) {
	tests := []struct {
		name          string
		state         AIState
		wantEnabled   bool
		minRadius     float64
		maxRadius     float64
		minOpacityTgt float64
	}{
		{"detect", AIStateDetect, true, 2.5, 4.0, 0.3},
		{"chase", AIStateChase, true, 4.0, 6.0, 0.5},
		{"attack", AIStateAttack, true, 6.0, 8.0, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCombatReadyAuraSystem(world, 42)
			sys.SetGenre("fantasy")

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&AIComponent{State: tt.state})

			entities := []*Entity{entity}
			sys.timeSinceCheck = sys.stateCheckInterval + 0.01
			sys.Update(entities, 0.01)

			comp, ok := entity.GetComponent("combat_ready_aura")
			if !ok {
				t.Fatal("expected combat_ready_aura component to be created")
			}
			ac := comp.(*CombatReadyAuraComponent)

			if ac.Enabled != tt.wantEnabled {
				t.Errorf("expected Enabled=%v, got %v", tt.wantEnabled, ac.Enabled)
			}
			if ac.AuraRadius < tt.minRadius || ac.AuraRadius > tt.maxRadius {
				t.Errorf("AuraRadius %f not in [%f, %f]", ac.AuraRadius, tt.minRadius, tt.maxRadius)
			}
			if ac.TargetOpacity < tt.minOpacityTgt {
				t.Errorf("TargetOpacity %f < %f", ac.TargetOpacity, tt.minOpacityTgt)
			}
			if ac.TriggerState != tt.state {
				t.Errorf("expected TriggerState %v, got %v", tt.state, ac.TriggerState)
			}
		})
	}
}

func TestCombatReadyAuraPassiveStatesFadeOut(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	aiComp := &AIComponent{State: AIStateAttack}
	entity.AddComponent(aiComp)

	entities := []*Entity{entity}

	// First: activate aura in attack state
	sys.timeSinceCheck = sys.stateCheckInterval + 0.01
	sys.Update(entities, 0.01)

	comp, _ := entity.GetComponent("combat_ready_aura")
	ac := comp.(*CombatReadyAuraComponent)
	if !ac.Enabled {
		t.Fatal("expected aura to be enabled in attack state")
	}

	// Simulate some opacity buildup
	ac.Opacity = 0.8

	// Switch to idle
	aiComp.State = AIStateIdle
	sys.timeSinceCheck = sys.stateCheckInterval + 0.01
	sys.Update(entities, 0.01)

	if ac.TargetOpacity != 0 {
		t.Errorf("expected TargetOpacity 0 for passive state, got %f", ac.TargetOpacity)
	}
}

func TestCombatReadyAuraFadeAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	ac := NewCombatReadyAuraComponent()
	ac.Enabled = true
	ac.TargetOpacity = 0.8
	ac.Opacity = 0.0
	ac.PulseSpeed = 1.0
	ac.BaseIntensity = 0.5
	ac.PulseAmplitude = 0.1
	entity.AddComponent(ac)

	entities := []*Entity{entity}

	// Run several frames of fade-in
	for i := 0; i < 10; i++ {
		sys.updatePulseAndFade(entities, 0.05)
	}

	if ac.Opacity <= 0 {
		t.Error("expected opacity to increase during fade-in")
	}
	if ac.Opacity > ac.TargetOpacity+0.01 {
		t.Errorf("opacity %f exceeded target %f", ac.Opacity, ac.TargetOpacity)
	}
}

func TestCombatReadyAuraPulseAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)

	entity := NewEntity(1)
	ac := NewCombatReadyAuraComponent()
	ac.Enabled = true
	ac.Opacity = 0.8
	ac.TargetOpacity = 0.8
	ac.PulseSpeed = 2.0
	ac.BaseIntensity = 0.6
	ac.PulseAmplitude = 0.15
	entity.AddComponent(ac)

	entities := []*Entity{entity}

	// Collect intensity samples over multiple frames
	intensities := make([]float64, 0, 20)
	for i := 0; i < 20; i++ {
		sys.updatePulseAndFade(entities, 0.05)
		intensities = append(intensities, ac.CurrentIntensity)
	}

	// Verify intensity varies (pulse is animating)
	minI, maxI := intensities[0], intensities[0]
	for _, v := range intensities[1:] {
		if v < minI {
			minI = v
		}
		if v > maxI {
			maxI = v
		}
	}
	if maxI-minI < 0.01 {
		t.Error("expected pulse animation to vary intensity")
	}
}

func TestCombatReadyAuraGenreColors(t *testing.T) {
	tests := []struct {
		genre string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"fantasy", 0.95, 0.60, 0.15},
		{"horror", 0.90, 0.10, 0.10},
		{"scifi", 0.15, 0.70, 0.95},
		{"cyberpunk", 0.85, 0.15, 0.90},
		{"postapoc", 0.90, 0.70, 0.15},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewCombatReadyAuraSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&AIComponent{State: AIStateAttack})

			entities := []*Entity{entity}
			sys.timeSinceCheck = sys.stateCheckInterval + 0.01
			sys.Update(entities, 0.01)

			comp, ok := entity.GetComponent("combat_ready_aura")
			if !ok {
				t.Fatal("expected combat_ready_aura component")
			}
			ac := comp.(*CombatReadyAuraComponent)

			// Check color is close to genre palette (within variation range)
			if math.Abs(ac.AuraR-tt.wantR) > 0.05 {
				t.Errorf("AuraR %f not near %f for genre %s", ac.AuraR, tt.wantR, tt.genre)
			}
			if math.Abs(ac.AuraG-tt.wantG) > 0.05 {
				t.Errorf("AuraG %f not near %f for genre %s", ac.AuraG, tt.wantG, tt.genre)
			}
			if math.Abs(ac.AuraB-tt.wantB) > 0.05 {
				t.Errorf("AuraB %f not near %f for genre %s", ac.AuraB, tt.wantB, tt.genre)
			}
		})
	}
}

func TestCombatReadyAuraThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateChase})

	entities := []*Entity{entity}

	// First update with small deltaTime (below interval)
	sys.Update(entities, 0.05)

	// Component should not be created yet (throttled)
	_, ok := entity.GetComponent("combat_ready_aura")
	if ok {
		t.Error("expected state check to be throttled on first small update")
	}

	// Update enough to pass throttle interval
	sys.Update(entities, 0.2)

	_, ok = entity.GetComponent("combat_ready_aura")
	if !ok {
		t.Error("expected combat_ready_aura after throttle interval passed")
	}
}

func TestCombatReadyAuraFadeOutDisables(t *testing.T) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)

	entity := NewEntity(1)
	ac := NewCombatReadyAuraComponent()
	ac.Enabled = true
	ac.Opacity = 0.05
	ac.TargetOpacity = 0.0
	ac.PulseSpeed = 1.0
	ac.BaseIntensity = 0.5
	entity.AddComponent(ac)

	entities := []*Entity{entity}

	// Run fade until fully faded
	for i := 0; i < 20; i++ {
		sys.updatePulseAndFade(entities, 0.05)
	}

	if ac.Enabled {
		t.Error("expected aura to be disabled after fading out")
	}
	if ac.Opacity > 0.01 {
		t.Errorf("expected opacity near 0 after fade, got %f", ac.Opacity)
	}
}

func TestClampCombatAura(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampCombatAura(tt.input)
		if got != tt.want {
			t.Errorf("clampCombatAura(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestBuildCombatAuraPalettes(t *testing.T) {
	palettes := buildCombatAuraPalettes()
	expected := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expected {
		if _, ok := palettes[genre]; !ok {
			t.Errorf("missing palette for genre %q", genre)
		}
	}
}

func BenchmarkCombatReadyAuraUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewCombatReadyAuraSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		e.AddComponent(&StubSprite{Visible: true})
		e.AddComponent(&AIComponent{State: AIStateChase})
		ac := NewCombatReadyAuraComponent()
		ac.Enabled = true
		ac.PulseSpeed = 1.0
		ac.BaseIntensity = 0.5
		ac.PulseAmplitude = 0.1
		ac.Opacity = 0.6
		ac.TargetOpacity = 0.6
		e.AddComponent(ac)
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
