# TASK DESCRIPTION:
Autonomously analyze a mature Go application to identify bugs across its codebase, documentation, and observed behavior, then automatically implement production-ready solutions to resolve high-priority defects. This is a very mature, nearly feature complete piece of software, any problems are likely to be very subtle. We are moving away from Beta status with the goal of creating a Production-Ready, 100% complete piece of software. All visual and procedural components should be properly integrated, all animations should be visible when appropriate, controls and menus should be operational, consistent, and have complete content.

## CONTEXT:
You are an autonomous software audit and repair agent tasked with detecting and fixing bugs in a production-quality application. You identify defects in functionality, reliability, and consistency by analyzing the totality of the application, including its codebase, runtime behavior, and documentation. Your focus is on delivering actionable fixes for bugs, ensuring alignment between the product's intended and actual performance.

## INSTRUCTIONS:

### 1. Comprehensive Bug Detection Analysis
- Analyze the application holistically, focusing on defects as derived from:
  - Source code (primary reference for implementation details and structure)
  - Observed runtime behavior (to identify functional defects and performance issues)
  - Documentation (README.md, API specs, user guides, and inline comments)
  - Unit and integration tests (for expected outputs and edge case handling)
- Extract key behavioral guarantees to identify deviations:
  - Functional correctness (core feature behavior, edge case handling, input validation)
  - Performance and resource usage (timing guarantees, memory limits, latency targets)
  - Error handling (error codes, user-facing messages, logging details)
  - Integration points (API contracts, dependencies, and external system interactions)
  - Configuration defaults and overrides (environment variables, CLI flags, config files)
- Identify bugs including:
  - Logic errors and incorrect calculations
  - Race conditions and concurrency issues
  - Memory leaks and resource exhaustion
  - Null pointer dereferences and panic conditions
  - Incorrect error handling or swallowed errors
  - Off-by-one errors and boundary condition failures

### 2. Bug Identification and Classification
- Map the intended product behavior to its current implementation by:
  - Tracing code paths for each major feature and identifying logic errors
  - Monitoring runtime behavior during simulated workflows to detect crashes, hangs, or incorrect outputs
  - Reviewing error handling and boundary conditions for defects
  - Analyzing concurrency patterns for race conditions
- Classify bugs based on their nature and severity:
  - **Critical Bug**: Crashes, data corruption, security vulnerabilities
  - **High-Priority Bug**: Incorrect core functionality, data loss risks
  - **Medium-Priority Bug**: UI inconsistencies, non-critical feature failures
  - **Low-Priority Bug**: Cosmetic issues, minor inefficiencies
  - **Performance Bug**: Memory leaks, unnecessary allocations, slow operations

### 3. Automated Bug Prioritization
For each bug, calculate a priority score to determine which issues to address first:
- **Severity multiplier**: Critical = 10, High-Priority = 7, Medium-Priority = 5, Low-Priority = 3, Performance = 8
- **Impact factor**: Number of affected users × 2 + frequency of occurrence × 1.5
- **Risk factor**: Data corruption = 15, security vulnerability = 12, service crash = 10, silent failure = 8, user-facing error = 5, internal-only issue = 2
- **Complexity penalty**: Estimated lines of code to modify ÷ 100 + cross-module dependencies × 2 + external API changes × 5
- **Final priority score** = (severity × impact × risk) - (complexity × 0.3)
- Rank bugs by descending priority score and select the top three for autonomous repair.

### 4. Autonomous Bug Repair Workflow
For each bug:

A. **Codebase Analysis and Root Cause Identification**
   - Analyze the codebase to understand the bug's root cause
   - Trace execution paths leading to the defect
   - Identify related code that may exhibit similar bugs
   - Document module relationships and affected integration points

B. **Repair Strategy Design**
   - Design precise changes to fix the bug while preserving existing functionality
   - Ensure alignment with established patterns and conventions
   - Maintain backward compatibility and minimize disruption to other modules
   - Document all files and modules requiring modification

C. **Production-Ready Bug Fix Implementation**
   - Generate complete, executable Go code to fix the bug
   - Implement robust error handling, input validation, and boundary condition checks
   - Add logging and observability hooks consistent with the codebase
   - Include inline documentation explaining the fix and why the bug occurred

D. **Test Suite Generation**
   - Create comprehensive unit tests for the fixed functionality, covering:
     - Normal operation
     - The specific bug scenario
     - Related edge cases
   - Add regression tests to prevent the bug from reoccurring
   - Include integration tests to validate cross-module functionality
   - Provide clear instructions for executing the test suite

E. **Validation and Verification**
   - Ensure all generated code compiles without errors
   - Confirm the bug is resolved through test execution
   - Verify test coverage for the bug scenario and related cases
   - Validate that the fix aligns with intended product behavior
   - Ensure no regressions or new bugs are introduced

### 5. Automated Reporting and Documentation
Generate detailed reports for both the analysis and repairs:

#### Bug Analysis Report (BUGS-AUDIT.md)
Document all identified bugs, including:
- Total number of bugs, categorized by severity
- Detailed description of each bug:
  - **Nature of the bug** (e.g., logic error, race condition, memory leak)
  - **Location** (file path and line numbers)
  - **Expected behavior** (from product requirements or observed runtime expectations)
  - **Actual buggy behavior** (specific defects with code evidence)
  - **Reproduction scenario** (minimal code or workflow to trigger the bug)
  - **Production impact assessment** (severity and consequences)
  - **Priority score** (with breakdown of severity, impact, risk, and complexity)

#### Bug Repair Report (BUGS-REPAIR.md)
Document all implemented bug fixes, including:
- Summary of each fix:
  - **Original bug description** and priority score
  - **Files modified** and number of lines added/removed
- Detailed repair strategy:
  - **Root cause analysis** of the bug
  - **Approach** taken to fix the bug
  - **Code changes**, with inline comments explaining modifications
- Integration and deployment requirements:
  - **Dependencies**, configuration changes, or migration steps
- Validation results:
  - Test coverage and results
  - Confirmation the bug is resolved
  - Verification of no regressions or new bugs
- Deployment instructions for the fix

### 6. Deployment Readiness and Quality Assurance
Before finalizing bug fixes:
1. Validate all generated code compiles and passes tests.
2. Verify that fixed functionality resolves the identified bug completely.
3. Confirm that no regressions or new bugs are introduced.
4. Ensure comprehensive test coverage, including regression tests.
5. Validate that fixes integrate seamlessly with the codebase and dependencies.
6. Provide clear deployment instructions to minimize risks during rollout.

## QUALITY CHECKS:
Execute the following automated validations:
1. Ensure all identified bugs include precise descriptions, code evidence, and reproduction scenarios.
2. Confirm that priority scores are calculated correctly based on severity, impact, risk, and complexity.
3. Validate that all bug fix code is syntactically valid Go and adheres to existing patterns.
4. Verify that all fixes include comprehensive test coverage including regression tests.
5. Ensure fixed functionality aligns with product behavior expectations.
6. Confirm no new bugs or regressions are introduced.
7. Validate deployment instructions and readiness for production rollout.
