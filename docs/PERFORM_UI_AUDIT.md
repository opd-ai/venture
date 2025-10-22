TASK: Conduct a systematic UI audit of Venture, a procedurally generated action-RPG built with Go/Ebiten, document all discovered issues with actionable fixes, and output findings to UI_AUDIT.md.

CONTEXT:
- **Game**: Venture - Procedural Action RPG
- **Technology Stack**: Go 1.24.7, Ebiten 2.9.2 game engine
- **Architecture**: ECS (Entity-Component-System) with procedural generation
- **UI System**: `pkg/rendering/ui` - Procedural UI element generation with genre-aware styling
- **Audit Scope**: UI elements, interactions, visual feedback, usability, and procedural generation consistency
- **Mental Model Approach**: Build progressive understanding of UI systems through systematic exploration and codebase analysis

## Venture-Specific Context

**Key Systems to Audit:**
- `pkg/rendering/ui/generator.go` - UI element generation (buttons, panels, health bars, labels, icons, frames)
- `pkg/rendering/palette/generator.go` - Genre-specific color palette generation
- `pkg/rendering/sprites/generator.go` - Procedural sprite generation
- `pkg/rendering/tiles/generator.go` - Tile-based rendering
- `pkg/engine/game.go` - Ebiten game loop integration (Update, Draw, Layout methods)
- `cmd/client/main.go` - Client application with UI initialization

**Testing Tools Available:**
- `cmd/rendertest/` - CLI tool for testing palette generation
- `cmd/client/` - Full game client with all systems integrated
- `go test -tags test ./pkg/rendering/ui/...` - Unit tests for UI generation

REQUIREMENTS:

## Pre-Audit Setup and Environment Verification

**Before starting the audit, verify the testing environment:**

1. **Build and run the client application:**
   ```bash
   cd /path/to/venture
   go build -o venture-client ./cmd/client
   ./venture-client -width 1280 -height 720 -genre fantasy -seed 12345
   ```

2. **Test UI element generation directly:**
   ```bash
   # Test color palette generation
   cd cmd/rendertest
   go run main.go -genre fantasy -seed 12345 -verbose
   
   # Run UI generator tests
   go test -tags test -v ./pkg/rendering/ui/...
   ```

3. **Check for compilation and dependency issues:**
   ```bash
   go build ./...
   go test -tags test ./pkg/...
   ```

4. **Verify Ebiten dependencies (Linux):**
   ```bash
   # Required X11 libraries
   dpkg -l | grep -E "libgl1-mesa-dev|libxcursor-dev|libxi-dev"
   ```

**Phase 1: Interface Discovery**

1. **Identify all UI components** by examining code:
   - Review `pkg/rendering/ui/types.go` for ElementType enum (Button, Panel, HealthBar, Label, Icon, Frame)
   - Check `pkg/engine/inventory_components.go` for inventory UI data structures
   - Examine `pkg/engine/game.go` Draw() method for rendering pipeline
   - Look for HUD, menu, dialog implementations in engine components
   
2. **Map the navigation flow** between UI states:
   - Review game state management in `pkg/engine/ecs.go`
   - Check for pause/menu state handling in game loop
   - Document transitions between gameplay, inventory, and menu states
   
3. **Document all interactive elements** and their expected behaviors:
   - Input handling in `pkg/engine/game.go` Update() method
   - Mouse/keyboard interaction patterns
   - Controller support (if implemented)
   
4. **Note procedural generation impact** on UI consistency:
   - Test UI elements across different genres (fantasy, scifi, horror, cyberpunk, postapoc)
   - Verify deterministic generation (same seed = same visuals)
   - Check color palette consistency via `pkg/rendering/palette/generator.go`

**Phase 2: Systematic Testing**

1. **Test each UI component** for:
   - **Visual clarity and readability**: 
     - Test at target resolution (800x600, 1280x720, 1920x1080)
     - Verify text rendering via Ebiten's text drawing APIs
     - Check border rendering in different border styles (Solid, Double, Ornate, Glow)
   - **Responsive feedback to user input**:
     - Test element state transitions (Normal → Hover → Pressed → Disabled)
     - Verify state changes are visually distinct in `generateButton()` method
     - Check cooldown visual feedback for actions
   - **Consistency with established patterns**:
     - Compare UI across all 5 genres using `rendertest` tool
     - Verify genre-specific styling in `selectBorderStyle()` and color selection
     - Check that deterministic generation works (same seed = same appearance)
   - **Accessibility of controls**:
     - Test keyboard navigation
     - Verify controller/gamepad support (if implemented)
     - Check color contrast ratios for text readability
   - **Performance** (frame rate impact, lag, stuttering):
     - Monitor FPS during UI-heavy screens (inventory, skill trees)
     - Use Go profiling: `go test -tags test -cpuprofile=cpu.prof -bench=.`
     - Profile UI generation: `go test -tags test -memprofile=mem.prof ./pkg/rendering/ui/...`
     - Check for memory leaks with large UI hierarchies
   - **Edge cases** (empty states, maximum values, rapid inputs):
     - Test empty inventory
     - Test full inventory (20 items capacity from `cmd/client/main.go`)
     - Test health bars at 0%, 30%, 60%, 100% values
     - Rapid button clicking and state changes

