package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestHeadgearTypeString(t *testing.T) {
	tests := []struct {
		hType HeadgearType
		want  string
	}{
		{HeadgearNone, "none"},
		{HeadgearCirclet, "circlet"},
		{HeadgearCrown, "crown"},
		{HeadgearWizardHat, "wizard_hat"},
		{HeadgearHood, "hood"},
		{HeadgearHornedHelm, "horned_helm"},
		{HeadgearWideBrim, "wide_brim"},
		{HeadgearTurban, "turban"},
		{HeadgearSkullCap, "skull_cap"},
		{HeadgearFullHelm, "full_helm"},
		{HeadgearPlumed, "plumed"},
		{HeadgearTiara, "tiara"},
		{HeadgearBandana, "bandana"},
		{HeadgearType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.hType.String()
			if got != tt.want {
				t.Errorf("HeadgearType(%d).String() = %q, want %q", tt.hType, got, tt.want)
			}
		})
	}
}

func TestSelectHeadgearForRole(t *testing.T) {
	tests := []struct {
		name  string
		role  string
		genre string
		seed  int64
	}{
		{"mage_fantasy", "mage", "fantasy", 42},
		{"warrior_fantasy", "warrior", "fantasy", 99},
		{"rogue_horror", "rogue", "horror", 777},
		{"merchant_cyberpunk", "merchant", "cyberpunk", 1234},
		{"priest_scifi", "priest", "sci-fi", 555},
		{"boss_fantasy", "boss", "fantasy", 888},
		{"ranger_postapoc", "ranger", "post-apocalyptic", 333},
		{"unknown_role", "unknown", "fantasy", 11},
		{"bard_default", "bard", "", 22},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hType := SelectHeadgearForRole(tt.role, tt.genre, tt.seed)
			if hType < HeadgearNone || hType >= HeadgearCount {
				t.Errorf("SelectHeadgearForRole(%q, %q, %d) = %d, out of range", tt.role, tt.genre, tt.seed, hType)
			}
		})
	}
}

func TestSelectHeadgearDeterministic(t *testing.T) {
	// Same inputs must produce same output
	for i := 0; i < 10; i++ {
		a := SelectHeadgearForRole("mage", "fantasy", 12345)
		b := SelectHeadgearForRole("mage", "fantasy", 12345)
		if a != b {
			t.Fatalf("Non-deterministic: got %d and %d for same inputs", a, b)
		}
	}
}

func TestSelectHeadgearVariety(t *testing.T) {
	// Different seeds should produce different headgear (at least some variety in 100 seeds)
	seen := make(map[HeadgearType]bool)
	for seed := int64(0); seed < 100; seed++ {
		h := SelectHeadgearForRole("warrior", "fantasy", seed)
		seen[h] = true
	}
	if len(seen) < 2 {
		t.Errorf("Expected variety across 100 seeds, got only %d distinct types", len(seen))
	}
}

func TestComputeHeadgearParams(t *testing.T) {
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
	}
	params := ComputeHeadgearParams(32, 32, headSpec, HeadgearCrown, DirDown, 42, "fantasy")

	if params.SpriteWidth != 32 || params.SpriteHeight != 32 {
		t.Errorf("dimensions mismatch: %dx%d", params.SpriteWidth, params.SpriteHeight)
	}
	if params.Type != HeadgearCrown {
		t.Errorf("type = %v, want HeadgearCrown", params.Type)
	}
	if params.PrimaryColor.A == 0 {
		t.Error("primary color has zero alpha")
	}
	if params.MaterialSheen < 0.3 || params.MaterialSheen > 0.8 {
		t.Errorf("material sheen %f out of expected range [0.3, 0.8]", params.MaterialSheen)
	}
}

func TestRenderHeadgearOverlayAllTypes(t *testing.T) {
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
	}

	types := []HeadgearType{
		HeadgearNone, HeadgearCirclet, HeadgearCrown, HeadgearWizardHat,
		HeadgearHood, HeadgearHornedHelm, HeadgearWideBrim, HeadgearTurban,
		HeadgearSkullCap, HeadgearFullHelm, HeadgearPlumed, HeadgearTiara,
		HeadgearBandana,
	}

	for _, hType := range types {
		t.Run(hType.String(), func(t *testing.T) {
			dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := HeadgearRenderParams{
				SpriteWidth:   32,
				SpriteHeight:  32,
				HeadSpec:      headSpec,
				Type:          hType,
				PrimaryColor:  color.RGBA{R: 180, G: 140, B: 80, A: 255},
				AccentColor:   color.RGBA{R: 200, G: 180, B: 100, A: 255},
				GemColor:      color.RGBA{R: 100, G: 200, B: 255, A: 255},
				Direction:     DirDown,
				MaterialSheen: 0.5,
				Seed:          42,
			}
			// Should not panic
			RenderHeadgearOverlay(dst, params)

			if hType == HeadgearNone {
				// No pixels should be modified for None type
				return
			}

			// At least some pixels should be non-zero for real headgear types
			nonZero := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					_, _, _, a := dst.At(x, y).RGBA()
					if a > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("HeadgearType %s rendered zero pixels", hType)
			}
		})
	}
}

