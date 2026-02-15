// Package engine provides the EquipmentWearTintComponent for per-entity
// aggregate equipment damage-state visual tinting. The render system uses
// these values to darken and reduce opacity of entity sprites based on
// the overall condition of their equipped items.
package engine

// EquipmentWearTintComponent stores computed visual degradation parameters
// aggregated from all equipped item damage states. This is a transient
// visual component—not persisted in saves.
type EquipmentWearTintComponent struct {
	// OpacityMultiplier reduces sprite alpha (1.0=full, 0.7=heavily worn)
	OpacityMultiplier float64

	// ColorDarken shifts sprite RGB toward black (0.0=none, 0.4=heavy)
	ColorDarken float64

	// CrackDensity controls overlay crack effect intensity (0.0–1.0)
	CrackDensity float64

	// EdgeRoughness increases edge jaggedness from wear (0.0–1.0)
	EdgeRoughness float64

	// Dirtiness adds grime overlay intensity (0.0–1.0)
	Dirtiness float64

	// Genre-specific wear tint color (RGB 0.0–1.0)
	TintR float64
	TintG float64
	TintB float64

	// Whether tint is currently active (at least one non-pristine item)
	Enabled bool

	// Number of equipped items contributing to the aggregate
	EquippedCount int

	// Worst damage state name for debugging
	WorstState string
}

// Type returns the component type identifier.
func (e *EquipmentWearTintComponent) Type() string {
	return "equipment_wear_tint"
}

// NewEquipmentWearTintComponent creates a wear tint component with pristine defaults.
func NewEquipmentWearTintComponent() *EquipmentWearTintComponent {
	return &EquipmentWearTintComponent{
		OpacityMultiplier: 1.0,
		ColorDarken:       0.0,
		CrackDensity:      0.0,
		EdgeRoughness:     0.0,
		Dirtiness:         0.0,
		TintR:             0.6,
		TintG:             0.5,
		TintB:             0.4,
		Enabled:           false,
	}
}
