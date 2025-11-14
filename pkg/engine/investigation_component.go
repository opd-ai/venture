package engine

import (
	"fmt"
	"time"
)

// InvestigationComponent tracks a player's environmental investigation abilities.
// It manages the active investigation state, discovered areas, and investigation
// mechanics such as revealing hidden story fragments and environmental clues.
type InvestigationComponent struct {
	// InvestigationRadius is the radius (in tiles) within which the player can investigate.
	// Default is 3.0 tiles, allowing examination of nearby environment.
	InvestigationRadius float64

	// IsInvestigating indicates if the player is currently performing an investigation action.
	IsInvestigating bool

	// InvestigationStartTime is when the current investigation began.
	InvestigationStartTime time.Time

	// InvestigationDuration is how long the investigation action lasts (seconds).
	InvestigationDuration float64

	// DiscoveredAreas tracks grid positions that have been investigated.
	// Key format: "x,y" where x and y are grid tile coordinates.
	DiscoveredAreas map[string]bool

	// InvestigationCooldown is the cooldown between investigation actions (seconds).
	InvestigationCooldown float64

	// CooldownElapsed tracks the current cooldown timer (seconds).
	CooldownElapsed float64

	// RevealedFragments tracks fragment entity IDs that were revealed through investigation.
	RevealedFragments map[uint64]bool

	// InvestigationSkillBonus increases investigation radius and success chance.
	// Values: 0.0 (no bonus) to 1.0 (expert investigator).
	// Higher values increase radius by up to 50% and improve hidden fragment detection.
	InvestigationSkillBonus float64

	// TotalInvestigations tracks how many times this player has investigated.
	TotalInvestigations int

	// LastInvestigationTime is when the last successful investigation occurred.
	LastInvestigationTime time.Time
}

// Type returns the component type identifier.
func (c *InvestigationComponent) Type() string {
	return "investigation"
}

// NewInvestigationComponent creates a new investigation component with default values.
func NewInvestigationComponent() *InvestigationComponent {
	return &InvestigationComponent{
		InvestigationRadius:     3.0, // 3 tile radius
		IsInvestigating:         false,
		InvestigationDuration:   2.0, // 2 seconds to complete
		DiscoveredAreas:         make(map[string]bool),
		InvestigationCooldown:   1.0, // 1 second cooldown
		CooldownElapsed:         1.0, // Start with cooldown ready
		RevealedFragments:       make(map[uint64]bool),
		InvestigationSkillBonus: 0.0,
		TotalInvestigations:     0,
	}
}

// StartInvestigation begins an investigation action.
// Returns true if investigation started, false if on cooldown.
func (c *InvestigationComponent) StartInvestigation() bool {
	if c.CooldownElapsed < c.InvestigationCooldown {
		return false // Still on cooldown
	}

	c.IsInvestigating = true
	c.InvestigationStartTime = time.Now()
	c.CooldownElapsed = 0.0
	c.TotalInvestigations++
	return true
}

// StopInvestigation ends the current investigation action.
func (c *InvestigationComponent) StopInvestigation() {
	c.IsInvestigating = false
	c.LastInvestigationTime = time.Now()
}

// IsInvestigationComplete returns true if the investigation action has completed.
func (c *InvestigationComponent) IsInvestigationComplete() bool {
	if !c.IsInvestigating {
		return false
	}

	elapsed := time.Since(c.InvestigationStartTime).Seconds()
	return elapsed >= c.InvestigationDuration
}

// Update advances the cooldown timer.
func (c *InvestigationComponent) Update(deltaTime float64) {
	if c.CooldownElapsed < c.InvestigationCooldown {
		c.CooldownElapsed += deltaTime
		// Clamp to max cooldown
		if c.CooldownElapsed > c.InvestigationCooldown {
			c.CooldownElapsed = c.InvestigationCooldown
		}
	}
}

// GetEffectiveRadius returns the investigation radius with skill bonus applied.
func (c *InvestigationComponent) GetEffectiveRadius() float64 {
	// Skill bonus increases radius by up to 50%
	radiusBonus := c.InvestigationRadius * 0.5 * c.InvestigationSkillBonus
	return c.InvestigationRadius + radiusBonus
}

// HasDiscoveredArea returns true if the given grid position has been investigated.
func (c *InvestigationComponent) HasDiscoveredArea(gridX, gridY int) bool {
	key := makeGridKey(gridX, gridY)
	return c.DiscoveredAreas[key]
}

// MarkAreaDiscovered marks a grid position as investigated.
func (c *InvestigationComponent) MarkAreaDiscovered(gridX, gridY int) {
	key := makeGridKey(gridX, gridY)
	c.DiscoveredAreas[key] = true
}

// HasRevealedFragment returns true if the given fragment was revealed by this player.
func (c *InvestigationComponent) HasRevealedFragment(fragmentID uint64) bool {
	return c.RevealedFragments[fragmentID]
}

// MarkFragmentRevealed marks a fragment as revealed by this player's investigation.
func (c *InvestigationComponent) MarkFragmentRevealed(fragmentID uint64) {
	c.RevealedFragments[fragmentID] = true
}

// GetDetectionChance returns the probability (0.0-1.0) of detecting a hidden fragment.
// Base chance is 0.6, increased up to 1.0 with investigation skill bonus.
func (c *InvestigationComponent) GetDetectionChance() float64 {
	baseChance := 0.6
	bonusChance := 0.4 * c.InvestigationSkillBonus
	return baseChance + bonusChance
}

// makeGridKey creates a string key for grid coordinates.
func makeGridKey(gridX, gridY int) string {
	return fmt.Sprintf("%d,%d", gridX, gridY)
}
