package sfx

import (
	"sync"

	"github.com/opd-ai/venture/pkg/audio"
)

// VarietyManager manages sound effect variety by caching multiple variants
// of each sound type to avoid repetitive audio. This creates more natural
// soundscapes by varying pitch, volume, and timbre for repeated sounds.
type VarietyManager struct {
	generator    *Generator
	variantCache map[string][]*audio.AudioSample
	mu           sync.RWMutex
	sampleRate   int
	seed         int64

	// Configuration
	variantsPerEffect int
	pitchVariance     float64
	volumeVariance    float64
}

// NewVarietyManager creates a new SFX variety manager.
func NewVarietyManager(sampleRate int, seed int64) *VarietyManager {
	return &VarietyManager{
		generator:         NewGenerator(sampleRate, seed),
		variantCache:      make(map[string][]*audio.AudioSample),
		sampleRate:        sampleRate,
		seed:              seed,
		variantsPerEffect: 5,
		pitchVariance:     2.0,
		volumeVariance:    0.2,
	}
}

// Generate creates a sound effect with automatic variety.
// Returns a randomly selected variant from the cache if available,
// or generates new variants if not.
func (vm *VarietyManager) Generate(effectType string, seed int64) *audio.AudioSample {
	vm.mu.RLock()
	variants, exists := vm.variantCache[effectType]
	vm.mu.RUnlock()

	if !exists {
		vm.generateVariants(effectType)
		vm.mu.RLock()
		variants = vm.variantCache[effectType]
		vm.mu.RUnlock()
	}

	if len(variants) == 0 {
		return vm.generator.Generate(effectType, seed)
	}

	idx := int(seed) % len(variants)
	if idx < 0 {
		idx = -idx
	}
	return variants[idx]
}

// generateVariants creates multiple variants of a sound effect type.
func (vm *VarietyManager) generateVariants(effectType string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, exists := vm.variantCache[effectType]; exists {
		return
	}

	variants := vm.generator.GenerateMultiVariant(
		effectType,
		vm.seed,
		vm.variantsPerEffect,
		vm.pitchVariance,
		vm.volumeVariance,
	)

	vm.variantCache[effectType] = variants
}

// SetVariantsPerEffect configures how many variants to cache per effect type.
func (vm *VarietyManager) SetVariantsPerEffect(count int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if count > 0 {
		vm.variantsPerEffect = count
	}
}

// SetPitchVariance configures pitch variation range (in semitones).
func (vm *VarietyManager) SetPitchVariance(variance float64) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if variance >= 0 {
		vm.pitchVariance = variance
	}
}

// SetVolumeVariance configures volume variation range (0.0-1.0).
func (vm *VarietyManager) SetVolumeVariance(variance float64) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if variance >= 0 && variance <= 1.0 {
		vm.volumeVariance = variance
	}
}

// ClearCache clears all cached variants.
func (vm *VarietyManager) ClearCache() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.variantCache = make(map[string][]*audio.AudioSample)
}

// GetCacheSize returns the total number of cached variants.
func (vm *VarietyManager) GetCacheSize() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	total := 0
	for _, variants := range vm.variantCache {
		total += len(variants)
	}
	return total
}