2. **Evaluate UI/UX patterns**:
   - **Information hierarchy**: HUD element priority and positioning
   - **Visual alignment and spacing**: Grid layouts, pixel-perfect alignment
   - **Color contrast and visibility**:
     - Test against `pkg/rendering/palette/generator.go` output
     - Verify Danger (red), Success (green), Accent colors are distinguishable
     - Check background/foreground contrast ratios (WCAG AA: 4.5:1 minimum)
   - **Text legibility at game resolution**: 
     - Test at minimum resolution (800x600)
     - Check font sizes in procedurally generated labels
   - **Icon clarity and meaning**: Verify ElementIcon generation produces recognizable shapes
   - **Tutorial/help availability**: Document in-game help or tutorial systems

3. **Test integration points**:
   - **UI updates during gameplay**:
     - Health bar updates on damage/healing
     - Inventory updates on item pickup
     - Experience/level progression UI updates
   - **Transitions between game states**:
     - Gameplay ↔ Inventory screen
     - Gameplay ↔ Pause menu
     - Check for visual artifacts or timing issues
   - **Pause/resume functionality**:
     - Verify `game.Paused` flag in `pkg/engine/game.go`
     - Test pause menu rendering and input handling
   - **Save/load impact on UI state**:
     - Check if UI state persists correctly
     - Verify no stale data in UI after loading
   - **Procedurally generated content display**:
     - Test item tooltips with generated item names/stats
     - Monster health bars with procedural entity data
     - Skill tree rendering with generated skills

**Phase 3: Issue Classification**

Categorize findings by severity and link to specific code locations:

- **Critical**: Blocks core functionality or causes crashes
  - Example: Game crashes when opening inventory
  - Action: Include stack trace and steps to reproduce
  - Link to relevant code: `pkg/engine/inventory_system.go` line numbers
  
- **High**: Significantly impairs usability or immersion
  - Example: Health bar doesn't update when taking damage
  - Action: Reference combat system integration in `pkg/engine/combat_system.go`
  - Test: Run combat tests with `go test -tags test ./pkg/engine/combat_test.go`
  
- **Medium**: Noticeable but doesn't prevent gameplay
  - Example: Button hover state color too similar to normal state
  - Action: Check color generation in `generateButton()` method
  - Fix: Adjust `lightenColor()` factor from 0.2 to 0.4
  
- **Low**: Minor polish or enhancement opportunities
  - Example: Border style inconsistent with genre theme
  - Action: Review `selectBorderStyle()` genre mappings
  - Test: Use rendertest to verify across all genres

## Debugging Workflows for Common Issues

### Workflow 1: UI Element Not Rendering

**Symptoms**: Button/panel/health bar not visible on screen

**Investigation Steps:**
1. Check if element generation succeeds:
   ```bash
   go test -tags test -v -run TestGenerator_Generate ./pkg/rendering/ui/
   ```

2. Verify Draw() method is called:
   - Add debug logging to `pkg/engine/game.go` Draw() method
   - Check if rendering systems are registered with world

3. Test element generation in isolation:
   ```go
   // Create test in /tmp/test_ui.go
   gen := ui.NewGenerator()
   config := ui.Config{Type: ui.ElementButton, Width: 100, Height: 30, GenreID: "fantasy", Seed: 12345}
   img, err := gen.Generate(config)
   // Save img to file for visual inspection
   ```

4. Check Ebiten image rendering:
   - Verify image.RGBA to ebiten.Image conversion
   - Check if image dimensions are valid (>0)
   - Verify image is drawn at correct screen coordinates

### Workflow 2: Performance Issues (FPS Drops)

**Symptoms**: Game stutters or FPS drops below 60 when UI is visible

**Investigation Steps:**
1. Profile the application:
   ```bash
   go build -o venture-client ./cmd/client
   go tool pprof -http=:8080 venture-client cpu.prof
   ```

