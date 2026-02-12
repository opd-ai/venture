TASK: Audit one Go sub-package for implementation completeness; produce structured findings.

EXECUTION MODE: Autonomous action — select package, audit, write files, report.

PACKAGE SELECTION:
1. Read root `AUDIT.md` to identify already-audited packages.
2. Pick ONE un-audited sub-package from `pkg/` or `cmd/`. Prefer packages that:
   - Are listed in root `AUDIT.md` but unchecked
   - Have high integration surface (many imports/importers)
   - Fall under core domains: engine, procgen, rendering, network, world
3. State the chosen package and rationale before proceeding.

AUDIT CHECKLIST — evaluate each item, cite file:line for every issue found:

| Category | What to check |
|---|---|
| **Stub/incomplete code** | Functions returning only `nil`/zero, `TODO`/`FIXME`/`placeholder` comments, empty method bodies |
| **ECS compliance** | Components must be pure data + `Type() string` only; no logic methods on components; systems must own all behavior |
| **Deterministic procgen** | All randomness via `rand.New(rand.NewSource(seed))`; no global `rand`, `time.Now()`, or OS entropy |
| **Network interfaces** | Variables use `net.Addr`/`net.PacketConn`/`net.Conn`/`net.Listener` — never concrete types; no type assertions to concrete net types |
| **Error handling** | No swallowed errors; all returned errors checked; structured logging with `logrus.WithFields` on error paths |
| **Test coverage** | Run `go test -cover ./path/to/pkg/...`; flag if below 65% target; note missing table-driven tests or benchmarks |
| **Doc coverage** | Exported types/functions have godoc comments; package has `doc.go` |
| **Integration points** | Verify registration in `system_init.go` / `handlers.go` where applicable; check serialize/deserialize support for persistent components |

OUTPUT FILES:

1. **Create `<package-dir>/AUDIT.md`** with this exact template:
```markdown
# Audit: <package-import-path>
**Date**: YYYY-MM-DD
**Status**: Complete | Incomplete | Needs Work

## Summary
<2-3 sentences: scope, overall health, critical risk>

## Issues Found
- [ ] <severity:high|med|low> <category> — <description> (`file.go:LINE`)
- [ ] ...

## Test Coverage
<percentage> (target: 65%)

## Integration Status
<How this package connects to engine, client, server; any missing registrations>

## Recommendations
1. <highest-priority fix>
2. <next fix>
```

2. **Update root `AUDIT.md`** — change the matching unchecked line to checked and append status:
```markdown
- [x] `path/to/AUDIT.md` — <Status> — <N> issues (<H> high, <M> med, <L> low)
```
   If the package is not yet listed, append a new checked entry.

FINAL REPORT (print to chat after files are written):
- Path to created AUDIT.md
- Test coverage percentage
- Top 3-5 critical findings, each with file:line
- `go vet ./path/to/pkg/...` pass/fail confirmation

SUCCESS CRITERIA:
- Sub-package AUDIT.md exists, follows template exactly, every issue has file:line
- Root AUDIT.md updated with checked entry
- `go vet` passes on audited package
- Findings reference codebase-specific standards (ECS purity, deterministic seeds, interface networking)