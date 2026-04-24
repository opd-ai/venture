// Package terrain provides async terrain generation with progress tracking.
package terrain

import (
	"fmt"
	"sync"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// AsyncLoader handles terrain generation in a background goroutine with progress tracking.
type AsyncLoader struct {
	mu       sync.RWMutex
	progress float64 // 0.0 to 1.0
	err      error
	result   *Terrain
	done     chan struct{}
	logger   *logrus.Entry
}

// NewAsyncLoader creates a new async terrain loader.
func NewAsyncLoader(logger *logrus.Logger) *AsyncLoader {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("component", "async_loader")
	}
	return &AsyncLoader{
		done:   make(chan struct{}),
		logger: logEntry,
	}
}

// StartGeneration begins terrain generation in a background goroutine.
// Returns immediately, allowing the caller to poll progress or wait for completion.
func (l *AsyncLoader) StartGeneration(generator procgen.Generator, seed int64, params procgen.GenerationParams) {
	go func() {
		defer close(l.done)

		if l.logger != nil {
			l.logger.WithFields(logrus.Fields{
				"seed":       seed,
				"genreID":    params.GenreID,
				"difficulty": params.Difficulty,
			}).Info("starting async terrain generation")
		}

		// Set progress to 10% when generation starts
		l.setProgress(0.1)

		// Generate terrain (this is the slow part: 12-50ms for composite)
		result, err := generator.Generate(seed, params)
		if err != nil {
			func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				l.err = err
				l.progress = 0.0
			}()
			if l.logger != nil {
				l.logger.WithError(err).Error("async terrain generation failed")
			}
			return
		}

		// Validate result type
		terrain, ok := result.(*Terrain)
		if !ok {
			func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				l.err = fmt.Errorf("generator returned invalid type: expected *Terrain, got %T", result)
				l.progress = 0.0
			}()
			if l.logger != nil {
				l.logger.WithField("type", fmt.Sprintf("%T", result)).Error("invalid terrain type")
			}
			return
		}

		// Set progress to 90% after generation completes
		l.setProgress(0.9)

		// Store result and mark complete
		func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			l.result = terrain
			l.progress = 1.0
		}()

		if l.logger != nil {
			l.logger.WithFields(logrus.Fields{
				"width":     terrain.Width,
				"height":    terrain.Height,
				"roomCount": len(terrain.Rooms),
			}).Info("async terrain generation complete")
		}
	}()
}

// GetProgress returns the current progress (0.0 to 1.0) and any error.
// Thread-safe for concurrent access.
func (l *AsyncLoader) GetProgress() (float64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.progress, l.err
}

// IsDone returns true if generation is complete (success or failure).
func (l *AsyncLoader) IsDone() bool {
	select {
	case <-l.done:
		return true
	default:
		return false
	}
}

// Wait blocks until generation completes, then returns the result or error.
func (l *AsyncLoader) Wait() (*Terrain, error) {
	<-l.done

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.err != nil {
		return nil, l.err
	}

	if l.result == nil {
		return nil, fmt.Errorf("terrain generation completed but result is nil")
	}

	return l.result, nil
}

// GetResult returns the current terrain result without blocking.
// Returns nil if generation is not yet complete.
func (l *AsyncLoader) GetResult() *Terrain {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.result
}

// setProgress updates the progress value (thread-safe helper).
func (l *AsyncLoader) setProgress(p float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.progress = p
}
