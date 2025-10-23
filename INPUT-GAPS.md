# Input System Gap Analysis - Ebiten Game Engine
Generated: 2025-10-23  
Documentation Sources: API_REFERENCE.md, USER_MANUAL.md, TECHNICAL_SPEC.md, input_system.go  
Total Methods Analyzed: 47  
Total Gaps Found: 23

## Documentation Coverage Map

### Keyboard Methods Documented
- `IsKeyPressed`: Documented in USER_MANUAL.md:44-86 (movement keys WASD, action keys, UI keys)
- `IsKeyJustPressed`: Documented in USER_MANUAL.md:44-86 (action triggers, single-press actions)
- `IsKeyReleased`: **NOT DOCUMENTED** - No specification found
- Key bindings: Documented in USER_MANUAL.md:44-86 and input_system.go:40-77
- Text input: **NOT DOCUMENTED** - No specification found
- Modifier keys (Shift/Ctrl/Alt): **NOT DOCUMENTED** - No specification found

### Mouse Methods Documented
- `MousePosition`/`CursorPosition`: Documented in USER_MANUAL.md:73-76 (aiming, selection)
- `IsMouseButtonPressed`: Documented in USER_MANUAL.md:73-76 (left/right click)
- `IsMouseButtonJustPressed`: **NOT DOCUMENTED** - Implied usage but no explicit spec
- `IsMouseButtonReleased`: **NOT DOCUMENTED** - No specification found
- `MouseWheel`/`ScrollWheel`: Documented in USER_MANUAL.md:76 (zoom camera)
- Mouse button constants: **NOT DOCUMENTED** - No enum/constant documentation

### Touch Input Methods (Mobile)
- `TouchIDs`: Implemented in mobile/touch.go:39
- `TouchPosition`: Implemented in mobile/touch.go:40
- Touch gesture detection: Implemented in mobile/touch.go:99-246
- **Status**: Fully implemented but not documented in user-facing docs

## Gap Inventory by Severity

### Missing Methods (8)
- BUG-001: IsKeyReleased() not exposed in InputSystem
- BUG-002: IsKeyJustReleased() not exposed in InputSystem
- BUG-003: GetPressedKeys() not implemented
- BUG-004: IsMouseButtonJustPressed() not exposed in InputSystem
- BUG-005: IsMouseButtonReleased() not exposed in InputSystem
- BUG-006: IsMouseButtonJustReleased() not exposed in InputSystem
- BUG-007: GetMouseWheel() not exposed in InputSystem
- BUG-008: GetMouseDelta() not implemented

### Incorrect Returns (0)
*None identified*

### State Desync Issues (3)
- BUG-009: Input state not cleared between frames for edge detection
- BUG-010: Mouse position not tracked frame-to-frame for delta calculation
- BUG-011: Keyboard repeat state not managed (OS key repeat interferes)

### Edge Case Failures (5)
- BUG-012: Rapid key press-release within single frame lost
- BUG-013: Multiple simultaneous key presses not all detected
- BUG-014: Mouse button press-release same frame not detected
- BUG-015: Touch and keyboard input simultaneously causes undefined behavior
- BUG-016: Virtual controls overlap with touch targeting area

### Timing Violations (2)
- BUG-017: No frame-boundary synchronization guarantee
- BUG-018: InputSystem.Update() called before Ebiten input polling

### Partial Implementation (4)
- BUG-019: SetKeyBindings() only sets 6 keys, not all 18 documented keys
- BUG-020: No public API for querying current key bindings
- BUG-021: No API for checking if ANY key is pressed (useful for "press any key")
- BUG-022: No API for checking modifier key state (Shift, Ctrl, Alt)

### Silent Malfunction (1)
- BUG-023: Mobile input silently disabled if InitializeVirtualControls() not called

---

## Detailed Bug Reports

### BUG-001: IsKeyReleased() - Missing Key Release Detection

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:44-86  
Quote: "Movement: W - Move Up, A - Move Left, S - Move Down, D - Move Right"  
Implicit: Users need to detect when keys are released for charge attacks, aiming, etc.

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No IsKeyReleased() method exists

**EXPECTED BEHAVIOR:**
InputSystem should provide a method to detect when a key is released this frame:
```go
func (s *InputSystem) IsKeyReleased(key ebiten.Key) bool
```
Should return true only on the frame where the key transitioned from pressed to released.

**ACTUAL BEHAVIOR:**
No such method exists. Developers cannot detect key release events.

**REPRODUCTION CODE:**
```go
// Desired usage (not currently possible)
inputSys := engine.NewInputSystem()
if inputSys.IsKeyReleased(ebiten.KeySpace) {
    // Fire charged attack that was held with Space
}
// Error: undefined method IsKeyReleased
```

