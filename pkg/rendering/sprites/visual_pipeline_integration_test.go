package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestEntityTypeToCreatureForm(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		wantForm   string
	}{
		{"wolf is quadruped", "wolf", "quadruped"},
		{"bear is quadruped", "bear", "quadruped"},
		{"horse is quadruped", "horse", "quadruped"},
		{"quadruped literal", "quadruped", "quadruped"},
		{"slime is blob", "slime", "blob"},
		{"ooze is blob", "ooze", "blob"},
		{"robot is mechanical", "robot", "mechanical"},
		{"golem is mechanical", "golem", "mechanical"},
		{"dragon is flying", "dragon", "flying"},
		{"bat is flying", "bat", "flying"},
		{"snake is serpentine", "snake", "serpentine"},
		{"worm is serpentine", "worm", "serpentine"},
		{"spider is arachnid", "spider", "arachnid"},
		{"scorpion is arachnid", "scorpion", "arachnid"},
		{"skeleton is undead", "skeleton", "undead"},
		{"zombie is undead", "zombie", "undead"},
		{"unknown defaults to quadruped", "unknown_creature", "quadruped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EntityTypeToCreatureForm(tt.entityType)
			if got != tt.wantForm {
				t.Errorf("EntityTypeToCreatureForm(%q) = %q, want %q", tt.entityType, got, tt.wantForm)
			}
		})
	}
}

func TestEntityTypeToCreatureForm_Deterministic(t *testing.T) {
	entityTypes := []string{"wolf", "slime", "robot", "dragon", "snake", "spider", "skeleton"}
	for _, et := range entityTypes {
		a := EntityTypeToCreatureForm(et)
		b := EntityTypeToCreatureForm(et)
		if a != b {
			t.Errorf("EntityTypeToCreatureForm(%q) not deterministic: %q vs %q", et, a, b)
		}
	}
}

func TestSurfaceTextureIntegration(t *testing.T) {
	tests := []struct {
		name     string
		form     string
		genre    string
		wantType SurfaceTextureType
	}{
		{"quadruped gets fur", "quadruped", "fantasy", TexFur},
		{"serpentine gets scales", "serpentine", "fantasy", TexScales},
		{"arachnid gets chitin", "arachnid", "horror", TexChitin},
		{"mechanical gets metal", "mechanical", "scifi", TexMetal},
		{"undead gets bone", "undead", "horror", TexBone},
		{"blob gets ooze", "blob", "fantasy", TexOoze},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			texSet := GenerateSurfaceTextureSet(42, tt.form, tt.genre)
			if texSet.TorsoTexture.Type != tt.wantType {
				t.Errorf("form %q: got texture type %v, want %v", tt.form, texSet.TorsoTexture.Type, tt.wantType)
			}
			if texSet.TorsoTexture.Intensity <= 0 {
				t.Errorf("form %q: torso texture intensity should be >0, got %f", tt.form, texSet.TorsoTexture.Intensity)
			}
		})
	}
}

func TestDepthEnhancementOnSpriteImage(t *testing.T) {
	// Create a test sprite with opaque pixels
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 120, G: 80, B: 60, A: 255})
		}
	}

	cfg := DefaultDepthEnhanceConfig(12345)
	zones := ApplyDepthEnhancement(img, cfg)
	if zones == 0 {
		t.Error("ApplyDepthEnhancement should process at least one zone")
	}

	// Verify pixels were modified (not all identical)
	firstPx := img.RGBAAt(10, 10)
	allSame := true
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			if img.RGBAAt(x, y) != firstPx {
				allSame = false
				break
			}
		}
		if !allSame {
			break
		}
	}
	if allSame {
		t.Error("depth enhancement should produce pixel variation")
	}
}

func TestDepthEnhancementForCreature(t *testing.T) {
	forms := []string{"quadruped", "serpentine", "arachnid", "flying", "blob"}

	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 4; y < 28; y++ {
				for x := 4; x < 28; x++ {
					img.SetRGBA(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
				}
			}
			cfg := DefaultDepthEnhanceConfig(99)
			zones := ApplyDepthEnhancementForCreature(img, form, cfg)
			if zones == 0 {
				t.Errorf("form %q should process at least one zone", form)
			}
		})
	}
}

func TestColorTemperatureIntegration(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk", "scifi", "postapoc", ""}

	for _, genre := range genres {
		t.Run("genre_"+genre, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 6; y < 26; y++ {
				for x := 6; x < 26; x++ {
					img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
				}
			}

			cfg := GenreColorTemperatureConfig(genre, 777)
			modified := ApplyColorTemperature(img, cfg)
			if modified == 0 {
				t.Errorf("genre %q: expected pixels to be modified", genre)
			}
		})
	}
}

func TestColorTemperatureGenreVariation(t *testing.T) {
	// Different genres should produce different color configs
	fantasy := GenreColorTemperatureConfig("fantasy", 42)
	horror := GenreColorTemperatureConfig("horror", 42)
	cyberpunk := GenreColorTemperatureConfig("cyberpunk", 42)

	if fantasy.WarmShift == horror.WarmShift && fantasy.CoolShift == horror.CoolShift {
		t.Error("fantasy and horror should have different color temperature settings")
	}
	if horror.WarmShift == cyberpunk.WarmShift && horror.CoolShift == cyberpunk.CoolShift {
		t.Error("horror and cyberpunk should have different color temperature settings")
	}
}

func TestSurfaceTextureAppliesOnlyToOpaque(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Only fill top-left quadrant
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 100, G: 80, B: 60, A: 255})
		}
	}

	params := SurfaceTextureParams{
		Type:           TexFur,
		Intensity:      0.5,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 80, G: 60, B: 40, A: 255},
		SecondaryColor: color.RGBA{R: 120, G: 100, B: 80, A: 255},
	}
	ApplySurfaceTexture(img, img.Bounds(), params, 42)

	// Transparent pixels in bottom-right should remain transparent
	for y := 8; y < 16; y++ {
		for x := 8; x < 16; x++ {
			if img.RGBAAt(x, y).A != 0 {
				t.Errorf("pixel (%d,%d) should remain transparent after surface texture", x, y)
			}
		}
	}
}

func TestFullPipelineOrder(t *testing.T) {
	// Verify that depth enhancement + color temperature applied to the same image
	// produces a different result than either alone, confirming both passes execute.
	makeImg := func() *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		for y := 8; y < 24; y++ {
			for x := 8; x < 24; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 128, G: 100, B: 80, A: 255})
			}
		}
		return img
	}

	// Depth only
	imgDepth := makeImg()
	ApplyDepthEnhancement(imgDepth, DefaultDepthEnhanceConfig(42))

	// Color temp only
	imgColor := makeImg()
	ApplyColorTemperature(imgColor, GenreColorTemperatureConfig("fantasy", 42))

	// Both in pipeline order
	imgBoth := makeImg()
	ApplyDepthEnhancement(imgBoth, DefaultDepthEnhanceConfig(42))
	ApplyColorTemperature(imgBoth, GenreColorTemperatureConfig("fantasy", 42))

	// imgBoth should differ from imgDepth (color temp was also applied)
	same := true
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			if imgBoth.RGBAAt(x, y) != imgDepth.RGBAAt(x, y) {
				same = false
				break
			}
		}
		if !same {
			break
		}
	}
	if same {
		t.Error("applying both depth + color temp should differ from depth alone")
	}
}
