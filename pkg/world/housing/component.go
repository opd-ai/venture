package housing

// HousingComponent is an ECS component for entities that represent housing plots.
type HousingComponent struct {
	PlotID string // Reference to the plot managed by the housing manager
}

// Type returns the component type identifier.
func (h HousingComponent) Type() string {
	return "housing"
}
