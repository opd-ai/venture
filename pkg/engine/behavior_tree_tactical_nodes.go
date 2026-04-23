// Package engine provides tactical behavior tree action and condition nodes.
// These nodes implement specific AI behaviors for combat, movement, and decision-making.
// All nodes follow the BehaviorNode interface and use the blackboard for state.
package engine

import (
	"fmt"
	"math"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
	"github.com/sirupsen/logrus"
)

// ==============================================================================
// HELPER FUNCTIONS - Common patterns for behavior tree nodes
// ==============================================================================

// TargetPositions holds the extracted position data for entity and target.
type TargetPositions struct {
	MyPos     *PositionComponent
	TargetPos *PositionComponent
	Target    *Entity
	Dist      float64
	Dx, Dy    float64
}

// btMovementParams holds configuration for the runMovementTick scaffold.
type btMovementParams struct {
	Speed    float64
	StopDist float64
}

// btAllyCtx holds the resolved faction/position/nearby context passed to the
// decide step in runMovementTick.
type btAllyCtx struct {
	Pos     *PositionComponent
	Faction *FactionComponent
	Nearby  []*Entity
}

// getFactionAndNearbyEntities extracts the faction component, position, and nearby
// entity list that movement-oriented behavior-tree Tick methods all acquire at
// their start. Returns ok=false if any required component is absent.
func getFactionAndNearbyEntities(entity *Entity, bb *Blackboard) (*FactionComponent, *PositionComponent, []*Entity, bool) {
	factionComp, hasFaction := entity.GetComponent("faction")
	if !hasFaction {
		return nil, nil, nil, false
	}
	faction, ok := factionComp.(*FactionComponent)
	if !ok {
		return nil, nil, nil, false
	}

	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return nil, nil, nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil, nil, false
	}

	nearbyVal, hasNearby := bb.Get("nearby_entities")
	if !hasNearby || nearbyVal == nil {
		return nil, nil, nil, false
	}
	nearby, ok := nearbyVal.([]*Entity)
	if !ok {
		return nil, nil, nil, false
	}

	return faction, pos, nearby, true
}

// runMovementTick is the shared scaffold for movement-oriented behavior-tree
// Tick methods. It handles faction lookup, position acquisition, nearby-entity
// resolution, movement application, velocity update, and timeMoving
// bookkeeping. The decide callback receives the resolved context and returns
// the target (targetX, targetY) to move toward. Return ok=false to signal no
// valid target, which causes NodeFailure.
func runMovementTick(
	entity *Entity,
	bb *Blackboard,
	dt float64,
	params btMovementParams,
	timeMoving *float64,
	decide func(ctx btAllyCtx, entity *Entity, bb *Blackboard) (targetX, targetY float64, ok bool),
) NodeStatus {
	faction, pos, nearby, resolved := getFactionAndNearbyEntities(entity, bb)
	if !resolved {
		return NodeFailure
	}

	ctx := btAllyCtx{Pos: pos, Faction: faction, Nearby: nearby}
	targetX, targetY, found := decide(ctx, entity, bb)
	if !found {
		return NodeFailure
	}

	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= params.StopDist {
		*timeMoving = 0
		return NodeSuccess
	}

	if dist > 0 {
		nx := dx / dist
		ny := dy / dist
		pos.X += nx * params.Speed * dt
		pos.Y += ny * params.Speed * dt

		if velComp, ok := entity.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX = nx * params.Speed
				vel.VY = ny * params.Speed
			}
		}
	}

	*timeMoving += dt
	return NodeRunning
}

