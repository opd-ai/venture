package engine

import (
	"math"
	"testing"
)

func TestAIStateBubbleComponentType(t *testing.T) {
	comp := NewAIStateBubbleComponent()
	if comp.Type() != "ai_state_bubble" {
		t.Errorf("expected type 'ai_state_bubble', got %q", comp.Type())
	}
}

func TestAIStateBubbleComponentDefaults(t *testing.T) {
	comp := NewAIStateBubbleComponent()
	if comp.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if comp.Opacity != 0 {
		t.Errorf("expected Opacity 0, got %f", comp.Opacity)
	}
	if comp.OffsetY != -20.0 {
		t.Errorf("expected OffsetY -20.0, got %f", comp.OffsetY)
	}
	if comp.BobAmplitude != 2.0 {
		t.Errorf("expected BobAmplitude 2.0, got %f", comp.BobAmplitude)
	}
	if comp.Symbol != "" {
		t.Errorf("expected empty Symbol, got %q", comp.Symbol)
	}
}

func TestNewAIStateBubbleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", sys.genreID)
	}
}

func TestAIStateBubbleSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)

	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, g := range genres {
		sys.SetGenre(g)
		if sys.genreID != g {
			t.Errorf("expected genre %q, got %q", g, sys.genreID)
		}
	}
}

func TestAIStateBubbleSetGenreInvalid(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("nonexistent")
	if sys.genreID != "fantasy" {
		t.Errorf("expected genre to remain 'fantasy' for invalid genre, got %q", sys.genreID)
	}
}

func TestAIStateBubbleSkipsWithoutAI(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})

	entities := []*Entity{entity}
	sys.Update(entities, 0.5)

	_, ok := entity.GetComponent("ai_state_bubble")
	if ok {
		t.Error("entity without AI should not get state bubble")
	}
}

func TestAIStateBubbleSkipsPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(NewStubInput())

	entities := []*Entity{entity}
	sys.Update(entities, 0.5)

	_, ok := entity.GetComponent("ai_state_bubble")
	if ok {
		t.Error("player entity should not get state bubble")
	}
}

func TestAIStateBubbleSkipsNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateIdle})

	entities := []*Entity{entity}
	sys.Update(entities, 0.5)

	_, ok := entity.GetComponent("ai_state_bubble")
	if ok {
		t.Error("entity without position should not get state bubble")
	}
}

func TestAIStateBubbleSkipsNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateIdle})

	entities := []*Entity{entity}
	sys.Update(entities, 0.5)

	_, ok := entity.GetComponent("ai_state_bubble")
	if ok {
		t.Error("entity without sprite should not get state bubble")
	}
}

func TestAIStateBubbleAllStates(t *testing.T) {
	tests := []struct {
		name           string
		state          AIState
		expectedSymbol string
		minOpacity     float64
	}{
		{"idle", AIStateIdle, "zzz", 0.40},
		{"patrol", AIStatePatrol, "~", 0.45},
		{"detect", AIStateDetect, "!", 0.75},
		{"chase", AIStateChase, "!!", 0.85},
		{"attack", AIStateAttack, "X", 0.90},
		{"flee", AIStateFlee, "...", 0.65},
		{"return", AIStateReturn, "<", 0.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewAIStateBubbleSystem(world, 42)
			sys.SetGenre("fantasy")

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&AIComponent{State: tt.state})

			entities := []*Entity{entity}
			// First update to trigger state check and start fade
			sys.Update(entities, 0.3)

			comp, ok := entity.GetComponent("ai_state_bubble")
			if !ok {
				t.Fatal("expected ai_state_bubble component")
			}
			bc := comp.(*AIStateBubbleComponent)
			if bc.Symbol != tt.expectedSymbol {
				t.Errorf("expected symbol %q, got %q", tt.expectedSymbol, bc.Symbol)
			}
			if bc.TargetOpacity < tt.minOpacity {
				t.Errorf("expected target opacity >= %f, got %f", tt.minOpacity, bc.TargetOpacity)
			}
			if !bc.Enabled {
				t.Error("expected bubble to be enabled")
			}
		})
	}
}

func TestAIStateBubbleGenreColors(t *testing.T) {
	tests := []struct {
		genre string
		minR  float64
		maxR  float64
	}{
		{"fantasy", 0.85, 1.0},   // Warm gold (high R)
		{"horror", 0.80, 1.0},    // Blood red (high R)
		{"scifi", 0.10, 0.35},    // Cyan (low R)
		{"cyberpunk", 0.80, 1.0}, // Neon magenta (high R)
		{"postapoc", 0.80, 1.0},  // Amber (high R)
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewAIStateBubbleSystem(world, 42)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&StubSprite{Visible: true})
			entity.AddComponent(&AIComponent{State: AIStateAttack})

			sys.Update([]*Entity{entity}, 0.3)

			comp, _ := entity.GetComponent("ai_state_bubble")
			bc := comp.(*AIStateBubbleComponent)
			if bc.SymbolR < tt.minR || bc.SymbolR > tt.maxR {
				t.Errorf("genre %s: expected R in [%f, %f], got %f", tt.genre, tt.minR, tt.maxR, bc.SymbolR)
			}
		})
	}
}

