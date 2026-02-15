package sprites

import (
	"image"
	"image/color"
	"testing"
)

// TestRenderRoleDetails_AllRoles verifies that every supported role renders
// without panics and produces non-empty detail pixels.
func TestRenderRoleDetails_AllRoles(t *testing.T) {
	roles := []VisualRole{
		RoleMage, RoleWarrior, RoleKnight, RoleRogue,
		RoleMerchant, RoleRanger, RolePriest,
	}
	directions := []string{"up", "down", "left", "right"}

	for _, role := range roles {
		for _, dir := range directions {
			t.Run(string(role)+"_"+dir, func(t *testing.T) {
				buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
				// Pre-fill center region so isOpaqueAt checks succeed
				fillCenterRegion(buf, 32, 32)

				RenderRoleDetails(buf, RoleDetailParams{
					Width:     32,
					Height:    32,
					Role:      role,
					Direction: dir,
					Seed:      42,
					Genre:     "fantasy",
				})

				if countNonBlackPixels(buf) == 0 {
					t.Errorf("role %s direction %s produced no detail pixels", role, dir)
				}
			})
		}
	}
}

// TestRenderRoleDetails_NilBuffer verifies nil buffer is handled gracefully.
func TestRenderRoleDetails_NilBuffer(t *testing.T) {
	// Should not panic
	RenderRoleDetails(nil, RoleDetailParams{
		Width: 32, Height: 32, Role: RoleMage, Direction: "down", Seed: 1,
	})
}

// TestRenderRoleDetails_ZeroDimensions verifies zero dimensions are handled.
func TestRenderRoleDetails_ZeroDimensions(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderRoleDetails(buf, RoleDetailParams{
		Width: 0, Height: 0, Role: RoleWarrior, Direction: "down", Seed: 1,
	})
	// Should produce no pixels
	if countNonBlackPixels(buf) != 0 {
		t.Error("zero dimensions should produce no detail pixels")
	}
}

// TestRenderRoleDetails_UnknownRole verifies unknown roles are skipped.
func TestRenderRoleDetails_UnknownRole(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderRoleDetails(buf, RoleDetailParams{
		Width: 32, Height: 32, Role: "unknown_role", Direction: "down", Seed: 1,
	})
	if countNonBlackPixels(buf) != 0 {
		t.Error("unknown role should produce no detail pixels")
	}
}

// TestRenderRoleDetails_Deterministic verifies same seed produces same output.
func TestRenderRoleDetails_Deterministic(t *testing.T) {
	roles := []VisualRole{
		RoleMage, RoleWarrior, RoleKnight, RoleRogue,
		RoleMerchant, RoleRanger, RolePriest,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			buf1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			fillCenterRegion(buf1, 32, 32)
			RenderRoleDetails(buf1, RoleDetailParams{
				Width: 32, Height: 32, Role: role, Direction: "down",
				Seed: 12345, Genre: "fantasy",
			})

			buf2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			fillCenterRegion(buf2, 32, 32)
			RenderRoleDetails(buf2, RoleDetailParams{
				Width: 32, Height: 32, Role: role, Direction: "down",
				Seed: 12345, Genre: "fantasy",
			})

			if !pixelDataEqual(buf1, buf2) {
				t.Errorf("role %s not deterministic for same seed", role)
			}
		})
	}
}

// TestRenderRoleDetails_DifferentSeeds verifies different seeds produce different output.
func TestRenderRoleDetails_DifferentSeeds(t *testing.T) {
	roles := []VisualRole{RoleMage, RoleWarrior, RoleRogue}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			buf1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			fillCenterRegion(buf1, 32, 32)
			RenderRoleDetails(buf1, RoleDetailParams{
				Width: 32, Height: 32, Role: role, Direction: "down",
				Seed: 100, Genre: "fantasy",
			})

			buf2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			fillCenterRegion(buf2, 32, 32)
			RenderRoleDetails(buf2, RoleDetailParams{
				Width: 32, Height: 32, Role: role, Direction: "down",
				Seed: 999, Genre: "fantasy",
			})

			if pixelDataEqual(buf1, buf2) {
				t.Errorf("role %s produced identical output for different seeds", role)
			}
		})
	}
}

// TestRenderRoleDetails_GenreAwareness verifies genre changes arcane/holy colors.
func TestRenderRoleDetails_GenreAwareness(t *testing.T) {
	tests := []struct {
		role   VisualRole
		genres []string
	}{
		{RoleMage, []string{"fantasy", "horror", "cyberpunk"}},
		{RoleKnight, []string{"fantasy", "horror", "cyberpunk"}},
		{RolePriest, []string{"fantasy", "horror", "cyberpunk"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			results := make([][]byte, len(tt.genres))
			for i, genre := range tt.genres {
				buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
				fillCenterRegion(buf, 32, 32)
				RenderRoleDetails(buf, RoleDetailParams{
					Width: 32, Height: 32, Role: tt.role, Direction: "down",
					Seed: 42, Genre: genre,
				})
				results[i] = make([]byte, len(buf.Pix))
				copy(results[i], buf.Pix)
			}

			// At least one genre pair should differ
			anyDiff := false
			for i := 1; i < len(results); i++ {
				if !bytesEqual(results[0], results[i]) {
					anyDiff = true
					break
				}
			}
			if !anyDiff {
				t.Errorf("role %s produced identical output across genres", tt.role)
			}
		})
	}
}

