# Toggle Fix - Inventory & Quest UI

## Issue
Pressing **I** would open the inventory, but pressing **I** again wouldn't close it. Same issue with **J** for quest log.

## Root Cause
The UI toggle logic was in the InputSystem callbacks, but InputSystem.Update() is called inside World.Update(). When a UI was visible, the game logic was:

```go
// In Game.Update()
InventoryUI.Update()
QuestUI.Update()

// Skip world update if UI is visible
if !InventoryUI.IsVisible() && !QuestUI.IsVisible() {
    World.Update(deltaTime) // InputSystem is here!
}
```

**Problem:** When the inventory was open, World.Update() was skipped, so InputSystem never checked for the toggle key press.

## Solution
Moved toggle key checking into the UI Update() methods themselves, so they always check for their toggle key regardless of visibility state.

### Changes Made

**1. `pkg/engine/inventory_ui.go`**
```go
func (ui *InventoryUI) Update() {
    // Always check for toggle key, even when not visible
    if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        ui.Toggle()
        return // Don't process other input on same frame
    }
    
    if !ui.visible || ui.playerEntity == nil {
        return
    }
    
    // ... rest of UI update logic
}
```

**2. `pkg/engine/quest_ui.go`**
```go
func (ui *QuestUI) Update() {
    // Always check for toggle key, even when not visible
    if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
        ui.Toggle()
        return // Don't process other input on same frame
    }
    
    if !ui.visible || ui.playerEntity == nil {
        return
    }
    
    // ... rest of UI update logic
}
```

## How It Works Now

1. **Every frame**, before checking if UI is visible:
   - Check if I key was just pressed → toggle inventory
   - Check if J key was just pressed → toggle quest log

2. **Then**, if UI is visible:
   - Process UI-specific input (mouse clicks, E/D keys, tab switching)

3. **Finally**, in Game.Update():
   - If no UI is visible → update world (gameplay continues)
   - If any UI is visible → skip world update (pause gameplay)

## Benefits
- ✅ Toggle works consistently (open and close)
- ✅ No dependency on InputSystem being in the update loop
- ✅ Early return prevents double-processing input on toggle frame
- ✅ Clean separation: UI handles its own input
- ✅ Redundancy: InputSystem callbacks still exist as fallback

## Testing
```bash
# Build
go build -o venture-client ./cmd/client/

# Run
./venture-client

# Test:
# 1. Press I → Inventory opens ✅
# 2. Press I again → Inventory closes ✅
# 3. Press J → Quest log opens ✅
# 4. Press J again → Quest log closes ✅
```

## Status
✅ **FIXED** - Toggle functionality now works correctly for both UIs!
