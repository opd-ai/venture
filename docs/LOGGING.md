Implement comprehensive structured logging using logrus throughout the entire Venture codebase. The implementation must follow these requirements:

## Core Requirements

1. **Package Integration**: Add logrus to all packages in `pkg/` (engine, procgen, rendering, audio, network, combat, world, saveload, mobile) and all commands in `cmd/` (client, server, and test utilities).

2. **Log Levels**: Use appropriate levels consistently:
   - **Debug**: Detailed system state, component creation, seed values, generation parameters
   - **Info**: Startup/shutdown, phase transitions, world generation, player connections/disconnections
   - **Warn**: Validation failures, retries, performance degradation, high latency
   - **Error**: Generation failures, network errors, invalid state, component errors
   - **Fatal**: Unrecoverable errors only (initialization failures, critical resource errors)

3. **Structured Fields**: Always use logrus.Fields for context:
   - Entity operations: `entityID`, `componentType`, `systemName`
   - Procgen: `seed`, `genreID`, `depth`, `difficulty`, `generatorType`
   - Network: `playerID`, `latency`, `packetSize`, `connectionState`
   - Performance: `duration`, `allocations`, `fps`, `entityCount`

4. **Logger Configuration**:
   - Client: JSON formatter for production, Text formatter for development
   - Server: Always JSON for log aggregation
   - Include timestamps, caller information (file:line)
   - Support environment variable `LOG_LEVEL` for runtime control
   - Test utilities: Text formatter with color

5. **Performance Considerations**:
   - No logging in hot paths (game loop, rendering, input handling) above Info level
   - Use conditional debug logging: `if log.GetLevel() >= log.DebugLevel`
   - Lazy evaluation for expensive field computation
   - Pool logrus.Fields objects if allocations become problematic

6. **Context Propagation**: Pass logger instances with contextual fields:
   ```go
   logger := log.WithFields(log.Fields{"system": "terrain", "seed": seed})
   ```

7. **Error Integration**: Wrap errors with context before logging:
   ```go
   logger.WithError(err).WithFields(log.Fields{...}).Error("operation failed")
   ```

8. **Special Cases**:
   - Network package: Log packet types, sizes, timing
   - Procgen: Log generation success/failure with validation results
   - Engine systems: Log entity lifecycle, component additions/removals
   - Combat: Log damage calculations, death events

Maintain deterministic behavior—logging should never affect game state or generation. Preserve all existing functionality while adding observability for debugging, profiling, and production monitoring.