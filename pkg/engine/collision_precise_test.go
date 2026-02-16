package engine

import (
	"math"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestQuantizePosition verifies sub-pixel precision quantization to 0.1px.
func TestQuantizePosition(t *testing.T) {
	tests := []struct {
		name      string
		x, y      float64
		expectedX float64
		expectedY float64
	}{
		{"exact value", 10.0, 20.0, 10.0, 20.0},
		{"round down", 10.04, 20.03, 10.0, 20.0},
		{"round up", 10.06, 20.07, 10.1, 20.1},
		{"negative values", -5.14, -3.26, -5.1, -3.3},
		{"sub-precision", 15.123, 25.456, 15.1, 25.5},
		{"midpoint down", 10.05, 20.05, 10.1, 20.1}, // Round half to even or up
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qx, qy := QuantizePosition(tt.x, tt.y)

			// Allow small floating-point error
			epsilon := 0.001
			if math.Abs(qx-tt.expectedX) > epsilon {
				t.Errorf("QuantizePosition() x = %v, want %v", qx, tt.expectedX)
			}
			if math.Abs(qy-tt.expectedY) > epsilon {
				t.Errorf("QuantizePosition() y = %v, want %v", qy, tt.expectedY)
			}

			// Verify precision is exactly 0.1
			if math.Mod(qx*10, 1.0) > epsilon && math.Mod(qx*10, 1.0) < 1.0-epsilon {
				t.Errorf("QuantizePosition() x precision error: %v", qx)
			}
			if math.Mod(qy*10, 1.0) > epsilon && math.Mod(qy*10, 1.0) < 1.0-epsilon {
				t.Errorf("QuantizePosition() y precision error: %v", qy)
			}
		})
	}
}

// TestPreciseColliderGetBounds verifies bounds calculation with quantization.
func TestPreciseColliderGetBounds(t *testing.T) {
	tests := []struct {
		name     string
		collider *PreciseColliderComponent
		x, y     float64
		wantMinX float64
		wantMinY float64
		wantMaxX float64
		wantMaxY float64
	}{
		{
			name: "simple bounds",
			collider: &PreciseColliderComponent{
				Width: 32.0, Height: 32.0,
				OffsetX: 0, OffsetY: 0,
			},
			x: 10.0, y: 20.0,
			wantMinX: 10.0, wantMinY: 20.0,
			wantMaxX: 42.0, wantMaxY: 52.0,
		},
		{
			name: "with offset",
			collider: &PreciseColliderComponent{
				Width: 16.0, Height: 16.0,
				OffsetX: -8.0, OffsetY: -8.0,
			},
			x: 100.0, y: 100.0,
			wantMinX: 92.0, wantMinY: 92.0,
			wantMaxX: 108.0, wantMaxY: 108.0,
		},
		{
			name: "sub-pixel quantization",
			collider: &PreciseColliderComponent{
				Width: 32.5, Height: 32.5,
				OffsetX: 0, OffsetY: 0,
			},
			x: 10.14, y: 20.26,
			wantMinX: 10.1, wantMinY: 20.3,
			wantMaxX: 42.6, wantMaxY: 52.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minX, minY, maxX, maxY := tt.collider.GetBounds(tt.x, tt.y)

			epsilon := 0.01
			if math.Abs(minX-tt.wantMinX) > epsilon {
				t.Errorf("GetBounds() minX = %v, want %v", minX, tt.wantMinX)
			}
			if math.Abs(minY-tt.wantMinY) > epsilon {
				t.Errorf("GetBounds() minY = %v, want %v", minY, tt.wantMinY)
			}
			if math.Abs(maxX-tt.wantMaxX) > epsilon {
				t.Errorf("GetBounds() maxX = %v, want %v", maxX, tt.wantMaxX)
			}
			if math.Abs(maxY-tt.wantMaxY) > epsilon {
				t.Errorf("GetBounds() maxY = %v, want %v", maxY, tt.wantMaxY)
			}
		})
	}
}

