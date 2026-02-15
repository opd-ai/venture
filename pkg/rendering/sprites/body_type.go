// Package sprites provides a body type variety system that assigns dramatically
// different body proportions to humanoid entities based on their seed. Each body
// type modifies template width, height, and shape preferences to produce
// immediately distinguishable silhouettes at 32×32 pixels from a top-down view.
package sprites

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// BodyType represents a distinct humanoid body build visible from top-down.
type BodyType int

const (
	// BodyTypeAverage is the baseline proportions.
	BodyTypeAverage BodyType = iota
	// BodyTypeStocky has a wide torso and shorter stature.
	BodyTypeStocky
	// BodyTypeLean has a narrow torso with elongated proportions.
	BodyTypeLean
	// BodyTypeMuscular has very wide shoulders and arms.
	BodyTypeMuscular
	// BodyTypeHeavy has wide, rounded proportions throughout.
	BodyTypeHeavy
	// BodyTypePetite has smaller, rounder proportions overall.
	BodyTypePetite
	// BodyTypeBroad has an extra-wide torso and shoulders.
	BodyTypeBroad
	// BodyTypeLanky has narrow body with longer limb reach.
	BodyTypeLanky

	// BodyTypeCount is the number of body types (must be last).
	BodyTypeCount
)

// String returns the body type name.
func (b BodyType) String() string {
	switch b {
	case BodyTypeAverage:
		return "average"
	case BodyTypeStocky:
		return "stocky"
	case BodyTypeLean:
		return "lean"
	case BodyTypeMuscular:
		return "muscular"
	case BodyTypeHeavy:
		return "heavy"
	case BodyTypePetite:
		return "petite"
	case BodyTypeBroad:
		return "broad"
	case BodyTypeLanky:
		return "lanky"
	default:
		return "unknown"
	}
}

// BodyTypeModifiers holds per-body-part scale factors and shape preferences
// that transform a base aerial template into a distinct body type silhouette.
type BodyTypeModifiers struct {
	// TorsoWidthScale multiplies torso RelativeWidth.
	TorsoWidthScale float64
	// TorsoHeightScale multiplies torso RelativeHeight.
	TorsoHeightScale float64
	// HeadWidthScale multiplies head RelativeWidth.
	HeadWidthScale float64
	// HeadHeightScale multiplies head RelativeHeight.
	HeadHeightScale float64
	// ArmWidthScale multiplies arm RelativeWidth.
	ArmWidthScale float64
	// ArmHeightScale multiplies arm RelativeHeight.
	ArmHeightScale float64
	// LegWidthScale multiplies leg RelativeWidth.
	LegWidthScale float64
	// LegHeightScale multiplies leg RelativeHeight.
	LegHeightScale float64
	// ShadowWidthScale multiplies shadow width to match body.
	ShadowWidthScale float64
	// PreferredTorsoShapes overrides torso shape types if non-nil.
	PreferredTorsoShapes []shapes.ShapeType
	// PreferredHeadShapes overrides head shape types if non-nil.
	PreferredHeadShapes []shapes.ShapeType
}

