package engine

import (
	"math"
	"testing"
)

// TestExpressionType_String tests expression type string conversion.
func TestExpressionType_String(t *testing.T) {
	tests := []struct {
		expType  ExpressionType
		expected string
	}{
		{ExpressionWave, "Wave"},
		{ExpressionCheer, "Cheer"},
		{ExpressionDance, "Dance"},
		{ExpressionLaugh, "Laugh"},
		{ExpressionCry, "Cry"},
		{ExpressionSit, "Sit"},
		{ExpressionPoint, "Point"},
		{ExpressionSalute, "Salute"},
		{ExpressionShrug, "Shrug"},
		{ExpressionThumbsUp, "ThumbsUp"},
		{ExpressionFacepalm, "Facepalm"},
		{ExpressionSleep, "Sleep"},
		{ExpressionType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.expType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestExpressionComponent_Type tests component type identification.
func TestExpressionComponent_Type(t *testing.T) {
	comp := ExpressionComponent{}
	if comp.Type() != "expression" {
		t.Errorf("Expected type 'expression', got %q", comp.Type())
	}
}

// TestSimpleAnimationSequence tests animation sequence implementation.
func TestSimpleAnimationSequence(t *testing.T) {
	tests := []struct {
		name       string
		frameCount int
		frameTime  float64
		loop       bool
	}{
		{"short non-loop", 4, 0.1, false},
		{"long loop", 16, 0.15, true},
		{"single frame", 1, 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := NewSimpleAnimationSequence(tt.frameCount, tt.frameTime, tt.loop)

			if seq.GetFrameCount() != tt.frameCount {
				t.Errorf("Expected %d frames, got %d", tt.frameCount, seq.GetFrameCount())
			}

			if seq.GetFrameTime() != tt.frameTime {
				t.Errorf("Expected frame time %f, got %f", tt.frameTime, seq.GetFrameTime())
			}

			if seq.ShouldLoop() != tt.loop {
				t.Errorf("Expected loop %t, got %t", tt.loop, seq.ShouldLoop())
			}

			// Test that interface methods work correctly
			var _ AnimationSequence = seq
		})
	}
}

// TestBaseExpression tests expression creation and properties.
func TestBaseExpression(t *testing.T) {
	tests := []struct {
		expType        ExpressionType
		expectedSound  string
		shouldHaveAnim bool
		isInfinite     bool
	}{
		{ExpressionWave, "wave", true, false},
		{ExpressionCheer, "cheer", true, false},
		{ExpressionDance, "music_note", true, false},
		{ExpressionLaugh, "laugh", true, false},
		{ExpressionCry, "sob", true, false},
		{ExpressionSit, "", true, true},
		{ExpressionPoint, "", true, false},
		{ExpressionSalute, "", true, false},
		{ExpressionShrug, "", true, false},
		{ExpressionThumbsUp, "approval", true, false},
		{ExpressionFacepalm, "thud", true, false},
		{ExpressionSleep, "snore", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.expType.String(), func(t *testing.T) {
			expr := NewBaseExpression(tt.expType)

			// Test Expression interface methods
			var _ Expression = expr

			// Check sound effect
			if expr.GetSoundEffect() != tt.expectedSound {
				t.Errorf("Expected sound %q, got %q", tt.expectedSound, expr.GetSoundEffect())
			}

			// Check animation
			anim := expr.GetAnimation()
			if anim == nil && tt.shouldHaveAnim {
				t.Error("Expected animation to be set")
			}

			if anim != nil {
				if anim.GetFrameCount() <= 0 {
					t.Error("Expected positive frame count")
				}
				if anim.GetFrameTime() <= 0 {
					t.Error("Expected positive frame time")
				}
				// Call ShouldLoop to test interface
				_ = anim.ShouldLoop()
			}

			// Check duration
			duration := expr.GetDuration()
			if tt.isInfinite {
				if !math.IsInf(duration, 1) {
					t.Errorf("Expected infinite duration, got %f", duration)
				}
			} else {
				if duration <= 0 || math.IsInf(duration, 1) {
					t.Errorf("Expected finite positive duration, got %f", duration)
				}
			}
		})
	}
}

// TestExpressionSystem_NewExpressionSystem tests system initialization.
func TestExpressionSystem_NewExpressionSystem(t *testing.T) {
	world := NewWorld()

	// Test with nil audio manager
	sys1 := NewExpressionSystem(world, nil)
	if sys1 == nil {
		t.Fatal("NewExpressionSystem returned nil")
	}
	if sys1.world != world {
		t.Error("Expected world to be set")
	}
	if sys1.expressionDefs == nil {
		t.Error("Expected expression definitions to be initialized")
	}
	if len(sys1.expressionDefs) != 12 {
		t.Errorf("Expected 12 expression definitions, got %d", len(sys1.expressionDefs))
	}

	// Test with audio manager
	audioMgr := NewAudioManager(44100, 12345)
	sys2 := NewExpressionSystem(world, audioMgr)
	if sys2.audioManager != audioMgr {
		t.Error("Expected audio manager to be set")
	}
}

// TestExpressionSystem_TriggerExpression tests expression triggering.
func TestExpressionSystem_TriggerExpression(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Create test entity
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// First trigger should succeed
	success := sys.TriggerExpression(entity.ID, ExpressionWave)
	if !success {
		t.Error("Expected first trigger to succeed")
	}

	// Verify component was added
	expCompRaw, ok := entity.GetComponent("expression")
	if !ok {
		t.Fatal("Expected expression component to be added")
	}

	expComp := expCompRaw.(*ExpressionComponent)
	if expComp.ActiveExpression != ExpressionWave {
		t.Errorf("Expected Wave expression, got %s", expComp.ActiveExpression.String())
	}
	if expComp.Cooldown <= 0 {
		t.Error("Expected cooldown to be set")
	}
	if expComp.ExpressionTime <= 0 {
		t.Error("Expected expression time to be set")
	}

	// Second trigger immediately should fail (cooldown)
	success = sys.TriggerExpression(entity.ID, ExpressionCheer)
	if success {
		t.Error("Expected second trigger to fail due to cooldown")
	}

	// Verify expression didn't change
	if expComp.ActiveExpression != ExpressionWave {
		t.Error("Expected expression to remain Wave")
	}
}

// TestExpressionSystem_TriggerExpression_InvalidEntity tests error handling.
func TestExpressionSystem_TriggerExpression_InvalidEntity(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Try to trigger on non-existent entity
	success := sys.TriggerExpression(99999, ExpressionWave)
	if success {
		t.Error("Expected trigger on invalid entity to fail")
	}
}

// TestExpressionSystem_Update tests expression timer updates.
func TestExpressionSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Create entity with expression
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	sys.TriggerExpression(entity.ID, ExpressionWave)

	expCompRaw, _ := entity.GetComponent("expression")
	expComp := expCompRaw.(*ExpressionComponent)

	initialTime := expComp.ExpressionTime
	initialCooldown := expComp.Cooldown

	// Update system
	sys.Update(1.0)

	// Verify timers decreased
	if expComp.ExpressionTime >= initialTime {
		t.Error("Expected expression time to decrease")
	}
	if expComp.Cooldown >= initialCooldown {
		t.Error("Expected cooldown to decrease")
	}

	// Update until expression expires
	for i := 0; i < 10; i++ {
		sys.Update(1.0)
	}

	// Verify expression cleared
	if expComp.ExpressionTime != 0 {
		t.Error("Expected expression time to reach 0")
	}
}

// TestExpressionSystem_Update_InfiniteDuration tests infinite duration expressions.
func TestExpressionSystem_Update_InfiniteDuration(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Create entity with sit expression (infinite duration)
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	sys.TriggerExpression(entity.ID, ExpressionSit)

	expCompRaw, _ := entity.GetComponent("expression")
	expComp := expCompRaw.(*ExpressionComponent)

	// Verify infinite duration
	if !math.IsInf(expComp.ExpressionTime, 1) {
		t.Error("Expected infinite expression time for Sit")
	}

	// Update multiple times
	for i := 0; i < 10; i++ {
		sys.Update(1.0)
	}

	// Verify expression still active (infinite)
	if !math.IsInf(expComp.ExpressionTime, 1) {
		t.Error("Expected expression time to remain infinite")
	}
}

// TestExpressionSystem_CancelExpression tests expression cancellation.
func TestExpressionSystem_CancelExpression(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Create entity with infinite expression
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	sys.TriggerExpression(entity.ID, ExpressionSleep)

	expCompRaw, _ := entity.GetComponent("expression")
	expComp := expCompRaw.(*ExpressionComponent)

	// Verify expression is active
	if !math.IsInf(expComp.ExpressionTime, 1) {
		t.Error("Expected infinite expression time")
	}

	// Cancel expression
	success := sys.CancelExpression(entity.ID)
	if !success {
		t.Error("Expected cancel to succeed")
	}

	// Verify expression cleared
	if expComp.ExpressionTime != 0 {
		t.Error("Expected expression time to be 0 after cancel")
	}

	// Test cancel on invalid entity
	success = sys.CancelExpression(99999)
	if success {
		t.Error("Expected cancel on invalid entity to fail")
	}
}

// TestExpressionSystem_IsOnCooldown tests cooldown checking.
func TestExpressionSystem_IsOnCooldown(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Should not be on cooldown initially
	if sys.IsOnCooldown(entity.ID) {
		t.Error("Expected entity to not be on cooldown initially")
	}

	// Trigger expression
	sys.TriggerExpression(entity.ID, ExpressionWave)

	// Should be on cooldown now
	if !sys.IsOnCooldown(entity.ID) {
		t.Error("Expected entity to be on cooldown after trigger")
	}

	// Update until cooldown expires
	for i := 0; i < 10; i++ {
		sys.Update(1.0)
	}

	// Should not be on cooldown anymore
	if sys.IsOnCooldown(entity.ID) {
		t.Error("Expected cooldown to expire")
	}
}

// TestExpressionSystem_GetActiveExpression tests active expression retrieval.
func TestExpressionSystem_GetActiveExpression(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Should return nil when no expression active
	active := sys.GetActiveExpression(entity.ID)
	if active != nil {
		t.Error("Expected nil active expression initially")
	}

	// Trigger expression
	sys.TriggerExpression(entity.ID, ExpressionDance)

	// Should return active expression
	active = sys.GetActiveExpression(entity.ID)
	if active == nil {
		t.Fatal("Expected active expression to be set")
	}
	if *active != ExpressionDance {
		t.Errorf("Expected Dance, got %s", active.String())
	}

	// Update until expression expires
	for i := 0; i < 10; i++ {
		sys.Update(1.0)
	}

	// Should return nil after expiration
	active = sys.GetActiveExpression(entity.ID)
	if active != nil {
		t.Error("Expected nil active expression after expiration")
	}
}

// TestExpressionSystem_GetDuration tests duration retrieval.
func TestExpressionSystem_GetDuration(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	tests := []struct {
		expType    ExpressionType
		isInfinite bool
	}{
		{ExpressionWave, false},
		{ExpressionDance, false},
		{ExpressionSit, true},
		{ExpressionSleep, true},
	}

	for _, tt := range tests {
		t.Run(tt.expType.String(), func(t *testing.T) {
			duration := sys.GetDuration(tt.expType)
			if tt.isInfinite {
				if !math.IsInf(duration, 1) {
					t.Errorf("Expected infinite duration, got %f", duration)
				}
			} else {
				if duration <= 0 {
					t.Errorf("Expected positive duration, got %f", duration)
				}
			}
		})
	}
}

// TestExpressionSystem_CooldownPreventsSpam tests spam prevention.
func TestExpressionSystem_CooldownPreventsSpam(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Trigger first expression
	success := sys.TriggerExpression(entity.ID, ExpressionWave)
	if !success {
		t.Fatal("Expected first trigger to succeed")
	}

	// Try to trigger multiple times rapidly
	successCount := 0
	for i := 0; i < 10; i++ {
		if sys.TriggerExpression(entity.ID, ExpressionCheer) {
			successCount++
		}
	}

	// Should all fail due to cooldown
	if successCount > 0 {
		t.Errorf("Expected 0 successful rapid triggers, got %d", successCount)
	}

	// Update to clear cooldown (3 seconds + margin)
	sys.Update(3.5)

	// Should succeed now
	success = sys.TriggerExpression(entity.ID, ExpressionLaugh)
	if !success {
		t.Error("Expected trigger to succeed after cooldown")
	}
}

// TestExpressionSystem_MultipleEntities tests multiple entities with expressions.
func TestExpressionSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewExpressionSystem(world, nil)

	// Create multiple entities with different expressions
	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()
	entity3 := world.CreateEntity()
	world.Update(0) // Process entity additions

	sys.TriggerExpression(entity1.ID, ExpressionWave)
	sys.TriggerExpression(entity2.ID, ExpressionDance)
	sys.TriggerExpression(entity3.ID, ExpressionSit)

	// Verify each has correct expression
	active1 := sys.GetActiveExpression(entity1.ID)
	active2 := sys.GetActiveExpression(entity2.ID)
	active3 := sys.GetActiveExpression(entity3.ID)

	if active1 == nil || *active1 != ExpressionWave {
		t.Error("Entity 1 should have Wave expression")
	}
	if active2 == nil || *active2 != ExpressionDance {
		t.Error("Entity 2 should have Dance expression")
	}
	if active3 == nil || *active3 != ExpressionSit {
		t.Error("Entity 3 should have Sit expression")
	}

	// Update system
	sys.Update(1.0)

	// All should still be updating correctly
	if !sys.IsOnCooldown(entity1.ID) {
		t.Error("Entity 1 should still be on cooldown")
	}
	if !sys.IsOnCooldown(entity2.ID) {
		t.Error("Entity 2 should still be on cooldown")
	}
	if !sys.IsOnCooldown(entity3.ID) {
		t.Error("Entity 3 should still be on cooldown")
	}
}
