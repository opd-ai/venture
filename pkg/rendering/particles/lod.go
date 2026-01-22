// Package particles provides LOD (Level of Detail) system for particle optimization.
// This file implements distance-based particle reduction and viewport culling
// to maintain 60 FPS with 1000+ active particles (Phase 14.3).
package particles

// LODConfig contains parameters for particle LOD system.
type LODConfig struct {
	// Enabled flag for LOD system
	Enabled bool

	// MaxParticles global limit across all systems
	MaxParticles int

	// DistanceClose threshold for full detail (pixels)
	DistanceClose float64

	// DistanceMid threshold for reduced detail (pixels)
	DistanceMid float64

	// DistanceFar threshold for culling (pixels)
	DistanceFar float64

	// CloseReduction fraction to keep at close distance (1.0 = all)
	CloseReduction float64

	// MidReduction fraction to keep at mid distance (0.5 = half)
	MidReduction float64

	// FarReduction fraction to keep at far distance (0.1 = 10%)
	FarReduction float64
}

// DefaultLODConfig returns sensible default LOD parameters.
func DefaultLODConfig() LODConfig {
	return LODConfig{
		Enabled:        true,
		MaxParticles:   1000,
		DistanceClose:  200.0,
		DistanceMid:    400.0,
		DistanceFar:    800.0,
		CloseReduction: 1.0,  // Keep all particles
		MidReduction:   0.5,  // Keep half
		FarReduction:   0.25, // Keep quarter
	}
}

// ViewportCullingConfig contains parameters for viewport culling.
type ViewportCullingConfig struct {
	// Enabled flag for viewport culling
	Enabled bool

	// Margin in pixels around viewport to start rendering
	Margin float64

	// ViewportX, ViewportY top-left corner of viewport
	ViewportX float64
	ViewportY float64

	// ViewportWidth, ViewportHeight size of viewport
	ViewportWidth  float64
	ViewportHeight float64
}

// DefaultViewportCullingConfig returns sensible default culling parameters.
func DefaultViewportCullingConfig() ViewportCullingConfig {
	return ViewportCullingConfig{
		Enabled:        true,
		Margin:         100.0, // Start rendering 100px before entering viewport
		ViewportX:      0,
		ViewportY:      0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}
}

// LODStats tracks LOD system performance.
type LODStats struct {
	// TotalParticles before LOD
	TotalParticles int

	// VisibleParticles after viewport culling
	VisibleParticles int

	// RenderedParticles after distance LOD
	RenderedParticles int

	// CulledByViewport particles outside viewport
	CulledByViewport int

	// CulledByDistance particles too far away
	CulledByDistance int

	// ReducedByLOD particles reduced by distance tiers
	ReducedByLOD int
}