2. Run benchmarks on UI generation:
   ```bash
   go test -tags test -bench=BenchmarkGenerate -benchmem ./pkg/rendering/ui/
   ```

3. Check for common performance issues:
   - Multiple allocations in hot path (Draw called every frame)
   - Creating new images every frame instead of caching
   - Excessive color calculations in nested loops
   - Large UI element dimensions causing memory pressure

4. Solutions:
   - Cache generated UI elements when not changing
   - Use object pooling for frequently created UI
   - Implement dirty flag to only regenerate when needed
   - Profile with: `go test -tags test -cpuprofile=cpu.prof -memprofile=mem.prof`

### Workflow 3: Color/Visual Inconsistencies

**Symptoms**: Colors don't match genre theme or vary between sessions with same seed

**Investigation Steps:**
1. Verify deterministic color generation:
   ```bash
   # Generate palette twice with same seed
   cd cmd/rendertest
   go run main.go -genre fantasy -seed 12345 > output1.txt
   go run main.go -genre fantasy -seed 12345 > output2.txt
   diff output1.txt output2.txt  # Should be identical
   ```

2. Check RNG seeding in UI generation:
   - Verify `rand.New(rand.NewSource(config.Seed))` in Generate()
   - Ensure no usage of global `math/rand` functions
   - Check for `time.Now()` usage (breaks determinism)

3. Test palette across genres:
   ```bash
   for genre in fantasy scifi horror cyberpunk postapoc; do
     go run cmd/rendertest/main.go -genre $genre -seed 12345
   done
   ```

4. Review color transformation functions:
   - Check `lightenColor()` and `darkenColor()` implementations
   - Verify color values stay within valid ranges (0-255)
   - Test extreme cases (black, white, near-black, near-white)

### Workflow 4: State Management Issues

**Symptoms**: Button stays in hover/pressed state, UI shows stale data

**Investigation Steps:**
1. Check state transitions in code:
   - Review `ElementState` enum in `pkg/rendering/ui/types.go`
   - Verify state changes in input handling code
   - Check if state is reset properly on transitions

2. Add state logging:
   ```go
   // In button handling code
   log.Printf("Button state transition: %s -> %s", oldState, newState)
   ```

3. Test state machine:
   - Create unit test for all state transitions
   - Verify edge cases: rapid clicks, mouse out during press
   - Test with different input methods (mouse, keyboard, gamepad)

4. Common fixes:
   - Reset state to Normal when mouse leaves element
   - Clear pressed state after action completes
   - Implement state timeout for stuck states

### Workflow 5: Layout and Positioning Issues

**Symptoms**: UI elements overlap, misaligned, or off-screen

**Investigation Steps:**
1. Check Layout() method in `pkg/engine/game.go`:
   - Verify returned screen dimensions match window size
   - Test with different resolutions via `-width` and `-height` flags

2. Review positioning calculations:
   - Check for integer overflow/underflow
   - Verify screen bounds checking
   - Test at minimum resolution (800x600) and maximum (1920x1080)

3. Add visual debug overlay:
   ```go
   // Draw bounding boxes for all UI elements
   ebitenutil.DrawRect(screen, x, y, width, height, color.RGBA{255, 0, 0, 128})
   ```

4. Test with different window resize modes:
   - Check Ebiten WindowResizingMode behavior
   - Verify UI scales/repositions correctly
   - Test fullscreen transitions

OUTPUT FORMAT:

Create `UI_AUDIT.md` with this exact structure:

```markdown
# UI Audit Report
**Game**: [Game Name/Version]
**Audit Date**: [ISO 8601 format]
**Auditor**: BotBot AI
**Total Issues Found**: [Number]

## Executive Summary
[2-3 sentence overview of audit scope and key findings]

## Issues by Severity

### Critical Issues
#### Issue #[N]: [Descriptive Title]
- **Component**: [Specific UI element]
- **Description**: [Clear explanation of the problem]
- **Steps to Reproduce**:
  1. [Step 1]
  2. [Step 2]
- **Expected Behavior**: [What should happen]
- **Actual Behavior**: [What currently happens]
- **Suggested Fix**: [Actionable solution with technical details]
- **Ebiten-Specific Considerations**: [Relevant engine constraints or APIs]

[Repeat for each critical issue]

### High Priority Issues
[Same structure as above]

### Medium Priority Issues
[Same structure as above]

### Low Priority Issues
[Same structure as above]

## Positive Observations
[List UI elements that work well and demonstrate good design]

## Recommendations Summary
1. [Prioritized fix recommendation 1]
2. [Prioritized fix recommendation 2]
[...]

## Technical Notes
- **Ebiten Version**: [If determinable]
- **Resolution Tested**: [Screen resolution]
- **Testing Environment**: [Browser/OS if applicable]
```

