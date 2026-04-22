package timerutil

import "time"

// StopAndDrain safely stops a timer and performs a non-blocking drain of the
// timer channel if it has already fired, preventing timer event leaks in
// cleanup paths.
func StopAndDrain(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
