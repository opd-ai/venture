package mobile

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// VirtualDPad represents an on-screen directional pad for movement.
type VirtualDPad struct {
	X, Y        float64 // Center position
	Radius      float64 // Outer radius
	InnerRadius float64 // Inner dead zone radius

	// Current state
	TouchID    ebiten.TouchID
	Active     bool
	DirectionX float64 // -1.0 to 1.0
	DirectionY float64 // -1.0 to 1.0

	// Visual settings
	OuterColor  color.Color
	InnerColor  color.Color
	ActiveColor color.Color
	Opacity     float64
}

// NewVirtualDPad creates a new virtual D-pad at the specified position.
func NewVirtualDPad(x, y, radius float64) *VirtualDPad {
	return &VirtualDPad{
		X:           x,
		Y:           y,
		Radius:      radius,
		InnerRadius: radius * 0.3,
		TouchID:     -1,
		OuterColor:  color.RGBA{100, 100, 100, 128},
		InnerColor:  color.RGBA{150, 150, 150, 200},
		ActiveColor: color.RGBA{200, 200, 255, 255},
		Opacity:     0.5,
	}
}

// Update processes touch input for the D-pad.
func (d *VirtualDPad) Update(touches map[ebiten.TouchID]*Touch) {
	// If we have an active touch, check if it's still active
	if d.TouchID >= 0 {
		if touch, exists := touches[d.TouchID]; exists && touch.Active {
			// Update direction based on touch position
			dx := float64(touch.X) - d.X
			dy := float64(touch.Y) - d.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance > d.InnerRadius {
				// Normalize to -1.0 to 1.0 range
				if distance > d.Radius {
					distance = d.Radius
				}
				d.DirectionX = dx / d.Radius
				d.DirectionY = dy / d.Radius
			} else {
				// Inside dead zone
				d.DirectionX = 0
				d.DirectionY = 0
			}
			d.Active = true
			return
		} else {
			// Touch released
			d.TouchID = -1
			d.Active = false
			d.DirectionX = 0
			d.DirectionY = 0
		}
	}

	// Look for new touch within D-pad area
	for id, touch := range touches {
		if !touch.Active {
			continue
		}

		dx := float64(touch.X) - d.X
		dy := float64(touch.Y) - d.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= d.Radius {
			// Touch started in D-pad area
			d.TouchID = id
			d.Active = true
			// Initial direction will be set on next Update
			break
		}
	}
}

// GetDirection returns the normalized direction (-1.0 to 1.0 for each axis).
func (d *VirtualDPad) GetDirection() (float64, float64) {
	return d.DirectionX, d.DirectionY
}

// IsActive returns true if the D-pad is currently being touched.
func (d *VirtualDPad) IsActive() bool {
	return d.Active
}

// Draw renders the D-pad on screen.
func (d *VirtualDPad) Draw(screen *ebiten.Image) {
	// Draw outer circle
	outerColor := d.OuterColor
	if d.Active {
		outerColor = d.ActiveColor
	}
	vector.DrawFilledCircle(screen, float32(d.X), float32(d.Y), float32(d.Radius), outerColor, true)

	// Draw inner circle (position indicator)
	innerX := d.X + d.DirectionX*d.Radius*0.5
	innerY := d.Y + d.DirectionY*d.Radius*0.5
	vector.DrawFilledCircle(screen, float32(innerX), float32(innerY), float32(d.InnerRadius), d.InnerColor, true)
}

// VirtualButton represents an on-screen button.
type VirtualButton struct {
	X, Y   float64 // Center position
	Radius float64

	// Current state
	TouchID ebiten.TouchID
	Active  bool
	Pressed bool // True for one frame when pressed

	// Visual settings
	Label       string
	NormalColor color.Color
	ActiveColor color.Color
	TextColor   color.Color
	Opacity     float64
}

// NewVirtualButton creates a new virtual button at the specified position.
func NewVirtualButton(x, y, radius float64, label string) *VirtualButton {
	return &VirtualButton{
		X:           x,
		Y:           y,
		Radius:      radius,
		TouchID:     -1,
		Label:       label,
		NormalColor: color.RGBA{100, 100, 100, 128},
		ActiveColor: color.RGBA{255, 200, 100, 255},
		TextColor:   color.RGBA{255, 255, 255, 255},
		Opacity:     0.5,
	}
}

// Update processes touch input for the button.
func (b *VirtualButton) Update(touches map[ebiten.TouchID]*Touch) {
	b.Pressed = false

	// If we have an active touch, check if it's still active
	if b.TouchID >= 0 {
		if touch, exists := touches[b.TouchID]; exists && touch.Active {
			b.Active = true
			return
		} else {
			// Touch released - trigger button press
			if b.Active {
				b.Pressed = true
			}
			b.TouchID = -1
			b.Active = false
		}
	}

	// Look for new touch within button area
	for id, touch := range touches {
		if !touch.Active {
			continue
		}

		dx := float64(touch.X) - b.X
		dy := float64(touch.Y) - b.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= b.Radius {
			// Touch started in button area
			b.TouchID = id
			b.Active = true
			break
		}
	}
}

