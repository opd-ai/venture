Conduct automated code review of venture repository packages, prioritizing by dependency depth (lowest first) and excluding packages with existing AUDIT.md files.

**Execution Mode:** Autonomous action - select package, review, create AUDIT.md.

**Package Selection Algorithm:**
1. Scan `pkg/` directory structure recursively
2. Exclude packages with existing `AUDIT.md` files
3. Analyze Go imports to determine dependency depth (packages with fewest internal dependencies = lowest depth)
4. Select package with lowest dependency depth that hasn't been audited
5. Prioritize foundational packages: `engine` → `procgen/*` → `rendering/*` → higher-level packages

**Review Process (per CODE_REVIEW_PLAN.md):**
1. **Static Analysis:** `go vet`, `gofmt -l`, compilation check
2. **Structure:** Package docs, file organization, naming conventions
3. **API Design:** Godoc coverage, error handling, interface compliance
4. **Pattern Compliance:** ECS components (data-only), generators (determinism), systems (stateless queries)
5. **Testing:** Coverage ≥65% (excluding Ebiten init functions), race detection
6. **Concurrency:** Resource safety, proper cleanup, no leaks
7. **Error Handling:** All returns checked, wrapped with context, validated inputs

**Output Format (`pkg/[SELECTED_PACKAGE]/AUDIT.md`):**
```markdown
# Code Review Audit: [package name]
**Date:** [ISO date]
**Reviewer:** GitHub Copilot
**Dependency Depth:** [N]

## Executive Summary
[Pass/Fail with brief assessment]

## Quality Gates
- [ ] Build success
- [ ] All tests pass
- [ ] Race-free
- [ ] Coverage ≥65%
[... remaining 14 gates from CODE_REVIEW_PLAN.md]

## Findings
### Critical (blocks merge)
[file:line - issue - fix]

### Major (should fix)
[file:line - issue - fix]

### Minor (nice-to-have)
[file:line - issue - fix]

## Recommendations
[Actionable next steps]
```

**Success Criteria:**
- Exactly one package selected and reviewed
- AUDIT.md created only in selected package directory
- All findings reference specific file:line locations
- Code snippets provided for non-trivial issues
- Selection logged: "Reviewing pkg/[name] (depth: N, no prior audit)"