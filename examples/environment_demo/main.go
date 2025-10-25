// Package main provides an interactive demonstration of Phase 5 environment visual enhancements.
//
// This demo showcases:
//   - Tile variation system with multiple variations per tile type
//   - Environmental objects (furniture, decorations, obstacles, hazards)
//   - Dynamic lighting with multiple falloff types
//   - Weather particle effects with intensity controls
//
// Controls:
//   - 1-5: Switch genre (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic)
//   - Arrow Keys: Move light source
//   - Space: Toggle weather effects
//   - W/S: Increase/decrease weather intensity
//   - L: Cycle light falloff type
//   - O: Spawn random object at mouse position
//   - R: Reset demo with new seed
//   - F: Toggle FPS display
//   - ESC: Exit
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/procgen/environment"
	"github.com/opd-ai/venture/pkg/rendering/lighting"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	tileSize     = 64
	gridWidth    = screenWidth / tileSize
	gridHeight   = screenHeight / tileSize
)

// Demo represents the environment demonstration state.
type Demo struct {
	// Generators
	tileGen    *tiles.Generator
	objGen     *environment.Generator
	paletteGen *palette.Generator
	lightSys   *lighting.System

	// State
	genre          string
	genreIndex     int
	genres         []string
	seed           int64
	palette        *palette.Palette
	tileVariations *tiles.VariationSet
	objects        []PlacedObject
	lightIndex     int
	falloffType    lighting.FalloffType

	// Weather
	weatherSystem    *particles.WeatherSystem
	weatherEnabled   bool
	weatherType      particles.WeatherType
	weatherIntensity particles.WeatherIntensity

	// Rendering
	tileLayer    *image.RGBA
	objectBuffer *ebiten.Image
	lightBuffer  *image.RGBA

	// UI
	showFPS   bool
	lastFPS   float64
	frameTime time.Duration
}

// PlacedObject represents an object placed in the world.
type PlacedObject struct {
	obj *environment.EnvironmentalObject
	pos image.Point
}

// NewDemo creates a new demo.
func NewDemo() (*Demo, error) {
	d := &Demo{
		tileGen:    tiles.NewGenerator(),
		objGen:     environment.NewGenerator(),
		paletteGen: palette.NewGenerator(),
		lightSys: lighting.NewSystemWithConfig(lighting.LightingConfig{
			AmbientColor:     color.RGBA{30, 30, 40, 255},
			AmbientIntensity: 0.2,
			MaxLights:        10,
		}),
		genres:           []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"},
		genreIndex:       0,
		seed:             time.Now().UnixNano(),
		lightIndex:       0,
		falloffType:      lighting.FalloffQuadratic,
		weatherEnabled:   true,
		weatherIntensity: particles.IntensityMedium,
		showFPS:          true,
		objects:          make([]PlacedObject, 0, 20),
		tileLayer:        image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight)),
		objectBuffer:     ebiten.NewImage(screenWidth, screenHeight),
		lightBuffer:      image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight)),
	}

	d.genre = d.genres[0]
	if err := d.regenerate(); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return d, nil
}

// regenerate regenerates all content for current genre/seed.
func (d *Demo) regenerate() error {
	startTime := time.Now()

	// Generate palette
	var err error
	d.palette, err = d.paletteGen.Generate(d.genre, d.seed)
	if err != nil {
		return fmt.Errorf("palette generation failed: %w", err)
	}

	// Generate tile variations
	tileConfig := tiles.Config{
		Type:    tiles.TileFloor,
		Width:   tileSize,
		Height:  tileSize,
		GenreID: d.genre,
		Seed:    d.seed,
		Variant: 0.5,
	}

	d.tileVariations, err = d.tileGen.GenerateVariations(tileConfig, 5)
	if err != nil {
		return fmt.Errorf("tile generation failed: %w", err)
	}

	// Render tiles to layer
	d.renderTiles()

	// Clear objects and spawn initial set
	d.objects = d.objects[:0]
	rng := rand.New(rand.NewSource(d.seed))
	for i := 0; i < 5; i++ {
		x := rng.Intn(gridWidth) * tileSize
		y := rng.Intn(gridHeight) * tileSize
		_ = d.spawnObjectAt(image.Point{X: x, Y: y})
	}

	// Setup weather
	weatherTypes := particles.GetGenreWeather(d.genre)
	if len(weatherTypes) > 0 {
		d.weatherType = weatherTypes[0]
	} else {
		d.weatherType = particles.WeatherRain
	}

	weatherConfig := particles.WeatherConfig{
		Type:      d.weatherType,
		Intensity: d.weatherIntensity,
		Width:     screenWidth,
		Height:    screenHeight,
		GenreID:   d.genre,
		Seed:      d.seed + 1000,
	}

	d.weatherSystem, err = particles.GenerateWeather(weatherConfig)
	if err != nil {
		return fmt.Errorf("weather generation failed: %w", err)
	}

	// Setup lighting
	d.lightSys.ClearLights()
	centerLight := lighting.Light{
		Type:      lighting.TypePoint,
		Position:  image.Point{X: screenWidth / 2, Y: screenHeight / 2},
		Color:     d.palette.Highlight1,
		Intensity: 1.0,
		Radius:    300.0,
		Falloff:   d.falloffType,
	}
	d.lightIndex = 0
	_ = d.lightSys.AddLight(centerLight)

	d.frameTime = time.Since(startTime)
	return nil
}

