package engine

// DropShadowComponent stores per-entity drop shadow visual parameters.
// The render system uses these to draw a soft elliptical shadow beneath entities.
type DropShadowComponent struct {
	// Ellipse dimensions in pixels
	ShadowWidth  float64
	ShadowHeight float64

	// Opacity 0.0 (invisible) to 1.0 (fully opaque)
	Opacity float64

	// Shadow color RGB (typically dark values)
	ColorR float64
	ColorG float64
	ColorB float64

	// Offset from entity center (e.g. for directional light)
	OffsetX float64
	OffsetY float64

	// Whether shadow is currently visible
	Enabled bool
}

// Type returns the component type identifier.
func (d *DropShadowComponent) Type() string {
	return "drop_shadow"
}

// NewDropShadowComponent creates a drop shadow with sensible defaults for a 32×32 entity.
func NewDropShadowComponent() *DropShadowComponent {
	return &DropShadowComponent{
		ShadowWidth:  20.0,
		ShadowHeight: 8.0,
		Opacity:      0.35,
		ColorR:       0.0,
		ColorG:       0.0,
		ColorB:       0.0,
		OffsetX:      0.0,
		OffsetY:      4.0,
		Enabled:      true,
	}
}
