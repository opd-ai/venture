package engine

// Alignment represents moral alignment
type Alignment struct {
	LawAxis  float64 // -1 (Chaotic) to +1 (Lawful)
	GoodAxis float64 // -1 (Evil) to +1 (Good)
}

// Deed represents a significant action
type Deed struct {
	Action      string
	Timestamp   int64
	AlignmentChange Alignment
}

// ReputationComponent tracks faction reputation and alignment
type ReputationComponent struct {
	Factions  map[string]float64 // Faction name → reputation (-100 to +100)
	Alignment Alignment
	KarmaDeed []Deed // History of significant actions
}

// Type returns the component type
func (r ReputationComponent) Type() string {
	return "reputation"
}
