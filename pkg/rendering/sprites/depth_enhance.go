// Package sprites provides form-aware volumetric depth enhancement for
// procedurally generated top-down entity sprites. ApplyDepthEnhancement
// analyzes the sprite's opaque regions and applies 3D-form-appropriate
// shading — spherical for heads, cylindrical for torsos, tubular for
// limbs — creating the illusion of volume when viewed from directly above.
// Contact shadows are added between overlapping body part regions.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// DepthFormType identifies the 3D form used for volumetric shading.
type DepthFormType int

const (
	// FormSphere represents a spherical form (head, eyes, round shields).
	FormSphere DepthFormType = iota
	// FormCylinder represents a vertical cylinder (torso seen from above).
	FormCylinder
	// FormTube represents a horizontal tube (arms, legs).
	FormTube
	// FormFlat represents a flat surface (shadows, ground contact).
	FormFlat
	// FormDome represents a low dome (helmets, shoulder pads).
	FormDome
	// FormMembrane represents a thin membrane (wings, capes).
	FormMembrane
)

// DepthEnhanceConfig controls the volumetric depth enhancement pass.
type DepthEnhanceConfig struct {
	// LightAzimuth is the horizontal light angle in radians (0 = north/up).
	LightAzimuth float64
	// LightElevation is the vertical light angle in radians (π/4 = 45°).
	LightElevation float64
	// SpecularPower controls specular highlight tightness (higher = tighter).
	SpecularPower float64
	// SpecularIntensity scales the specular highlight brightness (0-1).
	SpecularIntensity float64
	// DiffuseStrength scales diffuse lighting contribution (0-1).
	DiffuseStrength float64
	// ContactShadowStrength darkens pixels at body part overlap zones (0-1).
	ContactShadowStrength float64
	// SubsurfaceStrength adds warm edge glow to simulate skin translucency (0-1).
	SubsurfaceStrength float64
	// Seed for deterministic variation.
	Seed int64
}

// DefaultDepthEnhanceConfig returns sensible defaults for top-down sprites.
func DefaultDepthEnhanceConfig(seed int64) DepthEnhanceConfig {
	return DepthEnhanceConfig{
		LightAzimuth:          -math.Pi / 4, // NW light (315°)
		LightElevation:        math.Pi / 3,  // 60° from horizon (high overhead)
		SpecularPower:         16.0,
		SpecularIntensity:     0.35,
		DiffuseStrength:       0.30,
		ContactShadowStrength: 0.20,
		SubsurfaceStrength:    0.08,
		Seed:                  seed,
	}
}

// DepthZone describes a rectangular region of the sprite and its 3D form.
type DepthZone struct {
	X, Y, W, H int
	Form       DepthFormType
	// BaseHeight is the base height above ground (0.0-1.0). Higher parts cast shadows.
	BaseHeight float64
}

