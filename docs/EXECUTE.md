# Phase: Execute Next Planned Phase

## OBJECTIVE
Identify and implement the next incomplete Phase from docs/PLAN.md (short-term) or docs/ROADMAP_V4.md (long-term). Prioritize PLAN.md Phases. If PLAN.md is complete, verify thoroughly then delete it. Complete exactly ONE Phase following Venture's ECS architecture, procedural generation patterns, and testing standards. Work in complete sections, implement the entire objective defined by the planning document, do not stop until you are finished. You are not done until all tests pass.

## EXECUTION MODE
**Autonomous Action** - Implement the Phase immediately with full testing and documentation.

## CONTEXT
- **Project**: Venture (Go 1.24+, Ebiten 2.9) - Procedural action-RPG with ECS architecture
- **Current Version**: 3.0.0 Production (Phases 15-20 complete); V4.0 in progress (Phases 21-29 complete, Phase 30 status: planning)
- **Key Constraints**: Deterministic generation (seed-based), 60 FPS target, >65% test coverage per package

## IMPLEMENTATION WORKFLOW

### 1. Analysis (Read First)
- Check `docs/PLAN.md` for next incomplete Phase
- If PLAN.md doesn't exist or is complete, check `docs/ROADMAP_V4.md` for current development phases
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
- Install dependencies per `test.yml` workflow and use `xvfb` to run test suite.
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
- [ ] Update PLAN.md or ROADMAP_V4.md with completion status, avoid extra detail
- [ ] Do not create unnecessary new `*.md` or `docs/*.md` files—update existing documentation as required by the workflow, and write excellent godoc instead.

> **Checklist convention:** Mark completed items with `[x]` and incomplete items with `[ ]` for consistency.

## OUTPUT FORMAT
Provide brief status updates:
1. **Phase Identified**: "[Phase name from PLAN.md/ROADMAP_V4.md]"
2. **Implementation**: Concise progress during work
3. **Completion**: "✅ [Phase] complete. Coverage: X%. Tests passing."

## SUCCESS CRITERIA
- Exactly one Phase completed with zero regressions
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
- **Replace, Don't Accumulate**: When a system/technique is completely replaced (e.g., improved sprites, better generation algorithms), remove the old implementation entirely without backward compatibility concerns. Only keep code that still serves a purpose.
	- *Example*: If a new terrain generator fully supersedes the old one, delete the old generator code and tests instead of leaving them commented out or in an "old" directory. Do not keep deprecated code for reference—use version control history if needed.
