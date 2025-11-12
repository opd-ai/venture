package engine

import "time"

// SpellComboComponent tracks spell casting history for combo detection.
// It stores recent spell casts and known combo recipes for an entity.
type SpellComboComponent struct {
	// RecentCasts tracks spells cast in the last second for combo detection
	RecentCasts []RecentCast

	// KnownRecipes stores discovered combo recipes (spell1+spell2 -> result)
	KnownRecipes []ComboRecipe

	// ActiveCombo is the currently active combo effect (nil if none)
	ActiveCombo *ActiveCombo

	// ComboWindow is the time window for combo detection (default 1.0 second)
	ComboWindow float64
}

// Type returns the component type identifier.
func (s *SpellComboComponent) Type() string {
	return "spell_combo"
}

// RecentCast represents a recently cast spell for combo tracking.
type RecentCast struct {
	// SpellName is the name of the cast spell
	SpellName string

	// Element is the elemental type of the spell
	Element string

	// CastTime is when the spell was cast (Unix timestamp in seconds)
	CastTime float64

	// SlotIndex is which spell slot was used (0-4)
	SlotIndex int
}

// ComboRecipe defines a discovered spell combination.
type ComboRecipe struct {
	// Spell1Name is the name of the first spell
	Spell1Name string

	// Spell2Name is the name of the second spell
	Spell2Name string

	// Element1 is the element of the first spell
	Element1 string

	// Element2 is the element of the second spell
	Element2 string

	// ResultEffect describes the combined effect
	ResultEffect string

	// PowerMultiplier is the damage/effect boost (1.0 = no change, 2.0 = double)
	PowerMultiplier float64

	// IsSymmetric indicates if order matters (true = A+B same as B+A)
	IsSymmetric bool
}

// ActiveCombo represents an ongoing combo effect.
type ActiveCombo struct {
	// Spell1Name is the first spell in the combo
	Spell1Name string

	// Spell2Name is the second spell in the combo
	Spell2Name string

	// PowerMultiplier is the active bonus multiplier
	PowerMultiplier float64

	// EffectDescription describes what's happening
	EffectDescription string

	// StartTime is when the combo was triggered (Unix timestamp)
	StartTime float64

	// Duration is how long the combo effect lasts
	Duration float64
}

// AddRecentCast adds a spell cast to the history.
func (s *SpellComboComponent) AddRecentCast(spellName, element string, castTime float64, slotIndex int) {
	s.RecentCasts = append(s.RecentCasts, RecentCast{
		SpellName: spellName,
		Element:   element,
		CastTime:  castTime,
		SlotIndex: slotIndex,
	})
}

// CleanOldCasts removes casts outside the combo window.
func (s *SpellComboComponent) CleanOldCasts(currentTime float64) {
	cutoff := currentTime - s.ComboWindow

	// Keep only recent casts within the window
	validCasts := make([]RecentCast, 0, len(s.RecentCasts))
	for _, cast := range s.RecentCasts {
		if cast.CastTime >= cutoff {
			validCasts = append(validCasts, cast)
		}
	}
	s.RecentCasts = validCasts
}

// HasRecentCasts returns true if there are casts in the last window.
func (s *SpellComboComponent) HasRecentCasts(currentTime float64) bool {
	s.CleanOldCasts(currentTime)
	return len(s.RecentCasts) > 0
}

// GetRecentCastsCount returns the number of spells cast in the window.
func (s *SpellComboComponent) GetRecentCastsCount(currentTime float64) int {
	s.CleanOldCasts(currentTime)
	return len(s.RecentCasts)
}

// HasRecipe checks if a combo recipe is known.
func (s *SpellComboComponent) HasRecipe(spell1, spell2 string) bool {
	for _, recipe := range s.KnownRecipes {
		if recipe.IsSymmetric {
			if (recipe.Spell1Name == spell1 && recipe.Spell2Name == spell2) ||
				(recipe.Spell1Name == spell2 && recipe.Spell2Name == spell1) {
				return true
			}
		} else {
			if recipe.Spell1Name == spell1 && recipe.Spell2Name == spell2 {
				return true
			}
		}
	}
	return false
}

// AddRecipe adds a new combo recipe.
func (s *SpellComboComponent) AddRecipe(recipe ComboRecipe) {
	// Don't add duplicates
	if s.HasRecipe(recipe.Spell1Name, recipe.Spell2Name) {
		return
	}
	s.KnownRecipes = append(s.KnownRecipes, recipe)
}

// GetRecipe returns the matching recipe if found.
func (s *SpellComboComponent) GetRecipe(spell1, spell2 string) *ComboRecipe {
	for i := range s.KnownRecipes {
		recipe := &s.KnownRecipes[i]
		if recipe.IsSymmetric {
			if (recipe.Spell1Name == spell1 && recipe.Spell2Name == spell2) ||
				(recipe.Spell1Name == spell2 && recipe.Spell2Name == spell1) {
				return recipe
			}
		} else {
			if recipe.Spell1Name == spell1 && recipe.Spell2Name == spell2 {
				return recipe
			}
		}
	}
	return nil
}

// IsComboActive returns true if a combo is currently active.
func (s *SpellComboComponent) IsComboActive(currentTime float64) bool {
	if s.ActiveCombo == nil {
		return false
	}
	return currentTime < s.ActiveCombo.StartTime+s.ActiveCombo.Duration
}

// GetCurrentTime returns the current time in seconds since epoch.
// This is a helper for getting consistent timestamps.
func GetCurrentTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}
