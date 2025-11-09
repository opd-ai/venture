# Code Audit Report

## AUDIT SUMMARY
**Total Issues:** 498
**By Category:** CONCURRENCY ISSUE: 22, CRITICAL BUG: 465, ERROR HANDLING GAP: 9, RESOURCE LEAK: 2
**By Severity:** High: 489 | Medium: 0 | Low: 9

---

## DETAILED FINDINGS

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:43`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  41: 		}
  42: 
> 43: 		aiState := aiComp.(*AIComponent)
  44: 
  45: 		// Update timers
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:65`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  63: 		return // Can't do AI without position
  64: 	}
> 65: 	pos := posComp.(*PositionComponent)
  66: 
  67: 	// Check health for flee condition
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:141`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  139: 		// Stop movement while waiting
  140: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 141: 			vel := velComp.(*VelocityComponent)
  142: 			vel.VX = 0
  143: 			vel.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:161`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  159: 	// Move towards waypoint
  160: 	if velComp, ok := entity.GetComponent("velocity"); ok {
> 161: 		vel := velComp.(*VelocityComponent)
  162: 
  163: 		// Use default base speed (velocity component doesn't store speed)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:214`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  212: 		return
  213: 	}
> 214: 	attack := attackComp.(*AttackComponent)
  215: 
  216: 	// Check if in attack range
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:223`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  221: 		return
  222: 	}
> 223: 	targetP := targetPos.(*PositionComponent)
  224: 
  225: 	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:250`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  248: 		return
  249: 	}
> 250: 	attack := attackComp.(*AttackComponent)
  251: 
  252: 	// Check if in attack range
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:259`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  257: 		return
  258: 	}
> 259: 	targetP := targetPos.(*PositionComponent)
  260: 
  261: 	distance := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:273`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  271: 		// GAP-018 REPAIR: Set animation to attack when attacking
  272: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 273: 			anim := animComp.(*AnimationComponent)
  274: 			if anim.CurrentState != AnimationStateAttack {
  275: 				anim.SetState(AnimationStateAttack)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:315`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  313: 		velComp, ok := entity.GetComponent("velocity")
  314: 		if ok {
> 315: 			vel := velComp.(*VelocityComponent)
  316: 			vel.VX = 0
  317: 			vel.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:322`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  320: 		// GAP-018 REPAIR: Set animation to idle when stopped
  321: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 322: 			anim := animComp.(*AnimationComponent)
  323: 			if anim.CurrentState != AnimationStateIdle {
  324: 				anim.SetState(AnimationStateIdle)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:346`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  344: 	}
  345: 
> 346: 	health := healthComp.(*HealthComponent)
  347: 	if health.Max <= 0 {
  348: 		return false
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:361`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  359: 		return nil // No team component, can't determine enemies
  360: 	}
> 361: 	team := teamComp.(*TeamComponent)
  362: 
  363: 	var nearest *Entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:376`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  374: 			continue
  375: 		}
> 376: 		otherT := otherTeam.(*TeamComponent)
  377: 
  378: 		if !team.IsEnemy(otherT.TeamID) {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:385`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  383: 		otherHealth, ok := other.GetComponent("health")
  384: 		if ok {
> 385: 			h := otherHealth.(*HealthComponent)
  386: 			if h.IsDead() {
  387: 				continue
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:396`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  394: 			continue
  395: 		}
> 396: 		otherP := otherPos.(*PositionComponent)
  397: 
  398: 		dist := ai.getDistance(pos.X, pos.Y, otherP.X, otherP.Y)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:417`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  415: 	targetHealth, ok := target.GetComponent("health")
  416: 	if ok {
> 417: 		h := targetHealth.(*HealthComponent)
  418: 		if h.IsDead() {
  419: 			return false
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:428`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  426: 		return false
  427: 	}
> 428: 	targetP := targetPos.(*PositionComponent)
  429: 
  430: 	dist := ai.getDistance(pos.X, pos.Y, targetP.X, targetP.Y)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:440`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  438: 		return // No velocity component, can't move
  439: 	}
> 440: 	vel := velComp.(*VelocityComponent)
  441: 
  442: 	// Calculate direction
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:456`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  454: 		// This makes enemies visibly rotate towards their target
  455: 		if aimComp, ok := entity.GetComponent("aim"); ok {
> 456: 			aim := aimComp.(*AimComponent)
  457: 			aim.SetAimTarget(targetX, targetY)
  458: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:462`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  460: 		// GAP-018 REPAIR: Update animation state to walk when moving
  461: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 462: 			anim := animComp.(*AnimationComponent)
  463: 			if anim.CurrentState != AnimationStateWalk {
  464: 				anim.SetState(AnimationStateWalk)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:481`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  479: 	aiComp, ok := entity.GetComponent("ai")
  480: 	if ok {
> 481: 		aiC := aiComp.(*AIComponent)
  482: 		aiC.DetectionRange = detectionRange
  483: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ai_system.go:492`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  490: 		return AIStateIdle
  491: 	}
> 492: 	return aiComp.(*AIComponent).State
  493: }
  494: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:128`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  126: 	if s.playerEntity != nil {
  127: 		if posComp, ok := s.playerEntity.GetComponent("position"); ok {
> 128: 			pos := posComp.(*PositionComponent)
  129: 			playerX = pos.X
  130: 			playerY = pos.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:139`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  137: 	if s.enableViewportCull && s.cameraSystem != nil && s.cameraSystem.activeCamera != nil {
  138: 		if camComp, ok := s.cameraSystem.activeCamera.GetComponent("camera"); ok {
> 139: 			camera := camComp.(*CameraComponent)
  140: 			// Calculate viewport bounds with margin for sprites
  141: 			margin := 100.0 // Extra margin to start animating before entity enters view
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:172`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  170: 			continue
  171: 		}
> 172: 		pos := posComp.(*PositionComponent)
  173: 
  174: 		// Phase 14.2: Viewport culling check
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:566`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  564: 		facing := "down" // Default
  565: 		if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 566: 			vel := velComp.(*VelocityComponent)
  567: 			// Use velocity direction if moving, otherwise keep last facing
  568: 			if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:597`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  595: 		}
  596: 	} else if teamComp, ok := entity.GetComponent("team"); ok {
> 597: 		team := teamComp.(*TeamComponent)
  598: 		if team.TeamID == 2 { // Enemy team
  599: 			// Determine monster type based on entity characteristics
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:604`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  602: 			// Check if it's a boss (high damage indicates boss)
  603: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 604: 				attack := attackComp.(*AttackComponent)
  605: 				if attack.Damage > 20 {
  606: 					entityType = "boss"
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:614`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  612: 			// Check size based on collider
  613: 			if colliderComp, ok := entity.GetComponent("collider"); ok {
> 614: 				collider := colliderComp.(*ColliderComponent)
  615: 				if collider.Width > 48 {
  616: 					entityType = "monster" // Large monster
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/animation_system.go:627`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  625: 			facing := "down" // Default
  626: 			if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 627: 				vel := velComp.(*VelocityComponent)
  628: 				// Use velocity direction if moving, otherwise keep last facing
  629: 				if math.Abs(vel.VX) > 0.1 || math.Abs(vel.VY) > 0.1 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:86`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  84: 			return false
  85: 		}
> 86: 		health := healthComp.(*HealthComponent)
  87: 		return float64(health.Current)/float64(health.Max) < threshold
  88: 	})
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:104`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  102: 			return false
  103: 		}
> 104: 		pos := posComp.(*PositionComponent)
  105: 
  106: 		targetPos, ok := target.GetComponent("position")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:110`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  108: 			return false
  109: 		}
> 110: 		tPos := targetPos.(*PositionComponent)
  111: 
  112: 		dx := tPos.X - pos.X
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:129`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  127: 			return NodeFailure
  128: 		}
> 129: 		pos := posComp.(*PositionComponent)
  130: 
  131: 		// Get team component to identify enemies
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:135`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  133: 		var teamID int
  134: 		if hasTeam {
> 135: 			teamID = teamComp.(*TeamComponent).TeamID
  136: 		}
  137: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:156`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  154: 					continue
  155: 				}
> 156: 				otherTeamID := otherTeamComp.(*TeamComponent).TeamID
  157: 				if otherTeamID == teamID {
  158: 					continue // Same team
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:166`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  164: 					continue
  165: 				}
> 166: 				oPos := otherPos.(*PositionComponent)
  167: 
  168: 				dx := oPos.X - pos.X
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:201`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  199: 			return NodeFailure
  200: 		}
> 201: 		pos := posComp.(*PositionComponent)
  202: 
  203: 		targetPos, ok := target.GetComponent("position")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:207`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  205: 			return NodeFailure
  206: 		}
> 207: 		tPos := targetPos.(*PositionComponent)
  208: 
  209: 		// Calculate direction
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:227`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  225: 			return NodeFailure
  226: 		}
> 227: 		vel := velComp.(*VelocityComponent)
  228: 		vel.VX = dx
  229: 		vel.VY = dy
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:248`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  246: 			return NodeFailure
  247: 		}
> 248: 		attack := attackComp.(*AttackComponent)
  249: 
  250: 		// Check attack cooldown
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:278`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  276: 			return NodeFailure
  277: 		}
> 278: 		pos := posComp.(*PositionComponent)
  279: 
  280: 		targetPos, ok := target.GetComponent("position")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:284`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  282: 			return NodeFailure
  283: 		}
> 284: 		tPos := targetPos.(*PositionComponent)
  285: 
  286: 		// Calculate direction (away from target)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:307`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  305: 			return NodeFailure
  306: 		}
> 307: 		vel := velComp.(*VelocityComponent)
  308: 		vel.VX = dx
  309: 		vel.VY = dy
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:342`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  340: 			velComp, ok := entity.GetComponent("velocity")
  341: 			if ok {
> 342: 				vel := velComp.(*VelocityComponent)
  343: 				vel.VX = 0
  344: 				vel.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_actions.go:357`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  355: 			return NodeFailure
  356: 		}
> 357: 		vel := velComp.(*VelocityComponent)
  358: 		vel.VX = dx
  359: 		vel.VY = dy
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/behavior_tree_system.go:39`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  37: 		}
  38: 
> 39: 		btComp := treeComp.(*BehaviorTreeComponent)
  40: 		if !btComp.Enabled {
  41: 			continue
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:91`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  89: 		}
  90: 
> 91: 		camera := cameraComp.(*CameraComponent)
  92: 
  93: 		// Get entity position
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:98`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  96: 			continue
  97: 		}
> 98: 		pos := posComp.(*PositionComponent)
  99: 
  100: 		// Calculate target camera position (entity position + offset)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:163`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  161: 		}
  162: 
> 163: 		hitStop := hitStopComp.(*HitStopComponent)
  164: 		if hitStop.IsActive() {
  165: 			// Update hit-stop elapsed time with REAL delta time (not scaled)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:190`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  188: 	}
  189: 
> 190: 	shake := shakeComp.(*ScreenShakeComponent)
  191: 	if !shake.IsShaking() {
  192: 		return
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:228`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  226: 		return worldX, worldY
  227: 	}
> 228: 	camera := cameraComp.(*CameraComponent)
  229: 
  230: 	// Apply camera transform
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:255`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  253: 		return screenX, screenY
  254: 	}
> 255: 	camera := cameraComp.(*CameraComponent)
  256: 
  257: 	// Remove screen centering
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:307`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  305: 		return 0, 0
  306: 	}
> 307: 	camera := cameraComp.(*CameraComponent)
  308: 
  309: 	return camera.X, camera.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:331`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  329: 		return
  330: 	}
> 331: 	camera := cameraComp.(*CameraComponent)
  332: 
  333: 	// Add to existing shake (allows stacking)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:360`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  358: 	shakeComp, ok := s.activeCamera.GetComponent("screenShake")
  359: 	if ok {
> 360: 		advanced := shakeComp.(*ScreenShakeComponent)
  361: 		advanced.TriggerShake(intensity, duration)
  362: 		return
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:389`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  387: 	}
  388: 
> 389: 	hitStop := hitStopComp.(*HitStopComponent)
  390: 	hitStop.TriggerHitStop(duration, timeScale)
  391: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:405`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  403: 	}
  404: 
> 405: 	hitStop := hitStopComp.(*HitStopComponent)
  406: 	return hitStop.IsActive()
  407: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/camera_system.go:422`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  420: 	}
  421: 
> 422: 	hitStop := hitStopComp.(*HitStopComponent)
  423: 	return hitStop.GetTimeScale()
  424: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:79`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  77: 			continue
  78: 		}
> 79: 		playerPos := playerPosComp.(*PositionComponent)
  80: 
  81: 		// Get object entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:94`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  92: 			continue
  93: 		}
> 94: 		objPos := objPosComp.(*PositionComponent)
  95: 
  96: 		// Update object position to follow player (slightly offset above/in front)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:128`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  126: 		return false
  127: 	}
> 128: 	carriable := carrComp.(*CarriableComponent)
  129: 
  130: 	// Check if object can be picked up
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:141`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  139: 	// Remove velocity if object was moving
  140: 	if velComp, ok := object.GetComponent("velocity"); ok {
> 141: 		vel := velComp.(*VelocityComponent)
  142: 		vel.VX = 0
  143: 		vel.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:197`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  195: 		return
  196: 	}
> 197: 	carriable := carrComp.(*CarriableComponent)
  198: 
  199: 	// Calculate throw velocity based on weight
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:216`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  214: 	// Set velocity
  215: 	if velComp, ok := object.GetComponent("velocity"); ok {
> 216: 		vel := velComp.(*VelocityComponent)
  217: 		vel.VX = aimX * throwVel
  218: 		vel.VY = aimY * throwVel
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:258`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  256: 		return
  257: 	}
> 258: 	carriable := carrComp.(*CarriableComponent)
  259: 
  260: 	// Mark as not carried
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:297`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  295: 			continue
  296: 		}
> 297: 		carriable := carrComp.(*CarriableComponent)
  298: 
  299: 		// Skip if already carried or not pickupable
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/carry_system.go:309`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  307: 			continue
  308: 		}
> 309: 		pos := posComp.(*PositionComponent)
  310: 
  311: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_creation.go:1213`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1211: 	}
  1212: 
> 1213: 	health := healthComp.(*HealthComponent)
  1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_creation.go:1214`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1212: 
  1213: 	health := healthComp.(*HealthComponent)