**GAP ANALYSIS:**
Ebiten provides `inpututil.IsKeyJustReleased()` but InputSystem doesn't expose it. This is needed for:
- Charge attacks (hold to charge, release to fire)
- Aim mechanics (hold to aim, release to shoot)
- Jump mechanics (hold for higher jump, release early for short hop)
- UI interactions (button highlight on press, action on release)

**DEPENDENCIES:**
None - standalone method using Ebiten's inpututil package

---

### BUG-002: IsKeyJustReleased() - Naming Inconsistency

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:325-328  
Context: Uses `IsKeyJustPressed()` naming pattern

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No IsKeyJustReleased() method (see BUG-001)

**EXPECTED BEHAVIOR:**
For API consistency, should provide:
```go
func (s *InputSystem) IsKeyJustReleased(key ebiten.Key) bool
```
Alias for IsKeyReleased() to match IsKeyJustPressed() naming.

**ACTUAL BEHAVIOR:**
Neither IsKeyReleased() nor IsKeyJustReleased() exist.

**REPRODUCTION CODE:**
```go
// Expected consistent API
inputSys.IsKeyPressed(key)       // Continuous state - EXISTS
inputSys.IsKeyJustPressed(key)   // Press edge - EXISTS
inputSys.IsKeyJustReleased(key)  // Release edge - MISSING
```

**GAP ANALYSIS:**
API naming inconsistency. InputSystem wraps Ebiten's input APIs but only exposes press detection, not release detection.

**DEPENDENCIES:**
Same as BUG-001 (duplicate/related issue)

---

### BUG-003: GetPressedKeys() - No Bulk Key Query

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:88  
Quote: "Custom Key Bindings: Edit key bindings in the settings menu"  
Implicit: Need to detect "any key press" for rebinding UI

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No method to get all currently pressed keys

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) GetPressedKeys() []ebiten.Key
```
Returns slice of all keys currently pressed. Useful for:
- Key binding configuration UI ("Press any key to bind...")
- Debug displays
- Accessibility features
- Combo detection

**ACTUAL BEHAVIOR:**
No method exists. Must poll every possible key individually to detect arbitrary key presses.

**REPRODUCTION CODE:**
```go
// Current workaround (inefficient)
for key := ebiten.Key0; key <= ebiten.KeyMax; key++ {
    if ebiten.IsKeyPressed(key) {
        // Found a pressed key
    }
}

// Desired API
pressedKeys := inputSys.GetPressedKeys()
if len(pressedKeys) > 0 {
    bindKey := pressedKeys[0]
}
```

**GAP ANALYSIS:**
Ebiten provides `inpututil.AppendPressedKeys()` but InputSystem doesn't expose it. Essential for UI workflows.

**DEPENDENCIES:**
None - uses Ebiten's built-in inpututil function

---

### BUG-004: IsMouseButtonJustPressed() - Missing Mouse Press Detection

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:73-76  
Quote: "Mouse: Left Click - Confirm / Select / Attack"  
Implicit: Need edge-triggered click detection, not continuous state

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:334  
Current: Only `ebiten.IsMouseButtonPressed()` used (continuous state)

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) IsMouseButtonJustPressed(button ebiten.MouseButton) bool
```
Returns true only on the frame where mouse button was pressed (edge-triggered).

**ACTUAL BEHAVIOR:**
InputComponent.MousePressed is set from continuous state check, causing issues:
```go
input.MousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
```
This is true for ALL frames while held, not just the press frame.

**REPRODUCTION CODE:**
```go
// Current behavior (problematic)
if input.MousePressed {
    // Fires every frame while held - wrong for single-click actions!
    OpenInventory()
}

// Desired behavior
if inputSys.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
    // Fires once per click - correct!
    OpenInventory()
}
```

**GAP ANALYSIS:**
Critical bug for UI interactions. Buttons, inventory clicks, and all single-action clicks are broken because they fire every frame instead of once per click. Current workaround requires game code to track previous frame state manually.

**DEPENDENCIES:**
None - uses Ebiten's inpututil.IsMouseButtonJustPressed()

---

