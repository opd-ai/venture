# TASK DESCRIPTION:
Fix bugs documented in AUDIT.md sequentially, creating individual commits with descriptive messages.

## CONTEXT:
You are assisting with an ongoing audit. Focus exclusively on unresolved bugs in AUDIT.md. Each bug requires independent validation and minimal code changes before proceeding to the next.

## INSTRUCTIONS:
For each unresolved bug in AUDIT.md:

1. **Validation Phase**
   - Verify the bug still exists in current codebase
   - Trace code execution path to confirm the issue
   - Document the specific conditions that trigger the bug

2. **Analysis Phase**
   - Identify root cause through code inspection
   - Map all code paths affected by the bug
   - Determine minimal change scope needed

3. **Implementation Phase**
   - Apply the simplest fix that resolves the issue
   - Preserve all existing functionality
   - Add defensive coding where appropriate (nil checks, bounds validation)

4. **Verification Phase**
   - Manually trace execution with fix in place
   - Confirm edge cases are handled
   - Verify no regression in related functionality

5. **Commit Phase**
   - Stage only files directly related to this bug fix
   - Create commit with format: "Fix: [bug description] (#bug-id)"
   - Include brief explanation in commit body if needed

6. **Documentation Phase**
   - Update AUDIT.md status to "Resolved" with commit hash
   - Add inline code comments explaining the fix
   - Note any newly discovered issues

7. **Confirmation Gate**
   - Present summary of:
     * Root cause identified
     * Changes implemented
     * Verification performed
   - Wait for explicit confirmation before proceeding to next bug

## ALTERNATIVE VALIDATION METHODS:
Instead of formal testing, use these verification approaches:
- **Code path analysis**: Trace execution flow to confirm fix effectiveness
- **Defensive programming**: Add guards against invalid states
- **Invariant checking**: Ensure critical conditions remain true
- **Manual verification**: Step through logic with example inputs
- **Static analysis**: Use Go's built-in tools (go vet, staticcheck)

## FORMATTING REQUIREMENTS:
- Commit messages: One line, max 72 characters, imperative mood
- Code comments: Explain why, not what
- AUDIT.md updates: Include timestamp and commit reference

## QUALITY CRITERIA:
Before requesting confirmation, ensure:
- Bug's root cause is clearly understood
- Fix addresses cause, not just symptoms
- No new bugs introduced by changes
- Code remains readable and maintainable
- AUDIT.md accurately reflects resolution status

## EXAMPLE WORKFLOW:
```
Bug: Nil pointer dereference in auth handler

1. Validation: Confirmed user=nil causes panic at line 47
2. Analysis: findUserByUsername returns nil for unknown users
3. Implementation: Added nil check before accessing user.IsActive
4. Verification: Traced flow with valid/invalid usernames
5. Commit: "Fix: Add nil check in authenticateUser (#AUTH-001)"
6. Documentation: Updated AUDIT.md, added comment explaining nil case
7. Confirmation: Ready for review
```
