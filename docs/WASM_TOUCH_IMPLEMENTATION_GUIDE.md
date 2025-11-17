# WASM Touch Implementation Guide

## Overview

This document provides implementation patterns for adding touch support to UI systems in the Venture game. These patterns have been successfully applied to 12+ UI systems and ensure consistent mobile usability.

## Core Components

### 1. TouchInputHandler
**Location:** `pkg/mobile/touch.go`

```go
// Initialize in UI struct
touchHandler *mobile.TouchInputHandler

// In constructor
touchHandler: mobile.NewTouchInputHandler(),

// In Update() method
if ui.touchHandler != nil {
    ui.touchHandler.Update()
}

// For gesture detection
if direction, distance, detected := ui.touchHandler.GetSwipe(); detected {
    // Handle swipe for scrolling
}
```

### 2. TouchButton
**Location:** `pkg/mobile/ui.go`

**Minimum Size:** 44x44px (iOS/Android HIG compliance)

```go
// Create button
closeButton := mobile.NewTouchButton(
    float64(x), float64(y),  // Position
    44, 44,                   // Size (minimum 44x44)
    "✕",                      // Label (or icon)
    func() { ui.Hide() },     // OnTap callback
)

// Update button every frame
closeButton.Update()

// Draw button
closeButton.Draw(screen)

// Dynamic positioning (for responsive layout)
closeButton.SetPosition(float64(newX), float64(newY))
```

### 3. VirtualControlsLayout
**Location:** `pkg/mobile/controls.go`

**Already Integrated:** Virtual D-pad + action buttons + menu button managed by InputSystem.

```go
// Initialization (done automatically for mobile/WASM)
if mobile.IsTouchCapable() {
    inputSystem.InitializeVirtualControls(screenWidth, screenHeight)
}

// Virtual controls include:
// - D-pad (bottom-left) - Movement
// - A/B buttons (bottom-right) - Actions
// - Menu button (☰) (top-right) - Opens pause menu
```

### 4. Virtual Keyboard (WASM Only)
**Location:** `pkg/mobile/keyboard_wasm.go`

**Build Tags:** Only compiled for WASM (`//go:build js`)

```go
// Show keyboard for text input
mobile.ShowKeyboard()

// Hide keyboard when done
mobile.HideKeyboard()

// Platform check
if mobile.IsKeyboardSupported() {
    // Virtual keyboard available
}
```

## Implementation Patterns

### Pattern 1: Simple Menu with Close Button

**Example:** Inventory UI, Character Stats UI

```go
type MyMenuUI struct {
    visible      bool
    touchHandler *mobile.TouchInputHandler
    closeButton  *mobile.TouchButton
    screenWidth  int
    screenHeight int
}

func NewMyMenuUI(screenWidth, screenHeight int) *MyMenuUI {
    ui := &MyMenuUI{
        touchHandler: mobile.NewTouchInputHandler(),
        screenWidth:  screenWidth,
        screenHeight: screenHeight,
    }
    
    // Position close button in top-right of window
    ui.closeButton = mobile.NewTouchButton(
        float64(windowX+windowWidth-54),
        float64(windowY+10),
        44, 44,
        "✕",
        func() { ui.Hide() },
    )
    
    return ui
}

func (ui *MyMenuUI) Update() {
    // Update touch handler
    if ui.touchHandler != nil {
        ui.touchHandler.Update()
    }
    
    // Update close button
    if ui.closeButton != nil {
        ui.closeButton.Update()
    }
    
    // Handle keyboard shortcuts (preserved for desktop)
    if shouldClose, _ := HandleMenuInput(MenuKeys.MyMenu, ui.visible); shouldClose {
        ui.Hide()
        return
    }
}

func (ui *MyMenuUI) Draw(screen *ebiten.Image) {
    if !ui.visible {
        return
    }
    
    // Draw menu content...
    
    // Update button position (for dynamic layout)
    ui.closeButton.SetPosition(
        float64(windowX+windowWidth-54),
        float64(windowY+10),
    )
    
    // Draw close button
    ui.closeButton.Draw(screen)
}
```

### Pattern 2: Menu with Tabs

**Example:** Quest Log UI

