package engine

import (
	"time"
)

// ExpressionCombo represents a synchronized group expression performance.
type ExpressionCombo struct {
	// ComboType identifies the combo pattern
	ComboType ComboType
	// ParticipantIDs are the entity IDs performing the combo
	ParticipantIDs []uint64
	// ExpressionType is the shared expression being performed
	ExpressionType ExpressionType
	// StartTime is when the combo was initiated
	StartTime float64
	// Active indicates if the combo is currently running
	Active bool
	// SyncWindow is the time window for entities to join (seconds)
	SyncWindow float64
}

// ComboType identifies different combo patterns
type ComboType int

const (
	// ComboSynchronized is when 2+ entities do the same expression simultaneously
	ComboSynchronized ComboType = iota
	// ComboWave is a sequential wave pattern across entities
	ComboWave
	// ComboCircle is entities arranged in a circle performing together
	ComboCircle
)

// String returns the combo type name
func (c ComboType) String() string {
	names := []string{"Synchronized", "Wave", "Circle"}
	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// ExpressionComboComponent tracks active combos for an entity
type ExpressionComboComponent struct {
	// ActiveCombo is the currently active combo (nil if not in a combo)
	ActiveCombo *ExpressionCombo
	// ComboHistory tracks recent combos for achievements
	ComboHistory []ComboRecord
	// TotalCombos is the lifetime count of combos participated in
	TotalCombos int
}

// Type returns the component type
func (e ExpressionComboComponent) Type() string {
	return "expressioncombo"
}

// ComboRecord tracks a completed combo for history
type ComboRecord struct {
	ComboType      ComboType
	ExpressionType ExpressionType
	ParticipantIDs []uint64
	Timestamp      int64
}

// ExpressionComboSystem manages synchronized group expression performances.
// Detects when multiple entities perform the same expression within a time window
// and creates combo bonuses/achievements.
type ExpressionComboSystem struct {
	world       *World
	pendingCombos map[ExpressionType]*ExpressionCombo
	comboWindow   float64 // Time window for joining a combo (seconds)
	currentTime   float64 // Tracks game time for combo timing
}

// NewExpressionComboSystem creates a new combo system
func NewExpressionComboSystem(world *World) *ExpressionComboSystem {
	return &ExpressionComboSystem{
		world:         world,
		pendingCombos: make(map[ExpressionType]*ExpressionCombo),
		comboWindow:   2.0, // 2 second window to join a combo
		currentTime:   0.0,
	}
}

// Update processes combo detection and management
func (s *ExpressionComboSystem) Update(deltaTime float64) {
	s.currentTime += deltaTime

	// Check for new expressions that could start or join combos
	entities := s.world.GetEntitiesWith("expression")
	
	for _, entity := range entities {
		expCompRaw, ok := entity.GetComponent("expression")
		if !ok {
			continue
		}
		expComp := expCompRaw.(*ExpressionComponent)

		// Only consider active expressions (just started)
		if expComp.ExpressionTime <= 0 {
			continue
		}

		// Check if this entity just started an expression
		// (time remaining is very close to full duration)
		s.checkForComboJoin(entity.ID, expComp.ActiveExpression, s.currentTime)
	}

	// Clean up expired pending combos
	for expType, combo := range s.pendingCombos {
		if s.currentTime-combo.StartTime > combo.SyncWindow {
			// Finalize combo if it has enough participants
			if len(combo.ParticipantIDs) >= 2 {
				s.finalizeCombo(combo)
			}
			delete(s.pendingCombos, expType)
		}
	}
}

// checkForComboJoin checks if an entity starting an expression should join a combo
func (s *ExpressionComboSystem) checkForComboJoin(entityID uint64, expressionType ExpressionType, currentTime float64) {
	// Get or create pending combo for this expression type
	combo, exists := s.pendingCombos[expressionType]
	
	if !exists {
		// Create new pending combo
		combo = &ExpressionCombo{
			ComboType:      ComboSynchronized,
			ParticipantIDs: []uint64{entityID},
			ExpressionType: expressionType,
			StartTime:      currentTime,
			Active:         false,
			SyncWindow:     s.comboWindow,
		}
		s.pendingCombos[expressionType] = combo
		return
	}

	// Check if entity is already in the combo
	for _, id := range combo.ParticipantIDs {
		if id == entityID {
			return // Already in combo
		}
	}

	// Check if within sync window
	if currentTime-combo.StartTime <= combo.SyncWindow {
		// Add to combo
		combo.ParticipantIDs = append(combo.ParticipantIDs, entityID)
		
		// Activate combo if we have 2+ participants
		if len(combo.ParticipantIDs) >= 2 && !combo.Active {
			combo.Active = true
			s.notifyComboStart(combo)
		}
	}
}

// finalizeCombo completes a combo and records it in participant histories
func (s *ExpressionComboSystem) finalizeCombo(combo *ExpressionCombo) {
	if !combo.Active || len(combo.ParticipantIDs) < 2 {
		return
	}

	timestamp := time.Now().Unix()
	record := ComboRecord{
		ComboType:      combo.ComboType,
		ExpressionType: combo.ExpressionType,
		ParticipantIDs: combo.ParticipantIDs,
		Timestamp:      timestamp,
	}

	// Record in each participant's history
	for _, entityID := range combo.ParticipantIDs {
		entity, ok := s.world.GetEntity(entityID)
		if !ok || entity == nil {
			continue
		}

		// Get or create combo component
		comboCompRaw, ok := entity.GetComponent("expressioncombo")
		var comboComp *ExpressionComboComponent
		
		if !ok {
			comboComp = &ExpressionComboComponent{
				ComboHistory: []ComboRecord{},
				TotalCombos:  0,
			}
			entity.AddComponent(comboComp)
		} else {
			comboComp = comboCompRaw.(*ExpressionComboComponent)
		}

		// Add to history (keep last 10)
		comboComp.ComboHistory = append(comboComp.ComboHistory, record)
		if len(comboComp.ComboHistory) > 10 {
			comboComp.ComboHistory = comboComp.ComboHistory[1:]
		}
		comboComp.TotalCombos++
		comboComp.ActiveCombo = nil
	}
}

// notifyComboStart marks the combo as active for all participants
func (s *ExpressionComboSystem) notifyComboStart(combo *ExpressionCombo) {
	for _, entityID := range combo.ParticipantIDs {
		entity, ok := s.world.GetEntity(entityID)
		if !ok || entity == nil {
			continue
		}

		// Get or create combo component
		comboCompRaw, ok := entity.GetComponent("expressioncombo")
		var comboComp *ExpressionComboComponent
		
		if !ok {
			comboComp = &ExpressionComboComponent{
				ComboHistory: []ComboRecord{},
				TotalCombos:  0,
			}
			entity.AddComponent(comboComp)
		} else {
			comboComp = comboCompRaw.(*ExpressionComboComponent)
		}

		comboComp.ActiveCombo = combo
	}
}

// GetActiveCombo returns the active combo for an entity, if any
func (s *ExpressionComboSystem) GetActiveCombo(entityID uint64) *ExpressionCombo {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	comboCompRaw, ok := entity.GetComponent("expressioncombo")
	if !ok {
		return nil
	}

	comboComp := comboCompRaw.(*ExpressionComboComponent)
	return comboComp.ActiveCombo
}

// GetComboCount returns the total number of combos an entity has participated in
func (s *ExpressionComboSystem) GetComboCount(entityID uint64) int {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return 0
	}

	comboCompRaw, ok := entity.GetComponent("expressioncombo")
	if !ok {
		return 0
	}

	comboComp := comboCompRaw.(*ExpressionComboComponent)
	return comboComp.TotalCombos
}
