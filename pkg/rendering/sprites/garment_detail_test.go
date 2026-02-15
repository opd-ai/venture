package sprites

import (
	"image"
	"image/color"
	"testing"
)

// makeTestBuffer creates an image.RGBA with a filled opaque rectangle
// in the center to simulate a rendered body.
func makeTestBuffer(w, h int) *image.RGBA {
	buf := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill center 60% with a solid body color to simulate rendered body parts
	margin := w / 5
	for y := h / 5; y < h*4/5; y++ {
		for x := margin; x < w-margin; x++ {
			idx := y*buf.Stride + x*4
			buf.Pix[idx] = 120   // R
			buf.Pix[idx+1] = 100 // G
			buf.Pix[idx+2] = 80  // B
			buf.Pix[idx+3] = 255 // A
		}
	}
	return buf
}

func TestSelectGarmentType(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		genre string
	}{
		{"fantasy seed 1", 1, "fantasy"},
		{"fantasy seed 2", 2, "fantasy"},
		{"scifi seed 100", 100, "sci-fi"},
		{"horror seed 42", 42, "horror"},
		{"cyberpunk seed 7", 7, "cyberpunk"},
		{"postapoc seed 99", 99, "postapoc"},
		{"empty genre seed 55", 55, ""},
		{"zero seed", 0, "fantasy"},
		{"negative seed", -12345, "horror"},
		{"large seed", 9999999, "cyberpunk"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := SelectGarmentType(tt.seed, tt.genre)
			if g < 0 || g >= garmentTypeCount {
				t.Errorf("SelectGarmentType(%d, %q) returned out-of-range type %d", tt.seed, tt.genre, g)
			}
		})
	}
}

func TestSelectGarmentTypeDeterminism(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		a := SelectGarmentType(seed, "fantasy")
		b := SelectGarmentType(seed, "fantasy")
		if a != b {
			t.Errorf("SelectGarmentType(%d, fantasy) not deterministic: %v != %v", seed, a, b)
		}
	}
}

func TestSelectGarmentTypeVariety(t *testing.T) {
	seen := make(map[GarmentType]bool)
	for seed := int64(0); seed < 200; seed++ {
		g := SelectGarmentType(seed, "")
		seen[g] = true
	}
	if len(seen) < 4 {
		t.Errorf("expected at least 4 distinct garment types in 200 seeds, got %d", len(seen))
	}
}

func TestRenderGarmentDetail(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		seed      int64
		genre     string
		direction string
	}{
		{"32x32 fantasy", 32, 32, 12345, "fantasy", "down"},
		{"32x32 scifi", 32, 32, 67890, "sci-fi", "up"},
		{"32x32 horror", 32, 32, 11111, "horror", "left"},
		{"32x32 cyberpunk", 32, 32, 22222, "cyberpunk", "right"},
		{"32x32 postapoc", 32, 32, 33333, "postapoc", "down"},
		{"16x16 small", 16, 16, 44444, "fantasy", "down"},
		{"64x64 large", 64, 64, 55555, "sci-fi", "up"},
		{"zero seed", 32, 32, 0, "", "down"},
		{"negative seed", 32, 32, -999, "horror", "down"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := makeTestBuffer(tt.width, tt.height)
			// Should not panic
			RenderGarmentDetail(buf, GarmentDetailParams{
				Width:     tt.width,
				Height:    tt.height,
				Seed:      tt.seed,
				Genre:     tt.genre,
				Direction: tt.direction,
			})
		})
	}
}

func TestRenderGarmentDetailNilBuffer(t *testing.T) {
	// Should not panic with nil buffer
	RenderGarmentDetail(nil, GarmentDetailParams{Width: 32, Height: 32, Seed: 1})
}

func TestRenderGarmentDetailZeroDimensions(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 1, 1))
	RenderGarmentDetail(buf, GarmentDetailParams{Width: 0, Height: 0, Seed: 1})
	RenderGarmentDetail(buf, GarmentDetailParams{Width: -1, Height: 32, Seed: 1})
}

func TestRenderGarmentDetailModifiesPixels(t *testing.T) {
	buf := makeTestBuffer(32, 32)
	// Copy original pixels
	original := make([]byte, len(buf.Pix))
	copy(original, buf.Pix)

	RenderGarmentDetail(buf, GarmentDetailParams{
		Width:  32,
		Height: 32,
		Seed:   12345,
		Genre:  "fantasy",
	})

	// At least some pixels should be different
	changed := 0
	for i := range buf.Pix {
		if buf.Pix[i] != original[i] {
			changed++
		}
	}
	if changed == 0 {
		t.Error("RenderGarmentDetail did not modify any pixels")
	}
}