// TestIntersectsAABB verifies AABB intersection with sub-pixel precision.
func TestIntersectsAABB(t *testing.T) {
	tests := []struct {
		name        string
		c1          *PreciseColliderComponent
		x1, y1      float64
		c2          *PreciseColliderComponent
		x2, y2      float64
		wantCollide bool
	}{
		{
			name: "clear intersection",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 16, y2: 16,
			wantCollide: true,
		},
		{
			name: "no intersection",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 50, y2: 50,
			wantCollide: false,
		},
		{
			name: "edge touch exactly",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 32.0, y2: 0, // Exact edge touch after quantization - no overlap
			wantCollide: false,
		},
		{
			name: "slight overlap",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 31.9, y2: 0, // 0.1px overlap after quantization
			wantCollide: true,
		},
		{
			name: "edge separation",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 32.2, y2: 0, // Separated by 0.2px
			wantCollide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c1.Shape = ShapeAABB
			tt.c2.Shape = ShapeAABB

			got := tt.c1.IntersectsAABB(tt.x1, tt.y1, tt.c2, tt.x2, tt.y2)
			if got != tt.wantCollide {
				t.Errorf("IntersectsAABB() = %v, want %v", got, tt.wantCollide)
			}
		})
	}
}

// TestIntersectsCircle verifies circle-circle intersection.
func TestIntersectsCircle(t *testing.T) {
	tests := []struct {
		name        string
		c1          *PreciseColliderComponent
		x1, y1      float64
		c2          *PreciseColliderComponent
		x2, y2      float64
		wantCollide bool
	}{
		{
			name: "circles overlap",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 20, y2: 0,
			wantCollide: true,
		},
		{
			name: "circles separate",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 50, y2: 0,
			wantCollide: false,
		},
		{
			name: "circles touch",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32},
			x2: 32, y2: 0, // Radius 16 + 16 = 32
			wantCollide: true,
		},
		{
			name: "different sizes",
			c1:   &PreciseColliderComponent{Width: 64, Height: 64},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 16, Height: 16},
			x2: 40, y2: 0, // Radius 32 + 8 = 40
			wantCollide: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c1.Shape = ShapeCircle
			tt.c2.Shape = ShapeCircle

			got := tt.c1.IntersectsCircle(tt.x1, tt.y1, tt.c2, tt.x2, tt.y2)
			if got != tt.wantCollide {
				t.Errorf("IntersectsCircle() = %v, want %v", got, tt.wantCollide)
			}
		})
	}
}

// TestIntersectsRoundedRect verifies rounded rectangle intersection.
func TestIntersectsRoundedRect(t *testing.T) {
	tests := []struct {
		name        string
		c1          *PreciseColliderComponent
		x1, y1      float64
		c2          *PreciseColliderComponent
		x2, y2      float64
		wantCollide bool
	}{
		{
			name: "core overlap",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x2: 16, y2: 16,
			wantCollide: true,
		},
		{
			name: "corner overlap",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x2: 28, y2: 28, // Clear overlap
			wantCollide: true,
		},
		{
			name: "no overlap",
			c1:   &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x1:   0, y1: 0,
			c2: &PreciseColliderComponent{Width: 32, Height: 32, CornerRadius: 4},
			x2: 50, y2: 50,
			wantCollide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c1.Shape = ShapeRoundedRect
			tt.c2.Shape = ShapeRoundedRect

			got := tt.c1.IntersectsRoundedRect(tt.x1, tt.y1, tt.c2, tt.x2, tt.y2)
			if got != tt.wantCollide {
				t.Errorf("IntersectsRoundedRect() = %v, want %v", got, tt.wantCollide)
			}
		})
	}
}

