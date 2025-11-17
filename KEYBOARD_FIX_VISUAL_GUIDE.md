# WebAssembly Virtual Keyboard Fix - Visual Guide

## Problem Visualization

### Before Fix (❌ Broken)

```
┌─────────────────────────────────────────┐
│  Mobile Browser                         │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │ Document (touch-action: none)     │ │  ← BLOCKS ALL TOUCHES
│  │                                   │ │
│  │  ┌─────────────────────────────┐ │ │
│  │  │ Canvas (touch-action: none) │ │ │
│  │  │                             │ │ │
│  │  │   Game Render Area          │ │ │
│  │  │                             │ │ │
│  │  └─────────────────────────────┘ │ │
│  │                                   │ │
│  │  Hidden Input (off-screen)        │ │
│  │  ┌─────┐                          │ │
│  │  │Input│ (at -9999px, -9999px)   │ │  ← CAN'T BE FOCUSED
│  │  └─────┘ (pointerEvents: none)   │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘

User taps screen → Touch blocked by body
                 → Input can't receive focus
                 → Keyboard never appears ❌
```

### After Fix (✅ Working)

```
┌─────────────────────────────────────────┐
│  Mobile Browser                         │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │ Document (NO touch-action)        │ │  ← ALLOWS INPUT TOUCHES
│  │                                   │ │
│  │  ┌─────────────────────────────┐ │ │
│  │  │ Canvas (touch-action: none) │ │ │  ← Game controls work
│  │  │                             │ │ │
│  │  │   Game Render Area          │ │ │
│  │  │                             │ │ │
│  │  └─────────────────────────────┘ │ │
│  │                                   │ │
│  │  When ShowKeyboard() called:      │ │
│  │       ┌──────────────┐            │ │
│  │       │ Input (200px)│            │ │  ← ON-SCREEN, TAPPABLE
│  │       │ bottom-center│            │ │
│  │       │ opacity: 0.01│            │ │  ← Nearly invisible
│  │       └──────────────┘            │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘

User taps anywhere → ShowKeyboard() moves input on-screen
                  → Input receives focus
                  → Keyboard appears! ✅
```

---

## State Diagram

```
┌─────────────────────────────────────────────────────────┐
│                   Keyboard Lifecycle                     │
└─────────────────────────────────────────────────────────┘

    Initial State                After ShowKeyboard()
    ─────────────                ────────────────────
         
    ┌───────────┐                   ┌───────────┐
    │  Input    │                   │  Input    │
    │           │                   │           │
    │ Off-screen│   ShowKeyboard()  │ On-screen │
    │ -9999px   │ ────────────────> │ 50%, 80px │
    │           │                   │ opacity   │
    │ Hidden    │                   │ 0.01      │
    └───────────┘                   └───────────┘
         ↑                                 │
         │                                 │
         │         HideKeyboard()          │
         └─────────────────────────────────┘

         │                                 │
         │                                 │
         ▼                                 ▼
         
    No keyboard                     Keyboard visible
    Game controls                   Text input mode
    work normally                   Characters forwarded
```

---

## Touch Event Flow

### Before Fix (❌)

```
User Touch on Screen
    ↓
Document touchstart listener
    ↓
e.preventDefault() ← BLOCKS EVERYTHING
    ↓
❌ Input never receives touch
❌ Focus never happens
❌ Keyboard never shows
```

### After Fix (✅)

```
User Touch on Screen
    ↓
Document touchstart listener
    ↓
Check target element:
    │
    ├─ Canvas? → Allow (game controls)
    ├─ INPUT?  → Allow (keyboard focus) ✅
    ├─ Button? → Allow (UI)
    └─ Other?  → Prevent (no zoom/scroll)
    ↓
✅ Input receives touch
✅ Focus triggered
✅ Keyboard appears!
```

---

## CSS Cascade

### Before Fix

```css
/* DOCUMENT LEVEL */
body {
    touch-action: none;  ❌ Blocks everything
}

/* ELEMENT LEVEL */
input {
    /* inherits touch-action: none */
    /* can't be tapped or focused */
}
```

### After Fix

```css
/* DOCUMENT LEVEL */
body {
    /* NO touch-action */
    /* Allows normal touch behavior */
}

/* CANVAS SPECIFIC */
#gameCanvas {
    touch-action: none;  ✅ Only blocks on canvas
}

/* INPUT SPECIFIC */
input {
    touch-action: auto !important;     ✅ Explicitly allow
    pointer-events: auto !important;   ✅ Explicitly allow
}
```

---

## Code Flow

### ShowKeyboard() Function

```
┌────────────────────────────────────────────────┐
│ mobile.ShowKeyboard() called                   │
└────────────────────────────────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Init keyboard element│
        │ (if not exists)      │
        └──────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Clear input value    │
        │ lastInputValue = ""  │
        └──────────────────────┘
                    ↓
        ┌──────────────────────────────────┐
        │ Move input ON-SCREEN             │
        │ - left: 50%                      │
        │ - transform: translateX(-50%)    │
        │ - bottom: 80px                   │
        │ - top: auto                      │
        └──────────────────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Call focus()         │
        │ Programmatically     │
        └──────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Log focus state      │
        │ Verify success       │
        └──────────────────────┘
```

