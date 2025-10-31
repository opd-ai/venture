// Package engine provides puzzle components for Phase 11.2.
// Procedural Puzzle Generation System
//
// This file implements components for constraint-solving puzzles including
// pressure plates, lever sequences, block pushing, and timed challenges.
package engine

import (
	"fmt"
	"time"
)

// PuzzleState represents the current state of a puzzle.
type PuzzleState int

const (
	// PuzzleStateUnsolved indicates puzzle is not yet solved
	PuzzleStateUnsolved PuzzleState = iota
	// PuzzleStateSolving indicates puzzle is being actively worked on
	PuzzleStateSolving
	// PuzzleStateSolved indicates puzzle has been successfully solved
	PuzzleStateSolved
	// PuzzleStateFailed indicates puzzle was failed (if applicable)
	PuzzleStateFailed
)

// String returns human-readable puzzle state name.
func (s PuzzleState) String() string {
	switch s {
	case PuzzleStateUnsolved:
		return "Unsolved"
	case PuzzleStateSolving:
		return "Solving"
	case PuzzleStateSolved:
		return "Solved"
	case PuzzleStateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// PuzzleType defines the category of puzzle.
type PuzzleType string

const (
	// PuzzleTypePressurePlate requires stepping on specific plates
	PuzzleTypePressurePlate PuzzleType = "pressure_plate"
	// PuzzleTypeLeverSequence requires activating levers in correct order
	PuzzleTypeLeverSequence PuzzleType = "lever_sequence"
	// PuzzleTypeBlockPushing requires pushing blocks to specific locations
	PuzzleTypeBlockPushing PuzzleType = "block_pushing"
	// PuzzleTypeTimedChallenge requires completing within time limit
	PuzzleTypeTimedChallenge PuzzleType = "timed_challenge"
	// PuzzleTypeMemoryPattern requires repeating a shown pattern
	PuzzleTypeMemoryPattern PuzzleType = "memory_pattern"
	// PuzzleTypeColorMatching requires matching colors/symbols
	PuzzleTypeColorMatching PuzzleType = "color_matching"
)

// PuzzleComponent tracks puzzle state and solution.
type PuzzleComponent struct {
	// PuzzleID uniquely identifies this puzzle
	PuzzleID string

	// PuzzleType defines puzzle category
	PuzzleType PuzzleType

	// State tracks current puzzle state
	State PuzzleState

	// Solution stores the correct solution (element IDs in order)
	Solution []string

	// CurrentProgress tracks player progress (activated elements)
	CurrentProgress []string

	// ElementIDs lists all puzzle element entity IDs
	ElementIDs []uint64

	// Difficulty ranges from 1 (simple) to 10 (complex)
	Difficulty int

	// TimeLimit for timed puzzles (0 = no limit)
	TimeLimit float64

	// TimeElapsed tracks time spent on puzzle
	TimeElapsed float64

	// MaxAttempts for puzzles with retry limits (0 = unlimited)
	MaxAttempts int

	// Attempts tracks number of tries
	Attempts int

	// RewardEntityID is the entity unlocked on success (door, chest, etc.)
	RewardEntityID uint64

	// SolvedBy tracks which player solved it (multiplayer)
	SolvedBy uint64

	// SolvedAt tracks when puzzle was solved
	SolvedAt time.Time

	// Hints available for this puzzle
	Hints []string

	// HintsUsed tracks which hints were revealed
	HintsUsed int
}

// Type returns the component type identifier (implements Component interface).
func (p *PuzzleComponent) Type() string {
	return "puzzle"
}

// NewPuzzleComponent creates a new puzzle component.
func NewPuzzleComponent(puzzleID string, puzzleType PuzzleType, difficulty int) *PuzzleComponent {
	return &PuzzleComponent{
		PuzzleID:        puzzleID,
		PuzzleType:      puzzleType,
		State:           PuzzleStateUnsolved,
		Solution:        []string{},
		CurrentProgress: []string{},
		ElementIDs:      []uint64{},
		Difficulty:      difficulty,
		TimeLimit:       0,
		TimeElapsed:     0,
		MaxAttempts:     0,
		Attempts:        0,
		RewardEntityID:  0,
		SolvedBy:        0,
		SolvedAt:        time.Time{},
		Hints:           []string{},
		HintsUsed:       0,
	}
}

// AddElement registers a puzzle element entity.
func (p *PuzzleComponent) AddElement(elementID uint64, elementName string) {
	p.ElementIDs = append(p.ElementIDs, elementID)
}

// RecordProgress adds an activated element to progress.
func (p *PuzzleComponent) RecordProgress(elementName string) {
	p.CurrentProgress = append(p.CurrentProgress, elementName)
}

// CheckSolution verifies if current progress matches solution.
func (p *PuzzleComponent) CheckSolution() bool {
	if len(p.CurrentProgress) != len(p.Solution) {
		return false
	}

	for i, expected := range p.Solution {
		if i >= len(p.CurrentProgress) || p.CurrentProgress[i] != expected {
			return false
		}
	}

	return true
}

// IsSolved returns true if puzzle is in solved state.
func (p *PuzzleComponent) IsSolved() bool {
	return p.State == PuzzleStateSolved
}

// MarkSolved sets puzzle as solved and records solver.
func (p *PuzzleComponent) MarkSolved(solverID uint64) {
	p.State = PuzzleStateSolved
	p.SolvedBy = solverID
	p.SolvedAt = time.Now()
}

// Reset clears progress and resets puzzle state.
func (p *PuzzleComponent) Reset() {
	p.State = PuzzleStateUnsolved
	p.CurrentProgress = []string{}
	p.TimeElapsed = 0
	p.Attempts++
}

// HasTimedOut returns true if time limit exceeded.
func (p *PuzzleComponent) HasTimedOut() bool {
	return p.TimeLimit > 0 && p.TimeElapsed >= p.TimeLimit
}

// HasMaxAttemptsReached returns true if attempt limit reached.
func (p *PuzzleComponent) HasMaxAttemptsReached() bool {
	return p.MaxAttempts > 0 && p.Attempts >= p.MaxAttempts
}

// GetProgressPercent returns completion percentage (0-100).
func (p *PuzzleComponent) GetProgressPercent() float64 {
	if len(p.Solution) == 0 {
		return 0
	}
	return (float64(len(p.CurrentProgress)) / float64(len(p.Solution))) * 100.0
}

// GetNextHint returns the next available hint.
func (p *PuzzleComponent) GetNextHint() (string, error) {
	if p.HintsUsed >= len(p.Hints) {
		return "", fmt.Errorf("no more hints available")
	}

	hint := p.Hints[p.HintsUsed]
	p.HintsUsed++
	return hint, nil
}

// PuzzleElementComponent represents an interactive puzzle element.
type PuzzleElementComponent struct {
	// ElementName uniquely identifies this element within the puzzle
	ElementName string

	// PuzzleID links this element to its parent puzzle
	PuzzleID string

	// ElementType describes the visual/logical type
	ElementType string // "pressure_plate", "lever", "block", "rune", etc.

	// State tracks current activation state
	State int // Element-specific state (0=inactive, 1=active, etc.)

	// TargetState is the required state for solution
	TargetState int

	// InteractionRange is the distance for interaction (pixels)
	InteractionRange float64

	// CooldownTime prevents rapid re-activation (seconds)
	CooldownTime float64

	// CooldownElapsed tracks cooldown timer
	CooldownElapsed float64

	// IsInteractable indicates if element can currently be interacted with
	IsInteractable bool

	// RequiresItem optionally requires specific item to interact
	RequiresItem string

	// ActivationSound for audio feedback
	ActivationSound string

	// VisualFeedback for state changes (particle effect, color change, etc.)
	VisualFeedback string
}

// Type returns the component type identifier.
func (e *PuzzleElementComponent) Type() string {
	return "puzzleElement"
}

// NewPuzzleElementComponent creates a new puzzle element.
func NewPuzzleElementComponent(elementName, puzzleID, elementType string) *PuzzleElementComponent {
	return &PuzzleElementComponent{
		ElementName:      elementName,
		PuzzleID:         puzzleID,
		ElementType:      elementType,
		State:            0,
		TargetState:      0,
		InteractionRange: 32.0, // Default 32 pixels
		CooldownTime:     0.5,  // Default 0.5 second cooldown
		CooldownElapsed:  0,
		IsInteractable:   true,
		RequiresItem:     "",
		ActivationSound:  "puzzle_activate",
		VisualFeedback:   "sparkle",
	}
}

// Activate toggles element state (for binary elements).
func (e *PuzzleElementComponent) Activate() bool {
	if !e.IsInteractable || e.CooldownElapsed < e.CooldownTime {
		return false
	}

	// Toggle state
	if e.State == 0 {
		e.State = 1
	} else {
		e.State = 0
	}

	// Reset cooldown
	e.CooldownElapsed = 0

	return true
}

// SetState sets element to specific state.
func (e *PuzzleElementComponent) SetState(state int) bool {
	if !e.IsInteractable || e.CooldownElapsed < e.CooldownTime {
		return false
	}

	e.State = state
	e.CooldownElapsed = 0
	return true
}

// IsInCorrectState returns true if element matches target state.
func (e *PuzzleElementComponent) IsInCorrectState() bool {
	return e.State == e.TargetState
}

// IsOnCooldown returns true if cooldown is active.
func (e *PuzzleElementComponent) IsOnCooldown() bool {
	return e.CooldownElapsed < e.CooldownTime
}

// UpdateCooldown progresses cooldown timer.
func (e *PuzzleElementComponent) UpdateCooldown(deltaTime float64) {
	if e.CooldownElapsed < e.CooldownTime {
		e.CooldownElapsed += deltaTime
	}
}

// Reset returns element to initial state.
func (e *PuzzleElementComponent) Reset() {
	e.State = 0
	e.CooldownElapsed = 0
	e.IsInteractable = true
}
