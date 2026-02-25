# Venture Mobile Bug Audit & Remediation

## Objective
Identify and fix all gameplay-blocking bugs and critical defects in the Venture mobile builds (Android/iOS) by methodically testing the complete player journey from launch to endgame, prioritizing issues that prevent normal gameplay progression on touch-based devices.

## Execution Mode
**Autonomous Action** - Automatically detect and fix all discovered bugs without requiring approval.

## Platform Specifications

### Target Platforms
- **Android**: API 21+ (Android 5.0 Lollipop and later), ARM64/ARMv7
- **iOS**: iOS 12.0+, ARM64 (iPhone 6s and later, all iPad Air/Pro models)

### Build Verification
```bash
# Install mobile build tools
make mobile-deps

# Android builds
make android-verify          # Verify Android SDK/NDK setup
make android-apk             # Debug APK for testing
make android-apk-release     # Release APK (requires signing)
make android-aab             # Android App Bundle for Google Play
make android-install         # Build and install on connected device

# iOS builds (requires macOS)
make ios-xcframework         # Build iOS framework
make ios-simulator           # Build for iOS Simulator
make ios-device              # Build for iOS device
make ios-ipa                 # Build IPA for distribution
make ios-install             # Build and install on connected device
```

### Display Requirements
- **Screen Orientations**: Portrait, landscape, auto-rotation support
- **Screen Densities**: LDPI, MDPI, HDPI, XHDPI, XXHDPI, XXXHDPI (Android), @1x/@2x/@3x (iOS)
- **Safe Areas**: iPhone notch (X/11/12/13/14/15), Android punch-hole cameras
- **Aspect Ratios**: 16:9, 18:9, 19.5:9, 20:9 (various devices)

### Performance Targets
- **Frame Rate**: 30+ FPS on mid-range devices (Snapdragon 660/A11 Bionic)
- **Battery**: <20% drain per hour of gameplay
- **Memory**: <300MB on devices with 2GB+ RAM
- **Thermal**: No excessive heating during normal gameplay

## Testing Methodology: Player Journey Order

### Phase 1: Launch & Main Menu (5 minutes)

1. **Application Launch**
   - Build verification: `make android-apk` or `make ios-ipa`
   - App installs without errors (check adb logcat or Xcode console)
   - Launch succeeds without crashes
   - Splash screen displays correctly (if implemented)
   - App permissions requested (storage, notifications) at appropriate times

2. **Main Menu Navigation**
   - All menu buttons visible and functional (New Game, Load Game, Settings, Exit)
   - **Back button (Android) or swipe gesture (iOS)** returns to previous menu or exits app
   - Genre selection UI accessible and functional (scrollable list or grid)
   - Settings menu: graphics, audio, controls adjustable with touch
   - Exit paths work from all submenu states
   - **Touch target size**: Minimum 44x44 points (iOS HIG) or 48x48dp (Material Design)

3. **Mobile-Specific Display Tests**
   - **Orientation Handling**: Rotate device, verify UI reflows correctly (portrait ↔ landscape)
   - **Safe Area Insets**: UI elements avoid notch/punch-hole on iPhone X+ and Android devices
   - **Density Scaling**: UI elements appropriately sized on low-end (720p) and high-end (1440p) devices
   - **Status Bar**: System status bar visible or hidden as appropriate (fullscreen gameplay)

### Phase 2: Tutorial & First-Time Experience (5 minutes)

4. **Tutorial System**
   - Tutorial triggers on first launch or new game
   - All tutorial prompts display correctly (text readable on small screens)
   - Tutorial controls functional (tap to skip, swipe to advance, close button)
   - **Back button/swipe exits tutorial** without trapping player
   - Tutorial completion flag persists correctly in app data

5. **Initial Spawn**
   - Player entity spawns in walkable terrain
   - Camera centers on player
   - HUD elements render (health, mana, XP bar) in safe areas
   - **Virtual controls render**: Dual joysticks or d-pad + action buttons visible

### Phase 3: Core Gameplay Loop (10 minutes)

6. **Movement & Collision**
   - **Left virtual joystick** controls movement in all 8 directions
   - **Swipe gestures** (alternative input) move character
   - Collision detection prevents wall clipping
   - **Right virtual joystick** or **directional swipes** rotate player sprite (360° rotation)
   - **Touch controls responsive**: <100ms input latency
   - **Haptic feedback** on actions (if implemented, device-dependent)

7. **Combat System**
   - **Tap attack button** or **tap enemy** triggers attack animation
   - **Hold attack button** for charged attacks (if implemented)
   - Damage calculation applies to enemies
   - Player takes damage from enemy attacks
   - Health depletes correctly, death/revival system activates
   - Status effects apply and expire (visual indicators visible on small screens)
   - **Ability buttons (1-5 on-screen)** trigger learned spells

