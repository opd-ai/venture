# Venture WebAssembly Bug Audit & Remediation

## Objective
Identify and fix all gameplay-blocking bugs and critical defects in the Venture WebAssembly (WASM) build by methodically testing the complete player journey from launch to endgame, prioritizing issues that prevent normal gameplay in web browsers.

## Execution Mode
**Autonomous Action** - Automatically detect and fix all discovered bugs without requiring approval.

## Platform Specifications

### Target Browsers
- **Chrome/Chromium**: Version 90+ (recommended)
- **Firefox**: Version 88+
- **Safari**: Version 14.1+ (macOS/iOS)
- **Edge**: Version 90+ (Chromium-based)

### Build Verification
```bash
# Build WebAssembly version
make build-wasm

# Serve locally for testing
make serve-wasm
# Opens http://localhost:8080 in default browser

# Manual build steps
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/wasm/venture.wasm ./cmd/client
cp $(go env GOROOT)/lib/wasm/wasm_exec.js build/wasm/
cd build/wasm && python3 -m http.server 8080
```

### Deployment Verification
```bash
# Verify GitHub Pages deployment (reference: docs/GITHUB_PAGES.md)
# 1. Push to main branch triggers automatic deployment
# 2. Check GitHub Actions: .github/workflows/pages.yml
# 3. Access deployed version: https://<username>.github.io/<repository>/
# Example: https://opd-ai.github.io/venture/
```

### Browser Requirements
- **WebAssembly Support**: wasm32 architecture
- **WebGL/Canvas**: 2D rendering context
- **Web Audio API**: AudioContext for sound synthesis
- **Local Storage**: IndexedDB (preferred) or localStorage fallback (5MB limit)
- **Fullscreen API**: For immersive gameplay
- **Pointer Lock API**: For mouse capture (optional)

### Performance Targets
- **Frame Rate**: 30+ FPS on desktop browsers, 20+ FPS on mobile browsers
- **Memory**: <512MB WASM heap (browser limit)
- **Loading Time**: <5s initial load on broadband (2+ Mbps)
- **Bundle Size**: <10MB total (venture.wasm + wasm_exec.js)

## Testing Methodology: Player Journey Order

### Phase 0: HTML/CSS Validation & Mobile Web Audit (10 minutes)

**HTML Structure Validation**

1. **index.html Audit**
   - Valid HTML5 doctype: `<!DOCTYPE html>`
   - `<meta charset="UTF-8">` present
   - **Viewport meta tag**: `<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">`
     - `width=device-width`: Responsive to device width
     - `initial-scale=1.0`: No initial zoom
     - `maximum-scale=1.0, user-scalable=no`: Prevent pinch-zoom (game controls)
     - `viewport-fit=cover`: Extend to safe area on notch devices
   - **PWA meta tags**: `<meta name="apple-mobile-web-app-capable">`, `<meta name="mobile-web-app-capable">`
   - Semantic HTML: Proper heading hierarchy (h1, h2, h3)
   - Accessible elements: `alt` text on images, `aria-label` on buttons
   - Valid CSS: No syntax errors in `<style>` blocks

2. **game.html Audit**
   - Same viewport and PWA meta tags as index.html
   - **Canvas element**: `<canvas id="gameCanvas">` with fallback content
   - **Loading indicator**: Visible during WASM initialization
   - **Error messages**: Graceful degradation for unsupported browsers
   - **Hidden input element**: For mobile keyboard activation (if implemented)

**CSS Mobile Compatibility Audit**

3. **Responsive Layout**
   - **Flexbox/Grid**: Modern layout methods for cross-device compatibility
   - **Viewport units**: `vw`, `vh`, `vmin`, `vmax` for responsive sizing
   - **Media queries**: Breakpoints for mobile (`max-width: 768px`), tablet (`max-width: 1024px`), desktop
   - **Orientation queries**: `@media (orientation: portrait)` and `@media (orientation: landscape)`
   - **Safe area CSS**:
     ```css
     padding: env(safe-area-inset-top) env(safe-area-inset-right) 
              env(safe-area-inset-bottom) env(safe-area-inset-left);
     ```

