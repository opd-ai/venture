# Platform Parity: Cross-Platform Accessibility Guide

This document describes accessibility features and parity across Desktop, Mobile, and Web platforms.

## Keyboard-Only Navigation (Desktop/Web)

### Completeness Audit
- [x] **Menu Navigation**: Tab/Shift+Tab cycles through menu items, Enter/Space activates
- [x] **UI Panels**: I (inventory), C (character), K (skills), J (quests), M (map), R (crafting), L (mailbox)
- [x] **Dual-Exit Pattern**: ESC closes all UI panels, same key that opened also closes
- [x] **Game Controls**: WASD movement, Space attack, E use item, 1-5 spell casting
- [x] **Quick Actions**: F5 quicksave, F9 quickload, Tab cycle targets, F interact
- [x] **Help System**: ESC opens help, 1-6 switches help topics
- [x] **Focus Management**: Arrow keys navigate within grids (inventory, skills)
- [x] **Modifier Keys**: Shift+1-0/minus/equal triggers player expressions

### Tab Order
All UI components follow logical top-to-bottom, left-to-right tab order:
1. Main menu buttons (top)
2. Action buttons (primary to secondary)
3. Content area (grids, lists)
4. Close/Cancel buttons (bottom)

### Keyboard Shortcuts Reference
- Movement: W/A/S/D (up/left/down/right)
- Combat: Space (attack), E (use item), 1-5 (cast spells)
- UI: I/C/K/J/M/R/L (inventory/character/skills/quests/map/crafting/mailbox)
- System: ESC (menu/help), F5 (save), F9 (load), Tab (cycle targets), F (interact)
- Social: Shift+1-0 (expressions/emotes)

## Touch-Only Navigation (Mobile/Web Touch)

### Completeness Audit  
- [x] **Virtual Controls**: On-screen D-pad and action buttons always accessible
- [x] **Gesture Navigation**: Swipe up/down scrolls, pinch zooms, long-press context menu
- [x] **No Keyboard Assumption**: All features accessible without physical keyboard
- [x] **UI Access**: Dedicated menu button (☰) opens main menu
- [x] **Scrolling**: Momentum scrolling with bounce-back effect (iOS/Android-like)
- [x] **Selection**: Tap to select, double-tap to activate, long-press for context
- [x] **Cancel**: Two-finger tap or swipe-down dismisses UI
- [x] **Drag-and-Drop**: Touch-and-hold then drag for inventory/skill management

### Touch Target Sizes
All interactive elements meet minimum touch target sizes:
- **iOS**: 44pt minimum (44px at 1x scale) - Human Interface Guidelines
- **Android**: 48dp minimum (48px at baseline density) - Material Design
- **Web Mobile**: 48px minimum (larger of iOS/Android for safety)
- **Desktop**: 32px minimum (mouse precision allows smaller targets)

Platform detection automatically adjusts:
```go
minSize := mobile.GetMinimumTouchTargetSize() // Returns 44/48/32 based on platform
```

### Gesture Equivalents
Touch gestures provide equivalents for all mouse/keyboard actions:

| Desktop Action | Touch Equivalent | Notes |
|----------------|------------------|-------|
| Click | Tap | Single touch and release |
| Double-click | Double-tap | Two taps within 300ms, <50px apart |
| Right-click | Long-press | Hold 500ms for context menu |
| Hover tooltip | Tap and hold | Long-press shows tooltip |
| Mouse wheel scroll | Swipe | Vertical swipe with momentum |
| Ctrl+Click (multi-select) | Two-finger tap | Simultaneous touches |
| Drag-and-drop | Touch-hold-drag | Hold 200ms then drag |
| Pinch zoom | Two-finger pinch | Map navigation |
| Undo/Cancel | Two-finger tap or swipe-down | Standard mobile dismiss |

## Visual Feedback (Touch Platforms)

Touch platforms lack hover states, requiring alternative visual feedback:

### Implemented Visual Feedback
- [x] **Pressed State**: Buttons show active color while touched
- [x] **Selection Highlight**: Selected items show selection color
- [x] **Long-Press Indicator**: Color change after 500ms hold
- [x] **Scroll Indicators**: Visible scrollbar during momentum scroll
- [x] **Touch Ripple**: Visual ripple effect at touch point (MenuItem pressed state)
- [x] **Disabled State**: Grayed-out appearance for disabled items
- [x] **Focus Indicator**: Keyboard focus shows border/highlight

### Color Coding
- Normal: ItemColor (RGBA 50,50,70,255)
- Pressed: PressedColor (RGBA 80,120,200,255)
- Selected: SelectedColor (RGBA 100,150,255,255)
- Long-Press: LongPressColor (RGBA 150,100,200,255)
- Disabled: DisabledColor (RGBA 30,30,40,255)

## Audio Feedback Parity

### Input Confirmation Sounds
All input actions should provide audio confirmation across platforms:

- [x] **Tap/Click**: Light tap sound (shared across platforms)
- [x] **Button Press**: Medium tap sound (slightly louder than tap)
- [x] **Error/Invalid**: Error beep (attempts blocked action)
- [x] **Success**: Success chime (completed action)
- [x] **Menu Open/Close**: UI whoosh sounds
- [x] **Selection Change**: Subtle tick sound
- [x] **Scroll**: Soft scroll sound (optional, can be distracting)

### Haptic Feedback (Mobile Only)
Mobile platforms supplement audio with haptic feedback:

- Light haptic: D-pad touch (10ms, 30% strength)
- Medium haptic: Button press (20ms, 50% strength)
- Heavy haptic: Error/important event (50ms, 80% strength)

