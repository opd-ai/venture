# Code Audit Report

## AUDIT SUMMARY
**Total Issues:** 465
**Resolved:** 130
**Remaining:** 335
**By Category:** CRITICAL BUG: 465
**By Severity:** High: 465 | Medium: 0 | Low: 0

---

## RESOLUTION LOG

### 2025-01-09 - Batches 1-9 (106 issues) - Commits 24627b3 to 4cdc024
- ✅ Completed: mounting_system.go, squad_system.go, vehicle_combat_system.go
- ✅ Completed: rotation_system.go, progression_system.go, movement.go
- ✅ Completed: music_context.go, squad_behaviors.go
- ✅ Completed: item_spawning.go, objective_tracker_system.go
- ✅ Completed: crafting_ui.go
- ✅ Completed: rendering package (sprites/pool.go, cache/sprite_cache.go, particles/pool.go)

### 2025-01-09 - Batch 10: Combat System
**Commit:** (pending)
**Files Fixed:** 1
**Issues Resolved:** 24

- `pkg/engine/combat_system.go` - 24 issues - ✅ RESOLVED
  - All type assertions now use comma-ok idiom
  - Tests passing with xvfb
  - Critical combat system fully secured

---

## DETAILED FINDINGS

**Note:** This audit focuses on the most critical and verified issues found through static analysis.
The primary pattern identified is unchecked type assertions throughout the ECS (Entity-Component-System) implementation.

### Summary of Findings

The codebase contains numerous instances where `GetComponent` returns an interface that is then type-asserted without using the comma-ok idiom. While the component existence is checked, the type assertion itself is unchecked. This creates a vulnerability where unexpected component types could cause runtime panics.

**Recommended Fix:** Add comma-ok checks to all type assertions:
```go
// Instead of:
comp, ok := entity.GetComponent("ai")
if ok {
    aiComp := comp.(*AIComponent)  // Unchecked - can panic!
}

// Use:
comp, ok := entity.GetComponent("ai")
if ok {
    if aiComp, ok := comp.(*AIComponent); ok {
        // Safe to use aiComp
    }
}
```

---

### File: `pkg/engine/ai_system.go` (23 issues)

#### Line 43: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  40: 			continue
  41: 		}
  42: 
> 43: 		aiState := aiComp.(*AIComponent)
  44: 
  45: 		// Update timers
  46: 		aiState.UpdateStateTimer(deltaTime)
```

#### Line 65: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  62: 	if !ok {
  63: 		return // Can't do AI without position
  64: 	}
> 65: 	pos := posComp.(*PositionComponent)
  66: 
  67: 	// Check health for flee condition
  68: 	shouldFlee := ai.shouldFlee(entity, aiComp)
```

#### Line 141: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  138: 	if aiComp.IsWaitingAtWaypoint(deltaTime) {
  139: 		// Stop movement while waiting
  140: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 141: 			vel := velComp.(*VelocityComponent)
  142: 			vel.VX = 0
  143: 			vel.VY = 0
  144: 		}
```

#### Line 161: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  158: 
  159: 	// Move towards waypoint
  160: 	if velComp, ok := entity.GetComponent("velocity"); ok {
> 161: 		vel := velComp.(*VelocityComponent)
  162: 
  163: 		// Use default base speed (velocity component doesn't store speed)
  164: 		baseSpeed := 100.0 // Default pixels per second
```

#### Line 214: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  211: 	if !ok {
  212: 		return
  213: 	}
> 214: 	attack := attackComp.(*AttackComponent)
  215: 
  216: 	// Check if in attack range
  217: 	targetPos, ok := aiComp.Target.GetComponent("position")
```

#### Line 223: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  220: 		aiComp.ChangeState(AIStateIdle)
  221: 		return
  222: 	}
> 223: 	targetP := targetPos.(*PositionComponent)
  224: 
  225: 	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
  226: 
```

#### Line 250: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  247: 	if !ok {
  248: 		return
  249: 	}
> 250: 	attack := attackComp.(*AttackComponent)
  251: 
  252: 	// Check if in attack range
  253: 	targetPos, ok := aiComp.Target.GetComponent("position")
```

#### Line 259: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  256: 		aiComp.ChangeState(AIStateIdle)
  257: 		return
  258: 	}
> 259: 	targetP := targetPos.(*PositionComponent)
  260: 
  261: 	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
  262: 
```

#### Line 273: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  270: 	if attack.CanAttack() {
  271: 		// GAP-018 REPAIR: Set animation to attack when attacking
  272: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 273: 			anim := animComp.(*AnimationComponent)
  274: 			if anim.CurrentState != AnimationStateAttack {
  275: 				anim.SetState(AnimationStateAttack)
  276: 			}
```

#### Line 315: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  312: 		// Stop movement
  313: 		velComp, ok := entity.GetComponent("velocity")
  314: 		if ok {
> 315: 			vel := velComp.(*VelocityComponent)
  316: 			vel.VX = 0
  317: 			vel.VY = 0
  318: 		}
```

#### Line 322: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  319: 
  320: 		// GAP-018 REPAIR: Set animation to idle when stopped
  321: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 322: 			anim := animComp.(*AnimationComponent)
  323: 			if anim.CurrentState != AnimationStateIdle {
  324: 				anim.SetState(AnimationStateIdle)
  325: 			}
```

#### Line 346: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  343: 		return false
  344: 	}
  345: 
> 346: 	health := healthComp.(*HealthComponent)
  347: 	if health.Max <= 0 {
  348: 		return false
  349: 	}
```

#### Line 361: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  358: 	if !ok {
  359: 		return nil // No team component, can't determine enemies
  360: 	}
> 361: 	team := teamComp.(*TeamComponent)
  362: 
  363: 	var nearest *Entity
  364: 	nearestDist := detectionRange
```

#### Line 376: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  373: 		if !ok {
  374: 			continue
  375: 		}
> 376: 		otherT := otherTeam.(*TeamComponent)
  377: 
  378: 		if !team.IsEnemy(otherT.TeamID) {
  379: 			continue
```

#### Line 385: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  382: 		// Check if alive
  383: 		otherHealth, ok := other.GetComponent("health")
  384: 		if ok {
> 385: 			h := otherHealth.(*HealthComponent)
  386: 			if h.IsDead() {
  387: 				continue
  388: 			}
```

#### Line 396: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  393: 		if !ok {
  394: 			continue
  395: 		}
> 396: 		otherP := otherPos.(*PositionComponent)
  397: 
  398: 		dist := ai.getDistance(pos.X, pos.Y, otherP.X, otherP.Y)
  399: 		if dist < nearestDist {
```

#### Line 417: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  414: 	// Check if target is alive
  415: 	targetHealth, ok := target.GetComponent("health")
  416: 	if ok {
> 417: 		h := targetHealth.(*HealthComponent)
  418: 		if h.IsDead() {
  419: 			return false
  420: 		}
```

#### Line 428: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  425: 	if !ok {
  426: 		return false
  427: 	}
> 428: 	targetP := targetPos.(*PositionComponent)
  429: 
  430: 	dist := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
  431: 	return dist <= maxRange
```

#### Line 440: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  437: 	if !ok {
  438: 		return // No velocity component, can't move
  439: 	}
> 440: 	vel := velComp.(*VelocityComponent)
  441: 
  442: 	// Calculate direction
  443: 	dx := targetX - pos.X
```

#### Line 456: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  453: 		// Phase 10.1: Update aim component to face movement direction
  454: 		// This makes enemies visibly rotate towards their target
  455: 		if aimComp, ok := entity.GetComponent("aim"); ok {
> 456: 			aim := aimComp.(*AimComponent)
  457: 			aim.SetAimTarget(targetX, targetY)
  458: 		}
  459: 
```

#### Line 462: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  459: 
  460: 		// GAP-018 REPAIR: Update animation state to walk when moving
  461: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 462: 			anim := animComp.(*AnimationComponent)
  463: 			if anim.CurrentState != AnimationStateWalk {
  464: 				anim.SetState(AnimationStateWalk)
  465: 			}
```

#### Line 481: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  478: func (ai *AISystem) SetDetectionRange(entity *Entity, detectionRange float64) {
  479: 	aiComp, ok := entity.GetComponent("ai")
  480: 	if ok {
> 481: 		aiC := aiComp.(*AIComponent)
  482: 		aiC.DetectionRange = detectionRange
  483: 	}
  484: }
```

#### Line 492: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  489: 	if !ok {
  490: 		return AIStateIdle
  491: 	}
> 492: 	return aiComp.(*AIComponent).State
  493: }
  494: 
```

---

### File: `pkg/engine/animation_system.go` (8 issues)

#### Line 128: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  125: 	var playerX, playerY float64
  126: 	if s.playerEntity != nil {
  127: 		if posComp, ok := s.playerEntity.GetComponent("position"); ok {
> 128: 			pos := posComp.(*PositionComponent)
  129: 			playerX = pos.X
  130: 			playerY = pos.Y
  131: 		}
```

#### Line 139: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  136: 	hasViewport := false
  137: 	if s.enableViewportCull && s.cameraSystem != nil && s.cameraSystem.activeCamera != nil {
  138: 		if camComp, ok := s.cameraSystem.activeCamera.GetComponent("camera"); ok {
> 139: 			camera := camComp.(*CameraComponent)
  140: 			// Calculate viewport bounds with margin for sprites
  141: 			margin := 100.0 // Extra margin to start animating before entity enters view
  142: 			halfWidth := float64(s.cameraSystem.ScreenWidth) / (2.0 * camera.Zoom)
```

#### Line 172: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  169: 		if !hasPos {
  170: 			continue
  171: 		}
> 172: 		pos := posComp.(*PositionComponent)
  173: 
  174: 		// Phase 14.2: Viewport culling check
  175: 		if hasViewport {
```

#### Line 566: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  563: 		// GAP FIX: Determine facing direction based on velocity
  564: 		facing := "down" // Default
  565: 		if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 566: 			vel := velComp.(*VelocityComponent)
  567: 			// Use velocity direction if moving, otherwise keep last facing
  568: 			if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
  569: 				if math.Abs(vel.VX) > math.Abs(vel.VY) {
```

#### Line 597: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  594: 			config.Custom["hasShield"] = false // Could be enhanced to check actual equipment
  595: 		}
  596: 	} else if teamComp, ok := entity.GetComponent("team"); ok {
> 597: 		team := teamComp.(*TeamComponent)
  598: 		if team.TeamID == 2 { // Enemy team
  599: 			// Determine monster type based on entity characteristics
  600: 			entityType := "humanoid" // Default
```

#### Line 604: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  601: 
  602: 			// Check if it's a boss (high damage indicates boss)
  603: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 604: 				attack := attackComp.(*AttackComponent)
  605: 				if attack.Damage > 20 {
  606: 					entityType = "boss"
  607: 					config.Custom["isBoss"] = true
```

#### Line 614: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  611: 
  612: 			// Check size based on collider
  613: 			if colliderComp, ok := entity.GetComponent("collider"); ok {
> 614: 				collider := colliderComp.(*ColliderComponent)
  615: 				if collider.Width > 48 {
  616: 					entityType = "monster" // Large monster
  617: 				} else if collider.Width < 24 {
```

#### Line 627: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  624: 			// GAP FIX: Determine facing direction based on velocity
  625: 			facing := "down" // Default
  626: 			if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 627: 				vel := velComp.(*VelocityComponent)
  628: 				// Use velocity direction if moving, otherwise keep last facing
  629: 				if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
  630: 					if math.Abs(vel.VX) > math.Abs(vel.VY) {
```

---

### File: `pkg/engine/behavior_tree_actions.go` (16 issues)

#### Line 86: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  83: 		if !ok {
  84: 			return false
  85: 		}
> 86: 		health := healthComp.(*HealthComponent)
  87: 		return float64(health.Current)/float64(health.Max) < threshold
  88: 	})
  89: }
```

#### Line 104: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  101: 		if !ok {
  102: 			return false
  103: 		}
> 104: 		pos := posComp.(*PositionComponent)
  105: 
  106: 		targetPos, ok := target.GetComponent("position")
  107: 		if !ok {
```

#### Line 110: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  107: 		if !ok {
  108: 			return false
  109: 		}
> 110: 		tPos := targetPos.(*PositionComponent)
  111: 
  112: 		dx := tPos.X - pos.X
  113: 		dy := tPos.Y - pos.Y
```

#### Line 129: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  126: 		if !ok {
  127: 			return NodeFailure
  128: 		}
> 129: 		pos := posComp.(*PositionComponent)
  130: 
  131: 		// Get team component to identify enemies
  132: 		teamComp, hasTeam := entity.GetComponent("team")
```

#### Line 135: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  132: 		teamComp, hasTeam := entity.GetComponent("team")
  133: 		var teamID int
  134: 		if hasTeam {
> 135: 			teamID = teamComp.(*TeamComponent).TeamID
  136: 		}
  137: 
  138: 		// Find nearest enemy
```

#### Line 156: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  153: 				if !hasOtherTeam {
  154: 					continue
  155: 				}
> 156: 				otherTeamID := otherTeamComp.(*TeamComponent).TeamID
  157: 				if otherTeamID == teamID {
  158: 					continue // Same team
  159: 				}
```

#### Line 166: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  163: 				if !ok {
  164: 					continue
  165: 				}
> 166: 				oPos := otherPos.(*PositionComponent)
  167: 
  168: 				dx := oPos.X - pos.X
  169: 				dy := oPos.Y - pos.Y
```

#### Line 201: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  198: 		if !ok {
  199: 			return NodeFailure
  200: 		}
> 201: 		pos := posComp.(*PositionComponent)
  202: 
  203: 		targetPos, ok := target.GetComponent("position")
  204: 		if !ok {
```

#### Line 207: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  204: 		if !ok {
  205: 			return NodeFailure
  206: 		}
> 207: 		tPos := targetPos.(*PositionComponent)
  208: 
  209: 		// Calculate direction
  210: 		dx := tPos.X - pos.X
```

#### Line 227: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  224: 		if !ok {
  225: 			return NodeFailure
  226: 		}
> 227: 		vel := velComp.(*VelocityComponent)
  228: 		vel.VX = dx
  229: 		vel.VY = dy
  230: 
```

#### Line 248: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  245: 		if !ok {
  246: 			return NodeFailure
  247: 		}
> 248: 		attack := attackComp.(*AttackComponent)
  249: 
  250: 		// Check attack cooldown
  251: 		if attack.CooldownTimer > 0 {
```

#### Line 278: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  275: 		if !ok {
  276: 			return NodeFailure
  277: 		}
> 278: 		pos := posComp.(*PositionComponent)
  279: 
  280: 		targetPos, ok := target.GetComponent("position")
  281: 		if !ok {
```

#### Line 284: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  281: 		if !ok {
  282: 			return NodeFailure
  283: 		}
> 284: 		tPos := targetPos.(*PositionComponent)
  285: 
  286: 		// Calculate direction (away from target)
  287: 		dx := pos.X - tPos.X
```

#### Line 307: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  304: 		if !ok {
  305: 			return NodeFailure
  306: 		}
> 307: 		vel := velComp.(*VelocityComponent)
  308: 		vel.VX = dx
  309: 		vel.VY = dy
  310: 
```

#### Line 342: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  339: 			blackboard.Set("hasWanderTarget", false)
  340: 			velComp, ok := entity.GetComponent("velocity")
  341: 			if ok {
> 342: 				vel := velComp.(*VelocityComponent)
  343: 				vel.VX = 0
  344: 				vel.VY = 0
  345: 			}
```

#### Line 357: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  354: 		if !ok {
  355: 			return NodeFailure
  356: 		}
> 357: 		vel := velComp.(*VelocityComponent)
  358: 		vel.VX = dx
  359: 		vel.VY = dy
  360: 
```

---

### File: `pkg/engine/behavior_tree_system.go` (1 issues)

#### Line 39: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  36: 			continue
  37: 		}
  38: 
> 39: 		btComp := treeComp.(*BehaviorTreeComponent)
  40: 		if !btComp.Enabled {
  41: 			continue
  42: 		}
```

---

### File: `pkg/engine/camera_system.go` (12 issues)

#### Line 91: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  88: 			continue
  89: 		}
  90: 
> 91: 		camera := cameraComp.(*CameraComponent)
  92: 
  93: 		// Get entity position
  94: 		posComp, ok := entity.GetComponent("position")
```

#### Line 98: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  95: 		if !ok {
  96: 			continue
  97: 		}
> 98: 		pos := posComp.(*PositionComponent)
  99: 
  100: 		// Calculate target camera position (entity position + offset)
  101: 		targetX := pos.X + camera.OffsetX
```

#### Line 163: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  160: 			continue
  161: 		}
  162: 
> 163: 		hitStop := hitStopComp.(*HitStopComponent)
  164: 		if hitStop.IsActive() {
  165: 			// Update hit-stop elapsed time with REAL delta time (not scaled)
  166: 			hitStop.Elapsed += deltaTime
```

#### Line 190: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  187: 		return
  188: 	}
  189: 
> 190: 	shake := shakeComp.(*ScreenShakeComponent)
  191: 	if !shake.IsShaking() {
  192: 		return
  193: 	}
```

#### Line 228: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  225: 	if !ok {
  226: 		return worldX, worldY
  227: 	}
> 228: 	camera := cameraComp.(*CameraComponent)
  229: 
  230: 	// Apply camera transform
  231: 	screenX = (worldX - camera.X) * camera.Zoom
```

#### Line 255: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  252: 	if !ok {
  253: 		return screenX, screenY
  254: 	}
> 255: 	camera := cameraComp.(*CameraComponent)
  256: 
  257: 	// Remove screen centering
  258: 	worldX = screenX - float64(s.ScreenWidth)/2
```

#### Line 307: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  304: 	if !ok {
  305: 		return 0, 0
  306: 	}
> 307: 	camera := cameraComp.(*CameraComponent)
  308: 
  309: 	return camera.X, camera.Y
  310: }
```

#### Line 331: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  328: 	if !ok {
  329: 		return
  330: 	}
> 331: 	camera := cameraComp.(*CameraComponent)
  332: 
  333: 	// Add to existing shake (allows stacking)
  334: 	camera.ShakeIntensity += intensity
```

#### Line 360: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  357: 	// Try advanced shake component first
  358: 	shakeComp, ok := s.activeCamera.GetComponent("screenShake")
  359: 	if ok {
> 360: 		advanced := shakeComp.(*ScreenShakeComponent)
  361: 		advanced.TriggerShake(intensity, duration)
  362: 		return
  363: 	}
```

#### Line 389: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  386: 		return // No hit-stop component, silently ignore
  387: 	}
  388: 
> 389: 	hitStop := hitStopComp.(*HitStopComponent)
  390: 	hitStop.TriggerHitStop(duration, timeScale)
  391: }
  392: 
```

