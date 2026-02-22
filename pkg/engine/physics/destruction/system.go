package destruction

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
)

// System manages building structural integrity, damage propagation, and debris
type System struct {
	config         *Config
	buildings      map[string]*StructuralIntegrity
	debris         []*DebrisParticle
	fallingObjects []*FallingObject
	accumulator    float64 // Time accumulator for fixed timestep
	// Reusable buffers to reduce per-frame allocations
	debrisBuffer  []*DebrisParticle
	fallingBuffer []*FallingObject
}

// NewSystem creates a new destruction system
func NewSystem(config *Config) *System {
	if config == nil {
		config = DefaultConfig()
	}
	return &System{
		config:         config,
		buildings:      make(map[string]*StructuralIntegrity),
		debris:         make([]*DebrisParticle, 0, config.MaxDebrisParticles),
		fallingObjects: make([]*FallingObject, 0, config.MaxFallingObjects),
		accumulator:    0,
		debrisBuffer:   make([]*DebrisParticle, 0, config.MaxDebrisParticles),
		fallingBuffer:  make([]*FallingObject, 0, config.MaxFallingObjects),
	}
}

// Update updates structural integrity, damage propagation, and physics
func (s *System) Update(deltaTime float64) {
	if deltaTime <= 0 {
		return
	}

	s.accumulator += deltaTime
	fixedDelta := 1.0 / s.config.UpdateFrequency

	for s.accumulator >= fixedDelta {
		s.updateIntegrity(fixedDelta)
		s.updateDebris(fixedDelta)
		s.updateFallingObjects(fixedDelta)
		s.accumulator -= fixedDelta
	}
}

// RegisterBuilding adds a building to the integrity tracking system.
// Returns an error if buildingID is empty or dimensions are invalid (non-positive).
func (s *System) RegisterBuilding(buildingID string, width, height, floors int, material MaterialType) error {
	if buildingID == "" {
		return fmt.Errorf("buildingID cannot be empty")
	}
	if width <= 0 || height <= 0 || floors <= 0 {
		return fmt.Errorf("invalid building dimensions: width=%d height=%d floors=%d", width, height, floors)
	}

	supports := s.generateSupports(width, height, floors, material)

	totalHealth := float64(len(supports)) * 100.0

	s.buildings[buildingID] = &StructuralIntegrity{
		BuildingID:     buildingID,
		TotalHealth:    totalHealth,
		CurrentHealth:  1.0,
		Supports:       supports,
		DamagedAreas:   []DamageArea{},
		State:          IntegrityPristine,
		CollapseRisk:   0.0,
		LastDamageTime: 0,
	}
	return nil
}

// ApplyDamage applies damage to a building at a specific location
func (s *System) ApplyDamage(buildingID string, x, y, floor int, amount, radius float64) error {
	integrity, ok := s.buildings[buildingID]
	if !ok {
		return fmt.Errorf("building not found: %s", buildingID)
	}

	damage := DamageArea{
		X:            x,
		Y:            y,
		Floor:        floor,
		Radius:       radius,
		Severity:     amount,
		PropagateAge: 0,
	}
	integrity.DamagedAreas = append(integrity.DamagedAreas, damage)

	for i := range integrity.Supports {
		support := &integrity.Supports[i]
		if support.Floor != floor {
			continue
		}

		dx := float64(support.X - x)
		dy := float64(support.Y - y)
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= radius {
			falloff := 1.0 - (dist / radius)
			damageAmount := amount * falloff
			support.Health -= damageAmount
			if support.Health < 0 {
				support.Health = 0
			}
		}
	}

	s.updateBuildingState(buildingID)
	return nil
}

// GetIntegrity returns the structural integrity of a building
func (s *System) GetIntegrity(buildingID string) (*StructuralIntegrity, error) {
	integrity, ok := s.buildings[buildingID]
	if !ok {
		return nil, fmt.Errorf("building not found: %s", buildingID)
	}
	return integrity, nil
}

// RemoveBuilding removes a building from tracking.
// Returns an error if the building does not exist.
func (s *System) RemoveBuilding(buildingID string) error {
	if _, exists := s.buildings[buildingID]; !exists {
		return fmt.Errorf("building not found: %s", buildingID)
	}
	delete(s.buildings, buildingID)
	return nil
}

