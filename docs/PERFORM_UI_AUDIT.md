TASK: Conduct a systematic UI audit of a procedurally generated Zelda-style RPG built with Go/Ebiten, document all discovered issues with actionable fixes, and output findings to UI_AUDIT.md.

CONTEXT:
- **Game Type**: Top-down procedurally generated RPG in the style of classic Zelda games
- **Technology Stack**: Go programming language with Ebiten game engine
- **Audit Scope**: User interface elements, interactions, visual feedback, and usability
- **Mental Model Approach**: Build progressive understanding of UI systems through systematic exploration

REQUIREMENTS:

**Phase 1: Interface Discovery**
1. Identify all UI components (menus, HUD elements, dialogs, inventory screens, etc.)
2. Map the navigation flow between UI states
3. Document all interactive elements and their expected behaviors
4. Note procedural generation impact on UI consistency

**Phase 2: Systematic Testing**
1. Test each UI component for:
   - Visual clarity and readability
   - Responsive feedback to user input
   - Consistency with established patterns
   - Accessibility of controls
   - Performance (frame rate impact, lag, stuttering)
   - Edge cases (empty states, maximum values, rapid inputs)

2. Evaluate UI/UX patterns:
   - Information hierarchy
   - Visual alignment and spacing
   - Color contrast and visibility
   - Text legibility at game resolution
   - Icon clarity and meaning
   - Tutorial/help availability

3. Test integration points:
   - UI updates during gameplay
   - Transitions between game states
   - Pause/resume functionality
   - Save/load impact on UI state
   - Procedurally generated content display

**Phase 3: Issue Classification**
Categorize findings by severity:
- **Critical**: Blocks core functionality or causes crashes
- **High**: Significantly impairs usability or immersion
- **Medium**: Noticeable but doesn't prevent gameplay
- **Low**: Minor polish or enhancement opportunities

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
1. **Exploration Phase**: Interact with all visible UI elements systematically
2. **State Mapping**: Document UI state machine and transitions
3. **Stress Testing**: Test boundary conditions and rapid inputs
4. **Consistency Check**: Verify patterns hold across procedurally generated content
5. **Accessibility Review**: Evaluate usability for different player skill levels
6. **Performance Monitoring**: Note any UI-related performance degradation

EBITEN-SPECIFIC CONSIDERATIONS:
- Image rendering performance with large UI elements
- Input handling responsiveness (keyboard/gamepad)
- Draw call efficiency for UI layers
- Text rendering clarity at different resolutions
- Audio feedback integration with UI actions

EXAMPLE OUTPUT SNIPPET:

```markdown
### High Priority Issues

#### Issue #3: Inventory Grid Misalignment on Item Addition
- **Component**: Inventory Screen Grid Layout
- **Description**: When items are added to inventory during gameplay, grid cells shift by 2-3 pixels, causing visual jitter and misaligned item icons.
- **Steps to Reproduce**:
  1. Open inventory screen (press 'I')
  2. Return to gameplay
  3. Collect any item from the procedurally generated dungeon
  4. Reopen inventory screen
- **Expected Behavior**: Items should appear in properly aligned grid cells without layout shifts
- **Actual Behavior**: Grid elements shift slightly, icons appear offset from cell borders
- **Suggested Fix**: Implement fixed pixel coordinates for grid cells rather than calculating positions dynamically on each render. Consider using `ebiten.Image` with pre-calculated offsets stored in a constant array. Example:
  ```go
  var inventoryGridPositions = [20]image.Point{
      {X: 32, Y: 64}, {X: 96, Y: 64}, // etc.
  }
  ```
- **Ebiten-Specific Considerations**: Ensure grid rendering happens in a single draw call batch to prevent sub-pixel positioning artifacts
```