#### Line 405: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  402: 		return false
  403: 	}
  404: 
> 405: 	hitStop := hitStopComp.(*HitStopComponent)
  406: 	return hitStop.IsActive()
  407: }
  408: 
```

#### Line 422: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  419: 		return 1.0
  420: 	}
  421: 
> 422: 	hitStop := hitStopComp.(*HitStopComponent)
  423: 	return hitStop.GetTimeScale()
  424: }
  425: 
```

---

### File: `pkg/engine/carry_system.go` (9 issues)

#### Line 79: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  76: 		if !ok {
  77: 			continue
  78: 		}
> 79: 		playerPos := playerPosComp.(*PositionComponent)
  80: 
  81: 		// Get object entity
  82: 		object, ok := s.world.GetEntity(objectID)
```

#### Line 94: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  91: 		if !ok {
  92: 			continue
  93: 		}
> 94: 		objPos := objPosComp.(*PositionComponent)
  95: 
  96: 		// Update object position to follow player (slightly offset above/in front)
  97: 		objPos.X = playerPos.X
```

#### Line 128: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  125: 	if !ok {
  126: 		return false
  127: 	}
> 128: 	carriable := carrComp.(*CarriableComponent)
  129: 
  130: 	// Check if object can be picked up
  131: 	if !carriable.CanPickUp || carriable.IsCarried {
```

#### Line 141: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  138: 
  139: 	// Remove velocity if object was moving
  140: 	if velComp, ok := object.GetComponent("velocity"); ok {
> 141: 		vel := velComp.(*VelocityComponent)
  142: 		vel.VX = 0
  143: 		vel.VY = 0
  144: 	}
```

#### Line 197: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  194: 	if !ok {
  195: 		return
  196: 	}
> 197: 	carriable := carrComp.(*CarriableComponent)
  198: 
  199: 	// Calculate throw velocity based on weight
  200: 	baseVelocity := 300.0 // pixels per second
```

#### Line 216: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  213: 
  214: 	// Set velocity
  215: 	if velComp, ok := object.GetComponent("velocity"); ok {
> 216: 		vel := velComp.(*VelocityComponent)
  217: 		vel.VX = aimX * throwVel
  218: 		vel.VY = aimY * throwVel
  219: 	} else {
```

#### Line 258: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  255: 	if !ok {
  256: 		return
  257: 	}
> 258: 	carriable := carrComp.(*CarriableComponent)
  259: 
  260: 	// Mark as not carried
  261: 	carriable.Drop()
```

#### Line 297: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  294: 		if !ok {
  295: 			continue
  296: 		}
> 297: 		carriable := carrComp.(*CarriableComponent)
  298: 
  299: 		// Skip if already carried or not pickupable
  300: 		if carriable.IsCarried || !carriable.CanPickUp {
```

#### Line 309: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  306: 		if !ok {
  307: 			continue
  308: 		}
> 309: 		pos := posComp.(*PositionComponent)
  310: 
  311: 		// Calculate distance
  312: 		dx := pos.X - x
```

---

### File: `pkg/engine/character_creation.go` (4 issues)

#### Line 1213: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1210: 		return fmt.Errorf("player missing attack component")
  1211: 	}
  1212: 
> 1213: 	health := healthComp.(*HealthComponent)
  1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
  1216: 	attack := attackComp.(*AttackComponent)
```

#### Line 1214: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1211: 	}
  1212: 
  1213: 	health := healthComp.(*HealthComponent)
> 1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
  1216: 	attack := attackComp.(*AttackComponent)
  1217: 
```

#### Line 1215: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1212: 
  1213: 	health := healthComp.(*HealthComponent)
  1214: 	mana := manaComp.(*ManaComponent)
> 1215: 	stats := statsComp.(*StatsComponent)
  1216: 	attack := attackComp.(*AttackComponent)
  1217: 
  1218: 	// Apply class-specific stats
```

#### Line 1216: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1213: 	health := healthComp.(*HealthComponent)
  1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
