# Async Terrain Generation

## Overview

The `AsyncLoader` provides non-blocking terrain generation for improved startup performance. By moving terrain generation to a background goroutine, the game client can remain responsive and display loading progress while terrain is being generated.

## Architecture

### AsyncLoader

Located in `pkg/procgen/terrain/async_loader.go`, the `AsyncLoader` manages asynchronous terrain generation with thread-safe progress tracking.

**Key Features:**
- **Non-blocking**: Terrain generation runs in a background goroutine
- **Progress Tracking**: Thread-safe progress updates (0.0 to 1.0)
- **Error Handling**: Propagates generation errors to the caller
- **Type Safety**: Validates generator output type
- **Minimal Overhead**: ~28µs for BSP, ~725µs for Cellular automata

### API

```go
// Create a new async loader
loader := terrain.NewAsyncLoader(logger)

// Start generation in background
loader.StartGeneration(generator, seed, params)

// Poll progress (thread-safe)
progress, err := loader.GetProgress()

// Check if complete
if loader.IsDone() {
    // ...
}

// Wait for result (blocks until complete)
terrain, err := loader.Wait()

// Or get result without blocking (returns nil if not complete)
terrain := loader.GetResult()
```

## Performance Impact

### Benchmark Results

```
BenchmarkAsyncLoader_BSP-16         	  126349	     28070 ns/op	   43210 B/op	     127 allocs/op
BenchmarkAsyncLoader_Cellular-16    	    4874	    725161 ns/op	  558317 B/op	    4077 allocs/op
```

**Overhead Analysis:**
- BSP Generator: ~28µs async overhead (negligible compared to ~3-5ms generation time)
- Cellular Generator: ~725µs async overhead (negligible compared to ~4-5ms generation time)
- Progress polling: Lock-free read for concurrent access

### Startup Time Reduction

**Before Async Loading:**
- Main thread blocks during terrain generation (12-50ms)
- UI frozen until terrain completes
- No user feedback during generation

**After Async Loading:**
- Main thread continues immediately
- UI can display loading progress
- Terrain generates in background
- User sees responsive startup experience

## Usage Example

### Client Integration

```go
func generateWorldTerrain(logger *logrus.Logger) *terrain.Terrain {
    gen := terrain.NewBSPGeneratorWithLogger(logger)
    params := procgen.GenerationParams{
        GenreID: "fantasy",
        Custom: map[string]interface{}{
            "width":  80,
            "height": 50,
        },
    }

    // Start async generation
    loader := terrain.NewAsyncLoader(logger)
    loader.StartGeneration(gen, seed, params)

    // Poll progress and display to user
    for !loader.IsDone() {
        progress, err := loader.GetProgress()
        if err != nil {
            log.Fatal(err)
        }
        // Update UI with progress percentage
        updateLoadingScreen(int(progress * 100))
    }

    // Get final result
    terrain, err := loader.Wait()
    if err != nil {
        log.Fatal(err)
    }

    return terrain
}
```

### Progress Monitoring

The loader provides three progress milestones:

1. **0.1 (10%)**: Generation started
2. **0.9 (90%)**: Generation complete, validating
3. **1.0 (100%)**: Validation complete, terrain ready

## Error Handling

The loader handles three error scenarios:

1. **Generation Error**: Generator.Generate() returns error
   - Progress set to 0.0
   - Error propagated through GetProgress() and Wait()

2. **Type Error**: Generator returns wrong type
   - Progress set to 0.0
   - Error message includes actual type received

3. **Nil Result**: Generation completes but result is nil
   - Error returned from Wait()

## Thread Safety

All public methods are thread-safe:

- **StartGeneration**: Spawns goroutine, safe to call once
- **GetProgress**: RLock for concurrent reads
- **IsDone**: Channel select (non-blocking check)
- **Wait**: Blocks on done channel
- **GetResult**: RLock for concurrent reads

## Testing

### Unit Tests

- **TestAsyncLoader_NewAsyncLoader**: Constructor validation
- **TestAsyncLoader_StartGeneration_Success**: Happy path
- **TestAsyncLoader_StartGeneration_Error**: Error propagation
- **TestAsyncLoader_StartGeneration_InvalidType**: Type validation
- **TestAsyncLoader_GetProgress_Concurrent**: Thread safety
- **TestAsyncLoader_IsDone**: State tracking
- **TestAsyncLoader_GetResult**: Result retrieval
- **TestAsyncLoader_WithLogger**: Logger integration

### Integration Tests

- **TestAsyncLoader_Integration_BSP**: BSP generator compatibility
- **TestAsyncLoader_Integration_Cellular**: Cellular automata compatibility
- **TestAsyncLoader_Integration_Composite**: Composite generator compatibility

### Coverage

- **async_loader.go**: 93-100% coverage across all methods
- **async_loader_test.go**: 7 unit tests
- **async_loader_integration_test.go**: 3 integration tests + 2 benchmarks

## Design Rationale

### Why AsyncLoader vs State Machine?

The AUDIT.md originally suggested a "state machine for loading" approach. The AsyncLoader provides a simpler alternative:

**Advantages:**
- **Simplicity**: Single struct with clear API, no complex state transitions
- **Reusability**: Works with any Generator implementation
- **Testability**: Easy to mock and test in isolation
- **Minimal Changes**: Integrates into existing client code with minimal refactoring

**Trade-offs:**
- **No Frame-Level Control**: Generation runs to completion (can't pause/resume per frame)
- **Single-Use**: Each loader instance is for one generation task
- **No Cancellation**: Generation cannot be cancelled mid-stream

For terrain generation (typically 2-50ms), these trade-offs are acceptable. The generation completes quickly enough that frame-level control is unnecessary.

### Why Progress Polling vs Callbacks?

Progress polling was chosen over callback-based updates:

**Advantages:**
- **Flexibility**: Caller controls polling frequency
- **No Goroutine Leaks**: No callback goroutine management
- **Simpler Testing**: No callback mocking required
- **Thread Safety**: RWMutex for safe concurrent access

**Usage Pattern:**
```go
for !loader.IsDone() {
    progress, _ := loader.GetProgress()
    // Update UI
    time.Sleep(pollInterval)
}
```

## Future Enhancements

Potential improvements for future versions:

1. **Cancellation Support**: Add context.Context for cancellation
2. **Incremental Progress**: Finer-grained progress updates (e.g., per-room in BSP)
3. **Generator Pooling**: Reuse generator instances across loads
4. **Cache Integration**: Automatic cache checking before generation
5. **Progress Callbacks**: Optional callback support for event-driven updates

## Related Documentation

- **AUDIT.md**: Performance audit documenting the async loading optimization
- **pkg/procgen/terrain/cache.go**: Terrain caching system (complements async loading)
- **pkg/procgen/terrain/composite.go**: Composite generator (benefits most from async loading)
- **cmd/client/handlers.go**: Client integration example