### HideKeyboard() Function

```
┌────────────────────────────────────────────────┐
│ mobile.HideKeyboard() called                   │
└────────────────────────────────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Call blur()          │
        │ Dismiss keyboard     │
        └──────────────────────┘
                    ↓
        ┌──────────────────────────────────┐
        │ Move input OFF-SCREEN            │
        │ - left: -9999px                  │
        │ - top: -9999px                   │
        │ - bottom: auto                   │
        │ - transform: none                │
        └──────────────────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Clear input value    │
        │ lastInputValue = ""  │
        └──────────────────────┘
                    ↓
        ┌──────────────────────┐
        │ Log completion       │
        └──────────────────────┘
```

---

## Event Forwarding Chain

```
Mobile Keyboard
    ↓ (user types 'H')
Hidden Input Element
    ↓ (input event)
inputEventListener
    ↓
dispatchInputEvent(canvas, 'H')
    ↓
InputEvent dispatched to Canvas
    ↓
Ebiten AppendInputChars
    ↓
✅ Character 'H' received by game
```

---

## Browser Compatibility

### iOS Safari

```
Before: ❌ Programmatic focus() blocked by touch-action: none
After:  ✅ Input on-screen + touch-action: auto = keyboard works

Requirements:
- Input must be on-screen (not -9999px)
- Input must have touch-action: auto
- Font-size >= 16px (prevents auto-zoom)
- Opacity > 0 (must be "visible")
```

### Android Chrome

```
Before: ❌ Programmatic focus() blocked by touch-action: none
After:  ✅ Input on-screen + touch-action: auto = keyboard works

Requirements:
- Input must be on-screen
- Input must have touch-action: auto
- Tap or programmatic focus works
```

### Desktop Browsers

```
Before: ✅ Keyboard always available (physical)
After:  ✅ No changes - input hidden, keyboard works

Impact:
- Off-screen input doesn't interfere
- Mouse/keyboard input unchanged
- Game controls unaffected
```

---

## Positioning Details

### Off-Screen (Default)

```
position: fixed;
left: -9999px;
top: -9999px;
width: 200px;
height: 50px;
opacity: 0.01;

Purpose: Hide input during gameplay
Effect: No touch interference
```

### On-Screen (During Text Input)

```
position: fixed;
left: 50%;
transform: translateX(-50%);  ← Center horizontally
bottom: 80px;                 ← Above keyboard area
width: 200px;
height: 50px;
opacity: 0.01;                ← Nearly invisible
z-index: 999;

Purpose: Provide tappable target
Effect: User can tap to trigger keyboard
```

---

## Testing Checklist

### ✅ Functionality

- [ ] Keyboard appears when ShowKeyboard() called
- [ ] Keyboard dismisses when HideKeyboard() called
- [ ] Characters reach game via AppendInputChars
- [ ] Backspace works
- [ ] Enter key works
- [ ] Special keys work (Tab, Escape)

### ✅ Visual

- [ ] Input nearly invisible (opacity 0.01)
- [ ] Input doesn't block game view
- [ ] Input moves on-screen during text input
- [ ] Input moves off-screen during gameplay

### ✅ Touch Behavior

- [ ] Canvas touches work for game controls
- [ ] Input touches trigger focus
- [ ] No unwanted zoom/scroll
- [ ] Keyboard shows on tap

### ✅ Console

- [ ] No JavaScript errors
- [ ] [VentureKeyboard] logs appear
- [ ] Focus state logged correctly
- [ ] Input events logged

---

## Common Scenarios

### Scenario 1: Character Name Entry

```
1. Player enters character creation
   → ShowKeyboard() called
   → Input moves on-screen

2. Player taps screen
   → Input receives focus
   → Keyboard appears

3. Player types "Hero"
   → Characters forwarded to game
   → Name displayed

4. Player presses Enter
   → HideKeyboard() called
   → Input moves off-screen
   → Keyboard dismisses
```

### Scenario 2: Server Address Entry

```
1. Player selects multiplayer
   → ShowKeyboard() called
   → Input moves on-screen

2. Player taps screen
   → Input receives focus
   → Keyboard appears

3. Player types "localhost:8080"
   → Characters forwarded to game
   → Address displayed

4. Player presses Enter
   → HideKeyboard() called
   → Input moves off-screen
   → Connection attempted
```

### Scenario 3: Crafting Search

```
1. Player opens crafting UI
   → ShowKeyboard() called
   → Input moves on-screen

2. Player taps screen
   → Input receives focus
   → Keyboard appears

3. Player types "sword"
   → Characters forwarded to game
   → Recipes filtered

4. Player closes crafting
   → HideKeyboard() called
   → Input moves off-screen
   → Returns to gameplay
```

---

*Visual Guide Created: 2025-11-17*  
*Companion to: KEYBOARD_FIX_NOVEMBER_2025.md*