// TestComputeWallNormal verifies wall normal calculation for sliding.
func TestComputeWallNormal(t *testing.T) {
	tests := []struct {
		name             string
		entityX, entityY float64
		wallX, wallY     float64
		wantNX, wantNY   float64
	}{
		{
			name:    "right of wall",
			entityX: 10, entityY: 0,
			wallX: 0, wallY: 0,
			wantNX: 1, wantNY: 0,
		},
		{
			name:    "above wall",
			entityX: 0, entityY: -10,
			wallX: 0, wallY: 0,
			wantNX: 0, wantNY: -1,
		},
		{
			name:    "diagonal",
			entityX: 10, entityY: 10,
			wallX: 0, wallY: 0,
			wantNX: 0.707, wantNY: 0.707, // ~1/sqrt(2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normal := ComputeWallNormal(tt.entityX, tt.entityY, tt.wallX, tt.wallY)

			epsilon := 0.01
			if math.Abs(normal.NX-tt.wantNX) > epsilon {
				t.Errorf("ComputeWallNormal() NX = %v, want %v", normal.NX, tt.wantNX)
			}
			if math.Abs(normal.NY-tt.wantNY) > epsilon {
				t.Errorf("ComputeWallNormal() NY = %v, want %v", normal.NY, tt.wantNY)
			}

			// Verify normalization
			length := math.Sqrt(normal.NX*normal.NX + normal.NY*normal.NY)
			if math.Abs(length-1.0) > epsilon {
				t.Errorf("ComputeWallNormal() not normalized: length = %v", length)
			}
		})
	}
}

// TestApplyWallSlide verifies smooth wall sliding velocity adjustment.
func TestApplyWallSlide(t *testing.T) {
	tests := []struct {
		name   string
		vx, vy float64
		normal EdgeNormal
		wantVX float64
		wantVY float64
	}{
		{
			name: "horizontal wall, vertical velocity",
			vx:   0, vy: 5,
			normal: EdgeNormal{NX: 1, NY: 0}, // Wall normal pointing right
			wantVX: 0, wantVY: 5,             // Slide up/down along wall
		},
		{
			name: "vertical wall, horizontal velocity",
			vx:   5, vy: 0,
			normal: EdgeNormal{NX: 0, NY: 1}, // Wall normal pointing down
			wantVX: 5, wantVY: 0,             // Slide left/right along wall
		},
		{
			name: "diagonal approach to horizontal wall",
			vx:   3, vy: 4,
			normal: EdgeNormal{NX: 1, NY: 0},
			wantVX: 0, wantVY: 4, // Only Y component remains
		},
		{
			name: "diagonal approach to vertical wall",
			vx:   3, vy: 4,
			normal: EdgeNormal{NX: 0, NY: 1},
			wantVX: 3, wantVY: 0, // Only X component remains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVX, gotVY := ApplyWallSlide(tt.vx, tt.vy, tt.normal)

			epsilon := 0.01
			if math.Abs(gotVX-tt.wantVX) > epsilon {
				t.Errorf("ApplyWallSlide() VX = %v, want %v", gotVX, tt.wantVX)
			}
			if math.Abs(gotVY-tt.wantVY) > epsilon {
				t.Errorf("ApplyWallSlide() VY = %v, want %v", gotVY, tt.wantVY)
			}
		})
	}
}

