// Package engine provides tests for the Layer component.
package engine

import (
	"testing"
)

func TestLayerComponent_Type(t *testing.T) {
	l := NewLayerComponent()
	if got := l.Type(); got != "layer" {
		t.Errorf("Type() = %q, want %q", got, "layer")
	}
}

func TestNewLayerComponent(t *testing.T) {
	l := NewLayerComponent()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"CurrentLayer", l.CurrentLayer, 0},
		{"TargetLayer", l.TargetLayer, -1},
		{"TransitionProgress", l.TransitionProgress, 0.0},
		{"CanFly", l.CanFly, false},
		{"CanSwim", l.CanSwim, false},
		{"CanClimb", l.CanClimb, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestNewFlyingLayerComponent(t *testing.T) {
	l := NewFlyingLayerComponent()

	if !l.CanFly {
		t.Error("CanFly should be true for flying component")
	}
	if !l.CanSwim {
		t.Error("CanSwim should be true for flying component")
	}
	if !l.CanClimb {
		t.Error("CanClimb should be true for flying component")
	}
	if l.CurrentLayer != 0 {
		t.Errorf("CurrentLayer = %d, want 0", l.CurrentLayer)
	}
}

func TestNewSwimmingLayerComponent(t *testing.T) {
	l := NewSwimmingLayerComponent()

	if l.CanFly {
		t.Error("CanFly should be false for swimming component")
	}
	if !l.CanSwim {
		t.Error("CanSwim should be true for swimming component")
	}
	if l.CanClimb {
		t.Error("CanClimb should be false for swimming component")
	}
	if l.CurrentLayer != 1 {
		t.Errorf("CurrentLayer = %d, want 1 (water layer)", l.CurrentLayer)
	}
}

func TestLayerComponent_IsTransitioning(t *testing.T) {
	tests := []struct {
		name        string
		targetLayer int
		want        bool
	}{
		{"not transitioning", -1, false},
		{"transitioning to ground", 0, true},
		{"transitioning to water", 1, true},
		{"transitioning to platform", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayerComponent()
			l.TargetLayer = tt.targetLayer
			if got := l.IsTransitioning(); got != tt.want {
				t.Errorf("IsTransitioning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLayerComponent_StartTransition(t *testing.T) {
	l := NewLayerComponent()
	l.StartTransition(2)

	if l.TargetLayer != 2 {
		t.Errorf("TargetLayer = %d, want 2", l.TargetLayer)
	}
	if l.TransitionProgress != 0.0 {
		t.Errorf("TransitionProgress = %f, want 0.0", l.TransitionProgress)
	}
	if !l.IsTransitioning() {
		t.Error("IsTransitioning() should be true after StartTransition")
	}
}

func TestLayerComponent_UpdateTransition(t *testing.T) {
	tests := []struct {
		name          string
		initialLayer  int
		targetLayer   int
		progressDelta float64
		wantComplete  bool
		wantCurrent   int
		wantTarget    int
		wantProgress  float64
	}{
		{
			name:          "partial transition",
			initialLayer:  0,
			targetLayer:   2,
			progressDelta: 0.3,
			wantComplete:  false,
			wantCurrent:   0,
			wantTarget:    2,
			wantProgress:  0.3,
		},
		{
			name:          "complete transition",
			initialLayer:  0,
			targetLayer:   2,
			progressDelta: 1.0,
			wantComplete:  true,
			wantCurrent:   2,
			wantTarget:    -1,
			wantProgress:  0.0,
		},
		{
			name:          "over-complete transition",
			initialLayer:  0,
			targetLayer:   1,
			progressDelta: 1.5,
			wantComplete:  true,
			wantCurrent:   1,
			wantTarget:    -1,
			wantProgress:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayerComponent()
			l.CurrentLayer = tt.initialLayer
			l.StartTransition(tt.targetLayer)

			complete := l.UpdateTransition(tt.progressDelta)

			if complete != tt.wantComplete {
				t.Errorf("UpdateTransition() complete = %v, want %v", complete, tt.wantComplete)
			}
			if l.CurrentLayer != tt.wantCurrent {
				t.Errorf("CurrentLayer = %d, want %d", l.CurrentLayer, tt.wantCurrent)
			}
			if l.TargetLayer != tt.wantTarget {
				t.Errorf("TargetLayer = %d, want %d", l.TargetLayer, tt.wantTarget)
			}
			if l.TransitionProgress != tt.wantProgress {
				t.Errorf("TransitionProgress = %f, want %f", l.TransitionProgress, tt.wantProgress)
			}
		})
	}
}

func TestLayerComponent_UpdateTransition_NotTransitioning(t *testing.T) {
	l := NewLayerComponent()
	// Don't start transition

	complete := l.UpdateTransition(0.5)
	if complete {
		t.Error("UpdateTransition() should return false when not transitioning")
	}
}

func TestLayerComponent_CancelTransition(t *testing.T) {
	l := NewLayerComponent()
	l.StartTransition(2)
	l.UpdateTransition(0.5)

	// Cancel mid-transition
	l.CancelTransition()

	if l.TargetLayer != -1 {
		t.Errorf("TargetLayer = %d, want -1 after cancel", l.TargetLayer)
	}
	if l.TransitionProgress != 0.0 {
		t.Errorf("TransitionProgress = %f, want 0.0 after cancel", l.TransitionProgress)
	}
	if l.CurrentLayer != 0 {
		t.Errorf("CurrentLayer = %d, should remain 0 after cancel", l.CurrentLayer)
	}
}

func TestLayerComponent_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name        string
		component   LayerComponent
		targetLayer int
		want        bool
	}{
		// Flying entities
		{
			name:        "flying to ground",
			component:   NewFlyingLayerComponent(),
			targetLayer: 0,
			want:        true,
		},
		{
			name:        "flying to water",
			component:   NewFlyingLayerComponent(),
			targetLayer: 1,
			want:        true,
		},
		{
			name:        "flying to platform",
			component:   NewFlyingLayerComponent(),
			targetLayer: 2,
			want:        true,
		},
		// Swimming entities
		{
			name:        "swimming to ground",
			component:   NewSwimmingLayerComponent(),
			targetLayer: 0,
			want:        true,
		},
		{
			name:        "swimming to water",
			component:   NewSwimmingLayerComponent(),
			targetLayer: 1,
			want:        true,
		},
		{
			name:        "swimming to platform - cannot",
			component:   NewSwimmingLayerComponent(),
			targetLayer: 2,
			want:        false,
		},
		// Normal entities
		{
			name:        "normal to ground",
			component:   NewLayerComponent(),
			targetLayer: 0,
			want:        false, // Already on ground, but rule allows return from platform
		},
		{
			name:        "normal to water - cannot",
			component:   NewLayerComponent(),
			targetLayer: 1,
			want:        false,
		},
		{
			name:        "normal to platform - can climb",
			component:   NewLayerComponent(),
			targetLayer: 2,
			want:        true,
		},
		// Platform to ground
		{
			name: "platform to ground",
			component: LayerComponent{
				CurrentLayer: 2,
				TargetLayer:  -1,
				CanClimb:     true,
			},
			targetLayer: 0,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.component.CanTransitionTo(tt.targetLayer); got != tt.want {
				t.Errorf("CanTransitionTo(%d) = %v, want %v", tt.targetLayer, got, tt.want)
			}
		})
	}
}

func TestLayerComponent_GetEffectiveLayer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*LayerComponent)
		wantLayer int
	}{
		{
			name: "not transitioning",
			setup: func(l *LayerComponent) {
				l.CurrentLayer = 0
			},
			wantLayer: 0,
		},
		{
			name: "early transition - use current",
			setup: func(l *LayerComponent) {
				l.CurrentLayer = 0
				l.StartTransition(2)
				l.UpdateTransition(0.3)
			},
			wantLayer: 0,
		},
		{
			name: "halfway transition - use current",
			setup: func(l *LayerComponent) {
				l.CurrentLayer = 0
				l.StartTransition(2)
				l.UpdateTransition(0.5)
			},
			wantLayer: 0,
		},
		{
			name: "late transition - use target",
			setup: func(l *LayerComponent) {
				l.CurrentLayer = 0
				l.StartTransition(2)
				l.UpdateTransition(0.7)
			},
			wantLayer: 2,
		},
		{
			name: "almost complete - use target",
			setup: func(l *LayerComponent) {
				l.CurrentLayer = 0
				l.StartTransition(2)
				l.UpdateTransition(0.9)
			},
			wantLayer: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayerComponent()
			tt.setup(&l)
			if got := l.GetEffectiveLayer(); got != tt.wantLayer {
				t.Errorf("GetEffectiveLayer() = %d, want %d", got, tt.wantLayer)
			}
		})
	}
}

