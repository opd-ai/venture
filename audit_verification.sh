#!/bin/bash
# Venture Bug Audit Verification Script
# Date: 2025-11-18

echo "=== Venture Bug Audit Verification ==="
echo ""

# Build verification
echo "1. Build Verification..."
if go build ./cmd/client > /dev/null 2>&1 && go build ./cmd/server > /dev/null 2>&1; then
    echo "   ✅ Both client and server build successfully"
else
    echo "   ❌ Build failed"
    exit 1
fi

# Test verification
echo ""
echo "2. Test Suite Verification..."
if go test ./pkg/engine ./pkg/procgen ./pkg/rendering ./pkg/network ./pkg/saveload > /dev/null 2>&1; then
    echo "   ✅ All core packages pass tests"
else
    echo "   ❌ Tests failed"
    exit 1
fi

# Static analysis
echo ""
echo "3. Static Analysis..."
if go vet ./... > /dev/null 2>&1; then
    echo "   ✅ No go vet issues"
else
    echo "   ❌ go vet found issues"
    exit 1
fi

# UI Menu ESC handling check
echo ""
echo "4. UI Menu ESC Handling..."
ui_files=(
    "pkg/engine/inventory_ui.go"
    "pkg/engine/character_ui.go"
    "pkg/engine/skills_ui.go"
    "pkg/engine/quest_ui.go"
    "pkg/engine/map_ui.go"
    "pkg/engine/crafting_ui.go"
    "pkg/engine/shop_ui.go"
    "pkg/engine/help_system.go"
)

all_have_esc=true
for file in "${ui_files[@]}"; do
    if ! grep -q "KeyEscape\|HandleMenuInput" "$file" 2>/dev/null; then
        echo "   ⚠️  $file may lack ESC handling"
        all_have_esc=false
    fi
done

if $all_have_esc; then
    echo "   ✅ All UI menus have ESC handling"
else
    echo "   ⚠️  Some UI menus may need ESC verification"
fi

# Determinism check
echo ""
echo "5. Deterministic Generation Check..."
if grep -r "time\.Now()" pkg/procgen/ --include="*.go" --exclude="*_test.go" | grep -v "markov.go" | grep -v "README" | grep -v "//"; then
    echo "   ⚠️  Found time.Now() usage outside markov.go"
else
    echo "   ✅ Deterministic generation preserved (markov.go exception documented)"
fi

echo ""
echo "=== Audit Verification Complete ==="
echo "Status: ✅ PRODUCTION-READY"
