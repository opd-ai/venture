package parallel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name        string
		workerCount int
		expected    int
	}{
		{"zero workers", 0, 1},
		{"negative workers", -5, 1},
		{"normal count", 4, 4},
		{"max count", 64, 64},
		{"excessive count", 100, 64}, // Capped at 64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workerCount)
			if pool == nil {
				t.Fatal("NewWorkerPool returned nil")
			}
			if pool.WorkerCount() != tt.expected {
				t.Errorf("Expected %d workers, got %d", tt.expected, pool.WorkerCount())
			}
			if pool.IsRunning() {
				t.Error("New pool should not be running")
			}
		})
	}
}

func TestWorkerPoolStartStop(t *testing.T) {
	pool := NewWorkerPool(4)

	// Pool should not be running initially
	if pool.IsRunning() {
		t.Error("New pool should not be running")
	}

	// Start the pool
	pool.Start()
	if !pool.IsRunning() {
		t.Error("Pool should be running after Start()")
	}

	// Calling Start() again should be safe
	pool.Start()
	if !pool.IsRunning() {
		t.Error("Pool should still be running")
	}

	// Stop the pool
	pool.Stop()
	if pool.IsRunning() {
		t.Error("Pool should not be running after Stop()")
	}

	// Calling Stop() again should be safe
	pool.Stop()
}

func TestWorkerPoolTaskProcessing(t *testing.T) {
	pool := NewWorkerPool(4)
	pool.Start()
	defer pool.Stop()

	// Submit tasks
	taskCount := 100
	submittedTasks := make(map[int]bool)
	for i := 0; i < taskCount; i++ {
		task := Task{
			ID:   i,
			Type: TaskSpriteGeneration,
			Data: i * 2,
		}
		if !pool.Submit(task) {
			t.Errorf("Failed to submit task %d", i)
		}
		submittedTasks[i] = true
	}

	// Collect results
	receivedResults := make(map[int]bool)
	timeout := time.After(5 * time.Second)
	for i := 0; i < taskCount; i++ {
		select {
		case result := <-pool.Results():
			if result.Error != nil {
				t.Errorf("Task %d returned error: %v", result.TaskID, result.Error)
			}
			receivedResults[result.TaskID] = true
		case <-timeout:
			t.Fatal("Timeout waiting for results")
		}
	}

	// Verify all tasks were processed
	if len(receivedResults) != taskCount {
		t.Errorf("Expected %d results, got %d", taskCount, len(receivedResults))
	}

	for taskID := range submittedTasks {
		if !receivedResults[taskID] {
			t.Errorf("Task %d was not processed", taskID)
		}
	}
}

func TestWorkerPoolConcurrentSubmit(t *testing.T) {
	pool := NewWorkerPool(8)
	pool.Start()
	defer pool.Stop()

	// Concurrent task submission
	var wg sync.WaitGroup
	goroutineCount := 10
	tasksPerGoroutine := 50
	totalTasks := goroutineCount * tasksPerGoroutine

	var taskCounter int32

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				taskID := int(atomic.AddInt32(&taskCounter, 1)) - 1
				task := Task{
					ID:   taskID,
					Type: TaskTileRendering,
					Data: goroutineID*100 + j,
				}
				pool.Submit(task)
			}
		}(i)
	}

	// Wait for all submissions
	wg.Wait()

	// Collect results
	receivedCount := 0
	timeout := time.After(5 * time.Second)
	for receivedCount < totalTasks {
		select {
		case <-pool.Results():
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for results. Received %d/%d", receivedCount, totalTasks)
		}
	}

	if receivedCount != totalTasks {
		t.Errorf("Expected %d results, got %d", totalTasks, receivedCount)
	}
}

func TestWorkerPoolSubmitAfterStop(t *testing.T) {
	pool := NewWorkerPool(4)
	pool.Start()
	pool.Stop()

	// Attempting to submit after stop should fail
	task := Task{
		ID:   1,
		Type: TaskParticleUpdate,
		Data: nil,
	}
	if pool.Submit(task) {
		t.Error("Submit should fail after Stop()")
	}
}

func TestWorkerPoolTrySubmit(t *testing.T) {
	pool := NewWorkerPool(2)

	// TrySubmit should fail when pool not running
	task := Task{ID: 1, Type: TaskSpriteGeneration, Data: nil}
	if pool.TrySubmit(task) {
		t.Error("TrySubmit should fail when pool is not running")
	}

	pool.Start()
	defer pool.Stop()

	// Start draining results to avoid blocking
	go func() {
		for range pool.Results() {
		}
	}()

	// TrySubmit should succeed when queue has space
	for i := 0; i < 10; i++ {
		task := Task{ID: i, Type: TaskSpriteGeneration, Data: i}
		if !pool.TrySubmit(task) {
			t.Errorf("TrySubmit should succeed for task %d when queue has space", i)
		}
	}
}

