package engine

import (
	"math"
	"testing"
)

func TestPositionComponent(t *testing.T) {
	pos := &PositionComponent{X: 10.0, Y: 20.0}
	
	if pos.Type() != "position" {
		t.Errorf("Expected type 'position', got %s", pos.Type())
	}
	
	if pos.X != 10.0 || pos.Y != 20.0 {
		t.Errorf("Position values incorrect: got (%f, %f)", pos.X, pos.Y)
	}
}

func TestVelocityComponent(t *testing.T) {
	vel := &VelocityComponent{VX: 5.0, VY: -3.0}
	
	if vel.Type() != "velocity" {
		t.Errorf("Expected type 'velocity', got %s", vel.Type())
	}
}

func TestColliderComponent(t *testing.T) {
	collider := &ColliderComponent{
		Width:  10.0,
		Height: 15.0,
		Solid:  true,
	}
	
	if collider.Type() != "collider" {
		t.Errorf("Expected type 'collider', got %s", collider.Type())
	}
}

func TestColliderGetBounds(t *testing.T) {
	tests := []struct {
		name     string
		collider ColliderComponent
		x, y     float64
		wantMinX float64
		wantMinY float64
		wantMaxX float64
		wantMaxY float64
	}{
		{
			name:     "no offset",
			collider: ColliderComponent{Width: 10, Height: 20},
			x:        5, y: 10,
			wantMinX: 5, wantMinY: 10,
			wantMaxX: 15, wantMaxY: 30,
		},
		{
			name:     "with offset",
			collider: ColliderComponent{Width: 10, Height: 20, OffsetX: -5, OffsetY: -10},
			x:        5, y: 10,
			wantMinX: 0, wantMinY: 0,
			wantMaxX: 10, wantMaxY: 20,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minX, minY, maxX, maxY := tt.collider.GetBounds(tt.x, tt.y)
			
			if minX != tt.wantMinX || minY != tt.wantMinY || maxX != tt.wantMaxX || maxY != tt.wantMaxY {
				t.Errorf("GetBounds() = (%f,%f,%f,%f), want (%f,%f,%f,%f)",
					minX, minY, maxX, maxY,
					tt.wantMinX, tt.wantMinY, tt.wantMaxX, tt.wantMaxY)
			}
		})
	}
}

func TestColliderIntersects(t *testing.T) {
	tests := []struct {
		name      string
		c1        ColliderComponent
		x1, y1    float64
		c2        ColliderComponent
		x2, y2    float64
		wantIntersect bool
	}{
		{
			name:      "overlapping",
			c1:        ColliderComponent{Width: 10, Height: 10},
			x1:        0, y1: 0,
			c2:        ColliderComponent{Width: 10, Height: 10},
			x2:        5, y2: 5,
			wantIntersect: true,
		},
		{
			name:      "not overlapping",
			c1:        ColliderComponent{Width: 10, Height: 10},
			x1:        0, y1: 0,
			c2:        ColliderComponent{Width: 10, Height: 10},
			x2:        20, y2: 20,
			wantIntersect: false,
		},
		{
			name:      "touching edge",
			c1:        ColliderComponent{Width: 10, Height: 10},
			x1:        0, y1: 0,
			c2:        ColliderComponent{Width: 10, Height: 10},
			x2:        10, y2: 0,
			wantIntersect: false, // Touching but not overlapping
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.c1.Intersects(tt.x1, tt.y1, &tt.c2, tt.x2, tt.y2)
			if result != tt.wantIntersect {
				t.Errorf("Intersects() = %v, want %v", result, tt.wantIntersect)
			}
		})
	}
}

func TestBoundsComponent(t *testing.T) {
	bounds := &BoundsComponent{
		MinX: 0, MinY: 0,
		MaxX: 100, MaxY: 100,
	}
	
	if bounds.Type() != "bounds" {
		t.Errorf("Expected type 'bounds', got %s", bounds.Type())
	}
}