// GetDebrisCount returns the number of active debris particles
func (s *System) GetDebrisCount() int {
	return len(s.debris)
}

// GetFallingObjectCount returns the number of active falling objects
func (s *System) GetFallingObjectCount() int {
	return len(s.fallingObjects)
}

// GetDebris returns all active debris particles
func (s *System) GetDebris() []*DebrisParticle {
	return s.debris
}

// GetFallingObjects returns all active falling objects
func (s *System) GetFallingObjects() []*FallingObject {
	return s.fallingObjects
}

// generateSupports creates structural support points for a building
func (s *System) generateSupports(width, height, floors int, material MaterialType) []SupportPoint {
	supports := []SupportPoint{}
	props := GetMaterialProperties(material)

	for floor := 0; floor < floors; floor++ {
		isGroundFloor := floor == 0

		if isGroundFloor {
			for x := 0; x < width; x += 2 {
				supports = append(supports, SupportPoint{
					X:            x,
					Y:            0,
					Floor:        floor,
					Type:         SupportFoundation,
					Health:       1.0,
					LoadBearing:  true,
					LoadCapacity: props.Durability,
					CurrentLoad:  float64(floors-floor) / float64(floors),
				})
			}
		}

		supports = append(supports,
			SupportPoint{
				X:            0,
				Y:            height / 2,
				Floor:        floor,
				Type:         SupportWall,
				Health:       1.0,
				LoadBearing:  true,
				LoadCapacity: props.Durability,
				CurrentLoad:  float64(floors-floor) / float64(floors),
			},
			SupportPoint{
				X:            width - 1,
				Y:            height / 2,
				Floor:        floor,
				Type:         SupportWall,
				Health:       1.0,
				LoadBearing:  true,
				LoadCapacity: props.Durability,
				CurrentLoad:  float64(floors-floor) / float64(floors),
			},
		)

		if width > 10 {
			columnX := width / 2
			supports = append(supports, SupportPoint{
				X:            columnX,
				Y:            height / 2,
				Floor:        floor,
				Type:         SupportColumn,
				Health:       1.0,
				LoadBearing:  true,
				LoadCapacity: props.Durability * 1.5,
				CurrentLoad:  float64(floors-floor) / float64(floors),
			})
		}
	}

	return supports
}

// updateIntegrity checks and updates building structural integrity
func (s *System) updateIntegrity(deltaTime float64) {
	if !s.config.EnableIntegrityChecks {
		return
	}

	for buildingID := range s.buildings {
		s.propagateDamage(buildingID, deltaTime)
		s.checkCollapse(buildingID)
		s.updateBuildingState(buildingID)
	}
}

// propagateDamage spreads damage over time from damaged areas
func (s *System) propagateDamage(buildingID string, deltaTime float64) {
	integrity := s.buildings[buildingID]

	for i := range integrity.DamagedAreas {
		area := &integrity.DamagedAreas[i]
		area.PropagateAge += deltaTime

		if area.PropagateAge < 1.0 {
			continue
		}

		propagationAmount := area.Severity * s.config.DamagePropagationRate * deltaTime

		for j := range integrity.Supports {
			support := &integrity.Supports[j]
			if support.Floor != area.Floor {
				continue
			}

			dx := float64(support.X - area.X)
			dy := float64(support.Y - area.Y)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= area.Radius+2.0 && dist > area.Radius {
				support.Health -= propagationAmount * 0.1
				if support.Health < 0 {
					support.Health = 0
				}
			}
		}
	}
}

// checkCollapse determines if a building should collapse
func (s *System) checkCollapse(buildingID string) {
	integrity := s.buildings[buildingID]

	loadBearingSupports := 0
	damagedLoadBearing := 0

	for _, support := range integrity.Supports {
		if support.LoadBearing {
			loadBearingSupports++
			if support.Health < 0.3 {
				damagedLoadBearing++
			}
		}
	}

	if loadBearingSupports > 0 {
		integrity.CollapseRisk = float64(damagedLoadBearing) / float64(loadBearingSupports)
	}

	if integrity.CurrentHealth <= s.config.CollapseThreshold {
		s.triggerCollapse(buildingID)
	}
}

