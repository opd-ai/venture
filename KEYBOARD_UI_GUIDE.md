# Virtual Keyboard UI - Visual Guide

## Character Creation Screen Layout

### Name Input Step (WASM/Mobile)

```
┌─────────────────────────────────────────────────────────┐
│                  CHARACTER CREATION                      │
│                                                          │
│            Step 1 of 4: Choose Your Name                │
│                                                          │
│          Enter your character's name:                    │
│                                                          │
│         ┌──────────────────────────────┐                │
│         │ Brave_                       │ ← Input field  │
│         └──────────────────────────────┘                │
│                                                          │
│          Or tap a preset name below:                     │
│                                                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │
│  │ Warrior │ │  Mage   │ │  Rogue  │ │ Ranger  │      │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘      │
│                                                          │
│                    ┌─────────┐                           │
│                    │  Auto   │ ← Random name             │
│                    └─────────┘                           │
│                                                          │
│                                                          │
│     Press ENTER to continue | F2 to save as default     │
│                                                          │
│  ┌──────┐                              ┌──────┐         │
│  │ Back │                              │ Next │         │
│  └──────┘                              └──────┘         │
└─────────────────────────────────────────────────────────┘
                         ↑
                   Mobile keyboard
                 appears here when
                 input field tapped
```

### Preset Button Behavior

**"Warrior" Button:**
- Tap → Name instantly set to "Warrior"
- Continue button becomes active
- Can proceed immediately

**"Mage" Button:**
- Tap → Name instantly set to "Mage"
- Continue button becomes active
- Can proceed immediately

**"Rogue" Button:**
- Tap → Name instantly set to "Rogue"
- Continue button becomes active
- Can proceed immediately

**"Ranger" Button:**
- Tap → Name instantly set to "Ranger"
- Continue button becomes active
- Can proceed immediately

**"Auto" Button:**
- Tap → Random name generated
- Examples: "Braveblade", "Swiftheart", "Darkfist", "Eldereye"
- Name based on class selection
- Can tap multiple times for different names

### Visual States

#### Normal State (Button Ready)
```
┌─────────┐
│ Warrior │  ← Green border, white text
└─────────┘
```

#### Pressed State (User Tapping)
```
┌─────────┐
│ Warrior │  ← Darker background, visual feedback
└─────────┘
```

#### After Selection
```
┌──────────────────────────────┐
│ Warrior_                     │  ← Name appears in input field
└──────────────────────────────┘

Message: "Name set to: Warrior"  ← Confirmation message
```

## Keyboard Test Application UI

```
┌─────────────────────────────────────────────────────────┐
│               Virtual Keyboard Test                      │
│                                                          │
│  Platform: WASM=true, IsKeyboardSupported=true          │
│                                                          │
│  Instructions:                                           │
│  1. TAP anywhere to show keyboard                        │
│  2. TYPE on your device keyboard                         │
│  3. Press ENTER to hide keyboard                         │
│  4. Press ESC to clear text                              │
│                                                          │
│  ┌────────────────────────────────────────────────┐     │
│  │ [Tap to enter text]                            │     │
│  └────────────────────────────────────────────────┘     │
│   ↑ Green border when keyboard shown                    │
│                                                          │
│  Keyboard Status: SHOWN | Tap Count: 1 | Length: 0      │
│                                                          │
│  Event Log:                                              │
│    Application started                                   │
│    Platform: WASM=true                                   │
│    Tap detected (count: 1)                               │
│    Calling mobile.ShowKeyboard()                         │
│    Keyboard shown flag set to true                       │
│    Character received: 'a' (code: 97)                    │
│    Character received: 'b' (code: 98)                    │
└─────────────────────────────────────────────────────────┘
                         ↑
                   Mobile keyboard
                 appears at bottom
```

## Browser Console View

### Developer Tools (F12) - Console Tab

```
Console
────────────────────────────────────────────────────────
▼ KeyboardTest: Application started
▼ KeyboardTest: Platform: WASM=true
▼ [VentureKeyboard] Initializing virtual keyboard element
▼ [VentureKeyboard] Virtual keyboard element created and added to DOM
▼ [VentureKeyboard] Element ID: venture-keyboard-input, Type: text, InputMode: text
▼ KeyboardTest: Tap detected (count: 1)
▼ [VentureKeyboard] ShowKeyboard() called
▼ [VentureKeyboard] Keyboard element focused - mobile keyboard should appear
▼ [VentureKeyboard] Focus successful - active element is venture-keyboard-input
▼ [VentureKeyboard] Input event: new chars added: 'a'
▼ KeyboardTest: Character received: 'a' (code: 97)
▼ [VentureKeyboard] Input event: new chars added: 'b'
▼ KeyboardTest: Character received: 'b' (code: 98)
```

