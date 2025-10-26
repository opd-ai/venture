Implement structured logging with logrus across the Venture codebase.

**EXECUTION MODE**: Autonomous action - Implement directly without requiring approval for each change.

**OBJECTIVE**: 
Replace all logging statements with logrus-based structured logging that provides production-grade observability while maintaining deterministic behavior and performance targets (60 FPS, <500MB memory).

**SCOPE**:
- All packages: `pkg/{engine,procgen,rendering,audio,network,combat,world,saveload,mobile}`
- All commands: `cmd/{client,server,*test}`
- Exclude: `examples/`, test files (`*_test.go`)

**IMPLEMENTATION RULES**:

1. **Level Usage**:
   - Debug: Internal state, seeds, parameters (NOT in hot paths)
   - Info: Lifecycle events (startup/shutdown, connections, generation completion)
   - Warn: Non-fatal issues (validation, retries, latency >200ms)
   - Error: Failures (generation, network, I/O)
   - Fatal: Unrecoverable initialization errors only

2. **Structured Fields** (always use `logrus.Fields`):
   - Procgen: `seed`, `genreID`, `depth`, `difficulty`, `generatorType`
   - Entities: `entityID`, `componentType`, `systemName`
   - Network: `playerID`, `latency`, `packetSize`, `state`
   - Performance: `duration`, `count`, `fps` (Info level only)

3. **Logger Initialization**:
   - Client/Test utilities: `logging.TestUtilityLogger("name")` (text format, color)
   - Server: `logging.NewLogger(logging.Config{Format: logging.JSONFormat})` (JSON)
   - Pass logger instances via constructors/methods (no globals)

4. **Performance Safeguards**:
   - NO logging in: game loop updates, rendering per-frame, input handling per-event
   - Wrap Debug calls: `if logger.GetLevel() >= logrus.DebugLevel { ... }`
   - Summary logging only (e.g., "processed 50 entities" not per-entity logs)

5. **Integration Pattern**:
   ```go
   import "github.com/opd-ai/venture/pkg/logging"
   
   type System struct {
       logger *logrus.Entry
   }
   
   func NewSystem(logger *logrus.Logger) *System {
       return &System{logger: logging.SystemLogger(logger, "systemName")}
   }
   ```

6. **Error Handling**:
   ```go
   if err != nil {
       logger.WithError(err).WithFields(logrus.Fields{
           "operation": "action",
           "context": value,
       }).Error("operation failed")
       return fmt.Errorf("action: %w", err)
   }
   ```

**OUTPUT FORMAT**:
- Modified files with logrus imports and structured logging calls
- Brief summary: file count, level distribution, performance impact notes
- No separate documentation file needed (STRUCTURED_LOGGING_GUIDE.md already exists)

**SUCCESS CRITERIA**:
- Zero `log.Printf`/`fmt.Println` in non-test code
- All existing tests pass unchanged
- No behavior changes (determinism preserved)
- Performance targets maintained (verify no hot-path logging)

**CONSTRAINTS**:
- Do NOT modify test files (`*_test.go`)
- Do NOT modify examples/
- Preserve all function signatures
- Maintain existing error handling patterns