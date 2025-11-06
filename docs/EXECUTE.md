# TASK: Execute Next Planned Task

## OBJECTIVE
Identify and implement the next incomplete task from docs/PLAN.md (short-term) or docs/ROADMAP.md (long-term). Prioritize PLAN.md tasks. If PLAN.md is complete, verify thoroughly then delete it. Complete exactly ONE task following Venture's ECS architecture, procedural generation patterns, and testing standards.

## EXECUTION MODE
**Autonomous Action** - Implement the task immediately with full testing and documentation.

## CONTEXT
- **Project**: Venture (Go 1.24+, Ebiten 2.9) - Procedural action-RPG with ECS architecture
- **Current Phase**: Phase 9 Post-Beta Enhancement (see docs/ROADMAP.md)
- **Key Constraints**: Deterministic generation (seed-based), 60 FPS target, >65% test coverage per package

## IMPLEMENTATION WORKFLOW

### 1. Analysis (Read First)
- Check `docs/PLAN.md` for next incomplete task
- If PLAN.md doesn't exist or is complete, check `docs/ROADMAP.md` 
- Extract clear acceptance criteria and affected systems

### 2. Design (Before Coding)
- Review `.github/copilot-instructions.md` for project patterns
- Identify affected packages under `pkg/` (engine, procgen, rendering, etc.)
- Choose standard library or minimal dependencies (logrus, zenity, golang.org/x/image only)
- Document approach in code comments explaining WHY

### 3. Implementation
- **ECS Pattern**: Entities (IDs), Components (data only), Systems (logic)
- **Deterministic**: Use `rand.New(rand.NewSource(seed))`, never `time.Now()` or global `math/rand`
- **Functions**: <50 lines, single responsibility, explicit error handling
- **Naming**: MixedCaps (not snake_case), descriptive over abbreviated
- **No Ebiten in Tests**: Use stub implementations (StubInput, StubSprite, etc.)

### 4. Testing
- Table-driven tests for multiple scenarios
- Test success AND error paths
- Verify determinism: same seed = identical output
- Target: >65% coverage (exclude Ebiten-dependent rendering functions)
- Run: `go test -race ./...` before completion

### 5. Documentation
- GoDoc comments for all exported elements (start with element name)
- Update package `doc.go` if adding features
- Update relevant docs in `docs/` directory
- Add usage examples for CLI tools in cmd/

### 6. Validation & Reporting
- [ ] Passes `go fmt`, `go vet`, and `go test -race ./...`
- [ ] Test coverage >65% for affected packages
- [ ] No circular dependencies, follows pkg/ hierarchy
- [ ] Deterministic generation verified (if applicable)
- [ ] Update PLAN.md or ROADMAP.md with completion status

## OUTPUT FORMAT
Provide brief status updates:
1. **Task Identified**: "[Task name from PLAN.md/ROADMAP.md]"
2. **Implementation**: Concise progress during work
3. **Completion**: "✅ [Task] complete. Coverage: X%. Tests passing."

## SUCCESS CRITERIA
- Exactly one task completed with zero regressions
- All tests pass: `go test ./...`
- Code follows Venture conventions (see copilot-instructions.md)
- Planning document updated to reflect completion
- No performance degradation (<16.67ms frame time maintained)

## VENTURE-SPECIFIC RULES
- **Package Structure**: Follow `pkg/` organization, avoid circular deps
- **Generators**: Implement `procgen.Generator` interface with Validate()
- **Performance**: Profile before optimizing (use Makefile targets)
- **Logging**: Use structured logging with logrus (see pkg/logging/)
- **No Asset Files**: All content procedurally generated at runtime