// Package sprites provides size-based anatomy proportion scaling for entity sprites.
// Entity size classes (Tiny, Small, Medium, Large, Huge) produce dramatically
// different body proportions so entities of different sizes are immediately
// distinguishable by silhouette, even when rendered at the same pixel dimensions.
//
// Tiny creatures get oversized heads (chibi-like), while Huge creatures get
// massive torsos with proportionally small heads, creating an intimidating
// presence visible from the top-down aerial perspective.
package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// SizeClass represents an entity's physical size category.
type SizeClass int

const (
	// SizeClassMedium is the default baseline size.
	SizeClassMedium SizeClass = iota
	// SizeClassTiny represents very small entities (sprites, insects, critters).
	SizeClassTiny
	// SizeClassSmall represents smaller-than-average entities (goblins, halflings).
	SizeClassSmall
	// SizeClassLarge represents larger-than-average entities (ogres, large beasts).
	SizeClassLarge
	// SizeClassHuge represents massive entities (giants, dragons, bosses).
	SizeClassHuge
)

// String returns the size class name.
func (s SizeClass) String() string {
	switch s {
	case SizeClassTiny:
		return "tiny"
	case SizeClassSmall:
		return "small"
	case SizeClassMedium:
		return "medium"
	case SizeClassLarge:
		return "large"
	case SizeClassHuge:
		return "huge"
	default:
		return "medium"
	}
}

// ParseSizeClass converts a string size class to the SizeClass enum.
func ParseSizeClass(s string) SizeClass {
	switch s {
	case "tiny":
		return SizeClassTiny
	case "small":
		return SizeClassSmall
	case "large":
		return SizeClassLarge
	case "huge":
		return SizeClassHuge
	default:
		return SizeClassMedium
	}
}

// SizeScaleModifiers holds per-body-part proportion adjustments that transform
// an anatomy template to reflect entity size. Unlike BodyTypeModifiers (which
// vary body build within the same size), these change the fundamental
// head-to-body ratio to visually communicate scale.
type SizeScaleModifiers struct {
	// HeadWidthScale adjusts head width (>1 = bigger head).
	HeadWidthScale float64
	// HeadHeightScale adjusts head height.
	HeadHeightScale float64
	// TorsoWidthScale adjusts torso/body width.
	TorsoWidthScale float64
	// TorsoHeightScale adjusts torso/body height.
	TorsoHeightScale float64
	// ArmWidthScale adjusts arm/limb span.
	ArmWidthScale float64
	// ArmHeightScale adjusts arm/limb thickness.
	ArmHeightScale float64
	// LegWidthScale adjusts leg/foot spread.
	LegWidthScale float64
	// LegHeightScale adjusts leg visibility.
	LegHeightScale float64
	// ShadowWidthScale adjusts ground shadow to match body.
	ShadowWidthScale float64
	// ShadowOpacityScale adjusts shadow darkness (larger = darker shadow).
	ShadowOpacityScale float64
	// PreferredHeadShapes overrides head shape preferences if non-nil.
	PreferredHeadShapes []shapes.ShapeType
	// PreferredTorsoShapes overrides torso shape preferences if non-nil.
	PreferredTorsoShapes []shapes.ShapeType
}

