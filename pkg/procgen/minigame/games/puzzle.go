// PuzzleGame implements sliding tile and pattern matching puzzles.
// This file contains the PuzzleGame implementation where players must
// rearrange a shuffled grid to match the solution pattern.
package games

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
)

// PuzzleGame implements sliding tiles and pattern matching puzzles.
// Players must solve the puzzle within a move limit.
// Game duration: 3-7 minutes typical.
//
// Phase 27.2: Mini-Game Types
type PuzzleGame struct {
	rng        *rand.Rand
	difficulty float64
	gridSize   int
	maxMoves   int
	moves      int
	grid       [][]int
	solution   [][]int
	completed  bool
	playerWon  bool
	// LastRender contains the most recent render output for external consumption.
	LastRender *RenderOutput
}

// NewPuzzleGame creates a new puzzle game instance.
func NewPuzzleGame() *PuzzleGame {
	return &PuzzleGame{}
}

// Initialize sets up the puzzle game with the given seed and difficulty.
func (p *PuzzleGame) Initialize(seed int64, difficulty float64) error {
	if difficulty < 0 || difficulty > 1.0 {
		return fmt.Errorf("difficulty must be between 0 and 1, got %.2f", difficulty)
	}

	p.rng = rand.New(rand.NewSource(seed))
	p.difficulty = difficulty

	// Scale grid size based on difficulty: 3x3 to 5x5
	p.gridSize = 3 + int(difficulty*2)
	p.maxMoves = p.gridSize * p.gridSize * 2 // More moves for larger grids

	// Generate solution grid (ordered numbers)
	p.solution = make([][]int, p.gridSize)
	for i := 0; i < p.gridSize; i++ {
		p.solution[i] = make([]int, p.gridSize)
		for j := 0; j < p.gridSize; j++ {
			p.solution[i][j] = i*p.gridSize + j + 1
		}
	}

	// Generate shuffled grid
	p.grid = make([][]int, p.gridSize)
	for i := 0; i < p.gridSize; i++ {
		p.grid[i] = make([]int, p.gridSize)
		copy(p.grid[i], p.solution[i])
	}

	// Shuffle deterministically
	shuffles := 20 + int(difficulty*30)
	for i := 0; i < shuffles; i++ {
		row1, col1 := p.rng.Intn(p.gridSize), p.rng.Intn(p.gridSize)
		row2, col2 := p.rng.Intn(p.gridSize), p.rng.Intn(p.gridSize)
		p.grid[row1][col1], p.grid[row2][col2] = p.grid[row2][col2], p.grid[row1][col1]
	}

	p.completed = false
	p.playerWon = false
	p.moves = 0

	return nil
}

// Update advances the puzzle game by making a simulated move.
func (p *PuzzleGame) Update(deltaTime float64) error {
	if p.completed {
		return nil
	}

	// Simulate player making a move toward solution
	p.makeMove()
	p.moves++

	// Check for completion
	if p.isSolved() {
		p.completed = true
		p.playerWon = true
	} else if p.moves >= p.maxMoves {
		p.completed = true
		p.playerWon = false
	}

	return nil
}

// makeMove simulates a move toward the solution.
func (p *PuzzleGame) makeMove() {
	// Simple AI: occasionally move a random tile closer to its solution position
	if p.rng.Float64() < 0.7-p.difficulty*0.3 { // Success rate decreases with difficulty
		// Find a misplaced tile
		for i := 0; i < p.gridSize; i++ {
			for j := 0; j < p.gridSize; j++ {
				if p.grid[i][j] != p.solution[i][j] {
					// Try to swap with correct position
					targetRow := (p.grid[i][j] - 1) / p.gridSize
					targetCol := (p.grid[i][j] - 1) % p.gridSize
					p.grid[i][j], p.grid[targetRow][targetCol] = p.grid[targetRow][targetCol], p.grid[i][j]
					return
				}
			}
		}
	} else {
		// Random move (simulates player mistake)
		row1, col1 := p.rng.Intn(p.gridSize), p.rng.Intn(p.gridSize)
		row2, col2 := p.rng.Intn(p.gridSize), p.rng.Intn(p.gridSize)
		p.grid[row1][col1], p.grid[row2][col2] = p.grid[row2][col2], p.grid[row1][col1]
	}
}

