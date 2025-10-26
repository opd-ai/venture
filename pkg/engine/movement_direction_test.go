// Package engine provides movement direction tests.
// This file tests automatic facing direction updates based on velocity (Phase 3).
package engine

import (
	"math"
	"testing"
)

// TestMovementSystem_DirectionUpdate_CardinalDirections tests facing updates for N/S/E/W movement.
func TestMovementSystem_DirectionUpdate_CardinalDirections(t *testing.T) {
	tests := []struct {
		name      string
		vx, vy    float64
		wantDir   Direction
		wantState AnimationState
	}{
		{"moving right", 5.0, 0.0, DirRight, AnimationStateWalk},
		{"moving left", -5.0, 0.0, DirLeft, AnimationStateWalk},
		{"moving down", 0.0, 5.0, DirDown, AnimationStateWalk},
		{"moving up", 0.0, -5.0, DirUp, AnimationStateWalk},
		{"fast right", 10.0, 0.0, DirRight, AnimationStateRun},
		{"fast left", -10.0, 0.0, DirLeft, AnimationStateRun},
		{"fast down", 0.0, 10.0, DirDown, AnimationStateRun},
		{"fast up", 0.0, -10.0, DirUp, AnimationStateRun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewMovementSystem(8.0) // MaxSpeed 8.0 for walk/run distinction
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{VX: tt.vx, VY: tt.vy})
			entity.AddComponent(NewAnimationComponent(12345))

			world.Update(0)                           // Process pending additions
			system.Update(world.GetEntities(), 0.016) // ~60 FPS

			animComp, _ := entity.GetComponent("animation")
			anim := animComp.(*AnimationComponent)
			if anim.GetFacing() != tt.wantDir {
				t.Errorf("After velocity (%v, %v), facing = %v, want %v",
					tt.vx, tt.vy, anim.GetFacing(), tt.wantDir)
			}
			if anim.CurrentState != tt.wantState {
				t.Errorf("After velocity (%v, %v), state = %v, want %v",
					tt.vx, tt.vy, anim.CurrentState, tt.wantState)
			}
		})
	}
}

// TestMovementSystem_DirectionUpdate_DiagonalMovement tests horizontal priority for diagonal movement.
func TestMovementSystem_DirectionUpdate_DiagonalMovement(t *testing.T) {
	tests := []struct {
		name    string
		vx, vy  float64
		wantDir Direction
	}{
		{"diagonal up-right (H>V)", 5.0, 3.0, DirRight},
		{"diagonal down-right (H>V)", 5.0, 3.0, DirRight},
		{"diagonal up-left (H>V)", -5.0, 3.0, DirLeft},
		{"diagonal down-left (H>V)", -5.0, -3.0, DirLeft},
		{"diagonal up-right (V>H)", 3.0, 5.0, DirDown}, // vertical wins
		{"diagonal up-right (V>H)", 3.0, -5.0, DirUp},  // vertical wins
		{"perfect diagonal (H priority)", 5.0, 5.0, DirRight},
		{"perfect diagonal negative", -5.0, -5.0, DirLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewMovementSystem(0) // No speed limit
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{VX: tt.vx, VY: tt.vy})
			entity.AddComponent(NewAnimationComponent(12345))

			world.Update(0)
			system.Update(world.GetEntities(), 0.016)

			animComp, _ := entity.GetComponent("animation")
			anim := animComp.(*AnimationComponent)
			if anim.GetFacing() != tt.wantDir {
				t.Errorf("Diagonal velocity (%v, %v): facing = %v, want %v",
					tt.vx, tt.vy, anim.GetFacing(), tt.wantDir)
			}
		})
	}
}

