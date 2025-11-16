package engine

import "sync"

// nearbyResult holds temporary data for collision queries.
type nearbyResult struct {
	seen   map[uint64]bool
	result []*Entity
}

// nearbyResultPool provides pooling for collision query results to reduce allocations.
var nearbyResultPool = sync.Pool{
	New: func() interface{} {
		return &nearbyResult{
			seen:   make(map[uint64]bool, 32),
			result: make([]*Entity, 0, 32),
		}
	},
}

// getNearbyResult gets a pooled nearby result.
func getNearbyResult() *nearbyResult {
	nr := nearbyResultPool.Get().(*nearbyResult)
	// Reset the map and slice for reuse
	for k := range nr.seen {
		delete(nr.seen, k)
	}
	nr.result = nr.result[:0]
	return nr
}

// putNearbyResult returns a nearby result to the pool.
func putNearbyResult(nr *nearbyResult) {
	// Clear entity references to avoid memory leaks
	for i := range nr.result {
		nr.result[i] = nil
	}
	nr.result = nr.result[:0]

	// Clear map if it grew too large (avoid memory bloat)
	if len(nr.seen) > 128 {
		nr.seen = make(map[uint64]bool, 32)
	}

	nearbyResultPool.Put(nr)
}
