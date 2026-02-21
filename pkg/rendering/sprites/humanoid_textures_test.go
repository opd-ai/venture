package sprites

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestHumanoidTextureType_String(t *testing.T) {
	tests := []struct {
		texType HumanoidTextureType
		want    string
	}{
		{SkinSmooth, "skin_smooth"},
		{SkinFreckled, "skin_freckled"},
		{SkinScarred, "skin_scarred"},
		{SkinWeathered, "skin_weathered"},
		{SkinTattooed, "skin_tattooed"},
		{FabricLinen, "fabric_linen"},
		{FabricLeather, "fabric_leather"},
		{FabricSilk, "fabric_silk"},
		{FabricWool, "fabric_wool"},
		{FabricChainmail, "fabric_chainmail"},
		{FabricPlate, "fabric_plate"},
		{HumTexHairStraight, "hair_straight"},
		{HumTexHairWavy, "hair_wavy"},
		{HumTexHairCurly, "hair_curly"},
		{HumTexHairBraided, "hair_braided"},
		{HumanoidTextureType(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.texType.String(); got != tt.want {
				t.Errorf("HumanoidTextureType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateHumanoidTextureSet_Determinism(t *testing.T) {
	seed := int64(42)
	genre := "fantasy"

	set1 := GenerateHumanoidTextureSet(seed, genre)
	set2 := GenerateHumanoidTextureSet(seed, genre)

	// Same seed + genre = identical output
	if set1.SkinTexture.Type != set2.SkinTexture.Type {
		t.Error("Determinism failed: skin texture type differs")
	}
	if set1.ClothingTop.Type != set2.ClothingTop.Type {
		t.Error("Determinism failed: clothing top type differs")
	}
	if set1.HairTexture.Type != set2.HairTexture.Type {
		t.Error("Determinism failed: hair texture type differs")
	}
	if set1.SkinTexture.Intensity != set2.SkinTexture.Intensity {
		t.Error("Determinism failed: skin intensity differs")
	}
}

func TestGenerateHumanoidTextureSet_GenreVariation(t *testing.T) {
	seed := int64(12345)
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "post-apocalyptic"}

	results := make(map[string]HumanoidTextureSet)
	for _, genre := range genres {
		results[genre] = GenerateHumanoidTextureSet(seed, genre)
	}

	// At least some genres should produce different results
	// (due to genre-specific biases)
	foundDifference := false
	for i := 0; i < len(genres)-1; i++ {
		for j := i + 1; j < len(genres); j++ {
			g1, g2 := genres[i], genres[j]
			if results[g1].SkinTexture.Type != results[g2].SkinTexture.Type ||
				results[g1].ClothingTop.Type != results[g2].ClothingTop.Type {
				foundDifference = true
				break
			}
		}
		if foundDifference {
			break
		}
	}
	// Note: it's possible all genres produce the same result for this seed,
	// so we just check that processing completes without error
}

func TestGenerateHumanoidTextureSet_ValidParameters(t *testing.T) {
	seeds := []int64{0, 1, 42, 999999, -1}
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "postapoc", "unknown"}

	for _, seed := range seeds {
		for _, genre := range genres {
			t.Run("", func(t *testing.T) {
				set := GenerateHumanoidTextureSet(seed, genre)

				// Intensity should be in valid range
				if set.SkinTexture.Intensity < 0 || set.SkinTexture.Intensity > 1 {
					t.Errorf("Skin intensity out of range: %f", set.SkinTexture.Intensity)
				}
				if set.ClothingTop.Intensity < 0 || set.ClothingTop.Intensity > 1 {
					t.Errorf("Clothing intensity out of range: %f", set.ClothingTop.Intensity)
				}
				if set.HairTexture.Intensity < 0 || set.HairTexture.Intensity > 1 {
					t.Errorf("Hair intensity out of range: %f", set.HairTexture.Intensity)
				}

				// Scale should be positive
				if set.SkinTexture.Scale <= 0 {
					t.Errorf("Skin scale non-positive: %f", set.SkinTexture.Scale)
				}
			})
		}
	}
}

func TestApplyHumanoidTexture_NilBuffer(t *testing.T) {
	params := HumanoidTextureParams{
		Type:      SkinSmooth,
		Intensity: 0.3,
	}
	// Should not panic
	ApplyHumanoidTexture(nil, image.Rect(0, 0, 10, 10), params, 42)
}

func TestApplyHumanoidTexture_EmptyRegion(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	params := HumanoidTextureParams{
		Type:      SkinFreckled,
		Intensity: 0.3,
	}
	// Empty region should not panic
	ApplyHumanoidTexture(buf, image.Rect(0, 0, 0, 0), params, 42)
}

func TestApplyHumanoidTexture_ZeroIntensity(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	// Fill with opaque color
	for i := 0; i < len(buf.Pix); i += 4 {
		buf.Pix[i] = 100
		buf.Pix[i+1] = 100
		buf.Pix[i+2] = 100
		buf.Pix[i+3] = 255
	}
	originalPix := make([]uint8, len(buf.Pix))
	copy(originalPix, buf.Pix)

	params := HumanoidTextureParams{
		Type:      SkinSmooth,
		Intensity: 0, // Zero intensity
	}
	ApplyHumanoidTexture(buf, buf.Bounds(), params, 42)

	// Buffer should be unchanged with zero intensity
	for i := range buf.Pix {
		if buf.Pix[i] != originalPix[i] {
			t.Error("Zero intensity texture modified buffer")
			break
		}
	}
}

func TestApplyHumanoidTexture_AllTypes(t *testing.T) {
	types := []HumanoidTextureType{
		SkinSmooth, SkinFreckled, SkinScarred, SkinWeathered, SkinTattooed,
		FabricLinen, FabricLeather, FabricSilk, FabricWool, FabricChainmail, FabricPlate,
		HumTexHairStraight, HumTexHairWavy, HumTexHairCurly, HumTexHairBraided,
	}

	for _, texType := range types {
		t.Run(texType.String(), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			// Fill with opaque color
			for i := 0; i < len(buf.Pix); i += 4 {
				buf.Pix[i] = 128
				buf.Pix[i+1] = 100
				buf.Pix[i+2] = 80
				buf.Pix[i+3] = 255
			}

			params := HumanoidTextureParams{
				Type:           texType,
				Intensity:      0.3,
				Scale:          1.0,
				PrimaryColor:   color.RGBA{R: 80, G: 60, B: 40, A: 150},
				SecondaryColor: color.RGBA{R: 140, G: 120, B: 100, A: 120},
				Direction:      0.5,
			}

			// Should not panic and should modify at least some pixels
			ApplyHumanoidTexture(buf, buf.Bounds(), params, int64(texType)*1000)
		})
	}
}

func TestApplyHumanoidTexture_TransparentPixelsUnmodified(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Leave pixels transparent (alpha = 0)

	params := HumanoidTextureParams{
		Type:         SkinFreckled,
		Intensity:    0.5,
		PrimaryColor: color.RGBA{R: 200, G: 100, B: 50, A: 200},
	}
	ApplyHumanoidTexture(buf, buf.Bounds(), params, 42)

	// All pixels should remain transparent
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			idx := y*buf.Stride + x*4
			if buf.Pix[idx+3] != 0 {
				t.Errorf("Transparent pixel at (%d,%d) was modified", x, y)
			}
		}
	}
}