// TestMovementSystem_DirectionUpdate_JitterFiltering tests 0.1 threshold prevents small velocity changes.
func TestMovementSystem_DirectionUpdate_JitterFiltering(t *testing.T) {
	tests := []struct {
		name             string
		vx, vy           float64
		initialDir       Direction
		shouldUpdateDir  bool
		shouldUpdateAnim bool
	}{
		{"below threshold X", 0.05, 0.0, DirDown, false, false},
		{"below threshold Y", 0.0, 0.05, DirDown, false, false},
		{"below threshold both", 0.05, 0.05, DirDown, false, false},
		{"at threshold X", 0.1, 0.0, DirDown, false, false}, // Exactly 0.1 is not > 0.1
		{"above threshold X", 0.11, 0.0, DirDown, true, true},
		{"above threshold Y", 0.0, 0.11, DirDown, true, true},
		{"negative below threshold", -0.05, -0.05, DirRight, false, false},
		{"negative above threshold", -0.15, 0.0, DirRight, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewMovementSystem(0)
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{VX: tt.vx, VY: tt.vy})
			anim := NewAnimationComponent(12345)
			anim.SetFacing(tt.initialDir)
			anim.SetState(AnimationStateIdle)
			entity.AddComponent(anim)

			world.Update(0)
			system.Update(world.GetEntities(), 0.016)

			animComp, _ := entity.GetComponent("animation")
			anim = animComp.(*AnimationComponent)

			// Calculate expected direction if update should happen
			absVX := math.Abs(tt.vx)
			absVY := math.Abs(tt.vy)
			var expectedDir Direction
			if absVX > absVY {
				if tt.vx > 0 {
					expectedDir = DirRight
				} else {
					expectedDir = DirLeft
				}
			} else {
				if tt.vy > 0 {
					expectedDir = DirDown
				} else {
					expectedDir = DirUp
				}
			}

			if tt.shouldUpdateDir {
				if anim.GetFacing() != expectedDir {
					t.Errorf("Velocity (%v, %v) above threshold: facing = %v, want %v",
						tt.vx, tt.vy, anim.GetFacing(), expectedDir)
				}
			} else {
				if anim.GetFacing() != tt.initialDir {
					t.Errorf("Velocity (%v, %v) below threshold: facing changed from %v to %v",
						tt.vx, tt.vy, tt.initialDir, anim.GetFacing())
				}
			}

			// Check animation state update
			if tt.shouldUpdateAnim {
				if anim.CurrentState == AnimationStateIdle {
					t.Errorf("Velocity (%v, %v) above threshold: state still idle", tt.vx, tt.vy)
				}
			} else {
				if anim.CurrentState != AnimationStateIdle {
					t.Errorf("Velocity (%v, %v) below threshold: state changed to %v",
						tt.vx, tt.vy, anim.CurrentState)
				}
			}
		})
	}
}

// TestMovementSystem_DirectionUpdate_StationaryPreservesFacing tests facing persists when stopped.
func TestMovementSystem_DirectionUpdate_StationaryPreservesFacing(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, initialDir := range directions {
		t.Run(initialDir.String(), func(t *testing.T) {
			world := NewWorld()
			system := NewMovementSystem(0)
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{VX: 0, VY: 0}) // Stationary
			anim := NewAnimationComponent(12345)
			anim.SetFacing(initialDir)
			anim.SetState(AnimationStateWalk) // Was walking
			entity.AddComponent(anim)

			world.Update(0)
			system.Update(world.GetEntities(), 0.016)

			animComp, _ := entity.GetComponent("animation")
			anim = animComp.(*AnimationComponent)

			// Should preserve facing
			if anim.GetFacing() != initialDir {
				t.Errorf("Stationary entity facing changed from %v to %v",
					initialDir, anim.GetFacing())
			}

			// Should transition to idle
			if anim.CurrentState != AnimationStateIdle {
				t.Errorf("Stationary entity state = %v, want %v",
					anim.CurrentState, AnimationStateIdle)
			}
		})
	}
}

// TestMovementSystem_DirectionUpdate_MovementResume tests facing updates after stopping.
func TestMovementSystem_DirectionUpdate_MovementResume(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	vel := &VelocityComponent{VX: 5.0, VY: 0.0} // Moving right
	entity.AddComponent(vel)
	anim := NewAnimationComponent(12345)
	anim.SetFacing(DirUp) // Initially facing up
	entity.AddComponent(anim)

	world.Update(0)

	// First update: moving right
	system.Update(world.GetEntities(), 0.016)
	animComp, _ := entity.GetComponent("animation")
	anim = animComp.(*AnimationComponent)
	if anim.GetFacing() != DirRight {
		t.Errorf("After moving right, facing = %v, want %v", anim.GetFacing(), DirRight)
	}

	// Stop moving
	vel.VX = 0
	vel.VY = 0
	system.Update(world.GetEntities(), 0.016)
	animComp, _ = entity.GetComponent("animation")
	anim = animComp.(*AnimationComponent)
	if anim.GetFacing() != DirRight {
		t.Errorf("After stopping, facing changed from right to %v", anim.GetFacing())
	}

	// Resume moving left
	vel.VX = -5.0
	system.Update(world.GetEntities(), 0.016)
	animComp, _ = entity.GetComponent("animation")
	anim = animComp.(*AnimationComponent)
	if anim.GetFacing() != DirLeft {
		t.Errorf("After moving left, facing = %v, want %v", anim.GetFacing(), DirLeft)
	}
}

// TestMovementSystem_DirectionUpdate_NoAnimationComponent tests entities without animation.
func TestMovementSystem_DirectionUpdate_NoAnimationComponent(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 5.0, VY: 0.0})
	// No animation component

	world.Update(0)

	// Should not panic
	system.Update(world.GetEntities(), 0.016)

	// Position should still update
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 0 {
		t.Errorf("Position not updated without animation component: X = %v", pos.X)
	}
}

