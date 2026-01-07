# Error Handling Package

Comprehensive error handling framework for the Venture game system with structured error types, context enrichment, and correlation ID support for distributed tracing.

## Features

- **Structured Error Types**: 13 predefined error categories (Network, Validation, Configuration, etc.)
- **Error Wrapping**: Preserves error chains for `errors.Is` and `errors.As` compatibility
- **Context Enrichment**: Add arbitrary key-value pairs to errors
- **Correlation IDs**: UUID-based request tracking for distributed tracing
- **User-Friendly Messages**: Separate technical and user-facing error messages
- **Retryability Indicators**: Errors indicate if operations can be retried
- **Logging Integration**: Seamless integration with pkg/logging

## Quick Start

```go
import "github.com/opd-ai/venture/pkg/errors"

// Create a network error
err := errors.Network("connection timeout")

// Wrap an existing error with context
if err := conn.Read(buf); err != nil {
    return errors.NetworkWrap(err, "failed to read from connection").
        WithContext("host", host).
        WithContext("port", port)
}

// Use correlation IDs for request tracing
ctx := errors.WithCorrelationID(context.Background(), errors.NewCorrelationID())
err := errors.NewWithContext(ctx, errors.ErrorTypeNetwork, "connection failed")
```

## Error Types

| Type | Retryable | Use Case |
|------|-----------|----------|
| `ErrorTypeNetwork` | ✅ | Network connectivity, timeouts, protocol errors |
| `ErrorTypeValidation` | ❌ | Input validation failures |
| `ErrorTypeConfiguration` | ❌ | Configuration errors |
| `ErrorTypeGeneration` | ❌ | Procedural generation failures |
| `ErrorTypeSerialization` | ❌ | Data encoding/decoding errors |
| `ErrorTypeFileSystem` | ❌ | File I/O errors |
| `ErrorTypeDatabase` | ✅ | Database/persistence errors |
| `ErrorTypeAuthentication` | ❌ | Auth/authorization failures |
| `ErrorTypeRateLimit` | ✅ | Rate limiting errors |
| `ErrorTypeTimeout` | ✅ | Operation timeouts |
| `ErrorTypeConcurrency` | ❌ | Concurrency/locking errors |
| `ErrorTypeResource` | ✅ | Resource exhaustion (memory, CPU) |
| `ErrorTypeUnknown` | ❌ | Unclassified errors |

## Testing

Run tests with coverage:

```bash
go test -v -cover ./pkg/errors/...
```

Current coverage: **93.7%**

## Documentation

- [Error Handling Guide](../../docs/ERROR_HANDLING.md) - Comprehensive usage guide
- [Package Documentation](doc.go) - GoDoc reference
- [Example Demo](../../examples/error_handling_demo.go) - Working examples

## Integration

### With Logging

```go
import (
    "github.com/opd-ai/venture/pkg/errors"
    "github.com/opd-ai/venture/pkg/logging"
)

err := errors.Network("connection failed").
    WithCorrelationID(correlationID).
    WithContext("host", "example.com")

logging.LogError(logger, err, "operation failed")
// Automatically logs: error_type, correlation_id, error_context={host:...}, retryable
```

### With Context

```go
// Create context with correlation ID
ctx := errors.WithCorrelationID(context.Background(), errors.NewCorrelationID())

// Errors inherit correlation ID
err := errors.NewWithContext(ctx, errors.ErrorTypeValidation, "invalid input")

// Wrap errors with correlation ID from context
if err := operation(); err != nil {
    return errors.WrapWithContext(ctx, err, errors.ErrorTypeNetwork, "operation failed")
}
```

## Best Practices

1. **Always wrap errors with context**
   ```go
   return errors.NetworkWrap(err, "connection failed").
       WithContext("host", host).
       WithContext("port", port)
   ```

2. **Use specific error types**
   ```go
   // ✅ Good
   return errors.Network("connection timeout")
   
   // ❌ Bad
   return errors.New(errors.ErrorTypeUnknown, "something failed")
   ```

3. **Provide user-friendly messages**
   ```go
   err := errors.Network("TCP connection refused")
   err.UserMessage = "Cannot connect to server. Check your network."
   ```

4. **Include correlation IDs in request handlers**
   ```go
   ctx := errors.WithCorrelationID(r.Context(), errors.NewCorrelationID())
   // Pass ctx through call chain
   ```

5. **Check retryability before retrying**
   ```go
   if err, ok := errors.AsVentureError(err); ok && err.IsRetryable() {
       // Implement retry with backoff
   }
   ```

## Migration

The framework is designed for gradual adoption:

1. Start using in new code
2. Wrap errors at package boundaries
3. Add correlation IDs to request handlers
4. Update logging to use `logging.ErrorLogger`
5. Refactor critical paths incrementally

You can mix structured and standard errors - the framework handles both gracefully.

## Performance

- Error creation: ~100 ns/op
- Error wrapping: ~150 ns/op
- Context addition: ~50 ns/op per field
- UUID generation: ~500 ns/op

Minimal overhead suitable for production use.

## See Also

- [PLAN.md Phase 2](../../PLAN.md#phase-2-stability--error-handling) - Production readiness requirements
- [pkg/logging](../logging/) - Logging integration
- [pkg/recovery](../recovery/) - Panic recovery (Phase 1)
