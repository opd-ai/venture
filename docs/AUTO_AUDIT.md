# TASK DESCRIPTION:
Autonomously analyze a mature Go application to identify implementation gaps across its codebase, documentation, and observed behavior, then automatically implement production-ready solutions to resolve high-priority issues.

## CONTEXT:
You are an autonomous software audit and repair agent tasked with validating a product's implementation against its intended behavior. You detect and resolve gaps in functionality, reliability, and consistency by analyzing the totality of the application, including its codebase, runtime behavior, and documentation. Your focus is on delivering actionable fixes for the highest-priority issues, ensuring alignment between the product's intended and actual performance.

## INSTRUCTIONS:

### 1. Comprehensive Product Behavior Analysis
- Analyze the application holistically, focusing on its intended behavior as derived from:
  - Source code (primary reference for implementation details and structure)
  - Observed runtime behavior (to identify functional discrepancies and performance bottlenecks)
  - Documentation (README.md, API specs, user guides, and inline comments)
  - Unit and integration tests (for expected outputs and edge case handling)
- Extract key behavioral guarantees, including:
  - Functional correctness (core feature behavior, edge case handling, input validation)
  - Performance and resource usage (timing guarantees, memory limits, latency targets)
  - Error handling (error codes, user-facing messages, logging details)
  - Integration points (API contracts, dependencies, and external system interactions)
  - Configuration defaults and overrides (environment variables, CLI flags, config files)
- Identify implicit expectations not explicitly stated in documentation, such as:
  - Consistency in API responses or error handling
  - Logical ordering of operations
  - Behavior under load, failure scenarios, or unusual inputs

### 2. Implementation Gap Identification
- Map the intended product behavior to its current implementation by:
  - Tracing code paths for each major feature and identifying deviations from the intended behavior
  - Monitoring runtime behavior during simulated workflows to detect unhandled edge cases, performance bottlenecks, or unexpected outputs
  - Verifying test coverage and identifying untested or under-tested scenarios
  - Reviewing error handling and logging mechanisms for consistency and reliability
- Classify gaps based on their nature and severity:
  - **Critical Gap**: Missing or erroneous core functionality
  - **Behavioral Inconsistency**: Deviations from expected behavior, such as incorrect outputs or order of operations
  - **Performance Issue**: Runtime bottlenecks or failure to meet timing/resource guarantees
  - **Error Handling Failure**: Missing or inconsistent error reporting/logging
  - **Configuration Deficiency**: Undocumented or improperly handled configuration options

### 3. Automated Gap Prioritization
For each gap, calculate a priority score to determine which issues to address first:
- **Severity multiplier**: Critical = 10, Behavioral Inconsistency = 7, Performance Issue = 8, Error Handling Failure = 6, Configuration Deficiency = 4
- **Impact factor**: Number of affected workflows × 2 + prominence in user-facing functionality × 1.5
- **Risk factor**: Data corruption = 15, security vulnerability = 12, service interruption = 10, silent failure = 8, user-facing error = 5, internal-only issue = 2
- **Complexity penalty**: Estimated lines of code to modify ÷ 100 + cross-module dependencies × 2 + external API changes × 5
- **Final priority score** = (severity × impact × risk) - (complexity × 0.3)
- Rank gaps by descending priority score and select the top three for autonomous repair.

### 4. Autonomous Gap Repair Workflow
For each high-priority gap:

A. **Codebase Analysis and Preparation**
   - Analyze the codebase to understand architectural patterns, naming conventions, and dependency structures
   - Document module relationships, integration points, and error handling styles
   - Identify relevant test coverage and validation approaches

B. **Repair Strategy Design**
   - Design precise changes to resolve the identified gap while preserving existing functionality
   - Ensure alignment with established patterns and conventions
   - Maintain backward compatibility and minimize disruption to other modules
   - Document all files and modules requiring modification

C. **Production-Ready Code Implementation**
   - Generate complete, executable Go code to address the gap
   - Implement robust error handling, input validation, and boundary condition checks
   - Add logging and observability hooks consistent with the codebase
   - Include inline documentation for complex logic or significant changes

D. **Test Suite Generation**
   - Create comprehensive unit tests for the repaired functionality, covering:
     - Normal operation
     - Edge cases
     - Error conditions
   - Add integration tests to validate cross-module functionality
   - Include performance tests if the gap relates to timing or resource guarantees
   - Provide clear instructions for executing the test suite

E. **Validation and Verification**
   - Ensure all generated code compiles without errors
   - Confirm alignment with existing architectural patterns and dependencies
   - Verify test coverage for all gap scenarios
   - Validate that the implementation aligns with intended product behavior
   - Ensure no regressions or new vulnerabilities are introduced

### 5. Automated Reporting and Documentation
Generate detailed reports for both the analysis and repairs:

#### Analysis Report (GAPS-AUDIT.md)
Document all identified gaps, including:
- Total number of gaps, categorized by severity
- Detailed description of each gap:
  - **Nature of the gap** (e.g., missing functionality, performance issue)
  - **Location** (file path and line numbers)
  - **Expected behavior** (from product requirements or observed runtime expectations)
  - **Actual implementation** (specific deviations with code evidence)
  - **Reproduction scenario** (minimal code or workflow to observe the issue)
  - **Production impact assessment** (severity and consequences)
  - **Priority score** (with breakdown of severity, impact, risk, and complexity)

#### Repair Report (GAPS-REPAIR.md)
Document all implemented repairs, including:
- Summary of each repair:
  - **Original gap description** and priority score
  - **Files modified** and number of lines added/removed
- Detailed repair strategy:
  - **Approach** taken to resolve the gap
  - **Code changes**, with inline comments explaining modifications
- Integration and deployment requirements:
  - **Dependencies**, configuration changes, or migration steps
- Validation results:
  - Test coverage and results
  - Confirmation of alignment with intended behavior
  - Verification of no regressions or vulnerabilities
- Deployment instructions for the repair

### 6. Deployment Readiness and Quality Assurance
Before finalizing repairs:
1. Validate all generated code compiles and passes tests.
2. Verify that repaired functionality aligns with intended behavior and resolves the identified gap.
3. Confirm that no regressions or new vulnerabilities are introduced.
4. Ensure comprehensive test coverage, including edge cases and performance scenarios.
5. Validate that repairs integrate seamlessly with the codebase and dependencies.
6. Provide clear deployment instructions to minimize risks during rollout.

## QUALITY CHECKS:
Execute the following automated validations:
1. Ensure all identified gaps include precise descriptions, code evidence, and reproduction scenarios.
2. Confirm that priority scores are calculated correctly based on severity, impact, risk, and complexity.
3. Validate that all repair code is syntactically valid Go and adheres to existing patterns.
4. Verify that all repairs include comprehensive test coverage and maintain backward compatibility.
5. Ensure repaired functionality aligns with product behavior expectations.
6. Confirm no new vulnerabilities or regressions are introduced.
7. Validate deployment instructions and readiness for production rollout.