// GetBodyTypeModifiers returns the proportion modifiers for a body type.
func GetBodyTypeModifiers(bt BodyType) BodyTypeModifiers {
	switch bt {
	case BodyTypeStocky:
		return BodyTypeModifiers{
			TorsoWidthScale:      1.35,
			TorsoHeightScale:     0.90,
			HeadWidthScale:       1.10,
			HeadHeightScale:      1.05,
			ArmWidthScale:        1.20,
			ArmHeightScale:       1.10,
			LegWidthScale:        1.25,
			LegHeightScale:       0.85,
			ShadowWidthScale:     1.25,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
			PreferredHeadShapes:  []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		}
	case BodyTypeLean:
		return BodyTypeModifiers{
			TorsoWidthScale:      0.78,
			TorsoHeightScale:     1.10,
			HeadWidthScale:       0.90,
			HeadHeightScale:      1.05,
			ArmWidthScale:        0.80,
			ArmHeightScale:       1.10,
			LegWidthScale:        0.80,
			LegHeightScale:       1.15,
			ShadowWidthScale:     0.82,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeRectangle},
			PreferredHeadShapes:  []shapes.ShapeType{shapes.ShapeEllipse},
		}
	case BodyTypeMuscular:
		return BodyTypeModifiers{
			TorsoWidthScale:      1.25,
			TorsoHeightScale:     1.05,
			HeadWidthScale:       1.05,
			HeadHeightScale:      1.00,
			ArmWidthScale:        1.45,
			ArmHeightScale:       1.15,
			LegWidthScale:        1.15,
			LegHeightScale:       1.00,
			ShadowWidthScale:     1.30,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeBean},
			PreferredHeadShapes:  []shapes.ShapeType{shapes.ShapeCircle},
		}
	case BodyTypeHeavy:
		return BodyTypeModifiers{
			TorsoWidthScale:      1.40,
			TorsoHeightScale:     1.10,
			HeadWidthScale:       1.15,
			HeadHeightScale:      1.10,
			ArmWidthScale:        1.30,
			ArmHeightScale:       1.05,
			LegWidthScale:        1.30,
			LegHeightScale:       0.90,
			ShadowWidthScale:     1.35,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCircle},
			PreferredHeadShapes:  []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		}
	case BodyTypePetite:
		return BodyTypeModifiers{
			TorsoWidthScale:     0.82,
			TorsoHeightScale:    0.85,
			HeadWidthScale:      1.10,
			HeadHeightScale:     1.10,
			ArmWidthScale:       0.80,
			ArmHeightScale:      0.85,
			LegWidthScale:       0.80,
			LegHeightScale:      0.85,
			ShadowWidthScale:    0.80,
			PreferredHeadShapes: []shapes.ShapeType{shapes.ShapeCircle},
		}
	case BodyTypeBroad:
		return BodyTypeModifiers{
			TorsoWidthScale:      1.38,
			TorsoHeightScale:     1.00,
			HeadWidthScale:       1.08,
			HeadHeightScale:      1.00,
			ArmWidthScale:        1.35,
			ArmHeightScale:       1.00,
			LegWidthScale:        1.20,
			LegHeightScale:       1.00,
			ShadowWidthScale:     1.30,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeEllipse},
		}
	case BodyTypeLanky:
		return BodyTypeModifiers{
			TorsoWidthScale:      0.75,
			TorsoHeightScale:     1.12,
			HeadWidthScale:       0.88,
			HeadHeightScale:      1.00,
			ArmWidthScale:        0.85,
			ArmHeightScale:       1.25,
			LegWidthScale:        0.78,
			LegHeightScale:       1.20,
			ShadowWidthScale:     0.80,
			PreferredTorsoShapes: []shapes.ShapeType{shapes.ShapeCapsule},
			PreferredHeadShapes:  []shapes.ShapeType{shapes.ShapeEllipse},
		}
	default: // Average
		return BodyTypeModifiers{
			TorsoWidthScale:  1.00,
			TorsoHeightScale: 1.00,
			HeadWidthScale:   1.00,
			HeadHeightScale:  1.00,
			ArmWidthScale:    1.00,
			ArmHeightScale:   1.00,
			LegWidthScale:    1.00,
			LegHeightScale:   1.00,
			ShadowWidthScale: 1.00,
		}
	}
}

// DeriveBodyType deterministically selects a body type from a seed.
// Genre influences the distribution: some genres favor certain builds.
func DeriveBodyType(seed int64, genre string) BodyType {
	rng := rand.New(rand.NewSource(seed ^ 0x426F6479)) // "Body" XOR salt
	weights := genreBodyTypeWeights(genre)
	total := 0.0
	for _, w := range weights {
		total += w
	}
	roll := rng.Float64() * total
	cumulative := 0.0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return BodyType(i)
		}
	}
	return BodyTypeAverage
}

