# Choice Consequences Package

This package implements persistent choice tracking and consequence systems for the Venture action-RPG.

## Package Structure

The package is organized into focused, single-responsibility files:

### Core Files

- **`doc.go`** - Package-level documentation with usage examples and integration points
- **`types.go`** - Core domain types (PlayerChoice, NPCRelationship, ContentLock, QuestBranch, etc.) and the ECS component
- **`alignment.go`** - Player moral alignment types and logic (AlignmentShift, PlayerAlignment, AlignmentRequirement)
- **`choice_tracker.go`** - Main ChoiceTracker implementation with all business logic
- **`helpers.go`** - Internal utility functions (clamp, abs)

### Test Files

- **`manager_test.go`** - Comprehensive test suite (24 tests, 84.2% coverage)

## File Organization Rationale

**Why separate `alignment.go` from `types.go`?**
- Alignment logic is self-contained with its own types and methods
- Reduces `types.go` complexity from 196 to 122 lines
- Makes alignment system easier to understand and modify independently
- Common pattern in Go projects (e.g., `time.go` vs `format.go` in stdlib)

**Why `choice_tracker.go` instead of `manager.go`?**
- More descriptive name matches the primary type: `ChoiceTracker`
- Follows Go convention: primary implementation file named after main type
- Examples: `http.Client` → `client.go`, `sql.DB` → `db.go`

**Why extract `helpers.go`?**
- Separates pure utility functions from business logic
- Makes testing and reuse easier
- Reduces cognitive load when reading business logic

## Usage

See `doc.go` for comprehensive usage examples and integration patterns.

Quick start:

```go
import "github.com/opd-ai/venture/pkg/integration/choice_consequences"

// Create tracker
tracker := choice_consequences.NewChoiceTracker()

// Record a choice
choice := &choice_consequences.PlayerChoice{
    ChoiceID:    "quest_spare_bandit",
    StoryNodeID: "village_burned",
    Timestamp:   time.Now().Unix(),
    MoralAlignment: &choice_consequences.AlignmentShift{
        GoodEvil: 0.2,  // Good action
        LawChaos: -0.1, // Chaotic mercy
    },
    Irreversible: true,
}
tracker.RecordChoice("player123", choice)

// Check consequences
available := tracker.IsContentAvailable("player123", "quest_bandit_redemption")
attitude := tracker.GetNPCAttitude("player123", "villager_elder")
alignment := tracker.GetAlignment("player123")
```

## Testing

Run tests:
```bash
go test ./pkg/integration/choice_consequences/...
```

With coverage:
```bash
go test -cover ./pkg/integration/choice_consequences/...
```

## Quality Metrics

- **Test Coverage**: 84.2%
- **Tests**: 24 passing
- **Lines of Code**: ~750 (excluding tests)
- **Documentation**: 100% of exported symbols
- **Build Status**: ✅ Passing
- **Go Vet**: ✅ No issues

## Known Issues

See `AUDIT.md` for detailed implementation gap analysis. Key items:

1. **Incomplete feature**: `applyConsequence()` only handles 3 of 6 LockType values
2. **Untested code**: `abs()` helper function has 0% coverage (low risk)

## Contributing

When modifying this package:

1. **Maintain file organization** - add related types/functions to appropriate files
2. **Keep test coverage ≥80%** - add tests for new functionality
3. **Document exported symbols** - all public APIs need godoc comments
4. **Run `go vet`** - ensure no static analysis warnings
5. **Update AUDIT.md** - document any new implementation gaps

## Integration Points

This package integrates with:

- **V8 Branching Narratives** - Story graph management
- **V4 Reputation** - NPC relations tracking  
- **V4 Classes** - Class-specific content unlocking
- **V8 Companion Learning** - Alignment-based reactions

See `doc.go` for detailed integration examples.
