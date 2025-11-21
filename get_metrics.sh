#!/bin/bash

echo "=== REFACTORED FUNCTION METRICS COMPARISON ==="
echo

# Function 1: BuildEntitySnapshot
echo "1. BuildEntitySnapshot (pkg/network/snapshot_builder.go)"
echo "   BEFORE:"
jq -r '.functions | map(select(.name == "BuildEntitySnapshot" and .file == "pkg/network/snapshot_builder.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' baseline.json
echo "   AFTER:"
jq -r '.functions | map(select(.name == "BuildEntitySnapshot" and .file == "pkg/network/snapshot_builder.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' refactored.json
echo

# Function 2: addWaterHazards  
echo "2. addWaterHazards (pkg/procgen/terrain/maze.go)"
echo "   BEFORE:"
jq -r '.functions | map(select(.name == "addWaterHazards" and .file == "pkg/procgen/terrain/maze.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' baseline.json
echo "   AFTER:"
jq -r '.functions | map(select(.name == "addWaterHazards" and .file == "pkg/procgen/terrain/maze.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' refactored.json
echo

# Function 3: createPark
echo "3. createPark (pkg/procgen/terrain/city.go)"
echo "   BEFORE:"
jq -r '.functions | map(select(.name == "createPark" and .file == "pkg/procgen/terrain/city.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' baseline.json
echo "   AFTER:"
jq -r '.functions | map(select(.name == "createPark" and .file == "pkg/procgen/terrain/city.go")) | .[0] | "   Overall: \(.complexity.overall), Cyclomatic: \(.complexity.cyclomatic), Nesting: \(.complexity.nesting_depth), Lines: \(.lines.total)"' refactored.json
echo

# Helper functions created - count them
echo "   HELPER FUNCTIONS CREATED:"
jq -r '.functions | map(select(.file == "pkg/network/snapshot_builder.go" and (.name | startswith("serialize")))) | length' refactored.json | xargs echo "   - snapshot_builder.go:"
jq -r '.functions | map(select(.file == "pkg/procgen/terrain/maze.go" and (.name == "isValidWaterLocation" or .name == "isPointInRoom" or .name == "placeWaterPool" or .name == "addShallowWaterTiles"))) | length' refactored.json | xargs echo "   - maze.go:"
jq -r '.functions | map(select(.file == "pkg/procgen/terrain/city.go" and (.name | startswith("fill") or startswith("place") or startswith("maybe")))) | length' refactored.json | xargs echo "   - city.go:"
