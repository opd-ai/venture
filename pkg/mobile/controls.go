package mobile

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Platform parity fix: Input rate limiting and spam prevention

// InputRateLimiter prevents rapid input spam across all platforms.
// Platform parity fix: Addresses button mashing, accidental double-taps, lag-induced duplicates
type InputRateLimiter struct {
	lastInputTime map[string]time.Time
	cooldowns     map[string]time.Duration
	inputCounts   map[string]int // Platform parity fix: Track rapid inputs for anti-spam
	timeWindow    time.Duration  // Time window for counting rapid inputs
	maxInputs     int            // Maximum inputs allowed in time window
}

// NewInputRateLimiter creates a new input rate limiter.
// Platform parity fix: Configurable per-action cooldowns and spam detection
func NewInputRateLimiter() *InputRateLimiter {
	return &InputRateLimiter{
		lastInputTime: make(map[string]time.Time),
		cooldowns:     make(map[string]time.Duration),
		inputCounts:   make(map[string]int),
		timeWindow:    1 * time.Second, // 1 second window for spam detection
		maxInputs:     10,              // Max 10 inputs per second (prevents spam)
	}
}

// SetCooldown configures cooldown period for a specific action.
// Platform parity fix: Prevents double-tap on laggy connections or fat-finger errors
func (l *InputRateLimiter) SetCooldown(actionID string, cooldown time.Duration) {
	l.cooldowns[actionID] = cooldown
}

// CanExecute checks if an action can be executed (not on cooldown).
// Platform parity fix: Returns false if action is on cooldown or exceeds spam threshold
func (l *InputRateLimiter) CanExecute(actionID string) bool {
	// INTENTIONAL time.Now() usage: Input rate limiting timestamp (non-procgen operational timing).
	// This is NOT part of procedural content generation and does not affect determinism.
	now := time.Now()

	// Platform parity fix: Check cooldown
	if lastTime, exists := l.lastInputTime[actionID]; exists {
		cooldown := l.cooldowns[actionID]
		if cooldown == 0 {
			cooldown = 100 * time.Millisecond // Default 100ms cooldown
		}

		if now.Sub(lastTime) < cooldown {
			return false // Still on cooldown
		}
	}

	// Platform parity fix: Check spam threshold
	count := l.inputCounts[actionID]
	if count >= l.maxInputs {
		// Too many inputs in time window - likely spam or lag
		return false
	}

	return true
}

// RecordInput records that an action was executed.
// Platform parity fix: Updates cooldown and spam counter
func (l *InputRateLimiter) RecordInput(actionID string) {
	// INTENTIONAL time.Now() usage: Input rate limiting timestamp (non-procgen operational timing).
	// This is NOT part of procedural content generation and does not affect determinism.
	now := time.Now()
	l.lastInputTime[actionID] = now

	// Platform parity fix: Increment spam counter
	l.inputCounts[actionID]++
}

// Update cleans up old input counts from spam detection.
// Platform parity fix: Call every frame to reset spam counters after time window
func (l *InputRateLimiter) Update() {
	// INTENTIONAL time.Now() usage: Input rate limiting cleanup (non-procgen operational timing).
	// This is NOT part of procedural content generation and does not affect determinism.
	now := time.Now()

	// Platform parity fix: Reset spam counters for actions outside time window
	for actionID, lastTime := range l.lastInputTime {
		if now.Sub(lastTime) > l.timeWindow {
			l.inputCounts[actionID] = 0
		}
	}
}

// GetRemainingCooldown returns remaining cooldown time for an action.
// Platform parity fix: Allows UI to show cooldown indicators
func (l *InputRateLimiter) GetRemainingCooldown(actionID string) time.Duration {
	if lastTime, exists := l.lastInputTime[actionID]; exists {
		cooldown := l.cooldowns[actionID]
		if cooldown == 0 {
			cooldown = 100 * time.Millisecond
		}

		elapsed := time.Since(lastTime)
		if elapsed < cooldown {
			return cooldown - elapsed
		}
	}

	return 0
}

// GetSpamCount returns current spam input count for an action.
// Platform parity fix: Diagnostic tool for detecting input issues
func (l *InputRateLimiter) GetSpamCount(actionID string) int {
	return l.inputCounts[actionID]
}