> 1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
  1216: 	attack := attackComp.(*AttackComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_creation.go:1215`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1213: 	health := healthComp.(*HealthComponent)
  1214: 	mana := manaComp.(*ManaComponent)
> 1215: 	stats := statsComp.(*StatsComponent)
  1216: 	attack := attackComp.(*AttackComponent)
  1217: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_creation.go:1216`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1214: 	mana := manaComp.(*ManaComponent)
  1215: 	stats := statsComp.(*StatsComponent)
> 1216: 	attack := attackComp.(*AttackComponent)
  1217: 
  1218: 	// Apply class-specific stats
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_ui.go:143`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  141: 	}
  142: 
> 143: 	stats := statsComp.(*StatsComponent)
  144: 	var equipment *EquipmentComponent
  145: 	if hasEquip {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_ui.go:146`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  144: 	var equipment *EquipmentComponent
  145: 	if hasEquip {
> 146: 		equipment = equipComp.(*EquipmentComponent)
  147: 	}
  148: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_ui.go:185`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  183: 	// Level and Gold info
  184: 	if hasExp {
> 185: 		exp := expComp.(*ExperienceComponent)
  186: 		levelText := fmt.Sprintf("Level %d", exp.Level)
  187: 		text.Draw(img, levelText, basicfont.Face7x13, panelX+20, titleY+13,
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/character_ui.go:192`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  190: 
  191: 	if hasInv {
> 192: 		inv := invComp.(*InventoryComponent)
  193: 		goldText := fmt.Sprintf("Gold: %d", inv.Gold)
  194: 		text.Draw(img, goldText, basicfont.Face7x13, panelX+panelWidth-120, titleY+13,
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:55`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  53: 
  54: 	colliderComp, _ := entity.GetComponent("collider")
> 55: 	collider := colliderComp.(*ColliderComponent)
  56: 
  57: 	// Only check solid colliders (triggers don't block movement)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:84`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  82: 	collider1Comp, _ := entity.GetComponent("collider")
  83: 	collider2Comp, _ := other.GetComponent("collider")
> 84: 	collider1 := collider1Comp.(*ColliderComponent)
  85: 	collider2 := collider2Comp.(*ColliderComponent)
  86: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:85`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  83: 	collider2Comp, _ := other.GetComponent("collider")
  84: 	collider1 := collider1Comp.(*ColliderComponent)
> 85: 	collider2 := collider2Comp.(*ColliderComponent)
  86: 
  87: 	// Skip trigger colliders (they don't block movement)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:106`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  104: 	layer2Comp, hasLayer2 := other.GetComponent("layer")
  105: 	if hasLayer1 && hasLayer2 {
> 106: 		l1 := layer1Comp.(*LayerComponent)
  107: 		l2 := layer2Comp.(*LayerComponent)
  108: 		// Flying entities collide with all layers
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:107`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  105: 	if hasLayer1 && hasLayer2 {
  106: 		l1 := layer1Comp.(*LayerComponent)
> 107: 		l2 := layer2Comp.(*LayerComponent)
  108: 		// Flying entities collide with all layers
  109: 		if !l1.CanFly && !l2.CanFly {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:119`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  117: 	// Get other entity's current position
  118: 	pos2Comp, _ := other.GetComponent("position")
> 119: 	pos2 := pos2Comp.(*PositionComponent)
  120: 
  121: 	// Issue #20: Check intersection at predicted position with rotation support
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:130`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  128: 		angle2 := 0.0
  129: 		if hasRot1 {
> 130: 			angle1 = rot1Comp.(*RotationComponent).Angle
  131: 		}
  132: 		if hasRot2 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:133`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  131: 		}
  132: 		if hasRot2 {
> 133: 			angle2 = rot2Comp.(*RotationComponent).Angle
  134: 		}
  135: 		return collider1.IntersectsRotated(newX, newY, angle1, collider2, pos2.X, pos2.Y, angle2)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:221`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  219: 			layer2Comp, hasLayer2 := other.GetComponent("layer")
  220: 			if hasLayer1 && hasLayer2 {
> 221: 				l1 := layer1Comp.(*LayerComponent)
  222: 				l2 := layer2Comp.(*LayerComponent)
  223: 				// Flying entities collide with all layers
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:222`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  220: 			if hasLayer1 && hasLayer2 {
  221: 				l1 := layer1Comp.(*LayerComponent)
> 222: 				l2 := layer2Comp.(*LayerComponent)
  223: 				// Flying entities collide with all layers
  224: 				if !l1.CanFly && !l2.CanFly {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:243`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  241: 				angle2 := 0.0
  242: 				if hasRot1 {
> 243: 					angle1 = rot1Comp.(*RotationComponent).Angle
  244: 				}
  245: 				if hasRot2 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:246`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  244: 				}
  245: 				if hasRot2 {
> 246: 					angle2 = rot2Comp.(*RotationComponent).Angle
  247: 				}
  248: 				intersects = collider.IntersectsRotated(pos.X, pos.Y, angle1, otherCollider, otherPos.X, otherPos.Y, angle2)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:395`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  393: 		if e1.HasComponent("velocity") {
  394: 			vel1, _ := e1.GetComponent("velocity")
> 395: 			vel1.(*VelocityComponent).VX = 0
  396: 		}
  397: 		if e2.HasComponent("velocity") {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:399`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  397: 		if e2.HasComponent("velocity") {
  398: 			vel2, _ := e2.GetComponent("velocity")
> 399: 			vel2.(*VelocityComponent).VX = 0
  400: 		}
  401: 	} else {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:414`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  412: 		if e1.HasComponent("velocity") {
  413: 			vel1, _ := e1.GetComponent("velocity")
> 414: 			vel1.(*VelocityComponent).VY = 0
  415: 		}
  416: 		if e2.HasComponent("velocity") {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:418`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  416: 		if e2.HasComponent("velocity") {
  417: 			vel2, _ := e2.GetComponent("velocity")
> 418: 			vel2.(*VelocityComponent).VY = 0
  419: 		}
  420: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:432`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  430: 	colliderComp, _ := entity.GetComponent("collider")
  431: 
> 432: 	pos := posComp.(*PositionComponent)
  433: 	collider := colliderComp.(*ColliderComponent)
  434: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:433`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  431: 
  432: 	pos := posComp.(*PositionComponent)
> 433: 	collider := colliderComp.(*ColliderComponent)
  434: 
  435: 	// Try to find a valid position by moving away from walls
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:457`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  455: 			if entity.HasComponent("velocity") {
  456: 				vel, _ := entity.GetComponent("velocity")
> 457: 				velocity := vel.(*VelocityComponent)
  458: 
  459: 				// Stop velocity component that's moving into the wall
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/collision.go:474`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  472: 	if entity.HasComponent("velocity") {
  473: 		vel, _ := entity.GetComponent("velocity")
> 474: 		velocity := vel.(*VelocityComponent)
  475: 		velocity.VX = 0
  476: 		velocity.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:100`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  98: 			// Update attack cooldowns only for living entities
  99: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 100: 				attack := attackComp.(*AttackComponent)
  101: 				beforeCooldown := attack.CooldownTimer
  102: 				attack.UpdateCooldown(deltaTime)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:118`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  116: 		// Process status effects (for both living and dead entities)
  117: 		if statusComp, ok := entity.GetComponent("status_effect"); ok {
> 118: 			status := statusComp.(*StatusEffectComponent)
  119: 
  120: 			// Update status effect
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:135`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  133: 	for _, entity := range entities {
  134: 		if healthComp, ok := entity.GetComponent("health"); ok {
> 135: 			health := healthComp.(*HealthComponent)
  136: 			if health.IsDead() {
  137: 				if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:158`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  156: 	}
  157: 
> 158: 	health := healthComp.(*HealthComponent)
  159: 
  160: 	switch effect.EffectType {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:188`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  186: 		return false
  187: 	}
> 188: 	attack := attackComp.(*AttackComponent)
  189: 
  190: 	// Check cooldown
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:198`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  196: 	// If so, spawn a projectile instead of doing instant damage
  197: 	if equipComp, hasEquip := attacker.GetComponent("equipment"); hasEquip {
> 198: 		equipment := equipComp.(*EquipmentComponent)
  199: 		if weapon, hasWeapon := equipment.Slots[SlotMainHand]; hasWeapon && weapon != nil {
  200: 			if weapon.Stats.IsProjectile {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:215`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  213: 		return false
  214: 	}
> 215: 	health := targetHealth.(*HealthComponent)
  216: 
  217: 	// Check if target is already dead
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:236`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  234: 	var attackerStats *StatsComponent
  235: 	if attackerStatsComp != nil {
> 236: 		attackerStats = attackerStatsComp.(*StatsComponent)
  237: 	}
  238: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:243`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  241: 	var targetStats *StatsComponent
  242: 	if targetStatsComp != nil {
> 243: 		targetStats = targetStatsComp.(*StatsComponent)
  244: 	}
  245: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:301`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  299: 	// Check for shield first
  300: 	if shieldComp, hasShield := target.GetComponent("shield"); hasShield {
> 301: 		shield := shieldComp.(*ShieldComponent)
  302: 		if shield.IsActive() {
  303: 			// Shield absorbs damage
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:320`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  318: 	// Trigger attack animation for attacker
  319: 	if animComp, hasAnim := attacker.GetComponent("animation"); hasAnim {
> 320: 		anim := animComp.(*AnimationComponent)
  321: 
  322: 		// Log animation trigger for player when debugging
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:336`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  334: 			// Check if entity is moving to set appropriate idle/walk state
  335: 			if velComp, hasVel := attacker.GetComponent("velocity"); hasVel {
> 336: 				vel := velComp.(*VelocityComponent)
  337: 				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  338: 				if speed > 0.1 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:355`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  353: 	// Trigger hurt animation for target
  354: 	if animComp, hasAnim := target.GetComponent("animation"); hasAnim {
> 355: 		anim := animComp.(*AnimationComponent)
  356: 		anim.SetState(AnimationStateHit)
  357: 		// Set a callback to return to idle after hurt animation
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:361`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  359: 			// Check if entity is moving to set appropriate idle/walk state
  360: 			if velComp, hasVel := target.GetComponent("velocity"); hasVel {
> 361: 				vel := velComp.(*VelocityComponent)
  362: 				speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  363: 				if speed > 0.1 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:390`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  388: 	if s.particleSystem != nil && s.world != nil {
  389: 		if posComp, ok := target.GetComponent("position"); ok {
> 390: 			pos := posComp.(*PositionComponent)
  391: 			// Use timestamp for particle seed variation
  392: 			particleSeed := s.seed + int64(pos.X*1000) + int64(pos.Y*1000)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:402`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  400: 		// Check accessibility settings for visual flash
  401: 		if s.camera != nil && s.camera.Accessibility.ShouldApplyVisualFlash() {
> 402: 			feedback := feedbackComp.(*VisualFeedbackComponent)
  403: 			// Flash intensity scales with damage (0.3-1.0 range)
  404: 			flashIntensity := 0.3 + (finalDamage / 100.0)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:419`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  417: 		var maxHP float64 = 100 // Default
  418: 		if targetHealthComp != nil {
> 419: 			maxHP = targetHealthComp.(*HealthComponent).Max
  420: 		}
  421: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:469`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  467: 		return false
  468: 	}
> 469: 	attack := attackComp.(*AttackComponent)
  470: 
  471: 	if !attack.CanAttack() {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:476`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  474: 
  475: 	targetHealth, ok := target.GetComponent("health")
> 476: 	if !ok || targetHealth.(*HealthComponent).IsDead() {
  477: 		return false
  478: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:509`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  507: 	}
  508: 
> 509: 	health := healthComp.(*HealthComponent)
  510: 	health.Heal(amount)
  511: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:533`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  531: 	var attackerTeamID int
  532: 	if attackerTeam != nil {
> 533: 		attackerTeamID = attackerTeam.(*TeamComponent).TeamID
  534: 	}
  535: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:551`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  549: 		targetTeam, hasTeam := entity.GetComponent("team")
  550: 		if hasTeam {
> 551: 			team := targetTeam.(*TeamComponent)
  552: 			if !team.IsEnemy(attackerTeamID) {
  553: 				continue
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:559`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  557: 		// Check health
  558: 		healthComp, hasHealth := entity.GetComponent("health")
> 559: 		if !hasHealth || healthComp.(*HealthComponent).IsDead() {
  560: 			continue
  561: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:617`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  615: 		return nil
  616: 	}
> 617: 	pos := attackerPos.(*PositionComponent)
  618: 
  619: 	// Filter enemies by aim cone and find closest
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:629`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  627: 			continue
  628: 		}
> 629: 		ePos := enemyPos.(*PositionComponent)
  630: 
  631: 		// Calculate angle from attacker to enemy
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:672`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  670: 		return false
  671: 	}
> 672: 	attackerPos := attackerPosComp.(*PositionComponent)
  673: 
  674: 	// Get aim direction
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:677`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  675: 	var aimAngle float64
  676: 	if aimComp, hasAim := attacker.GetComponent("aim"); hasAim {
> 677: 		aim := aimComp.(*AimComponent)
  678: 		aimAngle = aim.AimAngle
  679: 	} else if rotComp, hasRot := attacker.GetComponent("rotation"); hasRot {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:680`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  678: 		aimAngle = aim.AimAngle
  679: 	} else if rotComp, hasRot := attacker.GetComponent("rotation"); hasRot {
> 680: 		rot := rotComp.(*RotationComponent)
  681: 		aimAngle = rot.Angle
  682: 	} else {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:688`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  686: 			return false
  687: 		}
> 688: 		targetPos := targetPosComp.(*PositionComponent)
  689: 		dx := targetPos.X - attackerPos.X
  690: 		dy := targetPos.Y - attackerPos.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/combat_system.go:712`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  710: 	// Get attacker stats for bonus damage
  711: 	if attackerStatsComp, hasStats := attacker.GetComponent("stats"); hasStats {
> 712: 		attackerStats := attackerStatsComp.(*StatsComponent)
  713: 		if attack.DamageType == combat.DamageMagical {
  714: 			baseDamage += attackerStats.MagicPower
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_system.go:317`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  315: 	}
  316: 
> 317: 	items := result.([]*item.Item)
  318: 	if len(items) == 0 {
  319: 		return s.createFallbackItem(recipe)
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/engine/crafting_ui.go:149`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  147: 	ui.visible = !ui.visible
  148: 	if !ui.visible {
> 149: 		ui.Close()
  150: 	}
  151: }
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/engine/crafting_ui.go:171`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  169: 			ui.Toggle()
  170: 		} else {
> 171: 			ui.Close()
  172: 		}
  173: 		return // Don't process other input on the same frame as toggle/close
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:182`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  180: 	// Check if player is currently crafting
  181: 	if progressComp, ok := ui.playerEntity.GetComponent("crafting_progress"); ok {
> 182: 		progress := progressComp.(*CraftingProgressComponent)
  183: 		if progress != nil {
  184: 			ui.showingProgress = true
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:196`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  194: 		return
  195: 	}
> 196: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  197: 
  198: 	// Convert map to slice for ordered iteration
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:403`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  401: 	}
  402: 
> 403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 	skill := skillComp.(*CraftingSkillComponent)
  405: 	inv := invComp.(*InventoryComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:404`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  402: 
  403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
> 404: 	skill := skillComp.(*CraftingSkillComponent)
  405: 	inv := invComp.(*InventoryComponent)
  406: 	recipes := knowledge.KnownRecipes
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:405`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  403: 	knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 	skill := skillComp.(*CraftingSkillComponent)
> 405: 	inv := invComp.(*InventoryComponent)
  406: 	recipes := knowledge.KnownRecipes
  407: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:430`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  428: 	if ui.stationEntity != nil {
  429: 		if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
> 430: 			station := stationComp.(*CraftingStationComponent)
  431: 			titleText = fmt.Sprintf("CRAFTING - %s Station (+5%% success, 25%% faster)", station.StationType.String())
  432: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:465`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  463: 	if ui.showingProgress {
  464: 		progressComp, _ := ui.playerEntity.GetComponent("crafting_progress")
> 465: 		progress := progressComp.(*CraftingProgressComponent)
  466: 		if progress != nil {
  467: 			progressPercent := (progress.ElapsedTimeSec / progress.RequiredTimeSec) * 100
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:616`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  614: 		if ui.stationEntity != nil {
  615: 			if stationComp, ok := ui.stationEntity.GetComponent("crafting_station"); ok {
> 616: 				station := stationComp.(*CraftingStationComponent)
  617: 				// Check if station type matches recipe type
  618: 				if station.StationType == recipe.Type {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:692`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  690: 	if ui.stationEntity == nil && ui.playerEntity != nil {
  691: 		if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 692: 			pos := posComp.(*PositionComponent)
  693: 			// Find nearest station within 100 pixels
  694: 			nearestStation, distance := ui.findNearestStation(pos.X, pos.Y, 100)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:697`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  695: 			if nearestStation != nil {
  696: 				if stationComp, ok := nearestStation.GetComponent("crafting_station"); ok {
> 697: 					station := stationComp.(*CraftingStationComponent)
  698: 					stationHint := fmt.Sprintf("Nearby: %s (%.0f units away) - Move closer to use station bonuses",
  699: 						station.StationType.String(), distance)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/crafting_ui.go:821`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  819: 		return false
  820: 	}
> 821: 	inventory := invComp.(*InventoryComponent)
  822: 
  823: 	// Check if player has all required materials
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/destructible_object_system.go:96`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  94: 		return
  95: 	}
> 96: 	pos := posComp.(*PositionComponent)
  97: 
  98: 	if s.logger != nil {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/destructible_object_system.go:140`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  138: 			continue
  139: 		}
> 140: 		entityPos := posComp.(*PositionComponent)
  141: 
  142: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/destructible_object_system.go:286`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  284: 			continue
  285: 		}
> 286: 		pos := posComp.(*PositionComponent)
  287: 
  288: 		// Check if within damage radius
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/destructible_object_system.go:299`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  297: 				continue
  298: 			}
> 299: 			destructibleObj := objComp.(*DestructibleObjectComponent)
  300: 
  301: 			// Apply damage
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:41`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  39: 	switch c.Type() {
  40: 	case "position":
> 41: 		e.position = c.(*PositionComponent)
  42: 	case "velocity":
  43: 		e.velocity = c.(*VelocityComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:43`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  41: 		e.position = c.(*PositionComponent)
  42: 	case "velocity":
> 43: 		e.velocity = c.(*VelocityComponent)
  44: 	case "health":
  45: 		e.health = c.(*HealthComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:45`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  43: 		e.velocity = c.(*VelocityComponent)
  44: 	case "health":
> 45: 		e.health = c.(*HealthComponent)
  46: 	case "collider":
  47: 		e.collider = c.(*ColliderComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:47`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  45: 		e.health = c.(*HealthComponent)
  46: 	case "collider":
> 47: 		e.collider = c.(*ColliderComponent)
  48: 	case "inventory":
  49: 		e.inventory = c.(*InventoryComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:49`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  47: 		e.collider = c.(*ColliderComponent)
  48: 	case "inventory":
> 49: 		e.inventory = c.(*InventoryComponent)
  50: 	case "stats":
  51: 		e.stats = c.(*StatsComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/ecs.go:51`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  49: 		e.inventory = c.(*InventoryComponent)
  50: 	case "stats":
> 51: 		e.stats = c.(*StatsComponent)
  52: 	}
  53: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/entity_spawning.go:53`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  51: 	}
  52: 
> 53: 	generatedEntities := result.([]*entity.Entity)
  54: 	if len(generatedEntities) == 0 {
  55: 		return 0, nil
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/faction_system.go:94`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  92: 		// Note: In the ECS, we'll need to handle multiple faction components
  93: 		// For now, we assume one component per faction via a different approach
> 94: 		fc := comp.(FactionComponent)
  95: 		if fc.FactionID == change.FactionID && fc.IsPlayerFaction {
  96: 			factionComp = &fc
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/faction_system.go:175`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  173: 
  174: 	if comp, ok := playerEntity.GetComponent("faction"); ok {
> 175: 		fc := comp.(FactionComponent)
  176: 		if fc.FactionID == factionID && fc.IsPlayerFaction {
  177: 			return fc.Reputation
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/faction_system.go:207`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  205: 	// Get NPC's faction
  206: 	if comp, ok := entity.GetComponent("faction"); ok {
> 207: 		fc := comp.(FactionComponent)
  208: 		if !fc.IsPlayerFaction {
  209: 			// This is an NPC faction member
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/faction_system.go:232`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  230: 	}
  231: 
> 232: 	victimFaction := comp.(FactionComponent)
  233: 	if victimFaction.IsPlayerFaction {
  234: 		return // Don't process if victim is player
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:78`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  76: 			continue
  77: 		}
> 78: 		hazard := hazComp.(*HazardComponent)
  79: 
  80: 		// Update hazard duration
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:97`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  95: 			continue
  96: 		}
> 97: 		hazPos := hazPosComp.(*PositionComponent)
  98: 
  99: 		// Sync zone tracker with current hazard state
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:145`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  143: 			continue
  144: 		}
> 145: 		entPos := entPosComp.(*PositionComponent)
  146: 
  147: 		// Query zones at entity position
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:232`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  230: 			continue
  231: 		}
> 232: 		entPos := entPosComp.(*PositionComponent)
  233: 
  234: 		// Calculate distance to hazard
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:291`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  289: 		return
  290: 	}
> 291: 	entPos := entPosComp.(*PositionComponent)
  292: 
  293: 	zones := s.zoneTracker.GetZonesAt(entPos.X, entPos.Y)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:350`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  348: 			continue
  349: 		}
> 350: 		hazard := hazComp.(*HazardComponent)
  351: 
  352: 		// Get hazard position
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hazard_system.go:357`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  355: 			continue
  356: 		}
> 357: 		hazPos := hazPosComp.(*PositionComponent)
  358: 
  359: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/help_system.go:324`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  322: 				continue
  323: 			}
> 324: 			health := comp.(*HealthComponent)
  325: 			if health.Current < health.Max*0.25 && !hs.ShowQuickHint {
  326: 				hs.ShowQuickHintFor("low_health")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/help_system.go:336`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  334: 				continue
  335: 			}
> 336: 			inv := comp.(*InventoryComponent)
  337: 			if len(inv.Items) >= inv.MaxItems && !hs.ShowQuickHint {
  338: 				hs.ShowQuickHintFor("inventory_full")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:97`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  95: 		return
  96: 	}
> 97: 	health := healthComp.(*HealthComponent)
  98: 
  99: 	// Health bar dimensions
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:149`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  147: 	// Draw level if available
  148: 	if hasExp {
> 149: 		exp := expComp.(*ExperienceComponent)
  150: 		levelText := fmt.Sprintf("Level: %d", exp.Level)
  151: 		h.drawText(levelText, x, y, color.White)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:157`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  155: 	// Draw stats if available
  156: 	if hasStats {
> 157: 		stats := statsComp.(*StatsComponent)
  158: 		h.drawText(fmt.Sprintf("ATK: %.0f", stats.Attack), x, y, color.RGBA{255, 200, 200, 255})
  159: 		y += lineHeight
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:172`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  170: 		return
  171: 	}
> 172: 	exp := expComp.(*ExperienceComponent)
  173: 
  174: 	// Experience bar dimensions
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:243`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  241: 		return // No aim component, skip indicator
  242: 	}
> 243: 	aim := aimComp.(*AimComponent)
  244: 
  245: 	// DEBUG: Compare aim vs rotation components
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/hud_system.go:247`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  245: 	// DEBUG: Compare aim vs rotation components
  246: 	if rotComp, ok := h.playerEntity.GetComponent("rotation"); ok {
> 247: 		rotation := rotComp.(*RotationComponent)
  248: 		fmt.Printf("[DEBUG] HUD: AimAngle=%.4f, RotationAngle=%.4f, RotationTarget=%.4f\n",
  249: 			aim.AimAngle, rotation.Angle, rotation.TargetAngle)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/input_system.go:507`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  505: 		}
  506: 
> 507: 		input := inputComp.(*EbitenInput)
  508: 		s.processInput(entity, input, deltaTime)
  509: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/input_system.go:658`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  656: 		aimComp, ok := entity.GetComponent("aim")
  657: 		if ok {
> 658: 			aim := aimComp.(*AimComponent)
  659: 
  660: 			// Convert screen coordinates to world coordinates
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/input_system.go:676`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  674: 	// Apply movement to velocity component if it exists
  675: 	if velComp, ok := entity.GetComponent("velocity"); ok {
> 676: 		velocity := velComp.(*VelocityComponent)
  677: 		velocity.VX = input.MoveX * s.MoveSpeed
  678: 		velocity.VY = input.MoveY * s.MoveSpeed
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/input_system.go:682`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  680: 		// GAP-018 REPAIR: Update player animation based on movement
  681: 		if animComp, ok := entity.GetComponent("animation"); ok {
> 682: 			anim := animComp.(*AnimationComponent)
  683: 			// Check if player is moving
  684: 			isMoving := (velocity.VX != 0 || velocity.VY != 0)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:71`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  69: 		return
  70: 	}
> 71: 	playerPos := playerPosComp.(*PositionComponent)
  72: 
  73: 	// Priority order for interactions:
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:103`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  101: 			continue
  102: 		}
> 103: 		entPos := entPosComp.(*PositionComponent)
  104: 
  105: 		// Get context action component
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:110`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  108: 			continue
  109: 		}
> 110: 		contextAction := ctxComp.(*ContextActionComponent)
  111: 
  112: 		// Skip if not available or on cooldown
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:241`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  239: 			continue
  240: 		}
> 241: 		elemPos := elemPosComp.(*PositionComponent)
  242: 
  243: 		// Get puzzle element component
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:248`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  246: 			continue
  247: 		}
> 248: 		puzzleElem := elemComp.(*PuzzleElementComponent)
  249: 
  250: 		// Skip if not interactable or on cooldown
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/interaction_system.go:287`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  285: 		}
  286: 
> 287: 		puzzle := puzzleComp.(*PuzzleComponent)
  288: 		if puzzle.PuzzleID == puzzleElem.PuzzleID {
  289: 			// Record progress
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_system.go:256`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  254: 	var healthComp *HealthComponent
  255: 	if hasHealth {
> 256: 		healthComp, _ = comp.(*HealthComponent)
  257: 	}
  258: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_system.go:301`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  299: 		return
  300: 	}
> 301: 	equipComp, _ := comp.(*EquipmentComponent)
  302: 	if equipComp == nil {
  303: 		return
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_system.go:381`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  379: 		return fmt.Errorf("entity %d has no position component, cannot drop item", entityID)
  380: 	}
> 381: 	pos := posComp.(*PositionComponent)
  382: 
  383: 	// Remove item from inventory (only after we know we can drop it)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_system.go:572`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  570: 		}
  571: 
> 572: 		equipment := equipComp.(*EquipmentComponent)
  573: 
  574: 		// Recalculate equipment stats if dirty
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_ui.go:113`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  111: 		return
  112: 	}
> 113: 	inventory := invComp.(*InventoryComponent)
  114: 
  115: 	// Calculate inventory window position
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_ui.go:231`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  229: 		return
  230: 	}
> 231: 	inventory := invComp.(*InventoryComponent)
  232: 
  233: 	// Draw semi-transparent overlay
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/inventory_ui.go:349`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  347: 		// Show equipped item if present
  348: 		if hasEquipment {
> 349: 			equipment := equipComp.(*EquipmentComponent)
  350: 			equipped := equipment.GetEquipped(slotInfo.slot)
  351: 			if equipped != nil {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:78`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  76: 	// Increase drop chance for bosses/elites
  77: 	if statsComp, ok := enemy.GetComponent("stats"); ok {
> 78: 		stats := statsComp.(*StatsComponent)
  79: 		if stats.Attack > 20 || stats.Defense > 20 {
  80: 			dropChance = 0.7 // 70% for strong enemies
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:93`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  91: 	depth := 1
  92: 	if expComp, ok := enemy.GetComponent("experience"); ok {
> 93: 		exp := expComp.(*ExperienceComponent)
  94: 		depth = exp.Level
  95: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:113`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  111: 	}
  112: 
> 113: 	items := result.([]*item.Item)
  114: 	if len(items) == 0 {
  115: 		return nil
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:133`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  131: 	// Increase drop chance for bosses/elites
  132: 	if statsComp, ok := enemy.GetComponent("stats"); ok {
> 133: 		stats := statsComp.(*StatsComponent)
  134: 		if stats.Attack > 20 || stats.Defense > 20 {
  135: 			dropChance = 0.2 // 20% for strong enemies
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:150`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  148: 
  149: 	if expComp, ok := enemy.GetComponent("experience"); ok {
> 150: 		exp := expComp.(*ExperienceComponent)
  151: 		depth = exp.Level
  152: 		difficulty = 0.3 + float64(depth)*0.05 // Scale with depth
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:170`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  168: 	}
  169: 
> 170: 	recipes := result.([]*Recipe)
  171: 	if len(recipes) == 0 {
  172: 		return nil
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:321`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  319: 		}
  320: 
> 321: 		inventory := playerInventory.(*InventoryComponent)
  322: 
  323: 		for _, itemEntity := range items {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:334`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  332: 			}
  333: 
> 334: 			itemData := itemEntityComp.(*ItemEntityComponent)
  335: 
  336: 			// Check distance for pickup (32 pixels = 1 tile)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:390`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  388: 			}
  389: 
> 390: 			recipeData := recipeEntityComp.(*RecipeEntityComponent)
  391: 
  392: 			// Check distance for pickup (32 pixels = 1 tile)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/item_spawning.go:403`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  401: 				}
  402: 
> 403: 				knowledge := knowledgeComp.(*RecipeKnowledgeComponent)
  404: 
  405: 				// Check if player already knows this recipe
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/lifetime_system.go:50`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  48: 		}
  49: 
> 50: 		lifetime := lifetimeComp.(*LifetimeComponent)
  51: 		lifetime.Elapsed += deltaTime
  52: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:299`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  297: 	// Draw player icon
  298: 	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 299: 		pos := posComp.(*PositionComponent)
  300: 		// Convert world position to tile coordinates (assuming 32px tiles)
  301: 		tileX := int(pos.X / 32)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:427`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  425: 	}
  426: 
> 427: 	pos := posComp.(*PositionComponent)
  428: 	// Convert world position to tile coordinates (assuming 32px tiles)
  429: 	centerX := int(pos.X / 32)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:512`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  510: 	// Draw player icon
  511: 	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
> 512: 		pos := posComp.(*PositionComponent)
  513: 		tileX := int(pos.X / 32)
  514: 		tileY := int(pos.Y / 32)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:537`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  535: 		}
  536: 
> 537: 		pos := posComp.(*PositionComponent)
  538: 		tileX := int(pos.X / 32)
  539: 		tileY := int(pos.Y / 32)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:557`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  555: 		// Check if enemy
  556: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 557: 			team := teamComp.(*TeamComponent)
  558: 			if team.TeamID == 2 { // Enemy team
  559: 				iconColor = color.RGBA{255, 100, 100, 255} // Red
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/map_ui.go:636`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  634: 	}
  635: 
> 636: 	pos := posComp.(*PositionComponent)
  637: 	tileX := int(pos.X / 32)
  638: 	tileY := int(pos.Y / 32)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:107`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  105: 	} else {
  106: 		if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
> 107: 			menuComp := menu.(*MenuComponent)
  108: 			menuComp.Active = !menuComp.Active
  109: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:126`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  124: 	}
  125: 	if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
> 126: 		return menu.(*MenuComponent).Active
  127: 	}
  128: 	return false
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:138`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  136: 
  137: 	menu, ok := ms.menuEntity.GetComponent("menu")
> 138: 	if !ok || !menu.(*MenuComponent).Active {
  139: 		return
  140: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:142`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  140: 	}
  141: 
> 142: 	menuComp := menu.(*MenuComponent)
  143: 
  144: 	// Update error message timeout
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:481`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  479: 
  480: 	menu, ok := ms.menuEntity.GetComponent("menu")
> 481: 	if !ok || !menu.(*MenuComponent).Active {
  482: 		return
  483: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/menu_system.go:485`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  483: 	}
  484: 
> 485: 	menuComp := menu.(*MenuComponent)
  486: 
  487: 	// Draw semi-transparent overlay
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/merchant_spawn.go:236`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  234: 			continue
  235: 		}
> 236: 		pos := posComp.(*PositionComponent)
  237: 
  238: 		// Calculate distance squared (avoid sqrt for performance)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/merchant_spawn.go:267`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  265: 			continue
  266: 		}
> 267: 		pos := posComp.(*PositionComponent)
  268: 
  269: 		dx := pos.X - x
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/merchant_spawn.go:308`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  306: 	}
  307: 
> 308: 	merchantData := merchComp.(*MerchantComponent)
  309: 	return fmt.Sprintf("Press S to talk to %s", merchantData.MerchantName)
  310: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:45`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  43: 		}
  44: 
> 45: 		mount := mountComp.(*MountComponent)
  46: 
  47: 		// Find the vehicle entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:60`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  58: 			continue
  59: 		}
> 60: 		vehiclePos := vehiclePosComp.(*PositionComponent)
  61: 
  62: 		// Update rider position to match vehicle + offset
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:65`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  63: 		riderPosComp, hasRiderPos := entity.GetComponent("position")
  64: 		if hasRiderPos {
> 65: 			riderPos := riderPosComp.(*PositionComponent)
  66: 			riderPos.X = vehiclePos.X + mount.MountOffset.X
  67: 			riderPos.Y = vehiclePos.Y + mount.MountOffset.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:74`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  72: 		riderRotComp, hasRiderRot := entity.GetComponent("rotation")
  73: 		if hasVehicleRot && hasRiderRot {
> 74: 			vehicleRot := vehicleRotComp.(*RotationComponent)
  75: 			riderRot := riderRotComp.(*RotationComponent)
  76: 			riderRot.Angle = vehicleRot.Angle
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:75`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  73: 		if hasVehicleRot && hasRiderRot {
  74: 			vehicleRot := vehicleRotComp.(*RotationComponent)
> 75: 			riderRot := riderRotComp.(*RotationComponent)
  76: 			riderRot.Angle = vehicleRot.Angle
  77: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:102`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  100: 	}
  101: 
> 102: 	vehicleData := vehicleComp.(*VehicleComponent)
  103: 
  104: 	// Check if vehicle has capacity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:121`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  119: 	}
  120: 
> 121: 	riderPos := riderPosComp.(*PositionComponent)
  122: 	vehiclePos := vehiclePosComp.(*PositionComponent)
  123: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:122`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  120: 
  121: 	riderPos := riderPosComp.(*PositionComponent)
> 122: 	vehiclePos := vehiclePosComp.(*PositionComponent)
  123: 
  124: 	// Calculate offset (preserve relative position)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:158`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  156: 	}
  157: 
> 158: 	mount := mountComp.(*MountComponent)
  159: 
  160: 	// Find vehicle and update passenger count
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:165`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  163: 		if exists && vehicle != nil {
  164: 			if vehicleComp, hasVehicle := vehicle.GetComponent("vehicle"); hasVehicle {
> 165: 				vehicleData := vehicleComp.(*VehicleComponent)
  166: 				vehicleData.RemovePassenger()
  167: 			}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/mounting_system.go:207`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  205: 	}
  206: 
> 207: 	mount := mountComp.(*MountComponent)
  208: 
  209: 	if ms.world != nil {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:66`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  64: 		}
  65: 
> 66: 		pos := posComp.(*PositionComponent)
  67: 		vel := velComp.(*VelocityComponent)
  68: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:67`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  65: 
  66: 		pos := posComp.(*PositionComponent)
> 67: 		vel := velComp.(*VelocityComponent)
  68: 
  69: 		// Apply speed limit if configured
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:87`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  85: 		if s.collisionSystem != nil && entity.HasComponent("collider") {
  86: 			colliderComp, _ := entity.GetComponent("collider")
> 87: 			collider := colliderComp.(*ColliderComponent)
  88: 
  89: 			// Only check solid, non-trigger colliders
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:164`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  162: 		// Apply bounds if entity has them
  163: 		if boundsComp, hasBounds := entity.GetComponent("bounds"); hasBounds {
> 164: 			bounds := boundsComp.(*BoundsComponent)
  165: 			pos.X, pos.Y = bounds.Clamp(pos.X, pos.Y)
  166: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:180`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  178: 		// Priority 1.4: Apply friction/drag to slow down entities
  179: 		if frictionComp, hasFriction := entity.GetComponent("friction"); hasFriction {
> 180: 			friction := frictionComp.(*FrictionComponent)
  181: 
  182: 			// Apply friction as exponential decay: v *= (1 - coefficient)^deltaTime
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:197`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  195: 		// Update animation state based on movement
  196: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 197: 			anim := animComp.(*AnimationComponent)
  198: 			speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  199: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:290`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  288: func SetVelocity(entity *Entity, vx, vy float64) {
  289: 	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 290: 		vel := velComp.(*VelocityComponent)
  291: 		vel.VX = vx
  292: 		vel.VY = vy
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:299`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  297: func GetPosition(entity *Entity) (x, y float64, ok bool) {
  298: 	if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 299: 		pos := posComp.(*PositionComponent)
  300: 		return pos.X, pos.Y, true
  301: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:308`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  306: func SetPosition(entity *Entity, x, y float64) {
  307: 	if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 308: 		pos := posComp.(*PositionComponent)
  309: 		pos.X = x
  310: 		pos.Y = y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/movement.go:382`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  380: 		return
  381: 	}
> 382: 	collider := colliderComp.(*ColliderComponent)
  383: 
  384: 	// Calculate tile coordinates from entity position using helper method
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:109`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  107: 		return MusicContextExploration
  108: 	}
> 109: 	playerPos := playerPosComp.(*PositionComponent)
  110: 
  111: 	// Check player health for danger state
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:113`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  111: 	// Check player health for danger state
  112: 	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
> 113: 		health := healthComp.(*HealthComponent)
  114: 		healthPercent := float64(health.Current) / float64(health.Max)
  115: 		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:145`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  143: 		playerTeam := 1 // Default player team
  144: 		if teamComp, hasTeam := playerEntity.GetComponent("team"); hasTeam {
> 145: 			playerTeam = teamComp.(*TeamComponent).TeamID
  146: 		}
  147: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:150`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  148: 		entityTeam := 0 // Default enemy team
  149: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 150: 			entityTeam = teamComp.(*TeamComponent).TeamID
  151: 		}
  152: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:162`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  160: 		// Check proximity to player
  161: 		if posComp, hasPos := entity.GetComponent("position"); hasPos {
> 162: 			entityPos := posComp.(*PositionComponent)
  163: 			distance := d.calculateDistance(playerPos.X, playerPos.Y, entityPos.X, entityPos.Y)
  164: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:170`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  168: 				// Check if it's a boss (high attack stat)
  169: 				if statsComp, hasStats := entity.GetComponent("stats"); hasStats {
> 170: 					stats := statsComp.(*StatsComponent)
  171: 					if stats.Attack >= d.BossAttackThreshold {
  172: 						hasBoss = true
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/music_context.go:192`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  190: 	// Check danger state last (lowest priority of active contexts)
  191: 	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
> 192: 		health := healthComp.(*HealthComponent)
  193: 		healthPercent := float64(health.Current) / float64(health.Max)
  194: 		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/narrative_system.go:49`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  47: 		}
  48: 
> 49: 		narrative := narComp.(*NarrativeComponent)
  50: 
  51: 		// Check for triggered events
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/narrative_system.go:128`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  126: 		// Check for boss or elite status
  127: 		if aiComp, ok := enemyEntity.GetComponent("ai"); ok {
> 128: 			ai := aiComp.(*AIComponent)
  129: 			// Boss entities typically have high detection range
  130: 			if ai.DetectionRange > BossDetectionRange {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:71`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  69: 		return
  70: 	}
> 71: 	tracker := comp.(*QuestTrackerComponent)
  72: 
  73: 	// For now, all enemies count as "enemy" or "monster"
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:102`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  100: 		return
  101: 	}
> 102: 	tracker := comp.(*QuestTrackerComponent)
  103: 
  104: 	// Update collect objectives
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:135`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  133: 		return
  134: 	}
> 135: 	tracker := comp.(*QuestTrackerComponent)
  136: 
  137: 	// Update UI interaction objectives (used in tutorial quests)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:169`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  167: 		return
  168: 	}
> 169: 	tracker := comp.(*QuestTrackerComponent)
  170: 
  171: 	// Update explore objectives
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:194`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  192: 		return
  193: 	}
> 194: 	pos := posComp.(*PositionComponent)
  195: 
  196: 	// Convert world coordinates to tile coordinates (assuming 32-pixel tiles)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:209`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  207: 		return
  208: 	}
> 209: 	tracker := comp.(*QuestTrackerComponent)
  210: 
  211: 	// Check each active quest
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/objective_tracker_system.go:328`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  326: 		return // No inventory to receive items
  327: 	}
> 328: 	inv := invComp.(*InventoryComponent)
  329: 
  330: 	// Get genre for item generation
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/particle_system.go:35`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  33: 		}
  34: 
> 35: 		emitter := comp.(*ParticleEmitterComponent)
  36: 
  37: 		// Update elapsed time for time-limited emitters
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/particle_system.go:71`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  69: 				// Position particles at entity's position
  70: 				if posComp, ok := entity.GetComponent("position"); ok {
> 71: 					pos := posComp.(*PositionComponent)
  72: 					ps.offsetParticles(system, pos.X, pos.Y)
  73: 				}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_combat_system.go:73`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  71: 			continue // Entity can't attack
  72: 		}
> 73: 		attack := attackComp.(*AttackComponent)
  74: 
  75: 		// Check if attack is ready (cooldown)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_combat_system.go:99`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  97: 		// This provides visual feedback that the attack button was pressed
  98: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 99: 			anim := animComp.(*AnimationComponent)
  100: 			anim.SetState(AnimationStateAttack)
  101: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_combat_system.go:105`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  103: 			anim.OnComplete = func() {
  104: 				if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
> 105: 					vel := velComp.(*VelocityComponent)
  106: 					speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
  107: 					if speed > 0.1 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_combat_system.go:123`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  121: 		// Check if entity has aim component (Phase 10.1)
  122: 		if aimComp, hasAim := entity.GetComponent("aim"); hasAim {
> 123: 			aim := aimComp.(*AimComponent)
  124: 			// Use aim direction for target selection with default aim cone (forgiving aim)
  125: 			target = FindEnemyInAimDirection(s.world, entity, aim.AimAngle, maxRange, DefaultAimCone)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_item_use_system.go:72`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  70: 			continue // Entity has no inventory
  71: 		}
> 72: 		inventory := invComp.(*InventoryComponent)
  73: 
  74: 		// Get hotbar component for selected item (if available)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_item_use_system.go:77`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  75: 		var selectedIndex int
  76: 		if hotbarComp, hasHotbar := entity.GetComponent("hotbar"); hasHotbar {
> 77: 			hotbar := hotbarComp.(*HotbarComponent)
  78: 			selectedIndex = hotbar.LastUsedIndex
  79: 			// Check if the slot has an item
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_item_use_system.go:169`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  167: 		return
  168: 	}
> 169: 	hotbar := hotbarComp.(*HotbarComponent)
  170: 	hotbar.LastUsedIndex = slotIndex
  171: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/player_spell_casting.go:57`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  55: 	// Get spell slots
  56: 	slotsComp, _ := player.GetComponent("spell_slots")
> 57: 	slots := slotsComp.(*SpellSlotComponent)
  58: 
  59: 	// If currently casting, don't start new cast
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:110`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  108: 	}
  109: 
> 110: 	exp := expComp.(*ExperienceComponent)
  111: 
  112: 	if ps.logger != nil && ps.logger.Logger.GetLevel() >= logrus.DebugLevel {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:172`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  170: 		return // No scaling defined
  171: 	}
> 172: 	scaling := scalingComp.(*LevelScalingComponent)
  173: 
  174: 	// Update health component
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:177`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  175: 	healthComp, ok := entity.GetComponent("health")
  176: 	if ok {
> 177: 		health := healthComp.(*HealthComponent)
  178: 		oldMax := health.Max
  179: 		health.Max = scaling.CalculateHealthForLevel(level)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:187`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  185: 	statsComp, ok := entity.GetComponent("stats")
  186: 	if ok {
> 187: 		stats := statsComp.(*StatsComponent)
  188: 		stats.Attack = scaling.CalculateAttackForLevel(level)
  189: 		stats.Defense = scaling.CalculateDefenseForLevel(level)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:205`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  203: 	}
  204: 
> 205: 	exp := expComp.(*ExperienceComponent)
  206: 	level := exp.Level
  207: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:230`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  228: 	}
  229: 
> 230: 	return expComp.(*ExperienceComponent).Level
  231: }
  232: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:244`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  242: 	}
  243: 
> 244: 	return expComp.(*ExperienceComponent).ProgressToNextLevel()
  245: }
  246: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:259`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  257: 	}
  258: 
> 259: 	exp := expComp.(*ExperienceComponent)
  260: 	if exp.SkillPoints <= 0 {
  261: 		return fmt.Errorf("entity has no skill points to spend")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:279`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  277: 	}
  278: 
> 279: 	return expComp.(*ExperienceComponent).SkillPoints
  280: }
  281: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/progression_system.go:299`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  297: 		entity.AddComponent(exp)
  298: 	} else {
> 299: 		exp = expComp.(*ExperienceComponent)
  300: 	}
  301: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_pool.go:43`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  41: // Returns a zeroed component ready for initialization.
  42: func (p *ProjectilePool) Get() *ProjectileComponent {
> 43: 	proj := p.pool.Get().(*ProjectileComponent)
  44: 	// Reset all fields to zero values
  45: 	proj.Damage = 0.0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_pool.go:87`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  85: // Get acquires a velocity component from the pool.
  86: func (p *VelocityPool) Get() *VelocityComponent {
> 87: 	vel := p.pool.Get().(*VelocityComponent)
  88: 	vel.VX = 0.0
  89: 	vel.VY = 0.0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_pool.go:120`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  118: // Get acquires a position component from the pool.
  119: func (p *PositionPool) Get() *PositionComponent {
> 120: 	pos := p.pool.Get().(*PositionComponent)
  121: 	pos.X = 0.0
  122: 	pos.Y = 0.0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system.go:472`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  470: 	case "fireball", "magic_missile", "ice_shard", "lightning_bolt":
  471: 		// Magical projectiles get glowing trails
> 472: 		magicColor := spriteComp.Color.(color.RGBA)
  473: 		trail = NewMagicTrailComponent(&magicColor)
  474: 	case "arrow", "bullet", "bolt":
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_debug.go:37`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  35: 		if _, ok := e.GetComponent("projectile"); ok {
  36: 			if pos, ok := e.GetComponent("position"); ok {
> 37: 				p := pos.(*PositionComponent)
  38: 				fmt.Printf("Projectile start: (%.2f, %.2f)\n", p.X, p.Y)
  39: 			}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_debug.go:54`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  52: 			projectileExists = true
  53: 			if pos, ok := e.GetComponent("position"); ok {
> 54: 				p := pos.(*PositionComponent)
  55: 				fmt.Printf("Projectile after update: (%.2f, %.2f)\n", p.X, p.Y)
  56: 			}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_debug.go:65`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  63: 	// Check health
  64: 	healthComp1, _ := target1.GetComponent("health")
> 65: 	health1 := healthComp1.(*HealthComponent)
  66: 	fmt.Printf("Target 1 health: %.2f\n", health1.Current)
  67: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_debug.go:69`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  67: 
  68: 	healthComp2, _ := target2.GetComponent("health")
> 69: 	health2 := healthComp2.(*HealthComponent)
  70: 	fmt.Printf("Target 2 health: %.2f\n", health2.Current)
  71: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_debug.go:73`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  71: 
  72: 	healthComp3, _ := target3.GetComponent("health")
> 73: 	health3 := healthComp3.(*HealthComponent)
  74: 	fmt.Printf("Target 3 health: %.2f\n", health3.Current)
  75: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_minimal.go:31`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  29: 	for _, e := range entities {
  30: 		if pos, ok := e.GetComponent("position"); ok {
> 31: 			p := pos.(*PositionComponent)
  32: 			fmt.Printf("  Entity ID=%d at (%.1f, %.1f)\n", e.ID, p.X, p.Y)
  33: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_minimal.go:43`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  41: 	// Check projectile position
  42: 	if projPos, ok := proj.GetComponent("position"); ok {
> 43: 		p := projPos.(*PositionComponent)
  44: 		fmt.Printf("Projectile now at (%.1f, %.1f)\n", p.X, p.Y)
  45: 	} else {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/projectile_system_test_minimal.go:51`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  49: 	// Check target health
  50: 	if healthComp, ok := target.GetComponent("health"); ok {
> 51: 		health := healthComp.(*HealthComponent)
  52: 		fmt.Printf("Target health: %.1f\n", health.Current)
  53: 		if health.Current == 100.0 {
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/engine/quality_system.go:91`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  88: 		return err
  89: 	}
  90: 
> 91: 	qs.mu.Lock()
  92: 	qs.config = config
  93: 	qs.mu.Unlock()
  94: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/engine/quality_system.go:105`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  102: func (qs *QualitySystem) SetQualityLevel(level quality.QualityLevel) {
  103: 	qs.adjuster.SetManualQuality(level)
  104: 
> 105: 	qs.mu.Lock()
  106: 	qs.config.ApplyLevel(level)
  107: 	qs.mu.Unlock()
  108: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/quest_ui.go:102`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  100: 		trackerComp, ok := ui.playerEntity.GetComponent("questtracker")
  101: 		if ok {
> 102: 			tracker := trackerComp.(*QuestTrackerComponent)
  103: 			var quests []*TrackedQuest
  104: 			if ui.currentTab == 0 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/quest_ui.go:186`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  184: 		return
  185: 	}
> 186: 	tracker := trackerComp.(*QuestTrackerComponent)
  187: 
  188: 	// Draw semi-transparent overlay
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:336`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  334: 			continue
  335: 		}
> 336: 		sprite := spriteComp.(*EbitenSprite)
  337: 
  338: 		if !sprite.Visible {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:376`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  374: 		return
  375: 	}
> 376: 	batchSpriteImage := firstSprite.(*EbitenSprite).Image
  377: 	if batchSpriteImage == nil {
  378: 		// No sprite image, draw entities individually
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:409`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  407: 		}
  408: 
> 409: 		pos := posComp.(*PositionComponent)
  410: 		sprite := spriteComp.(*EbitenSprite)
  411: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:410`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  408: 
  409: 		pos := posComp.(*PositionComponent)
> 410: 		sprite := spriteComp.(*EbitenSprite)
  411: 
  412: 		// DEBUG: Log sprite state for player
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:423`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  421: 		// Phase 4: Sync CurrentDirection from AnimationComponent.Facing
  422: 		if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 423: 			anim := animComp.(*AnimationComponent)
  424: 			sprite.CurrentDirection = int(anim.GetFacing())
  425: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:431`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  429: 		// CRITICAL: Must sync here for batch rendering path (drawEntity has its own sync)
  430: 		if rotComp, hasRot := entity.GetComponent("rotation"); hasRot {
> 431: 			rotation := rotComp.(*RotationComponent)
  432: 			sprite.Rotation = rotation.Angle
  433: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:474`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  472: 		var tintR, tintG, tintB, tintA float64 = 1.0, 1.0, 1.0, 1.0
  473: 		if feedbackComp, ok := entity.GetComponent("visual_feedback"); ok {
> 474: 			feedback := feedbackComp.(*VisualFeedbackComponent)
  475: 			flashAlpha = feedback.GetFlashAlpha()
  476: 			tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:611`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  609: 		return entities
  610: 	}
> 611: 	camera := camComp.(*CameraComponent)
  612: 
  613: 	// Calculate viewport bounds in world space with margin for sprites
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:657`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  655: 	}
  656: 
> 657: 	pos := posComp.(*PositionComponent)
  658: 	sprite := spriteComp.(*EbitenSprite)
  659: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:658`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  656: 
  657: 	pos := posComp.(*PositionComponent)
> 658: 	sprite := spriteComp.(*EbitenSprite)
  659: 
  660: 	if !sprite.Visible {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:669`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  667: 	// Phase 4: Sync CurrentDirection from AnimationComponent.Facing
  668: 	if animComp, hasAnim := entity.GetComponent("animation"); hasAnim {
> 669: 		anim := animComp.(*AnimationComponent)
  670: 		sprite.CurrentDirection = int(anim.GetFacing())
  671: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:676`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  674: 	// This enables 360° visual rotation for entities with rotation component
  675: 	if rotComp, hasRot := entity.GetComponent("rotation"); hasRot {
> 676: 		rotation := rotComp.(*RotationComponent)
  677: 		sprite.Rotation = rotation.Angle
  678: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:700`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  698: 	var layerTransitionAlpha float64 = 1.0
  699: 	if layerComp, hasLayer := entity.GetComponent("layer"); hasLayer {
> 700: 		layer := layerComp.(*LayerComponent)
  701: 		if layer.IsTransitioning() {
  702: 			// Calculate depth offset based on transition progress
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:732`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  730: 	var tintR, tintG, tintB, tintA float64 = 1.0, 1.0, 1.0, 1.0
  731: 	if feedbackComp, ok := entity.GetComponent("visual_feedback"); ok {
> 732: 		feedback := feedbackComp.(*VisualFeedbackComponent)
  733: 		flashAlpha = feedback.GetFlashAlpha()
  734: 		tintR, tintG, tintB, tintA = feedback.TintR, feedback.TintG, feedback.TintB, feedback.TintA
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:832`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  830: 	}
  831: 
> 832: 	health := healthComp.(*HealthComponent)
  833: 
  834: 	// Don't draw health bar for player (has HUD display)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:842`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  840: 	isBoss := false
  841: 	if attackComp, ok := entity.GetComponent("attack"); ok {
> 842: 		attack := attackComp.(*AttackComponent)
  843: 		isBoss = attack.Damage > 20 // Boss threshold
  844: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:908`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  906: 		}
  907: 
> 908: 		emitter := comp.(*ParticleEmitterComponent)
  909: 
  910: 		// Render each particle system
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:989`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  987: 		}
  988: 
> 989: 		pos := posComp.(*PositionComponent)
  990: 		collider := colliderComp.(*ColliderComponent)
  991: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:990`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  988: 
  989: 		pos := posComp.(*PositionComponent)
> 990: 		collider := colliderComp.(*ColliderComponent)
  991: 
  992: 		// Get collider bounds
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:1025`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1023: 	for _, entity := range entities {
  1024: 		if sprite, ok := entity.GetComponent("sprite"); ok {
> 1025: 			ebitenSprite := sprite.(*EbitenSprite)
  1026: 			cache = append(cache, entitySprite{
  1027: 				entity: entity,
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:1047`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1045: 		posJ, okJ := cache[j].entity.GetComponent("position")
  1046: 		if okI && okJ {
> 1047: 			yI := posI.(*PositionComponent).Y
  1048: 			yJ := posJ.(*PositionComponent).Y
  1049: 			if yI != yJ {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/render_system.go:1048`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1046: 		if okI && okJ {
  1047: 			yI := posI.(*PositionComponent).Y
> 1048: 			yJ := posJ.(*PositionComponent).Y
  1049: 			if yI != yJ {
  1050: 				return yI < yJ
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:48`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  46: 		if entity.HasComponent("input") && !entity.HasComponent("dead") {
  47: 			if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
> 48: 				health := healthComp.(*HealthComponent)
  49: 				if health.IsAlive() {
  50: 					livingPlayers = append(livingPlayers, entity)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:73`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  71: 		// Check for revival input (E key or interact button)
  72: 		inputComp, _ := livingPlayer.GetComponent("input")
> 73: 		input := inputComp.(*EbitenInput)
  74: 
  75: 		// Check if revival action key is pressed (E key = UseItemPressed)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:86`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  84: 			continue
  85: 		}
> 86: 		livingPos := livingPosComp.(*PositionComponent)
  87: 
  88: 		// Find closest dead player within range
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:98`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  96: 				continue
  97: 			}
> 98: 			deadPos := deadPosComp.(*PositionComponent)
  99: 
  100: 			// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:126`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  124: 		return
  125: 	}
> 126: 	health := healthComp.(*HealthComponent)
  127: 
  128: 	// Restore health (20% of max by default)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:168`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  166: 		return nil
  167: 	}
> 168: 	livingPos := livingPosComp.(*PositionComponent)
  169: 
  170: 	var revivablePlayers []*Entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/revival_system.go:183`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  181: 			continue
  182: 		}
> 183: 		deadPos := deadPosComp.(*PositionComponent)
  184: 
  185: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:42`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  40: 			continue
  41: 		}
> 42: 		rotation := rotComp.(*RotationComponent)
  43: 
  44: 		// Sync rotation target with aim component if present
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:48`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  46: 			aimComp, ok := entity.GetComponent("aim")
  47: 			if ok {
> 48: 				aim := aimComp.(*AimComponent)
  49: 
  50: 				// Update aim angle from position if target-based
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:54`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  52: 					posComp, ok := entity.GetComponent("position")
  53: 					if ok {
> 54: 						pos := posComp.(*PositionComponent)
  55: 						aim.UpdateAimAngle(pos.X, pos.Y)
  56: 					}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:96`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  94: 
  95: 	rotComp, _ := entity.GetComponent("rotation")
> 96: 	rotation := rotComp.(*RotationComponent)
  97: 
  98: 	aimComp, _ := entity.GetComponent("aim")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:99`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  97: 
  98: 	aimComp, _ := entity.GetComponent("aim")
> 99: 	aim := aimComp.(*AimComponent)
  100: 
  101: 	rotation.SetAngleImmediate(aim.AimAngle)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:120`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  118: 
  119: 	rotComp, _ := entity.GetComponent("rotation")
> 120: 	rotation := rotComp.(*RotationComponent)
  121: 
  122: 	rotation.SetAngleImmediate(angle)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:140`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  138: 
  139: 	rotComp, _ := entity.GetComponent("rotation")
> 140: 	rotation := rotComp.(*RotationComponent)
  141: 
  142: 	return rotation.Angle, true
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:161`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  159: 
  160: 	rotComp, _ := entity.GetComponent("rotation")
> 161: 	rotation := rotComp.(*RotationComponent)
  162: 
  163: 	rotation.SmoothRotation = enabled
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/rotation_system.go:182`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  180: 
  181: 	rotComp, _ := entity.GetComponent("rotation")
> 182: 	rotation := rotComp.(*RotationComponent)
  183: 
  184: 	rotation.RotationSpeed = speed
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/engine/shop_ui.go:140`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  138: 	ui.visible = !ui.visible
  139: 	if !ui.visible {
> 140: 		ui.Close()
  141: 	}
  142: }
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/engine/shop_ui.go:174`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  172: 			ui.Toggle()
  173: 		} else {
> 174: 			ui.Close()
  175: 		}
  176: 		// Also end dialog if dialog system is set
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/shop_ui.go:203`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  201: 		// Show merchant inventory
  202: 		if merchantComp, ok := ui.merchantEntity.GetComponent("merchant"); ok {
> 203: 			merchant := merchantComp.(*MerchantComponent)
  204: 			currentInventory = merchant.Inventory
  205: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/shop_ui.go:209`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  207: 		// Show player inventory
  208: 		if invComp, ok := ui.playerEntity.GetComponent("inventory"); ok {
> 209: 			inv := invComp.(*InventoryComponent)
  210: 			currentInventory = inv.Items
  211: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/shop_ui.go:349`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  347: 	}
  348: 
> 349: 	playerInv := playerInvComp.(*InventoryComponent)
  350: 	merchant := merchantComp.(*MerchantComponent)
  351: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/shop_ui.go:350`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  348: 
  349: 	playerInv := playerInvComp.(*InventoryComponent)
> 350: 	merchant := merchantComp.(*MerchantComponent)
  351: 
  352: 	// Draw semi-transparent overlay
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:54`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  52: 		return
  53: 	}
> 54: 	treeComp := comp.(*SkillTreeComponent)
  55: 
  56: 	// Get stats component
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:61`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  59: 		return // No stats to modify
  60: 	}
> 61: 	stats := statsComp.(*StatsComponent)
  62: 
  63: 	// Reset bonus stats (we'll recalculate from scratch)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:188`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  186: 	baseStatsComp, hasBaseStats := entity.GetComponent("base_stats")
  187: 	if hasBaseStats {
> 188: 		baseStats := baseStatsComp.(*BaseStatsComponent)
  189: 
  190: 		// Apply attack bonus
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:193`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  191: 		if bonuses.DamageBonus != 0 {
  192: 			if attackComp, ok := entity.GetComponent("attack"); ok {
> 193: 				attack := attackComp.(*AttackComponent)
  194: 				attack.Damage = baseStats.BaseAttack * (1.0 + bonuses.DamageBonus)
  195: 			}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:211`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  209: 		if bonuses.HealthBonus != 0 {
  210: 			if healthComp, ok := entity.GetComponent("health"); ok {
> 211: 				health := healthComp.(*HealthComponent)
  212: 				oldMax := health.Max
  213: 				health.Max = baseStats.BaseMaxHealth * (1.0 + bonuses.HealthBonus)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_progression_system.go:224`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  222: 		if bonuses.ManaRegenBonus != 0 {
  223: 			if manaComp, ok := entity.GetComponent("mana"); ok {
> 224: 				mana := manaComp.(*ManaComponent)
  225: 				mana.Regen = baseStats.BaseManaRegen * (1.0 + bonuses.ManaRegenBonus)
  226: 			}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_tree_loader.go:42`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  40: 	}
  41: 
> 42: 	trees := result.([]*skills.SkillTree)
  43: 	if len(trees) == 0 {
  44: 		return nil // No trees generated, not an error
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skill_tree_loader.go:59`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  57: 		comp, ok := player.GetComponent("skill_tree")
  58: 		if ok {
> 59: 			treeComp := comp.(*SkillTreeComponent)
  60: 			treeComp.Tree = mainTree
  61: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skills_ui.go:129`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  127: 
  128: 	if comp, ok := ui.playerEntity.GetComponent("skill_tree"); ok {
> 129: 		ui.skillTreeComp = comp.(*SkillTreeComponent)
  130: 	}
  131: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skills_ui.go:506`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  504: 		// Deduct skill points from experience component
  505: 		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 506: 			exp := expComp.(*ExperienceComponent)
  507: 			skill := ui.skillTreeComp.Tree.GetSkillByID(skillID)
  508: 			if skill != nil {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skills_ui.go:533`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  531: 		// Refund skill points to experience component
  532: 		if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 533: 			exp := expComp.(*ExperienceComponent)
  534: 			exp.SkillPoints += pointsRefunded
  535: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skills_ui.go:609`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  607: 
  608: 	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 609: 		return expComp.(*ExperienceComponent).SkillPoints
  610: 	}
  611: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/skills_ui.go:622`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  620: 
  621: 	if expComp, ok := ui.playerEntity.GetComponent("experience"); ok {
> 622: 		return expComp.(*ExperienceComponent).Level
  623: 	}
  624: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spatial_partition.go:64`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  62: 		return false
  63: 	}
> 64: 	pos := posComp.(*PositionComponent)
  65: 
  66: 	// Check if point is in bounds
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spatial_partition.go:135`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  133: 			continue
  134: 		}
> 135: 		pos := posComp.(*PositionComponent)
  136: 
  137: 		if queryBounds.Contains(pos.X, pos.Y) {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spatial_partition.go:172`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  170: 			continue
  171: 		}
> 172: 		pos := posComp.(*PositionComponent)
  173: 
  174: 		dx := pos.X - x
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:122`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  120: 			continue
  121: 		}
> 122: 		slots := spellComp.(*SpellSlotComponent)
  123: 
  124: 		// Update cooldowns
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:173`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  171: 		return
  172: 	}
> 173: 	mana := manaComp.(*ManaComponent)
  174: 
  175: 	if mana.Current < spell.Stats.ManaCost {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:194`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  192: 		return
  193: 	}
> 194: 	pos := posComp.(*PositionComponent)
  195: 
  196: 	// Apply spell effects based on type
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:244`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  242: 			continue
  243: 		}
> 244: 		health := healthComp.(*HealthComponent)
  245: 
  246: 		health.Current -= float64(spell.Stats.Damage)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:260`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  258: 			targetPos, hasPos := target.GetComponent("position")
  259: 			if hasPos {
> 260: 				pos := targetPos.(*PositionComponent)
  261: 				// Spawn element-specific particles
  262: 				s.spawnElementalHitEffect(pos.X, pos.Y, spell.Element, target.ID)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:300`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  298: 		return
  299: 	}
> 300: 	health := healthComp.(*HealthComponent)
  301: 
  302: 	health.Current += float64(spell.Stats.Healing)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:311`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  309: 		targetPos, hasPos := target.GetComponent("position")
  310: 		if hasPos {
> 311: 			pos := targetPos.(*PositionComponent)
  312: 			config := particles.Config{
  313: 				Type:     particles.ParticleMagic,
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:344`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  342: 	var casterTeamID int
  343: 	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
> 344: 		casterTeamID = teamComp.(*TeamComponent).TeamID
  345: 	}
  346: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:354`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  352: 		// Check if ally
  353: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 354: 			team := teamComp.(*TeamComponent)
  355: 			if !team.IsAlly(casterTeamID) {
  356: 				continue
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:368`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  366: 			continue
  367: 		}
> 368: 		health := healthComp.(*HealthComponent)
  369: 		if health.Current >= health.Max {
  370: 			continue // At full health
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:392`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  390: 	var casterTeamID int
  391: 	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
> 392: 		casterTeamID = teamComp.(*TeamComponent).TeamID
  393: 	}
  394: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:398`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  396: 		// Check if ally (including self)
  397: 		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
> 398: 			team := teamComp.(*TeamComponent)
  399: 			if !team.IsAlly(casterTeamID) {
  400: 				continue
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:476`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  474: 			healthComp, hasHealth := target.GetComponent("health")
  475: 			if hasHealth {
> 476: 				health := healthComp.(*HealthComponent)
  477: 				health.Current -= float64(spell.Stats.Damage)
  478: 				if health.Current < 0 {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:574`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  572: 		return
  573: 	}
> 574: 	pos := posComp.(*PositionComponent)
  575: 
  576: 	// Calculate teleport direction and distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:587`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  585: 	dirX, dirY := 0.0, 1.0 // Default: down
  586: 	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
> 587: 		vel := velComp.(*VelocityComponent)
  588: 		if vel.VX != 0 || vel.VY != 0 {
  589: 			// Normalize velocity to get direction
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:641`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  639: 		return
  640: 	}
> 641: 	pos := posComp.(*PositionComponent)
  642: 
  643: 	// Determine reveal radius from spell stats
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:708`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  706: 		posComp, hasPos := caster.GetComponent("position")
  707: 		if hasPos {
> 708: 			pos := posComp.(*PositionComponent)
  709: 			config := particles.Config{
  710: 				Type:     particles.ParticleDust,
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:748`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  746: 			continue
  747: 		}
> 748: 		collider := colliderComp.(*ColliderComponent)
  749: 
  750: 		// Skip non-solid colliders
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:760`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  758: 			continue
  759: 		}
> 760: 		pos := entityPos.(*PositionComponent)
  761: 
  762: 		// Get caster collider for size checking
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:765`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  763: 		casterCollider, hasCasterCollider := caster.GetComponent("collider")
  764: 		if hasCasterCollider {
> 765: 			cc := casterCollider.(*ColliderComponent)
  766: 			// Create temporary collider at target position
  767: 			tempCollider := &ColliderComponent{
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:876`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  874: 			break
  875: 		}
> 876: 		casterPosComp := casterPos.(*PositionComponent)
  877: 
  878: 		// Get caster's facing direction (use velocity or mouse aim)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:901`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  899: 				continue
  900: 			}
> 901: 			entityPosComp := entityPos.(*PositionComponent)
  902: 
  903: 			// Vector from caster to entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:933`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  931: 			break
  932: 		}
> 933: 		casterPosComp := casterPos.(*PositionComponent)
  934: 
  935: 		// Get caster's facing direction (use velocity or mouse aim)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:958`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  956: 				continue
  957: 			}
> 958: 			entityPosComp := entityPos.(*PositionComponent)
  959: 
  960: 			// Vector from caster to entity
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:997`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  995: 	// Try to use velocity for moving entities
  996: 	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
> 997: 		vel := velComp.(*VelocityComponent)
  998: 		if vel.VX != 0 || vel.VY != 0 {
  999: 			return vel.VX, vel.VY
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1005`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1003: 	// Fall back to direction towards target point
  1004: 	if posComp, hasPos := caster.GetComponent("position"); hasPos {
> 1005: 		pos := posComp.(*PositionComponent)
  1006: 		dirX = targetX - pos.X
  1007: 		dirY = targetY - pos.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1021`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1019: 		return false
  1020: 	}
> 1021: 	slots := spellComp.(*SpellSlotComponent)
  1022: 
  1023: 	// Check if already casting
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1044`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1042: 		return false
  1043: 	}
> 1044: 	mana := manaComp.(*ManaComponent)
  1045: 	if mana.Current < spell.Stats.ManaCost {
  1046: 		return false
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1062`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1060: 		return
  1061: 	}
> 1062: 	slots := spellComp.(*SpellSlotComponent)
  1063: 
  1064: 	slots.Casting = -1
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1218`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1216: 			continue
  1217: 		}
> 1218: 		mana := manaComp.(*ManaComponent)
  1219: 
  1220: 		// Regenerate mana
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1246`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1244: 	}
  1245: 
> 1246: 	spells := result.([]*magic.Spell)
  1247: 
  1248: 	// Create spell slots component if doesn't exist
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/spell_casting.go:1257`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  1255: 	} else {
  1256: 		slotsComp, _ := player.GetComponent("spell_slots")
> 1257: 		slots = slotsComp.(*SpellSlotComponent)
  1258: 	}
  1259: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:24`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  22: 			return NodeFailure
  23: 		}
> 24: 		pos := posComp.(*PositionComponent)
  25: 
  26: 		// Calculate distance to formation position
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:35`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  33: 			// Stop movement
  34: 			if velComp, ok := entity.GetComponent("velocity"); ok {
> 35: 				vel := velComp.(*VelocityComponent)
  36: 				vel.VX = 0
  37: 				vel.VY = 0
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:44`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  42: 		// Move towards formation position
  43: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 44: 			vel := velComp.(*VelocityComponent)
  45: 			speed := 80.0 // Formation movement speed
  46: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:70`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  68: 			return NodeFailure // Not in a squad
  69: 		}
> 70: 		squad := squadComp.(*SquadComponent)
  71: 
  72: 		// Get positions
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:77`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  75: 			return NodeFailure
  76: 		}
> 77: 		pos := posComp.(*PositionComponent)
  78: 
  79: 		targetPosComp, ok := target.GetComponent("position")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:83`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  81: 			return NodeFailure
  82: 		}
> 83: 		targetPos := targetPosComp.(*PositionComponent)
  84: 
  85: 		// Determine flanking position based on squad position index
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:111`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  109: 
  110: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 111: 			vel := velComp.(*VelocityComponent)
  112: 			speed := 90.0
  113: 			vel.VX = (dx / distance) * speed
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:129`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  127: 			return NodeFailure
  128: 		}
> 129: 		squad := squadComp.(*SquadComponent)
  130: 
  131: 		// Check if squad has a shared priority target
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:151`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  149: 			return NodeFailure
  150: 		}
> 151: 		squad := squadComp.(*SquadComponent)
  152: 
  153: 		// Check if we can alert (cooldown)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:164`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  162: 			return NodeFailure
  163: 		}
> 164: 		pos := posComp.(*PositionComponent)
  165: 
  166: 		// Find nearby squads within alert range (500 pixels)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:176`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  174: 
  175: 			otherSquadComp, _ := other.GetComponent("squad")
> 176: 			otherSquad := otherSquadComp.(*SquadComponent)
  177: 
  178: 			// Skip same squad
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:184`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  182: 
  183: 			otherPosComp, _ := other.GetComponent("position")
> 184: 			otherPos := otherPosComp.(*PositionComponent)
  185: 
  186: 			// Check distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:215`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  213: 			return NodeFailure
  214: 		}
> 215: 		squad := squadComp.(*SquadComponent)
  216: 
  217: 		// Check if squad leader ordered retreat
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:233`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  231: 		// Move in retreat direction
  232: 		if velComp, ok := entity.GetComponent("velocity"); ok {
> 233: 			vel := velComp.(*VelocityComponent)
  234: 			speed := 120.0 // Retreat faster
  235: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:253`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  251: 		// Get squad component
  252: 		squadComp, ok := entity.GetComponent("squad")
> 253: 		if !ok || !squadComp.(*SquadComponent).IsLeader() {
  254: 			return NodeFailure
  255: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:256`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  254: 			return NodeFailure
  255: 		}
> 256: 		squad := squadComp.(*SquadComponent)
  257: 
  258: 		// Leader makes tactical decisions for the squad
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:262`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  260: 		healthComp, ok := entity.GetComponent("health")
  261: 		if ok {
> 262: 			health := healthComp.(*HealthComponent)
  263: 			healthPercent := float64(health.Current) / float64(health.Max)
  264: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:272`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  270: 				if target, ok := blackboard.GetEntity("target"); ok {
  271: 					posComp, _ := entity.GetComponent("position")
> 272: 					pos := posComp.(*PositionComponent)
  273: 					targetPosComp, _ := target.GetComponent("position")
  274: 					targetPos := targetPosComp.(*PositionComponent)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:274`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  272: 					pos := posComp.(*PositionComponent)
  273: 					targetPosComp, _ := target.GetComponent("position")
> 274: 					targetPos := targetPosComp.(*PositionComponent)
  275: 
  276: 					// Retreat direction is away from target
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:306`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  304: 			return false
  305: 		}
> 306: 		return squadComp.(*SquadComponent).IsLeader()
  307: 	})
  308: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:317`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  315: 			return false
  316: 		}
> 317: 		squad := squadComp.(*SquadComponent)
  318: 
  319: 		target, ok := squad.SharedBlackboard.GetEntity("priority_target")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_behaviors.go:331`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  329: 			return false
  330: 		}
> 331: 		squad := squadComp.(*SquadComponent)
  332: 
  333: 		retreat, ok := squad.SharedBlackboard.GetBool("retreat_ordered")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:45`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  43: 			continue
  44: 		}
> 45: 		squad := squadComp.(*SquadComponent)
  46: 		squads[squad.SquadID] = append(squads[squad.SquadID], entity)
  47: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:62`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  60: 	for _, member := range members {
  61: 		squadComp, _ := member.GetComponent("squad")
> 62: 		squad := squadComp.(*SquadComponent)
  63: 		if squad.IsLeader() {
  64: 			leader = member
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:89`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  87: 		return
  88: 	}
> 89: 	leaderPos := leaderPosComp.(*PositionComponent)
  90: 
  91: 	// Get leader's rotation for oriented formations
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:94`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  92: 	leaderAngle := 0.0
  93: 	if rotComp, ok := leader.GetComponent("rotation"); ok {
> 94: 		rot := rotComp.(*RotationComponent)
  95: 		leaderAngle = rot.Angle
  96: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:106`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  104: 
  105: 		squadComp, _ := member.GetComponent("squad")
> 106: 		squad := squadComp.(*SquadComponent)
  107: 
  108: 		// Calculate target position based on formation type
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:120`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  118: 		// Store target position in member's blackboard
  119: 		if btComp, ok := member.GetComponent("behaviortree"); ok {
> 120: 			bt := btComp.(*BehaviorTreeComponent)
  121: 			bt.Blackboard.Set("formation_target_x", targetX)
  122: 			bt.Blackboard.Set("formation_target_y", targetY)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:189`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  187: 	for _, member := range members {
  188: 		squadComp, _ := member.GetComponent("squad")
> 189: 		squad := squadComp.(*SquadComponent)
  190: 		if squad.SharedBlackboard != nil {
  191: 			sharedBlackboard = squad.SharedBlackboard
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/squad_system.go:203`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  201: 	for _, member := range members {
  202: 		squadComp, _ := member.GetComponent("squad")
> 203: 		squad := squadComp.(*SquadComponent)
  204: 		squad.SharedBlackboard = sharedBlackboard
  205: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/station_spawn.go:281`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  279: 		}
  280: 
> 281: 		pos := posComp.(*PositionComponent)
  282: 
  283: 		// Calculate squared distance (avoid sqrt for performance)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/station_spawn.go:327`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  325: 		}
  326: 
> 327: 		pos := posComp.(*PositionComponent)
  328: 
  329: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/station_spawn.go:361`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  359: 	}
  360: 
> 361: 	station := stationComp.(*CraftingStationComponent)
  362: 
  363: 	recipeTypeName := ""
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_pool.go:35`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  33: //	ReleaseStatusEffect(effect)
  34: func NewStatusEffectComponent(effectType string, magnitude, duration, tickInterval float64) *StatusEffectComponent {
> 35: 	effect := statusEffectPool.Get().(*StatusEffectComponent)
  36: 
  37: 	// Initialize with provided values
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:54`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  52: 		// Update shield duration
  53: 		if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
> 54: 			shield := shieldComp.(*ShieldComponent)
  55: 			shield.Update(deltaTime)
  56: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:71`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  69: 		return
  70: 	}
> 71: 	health := healthComp.(*HealthComponent)
  72: 
  73: 	switch effect.EffectType {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:94`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  92: 		return
  93: 	}
> 94: 	stats := statsComp.(*StatsComponent)
  95: 
  96: 	switch effect.EffectType {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:145`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  143: 		return
  144: 	}
> 145: 	stats := statsComp.(*StatsComponent)
  146: 
  147: 	switch effect.EffectType {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:171`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  169: 	if shieldComp, hasShield := entity.GetComponent("shield"); hasShield {
  170: 		// Add to existing shield
> 171: 		shield := shieldComp.(*ShieldComponent)
  172: 		shield.Amount += amount
  173: 		if shield.Amount > shield.MaxAmount {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:200`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  198: 	// Apply damage to initial target
  199: 	if healthComp, hasHealth := initialTarget.GetComponent("health"); hasHealth {
> 200: 		health := healthComp.(*HealthComponent)
  201: 		health.TakeDamage(damage)
  202: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:255`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  253: 	if casterTeam, hasCasterTeam := caster.GetComponent("team"); hasCasterTeam {
  254: 		if targetTeam, hasTargetTeam := target.GetComponent("team"); hasTargetTeam {
> 255: 			ct := casterTeam.(*TeamComponent)
  256: 			tt := targetTeam.(*TeamComponent)
  257: 			return ct.IsEnemy(tt.TeamID)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/status_effect_system.go:256`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  254: 		if targetTeam, hasTargetTeam := target.GetComponent("team"); hasTargetTeam {
  255: 			ct := casterTeam.(*TeamComponent)
> 256: 			tt := targetTeam.(*TeamComponent)
  257: 			return ct.IsEnemy(tt.TeamID)
  258: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/terrain_collision_system.go:165`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  163: 	colliderComp, _ := entity.GetComponent("collider")
  164: 
> 165: 	pos := posComp.(*PositionComponent)
  166: 	collider := colliderComp.(*ColliderComponent)
  167: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/terrain_collision_system.go:166`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  164: 
  165: 	pos := posComp.(*PositionComponent)
> 166: 	collider := colliderComp.(*ColliderComponent)
  167: 
  168: 	// Get entity's layer (default to ground layer if no layer component)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/terrain_collision_system.go:172`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  170: 	if entity.HasComponent("layer") {
  171: 		layerComp, _ := entity.GetComponent("layer")
> 172: 		layerComponent := layerComp.(*LayerComponent)
  173: 		layer = layerComponent.GetEffectiveLayer()
  174: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tile_cache.go:129`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  127: 
  128: 	c.lruList.Remove(elem)
> 129: 	keyStr := elem.Value.(string)
  130: 	delete(c.cache, keyStr)
  131: }
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/engine/tile_cache.go:68`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  65: 		c.mu.RUnlock()
  66: 
  67: 		// Move to front of LRU list (write lock)
> 68: 		c.mu.Lock()
  69: 		c.lruList.MoveToFront(entry.elem)
  70: 		c.hits++
  71: 		c.mu.Unlock()
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tutorial_system.go:86`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  84: 							continue
  85: 						}
> 86: 						pos := comp.(*PositionComponent)
  87: 						// Simple distance check from origin (400, 300 typical spawn)
  88: 						distFromStart := (pos.X-400)*(pos.X-400) + (pos.Y-300)*(pos.Y-300)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tutorial_system.go:109`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  107: 							continue
  108: 						}
> 109: 						attack := comp.(*AttackComponent)
  110: 						// Check if attack cooldown is active (means they attacked)
  111: 						return attack.CooldownTimer > 0 || attack.CooldownTimer < attack.Cooldown
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tutorial_system.go:131`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  129: 							continue
  130: 						}
> 131: 						health := comp.(*HealthComponent)
  132: 						// Complete if health is damaged but still above 50%
  133: 						return health.Current < health.Max && health.Current > health.Max/2
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tutorial_system.go:153`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  151: 							continue
  152: 						}
> 153: 						inv := comp.(*InventoryComponent)
  154: 						return len(inv.Items) > 0
  155: 					}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/tutorial_system.go:174`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  172: 							continue
  173: 						}
> 174: 						exp := comp.(*ExperienceComponent)
  175: 						return exp.Level >= 2
  176: 					}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:61`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  59: 			continue
  60: 		}
> 61: 		combat := combatComp.(*VehicleCombatComponent)
  62: 
  63: 		// Update cooldowns
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:81`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  79: 		return
  80: 	}
> 81: 	v := vehicleComp.(*VehicleComponent)
  82: 
  83: 	// Check if can ram (cooldown ready and sufficient speed)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:92`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  90: 		return
  91: 	}
> 92: 	pos := posComp.(*PositionComponent)
  93: 
  94: 	// Check for entities in ramming range (small radius around vehicle)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:112`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  110: 			continue
  111: 		}
> 112: 		tPos := targetPos.(*PositionComponent)
  113: 
  114: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:125`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  123: 
  124: 			// Apply damage to target
> 125: 			health := targetHealth.(*HealthComponent)
  126: 			health.TakeDamage(damage)
  127: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:162`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  160: 		return
  161: 	}
> 162: 	pos := posComp.(*PositionComponent)
  163: 
  164: 	// Check if vehicle is player-controlled or has AI targeting
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:180`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  178: 		// Calculate direction to target
  179: 		targetPos, _ := target.GetComponent("position")
> 180: 		tPos := targetPos.(*PositionComponent)
  181: 		dx := tPos.X - pos.X
  182: 		dy := tPos.Y - pos.Y
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:193`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  191: 		targetHealth, hasHealth := target.GetComponent("health")
  192: 		if hasHealth {
> 193: 			health := targetHealth.(*HealthComponent)
  194: 			health.TakeDamage(combat.WeaponDamage)
  195: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:222`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  220: 	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
  221: 	if hasVehicle {
> 222: 		v := vehicleComp.(*VehicleComponent)
  223: 		if v.CurrentPassengers > 0 {
  224: 			return true
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_combat_system.go:256`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  254: 			continue
  255: 		}
> 256: 		tPos := targetPos.(*PositionComponent)
  257: 
  258: 		// Calculate distance
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_durability_system.go:43`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  41: 		}
  42: 
> 43: 		vehicle := vehicleComp.(*VehicleComponent)
  44: 
  45: 		// Skip if already destroyed
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_durability_system.go:67`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  65: 	}
  66: 
> 67: 	collider := collComp.(*ColliderComponent)
  68: 
  69: 	// In a full implementation, this would check if the collider
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_durability_system.go:114`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  112: 	}
  113: 
> 114: 	vehicle := vehicleComp.(*VehicleComponent)
  115: 	destroyed := vehicle.TakeDamage(damage)
  116: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_durability_system.go:131`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  129: 	}
  130: 
> 131: 	vehicle := vehicleComp.(*VehicleComponent)
  132: 	vehicle.Repair(amount)
  133: 	return true
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_movement_system.go:47`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  45: 		}
  46: 
> 47: 		vehicle := vehicleComp.(*VehicleComponent)
  48: 
  49: 		// Skip if vehicle is destroyed or out of fuel
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_movement_system.go:60`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  58: 			continue // Can't move without position
  59: 		}
> 60: 		pos := posComp.(*PositionComponent)
  61: 
  62: 		rotComp, hasRot := entity.GetComponent("rotation")
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_movement_system.go:66`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  64: 			continue // Need rotation for direction
  65: 		}
> 66: 		rot := rotComp.(*RotationComponent)
  67: 
  68: 		// Check if entity is being controlled (has input or is mounted)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/vehicle_movement_system.go:112`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  110: 	// Check if any entity is mounted on this vehicle
  111: 	vehicleComp, _ := entity.GetComponent("vehicle")
> 112: 	vehicle := vehicleComp.(*VehicleComponent)
  113: 	return vehicle.CurrentPassengers > 0
  114: }
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/visual_feedback_components.go:92`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  90: 		}
  91: 
> 92: 		feedback := feedbackComp.(*VisualFeedbackComponent)
  93: 
  94: 		// Update flash timer
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/engine/weather_system.go:160`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  158: 			// Apply transition opacity to particle color
  159: 			// Convert color.Color interface to color.RGBA for alpha manipulation
> 160: 			rgba := color.RGBAModel.Convert(p.Color).(color.RGBA)
  161: 			// Adjust alpha based on opacity
  162: 			originalAlpha := float64(rgba.A)
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/hostplay/host_and_play.go:72`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  70: 		if err == nil {
  71: 			// Port is available, close the test listener
> 72: 			listener.Close()
  73: 			return port, bindAddr, nil
  74: 		}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/hostplay/input_handler.go:91`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  89: 
  90: 	// Extract direction from input data
> 91: 	dx, dxOk := data["dx"].(float64)
  92: 	dy, dyOk := data["dy"].(float64)
  93: 	if !dxOk || !dyOk {
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/hostplay/input_handler.go:92`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  90: 	// Extract direction from input data
  91: 	dx, dxOk := data["dx"].(float64)
> 92: 	dy, dyOk := data["dy"].(float64)
  93: 	if !dxOk || !dyOk {
  94: 		return
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/hostplay/server_manager.go:142`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  140: 		return fmt.Errorf("failed to generate terrain: %w", err)
  141: 	}
> 142: 	sm.generatedTerrain = terrainResult.(*terrain.Terrain)
  143: 
  144: 	sm.logger.WithFields(logrus.Fields{
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/hostplay/server_manager.go:318`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  316: 		netComp, exists := entity.GetComponent("network")
  317: 		if exists && netComp != nil {
> 318: 			nc := netComp.(*engine.NetworkComponent)
  319: 			if nc.PlayerID == playerID {
  320: 				sm.world.RemoveEntity(entity.ID)
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/hostplay/server_manager.go:335`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  332: 
  333: // Stop gracefully stops the server and waits for the goroutine to exit.
  334: func (sm *ServerManager) Stop() error {
> 335: 	sm.mu.Lock()
  336: 	if !sm.running {
  337: 		sm.mu.Unlock()
  338: 		return nil
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/mobile/ui.go:520`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  518: 	}
  519: 
> 520: 	bgColor := n.BackgroundColor.(color.RGBA)
  521: 	bgColor.A = alpha
  522: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/mobile/ui.go:528`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  526: 	// Draw message text
  527: 	if n.Message != "" {
> 528: 		textColor := n.TextColor.(color.RGBA)
  529: 		textColor.A = alpha
  530: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/network/buffer_pool.go:33`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  31: //	*buf = append(*buf, data...)
  32: func AcquireBuffer() *[]byte {
> 33: 	return bufferPool.Get().(*[]byte)
  34: }
  35: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/buffer_stats.go:52`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  49: func (bs *BufferStats) RecordSend() {
  50: 	atomic.AddUint64(&bs.sent, 1)
  51: 
> 52: 	bs.mu.Lock()
  53: 	bs.currentSize++
  54: 	size := bs.currentSize
  55: 	bs.mu.Unlock()
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/buffer_stats.go:64`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  61: // RecordReceive records a successful receive from the channel.
  62: // Call this after successfully receiving from a channel.
  63: func (bs *BufferStats) RecordReceive() {
> 64: 	bs.mu.Lock()
  65: 	if bs.currentSize > 0 {
  66: 		bs.currentSize--
  67: 	}
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/buffer_stats.go:160`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  157: func (bs *BufferStats) Reset() {
  158: 	atomic.StoreUint64(&bs.sent, 0)
  159: 	atomic.StoreUint64(&bs.dropped, 0)
> 160: 	bs.mu.Lock()
  161: 	bs.currentSize = 0
  162: 	bs.lastWarnSize = 0
  163: 	bs.mu.Unlock()
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/network/client.go:290`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  288: 	// Close connection
  289: 	if c.conn != nil {
> 290: 		c.conn.Close()
  291: 	}
  292: 	c.mu.Unlock()
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/client.go:275`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  272: 
  273: // Disconnect closes the connection to the server.
  274: func (c *TCPClient) Disconnect() error {
> 275: 	c.mu.Lock()
  276: 	if !c.connected {
  277: 		c.mu.Unlock()
  278: 		return nil
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/client.go:334`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  331: 
  332: // SendInput queues an input command to send to the server.
  333: func (c *TCPClient) SendInput(inputType string, data []byte) error {
> 334: 	c.mu.Lock()
  335: 	if !c.connected {
  336: 		c.mu.Unlock()
  337: 		return fmt.Errorf("not connected")
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/client.go:450`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  447: 		}
  448: 
  449: 		// Update sequence number
> 450: 		c.mu.Lock()
  451: 		c.stateSeq = update.SequenceNumber
  452: 		c.mu.Unlock()
  453: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/client.go:481`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  478: 
  479: 		case <-pingTicker.C:
  480: 			// Send ping (empty input with type "ping")
> 481: 			c.mu.Lock()
  482: 			c.lastPing = time.Now()
  483: 			c.mu.Unlock()
  484: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/mock_server.go:188`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  185: // SimulatePlayerJoin simulates a player connecting.
  186: // Use this to test player connection handling.
  187: func (m *MockServer) SimulatePlayerJoin(playerID uint64) {
> 188: 	m.mu.Lock()
  189: 	m.Players[playerID] = true
  190: 	m.mu.Unlock()
  191: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/mock_server.go:202`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  199: // SimulatePlayerLeave simulates a player disconnecting.
  200: // Use this to test player disconnection handling.
  201: func (m *MockServer) SimulatePlayerLeave(playerID uint64) {
> 202: 	m.mu.Lock()
  203: 	delete(m.Players, playerID)
  204: 	m.mu.Unlock()
  205: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/network/priority_queue.go:52`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  50: func (h *priorityHeap) Push(x interface{}) {
  51: 	n := len(*h)
> 52: 	item := x.(*priorityItem)
  53: 	item.index = n
  54: 	*h = append(*h, item)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/network/priority_queue.go:113`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  111: 	}
  112: 
> 113: 	item := heap.Pop(&pq.heap).(*priorityItem)
  114: 	return item.update
  115: }
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/network/server.go:194`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  192: 	// Close listener
  193: 	if s.listener != nil {
> 194: 		s.listener.Close()
  195: 	}
  196: 
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/network/server.go:314`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  312: 
  313: 		if playerCount >= s.config.MaxPlayers {
> 314: 			conn.Close()
  315: 			s.errors <- fmt.Errorf("server full, rejected connection from %s", conn.RemoteAddr())
  316: 			continue
```

---

### ERROR HANDLING GAP: Close() error not checked
**File:** `pkg/network/server.go:567`
**Severity:** Low

**Description:** Close() method called without checking returned error
**Actual Behavior:** Close errors are silently ignored
**Correct Behavior:** Check error: if err := f.Close(); err != nil { ... }
**Impact:** Failed resource cleanup may go unnoticed
**Reproduction:** Force close failure and observe lack of error handling
**Code Reference:**
```go
  565: 		c.connected = false
  566: 		if c.conn != nil {
> 567: 			c.conn.Close()
  568: 		}
  569: 		close(c.updateSignal)
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:158`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  155: 
  156: // Start begins listening for client connections.
  157: func (s *TCPServer) Start() error {
> 158: 	s.clientsMu.Lock()
  159: 	if s.running {
  160: 		s.clientsMu.Unlock()
  161: 		return fmt.Errorf("server already running")
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:183`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  180: 
  181: // Stop shuts down the server.
  182: func (s *TCPServer) Stop() error {
> 183: 	s.clientsMu.Lock()
  184: 	if !s.running {
  185: 		s.clientsMu.Unlock()
  186: 		return nil
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:263`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  260: 	}
  261: 
  262: 	// Assign sequence number
> 263: 	s.stateMu.Lock()
  264: 	update.SequenceNumber = s.stateSeq
  265: 	s.stateSeq++
  266: 	s.stateMu.Unlock()
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:337`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  334: 		}
  335: 
  336: 		// Create client connection
> 337: 		s.clientsMu.Lock()
  338: 		playerID := s.nextPlayerID
  339: 		s.nextPlayerID++
  340: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:428`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  425: 		}
  426: 
  427: 		// Update last active
> 428: 		client.mu.Lock()
  429: 		client.lastActive = time.Now()
  430: 		client.mu.Unlock()
  431: 
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/network/server.go:526`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  523: 
  524: // disconnectClient removes a client from the server.
  525: func (s *TCPServer) disconnectClient(playerID uint64) {
> 526: 	s.clientsMu.Lock()
  527: 	client, exists := s.clients[playerID]
  528: 	if exists {
  529: 		client.disconnect()
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/procgen/terrain/multilevel.go:83`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  81: 		}
  82: 
> 83: 		terrain := result.(*Terrain)
  84: 		terrain.Level = i
  85: 		levels[i] = terrain
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/cache/sprite_cache.go:89`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  87: 		c.lru.MoveToFront(elem)
  88: 		c.stats.Hits++
> 89: 		return elem.Value.(*entry).image, true
  90: 	}
  91: 
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/cache/sprite_cache.go:106`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  104: 		// Update existing entry and move to front
  105: 		c.lru.MoveToFront(elem)
> 106: 		elem.Value.(*entry).image = img
  107: 		return
  108: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/cache/sprite_cache.go:143`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  141: func (c *SpriteCache) removeElement(elem *list.Element) {
  142: 	c.lru.Remove(elem)
> 143: 	e := elem.Value.(*entry)
  144: 	delete(c.cache, e.key)
  145: 	c.stats.TotalSize -= e.size
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/particles/pool.go:41`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  39: // Returns: Pooled ParticleSystem ready for use
  40: func NewParticleSystem(particles []Particle, pType ParticleType, config Config) *ParticleSystem {
> 41: 	ps := particleSystemPool.Get().(*ParticleSystem)
  42: 
  43: 	// Clear previous state
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/particles/pool.go:86`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  84: // Returns: Pointer to slice with 0 length, 100 capacity
  85: func AcquireParticleSlice() *[]Particle {
> 86: 	particles := particleSlicePool.Get().(*[]Particle)
  87: 	*particles = (*particles)[:0] // Reset length, keep capacity
  88: 	return particles
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/pool/image_pool.go:70`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  68: 		switch width {
  69: 		case SizePlayer:
> 70: 			return p.pool28.Get().(*ebiten.Image)
  71: 		case SizeSmall:
  72: 			return p.pool32.Get().(*ebiten.Image)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/pool/image_pool.go:72`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  70: 			return p.pool28.Get().(*ebiten.Image)
  71: 		case SizeSmall:
> 72: 			return p.pool32.Get().(*ebiten.Image)
  73: 		case SizeMedium:
  74: 			return p.pool64.Get().(*ebiten.Image)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/pool/image_pool.go:74`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  72: 			return p.pool32.Get().(*ebiten.Image)
  73: 		case SizeMedium:
> 74: 			return p.pool64.Get().(*ebiten.Image)
  75: 		case SizeLarge:
  76: 			return p.pool128.Get().(*ebiten.Image)
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/pool/image_pool.go:76`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  74: 			return p.pool64.Get().(*ebiten.Image)
  75: 		case SizeLarge:
> 76: 			return p.pool128.Get().(*ebiten.Image)
  77: 		}
  78: 	}
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/sprites/cache.go:123`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  121: 	}
  122: 
> 123: 	key := element.Value.(uint64)
  124: 	c.lruList.Remove(element)
  125: 	delete(c.cache, key)
```

---

### RESOURCE LEAK: Defer statement inside loop
**File:** `pkg/rendering/sprites/cache.go:342`
**Severity:** High

**Description:** Defer statement at line 342 inside for loop starting at line 339
**Actual Behavior:** All deferred calls accumulate and execute when function returns, not at each iteration
**Correct Behavior:** Extract loop body to separate function or manually close resources in loop
**Impact:** Resource exhaustion if loop runs many times - defers pile up
**Reproduction:** Run loop 1000+ times and monitor file descriptors/memory
**Code Reference:**
```go
  339: 	for w := 0; w < workers; w++ {
  340: 		wg.Add(1)
  341: 		go func() {
> 342: 			defer wg.Done()
  343: 			for job := range jobs {
  344: 				sprite, err := cg.Generate(job.config)
  345: 				resultsChan <- result{
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/rendering/sprites/cache.go:67`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  64: 	c.mutex.RUnlock()
  65: 
  66: 	if found {
> 67: 		c.mutex.Lock()
  68: 		// Move to front of LRU list (most recently used)
  69: 		c.lruList.MoveToFront(entry.element)
  70: 		c.hits++
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/rendering/sprites/cache.go:75`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  72: 		return entry.sprite
  73: 	}
  74: 
> 75: 	c.mutex.Lock()
  76: 	c.misses++
  77: 	c.mutex.Unlock()
  78: 	return nil
```

---

### CRITICAL BUG: Unchecked type assertion
**File:** `pkg/rendering/sprites/pool.go:34`
**Severity:** High

**Description:** Type assertion performed without checking if the assertion succeeded
**Actual Behavior:** Code panics if the type assertion fails
**Correct Behavior:** Use comma-ok idiom: value, ok := x.(Type); if !ok { ... }
**Impact:** Application crash when wrong type is encountered
**Reproduction:** Pass unexpected type to trigger panic
**Code Reference:**
```go
  32: // The returned image will be cleared (transparent).
  33: func (p *ImagePool) Get() *ebiten.Image {
> 34: 	img := p.pool.Get().(*ebiten.Image)
  35: 	// Clear the image for reuse
  36: 	img.Clear()
```

---

### RESOURCE LEAK: Defer statement inside loop
**File:** `pkg/rendering/sprites/pool.go:308`
**Severity:** High

**Description:** Defer statement at line 308 inside for loop starting at line 305
**Actual Behavior:** All deferred calls accumulate and execute when function returns, not at each iteration
**Correct Behavior:** Extract loop body to separate function or manually close resources in loop
**Impact:** Resource exhaustion if loop runs many times - defers pile up
**Reproduction:** Run loop 1000+ times and monitor file descriptors/memory
**Code Reference:**
```go
  305: 	for w := 0; w < workers; w++ {
  306: 		wg.Add(1)
  307: 		go func() {
> 308: 			defer wg.Done()
  309: 			for job := range jobs {
  310: 				sprite, err := cg.Generate(job.config)
  311: 				resultsChan <- result{
```

---

### CONCURRENCY ISSUE: Mutex locked without deferred unlock
**File:** `pkg/rendering/sprites/pool.go:81`
**Severity:** High

**Description:** Mutex.Lock() called without corresponding defer Mutex.Unlock()
**Actual Behavior:** Lock held indefinitely if function returns early via error or panic
**Correct Behavior:** Use: mu.Lock(); defer mu.Unlock()
**Impact:** Deadlock when other goroutines wait for lock that's never released
**Reproduction:** Trigger early return after lock and observe deadlock
**Code Reference:**
```go
  78: 	sp.mutex.RUnlock()
  79: 
  80: 	if !exists {
> 81: 		sp.mutex.Lock()
  82: 		// Double-check after acquiring write lock
  83: 		pool, exists = sp.pools[key]
  84: 		if !exists {
```