// TestRoleHelperFunctions verifies helper functions.
func TestRoleHelperFunctions(t *testing.T) {
	t.Run("maxDetail", func(t *testing.T) {
		if maxDetail(3, 5) != 5 {
			t.Error("maxDetail(3,5) should be 5")
		}
		if maxDetail(7, 2) != 7 {
			t.Error("maxDetail(7,2) should be 7")
		}
	})

	t.Run("isRoleEdgePixel", func(t *testing.T) {
		buf := image.NewRGBA(image.Rect(0, 0, 8, 8))
		// Set a single pixel
		setPixelSafe(buf, 4, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		if !isRoleEdgePixel(buf, 4, 4) {
			t.Error("isolated pixel should be edge pixel")
		}
		// Fill area
		for x := 3; x <= 5; x++ {
			for y := 3; y <= 5; y++ {
				setPixelSafe(buf, x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
		if isRoleEdgePixel(buf, 4, 4) {
			t.Error("interior pixel should not be edge pixel")
		}
	})

	t.Run("isHoodFront", func(t *testing.T) {
		if !isHoodFront(0.5, "down") {
			t.Error("angle 0.5 should be front for down-facing")
		}
	})

	t.Run("staffTipPosition", func(t *testing.T) {
		x, y := staffTipPosition(16, 16, "up", 32, 32)
		if x == 0 && y == 0 {
			t.Error("staff tip should not be at origin")
		}
	})

	t.Run("daggerPosition", func(t *testing.T) {
		x, y := daggerPosition(16, 16, "down", 32, 32)
		if x == 0 && y == 0 {
			t.Error("dagger position should not be at origin")
		}
	})

	t.Run("bladeDirection", func(t *testing.T) {
		dirs := []string{"up", "down", "left", "right", ""}
		for _, d := range dirs {
			bd := bladeDirection(d)
			if bd[0] == 0 && bd[1] == 0 {
				t.Errorf("bladeDirection(%q) returned zero vector", d)
			}
		}
	})

	t.Run("roleRearPosition", func(t *testing.T) {
		x, y := roleRearPosition(16, 16, "up", 32, 32)
		if y <= 16 {
			t.Error("rear of up-facing should be below center")
		}
		_ = x
	})

	t.Run("quiverPosition", func(t *testing.T) {
		x, y := quiverPosition(16, 16, "down", 32, 32)
		if x == 0 && y == 0 {
			t.Error("quiver should not be at origin")
		}
	})
}

// TestMageArcaneColor verifies genre-aware color selection.
func TestMageArcaneColor(t *testing.T) {
	genres := []string{"horror", "cyberpunk", "sci-fi", "post-apocalyptic", "fantasy", ""}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			c := mageArcaneColor(g, nil) // rng only used for default
			if c.A == 0 {
				t.Errorf("genre %q produced transparent arcane color", g)
			}
		})
	}
}

// TestKnightEmblemColor verifies genre-aware emblem colors.
func TestKnightEmblemColor(t *testing.T) {
	genres := []string{"horror", "cyberpunk", "fantasy", ""}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			c := knightEmblemColor(g, nil)
			if c.A == 0 {
				t.Errorf("genre %q produced transparent emblem color", g)
			}
		})
	}
}

// TestPriestHolyColor verifies genre-aware holy colors.
func TestPriestHolyColor(t *testing.T) {
	genres := []string{"horror", "cyberpunk", "sci-fi", "fantasy", ""}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			c := priestHolyColor(g, nil)
			if c.A == 0 {
				t.Errorf("genre %q produced transparent holy color", g)
			}
		})
	}
}

// TestRenderRoleDetails_SmallSprite verifies rendering on small sprites (16x16).
func TestRenderRoleDetails_SmallSprite(t *testing.T) {
	roles := []VisualRole{
		RoleMage, RoleWarrior, RoleKnight, RoleRogue,
		RoleMerchant, RoleRanger, RolePriest,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 16, 16))
			fillCenterRegion(buf, 16, 16)
			// Should not panic on small sprites
			RenderRoleDetails(buf, RoleDetailParams{
				Width: 16, Height: 16, Role: role, Direction: "down",
				Seed: 42, Genre: "fantasy",
			})
		})
	}
}

// TestRenderRoleDetails_LargeSprite verifies rendering on large sprites (64x64).
func TestRenderRoleDetails_LargeSprite(t *testing.T) {
	roles := []VisualRole{RoleMage, RoleWarrior, RoleKnight}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 64, 64))
			fillCenterRegion(buf, 64, 64)
			RenderRoleDetails(buf, RoleDetailParams{
				Width: 64, Height: 64, Role: role, Direction: "down",
				Seed: 42, Genre: "fantasy",
			})
			if countNonBlackPixels(buf) == 0 {
				t.Errorf("role %s produced no details at 64x64", role)
			}
		})
	}
}

// ============================================================================
// Test helpers
// ============================================================================

// fillCenterRegion fills the center 60% of the buffer with opaque grey
// to simulate a rendered body that role details will overlay.
func fillCenterRegion(buf *image.RGBA, w, h int) {
	marginX := w / 5
	marginY := h / 5
	body := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	for y := marginY; y < h-marginY; y++ {
		for x := marginX; x < w-marginX; x++ {
			setPixelSafe(buf, x, y, body)
		}
	}
}

// countNonBlackPixels counts pixels with non-zero RGBA values, excluding
// the pre-filled grey center.
func countNonBlackPixels(buf *image.RGBA) int {
	count := 0
	for i := 0; i < len(buf.Pix); i += 4 {
		r, g, b, a := buf.Pix[i], buf.Pix[i+1], buf.Pix[i+2], buf.Pix[i+3]
		if a > 0 && !(r == 128 && g == 128 && b == 128 && a == 255) {
			count++
		}
	}
	return count
}

// pixelDataEqual compares two RGBA buffers.
func pixelDataEqual(a, b *image.RGBA) bool {
	return bytesEqual(a.Pix, b.Pix)
}

// bytesEqual compares two byte slices.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
