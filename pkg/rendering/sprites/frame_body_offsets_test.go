package sprites

import (
	"math"
	"testing"
)

func TestComputeFrameOffsets_AllStates(t *testing.T) {
	states := []string{"idle", "walk", "run", "attack", "cast", "hit", "death", "unknown"}
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			for i := 0; i < 8; i++ {
				offsets := ComputeFrameOffsets(state, i, 8)
				if offsets == nil {
					t.Errorf("state=%s frame=%d returned nil", state, i)
				}
				// Verify no NaN or Inf values
				for part, offset := range offsets {
					if math.IsNaN(offset.DX) || math.IsInf(offset.DX, 0) {
						t.Errorf("state=%s frame=%d part=%d DX is NaN/Inf", state, i, part)
					}
					if math.IsNaN(offset.DY) || math.IsInf(offset.DY, 0) {
						t.Errorf("state=%s frame=%d part=%d DY is NaN/Inf", state, i, part)
					}
					if math.IsNaN(offset.Scale) || math.IsInf(offset.Scale, 0) {
						t.Errorf("state=%s frame=%d part=%d Scale is NaN/Inf", state, i, part)
					}
					// Scale should be positive and reasonable
					if offset.Scale < 0.5 || offset.Scale > 2.0 {
						t.Errorf("state=%s frame=%d part=%d Scale=%f out of [0.5,2.0]", state, i, part, offset.Scale)
					}
				}
			}
		})
	}
}

func TestComputeFrameOffsets_ZeroFrameCount(t *testing.T) {
	offsets := ComputeFrameOffsets("idle", 0, 0)
	if offsets != nil {
		t.Error("expected nil for zero frame count")
	}
}

func TestComputeFrameOffsets_Deterministic(t *testing.T) {
	for _, state := range []string{"idle", "walk", "attack"} {
		a := ComputeFrameOffsets(state, 3, 8)
		b := ComputeFrameOffsets(state, 3, 8)
		for part, oa := range a {
			ob, ok := b[part]
			if !ok {
				t.Errorf("state=%s: part %d missing in second call", state, part)
				continue
			}
			if oa.DX != ob.DX || oa.DY != ob.DY || oa.Scale != ob.Scale {
				t.Errorf("state=%s: non-deterministic for part %d", state, part)
			}
		}
	}
}

func TestComputeFrameOffsets_WalkLegsAlternate(t *testing.T) {
	// Frames 0-3 and 4-7 should have opposite leg DX signs (alternating gait)
	f0 := ComputeFrameOffsets("walk", 0, 8)
	f4 := ComputeFrameOffsets("walk", 4, 8)
	legA := f0[PartLegs].DX
	legB := f4[PartLegs].DX
	// At frame 0 (sin(0)=0) and frame 4 (sin(pi)≈0), both are near zero.
	// Use frames 1 and 5 where swing is pronounced.
	f1 := ComputeFrameOffsets("walk", 1, 8)
	f5 := ComputeFrameOffsets("walk", 5, 8)
	legA = f1[PartLegs].DX
	legB = f5[PartLegs].DX
	if legA*legB > 0 {
		t.Errorf("walk legs should alternate: frame1=%f frame5=%f", legA, legB)
	}
}

func TestComputeFrameOffsets_AttackArmLunge(t *testing.T) {
	// Wind-up frame (frame 0 of 8, t=0) arm should be near neutral
	f0 := ComputeFrameOffsets("attack", 0, 8)
	// Strike frame (frame 3 of 8, t=0.375) arm should be extended forward
	f3 := ComputeFrameOffsets("attack", 3, 8)
	armStart := f0[PartArms].DX
	armStrike := f3[PartArms].DX
	if armStrike <= armStart {
		t.Errorf("attack arm should lunge forward: start=%f strike=%f", armStart, armStrike)
	}
}