> 1216: 	attack := attackComp.(*AttackComponent)
  1217: 
  1218: 	// Apply class-specific stats
  1219: 	switch class {
```

---

### File: `pkg/engine/character_ui.go` (4 issues)

#### Line 143: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  140: 		return // Need stats at minimum
  141: 	}
  142: 
> 143: 	stats := statsComp.(*StatsComponent)
  144: 	var equipment *EquipmentComponent
  145: 	if hasEquip {
  146: 		equipment = equipComp.(*EquipmentComponent)
```

#### Line 146: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  143: 	stats := statsComp.(*StatsComponent)
  144: 	var equipment *EquipmentComponent
  145: 	if hasEquip {
> 146: 		equipment = equipComp.(*EquipmentComponent)
  147: 	}
  148: 
  149: 	// Draw semi-transparent overlay
```

#### Line 185: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  182: 
  183: 	// Level and Gold info
  184: 	if hasExp {
> 185: 		exp := expComp.(*ExperienceComponent)
  186: 		levelText := fmt.Sprintf("Level %d", exp.Level)
  187: 		text.Draw(img, levelText, basicfont.Face7x13, panelX+20, titleY+13,
  188: 			color.RGBA{100, 255, 100, 255})
```

#### Line 192: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  189: 	}
  190: 
  191: 	if hasInv {
> 192: 		inv := invComp.(*InventoryComponent)
  193: 		goldText := fmt.Sprintf("Gold: %d", inv.Gold)
  194: 		text.Draw(img, goldText, basicfont.Face7x13, panelX+panelWidth-120, titleY+13,
  195: 			color.RGBA{255, 215, 0, 255})
```

---

### File: `pkg/engine/collision.go` (20 issues)

#### Line 55: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  52: 	}
  53: 
  54: 	colliderComp, _ := entity.GetComponent("collider")
> 55: 	collider := colliderComp.(*ColliderComponent)
  56: 
  57: 	// Only check solid colliders (triggers don't block movement)
  58: 	if !collider.Solid || collider.IsTrigger {
```

#### Line 84: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  81: 	// Get collider components
  82: 	collider1Comp, _ := entity.GetComponent("collider")
  83: 	collider2Comp, _ := other.GetComponent("collider")
> 84: 	collider1 := collider1Comp.(*ColliderComponent)
  85: 	collider2 := collider2Comp.(*ColliderComponent)
  86: 
  87: 	// Skip trigger colliders (they don't block movement)
```

#### Line 85: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  82: 	collider1Comp, _ := entity.GetComponent("collider")
  83: 	collider2Comp, _ := other.GetComponent("collider")
  84: 	collider1 := collider1Comp.(*ColliderComponent)
> 85: 	collider2 := collider2Comp.(*ColliderComponent)
  86: 
  87: 	// Skip trigger colliders (they don't block movement)
  88: 	if collider1.IsTrigger || collider2.IsTrigger {
```

#### Line 106: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  103: 	layer1Comp, hasLayer1 := entity.GetComponent("layer")
  104: 	layer2Comp, hasLayer2 := other.GetComponent("layer")
  105: 	if hasLayer1 && hasLayer2 {
> 106: 		l1 := layer1Comp.(*LayerComponent)
  107: 		l2 := layer2Comp.(*LayerComponent)
  108: 		// Flying entities collide with all layers
  109: 		if !l1.CanFly && !l2.CanFly {
```

#### Line 107: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  104: 	layer2Comp, hasLayer2 := other.GetComponent("layer")
  105: 	if hasLayer1 && hasLayer2 {
  106: 		l1 := layer1Comp.(*LayerComponent)
> 107: 		l2 := layer2Comp.(*LayerComponent)
  108: 		// Flying entities collide with all layers
  109: 		if !l1.CanFly && !l2.CanFly {
  110: 			// Check if entities are on same effective terrain layer
```

#### Line 119: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  116: 
  117: 	// Get other entity's current position
  118: 	pos2Comp, _ := other.GetComponent("position")
> 119: 	pos2 := pos2Comp.(*PositionComponent)
  120: 
  121: 	// Issue #20: Check intersection at predicted position with rotation support
  122: 	rot1Comp, hasRot1 := entity.GetComponent("rotation")
```

#### Line 130: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  127: 		angle1 := 0.0
  128: 		angle2 := 0.0
  129: 		if hasRot1 {
> 130: 			angle1 = rot1Comp.(*RotationComponent).Angle
  131: 		}
  132: 		if hasRot2 {
  133: 			angle2 = rot2Comp.(*RotationComponent).Angle
```

#### Line 133: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  130: 			angle1 = rot1Comp.(*RotationComponent).Angle
  131: 		}
  132: 		if hasRot2 {
> 133: 			angle2 = rot2Comp.(*RotationComponent).Angle
  134: 		}
  135: 		return collider1.IntersectsRotated(newX, newY, angle1, collider2, pos2.X, pos2.Y, angle2)
  136: 	}
```

#### Line 221: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  218: 			layer1Comp, hasLayer1 := entity.GetComponent("layer")
  219: 			layer2Comp, hasLayer2 := other.GetComponent("layer")
  220: 			if hasLayer1 && hasLayer2 {
> 221: 				l1 := layer1Comp.(*LayerComponent)
  222: 				l2 := layer2Comp.(*LayerComponent)
  223: 				// Flying entities collide with all layers
  224: 				if !l1.CanFly && !l2.CanFly {
```

#### Line 222: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  219: 			layer2Comp, hasLayer2 := other.GetComponent("layer")
  220: 			if hasLayer1 && hasLayer2 {
  221: 				l1 := layer1Comp.(*LayerComponent)
> 222: 				l2 := layer2Comp.(*LayerComponent)
  223: 				// Flying entities collide with all layers
  224: 				if !l1.CanFly && !l2.CanFly {
  225: 					// Check if entities are on same effective terrain layer
```

#### Line 243: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  240: 				angle1 := 0.0
  241: 				angle2 := 0.0
  242: 				if hasRot1 {
> 243: 					angle1 = rot1Comp.(*RotationComponent).Angle
  244: 				}
  245: 				if hasRot2 {
  246: 					angle2 = rot2Comp.(*RotationComponent).Angle
```

#### Line 246: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  243: 					angle1 = rot1Comp.(*RotationComponent).Angle
  244: 				}
  245: 				if hasRot2 {
> 246: 					angle2 = rot2Comp.(*RotationComponent).Angle
  247: 				}
  248: 				intersects = collider.IntersectsRotated(pos.X, pos.Y, angle1, otherCollider, otherPos.X, otherPos.Y, angle2)
  249: 			} else {
```

#### Line 395: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  392: 		// Stop horizontal velocity
  393: 		if e1.HasComponent("velocity") {
  394: 			vel1, _ := e1.GetComponent("velocity")
> 395: 			vel1.(*VelocityComponent).VX = 0
  396: 		}
  397: 		if e2.HasComponent("velocity") {
  398: 			vel2, _ := e2.GetComponent("velocity")
```

#### Line 399: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  396: 		}
  397: 		if e2.HasComponent("velocity") {
  398: 			vel2, _ := e2.GetComponent("velocity")
> 399: 			vel2.(*VelocityComponent).VX = 0
  400: 		}
  401: 	} else {
  402: 		// Separate vertically
```

#### Line 414: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  411: 		// Stop vertical velocity
  412: 		if e1.HasComponent("velocity") {
  413: 			vel1, _ := e1.GetComponent("velocity")
> 414: 			vel1.(*VelocityComponent).VY = 0
  415: 		}
  416: 		if e2.HasComponent("velocity") {
  417: 			vel2, _ := e2.GetComponent("velocity")
```

#### Line 418: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  415: 		}
  416: 		if e2.HasComponent("velocity") {
  417: 			vel2, _ := e2.GetComponent("velocity")
> 418: 			vel2.(*VelocityComponent).VY = 0
  419: 		}
  420: 	}
  421: }
```

#### Line 432: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  429: 	posComp, _ := entity.GetComponent("position")
  430: 	colliderComp, _ := entity.GetComponent("collider")
  431: 
> 432: 	pos := posComp.(*PositionComponent)
  433: 	collider := colliderComp.(*ColliderComponent)
  434: 
  435: 	// Try to find a valid position by moving away from walls
```

#### Line 433: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  430: 	colliderComp, _ := entity.GetComponent("collider")
  431: 
  432: 	pos := posComp.(*PositionComponent)
> 433: 	collider := colliderComp.(*ColliderComponent)
  434: 
  435: 	// Try to find a valid position by moving away from walls
  436: 	// Check 8 directions around the entity
```

#### Line 457: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  454: 			// Stop movement in the blocked direction
  455: 			if entity.HasComponent("velocity") {
  456: 				vel, _ := entity.GetComponent("velocity")
> 457: 				velocity := vel.(*VelocityComponent)
  458: 
  459: 				// Stop velocity component that's moving into the wall
  460: 				if dir.dx != 0 {
```

#### Line 474: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  471: 	// If no direction works, stop all movement
  472: 	if entity.HasComponent("velocity") {
  473: 		vel, _ := entity.GetComponent("velocity")
> 474: 		velocity := vel.(*VelocityComponent)
  475: 		velocity.VX = 0
  476: 		velocity.VY = 0
  477: 	}
```

---

### File: `pkg/engine/combat_system.go` (30 issues)

#### Line 100: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  97: 		if !isDead {
  98: 			// Update attack cooldowns only for living entities
  99: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 100: 				attack := attackComp.(*AttackComponent)
  101: 				beforeCooldown := attack.CooldownTimer
  102: 				attack.UpdateCooldown(deltaTime)
  103: 
```

#### Line 118: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  115: 
  116: 		// Process status effects (for both living and dead entities)
  117: 		if statusComp, ok := entity.GetComponent("status_effect"); ok {
> 118: 			status := statusComp.(*StatusEffectComponent)
  119: 
  120: 			// Update status effect
  121: 			if ticked := status.Update(deltaTime); ticked {
```

#### Line 135: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  132: 	// Clean up dead entities
  133: 	for _, entity := range entities {
  134: 		if healthComp, ok := entity.GetComponent("health"); ok {
> 135: 			health := healthComp.(*HealthComponent)
  136: 			if health.IsDead() {
  137: 				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
  138: 					s.logger.WithFields(logrus.Fields{
```

#### Line 158: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  155: 		return
  156: 	}
  157: 
> 158: 	health := healthComp.(*HealthComponent)
  159: 
  160: 	switch effect.EffectType {
  161: 	case "poison", "burn":
```

#### Line 188: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  185: 	if !ok {
  186: 		return false
  187: 	}
> 188: 	attack := attackComp.(*AttackComponent)
  189: 
  190: 	// Check cooldown
  191: 	if !attack.CanAttack() {
```

#### Line 198: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  195: 	// Phase 10.2: Check if attacker has a projectile weapon equipped
  196: 	// If so, spawn a projectile instead of doing instant damage
  197: 	if equipComp, hasEquip := attacker.GetComponent("equipment"); hasEquip {
> 198: 		equipment := equipComp.(*EquipmentComponent)
  199: 		if weapon, hasWeapon := equipment.Slots[SlotMainHand]; hasWeapon && weapon != nil {
  200: 			if weapon.Stats.IsProjectile {
  201: 				// Spawn projectile for ranged weapon
```

#### Line 215: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  212: 	if !ok {
  213: 		return false
  214: 	}
> 215: 	health := targetHealth.(*HealthComponent)
  216: 
  217: 	// Check if target is already dead
  218: 	if health.IsDead() {
```

#### Line 236: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  233: 	attackerStatsComp, _ := attacker.GetComponent("stats")
  234: 	var attackerStats *StatsComponent
  235: 	if attackerStatsComp != nil {
> 236: 		attackerStats = attackerStatsComp.(*StatsComponent)
  237: 	}
  238: 
  239: 	// Get target stats
```

#### Line 243: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  240: 	targetStatsComp, _ := target.GetComponent("stats")
  241: 	var targetStats *StatsComponent
  242: 	if targetStatsComp != nil {
> 243: 		targetStats = targetStatsComp.(*StatsComponent)
  244: 	}
  245: 
  246: 	// Check for evasion
```

#### Line 301: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  298: 
  299: 	// Check for shield first
  300: 	if shieldComp, hasShield := target.GetComponent("shield"); hasShield {
> 301: 		shield := shieldComp.(*ShieldComponent)
  302: 		if shield.IsActive() {
  303: 			// Shield absorbs damage
  304: 			absorbed := shield.AbsorbDamage(finalDamage)
```

#### Line 320: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  317: 
  318: 	// Trigger attack animation for attacker
  319: 	if animComp, hasAnim := attacker.GetComponent("animation"); hasAnim {
> 320: 		anim := animComp.(*AnimationComponent)
  321: 
  322: 		// Log animation trigger for player when debugging
  323: 		if attacker.HasComponent("input") && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
```

#### Line 336: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  333: 		anim.OnComplete = func() {
  334: 			// Check if entity is moving to set appropriate idle/walk state
  335: 			if velComp, hasVel := attacker.GetComponent("velocity"); hasVel {
> 336: 				vel := velComp.(*VelocityComponent)
  337: 				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  338: 				if speed > 0.1 {
  339: 					anim.SetState(AnimationStateWalk)
```

#### Line 355: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  352: 
  353: 	// Trigger hurt animation for target
  354: 	if animComp, hasAnim := target.GetComponent("animation"); hasAnim {
> 355: 		anim := animComp.(*AnimationComponent)
  356: 		anim.SetState(AnimationStateHit)
  357: 		// Set a callback to return to idle after hurt animation
  358: 		anim.OnComplete = func() {
```

#### Line 361: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  358: 		anim.OnComplete = func() {
  359: 			// Check if entity is moving to set appropriate idle/walk state
  360: 			if velComp, hasVel := target.GetComponent("velocity"); hasVel {
> 361: 				vel := velComp.(*VelocityComponent)
  362: 				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  363: 				if speed > 0.1 {
  364: 					anim.SetState(AnimationStateWalk)
```

#### Line 390: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  387: 	// GAP-016 REPAIR: Spawn hit particles at target position
  388: 	if s.particleSystem != nil && s.world != nil {
  389: 		if posComp, ok := target.GetComponent("position"); ok {
> 390: 			pos := posComp.(*PositionComponent)
  391: 			// Use timestamp for particle seed variation
  392: 			particleSeed := s.seed + int64(pos.X*1000) + int64(pos.Y*1000)
  393: 			s.particleSystem.SpawnHitSparks(s.world, pos.X, pos.Y, particleSeed, s.genreID)
```

#### Line 402: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  399: 	if feedbackComp, ok := target.GetComponent("visual_feedback"); ok {
  400: 		// Check accessibility settings for visual flash
  401: 		if s.camera != nil && s.camera.Accessibility.ShouldApplyVisualFlash() {
> 402: 			feedback := feedbackComp.(*VisualFeedbackComponent)
  403: 			// Flash intensity scales with damage (0.3-1.0 range)
  404: 			flashIntensity := 0.3 + (finalDamage / 100.0)
  405: 			if flashIntensity > 1.0 {
```

#### Line 419: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  416: 		targetHealthComp, _ := target.GetComponent("health")
  417: 		var maxHP float64 = 100 // Default
  418: 		if targetHealthComp != nil {
> 419: 			maxHP = targetHealthComp.(*HealthComponent).Max
  420: 		}
  421: 
  422: 		// Calculate shake intensity based on damage relative to max HP
```

#### Line 469: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  466: 	if !ok {
  467: 		return false
  468: 	}
> 469: 	attack := attackComp.(*AttackComponent)
  470: 
  471: 	if !attack.CanAttack() {
  472: 		return false
```

#### Line 476: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  473: 	}
  474: 
  475: 	targetHealth, ok := target.GetComponent("health")
> 476: 	if !ok || targetHealth.(*HealthComponent).IsDead() {
  477: 		return false
  478: 	}
  479: 
```

#### Line 509: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  506: 		return
  507: 	}
  508: 
> 509: 	health := healthComp.(*HealthComponent)
  510: 	health.Heal(amount)
  511: }
  512: 
```

#### Line 533: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  530: 	attackerTeam, _ := attacker.GetComponent("team")
  531: 	var attackerTeamID int
  532: 	if attackerTeam != nil {
> 533: 		attackerTeamID = attackerTeam.(*TeamComponent).TeamID
  534: 	}
  535: 
  536: 	enemies := make([]*Entity, 0)
```

#### Line 551: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  548: 		// Check team
  549: 		targetTeam, hasTeam := entity.GetComponent("team")
  550: 		if hasTeam {
> 551: 			team := targetTeam.(*TeamComponent)
  552: 			if !team.IsEnemy(attackerTeamID) {
  553: 				continue
  554: 			}
```

#### Line 559: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  556: 
  557: 		// Check health
  558: 		healthComp, hasHealth := entity.GetComponent("health")
> 559: 		if !hasHealth || healthComp.(*HealthComponent).IsDead() {
  560: 			continue
  561: 		}
  562: 
```

#### Line 617: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  614: 	if !hasPos {
  615: 		return nil
  616: 	}
> 617: 	pos := attackerPos.(*PositionComponent)
  618: 
  619: 	// Filter enemies by aim cone and find closest
  620: 	var bestEnemy *Entity
```

#### Line 629: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  626: 		if !hasEnemyPos {
  627: 			continue
  628: 		}
> 629: 		ePos := enemyPos.(*PositionComponent)
  630: 
  631: 		// Calculate angle from attacker to enemy
  632: 		dx := ePos.X - pos.X
```

#### Line 672: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  669: 	if !hasPos {
  670: 		return false
  671: 	}
> 672: 	attackerPos := attackerPosComp.(*PositionComponent)
  673: 
  674: 	// Get aim direction
  675: 	var aimAngle float64
```

#### Line 677: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  674: 	// Get aim direction
  675: 	var aimAngle float64
  676: 	if aimComp, hasAim := attacker.GetComponent("aim"); hasAim {
> 677: 		aim := aimComp.(*AimComponent)
  678: 		aimAngle = aim.AimAngle
  679: 	} else if rotComp, hasRot := attacker.GetComponent("rotation"); hasRot {
  680: 		rot := rotComp.(*RotationComponent)
```

#### Line 680: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  677: 		aim := aimComp.(*AimComponent)
  678: 		aimAngle = aim.AimAngle
  679: 	} else if rotComp, hasRot := attacker.GetComponent("rotation"); hasRot {
> 680: 		rot := rotComp.(*RotationComponent)
  681: 		aimAngle = rot.Angle
  682: 	} else {
  683: 		// Fallback: aim at target
```

#### Line 688: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  685: 		if !hasTargetPos {
  686: 			return false
  687: 		}
> 688: 		targetPos := targetPosComp.(*PositionComponent)
  689: 		dx := targetPos.X - attackerPos.X
  690: 		dy := targetPos.Y - attackerPos.Y
  691: 		aimAngle = math.Atan2(dy, dx)
```

#### Line 712: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  709: 
  710: 	// Get attacker stats for bonus damage
  711: 	if attackerStatsComp, hasStats := attacker.GetComponent("stats"); hasStats {
> 712: 		attackerStats := attackerStatsComp.(*StatsComponent)
  713: 		if attack.DamageType == combat.DamageMagical {
  714: 			baseDamage += attackerStats.MagicPower
  715: 		} else {
```

---

### File: `pkg/engine/crafting_system.go` (1 issues)

#### Line 317: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  314: 		return s.createFallbackItem(recipe)
  315: 	}
  316: 
> 317: 	items := result.([]*item.Item)
  318: 	if len(items) == 0 {
  319: 		return s.createFallbackItem(recipe)
  320: 	}
```

---

### File: `pkg/engine/crafting_ui.go` (11 issues)

#### Line 182: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  179: 
  180: 	// Check if player is currently crafting
  181: 	if progressComp, ok := ui.playerEntity.GetComponent("crafting_progress"); ok {
> 182: 		progress := progressComp.(*CraftingProgressComponent)
  183: 		if progress != nil {
  184: 			ui.showingProgress = true
  185: 			return // Don't allow new crafts while one is in progress
```

#### Line 196: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  193: 		ui.showMessage("You don't know any recipes yet")
  194: 		return
  195: 	}
> 196: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  197: 
  198: 	// Convert map to slice for ordered iteration
  199: 	var recipeList []*Recipe
```

#### Line 403: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  400: 		return
  401: 	}
  402: 
> 403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 	skill := skillComp.(*CraftingSkillComponent)
  405: 	inv := invComp.(*InventoryComponent)
  406: 	recipes := knowledge.KnownRecipes
```

#### Line 404: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  401: 	}
  402: 
  403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
> 404: 	skill := skillComp.(*CraftingSkillComponent)
  405: 	inv := invComp.(*InventoryComponent)
  406: 	recipes := knowledge.KnownRecipes
  407: 
```

#### Line 405: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  402: 
  403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 	skill := skillComp.(*CraftingSkillComponent)
> 405: 	inv := invComp.(*InventoryComponent)
  406: 	recipes := knowledge.KnownRecipes
  407: 
  408: 	// Draw semi-transparent overlay
```

#### Line 430: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  427: 	titleText := "CRAFTING RECIPES"
  428: 	if ui.stationEntity != nil {
  429: 		if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
> 430: 			station := stationComp.(*CraftingStationComponent)
  431: 			titleText = fmt.Sprintf("CRAFTING - %s Station (+5%% success, 25%% faster)", station.StationType.String())
  432: 		}
  433: 	}
```

#### Line 465: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  462: 	instructionY := windowY + 90
  463: 	if ui.showingProgress {
  464: 		progressComp, _ := ui.playerEntity.GetComponent("crafting_progress")
> 465: 		progress := progressComp.(*CraftingProgressComponent)
  466: 		if progress != nil {
  467: 			progressPercent := (progress.ElapsedTimeSec / progress.RequiredTimeSec) * 100
  468: 			ebitenutil.DebugPrintAt(img, fmt.Sprintf("Crafting in progress... %.0f%%", progressPercent),
```

#### Line 616: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  613: 		// Show station bonus in success chance
  614: 		if ui.stationEntity != nil {
  615: 			if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
> 616: 				station := stationComp.(*CraftingStationComponent)
  617: 				// Check if station type matches recipe type
  618: 				if station.StationType == recipe.Type {
  619: 					bonusChance := successChance + station.BonusSuccessChance
```

#### Line 692: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  689: 	// Draw nearby station hint if not at a station
  690: 	if ui.stationEntity == nil && ui.playerEntity != nil {
  691: 		if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 692: 			pos := posComp.(*PositionComponent)
  693: 			// Find nearest station within 100 pixels
  694: 			nearestStation, distance := ui.findNearestStation(pos.X, pos.Y, 100)
  695: 			if nearestStation != nil {
```

#### Line 697: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  694: 			nearestStation, distance := ui.findNearestStation(pos.X, pos.Y, 100)
  695: 			if nearestStation != nil {
  696: 				if stationComp, ok := nearestStation.GetComponent("crafting_station"); ok {
> 697: 					station := stationComp.(*CraftingStationComponent)
  698: 					stationHint := fmt.Sprintf("Nearby: %s (%.0f units away) - Move closer to use station bonuses",
  699: 						station.StationType.String(), distance)
  700: 					ebitenutil.DebugPrintAt(img, stationHint, windowX+10, footerY-20)
```

#### Line 821: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  818: 	if !hasInv {
  819: 		return false
  820: 	}
> 821: 	inventory := invComp.(*InventoryComponent)
  822: 
  823: 	// Check if player has all required materials
  824: 	for _, mat := range recipe.Materials {
```

---

### File: `pkg/engine/destructible_object_system.go` (4 issues)

#### Line 96: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  93: 	if !ok {
  94: 		return
  95: 	}
> 96: 	pos := posComp.(*PositionComponent)
  97: 
  98: 	if s.logger != nil {
  99: 		s.logger.WithFields(logrus.Fields{
```

#### Line 140: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  137: 		if !ok {
  138: 			continue
  139: 		}
> 140: 		entityPos := posComp.(*PositionComponent)
  141: 
  142: 		// Calculate distance
  143: 		dx := entityPos.X - x
```

#### Line 286: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  283: 		if !ok {
  284: 			continue
  285: 		}
> 286: 		pos := posComp.(*PositionComponent)
  287: 
  288: 		// Check if within damage radius
  289: 		dx := pos.X - x
```

#### Line 299: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  296: 			if !ok {
  297: 				continue
  298: 			}
> 299: 			destructibleObj := objComp.(*DestructibleObjectComponent)
  300: 
  301: 			// Apply damage
  302: 			destroyed := destructibleObj.TakeDamage(damage)
```

---

### File: `pkg/engine/ecs.go` (6 issues)

#### Line 41: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  38: 	// Update fast-path cache for hot components
  39: 	switch c.Type() {
  40: 	case "position":
> 41: 		e.position = c.(*PositionComponent)
  42: 	case "velocity":
  43: 		e.velocity = c.(*VelocityComponent)
  44: 	case "health":
```

#### Line 43: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  40: 	case "position":
  41: 		e.position = c.(*PositionComponent)
  42: 	case "velocity":
> 43: 		e.velocity = c.(*VelocityComponent)
  44: 	case "health":
  45: 		e.health = c.(*HealthComponent)
  46: 	case "collider":
```

#### Line 45: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  42: 	case "velocity":
  43: 		e.velocity = c.(*VelocityComponent)
  44: 	case "health":
> 45: 		e.health = c.(*HealthComponent)
  46: 	case "collider":
  47: 		e.collider = c.(*ColliderComponent)
  48: 	case "inventory":
```

#### Line 47: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  44: 	case "health":
  45: 		e.health = c.(*HealthComponent)
  46: 	case "collider":
> 47: 		e.collider = c.(*ColliderComponent)
  48: 	case "inventory":
  49: 		e.inventory = c.(*InventoryComponent)
  50: 	case "stats":
```

#### Line 49: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  46: 	case "collider":
  47: 		e.collider = c.(*ColliderComponent)
  48: 	case "inventory":
> 49: 		e.inventory = c.(*InventoryComponent)
  50: 	case "stats":
  51: 		e.stats = c.(*StatsComponent)
  52: 	}
```

#### Line 51: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  48: 	case "inventory":
  49: 		e.inventory = c.(*InventoryComponent)
  50: 	case "stats":
> 51: 		e.stats = c.(*StatsComponent)
  52: 	}
  53: }
  54: 
```

---

### File: `pkg/engine/entity_spawning.go` (1 issues)

#### Line 53: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  50: 		return 0, fmt.Errorf("failed to generate entities: %w", err)
  51: 	}
  52: 
> 53: 	generatedEntities := result.([]*entity.Entity)
  54: 	if len(generatedEntities) == 0 {
  55: 		return 0, nil
  56: 	}
```

---

### File: `pkg/engine/faction_system.go` (4 issues)

#### Line 94: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  91: 		// Player has faction components, find the right one
  92: 		// Note: In the ECS, we'll need to handle multiple faction components
  93: 		// For now, we assume one component per faction via a different approach
> 94: 		fc := comp.(FactionComponent)
  95: 		if fc.FactionID == change.FactionID && fc.IsPlayerFaction {
  96: 			factionComp = &fc
  97: 		}
```

#### Line 175: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  172: 	}
  173: 
  174: 	if comp, ok := playerEntity.GetComponent("faction"); ok {
> 175: 		fc := comp.(FactionComponent)
  176: 		if fc.FactionID == factionID && fc.IsPlayerFaction {
  177: 			return fc.Reputation
  178: 		}
```

#### Line 207: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  204: func (fs *FactionSystem) UpdateNPCHostility(entity *Entity) {
  205: 	// Get NPC's faction
  206: 	if comp, ok := entity.GetComponent("faction"); ok {
> 207: 		fc := comp.(FactionComponent)
  208: 		if !fc.IsPlayerFaction {
  209: 			// This is an NPC faction member
  210: 			if fs.ShouldAttackPlayer(fc.FactionID) {
```

#### Line 232: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  229: 		return // Victim has no faction
  230: 	}
  231: 
> 232: 	victimFaction := comp.(FactionComponent)
  233: 	if victimFaction.IsPlayerFaction {
  234: 		return // Don't process if victim is player
  235: 	}
```

---

### File: `pkg/engine/hazard_system.go` (7 issues)

#### Line 78: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  75: 		if !ok {
  76: 			continue
  77: 		}
> 78: 		hazard := hazComp.(*HazardComponent)
  79: 
  80: 		// Update hazard duration
  81: 		hazard.Update(deltaTime)
```

#### Line 97: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  94: 		if !ok {
  95: 			continue
  96: 		}
> 97: 		hazPos := hazPosComp.(*PositionComponent)
  98: 
  99: 		// Sync zone tracker with current hazard state
  100: 		zone, exists := s.zoneTracker.GetZone(hazardEntity.ID)
```

#### Line 145: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  142: 		if !ok {
  143: 			continue
  144: 		}
> 145: 		entPos := entPosComp.(*PositionComponent)
  146: 
  147: 		// Query zones at entity position
  148: 		zones := s.zoneTracker.GetZonesAt(entPos.X, entPos.Y)
```

#### Line 232: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  229: 		if !ok {
  230: 			continue
  231: 		}
> 232: 		entPos := entPosComp.(*PositionComponent)
  233: 
  234: 		// Calculate distance to hazard
  235: 		dx := entPos.X - hazPos.X
```

#### Line 291: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  288: 	if !ok {
  289: 		return
  290: 	}
> 291: 	entPos := entPosComp.(*PositionComponent)
  292: 
  293: 	zones := s.zoneTracker.GetZonesAt(entPos.X, entPos.Y)
  294: 	for _, zone := range zones {
```

#### Line 350: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  347: 		if !ok {
  348: 			continue
  349: 		}
> 350: 		hazard := hazComp.(*HazardComponent)
  351: 
  352: 		// Get hazard position
  353: 		hazPosComp, ok := hazardEntity.GetComponent("position")
```

#### Line 357: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  354: 		if !ok {
  355: 			continue
  356: 		}
> 357: 		hazPos := hazPosComp.(*PositionComponent)
  358: 
  359: 		// Calculate distance
  360: 		dx := x - hazPos.X
```

---

### File: `pkg/engine/help_system.go` (2 issues)

#### Line 324: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  321: 			if !ok {
  322: 				continue
  323: 			}
> 324: 			health := comp.(*HealthComponent)
  325: 			if health.Current < health.Max*0.25 && !hs.ShowQuickHint {
  326: 				hs.ShowQuickHintFor("low_health")
  327: 			}
```

#### Line 336: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  333: 			if !ok {
  334: 				continue
  335: 			}
> 336: 			inv := comp.(*InventoryComponent)
  337: 			if len(inv.Items) >= inv.MaxItems && !hs.ShowQuickHint {
  338: 				hs.ShowQuickHintFor("inventory_full")
  339: 			}
```

---

### File: `pkg/engine/hud_system.go` (6 issues)

#### Line 97: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  94: 	if !ok {
  95: 		return
  96: 	}
> 97: 	health := healthComp.(*HealthComponent)
  98: 
  99: 	// Health bar dimensions
  100: 	barX := float32(20)
```

#### Line 149: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  146: 
  147: 	// Draw level if available
  148: 	if hasExp {
> 149: 		exp := expComp.(*ExperienceComponent)
  150: 		levelText := fmt.Sprintf("Level: %d", exp.Level)
  151: 		h.drawText(levelText, x, y, color.White)
  152: 		y += lineHeight
```

#### Line 157: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  154: 
  155: 	// Draw stats if available
  156: 	if hasStats {
> 157: 		stats := statsComp.(*StatsComponent)
  158: 		h.drawText(fmt.Sprintf("ATK: %.0f", stats.Attack), x, y, color.RGBA{255, 200, 200, 255})
  159: 		y += lineHeight
  160: 		h.drawText(fmt.Sprintf("DEF: %.0f", stats.Defense), x, y, color.RGBA{200, 200, 255, 255})
```

#### Line 172: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  169: 	if !ok {
  170: 		return
  171: 	}
> 172: 	exp := expComp.(*ExperienceComponent)
  173: 
  174: 	// Experience bar dimensions
  175: 	barX := float32(20)
```

#### Line 243: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  240: 	if !ok {
  241: 		return // No aim component, skip indicator
  242: 	}
> 243: 	aim := aimComp.(*AimComponent)
  244: 
  245: 	// DEBUG: Compare aim vs rotation components
  246: 	if rotComp, ok := h.playerEntity.GetComponent("rotation"); ok {
```

#### Line 247: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  244: 
  245: 	// DEBUG: Compare aim vs rotation components
  246: 	if rotComp, ok := h.playerEntity.GetComponent("rotation"); ok {
> 247: 		rotation := rotComp.(*RotationComponent)
  248: 		fmt.Printf("[DEBUG] HUD: AimAngle=%.4f, RotationAngle=%.4f, RotationTarget=%.4f\n",
  249: 			aim.AimAngle, rotation.Angle, rotation.TargetAngle)
  250: 	}
```

---

### File: `pkg/engine/input_system.go` (4 issues)

#### Line 507: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  504: 			continue
  505: 		}
  506: 
> 507: 		input := inputComp.(*EbitenInput)
  508: 		s.processInput(entity, input, deltaTime)
  509: 	}
  510: }
```

#### Line 658: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  655: 	if entity.HasComponent("aim") && s.cameraSystem != nil {
  656: 		aimComp, ok := entity.GetComponent("aim")
  657: 		if ok {
> 658: 			aim := aimComp.(*AimComponent)
  659: 
  660: 			// Convert screen coordinates to world coordinates
  661: 			worldX, worldY := s.cameraSystem.ScreenToWorld(float64(input.MouseX), float64(input.MouseY))
```

#### Line 676: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  673: 
  674: 	// Apply movement to velocity component if it exists
  675: 	if velComp, ok := entity.GetComponent("velocity"); ok {
> 676: 		velocity := velComp.(*VelocityComponent)
  677: 		velocity.VX = input.MoveX * s.MoveSpeed
  678: 		velocity.VY = input.MoveY * s.MoveSpeed
  679: 
```

#### Line 682: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  679: 
  680: 		// GAP-018 REPAIR: Update player animation based on movement
  681: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 682: 			anim := animComp.(*AnimationComponent)
  683: 			// Check if player is moving
  684: 			isMoving := (velocity.VX != 0 || velocity.VY != 0)
  685: 
```

---

### File: `pkg/engine/interaction_system.go` (6 issues)

#### Line 71: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  68: 	if !ok {
  69: 		return
  70: 	}
> 71: 	playerPos := playerPosComp.(*PositionComponent)
  72: 
  73: 	// Priority order for interactions:
  74: 	// 1. Context actions (doors, levers, etc.)
```

#### Line 103: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  100: 		if !ok {
  101: 			continue
  102: 		}
> 103: 		entPos := entPosComp.(*PositionComponent)
  104: 
  105: 		// Get context action component
  106: 		ctxComp, ok := entity.GetComponent("contextAction")
```

#### Line 110: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  107: 		if !ok {
  108: 			continue
  109: 		}
> 110: 		contextAction := ctxComp.(*ContextActionComponent)
  111: 
  112: 		// Skip if not available or on cooldown
  113: 		if !contextAction.CanInteract() {
```

#### Line 241: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  238: 		if !ok {
  239: 			continue
  240: 		}
> 241: 		elemPos := elemPosComp.(*PositionComponent)
  242: 
  243: 		// Get puzzle element component
  244: 		elemComp, ok := element.GetComponent("puzzleElement")
```

#### Line 248: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  245: 		if !ok {
  246: 			continue
  247: 		}
> 248: 		puzzleElem := elemComp.(*PuzzleElementComponent)
  249: 
  250: 		// Skip if not interactable or on cooldown
  251: 		if !puzzleElem.IsInteractable || puzzleElem.CooldownElapsed > 0 {
```

#### Line 287: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  284: 			continue
  285: 		}
  286: 
> 287: 		puzzle := puzzleComp.(*PuzzleComponent)
  288: 		if puzzle.PuzzleID == puzzleElem.PuzzleID {
  289: 			// Record progress
  290: 			puzzle.RecordProgress(puzzleElem.ElementName)
```

---

### File: `pkg/engine/inventory_system.go` (4 issues)

#### Line 256: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  253: 	comp, hasHealth := entity.GetComponent("health")
  254: 	var healthComp *HealthComponent
  255: 	if hasHealth {
> 256: 		healthComp, _ = comp.(*HealthComponent)
  257: 	}
  258: 
  259: 	// Apply effects based on consumable type
```

#### Line 301: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  298: 	if !ok {
  299: 		return
  300: 	}
> 301: 	equipComp, _ := comp.(*EquipmentComponent)
  302: 	if equipComp == nil {
  303: 		return
  304: 	}
```

#### Line 381: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  378: 		// Entity has no position - can't drop item in world
  379: 		return fmt.Errorf("entity %d has no position component, cannot drop item", entityID)
  380: 	}
> 381: 	pos := posComp.(*PositionComponent)
  382: 
  383: 	// Remove item from inventory (only after we know we can drop it)
  384: 	itm := invComp.RemoveItem(inventoryIndex)
```

#### Line 572: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  569: 			continue
  570: 		}
  571: 
> 572: 		equipment := equipComp.(*EquipmentComponent)
  573: 
  574: 		// Recalculate equipment stats if dirty
  575: 		// This ensures CachedStats is up-to-date for CharacterUI display
```

---

### File: `pkg/engine/inventory_ui.go` (3 issues)

#### Line 113: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  110: 	if !ok {
  111: 		return
  112: 	}
> 113: 	inventory := invComp.(*InventoryComponent)
  114: 
  115: 	// Calculate inventory window position
  116: 	windowWidth := ui.gridCols*ui.slotSize + ui.padding*2
```

#### Line 231: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  228: 	if !ok {
  229: 		return
  230: 	}
> 231: 	inventory := invComp.(*InventoryComponent)
  232: 
  233: 	// Draw semi-transparent overlay
  234: 	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
```

#### Line 349: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  346: 
  347: 		// Show equipped item if present
  348: 		if hasEquipment {
> 349: 			equipment := equipComp.(*EquipmentComponent)
  350: 			equipped := equipment.GetEquipped(slotInfo.slot)
  351: 			if equipped != nil {
  352: 				itemName := equipped.Name
```

---

### File: `pkg/engine/item_spawning.go` (10 issues)

#### Line 78: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  75: 
  76: 	// Increase drop chance for bosses/elites
  77: 	if statsComp, ok := enemy.GetComponent("stats"); ok {
> 78: 		stats := statsComp.(*StatsComponent)
  79: 		if stats.Attack > 20 || stats.Defense > 20 {
  80: 			dropChance = 0.7 // 70% for strong enemies
  81: 		}
```

#### Line 93: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  90: 	// Determine item depth from enemy stats
  91: 	depth := 1
  92: 	if expComp, ok := enemy.GetComponent("experience"); ok {
> 93: 		exp := expComp.(*ExperienceComponent)
  94: 		depth = exp.Level
  95: 	}
  96: 
```

#### Line 113: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  110: 		return nil
  111: 	}
  112: 
> 113: 	items := result.([]*item.Item)
  114: 	if len(items) == 0 {
  115: 		return nil
  116: 	}
```

#### Line 133: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  130: 
  131: 	// Increase drop chance for bosses/elites
  132: 	if statsComp, ok := enemy.GetComponent("stats"); ok {
> 133: 		stats := statsComp.(*StatsComponent)
  134: 		if stats.Attack > 20 || stats.Defense > 20 {
  135: 			dropChance = 0.2 // 20% for strong enemies
  136: 		}
```

#### Line 150: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  147: 	difficulty := 0.3 // Start lower for common recipes
  148: 
  149: 	if expComp, ok := enemy.GetComponent("experience"); ok {
> 150: 		exp := expComp.(*ExperienceComponent)
  151: 		depth = exp.Level
  152: 		difficulty = 0.3 + float64(depth)*0.05 // Scale with depth
  153: 	}
```

#### Line 170: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  167: 		return nil
  168: 	}
  169: 
> 170: 	recipes := result.([]*Recipe)
  171: 	if len(recipes) == 0 {
  172: 		return nil
  173: 	}
```

#### Line 321: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  318: 			continue
  319: 		}
  320: 
> 321: 		inventory := playerInventory.(*InventoryComponent)
  322: 
  323: 		for _, itemEntity := range items {
  324: 			_, hasItemPos := itemEntity.GetComponent("position")
```

#### Line 334: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  331: 				continue
  332: 			}
  333: 
> 334: 			itemData := itemEntityComp.(*ItemEntityComponent)
  335: 
  336: 			// Check distance for pickup (32 pixels = 1 tile)
  337: 			distance := GetDistance(player, itemEntity)
```

#### Line 390: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  387: 				continue
  388: 			}
  389: 
> 390: 			recipeData := recipeEntityComp.(*RecipeEntityComponent)
  391: 
  392: 			// Check distance for pickup (32 pixels = 1 tile)
  393: 			distance := GetDistance(player, recipeEntity)
```

#### Line 403: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  400: 					player.AddComponent(knowledgeComp)
  401: 				}
  402: 
> 403: 				knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 
  405: 				// Check if player already knows this recipe
  406: 				if knowledge.KnowsRecipe(recipeData.Recipe.ID) {
```

---

### File: `pkg/engine/lifetime_system.go` (1 issues)

#### Line 50: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  47: 			continue
  48: 		}
  49: 
> 50: 		lifetime := lifetimeComp.(*LifetimeComponent)
  51: 		lifetime.Elapsed += deltaTime
  52: 
  53: 		// Check if lifetime expired
```

---

### File: `pkg/engine/map_ui.go` (6 issues)

#### Line 299: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  296: 
  297: 	// Draw player icon
  298: 	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 299: 		pos := posComp.(*PositionComponent)
  300: 		// Convert world position to tile coordinates (assuming 32px tiles)
  301: 		tileX := int(pos.X / 32)
  302: 		tileY := int(pos.Y / 32)
```

#### Line 427: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  424: 		return
  425: 	}
  426: 
> 427: 	pos := posComp.(*PositionComponent)
  428: 	// Convert world position to tile coordinates (assuming 32px tiles)
  429: 	centerX := int(pos.X / 32)
  430: 	centerY := int(pos.Y / 32)
```

#### Line 512: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  509: 
  510: 	// Draw player icon
  511: 	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 512: 		pos := posComp.(*PositionComponent)
  513: 		tileX := int(pos.X / 32)
  514: 		tileY := int(pos.Y / 32)
  515: 
```

#### Line 537: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  534: 			continue
  535: 		}
  536: 
> 537: 		pos := posComp.(*PositionComponent)
  538: 		tileX := int(pos.X / 32)
  539: 		tileY := int(pos.Y / 32)
  540: 
```

#### Line 557: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  554: 
  555: 		// Check if enemy
  556: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 557: 			team := teamComp.(*TeamComponent)
  558: 			if team.TeamID == 2 { // Enemy team
  559: 				iconColor = color.RGBA{255, 100, 100, 255} // Red
  560: 			}
```

#### Line 636: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  633: 		return
  634: 	}
  635: 
> 636: 	pos := posComp.(*PositionComponent)
  637: 	tileX := int(pos.X / 32)
  638: 	tileY := int(pos.Y / 32)
  639: 
```

---

### File: `pkg/engine/menu_system.go` (6 issues)

#### Line 107: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  104: 		ms.world.Update(0) // Process entity addition
  105: 	} else {
  106: 		if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
> 107: 			menuComp := menu.(*MenuComponent)
  108: 			menuComp.Active = !menuComp.Active
  109: 
  110: 			// Rebuild main menu when opening
```

#### Line 126: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  123: 		return false
  124: 	}
  125: 	if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
> 126: 		return menu.(*MenuComponent).Active
  127: 	}
  128: 	return false
  129: }
```

#### Line 138: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  135: 	}
  136: 
  137: 	menu, ok := ms.menuEntity.GetComponent("menu")
> 138: 	if !ok || !menu.(*MenuComponent).Active {
  139: 		return
  140: 	}
  141: 
```

#### Line 142: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  139: 		return
  140: 	}
  141: 
> 142: 	menuComp := menu.(*MenuComponent)
  143: 
  144: 	// Update error message timeout
  145: 	if menuComp.ErrorTimeout > 0 {
```

#### Line 481: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  478: 	}
  479: 
  480: 	menu, ok := ms.menuEntity.GetComponent("menu")
> 481: 	if !ok || !menu.(*MenuComponent).Active {
  482: 		return
  483: 	}
  484: 
```

#### Line 485: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  482: 		return
  483: 	}
  484: 
> 485: 	menuComp := menu.(*MenuComponent)
  486: 
  487: 	// Draw semi-transparent overlay
  488: 	overlay := ebiten.NewImage(ms.screenWidth, ms.screenHeight)
```

---

### File: `pkg/engine/merchant_spawn.go` (3 issues)

#### Line 236: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  233: 		if !ok {
  234: 			continue
  235: 		}
> 236: 		pos := posComp.(*PositionComponent)
  237: 
  238: 		// Calculate distance squared (avoid sqrt for performance)
  239: 		dx := pos.X - x
```

#### Line 267: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  264: 		if !ok {
  265: 			continue
  266: 		}
