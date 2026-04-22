package webrtc

import (
	"time"

	"github.com/opd-ai/venture/pkg/network/federation/internal/timerutil"
)

// sendWithTimeout first attempts a non-blocking send and only allocates a timer
// when immediate delivery fails, then safely stops/drains the timer on return.
// It returns true when the message is sent and false when timeout elapses.
func sendWithTimeout[T any](ch chan<- T, msg T, timeout time.Duration) bool {
	select {
	case ch <- msg:
		return true
	default:
	}

	timer := time.NewTimer(timeout)
	defer timerutil.StopAndDrain(timer)

	select {
	case ch <- msg:
		return true
	case <-timer.C:
		return false
	}
}

// sendWithTimeoutOrDone sends with an immediate fast-path, then waits for
// successful send, cancellation via done, or timeout and returns the provided
// done/timeout errors for callers that need explicit error typing.
// If send and done are both ready in the blocking select, Go may choose either
// case; callers that require strict precedence should use distinct error values
// and handle either result as equivalent terminal state.
func sendWithTimeoutOrDone[T any](
	ch chan<- T,
	msg T,
	done <-chan struct{},
	timeout time.Duration,
	doneErr error,
	timeoutErr error,
) error {
	select {
	case ch <- msg:
		return nil
	case <-done:
		return doneErr
	default:
	}

	timer := time.NewTimer(timeout)
	defer timerutil.StopAndDrain(timer)

	select {
	case ch <- msg:
		return nil
	case <-done:
		return doneErr
	case <-timer.C:
		return timeoutErr
	}
}
