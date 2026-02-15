// Package sprites - aerial (top-down) anatomical templates for nonhumanoid creatures.
// These templates provide direction-aware, overhead-perspective body plans for
// creature types that should NOT reuse humanoid anatomy. Each creature type has
// a distinct silhouette visible from directly above.
package sprites

import (
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// QuadrupedAerialTemplate returns a top-down quadruped template (wolves, bears, horses).
// From above: oval body with four legs radiating outward, head at front edge.
// Direction shifts head position to indicate facing.
func QuadrupedAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "quadruped_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Shadow beneath the whole body
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.88,
		RelativeWidth:  0.80,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.35,
	}

	// Four legs visible as stubs radiating from body — wide spread
	template.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.55,
		RelativeWidth:  0.90,
		RelativeHeight: 0.50,
		PreferredPixelSize: &PixelDimensions{
			Width:  28,
			Height: 16,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCross, shapes.ShapeCapsule},
		ZIndex:     3,
		ColorRole:  "primary",
		Opacity:    0.85,
	}

	// Main body — large horizontal oval dominating the sprite
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.48,
		RelativeWidth:  0.65,
		RelativeHeight: 0.55,
		PreferredPixelSize: &PixelDimensions{
			Width:  20,
			Height: 18,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Tail at rear
	template.BodyPartLayout[PartTail] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.82,
		RelativeWidth:  0.15,
		RelativeHeight: 0.18,
		PreferredPixelSize: &PixelDimensions{
			Width:  5,
			Height: 6,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean},
		ZIndex:     8,
		ColorRole:  "secondary",
		Opacity:    0.9,
	}

	// Head — position varies by direction
	headX, headY := 0.50, 0.18
	headRotation := 0.0
	switch direction {
	case DirUp:
		headX, headY = 0.50, 0.15
	case DirDown:
		headX, headY = 0.50, 0.78
		template.BodyPartLayout[PartTail] = PartSpec{
			RelativeX: 0.5, RelativeY: 0.15,
			RelativeWidth: 0.15, RelativeHeight: 0.18,
			PreferredPixelSize: &PixelDimensions{Width: 5, Height: 6},
			ShapeTypes:         []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean},
			ZIndex:             8, ColorRole: "secondary", Opacity: 0.9,
		}
	case DirLeft:
		headX, headY = 0.20, 0.45
		headRotation = 270
		template.BodyPartLayout[PartTail] = PartSpec{
			RelativeX: 0.80, RelativeY: 0.45,
			RelativeWidth: 0.18, RelativeHeight: 0.15,
			PreferredPixelSize: &PixelDimensions{Width: 6, Height: 5},
			ShapeTypes:         []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean},
			ZIndex:             8, ColorRole: "secondary", Opacity: 0.9,
		}
	case DirRight:
		headX, headY = 0.80, 0.45
		headRotation = 90
		template.BodyPartLayout[PartTail] = PartSpec{
			RelativeX: 0.20, RelativeY: 0.45,
			RelativeWidth: 0.18, RelativeHeight: 0.15,
			PreferredPixelSize: &PixelDimensions{Width: 6, Height: 5},
			ShapeTypes:         []shapes.ShapeType{shapes.ShapeCapsule, shapes.ShapeBean},
			ZIndex:             8, ColorRole: "secondary", Opacity: 0.9,
		}
	}

	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.30,
		RelativeHeight: 0.25,
		PreferredPixelSize: &PixelDimensions{
			Width:  10,
			Height: 8,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeWedge, shapes.ShapeCircle},
		ZIndex:     15,
		ColorRole:  "secondary",
		Opacity:    1.0,
		Rotation:   headRotation,
	}

	return template
}