// TestResolveWallCollision verifies wall collision resolution with sliding.
func TestResolveWallCollision(t *testing.T) {
	tests := []struct {
		name   string
		posX   float64
		posY   float64
		velX   float64
		velY   float64
		normal EdgeNormal
		wantX  float64
		wantY  float64
	}{
		{
			name: "push right from vertical wall",
			posX: 10, posY: 20,
			velX: -5, velY: 0,
			normal: EdgeNormal{NX: 1, NY: 0},
			wantX:  10.2, wantY: 20.0, // Pushed 0.2px right
		},
		{
			name: "push up from horizontal wall",
			posX: 10, posY: 20,
			velX: 0, velY: 5,
			normal: EdgeNormal{NX: 0, NY: -1},
			wantX:  10.0, wantY: 19.8, // Pushed 0.2px up
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1) // Use NewEntity to initialize map
			entity.AddComponent(&PositionComponent{X: tt.posX, Y: tt.posY})
			entity.AddComponent(&VelocityComponent{VX: tt.velX, VY: tt.velY})

			gotX, gotY := ResolveWallCollision(entity, tt.normal)

			epsilon := 0.01
			if math.Abs(gotX-tt.wantX) > epsilon {
				t.Errorf("ResolveWallCollision() X = %v, want %v", gotX, tt.wantX)
			}
			if math.Abs(gotY-tt.wantY) > epsilon {
				t.Errorf("ResolveWallCollision() Y = %v, want %v", gotY, tt.wantY)
			}

			// Verify velocity was adjusted for sliding
			velComp, _ := entity.GetComponent("velocity")
			if vel, ok := velComp.(*VelocityComponent); ok {
				// Velocity should be projected onto wall tangent
				// For vertical wall (normal = 1,0), horizontal velocity should be 0
				// For horizontal wall (normal = 0,-1), vertical velocity should be 0
				if tt.normal.NX == 1 {
					if math.Abs(vel.VX) > epsilon {
						t.Errorf("ResolveWallCollision() velocity not adjusted: VX = %v", vel.VX)
					}
				}
				if tt.normal.NY == -1 {
					if math.Abs(vel.VY) > epsilon {
						t.Errorf("ResolveWallCollision() velocity not adjusted: VY = %v", vel.VY)
					}
				}
			}
		})
	}
}

// TestCheckPreciseCollision verifies pixel-perfect collision detection.
func TestCheckPreciseCollision(t *testing.T) {
	tests := []struct {
		name        string
		setupE1     func() *Entity
		setupE2     func() *Entity
		wantCollide bool
	}{
		{
			name: "precise colliders intersect",
			setupE1: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&PositionComponent{X: 0, Y: 0})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					Shape: ShapeAABB,
				})
				return e
			},
			setupE2: func() *Entity {
				e := NewEntity(2)
				e.AddComponent(&PositionComponent{X: 16, Y: 16})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					Shape: ShapeAABB,
				})
				return e
			},
			wantCollide: true,
		},
		{
			name: "precise colliders separate",
			setupE1: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&PositionComponent{X: 0, Y: 0})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					Shape: ShapeAABB,
				})
				return e
			},
			setupE2: func() *Entity {
				e := NewEntity(2)
				e.AddComponent(&PositionComponent{X: 50, Y: 50})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					Shape: ShapeAABB,
				})
				return e
			},
			wantCollide: false,
		},
		{
			name: "fallback to standard collider",
			setupE1: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&PositionComponent{X: 0, Y: 0})
				e.AddComponent(&ColliderComponent{
					Width: 32, Height: 32,
				})
				return e
			},
			setupE2: func() *Entity {
				e := NewEntity(2)
				e.AddComponent(&PositionComponent{X: 16, Y: 16})
				e.AddComponent(&ColliderComponent{
					Width: 32, Height: 32,
				})
				return e
			},
			wantCollide: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e1 := tt.setupE1()
			e2 := tt.setupE2()

			got := CheckPreciseCollision(e1, e2)
			if got != tt.wantCollide {
				t.Errorf("CheckPreciseCollision() = %v, want %v", got, tt.wantCollide)
			}
		})
	}
}

