# WASM Mobile Controls Layout

## Virtual Controls (Gameplay)

```
┌─────────────────────────────────────────────────────────────┐
│  HP: [████████████░░░░░░] 120/150     Level 5   XP: 450/500 │ HUD
│                                                      ┌──┐    │
│                                                      │☰ │    │ Menu Button
│                                                      └──┘    │ (44x44px)
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│                    GAME WORLD AREA                            │
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│                                                               │
│   ┌────┐                                          ┌──┐       │
│   │ ⬆  │                                          │B │       │ Action Buttons
│ ┌─┼────┼─┐                                      ┌─┴──┴─┐     │ (44x44px each)
│ │⬅│  ●│➡│                                      │  A  │     │
│ └─┼────┼─┘                                      └──────┘     │
│   │ ⬇  │                                                     │
│   └────┘                                                     │
│  Virtual D-Pad                                               │
│  (120x120px)                                                 │
└─────────────────────────────────────────────────────────────┘
```

## Menu Screen Layout (Example: Inventory)

```
┌─────────────────────────────────────────────────────────────┐
│  INVENTORY                                            ┌──┐   │
│                                                       │✕ │   │ Close Button
│                                                       └──┘   │ (44x44px)
│  ┌──┬──┬──┬──┬──┬──┬──┬──┐                                  │
│  │  │  │  │  │  │  │  │  │  ← Inventory Grid               │
│  ├──┼──┼──┼──┼──┼──┼──┼──┤    (48x48px slots)              │
│  │  │  │  │  │  │  │  │  │                                  │
│  ├──┼──┼──┼──┼──┼──┼──┼──┤                                  │
│  │  │  │  │  │  │  │  │  │                                  │
│  ├──┼──┼──┼──┼──┼──┼──┼──┤                                  │
│  │  │  │  │  │  │  │  │  │                                  │
│  └──┴──┴──┴──┴──┴──┴──┴──┘                                  │
│                                                               │
│  Item: Iron Sword                                            │
│  Damage: +15  Rarity: Uncommon                               │
│                                                               │
│  [Swipe ⬆⬇ to scroll if more items]                         │
│                                                               │
└─────────────────────────────────────────────────────────────┘

Note: Virtual controls (D-pad, A/B) are HIDDEN when menu is open
```

## Menu Screen with Tabs (Example: Quest Log)

```
┌─────────────────────────────────────────────────────────────┐
│  QUEST LOG                                            ┌──┐   │
│                                                       │✕ │   │ Close Button
│  ┌─────────────┬─────────────┐                       └──┘   │ (44x44px)
│  │   ACTIVE    │  COMPLETED  │  ← Tab Buttons               │
│  │  (selected) │             │    (120x44px each)           │
│  └─────────────┴─────────────┘                              │
│                                                               │
│  📋 Find the Lost Artifact                                   │
│     Location: Ancient Temple                                 │
│     Reward: 500 XP, Gold Ring                                │
│                                                               │
│  📋 Defeat the Goblin King                                   │
│     Location: Dark Forest                                    │
│     Reward: 750 XP, Magic Sword                              │
│                                                               │
│  📋 Gather 10 Healing Herbs                                  │
│     Progress: 7/10                                           │
│     Reward: 300 XP, Health Potion                            │
│                                                               │
│  [Swipe ⬆⬇ to scroll if more quests]                        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## NPC Dialog Layout (Pattern - Implementation Pending)

```
┌─────────────────────────────────────────────────────────────┐
│                                                               │
│  [NPC Portrait]                                               │
│                                                               │
│  "Greetings, traveler! I have wares if you have coin.        │
│   What brings you to these parts?"                           │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  1. "I'm looking for supplies."                        │  │ Choice 1
│  └────────────────────────────────────────────────────────┘  │ (400x44px)
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  2. "Do you know anything about the ancient temple?"   │  │ Choice 2
│  └────────────────────────────────────────────────────────┘  │ (400x44px)
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  3. "Farewell."                                        │  │ Choice 3
│  └────────────────────────────────────────────────────────┘  │ (400x44px)
│                                                               │
└─────────────────────────────────────────────────────────────┘

Note: Each choice is a touch-tappable button
      Keyboard shortcuts (1-3) still work on desktop