func TestAIStateBubbleBobAnimation(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateDetect})

	// First update creates component and starts animation
	sys.Update([]*Entity{entity}, 0.3)

	comp, _ := entity.GetComponent("ai_state_bubble")
	bc := comp.(*AIStateBubbleComponent)
	offsetBefore := bc.OffsetY

	// Several frames of animation
	sys.Update([]*Entity{entity}, 0.5)
	offsetAfter := bc.OffsetY

	if offsetBefore == offsetAfter {
		t.Error("expected bob animation to change OffsetY over time")
	}

	// Verify offset stays in reasonable range
	if bc.OffsetY < -24.0 || bc.OffsetY > -16.0 {
		t.Errorf("OffsetY %f outside expected bob range [-24, -16]", bc.OffsetY)
	}
}

func TestAIStateBubbleFadeTransition(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateAttack})

	// First update triggers state check and starts fading in
	sys.Update([]*Entity{entity}, 0.3)

	comp, _ := entity.GetComponent("ai_state_bubble")
	bc := comp.(*AIStateBubbleComponent)

	// Opacity should have started increasing toward target
	if bc.Opacity <= 0 {
		t.Error("expected opacity to increase after update")
	}
	if bc.TargetOpacity <= 0 {
		t.Error("expected non-zero target opacity for attack state")
	}
}

func TestAIStateBubbleStateTransition(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	aiComp := &AIComponent{State: AIStateIdle}
	entity.AddComponent(aiComp)

	sys.Update([]*Entity{entity}, 0.3)

	comp, _ := entity.GetComponent("ai_state_bubble")
	bc := comp.(*AIStateBubbleComponent)
	if bc.Symbol != "zzz" {
		t.Errorf("expected 'zzz' for idle, got %q", bc.Symbol)
	}

	// Transition to chase
	aiComp.State = AIStateChase
	sys.Update([]*Entity{entity}, 0.3)

	if bc.Symbol != "!!" {
		t.Errorf("expected '!!' for chase, got %q", bc.Symbol)
	}
	if bc.PreviousState != AIStateIdle {
		t.Error("expected PreviousState to be idle after transition")
	}
}

func TestAIStateBubbleScaleFlash(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	aiComp := &AIComponent{State: AIStateIdle}
	entity.AddComponent(aiComp)

	sys.Update([]*Entity{entity}, 0.3)

	// Trigger a state transition
	aiComp.State = AIStateAttack
	sys.Update([]*Entity{entity}, 0.3)

	comp, _ := entity.GetComponent("ai_state_bubble")
	bc := comp.(*AIStateBubbleComponent)

	// Scale should be elevated after transition (flash effect)
	sym := sys.symbols[AIStateAttack]
	if bc.Scale < sym.Scale {
		t.Errorf("expected scale >= %f after transition flash, got %f", sym.Scale, bc.Scale)
	}
}

func TestAIStateBubbleNilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := []*Entity{nil}
	// Should not panic
	sys.Update(entities, 0.1)
}

func TestAIStateBubbleBobPhaseWraps(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&StubSprite{Visible: true})
	entity.AddComponent(&AIComponent{State: AIStateAttack})

	sys.Update([]*Entity{entity}, 0.3)

	comp, _ := entity.GetComponent("ai_state_bubble")
	bc := comp.(*AIStateBubbleComponent)

	// Run many updates to wrap phase
	for i := 0; i < 100; i++ {
		sys.Update([]*Entity{entity}, 0.1)
	}

	if bc.BobPhase < 0 || bc.BobPhase > 2*math.Pi {
		t.Errorf("BobPhase %f should be within [0, 2π]", bc.BobPhase)
	}
}

func TestClampBubble(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}
	for _, tt := range tests {
		got := clampBubble(tt.input)
		if got != tt.expected {
			t.Errorf("clampBubble(%f) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestAIStateBubbleMultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 5)
	states := []AIState{AIStateIdle, AIStatePatrol, AIStateDetect, AIStateChase, AIStateAttack}
	expectedSymbols := []string{"zzz", "~", "!", "!!", "X"}

	for i := 0; i < 5; i++ {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 50), Y: 10})
		e.AddComponent(&StubSprite{Visible: true})
		e.AddComponent(&AIComponent{State: states[i]})
		entities[i] = e
	}

	sys.Update(entities, 0.3)

	for i, e := range entities {
		comp, ok := e.GetComponent("ai_state_bubble")
		if !ok {
			t.Fatalf("entity %d: expected ai_state_bubble component", i)
		}
		bc := comp.(*AIStateBubbleComponent)
		if bc.Symbol != expectedSymbols[i] {
			t.Errorf("entity %d: expected symbol %q, got %q", i, expectedSymbols[i], bc.Symbol)
		}
	}
}

func BenchmarkAIStateBubbleSystem(b *testing.B) {
	world := NewWorld()
	sys := NewAIStateBubbleSystem(world, 42)
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 200)
	states := []AIState{AIStateIdle, AIStatePatrol, AIStateDetect, AIStateChase, AIStateAttack, AIStateFlee}
	for i := 0; i < 200; i++ {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 5)})
		e.AddComponent(&StubSprite{Visible: true})
		e.AddComponent(&AIComponent{State: states[i%len(states)]})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