func TestBoundsClamp(t *testing.T) {
	tests := []struct {
		name   string
		bounds BoundsComponent
		x, y   float64
		wantX  float64
		wantY  float64
	}{
		{
			name:   "within bounds",
			bounds: BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100},
			x:      50, y: 50,
			wantX:  50, wantY: 50,
		},
		{
			name:   "below minimum",
			bounds: BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100},
			x:      -10, y: -5,
			wantX:  0, wantY: 0,
		},
		{
			name:   "above maximum",
			bounds: BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100},
			x:      110, y: 105,
			wantX:  100, wantY: 100,
		},
		{
			name:   "wrap around max",
			bounds: BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100, Wrap: true},
			x:      105, y: 110,
			wantX:  5, wantY: 10,
		},
		{
			name:   "wrap around min",
			bounds: BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100, Wrap: true},
			x:      -5, y: -10,
			wantX:  95, wantY: 90,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := tt.bounds.Clamp(tt.x, tt.y)
			
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("Clamp(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestMovementSystem(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0) // No speed limit
	
	// Create entity with position and velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 5})
	
	world.Update(0) // Process pending additions
	
	// Update for 1 second
	system.Update(world.GetEntities(), 1.0)
	
	pos, _ := entity.GetComponent("position")
	position := pos.(*PositionComponent)
	
	if position.X != 10 || position.Y != 5 {
		t.Errorf("Position after update = (%f, %f), want (10, 5)", position.X, position.Y)
	}
}

func TestMovementSystemWithSpeedLimit(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(10.0) // Max speed of 10
	
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100}) // Very fast
	
	world.Update(0)
	
	// Update for 0.1 seconds
	system.Update(world.GetEntities(), 0.1)
	
	vel, _ := entity.GetComponent("velocity")
	velocity := vel.(*VelocityComponent)
	
	// Speed should be clamped to 10
	speed := math.Sqrt(velocity.VX*velocity.VX + velocity.VY*velocity.VY)
	
	if math.Abs(speed-10.0) > 0.01 {
		t.Errorf("Speed = %f, want 10.0", speed)
	}
}

func TestMovementSystemWithBounds(t *testing.T) {
	world := NewWorld()
	system := NewMovementSystem(0)
	
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 95, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 0})
	entity.AddComponent(&BoundsComponent{MinX: 0, MinY: 0, MaxX: 100, MaxY: 100})
	
	world.Update(0)
	
	// Update for 1 second - should hit boundary
	system.Update(world.GetEntities(), 1.0)
	
	pos, _ := entity.GetComponent("position")
	position := pos.(*PositionComponent)
	
	// Should be clamped to max boundary
	if position.X != 100 {
		t.Errorf("Position.X = %f, want 100", position.X)
	}
	
	// Velocity should be stopped
	vel, _ := entity.GetComponent("velocity")
	velocity := vel.(*VelocityComponent)
	
	if velocity.VX != 0 {
		t.Errorf("Velocity.VX = %f, want 0", velocity.VX)
	}
}

func TestGetSetVelocity(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	
	SetVelocity(entity, 10, 20)
	
	vel, _ := entity.GetComponent("velocity")
	velocity := vel.(*VelocityComponent)
	
	if velocity.VX != 10 || velocity.VY != 20 {
		t.Errorf("Velocity = (%f, %f), want (10, 20)", velocity.VX, velocity.VY)
	}
}

func TestGetSetPosition(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	
	SetPosition(entity, 15, 25)
	
	x, y, ok := GetPosition(entity)
	
	if !ok {
		t.Fatal("GetPosition returned ok=false")
	}
	
	if x != 15 || y != 25 {
		t.Errorf("Position = (%f, %f), want (15, 25)", x, y)
	}
}

func TestGetDistance(t *testing.T) {
	e1 := NewEntity(1)
	e1.AddComponent(&PositionComponent{X: 0, Y: 0})
	
	e2 := NewEntity(2)
	e2.AddComponent(&PositionComponent{X: 3, Y: 4})
	
	distance := GetDistance(e1, e2)
	
	// Distance should be 5 (3-4-5 triangle)
	if math.Abs(distance-5.0) > 0.01 {
		t.Errorf("Distance = %f, want 5.0", distance)
	}
}

func TestMoveTowards(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	
	// Move towards (10, 0) at speed 5
	reached := MoveTowards(entity, 10, 0, 5, 1.0)
	
	if reached {
		t.Error("Should not have reached target immediately")
	}
	
	vel, _ := entity.GetComponent("velocity")
	velocity := vel.(*VelocityComponent)
	
	// Velocity should be (5, 0) - normalized direction * speed
	if math.Abs(velocity.VX-5.0) > 0.01 || math.Abs(velocity.VY) > 0.01 {
		t.Errorf("Velocity = (%f, %f), want (5, 0)", velocity.VX, velocity.VY)
	}
	
	// Test reaching target
	entity.AddComponent(&PositionComponent{X: 10, Y: 0})
	reached = MoveTowards(entity, 10, 0, 5, 1.0)
	
	if !reached {
		t.Error("Should have reached target")
	}
}