// Haptic feedback settings for mobile controls.
// These values provide tactile responsiveness without being intrusive.
const (
	hapticLightDuration   = 10 * time.Millisecond // Light tap for D-pad touch
	hapticLightMagnitude  = 0.3                   // 30% strength
	hapticButtonDuration  = 20 * time.Millisecond // Medium tap for button press
	hapticButtonMagnitude = 0.5                   // 50% strength
	hapticMinInterval     = 50 * time.Millisecond // Minimum time between haptics
)

// triggerHaptic triggers device vibration with rate limiting.
// lastHaptic should be a pointer to the last haptic time (0 for first call).
func triggerHaptic(duration time.Duration, magnitude float64, lastHaptic *time.Time) {
	// INTENTIONAL time.Now() usage: Haptic feedback rate limiting (non-procgen operational timing).
	// This is NOT part of procedural content generation and does not affect determinism.
	now := time.Now()

	// Rate limiting: only trigger if enough time has passed since last haptic
	if !lastHaptic.IsZero() && now.Sub(*lastHaptic) < hapticMinInterval {
		return
	}

	// Trigger vibration on mobile devices
	ebiten.Vibrate(&ebiten.VibrateOptions{
		Duration:  duration,
		Magnitude: magnitude,
	})

	*lastHaptic = now
}

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

	// Haptic feedback tracking
	lastHaptic time.Time
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

			// Trigger light haptic feedback when D-pad is touched
			triggerHaptic(hapticLightDuration, hapticLightMagnitude, &d.lastHaptic)

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

	// Haptic feedback tracking
	lastHaptic time.Time
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

				// Trigger medium haptic feedback when button is pressed
				triggerHaptic(hapticButtonDuration, hapticButtonMagnitude, &b.lastHaptic)
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
// Platform parity fix: Extended to provide complete action coverage for Mobile/WASM.
// All game actions available on Desktop are now accessible via virtual controls.
type VirtualControlsLayout struct {
	DPad            *VirtualDPad
	ActionButton    *VirtualButton
	SecondaryButton *VirtualButton
	MenuButton      *VirtualButton

	// Platform parity fix: Additional action buttons for complete input coverage
	// These buttons provide access to actions that are keyboard-only on Desktop
	SpellButtons    []*VirtualButton // Spell casting buttons (1-5)
	InventoryButton *VirtualButton   // Quick access to inventory
	TargetButton    *VirtualButton   // Cycle targets button
	InteractButton  *VirtualButton   // Interact/Use button (F key equivalent)

	// Platform parity fix: UI shortcut buttons for complete menu accessibility
	// These provide touch equivalents to all keyboard UI shortcuts
	CharacterButton *VirtualButton // Character sheet (C key equivalent)
	SkillsButton    *VirtualButton // Skills/Skill tree (K key equivalent)
	QuestLogButton  *VirtualButton // Quest log (J key equivalent)
	MapButton       *VirtualButton // World map (M key equivalent)

	// Extended controls visibility (can be toggled for cleaner UI)
	ShowSpellButtons bool
	ShowUIShortcuts  bool // Toggle visibility of UI shortcut buttons

	Visible      bool
	touchHandler *TouchInputHandler
	screenWidth  int
	screenHeight int
}