// SerpentineAerialTemplate returns a top-down serpent template (snakes, worms, wyrms).
// From above: sinuous elongated body, wedge-shaped head, tapering tail.
// Direction rotates the whole body orientation.
func SerpentineAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "serpentine_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Serpents cast a thin elongated shadow
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.88,
		RelativeWidth:  0.75,
		RelativeHeight: 0.12,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.30,
	}

	// Vertical orientation (up/down) — elongated vertically
	// Horizontal (left/right) — elongated horizontally
	isVertical := direction == DirUp || direction == DirDown

	var torsoW, torsoH float64
	var torsoPixW, torsoPixH int
	var headX, headY, tailX, tailY float64
	var headRot float64

	if isVertical {
		torsoW, torsoH = 0.25, 0.60
		torsoPixW, torsoPixH = 8, 19
		if direction == DirUp {
			headX, headY = 0.50, 0.12
			tailX, tailY = 0.50, 0.85
		} else {
			headX, headY = 0.50, 0.85
			tailX, tailY = 0.50, 0.12
			headRot = 180
		}
	} else {
		torsoW, torsoH = 0.60, 0.25
		torsoPixW, torsoPixH = 19, 8
		if direction == DirLeft {
			headX, headY = 0.12, 0.50
			tailX, tailY = 0.85, 0.50
			headRot = 270
		} else {
			headX, headY = 0.85, 0.50
			tailX, tailY = 0.12, 0.50
			headRot = 90
		}
	}

	// Tail segment — tapers to a point
	template.BodyPartLayout[PartTail] = PartSpec{
		RelativeX:      tailX,
		RelativeY:      tailY,
		RelativeWidth:  0.12,
		RelativeHeight: 0.15,
		PreferredPixelSize: &PixelDimensions{
			Width:  4,
			Height: 5,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeCapsule},
		ZIndex:     5,
		ColorRole:  "primary",
		Opacity:    0.9,
		Rotation:   headRot + 180, // tail points opposite head
	}

	// Main sinuous body
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.50,
		RelativeY:      0.48,
		RelativeWidth:  torsoW,
		RelativeHeight: torsoH,
		PreferredPixelSize: &PixelDimensions{
			Width:  torsoPixW,
			Height: torsoPixH,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWave, shapes.ShapeCapsule, shapes.ShapeBean},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Wedge-shaped head
	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.22,
		RelativeHeight: 0.18,
		PreferredPixelSize: &PixelDimensions{
			Width:  7,
			Height: 6,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle, shapes.ShapeEllipse},
		ZIndex:     15,
		ColorRole:  "secondary",
		Opacity:    1.0,
		Rotation:   headRot,
	}

	// Eyes on the head for personality
	template.BodyPartLayout[PartEyes] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.15,
		RelativeHeight: 0.08,
		PreferredPixelSize: &PixelDimensions{
			Width:  5,
			Height: 3,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle},
		ZIndex:     16,
		ColorRole:  "accent1",
		Opacity:    1.0,
	}

	return template
}

// ArachnidAerialTemplate returns a top-down arachnid/insect template (spiders, beetles).
// From above: compact round body with 8 legs radiating outward like a starburst.
// This is the most distinctive nonhumanoid silhouette — radial symmetry.
func ArachnidAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "arachnid_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Wide shadow matching leg spread
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.88,
		RelativeWidth:  0.90,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.30,
	}

	// Rear legs — back set of legs, spread wide
	template.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.62,
		RelativeWidth:  0.90,
		RelativeHeight: 0.30,
		PreferredPixelSize: &PixelDimensions{
			Width:  28,
			Height: 10,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeLightning, shapes.ShapeStar},
		ZIndex:     3,
		ColorRole:  "primary",
		Opacity:    0.85,
	}

	// Front legs — forward-reaching set
	template.BodyPartLayout[PartArms] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.35,
		RelativeWidth:  0.85,
		RelativeHeight: 0.28,
		PreferredPixelSize: &PixelDimensions{
			Width:  27,
			Height: 9,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeLightning, shapes.ShapeStar},
		ZIndex:     4,
		ColorRole:  "primary",
		Opacity:    0.85,
		Rotation:   15, // slight splay
	}

	// Abdomen — large rear segment
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.58,
		RelativeWidth:  0.45,
		RelativeHeight: 0.35,
		PreferredPixelSize: &PixelDimensions{
			Width:  14,
			Height: 11,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeCircle, shapes.ShapeOrganic},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Cephalothorax / head — smaller front segment
	headX := 0.50
	headY := 0.28
	switch direction {
	case DirUp:
		headY = 0.25
	case DirDown:
		headY = 0.68
	case DirLeft:
		headX = 0.30
		headY = 0.45
	case DirRight:
		headX = 0.70
		headY = 0.45
	}

	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.28,
		RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{
			Width:  9,
			Height: 7,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex:     12,
		ColorRole:  "secondary",
		Opacity:    1.0,
	}

	// Eyes — multiple eyes visible from above
	template.BodyPartLayout[PartEyes] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY - 0.02,
		RelativeWidth:  0.20,
		RelativeHeight: 0.08,
		PreferredPixelSize: &PixelDimensions{
			Width:  6,
			Height: 3,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle},
		ZIndex:     13,
		ColorRole:  "accent1",
		Opacity:    1.0,
	}

	return template
}