4. **Touch-Friendly Styles**
   - **Touch targets**: Minimum 44x44px (iOS) or 48x48px (Android)
   - **Touch-action property**: `touch-action: none` on canvas (prevent browser gestures)
   - **User-select**: `-webkit-user-select: none` to prevent text selection
   - **Tap highlight**: `-webkit-tap-highlight-color: transparent` for cleaner UI
   - **No hover-only interactions**: Ensure all hover states have touch equivalents

5. **Canvas Rendering Styles**
   - **Image rendering**: `image-rendering: pixelated` or `crisp-edges` for pixel art
   - **Canvas sizing**: Responsive canvas with `max-width: 100vw`, `max-height: 100vh`
   - **Aspect ratio**: `aspect-ratio: 16/9` or maintain aspect ratio with padding-bottom trick
   - **High-DPI support**: CSS scaling for Retina/4K displays

6. **Performance CSS**
   - **Hardware acceleration**: `transform: translateZ(0)` or `will-change: transform` for smooth animations
   - **Avoid expensive properties**: Minimize `box-shadow`, `filter`, `backdrop-filter` on mobile
   - **Reduce reflows**: Fixed positioning for HUD elements

**Mobile Web-Specific Tests**

7. **iOS Safari Compatibility**
   - **Address bar auto-hide**: Content not hidden by Safari UI chrome
   - **Safe area insets**: Notch/home indicator don't obscure UI (iPhone X+)
   - **Standalone mode**: Test "Add to Home Screen" (appears as native app)
   - **Touch callouts**: `-webkit-touch-callout: none` prevents long-press menu
   - **Momentum scrolling**: `-webkit-overflow-scrolling: touch` for smooth scrolling (if scrollable areas)

8. **Android Chrome Compatibility**
   - **Address bar behavior**: Content adjusts when address bar shows/hides
   - **Theme color**: `<meta name="theme-color" content="#000000">` matches app design
   - **PWA install prompt**: "Add to Home screen" banner appears (if manifest.json present)
   - **Chrome custom tabs**: External links open in custom tabs, not new browser window

9. **Touch Input Conflicts**
   - **Canvas touch events**: `touchstart`, `touchmove`, `touchend` not blocked by CSS
   - **Prevent default**: `event.preventDefault()` on canvas to stop browser scrolling
   - **Passive listeners**: Use passive event listeners for scroll performance
   - **Input focus**: Hidden input can receive focus for keyboard (mobile keyboard activation)

**HTML/CSS Validation Tools**

10. **Automated Validation**
    ```bash
    # W3C HTML Validator (requires internet or local validator)
    curl -H "Content-Type: text/html; charset=utf-8" \
         --data-binary @build/wasm/index.html \
         https://validator.w3.org/nu/?out=text
    
    curl -H "Content-Type: text/html; charset=utf-8" \
         --data-binary @build/wasm/game.html \
         https://validator.w3.org/nu/?out=text
    
    # CSS Validation
    # Extract CSS from HTML and validate (manual check in browser DevTools)
    # Chrome DevTools → Elements → Styles → Check for warnings
    
    # Lighthouse Mobile Audit
    # Chrome DevTools → Lighthouse → Mobile → Performance + Accessibility
    # Target scores: Performance >80, Accessibility >90
    ```

11. **Manual Mobile Testing**
    - **Real devices**: Test on actual iPhone (Safari) and Android (Chrome) devices
    - **Browser DevTools**: Chrome DevTools → Device Toolbar (Ctrl+Shift+M)
      - Test various device presets (iPhone SE, Pixel 5, iPad, etc.)
      - Rotate orientation (portrait ↔ landscape)
      - Throttle network (Slow 3G, Fast 3G, Offline)
      - Throttle CPU (4x slowdown for low-end devices)
    - **Responsive Design Mode**: Firefox DevTools → Responsive Design Mode
    - **Safari Developer Tools**: macOS Safari → Develop → Enter Responsive Design Mode

12. **Common Mobile Web Issues**
    - **Viewport not set**: Page zoomed out, requires pinch-zoom
    - **Touch targets too small**: Buttons <44px, hard to tap accurately
    - **Fixed positioning bugs**: iOS Safari fixed elements jump on scroll
    - **100vh bug**: `height: 100vh` includes address bar (use `height: 100svh` or JavaScript)
    - **Input zoom on focus**: iOS Safari zooms if input font-size <16px
    - **Landscape orientation**: UI elements off-screen or overlapping
    - **Notch/cutout overlaps**: UI hidden by iPhone notch or Android camera cutout