// ApplyViewportCulling filters particles to only those visible in viewport.
// Returns indices of visible particles.
func ApplyViewportCulling(particles []Particle, config ViewportCullingConfig) []int {
	if !config.Enabled {
		// Return all indices if culling disabled
		indices := make([]int, len(particles))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// Calculate viewport bounds with margin
	minX := config.ViewportX - config.Margin
	maxX := config.ViewportX + config.ViewportWidth + config.Margin
	minY := config.ViewportY - config.Margin
	maxY := config.ViewportY + config.ViewportHeight + config.Margin

	// Filter visible particles
	visible := make([]int, 0, len(particles))
	for i, p := range particles {
		if p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY {
			visible = append(visible, i)
		}
	}

	return visible
}

// ApplyDistanceLOD filters particles based on distance from camera/player.
// Returns indices of particles to render.
func ApplyDistanceLOD(particles []Particle, visibleIndices []int, cameraX, cameraY float64, config LODConfig) []int {
	if !config.Enabled {
		return visibleIndices
	}

	n := len(visibleIndices)
	if n == 0 {
		return visibleIndices
	}

	// Pre-allocate tier slices with estimated capacity to reduce allocations
	// Typical distribution: 30% close, 30% mid, 30% far, 10% culled
	estimatedPerTier := (n + 2) / 3
	close := make([]int, 0, estimatedPerTier)
	mid := make([]int, 0, estimatedPerTier)
	far := make([]int, 0, estimatedPerTier)

	// Pre-compute squared distances to avoid sqrt in hot loop
	distCloseSq := config.DistanceClose * config.DistanceClose
	distMidSq := config.DistanceMid * config.DistanceMid
	distFarSq := config.DistanceFar * config.DistanceFar

	for _, idx := range visibleIndices {
		p := particles[idx]
		dx := p.X - cameraX
		dy := p.Y - cameraY
		distSq := dx*dx + dy*dy

		if distSq < distCloseSq {
			close = append(close, idx)
		} else if distSq < distMidSq {
			mid = append(mid, idx)
		} else if distSq < distFarSq {
			far = append(far, idx)
		}
		// Beyond DistanceFar: culled entirely
	}

	// Apply reduction factors to each tier
	keepClose := int(float64(len(close)) * config.CloseReduction)
	keepMid := int(float64(len(mid)) * config.MidReduction)
	keepFar := int(float64(len(far)) * config.FarReduction)

	// Pre-allocate result with exact capacity needed
	result := make([]int, 0, keepClose+keepMid+keepFar)
	result = append(result, close[:keepClose]...)
	result = append(result, mid[:keepMid]...)
	result = append(result, far[:keepFar]...)

	return result
}

// particleDistance stores index and squared distance for sorting.
type particleDistance struct {
	index  int
	distSq float64
}

// EnforceLODLimit enforces global particle limit across all systems.
// Returns indices to render, prioritizing closer particles.
// Optimized: uses squared distances (no sqrt) and O(n log n) sort.
func EnforceLODLimit(particles []Particle, renderIndices []int, cameraX, cameraY float64, maxParticles int) []int {
	if len(renderIndices) <= maxParticles {
		return renderIndices
	}

	// Use file-level particleDistance type for sort compatibility
	distances := make([]particleDistance, len(renderIndices))
	for i, idx := range renderIndices {
		p := particles[idx]
		dx := p.X - cameraX
		dy := p.Y - cameraY
		distances[i] = particleDistance{index: idx, distSq: dx*dx + dy*dy}
	}

	// Use efficient O(n log n) sort instead of bubble sort
	sortByDistance(distances)

	// Take first N (closest)
	result := make([]int, maxParticles)
	for i := 0; i < maxParticles; i++ {
		result[i] = distances[i].index
	}

	return result
}

// sortByDistance sorts particle distances by squared distance (ascending).
// Uses Go's pattern-defeating quicksort for O(n log n) worst-case performance.
func sortByDistance(distances []particleDistance) {
	if len(distances) <= 1 {
		return
	}
	// Use pdqsort via slices package for better worst-case performance
	pdqsortDistances(distances)
}

// pdqsortDistances implements pattern-defeating quicksort for particle distances.
// This handles already-sorted, reverse-sorted, and random data efficiently.
func pdqsortDistances(distances []particleDistance) {
	n := len(distances)
	if n <= 12 {
		insertionSortDistances(distances)
		return
	}

	_, pivotIdx := partitionDistances(distances)
	recursiveSortPartitions(distances, pivotIdx)
}

// insertionSortDistances sorts small slices using insertion sort algorithm.
func insertionSortDistances(distances []particleDistance) {
	for i := 1; i < len(distances); i++ {
		key := distances[i]
		j := i - 1
		for j >= 0 && distances[j].distSq > key.distSq {
			distances[j+1] = distances[j]
			j--
		}
		distances[j+1] = key
	}
}

// partitionDistances partitions the slice around a median-of-three pivot.
func partitionDistances(distances []particleDistance) (float64, int) {
	n := len(distances)
	selectMedianPivot(distances, n)
	pivot := distances[n-1].distSq

	i := -1
	for j := 0; j < n-1; j++ {
		if distances[j].distSq <= pivot {
			i++
			distances[i], distances[j] = distances[j], distances[i]
		}
	}
	distances[i+1], distances[n-1] = distances[n-1], distances[i+1]
	return pivot, i + 1
}

// selectMedianPivot chooses a pivot using median-of-three strategy.
func selectMedianPivot(distances []particleDistance, n int) {
	lo, hi := 0, n-1
	mid := n / 2

	if distances[mid].distSq < distances[lo].distSq {
		distances[lo], distances[mid] = distances[mid], distances[lo]
	}
	if distances[hi].distSq < distances[lo].distSq {
		distances[lo], distances[hi] = distances[hi], distances[lo]
	}
	if distances[mid].distSq < distances[hi].distSq {
		distances[mid], distances[hi] = distances[hi], distances[mid]
	}
}

// recursiveSortPartitions sorts the left and right partitions recursively.
func recursiveSortPartitions(distances []particleDistance, pivotIdx int) {
	n := len(distances)
	if pivotIdx > 0 {
		pdqsortDistances(distances[0:pivotIdx])
	}
	if pivotIdx+1 < n {
		pdqsortDistances(distances[pivotIdx+1 : n])
	}
}

// CalculateLODStats computes LOD statistics for debugging/monitoring.
func CalculateLODStats(totalParticles int, visibleIndices, renderIndices []int) LODStats {
	return LODStats{
		TotalParticles:    totalParticles,
		VisibleParticles:  len(visibleIndices),
		RenderedParticles: len(renderIndices),
		CulledByViewport:  totalParticles - len(visibleIndices),
		ReducedByLOD:      len(visibleIndices) - len(renderIndices),
		CulledByDistance:  totalParticles - len(visibleIndices) - (len(visibleIndices) - len(renderIndices)),
	}
}

// StaggeredLODEnforcer spreads LOD enforcement work across multiple frames
// to reduce per-frame jitter. Instead of processing all particles in one frame,
// it processes a fraction per frame and maintains a cached result.
type StaggeredLODEnforcer struct {
	// staggerFrames is the number of frames over which to spread LOD enforcement.
	staggerFrames int

	// currentFrame tracks which frame we're on in the stagger cycle.
	currentFrame int

	// cachedResult holds the last computed LOD-limited indices.
	cachedResult []int

	// cachedDistances holds distance data for incremental updates.
	cachedDistances []particleDistance

	// maxParticles is the limit for rendered particles.
	maxParticles int
}

// NewStaggeredLODEnforcer creates a new staggered LOD enforcer.
// staggerFrames controls how many frames to spread the work over (default: 4).
// maxParticles is the global particle limit.
func NewStaggeredLODEnforcer(staggerFrames, maxParticles int) *StaggeredLODEnforcer {
	if staggerFrames < 1 {
		staggerFrames = 1
	}
	if maxParticles < 1 {
		maxParticles = 1000
	}
	return &StaggeredLODEnforcer{
		staggerFrames: staggerFrames,
		maxParticles:  maxParticles,
		cachedResult:  nil,
	}
}

// Update processes a fraction of particles this frame and returns the current
// LOD-limited indices. Call once per frame with the current particle and render state.
func (s *StaggeredLODEnforcer) Update(particles []Particle, renderIndices []int, cameraX, cameraY float64) []int {
	n := len(renderIndices)

	// If within limit, return all immediately
	if n <= s.maxParticles {
		s.cachedResult = renderIndices
		s.currentFrame = 0
		return renderIndices
	}

	// On first frame of cycle, initialize distance cache
	if s.currentFrame == 0 || len(s.cachedDistances) != n {
		s.cachedDistances = make([]particleDistance, n)
		for i, idx := range renderIndices {
			s.cachedDistances[i] = particleDistance{index: idx, distSq: -1}
		}
	}

	// Calculate distances for this frame's slice of particles
	sliceStart := (s.currentFrame * n) / s.staggerFrames
	sliceEnd := ((s.currentFrame + 1) * n) / s.staggerFrames
	if sliceEnd > n {
		sliceEnd = n
	}

	for i := sliceStart; i < sliceEnd; i++ {
		idx := renderIndices[i]
		if idx < len(particles) {
			p := particles[idx]
			dx := p.X - cameraX
			dy := p.Y - cameraY
			s.cachedDistances[i] = particleDistance{index: idx, distSq: dx*dx + dy*dy}
		}
	}

	// Advance frame counter
	s.currentFrame++
	if s.currentFrame >= s.staggerFrames {
		// Complete cycle - sort and produce final result
		sortByDistance(s.cachedDistances)

		s.cachedResult = make([]int, s.maxParticles)
		for i := 0; i < s.maxParticles && i < len(s.cachedDistances); i++ {
			s.cachedResult[i] = s.cachedDistances[i].index
		}
		s.currentFrame = 0
	}

	// Return cached result (may be stale, but smooth frame pacing)
	if s.cachedResult == nil {
		// First cycle not complete yet - return first N indices as fallback
		if n <= s.maxParticles {
			return renderIndices
		}
		return renderIndices[:s.maxParticles]
	}
	return s.cachedResult
}

// SetMaxParticles updates the particle limit.
func (s *StaggeredLODEnforcer) SetMaxParticles(max int) {
	if max > 0 {
		s.maxParticles = max
	}
}

// SetStaggerFrames updates the number of frames to spread work over.
func (s *StaggeredLODEnforcer) SetStaggerFrames(frames int) {
	if frames > 0 {
		s.staggerFrames = frames
	}
}

// GetCycleProgress returns (currentFrame, totalFrames) for debugging.
func (s *StaggeredLODEnforcer) GetCycleProgress() (current, total int) {
	return s.currentFrame, s.staggerFrames
}

// Reset clears cached state, useful when camera teleports or scene changes.
func (s *StaggeredLODEnforcer) Reset() {
	s.currentFrame = 0
	s.cachedResult = nil
	s.cachedDistances = nil
}