// BlobAerialTemplate returns a top-down amorphous blob template (slimes, oozes).
// From above: irregular circular mass, slight translucency, internal nucleus visible.
// Direction creates a subtle directional bulge indicating movement.
func BlobAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "blob_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Blob casts a wide, soft shadow
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.85,
		RelativeWidth:  0.80,
		RelativeHeight: 0.20,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeOrganic},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.40,
	}

	// Main blob mass — large, organic, slightly translucent
	blobX := 0.50
	blobY := 0.48
	switch direction {
	case DirUp:
		blobY = 0.44
	case DirDown:
		blobY = 0.52
	case DirLeft:
		blobX = 0.45
	case DirRight:
		blobX = 0.55
	}

	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      blobX,
		RelativeY:      blobY,
		RelativeWidth:  0.75,
		RelativeHeight: 0.70,
		PreferredPixelSize: &PixelDimensions{
			Width:  24,
			Height: 22,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeCircle, shapes.ShapeBean},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    0.80, // translucent mass
	}

	// Internal nucleus / core — visible through translucent body
	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      0.50,
		RelativeY:      0.42,
		RelativeWidth:  0.30,
		RelativeHeight: 0.25,
		PreferredPixelSize: &PixelDimensions{
			Width:  10,
			Height: 8,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeOrganic},
		ZIndex:     12,
		ColorRole:  "accent1",
		Opacity:    0.70,
	}

	// Surface highlight — specular reflection dot on top
	template.BodyPartLayout[PartEyes] = PartSpec{
		RelativeX:      0.42,
		RelativeY:      0.35,
		RelativeWidth:  0.12,
		RelativeHeight: 0.10,
		PreferredPixelSize: &PixelDimensions{
			Width:  4,
			Height: 3,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse},
		ZIndex:     14,
		ColorRole:  "highlight",
		Opacity:    0.60,
	}

	return template
}

// FlyingAerialTemplate returns a top-down winged creature template (dragons, bats, birds).
// From above: wide wingspan dominates, streamlined body center, head at front.
// Wings use dedicated PartWings for left+right symmetry.
func FlyingAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "flying_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Small shadow far below — creature is airborne
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.92,
		RelativeWidth:  0.35,
		RelativeHeight: 0.10,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.20,
	}

	// Wings — massive spread, the defining feature from above
	// Left wing
	leftWingX, leftWingY := 0.22, 0.48
	rightWingX, rightWingY := 0.78, 0.48
	wingRot := 0.0
	switch direction {
	case DirUp:
		leftWingX, leftWingY = 0.22, 0.50
		rightWingX, rightWingY = 0.78, 0.50
	case DirDown:
		leftWingX, leftWingY = 0.22, 0.45
		rightWingX, rightWingY = 0.78, 0.45
	case DirLeft:
		leftWingX, leftWingY = 0.48, 0.22
		rightWingX, rightWingY = 0.48, 0.78
		wingRot = 90
	case DirRight:
		leftWingX, leftWingY = 0.52, 0.22
		rightWingX, rightWingY = 0.52, 0.78
		wingRot = 90
	}

	template.BodyPartLayout[PartArms] = PartSpec{
		RelativeX:      leftWingX,
		RelativeY:      leftWingY,
		RelativeWidth:  0.35,
		RelativeHeight: 0.50,
		PreferredPixelSize: &PixelDimensions{
			Width:  11,
			Height: 16,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle, shapes.ShapeBean},
		ZIndex:     8,
		ColorRole:  "secondary",
		Opacity:    0.90,
		Rotation:   wingRot,
	}

	template.BodyPartLayout[PartWings] = PartSpec{
		RelativeX:      rightWingX,
		RelativeY:      rightWingY,
		RelativeWidth:  0.35,
		RelativeHeight: 0.50,
		PreferredPixelSize: &PixelDimensions{
			Width:  11,
			Height: 16,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeTriangle, shapes.ShapeBean},
		ZIndex:     8,
		ColorRole:  "secondary",
		Opacity:    0.90,
		Rotation:   wingRot,
	}

	// Streamlined body center
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.50,
		RelativeY:      0.48,
		RelativeWidth:  0.25,
		RelativeHeight: 0.50,
		PreferredPixelSize: &PixelDimensions{
			Width:  8,
			Height: 16,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeBean, shapes.ShapeCapsule},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Tail behind body
	tailX, tailY := 0.50, 0.82
	tailRot := 0.0
	switch direction {
	case DirDown:
		tailX, tailY = 0.50, 0.15
		tailRot = 180
	case DirLeft:
		tailX, tailY = 0.82, 0.50
		tailRot = 90
	case DirRight:
		tailX, tailY = 0.15, 0.50
		tailRot = 270
	}
	template.BodyPartLayout[PartTail] = PartSpec{
		RelativeX:      tailX,
		RelativeY:      tailY,
		RelativeWidth:  0.15,
		RelativeHeight: 0.20,
		PreferredPixelSize: &PixelDimensions{
			Width:  5,
			Height: 6,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeWedge, shapes.ShapeCapsule},
		ZIndex:     7,
		ColorRole:  "primary",
		Opacity:    0.90,
		Rotation:   tailRot,
	}

	// Head at front
	headX, headY := 0.50, 0.18
	headRot := 0.0
	switch direction {
	case DirDown:
		headX, headY = 0.50, 0.78
		headRot = 180
	case DirLeft:
		headX, headY = 0.18, 0.50
		headRot = 270
	case DirRight:
		headX, headY = 0.82, 0.50
		headRot = 90
	}

	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.22,
		RelativeHeight: 0.18,
		PreferredPixelSize: &PixelDimensions{
			Width:  7,
			Height: 6,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeEllipse, shapes.ShapeWedge},
		ZIndex:     15,
		ColorRole:  "secondary",
		Opacity:    1.0,
		Rotation:   headRot,
	}

	return template
}

