package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/postprocess"
	"github.com/sirupsen/logrus"
)

var (
	effectType = flag.String("effect", "colorgrading", "Effect type (motionblur, depthblur, colorgrading, vignette, chromatic, all)")
	genreID    = flag.String("genre", "", "Genre preset ID (fantasy, scifi, horror, cyberpunk, postapoc) - overrides individual settings")
	input      = flag.String("input", "", "Input image path (if empty, generates test pattern)")
	output     = flag.String("output", "postprocess.png", "Output PNG file path")
	width      = flag.Int("width", 400, "Test pattern width (if no input)")
	height     = flag.Int("height", 300, "Test pattern height (if no input)")
	verbose    = flag.Bool("verbose", false, "Show detailed processing information")

	// Color grading flags
	saturation  = flag.Float64("saturation", 1.0, "Color saturation (0.0-2.0, 1.0=normal)")
	contrast    = flag.Float64("contrast", 1.0, "Contrast (0.0-2.0, 1.0=normal)")
	brightness  = flag.Float64("brightness", 0.0, "Brightness adjustment (-1.0 to 1.0)")
	temperature = flag.Float64("temperature", 0.0, "Color temperature (-1.0=cool to 1.0=warm)")
	tint        = flag.Float64("tint", 0.0, "Color tint (-1.0=green to 1.0=magenta)")

	// Vignette flags
	vignetteIntensity = flag.Float64("vignette-intensity", 0.5, "Vignette darkness (0.0-1.0)")
	vignetteSoftness  = flag.Float64("vignette-softness", 0.6, "Vignette edge softness (0.0-1.0)")

	// Chromatic aberration flags
	chromaIntensity = flag.Float64("chroma-intensity", 0.2, "Chromatic aberration strength (0.0-1.0)")
	chromaSamples   = flag.Int("chroma-samples", 3, "Chromatic aberration samples (2-5)")

	// Motion blur flags
	motionVelocityX = flag.Float64("motion-vx", 0.0, "Motion blur horizontal velocity (pixels/frame)")
	motionVelocityY = flag.Float64("motion-vy", 0.0, "Motion blur vertical velocity (pixels/frame)")
	motionIntensity = flag.Float64("motion-intensity", 0.5, "Motion blur intensity (0.0-1.0)")
	motionSamples   = flag.Int("motion-samples", 7, "Motion blur samples (3-15)")

	// Depth blur flags
	depthFocal    = flag.Float64("depth-focal", 0.5, "Depth blur focal distance (0.0-1.0)")
	depthRange    = flag.Float64("depth-range", 0.2, "Depth blur focal range (0.0-1.0)")
	depthStrength = flag.Float64("depth-strength", 0.5, "Depth blur strength (0.0-1.0)")
	depthSamples  = flag.Int("depth-samples", 7, "Depth blur samples (3-15)")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("postprocesstest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"effect": *effectType,
		"genre":  *genreID,
		"output": *output,
	}).Info("Post-Process Test Tool started")

	// Load or generate test image
	var img *image.RGBA
	var err error

	if *input != "" {
		img, err = loadImage(*input)
		if err != nil {
			logger.WithError(err).Fatal("failed to load input image")
		}
		logger.WithField("path", *input).Info("loaded input image")
	} else {
		img = generateTestPattern(*width, *height)
		logger.WithFields(logrus.Fields{
			"width":  *width,
			"height": *height,
		}).Info("generated test pattern")
	}

	// Create processor
	processor := postprocess.NewProcessor()

	// Apply genre preset or individual settings
	if *genreID != "" {
		preset := postprocess.GetPresetByGenre(*genreID)
		processor.SetConfig(preset.Config)
		logger.WithFields(logrus.Fields{
			"genre":  *genreID,
			"preset": preset.Name,
		}).Info("applied genre preset")
	} else {
		// Configure individual effect
		config := configureEffect(*effectType)
		processor.SetConfig(config)
		logger.WithField("effect", *effectType).Info("configured individual effect")
	}

	// Process image
	var result *image.RGBA

	switch *effectType {
	case "motionblur":
		velMap := createVelocityMap(img.Bounds())
		result = processor.ApplyMotionBlur(img, velMap)
	case "depthblur":
		depthMap := postprocess.GenerateDepthMapFromY(img.Bounds())
		result = processor.ApplyDepthBlur(img, depthMap)
	case "colorgrading":
		result = processor.ApplyColorGrading(img)
	case "vignette":
		result = processor.ApplyVignette(img)
	case "chromatic":
		result = processor.ApplyChromaticAberration(img)
	case "all":
		velMap := createVelocityMap(img.Bounds())
		depthMap := postprocess.GenerateDepthMapFromY(img.Bounds())
		result = processor.ApplyAll(img, velMap, depthMap)
	default:
		logger.WithField("effect", *effectType).Fatal("unknown effect type")
	}

	logger.Info("post-processing applied successfully")

	// Save result
	if err := saveImage(result, *output); err != nil {
		logger.WithError(err).Fatal("failed to save output image")
	}

	logger.WithField("path", *output).Info("output saved successfully")

	// Display statistics
	displayStats(logger, img, result)
}