## Quick Reference: UI Debugging Checklist

Use this checklist to systematically investigate any UI issue:

### Initial Triage
- [ ] Can you reproduce the issue consistently?
- [ ] Does it occur with multiple seeds or just one?
- [ ] Does it affect all genres or just specific ones?
- [ ] Is it present in tests or only in the running game?
- [ ] Can you isolate it to a single UI element type?

### Code Review Checklist
- [ ] Check `pkg/rendering/ui/generator.go` for generation logic
- [ ] Review `pkg/rendering/palette/generator.go` for color issues  
- [ ] Examine `pkg/engine/game.go` for Update/Draw/Layout methods
- [ ] Look at relevant component definitions in `pkg/engine/components.go`
- [ ] Check system implementations for UI state management

### Testing Checklist
- [ ] Run unit tests: `go test -tags test ./pkg/rendering/ui/...`
- [ ] Run with race detector: `go test -tags test -race ./...`
- [ ] Test with multiple seeds: 12345, 67890, 11111, 99999
- [ ] Test all genres: fantasy, scifi, horror, cyberpunk, postapoc
- [ ] Test at multiple resolutions: 800x600, 1280x720, 1920x1080
- [ ] Profile CPU: `go test -tags test -cpuprofile=cpu.prof`
- [ ] Profile memory: `go test -tags test -memprofile=mem.prof`
- [ ] Run benchmarks: `go test -tags test -bench=. -benchmem`

### Determinism Verification
- [ ] Generate UI element twice with same parameters
- [ ] Compare outputs byte-by-byte or visually
- [ ] Check for `time.Now()` usage in call stack
- [ ] Verify all RNG uses seeded `rand.New()` not global `rand`
- [ ] Confirm no dependency on map iteration order
- [ ] Test across different OS/architectures if possible

### Performance Checklist
- [ ] Monitor FPS during issue occurrence
- [ ] Check frame time budget (target: <16ms per frame for 60 FPS)
- [ ] Count draw calls (use Ebiten debug mode)
- [ ] Measure UI generation time (should be <1ms per element)
- [ ] Check memory allocations per frame (aim for zero in hot path)
- [ ] Profile with pprof to identify bottlenecks

### Integration Checklist
- [ ] Verify component data is up-to-date
- [ ] Check system execution order
- [ ] Confirm UI reads latest entity state
- [ ] Test state transitions (gameplay ↔ menus)
- [ ] Verify input handling doesn't conflict
- [ ] Check for race conditions with `-race` flag

### Documentation Checklist
- [ ] Take screenshots of the issue
- [ ] Record exact reproduction steps
- [ ] Note environment details (OS, Go version, resolution)
- [ ] Document expected vs actual behavior
- [ ] List affected files and line numbers
- [ ] Suggest specific fix with code snippets
- [ ] Estimate severity and priority
- [ ] Link to related issues or code sections

## Common Commands Reference

### Building and Running
```bash
# Build client
go build -o venture-client ./cmd/client

# Run with specific settings
./venture-client -width 1280 -height 720 -genre fantasy -seed 12345 -verbose

# Build with optimizations
go build -ldflags="-s -w" -o venture-client ./cmd/client
```

### Testing
```bash
# Run all tests
go test -tags test ./...

# Test specific package
go test -tags test -v ./pkg/rendering/ui/

# Test with coverage
go test -tags test -cover -coverprofile=coverage.out ./pkg/rendering/...
go tool cover -html=coverage.out

# Test with race detection
go test -tags test -race ./...

# Run benchmarks
go test -tags test -bench=. -benchmem ./pkg/rendering/ui/
```

### Profiling
```bash
# CPU profiling
go test -tags test -cpuprofile=cpu.prof -bench=. ./pkg/rendering/ui/
go tool pprof -http=:8080 cpu.prof

# Memory profiling
go test -tags test -memprofile=mem.prof -bench=. ./pkg/rendering/ui/
go tool pprof -http=:8080 mem.prof

# Profile running application
go build -o venture-client ./cmd/client
./venture-client &
kill -USR1 $!  # If signal handling implemented
```

### Debugging with Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug client
dlv debug ./cmd/client -- -genre fantasy -seed 12345