### Phase 1: Launch & Main Menu (5 minutes)

1. **Application Launch**
   - Navigate to deployed URL or `http://localhost:8080`
   - Page loads without errors (check browser console: F12 → Console)
   - **HTML validation**: No markup errors, proper viewport meta tag
   - **CSS validation**: No style warnings, touch-action properly applied
   - **Loading screen** displays during WASM initialization
   - **Audio context activation**: Click anywhere or press key to enable audio (browser security requirement)
   - Canvas renders game at correct resolution (auto-scales to window)

2. **Main Menu Navigation**
   - All menu buttons visible and functional (New Game, Load Game, Settings, Exit)
   - **ESC key** returns to previous menu or shows exit confirmation
   - **Mouse click** navigates menus
   - Genre selection UI accessible and functional
   - Settings menu: graphics, audio, controls adjustable
   - Exit paths work from all submenu states
   - **Browser back button** doesn't navigate away from game (history.pushState handling)

3. **Browser-Specific Display Tests**
   - **Canvas Sizing**: Resize browser window, canvas scales correctly
   - **Fullscreen Mode**: Click fullscreen button or press F11, game enters fullscreen
   - **High-DPI Displays**: Retina/4K displays render without blurring (devicePixelRatio handling)
   - **Mobile Browser**: Test on Chrome Mobile/Safari iOS, touch controls activate
   - **Browser Zoom**: Ctrl+/- or pinch-zoom doesn't break layout

### Phase 2: Tutorial & First-Time Experience (5 minutes)

4. **Tutorial System**
   - Tutorial triggers on first launch or new game
   - All tutorial prompts display correctly
   - Tutorial controls functional (can skip, advance, close)
   - **ESC exits tutorial** without trapping player
   - Tutorial completion flag persists in IndexedDB/localStorage

5. **Initial Spawn**
   - Player entity spawns in walkable terrain
   - Camera centers on player
   - HUD elements render (health, mana, XP bar)
   - **Keyboard controls responsive**: WASD movement, mouse aim
   - **Mobile touch controls**: Virtual joysticks visible on touch devices

### Phase 3: Core Gameplay Loop (10 minutes)

6. **Movement & Collision**
   - **WASD movement** in all 8 directions
   - **Arrow keys** as alternative movement input
   - Collision detection prevents wall clipping
   - **Mouse aim** rotates player sprite (360° rotation system)
   - **Mouse cursor** displays correctly (canvas captures cursor in fullscreen)
   - **Touch controls** (mobile browsers): Left joystick moves, right joystick aims
   - Frame rate stable at 30+ FPS (check with browser DevTools: Performance tab)

7. **Combat System**
   - **Space bar or Left Mouse Click** triggers attack animation
   - **Right Mouse Click** for secondary attack (if implemented)
   - Damage calculation applies to enemies
   - Player takes damage from enemy attacks
   - Health depletes correctly, death/revival system activates
   - Status effects apply and expire (visual indicators present)
   - **Number keys 1-5** or **on-screen ability buttons** (touch) trigger learned spells

8. **Inventory System**
   - **I key** or **tap inventory icon** (touch) opens inventory menu
   - Items display with sprites and stats
   - **Mouse drag-and-drop** or **touch drag** or click/tap-to-equip works
   - **ESC or back button closes inventory** (dual-exit pattern)
   - Equipment slots update character stats in real-time
   - Inventory grid scrolls with mouse wheel or touch swipe

9. **Item Pickup & Loot**
   - **E key** or **tap item** interacts with chests/drops
   - Items added to inventory with pickup animation
   - Inventory full condition handled gracefully (notification displayed)
   - Loot rarity colors display correctly (Common/Uncommon/Rare/Epic/Legendary)
   - **Mouse hover** or **long-press (touch)** shows item tooltips

### Phase 4: Progression Systems (10 minutes)

10. **Character Sheet**
    - **C key** or **tap character icon** opens character menu
    - Stats display current values (STR, DEX, INT, VIT, etc.)
    - Level-up increases stats correctly
    - **ESC or back button closes menu**
    - **Tab key** or **swipe** cycles between character sheet tabs (if implemented)