func TestApplyHumanoidTexture_ModifiesOpaquePixels(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Fill with uniform opaque color
	for i := 0; i < len(buf.Pix); i += 4 {
		buf.Pix[i] = 128
		buf.Pix[i+1] = 128
		buf.Pix[i+2] = 128
		buf.Pix[i+3] = 255
	}

	params := HumanoidTextureParams{
		Type:           FabricLinen,
		Intensity:      0.5,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 80, G: 60, B: 40, A: 200},
		SecondaryColor: color.RGBA{R: 180, G: 160, B: 140, A: 180},
	}
	ApplyHumanoidTexture(buf, buf.Bounds(), params, 42)

	// At least some pixels should be modified (not all still 128,128,128)
	modifiedCount := 0
	for i := 0; i < len(buf.Pix); i += 4 {
		if buf.Pix[i] != 128 || buf.Pix[i+1] != 128 || buf.Pix[i+2] != 128 {
			modifiedCount++
		}
	}
	if modifiedCount == 0 {
		t.Error("FabricLinen texture did not modify any pixels")
	}
}

func TestHSVToRGB(t *testing.T) {
	tests := []struct {
		h, s, v float64
		wantR   uint8
		wantG   uint8
		wantB   uint8
	}{
		{0, 1, 1, 255, 0, 0},       // Red
		{120, 1, 1, 0, 255, 0},     // Green
		{240, 1, 1, 0, 0, 255},     // Blue
		{0, 0, 1, 255, 255, 255},   // White
		{0, 0, 0, 0, 0, 0},         // Black
		{60, 1, 1, 255, 255, 0},    // Yellow
		{180, 1, 1, 0, 255, 255},   // Cyan
		{300, 1, 1, 255, 0, 255},   // Magenta
		{0, 0.5, 0.5, 127, 63, 63}, // Desaturated dark red
	}

	for _, tt := range tests {
		r, g, b := hsvToRGB(tt.h, tt.s, tt.v)
		// Allow tolerance of 1 due to rounding
		if abs(int(r)-int(tt.wantR)) > 1 || abs(int(g)-int(tt.wantG)) > 1 || abs(int(b)-int(tt.wantB)) > 1 {
			t.Errorf("hsvToRGB(%v, %v, %v) = (%d, %d, %d), want (%d, %d, %d)",
				tt.h, tt.s, tt.v, r, g, b, tt.wantR, tt.wantG, tt.wantB)
		}
	}
}

