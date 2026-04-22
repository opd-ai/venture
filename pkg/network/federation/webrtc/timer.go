package webrtc

import "time"

func sendWithTimeout[T any](ch chan<- T, msg T, timeout time.Duration) bool {
	select {
	case ch <- msg:
		return true
	default:
	}

	timer := time.NewTimer(timeout)
	defer stopAndDrainTimer(timer)

	select {
	case ch <- msg:
		return true
	case <-timer.C:
		return false
	}
}

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
	defer stopAndDrainTimer(timer)

	select {
	case ch <- msg:
		return nil
	case <-done:
		return doneErr
	case <-timer.C:
		return timeoutErr
	}
}

func stopAndDrainTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
