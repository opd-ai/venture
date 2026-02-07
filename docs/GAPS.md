TASK: Generate a comprehensive instantiation audit report for the Go/Ebiten game codebase.

OBJECTIVE: Verify that all game systems, components, and structs defined across the codebase are properly instantiated in runtime code (client/server entrypoints and game loops), not only in test files.

SCOPE:
1. Scan all packages in the #codebase for:
   - Struct definitions (game systems, components, managers)
   - Interface implementations
   - System registration patterns

2. Cross-reference findings against:
   - Client entrypoint initialization code
   - Server entrypoint initialization code  
   - Main game loop(s)
   - System enable/disable configuration

3. Identify discrepancies where components are:
   - Defined but never instantiated in runtime
   - Instantiated but disabled by default
   - Only used in test files

OUTPUT FORMAT:
```
## Instantiation Audit Report

### Summary
- Total systems/components found: [N]
- Properly instantiated: [N]
- Missing from runtime: [N]
- Disabled by default: [N]

### Properly Instantiated Systems
[List with file locations]

### Issues Found

#### Not Instantiated in Runtime
- **[System/Component Name]**
  - Defined: [file:line]
  - Expected location: [entrypoint/gameloop file]
  - Status: Missing

#### Disabled by Default
- **[System Name]**
  - Location: [file:line]
  - Configuration: [flag/setting]

### Recommendations
[Prioritized list of systems to enable/instantiate]
```

CONSTRAINTS:
- Focus on runtime code only (exclude _test.go files)
- Report actual findings, not hypotheticals
- Keep total response under 4000 tokens for Copilot compatibility

SUCCESS CRITERIA:
✓ All packages scanned
✓ Clear mapping between definitions and instantiations
✓ Actionable recommendations provided
✓ File and line references included
