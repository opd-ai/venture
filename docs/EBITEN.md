TASK: Conduct a systematic UI audit of Venture, a procedural multiplayer action-RPG built with Go/Ebiten, document all discovered issues with actionable fixes, and output findings to AUDIT.md in the project root.

CONTEXT:
- **Game**: Venture - Fully procedural multiplayer action-RPG
- **Technology Stack**: Go 1.24+ with Ebiten 2.9.2 game engine
- **Architecture**: Entity-Component-System (ECS) with deterministic procedural generation
- **Audit Scope**: All UI systems including game HUD, menus, inventory, character progression, crafting, quests, and multiplayer interfaces
- **Genre Support**: Fantasy, sci-fi, horror, cyberpunk, post-apocalyptic themes with cross-genre blending
- **Platform Testing**: Desktop (Linux/macOS/Windows), WebAssembly (browser), and mobile (iOS/Android touch interface)

REQUIREMENTS:

**Phase 1: Interface Discovery**
1. Map all UI components in `pkg/rendering/ui/`:
    - Main menu, pause menu, settings, genre selection
    - Game HUD (health, mana, minimap, hotbar, chat)
    - Character systems (inventory, skills, progression, stats)
    - World interaction (crafting stations, merchant trading, quest log)
    - Multiplayer interfaces (lobby, player list, connection status)
    - Mobile-specific touch controls (`pkg/mobile/`)

2. Document navigation flow using dual-exit pattern (ESC + dedicated back buttons)
3. Test procedural UI adaptation across all five genre themes
4. Verify deterministic UI state with same world seeds

**Phase 2: Systematic Testing**
1. Test each UI component for:
    - Visual clarity across all genre color palettes (`pkg/rendering/palette/`)
    - 60 FPS performance target with complex procedural content
    - Responsive feedback (visual/audio through `pkg/audio/synthesis/`)
    - Cross-platform input handling (keyboard, gamepad, touch)
    - Network latency tolerance (200-5000ms multiplayer scenarios)
    - Save/load state preservation (`pkg/saveload/`)

2. Evaluate procedural generation integration:
    - Dynamic item tooltips with generated stats and descriptions
    - Procedural quest text rendering and formatting
    - Generated skill tree visualization and navigation
    - Real-time entity health bars and status effects
    - Merchant inventory with procedural items and pricing

3. Test ECS integration:
    - UI component updates through entity systems
    - Performance with 2000+ entities (target: 106 FPS achieved)
    - Component state synchronization in multiplayer
    - Memory usage (target: <500MB, achieved: 73MB)

**Phase 3: Venture-Specific Validation**
1. **Deterministic Testing**: Verify UI displays identical content with same seed across clients
2. **Genre Consistency**: Test UI theming across fantasy/sci-fi/horror/cyberpunk/post-apocalyptic
3. **Multiplayer Sync**: Validate UI state synchronization with client-side prediction
4. **Performance Benchmarks**: Ensure UI doesn't impact core performance targets
5. **Mobile Adaptation**: Test touch interface scaling and responsiveness
6. **Accessibility**: Verify keyboard navigation and screen reader compatibility

**Phase 4: Issue Classification**
Categorize by impact on Venture's core experience:
- **Critical**: Prevents core gameplay loop (exploration, combat, progression, multiplayer)
- **High**: Significantly degrades procedural content experience or genre immersion
- **Medium**: Affects usability but doesn't break deterministic generation
- **Low**: Polish opportunities that enhance but don't impair functionality

OUTPUT FORMAT:

Create `/AUDIT.md` in the project root with this structure:

