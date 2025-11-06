// Package network provides priority queue implementation for state updates.
package network

import (
	"container/heap"
	"sync"
)

// Priority constants for state updates.
// Higher values indicate higher priority and will be sent first.
const (
	// PriorityCritical is used for critical events (deaths, revivals)
	PriorityCritical uint8 = 255

	// PriorityHigh is used for important events (combat, damage)
	PriorityHigh uint8 = 200

	// PriorityNormal is the default priority for regular updates
	PriorityNormal uint8 = 128

	// PriorityLow is used for cosmetic updates (animations, particles)
	PriorityLow uint8 = 64
)

// priorityItem wraps a StateUpdate with metadata for the priority queue.
type priorityItem struct {
	update   *StateUpdate
	priority uint8
	index    int // Index in the heap (maintained by heap.Interface)
}

// priorityHeap implements heap.Interface for state updates.
// It's a max-heap based on priority (higher priority = popped first).
type priorityHeap []*priorityItem

func (h priorityHeap) Len() int { return len(h) }

// Less returns true if i should be ordered before j.
// We want higher priorities first, so reverse the comparison.
func (h priorityHeap) Less(i, j int) bool {
	return h[i].priority > h[j].priority
}

func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *priorityHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*priorityItem)
	item.index = n
	*h = append(*h, item)
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // Avoid memory leak
	item.index = -1 // For safety
	*h = old[0 : n-1]
	return item
}

// StateUpdatePriorityQueue is a thread-safe priority queue for state updates.
// Higher priority updates are dequeued first.
type StateUpdatePriorityQueue struct {
	heap priorityHeap
	mu   sync.RWMutex
	cap  int
}

// NewStateUpdatePriorityQueue creates a new priority queue with the given capacity.
func NewStateUpdatePriorityQueue(capacity int) *StateUpdatePriorityQueue {
	pq := &StateUpdatePriorityQueue{
		heap: make(priorityHeap, 0, capacity),
		cap:  capacity,
	}
	heap.Init(&pq.heap)
	return pq
}

// Push adds a state update to the queue with its priority.
// Returns false if the queue is full.
func (pq *StateUpdatePriorityQueue) Push(update *StateUpdate) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.heap) >= pq.cap {
		return false
	}

	item := &priorityItem{
		update:   update,
		priority: update.Priority,
	}
	heap.Push(&pq.heap, item)
	return true
}

// Pop removes and returns the highest priority state update.
// Returns nil if the queue is empty.
func (pq *StateUpdatePriorityQueue) Pop() *StateUpdate {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.heap) == 0 {
		return nil
	}

	item := heap.Pop(&pq.heap).(*priorityItem)
	return item.update
}

// Len returns the number of items in the queue.
func (pq *StateUpdatePriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.heap)
}

// Cap returns the capacity of the queue.
func (pq *StateUpdatePriorityQueue) Cap() int {
	return pq.cap
}

// IsEmpty returns true if the queue is empty.
func (pq *StateUpdatePriorityQueue) IsEmpty() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.heap) == 0
}

// IsFull returns true if the queue is full.
func (pq *StateUpdatePriorityQueue) IsFull() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.heap) >= pq.cap
}