// renderTiles renders the tile layer.
func (d *Demo) renderTiles() {
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			// Use position-based seed for deterministic variation selection
			seed := int64(x*1000 + y)
			tile := d.tileVariations.GetVariationBySeed(seed)
			if tile != nil {
				// Copy tile to tile layer
				dx := x * tileSize
				dy := y * tileSize
				for ty := 0; ty < tile.Bounds().Dy(); ty++ {
					for tx := 0; tx < tile.Bounds().Dx(); tx++ {
						if dx+tx < d.tileLayer.Bounds().Dx() && dy+ty < d.tileLayer.Bounds().Dy() {
							d.tileLayer.Set(dx+tx, dy+ty, tile.At(tx, ty))
						}
					}
				}
			}
		}
	}
}

// spawnObjectAt spawns a random object at the given position.
func (d *Demo) spawnObjectAt(pos image.Point) error {
	objectTypes := []environment.SubType{
		environment.SubTypeTable, environment.SubTypeChair, environment.SubTypeBed,
		environment.SubTypePlant, environment.SubTypeStatue, environment.SubTypeTorch,
		environment.SubTypeBarrel, environment.SubTypeCrate, environment.SubTypePillar,
		environment.SubTypeSpikes, environment.SubTypeFirePit,
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	subType := objectTypes[rng.Intn(len(objectTypes))]

	config := environment.Config{
		SubType: subType,
		Width:   tileSize,
		Height:  tileSize,
		GenreID: d.genre,
		Seed:    time.Now().UnixNano(),
	}

	obj, err := d.objGen.Generate(config)
	if err != nil {
		return err
	}

	d.objects = append(d.objects, PlacedObject{obj: obj, pos: pos})
	return nil
}

// Update handles input and updates game state.
func (d *Demo) Update() error {
	// Genre switching
	for i := 0; i < 5; i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(int(ebiten.Key1) + i)) {
			d.genreIndex = i
			d.genre = d.genres[i]
			d.seed = time.Now().UnixNano()
			return d.regenerate()
		}
	}

	// Get current light for movement
	light, err := d.lightSys.GetLight(d.lightIndex)
	if err == nil {
		moved := false
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			light.Position.X -= 5
			moved = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			light.Position.X += 5
			moved = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			light.Position.Y -= 5
			moved = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			light.Position.Y += 5
			moved = true
		}
		if moved {
			_ = d.lightSys.UpdateLight(d.lightIndex, light)
		}
	}

	// Toggle weather
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		d.weatherEnabled = !d.weatherEnabled
	}

	// Weather intensity
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		d.weatherIntensity = (d.weatherIntensity + 1) % 4
		config := d.weatherSystem.Config
		config.Intensity = d.weatherIntensity
		newWeather, err := particles.GenerateWeather(config)
		if err == nil {
			d.weatherSystem = newWeather
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		d.weatherIntensity = (d.weatherIntensity + 3) % 4
		config := d.weatherSystem.Config
		config.Intensity = d.weatherIntensity
		newWeather, err := particles.GenerateWeather(config)
		if err == nil {
			d.weatherSystem = newWeather
		}
	}

	// Cycle light falloff
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		d.falloffType = (d.falloffType + 1) % 4
		if light, err := d.lightSys.GetLight(d.lightIndex); err == nil {
			light.Falloff = d.falloffType
			_ = d.lightSys.UpdateLight(d.lightIndex, light)
		}
	}

	// Spawn object at mouse
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		mx, my := ebiten.CursorPosition()
		_ = d.spawnObjectAt(image.Point{X: mx, Y: my})
	}

	// Reset
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		d.seed = time.Now().UnixNano()
		return d.regenerate()
	}

	// Toggle FPS
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		d.showFPS = !d.showFPS
	}

	// Exit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("exit requested")
	}

	// Update weather
	if d.weatherEnabled {
		d.weatherSystem.Update(1.0 / 60.0)
	}

	return nil
}