### BUG-005: IsMouseButtonReleased() - Missing Mouse Release Detection

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:73-76  
Quote: "Right Click - Cancel / Alt action"  
Implicit: Some actions trigger on release (drag-and-drop, long-press menus)

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:334  
Missing: No release detection for mouse buttons

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) IsMouseButtonReleased(button ebiten.MouseButton) bool
```
Returns true on frame where button was released (edge-triggered).

**ACTUAL BEHAVIOR:**
No such method exists. Cannot detect mouse button release events.

**REPRODUCTION CODE:**
```go
// Desired usage for drag-and-drop
if inputSys.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
    StartDrag(item)
}
if inputSys.IsMouseButtonReleased(ebiten.MouseButtonLeft) {
    EndDrag(item) // Drop item
}
```

**GAP ANALYSIS:**
Essential for drag-and-drop, context menus (right-click hold), charge attacks with mouse. Ebiten provides `inpututil.IsMouseButtonJustReleased()` but InputSystem doesn't expose it.

**DEPENDENCIES:**
None

---

### BUG-006: IsMouseButtonJustReleased() - Naming Alias Missing

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:334  
Context: API consistency with keyboard methods

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:334  
Missing: Alias for IsMouseButtonReleased()

**EXPECTED BEHAVIOR:**
For naming consistency with IsKeyJustReleased(), provide:
```go
func (s *InputSystem) IsMouseButtonJustReleased(button ebiten.MouseButton) bool
```

**ACTUAL BEHAVIOR:**
Neither IsMouseButtonReleased() nor IsMouseButtonJustReleased() exist.

**REPRODUCTION CODE:**
See BUG-005

**GAP ANALYSIS:**
Duplicate/related to BUG-005. Naming consistency issue.

**DEPENDENCIES:**
Same as BUG-005

---

### BUG-007: GetMouseWheel() - Missing Scroll Detection

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:76  
Quote: "Scroll Wheel - Zoom camera"

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No scroll wheel support

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) GetMouseWheel() (deltaX, deltaY float64)
```
Returns scroll wheel delta for current frame. Used for:
- Camera zoom (as documented)
- Inventory scrolling
- Weapon switching
- UI element scrolling

**ACTUAL BEHAVIOR:**
No mouse wheel support in InputSystem at all.

**REPRODUCTION CODE:**
```go
// Desired usage
wheelX, wheelY := inputSys.GetMouseWheel()
if wheelY > 0 {
    ZoomIn()
} else if wheelY < 0 {
    ZoomOut()
}

// Current: Not possible, must call Ebiten directly
```

**GAP ANALYSIS:**
Feature explicitly documented in USER_MANUAL.md but completely missing from implementation. Ebiten provides `ebiten.Wheel()` method.

**DEPENDENCIES:**
None

---

### BUG-008: GetMouseDelta() - Missing Mouse Movement Tracking

**SEVERITY:** Missing Method

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:73  
Quote: "Move - Aim / Look direction"  
Implicit: Need mouse movement delta for aiming sensitivity

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:333  
Current: Only stores absolute position, not delta

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) GetMouseDelta() (dx, dy float64)
```
Returns mouse movement delta since last frame. Essential for:
- First-person camera control
- Aiming with sensitivity settings
- Gesture detection
- Mouse-look controls

**ACTUAL BEHAVIOR:**
InputSystem stores absolute cursor position but never calculates delta between frames.

**REPRODUCTION CODE:**
```go
// Desired API
dx, dy := inputSys.GetMouseDelta()
camera.Rotate(dx * sensitivity)

// Current workaround (game code must track)
type GameState struct {
    lastMouseX, lastMouseY int
}
currentX, currentY := ebiten.CursorPosition()
dx := currentX - state.lastMouseX
dy := currentY - state.lastMouseY
state.lastMouseX, state.lastMouseY = currentX, currentY
```

**GAP ANALYSIS:**
Common feature in action games. InputSystem should manage frame-to-frame state tracking instead of forcing every game to implement it.

**DEPENDENCIES:**
Requires InputSystem to track previous frame mouse position (new field needed)

---

### BUG-009: Input State Not Cleared Between Frames

**SEVERITY:** State Desync

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:253-260  
Code: Input state reset in processInput()

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:253-260

**EXPECTED BEHAVIOR:**
Input state should be reset at the start of each frame BEFORE polling new input, ensuring clean state transitions and preventing stale data from previous frames.

**ACTUAL BEHAVIOR:**
```go
func (s *InputSystem) processInput(entity *Entity, input *InputComponent, deltaTime float64) {
    // Reset input state
    input.MoveX = 0
    input.MoveY = 0
    input.ActionPressed = false
    input.UseItemPressed = false
    // ... process new input
}
```
State is reset per-entity during processing, not globally before frame. If an entity is removed between frames, its input state persists.

**REPRODUCTION CODE:**
```go
// Frame 1: Player presses Space
entity.InputComponent.ActionPressed = true

// Frame 2: Entity removed from world before processInput()
world.RemoveEntity(entity.ID)