### Success Indicators ✅
- ✅ Green "Focus successful" message
- ✅ "Input event: new chars added" messages
- ✅ Character codes displayed
- ✅ No error messages

### Failure Indicators ❌
- ❌ Red "Focus failed - active element is: CANVAS"
- ❌ No "Input event" messages
- ❌ Error: "Keyboard element is undefined"

## Mobile Device View

### iOS Safari - Character Creation
```
┌───────────────────────────────┐
│  ☰ ↻ venture-game.com    🔍   │  ← Safari UI
├───────────────────────────────┤
│                               │
│   CHARACTER CREATION          │
│                               │
│   Enter your character's      │
│   name:                       │
│                               │
│   ┌─────────────────────┐     │
│   │ Hero_               │     │
│   └─────────────────────┘     │
│                               │
│   Or tap a preset name:       │
│                               │
│   [Warrior] [Mage]            │
│   [Rogue] [Ranger]            │
│      [Auto]                   │
│                               │
├───────────────────────────────┤
│  Q W E R T Y U I O P          │  ← iOS Keyboard
│   A S D F G H J K L           │     appears here
│    Z X C V B N M              │
│     ┌───────┐                 │
│     │ Done  │                 │
│     └───────┘                 │
└───────────────────────────────┘
```

### Android Chrome - Character Creation
```
┌───────────────────────────────┐
│  ← venture-game.com      ⋮    │  ← Chrome UI
├───────────────────────────────┤
│                               │
│   CHARACTER CREATION          │
│                               │
│   Enter your character's      │
│   name:                       │
│                               │
│   ┌─────────────────────┐     │
│   │ Champion_           │     │
│   └─────────────────────┘     │
│                               │
│   Or tap a preset name:       │
│                               │
│   [Warrior] [Mage]            │
│   [Rogue] [Ranger]            │
│      [Auto]                   │
│                               │
├───────────────────────────────┤
│  q w e r t y u i o p          │  ← Android Keyboard
│   a s d f g h j k l           │     appears here
│    z x c v b n m              │
│  ┌───────────────┐            │
│  │    Return     │            │
│  └───────────────┘            │
└───────────────────────────────┘
```

## User Flow Diagram

### Scenario 1: Keyboard Works
```
User taps name field
        ↓
Virtual keyboard appears
        ↓
User types name
        ↓
Characters appear in field
        ↓
User presses Enter/Done
        ↓
Keyboard hides
        ↓
User taps Next
        ↓
Proceeds to class selection
```

### Scenario 2: Keyboard Doesn't Work
```
User taps name field
        ↓
Keyboard doesn't appear
        ↓
User sees preset buttons
        ↓
User taps "Warrior" button
        ↓
Name instantly set to "Warrior"
        ↓
User taps Next
        ↓
Proceeds to class selection
```

### Scenario 3: User Prefers Quick Selection
```
User sees character creation
        ↓
User taps "Auto" button
        ↓
Random name generated
        ↓
If user likes it: tap Next
        ↓
If not: tap "Auto" again for new name
        ↓
Proceeds to class selection
```

## Visual Design Notes

### Button Styling
- **Size:** 100px width × 36px height (meets touch target guidelines)
- **Spacing:** 10px horizontal gap between buttons
- **Colors:** 
  - Normal: Dark background (40,40,50), White text
  - Hover/Press: Lighter background (60,60,70), White text
  - Border: Green (100,200,100) when active

### Input Field Styling
- **Size:** 300px width × 30px height
- **Colors:**
  - Background: Dark (40,40,50)
  - Border: Blue-gray (150,150,200)
  - Text: White with cursor (underscore)

### Layout Positioning
- Input field: Centered horizontally, 150px from panel top
- Preset hint text: 35px below input field
- Preset buttons: Centered horizontally, arranged in 2 rows
  - Row 1: Warrior, Mage, Rogue, Ranger (4 buttons)
  - Row 2: Auto (1 button, centered)

## Accessibility Features

✅ **Touch Target Size:** All buttons meet 44px minimum (iOS/Android guidelines)
✅ **Visual Feedback:** Pressed state provides immediate feedback
✅ **Clear Labels:** Simple, recognizable class names
✅ **Fallback Always Available:** Never blocked from progression
✅ **Confirmation Message:** Shows what was selected

## Platform-Specific Notes

### Desktop Browsers
- Physical keyboard works normally
- Preset buttons still available but not necessary
- Can type freely without tapping buttons

### Mobile Browsers (iOS/Android)
- Virtual keyboard should appear when tapping input
- If keyboard fails, preset buttons provide alternative
- "Auto" button useful for quick testing/demo
- Touch targets optimized for finger input

### Tablet Browsers
- Hybrid behavior: may have virtual or physical keyboard
- Preset buttons provide convenient quick selection
- Layout adjusts to larger screen size
