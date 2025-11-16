package ui

import (
	"github.com/opd-ai/venture/pkg/rendering/display"
)

// UIScaler wraps display.Scaler for UI-specific scaling operations.
// Provides convenience methods for common UI scaling patterns.
type UIScaler struct {
	scaler *display.Scaler
}

// NewUIScaler creates a UI scaler from display config.
func NewUIScaler(cfg *display.Config) *UIScaler {
	return &UIScaler{
		scaler: display.NewScaler(cfg),
	}
}

// ScaleFont scales font size with UI-specific minimums.
func (s *UIScaler) ScaleFont(size int) int {
	return s.scaler.ScaleFontSize(size)
}

// ScaleButton scales button dimensions.
func (s *UIScaler) ScaleButton(width, height int) (int, int) {
	return s.scaler.ScaleSize(width, height)
}

// ScalePanel scales panel dimensions.
func (s *UIScaler) ScalePanel(width, height int) (int, int) {
	return s.scaler.ScaleSize(width, height)
}

// ScaleMargin scales margin/padding values.
func (s *UIScaler) ScaleMargin(margin int) int {
	return s.scaler.ScaleWidth(margin)
}

// ScalePadding scales padding values.
func (s *UIScaler) ScalePadding(padding int) int {
	return s.scaler.ScaleWidth(padding)
}

// ScaleBorder scales border thickness.
func (s *UIScaler) ScaleBorder(thickness int) int {
	scaled := s.scaler.ScaleWidth(thickness)
	if scaled < 1 {
		scaled = 1 // Minimum 1px border
	}
	return scaled
}

// ScaleIconSize scales icon dimensions (enforces square).
func (s *UIScaler) ScaleIconSize(size int) int {
	return s.scaler.ScaleWidth(size)
}

// ScalePosition scales UI element position.
func (s *UIScaler) ScalePosition(x, y int) (int, int) {
	return s.scaler.ScalePosition(x, y)
}

// GetScaleFactor returns underlying scale factor.
func (s *UIScaler) GetScaleFactor() float64 {
	return s.scaler.GetScaleFactor()
}

// ScaleMenuItemHeight scales menu item height (minimum 20px).
func (s *UIScaler) ScaleMenuItemHeight(height int) int {
	scaled := s.scaler.ScaleHeight(height)
	if scaled < 20 {
		scaled = 20
	}
	return scaled
}

// ScaleScrollbarWidth scales scrollbar width (minimum 8px).
func (s *UIScaler) ScaleScrollbarWidth(width int) int {
	scaled := s.scaler.ScaleWidth(width)
	if scaled < 8 {
		scaled = 8
	}
	return scaled
}
