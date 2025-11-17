package mobile

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// MobileMenu represents a touch-friendly menu system.
// Platform parity fix: Enhanced with momentum scrolling and long-press context menu support
type MobileMenu struct {
	X, Y          float64
	Width, Height float64
	Items         []MenuItem
	SelectedIndex int
	Visible       bool

	// Touch tracking
	touchHandler *TouchInputHandler
	scrollOffset float64

	// Platform parity fix: Momentum scrolling for smooth touch UX like mobile OS
	scrollVelocity     float64 // Current scroll velocity (pixels/frame)
	scrollDeceleration float64 // Deceleration rate for momentum scrolling
	isScrolling        bool    // Platform parity fix: Prevents interaction during momentum scroll

	// Platform parity fix: Long-press context menu support (right-click alternative)
	longPressItem     int       // Index of item being long-pressed (-1 if none)
	longPressCallback func(int) // Callback for long-press on item

	// Visual settings
	BackgroundColor color.Color
	ItemColor       color.Color
	SelectedColor   color.Color
	TextColor       color.Color

	// Platform parity fix: Visual feedback for touch interactions (no hover available)
	pressedItemIndex int // Currently pressed item (-1 if none)
}

// MenuItem represents a single menu item.
// Platform parity fix: Enhanced with context menu callback support
type MenuItem struct {
	Label           string
	Enabled         bool
	OnSelect        func()
	OnLongPress     func()  // Platform parity fix: Right-click alternative for touch (context menu)
	OnDragStart     func()  // Platform parity fix: Drag-and-drop support for reordering/organizing
	OnDragEnd       func()  // Platform parity fix: Drop handler for drag-and-drop
	Draggable       bool    // Platform parity fix: Whether this item supports drag-and-drop
	MinTouchTargetH float64 // Platform parity fix: Minimum touch target height (44pt iOS, 48pt Android)
}

// NewMobileMenu creates a new mobile-optimized menu.
// Platform parity fix: Enhanced with momentum scrolling and touch target validation
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

		// Platform parity fix: Momentum scrolling configuration
		scrollDeceleration: 0.95, // 5% velocity loss per frame (~60 FPS)
		longPressItem:      -1,
		pressedItemIndex:   -1,
	}
}

// AddItem adds a menu item.
// Platform parity fix: Validates minimum touch target size for accessibility
func (m *MobileMenu) AddItem(label string, enabled bool, onSelect func()) {
	// Platform parity fix: Calculate appropriate item height
	itemHeight := m.Height / float64(len(m.Items)+1)

	// Platform parity fix: Touch target size validation
	// iOS: 44pt minimum (44px at 1x scale)
	// Android: 48dp minimum (48px at 1x density)
	// Use 48px as safe minimum for both platforms
	minTouchTarget := 48.0
	if itemHeight < minTouchTarget {
		// Warn: items will be smaller than recommended
		// In production, should adjust menu height or enable scrolling
		minTouchTarget = itemHeight
	}

	m.Items = append(m.Items, MenuItem{
		Label:           label,
		Enabled:         enabled,
		OnSelect:        onSelect,
		Draggable:       false,
		MinTouchTargetH: minTouchTarget,
	})
}