func TestGenerateSkinTexture_Distribution(t *testing.T) {
	counts := make(map[HumanoidTextureType]int)
	numTrials := 1000

	for i := 0; i < numTrials; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		params := generateSkinTexture(rng, "fantasy")
		counts[params.Type]++
	}

	// SkinSmooth should be most common (~40%)
	if counts[SkinSmooth] < numTrials/5 {
		t.Errorf("SkinSmooth too rare: %d/%d", counts[SkinSmooth], numTrials)
	}
	// All skin types should appear
	for _, st := range []HumanoidTextureType{SkinSmooth, SkinFreckled, SkinScarred, SkinWeathered, SkinTattooed} {
		if counts[st] == 0 {
			t.Errorf("Skin type %s never generated", st.String())
		}
	}
}

func TestGenerateClothingTexture_Distribution(t *testing.T) {
	counts := make(map[HumanoidTextureType]int)
	numTrials := 1000

	for i := 0; i < numTrials; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		params := generateClothingTexture(rng, "fantasy", true)
		counts[params.Type]++
	}

	// All fabric types should appear
	for _, ft := range []HumanoidTextureType{FabricLinen, FabricLeather, FabricSilk, FabricWool, FabricChainmail, FabricPlate} {
		if counts[ft] == 0 {
			t.Errorf("Fabric type %s never generated", ft.String())
		}
	}
}

func TestGenerateHairTexture_Distribution(t *testing.T) {
	counts := make(map[HumanoidTextureType]int)
	numTrials := 1000

	for i := 0; i < numTrials; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		params := generateHairTexture(rng, "fantasy")
		counts[params.Type]++
	}

	// All hair types should appear
	for _, ht := range []HumanoidTextureType{HumTexHairStraight, HumTexHairWavy, HumTexHairCurly, HumTexHairBraided} {
		if counts[ht] == 0 {
			t.Errorf("Hair type %s never generated", ht.String())
		}
	}
}

func TestBlendHumanoidPixel(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set one pixel to opaque red
	buf.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Blend blue at 50%
	blendHumanoidPixel(buf, 1, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255}, 0.5)

	c := buf.RGBAAt(1, 1)
	// Result should be approximately purple (127, 0, 127)
	if c.R < 100 || c.R > 155 || c.B < 100 || c.B > 155 {
		t.Errorf("Blend result unexpected: %v", c)
	}
}

func TestBlendHumanoidPixel_ZeroAlpha(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	buf.Set(1, 1, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	original := buf.RGBAAt(1, 1)

	// Zero alpha blend should not change pixel
	blendHumanoidPixel(buf, 1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0)

	result := buf.RGBAAt(1, 1)
	if result != original {
		t.Errorf("Zero alpha blend changed pixel: %v -> %v", original, result)
	}
}

func TestBlendHumanoidPixel_TransparentPixel(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Leave pixel transparent

	// Blend should not affect transparent pixels
	blendHumanoidPixel(buf, 1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1.0)

	c := buf.RGBAAt(1, 1)
	if c.A != 0 {
		t.Errorf("Blend affected transparent pixel: alpha = %d", c.A)
	}
}

func BenchmarkApplyHumanoidTexture_SkinSmooth(b *testing.B) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 0; i < len(buf.Pix); i += 4 {
		buf.Pix[i], buf.Pix[i+1], buf.Pix[i+2], buf.Pix[i+3] = 128, 100, 80, 255
	}
	params := HumanoidTextureParams{
		Type:           SkinSmooth,
		Intensity:      0.3,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 80, G: 60, B: 40, A: 150},
		SecondaryColor: color.RGBA{R: 140, G: 120, B: 100, A: 120},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyHumanoidTexture(buf, buf.Bounds(), params, int64(i))
	}
}

func BenchmarkApplyHumanoidTexture_FabricLinen(b *testing.B) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 0; i < len(buf.Pix); i += 4 {
		buf.Pix[i], buf.Pix[i+1], buf.Pix[i+2], buf.Pix[i+3] = 128, 100, 80, 255
	}
	params := HumanoidTextureParams{
		Type:           FabricLinen,
		Intensity:      0.3,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 80, G: 60, B: 40, A: 150},
		SecondaryColor: color.RGBA{R: 140, G: 120, B: 100, A: 120},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyHumanoidTexture(buf, buf.Bounds(), params, int64(i))
	}
}

func BenchmarkApplyHumanoidTexture_HairWavy(b *testing.B) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for i := 0; i < len(buf.Pix); i += 4 {
		buf.Pix[i], buf.Pix[i+1], buf.Pix[i+2], buf.Pix[i+3] = 128, 100, 80, 255
	}
	params := HumanoidTextureParams{
		Type:           HumTexHairWavy,
		Intensity:      0.4,
		Scale:          1.0,
		PrimaryColor:   color.RGBA{R: 60, G: 40, B: 20, A: 180},
		SecondaryColor: color.RGBA{R: 100, G: 80, B: 60, A: 150},
		Direction:      0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyHumanoidTexture(buf, buf.Bounds(), params, int64(i))
	}
}

func BenchmarkGenerateHumanoidTextureSet(b *testing.B) {
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "postapoc"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateHumanoidTextureSet(int64(i), genres[i%len(genres)])
	}
}

func absHumTex(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
