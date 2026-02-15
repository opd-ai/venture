// Package engine provides the AfterimageComponent storing per-entity afterimage
// ghost data for the DodgeAfterimageSystem. Each ghost records a past position
// and a fading opacity to create translucent motion trail copies.
package engine

// AfterimageGhost represents a single fading ghost copy at a past position.
type AfterimageGhost struct {
	X, Y    float64
	Opacity float64 // 1.0 = fully visible, fades toward 0.0
}

// AfterimageComponent holds afterimage visual state for fast-moving entities.
// Pure data — all logic lives in DodgeAfterimageSystem.
type AfterimageComponent struct {
	Ghosts   []AfterimageGhost
	MaxGhost int     // Maximum number of active ghosts (default 5)
	Decay    float64 // Opacity decrease per second (default 3.0)
	Interval float64 // Minimum seconds between ghost spawns (default 0.06)
	// Tint color (genre-driven)
	TintR, TintG, TintB float64
	// Internal timing
	TimeSinceSpawn float64
	// Speed threshold squared for triggering afterimages
	SpeedThresholdSq float64
}

// Type implements the Component interface.
func (a *AfterimageComponent) Type() string { return "afterimage" }

// NewAfterimageComponent creates an AfterimageComponent with sensible defaults.
func NewAfterimageComponent() *AfterimageComponent {
	return &AfterimageComponent{
		Ghosts:           make([]AfterimageGhost, 0, 5),
		MaxGhost:         5,
		Decay:            3.0,
		Interval:         0.06,
		TintR:            1.0,
		TintG:            0.85,
		TintB:            0.4,
		SpeedThresholdSq: 14400.0, // 120 px/s
	}
}
