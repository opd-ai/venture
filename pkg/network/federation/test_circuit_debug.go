package federation

import (
"errors"
"fmt"
"testing"
"time"
)

func TestCircuitBreakerHalfOpenMaxRequestsDebug(t *testing.T) {
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
fmt.Printf("After wait, state: %v\n", cb.State())

// Allow max requests
err1 := cb.Call(func() error { return nil })
fmt.Printf("After call 1: state=%v, err=%v, halfOpenRequests=%d\n", cb.State(), err1, cb.halfOpenRequests)

err2 := cb.Call(func() error { return nil })
fmt.Printf("After call 2: state=%v, err=%v, halfOpenRequests=%d\n", cb.State(), err2, cb.halfOpenRequests)

// Third request should be blocked (still in half-open state)
err3 := cb.Call(func() error { return nil })
fmt.Printf("After call 3: state=%v, err=%v, halfOpenRequests=%d\n", cb.State(), err3, cb.halfOpenRequests)

if err3 == nil {
t.Error("Expected call to be blocked after max half-open requests")
}
}
