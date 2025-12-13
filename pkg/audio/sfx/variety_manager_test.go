package sfx

import (
	"testing"
)

func TestVarietyManager_NewVarietyManager(t *testing.T) {
	vm := NewVarietyManager(44100, 12345)
	if vm == nil {
		t.Fatal("expected non-nil VarietyManager")
	}
	if vm.sampleRate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", vm.sampleRate)
	}
	if vm.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", vm.seed)
	}
	if vm.variantsPerEffect != 5 {
		t.Errorf("expected 5 variants per effect, got %d", vm.variantsPerEffect)
	}
}

func TestVarietyManager_Generate(t *testing.T) {
	vm := NewVarietyManager(44100, 12345)

	tests := []struct {
		name       string
		effectType string
		seed       int64
	}{
		{"impact", "impact", 100},
		{"explosion", "explosion", 200},
		{"magic", "magic", 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := vm.Generate(tt.effectType, tt.seed)
			if sample == nil {
				t.Fatal("expected non-nil AudioSample")
			}
			if len(sample.Data) == 0 {
				t.Error("expected non-empty audio data")
			}
		})
	}
}

func TestVarietyManager_VariantCaching(t *testing.T) {
	vm := NewVarietyManager(44100, 12345)
	vm.SetVariantsPerEffect(3)

	// Generate first sample (should create cache)
	sample1 := vm.Generate("impact", 100)
	if sample1 == nil {
		t.Fatal("expected non-nil sample")
	}

	// Check cache was populated
	cacheSize := vm.GetCacheSize()
	if cacheSize != 3 {
		t.Errorf("expected 3 cached variants, got %d", cacheSize)
	}

	// Generate second sample (should use cache)
	sample2 := vm.Generate("impact", 200)
	if sample2 == nil {
		t.Fatal("expected non-nil sample")
	}

	// Cache size should not have changed
	newCacheSize := vm.GetCacheSize()
	if newCacheSize != cacheSize {
		t.Errorf("cache size changed unexpectedly: %d -> %d", cacheSize, newCacheSize)
	}
}

func TestVarietyManager_Determinism(t *testing.T) {
	vm1 := NewVarietyManager(44100, 12345)
	vm2 := NewVarietyManager(44100, 12345)

	sample1 := vm1.Generate("impact", 999)
	sample2 := vm2.Generate("impact", 999)

	if len(sample1.Data) != len(sample2.Data) {
		t.Errorf("sample lengths differ: %d vs %d", len(sample1.Data), len(sample2.Data))
	}

	// Check first 100 samples for determinism
	checkLen := 100
	if len(sample1.Data) < checkLen {
		checkLen = len(sample1.Data)
	}
	for i := 0; i < checkLen; i++ {
		if sample1.Data[i] != sample2.Data[i] {
			t.Errorf("sample data differs at index %d: %f vs %f", i, sample1.Data[i], sample2.Data[i])
			break
		}
	}
}

func TestVarietyManager_ClearCache(t *testing.T) {
	vm := NewVarietyManager(44100, 12345)

	vm.Generate("impact", 100)
	cacheSize := vm.GetCacheSize()
	if cacheSize == 0 {
		t.Error("expected non-zero cache size")
	}

	vm.ClearCache()
	newSize := vm.GetCacheSize()
	if newSize != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", newSize)
	}
}

func TestVarietyManager_Configuration(t *testing.T) {
	vm := NewVarietyManager(44100, 12345)

	vm.SetVariantsPerEffect(10)
	vm.SetPitchVariance(3.0)
	vm.SetVolumeVariance(0.5)

	if vm.variantsPerEffect != 10 {
		t.Errorf("expected 10 variants, got %d", vm.variantsPerEffect)
	}
	if vm.pitchVariance != 3.0 {
		t.Errorf("expected pitch variance 3.0, got %f", vm.pitchVariance)
	}
	if vm.volumeVariance != 0.5 {
		t.Errorf("expected volume variance 0.5, got %f", vm.volumeVariance)
	}
}

func BenchmarkVarietyManager_Generate(b *testing.B) {
	vm := NewVarietyManager(44100, 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vm.Generate("impact", int64(i))
	}
}