```go
type TabMenuUI struct {
    visible       bool
    touchHandler  *mobile.TouchInputHandler
    closeButton   *mobile.TouchButton
    tab1Button    *mobile.TouchButton
    tab2Button    *mobile.TouchButton
    currentTab    int
}

func NewTabMenuUI() *TabMenuUI {
    ui := &TabMenuUI{
        touchHandler: mobile.NewTouchInputHandler(),
        currentTab:   0,
    }
    
    // Close button (top-right)
    ui.closeButton = mobile.NewTouchButton(
        100, 10, 44, 44, "✕",
        func() { ui.Hide() },
    )
    
    // Tab buttons (120x44 minimum for tab labels)
    ui.tab1Button = mobile.NewTouchButton(
        50, 50, 120, 44, "Tab 1",
        func() { ui.currentTab = 0 },
    )
    
    ui.tab2Button = mobile.NewTouchButton(
        180, 50, 120, 44, "Tab 2",
        func() { ui.currentTab = 1 },
    )
    
    return ui
}

func (ui *TabMenuUI) Update() {
    ui.touchHandler.Update()
    ui.closeButton.Update()
    ui.tab1Button.Update()
    ui.tab2Button.Update()
    
    // Keyboard shortcuts (1/2 keys for tabs on desktop)
    if inpututil.IsKeyJustPressed(ebiten.Key1) {
        ui.currentTab = 0
    }
    if inpututil.IsKeyJustPressed(ebiten.Key2) {
        ui.currentTab = 1
    }
}

func (ui *TabMenuUI) Draw(screen *ebiten.Image) {
    // Draw tabs with visual active state
    ui.tab1Button.BackgroundColor = getTabColor(ui.currentTab == 0)
    ui.tab2Button.BackgroundColor = getTabColor(ui.currentTab == 1)
    
    ui.tab1Button.Draw(screen)
    ui.tab2Button.Draw(screen)
    ui.closeButton.Draw(screen)
}

func getTabColor(active bool) color.Color {
    if active {
        return color.RGBA{80, 120, 200, 255} // Highlighted
    }
    return color.RGBA{50, 50, 70, 255} // Normal
}
```

### Pattern 3: Scrollable Menu

**Example:** Help System, Settings UI

```go
type ScrollableMenuUI struct {
    touchHandler *mobile.TouchInputHandler
    closeButton  *mobile.TouchButton
    scrollOffset float64
}

func (ui *ScrollableMenuUI) Update() {
    ui.touchHandler.Update()
    
    // Handle swipe for scrolling
    if direction, distance, detected := ui.touchHandler.GetSwipe(); detected {
        // Vertical swipe (direction > 1.0 or < -1.0)
        if direction > 1.0 || direction < -1.0 {
            if direction < 0 {
                ui.scrollOffset += distance * 0.5  // Swipe up
            } else {
                ui.scrollOffset -= distance * 0.5  // Swipe down
            }
            // Clamp scroll offset
            if ui.scrollOffset < 0 {
                ui.scrollOffset = 0
            }
        }
    }
}
```

### Pattern 4: Text Input with Virtual Keyboard

**Example:** Character Creation, Server Address Input

```go
type TextInputUI struct {
    inputBuffer   []rune
    keyboardShown bool
    inputActive   bool
}

func (ui *TextInputUI) Update() {
    // Show keyboard when input field is focused
    if ui.inputActive && !ui.keyboardShown {
        mobile.ShowKeyboard()
        ui.keyboardShown = true
    }
    
    // Capture keyboard input
    ui.inputBuffer = ebiten.AppendInputChars(ui.inputBuffer[:0])
    
    // Process input
    for _, ch := range ui.inputBuffer {
        if ch >= 32 && ch <= 126 {
            ui.text += string(ch)
        }
    }
    
    // Handle backspace
    if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(ui.text) > 0 {
        ui.text = ui.text[:len(ui.text)-1]
    }
    
    // Handle submit (Enter key)
    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        mobile.HideKeyboard()
        ui.keyboardShown = false
        ui.inputActive = false
        ui.handleSubmit()
    }
}
```

### Pattern 5: Dialog with Touch Choices

**Example:** NPC Dialog System (IMPLEMENTATION NEEDED)

```go
type DialogUI struct {
    visible       bool
    dialogText    string
    options       []DialogOption
    touchHandler  *mobile.TouchInputHandler
    choiceButtons []*mobile.TouchButton
}

func (ui *DialogUI) SetDialog(text string, options []DialogOption) {
    ui.dialogText = text
    ui.options = options
    
    // Create touch button for each choice
    ui.choiceButtons = make([]*mobile.TouchButton, len(options))
    for i, option := range options {
        btnLabel := fmt.Sprintf("%d. %s", i+1, option.Text)
        btnY := 200 + i*50 // Vertical spacing
        
        ui.choiceButtons[i] = mobile.NewTouchButton(
            100, float64(btnY),
            400, 44, // Wide button for text
            btnLabel,
            func(idx int) func() {
                return func() { ui.selectChoice(idx) }
            }(i),
        )
        
        // Disable button if option is disabled
        ui.choiceButtons[i].Enabled = option.Enabled
    }
}

func (ui *DialogUI) Update() {
    ui.touchHandler.Update()
    
    // Update all choice buttons
    for _, btn := range ui.choiceButtons {
        btn.Update()
    }
    
    // Keyboard shortcuts (1-9 keys for desktop)
    for i := 0; i < len(ui.options) && i < 9; i++ {
        key := ebiten.Key(ebiten.Key1 + i)
        if inpututil.IsKeyJustPressed(key) && ui.options[i].Enabled {
            ui.selectChoice(i)
        }
    }
}

func (ui *DialogUI) Draw(screen *ebiten.Image) {
    if !ui.visible {
        return
    }
    
    // Draw dialog text...
    
    // Draw choice buttons
    for _, btn := range ui.choiceButtons {
        btn.Draw(screen)
    }
}

func (ui *DialogUI) selectChoice(index int) {
    if index >= 0 && index < len(ui.options) && ui.options[index].Enabled {
        // Process dialog choice
        action := ui.options[index].Action
        // Handle action...
    }
}
```

