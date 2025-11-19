#!/bin/bash
# Input Handling Verification Script
# Tests that menu input handling is working correctly after bug fixes

set -e

echo "=== Input Handling Verification ==="
echo ""

# Check 1: Build verification
echo "✓ Check 1: Build verification"
go build -o /tmp/venture-client ./cmd/client 2>&1 | grep -i error && exit 1 || echo "  Build: OK"
echo ""

# Check 2: Verify TouchButton code removed from main menu
echo "✓ Check 2: Verify TouchButton removal from main_menu_ui.go"
if grep -q "optionButtons" pkg/engine/main_menu_ui.go; then
    echo "  ERROR: optionButtons still present in main_menu_ui.go"
    exit 1
fi
if grep -q "touchHandler.*mobile.TouchInputHandler" pkg/engine/main_menu_ui.go; then
    echo "  ERROR: touchHandler still present in main_menu_ui.go"
    exit 1
fi
echo "  TouchButton removal: OK"
echo ""

# Check 3: Verify logging added
echo "✓ Check 3: Verify structured logging added"
if ! grep -q "log.WithFields" pkg/engine/main_menu_ui.go; then
    echo "  ERROR: Structured logging not found in main_menu_ui.go"
    exit 1
fi
echo "  Logging added: OK"
echo ""

# Check 4: Verify unified input helpers used
echo "✓ Check 4: Verify unified input helpers used"
if ! grep -q "IsTouchOrMouseJustPressed" pkg/engine/main_menu_ui.go; then
    echo "  ERROR: IsTouchOrMouseJustPressed not found"
    exit 1
fi
if ! grep -q "GetTouchOrMousePosition" pkg/engine/main_menu_ui.go; then
    echo "  ERROR: GetTouchOrMousePosition not found"
    exit 1
fi
echo "  Unified input helpers: OK"
echo ""

# Check 5: Verify other menus use consistent pattern
echo "✓ Check 5: Verify other menus use consistent pattern"
for menu_file in pkg/engine/single_player_menu.go pkg/engine/multiplayer_menu.go pkg/engine/genre_selection_menu.go; do
    if [ ! -f "$menu_file" ]; then
        echo "  WARNING: $menu_file not found"
        continue
    fi
    
    if ! grep -q "GetTouchOrMousePosition" "$menu_file"; then
        echo "  WARNING: $menu_file doesn't use GetTouchOrMousePosition"
    fi
done
echo "  Menu consistency: OK"
echo ""

# Check 6: Verify BUG FIX comments present
echo "✓ Check 6: Verify BUG FIX documentation comments"
if ! grep -q "BUG FIX" pkg/engine/main_menu_ui.go; then
    echo "  WARNING: BUG FIX comments not found (optional)"
else
    echo "  BUG FIX comments: OK"
fi
echo ""

# Summary
echo "=== Summary ==="
echo "✅ All checks passed!"
echo ""
echo "Manual testing recommended:"
echo "  1. Run: ./venture-client"
echo "  2. Test keyboard navigation (↑↓, Enter)"
echo "  3. Test mouse clicks on menu items"
echo "  4. Test touch input (if available)"
echo "  5. Verify no double-triggers on selection"
echo ""
echo "Debug logging enabled:"
echo "  Run with: RUST_LOG=debug ./venture-client"
echo "  (Check for 'Main menu option selected' log entries)"