func TestRenderHeadgearAllDirections(t *testing.T) {
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
	}

	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}
	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := HeadgearRenderParams{
				SpriteWidth:   32,
				SpriteHeight:  32,
				HeadSpec:      headSpec,
				Type:          HeadgearPlumed,
				PrimaryColor:  color.RGBA{R: 150, G: 150, B: 150, A: 255},
				AccentColor:   color.RGBA{R: 200, G: 50, B: 50, A: 255},
				GemColor:      color.RGBA{R: 100, G: 200, B: 255, A: 255},
				Direction:     dir,
				MaterialSheen: 0.5,
				Seed:          42,
			}
			RenderHeadgearOverlay(dst, params)
		})
	}
}

func TestRenderHeadgearAllGenreColors(t *testing.T) {
	genres := []string{"fantasy", "horror", "sci-fi", "cyberpunk", "post-apocalyptic", ""}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			headSpec := PartSpec{
				RelativeX:      0.5,
				RelativeY:      0.25,
				RelativeWidth:  0.35,
				RelativeHeight: 0.35,
			}
			params := ComputeHeadgearParams(32, 32, headSpec, HeadgearCrown, DirDown, 42, genre)
			if params.PrimaryColor.A == 0 {
				t.Error("primary color has zero alpha")
			}
		})
	}
}

func TestHeadgearFitsGenre(t *testing.T) {
	tests := []struct {
		hType HeadgearType
		genre string
		want  bool
	}{
		{HeadgearCrown, "fantasy", true},
		{HeadgearCrown, "horror", false},
		{HeadgearHood, "cyberpunk", true},
		{HeadgearCirclet, "sci-fi", true},
		{HeadgearWideBrim, "sci-fi", false},
		{HeadgearBandana, "post-apocalyptic", true},
		{HeadgearFullHelm, "sci-fi", true},
	}
	for _, tt := range tests {
		t.Run(tt.hType.String()+"_"+tt.genre, func(t *testing.T) {
			got := headgearFitsGenre(tt.hType, tt.genre)
			if got != tt.want {
				t.Errorf("headgearFitsGenre(%v, %q) = %v, want %v", tt.hType, tt.genre, got, tt.want)
			}
		})
	}
}

func TestHeadgearCandidatesForRole(t *testing.T) {
	roles := []string{"mage", "warrior", "rogue", "ranger", "priest", "merchant", "bard", "boss", "unknown"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			candidates := headgearCandidatesForRole(role)
			if len(candidates) == 0 {
				t.Errorf("no candidates for role %q", role)
			}
			for _, c := range candidates {
				if c < HeadgearNone || c >= HeadgearCount {
					t.Errorf("invalid candidate %d for role %q", c, role)
				}
			}
		})
	}
}

func TestHeadgearUtilityFunctions(t *testing.T) {
	// lightenHeadgear
	c := lightenHeadgear(color.RGBA{R: 100, G: 100, B: 100, A: 255}, 50)
	if c.R != 150 || c.G != 150 || c.B != 150 {
		t.Errorf("lightenHeadgear: got %v", c)
	}
	// Overflow protection
	c2 := lightenHeadgear(color.RGBA{R: 240, G: 240, B: 240, A: 255}, 50)
	if c2.R != 255 {
		t.Errorf("lightenHeadgear overflow: got R=%d", c2.R)
	}

	// darkenHeadgear
	d := darkenHeadgear(color.RGBA{R: 200, G: 100, B: 50, A: 255}, 0.5)
	if d.R != 100 || d.G != 50 || d.B != 25 {
		t.Errorf("darkenHeadgear: got %v", d)
	}

	// blendHeadgear
	b := blendHeadgear(
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		color.RGBA{R: 100, G: 100, B: 100, A: 255},
		0.5,
	)
	if b.R != 50 || b.G != 50 || b.B != 50 {
		t.Errorf("blendHeadgear: got %v", b)
	}
}

func BenchmarkRenderHeadgearOverlay(b *testing.B) {
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
	}
	params := HeadgearRenderParams{
		SpriteWidth:   32,
		SpriteHeight:  32,
		HeadSpec:      headSpec,
		Type:          HeadgearCrown,
		PrimaryColor:  color.RGBA{R: 180, G: 140, B: 80, A: 255},
		AccentColor:   color.RGBA{R: 200, G: 180, B: 100, A: 255},
		GemColor:      color.RGBA{R: 100, G: 200, B: 255, A: 255},
		Direction:     DirDown,
		MaterialSheen: 0.5,
		Seed:          42,
	}
	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderHeadgearOverlay(dst, params)
	}
}