// IsPressed returns true for one frame when the button is pressed.
func (b *VirtualButton) IsPressed() bool {
	return b.Pressed
}

// IsActive returns true while the button is being touched.
func (b *VirtualButton) IsActive() bool {
	return b.Active
}

// Draw renders the button on screen.
func (b *VirtualButton) Draw(screen *ebiten.Image) {
	// Draw button circle
	buttonColor := b.NormalColor
	if b.Active {
		buttonColor = b.ActiveColor
	}
	vector.DrawFilledCircle(screen, float32(b.X), float32(b.Y), float32(b.Radius), buttonColor, true)

	// Draw button border
	vector.StrokeCircle(screen, float32(b.X), float32(b.Y), float32(b.Radius), 2, b.TextColor, true)

	// Draw label text centered in button
	if b.Label != "" {
		// Measure text dimensions
		bounds, _ := font.BoundString(basicfont.Face7x13, b.Label)
		textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

		// Center text in button
		textX := int(b.X) - textWidth/2
		textY := int(b.Y) + textHeight/2

		// Draw text
		d := &font.Drawer{
			Dst:  screen,
			Src:  &image.Uniform{b.TextColor},
			Face: basicfont.Face7x13,
			Dot:  fixed.P(textX, textY),
		}
		d.DrawString(b.Label)
	}
}

// VirtualControlsLayout manages the complete virtual control layout.
type VirtualControlsLayout struct {
	DPad            *VirtualDPad
	ActionButton    *VirtualButton
	SecondaryButton *VirtualButton
	MenuButton      *VirtualButton

	Visible      bool
	touchHandler *TouchInputHandler
}

// NewVirtualControlsLayout creates a complete virtual control layout for a given screen size.
func NewVirtualControlsLayout(screenWidth, screenHeight int) *VirtualControlsLayout {
	// Calculate positions based on screen size
	dpadSize := float64(screenHeight) * 0.15
	buttonSize := float64(screenHeight) * 0.08
	margin := float64(screenHeight) * 0.05

	// D-pad on bottom left
	dpadX := margin + dpadSize
	dpadY := float64(screenHeight) - margin - dpadSize

	// Action buttons on bottom right
	actionX := float64(screenWidth) - margin - buttonSize*2.5
	actionY := float64(screenHeight) - margin - buttonSize

	secondaryX := float64(screenWidth) - margin - buttonSize
	secondaryY := float64(screenHeight) - margin - buttonSize*2.5

	// Menu button on top right
	menuX := float64(screenWidth) - margin - buttonSize
	menuY := margin + buttonSize

	return &VirtualControlsLayout{
		DPad:            NewVirtualDPad(dpadX, dpadY, dpadSize),
		ActionButton:    NewVirtualButton(actionX, actionY, buttonSize, "A"),
		SecondaryButton: NewVirtualButton(secondaryX, secondaryY, buttonSize, "B"),
		MenuButton:      NewVirtualButton(menuX, menuY, buttonSize*0.7, "☰"),
		Visible:         true,
		touchHandler:    NewTouchInputHandler(),
	}
}

// Update processes touch input for all virtual controls.
func (l *VirtualControlsLayout) Update() {
	if !l.Visible {
		return
	}

	l.touchHandler.Update()
	touches := make(map[ebiten.TouchID]*Touch)
	for _, touch := range l.touchHandler.GetActiveTouches() {
		touches[touch.ID] = touch
	}

	l.DPad.Update(touches)
	l.ActionButton.Update(touches)
	l.SecondaryButton.Update(touches)
	l.MenuButton.Update(touches)
}

// Draw renders all virtual controls on screen.
func (l *VirtualControlsLayout) Draw(screen *ebiten.Image) {
	if !l.Visible {
		return
	}

	l.DPad.Draw(screen)
	l.ActionButton.Draw(screen)
	l.SecondaryButton.Draw(screen)
	l.MenuButton.Draw(screen)
}

// GetMovementInput returns normalized movement direction from D-pad.
func (l *VirtualControlsLayout) GetMovementInput() (float64, float64) {
	return l.DPad.GetDirection()
}

// IsActionPressed returns true when the main action button is pressed.
func (l *VirtualControlsLayout) IsActionPressed() bool {
	return l.ActionButton.IsPressed()
}

// IsSecondaryPressed returns true when the secondary action button is pressed.
func (l *VirtualControlsLayout) IsSecondaryPressed() bool {
	return l.SecondaryButton.IsPressed()
}

// IsMenuPressed returns true when the menu button is pressed.
func (l *VirtualControlsLayout) IsMenuPressed() bool {
	return l.MenuButton.IsPressed()
}

// SetVisible controls whether virtual controls are shown and active.
func (l *VirtualControlsLayout) SetVisible(visible bool) {
	l.Visible = visible
}