// triggerCollapse initiates building collapse and debris generation
func (s *System) triggerCollapse(buildingID string) {
	integrity := s.buildings[buildingID]
	if integrity.State == IntegrityCollapsed {
		return
	}

	integrity.State = IntegrityCollapsed

	if s.config.EnableDebris {
		s.generateCollapseDebris(integrity)
	}
}

// hashBuildingID generates a deterministic hash from a building ID for seed derivation
func hashBuildingID(buildingID string) int64 {
	var hash int64 = 17
	for _, c := range buildingID {
		hash = hash*31 + int64(c)
	}
	return hash
}

// generateCollapseDebris creates debris particles from a collapsed building.
// Uses deterministic seed derived from config.Seed and building ID for multiplayer sync.
func (s *System) generateCollapseDebris(integrity *StructuralIntegrity) {
	debrisCount := len(integrity.Supports) * 3
	if debrisCount > s.config.MaxDebrisParticles-len(s.debris) {
		debrisCount = s.config.MaxDebrisParticles - len(s.debris)
	}

	// Derive deterministic seed from config seed and building ID hash
	buildingHash := hashBuildingID(integrity.BuildingID)
	seed := s.config.Seed ^ buildingHash
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < debrisCount; i++ {
		support := integrity.Supports[i%len(integrity.Supports)]

		debris := &DebrisParticle{
			X:        float64(support.X*32) + rng.Float64()*16 - 8,
			Y:        float64(support.Y*32) + rng.Float64()*16 - 8,
			VelX:     (rng.Float64() - 0.5) * 200,
			VelY:     (rng.Float64() - 0.5) * 200,
			RotVel:   (rng.Float64() - 0.5) * 6.28,
			Angle:    rng.Float64() * 6.28,
			Mass:     0.5 + rng.Float64()*0.5,
			Size:     4 + rng.Float64()*12,
			Material: MaterialStone,
			Color:    color.RGBA{R: 128, G: 128, B: 128, A: 255},
			Life:     s.config.DebrisLifetime,
		}

		s.debris = append(s.debris, debris)
	}
}

// updateBuildingState updates the integrity state based on health
func (s *System) updateBuildingState(buildingID string) {
	integrity := s.buildings[buildingID]

	supportHealth := 0.0
	for _, support := range integrity.Supports {
		supportHealth += support.Health
	}

	if len(integrity.Supports) > 0 {
		integrity.CurrentHealth = supportHealth / float64(len(integrity.Supports))
	}

	if integrity.CurrentHealth >= 0.99 {
		integrity.State = IntegrityPristine
	} else if integrity.CurrentHealth >= 0.5 {
		integrity.State = IntegrityDamaged
	} else if integrity.CurrentHealth > s.config.CollapseThreshold {
		integrity.State = IntegrityCritical
	} else {
		integrity.State = IntegrityCollapsed
	}
}

// updateDebris updates debris particle physics
func (s *System) updateDebris(deltaTime float64) {
	// Reuse buffer to reduce per-frame allocations
	s.debrisBuffer = s.debrisBuffer[:0]

	// Ensure capacity
	if cap(s.debrisBuffer) < len(s.debris) {
		s.debrisBuffer = make([]*DebrisParticle, 0, len(s.debris))
	}

	for _, d := range s.debris {
		d.Life -= deltaTime
		if d.Life <= 0 {
			continue
		}

		d.VelY += s.config.Gravity * deltaTime

		d.VelX *= (1.0 - s.config.AirResistance)
		d.VelY *= (1.0 - s.config.AirResistance)

		d.X += d.VelX * deltaTime
		d.Y += d.VelY * deltaTime
		d.Angle += d.RotVel * deltaTime

		s.debrisBuffer = append(s.debrisBuffer, d)
	}

	// Swap buffers to avoid allocation
	s.debris, s.debrisBuffer = s.debrisBuffer, s.debris
}

