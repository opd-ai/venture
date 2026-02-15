package sprites

import (
	"math"
	"testing"
)

// TestComputeCreatureFrameOffsets_HumanoidDelegation verifies humanoid types
// delegate to the standard ComputeFrameOffsets.
func TestComputeCreatureFrameOffsets_HumanoidDelegation(t *testing.T) {
	humanoidTypes := []string{"humanoid", "player", "npc", "knight", "mage", "warrior"}
	for _, et := range humanoidTypes {
		for _, state := range []string{"idle", "walk", "attack", "hit", "death"} {
			creature := ComputeCreatureFrameOffsets(state, 3, 8, et)
			standard := ComputeFrameOffsets(state, 3, 8)
			if len(creature) != len(standard) {
				t.Errorf("humanoid %q state %q: offset count %d != standard %d", et, state, len(creature), len(standard))
			}
			for part, co := range creature {
				so, ok := standard[part]
				if !ok {
					t.Errorf("humanoid %q state %q: extra part %v", et, state, part)
					continue
				}
				if co.DX != so.DX || co.DY != so.DY || co.Scale != so.Scale {
					t.Errorf("humanoid %q state %q part %v: mismatch", et, state, part)
				}
			}
		}
	}
}

// TestComputeCreatureFrameOffsets_ZeroFrameCount returns nil for zero frames.
func TestComputeCreatureFrameOffsets_ZeroFrameCount(t *testing.T) {
	for _, et := range []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie", "player"} {
		offsets := ComputeCreatureFrameOffsets("idle", 0, 0, et)
		if offsets != nil {
			t.Errorf("entityType %q: expected nil for zero frameCount, got %v", et, offsets)
		}
	}
}

// TestComputeCreatureFrameOffsets_AllCreatureTypes verifies every creature type
// produces non-nil offsets for all animation states.
func TestComputeCreatureFrameOffsets_AllCreatureTypes(t *testing.T) {
	types := []struct {
		entityType string
		category   string
	}{
		{"wolf", "quadruped"},
		{"bear", "quadruped"},
		{"horse", "quadruped"},
		{"snake", "serpentine"},
		{"worm", "serpentine"},
		{"spider", "arachnid"},
		{"beetle", "arachnid"},
		{"dragon", "flying"},
		{"bat", "flying"},
		{"slime", "blob"},
		{"ooze", "blob"},
		{"robot", "mechanical"},
		{"golem", "mechanical"},
		{"zombie", "undead"},
		{"skeleton", "undead"},
	}

	states := []string{"idle", "walk", "run", "attack", "hit", "death", "cast"}
	frameCount := 8

	for _, tt := range types {
		for _, state := range states {
			for i := 0; i < frameCount; i++ {
				offsets := ComputeCreatureFrameOffsets(state, i, frameCount, tt.entityType)
				if offsets == nil {
					t.Errorf("%s(%s) state=%s frame=%d: nil offsets", tt.entityType, tt.category, state, i)
					continue
				}
				if len(offsets) == 0 {
					t.Errorf("%s(%s) state=%s frame=%d: empty offsets", tt.entityType, tt.category, state, i)
				}
			}
		}
	}
}

// TestComputeCreatureFrameOffsets_Deterministic verifies same inputs produce same outputs.
func TestComputeCreatureFrameOffsets_Deterministic(t *testing.T) {
	types := []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie"}
	for _, et := range types {
		for _, state := range []string{"idle", "walk", "attack"} {
			a := ComputeCreatureFrameOffsets(state, 4, 8, et)
			b := ComputeCreatureFrameOffsets(state, 4, 8, et)
			for part, ao := range a {
				bo := b[part]
				if ao.DX != bo.DX || ao.DY != bo.DY || ao.Scale != bo.Scale {
					t.Errorf("%s state=%s part=%v: not deterministic", et, state, part)
				}
			}
		}
	}
}

// TestComputeCreatureFrameOffsets_DifferentFromHumanoid verifies creature-specific
// offsets differ from the default humanoid offsets.
func TestComputeCreatureFrameOffsets_DifferentFromHumanoid(t *testing.T) {
	types := []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie"}
	for _, et := range types {
		creatureIdle := ComputeCreatureFrameOffsets("idle", 2, 8, et)
		humanoidIdle := ComputeFrameOffsets("idle", 2, 8)

		// At least one offset should differ (creatures have different body plans)
		anyDiff := false
		for part, co := range creatureIdle {
			ho, ok := humanoidIdle[part]
			if !ok {
				anyDiff = true
				break
			}
			if co.DX != ho.DX || co.DY != ho.DY || co.Scale != ho.Scale {
				anyDiff = true
				break
			}
		}
		if !anyDiff && len(creatureIdle) == len(humanoidIdle) {
			t.Errorf("%s idle offsets identical to humanoid — creature animation not differentiated", et)
		}
	}
}