> 267: 		pos := posComp.(*PositionComponent)
  268: 
  269: 		dx := pos.X - x
  270: 		dy := pos.Y - y
```

#### Line 308: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  305: 		return ""
  306: 	}
  307: 
> 308: 	merchantData := merchComp.(*MerchantComponent)
  309: 	return fmt.Sprintf("Press S to talk to %s", merchantData.MerchantName)
  310: }
  311: 
```

---

### File: `pkg/engine/mounting_system.go` (11 issues)

#### Line 45: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  42: 			continue
  43: 		}
  44: 
> 45: 		mount := mountComp.(*MountComponent)
  46: 
  47: 		// Find the vehicle entity
  48: 		vehicle := ms.findEntity(entities, mount.MountedEntityID)
```

#### Line 60: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  57: 		if !hasVehiclePos {
  58: 			continue
  59: 		}
> 60: 		vehiclePos := vehiclePosComp.(*PositionComponent)
  61: 
  62: 		// Update rider position to match vehicle + offset
  63: 		riderPosComp, hasRiderPos := entity.GetComponent("position")
```

#### Line 65: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  62: 		// Update rider position to match vehicle + offset
  63: 		riderPosComp, hasRiderPos := entity.GetComponent("position")
  64: 		if hasRiderPos {
> 65: 			riderPos := riderPosComp.(*PositionComponent)
  66: 			riderPos.X = vehiclePos.X + mount.MountOffset.X
  67: 			riderPos.Y = vehiclePos.Y + mount.MountOffset.Y
  68: 		}
```

#### Line 74: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  71: 		vehicleRotComp, hasVehicleRot := vehicle.GetComponent("rotation")
  72: 		riderRotComp, hasRiderRot := entity.GetComponent("rotation")
  73: 		if hasVehicleRot && hasRiderRot {
> 74: 			vehicleRot := vehicleRotComp.(*RotationComponent)
  75: 			riderRot := riderRotComp.(*RotationComponent)
  76: 			riderRot.Angle = vehicleRot.Angle
  77: 		}
```

#### Line 75: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  72: 		riderRotComp, hasRiderRot := entity.GetComponent("rotation")
  73: 		if hasVehicleRot && hasRiderRot {
  74: 			vehicleRot := vehicleRotComp.(*RotationComponent)
> 75: 			riderRot := riderRotComp.(*RotationComponent)
  76: 			riderRot.Angle = vehicleRot.Angle
  77: 		}
  78: 	}
```

#### Line 102: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  99: 		return fmt.Errorf("entity is not a vehicle")
  100: 	}
  101: 
> 102: 	vehicleData := vehicleComp.(*VehicleComponent)
  103: 
  104: 	// Check if vehicle has capacity
  105: 	if !vehicleData.CanAddPassenger() {
```

#### Line 121: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  118: 		return fmt.Errorf("missing position component")
  119: 	}
  120: 
> 121: 	riderPos := riderPosComp.(*PositionComponent)
  122: 	vehiclePos := vehiclePosComp.(*PositionComponent)
  123: 
  124: 	// Calculate offset (preserve relative position)
```

#### Line 122: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  119: 	}
  120: 
  121: 	riderPos := riderPosComp.(*PositionComponent)
> 122: 	vehiclePos := vehiclePosComp.(*PositionComponent)
  123: 
  124: 	// Calculate offset (preserve relative position)
  125: 	offsetX := riderPos.X - vehiclePos.X
```

#### Line 158: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  155: 		return fmt.Errorf("rider is not mounted")
  156: 	}
  157: 
> 158: 	mount := mountComp.(*MountComponent)
  159: 
  160: 	// Find vehicle and update passenger count
  161: 	if ms.world != nil {
```

#### Line 165: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  162: 		vehicle, exists := ms.world.GetEntity(mount.MountedEntityID)
  163: 		if exists && vehicle != nil {
  164: 			if vehicleComp, hasVehicle := vehicle.GetComponent("vehicle"); hasVehicle {
> 165: 				vehicleData := vehicleComp.(*VehicleComponent)
  166: 				vehicleData.RemovePassenger()
  167: 			}
  168: 		}
```

#### Line 207: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  204: 		return nil
  205: 	}
  206: 