// Draw renders the demo.
func (d *Demo) Draw(screen *ebiten.Image) {
	startTime := time.Now()

	// Copy tile layer to light buffer
	copy(d.lightBuffer.Pix, d.tileLayer.Pix)

	// Draw objects to object buffer
	d.objectBuffer.Clear()
	for _, obj := range d.objects {
		// Convert RGBA to Ebiten image
		ebitenImg := ebiten.NewImageFromImage(obj.obj.Sprite)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(obj.pos.X), float64(obj.pos.Y))
		d.objectBuffer.DrawImage(ebitenImg, opts)
	}

	// Composite objects onto light buffer
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			r, g, b, a := d.objectBuffer.At(x, y).RGBA()
			if a > 0 {
				d.lightBuffer.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
			}
		}
	}

	// Apply lighting
	litImage := d.lightSys.ApplyLighting(d.lightBuffer)
	screen.DrawImage(ebiten.NewImageFromImage(litImage), nil)

	// Draw weather
	if d.weatherEnabled {
		d.drawWeather(screen)
	}

	// Draw light position indicator
	if light, err := d.lightSys.GetLight(d.lightIndex); err == nil {
		lx, ly := float64(light.Position.X), float64(light.Position.Y)
		ebitenutil.DrawCircle(screen, lx, ly, 10, color.RGBA{255, 255, 0, 200})
	}

	// Draw UI
	d.drawUI(screen)

	d.frameTime = time.Since(startTime)
	d.lastFPS = ebiten.ActualFPS()
}

// drawWeather draws weather particles on screen.
func (d *Demo) drawWeather(screen *ebiten.Image) {
	for _, p := range d.weatherSystem.Particles {
		if p.Life > 0 {
			r, g, b, _ := p.Color.RGBA()
			alpha := uint8(p.Life * 255)
			col := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), alpha}
			ebitenutil.DrawCircle(screen, p.X, p.Y, p.Size, col)
		}
	}
}

// drawUI renders the UI overlay.
func (d *Demo) drawUI(screen *ebiten.Image) {
	y := 10
	ebitenutil.DebugPrintAt(screen, "Phase 5 Environment Demo", 10, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Genre: %s (1-5 to switch)", d.genre), 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weather: %s [%s] (Space: toggle, W/S: intensity)",
		d.weatherType.String(), d.weatherIntensity.String()), 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Light Falloff: %s (L: cycle)", d.falloffType.String()), 10, y)
	y += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Objects: %d (O: spawn at mouse)", len(d.objects)), 10, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "Controls: Arrow Keys: move light | R: reset | ESC: exit", 10, y)

	if d.showFPS {
		y += 20
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", d.lastFPS), 10, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %.2fms", float64(d.frameTime.Microseconds())/1000.0), 10, y)
		y += 15
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Particles: %d", len(d.weatherSystem.Particles)), 10, y)
	}

	// Genre list
	y = screenHeight - 100
	ebitenutil.DebugPrintAt(screen, "Genres:", 10, y)
	for i, g := range d.genres {
		y += 15
		prefix := "  "
		if i == d.genreIndex {
			prefix = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%d: %s", prefix, i+1, g), 10, y)
	}
}

// Layout returns the screen dimensions.
func (d *Demo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	demo, err := NewDemo()
	if err != nil {
		log.Fatalf("Failed to create demo: %v", err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Venture - Phase 5 Environment Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(demo); err != nil {
		if err.Error() != "exit requested" {
			log.Fatal(err)
		}
	}
	log.Println("Demo exited normally")
}
