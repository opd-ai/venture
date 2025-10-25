package mobile

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// MobileMenu represents a touch-friendly menu system.
type MobileMenu struct {
	X, Y          float64
	Width, Height float64
	Items         []MenuItem
	SelectedIndex int
	Visible       bool

	// Touch tracking
	touchHandler *TouchInputHandler
	scrollOffset float64

	// Visual settings
	BackgroundColor color.Color
	ItemColor       color.Color
	SelectedColor   color.Color
	TextColor       color.Color
}

// MenuItem represents a single menu item.
type MenuItem struct {
	Label    string
	Enabled  bool
	OnSelect func()
}

// NewMobileMenu creates a new mobile-optimized menu.
func NewMobileMenu(x, y, width, height float64) *MobileMenu {
	return &MobileMenu{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Items:           make([]MenuItem, 0),
		touchHandler:    NewTouchInputHandler(),
		BackgroundColor: color.RGBA{20, 20, 30, 230},
		ItemColor:       color.RGBA{50, 50, 70, 255},
		SelectedColor:   color.RGBA{100, 150, 255, 255},
		TextColor:       color.RGBA{255, 255, 255, 255},
	}
}

// AddItem adds a menu item.
func (m *MobileMenu) AddItem(label string, enabled bool, onSelect func()) {
	m.Items = append(m.Items, MenuItem{
		Label:    label,
		Enabled:  enabled,
		OnSelect: onSelect,
	})
}

// Update processes touch input for the menu.
func (m *MobileMenu) Update() {
	if !m.Visible {
		return
	}

	m.touchHandler.Update()

	// Handle tap on menu items
	if m.touchHandler.IsTapping() {
		tapX, tapY := m.touchHandler.GetTapPosition()
		itemHeight := m.Height / float64(len(m.Items))

		for i := range m.Items {
			itemY := m.Y + float64(i)*itemHeight + m.scrollOffset

			if float64(tapX) >= m.X && float64(tapX) <= m.X+m.Width &&
				float64(tapY) >= itemY && float64(tapY) <= itemY+itemHeight {
				// Tapped on item
				if m.Items[i].Enabled && m.Items[i].OnSelect != nil {
					m.Items[i].OnSelect()
				}
				break
			}
		}
	}

	// Handle swipe for scrolling (if menu is longer than visible area)
	if direction, distance, detected := m.touchHandler.GetSwipe(); detected {
		// Vertical swipe for scrolling
		if direction > -1.0 && direction < 1.0 {
			// Swipe up/down
			m.scrollOffset += distance * 0.5
			// Clamp scroll offset
			maxScroll := float64(len(m.Items))*50.0 - m.Height
			if m.scrollOffset > 0 {
				m.scrollOffset = 0
			} else if m.scrollOffset < -maxScroll && maxScroll > 0 {
				m.scrollOffset = -maxScroll
			}
		}
	}
}

// Draw renders the menu on screen.
func (m *MobileMenu) Draw(screen *ebiten.Image) {
	if !m.Visible {
		return
	}

	// Draw background
	vector.DrawFilledRect(screen, float32(m.X), float32(m.Y), float32(m.Width), float32(m.Height), m.BackgroundColor, true)

	// Draw menu items
	itemHeight := m.Height / float64(len(m.Items))
	for i, item := range m.Items {
		itemY := m.Y + float64(i)*itemHeight + m.scrollOffset

		// Skip items outside visible area
		if itemY+itemHeight < m.Y || itemY > m.Y+m.Height {
			continue
		}

		itemColor := m.ItemColor
		if i == m.SelectedIndex {
			itemColor = m.SelectedColor
		}
		if !item.Enabled {
			itemColor = color.RGBA{30, 30, 40, 255}
		}

		// Draw item background
		vector.DrawFilledRect(screen, float32(m.X+5), float32(itemY+2), float32(m.Width-10), float32(itemHeight-4), itemColor, true)

		// Draw item text
		if item.Label != "" {
			textColor := color.RGBA{255, 255, 255, 255}
			if !item.Enabled {
				textColor = color.RGBA{100, 100, 100, 255}
			}

			// Calculate text position (left-aligned with padding)
			textX := int(m.X) + 15
			textY := int(itemY+itemHeight/2) + 6 // Center vertically

			// Draw the text
			d := &font.Drawer{
				Dst:  screen,
				Src:  &image.Uniform{textColor},
				Face: basicfont.Face7x13,
				Dot:  fixed.P(textX, textY),
			}
			d.DrawString(item.Label)
		}
	}
}

