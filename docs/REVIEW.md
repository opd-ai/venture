Conduct automated code review of recently changed files in the venture repository, analyzing the last 3 commits and autonomously resolving true positive issues.

**Execution Mode:** Autonomous action - identify changed files, review, resolve issues, create AUDIT.md.

**File Selection Algorithm:**
1. Execute `git log -3 --name-only --pretty=format:` to get files changed in last 3 commits
2. Filter for `.go` files only
3. Exclude files in directories with existing `AUDIT.md` for the current date
4. Prioritize files by change frequency (most changed = highest priority)
5. Select top file that hasn't been audited today

**Before You Begin (install dependencies):**
- Install build dependencies from README.md and tests.yml using `sudo apt install`
- Install xvfb using `sudo apt install`

**Review Process (per DEVELOPMENT.md and CONTRIBUTING.md):**
1. **Static Analysis:** `go vet`, `gofmt -l`, compilation check on changed file
2. **Structure:** Package docs, file organization, naming conventions
3. **API Design:** Godoc coverage, error handling, interface compliance
4. **Pattern Compliance:** ECS components (data-only), generators (determinism), systems (stateless queries)
5. **Testing:** Coverage ≥40% (≥30% for X11/Wayland/Ebiten-dependent packages), race detection
6. **Concurrency:** Resource safety, proper cleanup, no leaks
7. **Error Handling:** All returns checked, wrapped with context, validated inputs

**Issue Analysis & Resolution:**
For each finding:
1. **Determine False Positive:** Check if issue violates actual project guidelines from copilot-instructions.md
    - Ignore style preferences not in coding standards
    - Ignore issues in generated code or vendor directories
    - Ignore Ebiten initialization warnings (documented as untestable)
2. **Resolve True Positives:** If issue is valid:
    - Apply fix directly to the file
    - Run `go fmt` on modified file
    - Verify fix with `go vet` and `go test` for affected package
    - Document fix in AUDIT.md

**Output Format (`pkg/[PACKAGE_DIR]/AUDIT.md` or `cmd/[CMD_DIR]/AUDIT.md`):**
```markdown
# Code Review Audit: [file path]
**Date:** [ISO date]
**Reviewer:** GitHub Copilot
**Commits Analyzed:** Last 3
**Change Frequency:** [N times]

## Executive Summary
[Pass/Fail with brief assessment and auto-fix summary]

## Quality Gates
- [ ] Build success
- [ ] All tests pass
- [ ] Race-free
- [ ] Coverage ≥40% (≥30% for X11/Wayland/Ebiten-dependent packages)
[... remaining quality gates per DEVELOPMENT.md]

## Findings & Resolutions
### Critical (blocks merge)
**[file:line - issue]**
- Status: [RESOLVED/FALSE_POSITIVE/REQUIRES_MANUAL]
- Rationale: [why false positive or fix explanation]
- Fix Applied: [code diff if resolved]

### Major (should fix)
**[file:line - issue]**
- Status: [RESOLVED/FALSE_POSITIVE/REQUIRES_MANUAL]
- Rationale: [why false positive or fix explanation]
- Fix Applied: [code diff if resolved]

### Minor (nice-to-have)
**[file:line - issue]**
- Status: [RESOLVED/FALSE_POSITIVE/REQUIRES_MANUAL]
- Rationale: [why false positive or fix explanation]
- Fix Applied: [code diff if resolved]

## Auto-Fix Summary
- Files Modified: [N]
- Issues Resolved: [N]
- False Positives: [N]
- Manual Review Required: [N]

## Recommendations
[Actionable next steps for remaining issues]
```

**Success Criteria:**
- Exactly one file selected and reviewed from last 3 commits
- AUDIT.md created in appropriate package/cmd directory
- All true positive issues automatically resolved where possible
- False positives documented with rationale
- All findings reference specific file:line locations
- Modified files pass `go fmt`, `go vet`, and package tests
- Selection logged: "Reviewing [file] (changed N times in last 3 commits)"