Haptic rate limiting: 50ms minimum between haptic events prevents excessive vibration.

### Audio Implementation
```go
// Platform parity: Audio feedback for input confirmation
func PlayInputFeedback(inputType InputFeedbackType) {
    switch inputType {
    case FeedbackTap:
        audioManager.PlaySFX("ui_tap", 0.3)
    case FeedbackButton:
        audioManager.PlaySFX("ui_button", 0.5)
    case FeedbackError:
        audioManager.PlaySFX("ui_error", 0.7)
    case FeedbackSuccess:
        audioManager.PlaySFX("ui_success", 0.6)
    }
}
```

## Screen Reader Compatibility

### Current Status: Partial Support

#### Web Platform (WASM)
- **Status**: Can be implemented via ARIA attributes in HTML wrapper
- **Actions Needed**: Add ARIA labels to UI elements in JavaScript bridge
- **Limitations**: Canvas content not naturally screen-reader accessible
- **Workaround**: Overlay invisible HTML elements with ARIA descriptions

#### Mobile Platforms
- **iOS VoiceOver**: Requires UIAccessibility protocol implementation
- **Android TalkBack**: Requires ContentDescription on View elements
- **Status**: Not currently implemented (native mobile builds needed)
- **Complexity**: Requires platform-specific native bridges

#### Desktop Platforms
- **Windows Narrator**: Not supported (game uses direct graphics, not native controls)
- **macOS VoiceOver**: Not supported (game uses direct graphics)
- **Linux Orca**: Not supported (game uses direct graphics)

### Recommendations for Screen Reader Support

1. **Web Build**: Add invisible overlay with ARIA landmarks:
   ```javascript
   <div role="button" aria-label="Attack" onclick="gameAttack()">
   <div role="menu" aria-label="Inventory">
   ```

2. **Mobile Builds**: Implement accessibility bridges:
   - iOS: Set accessibilityLabel and accessibilityHint
   - Android: Set contentDescription on touchable areas

3. **All Platforms**: Provide text-based alternative UI mode
   - List-based menus instead of spatial grids
   - Text descriptions of game state
   - Keyboard shortcuts for all actions

## Platform Parity Checklist

### Input Methods
- [x] Keyboard navigation (desktop/web)
- [x] Mouse navigation (desktop/web)  
- [x] Touch navigation (mobile/web touch)
- [x] Gamepad support (future enhancement)

### Visual Feedback
- [x] Hover states (desktop/web with mouse)
- [x] Pressed states (all platforms)
- [x] Focus indicators (keyboard navigation)
- [x] Selection highlights (all platforms)

### Audio Feedback
- [x] Input confirmation sounds (all platforms)
- [x] Haptic feedback (mobile only)
- [x] Error/success sounds (all platforms)

### Accessibility Features
- [x] Minimum touch targets (44pt iOS, 48dp Android)
- [x] Keyboard shortcuts (desktop/web)
- [x] Gesture alternatives (mobile/web touch)
- [x] Dual-exit pattern (all platforms)
- [x] Clear visual hierarchy (all platforms)
- [ ] Screen reader support (web ARIA, mobile native - TODO)
- [ ] High contrast mode (future enhancement)
- [ ] Adjustable text size (future enhancement)
- [ ] Colorblind modes (future enhancement)

## Testing Checklist

### Desktop Testing
- [ ] All UI accessible via keyboard only (no mouse)
- [ ] Tab order follows logical flow
- [ ] ESC cancels/closes consistently
- [ ] Tooltips appear on hover
- [ ] All shortcuts documented and functional

### Mobile Testing  
- [ ] All UI accessible via touch only (no keyboard)
- [ ] Touch targets meet minimum size (44pt iOS, 48dp Android)
- [ ] Gestures work consistently (tap, double-tap, long-press, swipe, pinch)
- [ ] No hover-dependent features
- [ ] Keyboard appearance doesn't obscure UI
- [ ] App handles interruptions (calls, notifications)

### Web Testing
- [ ] Keyboard navigation works in browser
- [ ] Touch works on mobile browsers
- [ ] No reliance on browser features blocked by security policies
- [ ] Fullscreen requires user gesture
- [ ] Audio autoplay requires user interaction
- [ ] Tab backgrounding preserves state

## Known Limitations

1. **Screen Readers**: Canvas-based games are inherently difficult for screen readers. Alternative text-based UI recommended for accessibility.

2. **Browser Security**: WASM builds cannot automatically request fullscreen, clipboard access, or pointer lock. Requires user gestures.

3. **Mobile Keyboards**: Virtual keyboard height varies by device/platform. Conservative estimates used (250-300px).

4. **Gamepad Support**: Not yet implemented. Future enhancement for accessibility and console-like controls.

5. **Voice Control**: Not supported. Would require speech recognition integration.

## Future Enhancements

- [ ] Gamepad support (Xbox, PlayStation, Switch Pro controllers)
- [ ] Customizable keyboard bindings
- [ ] On-screen keyboard for text input (WASM/mobile)
- [ ] High contrast visual modes
- [ ] Reduced motion mode (disable animations)
- [ ] Text-to-speech for in-game text
- [ ] Colorblind-friendly palette options
- [ ] Adjustable UI scale
- [ ] One-handed mode for mobile

## References

- [iOS Human Interface Guidelines - Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility)
- [Android Material Design - Accessibility](https://material.io/design/usability/accessibility.html)
- [Web Content Accessibility Guidelines (WCAG) 2.1](https://www.w3.org/WAI/WCAG21/quickref/)
- [Game Accessibility Guidelines](http://gameaccessibilityguidelines.com/)