// GetSizeScaleModifiers returns proportion scaling factors for the given size class.
//
// Design rationale for top-down aerial view:
//   - Tiny: Chibi-like oversized head dominates the sprite. Body is compact.
//     Conveys "small and cute/weak" at a glance.
//   - Small: Slightly larger head, narrower body. Subtle but recognizable.
//   - Medium: Baseline (1.0x all scales). No modification.
//   - Large: Head shrinks relative to body, torso widens. Conveys "big and strong."
//   - Huge: Tiny head relative to massive torso. Arms and legs thicken.
//     Conveys "enormous and intimidating." Shadow grows dark and wide.
func GetSizeScaleModifiers(sc SizeClass) SizeScaleModifiers {
	switch sc {
	case SizeClassTiny:
		return SizeScaleModifiers{
			HeadWidthScale:     1.35,
			HeadHeightScale:    1.40,
			TorsoWidthScale:    0.75,
			TorsoHeightScale:   0.70,
			ArmWidthScale:      0.65,
			ArmHeightScale:     0.60,
			LegWidthScale:      0.60,
			LegHeightScale:     0.55,
			ShadowWidthScale:   0.70,
			ShadowOpacityScale: 0.75,
			PreferredHeadShapes: []shapes.ShapeType{
				shapes.ShapeCircle,
			},
			PreferredTorsoShapes: []shapes.ShapeType{
				shapes.ShapeEllipse,
			},
		}

	case SizeClassSmall:
		return SizeScaleModifiers{
			HeadWidthScale:     1.15,
			HeadHeightScale:    1.12,
			TorsoWidthScale:    0.88,
			TorsoHeightScale:   0.90,
			ArmWidthScale:      0.85,
			ArmHeightScale:     0.85,
			LegWidthScale:      0.85,
			LegHeightScale:     0.82,
			ShadowWidthScale:   0.88,
			ShadowOpacityScale: 0.90,
		}

	case SizeClassLarge:
		return SizeScaleModifiers{
			HeadWidthScale:     0.82,
			HeadHeightScale:    0.80,
			TorsoWidthScale:    1.25,
			TorsoHeightScale:   1.15,
			ArmWidthScale:      1.30,
			ArmHeightScale:     1.20,
			LegWidthScale:      1.25,
			LegHeightScale:     1.15,
			ShadowWidthScale:   1.25,
			ShadowOpacityScale: 1.15,
			PreferredTorsoShapes: []shapes.ShapeType{
				shapes.ShapeRectangle, shapes.ShapeBean,
			},
		}

	case SizeClassHuge:
		return SizeScaleModifiers{
			HeadWidthScale:     0.65,
			HeadHeightScale:    0.60,
			TorsoWidthScale:    1.45,
			TorsoHeightScale:   1.35,
			ArmWidthScale:      1.50,
			ArmHeightScale:     1.40,
			LegWidthScale:      1.45,
			LegHeightScale:     1.30,
			ShadowWidthScale:   1.45,
			ShadowOpacityScale: 1.30,
			PreferredTorsoShapes: []shapes.ShapeType{
				shapes.ShapeRectangle, shapes.ShapeBean, shapes.ShapeEllipse,
			},
		}

	default: // Medium
		return SizeScaleModifiers{
			HeadWidthScale:     1.0,
			HeadHeightScale:    1.0,
			TorsoWidthScale:    1.0,
			TorsoHeightScale:   1.0,
			ArmWidthScale:      1.0,
			ArmHeightScale:     1.0,
			LegWidthScale:      1.0,
			LegHeightScale:     1.0,
			ShadowWidthScale:   1.0,
			ShadowOpacityScale: 1.0,
		}
	}
}

// ApplySizeScaling modifies an AnatomicalTemplate's proportions based on entity
// size class. Returns a new template with adjusted body part dimensions. Medium
// size returns the template unmodified. This should be applied AFTER body type
// modifications so size and build stack correctly.
func ApplySizeScaling(template AnatomicalTemplate, sizeClass string) AnatomicalTemplate {
	sc := ParseSizeClass(sizeClass)
	if sc == SizeClassMedium {
		return template
	}

	mods := GetSizeScaleModifiers(sc)
	result := AnatomicalTemplate{
		Name:           template.Name + "_size_" + sc.String(),
		BodyPartLayout: make(map[BodyPart]PartSpec, len(template.BodyPartLayout)),
	}

	for part, spec := range template.BodyPartLayout {
		modified := spec
		switch part {
		case PartHead, PartHelmet:
			modified.RelativeWidth *= mods.HeadWidthScale
			modified.RelativeHeight *= mods.HeadHeightScale
			if mods.PreferredHeadShapes != nil {
				modified.ShapeTypes = mods.PreferredHeadShapes
			}
		case PartTorso, PartArmor:
			modified.RelativeWidth *= mods.TorsoWidthScale
			modified.RelativeHeight *= mods.TorsoHeightScale
			if mods.PreferredTorsoShapes != nil {
				modified.ShapeTypes = mods.PreferredTorsoShapes
			}
		case PartArms:
			modified.RelativeWidth *= mods.ArmWidthScale
			modified.RelativeHeight *= mods.ArmHeightScale
		case PartLegs:
			modified.RelativeWidth *= mods.LegWidthScale
			modified.RelativeHeight *= mods.LegHeightScale
		case PartShadow:
			modified.RelativeWidth *= mods.ShadowWidthScale
			modified.Opacity *= mods.ShadowOpacityScale
			if modified.Opacity > 0.55 {
				modified.Opacity = 0.55
			}
		}

		// Scale PreferredPixelSize if present
		if modified.PreferredPixelSize != nil {
			pps := *modified.PreferredPixelSize
			switch part {
			case PartHead, PartHelmet:
				pps.Width = scalePixelDim(pps.Width, mods.HeadWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.HeadHeightScale)
			case PartTorso, PartArmor:
				pps.Width = scalePixelDim(pps.Width, mods.TorsoWidthScale)
				pps.Height = scalePixelDim(pps.Height, mods.TorsoHeightScale)
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