// NewVirtualControlsLayout creates a complete virtual control layout for a given screen size.
// Platform parity fix: Now includes all action buttons to ensure complete input coverage
// for Mobile and WASM platforms matching Desktop keyboard functionality.
func NewVirtualControlsLayout(screenWidth, screenHeight int) *VirtualControlsLayout {
	// Calculate positions based on screen size
	dpadSize := float64(screenHeight) * 0.15
	buttonSize := float64(screenHeight) * 0.08
	smallButtonSize := float64(screenHeight) * 0.06
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

	// Platform parity fix: Inventory button on top right (next to menu)
	inventoryX := menuX - buttonSize*1.3
	inventoryY := menuY

	// Platform parity fix: Target cycle button above D-pad
	targetX := dpadX
	targetY := dpadY - dpadSize - smallButtonSize - margin*0.5

	// Platform parity fix: Interact button near action buttons
	interactX := actionX - buttonSize*1.5
	interactY := actionY

	// Platform parity fix: Spell buttons arranged horizontally above action area
	spellButtonY := actionY - buttonSize*1.8
	spellButtons := make([]*VirtualButton, 5)
	for i := 0; i < 5; i++ {
		spellX := float64(screenWidth)/2 - smallButtonSize*3 + float64(i)*smallButtonSize*1.3
		spellButtons[i] = NewVirtualButton(spellX, spellButtonY, smallButtonSize, string(rune('1'+i)))
	}

	// Platform parity fix: UI shortcut buttons arranged on top left (below HUD area)
	// These provide touch equivalents for Character (C), Skills (K), Quest (J), Map (M)
	uiButtonY := margin + buttonSize*0.7
	uiButtonSpacing := smallButtonSize * 1.3
	characterX := margin + smallButtonSize
	skillsX := characterX + uiButtonSpacing
	questX := skillsX + uiButtonSpacing
	mapX := questX + uiButtonSpacing

	return &VirtualControlsLayout{
		DPad:            NewVirtualDPad(dpadX, dpadY, dpadSize),
		ActionButton:    NewVirtualButton(actionX, actionY, buttonSize, "A"),
		SecondaryButton: NewVirtualButton(secondaryX, secondaryY, buttonSize, "B"),
		MenuButton:      NewVirtualButton(menuX, menuY, buttonSize*0.7, "☰"),
		InventoryButton: NewVirtualButton(inventoryX, inventoryY, buttonSize*0.7, "I"),
		TargetButton:    NewVirtualButton(targetX, targetY, smallButtonSize, "↹"),
		InteractButton:  NewVirtualButton(interactX, interactY, smallButtonSize, "F"),
		SpellButtons:    spellButtons,
		// Platform parity fix: UI shortcut buttons for complete menu accessibility
		CharacterButton:  NewVirtualButton(characterX, uiButtonY, smallButtonSize, "C"),
		SkillsButton:     NewVirtualButton(skillsX, uiButtonY, smallButtonSize, "K"),
		QuestLogButton:   NewVirtualButton(questX, uiButtonY, smallButtonSize, "J"),
		MapButton:        NewVirtualButton(mapX, uiButtonY, smallButtonSize, "M"),
		ShowSpellButtons: true,
		ShowUIShortcuts:  true,
		Visible:          true,
		touchHandler:     NewTouchInputHandler(),
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
	}
}

// Update processes touch input for all virtual controls.
// Platform parity fix: Now updates all extended controls including spell buttons.
func (l *VirtualControlsLayout) Update() {
	if !l.Visible {
		return
	}

	touches := l.collectActiveTouches()
	l.updateCoreControls(touches)
	l.updateExtendedControls(touches)
	l.updateUIShortcuts(touches)
	l.updateSpellButtons(touches)
}

// collectActiveTouches retrieves and maps all active touches from the touch handler.
func (l *VirtualControlsLayout) collectActiveTouches() map[ebiten.TouchID]*Touch {
	l.touchHandler.Update()
	touches := make(map[ebiten.TouchID]*Touch)
	for _, touch := range l.touchHandler.GetActiveTouches() {
		touches[touch.ID] = touch
	}
	return touches
}

// updateCoreControls updates the primary D-pad and action buttons.
func (l *VirtualControlsLayout) updateCoreControls(touches map[ebiten.TouchID]*Touch) {
	l.DPad.Update(touches)
	l.ActionButton.Update(touches)
	l.SecondaryButton.Update(touches)
	l.MenuButton.Update(touches)
}

// updateExtendedControls updates the inventory, target, and interact buttons.
func (l *VirtualControlsLayout) updateExtendedControls(touches map[ebiten.TouchID]*Touch) {
	if l.InventoryButton != nil {
		l.InventoryButton.Update(touches)
	}
	if l.TargetButton != nil {
		l.TargetButton.Update(touches)
	}
	if l.InteractButton != nil {
		l.InteractButton.Update(touches)
	}
}

// updateUIShortcuts updates the character, skills, quest, and map shortcut buttons.
func (l *VirtualControlsLayout) updateUIShortcuts(touches map[ebiten.TouchID]*Touch) {
	if !l.ShowUIShortcuts {
		return
	}
	if l.CharacterButton != nil {
		l.CharacterButton.Update(touches)
	}
	if l.SkillsButton != nil {
		l.SkillsButton.Update(touches)
	}
	if l.QuestLogButton != nil {
		l.QuestLogButton.Update(touches)
	}
	if l.MapButton != nil {
		l.MapButton.Update(touches)
	}
}