// TestMovementSystem_DirectionUpdate_ActionStates tests facing doesn't update during attacks.
func TestMovementSystem_DirectionUpdate_ActionStates(t *testing.T) {
	actionStates := []AnimationState{
		AnimationStateAttack,
		AnimationStateHit,
		AnimationStateDeath,
		AnimationStateCast,
	}

	for _, state := range actionStates {
		t.Run(state.String(), func(t *testing.T) {
			world := NewWorld()
			system := NewMovementSystem(0)
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{VX: 5.0, VY: 0.0}) // Moving right
			anim := NewAnimationComponent(12345)
			anim.SetFacing(DirUp) // Facing up
			anim.SetState(state)  // In action state
			entity.AddComponent(anim)

			world.Update(0)
			system.Update(world.GetEntities(), 0.016)

			animComp, _ := entity.GetComponent("animation")
			anim = animComp.(*AnimationComponent)

			// Facing should NOT change during action states
			if anim.GetFacing() != DirUp {
				t.Errorf("During %v, facing changed from up to %v",
					state, anim.GetFacing())
			}

			// State should NOT change during action states
			if anim.CurrentState != state {
				t.Errorf("During %v, state changed to %v", state, anim.CurrentState)
			}
		})
	}
}

// TestMovementSystem_DirectionUpdate_MultipleEntities tests direction updates work for multiple entities.
func TestMovementSystem_DirectionUpdate_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0)

	// Create 4 entities moving in different directions
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity1.AddComponent(&VelocityComponent{VX: 5.0, VY: 0.0}) // Right
	entity1.AddComponent(NewAnimationComponent(1))

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity2.AddComponent(&VelocityComponent{VX: -5.0, VY: 0.0}) // Left
	entity2.AddComponent(NewAnimationComponent(2))

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity3.AddComponent(&VelocityComponent{VX: 0.0, VY: 5.0}) // Down
	entity3.AddComponent(NewAnimationComponent(3))

	entity4 := world.CreateEntity()
	entity4.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity4.AddComponent(&VelocityComponent{VX: 0.0, VY: -5.0}) // Up
	entity4.AddComponent(NewAnimationComponent(4))

	world.Update(0)
	system.Update(world.GetEntities(), 0.016)

	// Verify each entity has correct facing
	entities := []*Entity{entity1, entity2, entity3, entity4}
	expectedDirs := []Direction{DirRight, DirLeft, DirDown, DirUp}
	for i, entity := range entities {
		animComp, _ := entity.GetComponent("animation")
		anim := animComp.(*AnimationComponent)
		if anim.GetFacing() != expectedDirs[i] {
			t.Errorf("Entity %d facing = %v, want %v",
				i+1, anim.GetFacing(), expectedDirs[i])
		}
	}
}

// TestMovementSystem_DirectionUpdate_FrictionPreservesFacing tests facing persists as friction slows entity.
func TestMovementSystem_DirectionUpdate_FrictionPreservesFacing(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 5.0, VY: 0.0}) // Moving right
	entity.AddComponent(&FrictionComponent{Coefficient: 0.5}) // High friction
	anim := NewAnimationComponent(12345)
	entity.AddComponent(anim)

	world.Update(0)

	// First update: moving right, facing right
	system.Update(world.GetEntities(), 0.016)
	animComp, _ := entity.GetComponent("animation")
	anim = animComp.(*AnimationComponent)
	if anim.GetFacing() != DirRight {
		t.Fatalf("Initial facing = %v, want right", anim.GetFacing())
	}

	// Continue updating as friction slows entity
	for i := 0; i < 10; i++ {
		system.Update(world.GetEntities(), 0.016)
		animComp, _ = entity.GetComponent("animation")
		anim = animComp.(*AnimationComponent)

		velComp, _ := entity.GetComponent("velocity")
		vel := velComp.(*VelocityComponent)
		if vel.VX > 0.1 {
			// Still moving right (above threshold)
			if anim.GetFacing() != DirRight {
				t.Errorf("Iteration %d: while slowing (VX=%v), facing changed to %v",
					i, vel.VX, anim.GetFacing())
			}
		} else if vel.VX > 0 {
			// Below threshold but still positive
			if anim.GetFacing() != DirRight {
				t.Errorf("Iteration %d: below threshold (VX=%v), facing changed to %v",
					i, vel.VX, anim.GetFacing())
			}
		} else {
			// Stopped by friction
			if anim.GetFacing() != DirRight {
				t.Errorf("Iteration %d: after stopping (VX=%v), facing changed to %v",
					i, vel.VX, anim.GetFacing())
			}
			break
		}
	}
}

// BenchmarkMovementSystem_DirectionUpdate benchmarks direction update performance.
func BenchmarkMovementSystem_DirectionUpdate(b *testing.B) {
	world := NewWorld()
	system := NewMovementSystem(0)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	vel := &VelocityComponent{VX: 5.0, VY: 3.0}
	entity.AddComponent(vel)
	entity.AddComponent(NewAnimationComponent(12345))
	world.Update(0)

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate velocity to trigger direction changes
		if i%4 == 0 {
			vel.VX, vel.VY = 5.0, 0.0
		} else if i%4 == 1 {
			vel.VX, vel.VY = 0.0, 5.0
		} else if i%4 == 2 {
			vel.VX, vel.VY = -5.0, 0.0
		} else {
			vel.VX, vel.VY = 0.0, -5.0
		}
		system.Update(entities, 0.016)
	}
}
