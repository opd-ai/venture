package federation

import (
"errors"
"fmt"
"testing"
"time"
)

func TestCircuitBreakerDebug(t *testing.T) {
config := CircuitBreakerConfig{
FailureThreshold:    2,
SuccessThreshold:    5,
OpenTimeout:         100 * time.Millisecond,
HalfOpenMaxRequests: 2,
}
cb := NewCircuitBreaker(config)

// Trigger circuit to open
cb.Call(func() error { return errors.New("failure 1") })
cb.Call(func() error { return errors.New("failure 2") })
fmt.Printf("After failures, state: %v\n", cb.State())

// Wait for transition to half-open
time.Sleep(150 * time.Millisecond)
fmt.Printf("After wait, calling beforeRequest...\n")

// Manually check beforeRequest
allowed1 := cb.beforeRequest()
fmt.Printf("First beforeRequest: allowed=%v, state=%v\n", allowed1, cb.State())
cb.mu.RLock()
fmt.Printf("halfOpenRequests=%d\n", cb.halfOpenRequests)
cb.mu.RUnlock()

allowed2 := cb.beforeRequest()
fmt.Printf("Second beforeRequest: allowed=%v, state=%v\n", allowed2, cb.State())
cb.mu.RLock()
fmt.Printf("halfOpenRequests=%d\n", cb.halfOpenRequests)
cb.mu.RUnlock()

allowed3 := cb.beforeRequest()
fmt.Printf("Third beforeRequest: allowed=%v, state=%v\n", allowed3, cb.State())
cb.mu.RLock()
fmt.Printf("halfOpenRequests=%d\n", cb.halfOpenRequests)
cb.mu.RUnlock()

if allowed3 {
t.Error("Expected third request to be blocked")
}
}