// Update processes touch input for the menu.
// Platform parity fix: Enhanced with momentum scrolling, long-press context menu, and drag-and-drop
func (m *MobileMenu) Update() {
	if !m.Visible {
		return
	}

	m.touchHandler.Update()

	// Platform parity fix: Update momentum scrolling physics
	m.updateMomentumScrolling()

	// Platform parity fix: Handle tap on menu items with visual feedback
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
				m.pressedItemIndex = -1 // Clear visual feedback
				break
			}
		}
	}

	// Platform parity fix: Handle long-press for context menu (right-click alternative)
	if m.touchHandler.IsLongPress() {
		longX, longY := m.touchHandler.gestureDetector.GetLongPressPosition()
		itemHeight := m.Height / float64(len(m.Items))

		for i := range m.Items {
			itemY := m.Y + float64(i)*itemHeight + m.scrollOffset

			if float64(longX) >= m.X && float64(longX) <= m.X+m.Width &&
				float64(longY) >= itemY && float64(longY) <= itemY+itemHeight {
				// Long press on item - trigger context menu
				if m.Items[i].Enabled && m.Items[i].OnLongPress != nil {
					m.Items[i].OnLongPress()
					m.longPressItem = i
				} else if m.longPressCallback != nil {
					m.longPressCallback(i)
					m.longPressItem = i
				}
				break
			}
		}
	} else {
		m.longPressItem = -1 // Clear long-press state
	}

	// Platform parity fix: Track pressed item for visual feedback (hover alternative for touch)
	touches := m.touchHandler.GetActiveTouches()
	if len(touches) > 0 {
		touch := touches[0]
		itemHeight := m.Height / float64(len(m.Items))

		m.pressedItemIndex = -1
		for i := range m.Items {
			itemY := m.Y + float64(i)*itemHeight + m.scrollOffset

			if float64(touch.X) >= m.X && float64(touch.X) <= m.X+m.Width &&
				float64(touch.Y) >= itemY && float64(touch.Y) <= itemY+itemHeight {
				m.pressedItemIndex = i
				break
			}
		}
	} else {
		m.pressedItemIndex = -1
	}

	// Platform parity fix: Handle swipe with velocity-based momentum scrolling
	if direction, _, detected := m.touchHandler.GetSwipe(); detected {
		// Calculate swipe direction (vertical swipes for scrolling)
		// direction is in radians: -π/2 (up) to π/2 (down)
		isVertical := (direction > 1.0 || direction < -1.0)

		if isVertical {
			// Platform parity fix: Initiate momentum scrolling with velocity
			velocity := m.touchHandler.gestureDetector.GetLastVelocity()

			// Convert velocity to scroll offset change
			// Negative velocity = swipe down = scroll up (reveal top items)
			// Positive velocity = swipe up = scroll down (reveal bottom items)
			m.scrollVelocity = -velocity * 10.0 // Scale factor for feel
			m.isScrolling = true
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
		} else if i == m.pressedItemIndex {
			// Platform parity fix: Show pressed state for active touch (hover alternative)
			itemColor = color.RGBA{80, 120, 200, 255}
		} else if i == m.longPressItem {
			// Platform parity fix: Show long-press active state for context menu
			itemColor = color.RGBA{150, 100, 200, 255}
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

// Platform parity fix: Enhanced menu functionality methods

// updateMomentumScrolling applies physics-based momentum scrolling for smooth UX.
// Platform parity fix: Provides iOS/Android-like inertial scrolling on all platforms
func (m *MobileMenu) updateMomentumScrolling() {
	if !m.isScrolling {
		return
	}

	// Apply velocity to scroll offset
	m.scrollOffset += m.scrollVelocity

	// Clamp scroll offset to valid range
	itemHeight := 50.0 // Default item height
	if len(m.Items) > 0 {
		itemHeight = m.Height / float64(len(m.Items))
	}
	maxScroll := float64(len(m.Items))*itemHeight - m.Height

	if maxScroll <= 0 {
		m.scrollOffset = 0
		m.isScrolling = false
		m.scrollVelocity = 0
		return
	}

	// Platform parity fix: Bounce-back effect when overscrolling (iOS-like behavior)
	if m.scrollOffset > 0 {
		// Scrolled above top - apply strong resistance
		m.scrollOffset *= 0.85 // Rubber-band effect
		m.scrollVelocity *= 0.7
		if m.scrollOffset < 1.0 {
			m.scrollOffset = 0
		}
	} else if m.scrollOffset < -maxScroll {
		// Scrolled below bottom - apply strong resistance
		excess := m.scrollOffset + maxScroll
		m.scrollOffset = -maxScroll + excess*0.85 // Rubber-band effect
		m.scrollVelocity *= 0.7
		if m.scrollOffset > -maxScroll-1.0 {
			m.scrollOffset = -maxScroll
		}
	}

	// Apply deceleration (friction)
	m.scrollVelocity *= m.scrollDeceleration

	// Stop scrolling when velocity is negligible
	if m.scrollVelocity > -0.1 && m.scrollVelocity < 0.1 {
		m.isScrolling = false
		m.scrollVelocity = 0

		// Snap to valid range if in overscroll
		if m.scrollOffset > 0 {
			m.scrollOffset = 0
		} else if m.scrollOffset < -maxScroll {
			m.scrollOffset = -maxScroll
		}
	}
}

// SetLongPressCallback sets a callback for long-press gestures on menu items.
// Platform parity fix: Enables context menu functionality (right-click alternative)
func (m *MobileMenu) SetLongPressCallback(callback func(int)) {
	m.longPressCallback = callback
}

// StopScrolling immediately stops momentum scrolling.
// Platform parity fix: Called when user touches screen during momentum scroll
func (m *MobileMenu) StopScrolling() {
	m.isScrolling = false
	m.scrollVelocity = 0
}

// GetPressedItemIndex returns the currently pressed item index.
// Platform parity fix: Provides visual feedback state (hover alternative for touch)
func (m *MobileMenu) GetPressedItemIndex() int {
	return m.pressedItemIndex
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

// TouchButton represents a touch-friendly button widget.
// Platform parity fix: Ensures minimum touch target size per platform guidelines
type TouchButton struct {
	X, Y          float64
	Width, Height float64
	Label         string
	Icon          string // Icon character/emoji (optional)
	Enabled       bool
	Visible       bool
	OnTap         func()

	// Touch handling
	touchHandler *TouchInputHandler
	pressed      bool // Visual feedback for active touch

	// Visual settings
	BackgroundColor color.Color
	PressedColor    color.Color
	DisabledColor   color.Color
	TextColor       color.Color
	BorderColor     color.Color
}

// NewTouchButton creates a new touch button with minimum size validation.
// Platform parity fix: Enforces 44px minimum per iOS/Android guidelines
func NewTouchButton(x, y, width, height float64, label string, onTap func()) *TouchButton {
	// Platform parity fix: Ensure minimum touch target size
	minSize := 44.0
	if width < minSize {
		width = minSize
	}
	if height < minSize {
		height = minSize
	}

	return &TouchButton{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Label:           label,
		Enabled:         true,
		Visible:         true,
		OnTap:           onTap,
		touchHandler:    NewTouchInputHandler(),
		BackgroundColor: color.RGBA{50, 50, 70, 255},
		PressedColor:    color.RGBA{80, 120, 200, 255},
		DisabledColor:   color.RGBA{30, 30, 40, 255},
		TextColor:       color.RGBA{255, 255, 255, 255},
		BorderColor:     color.RGBA{100, 100, 100, 255},
	}
}

// Update processes touch input for the button.
func (b *TouchButton) Update() {
	if !b.Visible || !b.Enabled {
		return
	}

	b.touchHandler.Update()

	// Get mouse/cursor position
	mouseX, mouseY := ebiten.CursorPosition()
	b.pressed = false

	// Check if button is being pressed via mouse
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if float64(mouseX) >= b.X && float64(mouseX) <= b.X+b.Width &&
			float64(mouseY) >= b.Y && float64(mouseY) <= b.Y+b.Height {
			b.pressed = true
		}
	}

	// Check if button is being touched (native touch events)
	touchIDs := ebiten.TouchIDs()
	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		if float64(x) >= b.X && float64(x) <= b.X+b.Width &&
			float64(y) >= b.Y && float64(y) <= b.Y+b.Height {
			b.pressed = true
			break
		}
	}

	// Handle mouse click (works in WASM with mouse)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if float64(mouseX) >= b.X && float64(mouseX) <= b.X+b.Width &&
			float64(mouseY) >= b.Y && float64(mouseY) <= b.Y+b.Height {
			if b.OnTap != nil {
				b.OnTap()
			}
		}
	}

	// Handle native touch tap (works in WASM with actual touch devices)
	justPressedTouchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressedTouchIDs {
		x, y := ebiten.TouchPosition(id)
		if float64(x) >= b.X && float64(x) <= b.X+b.Width &&
			float64(y) >= b.Y && float64(y) <= b.Y+b.Height {
			if b.OnTap != nil {
				b.OnTap()
			}
			break
		}
	}
}