# Common delve commands
(dlv) break pkg/rendering/ui/generator.go:25
(dlv) continue
(dlv) print config
(dlv) locals
(dlv) stack
(dlv) next
(dlv) step
```

### Visual Testing
```bash
# Test palette generation
cd cmd/rendertest
go run main.go -genre fantasy -seed 12345 -verbose

# Test all genres
for genre in fantasy scifi horror cyberpunk postapoc; do
  echo "Testing $genre..."
  go run main.go -genre $genre -seed 12345
done

# Compare outputs
go run main.go -genre fantasy -seed 12345 > /tmp/test1.txt
go run main.go -genre fantasy -seed 12345 > /tmp/test2.txt  
diff /tmp/test1.txt /tmp/test2.txt
```

## Additional Resources

### Code Documentation
- [Architecture Overview](./ARCHITECTURE.md)
- [Development Guide](./DEVELOPMENT.md)
- [Technical Specification](./TECHNICAL_SPEC.md)
- [Implemented Phases](./IMPLEMENTED_PHASES.md)

### Package Documentation
```bash
# View package documentation
go doc github.com/opd-ai/venture/pkg/rendering/ui
go doc github.com/opd-ai/venture/pkg/engine

# Generate HTML documentation
godoc -http=:6060
# Visit http://localhost:6060/pkg/github.com/opd-ai/venture/
```

### Key Files for UI Debugging
- `pkg/rendering/ui/generator.go` - Main UI generation logic
- `pkg/rendering/ui/types.go` - UI element types and configuration
- `pkg/rendering/palette/generator.go` - Color palette generation
- `pkg/engine/game.go` - Ebiten integration (Update/Draw/Layout)
- `pkg/engine/components.go` - Core component definitions
- `cmd/client/main.go` - Client application entry point
- `cmd/rendertest/main.go` - CLI tool for testing rendering

### Ebiten Resources
- [Ebiten Documentation](https://ebitengine.org/en/documents/)
- [Ebiten Examples](https://ebitengine.org/en/examples/)
- [Ebiten GitHub](https://github.com/hajimehoshi/ebiten)

### Go Testing Resources
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Profiling](https://go.dev/blog/pprof)
- [Delve Debugger](https://github.com/go-delve/delve)

**If No Issues Found**:
```markdown
# UI Audit Report
**Game**: [Game Name/Version]
**Audit Date**: [ISO 8601 format]
**Auditor**: BotBot AI
**Total Issues Found**: 0

## 🎉 Excellent News!

After systematic exploration and testing of all discoverable UI components, **no issues were identified**. The interface demonstrates:

- ✓ Consistent visual design
- ✓ Responsive user interactions
- ✓ Clear information hierarchy
- ✓ Smooth state transitions
- ✓ Appropriate feedback mechanisms
- ✓ Stable performance

## Tested Components
[List all UI elements examined]

## Commendations
[Specific examples of well-executed UI/UX patterns]

Keep up the outstanding work! 🏆
```

QUALITY CRITERIA:
- ✓ Every issue includes reproducible steps
- ✓ Suggested fixes are technically feasible for Go/Ebiten
- ✓ Severity ratings are justified and consistent
- ✓ Report is actionable for developers
- ✓ Language is professional but encouraging
- ✓ All UI components mentioned are actually discoverable through interaction
- ✓ Mental model demonstrates logical exploration progression

TESTING METHODOLOGY:

## Systematic Testing Approach

1. **Exploration Phase**: Interact with all visible UI elements systematically
   - Start client with known seed: `./venture-client -seed 12345 -genre fantasy`
   - Test each interactive element (if implemented): buttons, menus, inventory
   - Document actual UI elements found vs. expected from code
   - Take screenshots for documentation

2. **State Mapping**: Document UI state machine and transitions
   - Create state diagram: Gameplay → Inventory → Pause Menu → etc.
   - Test all transition paths
   - Verify state persistence across transitions
   - Check for orphaned/unreachable states

3. **Stress Testing**: Test boundary conditions and rapid inputs
   - Test maximum capacity: 20 items in inventory (from `InventoryComponent.Capacity`)
   - Test rapid inputs: Click buttons rapidly, spam keyboard inputs
   - Test extreme values: Health at 0, 1, max value
   - Test long sessions: Run game for 30+ minutes, check for memory leaks

4. **Consistency Check**: Verify patterns hold across procedurally generated content
   - Test with multiple seeds: 12345, 67890, 11111, etc.
   - Test all genres: fantasy, scifi, horror, cyberpunk, postapoc
   - Use rendertest tool to compare palettes:
     ```bash
     go run cmd/rendertest/main.go -genre fantasy -seed 12345 -verbose
     go run cmd/rendertest/main.go -genre scifi -seed 12345 -verbose
     ```
   - Verify visual consistency within each genre
   - Ensure determinism: same seed + genre = same visuals

5. **Accessibility Review**: Evaluate usability for different player skill levels
   - Test with keyboard-only input (if implemented)
   - Test with gamepad/controller (if implemented)
   - Check text readability at minimum resolution
   - Verify color-blind friendly palettes (check contrast)
   - Test with different window sizes via Ebiten resizing

6. **Performance Monitoring**: Note any UI-related performance degradation
   - Monitor FPS with Ebiten debug info (if enabled)
   - Profile CPU usage: `go tool pprof cpu.prof`
   - Profile memory usage: `go tool pprof mem.prof`
   - Check draw call count (Ebiten debug mode)
   - Test on lower-spec hardware if available

## Code-Level Debugging Techniques

### Using Go Debugger (Delve)
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug client application
dlv debug ./cmd/client -- -width 800 -height 600 -genre fantasy

# Set breakpoints in UI code
(dlv) break pkg/rendering/ui/generator.go:25
(dlv) break pkg/engine/game.go:54

# Continue execution and inspect
(dlv) continue
(dlv) print config
(dlv) print img.Bounds()
```

