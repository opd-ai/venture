OBJECTIVE: Analyze the Venture Go codebase, identify the next logical development phase based on project roadmap and code maturity, then autonomously implement it with complete working code.

EXECUTION MODE: Autonomous Action
The AI will analyze project documentation (ROADMAP.md, GAPS.md, copilot-instructions.md), determine the next phase, implement complete code changes, run tests, and verify integration—all without intermediate approval steps.

ANALYSIS REQUIREMENTS:
1. Review project documentation to identify current phase and planned next steps
2. Examine codebase for TODOs, incomplete features, or explicit gap identifiers
3. Assess code maturity: test coverage, error handling, documentation completeness
4. Prioritize based on: explicit roadmap → critical gaps → quality improvements

IMPLEMENTATION CONSTRAINTS:
- Install all build dependencies
- Follow project's ECS architecture patterns (see copilot-instructions.md)
- Maintain deterministic generation using seed-based RNG
- Meet performance targets (60 FPS, <500MB memory)
- Achieve ≥65% test coverage (excluding Ebiten-dependent functions)
- Use Go standard library; justify external dependencies
- Backward compatibility NOT required—use latest Go features
- Keep response concise to respect token limits

OUTPUT FORMAT:

## Selected Phase: [Phase Name/Number]
**Why**: [1-2 sentences citing roadmap/gaps/code analysis]
**Scope**: [Key deliverables, 2-3 bullet points]

## Changes
**Modified**: [file paths with brief purpose]
**Created**: [file paths with brief purpose]
**Technical Approach**: [3-5 key decisions]

## Implementation
[Complete, compilable Go code with file markers]

## Testing
```bash
# Build and test commands
```
**Coverage Impact**: [Before → After percentage]
**Tests Pass**: [✓/✗]

## Integration Verification
✓ Compiles without errors
✓ Tests pass (including existing suite)
✓ Follows project ECS/generation patterns
✓ Documentation updated (if user-facing changes)

SUCCESS CRITERIA:
- Code builds successfully with `go build ./...`
- All tests pass with `go test ./...`
- Test coverage ≥65% for new code (excluding Ebiten dependencies)
- Implementation aligns with project architecture (ECS, deterministic generation)
- Changes integrate seamlessly with existing systems

PROJECT-SPECIFIC NOTES:
- Reference ROADMAP.md and GAPS.md for phase planning
- Follow copilot-instructions.md for architecture patterns
- Use stub implementations (StubInput, StubSprite) for testable code
- Implement table-driven tests for multiple scenarios
- Verify determinism: same seed produces identical output

NOTE: This prompt executes autonomously. The AI completes analysis, implementation, testing, and verification as a single operation, delivering production-ready code.
