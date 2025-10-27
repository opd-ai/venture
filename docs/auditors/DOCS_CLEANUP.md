# Task: Audit and Document Public API Design

**Objective:** Analyze the public API surface of this Ebiten-based Go 2D game to ensure it is complete, safe, and extensible. Produce a comprehensive audit document identifying design issues and improvement opportunities.

**Execution Mode:** Report generation - Output findings to `docs/API_AUDIT.md` without making code changes.

**Scope:**
Examine all exported (public) Go identifiers across the codebase:
- Structs, interfaces, functions, methods, constants
- Constructor patterns and initialization safety
- Interface consistency and design patterns
- Default value safety and zero-value usability

**Evaluation Criteria:**
1. **Completeness** - API provides sufficient functionality for downstream modifications/mods
2. **Safety** - Design prevents common misuse patterns ("footgun" elimination)
3. **Consistency** - Similar operations follow uniform patterns across packages
4. **Constructors** - All structs have constructor functions with sane defaults
5. **Documentation** - Public APIs have godoc comments
6. **Zero-value Safety** - Structs are usable with zero values or have required constructors
7. **Mod-Friendliness** - Extensibility points (interfaces, hooks, callbacks) exist where appropriate

**Output Format:**
Create `docs/API_AUDIT.md` with:
```
# API Audit Report

## Executive Summary
- Total packages/exports analyzed
- Critical issues count
- Safety concerns count
- Consistency violations count

## Package-by-Package Analysis
For each package in pkg/:
### pkg/<package_name>
- **Exported Types:** [count]
- **Constructor Coverage:** [X/Y have constructors]
- **Issues Found:** [count]
- **Findings:**
  - [Issue description with examples]
  - [Remediation recommendation]

## Cross-Package Concerns
- Interface consistency issues
- Naming convention violations
- Pattern inconsistencies

## Remediation Priorities
1. Critical (breaks safety/causes panics)
2. High (footguns, inconsistencies)
3. Medium (missing constructors, docs)
4. Low (nice-to-have improvements)

## Recommendations Summary
[Actionable list of changes organized by priority]
```

**Success Criteria:**
- All packages under `pkg/` analyzed
- Each exported type evaluated for constructor presence
- Concrete examples provided for issues found
- Remediation steps are specific and actionable
- Report is organized by severity/priority