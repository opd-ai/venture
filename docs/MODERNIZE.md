**Objective:** Modernize codebase by removing all backward-compatibility code and deprecated features in favor of current implementations.

**Execution Mode:** Autonomous action with automatic implementation.

**Scope:**
1. Identify and remove deprecated code paths, legacy feature flags, and compatibility shims
2. Delete associated tests for removed features
3. Update imports and dependencies to use latest patterns only
4. Simplify conditional logic that branches on feature availability

**Constraints:**
- Preserve ALL currently active features and their tests
- Maintain deterministic generation (seed-based RNG)
- Keep ECS architecture patterns intact
- Ensure test coverage remains ≥65% per package

**Output:**
- Modified files with backward-compatibility removed
- Brief summary of removed features/code paths
- Confirmation that all tests pass post-cleanup

**Success Criteria:**
- `go test ./...` passes
- `go build ./cmd/client` and `./cmd/server` succeed
- `gofmt -w -s` applied
- No deprecated feature flags or version checks remain
- Code complexity reduced (fewer conditional branches)