```

## Character Creation Layout

```
┌─────────────────────────────────────────────────────────────┐
│  CHARACTER CREATION                                           │
│                                                               │
│  Name: ┌──────────────────────────────────────┐             │
│        │ [Tap to type] Adventurer            │🔤           │ Name Input
│        └──────────────────────────────────────┘             │ + Keyboard Icon
│                                                               │
│  Class:                                                       │
│   ◉ Warrior    ○ Mage      ○ Rogue                          │ Touch Radio
│   ○ Ranger     ○ Cleric    ○ Necromancer                    │ Buttons
│                                                               │
│  Stats:                                                       │
│   Strength:   10  [-] [+]                                    │ +/- Buttons
│   Magic:       8  [-] [+]                                    │ (44x44px)
│   Agility:    12  [-] [+]                                    │
│                                                               │
│  ┌────────────┐                            ┌──────────────┐  │
│  │    BACK    │                            │    CREATE    │  │ Action Buttons
│  └────────────┘                            └──────────────┘  │ (120x44px)
│                                                               │
└─────────────────────────────────────────────────────────────┘

When name field is tapped:
┌─────────────────────────────────────────────────────────────┐
│                                                               │
│  [Virtual keyboard appears at bottom of screen]              │
│  ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐                           │
│  │Q │W │E │R │T │Y │U │I │O │P │                           │
│  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘                           │
│  ┌──┬──┬──┬──┬──┬──┬──┬──┬──┐                              │
│  │A │S │D │F │G │H │J │K │L │                              │
│  └──┴──┴──┴──┴──┴──┴──┴──┴──┘                              │
│  ┌──┬──┬──┬──┬──┬──┬──┬──┬──┐                              │
│  │Z │X │C │V │B │N │M │⌫ │  │                              │
│  └──┴──┴──┴──┴──┴──┴──┴──┴──┘                              │
│  ┌────────────────────────────────┬──────┐                  │
│  │          SPACE                 │ENTER │                  │
│  └────────────────────────────────┴──────┘                  │
└─────────────────────────────────────────────────────────────┘
```

## Tutorial Screen Layout

```
┌─────────────────────────────────────────────────────────────┐
│                                                               │
│  TUTORIAL                                                     │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                                                         │ │
│  │  Step 1: Movement                                       │ │
│  │                                                         │ │
│  │  Use the virtual D-pad on the left side of the screen  │ │
│  │  to move your character.                                │ │
│  │                                                         │ │
│  │  • Drag in any direction to move                        │ │
│  │  • Release to stop                                      │ │
│  │                                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌──────────────┐                       ┌────────────────┐   │
│  │SKIP TUTORIAL │                       │      NEXT      │   │ Buttons
│  └──────────────┘                       └────────────────┘   │ (120x44px)
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Control Size Reference

### Minimum Touch Targets (iOS/Android HIG)
- **Buttons:** 44x44px minimum ✅
- **Inventory Slots:** 48x48px minimum ✅
- **Tab Buttons:** 120x44px minimum ✅
- **Dialog Choices:** 400x44px ✅
- **Virtual D-pad:** 120x120px ✅
- **Spacing:** 8px minimum between elements ✅

### Virtual Controls Positioning
- **D-pad:** Bottom-left (margin: 5% of screen height)
- **A/B Buttons:** Bottom-right (margin: 5% of screen height)
- **Menu Button:** Top-right (margin: 5% of screen height)
- **Close Buttons:** Top-right of windows (10px from edge)

### Menu Button Behavior
```
Tap ☰ button → Opens Pause/Menu Screen
                ├─ Resume Game
                ├─ Inventory
                ├─ Character
                ├─ Skills
                ├─ Quests
                ├─ Map
                ├─ Crafting
                ├─ Settings
                └─ Save & Quit
```

### Virtual Controls Auto-Hide
```
State: In Gameplay
├─ Virtual controls visible (D-pad, A/B, ☰)
└─ Player taps ☰ → Menu opens
    ├─ Virtual controls HIDDEN
    └─ Player closes menu
        └─ Virtual controls VISIBLE again
```

## Color Coding

### Button States
- **Normal:** Dark blue-gray (#323246)
- **Pressed:** Light blue (#5078C8)
- **Disabled:** Dark gray (#1E1E28)
- **Active Tab:** Light blue (#5078C8)
- **Inactive Tab:** Dark blue-gray (#323246)

### Visual Feedback
- ✅ Color change on press (Normal → Pressed)
- ✅ Border highlight on active elements
- ✅ Disabled state grayed out
- ✅ Tab highlighting for active tab

## Accessibility Notes

1. **Touch Targets:** All interactive elements ≥44x44px
2. **Spacing:** Minimum 8px between tappable elements
3. **Visual Feedback:** Immediate color change on touch
4. **Haptic Feedback:** Available on supported devices
5. **Keyboard Fallback:** All touch controls have keyboard equivalents for desktop

---

**Legend:**
- ┌─┐ : Borders/boxes
- ● : Center point
- ⬆⬇⬅➡ : Direction indicators
- 📋 : Quest icon
- 🔤 : Keyboard icon
- ☰ : Menu icon (hamburger)
- ✕ : Close icon
- [-][+] : Increment/decrement buttons