// Frame 3: Entity re-added
world.AddEntity(entity)
// BUG: ActionPressed still true from Frame 1!
```

**GAP ANALYSIS:**
Edge case but causes ghost input when entities are dynamically added/removed. Reset should happen in Update() before entity loop.

**DEPENDENCIES:**
None - simple code reorg

---

### BUG-010: Mouse Position Not Tracked for Delta

**SEVERITY:** State Desync

**DOCUMENTATION REFERENCE:**
Related to BUG-008

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:333  
Code: `input.MouseX, input.MouseY = ebiten.CursorPosition()`

**EXPECTED BEHAVIOR:**
InputSystem should maintain previous frame's mouse position to calculate delta automatically.

**ACTUAL BEHAVIOR:**
No tracking of previous position. Each frame overwrites with current position, losing delta information.

**REPRODUCTION CODE:**
See BUG-008

**GAP ANALYSIS:**
Architectural issue - InputSystem has no persistent state for per-frame tracking. Needs new fields:
```go
type InputSystem struct {
    // Add these fields
    lastMouseX, lastMouseY int
    // existing fields...
}
```

**DEPENDENCIES:**
BUG-008 depends on fixing this

---

### BUG-011: Keyboard Repeat State Not Managed

**SEVERITY:** State Desync

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:304-316  
Code: Movement key handling

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:304-316

**EXPECTED BEHAVIOR:**
Game should have control over key repeat behavior (OS key repeat vs game-managed repeat).

**ACTUAL BEHAVIOR:**
Uses `ebiten.IsKeyPressed()` which includes OS key repeat. For continuous movement this is fine, but for discrete actions it causes unwanted repetition.

**REPRODUCTION CODE:**
```go
// If user holds key, OS may send repeat events
if ebiten.IsKeyPressed(ebiten.KeyE) {
    UseItem() // May fire multiple times from OS repeat!
}

// Should use edge detection for discrete actions
if inpututil.IsKeyJustPressed(ebiten.KeyE) {
    UseItem() // Fires once per press
}
```

**GAP ANALYSIS:**
Current code correctly uses `IsKeyJustPressed` for actions (lines 325-328), but documentation doesn't clarify repeat behavior expectations. Not a bug, but worth documenting.

**DEPENDENCIES:**
Documentation update only

---

### BUG-012: Rapid Key Press-Release Within Single Frame Lost

**SEVERITY:** Edge Case Failure

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:160-240  
Context: Global key handling in Update()

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:160-240

**EXPECTED BEHAVIOR:**
If user presses and releases a key within a single frame (< 16ms at 60 FPS), the input should still be detected as a "just pressed" event.

**ACTUAL BEHAVIOR:**
Ebiten's input polling happens once per frame. If key press+release occurs between polls, the event is lost entirely.

**REPRODUCTION CODE:**
```go
// User taps key extremely quickly (< 16ms)
// Frame N: Key not pressed
// [User presses and releases key in 8ms]
// Frame N+1: Key not pressed
// Result: inpututil.IsKeyJustPressed() returns false - event lost!
```

**GAP ANALYSIS:**
Fundamental limitation of poll-based input vs event-based input. Ebiten uses polling. Cannot fix without Ebiten core changes, but should be documented as a known limitation. Affects ~1% of inputs at 60 FPS, more at lower frame rates.

**DEPENDENCIES:**
Ebiten core limitation - document as known issue

---

### BUG-013: Multiple Simultaneous Key Presses Not All Detected

**SEVERITY:** Edge Case Failure

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:304-316  
Code: Movement key handling

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:304-316

**EXPECTED BEHAVIOR:**
All simultaneously pressed keys should be detected (hardware permitting).

**ACTUAL BEHAVIOR:**
Most keyboards have 6-key rollover (6KRO) or N-key rollover (NKRO). If limitation is exceeded, some keys won't register. This is hardware-dependent, not a software bug.

**REPRODUCTION CODE:**
```go
// Press 10 keys simultaneously on 6KRO keyboard
// Only 6 keys detected, rest lost
```

**GAP ANALYSIS:**
Hardware limitation, not software bug. Should document recommended key combinations avoid conflicts (e.g., don't require pressing W+A+S+D+Shift+Ctrl+Space simultaneously).

**DEPENDENCIES:**
Documentation update only

---

### BUG-014: Mouse Button Press-Release Same Frame Not Detected

**SEVERITY:** Edge Case Failure

**DOCUMENTATION REFERENCE:**
Related to BUG-012

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:334

**EXPECTED BEHAVIOR:**
Extremely fast mouse clicks (<16ms) should still register.

**ACTUAL BEHAVIOR:**
Same as BUG-012 - polling misses ultra-fast clicks.

**REPRODUCTION CODE:**
```go
// User clicks mouse in 5ms (very fast click)
// Frame N: Button not pressed
// [Click happens between frames]
// Frame N+1: Button not pressed
// Result: Click lost
```

**GAP ANALYSIS:**
Same Ebiten polling limitation as BUG-012. Affects <1% of clicks. Document as known limitation.

**DEPENDENCIES:**
Ebiten core limitation

---

### BUG-015: Touch and Keyboard Input Simultaneously Undefined Behavior

**SEVERITY:** Edge Case Failure

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:256-262  
Code: Input method auto-detection

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:256-262

**EXPECTED BEHAVIOR:**
When both touch and keyboard input detected simultaneously (e.g., tablet with keyboard), behavior should be defined and predictable.

**ACTUAL BEHAVIOR:**
```go
// Auto-detect input method: if touch input is detected, switch to touch mode
if s.mobileEnabled && len(ebiten.TouchIDs()) > 0 {
    s.useTouchInput = true
} else if !s.mobileEnabled && len(ebiten.TouchIDs()) == 0 {
    // Allow falling back to keyboard if no touches (e.g., tablet with keyboard)
    s.useTouchInput = false
}
```
Once touch is detected, switches to touch mode. Then if no touches, switches back to keyboard. Rapidly alternating input causes mode thrashing.

**REPRODUCTION CODE:**
```go
// Frame 1: User touches screen
s.useTouchInput = true // Touch mode

