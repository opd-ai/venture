# TASK: Execute Next Planned Item for Venture Go Project

## EXECUTION MODE: Autonomous Action
Implement the next task(s) directly. No user approval needed between steps.

## OBJECTIVE
Read task files in strict priority order, find the first incomplete task, and implement it with tests and documentation. Update the task file upon completion.

## TASK FILE PRIORITY (strict — never skip)

1. **AUDIT.md** — immediate fixes (always check first)
2. **PLAN.md** — medium-term work (only if AUDIT.md is absent or fully complete)
3. **ROADMAP.md** — long-term goals (only if both above are absent or fully complete)

**Rules:**
- Never work on a lower-priority file while a higher-priority file has open items — including optional items.
- If a file's tasks are all complete, delete it, then proceed to the next file.
- Process tasks in document order. Do not reorder or skip.

.## TASK GROUPING (1–20 tasks per execution)
You may batch 2–20 tasks if they share **any** of these traits:
- Same package, module, or closely related packages
- Similar in nature (e.g., all are doc fixes, all are test gaps, all are validation bugs)
- Combined diff stays under 800 lines
- Can be validated together without complex test isolation

AND this trait(required):
- Overlapping code context (you'd read similar code for each)

Single tasks that are large or span multiple unrelated packages should be executed alone.

## IMPLEMENTATION PROCESS

1. **Identify**: Read the active task file. Find the first `[ ]` item(s). Note acceptance criteria.
2. **Plan**: Write a brief approach as code comments before implementing.
3. **Implement**: Minimal viable solution. Prefer standard library; external deps require >1K GitHub stars and recent maintenance.
4. **Test**: Table-driven unit tests, >80% coverage on new/changed business logic, including error paths.
5. **Validate**: `go fmt`, `go vet`, `go test ./affected/packages/...` must all pass. No regressions.
6. **Document**: Godoc on exported symbols. Update README only if public API or usage changes.
7. **Update task file**: Mark completed items `[x]` with date and brief summary. Delete the file if all items are now complete.

## CODE STANDARDS (Venture-specific)

- **ECS**: Components are pure data with `Type() string` only. Logic belongs in Systems.
- **Determinism**: All procedural generation uses `rand.New(rand.NewSource(seed))`. Never `time.Now()` or global `math/rand` for gameplay state.
- **Logging**: `logrus.WithFields(logrus.Fields{...})` — never `fmt.Print`/`log.Fatal`. Standard fields: `seed`, `genre`, `entityID`, `playerID`, `system_name`, `component_type`.
- **Networking**: Use interface types (`net.Addr`, `net.PacketConn`, `net.Conn`, `net.Listener`). No type assertions to concrete types.
- **Functions**: ≤30 lines, single responsibility, all errors handled explicitly.
- **Testing**: Table-driven with `t.Run`. Use stubs (`StubInput`, `StubSprite`) to avoid Ebiten runtime dependency. Target ≥40% per package (≥30% for X11/Wayland/Ebiten-dependent packages).

## EXPECTED OUTPUT
- Working code changes with passing tests
- Updated task file reflecting completed work
- Brief summary of what was done and which task(s) were completed

## SUCCESS CRITERIA
- All existing tests still pass (`go test ./...` — no regressions)
- New code has unit tests covering success and error paths
- Task file accurately reflects current state
- Code passes `go fmt` and `go vet`

## SIMPLICITY RULE
If your solution needs more than 3 levels of abstraction, redesign for clarity. Boring and maintainable beats clever and elegant.