// MechanicalAerialTemplate returns a top-down mechanical/construct template (robots, golems).
// From above: boxy chassis with angular limbs, sensor head, geometric shapes.
// Proportions: head ~20%, chassis ~55%, legs ~25% (compact from above).
func MechanicalAerialTemplate(direction Direction) AnatomicalTemplate {
	template := AnatomicalTemplate{
		Name:           "mechanical_aerial_" + string(direction),
		BodyPartLayout: make(map[BodyPart]PartSpec),
	}

	// Hard-edged shadow
	template.BodyPartLayout[PartShadow] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.90,
		RelativeWidth:  0.70,
		RelativeHeight: 0.15,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeEllipse, shapes.ShapeRectangle},
		ZIndex:         0,
		ColorRole:      "shadow",
		Opacity:        0.35,
	}

	// Mechanical legs / treads — visible as stubs beneath chassis
	template.BodyPartLayout[PartLegs] = PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.80,
		RelativeWidth:  0.60,
		RelativeHeight: 0.20,
		PreferredPixelSize: &PixelDimensions{
			Width:  19,
			Height: 6,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule},
		ZIndex:     5,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Main chassis — boxy, geometric, dominates the sprite
	template.BodyPartLayout[PartTorso] = PartSpec{
		RelativeX:      0.50,
		RelativeY:      0.48,
		RelativeWidth:  0.65,
		RelativeHeight: 0.50,
		PreferredPixelSize: &PixelDimensions{
			Width:  20,
			Height: 16,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
		ZIndex:     10,
		ColorRole:  "primary",
		Opacity:    1.0,
	}

	// Mechanical arms / weapon mounts
	armX := 0.50
	armY := 0.50
	armRot := 0.0
	switch direction {
	case DirUp:
		armY = 0.52
	case DirDown:
		armY = 0.48
	case DirLeft:
		armX = 0.30
		armRot = 270
	case DirRight:
		armX = 0.70
		armRot = 90
	}
	template.BodyPartLayout[PartArms] = PartSpec{
		RelativeX:      armX,
		RelativeY:      armY,
		RelativeWidth:  0.80,
		RelativeHeight: 0.28,
		PreferredPixelSize: &PixelDimensions{
			Width:  25,
			Height: 9,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeCapsule, shapes.ShapeGear},
		ZIndex:     8,
		ColorRole:  "secondary",
		Opacity:    1.0,
		Rotation:   armRot,
	}

	// Sensor head — small, geometric, on top
	headX, headY := 0.50, 0.20
	switch direction {
	case DirDown:
		headY = 0.75
	case DirLeft:
		headX = 0.25
		headY = 0.45
	case DirRight:
		headX = 0.75
		headY = 0.45
	}

	template.BodyPartLayout[PartHead] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY,
		RelativeWidth:  0.28,
		RelativeHeight: 0.22,
		PreferredPixelSize: &PixelDimensions{
			Width:  9,
			Height: 7,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeRectangle, shapes.ShapeHexagon, shapes.ShapeOctagon},
		ZIndex:     15,
		ColorRole:  "accent1",
		Opacity:    1.0,
	}

	// Sensor eye / antenna on head
	template.BodyPartLayout[PartEyes] = PartSpec{
		RelativeX:      headX,
		RelativeY:      headY - 0.02,
		RelativeWidth:  0.10,
		RelativeHeight: 0.08,
		PreferredPixelSize: &PixelDimensions{
			Width:  3,
			Height: 3,
		},
		ShapeTypes: []shapes.ShapeType{shapes.ShapeCircle, shapes.ShapeCrystal},
		ZIndex:     16,
		ColorRole:  "accent1",
		Opacity:    1.0,
	}

	return template
}

