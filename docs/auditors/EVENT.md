Create a comprehensive PLAN.md document that addresses critical implementation gaps and bugs in the game mechanics:

**Core Systems Integration**
- Audit and verify all game systems are properly instantiated and connected in the ECS architecture
- Ensure events flow correctly between systems (combat → status effects → UI updates)
- Validate status effect application, duration tracking, and removal mechanics
- Confirm all procedural generation systems (audio, visual, storytelling) are operational

**Input System & UI**
- Audit input handling for keyboard/mouse interactions across all game states
- Verify UI labels match actual key bindings and controller mappings
- Debug menu navigation issues, focus handling, and input response problems
- Test input during combat, inventory management, skill selection, and dialogue
- Ensure input state correctly reflects player actions in multiplayer scenarios

**Death & Revival Mechanics**
- Implement complete player/monster death state: zero health triggers immobilization
- Disable all actions (attacks, spells, movement, items) for dead entities
- Spawn dropped items at death location with physics-based positioning
- Create multiplayer revival system: teammate touch restores 20% health
- Add appropriate death animations, sound effects, and UI feedback
- Ensure network synchronization of death/revival states in multiplayer

**Critical Bug Fixes**
- Debug fog-of-war visibility calculations and rendering
- Fix status effect edge cases (stacking, expiration, conflicts)
- Resolve input conflicts between game modes (exploration vs combat vs menu)
- Test and fix multiplayer desync issues related to death/revival
- Validate combat calculations (damage, defense, critical hits)

**Testing Requirements**
- Create test scenarios for each identified gap
- Verify deterministic behavior across client/server
- Test edge cases: simultaneous deaths, revival during combat, dropped item interactions
- Ensure 60 FPS performance with all systems active

Organize PLAN.md with: Issue identification → Root cause analysis → Implementation steps → Testing verification → Success criteria. Prioritize by severity: game-breaking bugs first, then core mechanics, then polish.