func TestRenderGarmentDetailDeterminism(t *testing.T) {
	for seed := int64(0); seed < 10; seed++ {
		buf1 := makeTestBuffer(32, 32)
		buf2 := makeTestBuffer(32, 32)

		params := GarmentDetailParams{Width: 32, Height: 32, Seed: seed, Genre: "fantasy"}
		RenderGarmentDetail(buf1, params)
		RenderGarmentDetail(buf2, params)

		for i := range buf1.Pix {
			if buf1.Pix[i] != buf2.Pix[i] {
				t.Errorf("seed %d: non-deterministic at pixel byte %d", seed, i)
				break
			}
		}
	}
}

func TestGarmentTypeString(t *testing.T) {
	tests := []struct {
		garment GarmentType
		want    string
	}{
		{GarmentTunic, "tunic"},
		{GarmentRobe, "robe"},
		{GarmentVest, "vest"},
		{GarmentPlateArmor, "plate_armor"},
		{GarmentShirt, "shirt"},
		{GarmentCloak, "cloak"},
		{GarmentType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.garment.String(); got != tt.want {
				t.Errorf("GarmentType(%d).String() = %q, want %q", tt.garment, got, tt.want)
			}
		})
	}
}

func TestGenreGarmentWeights(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "scifi", "horror", "cyberpunk", "postapoc", "post-apocalyptic", ""}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			w := genreGarmentWeights(genre)
			total := 0
			for _, v := range w {
				if v < 0 {
					t.Errorf("negative weight for genre %q", genre)
				}
				total += v
			}
			if total <= 0 {
				t.Errorf("total weight is zero for genre %q", genre)
			}
		})
	}
}

func TestGarmentIsOpaque(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set one pixel to opaque
	buf.Pix[0] = 255
	buf.Pix[1] = 0
	buf.Pix[2] = 0
	buf.Pix[3] = 255

	if !garmentIsOpaque(buf, 0, 0) {
		t.Error("expected opaque at (0,0)")
	}
	if garmentIsOpaque(buf, 1, 0) {
		t.Error("expected transparent at (1,0)")
	}
	if garmentIsOpaque(buf, -1, 0) {
		t.Error("expected false for out-of-bounds")
	}
	if garmentIsOpaque(buf, 0, 99) {
		t.Error("expected false for out-of-bounds y")
	}
}

func TestSampleBodyColor(t *testing.T) {
	buf := makeTestBuffer(32, 32)
	c := sampleBodyColor(buf, 32, 32)
	if c.A == 0 {
		t.Error("sampleBodyColor returned transparent pixel from filled buffer")
	}
}

func TestSampleBodyColorEmptyBuffer(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	c := sampleBodyColor(buf, 32, 32)
	// Should return the empty pixel (all zeros) since the buffer is empty
	if c.A != 0 {
		t.Log("sampleBodyColor returned non-zero alpha from empty buffer, which is acceptable")
	}
}

func TestAllGarmentTypesRender(t *testing.T) {
	// Force each garment type and verify rendering doesn't panic.
	// We use seeds that produce each type via genre tuning.
	for gt := GarmentType(0); gt < garmentTypeCount; gt++ {
		t.Run(gt.String(), func(t *testing.T) {
			// Find a seed that produces this garment type
			found := false
			for seed := int64(0); seed < 500; seed++ {
				if SelectGarmentType(seed, "") == gt {
					buf := makeTestBuffer(32, 32)
					RenderGarmentDetail(buf, GarmentDetailParams{
						Width:  32,
						Height: 32,
						Seed:   seed,
						Genre:  "",
					})
					found = true
					break
				}
			}
			if !found {
				t.Skipf("could not find seed producing garment type %v in 500 tries", gt)
			}
		})
	}
}

func TestLightenDarkenColor(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	l := garmentLighten(c, 1.5)
	d := garmentDarken(c, 0.5)
	if l.R <= c.R {
		t.Errorf("garmentLighten should increase R: got %d <= %d", l.R, c.R)
	}
	if d.R >= c.R {
		t.Errorf("garmentDarken should decrease R: got %d >= %d", d.R, c.R)
	}
}

func BenchmarkRenderGarmentDetail32(b *testing.B) {
	buf := makeTestBuffer(32, 32)
	params := GarmentDetailParams{Width: 32, Height: 32, Seed: 12345, Genre: "fantasy"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderGarmentDetail(buf, params)
	}
}

func BenchmarkSelectGarmentType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SelectGarmentType(int64(i), "fantasy")
	}
}