> 207: 	mount := mountComp.(*MountComponent)
  208: 
  209: 	if ms.world != nil {
  210: 		vehicle, exists := ms.world.GetEntity(mount.MountedEntityID)
```

---

### File: `pkg/engine/movement.go` (10 issues)

#### Line 66: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  63: 			continue
  64: 		}
  65: 
> 66: 		pos := posComp.(*PositionComponent)
  67: 		vel := velComp.(*VelocityComponent)
  68: 
  69: 		// Apply speed limit if configured
```

#### Line 67: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  64: 		}
  65: 
  66: 		pos := posComp.(*PositionComponent)
> 67: 		vel := velComp.(*VelocityComponent)
  68: 
  69: 		// Apply speed limit if configured
  70: 		if s.MaxSpeed > 0 {
```

#### Line 87: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  84: 		// If collision system is set, validate position before moving
  85: 		if s.collisionSystem != nil && entity.HasComponent("collider") {
  86: 			colliderComp, _ := entity.GetComponent("collider")
> 87: 			collider := colliderComp.(*ColliderComponent)
  88: 
  89: 			// Only check solid, non-trigger colliders
  90: 			if collider.Solid && !collider.IsTrigger {
```

#### Line 164: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  161: 
  162: 		// Apply bounds if entity has them
  163: 		if boundsComp, hasBounds := entity.GetComponent("bounds"); hasBounds {
> 164: 			bounds := boundsComp.(*BoundsComponent)
  165: 			pos.X, pos.Y = bounds.Clamp(pos.X, pos.Y)
  166: 
  167: 			// Stop movement at boundaries if not wrapping
```

#### Line 180: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  177: 
  178: 		// Priority 1.4: Apply friction/drag to slow down entities
  179: 		if frictionComp, hasFriction := entity.GetComponent("friction"); hasFriction {
> 180: 			friction := frictionComp.(*FrictionComponent)
  181: 
  182: 			// Apply friction as exponential decay: v *= (1 - coefficient)^deltaTime
  183: 			// For small deltaTime and coefficient, this approximates: v *= (1 - coefficient * deltaTime)
```

#### Line 197: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  194: 
  195: 		// Update animation state based on movement
  196: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 197: 			anim := animComp.(*AnimationComponent)
  198: 			speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  199: 
  200: 			// DON'T override attack/hit/death/cast animations - let them finish
```

#### Line 290: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  287: } // SetVelocity is a helper to set entity velocity.
  288: func SetVelocity(entity *Entity, vx, vy float64) {
  289: 	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 290: 		vel := velComp.(*VelocityComponent)
  291: 		vel.VX = vx
  292: 		vel.VY = vy
  293: 	}
```

#### Line 299: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  296: // GetPosition is a helper to get entity position.
  297: func GetPosition(entity *Entity) (x, y float64, ok bool) {
  298: 	if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 299: 		pos := posComp.(*PositionComponent)
  300: 		return pos.X, pos.Y, true
  301: 	}
  302: 	return 0, 0, false
```

#### Line 308: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  305: // SetPosition is a helper to set entity position.
  306: func SetPosition(entity *Entity, x, y float64) {
  307: 	if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 308: 		pos := posComp.(*PositionComponent)
  309: 		pos.X = x
  310: 		pos.Y = y
  311: 	}
```

#### Line 382: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  379: 	if !hasCollider {
  380: 		return
  381: 	}
> 382: 	collider := colliderComp.(*ColliderComponent)
  383: 
  384: 	// Calculate tile coordinates from entity position using helper method
  385: 	tileX, tileY := terrainChecker.worldToTileCoords(pos.X, pos.Y)
```

---

### File: `pkg/engine/music_context.go` (7 issues)

#### Line 109: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  106: 	if !hasPos {
  107: 		return MusicContextExploration
  108: 	}
> 109: 	playerPos := playerPosComp.(*PositionComponent)
  110: 
  111: 	// Check player health for danger state
  112: 	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
```

#### Line 113: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  110: 
  111: 	// Check player health for danger state
  112: 	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
> 113: 		health := healthComp.(*HealthComponent)
  114: 		healthPercent := float64(health.Current) / float64(health.Max)
  115: 		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
  116: 			// Danger state active, but continue checking for combat/boss
```

#### Line 145: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  142: 		// Check team (enemies are on different team than player)
  143: 		playerTeam := 1 // Default player team
  144: 		if teamComp, hasTeam := playerEntity.GetComponent("team"); hasTeam {
> 145: 			playerTeam = teamComp.(*TeamComponent).TeamID
  146: 		}
  147: 
  148: 		entityTeam := 0 // Default enemy team
```

#### Line 150: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  147: 
  148: 		entityTeam := 0 // Default enemy team
  149: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 150: 			entityTeam = teamComp.(*TeamComponent).TeamID
  151: 		}
  152: 
  153: 		if entityTeam == playerTeam {
```

#### Line 162: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  159: 
  160: 		// Check proximity to player
  161: 		if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 162: 			entityPos := posComp.(*PositionComponent)
  163: 			distance := d.calculateDistance(playerPos.X, playerPos.Y, entityPos.X, entityPos.Y)
  164: 
  165: 			if distance <= d.CombatRadius {
```

#### Line 170: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  167: 
  168: 				// Check if it's a boss (high attack stat)
  169: 				if statsComp, hasStats := entity.GetComponent("stats"); hasStats {
> 170: 					stats := statsComp.(*StatsComponent)
  171: 					if stats.Attack >= d.BossAttackThreshold {
  172: 						hasBoss = true
  173: 					}
```

#### Line 192: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  189: 
  190: 	// Check danger state last (lowest priority of active contexts)
  191: 	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
> 192: 		health := healthComp.(*HealthComponent)
  193: 		healthPercent := float64(health.Current) / float64(health.Max)
  194: 		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
  195: 			return MusicContextDanger
```

---

### File: `pkg/engine/narrative_system.go` (2 issues)

#### Line 49: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  46: 			continue
  47: 		}
  48: 
> 49: 		narrative := narComp.(*NarrativeComponent)
  50: 
  51: 		// Check for triggered events
  52: 		triggeredEvents := narrative.CheckTriggerConditions()
```

#### Line 128: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  125: 	if enemyEntity != nil {
  126: 		// Check for boss or elite status
  127: 		if aiComp, ok := enemyEntity.GetComponent("ai"); ok {
> 128: 			ai := aiComp.(*AIComponent)
  129: 			// Boss entities typically have high detection range
  130: 			if ai.DetectionRange > BossDetectionRange {
  131: 				importance = 0.8 // Boss fight
```

---

### File: `pkg/engine/objective_tracker_system.go` (7 issues)

#### Line 71: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  68: 	if !ok {
  69: 		return
  70: 	}
> 71: 	tracker := comp.(*QuestTrackerComponent)
  72: 
  73: 	// For now, all enemies count as "enemy" or "monster"
  74: 	// In future, could extract type from entity components
```

#### Line 102: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  99: 	if !ok {
  100: 		return
  101: 	}
> 102: 	tracker := comp.(*QuestTrackerComponent)
  103: 
  104: 	// Update collect objectives
  105: 	for _, tracked := range tracker.ActiveQuests {
```

#### Line 135: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  132: 	if !ok {
  133: 		return
  134: 	}
> 135: 	tracker := comp.(*QuestTrackerComponent)
  136: 
  137: 	// Update UI interaction objectives (used in tutorial quests)
  138: 	for _, tracked := range tracker.ActiveQuests {
```

#### Line 169: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  166: 	if !ok {
  167: 		return
  168: 	}
> 169: 	tracker := comp.(*QuestTrackerComponent)
  170: 
  171: 	// Update explore objectives
  172: 	for _, tracked := range tracker.ActiveQuests {
```

#### Line 194: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  191: 	if !ok {
  192: 		return
  193: 	}
> 194: 	pos := posComp.(*PositionComponent)
  195: 
  196: 	// Convert world coordinates to tile coordinates (assuming 32-pixel tiles)
  197: 	tileX := int(pos.X / 32)
```

#### Line 209: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  206: 	if !ok {
  207: 		return
  208: 	}
> 209: 	tracker := comp.(*QuestTrackerComponent)
  210: 
  211: 	// Check each active quest
  212: 	for _, tracked := range tracker.ActiveQuests {
```

#### Line 328: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  325: 	if !ok {
  326: 		return // No inventory to receive items
  327: 	}
> 328: 	inv := invComp.(*InventoryComponent)
  329: 
  330: 	// Get genre for item generation
  331: 	genreID := "fantasy" // Default genre
```

---

### File: `pkg/engine/particle_system.go` (2 issues)

#### Line 35: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  32: 			continue
  33: 		}
  34: 
> 35: 		emitter := comp.(*ParticleEmitterComponent)
  36: 
  37: 		// Update elapsed time for time-limited emitters
  38: 		if emitter.EmissionTime > 0 {
```

#### Line 71: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  68: 
  69: 				// Position particles at entity's position
  70: 				if posComp, ok := entity.GetComponent("position"); ok {
> 71: 					pos := posComp.(*PositionComponent)
  72: 					ps.offsetParticles(system, pos.X, pos.Y)
  73: 				}
  74: 
```

---

### File: `pkg/engine/player_combat_system.go` (4 issues)

#### Line 73: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  70: 		if !ok {
  71: 			continue // Entity can't attack
  72: 		}
> 73: 		attack := attackComp.(*AttackComponent)
  74: 
  75: 		// Check if attack is ready (cooldown)
  76: 		if !attack.CanAttack() {
```

#### Line 99: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  96: 		// ALWAYS trigger attack animation, even if no target
  97: 		// This provides visual feedback that the attack button was pressed
  98: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 99: 			anim := animComp.(*AnimationComponent)
  100: 			anim.SetState(AnimationStateAttack)
  101: 
  102: 			// Set OnComplete callback to return to idle/walk
```

#### Line 105: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  102: 			// Set OnComplete callback to return to idle/walk
  103: 			anim.OnComplete = func() {
  104: 				if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 105: 					vel := velComp.(*VelocityComponent)
  106: 					speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  107: 					if speed > 0.1 {
  108: 						anim.SetState(AnimationStateWalk)
```

#### Line 123: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  120: 
  121: 		// Check if entity has aim component (Phase 10.1)
  122: 		if aimComp, hasAim := entity.GetComponent("aim"); hasAim {
> 123: 			aim := aimComp.(*AimComponent)
  124: 			// Use aim direction for target selection with default aim cone (forgiving aim)
  125: 			target = FindEnemyInAimDirection(s.world, entity, aim.AimAngle, maxRange, DefaultAimCone)
  126: 
```

---

### File: `pkg/engine/player_item_use_system.go` (3 issues)

#### Line 72: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  69: 		if !ok {
  70: 			continue // Entity has no inventory
  71: 		}
> 72: 		inventory := invComp.(*InventoryComponent)
  73: 
  74: 		// Get hotbar component for selected item (if available)
  75: 		var selectedIndex int
```

#### Line 77: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  74: 		// Get hotbar component for selected item (if available)
  75: 		var selectedIndex int
  76: 		if hotbarComp, hasHotbar := entity.GetComponent("hotbar"); hasHotbar {
> 77: 			hotbar := hotbarComp.(*HotbarComponent)
  78: 			selectedIndex = hotbar.LastUsedIndex
  79: 			// Check if the slot has an item
  80: 			if selectedIndex == -1 || hotbar.GetSlot(selectedIndex) == nil {
```

#### Line 169: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  166: 	if !hasHotbar {
  167: 		return
  168: 	}
> 169: 	hotbar := hotbarComp.(*HotbarComponent)
  170: 	hotbar.LastUsedIndex = slotIndex
  171: }
  172: 
```

---

### File: `pkg/engine/player_spell_casting.go` (1 issues)

#### Line 57: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  54: 
  55: 	// Get spell slots
  56: 	slotsComp, _ := player.GetComponent("spell_slots")
> 57: 	slots := slotsComp.(*SpellSlotComponent)
  58: 
  59: 	// If currently casting, don't start new cast
  60: 	if slots.IsCasting() {
```

---

### File: `pkg/engine/progression_system.go` (10 issues)

#### Line 110: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  107: 		return fmt.Errorf("entity does not have experience component")
  108: 	}
  109: 
> 110: 	exp := expComp.(*ExperienceComponent)
  111: 
  112: 	if ps.logger != nil && ps.logger.Logger.GetLevel() >= logrus.DebugLevel {
  113: 		ps.logger.WithFields(logrus.Fields{
```

#### Line 172: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  169: 	if !ok {
  170: 		return // No scaling defined
  171: 	}
> 172: 	scaling := scalingComp.(*LevelScalingComponent)
  173: 
  174: 	// Update health component
  175: 	healthComp, ok := entity.GetComponent("health")
```

#### Line 177: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  174: 	// Update health component
  175: 	healthComp, ok := entity.GetComponent("health")
  176: 	if ok {
> 177: 		health := healthComp.(*HealthComponent)
  178: 		oldMax := health.Max
  179: 		health.Max = scaling.CalculateHealthForLevel(level)
  180: 		// Increase current health by the same amount
```

#### Line 187: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  184: 	// Update stats component
  185: 	statsComp, ok := entity.GetComponent("stats")
  186: 	if ok {
> 187: 		stats := statsComp.(*StatsComponent)
  188: 		stats.Attack = scaling.CalculateAttackForLevel(level)
  189: 		stats.Defense = scaling.CalculateDefenseForLevel(level)
  190: 		stats.MagicPower = scaling.CalculateMagicPowerForLevel(level)
```

#### Line 205: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  202: 		return 10
  203: 	}
  204: 
> 205: 	exp := expComp.(*ExperienceComponent)
  206: 	level := exp.Level
  207: 
  208: 	// Base XP = 10 * level
```

#### Line 230: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  227: 		return 1
  228: 	}
  229: 
> 230: 	return expComp.(*ExperienceComponent).Level
  231: }
  232: 
  233: // GetXPProgress returns the XP progress as a value between 0.0 and 1.0.
```

#### Line 244: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  241: 		return 0.0
  242: 	}
  243: 
> 244: 	return expComp.(*ExperienceComponent).ProgressToNextLevel()
  245: }
  246: 
  247: // SpendSkillPoint spends a skill point for an entity.
```

#### Line 259: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  256: 		return fmt.Errorf("entity does not have experience component")
  257: 	}
  258: 
> 259: 	exp := expComp.(*ExperienceComponent)
  260: 	if exp.SkillPoints <= 0 {
  261: 		return fmt.Errorf("entity has no skill points to spend")
  262: 	}
```

#### Line 279: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  276: 		return 0
  277: 	}
  278: 
> 279: 	return expComp.(*ExperienceComponent).SkillPoints
  280: }
  281: 
  282: // InitializeEntityAtLevel sets up an entity at a specific level.
```

#### Line 299: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  296: 		exp = NewExperienceComponent()
  297: 		entity.AddComponent(exp)
  298: 	} else {
> 299: 		exp = expComp.(*ExperienceComponent)
  300: 	}
  301: 
  302: 	// Set level and XP
```

---

### File: `pkg/engine/projectile_pool.go` (3 issues)

#### Line 43: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  40: // Get acquires a projectile component from the pool.
  41: // Returns a zeroed component ready for initialization.
  42: func (p *ProjectilePool) Get() *ProjectileComponent {
> 43: 	proj := p.pool.Get().(*ProjectileComponent)
  44: 	// Reset all fields to zero values
  45: 	proj.Damage = 0.0
  46: 	proj.Speed = 0.0
```

#### Line 87: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  84: 
  85: // Get acquires a velocity component from the pool.
  86: func (p *VelocityPool) Get() *VelocityComponent {
> 87: 	vel := p.pool.Get().(*VelocityComponent)
  88: 	vel.VX = 0.0
  89: 	vel.VY = 0.0
  90: 	return vel
```

#### Line 120: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  117: 
  118: // Get acquires a position component from the pool.
  119: func (p *PositionPool) Get() *PositionComponent {
> 120: 	pos := p.pool.Get().(*PositionComponent)
  121: 	pos.X = 0.0
  122: 	pos.Y = 0.0
  123: 	return pos
```

---

### File: `pkg/engine/projectile_system.go` (1 issues)

#### Line 472: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  469: 	switch projComp.ProjectileType {
  470: 	case "fireball", "magic_missile", "ice_shard", "lightning_bolt":
  471: 		// Magical projectiles get glowing trails
> 472: 		magicColor := spriteComp.Color.(color.RGBA)
  473: 		trail = NewMagicTrailComponent(&magicColor)
  474: 	case "arrow", "bullet", "bolt":
  475: 		// Physical projectiles get subtle trails
```

---

### File: `pkg/engine/projectile_system_test_debug.go` (5 issues)

#### Line 37: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  34: 	for _, e := range entities {
  35: 		if _, ok := e.GetComponent("projectile"); ok {
  36: 			if pos, ok := e.GetComponent("position"); ok {
> 37: 				p := pos.(*PositionComponent)
  38: 				fmt.Printf("Projectile start: (%.2f, %.2f)\n", p.X, p.Y)
  39: 			}
  40: 		}
```

#### Line 54: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  51: 		if _, ok := e.GetComponent("projectile"); ok {
  52: 			projectileExists = true
  53: 			if pos, ok := e.GetComponent("position"); ok {
> 54: 				p := pos.(*PositionComponent)
  55: 				fmt.Printf("Projectile after update: (%.2f, %.2f)\n", p.X, p.Y)
  56: 			}
  57: 		}
```

#### Line 65: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  62: 
  63: 	// Check health
  64: 	healthComp1, _ := target1.GetComponent("health")
> 65: 	health1 := healthComp1.(*HealthComponent)
  66: 	fmt.Printf("Target 1 health: %.2f\n", health1.Current)
  67: 
  68: 	healthComp2, _ := target2.GetComponent("health")
```

#### Line 69: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  66: 	fmt.Printf("Target 1 health: %.2f\n", health1.Current)
  67: 
  68: 	healthComp2, _ := target2.GetComponent("health")
> 69: 	health2 := healthComp2.(*HealthComponent)
  70: 	fmt.Printf("Target 2 health: %.2f\n", health2.Current)
  71: 
  72: 	healthComp3, _ := target3.GetComponent("health")
```

#### Line 73: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  70: 	fmt.Printf("Target 2 health: %.2f\n", health2.Current)
  71: 
  72: 	healthComp3, _ := target3.GetComponent("health")
> 73: 	health3 := healthComp3.(*HealthComponent)
  74: 	fmt.Printf("Target 3 health: %.2f\n", health3.Current)
  75: }
  76: 
```

---

### File: `pkg/engine/projectile_system_test_minimal.go` (3 issues)

#### Line 31: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  28: 	fmt.Printf("Entities with health and position: %d\n", len(entities))
  29: 	for _, e := range entities {
  30: 		if pos, ok := e.GetComponent("position"); ok {
> 31: 			p := pos.(*PositionComponent)
  32: 			fmt.Printf("  Entity ID=%d at (%.1f, %.1f)\n", e.ID, p.X, p.Y)
  33: 		}
  34: 	}
```

#### Line 43: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  40: 
  41: 	// Check projectile position
  42: 	if projPos, ok := proj.GetComponent("position"); ok {
> 43: 		p := projPos.(*PositionComponent)
  44: 		fmt.Printf("Projectile now at (%.1f, %.1f)\n", p.X, p.Y)
  45: 	} else {
  46: 		fmt.Println("Projectile despawned (collision occurred?)")
```

#### Line 51: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  48: 
  49: 	// Check target health
  50: 	if healthComp, ok := target.GetComponent("health"); ok {
> 51: 		health := healthComp.(*HealthComponent)
  52: 		fmt.Printf("Target health: %.1f\n", health.Current)
  53: 		if health.Current == 100.0 {
  54: 			t.Error("Target should have taken damage")
```

---

### File: `pkg/engine/quest_ui.go` (2 issues)

#### Line 102: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  99: 	if ui.playerEntity != nil {
  100: 		trackerComp, ok := ui.playerEntity.GetComponent("questtracker")
  101: 		if ok {
> 102: 			tracker := trackerComp.(*QuestTrackerComponent)
  103: 			var quests []*TrackedQuest
  104: 			if ui.currentTab == 0 {
  105: 				quests = tracker.ActiveQuests
```

#### Line 186: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  183: 	if !ok {
  184: 		return
  185: 	}
> 186: 	tracker := trackerComp.(*QuestTrackerComponent)
  187: 
  188: 	// Draw semi-transparent overlay
  189: 	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
```

---

### File: `pkg/engine/render_system.go` (22 issues)

#### Line 336: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  333: 		if !hasSprite {
  334: 			continue
  335: 		}
> 336: 		sprite := spriteComp.(*EbitenSprite)
  337: 
  338: 		if !sprite.Visible {
  339: 			continue
```

#### Line 376: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  373: 	if !hasSprite {
  374: 		return
  375: 	}