// updateSpellButtons updates all spell casting buttons.
func (l *VirtualControlsLayout) updateSpellButtons(touches map[ebiten.TouchID]*Touch) {
	if !l.ShowSpellButtons {
		return
	}
	for _, btn := range l.SpellButtons {
		if btn != nil {
			btn.Update(touches)
		}
	}
}

// Draw renders all virtual controls on screen.
// Platform parity fix: Now renders all extended controls including spell buttons.
func (l *VirtualControlsLayout) Draw(screen *ebiten.Image) {
	if !l.Visible {
		return
	}

	l.drawCoreControls(screen)
	l.drawExtendedControls(screen)
	l.drawUIShortcuts(screen)
	l.drawSpellButtons(screen)
}

// drawCoreControls renders the primary D-pad and action buttons.
func (l *VirtualControlsLayout) drawCoreControls(screen *ebiten.Image) {
	l.DPad.Draw(screen)
	l.ActionButton.Draw(screen)
	l.SecondaryButton.Draw(screen)
	l.MenuButton.Draw(screen)
}

// drawExtendedControls renders the inventory, target, and interact buttons.
func (l *VirtualControlsLayout) drawExtendedControls(screen *ebiten.Image) {
	if l.InventoryButton != nil {
		l.InventoryButton.Draw(screen)
	}
	if l.TargetButton != nil {
		l.TargetButton.Draw(screen)
	}
	if l.InteractButton != nil {
		l.InteractButton.Draw(screen)
	}
}

// drawUIShortcuts renders the character, skills, quest, and map shortcut buttons.
func (l *VirtualControlsLayout) drawUIShortcuts(screen *ebiten.Image) {
	if !l.ShowUIShortcuts {
		return
	}
	if l.CharacterButton != nil {
		l.CharacterButton.Draw(screen)
	}
	if l.SkillsButton != nil {
		l.SkillsButton.Draw(screen)
	}
	if l.QuestLogButton != nil {
		l.QuestLogButton.Draw(screen)
	}
	if l.MapButton != nil {
		l.MapButton.Draw(screen)
	}
}