// Show displays the menu.
func (m *MobileMenu) Show() {
	m.Visible = true
}

// Hide hides the menu.
func (m *MobileMenu) Hide() {
	m.Visible = false
}

// Toggle toggles menu visibility.
func (m *MobileMenu) Toggle() {
	m.Visible = !m.Visible
}

// IsVisible returns true if the menu is visible.
func (m *MobileMenu) IsVisible() bool {
	return m.Visible
}

// MobileHUD represents a mobile-optimized heads-up display.
type MobileHUD struct {
	ScreenWidth  int
	ScreenHeight int
	Orientation  Orientation

	// HUD elements
	HealthBar    *ProgressBar
	ManaBar      *ProgressBar
	ExpBar       *ProgressBar
	Minimap      *MinimapWidget
	Notification *NotificationWidget

	// Visibility
	Visible bool
}

// NewMobileHUD creates a new mobile-optimized HUD.
func NewMobileHUD(screenWidth, screenHeight int) *MobileHUD {
	orientation := GetOrientation(screenWidth, screenHeight)

	hud := &MobileHUD{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Orientation:  orientation,
		Visible:      true,
	}

	// Position HUD elements based on orientation
	hud.LayoutElements()

	return hud
}

// LayoutElements positions HUD elements based on screen orientation.
func (h *MobileHUD) LayoutElements() {
	margin := 10.0
	barWidth := 150.0
	barHeight := 20.0

	if h.Orientation == OrientationLandscape {
		// Top-left corner for stats in landscape
		h.HealthBar = NewProgressBar(margin, margin, barWidth, barHeight, color.RGBA{200, 50, 50, 255})
		h.ManaBar = NewProgressBar(margin, margin+barHeight+5, barWidth, barHeight, color.RGBA{50, 100, 200, 255})
		h.ExpBar = NewProgressBar(margin, float64(h.ScreenHeight)-margin-barHeight, barWidth*2, barHeight*0.5, color.RGBA{255, 215, 0, 255})
	} else {
		// Top of screen for portrait
		h.HealthBar = NewProgressBar(margin, margin, barWidth, barHeight, color.RGBA{200, 50, 50, 255})
		h.ManaBar = NewProgressBar(margin+barWidth+5, margin, barWidth, barHeight, color.RGBA{50, 100, 200, 255})
		h.ExpBar = NewProgressBar(margin, float64(h.ScreenHeight)-margin-barHeight, float64(h.ScreenWidth)-margin*2, barHeight*0.5, color.RGBA{255, 215, 0, 255})
	}

	// Minimap in top-right
	minimapSize := 100.0
	h.Minimap = NewMinimapWidget(float64(h.ScreenWidth)-margin-minimapSize, margin, minimapSize, minimapSize)

	// Notification in center-top
	h.Notification = NewNotificationWidget(float64(h.ScreenWidth)/2-150, margin+30, 300, 50)
}

// UpdateOrientation updates HUD layout if orientation changes.
func (h *MobileHUD) UpdateOrientation(screenWidth, screenHeight int) {
	newOrientation := GetOrientation(screenWidth, screenHeight)
	if newOrientation != h.Orientation {
		h.ScreenWidth = screenWidth
		h.ScreenHeight = screenHeight
		h.Orientation = newOrientation
		h.LayoutElements()
	}
}

// Update updates HUD elements.
func (h *MobileHUD) Update(deltaTime float64) {
	if h.Notification != nil {
		h.Notification.Update(deltaTime)
	}
}