> 376: 	batchSpriteImage := firstSprite.(*EbitenSprite).Image
  377: 	if batchSpriteImage == nil {
  378: 		// No sprite image, draw entities individually
  379: 		for _, entity := range entities {
```

#### Line 409: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  406: 			continue
  407: 		}
  408: 
> 409: 		pos := posComp.(*PositionComponent)
  410: 		sprite := spriteComp.(*EbitenSprite)
  411: 
  412: 		// DEBUG: Log sprite state for player
```

#### Line 410: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  407: 		}
  408: 
  409: 		pos := posComp.(*PositionComponent)
> 410: 		sprite := spriteComp.(*EbitenSprite)
  411: 
  412: 		// DEBUG: Log sprite state for player
  413: 		if entity.HasComponent("input") && sprite.Image == nil {
```

#### Line 423: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  420: 
  421: 		// Phase 4: Sync CurrentDirection from AnimationComponent.Facing
  422: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 423: 			anim := animComp.(*AnimationComponent)
  424: 			sprite.CurrentDirection = int(anim.GetFacing())
  425: 		}
  426: 
```

#### Line 431: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  428: 		// This enables 360° visual rotation for entities with rotation component
  429: 		// CRITICAL: Must sync here for batch rendering path (drawEntity has its own sync)
  430: 		if rotComp, hasRot := entity.GetComponent("rotation"); hasRot {
> 431: 			rotation := rotComp.(*RotationComponent)
  432: 			sprite.Rotation = rotation.Angle
  433: 		}
  434: 
```

#### Line 474: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  471: 		var flashAlpha float64
  472: 		var tintR, tintG, tintB, tintA float64 = 1.0, 1.0, 1.0, 1.0
  473: 		if feedbackComp, ok := entity.GetComponent("visual_feedback"); ok {
> 474: 			feedback := feedbackComp.(*VisualFeedbackComponent)
  475: 			flashAlpha = feedback.GetFlashAlpha()
  476: 			tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
  477: 		}
```

#### Line 611: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  608: 	if !ok {
  609: 		return entities
  610: 	}
> 611: 	camera := camComp.(*CameraComponent)
  612: 
  613: 	// Calculate viewport bounds in world space with margin for sprites
  614: 	margin := 100.0 // Extra space to render sprites partially off-screen
```

#### Line 657: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  654: 		return
  655: 	}
  656: 
> 657: 	pos := posComp.(*PositionComponent)
  658: 	sprite := spriteComp.(*EbitenSprite)
  659: 
  660: 	if !sprite.Visible {
```

#### Line 658: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  655: 	}
  656: 
  657: 	pos := posComp.(*PositionComponent)
> 658: 	sprite := spriteComp.(*EbitenSprite)
  659: 
  660: 	if !sprite.Visible {
  661: 		if entity.HasComponent("input") {
```

#### Line 669: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  666: 
  667: 	// Phase 4: Sync CurrentDirection from AnimationComponent.Facing
  668: 	if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 669: 		anim := animComp.(*AnimationComponent)
  670: 		sprite.CurrentDirection = int(anim.GetFacing())
  671: 	}
  672: 
```

#### Line 676: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  673: 	// Phase 10.1: Sync sprite rotation from RotationComponent if present
  674: 	// This enables 360° visual rotation for entities with rotation component
  675: 	if rotComp, hasRot := entity.GetComponent("rotation"); hasRot {
> 676: 		rotation := rotComp.(*RotationComponent)
  677: 		sprite.Rotation = rotation.Angle
  678: 
  679: 		// DEBUG: Log rotation values for player entity (entity ID 1)
```

#### Line 700: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  697: 	var layerTransitionYOffset float64
  698: 	var layerTransitionAlpha float64 = 1.0
  699: 	if layerComp, hasLayer := entity.GetComponent("layer"); hasLayer {
> 700: 		layer := layerComp.(*LayerComponent)
  701: 		if layer.IsTransitioning() {
  702: 			// Calculate depth offset based on transition progress
  703: 			// Moving up to higher layer (platform): negative offset (entity rises)
```

#### Line 732: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  729: 	var flashAlpha float64
  730: 	var tintR, tintG, tintB, tintA float64 = 1.0, 1.0, 1.0, 1.0
  731: 	if feedbackComp, ok := entity.GetComponent("visual_feedback"); ok {
> 732: 		feedback := feedbackComp.(*VisualFeedbackComponent)
  733: 		flashAlpha = feedback.GetFlashAlpha()
  734: 		tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
  735: 	}
```

#### Line 832: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  829: 		return
  830: 	}
  831: 
> 832: 	health := healthComp.(*HealthComponent)
  833: 
  834: 	// Don't draw health bar for player (has HUD display)
  835: 	if entity.HasComponent("input") {
```

#### Line 842: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  839: 	// Check if entity is a boss (high attack indicates boss)
  840: 	isBoss := false
  841: 	if attackComp, ok := entity.GetComponent("attack"); ok {
> 842: 		attack := attackComp.(*AttackComponent)
  843: 		isBoss = attack.Damage > 20 // Boss threshold
  844: 	}
  845: 
```

#### Line 908: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  905: 			continue
  906: 		}
  907: 
> 908: 		emitter := comp.(*ParticleEmitterComponent)
  909: 
  910: 		// Render each particle system
  911: 		for _, system := range emitter.Systems {
```

#### Line 989: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  986: 			continue
  987: 		}
  988: 
> 989: 		pos := posComp.(*PositionComponent)
  990: 		collider := colliderComp.(*ColliderComponent)
  991: 
  992: 		// Get collider bounds
```

#### Line 990: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  987: 		}
  988: 
  989: 		pos := posComp.(*PositionComponent)
> 990: 		collider := colliderComp.(*ColliderComponent)
  991: 
  992: 		// Get collider bounds
  993: 		minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)
```

#### Line 1025: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1022: 	// Collect entities with sprites and cache their sprite components
  1023: 	for _, entity := range entities {
  1024: 		if sprite, ok := entity.GetComponent("sprite"); ok {
> 1025: 			ebitenSprite := sprite.(*EbitenSprite)
  1026: 			cache = append(cache, entitySprite{
  1027: 				entity: entity,
  1028: 				sprite: ebitenSprite,
```

#### Line 1047: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1044: 		posI, okI := cache[i].entity.GetComponent("position")
  1045: 		posJ, okJ := cache[j].entity.GetComponent("position")
  1046: 		if okI && okJ {
> 1047: 			yI := posI.(*PositionComponent).Y
  1048: 			yJ := posJ.(*PositionComponent).Y
  1049: 			if yI != yJ {
  1050: 				return yI < yJ
```

#### Line 1048: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1045: 		posJ, okJ := cache[j].entity.GetComponent("position")
  1046: 		if okI && okJ {
  1047: 			yI := posI.(*PositionComponent).Y
> 1048: 			yJ := posJ.(*PositionComponent).Y
  1049: 			if yI != yJ {
  1050: 				return yI < yJ
  1051: 			}
```

---

### File: `pkg/engine/revival_system.go` (7 issues)

#### Line 48: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  45: 	for _, entity := range entities {
  46: 		if entity.HasComponent("input") && !entity.HasComponent("dead") {
  47: 			if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
> 48: 				health := healthComp.(*HealthComponent)
  49: 				if health.IsAlive() {
  50: 					livingPlayers = append(livingPlayers, entity)
  51: 				}
```

#### Line 73: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  70: 	for _, livingPlayer := range livingPlayers {
  71: 		// Check for revival input (E key or interact button)
  72: 		inputComp, _ := livingPlayer.GetComponent("input")
> 73: 		input := inputComp.(*EbitenInput)
  74: 
  75: 		// Check if revival action key is pressed (E key = UseItemPressed)
  76: 		// In this context, E key serves dual purpose: use item / interact / revive
```

#### Line 86: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  83: 		if !hasLivingPos {
  84: 			continue
  85: 		}
> 86: 		livingPos := livingPosComp.(*PositionComponent)
  87: 
  88: 		// Find closest dead player within range
  89: 		var closestDeadPlayer *Entity
```

#### Line 98: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  95: 			if !hasDeadPos {
  96: 				continue
  97: 			}
> 98: 			deadPos := deadPosComp.(*PositionComponent)
  99: 
  100: 			// Calculate distance
  101: 			dx := deadPos.X - livingPos.X
```

#### Line 126: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  123: 	if !hasHealth {
  124: 		return
  125: 	}
> 126: 	health := healthComp.(*HealthComponent)
  127: 
  128: 	// Restore health (20% of max by default)
  129: 	restoredHealth := health.Max * s.RevivalAmount
```

#### Line 168: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  165: 	if !hasPos {
  166: 		return nil
  167: 	}
> 168: 	livingPos := livingPosComp.(*PositionComponent)
  169: 
  170: 	var revivablePlayers []*Entity
  171: 
```

#### Line 183: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  180: 		if !hasDeadPos {
  181: 			continue
  182: 		}
> 183: 		deadPos := deadPosComp.(*PositionComponent)
  184: 
  185: 		// Calculate distance
  186: 		dx := deadPos.X - livingPos.X
```

---

### File: `pkg/engine/rotation_system.go` (9 issues)

#### Line 42: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  39: 		if !ok {
  40: 			continue
  41: 		}
> 42: 		rotation := rotComp.(*RotationComponent)
  43: 
  44: 		// Sync rotation target with aim component if present
  45: 		if entity.HasComponent("aim") {
```

#### Line 48: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  45: 		if entity.HasComponent("aim") {
  46: 			aimComp, ok := entity.GetComponent("aim")
  47: 			if ok {
> 48: 				aim := aimComp.(*AimComponent)
  49: 
  50: 				// Update aim angle from position if target-based
  51: 				if entity.HasComponent("position") {
```

#### Line 54: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  51: 				if entity.HasComponent("position") {
  52: 					posComp, ok := entity.GetComponent("position")
  53: 					if ok {
> 54: 						pos := posComp.(*PositionComponent)
  55: 						aim.UpdateAimAngle(pos.X, pos.Y)
  56: 					}
  57: 				}
```

#### Line 96: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  93: 	}
  94: 
  95: 	rotComp, _ := entity.GetComponent("rotation")
> 96: 	rotation := rotComp.(*RotationComponent)
  97: 
  98: 	aimComp, _ := entity.GetComponent("aim")
  99: 	aim := aimComp.(*AimComponent)
```

#### Line 99: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  96: 	rotation := rotComp.(*RotationComponent)
  97: 
  98: 	aimComp, _ := entity.GetComponent("aim")
> 99: 	aim := aimComp.(*AimComponent)
  100: 
  101: 	rotation.SetAngleImmediate(aim.AimAngle)
  102: 	return true
```

#### Line 120: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  117: 	}
  118: 
  119: 	rotComp, _ := entity.GetComponent("rotation")
> 120: 	rotation := rotComp.(*RotationComponent)
  121: 
  122: 	rotation.SetAngleImmediate(angle)
  123: 	return true
```

#### Line 140: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  137: 	}
  138: 
  139: 	rotComp, _ := entity.GetComponent("rotation")
> 140: 	rotation := rotComp.(*RotationComponent)
  141: 
  142: 	return rotation.Angle, true
  143: }
```

#### Line 161: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  158: 	}
  159: 
  160: 	rotComp, _ := entity.GetComponent("rotation")
> 161: 	rotation := rotComp.(*RotationComponent)
  162: 
  163: 	rotation.SmoothRotation = enabled
  164: 	return true
```

#### Line 182: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  179: 	}
  180: 
  181: 	rotComp, _ := entity.GetComponent("rotation")
> 182: 	rotation := rotComp.(*RotationComponent)
  183: 
  184: 	rotation.RotationSpeed = speed
  185: 	return true
```

---

### File: `pkg/engine/shop_ui.go` (4 issues)

#### Line 203: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  200: 	if ui.mode == ShopModeBuy {
  201: 		// Show merchant inventory
  202: 		if merchantComp, ok := ui.merchantEntity.GetComponent("merchant"); ok {
> 203: 			merchant := merchantComp.(*MerchantComponent)
  204: 			currentInventory = merchant.Inventory
  205: 		}
  206: 	} else {
```

#### Line 209: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  206: 	} else {
  207: 		// Show player inventory
  208: 		if invComp, ok := ui.playerEntity.GetComponent("inventory"); ok {
> 209: 			inv := invComp.(*InventoryComponent)
  210: 			currentInventory = inv.Items
  211: 		}
  212: 	}
```

#### Line 349: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  346: 		return
  347: 	}
  348: 
> 349: 	playerInv := playerInvComp.(*InventoryComponent)
  350: 	merchant := merchantComp.(*MerchantComponent)
  351: 
  352: 	// Draw semi-transparent overlay
```

#### Line 350: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  347: 	}
  348: 
  349: 	playerInv := playerInvComp.(*InventoryComponent)
> 350: 	merchant := merchantComp.(*MerchantComponent)
  351: 
  352: 	// Draw semi-transparent overlay
  353: 	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
```

---

### File: `pkg/engine/skill_progression_system.go` (6 issues)

#### Line 54: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  51: 	if !ok {
  52: 		return
  53: 	}
> 54: 	treeComp := comp.(*SkillTreeComponent)
  55: 
  56: 	// Get stats component
  57: 	statsComp, ok := entity.GetComponent("stats")
```

#### Line 61: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  58: 	if !ok {
  59: 		return // No stats to modify
  60: 	}
> 61: 	stats := statsComp.(*StatsComponent)
  62: 
  63: 	// Reset bonus stats (we'll recalculate from scratch)
  64: 	bonuses := &SkillBonuses{
```

#### Line 188: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  185: 	// Apply attack/defense/magic bonuses using base stats
  186: 	baseStatsComp, hasBaseStats := entity.GetComponent("base_stats")
  187: 	if hasBaseStats {
> 188: 		baseStats := baseStatsComp.(*BaseStatsComponent)
  189: 
  190: 		// Apply attack bonus
  191: 		if bonuses.DamageBonus != 0 {
```

#### Line 193: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  190: 		// Apply attack bonus
  191: 		if bonuses.DamageBonus != 0 {
  192: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 193: 				attack := attackComp.(*AttackComponent)
  194: 				attack.Damage = baseStats.BaseAttack * (1.0 + bonuses.DamageBonus)
  195: 			}
  196: 		}
```

#### Line 211: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  208: 		// Apply health bonus
  209: 		if bonuses.HealthBonus != 0 {
  210: 			if healthComp, ok := entity.GetComponent("health"); ok {
> 211: 				health := healthComp.(*HealthComponent)
  212: 				oldMax := health.Max
  213: 				health.Max = baseStats.BaseMaxHealth * (1.0 + bonuses.HealthBonus)
  214: 				// Scale current health proportionally
```

#### Line 224: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  221: 		// Apply mana regen bonus
  222: 		if bonuses.ManaRegenBonus != 0 {
  223: 			if manaComp, ok := entity.GetComponent("mana"); ok {
> 224: 				mana := manaComp.(*ManaComponent)
  225: 				mana.Regen = baseStats.BaseManaRegen * (1.0 + bonuses.ManaRegenBonus)
  226: 			}
  227: 		}
```

---

### File: `pkg/engine/skill_tree_loader.go` (2 issues)

#### Line 42: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  39: 		return err
  40: 	}
  41: 
> 42: 	trees := result.([]*skills.SkillTree)
  43: 	if len(trees) == 0 {
  44: 		return nil // No trees generated, not an error
  45: 	}
```

#### Line 59: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  56: 		// Update existing component with new tree
  57: 		comp, ok := player.GetComponent("skill_tree")
  58: 		if ok {
> 59: 			treeComp := comp.(*SkillTreeComponent)
  60: 			treeComp.Tree = mainTree
  61: 		}
  62: 	}
```

---

### File: `pkg/engine/skills_ui.go` (5 issues)

#### Line 129: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  126: 	}
  127: 
  128: 	if comp, ok := ui.playerEntity.GetComponent("skill_tree"); ok {
> 129: 		ui.skillTreeComp = comp.(*SkillTreeComponent)
  130: 	}
  131: }
  132: 
```

#### Line 506: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  503: 	if ui.skillTreeComp.LearnSkill(skillID, availablePoints) {
  504: 		// Deduct skill points from experience component
  505: 		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 506: 			exp := expComp.(*ExperienceComponent)
  507: 			skill := ui.skillTreeComp.Tree.GetSkillByID(skillID)
  508: 			if skill != nil {
  509: 				exp.SkillPoints -= skill.Requirements.SkillPoints
```

#### Line 533: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  530: 	if pointsRefunded > 0 {
  531: 		// Refund skill points to experience component
  532: 		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 533: 			exp := expComp.(*ExperienceComponent)
  534: 			exp.SkillPoints += pointsRefunded
  535: 		}
  536: 
```

#### Line 609: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  606: 	}
  607: 
  608: 	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 609: 		return expComp.(*ExperienceComponent).SkillPoints
  610: 	}
  611: 
  612: 	return 0
```

#### Line 622: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  619: 	}
  620: 
  621: 	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 622: 		return expComp.(*ExperienceComponent).Level
  623: 	}
  624: 
  625: 	return 1
```

---

### File: `pkg/engine/spatial_partition.go` (3 issues)

#### Line 64: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  61: 	if !ok {
  62: 		return false
  63: 	}
> 64: 	pos := posComp.(*PositionComponent)
  65: 
  66: 	// Check if point is in bounds
  67: 	if !q.bounds.Contains(pos.X, pos.Y) {
```

#### Line 135: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  132: 		if !ok {
  133: 			continue
  134: 		}
> 135: 		pos := posComp.(*PositionComponent)
  136: 
  137: 		if queryBounds.Contains(pos.X, pos.Y) {
  138: 			*result = append(*result, entity)
```

#### Line 172: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  169: 		if !ok {
  170: 			continue
  171: 		}
> 172: 		pos := posComp.(*PositionComponent)
  173: 
  174: 		dx := pos.X - x
  175: 		dy := pos.Y - y
```

---

### File: `pkg/engine/spell_casting.go` (32 issues)

#### Line 122: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  119: 		if !hasSpells {
  120: 			continue
  121: 		}
> 122: 		slots := spellComp.(*SpellSlotComponent)
  123: 
  124: 		// Update cooldowns
  125: 		for i := range slots.Cooldowns {
```

#### Line 173: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  170: 	if !hasMana {
  171: 		return
  172: 	}
> 173: 	mana := manaComp.(*ManaComponent)
  174: 
  175: 	if mana.Current < spell.Stats.ManaCost {
  176: 		// Not enough mana - show notification to player
```

#### Line 194: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  191: 	if !hasPos {
  192: 		return
  193: 	}
> 194: 	pos := posComp.(*PositionComponent)
  195: 
  196: 	// Apply spell effects based on type
  197: 	switch spell.Type {
```

#### Line 244: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  241: 		if !hasHealth {
  242: 			continue
  243: 		}
> 244: 		health := healthComp.(*HealthComponent)
  245: 
  246: 		health.Current -= float64(spell.Stats.Damage)
  247: 		if health.Current < 0 {
```

#### Line 260: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  257: 		if s.particleSys != nil {
  258: 			targetPos, hasPos := target.GetComponent("position")
  259: 			if hasPos {
> 260: 				pos := targetPos.(*PositionComponent)
  261: 				// Spawn element-specific particles
  262: 				s.spawnElementalHitEffect(pos.X, pos.Y, spell.Element, target.ID)
  263: 			}
```

#### Line 300: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  297: 	if !hasHealth {
  298: 		return
  299: 	}