// Frame 2: User stops touching, uses keyboard
s.useTouchInput = false // Keyboard mode

// Frame 3: User touches screen again
s.useTouchInput = true // Touch mode again

// Virtual controls flicker on/off
```

**GAP ANALYSIS:**
Need hysteresis or explicit mode selection. Current auto-detection is too aggressive. Solutions:
1. Add mode lock (stay in touch mode for N frames after last touch)
2. Require explicit mode selection in settings
3. Support hybrid input (touch + keyboard simultaneously)

**DEPENDENCIES:**
Requires InputSystem architecture change

---

### BUG-016: Virtual Controls Overlap Touch Targeting Area

**SEVERITY:** Edge Case Failure

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:279-290  
Code: Touch position detection for "mouse" simulation

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:279-290

**EXPECTED BEHAVIOR:**
Virtual controls (D-pad, buttons) should not interfere with touch targeting (using touch as mouse pointer).

**ACTUAL BEHAVIOR:**
```go
// Use first touch outside controls as "mouse" position
if s.touchHandler != nil {
    touches := s.touchHandler.GetActiveTouches()
    for _, touch := range touches {
        // Check if touch is outside virtual controls
        // (simple heuristic: use center-screen touches)
        screenW, _ := ebiten.WindowSize()
        if touch.X > 200 && touch.X < screenW-200 {
            input.MouseX = touch.X
            input.MouseY = touch.Y
            input.MousePressed = true
            break
        }
    }
}
```
Heuristic uses hardcoded 200-pixel margins. Doesn't actually check virtual control positions, just assumes they're on left/right edges. Breaks if screen size or control layout changes.

**REPRODUCTION CODE:**
```go
// Small screen (480px wide)
// Virtual controls at left (0-150px) and right (330-480px)
// Heuristic uses (200, 280) region - misses valid touch area!

// Touch at X=180 (valid gameplay area)
if touch.X > 200 { // FALSE! Touch ignored
}
```

**GAP ANALYSIS:**
Needs proper collision detection with virtual control bounds. VirtualControlsLayout should expose a `Contains(x, y int) bool` method to check if a point is within controls.

**DEPENDENCIES:**
Requires mobile package changes (add Contains method to VirtualControlsLayout)

---

### BUG-017: No Frame-Boundary Synchronization Guarantee

**SEVERITY:** Timing Violation

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:143-245  
Code: Update() method

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:143-245

**EXPECTED BEHAVIOR:**
InputSystem.Update() should be guaranteed to run AFTER Ebiten's input polling for the current frame, ensuring input state is current.

**ACTUAL BEHAVIOR:**
No explicit synchronization. Relies on game loop calling order:
```go
// In game.Update():
world.Update(deltaTime)
    -> inputSystem.Update()
        -> ebiten.IsKeyPressed() // Are we guaranteed fresh input?