// InferDepthZones partitions the sprite into depth zones based on opacity
// clusters. For 32×32 sprites this produces head/torso/limb zones with
// appropriate 3D form types. The result is deterministic for a given seed.
func InferDepthZones(img *image.RGBA, seed int64) []DepthZone {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return nil
	}

	// Find bounding box of opaque pixels
	minX, minY, maxX, maxY := w, h, 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.Pix[(y*w+x)*4+3] > 20 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < minX || maxY < minY {
		return nil
	}

	bw := maxX - minX + 1
	bh := maxY - minY + 1

	// Compute horizontal density per row to detect major body segments
	rowDensity := make([]float64, bh)
	for y := 0; y < bh; y++ {
		count := 0
		for x := 0; x < bw; x++ {
			if img.Pix[((minY+y)*w+(minX+x))*4+3] > 20 {
				count++
			}
		}
		rowDensity[y] = float64(count) / float64(bw)
	}

	// Split into top (head ~35%), middle (torso ~50%), bottom (legs ~15%)
	headEnd := bh * 35 / 100
	torsoEnd := bh * 85 / 100
	if headEnd < 2 {
		headEnd = 2
	}
	if torsoEnd <= headEnd {
		torsoEnd = headEnd + 1
	}

	// Find per-zone horizontal extent
	headBounds := zoneBounds(img, minX, minY, minX+bw, minY+headEnd, w)
	torsoBounds := zoneBounds(img, minX, minY+headEnd, minX+bw, minY+torsoEnd, w)
	legsBounds := zoneBounds(img, minX, minY+torsoEnd, minX+bw, minY+bh, w)

	zones := make([]DepthZone, 0, 3)

	if headBounds.W > 0 && headBounds.H > 0 {
		zones = append(zones, DepthZone{
			X: headBounds.X, Y: headBounds.Y, W: headBounds.W, H: headBounds.H,
			Form: FormSphere, BaseHeight: 0.9,
		})
	}
	if torsoBounds.W > 0 && torsoBounds.H > 0 {
		zones = append(zones, DepthZone{
			X: torsoBounds.X, Y: torsoBounds.Y, W: torsoBounds.W, H: torsoBounds.H,
			Form: FormCylinder, BaseHeight: 0.5,
		})
	}
	if legsBounds.W > 0 && legsBounds.H > 0 {
		zones = append(zones, DepthZone{
			X: legsBounds.X, Y: legsBounds.Y, W: legsBounds.W, H: legsBounds.H,
			Form: FormTube, BaseHeight: 0.15,
		})
	}

	return zones
}

type rect struct{ X, Y, W, H int }

func zoneBounds(img *image.RGBA, x0, y0, x1, y1, stride int) rect {
	zMinX, zMinY, zMaxX, zMaxY := x1, y1, x0, y0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if img.Pix[(y*stride+x)*4+3] > 20 {
				if x < zMinX {
					zMinX = x
				}
				if x > zMaxX {
					zMaxX = x
				}
				if y < zMinY {
					zMinY = y
				}
				if y > zMaxY {
					zMaxY = y
				}
			}
		}
	}
	if zMaxX < zMinX || zMaxY < zMinY {
		return rect{}
	}
	return rect{X: zMinX, Y: zMinY, W: zMaxX - zMinX + 1, H: zMaxY - zMinY + 1}
}