// UndeadAerialTemplate returns a top-down undead template (skeletons, ghosts, zombies).
// Based on humanoid aerial proportions but with ghostly/skeletal modifications.
// Ghosts get translucency, skeletons get angular shapes, zombies get organic decay.
func UndeadAerialTemplate(direction Direction) AnatomicalTemplate {
	// Start from humanoid aerial as base
	template := HumanoidAerialTemplate(direction)
	template.Name = "undead_aerial_" + string(direction)

	// Make torso more gaunt / irregular
	if torso, ok := template.BodyPartLayout[PartTorso]; ok {
		torso.RelativeWidth = 0.65
		torso.Opacity = 0.85
		torso.ShapeTypes = []shapes.ShapeType{shapes.ShapeOrganic, shapes.ShapeBean, shapes.ShapeSkull}
		template.BodyPartLayout[PartTorso] = torso
	}

	// Head uses skull shape
	if head, ok := template.BodyPartLayout[PartHead]; ok {
		head.ShapeTypes = []shapes.ShapeType{shapes.ShapeSkull, shapes.ShapeCircle}
		head.Opacity = 0.90
		template.BodyPartLayout[PartHead] = head
	}

	// Ghostly arms
	if arms, ok := template.BodyPartLayout[PartArms]; ok {
		arms.Opacity = 0.70
		arms.ShapeTypes = []shapes.ShapeType{shapes.ShapeWave, shapes.ShapeCapsule}
		template.BodyPartLayout[PartArms] = arms
	}

	return template
}

// SelectNonhumanoidAerialTemplate selects the appropriate aerial template for
// a nonhumanoid entity type with optional genre variation.
func SelectNonhumanoidAerialTemplate(entityType, genre string, direction Direction) AnatomicalTemplate {
	var template AnatomicalTemplate

	switch entityType {
	case "quadruped", "wolf", "bear", "animal", "beast", "horse":
		template = QuadrupedAerialTemplate(direction)
	case "blob", "slime", "amoeba", "ooze":
		template = BlobAerialTemplate(direction)
	case "mechanical", "robot", "golem", "construct", "android":
		template = MechanicalAerialTemplate(direction)
	case "flying", "bird", "dragon", "bat", "wyvern":
		template = FlyingAerialTemplate(direction)
	case "serpentine", "snake", "worm", "tentacle", "wyrm":
		template = SerpentineAerialTemplate(direction)
	case "arachnid", "spider", "insect", "beetle", "scorpion":
		template = ArachnidAerialTemplate(direction)
	case "undead", "skeleton", "ghost", "zombie", "lich":
		template = UndeadAerialTemplate(direction)
	default:
		// Unknown nonhumanoid — use quadruped as safest generic creature
		template = QuadrupedAerialTemplate(direction)
	}

	// Apply genre variation if specified
	switch genre {
	case "fantasy":
		return ApplyFantasyVariation(template)
	case "scifi", "sci-fi":
		return ApplySciFiVariation(template)
	case "horror":
		return ApplyHorrorVariation(template)
	case "cyberpunk":
		return ApplyCyberpunkVariation(template)
	case "postapoc", "post-apocalyptic":
		return ApplyPostApocVariation(template)
	}

	return template
}