func TestComputeFrameOffsets_IdleBreathing(t *testing.T) {
	// Torso scale should oscillate during idle
	scales := make([]float64, 8)
	for i := 0; i < 8; i++ {
		offsets := ComputeFrameOffsets("idle", i, 8)
		scales[i] = offsets[PartTorso].Scale
	}
	// Check that scales vary (not all identical)
	allSame := true
	for _, s := range scales[1:] {
		if math.Abs(s-scales[0]) > 1e-6 {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("idle torso scale should oscillate (breathing), but all frames identical")
	}
}

func TestComputeFrameOffsets_DeathProgressive(t *testing.T) {
	// Head scale should decrease progressively during death
	f0 := ComputeFrameOffsets("death", 0, 8)
	f7 := ComputeFrameOffsets("death", 7, 8)
	if f7[PartHead].Scale >= f0[PartHead].Scale {
		t.Errorf("death head should shrink: frame0=%f frame7=%f", f0[PartHead].Scale, f7[PartHead].Scale)
	}
	// Shadow should grow
	if f7[PartShadow].Scale <= f0[PartShadow].Scale {
		t.Errorf("death shadow should grow: frame0=%f frame7=%f", f0[PartShadow].Scale, f7[PartShadow].Scale)
	}
}

func TestComputeFrameOffsets_OffsetBounds(t *testing.T) {
	// All offsets should be within reasonable bounds for 32x32 sprites
	// DX/DY as fraction of sprite: should be < 0.2 (6.4px at 32px)
	states := []string{"idle", "walk", "run", "attack", "cast", "hit", "death"}
	for _, state := range states {
		for i := 0; i < 8; i++ {
			offsets := ComputeFrameOffsets(state, i, 8)
			for part, o := range offsets {
				if math.Abs(o.DX) > 0.2 {
					t.Errorf("state=%s frame=%d part=%d DX=%f exceeds ±0.2", state, i, part, o.DX)
				}
				if math.Abs(o.DY) > 0.2 {
					t.Errorf("state=%s frame=%d part=%d DY=%f exceeds ±0.2", state, i, part, o.DY)
				}
			}
		}
	}
}

func BenchmarkComputeFrameOffsets(b *testing.B) {
	states := []string{"idle", "walk", "attack"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, state := range states {
			ComputeFrameOffsets(state, i%8, 8)
		}
	}
}

func TestApplyFrameOffsetsToSpec(t *testing.T) {
	spec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.5,
		RelativeWidth:  0.4,
		RelativeHeight: 0.3,
		ZIndex:         10,
	}
	offsets := FrameOffsetMap{
		PartTorso: {DX: 0.02, DY: -0.01, Scale: 1.1},
	}

	// Part present in offsets — should be modified
	result := applyFrameOffsetsToSpec(spec, PartTorso, offsets, 32, 32)
	if math.Abs(result.RelativeX-0.52) > 1e-6 {
		t.Errorf("expected RelativeX=0.52, got %f", result.RelativeX)
	}
	if math.Abs(result.RelativeY-0.49) > 1e-6 {
		t.Errorf("expected RelativeY=0.49, got %f", result.RelativeY)
	}
	if math.Abs(result.RelativeWidth-0.44) > 1e-6 {
		t.Errorf("expected RelativeWidth=0.44, got %f", result.RelativeWidth)
	}
	if math.Abs(result.RelativeHeight-0.33) > 1e-6 {
		t.Errorf("expected RelativeHeight=0.33, got %f", result.RelativeHeight)
	}

	// Part NOT in offsets — should return unmodified
	unchanged := applyFrameOffsetsToSpec(spec, PartHead, offsets, 32, 32)
	if unchanged.RelativeX != spec.RelativeX || unchanged.RelativeY != spec.RelativeY {
		t.Error("spec for absent part should be unchanged")
	}

	// Scale of 1.0 should not modify dimensions
	offsets2 := FrameOffsetMap{
		PartHead: {DX: 0.01, DY: 0, Scale: 1.0},
	}
	result2 := applyFrameOffsetsToSpec(spec, PartHead, offsets2, 32, 32)
	if result2.RelativeWidth != spec.RelativeWidth {
		t.Errorf("scale=1.0 should not change width: got %f", result2.RelativeWidth)
	}
}
