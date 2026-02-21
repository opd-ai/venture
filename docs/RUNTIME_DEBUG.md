TASK: Run the Venture game client, capture logs, and fix all errors and harmful warnings iteratively.

EXECUTION MODE: Autonomous action — run, diagnose, fix, repeat until clean.

---

## Objective

Eliminate all runtime errors and harmful warnings from the Venture game client by iteratively running the client, analyzing logs, and fixing root causes in the codebase.

---

## Procedure

### Phase 1: Error Resolution Loop

1. Run the client with a 60-second timeout, capturing all output:
   ```bash
   timeout 60s go run ./cmd/client 2>&1 | tee venture.log
   ```
2. Extract errors:
   ```bash
   grep 'ERRO' venture.log
   ```
3. If errors exist:
   - For each unique error, trace it to its source file and line using the log context fields (`system_name`, `component_type`, etc.) and stack traces.
   - Fix the root cause in the codebase. Do not suppress or silence errors.
   - Ensure `go build ./cmd/client` and `go vet ./cmd/client` pass after each fix.
   - Return to step 1.
4. If no errors exist, proceed to Phase 2.

### Phase 2: Warning Triage Loop

1. Extract warnings from the most recent `venture.log`:
   ```bash
   grep 'WARN' venture.log
   ```
2. For each unique warning, evaluate whether it indicates:
   - **A real problem** (nil components, missing registrations, failed initialization, data corruption, resource leaks) → Fix the root cause.
   - **A benign condition** (optional feature unavailable, expected fallback path, informational notice) → Skip. Do not fix warnings that are functioning as intended.
3. If any warnings were fixed, re-run the client (return to Phase 1 step 1 to confirm no new errors were introduced).
4. If no fixable warnings remain, proceed to Phase 3.

### Phase 3: Final Verification

1. Run the client one final time:
   ```bash
   timeout 60s go run ./cmd/client 2>&1 | tee venture.log
   ```
2. Confirm:
   - `grep 'ERRO' venture.log` returns zero results.
   - All remaining `WARN` lines are benign (list them with justification).
   - `go build ./...` succeeds.
   - `go vet ./...` passes.

---

## Constraints

- Do not suppress errors by removing log statements or lowering log levels.
- Do not modify logging infrastructure or log format.
- Do not break existing tests: `go test ./...` for any modified package must pass.
- Maintain ECS architecture rules: components are pure data, systems own logic.
- Preserve deterministic procgen: no `time.Now()` or global `rand` calls.

## Output

After all phases complete, report:

1. **Errors fixed**: List each error message, root cause file:line, and fix applied.
2. **Warnings fixed**: List each warning, why it was harmful, and fix applied.
3. **Warnings skipped**: List each remaining warning with justification for why it is benign.
4. **Verification**: Confirm `go build`, `go vet`, and `go test` status for all modified packages.