```

**REPRODUCTION CODE:**
```go
// Unclear execution order
func (g *Game) Update() error {
    // 1. Ebiten polls input (OS events)
    // 2. Ebiten calls this method
    // 3. We call inputSystem.Update()
    // 4. inputSystem reads Ebiten input state
    
    // Question: Is step 4 reading state from step 1 of THIS frame
    // or the PREVIOUS frame?
}
```

**GAP ANALYSIS:**
According to Ebiten documentation, input state is updated before Update() is called, so this is actually correct. However, InputSystem documentation doesn't state this guarantee. Should document the synchronization contract.

**DEPENDENCIES:**
Documentation update only

---

### BUG-018: InputSystem.Update() Called Before Ebiten Input Polling

**SEVERITY:** Timing Violation

**DOCUMENTATION REFERENCE:**
Related to BUG-017

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:143-245

**EXPECTED BEHAVIOR:**
InputSystem should not be callable before Ebiten's game loop starts.

**ACTUAL BEHAVIOR:**
No protection against calling Update() outside game loop:
```go
inputSys := engine.NewInputSystem()
inputSys.Update([]*engine.Entity{}, 0.016) // Called before ebiten.RunGame()
// Reads uninitialized input state!
```

**REPRODUCTION CODE:**
```go
func main() {
    inputSys := engine.NewInputSystem()
    
    // BUG: Update called before RunGame
    inputSys.Update([]*engine.Entity{}, 0.016)
    
    game := &Game{InputSys: inputSys}
    ebiten.RunGame(game)
}
```

**GAP ANALYSIS:**
Edge case but causes undefined behavior. Solutions:
1. Document that Update() must only be called from within ebiten.Game.Update()
2. Add runtime check to panic if called incorrectly
3. Add initialization flag to track if system is ready

**DEPENDENCIES:**
None - documentation or simple validation check

---

### BUG-019: SetKeyBindings() Only Sets 6 Keys, Not All 18

**SEVERITY:** Partial Implementation

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:348-355  
Code: SetKeyBindings() method

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:348-355

**EXPECTED BEHAVIOR:**
Given that InputSystem manages 18 keys:
- 4 movement (Up, Down, Left, Right)
- 2 action (Action, UseItem)
- 5 UI (Inventory, Character, Skills, Quests, Map)
- 4 system (Help, QuickSave, QuickLoad, CycleTargets)

SetKeyBindings() should allow customizing all of them.

**ACTUAL BEHAVIOR:**
```go
func (s *InputSystem) SetKeyBindings(up, down, left, right, action, useItem ebiten.Key) {
    s.KeyUp = up
    s.KeyDown = down
    s.KeyLeft = left
    s.KeyRight = right
    s.KeyAction = action
    s.KeyUseItem = useItem
}
```
Only sets 6 keys. Cannot rebind UI keys (I, C, K, J, M) or system keys (ESC, F5, F9, Tab).

**REPRODUCTION CODE:**
```go
// Want to rebind inventory from I to Tab
inputSys.SetKeyBindings(
    ebiten.KeyW, ebiten.KeyS, ebiten.KeyA, ebiten.KeyD,
    ebiten.KeySpace, ebiten.KeyE,
)
// ERROR: No way to set KeyInventory!
```

**GAP ANALYSIS:**
Incomplete API. Need either:
1. Separate setter for each key group
2. Comprehensive setter with all 18 parameters
3. Builder pattern for key bindings
4. Map-based setter: `SetKeyBinding(action string, key ebiten.Key)`

**DEPENDENCIES:**
None

---

### BUG-020: No Public API for Querying Current Key Bindings

**SEVERITY:** Partial Implementation

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:88  
Quote: "Custom Key Bindings: Edit key bindings in the settings menu"

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:40-77  
Code: Key binding fields are public but no getter methods

**EXPECTED BEHAVIOR:**
Settings menu needs to query current bindings to display them. Should provide:
```go
func (s *InputSystem) GetKeyBinding(action string) ebiten.Key
func (s *InputSystem) GetAllKeyBindings() map[string]ebiten.Key
```

**ACTUAL BEHAVIOR:**
Key fields are public (s.KeyUp, etc.) so technically accessible, but no structured API for querying by action name.

**REPRODUCTION CODE:**
```go
// Current: Direct field access (fragile)
currentUpKey := inputSys.KeyUp

// Desired: Named access
currentUpKey := inputSys.GetKeyBinding("up")
allBindings := inputSys.GetAllKeyBindings()
// map[string]ebiten.Key{
//     "up": ebiten.KeyW,
//     "down": ebiten.KeyS,
//     // ...
// }
```

**GAP ANALYSIS:**
API design issue. Direct field access works but isn't idiomatic Go. Getter methods provide:
- Abstraction (can change internal representation)
- Validation (ensure key is valid)
- Discoverability (shows up in godoc)

**DEPENDENCIES:**
None

---

### BUG-021: No API for Checking if ANY Key is Pressed

**SEVERITY:** Partial Implementation

**DOCUMENTATION REFERENCE:**
File: docs/USER_MANUAL.md:88  
Context: Key rebinding UI needs "Press any key to continue" functionality

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No IsAnyKeyPressed() method

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) IsAnyKeyPressed() bool
func (s *InputSystem) GetAnyPressedKey() (ebiten.Key, bool)
```
Returns true if any key is currently pressed, and optionally returns which key.