// Draw renders the HUD on screen.
func (h *MobileHUD) Draw(screen *ebiten.Image) {
	if !h.Visible {
		return
	}

	if h.HealthBar != nil {
		h.HealthBar.Draw(screen)
	}
	if h.ManaBar != nil {
		h.ManaBar.Draw(screen)
	}
	if h.ExpBar != nil {
		h.ExpBar.Draw(screen)
	}
	if h.Minimap != nil {
		h.Minimap.Draw(screen)
	}
	if h.Notification != nil {
		h.Notification.Draw(screen)
	}
}

// SetHealth sets the health value (0.0 to 1.0).
func (h *MobileHUD) SetHealth(value float64) {
	if h.HealthBar != nil {
		h.HealthBar.SetValue(value)
	}
}

// SetMana sets the mana value (0.0 to 1.0).
func (h *MobileHUD) SetMana(value float64) {
	if h.ManaBar != nil {
		h.ManaBar.SetValue(value)
	}
}

// SetExperience sets the experience value (0.0 to 1.0).
func (h *MobileHUD) SetExperience(value float64) {
	if h.ExpBar != nil {
		h.ExpBar.SetValue(value)
	}
}

// ShowNotification displays a notification message.
func (h *MobileHUD) ShowNotification(message string, duration float64) {
	if h.Notification != nil {
		h.Notification.Show(message, duration)
	}
}