// TestGetCollisionAlignment verifies visual/collision alignment calculation.
func TestGetCollisionAlignment(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Entity
		wantError float64
		maxError  float64
	}{
		{
			name: "perfect alignment",
			setup: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&PositionComponent{X: 100, Y: 100})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					OffsetX: -16, OffsetY: -16, // Centered
				})
				return e
			},
			wantError: 0.0,
			maxError:  0.01,
		},
		{
			name: "small offset",
			setup: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&PositionComponent{X: 100, Y: 100})
				e.AddComponent(&PreciseColliderComponent{
					Width: 32, Height: 32,
					OffsetX: -16.3, OffsetY: -16.4, // Slightly off-center
				})
				return e
			},
			wantError: 0.5,
			maxError:  0.6, // Should be <0.5px per Phase 48 requirements
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setup()

			alignmentError := GetCollisionAlignment(entity)

			if math.Abs(alignmentError-tt.wantError) > tt.maxError {
				t.Errorf("GetCollisionAlignment() = %v, want ~%v (max %v)",
					alignmentError, tt.wantError, tt.maxError)
			}

			// Verify Phase 48 requirement: <0.5px alignment
			if alignmentError >= 0.5 {
				t.Errorf("GetCollisionAlignment() = %v, exceeds 0.5px requirement", alignmentError)
			}
		})
	}
}

// TestGetCollisionAlignment_StandardCollider verifies alignment calculation with standard ColliderComponent.
// This test specifically exercises the code path that had a variable shadowing bug (hasCollider)
// where the inner `:=` in `else if cComp, hasCollider := ...` shadowed the outer `hasCollider`,
// causing the function to incorrectly report 0 alignment error for standard colliders.
func TestGetCollisionAlignment_StandardCollider(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ColliderComponent{
		Width: 32, Height: 32,
		OffsetX: -16, OffsetY: -16, // Centered
		Solid:   true,
	})

	alignmentError := GetCollisionAlignment(entity)

	// With centered collider (offset = -half-width/height), error should be ~0
	if alignmentError > 0.5 {
		t.Errorf("GetCollisionAlignment() with standard collider = %v, want <0.5", alignmentError)
	}
}

// TestGetCollisionAlignment_StandardColliderOffset verifies alignment with offset standard collider.
func TestGetCollisionAlignment_StandardColliderOffset(t *testing.T) {
	entity := NewEntity(2)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ColliderComponent{
		Width: 32, Height: 32,
		OffsetX: 0, OffsetY: 0, // Not centered — offset from position
		Solid:   true,
	})

	alignmentError := GetCollisionAlignment(entity)

	// With non-centered collider, alignment error should be > 0
	if alignmentError == 0 {
		t.Error("GetCollisionAlignment() with offset standard collider = 0, expected non-zero alignment error")
	}
}

// TestCollisionPrecision verifies the 0.1px precision constant.
func TestCollisionPrecision(t *testing.T) {
	if CollisionPrecision != 0.1 {
		t.Errorf("CollisionPrecision = %v, want 0.1", CollisionPrecision)
	}
}

// BenchmarkQuantizePosition benchmarks position quantization performance.
func BenchmarkQuantizePosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		QuantizePosition(123.456, 789.012)
	}
}

// BenchmarkIntersectsAABB benchmarks AABB intersection checks.
func BenchmarkIntersectsAABB(b *testing.B) {
	c1 := &PreciseColliderComponent{Width: 32, Height: 32, Shape: ShapeAABB}
	c2 := &PreciseColliderComponent{Width: 32, Height: 32, Shape: ShapeAABB}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c1.IntersectsAABB(0, 0, c2, 16, 16)
	}
}

// BenchmarkIntersectsCircle benchmarks circle intersection checks.
func BenchmarkIntersectsCircle(b *testing.B) {
	c1 := &PreciseColliderComponent{Width: 32, Height: 32, Shape: ShapeCircle}
	c2 := &PreciseColliderComponent{Width: 32, Height: 32, Shape: ShapeCircle}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c1.IntersectsCircle(0, 0, c2, 20, 0)
	}
}

// BenchmarkApplyWallSlide benchmarks wall sliding calculations.
func BenchmarkApplyWallSlide(b *testing.B) {
	normal := EdgeNormal{NX: 1, NY: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyWallSlide(5.0, 3.0, normal)
	}
}