### Adding Debug Logging
```go
// In pkg/rendering/ui/generator.go
import "log"

func (g *Generator) Generate(config Config) (*image.RGBA, error) {
    log.Printf("[UI DEBUG] Generating %s element: %dx%d, genre=%s, seed=%d",
        config.Type, config.Width, config.Height, config.GenreID, config.Seed)
    
    // ... existing code ...
    
    log.Printf("[UI DEBUG] Element generated successfully")
    return img, nil
}
```

### Using Ebiten Debug Mode
```go
// In pkg/engine/game.go Draw() method
import "github.com/hajimehoshi/ebiten/v2/ebitenutil"

func (g *Game) Draw(screen *ebiten.Image) {
    // Existing rendering code...
    
    // Add debug overlay
    ebitenutil.DebugPrint(screen, fmt.Sprintf(
        "FPS: %.2f\nEntities: %d\nSystems: %d",
        ebiten.ActualFPS(),
        len(g.World.entities),
        len(g.World.systems),
    ))
}
```

### Testing UI Generation Directly
Create test file in `/tmp/test_ui_generation.go`:
```go
package main

import (
    "image/png"
    "log"
    "os"
    
    "github.com/opd-ai/venture/pkg/rendering/ui"
)

func main() {
    gen := ui.NewGenerator()
    
    // Test button generation
    config := ui.Config{
        Type:    ui.ElementButton,
        Width:   200,
        Height:  50,
        GenreID: "fantasy",
        Seed:    12345,
        Text:    "Test Button",
        State:   ui.StateNormal,
    }
    
    img, err := gen.Generate(config)
    if err != nil {
        log.Fatalf("Generation failed: %v", err)
    }
    
    // Save to file
    f, _ := os.Create("/tmp/test_button.png")
    defer f.Close()
    png.Encode(f, img)
    
    log.Printf("Button generated: %v", img.Bounds())
}
```

Run with: `go run /tmp/test_ui_generation.go`

EBITEN-SPECIFIC CONSIDERATIONS:

## Ebiten Engine Constraints and Best Practices

### Image Rendering Performance
- **Issue**: Large UI elements (>1024x1024) may cause performance degradation
- **Check**: Review UI element dimensions in code, ensure reasonable sizes
- **Test**: Benchmark with `go test -tags test -bench=. ./pkg/rendering/ui/`
- **Solution**: Implement caching for static UI elements, use texture atlases
- **Code Reference**: `pkg/rendering/ui/generator.go` - image.NewRGBA() calls

### Input Handling Responsiveness
- **Issue**: Input lag on slower systems or with high poll rates
- **Check**: Review Update() method in `pkg/engine/game.go`
- **Test**: Test with rapid keyboard/mouse inputs
- **Solution**: 
  - Use `ebiten.IsKeyPressed()` for continuous input
  - Use state changes for discrete actions
  - Implement input buffering if needed
- **Code Reference**: Client input handling (not yet fully implemented in main.go)

### Draw Call Efficiency
- **Issue**: Too many draw calls per frame (>1000) can reduce FPS
- **Check**: Count DrawImage() calls in rendering code
- **Test**: Use Ebiten's TPS/FPS display
- **Solution**:
  - Batch similar UI elements into single texture
  - Use image.SubImage() for sprite sheets
  - Cache generated UI elements