8. **Inventory System**
   - **Tap inventory icon** opens inventory menu
   - Items display with sprites and stats (text readable, icons clear)
   - **Drag-and-drop or tap-to-equip** works on touch screens
   - **Back button/swipe closes inventory** (dual-exit pattern)
   - Equipment slots update character stats in real-time
   - **Pinch-to-zoom** or **double-tap** for item detail view (optional)

9. **Item Pickup & Loot**
   - **Tap item** or **tap "E" button** interacts with chests/drops
   - Items added to inventory with pickup animation
   - Inventory full condition handled gracefully (toast notification)
   - Loot rarity colors display correctly (visible on all display densities)
   - **Long-press item** shows tooltip (if implemented)

### Phase 4: Progression Systems (10 minutes)

10. **Character Sheet**
    - **Tap character icon** or **swipe right** opens character menu
    - Stats display current values (STR, DEX, INT, VIT, etc.)
    - Level-up increases stats correctly
    - **Back button/swipe closes menu**
    - **Swipe left/right** cycles between character sheet tabs (if implemented)

11. **Skills & Abilities**
    - **Tap skills icon** opens skill tree menu
    - Skill nodes display with prerequisites (locked/unlocked visual states)
    - **Tap skill node** allocates skill points
    - **Ability buttons (1-5 on-screen)** trigger learned spells
    - Mana consumption and cooldowns function (visual indicators)
    - **Long-press skill** shows description (if implemented)
    - **Pinch-to-zoom** skill tree (if implemented)

12. **Quest System**
    - **Tap quest icon** opens quest log
    - Active quests display objectives with progress tracking
    - Quest progress updates on completion
    - Rewards granted (XP, items, gold) with notification
    - Completed quests marked correctly (visual checkmark)
    - **Tap quest** expands/collapses quest details
    - **Swipe to dismiss** completed quests (optional)

### Phase 5: World Interaction (5 minutes)

13. **NPC Interaction**
    - **Tap NPC** or **tap "F" button** triggers merchant/NPC dialog
    - Dialog UI displays text correctly (readable font size, text wrapping)
    - Shop interface allows buy/sell (tap items, confirm purchase)
    - Prices scale by rarity
    - Transaction validation works (insufficient gold shows error toast)
    - **Back button/swipe exits dialog** without locking player

14. **Crafting System**
    - **Tap crafting icon** opens crafting menu
    - Recipes display with requirements (materials + crafting station)
    - **Tap recipe** shows details, **tap "Craft"** initiates crafting
    - Item crafting consumes materials correctly
    - Crafted items added to inventory
    - Station proximity validation works (error message if too far)

15. **Map & Navigation**
    - **Tap map icon** or **pinch out gesture** opens map overlay
    - Explored areas visible, unexplored dark (fog of war)
    - Player position marker accurate and updates in real-time
    - **Back button/swipe closes map**
    - **Drag gesture** pans map
    - **Pinch gesture** zooms map

### Phase 6: Multiplayer (5 minutes)

16. **Connection & Sync**
    - **Note**: Mobile multiplayer may be limited or disabled depending on implementation
    - Server connection: Enter server IP in settings or use LAN discovery
    - Client connects to server (Wi-Fi or cellular data)
    - Player entities spawn for all clients
    - Movement synchronizes across clients
    - Combat damage replicates correctly
    - **Network indicator**: Signal strength/ping display

17. **Social Systems (V5+)**
    - **Tap chat icon** opens virtual keyboard for text input
    - Text messages send/receive in chat window (text size readable)
    - Trade system initiates (if implemented)
    - Player names display above entities (readable on small screens)

### Phase 7: Save/Load (5 minutes)

18. **Save System**
    - **Tap save button** or **auto-save** triggers without errors
    - Save file created in app data directory (Android: `getFilesDir()`, iOS: Documents)
    - Save includes all state (position, inventory, quests, skills)
    - Multiple save slots supported (named saves with thumbnails)
    - Save timestamp displays correctly

19. **Load System**
    - **Tap load button** restores saved state
    - Load Game menu lists save files with metadata (date, time, level)
    - Load preserves world seed (deterministic generation verified)
    - No state corruption on load (inventory, quests, skills intact)
    - **Back button/swipe exits load menu** without loading

## Mobile-Specific Tests