// ProgressBar represents a progress bar widget.
type ProgressBar struct {
	X, Y            float64
	Width, Height   float64
	Value           float64 // 0.0 to 1.0
	Color           color.Color
	BackgroundColor color.Color
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(x, y, width, height float64, barColor color.Color) *ProgressBar {
	return &ProgressBar{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Value:           1.0,
		Color:           barColor,
		BackgroundColor: color.RGBA{30, 30, 30, 200},
	}
}

// SetValue sets the progress value (0.0 to 1.0).
func (p *ProgressBar) SetValue(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	p.Value = value
}

// Draw renders the progress bar.
func (p *ProgressBar) Draw(screen *ebiten.Image) {
	// Draw background
	vector.DrawFilledRect(screen, float32(p.X), float32(p.Y), float32(p.Width), float32(p.Height), p.BackgroundColor, true)

	// Draw progress fill
	fillWidth := p.Width * p.Value
	vector.DrawFilledRect(screen, float32(p.X), float32(p.Y), float32(fillWidth), float32(p.Height), p.Color, true)

	// Draw border
	borderColor := color.RGBA{100, 100, 100, 255}
	vector.StrokeRect(screen, float32(p.X), float32(p.Y), float32(p.Width), float32(p.Height), 1, borderColor, true)
}

// MinimapWidget represents a minimap widget.
type MinimapWidget struct {
	X, Y            float64
	Width, Height   float64
	BackgroundColor color.Color

	// World data for rendering (set externally)
	TerrainWidth  int
	TerrainHeight int
	TileData      [][]int // 2D array of tile types
	PlayerX       int     // Player tile position
	PlayerY       int     // Player tile position
	FogOfWar      [][]bool
}

// NewMinimapWidget creates a new minimap widget.
func NewMinimapWidget(x, y, width, height float64) *MinimapWidget {
	return &MinimapWidget{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		BackgroundColor: color.RGBA{20, 20, 30, 200},
	}
}

// Draw renders the minimap.
func (m *MinimapWidget) Draw(screen *ebiten.Image) {
	// Draw background
	vector.DrawFilledRect(screen, float32(m.X), float32(m.Y), float32(m.Width), float32(m.Height), m.BackgroundColor, true)

	// Draw border
	borderColor := color.RGBA{100, 100, 100, 255}
	vector.StrokeRect(screen, float32(m.X), float32(m.Y), float32(m.Width), float32(m.Height), 2, borderColor, true)

	// Draw minimap content if terrain data is available
	if m.TileData != nil && m.TerrainWidth > 0 && m.TerrainHeight > 0 {
		// Calculate tile scaling to fit terrain in minimap
		scaleX := m.Width / float64(m.TerrainWidth)
		scaleY := m.Height / float64(m.TerrainHeight)
		tileScale := scaleX
		if scaleY < scaleX {
			tileScale = scaleY
		}

		// Draw terrain tiles
		for y := 0; y < m.TerrainHeight && y < len(m.TileData); y++ {
			for x := 0; x < m.TerrainWidth && x < len(m.TileData[y]); x++ {
				// Check fog of war
				if m.FogOfWar != nil && y < len(m.FogOfWar) && x < len(m.FogOfWar[y]) {
					if !m.FogOfWar[y][x] {
						continue // Skip unexplored tiles
					}
				}

				// Get tile color based on type
				tileType := m.TileData[y][x]
				tileColor := m.getTileColorForType(tileType)

				// Calculate pixel position
				pixelX := float32(m.X) + float32(float64(x)*tileScale)
				pixelY := float32(m.Y) + float32(float64(y)*tileScale)
				pixelSize := float32(tileScale)

				if pixelSize < 1 {
					pixelSize = 1
				}

				vector.DrawFilledRect(screen, pixelX, pixelY, pixelSize, pixelSize, tileColor, true)
			}
		}

		// Draw player icon
		if m.PlayerX >= 0 && m.PlayerX < m.TerrainWidth && m.PlayerY >= 0 && m.PlayerY < m.TerrainHeight {
			pixelX := float32(m.X) + float32(float64(m.PlayerX)*tileScale)
			pixelY := float32(m.Y) + float32(float64(m.PlayerY)*tileScale)

			// Draw player as bright circle
			vector.DrawFilledCircle(screen, pixelX, pixelY, 3, color.RGBA{100, 200, 255, 255}, true)
		}
	}
}

// getTileColorForType returns a color for a given tile type.
func (m *MinimapWidget) getTileColorForType(tileType int) color.Color {
	// Map tile types to colors (simplified version)
	// 0 = wall/solid, 1 = floor/walkable
	switch tileType {
	case 0:
		return color.RGBA{60, 60, 60, 255} // Dark gray for walls
	case 1:
		return color.RGBA{150, 150, 150, 255} // Light gray for floor
	default:
		return color.RGBA{100, 100, 100, 255} // Default gray
	}
}

// NotificationWidget displays temporary notifications.
type NotificationWidget struct {
	X, Y            float64
	Width, Height   float64
	Message         string
	Visible         bool
	Duration        float64
	Remaining       float64
	BackgroundColor color.Color
	TextColor       color.Color
}

// NewNotificationWidget creates a new notification widget.
func NewNotificationWidget(x, y, width, height float64) *NotificationWidget {
	return &NotificationWidget{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		BackgroundColor: color.RGBA{50, 50, 50, 220},
		TextColor:       color.RGBA{255, 255, 255, 255},
	}
}

// Show displays a notification for the specified duration.
func (n *NotificationWidget) Show(message string, duration float64) {
	n.Message = message
	n.Duration = duration
	n.Remaining = duration
	n.Visible = true
}

// Update updates the notification timer.
func (n *NotificationWidget) Update(deltaTime float64) {
	if n.Visible {
		n.Remaining -= deltaTime
		if n.Remaining <= 0 {
			n.Visible = false
		}
	}
}

// Draw renders the notification.
func (n *NotificationWidget) Draw(screen *ebiten.Image) {
	if !n.Visible {
		return
	}

	// Fade out in last second
	alpha := uint8(255)
	if n.Remaining < 1.0 {
		alpha = uint8(n.Remaining * 255)
	}

	bgColor := n.BackgroundColor.(color.RGBA)
	bgColor.A = alpha

	// Draw background
	vector.DrawFilledRect(screen, float32(n.X), float32(n.Y), float32(n.Width), float32(n.Height), bgColor, true)

	// Draw message text
	if n.Message != "" {
		textColor := n.TextColor.(color.RGBA)
		textColor.A = alpha

		// Center text horizontally and vertically
		bounds, _ := font.BoundString(basicfont.Face7x13, n.Message)
		textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

		textX := int(n.X + (n.Width-float64(textWidth))/2)
		textY := int(n.Y + (n.Height+float64(textHeight))/2)

		d := &font.Drawer{
			Dst:  screen,
			Src:  &image.Uniform{textColor},
			Face: basicfont.Face7x13,
			Dot:  fixed.P(textX, textY),
		}
		d.DrawString(n.Message)
	}
}