// GetTargetPositions extracts target from blackboard and returns both positions.
// Returns nil if target or positions are unavailable.
func GetTargetPositions(entity *Entity, blackboard *Blackboard) *TargetPositions {
	targetVal, ok := blackboard.Get("target")
	if !ok || targetVal == nil {
		return nil
	}
	targetEntity, ok := targetVal.(*Entity)
	if !ok || targetEntity == nil {
		return nil
	}

	myPos, ok := entity.GetComponent("position")
	if !ok {
		return nil
	}
	myPosComp, ok := myPos.(*PositionComponent)
	if !ok {
		return nil
	}

	targetPos, ok := targetEntity.GetComponent("position")
	if !ok {
		return nil
	}
	targetPosComp, ok := targetPos.(*PositionComponent)
	if !ok {
		return nil
	}

	dx := targetPosComp.X - myPosComp.X
	dy := targetPosComp.Y - myPosComp.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	return &TargetPositions{
		MyPos:     myPosComp,
		TargetPos: targetPosComp,
		Target:    targetEntity,
		Dist:      dist,
		Dx:        dx,
		Dy:        dy,
	}
}

// ==============================================================================
// CONDITION NODES - Check entity state or world conditions
// ==============================================================================

// HealthBelowNode checks if entity health is below a threshold percentage.
type HealthBelowNode struct {
	name      string
	threshold float64 // 0.0-1.0 percentage
}

// NewHealthBelowNode creates a condition node that checks health percentage.
func NewHealthBelowNode(name string, threshold float64) *HealthBelowNode {
	return &HealthBelowNode{
		name:      name,
		threshold: threshold,
	}
}