// ApplyDepthEnhancement applies form-aware volumetric shading to the sprite.
// Zones are inferred automatically from the sprite's opaque regions.
// The image is modified in place. Returns the number of zones processed.
func ApplyDepthEnhancement(img *image.RGBA, cfg DepthEnhanceConfig) int {
	zones := InferDepthZones(img, cfg.Seed)
	if len(zones) == 0 {
		return 0
	}

	// Build a height map across the entire sprite
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	heightMap := make([]float64, w*h)

	for _, zone := range zones {
		applyZoneHeightMap(heightMap, zone, w)
	}

	// Compute light direction vector in 3D
	cosAz := math.Cos(cfg.LightAzimuth)
	sinAz := math.Sin(cfg.LightAzimuth)
	cosEl := math.Cos(cfg.LightElevation)
	sinEl := math.Sin(cfg.LightElevation)
	lightDir := [3]float64{sinAz * cosEl, -cosAz * cosEl, sinEl}

	rng := rand.New(rand.NewSource(cfg.Seed))
	_ = rng // reserved for future dithering variation

	// Apply lighting per pixel
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if img.Pix[idx+3] < 20 {
				continue
			}

			px := float64(x)
			py := float64(y)

			// Find which zone this pixel belongs to
			zone := findZone(zones, x, y)
			if zone == nil {
				continue
			}

			// Compute surface normal for this pixel based on zone form
			normal := computeFormNormal(zone, px, py)

			// Diffuse: N · L
			diffuse := normal[0]*lightDir[0] + normal[1]*lightDir[1] + normal[2]*lightDir[2]
			if diffuse < 0 {
				diffuse = 0
			}

			// Specular: (R · V)^power where V = (0,0,1) (camera is straight down)
			// R = 2*(N·L)*N - L
			ndotl := normal[0]*lightDir[0] + normal[1]*lightDir[1] + normal[2]*lightDir[2]
			if ndotl < 0 {
				ndotl = 0
			}
			reflZ := 2*ndotl*normal[2] - lightDir[2]
			specular := 0.0
			if reflZ > 0 {
				specular = math.Pow(reflZ, cfg.SpecularPower)
			}

			// Contact shadow: darken pixels near zone boundaries where higher zones overlap
			contactShadow := computeContactShadow(zones, zone, x, y, heightMap, w)

			// Subsurface scattering approximation: warm glow at edges of skin-like parts
			sss := 0.0
			if zone.Form == FormSphere || zone.Form == FormCylinder {
				edgeness := 1.0 - normal[2] // how much the surface faces sideways
				sss = edgeness * cfg.SubsurfaceStrength
			}

			// Apply to pixel color
			r := float64(img.Pix[idx])
			g := float64(img.Pix[idx+1])
			b := float64(img.Pix[idx+2])

			// Diffuse contribution
			diffFactor := cfg.DiffuseStrength * diffuse
			r += diffFactor * 40.0
			g += diffFactor * 38.0
			b += diffFactor * 35.0

			// Specular highlight (white)
			specFactor := cfg.SpecularIntensity * specular
			r += specFactor * 200.0
			g += specFactor * 195.0
			b += specFactor * 185.0

			// Contact shadow (darken)
			if contactShadow > 0 {
				shadowFactor := 1.0 - contactShadow*cfg.ContactShadowStrength
				r *= shadowFactor
				g *= shadowFactor
				b *= shadowFactor
			}

			// Subsurface scattering (warm reddish glow at edges)
			if sss > 0 {
				r += sss * 30.0
				g += sss * 8.0
				b += sss * 3.0
			}

			img.Pix[idx] = depthClamp(r)
			img.Pix[idx+1] = depthClamp(g)
			img.Pix[idx+2] = depthClamp(b)
		}
	}

	return len(zones)
}

// applyZoneHeightMap fills the height map for pixels in a zone.
func applyZoneHeightMap(heightMap []float64, zone DepthZone, stride int) {
	for y := zone.Y; y < zone.Y+zone.H; y++ {
		for x := zone.X; x < zone.X+zone.W; x++ {
			h := computeHeight(zone, float64(x), float64(y))
			idx := y*stride + x
			if idx >= 0 && idx < len(heightMap) && h > heightMap[idx] {
				heightMap[idx] = h
			}
		}
	}
}

// computeHeight returns the height above ground for a pixel in a zone.
func computeHeight(zone DepthZone, px, py float64) float64 {
	cx := float64(zone.X) + float64(zone.W)/2.0
	cy := float64(zone.Y) + float64(zone.H)/2.0
	rx := float64(zone.W) / 2.0
	ry := float64(zone.H) / 2.0
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}

	nx := (px - cx) / rx
	ny := (py - cy) / ry

	switch zone.Form {
	case FormSphere:
		r2 := nx*nx + ny*ny
		if r2 >= 1 {
			return zone.BaseHeight
		}
		return zone.BaseHeight + math.Sqrt(1-r2)*0.4

	case FormCylinder:
		// Cylinder runs top-to-bottom; height depends on horizontal distance from center
		if math.Abs(nx) >= 1 {
			return zone.BaseHeight
		}
		return zone.BaseHeight + math.Sqrt(1-nx*nx)*0.25

	case FormTube:
		// Tube runs left-to-right; height depends on vertical distance
		if math.Abs(ny) >= 1 {
			return zone.BaseHeight
		}
		return zone.BaseHeight + math.Sqrt(1-ny*ny)*0.15

	case FormDome:
		r2 := nx*nx + ny*ny
		if r2 >= 1 {
			return zone.BaseHeight
		}
		return zone.BaseHeight + math.Sqrt(1-r2)*0.3

	case FormMembrane:
		return zone.BaseHeight + 0.02

	default:
		return zone.BaseHeight
	}
}