// drawSpellButtons renders all spell casting buttons.
func (l *VirtualControlsLayout) drawSpellButtons(screen *ebiten.Image) {
	if !l.ShowSpellButtons {
		return
	}
	for _, btn := range l.SpellButtons {
		if btn != nil {
			btn.Draw(screen)
		}
	}
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

// IsVisible returns true if virtual controls are currently visible.
// Platform parity fix: Allows checking visibility state for UI logic.
func (l *VirtualControlsLayout) IsVisible() bool {
	return l.Visible
}

// Platform parity fix: Additional input getter methods for complete action coverage

// IsInventoryPressed returns true when the inventory button is pressed.
// Platform parity fix: Provides touch equivalent of 'I' key on Desktop.
func (l *VirtualControlsLayout) IsInventoryPressed() bool {
	return l.InventoryButton != nil && l.InventoryButton.IsPressed()
}

// IsTargetPressed returns true when the target cycle button is pressed.
// Platform parity fix: Provides touch equivalent of 'Tab' key on Desktop.
func (l *VirtualControlsLayout) IsTargetPressed() bool {
	return l.TargetButton != nil && l.TargetButton.IsPressed()
}

// IsInteractPressed returns true when the interact button is pressed.
// Platform parity fix: Provides touch equivalent of 'F' key on Desktop.
func (l *VirtualControlsLayout) IsInteractPressed() bool {
	return l.InteractButton != nil && l.InteractButton.IsPressed()
}

// IsCharacterPressed returns true when the character button is pressed.
// Platform parity fix: Provides touch equivalent of 'C' key on Desktop.
func (l *VirtualControlsLayout) IsCharacterPressed() bool {
	return l.ShowUIShortcuts && l.CharacterButton != nil && l.CharacterButton.IsPressed()
}

// IsSkillsPressed returns true when the skills button is pressed.
// Platform parity fix: Provides touch equivalent of 'K' key on Desktop.
func (l *VirtualControlsLayout) IsSkillsPressed() bool {
	return l.ShowUIShortcuts && l.SkillsButton != nil && l.SkillsButton.IsPressed()
}

// IsQuestLogPressed returns true when the quest log button is pressed.
// Platform parity fix: Provides touch equivalent of 'J' key on Desktop.
func (l *VirtualControlsLayout) IsQuestLogPressed() bool {
	return l.ShowUIShortcuts && l.QuestLogButton != nil && l.QuestLogButton.IsPressed()
}

// IsMapPressed returns true when the map button is pressed.
// Platform parity fix: Provides touch equivalent of 'M' key on Desktop.
func (l *VirtualControlsLayout) IsMapPressed() bool {
	return l.ShowUIShortcuts && l.MapButton != nil && l.MapButton.IsPressed()
}

// IsSpellPressed returns true when the specified spell button (1-5) is pressed.
// Platform parity fix: Provides touch equivalent of number keys on Desktop.
func (l *VirtualControlsLayout) IsSpellPressed(slot int) bool {
	if !l.ShowSpellButtons || slot < 1 || slot > len(l.SpellButtons) {
		return false
	}
	btn := l.SpellButtons[slot-1]
	return btn != nil && btn.IsPressed()
}

// SetSpellButtonsVisible controls visibility of spell buttons.
// Platform parity fix: Allows hiding spell buttons when not in combat.
func (l *VirtualControlsLayout) SetSpellButtonsVisible(visible bool) {
	l.ShowSpellButtons = visible
}

// SetUIShortcutsVisible controls visibility of UI shortcut buttons.
// Platform parity fix: Allows toggling UI shortcuts visibility for cleaner interface.
func (l *VirtualControlsLayout) SetUIShortcutsVisible(visible bool) {
	l.ShowUIShortcuts = visible
}

// Resize recalculates control positions for new screen dimensions.
// Platform parity fix: Handles orientation changes on mobile devices.
func (l *VirtualControlsLayout) Resize(screenWidth, screenHeight int) {
	l.screenWidth = screenWidth
	l.screenHeight = screenHeight

	sizes := calculateControlSizes(screenHeight)
	l.resizeCoreControls(screenWidth, screenHeight, sizes)
	l.resizeExtendedControls(sizes)
	l.resizeUIShortcutButtons(sizes)
	l.resizeSpellButtons(screenWidth, sizes)
}

// calculateControlSizes computes the size dimensions for all control elements.
func calculateControlSizes(screenHeight int) controlSizes {
	return controlSizes{
		dpadSize:        float64(screenHeight) * 0.15,
		buttonSize:      float64(screenHeight) * 0.08,
		smallButtonSize: float64(screenHeight) * 0.06,
		margin:          float64(screenHeight) * 0.05,
	}
}

// controlSizes holds the calculated size dimensions for controls.
type controlSizes struct {
	dpadSize        float64
	buttonSize      float64
	smallButtonSize float64
	margin          float64
}

// resizeCoreControls repositions the D-pad and primary action buttons.
func (l *VirtualControlsLayout) resizeCoreControls(screenWidth, screenHeight int, sizes controlSizes) {
	l.DPad.X = sizes.margin + sizes.dpadSize
	l.DPad.Y = float64(screenHeight) - sizes.margin - sizes.dpadSize
	l.DPad.Radius = sizes.dpadSize

	l.ActionButton.X = float64(screenWidth) - sizes.margin - sizes.buttonSize*2.5
	l.ActionButton.Y = float64(screenHeight) - sizes.margin - sizes.buttonSize
	l.ActionButton.Radius = sizes.buttonSize

	l.SecondaryButton.X = float64(screenWidth) - sizes.margin - sizes.buttonSize
	l.SecondaryButton.Y = float64(screenHeight) - sizes.margin - sizes.buttonSize*2.5
	l.SecondaryButton.Radius = sizes.buttonSize

	l.MenuButton.X = float64(screenWidth) - sizes.margin - sizes.buttonSize
	l.MenuButton.Y = sizes.margin + sizes.buttonSize
	l.MenuButton.Radius = sizes.buttonSize * 0.7
}

// resizeExtendedControls repositions the inventory, target, and interact buttons.
func (l *VirtualControlsLayout) resizeExtendedControls(sizes controlSizes) {
	if l.InventoryButton != nil {
		l.InventoryButton.X = l.MenuButton.X - sizes.buttonSize*1.3
		l.InventoryButton.Y = l.MenuButton.Y
		l.InventoryButton.Radius = sizes.buttonSize * 0.7
	}

	if l.TargetButton != nil {
		l.TargetButton.X = l.DPad.X
		l.TargetButton.Y = l.DPad.Y - sizes.dpadSize - sizes.smallButtonSize - sizes.margin*0.5
		l.TargetButton.Radius = sizes.smallButtonSize
	}

	if l.InteractButton != nil {
		l.InteractButton.X = l.ActionButton.X - sizes.buttonSize*1.5
		l.InteractButton.Y = l.ActionButton.Y
		l.InteractButton.Radius = sizes.smallButtonSize
	}
}

// resizeUIShortcutButtons repositions the character, skills, quest, and map buttons.
func (l *VirtualControlsLayout) resizeUIShortcutButtons(sizes controlSizes) {
	uiButtonY := sizes.margin + sizes.buttonSize*0.7
	uiButtonSpacing := sizes.smallButtonSize * 1.3
	characterX := sizes.margin + sizes.smallButtonSize
	skillsX := characterX + uiButtonSpacing
	questLogX := skillsX + uiButtonSpacing
	mapX := questLogX + uiButtonSpacing

	if l.CharacterButton != nil {
		l.CharacterButton.X = characterX
		l.CharacterButton.Y = uiButtonY
		l.CharacterButton.Radius = sizes.smallButtonSize
	}
	if l.SkillsButton != nil {
		l.SkillsButton.X = skillsX
		l.SkillsButton.Y = uiButtonY
		l.SkillsButton.Radius = sizes.smallButtonSize
	}
	if l.QuestLogButton != nil {
		l.QuestLogButton.X = questLogX
		l.QuestLogButton.Y = uiButtonY
		l.QuestLogButton.Radius = sizes.smallButtonSize
	}
	if l.MapButton != nil {
		l.MapButton.X = mapX
		l.MapButton.Y = uiButtonY
		l.MapButton.Radius = sizes.smallButtonSize
	}
}

// resizeSpellButtons repositions all spell casting buttons.
func (l *VirtualControlsLayout) resizeSpellButtons(screenWidth int, sizes controlSizes) {
	if l.SpellButtons == nil {
		return
	}
	spellButtonY := l.ActionButton.Y - sizes.buttonSize*1.8
	for i := range l.SpellButtons {
		if l.SpellButtons[i] != nil {
			l.SpellButtons[i].X = float64(screenWidth)/2 - sizes.smallButtonSize*3 + float64(i)*sizes.smallButtonSize*1.3
			l.SpellButtons[i].Y = spellButtonY
			l.SpellButtons[i].Radius = sizes.smallButtonSize
		}
	}
}

// Platform parity fix: Cancel/undo gesture patterns for consistent UX

// CancelGesture type moved to types.go

// SelectionState tracks selection state across input methods.
// Platform parity fix: Unified selection/deselection for mouse/touch/keyboard
type SelectionState struct {
	selectedItems map[int]bool
	lastSelected  int
	multiSelect   bool
}

// NewSelectionState creates a new selection state tracker.
func NewSelectionState(multiSelect bool) *SelectionState {
	return &SelectionState{
		selectedItems: make(map[int]bool),
		lastSelected:  -1,
		multiSelect:   multiSelect,
	}
}

// Select selects an item.
// Platform parity fix: Works for click, tap, or keyboard selection
func (s *SelectionState) Select(itemID int, isMultiSelectGesture bool) {
	if s.multiSelect && isMultiSelectGesture {
		// Platform parity fix: Multi-select (Ctrl+Click, Shift+Tap, etc.)
		s.selectedItems[itemID] = true
	} else {
		// Platform parity fix: Single select - clear others
		s.selectedItems = make(map[int]bool)
		s.selectedItems[itemID] = true
	}
	s.lastSelected = itemID
}

// Deselect deselects an item.
// Platform parity fix: Works for click-away, tap-away, or keyboard deselection
func (s *SelectionState) Deselect(itemID int) {
	delete(s.selectedItems, itemID)
}

// DeselectAll clears all selections.
// Platform parity fix: Escape key, tap empty space, click away all do this
func (s *SelectionState) DeselectAll() {
	s.selectedItems = make(map[int]bool)
	s.lastSelected = -1
}

// IsSelected checks if an item is selected.
func (s *SelectionState) IsSelected(itemID int) bool {
	return s.selectedItems[itemID]
}

// GetSelectedItems returns all selected item IDs.
func (s *SelectionState) GetSelectedItems() []int {
	items := make([]int, 0, len(s.selectedItems))
	for id := range s.selectedItems {
		items = append(items, id)
	}
	return items
}

// GetSelectionCount returns number of selected items.
func (s *SelectionState) GetSelectionCount() int {
	return len(s.selectedItems)
}
