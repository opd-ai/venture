// Package engine provides advanced behavior tree action and condition nodes.
// These nodes implement intelligent AI behaviors for tactical decision-making
// including item usage, environmental interaction, and formation tactics.
package engine

import (
	"fmt"
	"math"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// ==============================================================================
// ADVANCED ACTION NODES - Complex tactical behaviors
// ==============================================================================

// UseConsumableNode checks inventory for consumables and uses them when conditions are met.
type UseConsumableNode struct {
	name            string
	consumableType  item.ConsumableType
	healthThreshold float64 // 0-1 ratio, only used for health items
	cooldown        float64 // Minimum time between uses
	currentCooldown float64
	logger          *logrus.Entry
}

// NewUseConsumableNode creates a node that uses consumables from inventory.
// For potions, healthThreshold specifies when to use (e.g., 0.3 = when below 30% health).
func NewUseConsumableNode(name string, consumableType item.ConsumableType, healthThreshold, cooldown float64) *UseConsumableNode {
	return &UseConsumableNode{
		name:            name,
		consumableType:  consumableType,
		healthThreshold: healthThreshold,
		cooldown:        cooldown,
		currentCooldown: 0,
	}
}

// Tick attempts to use a consumable from inventory.
func (n *UseConsumableNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Update cooldown
	if n.currentCooldown > 0 {
		n.currentCooldown -= deltaTime
		return NodeFailure // On cooldown, fail to try other options
	}

	// Get health to check if we should use healing items
	healthComp, hasHealth := e.GetComponent("health")
	if !hasHealth {
		return NodeFailure
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok || health.Max <= 0 {
		return NodeFailure
	}

	healthRatio := health.Current / health.Max

	// For potions/food, only use when below threshold
	if n.consumableType == item.ConsumablePotion || n.consumableType == item.ConsumableFood {
		if healthRatio >= n.healthThreshold {
			return NodeFailure // Health is fine, don't use
		}
	}

	// Get inventory
	invComp, hasInv := e.GetComponent("inventory")
	if !hasInv {
		return NodeFailure
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok || len(inv.Items) == 0 {
		return NodeFailure
	}

	// Find matching consumable
	var foundItem *item.Item
	var foundIndex int = -1
	for idx, itm := range inv.Items {
		if itm.Type == item.TypeConsumable && itm.ConsumableType == n.consumableType {
			foundItem = itm
			foundIndex = idx
			break
		}
	}

	if foundItem == nil || foundIndex < 0 {
		return NodeFailure // No matching consumable
	}

	// Use the consumable
	switch n.consumableType {
	case item.ConsumablePotion, item.ConsumableFood:
		// Heal based on item value/tier (simple formula)
		healAmount := float64(foundItem.Stats.Value) * 0.5
		if healAmount < 10 {
			healAmount = 10
		}
		health.Current += healAmount
		if health.Current > health.Max {
			health.Current = health.Max
		}
	case item.ConsumableBomb:
		// Store bomb use event in blackboard for other systems to process
		posComp, hasPos := e.GetComponent("position")
		if hasPos {
			if pos, ok := posComp.(*PositionComponent); ok {
				blackboard.Set("item_used", map[string]interface{}{
					"type":     "bomb",
					"item":     foundItem,
					"position": []float64{pos.X, pos.Y},
					"entity":   e.ID,
				})
			}
		}
	case item.ConsumableScroll:
		// Store scroll use event for spell effect processing
		blackboard.Set("item_used", map[string]interface{}{
			"type":           "scroll",
			"item":           foundItem,
			"entity":         e.ID,
			"spell_effect":   foundItem.SpellEffectID,
			"spell_duration": foundItem.SpellDuration,
		})
	}

	// Remove item from inventory
	inv.RemoveItem(foundIndex)

	// Start cooldown
	n.currentCooldown = n.cooldown

	// Log the action
	blackboard.Set("last_item_use", map[string]interface{}{
		"entity": e.ID,
		"item":   foundItem.Name,
		"type":   n.consumableType.String(),
	})

	return NodeSuccess
}

// Reset resets cooldown.
func (n *UseConsumableNode) Reset() {
	n.currentCooldown = 0
}

// String returns the node description.
func (n *UseConsumableNode) String() string {
	return fmt.Sprintf("UseConsumable(%s, type=%s, threshold=%.0f%%)",
		n.name, n.consumableType.String(), n.healthThreshold*100)
}

// RetreatToAllyNode moves the entity toward the nearest ally for protection.
type RetreatToAllyNode struct {
	name         string
	speed        float64
	searchRadius float64
	minAllyDist  float64 // Stop when this close to ally
	timeMoving   float64
}

// NewRetreatToAllyNode creates a node that retreats toward allies.
func NewRetreatToAllyNode(name string, speed, searchRadius, minAllyDist float64) *RetreatToAllyNode {
	return &RetreatToAllyNode{
		name:         name,
		speed:        speed,
		searchRadius: searchRadius,
		minAllyDist:  minAllyDist,
	}
}

// Tick moves entity toward nearest ally.
func (n *RetreatToAllyNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	params := btMovementParams{Speed: n.speed, StopDist: n.minAllyDist}
	return runMovementTick(e, blackboard, deltaTime, params, &n.timeMoving,
		func(ctx btAllyCtx, _ *Entity, _ *Blackboard) (float64, float64, bool) {
			var nearestAlly *Entity
			nearestDist := n.searchRadius + 1
			for _, other := range ctx.Nearby {
				if other == e {
					continue
				}
				otherFaction, hasF := other.GetComponent("faction")
				if !hasF {
					continue
				}
				of, ok := otherFaction.(*FactionComponent)
				if !ok || of.FactionID != ctx.Faction.FactionID {
					continue
				}
				otherPos, hasP := other.GetComponent("position")
				if !hasP {
					continue
				}
				op, ok := otherPos.(*PositionComponent)
				if !ok {
					continue
				}
				dx := op.X - ctx.Pos.X
				dy := op.Y - ctx.Pos.Y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < nearestDist && dist > n.minAllyDist {
					nearestAlly = other
					nearestDist = dist
				}
			}
			if nearestAlly == nil {
				return 0, 0, false
			}
			allyPos, _ := nearestAlly.GetComponent("position")
			ap := allyPos.(*PositionComponent)
			return ap.X, ap.Y, true
		},
	)
}

// Reset resets movement timer.
func (n *RetreatToAllyNode) Reset() {
	n.timeMoving = 0
}

// String returns the node description.
func (n *RetreatToAllyNode) String() string {
	return fmt.Sprintf("RetreatToAlly(%s, radius=%.0f)", n.name, n.searchRadius)
}

// InteractWithEnvironmentNode attempts to interact with nearby objects.
type InteractWithEnvironmentNode struct {
	name             string
	interactionRange float64
	interactionType  string // "lever", "trap", "door", "hazard"
	cooldown         float64
	currentCooldown  float64
}

// NewInteractWithEnvironmentNode creates a node for environment interaction.
func NewInteractWithEnvironmentNode(name, interactionType string, range_, cooldown float64) *InteractWithEnvironmentNode {
	return &InteractWithEnvironmentNode{
		name:             name,
		interactionType:  interactionType,
		interactionRange: range_,
		cooldown:         cooldown,
	}
}

// Tick attempts to interact with a nearby environmental object.
func (n *InteractWithEnvironmentNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Update cooldown
	if n.currentCooldown > 0 {
		n.currentCooldown -= deltaTime
		return NodeFailure
	}

	// Get position
	posComp, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Get interactive objects from blackboard (set by environment system)
	interactablesVal, hasInt := blackboard.Get("nearby_interactables")
	if !hasInt || interactablesVal == nil {
		return NodeFailure
	}

	interactables, ok := interactablesVal.([]*Entity)
	if !ok || len(interactables) == 0 {
		return NodeFailure
	}

	// Find matching interactable in range
	for _, interactable := range interactables {
		// Check interaction type tag
		tagComp, hasTag := interactable.GetComponent("tag")
		if !hasTag {
			continue
		}
		if tag, ok := tagComp.(*TagComponent); ok {
			// Check if this matches our interaction type
			matchesType := false
			for _, t := range tag.Tags {
				if t == n.interactionType {
					matchesType = true
					break
				}
			}
			if !matchesType {
				continue
			}
		}

		// Check distance
		intPos, hasPos := interactable.GetComponent("position")
		if !hasPos {
			continue
		}
		ip, ok := intPos.(*PositionComponent)
		if !ok {
			continue
		}

		dx := ip.X - pos.X
		dy := ip.Y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= n.interactionRange {
			// Trigger interaction
			blackboard.Set("environment_interaction", map[string]interface{}{
				"entity":       e.ID,
				"interactable": interactable.ID,
				"type":         n.interactionType,
				"position":     []float64{ip.X, ip.Y},
			})

			n.currentCooldown = n.cooldown
			return NodeSuccess
		}
	}

	return NodeFailure
}

// Reset resets cooldown.
func (n *InteractWithEnvironmentNode) Reset() {
	n.currentCooldown = 0
}

// String returns the node description.
func (n *InteractWithEnvironmentNode) String() string {
	return fmt.Sprintf("InteractEnv(%s, type=%s)", n.name, n.interactionType)
}

// AmbushNode sets up an ambush at a strategic position and waits for targets.
type AmbushNode struct {
	name          string
	waitTime      float64 // Max time to wait in ambush
	elapsedTime   float64
	ambushRange   float64 // Range at which to spring ambush
	setupComplete bool
}

// NewAmbushNode creates a node that sets up ambushes.
func NewAmbushNode(name string, waitTime, ambushRange float64) *AmbushNode {
	return &AmbushNode{
		name:        name,
		waitTime:    waitTime,
		ambushRange: ambushRange,
	}
}

// Tick manages ambush setup and execution.
func (n *AmbushNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Check if we have an ambush position set
	ambushPosVal, hasPos := blackboard.Get("ambush_position")
	if !hasPos || ambushPosVal == nil {
		// No ambush position - try to find one
		// Look for a position with cover near expected enemy paths
		posComp, ok := e.GetComponent("position")
		if !ok {
			return NodeFailure
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			return NodeFailure
		}

		// Use current position or nearby cover as ambush point
		// In a full implementation, this would use pathfinding data
		rng := blackboard.GetRNG()
		offsetX := (rng.Float64() - 0.5) * 100 // Random offset within 50 units
		offsetY := (rng.Float64() - 0.5) * 100

		blackboard.Set("ambush_position", []float64{pos.X + offsetX, pos.Y + offsetY})
		return NodeRunning
	}

	// Get ambush position
	ambushPos, ok := ambushPosVal.([]float64)
	if !ok || len(ambushPos) < 2 {
		return NodeFailure
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

	// Move to ambush position if not there yet
	if !n.setupComplete {
		dx := ambushPos[0] - pos.X
		dy := ambushPos[1] - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > 10 {
			// Move toward ambush position
			speed := 100.0 // Movement speed
			if dist > 0 {
				pos.X += (dx / dist) * speed * deltaTime
				pos.Y += (dy / dist) * speed * deltaTime
			}
			return NodeRunning
		}
		n.setupComplete = true
		blackboard.Set("in_ambush", true)
	}

	// Update wait time
	n.elapsedTime += deltaTime
	if n.elapsedTime >= n.waitTime {
		// Ambush timed out - clear and fail
		n.Reset()
		blackboard.Set("ambush_position", nil)
		blackboard.Set("in_ambush", false)
		return NodeFailure
	}

	// Check for targets in ambush range
	targetVal, hasTarget := blackboard.Get("target")
	if hasTarget && targetVal != nil {
		target, ok := targetVal.(*Entity)
		if ok && target != nil {
			targetPos, hasTP := target.GetComponent("position")
			if hasTP {
				tp, ok := targetPos.(*PositionComponent)
				if ok {
					dx := tp.X - pos.X
					dy := tp.Y - pos.Y
					dist := math.Sqrt(dx*dx + dy*dy)

					if dist <= n.ambushRange {
						// Target in range - spring ambush!
						blackboard.Set("ambush_triggered", map[string]interface{}{
							"attacker": e.ID,
							"target":   target.ID,
							"position": []float64{pos.X, pos.Y},
						})
						blackboard.Set("in_ambush", false)
						n.Reset()
						return NodeSuccess
					}
				}
			}
		}
	}

	return NodeRunning
}

// Reset resets ambush state.
func (n *AmbushNode) Reset() {
	n.elapsedTime = 0
	n.setupComplete = false
}

// String returns the node description.
func (n *AmbushNode) String() string {
	return fmt.Sprintf("Ambush(%s, wait=%.0fs, range=%.0f)", n.name, n.waitTime, n.ambushRange)
}

// FormationNode maintains tactical formation with squad members.
type FormationNode struct {
	name           string
	formationType  FormationType
	spacing        float64 // Distance between formation members
	speed          float64
	formationSlot  int // This entity's slot in formation (-1 = auto-assign)
	timeInPosition float64
}

// NewFormationNode creates a node that maintains formation.
func NewFormationNode(name string, formationType FormationType, spacing, speed float64) *FormationNode {
	return &FormationNode{
		name:          name,
		formationType: formationType,
		spacing:       spacing,
		speed:         speed,
		formationSlot: -1,
	}
}

// Tick maintains formation position relative to squad leader.
func (n *FormationNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Get squad info
	squadComp, hasSquad := e.GetComponent("squad")
	if !hasSquad {
		return NodeFailure
	}
	squad, ok := squadComp.(*SquadComponent)
	if !ok {
		return NodeFailure
	}

	// Get leader from blackboard
	leaderVal, hasLeader := blackboard.Get("squad_leader")
	if !hasLeader || leaderVal == nil {
		// No leader - we might be the leader
		if squad.Role == SquadRoleLeader {
			return NodeSuccess // Leaders don't follow formation
		}
		return NodeFailure
	}

	leader, ok := leaderVal.(*Entity)
	if !ok || leader == nil {
		return NodeFailure
	}

	// Get leader position
	leaderPosComp, hasPos := leader.GetComponent("position")
	if !hasPos {
		return NodeFailure
	}
	leaderPos, ok := leaderPosComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Assign formation slot if not set
	if n.formationSlot < 0 {
		slotVal, hasSlot := blackboard.Get("formation_slot")
		if hasSlot {
			if slot, ok := slotVal.(int); ok {
				n.formationSlot = slot
			}
		}
		if n.formationSlot < 0 {
			n.formationSlot = blackboard.GetRNG().Intn(8) // Random slot 0-7
			blackboard.Set("formation_slot", n.formationSlot)
		}
	}

	// Calculate target position based on formation type
	targetX, targetY := n.calculateFormationPosition(leaderPos.X, leaderPos.Y, n.formationSlot, blackboard)

	// Get entity position
	posComp, ok := e.GetComponent("position")
	if !ok {
		return NodeFailure
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return NodeFailure
	}

	// Move toward formation position
	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 5 {
		// In position
		n.timeInPosition += deltaTime
		// Stop velocity
		if velComp, ok := e.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX = 0
				vel.VY = 0
			}
		}
		return NodeSuccess
	}

	// Move toward position
	if dist > 0 {
		nx := dx / dist
		ny := dy / dist
		pos.X += nx * n.speed * deltaTime
		pos.Y += ny * n.speed * deltaTime

		// Update velocity
		if velComp, ok := e.GetComponent("velocity"); ok {
			if vel, ok := velComp.(*VelocityComponent); ok {
				vel.VX = nx * n.speed
				vel.VY = ny * n.speed
			}
		}
	}

	n.timeInPosition = 0
	return NodeRunning
}

// calculateFormationPosition computes target position for a given slot.
func (n *FormationNode) calculateFormationPosition(leaderX, leaderY float64, slot int, blackboard *Blackboard) (float64, float64) {
	// Get facing direction from blackboard or default to 0 (right)
	facing := 0.0
	if facingVal, hasFacing := blackboard.Get("squad_facing"); hasFacing {
		if f, ok := facingVal.(float64); ok {
			facing = f
		}
	}

	switch n.formationType {
	case FormationLine:
		// Horizontal line perpendicular to facing
		offset := float64(slot-4) * n.spacing // Center around leader
		perpX := -math.Sin(facing)
		perpY := math.Cos(facing)
		return leaderX + perpX*offset, leaderY + perpY*offset

	case FormationColumn:
		// Vertical column behind leader
		offset := float64(slot+1) * n.spacing
		backX := -math.Cos(facing)
		backY := -math.Sin(facing)
		return leaderX + backX*offset, leaderY + backY*offset

	case FormationWedge:
		// V-formation behind leader
		row := (slot / 2) + 1
		side := (slot%2)*2 - 1 // -1 or 1
		backX := -math.Cos(facing)
		backY := -math.Sin(facing)
		perpX := -math.Sin(facing)
		perpY := math.Cos(facing)
		return leaderX + backX*float64(row)*n.spacing + perpX*float64(side*row)*n.spacing*0.5,
			leaderY + backY*float64(row)*n.spacing + perpY*float64(side*row)*n.spacing*0.5

	case FormationCircle:
		// Circle around leader
		angle := float64(slot) * (2 * math.Pi / 8) // 8 positions
		return leaderX + math.Cos(angle)*n.spacing*2, leaderY + math.Sin(angle)*n.spacing*2

	case FormationScatter:
		// Scattered positions using deterministic offsets
		rng := blackboard.GetRNG()
		angle := rng.Float64() * 2 * math.Pi
		dist := n.spacing * (1 + rng.Float64())
		return leaderX + math.Cos(angle)*dist, leaderY + math.Sin(angle)*dist

	default:
		return leaderX, leaderY
	}
}

// Reset resets formation state.
func (n *FormationNode) Reset() {
	n.timeInPosition = 0
}

// String returns the node description.
func (n *FormationNode) String() string {
	typeNames := []string{"Line", "Column", "Wedge", "Circle", "Scatter"}
	typeName := "Unknown"
	if int(n.formationType) < len(typeNames) {
		typeName = typeNames[n.formationType]
	}
	return fmt.Sprintf("Formation(%s, type=%s)", n.name, typeName)
}

// ==============================================================================
// ADVANCED CONDITION NODES - Tactical awareness checks
// ==============================================================================

// HasConsumableNode checks if the entity has a specific consumable type.
type HasConsumableNode struct {
	name           string
	consumableType item.ConsumableType
}

// NewHasConsumableNode creates a condition that checks for consumables.
func NewHasConsumableNode(name string, consumableType item.ConsumableType) *HasConsumableNode {
	return &HasConsumableNode{
		name:           name,
		consumableType: consumableType,
	}
}

// Tick checks inventory for matching consumable.
func (n *HasConsumableNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	invComp, hasInv := e.GetComponent("inventory")
	if !hasInv {
		return NodeFailure
	}
	inv, ok := invComp.(*InventoryComponent)
	if !ok || len(inv.Items) == 0 {
		return NodeFailure
	}

	for _, itm := range inv.Items {
		if itm.Type == item.TypeConsumable && itm.ConsumableType == n.consumableType {
			return NodeSuccess
		}
	}
	return NodeFailure
}

// Reset is a no-op for condition nodes.
func (n *HasConsumableNode) Reset() {}

// String returns the node description.
func (n *HasConsumableNode) String() string {
	return fmt.Sprintf("HasConsumable(%s, %s)", n.name, n.consumableType.String())
}

// IsOutnumberedNode checks if entity is outnumbered by enemies.
type IsOutnumberedNode struct {
	name   string
	range_ float64
	ratio  float64 // e.g., 2.0 means outnumbered if 2x more enemies than allies
}

// NewIsOutnumberedNode creates a condition that checks numerical disadvantage.
func NewIsOutnumberedNode(name string, range_, ratio float64) *IsOutnumberedNode {
	return &IsOutnumberedNode{
		name:   name,
		range_: range_,
		ratio:  ratio,
	}
}

// Tick counts nearby allies and enemies to check ratio.
func (n *IsOutnumberedNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	faction, pos, nearbyEntities, ok := getFactionAndNearbyEntities(e, blackboard)
	if !ok {
		return NodeFailure
	}

	allyCount := 0
	enemyCount := 0

	for _, other := range nearbyEntities {
		if other == e {
			continue
		}

		// Check distance
		otherPos, hasPos := other.GetComponent("position")
		if !hasPos {
			continue
		}
		op, ok := otherPos.(*PositionComponent)
		if !ok {
			continue
		}
		dx := op.X - pos.X
		dy := op.Y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > n.range_ {
			continue
		}

		// Check faction
		otherFaction, hasF := other.GetComponent("faction")
		if !hasF {
			continue
		}
		of, ok := otherFaction.(*FactionComponent)
		if !ok {
			continue
		}

		if of.FactionID == faction.FactionID {
			allyCount++
		} else {
			enemyCount++
		}
	}

	// Check if outnumbered
	if allyCount > 0 && float64(enemyCount)/float64(allyCount) >= n.ratio {
		blackboard.Set("enemy_ratio", float64(enemyCount)/float64(allyCount))
		return NodeSuccess
	}

	return NodeFailure
}

// Reset is a no-op.
func (n *IsOutnumberedNode) Reset() {}

// String returns the node description.
func (n *IsOutnumberedNode) String() string {
	return fmt.Sprintf("IsOutnumbered(%s, ratio=%.1f)", n.name, n.ratio)
}

// IsInCoverNode checks if entity is in a covered position.
type IsInCoverNode struct {
	name string
}

// NewIsInCoverNode creates a condition that checks cover status.
func NewIsInCoverNode(name string) *IsInCoverNode {
	return &IsInCoverNode{name: name}
}

// Tick checks blackboard for cover status.
func (n *IsInCoverNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	atCover, hasCover := blackboard.Get("at_cover")
	if !hasCover {
		return NodeFailure
	}
	if ac, ok := atCover.(bool); ok && ac {
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op.
func (n *IsInCoverNode) Reset() {}

// String returns the node description.
func (n *IsInCoverNode) String() string {
	return fmt.Sprintf("IsInCover(%s)", n.name)
}

// CanSeeTargetNode checks line of sight to target.
type CanSeeTargetNode struct {
	name     string
	maxRange float64
}

// NewCanSeeTargetNode creates a condition that checks visibility.
func NewCanSeeTargetNode(name string, maxRange float64) *CanSeeTargetNode {
	return &CanSeeTargetNode{
		name:     name,
		maxRange: maxRange,
	}
}

// Tick checks if target is visible within range.
func (n *CanSeeTargetNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	if tp.Dist > n.maxRange {
		return NodeFailure
	}

	// Check line of sight (simplified - would use raycasting in full implementation)
	// For now, check blackboard for obstacles
	losBlocked, hasLOS := blackboard.Get("los_blocked")
	if hasLOS {
		if blocked, ok := losBlocked.(bool); ok && blocked {
			return NodeFailure
		}
	}

	return NodeSuccess
}

// Reset is a no-op.
func (n *CanSeeTargetNode) Reset() {}

// String returns the node description.
func (n *CanSeeTargetNode) String() string {
	return fmt.Sprintf("CanSeeTarget(%s, range=%.0f)", n.name, n.maxRange)
}

// IsAmbushingNode checks if entity is in ambush state.
type IsAmbushingNode struct {
	name string
}

// NewIsAmbushingNode creates a condition that checks ambush state.
func NewIsAmbushingNode(name string) *IsAmbushingNode {
	return &IsAmbushingNode{name: name}
}

// Tick checks blackboard for ambush state.
func (n *IsAmbushingNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	inAmbush, has := blackboard.Get("in_ambush")
	if !has {
		return NodeFailure
	}
	if ia, ok := inAmbush.(bool); ok && ia {
		return NodeSuccess
	}
	return NodeFailure
}

// Reset is a no-op.
func (n *IsAmbushingNode) Reset() {}

// String returns the node description.
func (n *IsAmbushingNode) String() string {
	return fmt.Sprintf("IsAmbushing(%s)", n.name)
}

// ==============================================================================
// SQUAD COORDINATION NODES - Multi-entity tactics
// ==============================================================================

// CoordinatedAttackNode synchronizes attack timing with squad.
type CoordinatedAttackNode struct {
	name        string
	attackRange float64
	damage      int
	cooldown    float64
	currentCD   float64
	signalSent  bool
}

// NewCoordinatedAttackNode creates a node for synchronized squad attacks.
func NewCoordinatedAttackNode(name string, range_ float64, damage int, cooldown float64) *CoordinatedAttackNode {
	return &CoordinatedAttackNode{
		name:        name,
		attackRange: range_,
		damage:      damage,
		cooldown:    cooldown,
	}
}

// Tick coordinates attack with squad members.
func (n *CoordinatedAttackNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	// Update cooldown
	if n.currentCD > 0 {
		n.currentCD -= deltaTime
		return NodeRunning
	}

	// Check if squad attack is coordinated
	_, hasSignal := blackboard.Get("squad_attack_signal")
	if !hasSignal && !n.signalSent {
		// Send coordination signal
		blackboard.Set("squad_attack_signal", map[string]interface{}{
			"initiator": e.ID,
			"ready":     true,
		})
		n.signalSent = true
		return NodeRunning
	}

	// Check if all squad members ready
	readyCount, hasReady := blackboard.Get("squad_ready_count")
	squadSize, hasSize := blackboard.Get("squad_size")

	if hasReady && hasSize {
		rc, okR := readyCount.(int)
		ss, okS := squadSize.(int)
		if okR && okS && rc < ss {
			// Wait for others
			return NodeRunning
		}
	}

	// Perform attack
	tp := GetTargetPositions(e, blackboard)
	if tp == nil {
		return NodeFailure
	}

	if tp.Dist > n.attackRange {
		return NodeFailure
	}

	// Deal damage
	targetHealth, ok := tp.Target.GetComponent("health")
	if !ok {
		return NodeFailure
	}
	health, ok := targetHealth.(*HealthComponent)
	if !ok {
		return NodeFailure
	}

	// Bonus damage for coordinated attack
	coordDamage := n.damage + n.damage/4 // 25% bonus for coordination
	health.Current -= float64(coordDamage)
	if health.Current < 0 {
		health.Current = 0
	}

	// Log coordinated attack
	blackboard.Set("coordinated_attack", map[string]interface{}{
		"attacker":    e.ID,
		"target":      tp.Target.ID,
		"damage":      coordDamage,
		"coordinated": true,
	})

	// Reset state
	n.currentCD = n.cooldown
	n.signalSent = false
	blackboard.Set("squad_attack_signal", nil)

	return NodeSuccess
}

// Reset resets coordination state.
func (n *CoordinatedAttackNode) Reset() {
	n.currentCD = 0
	n.signalSent = false
}

// String returns the node description.
func (n *CoordinatedAttackNode) String() string {
	return fmt.Sprintf("CoordinatedAttack(%s, dmg=%d)", n.name, n.damage)
}

// ProtectAllyNode moves to shield a low-health ally.
type ProtectAllyNode struct {
	name            string
	protectRange    float64
	healthThreshold float64
	speed           float64
	timeProtecting  float64
}

// NewProtectAllyNode creates a node that protects wounded allies.
func NewProtectAllyNode(name string, range_, threshold, speed float64) *ProtectAllyNode {
	return &ProtectAllyNode{
		name:            name,
		protectRange:    range_,
		healthThreshold: threshold,
		speed:           speed,
	}
}

// Tick finds and protects low-health allies.
func (n *ProtectAllyNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	e, ok := entityFromContext(entity)
	if !ok {
		return NodeFailure
	}
	faction, pos, nearbyEntities, ok := getFactionAndNearbyEntities(e, blackboard)
	if !ok {
		return NodeFailure
	}

	// Find wounded ally to protect
	var woundedAlly *Entity
	lowestHealth := n.healthThreshold

	for _, other := range nearbyEntities {
		if other == e {
			continue
		}

		// Check faction
		otherFaction, hasF := other.GetComponent("faction")
		if !hasF {
			continue
		}
		of, ok := otherFaction.(*FactionComponent)
		if !ok || of.FactionID != faction.FactionID {
			continue
		}

		// Check health
		otherHealth, hasH := other.GetComponent("health")
		if !hasH {
			continue
		}
		oh, ok := otherHealth.(*HealthComponent)
		if !ok || oh.Max <= 0 {
			continue
		}

		healthRatio := oh.Current / oh.Max
		if healthRatio < lowestHealth {
			// Check distance
			otherPos, hasP := other.GetComponent("position")
			if !hasP {
				continue
			}
			op, ok := otherPos.(*PositionComponent)
			if !ok {
				continue
			}
			dx := op.X - pos.X
			dy := op.Y - pos.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= n.protectRange {
				woundedAlly = other
				lowestHealth = healthRatio
			}
		}
	}

	if woundedAlly == nil {
		return NodeFailure
	}

	// Move to protect ally (position between ally and threat)
	targetVal, hasTarget := blackboard.Get("target")
	allyPos, _ := woundedAlly.GetComponent("position")
	ap := allyPos.(*PositionComponent)

	var protectX, protectY float64
	if hasTarget && targetVal != nil {
		target, ok := targetVal.(*Entity)
		if ok && target != nil {
			targetPos, hasTP := target.GetComponent("position")
			if hasTP {
				tp := targetPos.(*PositionComponent)
				// Position between ally and threat
				protectX = (ap.X + tp.X) / 2
				protectY = (ap.Y + tp.Y) / 2
			} else {
				protectX = ap.X
				protectY = ap.Y
			}
		} else {
			protectX = ap.X
			protectY = ap.Y
		}
	} else {
		// No threat, just stay near ally
		protectX = ap.X
		protectY = ap.Y
	}

	// Move toward protect position
	dx := protectX - pos.X
	dy := protectY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 10 {
		n.timeProtecting += deltaTime
		return NodeSuccess
	}

	if dist > 0 {
		pos.X += (dx / dist) * n.speed * deltaTime
		pos.Y += (dy / dist) * n.speed * deltaTime
	}

	// Store protect target for visual feedback
	blackboard.Set("protecting_ally", woundedAlly.ID)

	n.timeProtecting = 0
	return NodeRunning
}

// Reset resets protection state.
func (n *ProtectAllyNode) Reset() {
	n.timeProtecting = 0
}

// String returns the node description.
func (n *ProtectAllyNode) String() string {
	return fmt.Sprintf("ProtectAlly(%s, threshold=%.0f%%)", n.name, n.healthThreshold*100)
}
