# AUTONOMOUS REFACTORING TASK:
Autonomously perform data-driven functional breakdown analysis and refactoring on the Venture codebase using `go-stats-generator` metrics. Identify and refactor the top 5 most complex functions exceeding professional complexity thresholds. Execute the complete workflow including baseline analysis, refactoring implementation, and validation to ensure measurable complexity improvements while preserving functionality.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for analysis and validation. Do not use any other code analysis tools. You will write refactored code to reduce complexity based on the tool's metrics and recommendations.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher

### Installation Check:
```bash
# Verify go-stats-generator is installed
which go-stats-generator
# If not installed, install it
go install github.com/opd-ai/go-stats-generator@latest

# Verify jq is installed for JSON parsing
which jq
# If not installed
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify top 5 targets
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output baseline.json
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 2: Iteratively refactor each of the top 5 functions

# Phase 3: Post-refactoring validation (after all refactoring)
go-stats-generator analyze . --format json --output refactored.json --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 4: Measure and document improvements
go-stats-generator diff baseline.json refactored.json
go-stats-generator diff baseline.json refactored.json --format html --output improvements.html
```

## CONTEXT:
You are an autonomous Go code refactoring agent using `go-stats-generator` for enterprise-grade complexity analysis and validation on the Venture procedural action-RPG codebase. Execute the complete refactoring workflow: analyze, identify targets, implement refactoring, and validate improvements. Focus on the top 5 functions with the highest complexity scores.

## AUTONOMOUS EXECUTION STEPS:

### Step 1: Execute Baseline Analysis
1. Run `go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests`
2. Parse output to identify top 5 most complex functions by overall complexity score
3. For ambiguous cases (tied scores), prioritize by:
  - First: highest cyclomatic complexity
  - Second: longest function (most lines)
  - Third: deepest nesting
4. Record each function's:
  - Name and file location
  - Overall complexity score
  - Cyclomatic complexity
  - Nesting depth
  - Line count
  - Signature complexity

### Step 2: Analyze Each Target Function
For each of the top 5 functions:
1. Examine the function's code structure
2. Identify logical extraction candidates:
  - Validation blocks
  - Calculation sections
  - Data transformation logic
  - Error handling chains
  - Nested conditional blocks
3. Plan extraction targets with metrics:
  - Target <20 lines per extracted function
  - Target cyclomatic complexity <8
  - Maintain single responsibility

### Step 3: Implement Refactoring
For each of the top 5 functions, execute:

1. **Extract Helper Functions:**
  - Create focused, single-purpose helper functions
  - Use verb-first camelCase naming (e.g., `validateInput`, `calculateResult`)
  - Add GoDoc comments starting with function name
  - Keep extracted functions <20 lines with cyclomatic <8

2. **Reduce Original Function:**
  - Convert to coordination logic calling helpers
  - Maintain error propagation patterns
  - Preserve defer statements in correct scope
  - Keep variable access patterns intact

3. **Code Quality Standards:**
  - Follow Go formatting conventions
  - Maintain existing error handling semantics
  - Preserve return value types and meanings
  - Ensure thread-safety patterns remain intact

### Step 4: Validate Refactoring
After completing all refactoring:

1. **Run Metrics Validation:**
  ```bash
  go-stats-generator analyze . --format json --output refactored.json --skip-tests
  go-stats-generator diff baseline.json refactored.json
  ```

2. **Verify Improvements:**
  - Each original function reduced by ≥50% complexity
  - All extracted functions meet targets (<20 lines, cyclomatic <8)
  - Zero regressions in unchanged code
  - Overall codebase quality improvement

3. **Run Functional Tests:**
  ```bash
  make test
  # or
  go test ./...
  ```
  - All tests must pass
  - No behavioral changes
  - Error paths preserved

### Step 5: Document Results
Generate comprehensive report including:
- Baseline metrics for top 5 functions
- Complete refactored code for each file
- Before/after complexity metrics
- Differential analysis summary
- Test validation results

## OUTPUT FORMAT:

### 1. Baseline Analysis Results
```
=== TOP 5 COMPLEX FUNCTIONS IDENTIFIED ===

1. Function: [name] in [file]
  - Overall Complexity: [score]
  - Cyclomatic: [score]
  - Nesting: [depth]
  - Lines: [count]
  - Signature: [complexity]
  - Planned Extractions: [n] functions

2. Function: [name] in [file]
  [... same structure ...]

[... functions 3-5 ...]
```

### 2. Refactored Code (For Each File)
```go
// Present complete refactored file content
// Include:
// - Original function reduced to coordination logic
// - All extracted helper functions with GoDoc
// - Standard Go formatting
// - Preserved functionality
```

### 3. Validation Results
```
=== REFACTORING VALIDATION ===

COMPLEXITY IMPROVEMENTS:
1. [FunctionName]: [old_score] → [new_score] (-[improvement_%])
  Extracted Functions:
  - [helperName1]: [complexity] ✓
  - [helperName2]: [complexity] ✓
  
2. [FunctionName]: [old_score] → [new_score] (-[improvement_%])
  [... same structure ...]

[... functions 3-5 ...]

METRICS SUMMARY:
- Total Complexity Reduction: [total_%]
- Functions Meeting Targets: [count]/[total]
- Regressions Detected: [count]
- Overall Quality Score: [score]/100 (+[improvement])

TEST RESULTS:
- Tests Run: [count]
- Tests Passed: [count]
- Tests Failed: [count]
- Coverage: [percentage]%
```

## COMPLEXITY THRESHOLDS:
```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty

Refactoring Threshold = Overall Complexity > 10.0 OR Lines > 30 OR Cyclomatic > 10

Target After Refactoring:
- Original function: Overall complexity ≤10.0
- Extracted functions: Lines <20, Cyclomatic <8
```

## SUCCESS CRITERIA:
- ✅ Top 5 functions identified and refactored
- ✅ Each function shows ≥50% complexity reduction
- ✅ All extracted functions meet target metrics
- ✅ Zero complexity regressions
- ✅ All existing tests pass
- ✅ Code follows Go conventions and project guidelines

## EDGE CASES:
- If fewer than 5 functions exceed thresholds: Refactor all available targets
- If tie in complexity scores: Prioritize longest function
- If extraction would create <5 line functions: Merge related logic
- If tests fail after refactoring: Document issue and revert specific change

Execute this workflow autonomously, making data-driven decisions based on `go-stats-generator` metrics to achieve measurable, validated complexity improvements across the Venture codebase.