// isSolved checks if the puzzle is solved.
func (p *PuzzleGame) isSolved() bool {
	for i := 0; i < p.gridSize; i++ {
		for j := 0; j < p.gridSize; j++ {
			if p.grid[i][j] != p.solution[i][j] {
				return false
			}
		}
	}
	return true
}

// PrepareRender computes the current visual state for the puzzle game.
// This validates screen dimensions and populates internal render data.
// Implements engine.MiniGame interface.
func (p *PuzzleGame) PrepareRender(screenWidth, screenHeight int) error {
	if err := validateScreenDimensions(screenWidth, screenHeight); err != nil {
		return fmt.Errorf("puzzle game prepare render: %w", err)
	}
	if p.rng == nil {
		return fmt.Errorf("puzzle game prepare render: game not initialized")
	}

	status := "Playing"
	if p.completed {
		if p.playerWon {
			status = "Won"
		} else {
			status = "Lost"
		}
	}

	elements := make([]RenderElement, 0, 2+p.gridSize*p.gridSize)

	// Move counter
	elements = append(elements, RenderElement{
		Type:  "text",
		X:     screenWidth / 2,
		Y:     10,
		Label: fmt.Sprintf("Moves: %d / %d", p.moves, p.maxMoves),
	})

	// Grid tiles
	tileSize := (screenHeight - 60) / p.gridSize
	maxTile := (screenWidth - 20) / p.gridSize
	if tileSize > maxTile {
		tileSize = maxTile
	}
	startX := (screenWidth - p.gridSize*tileSize) / 2
	startY := 40

	for i := 0; i < p.gridSize; i++ {
		for j := 0; j < p.gridSize; j++ {
			correct := p.grid[i][j] == p.solution[i][j]
			elements = append(elements, RenderElement{
				Type:        "tile",
				X:           startX + j*tileSize,
				Y:           startY + i*tileSize,
				W:           tileSize - 2,
				H:           tileSize - 2,
				Label:       fmt.Sprintf("%d", p.grid[i][j]),
				Value:       float64(p.grid[i][j]),
				Highlighted: correct,
			})
		}
	}

	p.LastRender = &RenderOutput{
		Title:    "Puzzle",
		Status:   status,
		Width:    screenWidth,
		Height:   screenHeight,
		Elements: elements,
	}
	return nil
}

// GetRenderOutput returns the computed visual state from the last PrepareRender call.
// Implements engine.MiniGame interface.
func (p *PuzzleGame) GetRenderOutput() engine.MiniGameRenderOutput {
	if p.LastRender == nil {
		return nil
	}
	return p.LastRender
}

// Render draws the puzzle game to the screen.
// Computes visual state including the grid, solution comparison, and move counter.
// DEPRECATED: Use PrepareRender() and GetRenderOutput() instead.
// Kept for backward compatibility with existing code.
func (p *PuzzleGame) Render(screen engine.ImageProvider) error {
	w, h, err := validateScreen(screen)
	if err != nil {
		return fmt.Errorf("puzzle game render: %w", err)
	}
	return p.PrepareRender(w, h)
}

// IsComplete returns true when the game has finished.
func (p *PuzzleGame) IsComplete() bool {
	return p.completed
}

// GetReward returns the reward for solving the puzzle.
func (p *PuzzleGame) GetReward() *engine.Reward {
	if !p.completed || !p.playerWon {
		return nil
	}

	// Reward scales with difficulty and unused moves
	movesLeft := p.maxMoves - p.moves
	bonusMultiplier := 1.0 + float64(movesLeft)/float64(p.maxMoves)
	goldReward := int(40.0 * bonusMultiplier * (1.0 + p.difficulty))
	xpReward := 20.0 + (p.difficulty * 40.0)

	return &engine.Reward{
		Gold:  goldReward,
		XP:    xpReward,
		Items: nil,
	}
}