// computeFormNormal returns the surface normal (x,y,z) for the 3D form at a pixel.
func computeFormNormal(zone *DepthZone, px, py float64) [3]float64 {
	cx := float64(zone.X) + float64(zone.W)/2.0
	cy := float64(zone.Y) + float64(zone.H)/2.0
	rx := float64(zone.W) / 2.0
	ry := float64(zone.H) / 2.0
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}

	nx := (px - cx) / rx
	ny := (py - cy) / ry

	switch zone.Form {
	case FormSphere:
		r2 := nx*nx + ny*ny
		if r2 >= 1 {
			// Edge pixel — normal points outward horizontally
			len := math.Sqrt(r2)
			return [3]float64{nx / len, ny / len, 0.1}
		}
		nz := math.Sqrt(1 - r2)
		return normalize3(nx, ny, nz)

	case FormCylinder:
		// Cylinder axis is vertical (Y); normal radiates horizontally (X)
		if math.Abs(nx) >= 1 {
			sign := 1.0
			if nx < 0 {
				sign = -1
			}
			return [3]float64{sign, 0, 0.1}
		}
		nz := math.Sqrt(1 - nx*nx)
		return normalize3(nx, 0, nz)

	case FormTube:
		if math.Abs(ny) >= 1 {
			sign := 1.0
			if ny < 0 {
				sign = -1
			}
			return [3]float64{0, sign, 0.1}
		}
		nz := math.Sqrt(1 - ny*ny)
		return normalize3(0, ny, nz)

	case FormDome:
		r2 := nx*nx + ny*ny
		if r2 >= 1 {
			len := math.Sqrt(r2)
			return [3]float64{nx / len, ny / len, 0.15}
		}
		nz := math.Sqrt(1-r2) * 0.6
		return normalize3(nx, ny, nz)

	case FormMembrane:
		return [3]float64{0, 0, 1} // nearly flat

	default:
		return [3]float64{0, 0, 1}
	}
}

func normalize3(x, y, z float64) [3]float64 {
	length := math.Sqrt(x*x + y*y + z*z)
	if length < 0.001 {
		return [3]float64{0, 0, 1}
	}
	return [3]float64{x / length, y / length, z / length}
}

// computeContactShadow returns shadow intensity (0-1) for a pixel based on
// proximity to higher-elevation zones.
func computeContactShadow(zones []DepthZone, current *DepthZone, x, y int, heightMap []float64, stride int) float64 {
	if current == nil {
		return 0
	}

	// Check if this pixel is near the boundary of the current zone
	dx := distToEdge(x, current.X, current.X+current.W)
	dy := distToEdge(y, current.Y, current.Y+current.H)
	edgeDist := math.Min(float64(dx), float64(dy))

	maxShadowDist := 3.0
	if edgeDist >= maxShadowDist {
		return 0
	}

	// Check if any adjacent zone is higher
	for i := range zones {
		other := &zones[i]
		if other == current {
			continue
		}
		if other.BaseHeight <= current.BaseHeight {
			continue
		}

		// Distance from pixel to other zone
		nearX := clampInt(x, other.X, other.X+other.W-1)
		nearY := clampInt(y, other.Y, other.Y+other.H-1)
		dist := math.Sqrt(float64((x-nearX)*(x-nearX) + (y-nearY)*(y-nearY)))

		if dist < maxShadowDist {
			return (1.0 - dist/maxShadowDist) * 0.5
		}
	}

	return 0
}