// genreBodyTypeWeights returns per-body-type probability weights for a genre.
// Order matches BodyType enum: Average, Stocky, Lean, Muscular, Heavy, Petite, Broad, Lanky.
func genreBodyTypeWeights(genre string) [8]float64 {
	switch genre {
	case "fantasy":
		// Wide variety — heroes of all shapes
		return [8]float64{1.0, 1.2, 1.0, 1.3, 0.8, 0.9, 1.0, 0.8}
	case "horror":
		// Extremes: gaunt or hulking
		return [8]float64{0.5, 0.6, 1.5, 0.4, 1.3, 0.8, 0.5, 1.4}
	case "cyberpunk":
		// Lean and augmented or heavy-set
		return [8]float64{0.8, 0.7, 1.4, 1.2, 0.9, 0.6, 1.1, 1.3}
	case "sci-fi", "scifi":
		// More uniform but with lean/muscular bias
		return [8]float64{1.2, 0.7, 1.3, 1.1, 0.6, 0.8, 0.8, 1.0}
	case "post-apocalyptic", "postapoc":
		// Survival builds: stocky, lean survivors
		return [8]float64{0.8, 1.4, 1.3, 0.9, 0.6, 0.7, 1.0, 1.1}
	default:
		// Equal distribution
		return [8]float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	}
}

// ApplyBodyTypeToTemplate modifies an AnatomicalTemplate's proportions and shape
// preferences according to the given body type modifiers. Returns a new template.
func ApplyBodyTypeToTemplate(template AnatomicalTemplate, bt BodyType) AnatomicalTemplate {
	if bt == BodyTypeAverage {
		return template
	}

	mods := GetBodyTypeModifiers(bt)
	result := AnatomicalTemplate{
		Name:           template.Name + "_" + bt.String(),
		BodyPartLayout: make(map[BodyPart]PartSpec, len(template.BodyPartLayout)),
	}

	for part, spec := range template.BodyPartLayout {
		modified := spec
		switch part {
		case PartTorso, PartArmor:
			modified.RelativeWidth *= mods.TorsoWidthScale
			modified.RelativeHeight *= mods.TorsoHeightScale
			if mods.PreferredTorsoShapes != nil {
				modified.ShapeTypes = mods.PreferredTorsoShapes
			}
		case PartHead, PartHelmet:
			modified.RelativeWidth *= mods.HeadWidthScale
			modified.RelativeHeight *= mods.HeadHeightScale
			if mods.PreferredHeadShapes != nil {
				modified.ShapeTypes = mods.PreferredHeadShapes
			}
		case PartArms:
			modified.RelativeWidth *= mods.ArmWidthScale
			modified.RelativeHeight *= mods.ArmHeightScale
		case PartLegs:
			modified.RelativeWidth *= mods.LegWidthScale
			modified.RelativeHeight *= mods.LegHeightScale
		case PartShadow:
			modified.RelativeWidth *= mods.ShadowWidthScale
		}

		// Scale PreferredPixelSize if present
		if modified.PreferredPixelSize != nil {
			pps := *modified.PreferredPixelSize
			switch part {
			case PartTorso, PartArmor:
				pps.Width = scalePixelDim(pps.Width, mods.TorsoWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.TorsoHeightScale)
			case PartHead, PartHelmet:
				pps.Width = scalePixelDim(pps.Width, mods.HeadWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.HeadHeightScale)
			case PartArms:
				pps.Width = scalePixelDim(pps.Width, mods.ArmWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.ArmHeightScale)
			case PartLegs:
				pps.Width = scalePixelDim(pps.Width, mods.LegWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.LegHeightScale)
			}
			modified.PreferredPixelSize = &pps
		}

		result.BodyPartLayout[part] = modified
	}

	return result
}

// scalePixelDim scales a pixel dimension by a factor, clamping to [1, max].
func scalePixelDim(original int, factor float64) int {
	scaled := int(float64(original)*factor + 0.5)
	if scaled < 1 {
		scaled = 1
	}
	return scaled
}