- **Optimization**: Implement spatial partitioning for only drawing visible UI

### Text Rendering Clarity
- **Issue**: Text may appear blurry at certain scales or resolutions
- **Check**: Test at multiple resolutions (800x600, 1280x720, 1920x1080)
- **Test**: Generate labels with rendertest tool
- **Solution**:
  - Use integer pixel coordinates (avoid float positions)
  - Consider using bitmap fonts for pixel-perfect rendering
  - Test font sizes: 12pt, 14pt, 16pt for readability
- **Code Reference**: `pkg/rendering/ui/generator.go` - generateLabel() method

### Audio Feedback Integration
- **Issue**: UI sound effects not synchronized with visual feedback
- **Check**: Review audio system integration (pkg/audio/)
- **Test**: Test button clicks, menu navigation sounds
- **Solution**:
  - Play sounds immediately on input event
  - Use Ebiten's audio package for low-latency playback
  - Cache audio samples for frequently used sounds
- **Code Reference**: Audio synthesis system (Phase 4 implementation)

### Memory Management
- **Issue**: Memory leaks from unreleased Ebiten images
- **Check**: Profile with `go test -tags test -memprofile=mem.prof`
- **Test**: Run game for extended period, monitor memory usage
- **Solution**:
  - Call img.Dispose() on unused images
  - Use object pools for temporary images
  - Implement resource manager for UI textures
- **Best Practice**: Follow Ebiten's image lifecycle guidelines

### Screen Scaling and Resolution
- **Issue**: UI elements don't scale properly with window resizing
- **Check**: Test Layout() method with various window sizes
- **Test**: Enable window resizing, drag to different sizes
- **Solution**:
  - Implement UI scaling factor based on Layout() return values
  - Use relative positioning (percentages) rather than fixed pixels
  - Regenerate UI at different sizes for key breakpoints
- **Code Reference**: `pkg/engine/game.go` - Layout() method (line 60)

### Goroutine Safety
- **Issue**: Race conditions when updating UI from multiple goroutines
- **Check**: Run with `go test -tags test -race ./...`
- **Test**: Use race detector during development
- **Solution**:
  - Only update UI from main game loop (Update/Draw)
  - Use channels to communicate UI updates from other systems
  - Lock shared UI state with sync.Mutex if necessary
- **Warning**: Ebiten's Update/Draw must only be called from main goroutine

## Venture-Specific UI Architecture

### Component-Based UI System
- **Current State**: UI generation is procedural but not yet integrated with ECS
- **Expected Integration**: 
  - Create UIComponent for entity-based UI elements (health bars above enemies)
  - Implement UISystem to manage UI lifecycle and rendering
  - Add UIElement entities for interactive interface elements
- **Code Locations**:
  - Components: `pkg/engine/components.go`
  - Systems: `pkg/engine/` (movement, collision, combat, etc.)
  - UI Generation: `pkg/rendering/ui/generator.go`

### Genre-Aware Styling
- **Implementation**: Uses `GenreID` to select color palettes and border styles
- **Test Command**: 
  ```bash
  for genre in fantasy scifi horror cyberpunk postapoc; do
    ./venture-client -genre $genre -seed 12345
  done
  ```
- **Validation**: Each genre should have distinct visual characteristics:
  - Fantasy: Ornate borders, warm earthy colors
  - Sci-Fi: Clean lines, neon accents, cool colors
  - Horror: Dark tones, rough textures, high contrast
  - Cyberpunk: Glowing edges, pinks/purples, digital aesthetics
  - Post-Apocalyptic: Worn textures, muted colors, rust tones
- **Code Reference**: `pkg/rendering/palette/generator.go`, `selectBorderStyle()` in UI generator

### Deterministic Visual Generation
- **Requirement**: Same seed must produce identical visuals for multiplayer sync
- **Test Procedure**:
  ```bash
  # Generate UI element twice
  go run /tmp/test_ui.go -seed 12345 > /tmp/out1.txt
  go run /tmp/test_ui.go -seed 12345 > /tmp/out2.txt
  diff /tmp/out1.txt /tmp/out2.txt  # Should be empty (no differences)
  ```
- **Common Pitfalls**:
  - Using `time.Now()` anywhere in generation pipeline
  - Using global `math/rand` instead of seeded `rand.New()`
  - Depending on map iteration order (non-deterministic in Go)
  - Using floating-point operations that vary across architectures
- **Validation**: Run generation tests multiple times, compare output