11. **Skills & Abilities**
    - **K key** or **tap skills icon** opens skill tree menu
    - Skill nodes display with prerequisites (locked/unlocked visual states)
    - **Mouse click** or **tap** allocates skill points
    - **Number keys 1-5** or **on-screen buttons** trigger learned spells
    - Mana consumption and cooldowns function (visual indicators)
    - **Mouse hover** or **long-press** shows skill descriptions

12. **Quest System**
    - **J key** or **tap quest icon** opens quest log
    - Active quests display objectives with progress tracking
    - Quest progress updates on completion
    - Rewards granted (XP, items, gold) with notification
    - Completed quests marked correctly (visual checkmark)
    - **Mouse click** or **tap** expands/collapses quest details

### Phase 5: World Interaction (5 minutes)

13. **NPC Interaction**
    - **F key** or **tap NPC** triggers merchant/NPC dialog
    - Dialog UI displays text correctly (readable font, no overflow)
    - Shop interface allows buy/sell (mouse click or tap on items)
    - Prices scale by rarity
    - Transaction validation works (insufficient gold prevents purchase)
    - **ESC or back button exits dialog** without locking player

14. **Crafting System**
    - **R key** or **tap crafting icon** opens crafting menu
    - Recipes display with requirements (materials + crafting station)
    - **Mouse click** or **tap** on recipe initiates crafting
    - Item crafting consumes materials correctly
    - Crafted items added to inventory
    - Station proximity validation works (error message if too far)

15. **Map & Navigation**
    - **M key** or **tap map icon** opens map overlay
    - Explored areas visible, unexplored dark (fog of war)
    - Player position marker accurate and updates in real-time
    - **ESC or back button closes map**
    - **Mouse drag** or **touch drag** pans map (if implemented)
    - **Mouse wheel** or **pinch gesture** zooms map (if implemented)

### Phase 6: Multiplayer (5 minutes)

16. **Connection & Sync**
    - **Note**: WebSocket connections may be restricted by CORS/SSL in browsers
    - **Local Testing**: Server on `ws://localhost:8080`, client connects
    - **Production**: Requires WSS (secure WebSocket) for HTTPS sites
    - Client connects: Enter server URL in settings or use default
    - Player entities spawn for all clients
    - Movement synchronizes across clients (tested with multiple browser tabs)
    - Combat damage replicates correctly
    - **Network indicator**: Ping/connection status displayed

17. **Social Systems (V5+)**
    - **Enter key** or **tap chat icon** opens text input
    - Text messages send/receive in chat window
    - Trade system initiates (if implemented)
    - Player names display above entities

### Phase 7: Save/Load (5 minutes)

18. **Save System**
    - **F5 quick save** or **tap save button** triggers without errors
    - Save data stored in **IndexedDB** (preferred) or **localStorage** (5MB fallback)
    - **Browser Storage Check**: DevTools → Application → IndexedDB/Local Storage
    - Save includes all state (position, inventory, quests, skills)
    - Multiple save slots supported (named saves)
    - Save timestamp displays correctly

19. **Load System**
    - **F9 quick load** or **tap load button** restores saved state
    - Load Game menu lists save files with metadata (date, time, level)
    - Load preserves world seed (deterministic generation verified)
    - No state corruption on load (inventory, quests, skills intact)
    - **ESC or back button exits load menu** without loading

### Browser-Specific Tests

### Mobile Web Browser Tests

**iOS Safari Mobile**
- **Address Bar Auto-Hide**: Scroll down to hide Safari chrome, verify game area expands
- **Safe Area Insets**: On iPhone X+ (notch devices), verify UI doesn't overlap notch/home indicator
- **Standalone Mode (PWA)**:
  - Add to Home Screen from Safari share menu
  - Launch from home screen icon
  - Verify full-screen mode (no Safari UI)
  - Status bar color matches theme
- **Touch Gestures**:
  - Single tap: Select/interact
  - Double tap: Zoom prevention (should not zoom)
  - Pinch: Zoom prevention on canvas (should not zoom game)
  - Swipe: Edge swipes for back navigation (should not interfere with game)