### Touch Input & Gestures
- **Virtual Joystick**: Left stick moves, right stick aims (dual joystick mode)
- **Alternative Layouts**: D-pad + buttons, swipe-based controls (configurable in settings)
- **Touch Precision**: Tap targets ≥44pt (iOS) / 48dp (Android)
- **Multi-Touch**: Simultaneous input (movement + attack + ability)
- **Gesture Conflicts**: Swipe vs. scroll, pinch vs. drag resolution
- **Touch Release**: Actions trigger on touch-up, not touch-down (prevents accidental inputs)

### Screen Orientation & Layout
- **Portrait Mode**: UI elements stacked vertically, joysticks at bottom
- **Landscape Mode**: UI elements in corners, joysticks on left/right edges
- **Auto-Rotation**: Smooth transition between orientations without UI glitches
- **Orientation Lock**: Settings option to lock to portrait or landscape

### Device-Specific Edge Cases
- **Notch/Cutout Devices**: iPhone X/11/12/13/14/15, Android punch-hole cameras
  - UI elements avoid obscured areas (use safe area insets)
  - Fullscreen mode respects safe area boundaries
- **Foldable Devices**: Samsung Galaxy Fold, Surface Duo
  - App continues when folding/unfolding
  - UI reflows to new aspect ratio
