# Error Handling Guide

This document describes the comprehensive error handling framework implemented in Phase 2 of the production readiness plan (PLAN.md).

## Overview

The error handling framework provides:
- **Structured error types** for common failure categories
- **Context enrichment** with arbitrary key-value pairs
- **Correlation ID support** for distributed request tracing
- **User-friendly messages** separate from technical details
- **Error wrapping** that preserves the error chain (errors.Is/As support)
- **Integration with logging** for structured log output

## Quick Start

### Creating Errors

```go
import "github.com/opd-ai/venture/pkg/errors"

// Create a network error
err := errors.Network("connection timeout")

// Create a validation error
err := errors.Validation("invalid username: must be 3-20 characters")

// Wrap an existing error
if err := conn.Read(buf); err != nil {
    return errors.NetworkWrap(err, "failed to read from connection")
}
```

### Adding Context

```go
err := errors.Validation("invalid input").
    WithContext("field", "username").
    WithContext("value", username).
    WithContext("min_length", 3).
    WithContext("max_length", 20)
```

### Correlation IDs for Distributed Tracing

```go
import (
    "context"
    "github.com/opd-ai/venture/pkg/errors"
)

// In request handler: create correlation ID
ctx := context.Background()
correlationID := errors.NewCorrelationID()
ctx = errors.WithCorrelationID(ctx, correlationID)

// Errors inherit correlation ID from context
err := errors.NewWithContext(ctx, errors.ErrorTypeNetwork, "connection failed")

// Wrap existing errors with correlation ID
if err := operation(); err != nil {
    return errors.WrapWithContext(ctx, err, errors.ErrorTypeNetwork, "operation failed")
}
```

### User-Friendly Messages

```go
err := errors.Network("TCP connection refused on 127.0.0.1:8080")
err.UserMessage = "Cannot connect to game server. Please check your network."

// For display to users
fmt.Println(err.GetUserMessage()) // "Cannot connect to game server..."

// For logs
log.Error(err.Error()) // "[Network] TCP connection refused on 127.0.0.1:8080"
```

### Integration with Logging

```go
import (
    "github.com/opd-ai/venture/pkg/errors"
    "github.com/opd-ai/venture/pkg/logging"
    "github.com/sirupsen/logrus"
)

logger := logrus.New()

err := errors.Network("connection failed").
    WithCorrelationID("req-123").
    WithContext("host", "game-server.example.com").
    WithContext("port", 8080)

// Automatic extraction of error fields for logging
logging.ErrorLogger(logger, err).Error("network operation failed")
// Output includes: error_type=Network, correlation_id=req-123, error_context={host:game-server.example.com, port:8080}

// Or use convenience function
logging.LogError(logger, err, "network operation failed")
```

## Error Types

The framework provides 13 predefined error types:

| Error Type | Retryable | Use Case |
|------------|-----------|----------|
| `ErrorTypeNetwork` | Yes | Network connectivity, timeouts, protocol errors |
| `ErrorTypeValidation` | No | Input validation failures |
| `ErrorTypeConfiguration` | No | Configuration errors |
| `ErrorTypeGeneration` | No | Procedural generation failures |
| `ErrorTypeSerialization` | No | Data encoding/decoding errors |
| `ErrorTypeFileSystem` | No | File I/O errors |
| `ErrorTypeDatabase` | Yes | Database/persistence errors |
| `ErrorTypeAuthentication` | No | Auth/authorization failures |
| `ErrorTypeRateLimit` | Yes | Rate limiting errors |
| `ErrorTypeTimeout` | Yes | Operation timeouts |
| `ErrorTypeConcurrency` | No | Concurrency/locking errors |
| `ErrorTypeResource` | Yes | Resource exhaustion (memory, CPU, etc.) |
| `ErrorTypeUnknown` | No | Unclassified errors |

### Retryability

Errors indicate whether the operation can be retried:

```go
err := errors.Network("temporary connection failure")
if err.IsRetryable() {
    // Implement retry logic with exponential backoff
    time.Sleep(backoffDuration)
    return retryOperation()
}
```

**Retryable errors:** Network, Timeout, Database, RateLimit, Resource
**Non-retryable errors:** Validation, Configuration, Generation, Serialization, Authentication

## Best Practices

### 1. Always Wrap Errors with Context

**❌ Bad:**
```go
if err != nil {
    return err // Lost context
}
```

**✅ Good:**
```go
if err != nil {
    return errors.NetworkWrap(err, "failed to connect to game server").
        WithContext("host", serverHost).
        WithContext("port", serverPort)
}
```

### 2. Choose Appropriate Error Types

**❌ Bad:**
```go
err := errors.New(errors.ErrorTypeUnknown, "connection failed")
```

**✅ Good:**
```go
err := errors.Network("connection failed")
```

### 3. Add User-Friendly Messages for Client Errors

**❌ Bad:**
```go
// Technical message shown to user
return errors.Validation("username regex validation failed: ^[a-zA-Z0-9_]{3,20}$")
```

**✅ Good:**
```go
err := errors.Validation("username must be 3-20 alphanumeric characters or underscores")
err.UserMessage = "Please choose a username between 3 and 20 characters (letters, numbers, and underscores only)"
return err
```

### 4. Include Correlation IDs in Request Handlers

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Create context with correlation ID
    ctx := r.Context()
    correlationID := errors.NewCorrelationID()
    ctx = errors.WithCorrelationID(ctx, correlationID)
    
    // Pass context through call chain
    if err := processRequest(ctx, data); err != nil {
        // Error automatically includes correlation ID
        logging.ErrorLogger(logger, err).Error("request processing failed")
        http.Error(w, err.(*errors.VentureError).GetUserMessage(), http.StatusInternalServerError)
    }
}