func distToEdge(v, lo, hi int) int {
	dLo := v - lo
	dHi := hi - 1 - v
	if dLo < dHi {
		return dLo
	}
	return dHi
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func findZone(zones []DepthZone, x, y int) *DepthZone {
	for i := range zones {
		z := &zones[i]
		if x >= z.X && x < z.X+z.W && y >= z.Y && y < z.Y+z.H {
			return z
		}
	}
	return nil
}

func depthClamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// ApplyDepthEnhancementForCreature applies volumetric shading with
// creature-appropriate zone forms. Nonhumanoid creatures use different
// body region splits and 3D forms (e.g. quadrupeds get a flat oval body,
// serpents get a long tube, arachnids get a dome center with radial legs).
func ApplyDepthEnhancementForCreature(img *image.RGBA, form string, cfg DepthEnhanceConfig) int {
	switch form {
	case "quadruped":
		return applyQuadrupedDepth(img, cfg)
	case "serpentine":
		return applySerpentDepth(img, cfg)
	case "arachnid":
		return applyArachnidDepth(img, cfg)
	case "flying":
		return applyFlyingDepth(img, cfg)
	case "blob":
		return applyBlobDepth(img, cfg)
	default:
		return ApplyDepthEnhancement(img, cfg)
	}
}

func applyQuadrupedDepth(img *image.RGBA, cfg DepthEnhanceConfig) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bb := opaqueBounds(img, w, h)
	if bb.W == 0 {
		return 0
	}

	// Quadruped from above: large oval body (dome), small head at top
	headH := bb.H * 25 / 100
	zones := []DepthZone{
		{X: bb.X + bb.W/4, Y: bb.Y, W: bb.W / 2, H: headH, Form: FormSphere, BaseHeight: 0.8},
		{X: bb.X, Y: bb.Y + headH, W: bb.W, H: bb.H - headH, Form: FormDome, BaseHeight: 0.5},
	}
	return applyZonedDepth(img, zones, cfg, w, h)
}

func applySerpentDepth(img *image.RGBA, cfg DepthEnhanceConfig) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bb := opaqueBounds(img, w, h)
	if bb.W == 0 {
		return 0
	}

	// Serpent: one long tube
	zones := []DepthZone{
		{X: bb.X, Y: bb.Y, W: bb.W, H: bb.H, Form: FormTube, BaseHeight: 0.3},
	}
	return applyZonedDepth(img, zones, cfg, w, h)
}

func applyArachnidDepth(img *image.RGBA, cfg DepthEnhanceConfig) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bb := opaqueBounds(img, w, h)
	if bb.W == 0 {
		return 0
	}

	// Arachnid: central dome body
	inset := bb.W / 4
	zones := []DepthZone{
		{X: bb.X + inset, Y: bb.Y + inset, W: bb.W - 2*inset, H: bb.H - 2*inset, Form: FormDome, BaseHeight: 0.6},
		{X: bb.X, Y: bb.Y, W: bb.W, H: bb.H, Form: FormFlat, BaseHeight: 0.1},
	}
	return applyZonedDepth(img, zones, cfg, w, h)
}

func applyFlyingDepth(img *image.RGBA, cfg DepthEnhanceConfig) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bb := opaqueBounds(img, w, h)
	if bb.W == 0 {
		return 0
	}

	// Flying creature: narrow body center, wide membrane wings
	bodyW := bb.W / 3
	bodyX := bb.X + (bb.W-bodyW)/2
	zones := []DepthZone{
		{X: bodyX, Y: bb.Y, W: bodyW, H: bb.H, Form: FormCylinder, BaseHeight: 0.7},
		{X: bb.X, Y: bb.Y, W: bb.W, H: bb.H, Form: FormMembrane, BaseHeight: 0.5},
	}
	return applyZonedDepth(img, zones, cfg, w, h)
}

func applyBlobDepth(img *image.RGBA, cfg DepthEnhanceConfig) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bb := opaqueBounds(img, w, h)
	if bb.W == 0 {
		return 0
	}

	// Blob: one large sphere
	zones := []DepthZone{
		{X: bb.X, Y: bb.Y, W: bb.W, H: bb.H, Form: FormSphere, BaseHeight: 0.4},
	}
	return applyZonedDepth(img, zones, cfg, w, h)
}

func opaqueBounds(img *image.RGBA, w, h int) rect {
	minX, minY, maxX, maxY := w, h, 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.Pix[(y*w+x)*4+3] > 20 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < minX {
		return rect{}
	}
	return rect{X: minX, Y: minY, W: maxX - minX + 1, H: maxY - minY + 1}
}