// configureEffect creates a config for the specified effect type
func configureEffect(effect string) postprocess.Config {
	config := postprocess.DefaultConfig()

	// Disable all effects by default
	config.MotionBlur.Enabled = false
	config.DepthBlur.Enabled = false
	config.ColorGrading.Enabled = false
	config.Vignette.Enabled = false
	config.ChromaticAberration.Enabled = false

	// Enable and configure the requested effect
	switch effect {
	case "motionblur":
		config.MotionBlur.Enabled = true
		config.MotionBlur.VelocityX = *motionVelocityX
		config.MotionBlur.VelocityY = *motionVelocityY
		config.MotionBlur.Intensity = *motionIntensity
		config.MotionBlur.Samples = *motionSamples

	case "depthblur":
		config.DepthBlur.Enabled = true
		config.DepthBlur.FocalDistance = *depthFocal
		config.DepthBlur.FocalRange = *depthRange
		config.DepthBlur.BlurStrength = *depthStrength
		config.DepthBlur.Samples = *depthSamples

	case "colorgrading":
		config.ColorGrading.Enabled = true
		config.ColorGrading.Saturation = *saturation
		config.ColorGrading.Contrast = *contrast
		config.ColorGrading.Brightness = *brightness
		config.ColorGrading.Temperature = *temperature
		config.ColorGrading.Tint = *tint

	case "vignette":
		config.Vignette.Enabled = true
		config.Vignette.Intensity = *vignetteIntensity
		config.Vignette.Softness = *vignetteSoftness

	case "chromatic":
		config.ChromaticAberration.Enabled = true
		config.ChromaticAberration.Intensity = *chromaIntensity
		config.ChromaticAberration.Samples = *chromaSamples

	case "all":
		// Enable all effects with flag values
		config.MotionBlur.Enabled = true
		config.MotionBlur.VelocityX = *motionVelocityX
		config.MotionBlur.VelocityY = *motionVelocityY
		config.MotionBlur.Intensity = *motionIntensity
		config.MotionBlur.Samples = *motionSamples

		config.DepthBlur.Enabled = false // Mutually exclusive with motion blur

		config.ColorGrading.Enabled = true
		config.ColorGrading.Saturation = *saturation
		config.ColorGrading.Contrast = *contrast
		config.ColorGrading.Brightness = *brightness
		config.ColorGrading.Temperature = *temperature
		config.ColorGrading.Tint = *tint

		config.Vignette.Enabled = true
		config.Vignette.Intensity = *vignetteIntensity
		config.Vignette.Softness = *vignetteSoftness

		config.ChromaticAberration.Enabled = true
		config.ChromaticAberration.Intensity = *chromaIntensity
		config.ChromaticAberration.Samples = *chromaSamples
	}

	return config
}

// createVelocityMap creates a test velocity map with uniform or radial motion
func createVelocityMap(bounds image.Rectangle) *postprocess.VelocityMap {
	if *motionVelocityX != 0.0 || *motionVelocityY != 0.0 {
		// Use uniform velocity from flags
		return postprocess.CreateUniformVelocityMap(bounds, *motionVelocityX, *motionVelocityY)
	}

	// Create a radial velocity map for demo (expansion effect)
	centerX := bounds.Dx() / 2
	centerY := bounds.Dy() / 2
	return postprocess.CreateRadialVelocityMap(bounds, centerX, centerY, 2.0)
}

// generateTestPattern creates a colorful test pattern image
func generateTestPattern(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a gradient pattern with some geometric shapes
			r := uint8((float64(x) / float64(width)) * 255.0)
			g := uint8((float64(y) / float64(height)) * 255.0)
			b := uint8(((float64(x) + float64(y)) / float64(width+height)) * 255.0)

			// Add some bright spots for bloom testing
			dx := float64(x - width/2)
			dy := float64(y - height/2)
			dist := dx*dx + dy*dy
			maxDist := float64(width*width + height*height) / 16.0

			if dist < maxDist {
				// Bright center
				brightness := 1.0 - (dist / maxDist)
				r = uint8(float64(r) + brightness*100.0)
				if r > 255 {
					r = 255
				}
				g = uint8(float64(g) + brightness*100.0)
				if g > 255 {
					g = 255
				}
				b = uint8(float64(b) + brightness*100.0)
				if b > 255 {
					b = 255
				}
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// loadImage loads a PNG image from a file
func loadImage(path string) (*image.RGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode PNG: %w", err)
	}

	// Convert to RGBA if needed
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba, nil
}

// saveImage saves an image to a PNG file
func saveImage(img *image.RGBA, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}

// displayStats shows statistics about the processing
func displayStats(logger *logrus.Logger, input, output *image.RGBA) {
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := width * height

	logger.WithFields(logrus.Fields{
		"width":  width,
		"height": height,
		"pixels": pixels,
	}).Info("image dimensions")

	// Calculate average brightness change
	var inputBrightness, outputBrightness float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ir, ig, ib, _ := input.At(x, y).RGBA()
			inputBrightness += float64(ir+ig+ib) / 3.0

			or, og, ob, _ := output.At(x, y).RGBA()
			outputBrightness += float64(or+og+ob) / 3.0
		}
	}

	inputBrightness /= float64(pixels) * 65535.0
	outputBrightness /= float64(pixels) * 65535.0

	brightnessChange := (outputBrightness - inputBrightness) / inputBrightness * 100.0

	logger.WithFields(logrus.Fields{
		"input_brightness":  fmt.Sprintf("%.3f", inputBrightness),
		"output_brightness": fmt.Sprintf("%.3f", outputBrightness),
		"change_percent":    fmt.Sprintf("%.1f%%", brightnessChange),
	}).Info("brightness analysis")
}
