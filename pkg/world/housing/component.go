package housing

import "encoding/json"

// HousingComponent is an ECS component for entities that represent housing plots.
type HousingComponent struct {
	PlotID string // Reference to the plot managed by the housing manager
}

// Type returns the component type identifier.
func (h HousingComponent) Type() string {
	return "housing"
}

// Serialize converts the component to JSON bytes for persistence.
func (h *HousingComponent) Serialize() ([]byte, error) {
	return json.Marshal(h)
}

// Deserialize restores the component from JSON bytes.
func (h *HousingComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, h)
}
