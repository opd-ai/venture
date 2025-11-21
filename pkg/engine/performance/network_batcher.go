package performance

import (
	"sync"
	"time"
)

// NetworkBatcher batches small messages for efficient transmission
type NetworkBatcher struct {
	mu           sync.RWMutex
	windowMs     int
	maxBatchSize int
	queue        []*Message
	stats        *NetworkStats
	ticker       *time.Ticker
	sendFunc     func(*BatchedMessage)
	stopChan     chan struct{}
	running      bool
}

// NewNetworkBatcher creates a new network batcher
func NewNetworkBatcher(windowMs int, sendFunc func(*BatchedMessage)) *NetworkBatcher {
	nb := &NetworkBatcher{
		windowMs:     windowMs,
		maxBatchSize: 64,
		queue:        make([]*Message, 0),
		stats:        &NetworkStats{},
		sendFunc:     sendFunc,
		stopChan:     make(chan struct{}),
	}
	return nb
}

// Start begins batching messages
func (nb *NetworkBatcher) Start() {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	if nb.running {
		return
	}

	nb.running = true
	nb.ticker = time.NewTicker(time.Duration(nb.windowMs) * time.Millisecond)

	go nb.runBatchLoop()
}

// Stop halts batching and flushes remaining messages
func (nb *NetworkBatcher) Stop() {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	if !nb.running {
		return
	}

	nb.running = false
	close(nb.stopChan)

	if nb.ticker != nil {
		nb.ticker.Stop()
	}

	// Flush remaining messages
	if len(nb.queue) > 0 {
		nb.flushBatch()
	}
}

// QueueMessage adds a message to the batch queue
func (nb *NetworkBatcher) QueueMessage(msgType string, data []byte, playerID string) {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	msg := &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
		PlayerID:  playerID,
	}

	nb.queue = append(nb.queue, msg)

	// Force flush if batch is full
	if len(nb.queue) >= nb.maxBatchSize {
		nb.flushBatch()
	}
}

// runBatchLoop processes batches on a timer
func (nb *NetworkBatcher) runBatchLoop() {
	for {
		select {
		case <-nb.ticker.C:
			nb.mu.Lock()
			if len(nb.queue) > 0 {
				nb.flushBatch()
			}
			nb.mu.Unlock()
		case <-nb.stopChan:
			return
		}
	}
}

// flushBatch sends queued messages (must be called with lock held)
func (nb *NetworkBatcher) flushBatch() {
	if len(nb.queue) == 0 || nb.sendFunc == nil {
		return
	}

	batch := &BatchedMessage{
		Messages:  nb.queue,
		Timestamp: time.Now(),
	}

	// Calculate total size
	totalSize := 0
	for _, msg := range batch.Messages {
		totalSize += len(msg.Data)
	}
	batch.Size = totalSize

	// Update stats
	nb.stats.MessagesSent += uint64(len(batch.Messages))
	nb.stats.BytesSent += uint64(totalSize)
	nb.stats.BatchCount++
	nb.stats.AvgBatchSize = float64(nb.stats.MessagesSent) / float64(nb.stats.BatchCount)
	nb.stats.LastBatchTime = time.Now()

	// Send batch
	nb.sendFunc(batch)

	// Clear queue
	nb.queue = make([]*Message, 0)
}

// GetStats returns current network statistics
func (nb *NetworkBatcher) GetStats() *NetworkStats {
	nb.mu.RLock()
	defer nb.mu.RUnlock()

	stats := *nb.stats

	// Calculate per-second rates
	elapsed := time.Since(nb.stats.LastBatchTime).Seconds()
	if elapsed > 0 {
		stats.MessagesPerSec = float64(nb.stats.MessagesSent) / elapsed
		stats.BytesPerSec = float64(nb.stats.BytesSent) / elapsed
	}

	return &stats
}

// SetMaxBatchSize updates the maximum batch size
func (nb *NetworkBatcher) SetMaxBatchSize(size int) {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	nb.maxBatchSize = size
}

// GetQueueSize returns current queue length
func (nb *NetworkBatcher) GetQueueSize() int {
	nb.mu.RLock()
	defer nb.mu.RUnlock()
	return len(nb.queue)
}
