package engine

// ReputationSystem manages faction reputation and alignment
type ReputationSystem struct {
	world *World
}

// NewReputationSystem creates a new reputation system
func NewReputationSystem(world *World) *ReputationSystem {
	return &ReputationSystem{world: world}
}

// Update processes reputation changes
func (s *ReputationSystem) Update(deltaTime float64) {
	// Reputation changes are event-driven, not time-based
	// This is here for consistency with other systems
}

// ModifyReputation changes reputation with a faction
func (s *ReputationSystem) ModifyReputation(entityID uint64, faction string, delta float64) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}
	
	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		// Create new reputation component
		repComp := &ReputationComponent{
			Factions: make(map[string]float64),
			Alignment: Alignment{
				LawAxis:  0.0,
				GoodAxis: 0.0,
			},
			KarmaDeed: []Deed{},
		}
		entity.AddComponent(repComp)
		repCompRaw, _ = entity.GetComponent("reputation")
	}
	
	repComp := repCompRaw.(*ReputationComponent)
	
	// Modify reputation (clamped to -100 to +100)
	current := repComp.Factions[faction]
	repComp.Factions[faction] = clamp(current+delta, -100.0, 100.0)
}

// ShiftAlignment changes moral alignment
func (s *ReputationSystem) ShiftAlignment(entityID uint64, lawDelta, goodDelta float64) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}
	
	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		return
	}
	
	repComp := repCompRaw.(*ReputationComponent)
	
	// Shift alignment (clamped to -1 to +1)
	repComp.Alignment.LawAxis = clamp(repComp.Alignment.LawAxis+lawDelta, -1.0, 1.0)
	repComp.Alignment.GoodAxis = clamp(repComp.Alignment.GoodAxis+goodDelta, -1.0, 1.0)
}

// RecordDeed records a significant action
func (s *ReputationSystem) RecordDeed(entityID uint64, action string, alignmentChange Alignment) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}
	
	repCompRaw, ok := entity.GetComponent("reputation")
	if !ok {
		return
	}
	
	repComp := repCompRaw.(*ReputationComponent)
	
	deed := Deed{
		Action:          action,
		Timestamp:       0, // TODO: Get current timestamp
		AlignmentChange: alignmentChange,
	}
	
	repComp.KarmaDeed = append(repComp.KarmaDeed, deed)
	
	// Apply alignment change
	repComp.Alignment.LawAxis = clamp(repComp.Alignment.LawAxis+alignmentChange.LawAxis, -1.0, 1.0)
	repComp.Alignment.GoodAxis = clamp(repComp.Alignment.GoodAxis+alignmentChange.GoodAxis, -1.0, 1.0)
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