> 300: 	health := healthComp.(*HealthComponent)
  301: 
  302: 	health.Current += float64(spell.Stats.Healing)
  303: 	if health.Current > health.Max {
```

#### Line 311: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  308: 	if s.particleSys != nil {
  309: 		targetPos, hasPos := target.GetComponent("position")
  310: 		if hasPos {
> 311: 			pos := targetPos.(*PositionComponent)
  312: 			config := particles.Config{
  313: 				Type:     particles.ParticleMagic,
  314: 				Count:    20,
```

#### Line 344: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  341: 	// Get caster's team
  342: 	var casterTeamID int
  343: 	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
> 344: 		casterTeamID = teamComp.(*TeamComponent).TeamID
  345: 	}
  346: 
  347: 	for _, entity := range entities {
```

#### Line 354: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  351: 
  352: 		// Check if ally
  353: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 354: 			team := teamComp.(*TeamComponent)
  355: 			if !team.IsAlly(casterTeamID) {
  356: 				continue
  357: 			}
```

#### Line 368: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  365: 		if !hasHealth {
  366: 			continue
  367: 		}
> 368: 		health := healthComp.(*HealthComponent)
  369: 		if health.Current >= health.Max {
  370: 			continue // At full health
  371: 		}
```

#### Line 392: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  389: 	// Get caster's team
  390: 	var casterTeamID int
  391: 	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
> 392: 		casterTeamID = teamComp.(*TeamComponent).TeamID
  393: 	}
  394: 
  395: 	for _, entity := range entities {
```

#### Line 398: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  395: 	for _, entity := range entities {
  396: 		// Check if ally (including self)
  397: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 398: 			team := teamComp.(*TeamComponent)
  399: 			if !team.IsAlly(casterTeamID) {
  400: 				continue
  401: 			}
```

#### Line 476: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  473: 		if spell.Stats.Damage > 0 {
  474: 			healthComp, hasHealth := target.GetComponent("health")
  475: 			if hasHealth {
> 476: 				health := healthComp.(*HealthComponent)
  477: 				health.Current -= float64(spell.Stats.Damage)
  478: 				if health.Current < 0 {
  479: 					health.Current = 0
```

#### Line 574: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  571: 	if !hasPos {
  572: 		return
  573: 	}
> 574: 	pos := posComp.(*PositionComponent)
  575: 
  576: 	// Calculate teleport direction and distance
  577: 	// Use spell range as max teleport distance
```

#### Line 587: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  584: 	// Use velocity direction if available, otherwise default direction
  585: 	dirX, dirY := 0.0, 1.0 // Default: down
  586: 	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
> 587: 		vel := velComp.(*VelocityComponent)
  588: 		if vel.VX != 0 || vel.VY != 0 {
  589: 			// Normalize velocity to get direction
  590: 			mag := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
```

#### Line 641: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  638: 	if !hasPos {
  639: 		return
  640: 	}
> 641: 	pos := posComp.(*PositionComponent)
  642: 
  643: 	// Determine reveal radius from spell stats
  644: 	revealRadius := spell.Stats.AreaSize
```

#### Line 708: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  705: 	if s.particleSys != nil {
  706: 		posComp, hasPos := caster.GetComponent("position")
  707: 		if hasPos {
> 708: 			pos := posComp.(*PositionComponent)
  709: 			config := particles.Config{
  710: 				Type:     particles.ParticleDust,
  711: 				Count:    25,
```

#### Line 748: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  745: 		if !hasCollider {
  746: 			continue
  747: 		}
> 748: 		collider := colliderComp.(*ColliderComponent)
  749: 
  750: 		// Skip non-solid colliders
  751: 		if !collider.Solid {
```

#### Line 760: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  757: 		if !hasPos {
  758: 			continue
  759: 		}
> 760: 		pos := entityPos.(*PositionComponent)
  761: 
  762: 		// Get caster collider for size checking
  763: 		casterCollider, hasCasterCollider := caster.GetComponent("collider")
```

#### Line 765: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  762: 		// Get caster collider for size checking
  763: 		casterCollider, hasCasterCollider := caster.GetComponent("collider")
  764: 		if hasCasterCollider {
> 765: 			cc := casterCollider.(*ColliderComponent)
  766: 			// Create temporary collider at target position
  767: 			tempCollider := &ColliderComponent{
  768: 				Width:   cc.Width,
```

#### Line 876: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  873: 		if !hasCasterPos {
  874: 			break
  875: 		}
> 876: 		casterPosComp := casterPos.(*PositionComponent)
  877: 
  878: 		// Get caster's facing direction (use velocity or mouse aim)
  879: 		dirX, dirY := s.getCasterDirection(caster, x, y)
```

#### Line 901: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  898: 			if !hasPos {
  899: 				continue
  900: 			}
> 901: 			entityPosComp := entityPos.(*PositionComponent)
  902: 
  903: 			// Vector from caster to entity
  904: 			toEntityX := entityPosComp.X - casterPosComp.X
```

#### Line 933: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  930: 		if !hasCasterPos {
  931: 			break
  932: 		}
> 933: 		casterPosComp := casterPos.(*PositionComponent)
  934: 
  935: 		// Get caster's facing direction (use velocity or mouse aim)
  936: 		dirX, dirY := s.getCasterDirection(caster, x, y)
```

#### Line 958: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  955: 			if !hasPos {
  956: 				continue
  957: 			}
> 958: 			entityPosComp := entityPos.(*PositionComponent)
  959: 
  960: 			// Vector from caster to entity
  961: 			toEntityX := entityPosComp.X - casterPosComp.X
```

#### Line 997: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  994: func (s *SpellCastingSystem) getCasterDirection(caster *Entity, targetX, targetY float64) (dirX, dirY float64) {
  995: 	// Try to use velocity for moving entities
  996: 	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
> 997: 		vel := velComp.(*VelocityComponent)
  998: 		if vel.VX != 0 || vel.VY != 0 {
  999: 			return vel.VX, vel.VY
  1000: 		}
```

#### Line 1005: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1002: 
  1003: 	// Fall back to direction towards target point
  1004: 	if posComp, hasPos := caster.GetComponent("position"); hasPos {
> 1005: 		pos := posComp.(*PositionComponent)
  1006: 		dirX = targetX - pos.X
  1007: 		dirY = targetY - pos.Y
  1008: 		return dirX, dirY
```

#### Line 1021: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1018: 	if !hasSpells {
  1019: 		return false
  1020: 	}
> 1021: 	slots := spellComp.(*SpellSlotComponent)
  1022: 
  1023: 	// Check if already casting
  1024: 	if slots.IsCasting() {
```

#### Line 1044: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1041: 	if !hasMana {
  1042: 		return false
  1043: 	}
> 1044: 	mana := manaComp.(*ManaComponent)
  1045: 	if mana.Current < spell.Stats.ManaCost {
  1046: 		return false
  1047: 	}
```

#### Line 1062: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1059: 	if !hasSpells {
  1060: 		return
  1061: 	}
> 1062: 	slots := spellComp.(*SpellSlotComponent)
  1063: 
  1064: 	slots.Casting = -1
  1065: 	slots.CastingBar = 0
```

#### Line 1218: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1215: 		if !hasMana {
  1216: 			continue
  1217: 		}
> 1218: 		mana := manaComp.(*ManaComponent)
  1219: 
  1220: 		// Regenerate mana
  1221: 		mana.Current += int(mana.Regen * deltaTime)
```

#### Line 1246: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1243: 		return err
  1244: 	}
  1245: 
> 1246: 	spells := result.([]*magic.Spell)
  1247: 
  1248: 	// Create spell slots component if doesn't exist
  1249: 	var slots *SpellSlotComponent
```

#### Line 1257: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  1254: 		player.AddComponent(slots)
  1255: 	} else {
  1256: 		slotsComp, _ := player.GetComponent("spell_slots")
> 1257: 		slots = slotsComp.(*SpellSlotComponent)
  1258: 	}
  1259: 
  1260: 	// Equip spells to slots
```

---

### File: `pkg/engine/squad_behaviors.go` (22 issues)

#### Line 24: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  21: 		if !ok {
  22: 			return NodeFailure
  23: 		}
> 24: 		pos := posComp.(*PositionComponent)
  25: 
  26: 		// Calculate distance to formation position
  27: 		dx := targetX - pos.X
```

#### Line 35: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  32: 		if distance < 5.0 {
  33: 			// Stop movement
  34: 			if velComp, ok := entity.GetComponent("velocity"); ok {
> 35: 				vel := velComp.(*VelocityComponent)
  36: 				vel.VX = 0
  37: 				vel.VY = 0
  38: 			}
```

#### Line 44: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  41: 
  42: 		// Move towards formation position
  43: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 44: 			vel := velComp.(*VelocityComponent)
  45: 			speed := 80.0 // Formation movement speed
  46: 
  47: 			// Normalize direction
```

#### Line 70: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  67: 		if !ok {
  68: 			return NodeFailure // Not in a squad
  69: 		}
> 70: 		squad := squadComp.(*SquadComponent)
  71: 
  72: 		// Get positions
  73: 		posComp, ok := entity.GetComponent("position")
```

#### Line 77: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  74: 		if !ok {
  75: 			return NodeFailure
  76: 		}
> 77: 		pos := posComp.(*PositionComponent)
  78: 
  79: 		targetPosComp, ok := target.GetComponent("position")
  80: 		if !ok {
```

#### Line 83: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  80: 		if !ok {
  81: 			return NodeFailure
  82: 		}
> 83: 		targetPos := targetPosComp.(*PositionComponent)
  84: 
  85: 		// Determine flanking position based on squad position index
  86: 		// Even indices flank left, odd indices flank right
```

#### Line 111: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  108: 		}
  109: 
  110: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 111: 			vel := velComp.(*VelocityComponent)
  112: 			speed := 90.0
  113: 			vel.VX = (dx / distance) * speed
  114: 			vel.VY = (dy / distance) * speed
```

#### Line 129: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  126: 		if !ok {
  127: 			return NodeFailure
  128: 		}
> 129: 		squad := squadComp.(*SquadComponent)
  130: 
  131: 		// Check if squad has a shared priority target
  132: 		priorityTarget, ok := squad.SharedBlackboard.GetEntity("priority_target")
```

#### Line 151: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  148: 		if !ok {
  149: 			return NodeFailure
  150: 		}
> 151: 		squad := squadComp.(*SquadComponent)
  152: 
  153: 		// Check if we can alert (cooldown)
  154: 		currentTime, _ := blackboard.GetFloat64("game_time")
```

#### Line 164: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  161: 		if !ok {
  162: 			return NodeFailure
  163: 		}
> 164: 		pos := posComp.(*PositionComponent)
  165: 
  166: 		// Find nearby squads within alert range (500 pixels)
  167: 		alertRange := 500.0
```

#### Line 176: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  173: 			}
  174: 
  175: 			otherSquadComp, _ := other.GetComponent("squad")
> 176: 			otherSquad := otherSquadComp.(*SquadComponent)
  177: 
  178: 			// Skip same squad
  179: 			if otherSquad.SquadID == squad.SquadID {
```

#### Line 184: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  181: 			}
  182: 
  183: 			otherPosComp, _ := other.GetComponent("position")
> 184: 			otherPos := otherPosComp.(*PositionComponent)
  185: 
  186: 			// Check distance
  187: 			dx := otherPos.X - pos.X
```

#### Line 215: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  212: 		if !ok {
  213: 			return NodeFailure
  214: 		}
> 215: 		squad := squadComp.(*SquadComponent)
  216: 
  217: 		// Check if squad leader ordered retreat
  218: 		shouldRetreat, ok := squad.SharedBlackboard.GetBool("retreat_ordered")
```

#### Line 233: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  230: 
  231: 		// Move in retreat direction
  232: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 233: 			vel := velComp.(*VelocityComponent)
  234: 			speed := 120.0 // Retreat faster
  235: 
  236: 			// Normalize direction
```

#### Line 253: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  250: 	return NewActionNode("SquadLeaderDecision", func(entity *Entity, blackboard *Blackboard, deltaTime float64) NodeStatus {
  251: 		// Get squad component
  252: 		squadComp, ok := entity.GetComponent("squad")
> 253: 		if !ok || !squadComp.(*SquadComponent).IsLeader() {
  254: 			return NodeFailure
  255: 		}
  256: 		squad := squadComp.(*SquadComponent)
```

#### Line 256: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  253: 		if !ok || !squadComp.(*SquadComponent).IsLeader() {
  254: 			return NodeFailure
  255: 		}
> 256: 		squad := squadComp.(*SquadComponent)
  257: 
  258: 		// Leader makes tactical decisions for the squad
  259: 		// Check if squad should retreat (simple heuristic: if leader health low)
```

#### Line 262: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  259: 		// Check if squad should retreat (simple heuristic: if leader health low)
  260: 		healthComp, ok := entity.GetComponent("health")
  261: 		if ok {
> 262: 			health := healthComp.(*HealthComponent)
  263: 			healthPercent := float64(health.Current) / float64(health.Max)
  264: 
  265: 			if healthPercent < 0.3 {
```

#### Line 272: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  269: 				// Calculate retreat direction (away from current target)
  270: 				if target, ok := blackboard.GetEntity("target"); ok {
  271: 					posComp, _ := entity.GetComponent("position")
> 272: 					pos := posComp.(*PositionComponent)
  273: 					targetPosComp, _ := target.GetComponent("position")
  274: 					targetPos := targetPosComp.(*PositionComponent)
  275: 
```

#### Line 274: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  271: 					posComp, _ := entity.GetComponent("position")
  272: 					pos := posComp.(*PositionComponent)
  273: 					targetPosComp, _ := target.GetComponent("position")
> 274: 					targetPos := targetPosComp.(*PositionComponent)
  275: 
  276: 					// Retreat direction is away from target
  277: 					dx := pos.X - targetPos.X
```

#### Line 306: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  303: 		if !ok {
  304: 			return false
  305: 		}
> 306: 		return squadComp.(*SquadComponent).IsLeader()
  307: 	})
  308: }
  309: 
```

#### Line 317: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  314: 		if !ok {
  315: 			return false
  316: 		}
> 317: 		squad := squadComp.(*SquadComponent)
  318: 
  319: 		target, ok := squad.SharedBlackboard.GetEntity("priority_target")
  320: 		return ok && target != nil
```

#### Line 331: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  328: 		if !ok {
  329: 			return false
  330: 		}
> 331: 		squad := squadComp.(*SquadComponent)
  332: 
  333: 		retreat, ok := squad.SharedBlackboard.GetBool("retreat_ordered")
  334: 		return ok && retreat
```

---

### File: `pkg/engine/squad_system.go` (8 issues)

#### Line 45: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  42: 		if !ok {
  43: 			continue
  44: 		}
> 45: 		squad := squadComp.(*SquadComponent)
  46: 		squads[squad.SquadID] = append(squads[squad.SquadID], entity)
  47: 	}
  48: 	return squads
```

#### Line 62: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  59: 	var leaderSquad *SquadComponent
  60: 	for _, member := range members {
  61: 		squadComp, _ := member.GetComponent("squad")
> 62: 		squad := squadComp.(*SquadComponent)
  63: 		if squad.IsLeader() {
  64: 			leader = member
  65: 			leaderSquad = squad
```

#### Line 89: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  86: 	if !ok {
  87: 		return
  88: 	}
> 89: 	leaderPos := leaderPosComp.(*PositionComponent)
  90: 
  91: 	// Get leader's rotation for oriented formations
  92: 	leaderAngle := 0.0
```

#### Line 94: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  91: 	// Get leader's rotation for oriented formations
  92: 	leaderAngle := 0.0
  93: 	if rotComp, ok := leader.GetComponent("rotation"); ok {
> 94: 		rot := rotComp.(*RotationComponent)
  95: 		leaderAngle = rot.Angle
  96: 	}
  97: 
```

#### Line 106: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  103: 		}
  104: 
  105: 		squadComp, _ := member.GetComponent("squad")
> 106: 		squad := squadComp.(*SquadComponent)
  107: 
  108: 		// Calculate target position based on formation type
  109: 		targetX, targetY := s.calculateFormationPosition(
```

#### Line 120: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  117: 
  118: 		// Store target position in member's blackboard
  119: 		if btComp, ok := member.GetComponent("behaviortree"); ok {
> 120: 			bt := btComp.(*BehaviorTreeComponent)
  121: 			bt.Blackboard.Set("formation_target_x", targetX)
  122: 			bt.Blackboard.Set("formation_target_y", targetY)
  123: 		}
```

#### Line 189: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  186: 	var sharedBlackboard *Blackboard
  187: 	for _, member := range members {
  188: 		squadComp, _ := member.GetComponent("squad")
> 189: 		squad := squadComp.(*SquadComponent)
  190: 		if squad.SharedBlackboard != nil {
  191: 			sharedBlackboard = squad.SharedBlackboard
  192: 			break
```

#### Line 203: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  200: 	// Ensure all members reference the same shared blackboard
  201: 	for _, member := range members {
  202: 		squadComp, _ := member.GetComponent("squad")
> 203: 		squad := squadComp.(*SquadComponent)
  204: 		squad.SharedBlackboard = sharedBlackboard
  205: 	}
  206: }
```

---

### File: `pkg/engine/station_spawn.go` (3 issues)

#### Line 281: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  278: 			continue
  279: 		}
  280: 
> 281: 		pos := posComp.(*PositionComponent)
  282: 
  283: 		// Calculate squared distance (avoid sqrt for performance)
  284: 		dx := pos.X - centerX
```

#### Line 327: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  324: 			continue
  325: 		}
  326: 
> 327: 		pos := posComp.(*PositionComponent)
  328: 
  329: 		// Calculate distance
  330: 		dx := pos.X - centerX
```

#### Line 361: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  358: 		return ""
  359: 	}
  360: 
> 361: 	station := stationComp.(*CraftingStationComponent)
  362: 
  363: 	recipeTypeName := ""
  364: 	switch station.StationType {
```

---

### File: `pkg/engine/status_effect_pool.go` (1 issues)

#### Line 35: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  32: //	// ... later when effect expires ...
  33: //	ReleaseStatusEffect(effect)
  34: func NewStatusEffectComponent(effectType string, magnitude, duration, tickInterval float64) *StatusEffectComponent {
> 35: 	effect := statusEffectPool.Get().(*StatusEffectComponent)
  36: 
  37: 	// Initialize with provided values
  38: 	effect.EffectType = effectType
```

---

### File: `pkg/engine/status_effect_system.go` (8 issues)

#### Line 54: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  51: 
  52: 		// Update shield duration
  53: 		if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
> 54: 			shield := shieldComp.(*ShieldComponent)
  55: 			shield.Update(deltaTime)
  56: 
  57: 			// Remove depleted shields
```

#### Line 71: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  68: 	if !hasHealth {
  69: 		return
  70: 	}