- **Tablet Support**: iPad, Android tablets (10"+ screens)
  - Larger touch targets, optional keyboard/mouse support

### Performance & Resource Management
- **Background/Foreground Transitions**:
  - App pauses when backgrounded (home button, notification, incoming call)
  - App resumes correctly when foregrounded
  - Audio pauses/resumes appropriately
  - Network connections maintained or gracefully reconnected
- **Low Memory Warnings**: App reduces cache/texture quality, doesn't crash
- **Battery Optimization**: No excessive CPU/GPU usage when idle
- **Thermal Throttling**: Frame rate degrades gracefully under thermal limits

### Platform-Specific Features
- **Android**:
  - Back button navigation (exits menus, then app with confirmation)
  - Recent apps multitasking (app thumbnail shows game screen)
  - Volume buttons control game audio (not system volume)
  - Notification system (quest completion, multiplayer invites)
- **iOS**:
  - Home indicator swipe gesture (bottom edge) doesn't conflict with game controls
  - Control Center swipe (top-right edge) pauses game
  - App clip or widget support (if implemented)
  - iCloud save sync (if implemented)

### Accessibility
- **Text Size**: Respects system text size settings (Android) or Dynamic Type (iOS)
- **Color Blindness**: Color-blind mode or high-contrast mode (optional)
- **Haptic Feedback**: Configurable in settings, respects system haptic settings
- **VoiceOver/TalkBack**: Basic screen reader support (menu navigation)

## Bug Categories & Priority

### Critical (P1 - Gameplay Blockers)
- **Menu Traps:** UI states where back button/swipe fails to exit
- **Broken Controls:** Non-functional touch areas, virtual joystick drift
- **Progression Blockers:** Infinite loops, deadlocks, unexitable states
- **Missing UI**: Touch targets too small (<44pt/48dp), buttons off-screen
- **Spawn Failures:** Player spawns in walls or out-of-bounds
- **Crash on Background**: App crashes when minimized or interrupted

### High Priority (P2 - Visual/UX Defects)
- **Rendering Failures:** Black screens, missing sprites, broken animations
- **UI Layout Issues:** Overlapping elements, text cutoff by notch, unreadable text
- **Input Edge Cases:** Touch not registering, gesture conflicts, multi-touch failures
- **Audio Failures:** Missing SFX, music playback errors, no audio after background
- **Orientation Bugs**: UI broken after rotation, safe area violations

### Medium Priority (P3 - System Bugs)
- **Memory Leaks:** Unbounded cache growth, textures not freed
- **Race Conditions:** Detected by `go test -race`
- **Save Corruption:** State persistence failures across sessions
- **Network Edge Cases:** Cellular data fails, Wi-Fi reconnection issues
- **Performance Degradation**: FPS drops below 30, battery drain >30%/hour

## Automated Detection

### Static Analysis
```bash
go vet ./...                           # Suspicious constructs
go test -race ./pkg/engine/... ./pkg/rendering/ui/... ./pkg/mobile/...  # Race conditions
```

### Pattern-Based Search
```bash
grep -rn "panic\|TODO.*block\|FIXME.*critical" pkg/ cmd/
grep -rn "for \{" pkg/engine/ pkg/rendering/     # Infinite loops
grep -rn "time\.Sleep.*second" pkg/engine/        # Blocking in game loop
grep -rn "ebiten\.NewImage" pkg/engine/ pkg/procgen/  # Ebiten calls in non-rendering code

# Mobile-specific checks
grep -rn "TouchID\|TouchPosition\|TouchState" pkg/mobile/
grep -rn "VirtualJoystick\|DualJoystick" pkg/mobile/
```

### Mobile-Specific Checks
```bash
# Check touch input handling
grep -rn "ebiten\.TouchIDs\|CursorPosition" pkg/mobile/

# Verify safe area handling (notch/cutout)
grep -rn "SafeAreaInsets\|WindowInsets" pkg/mobile/

# Check orientation handling
grep -rn "DeviceOrientation\|ScreenOrientation" pkg/mobile/
```

### Device Testing
- **Android**: Test on emulator (AVD) and physical devices (low-end, mid-range, flagship)
- **iOS**: Test on Simulator and physical devices (iPhone SE, iPhone 14, iPad)
- **Log Monitoring**:
  - Android: `adb logcat | grep venture`
  - iOS: Xcode Console or `idevicesyslog`

## Fix Requirements

### Code Quality
- Maintain ECS architecture (no logic in components)
- Preserve deterministic generation (seed-based RNG only)
- Follow dual-exit UI pattern (back button/swipe + close button)
- Ensure ≥40% test coverage after fixes (≥30% for packages depending on X11/Wayland/Ebiten; pkg/mobile/ currently high coverage)
- Pass `go test ./...` without errors

### Fix Documentation
Add inline comment for each fix:
```go
// BUG FIX: [Phase] - [Issue Description]
// Resolution: [What changed and why]
// Platform: Mobile (Android/iOS)
// Example:
// BUG FIX: Phase 3.8 - Inventory back button not bound to Hide()
// Resolution: Added BackButton() handler in mobile.UI for inventory menu
// Platform: Mobile (Android back button, iOS swipe gesture)
```

### Touch Input Guidelines (Reference: pkg/mobile/)
- Use `mobile.TouchManager` for touch event handling
- Implement rate limiting via `mobile.InputRateLimiter` to prevent spam
- Support both `DualJoystick` and alternative control schemes
- Minimum touch target: 44pt (iOS) / 48dp (Android)
- Touch feedback: Visual highlight + optional haptics

## Success Criteria

- ✅ Complete player journey tested (Phase 1-7) without blocks on Android and iOS
- ✅ All builds compile: `make android-apk` and `make ios-ipa`
- ✅ Full test suite passes: `go test ./...` (including pkg/mobile/)
- ✅ No race conditions: `go test -race ./...`
- ✅ All UI menus have functional exit paths (back button/swipe + close button)
- ✅ Core gameplay loop playable: launch → tutorial → combat → loot → progression → save/load
- ✅ Touch controls responsive (<100ms latency) on mid-range devices
- ✅ Performance targets met (30+ FPS, <300MB memory, <20% battery/hour)
- ✅ Orientation handling robust (portrait/landscape, no UI breaks)
- ✅ Safe area respected on notch/cutout devices
- ✅ Background/foreground transitions graceful (no crashes, state preserved)

## Constraints

- **Do Not Break:** Deterministic generation (same seed = same world)
- **Do Not Modify:** Public API signatures without version bump justification
- **Do Not Skip:** Test validation for each fix on both Android and iOS
- **Do Not Add:** New features or refactoring beyond bug fixes
- **Do Not Assume:** Device capabilities (check for notch, haptics, high-DPI at runtime)
- **Do Not Hardcode:** Screen dimensions or touch positions (use relative positioning)

## Execution Order

1. Fix Phase 1 (Launch & Main Menu) blockers on Android and iOS
2. Fix Phase 2 (Tutorial) blockers
3. Fix Phase 3 (Core Gameplay) blockers, prioritize touch input issues
4. Fix Phase 4 (Progression) issues
5. Fix Phase 5 (World Interaction) issues
6. Fix Phase 6 (Multiplayer) issues (if implemented for mobile)
7. Fix Phase 7 (Save/Load) issues
8. Fix Mobile-Specific Tests (touch, orientation, background/foreground, safe areas)
9. Validate all fixes with test suite on Android and iOS devices
10. Report summary of bugs found and fixed by phase and platform

## Cross-References

- **Mobile Build Process**: See `docs/MOBILE_BUILD.md` for detailed build instructions
- **Touch Input System**: See `pkg/mobile/touch.go` and `pkg/mobile/controls.go` for implementation
- **Cross-Platform Builds**: See `docs/CROSS_PLATFORM_BUILDS.md` for mobile-specific build flags
- **Performance Optimization**: See `docs/PERFORMANCE.md` for mobile performance tuning
- **Architecture**: See `docs/ARCHITECTURE.md` for ECS system design