- **Keyboard Input**:
  - Tap text field (if implemented), keyboard appears
  - Keyboard doesn't obscure critical UI
  - Dismiss keyboard (tap outside), game resumes
- **Orientation Changes**:
  - Rotate device portrait → landscape
  - UI reflows correctly, no broken layouts
  - Canvas maintains aspect ratio
- **Performance**:
  - Frame rate: 20+ FPS on iPhone 8+
  - Smooth scrolling (if scrollable areas)
  - No lag on touch input (<100ms)

**Android Chrome Mobile**
- **Address Bar Behavior**: Scroll to show/hide address bar, game area adjusts smoothly
- **Theme Color**: Address bar color matches `<meta name="theme-color">`
- **PWA Install Prompt**: "Add to Home screen" banner appears (if manifest.json exists)
- **Touch Gestures**:
  - Same as iOS Safari tests above
  - Pull-to-refresh disabled on game canvas
- **Keyboard Input**:
  - Virtual keyboard activation
  - Hardware keyboard (if connected via USB/Bluetooth)
- **Orientation Lock**: Test auto-rotate and manual orientation lock
- **Performance**:
  - Frame rate: 20+ FPS on mid-range devices (Snapdragon 660+)
  - Battery: <25% drain per hour

**Tablet Browsers (iPad/Android Tablets)**
- **Larger Canvas**: Game utilizes larger screen real estate
- **Touch Targets**: Still functional at larger sizes (no upscaling issues)
- **Split-Screen Mode**: Game continues when in split-screen multitasking
- **External Input**: Bluetooth mouse/keyboard support (optional)

**Mobile Web Edge Cases**
- **Low-End Devices**:
  - Test on devices with <2GB RAM
  - Verify game doesn't crash, degrades gracefully (lower FPS acceptable)
- **Slow Network**:
  - Throttle to Slow 3G (DevTools)
  - Verify loading screen shows progress
  - Timeout handling for WASM download
- **Offline Mode**:
  - Load game once, go offline
  - Refresh page, verify service worker caches assets (if implemented)
- **Interrupted Loading**:
  - Close tab during WASM loading
  - Reopen, verify recovers gracefully
- **Background Tab**:
  - Switch to another tab, wait 5+ minutes
  - Return, verify game resumes or shows resume prompt

### WebAssembly & JavaScript Interop
- **Console Errors**: No errors in browser console (F12 → Console)
- **WASM Initialization**: `wasm_exec.js` loads correctly, Go runtime starts
- **Memory Management**: WASM heap doesn't exceed 512MB (check in DevTools → Memory)
- **Garbage Collection**: No excessive GC pauses causing frame drops

### Audio Context Activation
- **User Gesture Requirement**: Audio context requires click/keypress before playing sounds
- **Audio Playback**: Music and SFX play correctly after activation
- **Audio Suspend/Resume**: Switching tabs pauses audio, returning resumes it
- **Volume Control**: In-game volume settings work independently of system volume

### Canvas & Rendering
- **Canvas Element**: `<canvas>` element renders game graphics
- **2D Context**: `getContext('2d')` used for Ebiten rendering
- **High-DPI**: `devicePixelRatio` handled correctly for crisp rendering
- **Fullscreen API**: `requestFullscreen()` works without breaking rendering
- **Pointer Lock**: `requestPointerLock()` captures mouse for camera control (optional)

### Storage & Persistence
- **IndexedDB**:
  - Save data persists across browser sessions
  - Quota limits respected (browsers typically allow 50MB+)
  - Error handling for quota exceeded errors
- **localStorage Fallback**:
  - Used if IndexedDB unavailable
  - 5MB limit enforced (warn user if save data large)
  - JSON serialization of game state