> 71: 	health := healthComp.(*HealthComponent)
  72: 
  73: 	switch effect.EffectType {
  74: 	case "burning":
```

#### Line 94: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  91: 	if !hasStats {
  92: 		return
  93: 	}
> 94: 	stats := statsComp.(*StatsComponent)
  95: 
  96: 	switch effect.EffectType {
  97: 	case "strength":
```

#### Line 145: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  142: 	if !hasStats {
  143: 		return
  144: 	}
> 145: 	stats := statsComp.(*StatsComponent)
  146: 
  147: 	switch effect.EffectType {
  148: 	case "strength":
```

#### Line 171: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  168: 	// Check if shield already exists
  169: 	if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
  170: 		// Add to existing shield
> 171: 		shield := shieldComp.(*ShieldComponent)
  172: 		shield.Amount += amount
  173: 		if shield.Amount > shield.MaxAmount {
  174: 			shield.MaxAmount = shield.Amount
```

#### Line 200: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  197: 
  198: 	// Apply damage to initial target
  199: 	if healthComp, hasHealth := initialTarget.GetComponent("health"); hasHealth {
> 200: 		health := healthComp.(*HealthComponent)
  201: 		health.TakeDamage(damage)
  202: 
  203: 		// Apply shocked effect
```

#### Line 255: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  252: 	// Check team if available
  253: 	if casterTeam, hasCasterTeam := caster.GetComponent("team"); hasCasterTeam {
  254: 		if targetTeam, hasTargetTeam := target.GetComponent("team"); hasTargetTeam {
> 255: 			ct := casterTeam.(*TeamComponent)
  256: 			tt := targetTeam.(*TeamComponent)
  257: 			return ct.IsEnemy(tt.TeamID)
  258: 		}
```

#### Line 256: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  253: 	if casterTeam, hasCasterTeam := caster.GetComponent("team"); hasCasterTeam {
  254: 		if targetTeam, hasTargetTeam := target.GetComponent("team"); hasTargetTeam {
  255: 			ct := casterTeam.(*TeamComponent)
> 256: 			tt := targetTeam.(*TeamComponent)
  257: 			return ct.IsEnemy(tt.TeamID)
  258: 		}
  259: 	}
```

---

### File: `pkg/engine/terrain_collision_system.go` (3 issues)

#### Line 165: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  162: 	posComp, _ := entity.GetComponent("position")
  163: 	colliderComp, _ := entity.GetComponent("collider")
  164: 
> 165: 	pos := posComp.(*PositionComponent)
  166: 	collider := colliderComp.(*ColliderComponent)
  167: 
  168: 	// Get entity's layer (default to ground layer if no layer component)
```

#### Line 166: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  163: 	colliderComp, _ := entity.GetComponent("collider")
  164: 
  165: 	pos := posComp.(*PositionComponent)
> 166: 	collider := colliderComp.(*ColliderComponent)
  167: 
  168: 	// Get entity's layer (default to ground layer if no layer component)
  169: 	layer := 0
```

#### Line 172: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  169: 	layer := 0
  170: 	if entity.HasComponent("layer") {
  171: 		layerComp, _ := entity.GetComponent("layer")
> 172: 		layerComponent := layerComp.(*LayerComponent)
  173: 		layer = layerComponent.GetEffectiveLayer()
  174: 	}
  175: 
```

---

### File: `pkg/engine/tile_cache.go` (1 issues)

#### Line 129: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  126: 	}
  127: 
  128: 	c.lruList.Remove(elem)
> 129: 	keyStr := elem.Value.(string)
  130: 	delete(c.cache, keyStr)
  131: }
  132: 
```

---

### File: `pkg/engine/tutorial_system.go` (5 issues)

#### Line 86: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  83: 						if !ok {
  84: 							continue
  85: 						}
> 86: 						pos := comp.(*PositionComponent)
  87: 						// Simple distance check from origin (400, 300 typical spawn)
  88: 						distFromStart := (pos.X-400)*(pos.X-400) + (pos.Y-300)*(pos.Y-300)
  89: 						return distFromStart > 2500 // ~50 units
```

#### Line 109: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  106: 						if !ok {
  107: 							continue
  108: 						}
> 109: 						attack := comp.(*AttackComponent)
  110: 						// Check if attack cooldown is active (means they attacked)
  111: 						return attack.CooldownTimer > 0 || attack.CooldownTimer < attack.Cooldown
  112: 					}
```

#### Line 131: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  128: 						if !ok {
  129: 							continue
  130: 						}
> 131: 						health := comp.(*HealthComponent)
  132: 						// Complete if health is damaged but still above 50%
  133: 						return health.Current < health.Max && health.Current > health.Max/2
  134: 					}
```

#### Line 153: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  150: 						if !ok {
  151: 							continue
  152: 						}
> 153: 						inv := comp.(*InventoryComponent)
  154: 						return len(inv.Items) > 0
  155: 					}
  156: 				}
```

#### Line 174: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  171: 						if !ok {
  172: 							continue
  173: 						}
> 174: 						exp := comp.(*ExperienceComponent)
  175: 						return exp.Level >= 2
  176: 					}
  177: 				}
```

---

### File: `pkg/engine/vehicle_combat_system.go` (10 issues)

#### Line 61: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  58: 		if !hasCombat {
  59: 			continue
  60: 		}
> 61: 		combat := combatComp.(*VehicleCombatComponent)
  62: 
  63: 		// Update cooldowns
  64: 		combat.UpdateCooldowns(deltaTime)
```

#### Line 81: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  78: 	if !hasVehicle {
  79: 		return
  80: 	}
> 81: 	v := vehicleComp.(*VehicleComponent)
  82: 
  83: 	// Check if can ram (cooldown ready and sufficient speed)
  84: 	if !combat.CanRam(v.Speed) {
```

#### Line 92: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  89: 	if !hasPos {
  90: 		return
  91: 	}
> 92: 	pos := posComp.(*PositionComponent)
  93: 
  94: 	// Check for entities in ramming range (small radius around vehicle)
  95: 	rammingRadius := 20.0 // pixels
```

#### Line 112: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  109: 		if !hasTargetPos {
  110: 			continue
  111: 		}
> 112: 		tPos := targetPos.(*PositionComponent)
  113: 
  114: 		// Calculate distance
  115: 		dx := tPos.X - pos.X
```

#### Line 125: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  122: 			damage := combat.CalculateRammingDamage(v.Speed)
  123: 
  124: 			// Apply damage to target
> 125: 			health := targetHealth.(*HealthComponent)
  126: 			health.TakeDamage(damage)
  127: 
  128: 			// Execute ramming attack (set cooldown)
```

#### Line 162: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  159: 	if !hasPos {
  160: 		return
  161: 	}
> 162: 	pos := posComp.(*PositionComponent)
  163: 
  164: 	// Check if vehicle is player-controlled or has AI targeting
  165: 	hasControl := vcs.hasTargetInput(vehicle)
```

#### Line 180: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  177: 	if vcs.projectileSystem != nil {
  178: 		// Calculate direction to target
  179: 		targetPos, _ := target.GetComponent("position")
> 180: 		tPos := targetPos.(*PositionComponent)
  181: 		dx := tPos.X - pos.X
  182: 		dy := tPos.Y - pos.Y
  183: 		targetAngle := math.Atan2(dy, dx)
```

#### Line 193: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  190: 		// Direct damage application (fallback if no projectile system)
  191: 		targetHealth, hasHealth := target.GetComponent("health")
  192: 		if hasHealth {
> 193: 			health := targetHealth.(*HealthComponent)
  194: 			health.TakeDamage(combat.WeaponDamage)
  195: 		}
  196: 	}
```

#### Line 222: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  219: 	// Check if mounted by player
  220: 	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
  221: 	if hasVehicle {
> 222: 		v := vehicleComp.(*VehicleComponent)
  223: 		if v.CurrentPassengers > 0 {
  224: 			return true
  225: 		}
```

#### Line 256: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  253: 		if !hasTargetPos {
  254: 			continue
  255: 		}
> 256: 		tPos := targetPos.(*PositionComponent)
  257: 
  258: 		// Calculate distance
  259: 		dx := tPos.X - pos.X
```

---

### File: `pkg/engine/vehicle_durability_system.go` (4 issues)

#### Line 43: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  40: 			continue
  41: 		}
  42: 
> 43: 		vehicle := vehicleComp.(*VehicleComponent)
  44: 
  45: 		// Skip if already destroyed
  46: 		if vehicle.IsDestroyed() {
```

#### Line 67: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  64: 		return
  65: 	}
  66: 
> 67: 	collider := collComp.(*ColliderComponent)
  68: 
  69: 	// In a full implementation, this would check if the collider
  70: 	// recently hit a solid object and apply damage based on speed.
```

#### Line 114: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  111: 		return false
  112: 	}
  113: 
> 114: 	vehicle := vehicleComp.(*VehicleComponent)
  115: 	destroyed := vehicle.TakeDamage(damage)
  116: 
  117: 	if destroyed {
```

#### Line 131: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  128: 		return false
  129: 	}
  130: 
> 131: 	vehicle := vehicleComp.(*VehicleComponent)
  132: 	vehicle.Repair(amount)
  133: 	return true
  134: }
```

---

### File: `pkg/engine/vehicle_movement_system.go` (4 issues)

#### Line 47: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  44: 			continue
  45: 		}
  46: 
> 47: 		vehicle := vehicleComp.(*VehicleComponent)
  48: 
  49: 		// Skip if vehicle is destroyed or out of fuel
  50: 		if vehicle.IsDestroyed() || vehicle.IsFuelDepleted() {
```

#### Line 60: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  57: 		if !hasPos {
  58: 			continue // Can't move without position
  59: 		}
> 60: 		pos := posComp.(*PositionComponent)
  61: 
  62: 		rotComp, hasRot := entity.GetComponent("rotation")
  63: 		if !hasRot {
```

#### Line 66: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  63: 		if !hasRot {
  64: 			continue // Need rotation for direction
  65: 		}
> 66: 		rot := rotComp.(*RotationComponent)
  67: 
  68: 		// Check if entity is being controlled (has input or is mounted)
  69: 		hasControl := vms.hasControlInput(entity)
```

#### Line 112: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  109: 
  110: 	// Check if any entity is mounted on this vehicle
  111: 	vehicleComp, _ := entity.GetComponent("vehicle")
> 112: 	vehicle := vehicleComp.(*VehicleComponent)
  113: 	return vehicle.CurrentPassengers > 0
  114: }
  115: 
```

---

### File: `pkg/engine/visual_feedback_components.go` (1 issues)

#### Line 92: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  89: 			continue
  90: 		}
  91: 
> 92: 		feedback := feedbackComp.(*VisualFeedbackComponent)
  93: 
  94: 		// Update flash timer
  95: 		if feedback.FlashTimer > 0 {
```

---

### File: `pkg/engine/weather_system.go` (1 issues)

#### Line 160: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  157: 
  158: 			// Apply transition opacity to particle color
  159: 			// Convert color.Color interface to color.RGBA for alpha manipulation
> 160: 			rgba := color.RGBAModel.Convert(p.Color).(color.RGBA)
  161: 			// Adjust alpha based on opacity
  162: 			originalAlpha := float64(rgba.A)
  163: 			rgba.A = uint8(originalAlpha * opacity)
```

---

### File: `pkg/hostplay/input_handler.go` (2 issues)

#### Line 91: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  88: 	}
  89: 
  90: 	// Extract direction from input data
> 91: 	dx, dxOk := data["dx"].(float64)
  92: 	dy, dyOk := data["dy"].(float64)
  93: 	if !dxOk || !dyOk {
  94: 		return
```

#### Line 92: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  89: 
  90: 	// Extract direction from input data
  91: 	dx, dxOk := data["dx"].(float64)
> 92: 	dy, dyOk := data["dy"].(float64)
  93: 	if !dxOk || !dyOk {
  94: 		return
  95: 	}
```

---

### File: `pkg/hostplay/server_manager.go` (2 issues)

#### Line 142: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  139: 	if err != nil {
  140: 		return fmt.Errorf("failed to generate terrain: %w", err)
  141: 	}
> 142: 	sm.generatedTerrain = terrainResult.(*terrain.Terrain)
  143: 
  144: 	sm.logger.WithFields(logrus.Fields{
  145: 		"width":     sm.generatedTerrain.Width,
```

#### Line 318: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  315: 	for _, entity := range sm.world.GetEntities() {
  316: 		netComp, exists := entity.GetComponent("network")
  317: 		if exists && netComp != nil {
> 318: 			nc := netComp.(*engine.NetworkComponent)
  319: 			if nc.PlayerID == playerID {
  320: 				sm.world.RemoveEntity(entity.ID)
  321: 				sm.logger.WithFields(logrus.Fields{
```

---

### File: `pkg/mobile/ui.go` (2 issues)

#### Line 520: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  517: 		alpha = uint8(n.Remaining * 255)
  518: 	}
  519: 
> 520: 	bgColor := n.BackgroundColor.(color.RGBA)
  521: 	bgColor.A = alpha
  522: 
  523: 	// Draw background
```

#### Line 528: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  525: 
  526: 	// Draw message text
  527: 	if n.Message != "" {
> 528: 		textColor := n.TextColor.(color.RGBA)
  529: 		textColor.A = alpha
  530: 
  531: 		// Center text horizontally and vertically
```

---

### File: `pkg/network/buffer_pool.go` (1 issues)

#### Line 33: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  30: //	defer ReleaseBuffer(buf)
  31: //	*buf = append(*buf, data...)
  32: func AcquireBuffer() *[]byte {
> 33: 	return bufferPool.Get().(*[]byte)
  34: }
  35: 
  36: // ReleaseBuffer returns a buffer to the pool for reuse.
```

---

### File: `pkg/network/priority_queue.go` (2 issues)

#### Line 52: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  49: 
  50: func (h *priorityHeap) Push(x interface{}) {
  51: 	n := len(*h)
> 52: 	item := x.(*priorityItem)
  53: 	item.index = n
  54: 	*h = append(*h, item)
  55: }
```

#### Line 113: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  110: 		return nil
  111: 	}
  112: 
> 113: 	item := heap.Pop(&pq.heap).(*priorityItem)
  114: 	return item.update
  115: }
  116: 
```

---

### File: `pkg/procgen/terrain/multilevel.go` (1 issues)

#### Line 83: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  80: 			return nil, fmt.Errorf("failed to generate level %d: %w", i, err)
  81: 		}
  82: 
> 83: 		terrain := result.(*Terrain)
  84: 		terrain.Level = i
  85: 		levels[i] = terrain
  86: 	}
```

---

### File: `pkg/rendering/cache/sprite_cache.go` (3 issues)

#### Line 89: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  86: 		// Move to front (most recently used)
  87: 		c.lru.MoveToFront(elem)
  88: 		c.stats.Hits++
> 89: 		return elem.Value.(*entry).image, true
  90: 	}
  91: 
  92: 	c.stats.Misses++
```

#### Line 106: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  103: 	if elem, ok := c.cache[key]; ok {
  104: 		// Update existing entry and move to front
  105: 		c.lru.MoveToFront(elem)
> 106: 		elem.Value.(*entry).image = img
  107: 		return
  108: 	}
  109: 
```

#### Line 143: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  140: // removeElement removes a specific element from the cache.
  141: func (c *SpriteCache) removeElement(elem *list.Element) {
  142: 	c.lru.Remove(elem)
> 143: 	e := elem.Value.(*entry)
  144: 	delete(c.cache, e.key)
  145: 	c.stats.TotalSize -= e.size
  146: 	c.stats.EntryCount--
```

---

### File: `pkg/rendering/particles/pool.go` (2 issues)

#### Line 41: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  38: //
  39: // Returns: Pooled ParticleSystem ready for use
  40: func NewParticleSystem(particles []Particle, pType ParticleType, config Config) *ParticleSystem {
> 41: 	ps := particleSystemPool.Get().(*ParticleSystem)
  42: 
  43: 	// Clear previous state
  44: 	ps.Particles = ps.Particles[:0]
```

#### Line 86: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  83: //
  84: // Returns: Pointer to slice with 0 length, 100 capacity
  85: func AcquireParticleSlice() *[]Particle {
> 86: 	particles := particleSlicePool.Get().(*[]Particle)
  87: 	*particles = (*particles)[:0] // Reset length, keep capacity
  88: 	return particles
  89: }
```

---

### File: `pkg/rendering/pool/image_pool.go` (4 issues)

#### Line 70: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  67: 	if width == height {
  68: 		switch width {
  69: 		case SizePlayer:
> 70: 			return p.pool28.Get().(*ebiten.Image)
  71: 		case SizeSmall:
  72: 			return p.pool32.Get().(*ebiten.Image)
  73: 		case SizeMedium:
```

#### Line 72: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  69: 		case SizePlayer:
  70: 			return p.pool28.Get().(*ebiten.Image)
  71: 		case SizeSmall:
> 72: 			return p.pool32.Get().(*ebiten.Image)
  73: 		case SizeMedium:
  74: 			return p.pool64.Get().(*ebiten.Image)
  75: 		case SizeLarge:
```

#### Line 74: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  71: 		case SizeSmall:
  72: 			return p.pool32.Get().(*ebiten.Image)
  73: 		case SizeMedium:
> 74: 			return p.pool64.Get().(*ebiten.Image)
  75: 		case SizeLarge:
  76: 			return p.pool128.Get().(*ebiten.Image)
  77: 		}
```

#### Line 76: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  73: 		case SizeMedium:
  74: 			return p.pool64.Get().(*ebiten.Image)
  75: 		case SizeLarge:
> 76: 			return p.pool128.Get().(*ebiten.Image)
  77: 		}
  78: 	}
  79: 
```

---

### File: `pkg/rendering/sprites/cache.go` (1 issues)

#### Line 123: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  120: 		return
  121: 	}
  122: 
> 123: 	key := element.Value.(uint64)
  124: 	c.lruList.Remove(element)
  125: 	delete(c.cache, key)
  126: }
```

---

### File: `pkg/rendering/sprites/pool.go` (1 issues)

#### Line 34: Unchecked type assertion
**Severity:** High

**Description:** Type assertion performed without checking if assertion succeeded. While GetComponent checks existence, the type assertion itself is unchecked.

**Code Reference:**
```go
  31: // Get retrieves an image from the pool or creates a new one.
  32: // The returned image will be cleared (transparent).
  33: func (p *ImagePool) Get() *ebiten.Image {
> 34: 	img := p.pool.Get().(*ebiten.Image)
  35: 	// Clear the image for reuse
  36: 	img.Clear()
  37: 	return img
```

---

