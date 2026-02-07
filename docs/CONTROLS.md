# Controls Reference

**Venture - Fully Procedural Multiplayer Action-RPG**  
**Version:** 1.0.0  
**Last Updated:** February 2026

This document provides comprehensive keyboard, mouse, and gamepad control mappings for all platforms.

---

## Table of Contents

1. [Desktop Controls](#desktop-controls)
   - [Keyboard](#keyboard)
   - [Mouse](#mouse)
   - [Gamepad](#gamepad)
2. [Mobile Controls](#mobile-controls)
3. [WebAssembly Controls](#webassembly-controls)
4. [Customization](#customization)
5. [Accessibility Options](#accessibility-options)

---

## Desktop Controls

### Keyboard

#### Movement & Navigation
| Key | Action |
|-----|--------|
| **W** / **↑** | Move Up |
| **S** / **↓** | Move Down |
| **A** / **←** | Move Left |
| **D** / **→** | Move Right |
| **Shift** | Sprint (2x speed, consumes stamina) |
| **Space** | Jump / Climb |
| **Ctrl** | Crouch / Sneak |

#### Combat
| Key | Action |
|-----|--------|
| **Space** | Attack |
| **E** | Use Item / Open Chest |
| **F** | Interact / Context Action (NPCs, merchants) |
| **Tab** | Target Cycling |
| **1-5** | Cast Spells (hotbar slots 1-5) |

#### Interface
| Key | Action |
|-----|--------|
| **I** | Inventory |
| **C** | Character Sheet |
| **K** | Skills Menu |
| **J** | Quest Log |
| **M** | Map |
| **R** | Crafting Menu |
| **G** | Gallery (Image Gallery) |
| **H** | Housing Menu |
| **Esc** | Main Menu / Cancel |

#### Social & Communication
| Key | Action |
|-----|--------|
| **Enter** | Open Chat |
| **/g + Enter** | Global Chat |
| **/l + Enter** | Local Chat |
| **/p + Enter** | Party Chat |
| **/w [player] + Enter** | Whisper |
| **Shift + 1-9** | Emote 1-9 |
| **Shift + 0** | Emote 10 |
| **Shift + -** | Emote 11 (Facepalm) |
| **Shift + =** | Emote 12 (Sleep) |

#### Miscellaneous
| Key | Action |
|-----|--------|
| **F1** | Help / Tutorial |
| **F2** | Screenshot |
| **F3** | Toggle Debug Overlay |
| **F5** | Quick Save |
| **F9** | Quick Load |
| **F11** | Toggle Fullscreen |
| **~** (Tilde) | Console (developer mode) |

---

### Mouse

#### Basic Actions
| Button/Action | Function |
|---------------|----------|
| **Left Click** | Select / Attack / Interact |
| **Right Click** | Context Menu / Aim |
| **Middle Click** | Ping Location (multiplayer) |
| **Scroll Up** | Zoom In |
| **Scroll Down** | Zoom Out |
| **Mouse Movement** | Camera Rotation / Aim Direction |

#### Advanced
| Action | Function |
|--------|----------|
| **Ctrl + Left Click** | Multi-Select (inventory) |
| **Shift + Left Click** | Quick Transfer (inventory ↔ chest) |
| **Alt + Left Click** | Link Item (chat) |
| **Right Click + Drag** | Area Selection (trade, crafting) |

---

### Gamepad

**Supported Controllers:** Xbox, PlayStation, Nintendo Switch Pro, Generic (DirectInput/XInput)

#### Movement & Combat
| Button | Action |
|--------|--------|
| **Left Stick** | Movement |
| **Right Stick** | Camera / Aim |
| **L3** (Left Stick Press) | Sprint |
| **R3** (Right Stick Press) | Target Lock |
| **A / Cross** | Jump / Interact |
| **B / Circle** | Dodge Roll |
| **X / Square** | Primary Attack |
| **Y / Triangle** | Secondary Attack |
| **LB / L1** | Cycle Weapons Left |
| **RB / R1** | Cycle Weapons Right |
| **LT / L2** | Block / Defend |
| **RT / R2** | Special Ability / Spell Cast |

#### Interface
| Button | Action |
|--------|--------|
| **D-Pad Up** | Inventory |
| **D-Pad Down** | Map |
| **D-Pad Left** | Quest Log |
| **D-Pad Right** | Character Sheet |
| **Start** | Main Menu |
| **Select / Back** | Toggle UI Visibility |

---

## Mobile Controls

**Platforms:** iOS, Android (touchscreen)

### Touch Gestures
| Gesture | Action |
|---------|--------|
| **Single Tap** | Move To / Select |
| **Double Tap** | Attack / Interact |
| **Long Press** | Context Menu |
| **Swipe** | Camera Rotation |
| **Pinch In/Out** | Zoom |
| **Two-Finger Tap** | Quick Action (use last ability) |

### Virtual Controls
- **Virtual Joystick** (bottom-left): Movement
- **Action Buttons** (bottom-right): Attack, Jump, Interact
- **Hotbar** (top): Quick access to items/spells (1-4 slots)
- **Menu Button** (top-left): Opens radial menu for inventory, map, etc.

### Mobile-Specific Features
- **Tap to Move:** Tap location to auto-path (optional, enabled by default)
- **Auto-Attack:** Toggle continuous attack on target (Settings → Combat → Auto-Attack)
- **One-Handed Mode:** UI layout adjusts for left/right hand preference

---

## WebAssembly Controls

**Browser Compatibility:** Chrome, Firefox, Safari, Edge (latest versions)

### Keyboard + Mouse
Same as [Desktop Controls](#desktop-controls), with the following caveats:

- **Fullscreen (F11):** Requires manual browser fullscreen (browser restrictions)
- **Alt Key:** May trigger browser menu (use Ctrl as alternative modifier)
- **Function Keys (F1-F12):** May conflict with browser shortcuts (rebind recommended)

### Touch (Tablet Browser)
Same as [Mobile Controls](#mobile-controls)

### Known Limitations
- **Clipboard Access:** Copy/paste restricted by browser security (use in-game text fields only)
- **Audio Latency:** ~50-100ms delay on some browsers (use desktop for best audio experience)
- **Pointer Lock:** May require re-clicking canvas after alt-tab

---

## Customization

### Rebinding Controls

**Desktop:**
1. Main Menu → Settings → Controls
2. Click on action to rebind
3. Press new key/button
4. Click "Save" to apply changes

**Mobile:**
1. Settings → Touch Controls
2. Enable "Custom Layout Mode"
3. Drag buttons to desired positions
4. Tap "Lock Layout" when done

### Control Presets
| Preset | Description |
|--------|-------------|
| **Default** | Optimized for general play |
| **WASD** | Traditional FPS layout (default) |
| **ESDF** | More accessible for small hands |
| **Southpaw** | Left-handed mouse users |
| **Gamepad Classic** | Traditional console-style mapping |
| **Custom** | User-defined bindings |

### Import/Export Bindings
- **Export:** Settings → Controls → Export (saves `.venturecontrols` JSON file)
- **Import:** Settings → Controls → Import (select `.venturecontrols` file)
- **Share:** Upload to community hub for others to download

---

## Accessibility Options

### Visual Assistance
- **High Contrast Mode:** Increase UI element contrast (Settings → Accessibility)
- **Colorblind Modes:** 3 options (Protanopia, Deuteranopia, Tritanopia)
- **Large Text:** 150% UI text scaling
- **Reduced Motion:** Disable camera shake and excessive particle effects

### Auditory Assistance
- **Subtitles:** Enable for all NPC dialog and sound effects
- **Visual Sound Indicators:** Directional indicators for off-screen sounds
- **Mono Audio:** Combine stereo channels (hearing impaired option)

### Motor Assistance
- **Auto-Aim:** Soft target assist (Settings → Combat → Auto-Aim Strength: 0-100%)
- **Sticky Keys:** Press once to toggle sprint/crouch (no hold required)
- **One-Button Mode:** Simplify combat to single button (mobile only)
- **Camera Speed:** Adjust camera rotation sensitivity (0.1x - 5.0x)

### Input Method Alternatives
- **Keyboard-Only Mode:** All actions accessible via keyboard (no mouse required)
- **Mouse-Only Mode:** Point-and-click interface (no keyboard required, limited features)
- **Voice Commands:** Experimental (requires microphone, opt-in beta)

---

## Tips & Tricks

### Efficient Hotbar Management
- **Bind frequently-used items** to 1-4 for quick access
- **Group similar items:** Potions on 1-2, offensive spells on 3-4, support on 5-6
- **Use Shift+Number** for alternate hotbar (16 total slots)

### Camera Control Mastery
- **Hold Right Click + Move Mouse:** Free camera rotation (no character movement)
- **Mouse Scroll:** Quick zoom for situational awareness (close combat vs. exploration)
- **R3 / Middle Click:** Snap camera behind character

### Multiplayer Communication
- **Ctrl+Enter:** Repeat last message (useful for coordinating in combat)
- **Tab Cycling:** Cycle through party members for quick whispers
- **Ping System:** Middle-click location to alert team (shows arrow on map)

### Combat Techniques
- **Animation Canceling:** Press dodge (B) immediately after attack lands to cancel recovery
- **Kiting:** Backpedal (S) while attacking to maintain distance
- **Combo Chaining:** Queue next attack during current animation for seamless combo

---

## Troubleshooting

### Controls Not Responding
1. **Verify Ebiten Focus:** Click game window to ensure it has keyboard/mouse focus
2. **Check Conflicting Software:** Disable overlays (Discord, Steam) that intercept inputs
3. **Reset to Defaults:** Settings → Controls → Reset to Default

### Gamepad Not Detected
1. **OS Recognition:** Ensure controller recognized by OS (test in OS settings)
2. **Restart Client:** Disconnect controller, close game, reconnect, relaunch
3. **Update Drivers:** Xbox controllers may need latest drivers on Windows

### Mobile Touch Unresponsive
1. **Screen Protector Interference:** Remove or replace with high-sensitivity protector
2. **Background Apps:** Close memory-intensive apps to free resources
3. **Update OS:** Ensure latest iOS/Android version

---

## Version History

### v1.0.0 (February 2026)
- Housing menu (H key)
- Gallery menu (G key)
- Enhanced gamepad support
- Full control reference aligned with implementation

---

## Related Documentation

- **[Getting Started](GETTING_STARTED.md):** First-time setup and tutorials
- **[User Manual](USER_MANUAL.md):** Comprehensive feature guide
- **[Accessibility](ACCESSIBILITY.md):** Detailed accessibility options
- **[Troubleshooting](TROUBLESHOOTING.md):** Common issues and solutions

---

**Questions?** See [FAQ](FAQ.md) or join our community Discord.

**Report Control Bugs:** [GitHub Issues](https://github.com/opd-ai/venture/issues) with `controls` label.