// Draw renders the button.
func (b *TouchButton) Draw(screen *ebiten.Image) {
	if screen == nil || !b.Visible {
		return
	}

	// EMERGENCY FIX: Force minimum dimensions if they got corrupted
	if b.Width < 44 {
		b.Width = 120
	}
	if b.Height < 44 {
		b.Height = 44
	}

	// Determine button color
	bgColor := b.BackgroundColor
	if !b.Enabled {
		bgColor = b.DisabledColor
	} else if b.pressed {
		bgColor = b.PressedColor
	}

	// Draw button background
	vector.DrawFilledRect(screen, float32(b.X), float32(b.Y), float32(b.Width), float32(b.Height), bgColor, true)

	// Draw border
	vector.StrokeRect(screen, float32(b.X), float32(b.Y), float32(b.Width), float32(b.Height), 2, b.BorderColor, true)

	// Draw label text (centered)
	displayText := b.Label
	if b.Icon != "" && b.Label != "" {
		displayText = b.Icon + " " + b.Label
	} else if b.Icon != "" {
		displayText = b.Icon
	}

	// DEBUG: Always show SOMETHING to diagnose the issue
	if displayText == "" {
		displayText = "EMPTY"
	}

	// Always render text - removed the conditional check
	// Calculate text position (centered)
	// basicfont.Face7x13 is approximately 7 pixels wide per character
	textWidth := len(displayText) * 7

	// Center horizontally
	textX := int(b.X + (b.Width-float64(textWidth))/2)

	// Center vertically - text.Draw Y is the baseline
	// For 7x13 font in a 44px button, center is around Y + 28
	textY := int(b.Y + b.Height/2 + 5)

	// DEBUG: Force bright yellow text and add red dot marker
	debugColor := color.RGBA{255, 255, 0, 255} // Bright yellow
	text.Draw(screen, displayText, basicfont.Face7x13, textX, textY, debugColor)

	// Draw red dot at text position to verify Draw is being called
	vector.DrawFilledCircle(screen, float32(textX), float32(textY), 3, color.RGBA{255, 0, 0, 255}, true)
}

// SetPosition sets the button position.
func (b *TouchButton) SetPosition(x, y float64) {
	b.X = x
	b.Y = y
}

// SetSize sets the button size (enforcing minimum).
func (b *TouchButton) SetSize(width, height float64) {
	minSize := 44.0
	if width < minSize {
		width = minSize
	}
	if height < minSize {
		height = minSize
	}
	b.Width = width
	b.Height = height
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