// updateFallingObjects updates falling object physics with realistic gravity and bouncing
func (s *System) updateFallingObjects(deltaTime float64) {
	// Reuse buffer to reduce per-frame allocations
	s.fallingBuffer = s.fallingBuffer[:0]

	// Ensure capacity
	if cap(s.fallingBuffer) < len(s.fallingObjects) {
		s.fallingBuffer = make([]*FallingObject, 0, len(s.fallingObjects))
	}

	for _, obj := range s.fallingObjects {
		if obj.IsGrounded() && obj.Bounces >= obj.MaxBounces {
			continue
		}

		props := GetMaterialProperties(obj.Material)

		obj.VelZ -= s.config.Gravity * deltaTime

		obj.VelX *= (1.0 - s.config.AirResistance)
		obj.VelY *= (1.0 - s.config.AirResistance)
		obj.VelZ *= (1.0 - s.config.AirResistance)

		obj.X += obj.VelX * deltaTime
		obj.Y += obj.VelY * deltaTime
		obj.Z += obj.VelZ * deltaTime
		obj.Angle += obj.RotVel * deltaTime

		if obj.Z <= 0 {
			obj.Z = 0
			obj.VelZ = -obj.VelZ * props.Bounciness

			obj.VelX *= (1.0 - props.Friction*s.config.GroundFriction)
			obj.VelY *= (1.0 - props.Friction*s.config.GroundFriction)
			obj.RotVel *= (1.0 - props.Friction)

			obj.Bounces++

			if math.Abs(obj.VelZ) < 10 && obj.Bounces >= obj.MaxBounces {
				obj.OnGround = true
			}
		}

		s.fallingBuffer = append(s.fallingBuffer, obj)
	}

	// Swap buffers to avoid allocation
	s.fallingObjects, s.fallingBuffer = s.fallingBuffer, s.fallingObjects
}

// SpawnFallingObject adds a falling object to the simulation with zero initial velocity.
// For deterministic initial velocity, use SpawnFallingObjectWithSeed.
// Returns an error if the maximum number of falling objects has been reached.
func (s *System) SpawnFallingObject(x, y, z float64, material MaterialType, width, height float64) error {
	return s.spawnFallingObjectInternal(x, y, z, material, width, height, 0, 0, 0, 0)
}

// SpawnFallingObjectWithSeed adds a falling object with deterministic random initial velocity.
// The seed is combined with the system's config seed for reproducible results across runs.
// This is useful for network sync scenarios where client and server must produce identical physics.
// Returns an error if the maximum number of falling objects has been reached.
func (s *System) SpawnFallingObjectWithSeed(x, y, z float64, material MaterialType, width, height float64, seed int64) error {
	// Derive deterministic velocity from seed
	rng := rand.New(rand.NewSource(s.config.Seed ^ seed))
	velX := (rng.Float64() - 0.5) * 100 // -50 to 50 initial horizontal velocity
	velY := (rng.Float64() - 0.5) * 100
	velZ := rng.Float64() * 50 // 0 to 50 initial upward velocity (will fall due to gravity)
	rotVel := (rng.Float64() - 0.5) * 3.14 // -π/2 to π/2 rotation

	return s.spawnFallingObjectInternal(x, y, z, material, width, height, velX, velY, velZ, rotVel)
}

// spawnFallingObjectInternal is the internal implementation for spawning falling objects.
func (s *System) spawnFallingObjectInternal(x, y, z float64, material MaterialType, width, height, velX, velY, velZ, rotVel float64) error {
	if len(s.fallingObjects) >= s.config.MaxFallingObjects {
		return fmt.Errorf("cannot spawn falling object: limit reached (%d/%d)",
			len(s.fallingObjects), s.config.MaxFallingObjects)
	}

	obj := &FallingObject{
		X:          x,
		Y:          y,
		Z:          z,
		VelX:       velX,
		VelY:       velY,
		VelZ:       velZ,
		RotVel:     rotVel,
		Angle:      0,
		Mass:       width * height * GetMaterialProperties(material).Density,
		Width:      width,
		Height:     height,
		Material:   material,
		OnGround:   false,
		Bounces:    0,
		MaxBounces: 3,
	}

	s.fallingObjects = append(s.fallingObjects, obj)
	return nil
}