## Best Practices

### 1. Touch Target Sizes
- **Minimum:** 44x44px (iOS/Android guideline)
- **Recommended:** 48x48px or larger for frequently used controls
- **Spacing:** Minimum 8px between tappable elements

### 2. Button Placement
- **Close buttons:** Top-right corner of windows
- **Navigation buttons:** Bottom of screen (easier thumb reach)
- **Primary actions:** Bottom-right (right-handed accessibility)
- **Secondary/Cancel:** Bottom-left

### 3. Visual Feedback
- **Pressed state:** Color change when button is tapped
- **Disabled state:** Grayed out when not available
- **Hover state:** Not needed for touch (but keep for mouse on desktop)

### 4. Keyboard Compatibility
- **Always preserve keyboard shortcuts** for desktop users
- Use `HandleMenuInput()` for standard menu navigation
- Keep number keys (1-9) for quick dialog choices on desktop
- Keep arrow keys for list navigation on desktop

### 5. Virtual Controls Visibility
- Hide virtual D-pad and buttons when menus are open
- Use `InputSystem.SetVirtualControlsVisible(false)` in menu `Show()`
- Restore visibility in menu `Hide()`

### 6. Platform Detection
```go
// Check if platform supports touch
if mobile.IsTouchCapable() {
    // Initialize touch-specific UI
}

// Check if virtual keyboard is available (WASM only)
if mobile.IsKeyboardSupported() {
    // Show keyboard for text input
}
```

## Testing Checklist

### Per-UI System Test
- [ ] Menu opens via touch button (not keyboard only)
- [ ] Menu closes via touch button or ESC key
- [ ] All interactive elements ≥44x44px
- [ ] Buttons have visual pressed state
- [ ] Swipe gestures work for scrolling (if applicable)
- [ ] Text input shows virtual keyboard (if applicable)
- [ ] Keyboard shortcuts still work on desktop
- [ ] Virtual controls hide when menu is shown

### Full Game Test (Mobile Browser)
- [ ] Complete game playthrough with zero physical keyboard
- [ ] Character creation via touch + virtual keyboard
- [ ] All menus accessible via touch
- [ ] NPC interactions work via touch
- [ ] Combat via virtual controls
- [ ] Save/load via touch

## Common Issues and Solutions

### Issue: Button not responding to touch
**Solution:** Ensure `Update()` is called every frame and TouchInputHandler is initialized.

### Issue: Virtual keyboard doesn't appear
**Solution:** Check build tags (`//go:build js`). Virtual keyboard only works in WASM builds.

### Issue: Touch and mouse both trigger
**Solution:** Use `GetTouchOrMousePosition()` and `IsTouchOrMouseJustPressed()` helpers.

### Issue: Buttons too small on mobile
**Solution:** Enforce 44x44px minimum via `NewTouchButton()` (already enforced in constructor).

### Issue: Virtual controls overlap menu
**Solution:** Call `SetVirtualControlsVisible(false)` when menu opens.

## References

- [iOS Human Interface Guidelines - Touch Targets](https://developer.apple.com/design/human-interface-guidelines/inputs/touchscreen-gestures)
- [Android Material Design - Touch Targets](https://material.io/design/usability/accessibility.html#layout-and-typography)
- [WCAG 2.1 - Target Size](https://www.w3.org/WAI/WCAG21/Understanding/target-size.html)

## Examples in Codebase

**Complete Implementations:**
- `pkg/engine/inventory_ui.go` - Close button + swipe scroll
- `pkg/engine/quest_ui.go` - Tab buttons + close button
- `pkg/engine/tutorial_system.go` - Next/Skip buttons
- `pkg/engine/help_system.go` - Close button + swipe scroll
- `pkg/engine/character_creation.go` - Virtual keyboard + touch buttons

**Touch Components:**
- `pkg/mobile/touch.go` - TouchInputHandler, gesture detection
- `pkg/mobile/ui.go` - TouchButton component
- `pkg/mobile/controls.go` - VirtualControlsLayout
- `pkg/mobile/keyboard_wasm.go` - Virtual keyboard (WASM)

---

**Last Updated:** 2025-11-17  
**Version:** 1.0  
**Status:** Production Ready
