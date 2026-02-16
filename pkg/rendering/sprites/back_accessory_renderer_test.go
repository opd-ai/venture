package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestBackAccessoryTypeString(t *testing.T) {
	tests := []struct {
		name     string
		aType    BackAccessoryType
		expected string
	}{
		{"none", BackAccessoryNone, "none"},
		{"cape", BackAccessoryCape, "cape"},
		{"cloak", BackAccessoryCloak, "cloak"},
		{"quiver", BackAccessoryQuiver, "quiver"},
		{"backpack", BackAccessoryBackpack, "backpack"},
		{"banner", BackAccessoryBanner, "banner"},
		{"scarf", BackAccessoryScarf, "scarf"},
		{"wing_cape", BackAccessoryWingCape, "wing_cape"},
		{"unknown", BackAccessoryType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aType.String(); got != tt.expected {
				t.Errorf("BackAccessoryType(%d).String() = %q, want %q", tt.aType, got, tt.expected)
			}
		})
	}
}

func TestSelectBackAccessoryForRole(t *testing.T) {
	tests := []struct {
		name  string
		role  string
		genre string
		seed  int64
	}{
		{"warrior_fantasy", "warrior", "fantasy", 12345},
		{"mage_horror", "mage", "horror", 67890},
		{"ranger_fantasy", "ranger", "fantasy", 11111},
		{"merchant_cyberpunk", "merchant", "cyberpunk", 22222},
		{"player_scifi", "player", "sci-fi", 33333},
		{"priest_default", "priest", "", 44444},
		{"rogue_postapoc", "rogue", "post-apocalyptic", 55555},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectBackAccessoryForRole(tt.role, tt.genre, tt.seed)
			if result < BackAccessoryNone || result >= backAccessoryCount {
				t.Errorf("SelectBackAccessoryForRole(%q, %q, %d) = %d, out of range [0, %d)",
					tt.role, tt.genre, tt.seed, result, backAccessoryCount)
			}
		})
	}
}

func TestSelectBackAccessoryDeterministic(t *testing.T) {
	role, genre, seed := "warrior", "fantasy", int64(42)
	a := SelectBackAccessoryForRole(role, genre, seed)
	b := SelectBackAccessoryForRole(role, genre, seed)
	if a != b {
		t.Errorf("SelectBackAccessoryForRole not deterministic: %d != %d", a, b)
	}
}

func TestSelectBackAccessoryVariety(t *testing.T) {
	seen := make(map[BackAccessoryType]bool)
	for seed := int64(0); seed < 200; seed++ {
		result := SelectBackAccessoryForRole("warrior", "fantasy", seed)
		seen[result] = true
	}
	// Should produce at least 3 different types including None
	if len(seen) < 3 {
		t.Errorf("Expected at least 3 different accessory types from 200 seeds, got %d", len(seen))
	}
}

func TestComputeBackAccessoryParams(t *testing.T) {
	torsoSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.75,
		RelativeHeight: 0.50,
	}

	tests := []struct {
		name      string
		aType     BackAccessoryType
		direction Direction
		genre     string
		seed      int64
	}{
		{"cape_down", BackAccessoryCape, DirDown, "fantasy", 100},
		{"cloak_up", BackAccessoryCloak, DirUp, "horror", 200},
		{"quiver_left", BackAccessoryQuiver, DirLeft, "fantasy", 300},
		{"backpack_right", BackAccessoryBackpack, DirRight, "cyberpunk", 400},
		{"banner_down", BackAccessoryBanner, DirDown, "fantasy", 500},
		{"scarf_down", BackAccessoryScarf, DirDown, "sci-fi", 600},
		{"wing_cape_down", BackAccessoryWingCape, DirDown, "fantasy", 700},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ComputeBackAccessoryParams(32, 32, torsoSpec, tt.aType, tt.direction, tt.seed, tt.genre)
			if params.SpriteWidth != 32 || params.SpriteHeight != 32 {
				t.Errorf("Expected 32x32, got %dx%d", params.SpriteWidth, params.SpriteHeight)
			}
			if params.AccessoryType != tt.aType {
				t.Errorf("Expected type %d, got %d", tt.aType, params.AccessoryType)
			}
			if params.PrimaryColor.A != 255 {
				t.Error("Primary color should be fully opaque")
			}
		})
	}
}