### Performance & Optimization
- **Loading Optimization**:
  - Lazy loading of assets (don't load all sprites upfront)
  - WASM streaming compilation (`WebAssembly.instantiateStreaming`)
  - Compression: gzip/brotli encoding for `venture.wasm` (server-side)
- **Runtime Performance**:
  - Profiling with Chrome DevTools (Performance tab)
  - Target: <33ms frame time (30 FPS minimum)
  - Memory profiling (heap snapshots, allocation timeline)

### Browser Compatibility
- **Chrome/Edge (Chromium)**:
  - Full WebAssembly support
  - High-performance WebGL/Canvas rendering
  - IndexedDB with generous quota
- **Firefox**:
  - WebAssembly support (may be slightly slower than Chrome)
  - IndexedDB support
  - Test canvas rendering performance
- **Safari (macOS/iOS)**:
  - WebAssembly support (version 14.1+)
  - IndexedDB quota restrictions (50MB typical)
  - Canvas performance lower than Chrome/Firefox
  - Audio context limitations (requires user gesture)
  - iOS Safari: Test in standalone mode (Add to Home Screen)

### Network & CORS
- **GitHub Pages Deployment**:
  - HTTPS required for production (secure context)
  - CORS headers for assets (typically allowed on same origin)
  - WebSocket multiplayer requires WSS (secure WebSocket)
- **Local Development**:
  - `http://localhost:8080` allowed (secure context)
  - CORS not an issue for same-origin requests
  - WebSocket `ws://` allowed locally

### Browser-Specific Edge Cases
- **Tab Visibility**:
  - Game pauses when tab not visible (Page Visibility API)
  - Resumes when tab focused again
  - Network connections maintained (WebSocket keepalive)
- **Browser Back/Forward**:
  - Back button doesn't navigate away from game (history API)
  - Forward button doesn't break game state
- **Browser Refresh**:
  - F5 refresh loses unsaved progress (warn user or auto-save)
  - Hard refresh (Ctrl+F5) clears WASM cache
- **Private/Incognito Mode**:
  - IndexedDB may be disabled or cleared on close
  - Warn user about save data limitations

## Bug Categories & Priority

### Critical (P1 - Gameplay Blockers)
- **Menu Traps:** UI states where ESC fails to exit
- **Broken Controls:** Non-functional keybinds (WASD, Space, E, F, etc.) or touch controls
- **Progression Blockers:** Infinite loops, deadlocks, unexitable states
- **Missing Menus:** Advertised features without accessible UI
- **Spawn Failures:** Player spawns in walls or out-of-bounds
- **WASM Crashes:** Uncaught exceptions, memory exhaustion, runtime panics
- **Audio Failures:** Audio context never activates, no sound playback
- **Mobile Web Blockers:**
  - Viewport not set (page requires pinch-zoom to use)
  - Touch targets <44px (impossible to tap on mobile)
  - Canvas not responsive (fixed size, off-screen on mobile)
  - Safe area violations (UI hidden by notch on iPhone X+)
  - Orientation change breaks layout completely

### High Priority (P2 - Visual/UX Defects)
- **Rendering Failures:** Black canvas, missing sprites, broken animations
- **UI Layout Issues:** Overlapping elements, off-canvas buttons, unreadable text
- **Input Edge Cases:** Keyboard/mouse conflicts, touch gesture failures
- **Storage Failures:** IndexedDB quota exceeded, localStorage 5MB limit hit
- **Fullscreen Bugs:** Canvas broken in fullscreen, ESC exits fullscreen instead of menu
- **Mobile Web UX Issues:**
  - Text too small to read on mobile (<16px, causes zoom-on-focus)
  - Buttons too close together (fat-finger errors)
  - Landscape orientation unusable (UI off-screen)
  - Address bar doesn't auto-hide (game area too small)
  - Keyboard obscures game UI (text input overlap)

### Medium Priority (P3 - System Bugs)
- **Memory Leaks:** WASM heap growth beyond 512MB
- **Performance Degradation**: FPS drops below 20, input latency >200ms
- **Save Corruption:** State persistence failures in IndexedDB/localStorage
- **Network Edge Cases:** WebSocket disconnects, CORS errors, WSS certificate issues
- **Browser Quirks**: Safari-specific bugs, Firefox rendering differences

### Automated Detection

### HTML/CSS Validation
```bash
# Validate HTML structure
grep -E "<!DOCTYPE html>|<meta.*viewport|<meta.*charset" build/wasm/*.html

# Check for required meta tags
grep "viewport-fit=cover" build/wasm/*.html  # Safe area support
grep "user-scalable=no" build/wasm/*.html    # Zoom prevention
grep "apple-mobile-web-app-capable" build/wasm/*.html  # iOS PWA

# Validate CSS touch properties
grep -E "touch-action|user-select|-webkit-tap-highlight" build/wasm/*.html

# Check canvas styling
grep -E "image-rendering.*pixelated|crisp-edges" build/wasm/*.html

# Lighthouse CI (if installed)
lighthouse http://localhost:8080 \
  --only-categories=performance,accessibility,best-practices,pwa \
  --emulated-form-factor=mobile \
  --output=html \
  --output-path=./lighthouse-mobile-report.html
```

### Mobile Compatibility Checks
```bash
# Check for mobile-unfriendly patterns
grep -rn "hover:" build/wasm/*.html  # Hover-only interactions (should have touch alternative)
grep -rn "cursor: pointer" build/wasm/*.html  # Ensure touch-action also set

# Verify viewport units usage
grep -E "vw|vh|vmin|vmax|svh|dvh" build/wasm/*.html

# Check for safe area CSS
grep "env(safe-area-inset" build/wasm/*.html

# Validate touch target sizes (manual check in DevTools)
# Chrome DevTools → Settings → Show rulers → Measure elements (should be ≥44px)
```

### Static Analysis
```bash
go vet ./...                           # Suspicious constructs
go test -race ./pkg/engine/... ./pkg/rendering/ui/...  # Race conditions (note: race detector not available in WASM)
```

### Pattern-Based Search
```bash
grep -rn "panic\|TODO.*block\|FIXME.*critical" pkg/ cmd/
grep -rn "for \{" pkg/engine/ pkg/rendering/     # Infinite loops
grep -rn "time\.Sleep.*second" pkg/engine/        # Blocking in game loop
grep -rn "ebiten\.NewImage" pkg/engine/ pkg/procgen/  # Ebiten calls in non-rendering code

# WASM-specific checks
grep -rn "syscall/js" pkg/                         # JavaScript interop
grep -rn "localStorage\|IndexedDB" pkg/saveload/   # Browser storage APIs
```

### Browser DevTools Checks
```javascript
// Open browser console (F12) and run:

// Check WASM initialization
console.log(WebAssembly.validate(new Uint8Array([0, 97, 115, 109, 1, 0, 0, 0])));  // Should return true

// Check canvas rendering
document.querySelector('canvas').getContext('2d');  // Should return CanvasRenderingContext2D

// Check IndexedDB
indexedDB.databases();  // Should list 'venture' database if saves exist

// Check localStorage
Object.keys(localStorage).filter(k => k.startsWith('venture_'));  // Should list save keys

// Monitor memory usage
performance.memory.usedJSHeapSize / 1024 / 1024;  // Current MB usage
performance.memory.jsHeapSizeLimit / 1024 / 1024;  // Max MB limit
```

### Network Monitoring
- **Chrome DevTools → Network**: Check `venture.wasm` loads (200 OK), size <10MB
- **Chrome DevTools → WebSockets**: Monitor WebSocket connections (if multiplayer enabled)
- **CORS Errors**: Check console for "Access-Control-Allow-Origin" errors

## Fix Requirements

### Code Quality
- Maintain ECS architecture (no logic in components)
- Preserve deterministic generation (seed-based RNG only)
- Follow dual-exit UI pattern (ESC + back button/click)
- Ensure ≥40% test coverage after fixes (≥30% for packages depending on X11/Wayland/Ebiten)
- Pass `go test ./...` without errors (note: some tests may not run in WASM environment)

### Fix Documentation
Add inline comment for each fix:
```go
// BUG FIX: [Phase] - [Issue Description]
// Resolution: [What changed and why]
// Platform: WASM (Browser)
// Example:
// BUG FIX: Phase 3.7 - Inventory ESC key not bound to Hide()
// Resolution: Added inpututil.IsKeyJustPressed(ebiten.KeyEscape) check in Update()
// Platform: WASM (all browsers)
```

### WASM-Specific Guidelines
- **Avoid Blocking Operations**: No `time.Sleep()` in main thread (use goroutines or frame-based delays)
- **Memory Management**: Free resources aggressively, target <512MB heap
- **Storage Limits**: Warn users if save data >5MB (localStorage limit), use IndexedDB for larger saves
- **Audio Context**: Defer audio initialization until user gesture (click/keypress)
- **Fullscreen API**: Handle fullscreen change events (user can ESC to exit)

## Success Criteria

- ✅ Complete player journey tested (Phase 0-7) without blocks on Chrome, Firefox, Safari (desktop + mobile)
- ✅ Build compiles: `make build-wasm`
- ✅ Full test suite passes: `go test ./...` (excluding WASM-incompatible tests)
- ✅ No race conditions: `go test -race ./...` (on non-WASM platforms)
- ✅ **HTML/CSS validation passed** (Phase 0):
  - Valid HTML5 markup (DOCTYPE, charset, viewport meta tag)
  - Mobile-friendly viewport configuration
  - Touch-action CSS properly applied to canvas
  - Safe area insets respected (iPhone notch, Android cutouts)
  - Responsive layout works on all screen sizes (320px → 3840px)
- ✅ **Mobile web compatibility** (Phase 0):
  - Touch targets ≥44px (iOS) / 48dp (Android)
  - Orientation changes handled gracefully
  - iOS Safari standalone mode (PWA) functional
  - Android Chrome PWA install prompt works
  - Lighthouse mobile audit: Performance >70, Accessibility >90
- ✅ All UI menus have functional exit paths (ESC + close button + back button/swipe on mobile)
- ✅ Core gameplay loop playable: launch → tutorial → combat → loot → progression → save/load
- ✅ Performance targets met (30+ FPS desktop, 20+ FPS mobile browsers)
- ✅ Storage persistence works (IndexedDB + localStorage fallback)
- ✅ Audio plays correctly after user gesture (click/keypress/tap)
- ✅ Fullscreen mode works without breaking rendering (desktop + mobile)
- ✅ GitHub Pages deployment successful (https://<username>.github.io/<repo>/)
- ✅ No console errors in browser DevTools (desktop + mobile browsers)

## Constraints

- **Do Not Break:** Deterministic generation (same seed = same world)
- **Do Not Modify:** Public API signatures without version bump justification
- **Do Not Skip:** Test validation for each fix on Chrome, Firefox, and Safari
- **Do Not Add:** New features or refactoring beyond bug fixes
- **Do Not Exceed:** 512MB WASM heap limit, 5MB localStorage limit
- **Do Not Block:** Main thread with synchronous operations (use async/goroutines)

## Execution Order

1. **Fix Phase 0 (HTML/CSS & Mobile Web Audit) issues first**:
   - Validate HTML5 markup (index.html, game.html)
   - Fix viewport meta tag if missing or incorrect
   - Apply touch-action CSS to canvas
   - Implement safe area insets (env(safe-area-inset-*))
   - Fix touch target sizes (<44px buttons)
   - Test responsive layout (320px mobile → 3840px desktop)
   - Run Lighthouse mobile audit, fix critical issues
2. Fix Phase 1 (Launch & Main Menu) blockers on all browsers (desktop + mobile)
3. Fix Phase 2 (Tutorial) blockers
4. Fix Phase 3 (Core Gameplay) blockers, prioritize input handling (keyboard + touch)
5. Fix Phase 4 (Progression) issues
6. Fix Phase 5 (World Interaction) issues
7. Fix Phase 6 (Multiplayer) issues (verify WebSocket WSS for production)
8. Fix Phase 7 (Save/Load) issues, prioritize IndexedDB implementation
9. Fix Browser-Specific Tests (audio context, fullscreen, storage, performance)
10. **Fix Mobile Web Browser Tests** (iOS Safari, Android Chrome, tablets, edge cases)
11. Validate all fixes with test suite and browser DevTools on Chrome, Firefox, Safari (desktop + mobile)
12. Deploy to GitHub Pages and verify production build on real mobile devices
13. Report summary of bugs found and fixed by phase, browser, and platform (desktop/mobile)

## Cross-References

- **WebAssembly Deployment**: See `docs/GITHUB_PAGES.md` for deployment process
- **Build Process**: See `Makefile` target `build-wasm` and `serve-wasm`
- **Web Files**: See `build/wasm/` directory for HTML and WASM artifacts
- **Storage System**: See `pkg/saveload/` for IndexedDB/localStorage implementation
- **Mobile Touch Controls**: See `pkg/mobile/` for touch input on mobile browsers
- **Performance Optimization**: See `docs/PERFORMANCE.md` for WASM-specific optimization
- **Architecture**: See `docs/ARCHITECTURE.md` for ECS system design