func TestOnSameLayer(t *testing.T) {
	tests := []struct {
		name string
		l1   *LayerComponent
		l2   *LayerComponent
		want bool
	}{
		{
			name: "both on ground",
			l1:   &LayerComponent{CurrentLayer: 0, TargetLayer: -1},
			l2:   &LayerComponent{CurrentLayer: 0, TargetLayer: -1},
			want: true,
		},
		{
			name: "different layers",
			l1:   &LayerComponent{CurrentLayer: 0, TargetLayer: -1},
			l2:   &LayerComponent{CurrentLayer: 2, TargetLayer: -1},
			want: false,
		},
		{
			name: "nil component - assume ground",
			l1:   nil,
			l2:   &LayerComponent{CurrentLayer: 0, TargetLayer: -1},
			want: true,
		},
		{
			name: "both nil - both ground",
			l1:   nil,
			l2:   nil,
			want: true,
		},
		{
			name: "nil vs platform",
			l1:   nil,
			l2:   &LayerComponent{CurrentLayer: 2, TargetLayer: -1},
			want: false,
		},
		{
			name: "one transitioning late - same effective",
			l1:   &LayerComponent{CurrentLayer: 0, TargetLayer: 2, TransitionProgress: 0.7},
			l2:   &LayerComponent{CurrentLayer: 2, TargetLayer: -1},
			want: true,
		},
		{
			name: "one transitioning early - different effective",
			l1:   &LayerComponent{CurrentLayer: 0, TargetLayer: 2, TransitionProgress: 0.3},
			l2:   &LayerComponent{CurrentLayer: 2, TargetLayer: -1},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OnSameLayer(tt.l1, tt.l2); got != tt.want {
				t.Errorf("OnSameLayer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLayerComponent_Clone(t *testing.T) {
	original := LayerComponent{
		CurrentLayer:       1,
		TargetLayer:        2,
		TransitionProgress: 0.6,
		CanFly:             true,
		CanSwim:            true,
		CanClimb:           false,
	}

	clone := original.Clone()

	// Verify values match
	if clone.CurrentLayer != original.CurrentLayer {
		t.Errorf("Clone CurrentLayer = %d, want %d", clone.CurrentLayer, original.CurrentLayer)
	}
	if clone.TargetLayer != original.TargetLayer {
		t.Errorf("Clone TargetLayer = %d, want %d", clone.TargetLayer, original.TargetLayer)
	}
	if clone.TransitionProgress != original.TransitionProgress {
		t.Errorf("Clone TransitionProgress = %f, want %f", clone.TransitionProgress, original.TransitionProgress)
	}
	if clone.CanFly != original.CanFly {
		t.Errorf("Clone CanFly = %v, want %v", clone.CanFly, original.CanFly)
	}
	if clone.CanSwim != original.CanSwim {
		t.Errorf("Clone CanSwim = %v, want %v", clone.CanSwim, original.CanSwim)
	}
	if clone.CanClimb != original.CanClimb {
		t.Errorf("Clone CanClimb = %v, want %v", clone.CanClimb, original.CanClimb)
	}

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clone.CurrentLayer = 99
	if original.CurrentLayer == 99 {
		t.Error("Clone modification affected original")
	}
}

// BenchmarkLayerComponent_GetEffectiveLayer benchmarks the hot path.
func BenchmarkLayerComponent_GetEffectiveLayer(b *testing.B) {
	l := NewLayerComponent()
	l.StartTransition(2)
	l.UpdateTransition(0.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.GetEffectiveLayer()
	}
}

// BenchmarkOnSameLayer benchmarks layer comparison (hot path in collision detection).
func BenchmarkOnSameLayer(b *testing.B) {
	l1 := &LayerComponent{CurrentLayer: 0, TargetLayer: -1}
	l2 := &LayerComponent{CurrentLayer: 0, TargetLayer: -1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = OnSameLayer(l1, l2)
	}
}
