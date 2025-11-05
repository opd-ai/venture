# Mobile WASM Keyboard Quick Reference

## For Developers: Adding Text Input Support

If you need to add a new text input field to Venture, follow this pattern to ensure it works on mobile WASM:

### 1. Import the mobile package

```go
import "github.com/opd-ai/venture/pkg/mobile"
```

### 2. Add keyboard state tracking

```go
type YourInputComponent struct {
    // ... other fields ...
    keyboardShown bool // Tracks whether mobile keyboard is currently shown
}
```

### 3. Show keyboard when input becomes active

```go
func (c *YourInputComponent) Update() {
    // MOBILE/WASM: Show keyboard when entering text input
    if !c.keyboardShown && mobile.IsWASM() {
        mobile.ShowKeyboard()
        c.keyboardShown = true
    }
    
    // Your existing input handling code...
    inputChars := ebiten.AppendInputChars(nil)
    // ... process input ...
}
```

### 4. Hide keyboard when input completes or is cancelled

```go
// On Enter/Submit
if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
    // MOBILE/WASM: Hide keyboard when completing
    if c.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        c.keyboardShown = false
    }
    // ... process submission ...
}

// On Escape/Cancel
if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
    // MOBILE/WASM: Hide keyboard when cancelling
    if c.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        c.keyboardShown = false
    }
    // ... cancel action ...
}
```

### 5. Reset keyboard state when transitioning away

```go
func (c *YourInputComponent) Reset() {
    // MOBILE/WASM: Ensure keyboard is hidden when resetting
    if c.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        c.keyboardShown = false
    }
    // ... other reset logic ...
}
```

## API Reference

### Functions

#### `mobile.ShowKeyboard()`
- **Purpose**: Displays the native mobile keyboard
- **Platform**: WASM only (no-op on desktop/native)
- **When to call**: When entering a text input state
- **Safe to call**: Multiple times (idempotent on WASM)

#### `mobile.HideKeyboard()`
- **Purpose**: Dismisses the native mobile keyboard
- **Platform**: WASM only (no-op on desktop/native)
- **When to call**: When leaving a text input state
- **Safe to call**: Multiple times (idempotent on WASM)

#### `mobile.IsWASM()`
- **Purpose**: Checks if running in WebAssembly/browser environment
- **Returns**: `true` on WASM, `false` on desktop/native mobile
- **Use case**: Conditional keyboard logic

### Example: Simple Text Input

```go
package engine

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/opd-ai/venture/pkg/mobile"
)

type SimpleTextInput struct {
    text          string
    active        bool
    keyboardShown bool
}

func (s *SimpleTextInput) Activate() {
    s.active = true
    // Show keyboard on activation
    if mobile.IsWASM() {
        mobile.ShowKeyboard()
        s.keyboardShown = true
    }
}

func (s *SimpleTextInput) Deactivate() {
    s.active = false
    // Hide keyboard on deactivation
    if s.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        s.keyboardShown = false
    }
}

func (s *SimpleTextInput) Update() {
    if !s.active {
        return
    }
    
    // Handle text input
    chars := ebiten.AppendInputChars(nil)
    for _, r := range chars {
        s.text += string(r)
    }
    
    // Handle backspace
    if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(s.text) > 0 {
        s.text = s.text[:len(s.text)-1]
    }
    
    // Submit on Enter
    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        s.Deactivate()
        // Process s.text...
    }
    
    // Cancel on Escape
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        s.Deactivate()
    }
}
```

## Testing Checklist

When implementing new text input:

- [ ] Import `pkg/mobile` package
- [ ] Add `keyboardShown bool` field
- [ ] Call `mobile.ShowKeyboard()` when input becomes active
- [ ] Call `mobile.HideKeyboard()` on completion/cancellation
- [ ] Guard calls with `mobile.IsWASM()` check
- [ ] Test on desktop (should work unchanged)
- [ ] Test WASM build compiles
- [ ] Test on real mobile device if possible

## Common Patterns

### Pattern 1: Modal Text Input
```go
// Show keyboard when modal opens
func (m *Modal) Show() {
    m.visible = true
    if mobile.IsWASM() {
        mobile.ShowKeyboard()
        m.keyboardShown = true
    }
}

// Hide keyboard when modal closes
func (m *Modal) Hide() {
    m.visible = false
    if m.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        m.keyboardShown = false
    }
}
```

### Pattern 2: Multi-step Input
```go
// Reset keyboard state between steps
func (w *Wizard) NextStep() {
    // Hide keyboard when leaving text input step
    if w.currentStep == stepTextInput && w.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        w.keyboardShown = false
    }
    
    w.currentStep++
    
    // Show keyboard when entering new text input step
    if w.currentStep == stepAnotherTextInput && mobile.IsWASM() {
        mobile.ShowKeyboard()
        w.keyboardShown = true
    }
}
```

### Pattern 3: Focus Management
```go
// Track which input has focus
type Form struct {
    inputs        []*TextInput
    focusedIndex  int
    keyboardShown bool
}

func (f *Form) SetFocus(index int) {
    // No change needed
    if f.focusedIndex == index {
        return
    }
    
    // Hide keyboard if it was shown
    if f.keyboardShown && mobile.IsWASM() {
        mobile.HideKeyboard()
        f.keyboardShown = false
    }
    
    f.focusedIndex = index
    
    // Show keyboard for new focused input
    if mobile.IsWASM() {
        mobile.ShowKeyboard()
        f.keyboardShown = true
    }
}
```

## Troubleshooting

### Keyboard doesn't appear on mobile
- ✓ Check `mobile.IsWASM()` returns true (use console.log in JS to verify)
- ✓ Verify `ShowKeyboard()` is called when input becomes active
- ✓ Ensure WASM build includes keyboard_wasm.go (check build tags)
- ✓ Test on real device (emulators may behave differently)

### Keyboard doesn't dismiss
- ✓ Check `HideKeyboard()` is called on input completion
- ✓ Verify state transitions properly clear `keyboardShown` flag
- ✓ Check for early returns that skip keyboard hide logic

### Keyboard appears at wrong time
- ✓ Review state management logic
- ✓ Ensure `keyboardShown` flag is reset on transitions
- ✓ Check for multiple components showing keyboard simultaneously

### Desktop builds broken
- ✓ Verify `mobile.IsWASM()` guards keyboard calls
- ✓ Check build compiles without WASM-specific code
- ✓ Ensure no WASM-only imports in desktop code paths

## Build Tags Reference

### keyboard_wasm.go
```go
//go:build js
// +build js

// WASM-specific implementation with syscall/js
```

### keyboard_default.go
```go
//go:build !js
// +build !js

// No-op implementation for desktop/native
```

### Your Code (No Build Tags Needed)
```go
package engine

import "github.com/opd-ai/venture/pkg/mobile"

// Works on all platforms
// mobile.ShowKeyboard() is no-op on desktop
```

## Performance Notes

- `ShowKeyboard()` and `HideKeyboard()` are fast (~0.01ms)
- State tracking adds 1 byte per component (negligible)
- No performance impact on desktop (no-op functions optimized away)
- Safe to call in Update loop (checks prevent redundant operations)

## See Also

- `MOBILE_WASM_KEYBOARD_FIX.md` - Full implementation documentation
- `pkg/mobile/keyboard_wasm.go` - WASM implementation source
- `pkg/engine/character_creation.go` - Reference implementation
- `pkg/engine/server_address_input.go` - Reference implementation