// Tick checks if entity health is below threshold.
func (n *HealthBelowNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	healthComp, ok := e.GetComponent("health")
	if !ok {
		return NodeFailure
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok || health.Max <= 0 {
		return NodeFailure
	}

	ratio := health.Current / health.Max
	if ratio < n.threshold {
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op for condition nodes.
func (n *HealthBelowNode) Reset() {}

// String returns the node description.
func (n *HealthBelowNode) String() string {
	return fmt.Sprintf("HealthBelow(%s, %.0f%%)", n.name, n.threshold*100)
}

// HasTargetNode checks if the entity has a valid target in the blackboard.
type HasTargetNode struct {
	name string
}

// NewHasTargetNode creates a condition node that checks for a target.
func NewHasTargetNode(name string) *HasTargetNode {
	return &HasTargetNode{name: name}
}

// Tick checks if blackboard has a valid target.
func (n *HasTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	target, ok := blackboard.Get("target")
	if !ok || target == nil {
		return NodeFailure
	}
	// Check if target is still valid (not nil entity)
	if targetEntity, ok := target.(*Entity); ok && targetEntity != nil {
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op for condition nodes.
func (n *HasTargetNode) Reset() {}

// String returns the node description.
func (n *HasTargetNode) String() string {
	return fmt.Sprintf("HasTarget(%s)", n.name)
}

// InRangeNode checks if the target is within a specified distance.
type InRangeNode struct {
	name     string
	distance float64
}

// NewInRangeNode creates a condition node that checks target distance.
func NewInRangeNode(name string, distance float64) *InRangeNode {
	return &InRangeNode{
		name:     name,
		distance: distance,
	}
}

// Tick checks if target is within range.
func (n *InRangeNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	if tp.Dist <= n.distance {
		blackboard.Set("target_distance", tp.Dist)
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op for condition nodes.
func (n *InRangeNode) Reset() {}

// String returns the node description.
func (n *InRangeNode) String() string {
	return fmt.Sprintf("InRange(%s, %.0f)", n.name, n.distance)
}

// HasAlliesNearbyNode checks if friendly entities are within range.
type HasAlliesNearbyNode struct {
	name     string
	distance float64
	minCount int
}

// NewHasAlliesNearbyNode creates a condition node that checks for nearby allies.
func NewHasAlliesNearbyNode(name string, distance float64, minCount int) *HasAlliesNearbyNode {
	return &HasAlliesNearbyNode{
		name:     name,
		distance: distance,
		minCount: minCount,
	}
}

// Tick checks for nearby allies using faction matching.
func (n *HasAlliesNearbyNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Get entity faction
	factionComp, ok := e.GetComponent("faction")
	if !ok {
		return NodeFailure
	}
	faction, ok := factionComp.(*FactionComponent)
	if !ok {
		return NodeFailure
	}

	// Get our position
	posComp, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Get nearby entities from blackboard (set by AI system during update)
	nearbyEntitiesVal, hasNearby := blackboard.Get("nearby_entities")
	if !hasNearby || nearbyEntitiesVal == nil {
		return NodeFailure
	}

	entities, ok := nearbyEntitiesVal.([]*Entity)
	if !ok {
		return NodeFailure
	}

	allyCount := 0
	for _, other := range entities {
		if other == e {
			continue
		}
		// Check faction match
		otherFaction, ok := other.GetComponent("faction")
		if !ok {
			continue
		}
		of, ok := otherFaction.(*FactionComponent)
		if !ok || of.FactionID != faction.FactionID {
			continue
		}

		// Check distance
		otherPos, ok := other.GetComponent("position")
		if !ok {
			continue
		}
		op, ok := otherPos.(*PositionComponent)
		if !ok {
			continue
		}

		dx := op.X - pos.X
		dy := op.Y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= n.distance {
			allyCount++
		}
	}

	if allyCount >= n.minCount {
		blackboard.Set("ally_count", allyCount)
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op for condition nodes.
func (n *HasAlliesNearbyNode) Reset() {}

// String returns the node description.
func (n *HasAlliesNearbyNode) String() string {
	return fmt.Sprintf("HasAlliesNearby(%s, dist=%.0f, min=%d)", n.name, n.distance, n.minCount)
}

// ==============================================================================
// ACTION NODES - Perform behaviors that modify entity state
// ==============================================================================

// MoveToTargetNode moves entity toward its target.
type MoveToTargetNode struct {
	name       string
	speed      float64
	stopDist   float64
	timeMoving float64
}

// NewMoveToTargetNode creates an action node that moves toward target.
func NewMoveToTargetNode(name string, speed, stopDist float64) *MoveToTargetNode {
	return &MoveToTargetNode{
		name:     name,
		speed:    speed,
		stopDist: stopDist,
	}
}

// Tick moves entity toward target each frame.
func (n *MoveToTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	// Check if close enough
	if tp.Dist <= n.stopDist {
		n.timeMoving = 0
		return NodeSuccess
	}

	// Normalize and move
	if tp.Dist > 0 {
		nx := tp.Dx / tp.Dist
		ny := tp.Dy / tp.Dist
		tp.MyPos.X += nx * n.speed * deltaTime
		tp.MyPos.Y += ny * n.speed * deltaTime
	}

	// Update velocity component if present for animation
	if velComp, ok := e.GetComponent("velocity"); ok {
		if vel, ok := velComp.(*VelocityComponent); ok {
			if tp.Dist > 0 {
				vel.VX = tp.Dx / tp.Dist * n.speed
				vel.VY = tp.Dy / tp.Dist * n.speed
			}
		}
	}

	n.timeMoving += deltaTime
	return NodeRunning
}

// Reset resets movement timer.
func (n *MoveToTargetNode) Reset() {
	n.timeMoving = 0
}

// String returns the node description.
func (n *MoveToTargetNode) String() string {
	return fmt.Sprintf("MoveToTarget(%s, speed=%.0f)", n.name, n.speed)
}

// FleeFromTargetNode moves entity away from its target.
type FleeFromTargetNode struct {
	name       string
	speed      float64
	safeDist   float64
	timeMoving float64
}

// NewFleeFromTargetNode creates an action node that flees from target.
func NewFleeFromTargetNode(name string, speed, safeDist float64) *FleeFromTargetNode {
	return &FleeFromTargetNode{
		name:     name,
		speed:    speed,
		safeDist: safeDist,
	}
}

// Tick moves entity away from target each frame.
func (n *FleeFromTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	// Check if far enough (note: for fleeing, we want distance AWAY from target)
	if tp.Dist >= n.safeDist {
		n.timeMoving = 0
		return NodeSuccess
	}

	// Normalize and move away (reverse direction)
	if tp.Dist > 0 {
		nx := -tp.Dx / tp.Dist
		ny := -tp.Dy / tp.Dist
		tp.MyPos.X += nx * n.speed * deltaTime
		tp.MyPos.Y += ny * n.speed * deltaTime
	} else {
		// If exactly on top, pick random direction
		angle := blackboard.GetRNG().Float64() * 2 * math.Pi
		tp.MyPos.X += math.Cos(angle) * n.speed * deltaTime
		tp.MyPos.Y += math.Sin(angle) * n.speed * deltaTime
	}

	n.timeMoving += deltaTime
	return NodeRunning
}

// Reset resets movement timer.
func (n *FleeFromTargetNode) Reset() {
	n.timeMoving = 0
}

// String returns the node description.
func (n *FleeFromTargetNode) String() string {
	return fmt.Sprintf("FleeFromTarget(%s, speed=%.0f, safe=%.0f)", n.name, n.speed, n.safeDist)
}

// SeekCoverNode moves entity toward nearest cover position.
type SeekCoverNode struct {
	name          string
	speed         float64
	coverDistance float64
	timeMoving    float64
}

// NewSeekCoverNode creates an action node that seeks cover.
func NewSeekCoverNode(name string, speed, coverDist float64) *SeekCoverNode {
	return &SeekCoverNode{
		name:          name,
		speed:         speed,
		coverDistance: coverDist,
	}
}

// Tick finds and moves toward cover position each frame.
func (n *SeekCoverNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Check if already at cover
	atCoverVal, hasAtCover := blackboard.Get("at_cover")
	if hasAtCover && atCoverVal != nil {
		if ac, ok := atCoverVal.(bool); ok && ac {
			return NodeSuccess
		}
	}

	// Get entity position
	myPos, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	myPosComp, ok := myPos.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Get or calculate cover position
	coverPosVal, hasCover := blackboard.Get("cover_position")
	var coverX, coverY float64

	if !hasCover || coverPosVal == nil {
		// Find cover position - move perpendicular to threat direction
		targetVal, hasTarget := blackboard.Get("target")
		if !hasTarget || targetVal == nil {
			// No threat, just hold position
			blackboard.Set("at_cover", true)
			return NodeSuccess
		}

		targetEntity, ok := targetVal.(*Entity)
		if !ok || targetEntity == nil {
			blackboard.Set("at_cover", true)
			return NodeSuccess
		}

		targetPos, ok := targetEntity.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		targetPosComp, ok := targetPos.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Calculate perpendicular direction for flanking cover
		dx := myPosComp.X - targetPosComp.X
		dy := myPosComp.Y - targetPosComp.Y

		// Perpendicular: rotate 90 degrees + add distance away
		perpX := -dy
		perpY := dx
		perpLen := math.Sqrt(perpX*perpX + perpY*perpY)
		if perpLen > 0 {
			perpX /= perpLen
			perpY /= perpLen
		}

		// Pick left or right perpendicular randomly
		if blackboard.GetRNG().Float64() < 0.5 {
			perpX = -perpX
			perpY = -perpY
		}

		coverX = myPosComp.X + perpX*n.coverDistance
		coverY = myPosComp.Y + perpY*n.coverDistance
		blackboard.Set("cover_position", []float64{coverX, coverY})
	} else {
		cp, ok := coverPosVal.([]float64)
		if !ok || len(cp) < 2 {
			return NodeFailure
		}
		coverX = cp[0]
		coverY = cp[1]
	}

	// Move toward cover
	dx := coverX - myPosComp.X
	dy := coverY - myPosComp.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 5.0 { // Reached cover
		blackboard.Set("at_cover", true)
		blackboard.Set("cover_position", nil)
		n.timeMoving = 0
		return NodeSuccess
	}

	// Move toward cover position
	if dist > 0 {
		nx := dx / dist
		ny := dy / dist
		myPosComp.X += nx * n.speed * deltaTime
		myPosComp.Y += ny * n.speed * deltaTime
	}

	n.timeMoving += deltaTime
	return NodeRunning
}

// Reset resets movement state.
func (n *SeekCoverNode) Reset() {
	n.timeMoving = 0
}

// String returns the node description.
func (n *SeekCoverNode) String() string {
	return fmt.Sprintf("SeekCover(%s, speed=%.0f)", n.name, n.speed)
}

// FlankTargetNode moves entity to flank position relative to target.
type FlankTargetNode struct {
	name       string
	speed      float64
	flankDist  float64
	timeMoving float64
}

// NewFlankTargetNode creates an action node that flanks the target.
func NewFlankTargetNode(name string, speed, flankDist float64) *FlankTargetNode {
	return &FlankTargetNode{
		name:      name,
		speed:     speed,
		flankDist: flankDist,
	}
}

// Tick moves entity to flank position each frame.
func (n *FlankTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	targetVal, ok := blackboard.Get("target")
	if !ok || targetVal == nil {
		return NodeFailure
	}
	targetEntity, ok := targetVal.(*Entity)
	if !ok || targetEntity == nil {
		return NodeFailure
	}

	// Get positions
	myPos, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	myPosComp, ok := myPos.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	targetPos, ok := targetEntity.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	targetPosComp, ok := targetPos.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Calculate flank position - perpendicular to direct line at flankDist
	dx := myPosComp.X - targetPosComp.X
	dy := myPosComp.Y - targetPosComp.Y

	// Get perpendicular direction
	perpX := -dy
	perpY := dx
	perpLen := math.Sqrt(perpX*perpX + perpY*perpY)
	if perpLen > 0 {
		perpX /= perpLen
		perpY /= perpLen
	}

	// Flank position: beside target at flankDist perpendicular
	flankX := targetPosComp.X + perpX*n.flankDist
	flankY := targetPosComp.Y + perpY*n.flankDist

	// Check current distance to flank position
	flankDx := flankX - myPosComp.X
	flankDy := flankY - myPosComp.Y
	flankDist := math.Sqrt(flankDx*flankDx + flankDy*flankDy)

	if flankDist <= 10.0 { // Reached flank position
		n.timeMoving = 0
		return NodeSuccess
	}

	// Move toward flank position
	if flankDist > 0 {
		nx := flankDx / flankDist
		ny := flankDy / flankDist
		myPosComp.X += nx * n.speed * deltaTime
		myPosComp.Y += ny * n.speed * deltaTime
	}

	n.timeMoving += deltaTime
	return NodeRunning
}

// Reset resets movement timer.
func (n *FlankTargetNode) Reset() {
	n.timeMoving = 0
}

// String returns the node description.
func (n *FlankTargetNode) String() string {
	return fmt.Sprintf("FlankTarget(%s, dist=%.0f)", n.name, n.flankDist)
}

// PatrolNode moves entity along a patrol path.
type PatrolNode struct {
	name         string
	speed        float64
	waypointIdx  int
	waypoints    [][]float64 // [[x1,y1], [x2,y2], ...]
	waitTime     float64
	currentWait  float64
	waitDuration float64
}

// NewPatrolNode creates an action node that patrols waypoints.
func NewPatrolNode(name string, speed, waitDuration float64) *PatrolNode {
	return &PatrolNode{
		name:         name,
		speed:        speed,
		waypointIdx:  0,
		waypoints:    nil,
		waitDuration: waitDuration,
	}
}

// Tick moves entity along patrol path.
func (n *PatrolNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Get waypoints from blackboard if not set
	if n.waypoints == nil {
		wp, hasWP := blackboard.Get("patrol_waypoints")
		if !hasWP || wp == nil {
			return NodeFailure
		}
		waypoints, ok := wp.([][]float64)
		if !ok || len(waypoints) == 0 {
			return NodeFailure
		}
		n.waypoints = waypoints
	}

	// Wait at waypoint
	if n.currentWait > 0 {
		n.currentWait -= deltaTime
		return NodeRunning
	}

	// Get entity position
	posComp, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Get current waypoint
	currentWP := n.waypoints[n.waypointIdx]
	if len(currentWP) < 2 {
		return NodeFailure
	}
	wpX, wpY := currentWP[0], currentWP[1]

	// Move toward waypoint
	dx := wpX - pos.X
	dy := wpY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 5.0 { // Reached waypoint
		// Start wait timer
		n.currentWait = n.waitDuration
		// Move to next waypoint (loop)
		n.waypointIdx = (n.waypointIdx + 1) % len(n.waypoints)
		return NodeRunning
	}

	// Move toward waypoint
	if dist > 0 {
		nx := dx / dist
		ny := dy / dist
		pos.X += nx * n.speed * deltaTime
		pos.Y += ny * n.speed * deltaTime
	}

	return NodeRunning
}

// Reset resets patrol state.
func (n *PatrolNode) Reset() {
	n.waypointIdx = 0
	n.currentWait = 0
}

// String returns the node description.
func (n *PatrolNode) String() string {
	return fmt.Sprintf("Patrol(%s, speed=%.0f, wait=%.1fs)", n.name, n.speed, n.waitDuration)
}

// AttackTargetNode performs an attack on the target.
type AttackTargetNode struct {
	name        string
	attackRange float64
	damage      int
	cooldown    float64
	currentCD   float64
	attackType  string // "melee", "ranged"
	logger      *logrus.Entry
}

// NewAttackTargetNode creates an action node that attacks the target.
func NewAttackTargetNode(name string, attackRange float64, damage int, cooldown float64, attackType string) *AttackTargetNode {
	return &AttackTargetNode{
		name:        name,
		attackRange: attackRange,
		damage:      damage,
		cooldown:    cooldown,
		attackType:  attackType,
	}
}

// Tick attempts to attack target each frame.
func (n *AttackTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Update cooldown
	if n.currentCD > 0 {
		n.currentCD -= deltaTime
		return NodeRunning
	}

	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	if tp.Dist > n.attackRange {
		return NodeFailure // Out of range
	}

	// Deal damage to target
	targetHealth, ok := tp.Target.GetComponent("health")
	if !ok {
		return NodeFailure
	}
	health, ok := targetHealth.(*HealthComponent)
	if !ok {
		return NodeFailure
	}

	health.Current -= float64(n.damage)
	if health.Current < 0 {
		health.Current = 0
	}

	// Set attack event in blackboard for visual feedback
	blackboard.Set("last_attack", map[string]interface{}{
		"attacker":    e.ID,
		"target":      tp.Target.ID,
		"damage":      n.damage,
		"attack_type": n.attackType,
	})

	// Start cooldown
	n.currentCD = n.cooldown
	return NodeSuccess
}

// Reset resets attack cooldown.
func (n *AttackTargetNode) Reset() {
	n.currentCD = 0
}

// String returns the node description.
func (n *AttackTargetNode) String() string {
	return fmt.Sprintf("AttackTarget(%s, dmg=%d, cd=%.1fs)", n.name, n.damage, n.cooldown)
}

// CallForHelpNode signals allies within range.
type CallForHelpNode struct {
	name   string
	radius float64
	called bool
}

// NewCallForHelpNode creates an action node that calls nearby allies.
func NewCallForHelpNode(name string, radius float64) *CallForHelpNode {
	return &CallForHelpNode{
		name:   name,
		radius: radius,
	}
}

// Tick broadcasts help signal to nearby allies.
func (n *CallForHelpNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	if n.called {
		return NodeSuccess // Already called this tick
	}

	// Get entity faction
	factionComp, ok := e.GetComponent("faction")
	if !ok {
		return NodeFailure
	}
	faction, ok := factionComp.(*FactionComponent)
	if !ok {
		return NodeFailure
	}

	// Get our position
	posComp, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Get target to share
	targetVal, _ := blackboard.Get("target")

	// Store help call event for AI system to process
	blackboard.Set("help_call", map[string]interface{}{
		"caller":   e.ID,
		"faction":  faction.FactionID,
		"position": []float64{pos.X, pos.Y},
		"radius":   n.radius,
		"target":   targetVal,
	})

	n.called = true
	return NodeSuccess
}

// Reset allows another help call.
func (n *CallForHelpNode) Reset() {
	n.called = false
}

// String returns the node description.
func (n *CallForHelpNode) String() string {
	return fmt.Sprintf("CallForHelp(%s, radius=%.0f)", n.name, n.radius)
}

// WaitNode waits for a specified duration.
type WaitNode struct {
	name     string
	duration float64
	elapsed  float64
}

// NewWaitNode creates an action node that waits.
func NewWaitNode(name string, duration float64) *WaitNode {
	return &WaitNode{
		name:     name,
		duration: duration,
	}
}

// Tick waits for duration.
func (n *WaitNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	n.elapsed += deltaTime
	if n.elapsed >= n.duration {
		n.elapsed = 0
		return NodeSuccess
	}
	return NodeRunning
}

// Reset resets wait timer.
func (n *WaitNode) Reset() {
	n.elapsed = 0
}

// String returns the node description.
func (n *WaitNode) String() string {
	return fmt.Sprintf("Wait(%s, %.1fs)", n.name, n.duration)
}

// RandomSelectorNode selects a random child to execute.
type RandomSelectorNode struct {
	name     string
	children []BehaviorNode
	selected int
	running  bool
}

// NewRandomSelectorNode creates a random selector node.
func NewRandomSelectorNode(name string, children ...BehaviorNode) *RandomSelectorNode {
	return &RandomSelectorNode{
		name:     name,
		children: children,
		selected: -1,
	}
}

// Tick selects and executes a random child.
func (r *RandomSelectorNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	if len(r.children) == 0 {
		return NodeFailure
	}

	// Select random child if not already running one
	if !r.running {
		r.selected = blackboard.GetRNG().Intn(len(r.children))
		r.running = true
	}

	status := r.children[r.selected].Tick(entity, blackboard, deltaTime)
	if status != NodeRunning {
		r.running = false
	}
	return status
}

// Reset resets selection state.
func (r *RandomSelectorNode) Reset() {
	r.selected = -1
	r.running = false
	for _, child := range r.children {
		child.Reset()
	}
}

// String returns the node description.
func (r *RandomSelectorNode) String() string {
	return fmt.Sprintf("RandomSelector(%s, %d children)", r.name, len(r.children))
}

// SucceederNode always returns success.
type SucceederNode struct {
	name  string
	child BehaviorNode
}

// NewSucceederNode creates a decorator that always succeeds.
func NewSucceederNode(name string, child BehaviorNode) *SucceederNode {
	return &SucceederNode{
		name:  name,
		child: child,
	}
}

// Tick executes child and returns success regardless.
func (s *SucceederNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	status := s.child.Tick(entity, blackboard, deltaTime)
	if status == NodeRunning {
		return NodeRunning
	}
	return NodeSuccess
}

// Reset resets child.
func (s *SucceederNode) Reset() {
	s.child.Reset()
}

// String returns the node description.
func (s *SucceederNode) String() string {
	return fmt.Sprintf("Succeeder(%s)", s.name)
}

// FailerNode always returns failure.
type FailerNode struct {
	name  string
	child BehaviorNode
}

// NewFailerNode creates a decorator that always fails.
func NewFailerNode(name string, child BehaviorNode) *FailerNode {
	return &FailerNode{
		name:  name,
		child: child,
	}
}

// Tick executes child and returns failure regardless.
func (f *FailerNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	status := f.child.Tick(entity, blackboard, deltaTime)
	if status == NodeRunning {
		return NodeRunning
	}
	return NodeFailure
}

// Reset resets child.
func (f *FailerNode) Reset() {
	f.child.Reset()
}

// String returns the node description.
func (f *FailerNode) String() string {
	return fmt.Sprintf("Failer(%s)", f.name)
}