EXAMPLE OUTPUT SNIPPET:

```markdown
### High Priority Issues

#### Issue #3: Button State Transitions Not Deterministic with Different Seeds
- **Component**: UI Button Generator (`pkg/rendering/ui/generator.go:64-100`)
- **Description**: Button visual appearance varies unpredictably between seeds due to random color selection from palette. While the palette itself is deterministic, the random selection of `colorIndex` in `generateButton()` uses the config seed, but the color chosen may not provide sufficient contrast for different states (normal, hover, pressed).
- **Steps to Reproduce**:
  1. Generate button with seed 12345: `go run /tmp/test_ui.go -seed 12345 -state normal`
  2. Generate same button with seed 67890: `go run /tmp/test_ui.go -seed 67890 -state normal`
  3. Observe that some seeds produce low-contrast buttons where hover state is barely distinguishable
- **Expected Behavior**: All generated buttons should have clearly distinguishable states regardless of seed, with sufficient contrast ratios (WCAG AA: 3:1 minimum for UI components)
- **Actual Behavior**: Some seed values produce buttons where normal and hover states have insufficient contrast, making interaction feedback unclear
- **Root Cause**: Line 69-70 in generator.go:
  ```go
  colorIndex := rng.Intn(len(pal.Colors))
  baseColor := pal.Colors[colorIndex]
  ```
  This selects a random color without checking if it provides adequate contrast for state variations.
- **Suggested Fix**: Implement color selection with contrast validation:
  ```go
  // pkg/rendering/ui/generator.go, line 69
  func (g *Generator) selectButtonColor(pal *palette.Palette, rng *rand.Rand) color.Color {
      // Try up to 10 colors to find one with good contrast potential
      for i := 0; i < 10; i++ {
          colorIndex := rng.Intn(len(pal.Colors))
          baseColor := pal.Colors[colorIndex]
          
          // Calculate luminance to ensure we can darken/lighten effectively
          r, gr, b, _ := baseColor.RGBA()
          luminance := (0.299*float64(r) + 0.587*float64(gr) + 0.114*float64(b)) / 65535.0
          
          // Good range for color manipulation: not too dark, not too light
          if luminance > 0.2 && luminance < 0.8 {
              return baseColor
          }
      }
      // Fallback to primary color which should be well-tested
      return pal.Primary
  }
  ```
- **Testing**: 
  - Test with seeds: 12345, 67890, 11111, 99999, 54321
  - Verify contrast ratios with tool or manual calculation
  - Run benchmark to ensure performance impact is minimal (<1ms)
- **Ebiten-Specific Considerations**: 
  - Color transformations (`lightenColor`, `darkenColor`) must maintain RGB validity (0-255)
  - Cache selected button colors to avoid recalculation every frame
  - Consider pre-generating button states on initialization rather than per-frame

#### Issue #7: Health Bar Update Lag in Combat
- **Component**: Health Bar UI Element and Combat System Integration
- **Description**: Health bars don't update immediately when entities take damage. There's a visible 1-2 frame delay between damage application and visual feedback.
- **Steps to Reproduce**:
  1. Build and run client: `go build ./cmd/client && ./venture-client`
  2. Engage in combat with an enemy
  3. Observe health bar during damage events
  4. Notice delay between damage number and bar update
- **Expected Behavior**: Health bar should update in the same frame as damage application
- **Actual Behavior**: 1-2 frame delay between HealthComponent update and UI refresh
- **Root Cause**: 
  - Combat system updates HealthComponent in Update() phase
  - UI rendering happens in Draw() phase  
  - No explicit UI update trigger when health changes
- **Suggested Fix**:
  1. Add dirty flag to HealthComponent:
     ```go
     // pkg/engine/components.go
     type HealthComponent struct {
         Current    float64
         Max        float64
         IsDirty    bool  // New field
     }
     ```
  2. Set dirty flag in combat system when health changes:
     ```go
     // pkg/engine/combat_system.go
     healthComp.Current -= damage
     healthComp.IsDirty = true
     ```
  3. Check dirty flag in rendering system and regenerate health bar if needed
  4. Clear dirty flag after rendering
- **Alternative Solution**: Implement event system for health changes
- **Testing**:
  - Run combat tests: `go test -tags test -v ./pkg/engine/combat_test.go`
  - Manual testing with frame-by-frame inspection
  - Verify no performance impact with many entities
- **Ebiten-Specific Considerations**: 
  - Ensure Draw() reads latest component state
  - Consider caching health bar images and only regenerating on change
  - Batch health bar updates for multiple entities
```
