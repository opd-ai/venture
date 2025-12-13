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
// ... (additional details omitted for brevity)
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
// ... (additional details omitted for brevity)
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
// ... (additional details omitted for brevity)
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
// ... (additional details omitted for brevity)
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
// ... (additional details omitted for brevity)
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
// ... (additional details omitted for brevity)
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
    
// ... (additional details omitted for brevity)
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