func processRequest(ctx context.Context, data interface{}) error {
    // Errors created from context inherit correlation ID
    if err := validate(data); err != nil {
        return errors.WrapWithContext(ctx, err, errors.ErrorTypeValidation, "validation failed")
    }
    return nil
}
```

### 5. Log Errors with Full Context

```go
// Automatic error field extraction
err := errors.Network("connection timeout").
    WithCorrelationID(correlationID).
    WithContext("host", host).
    WithContext("retry_count", retries)

logging.LogError(logger, err, "network operation failed")
// Logs: level=warn error_type=Network correlation_id=... host=... retry_count=... retryable=true
```

## Error Chain Support

The framework fully supports Go 1.13+ error wrapping:

```go
baseErr := fmt.Errorf("io error: %w", syscall.ECONNREFUSED)
networkErr := errors.NetworkWrap(baseErr, "connection failed")
timeoutErr := errors.Wrap(networkErr, errors.ErrorTypeTimeout, "operation timed out")

// Check error types
if errors.Is(timeoutErr, errors.ErrorTypeTimeout) {
    // Handle timeout
}

// Extract VentureError
if ventureErr, ok := errors.AsVentureError(timeoutErr); ok {
    log.WithFields(ventureErr.Context).Error(ventureErr.Message)
}

// Check wrapped error
if errors.Is(timeoutErr, syscall.ECONNREFUSED) {
    // Original syscall error is in the chain
}
```

## Migration Guide

### Converting Existing Code

**Before (simple errors):**
```go
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}
```

**After (structured errors):**
```go
if err != nil {
    return errors.NetworkWrap(err, "failed to connect to game server").
        WithContext("host", host).
        WithContext("port", port)
}
```

**Before (validation):**
```go
if username == "" {
    return fmt.Errorf("username is required")
}
```

**After (structured validation):**
```go
if username == "" {
    return errors.Validation("username is required").
        WithContext("field", "username")
}
```

### Gradual Adoption

The error framework is designed for gradual adoption:

1. **Start with new code**: Use structured errors in new functions and handlers
2. **Wrap at boundaries**: Convert errors at package boundaries using `errors.Wrap`
3. **Add correlation IDs**: Integrate correlation IDs in request handlers first
4. **Enhance logging**: Update log statements to use `logging.ErrorLogger`
5. **Refactor critical paths**: Update error handling in critical paths (network, validation)

You can mix structured and standard errors - the framework handles both gracefully.

## Performance Considerations

### Error Creation Overhead

Error creation has minimal overhead:
- `errors.New()`: ~100 ns/op
- `errors.Wrap()`: ~150 ns/op  
- `errors.WithContext()`: ~50 ns/op per field

### Correlation ID Generation

- UUID v4 generation: ~500 ns/op
- Reuse correlation IDs across a request lifetime
- Store in context.Context, not in every error

### Best Practices for Performance

```go
// ✅ Good: Reuse correlation ID
ctx := errors.WithCorrelationID(context.Background(), errors.NewCorrelationID())
for _, item := range items {
    if err := process(ctx, item); err != nil {
        return errors.WrapWithContext(ctx, err, errors.ErrorTypeValidation, "processing failed")
    }
}

// ❌ Bad: Generate new correlation ID per error
for _, item := range items {
    if err := process(ctx, item); err != nil {
        return errors.Wrap(err, errors.ErrorTypeValidation, "processing failed").
            WithCorrelationID(errors.NewCorrelationID()) // Don't do this
    }
}
```

## Testing

### Testing Error Types

```go
func TestOperation(t *testing.T) {
    err := Operation()
    
    // Check error type
    if !errors.Is(err, errors.ErrorTypeNetwork) {
        t.Errorf("expected network error, got %v", err)
    }
    
    // Extract and check VentureError
    ventureErr, ok := errors.AsVentureError(err)
    if !ok {
        t.Fatal("expected VentureError")
    }
    
    if !ventureErr.Retryable {
        t.Error("expected retryable error")
    }
    
    // Check context
    if ventureErr.Context["host"] != "example.com" {
        t.Errorf("expected host=example.com, got %v", ventureErr.Context["host"])
    }
}
```

### Testing with Correlation IDs

```go
func TestHandlerWithCorrelation(t *testing.T) {
    ctx := context.Background()
    correlationID := "test-correlation-123"
    ctx = errors.WithCorrelationID(ctx, correlationID)
    
    err := handler(ctx, data)
    
    ventureErr, _ := errors.AsVentureError(err)
    if ventureErr.CorrelationID != correlationID {
        t.Errorf("expected correlation ID %s, got %s", correlationID, ventureErr.CorrelationID)
    }
}
```

## Examples

See the [pkg/errors package documentation](../pkg/errors/doc.go) for comprehensive usage examples.

## Related Documentation

- [PLAN.md](../PLAN.md) - Production readiness plan (Phase 2: Error Handling)
- [pkg/errors](../pkg/errors/) - Error package implementation
- [pkg/logging](../pkg/logging/) - Logging integration
- [SECURITY.md](../SECURITY.md) - Security considerations for error messages

## Future Enhancements

Potential future improvements (not part of Phase 2):
- [ ] Error metrics collection (error rate by type)
- [ ] Alerting integration based on error types
- [ ] Error aggregation and deduplication
- [ ] Distributed tracing integration (OpenTelemetry)
- [ ] Error rate limiting to prevent log flooding
