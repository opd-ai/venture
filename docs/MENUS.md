## In-Game Menu Navigation Standardization

### Objective
Establish and enforce a consistent, user-friendly navigation pattern across all in-game menus in the Venture action-RPG client.

### Requirements

**1. Menu Activation**
- Each menu must be assigned a unique, mnemonic letter key (e.g., "I" for Inventory, "M" for Map, "C" for Character, "S" for Skills, "Q" for Quests)
- The assigned key should be intuitive and match the first letter of the menu name when possible
- Document all key assignments in a central configuration or constants file for maintainability

**2. Menu Exit Behavior (Critical)**
- Every menu MUST support TWO exit mechanisms:
  - **Toggle key**: The same letter key that opened the menu (e.g., pressing "I" again closes Inventory)
  - **Universal exit**: The Escape key must close any open menu
- Both exit methods must function identically and simultaneously—users should be able to use either at any time

**3. Implementation Checklist**
- Audit all existing menu systems (inventory, character sheet, skills, quests, map, settings, etc.)
- Verify each menu's input handler processes both its toggle key and Escape key for closure
- Ensure no menu is "trapped" (requiring specific exit actions other than these two methods)
- Test edge cases: rapid key presses, multiple menus open simultaneously (if applicable), menu transitions

**4. User Experience Considerations**
- Provide visual indicators showing the assigned key for each menu (e.g., "[I] Inventory" in UI)
- Display "Press [KEY] or [ESC] to close" hint within active menus
- Ensure consistent behavior across all game states (combat, exploration, multiplayer)
- Handle conflicts gracefully if multiple systems attempt to capture the same keys

**5. Testing Validation**
- Create test cases for each menu's open/close cycle
- Verify both exit methods work from various game states
- Confirm no regression in existing functionality
- Test with both keyboard and potential gamepad inputs (if applicable)

### Expected Outcome
A polished, predictable menu system where players can intuitively navigate using consistent key bindings, with the flexibility to exit any menu using either the toggle key or Escape, enhancing overall game usability and reducing player frustration.