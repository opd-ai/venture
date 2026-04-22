package timerutil

import "time"

// StopAndDrain safely stops timer and drains one pending tick if already fired.
func StopAndDrain(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
