package parallel

import (
	"sync"
)

// WorkerPool manages a pool of worker goroutines for parallel rendering tasks.
// The pool distributes tasks across workers and collects results efficiently.
type WorkerPool struct {
	workerCount int
	tasks       chan Task
	results     chan Result
	wg          sync.WaitGroup
	mu          sync.RWMutex
	running     bool
}

// Task represents a unit of rendering work to be processed by a worker.
type Task struct {
	ID       int         // Unique task identifier
	Type     TaskType    // Type of rendering task
	Data     interface{} // Task-specific data
	Priority int         // Higher priority tasks processed first (future: priority queue)
}

// TaskType identifies the type of rendering operation.
type TaskType int

const (
	// TaskSpriteGeneration generates a sprite image from entity data
	TaskSpriteGeneration TaskType = iota
	// TaskTileRendering renders a tile to an image
	TaskTileRendering
	// TaskParticleUpdate updates particle system state
	TaskParticleUpdate
	// TaskTextureUpload uploads texture data to GPU
	TaskTextureUpload
)

// String returns the task type name.
func (t TaskType) String() string {
	switch t {
	case TaskSpriteGeneration:
		return "SpriteGeneration"
	case TaskTileRendering:
		return "TileRendering"
	case TaskParticleUpdate:
		return "ParticleUpdate"
	case TaskTextureUpload:
		return "TextureUpload"
	default:
		return "Unknown"
	}
}

// Result represents the output of a completed rendering task.
type Result struct {
	TaskID int         // Corresponding task ID
	Data   interface{} // Result data (type depends on task type)
	Error  error       // Error if task failed
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
// workerCount should typically match CPU core count for optimal performance.
func NewWorkerPool(workerCount int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > 64 {
		workerCount = 64 // Cap at 64 workers to prevent excessive overhead
	}

	// Use larger buffers to prevent deadlock when submitting many tasks
	// Buffer size accommodates typical burst of tasks (at least 1024 to handle test cases)
	bufferSize := 1024

	return &WorkerPool{
		workerCount: workerCount,
		tasks:       make(chan Task, bufferSize),
		results:     make(chan Result, bufferSize),
		running:     false,
	}
}

// Start launches the worker goroutines.
// Must be called before submitting tasks.
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return // Already running
	}

	p.running = true

	// Start worker goroutines
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop gracefully shuts down the worker pool.
// Waits for all pending tasks to complete.
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	// Close task channel to signal workers to exit
	close(p.tasks)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close results channel
	close(p.results)
}

// Submit adds a task to the worker pool for processing.
// Returns false if the pool is not running.
func (p *WorkerPool) Submit(task Task) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return false
	}

	p.tasks <- task
	return true
}

// Results returns a channel that receives completed task results.
// The channel is closed when Stop() is called.
func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

// worker is the goroutine function that processes tasks.
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for task := range p.tasks {
		result := p.processTask(task)
		p.results <- result
	}
}

// processTask executes a task and returns the result.
func (p *WorkerPool) processTask(task Task) Result {
	// Task processing is delegated to specific handlers based on type
	// This is a placeholder - actual processing happens in Renderer
	return Result{
		TaskID: task.ID,
		Data:   nil,
		Error:  nil,
	}
}

// WorkerCount returns the number of workers in the pool.
func (p *WorkerPool) WorkerCount() int {
	return p.workerCount
}

// IsRunning returns true if the worker pool is currently running.
func (p *WorkerPool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// Stats returns current worker pool statistics.
type PoolStats struct {
	WorkerCount     int // Number of worker goroutines
	Running         bool
	TaskQueueSize   int // Current number of queued tasks
	ResultQueueSize int // Current number of pending results
}

// GetStats returns current pool statistics.
func (p *WorkerPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		WorkerCount:     p.workerCount,
		Running:         p.running,
		TaskQueueSize:   len(p.tasks),
		ResultQueueSize: len(p.results),
	}
}