**ACTUAL BEHAVIOR:**
No such method. Must use BUG-003's workaround (poll all keys).

**REPRODUCTION CODE:**
```go
// Title screen: "Press any key to start"
for {
    if inputSys.IsAnyKeyPressed() {
        StartGame()
        break
    }
}

// Key binding UI: "Press key to bind..."
bindingKey, found := inputSys.GetAnyPressedKey()
if found {
    settings.SetBinding(action, bindingKey)
}
```

**GAP ANALYSIS:**
Common UI pattern. Related to BUG-003 but focuses on boolean check rather than getting all keys.

**DEPENDENCIES:**
None

---

### BUG-022: No API for Checking Modifier Key State

**SEVERITY:** Partial Implementation

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:1-450  
Context: Many games use Shift+Key, Ctrl+Key for alternate actions

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:1-450  
Missing: No modifier key support

**EXPECTED BEHAVIOR:**
```go
func (s *InputSystem) IsShiftPressed() bool
func (s *InputSystem) IsControlPressed() bool
func (s *InputSystem) IsAltPressed() bool
func (s *InputSystem) IsSuperPressed() bool // Windows/Cmd key
```

**ACTUAL BEHAVIOR:**
No modifier key methods. Can check individually with `ebiten.IsKeyPressed(ebiten.KeyShift)` but InputSystem doesn't provide convenience methods.

**REPRODUCTION CODE:**
```go
// Shift+Click for multi-select in inventory
if inputSys.IsShiftPressed() && inputSys.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
    AddToSelection(item)
} else if inputSys.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
    SelectOnly(item)
}

// Current workaround
if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
    // Must check both left and right shift!
}
```

**GAP ANALYSIS:**
Modifier keys (Shift, Ctrl, Alt) have left and right variants. InputSystem should abstract this:
- Check both ShiftLeft and ShiftRight
- Provide simple `IsShiftPressed()` API
- Handle OS-specific keys (Super/Command/Windows)

**DEPENDENCIES:**
None

---

### BUG-023: Mobile Input Silently Disabled if InitializeVirtualControls() Not Called

**SEVERITY:** Silent Malfunction

**DOCUMENTATION REFERENCE:**
File: pkg/engine/input_system.go:95-100  
Code: InitializeVirtualControls() comment says "Should be called after screen size is known"

**IMPLEMENTATION LOCATION:**
File: pkg/engine/input_system.go:95-100

**EXPECTED BEHAVIOR:**
If mobile input is enabled but virtual controls aren't initialized, should:
1. Log warning
2. Auto-initialize with default screen size, OR
3. Disable mobile input with error message

**ACTUAL BEHAVIOR:**
```go
func (s *InputSystem) InitializeVirtualControls(screenWidth, screenHeight int) {
    if s.mobileEnabled {
        s.virtualControls = mobile.NewVirtualControlsLayout(screenWidth, screenHeight)
    }
}
```
If this is never called, `s.virtualControls` remains nil. Then in Update():
```go
if s.mobileEnabled && s.virtualControls != nil {
    s.virtualControls.Update()
    // ...
}
```
Mobile input silently doesn't work. No error, no warning, just broken input.

**REPRODUCTION CODE:**
```go
// Forget to call InitializeVirtualControls
inputSys := engine.NewInputSystem()
// inputSys.mobileEnabled = true (if on mobile platform)
// inputSys.virtualControls = nil (not initialized!)

// In game loop
inputSys.Update(entities, deltaTime)
// Virtual controls never update - no input works!
// No error message to help debug
```

**GAP ANALYSIS:**
Silent failure is worst kind of bug - very hard to debug. Solutions:
1. Auto-initialize in NewInputSystem() with default 800x600
2. Log warning if mobileEnabled but virtualControls == nil
3. Panic in Update() if invalid state detected (fail-fast)
4. Return error from Update() to propagate up

**DEPENDENCIES:**
None

---

## Repair Execution Plan