func TestRenderBackAccessoryOverlay(t *testing.T) {
	torsoSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.75,
		RelativeHeight: 0.50,
	}

	accessoryTypes := []BackAccessoryType{
		BackAccessoryCape,
		BackAccessoryCloak,
		BackAccessoryQuiver,
		BackAccessoryBackpack,
		BackAccessoryBanner,
		BackAccessoryScarf,
		BackAccessoryWingCape,
	}

	for _, aType := range accessoryTypes {
		t.Run(aType.String(), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := ComputeBackAccessoryParams(32, 32, torsoSpec, aType, DirDown, 12345, "fantasy")
			RenderBackAccessoryOverlay(buf, params)

			// Verify that some pixels were drawn (non-zero)
			nonZero := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					r, g, b, a := buf.At(x, y).RGBA()
					if r > 0 || g > 0 || b > 0 || a > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("RenderBackAccessoryOverlay(%s) produced zero pixels", aType.String())
			}
		})
	}
}

func TestRenderBackAccessoryNone(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	params := BackAccessoryParams{
		SpriteWidth:   32,
		SpriteHeight:  32,
		AccessoryType: BackAccessoryNone,
	}
	RenderBackAccessoryOverlay(buf, params)

	// Should draw nothing
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			r, g, b, a := buf.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 || a > 0 {
				t.Fatal("BackAccessoryNone should produce zero pixels")
			}
		}
	}
}

func TestRenderBackAccessoryDirections(t *testing.T) {
	torsoSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.75,
		RelativeHeight: 0.50,
	}

	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}
	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := ComputeBackAccessoryParams(32, 32, torsoSpec, BackAccessoryCape, dir, 42, "fantasy")
			RenderBackAccessoryOverlay(buf, params)

			nonZero := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					_, _, _, a := buf.At(x, y).RGBA()
					if a > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("Direction %s produced zero pixels", dir)
			}
		})
	}
}

func TestAccessoryUtilityFunctions(t *testing.T) {
	t.Run("accessoryBlend", func(t *testing.T) {
		a := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		b := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		mid := accessoryBlend(a, b, 0.5)
		// Should be approximately (127, 127, 0)
		if mid.R < 120 || mid.R > 135 || mid.G < 120 || mid.G > 135 {
			t.Errorf("accessoryBlend midpoint unexpected: %v", mid)
		}
	})

	t.Run("accessoryHSV", func(t *testing.T) {
		// Red
		c := accessoryHSV(0, 1.0, 1.0)
		if c.R < 250 || c.G > 5 || c.B > 5 {
			t.Errorf("accessoryHSV(0,1,1) expected red, got %v", c)
		}
		// Green
		c = accessoryHSV(120, 1.0, 1.0)
		if c.G < 250 || c.R > 5 || c.B > 5 {
			t.Errorf("accessoryHSV(120,1,1) expected green, got %v", c)
		}
	})

	t.Run("accessoryMaxI", func(t *testing.T) {
		if accessoryMaxI(3, 5) != 5 {
			t.Error("accessoryMaxI(3,5) should be 5")
		}
		if accessoryMaxI(5, 3) != 5 {
			t.Error("accessoryMaxI(5,3) should be 5")
		}
	})

	t.Run("directionOffsetX", func(t *testing.T) {
		if directionOffsetX(DirUp, 0.5, 32) != 0 {
			t.Error("DirUp should have 0 offset")
		}
		left := directionOffsetX(DirLeft, 0.5, 32)
		right := directionOffsetX(DirRight, 0.5, 32)
		if left <= 0 {
			t.Errorf("DirLeft expected positive offset, got %d", left)
		}
		if right >= 0 {
			t.Errorf("DirRight expected negative offset, got %d", right)
		}
	})
}

func TestRoleBackAccessoryWeights(t *testing.T) {
	roles := []string{"warrior", "mage", "rogue", "ranger", "merchant", "priest", "player", "unknown"}
	genres := []string{"horror", "cyberpunk", "sci-fi", "post-apocalyptic", "fantasy", ""}

	for _, role := range roles {
		for _, genre := range genres {
			t.Run(role+"_"+genre, func(t *testing.T) {
				weights := roleBackAccessoryWeights(role, genre)
				total := 0.0
				for _, w := range weights {
					if w < 0 {
						t.Errorf("Negative weight for role=%q genre=%q", role, genre)
					}
					total += w
				}
				if total <= 0 {
					t.Errorf("Zero total weight for role=%q genre=%q", role, genre)
				}
			})
		}
	}
}

func BenchmarkRenderBackAccessory(b *testing.B) {
	torsoSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.50,
		RelativeWidth:  0.75,
		RelativeHeight: 0.50,
	}
	params := ComputeBackAccessoryParams(32, 32, torsoSpec, BackAccessoryCape, DirDown, 42, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
		RenderBackAccessoryOverlay(buf, params)
	}
}