// TestComputeCreatureFrameOffsets_OffsetsInRange verifies offset values stay
// within reasonable bounds to prevent sprite glitches.
func TestComputeCreatureFrameOffsets_OffsetsInRange(t *testing.T) {
	types := []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie"}
	states := []string{"idle", "walk", "run", "attack", "hit", "death"}
	const maxDelta = 0.2  // Max positional offset as fraction of sprite
	const maxScale = 1.5  // Max scale multiplier
	const minScale = 0.3  // Min scale multiplier

	for _, et := range types {
		for _, state := range states {
			for i := 0; i < 8; i++ {
				offsets := ComputeCreatureFrameOffsets(state, i, 8, et)
				for part, o := range offsets {
					if math.Abs(o.DX) > maxDelta {
						t.Errorf("%s %s frame=%d part=%v: DX=%.3f exceeds ±%.1f", et, state, i, part, o.DX, maxDelta)
					}
					if math.Abs(o.DY) > maxDelta {
						t.Errorf("%s %s frame=%d part=%v: DY=%.3f exceeds ±%.1f", et, state, i, part, o.DY, maxDelta)
					}
					if o.Scale != 0 && (o.Scale > maxScale || o.Scale < minScale) {
						t.Errorf("%s %s frame=%d part=%v: Scale=%.3f outside [%.1f, %.1f]", et, state, i, part, o.Scale, minScale, maxScale)
					}
				}
			}
		}
	}
}

// TestCreatureCategory verifies entity type to category mapping.
func TestCreatureCategory(t *testing.T) {
	tests := []struct {
		entityType string
		want       string
	}{
		{"wolf", "quadruped"},
		{"bear", "quadruped"},
		{"snake", "serpentine"},
		{"wyrm", "serpentine"},
		{"spider", "arachnid"},
		{"beetle", "arachnid"},
		{"dragon", "flying"},
		{"bat", "flying"},
		{"slime", "blob"},
		{"ooze", "blob"},
		{"robot", "mechanical"},
		{"golem", "mechanical"},
		{"zombie", "undead"},
		{"ghost", "undead"},
		{"unknown_thing", ""},
	}
	for _, tt := range tests {
		got := creatureCategory(tt.entityType)
		if got != tt.want {
			t.Errorf("creatureCategory(%q) = %q, want %q", tt.entityType, got, tt.want)
		}
	}
}

// TestCreatureFrameOffsets_WalkDiffersFromIdle verifies walk animations actually differ.
func TestCreatureFrameOffsets_WalkDiffersFromIdle(t *testing.T) {
	types := []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie"}
	for _, et := range types {
		idle := ComputeCreatureFrameOffsets("idle", 3, 8, et)
		walk := ComputeCreatureFrameOffsets("walk", 3, 8, et)

		anyDiff := false
		allParts := make(map[BodyPart]bool)
		for p := range idle {
			allParts[p] = true
		}
		for p := range walk {
			allParts[p] = true
		}

		for p := range allParts {
			io, iok := idle[p]
			wo, wok := walk[p]
			if iok != wok {
				anyDiff = true
				break
			}
			if iok && wok && (io.DX != wo.DX || io.DY != wo.DY || io.Scale != wo.Scale) {
				anyDiff = true
				break
			}
		}
		if !anyDiff {
			t.Errorf("%s: walk offsets identical to idle — no walk animation", et)
		}
	}
}

// TestCreatureFrameOffsets_FlyingHasWings verifies flying creatures animate wings.
func TestCreatureFrameOffsets_FlyingHasWings(t *testing.T) {
	for _, state := range []string{"idle", "walk", "attack"} {
		offsets := ComputeCreatureFrameOffsets(state, 2, 8, "dragon")
		if _, hasWings := offsets[PartWings]; !hasWings {
			t.Errorf("dragon %s: missing PartWings offset — wings should animate", state)
		}
	}
}

// TestCreatureFrameOffsets_QuadrupedHasTail verifies quadruped idle animates tail.
func TestCreatureFrameOffsets_QuadrupedHasTail(t *testing.T) {
	offsets := ComputeCreatureFrameOffsets("idle", 2, 8, "wolf")
	if _, hasTail := offsets[PartTail]; !hasTail {
		t.Errorf("wolf idle: missing PartTail offset — tail should wag")
	}
}

// TestComputeCreatureFrameOffsets_UnknownFallback verifies unknown types fall
// back to humanoid offsets.
func TestComputeCreatureFrameOffsets_UnknownFallback(t *testing.T) {
	creature := ComputeCreatureFrameOffsets("idle", 3, 8, "unknown_creature")
	standard := ComputeFrameOffsets("idle", 3, 8)
	if len(creature) != len(standard) {
		t.Errorf("unknown type should fall back to humanoid, got %d offsets vs %d", len(creature), len(standard))
	}
}

// BenchmarkComputeCreatureFrameOffsets benchmarks creature offset computation.
func BenchmarkComputeCreatureFrameOffsets(b *testing.B) {
	types := []string{"wolf", "snake", "spider", "dragon", "slime", "robot", "zombie", "player"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		et := types[i%len(types)]
		_ = ComputeCreatureFrameOffsets("walk", i%8, 8, et)
	}
}