func TestWorkerPoolTrySubmitWhenFull(t *testing.T) {
	// Create pool with small buffer to test queue full scenario
	pool := NewWorkerPool(1)
	pool.Start()
	defer pool.Stop()

	// Fill the task queue without draining results
	// This will eventually fill up because the worker will block on results
	filled := 0
	for i := 0; i < 3000; i++ {
		task := Task{ID: i, Type: TaskSpriteGeneration, Data: i}
		if !pool.TrySubmit(task) {
			filled = i
			break
		}
	}

	// We should have been able to submit some tasks before queue filled
	if filled == 0 {
		t.Error("Expected some tasks to be submitted before queue filled")
	}

	// Now drain results to allow more submissions
	go func() {
		for range pool.Results() {
		}
	}()

	// Give workers time to process
	time.Sleep(10 * time.Millisecond)

	// Should be able to submit again
	task := Task{ID: 9999, Type: TaskSpriteGeneration, Data: 9999}
	if !pool.TrySubmit(task) {
		t.Error("TrySubmit should succeed after draining results")
	}
}

func TestWorkerPoolStats(t *testing.T) {
	pool := NewWorkerPool(8)

	// Initial stats
	stats := pool.GetStats()
	if stats.WorkerCount != 8 {
		t.Errorf("Expected 8 workers, got %d", stats.WorkerCount)
	}
	if stats.Running {
		t.Error("Pool should not be running initially")
	}

	// Start and check stats
	pool.Start()
	defer pool.Stop()

	stats = pool.GetStats()
	if !stats.Running {
		t.Error("Pool should be running")
	}
	if stats.WorkerCount != 8 {
		t.Errorf("Expected 8 workers, got %d", stats.WorkerCount)
	}

	// Submit tasks and check queue sizes
	for i := 0; i < 10; i++ {
		pool.Submit(Task{
			ID:   i,
			Type: TaskSpriteGeneration,
			Data: nil,
		})
	}

	// Note: Queue sizes may vary due to concurrent processing
	// Just verify stats are accessible
	stats = pool.GetStats()
	if stats.WorkerCount != 8 {
		t.Errorf("Expected 8 workers, got %d", stats.WorkerCount)
	}
}

func TestTaskTypeString(t *testing.T) {
	tests := []struct {
		taskType TaskType
		expected string
	}{
		{TaskSpriteGeneration, "SpriteGeneration"},
		{TaskTileRendering, "TileRendering"},
		{TaskParticleUpdate, "ParticleUpdate"},
		{TaskTextureUpload, "TextureUpload"},
		{TaskType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.taskType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWorkerPoolGracefulShutdown(t *testing.T) {
	pool := NewWorkerPool(4)
	pool.Start()

	// Submit many tasks
	taskCount := 1000
	for i := 0; i < taskCount; i++ {
		pool.Submit(Task{
			ID:   i,
			Type: TaskSpriteGeneration,
			Data: i,
		})
	}

	// Stop should wait for all tasks to complete
	stopDone := make(chan bool)
	go func() {
		pool.Stop()
		stopDone <- true
	}()

	// Collect all results
	receivedCount := 0
	for receivedCount < taskCount {
		<-pool.Results()
		receivedCount++
	}

	// Wait for stop to complete
	select {
	case <-stopDone:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not complete within timeout")
	}

	if pool.IsRunning() {
		t.Error("Pool should not be running after Stop()")
	}
}

// Benchmark worker pool throughput
func BenchmarkWorkerPoolThroughput(b *testing.B) {
	pool := NewWorkerPool(8)
	pool.Start()
	defer pool.Stop()

	// Start a goroutine to drain results concurrently to avoid deadlock
	// when submit fills the task buffer and workers fill the result buffer
	var received int64
	done := make(chan struct{})
	go func() {
		for range pool.Results() {
			atomic.AddInt64(&received, 1)
		}
		close(done)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(Task{
			ID:   i,
			Type: TaskSpriteGeneration,
			Data: i,
		})
	}
	b.StopTimer()

	// Wait for all results to be received
	for atomic.LoadInt64(&received) < int64(b.N) {
		time.Sleep(time.Microsecond)
	}
}

// Benchmark concurrent submissions
func BenchmarkWorkerPoolConcurrent(b *testing.B) {
	pool := NewWorkerPool(8)
	pool.Start()
	defer pool.Stop()

	// Start a goroutine to drain results concurrently to avoid deadlock
	var received int64
	done := make(chan struct{})
	go func() {
		for range pool.Results() {
			atomic.AddInt64(&received, 1)
		}
		close(done)
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		taskID := 0
		for pb.Next() {
			pool.Submit(Task{
				ID:   taskID,
				Type: TaskTileRendering,
				Data: taskID,
			})
			taskID++
		}
	})
	b.StopTimer()

	// Wait for all results to be received
	for atomic.LoadInt64(&received) < int64(b.N) {
		time.Sleep(time.Microsecond)
	}
}

// Benchmark pool creation
func BenchmarkNewWorkerPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pool := NewWorkerPool(8)
		_ = pool
	}
}

// Benchmark Start/Stop overhead
func BenchmarkWorkerPoolStartStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pool := NewWorkerPool(4)
		pool.Start()
		pool.Stop()
	}
}