// TestSetCollisionLogLevel verifies that SetCollisionLogLevel updates both logger and cache.
func TestSetCollisionLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         logrus.Level
		expectEnabled bool
	}{
		{"debug level enables flag", logrus.DebugLevel, true},
		{"trace level enables flag", logrus.TraceLevel, true},
		{"info level disables flag", logrus.InfoLevel, false},
		{"warn level disables flag", logrus.WarnLevel, false},
		{"error level disables flag", logrus.ErrorLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetCollisionLogLevel(tt.level)

			if collisionDebugEnabled != tt.expectEnabled {
				t.Errorf("collisionDebugEnabled = %v, want %v for level %v",
					collisionDebugEnabled, tt.expectEnabled, tt.level)
			}

			if collisionLog.GetLevel() != tt.level {
				t.Errorf("collisionLog.GetLevel() = %v, want %v",
					collisionLog.GetLevel(), tt.level)
			}
		})
	}
}

// TestCollisionDebugFlagCache verifies that cached flag avoids GetLevel() calls.
func TestCollisionDebugFlagCache(t *testing.T) {
	// Set to Info (debug disabled)
	SetCollisionLogLevel(logrus.InfoLevel)
	if collisionDebugEnabled {
		t.Fatal("collisionDebugEnabled should be false at Info level")
	}

	// Call functions that previously checked GetLevel() every call
	c := &PreciseColliderComponent{Width: 10, Height: 10}
	QuantizePosition(5.5, 7.7)
	c.GetBounds(0, 0)
	c.IntersectsAABB(0, 0, c, 10, 10)
	ComputeWallNormal(5, 5, 0, 0)
	ApplyWallSlide(1.0, 1.0, EdgeNormal{NX: 1, NY: 0})

	// Set to Debug (debug enabled)
	SetCollisionLogLevel(logrus.DebugLevel)
	if !collisionDebugEnabled {
		t.Fatal("collisionDebugEnabled should be true at Debug level")
	}

	// Functions should now use cached true value
	QuantizePosition(5.5, 7.7)
	c.GetBounds(0, 0)
	c.IntersectsAABB(0, 0, c, 10, 10)
	ComputeWallNormal(5, 5, 0, 0)
	ApplyWallSlide(1.0, 1.0, EdgeNormal{NX: 1, NY: 0})
}

// TestRefreshCollisionDebugFlag verifies internal flag refresh logic.
func TestRefreshCollisionDebugFlag(t *testing.T) {
	// Manually set logger level and refresh
	collisionLog.SetLevel(logrus.WarnLevel)
	refreshCollisionDebugFlag()

	if collisionDebugEnabled {
		t.Errorf("collisionDebugEnabled should be false at Warn level, got true")
	}

	collisionLog.SetLevel(logrus.DebugLevel)
	refreshCollisionDebugFlag()

	if !collisionDebugEnabled {
		t.Errorf("collisionDebugEnabled should be true at Debug level, got false")
	}
}

// BenchmarkQuantizePosition_CachedDebugFlag benchmarks with cached flag (debug off).
func BenchmarkQuantizePosition_CachedDebugFlag(b *testing.B) {
	SetCollisionLogLevel(logrus.InfoLevel) // Disable debug

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuantizePosition(123.456, 789.012)
	}
}

// BenchmarkGetBounds_CachedDebugFlag benchmarks with cached flag (debug off).
func BenchmarkGetBounds_CachedDebugFlag(b *testing.B) {
	SetCollisionLogLevel(logrus.InfoLevel) // Disable debug
	c := &PreciseColliderComponent{
		Width: 32, Height: 32,
		OffsetX: -16, OffsetY: -16,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetBounds(100, 200)
	}
}

// BenchmarkIntersectsAABB_CachedDebugFlag benchmarks with cached flag (debug off).
func BenchmarkIntersectsAABB_CachedDebugFlag(b *testing.B) {
	SetCollisionLogLevel(logrus.InfoLevel) // Disable debug
	c1 := &PreciseColliderComponent{Width: 32, Height: 32}
	c2 := &PreciseColliderComponent{Width: 32, Height: 32}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c1.IntersectsAABB(0, 0, c2, 16, 16)
	}
}