func applyZonedDepth(img *image.RGBA, zones []DepthZone, cfg DepthEnhanceConfig, w, h int) int {
	heightMap := make([]float64, w*h)
	for _, zone := range zones {
		applyZoneHeightMap(heightMap, zone, w)
	}

	cosAz := math.Cos(cfg.LightAzimuth)
	sinAz := math.Sin(cfg.LightAzimuth)
	cosEl := math.Cos(cfg.LightElevation)
	sinEl := math.Sin(cfg.LightElevation)
	lightDir := [3]float64{sinAz * cosEl, -cosAz * cosEl, sinEl}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if img.Pix[idx+3] < 20 {
				continue
			}

			zone := findZone(zones, x, y)
			if zone == nil {
				continue
			}

			normal := computeFormNormal(zone, float64(x), float64(y))

			ndotl := normal[0]*lightDir[0] + normal[1]*lightDir[1] + normal[2]*lightDir[2]
			if ndotl < 0 {
				ndotl = 0
			}

			reflZ := 2*ndotl*normal[2] - lightDir[2]
			specular := 0.0
			if reflZ > 0 {
				specular = math.Pow(reflZ, cfg.SpecularPower)
			}

			contactShadow := computeContactShadow(zones, zone, x, y, heightMap, w)

			sss := 0.0
			if zone.Form == FormSphere || zone.Form == FormDome {
				edgeness := 1.0 - normal[2]
				sss = edgeness * cfg.SubsurfaceStrength
			}

			r := float64(img.Pix[idx])
			g := float64(img.Pix[idx+1])
			b := float64(img.Pix[idx+2])

			diffFactor := cfg.DiffuseStrength * ndotl
			r += diffFactor * 40.0
			g += diffFactor * 38.0
			b += diffFactor * 35.0

			specFactor := cfg.SpecularIntensity * specular
			r += specFactor * 200.0
			g += specFactor * 195.0
			b += specFactor * 185.0

			if contactShadow > 0 {
				sf := 1.0 - contactShadow*cfg.ContactShadowStrength
				r *= sf
				g *= sf
				b *= sf
			}

			if sss > 0 {
				r += sss * 30.0
				g += sss * 8.0
				b += sss * 3.0
			}

			img.Pix[idx] = depthClamp(r)
			img.Pix[idx+1] = depthClamp(g)
			img.Pix[idx+2] = depthClamp(b)
		}
	}
	return len(zones)
}

// MakeTestSprite creates a simple test sprite with head/torso/leg-like regions.
// Exported for use in tests.
func MakeTestSprite(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Head: upper 35%, centered oval
	depthDrawFilledOval(img, w/4, 1, w/2, h*35/100-2, color.RGBA{R: 200, G: 160, B: 130, A: 255})
	// Torso: middle 50%
	depthDrawFilledRect(img, w/4-1, h*35/100, w/2+2, h*50/100, color.RGBA{R: 100, G: 80, B: 150, A: 255})
	// Legs: bottom 15%
	depthDrawFilledRect(img, w/3, h*85/100, w/3, h*15/100, color.RGBA{R: 80, G: 70, B: 60, A: 255})

	return img
}

func depthDrawFilledOval(img *image.RGBA, ox, oy, ow, oh int, c color.RGBA) {
	cx := float64(ox) + float64(ow)/2.0
	cy := float64(oy) + float64(oh)/2.0
	rx := float64(ow) / 2.0
	ry := float64(oh) / 2.0
	for y := oy; y < oy+oh; y++ {
		for x := ox; x < ox+ow; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			if dx*dx+dy*dy <= 1.0 {
				img.Set(x, y, c)
			}
		}
	}
}

func depthDrawFilledRect(img *image.RGBA, ox, oy, ow, oh int, c color.RGBA) {
	for y := oy; y < oy+oh; y++ {
		for x := ox; x < ox+ow; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}