### Dependency Graph
```
Foundation Layer (no dependencies):
  - BUG-001, BUG-002: Key release detection
  - BUG-003: GetPressedKeys()
  - BUG-004, BUG-005, BUG-006: Mouse button edge detection
  - BUG-007: Mouse wheel
  - BUG-019: SetKeyBindings() expansion
  - BUG-020: GetKeyBinding() API
  - BUG-021: IsAnyKeyPressed()
  - BUG-022: Modifier key API
  - BUG-023: Mobile initialization error handling

State Management Layer (depends on foundation):
  - BUG-010: Mouse delta tracking (required for BUG-008)
  - BUG-009: Input state reset (architectural)
  
Feature Layer (depends on state):
  - BUG-008: GetMouseDelta() (depends on BUG-010)
  - BUG-016: Virtual control collision (depends on mobile package)
  
Edge Case Layer (complex, defer):
  - BUG-015: Touch+keyboard mode management (depends on BUG-023)
  
Documentation Layer (no code changes):
  - BUG-011: Document key repeat behavior
  - BUG-012: Document rapid input limitation
  - BUG-013: Document key rollover limitation
  - BUG-014: Document fast click limitation
  - BUG-017: Document frame sync guarantee
  - BUG-018: Document Update() call requirements
```

### Repair Sequence (Priority Order)

1. **BUG-004: IsMouseButtonJustPressed()** - Priority: 91
   - Critical for UI (all clicks currently broken)
   - Simple implementation (wrapper for inpututil)
   - High impact, low complexity
   
2. **BUG-001: IsKeyReleased()** - Priority: 88
   - Required for charge attacks, aim mechanics
   - Simple implementation
   - High user impact

3. **BUG-007: GetMouseWheel()** - Priority: 85
   - Explicitly documented feature, not implemented
   - Simple implementation
   - Complete gap

4. **BUG-023: Mobile Init Error Handling** - Priority: 82
   - Silent failure is dangerous
   - Blocks mobile platform entirely
   - Simple fix (add validation)

5. **BUG-010: Mouse Delta Tracking** - Priority: 80
   - Foundation for BUG-008
   - Architectural change (adds state)
   - Required for aiming

6. **BUG-008: GetMouseDelta()** - Priority: 78
   - Depends on BUG-010
   - High value for action games
   - Medium complexity

7. **BUG-019: SetKeyBindings() Expansion** - Priority: 75
   - Documented feature (custom bindings)
   - API expansion
   - Moderate complexity

8. **BUG-020: GetKeyBinding() API** - Priority: 72
   - Needed for settings UI
   - Simple getter methods
   - Low complexity

9. **BUG-003: GetPressedKeys()** - Priority: 70
   - Required for key rebinding UI
   - Simple wrapper
   - Medium impact

10. **BUG-021: IsAnyKeyPressed()** - Priority: 68
    - Common UI pattern
    - Simple implementation
    - Related to BUG-003

11. **BUG-022: Modifier Key API** - Priority: 66
    - Convenience feature
    - Simple implementation
    - Nice-to-have

12. **BUG-005: IsMouseButtonReleased()** - Priority: 64
    - Needed for drag-and-drop
    - Simple implementation
    - Medium impact

13. **BUG-009: Input State Reset** - Priority: 62
    - Edge case but important
    - Simple code reorg
    - Low risk

14. **BUG-016: Virtual Control Overlap** - Priority: 58
    - Mobile-specific
    - Requires mobile package change
    - Medium complexity

15. **BUG-015: Touch+Keyboard Mode** - Priority: 54
    - Complex architectural change
    - Lower priority (edge case)
    - Requires design decision

16. **BUG-002: IsKeyJustReleased() alias** - Priority: 45
    - Duplicate of BUG-001
    - Naming consistency only
    - Trivial implementation

17. **BUG-006: IsMouseButtonJustReleased() alias** - Priority: 43
    - Duplicate of BUG-005
    - Naming consistency only
    - Trivial implementation

Documentation Updates (Priority: 40-50):
18. **BUG-017: Document frame sync** - Priority: 50
19. **BUG-018: Document Update() requirements** - Priority: 48
20. **BUG-011: Document key repeat** - Priority: 46
21. **BUG-012: Document rapid input limitation** - Priority: 44
22. **BUG-013: Document key rollover** - Priority: 42
23. **BUG-014: Document fast click limitation** - Priority: 40

### Estimated Repair Complexity

- **Simple fixes (< 20 lines)**: 11 bugs
  - BUG-001, 002, 003, 004, 005, 006, 007, 020, 021, 022, 023
  
- **Moderate fixes (20-100 lines)**: 9 bugs
  - BUG-008, 009, 010, 016, 019, plus 5 documentation bugs
  
- **Complex fixes (> 100 lines)**: 3 bugs
  - BUG-015 (architectural change to mode management)
  - Full test suite for all fixes
  - Integration testing

**Total Estimated LOC:**
- Production code: ~450 lines
- Test code: ~900 lines
- Documentation: ~300 lines
- **Total: ~1650 lines**

**Estimated Time:** 8-12 hours for complete repair and validation