```markdown
# Venture UI Audit Report
**Game**: Venture v[Phase 9] - Procedural Multiplayer Action-RPG
**Audit Date**: [ISO 8601 format]
**Auditor**: GitHub Copilot
**Technology**: Go 1.24+ / Ebiten 2.9.2 / ECS Architecture
**Total Issues Found**: [Number]

## Executive Summary
[2-3 sentence overview focusing on procedural generation UI integration and multiplayer synchronization]

## Issues by Severity

### Critical Issues
#### Issue #[N]: [Descriptive Title]
- **Component**: [Specific UI system in pkg/rendering/ui/]
- **Genre Impact**: [Which genres affected: fantasy/sci-fi/horror/cyberpunk/post-apocalyptic]
- **Platform**: [Desktop/Web/Mobile or All]
- **Description**: [Problem explanation with ECS/procedural context]
- **Steps to Reproduce**:
  1. [Include seed value for deterministic testing]
  2. [Specify genre and difficulty parameters]
  3. [Detail multiplayer context if applicable]
- **Expected Behavior**: [What should happen per Venture's design patterns]
- **Actual Behavior**: [Current problematic behavior]
- **Suggested Fix**: [Actionable solution with Go/Ebiten code examples]
- **ECS Integration**: [How fix affects entity/component/system architecture]
- **Performance Impact**: [Frame rate or memory considerations]
- **Testing Verification**: [How to verify fix maintains determinism]

### High Priority Issues
[Same structure, focusing on genre immersion and procedural content display]

### Medium Priority Issues  
[Same structure, emphasizing usability and polish]

### Low Priority Issues
[Same structure, covering enhancement opportunities]

## Procedural Generation UI Integration
- **Genre Theming**: [How well UI adapts to different genres]
- **Dynamic Content**: [Procedural item/quest/skill display quality]
- **Performance**: [UI impact on generation and rendering performance]
- **Determinism**: [UI state consistency across clients with same seed]

## Multiplayer UI Assessment
- **State Synchronization**: [UI updates with network prediction]
- **Lag Tolerance**: [Interface responsiveness at 200-5000ms latency]
- **Player Communication**: [Chat, status indicators, lobby functionality]
- **Connection Feedback**: [Network status and error reporting]

## Cross-Platform Compatibility
- **Desktop**: [Windows/macOS/Linux keyboard and mouse experience]
- **WebAssembly**: [Browser deployment UI considerations]
- **Mobile**: [Touch interface adaptation and responsiveness]

## Positive Observations
[Highlight excellent UI/procedural integration examples and ECS patterns]

## Recommendations Summary
1. [Prioritized fixes with estimated development effort]
2. [Architecture improvements for better procedural integration]
3. [Performance optimizations for UI rendering]
4. [Multiplayer UI enhancements]

## Technical Implementation Notes
- **Ebiten Version**: 2.9.2
- **Go Version**: 1.24+
- **Resolution Tested**: [Multiple resolutions including mobile viewports]
- **Seed Used**: [For deterministic reproduction]
- **Network Conditions**: [Latency scenarios tested]
- **Entity Count**: [UI performance with various entity loads]
```

**Quality Standards Specific to Venture**:
- ✓ Every issue tested with deterministic seeds for reproducibility
- ✓ Fixes consider ECS architecture and component design patterns
- ✓ Performance impact assessed against 60 FPS / <500MB targets
- ✓ Multiplayer implications evaluated for all UI changes
- ✓ Genre theming compatibility verified across all five themes
- ✓ Mobile touch interface considerations included where applicable
- ✓ Code examples follow Venture's established patterns in pkg/rendering/ui/

**Testing Methodology for Venture**:
1. **Deterministic Exploration**: Use fixed seeds (e.g., 12345) for consistent UI state testing
2. **Genre Rotation**: Test each UI component across all five supported genres
3. **Multiplayer Simulation**: Test UI with local server and multiple clients
4. **Performance Monitoring**: Use `go test -bench` patterns to verify UI performance
5. **Cross-Platform Validation**: Test desktop, WebAssembly, and mobile builds
6. **ECS Integration**: Verify UI components properly interact with entity systems
7. **Save/Load Consistency**: Test UI state preservation through save/load cycles

**Venture-Specific Considerations**:
- Procedural content must render consistently with same generation parameters
- UI theming should seamlessly adapt to genre blending scenarios
- Network UI must handle Venture's high-latency tolerance (onion services)
- Touch controls must maintain action-RPG responsiveness on mobile
- Entity UI (health bars, status effects) must perform with 2000+ entities
- Crafting and merchant UIs must handle infinite procedural item variety
- Quest log must accommodate procedural quest text and objectives
- Skill tree UI must dynamically render procedural skill graphs
- Chat system must integrate with Venture's multiplayer architecture

**Example Venture-Specific Issue**:

```markdown
#### Issue #1: Procedural Item Tooltip Overflow in Cyberpunk Genre
- **Component**: pkg/rendering/ui/inventory.go - item tooltip rendering
- **Genre Impact**: Cyberpunk (neon color palette causes text contrast issues)
- **Platform**: All platforms
- **Description**: When displaying procedural items with long generated names (e.g., "Quantum-Enhanced Tactical Neural Interface Mk VII"), tooltips extend beyond screen bounds in cyberpunk genre due to bright neon backgrounds reducing text contrast.
- **Steps to Reproduce**:
  1. Start with seed 12345, cyberpunk genre, difficulty 0.8
  2. Generate items until tooltip with 40+ character name appears
  3. Hover over item in inventory screen
- **Expected Behavior**: Tooltip should wrap text and maintain readability across all genres
- **Actual Behavior**: Text overflows, neon background makes text illegible
- **Suggested Fix**: Implement dynamic text wrapping in renderTooltip() function:
  ```go
  func (ui *InventoryUI) renderTooltip(item *procgen.Item, palette *palette.GenrePalette) {
        maxWidth := ui.screenWidth * 0.3
        wrappedText := ui.wrapText(item.GetDisplayName(), maxWidth)
        // Use genre-appropriate contrast for background
        bgColor := palette.GetContrastBackground()
        ui.renderTextWithBackground(wrappedText, bgColor)
  }
  ```
- **ECS Integration**: Tooltip rendering system should query ItemComponent and use cached display names
- **Performance Impact**: Text wrapping adds ~0.1ms per tooltip, negligible for 60 FPS target
- **Testing Verification**: Test with all genre palettes using deterministic item generation
```