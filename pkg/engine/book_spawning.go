// Package engine provides book and bookshelf spawning utilities for V4.0 integration.
// This file implements functions to spawn procedurally generated books and bookshelves into the game world.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// BookshelfSpawnData contains the necessary information to spawn a bookshelf with books.
// This avoids import cycles by not depending on pkg/procgen/book directly.
type BookshelfSpawnData struct {
	Books        []*BookComponent // Pre-generated books (3-8 per shelf)
	ShelfColor   color.RGBA       // Bookshelf sprite color
	ShelfSize    int              // Sprite size
	ColliderSize float64          // Collider dimensions
}

// SpawnBookshelvesInTerrain spawns procedurally generated bookshelves with books into terrain rooms.
// Places bookshelves in larger rooms, with each shelf containing 3-8 books.
// Returns the number of bookshelves spawned.
//
// Note: This function expects book generation to be done externally to avoid import cycles.
// Call this from cmd/client with books generated via bookgen.NewGenerator().
func SpawnBookshelvesInTerrain(world *World, terr *terrain.Terrain, bookshelves []BookshelfSpawnData, seed int64) (int, error) {
	if len(terr.Rooms) < 2 {
		return 0, fmt.Errorf("insufficient rooms for bookshelf spawning (need at least 2, got %d)", len(terr.Rooms))
	}

	if len(bookshelves) == 0 {
		return 0, nil
	}

	// Create RNG for room selection
	rng := rand.New(rand.NewSource(seed + 7000))

	// Select larger rooms for bookshelf placement (library-like rooms)
	// Filter rooms with area >= 40 tiles (minimum 8x5 or 5x8)
	libraryRooms := make([]*terrain.Room, 0)
	for i := 1; i < len(terr.Rooms); i++ { // Skip first room (player spawn)
		room := terr.Rooms[i]
		area := room.Width * room.Height
		if area >= 40 {
			libraryRooms = append(libraryRooms, room)
		}
	}

	// If no large rooms, fall back to medium-sized rooms (>= 25 tiles)
	if len(libraryRooms) == 0 {
		for i := 1; i < len(terr.Rooms); i++ {
			room := terr.Rooms[i]
			area := room.Width * room.Height
			if area >= 25 {
				libraryRooms = append(libraryRooms, room)
			}
		}
	}

	if len(libraryRooms) == 0 {
		return 0, fmt.Errorf("no available rooms for bookshelf spawning")
	}

	// Shuffle rooms
	rng.Shuffle(len(libraryRooms), func(i, j int) {
		libraryRooms[i], libraryRooms[j] = libraryRooms[i], libraryRooms[j]
	})

	spawned := 0
	for i, shelfData := range bookshelves {
		if i >= len(libraryRooms) {
			break // No more rooms available
		}

		room := libraryRooms[i]

		// Place bookshelf against a wall (edges of room)
		// Randomly choose wall: 0=left, 1=top, 2=right, 3=bottom
		wall := rng.Intn(4)
		var spawnX, spawnY float64

		switch wall {
		case 0: // Left wall
			spawnX = float64(room.X*32) + 16 // Near left edge
			spawnY = float64((room.Y + room.Height/2) * 32)
		case 1: // Top wall
			spawnX = float64((room.X + room.Width/2) * 32)
			spawnY = float64(room.Y*32) + 16 // Near top edge
		case 2: // Right wall
			spawnX = float64((room.X+room.Width-1)*32) - 16 // Near right edge
			spawnY = float64((room.Y + room.Height/2) * 32)
		case 3: // Bottom wall
			spawnX = float64((room.X + room.Width/2) * 32)
			spawnY = float64((room.Y+room.Height-1)*32) - 16 // Near bottom edge
		}

		// Create bookshelf entity with books
		entity := createBookshelfEntity(world, shelfData, spawnX, spawnY)
		if entity != nil {
			spawned++
		}
	}

	return spawned, nil
}

// createBookshelfEntity creates a bookshelf entity with pre-generated books.
func createBookshelfEntity(world *World, shelfData BookshelfSpawnData, x, y float64) *Entity {
	entity := world.CreateEntity()

	// Add position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Add bookshelf sprite
	sprite := NewSpriteComponent(float64(shelfData.ShelfSize), float64(shelfData.ShelfSize), shelfData.ShelfColor)
	sprite.Layer = 7 // Below NPCs/vehicles but above ground
	sprite.Visible = true
	entity.AddComponent(sprite)

	// Add collider (bookshelves are solid furniture)
	entity.AddComponent(&ColliderComponent{
		Width:     shelfData.ColliderSize,
		Height:    shelfData.ColliderSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     3, // Furniture layer
		OffsetX:   -shelfData.ColliderSize / 2,
		OffsetY:   -shelfData.ColliderSize / 2,
	})

	// Add bookshelf component (marks as container for books)
	bookshelfComp := NewBookshelfComponent(8, "fantasy") // Max 8 books per shelf
	entity.AddComponent(bookshelfComp)

	// Spawn individual book entities and add them to the bookshelf
	for _, bookComp := range shelfData.Books {
		bookEntity := createBookEntity(world, bookComp, x, y)
		if bookEntity != nil {
			bookshelfComp.AddBook(bookEntity.ID)
		}
	}

	// Add dialog component for interaction
	dialogProvider := NewBookshelfDialogProvider(bookshelfComp.GetBookCount())
	dialogComp := NewDialogComponent(dialogProvider)
	entity.AddComponent(dialogComp)

	return entity
}

// createBookEntity creates a book entity from a BookComponent.
func createBookEntity(world *World, bookComp *BookComponent, shelfX, shelfY float64) *Entity {
	entity := world.CreateEntity()

	// Add position (same as bookshelf initially)
	entity.AddComponent(&PositionComponent{X: shelfX, Y: shelfY})

	// Add the book component
	entity.AddComponent(bookComp)

	// Add sprite (book appearance)
	bookColor := getBookColor(bookComp.BookType)
	sprite := NewSpriteComponent(16, 16, bookColor) // Books are small
	sprite.Layer = 11                               // Above everything when picked up
	sprite.Visible = false                          // Start invisible (on shelf)
	entity.AddComponent(sprite)

	return entity
}

// getBookColor returns the sprite color for a book type.
func getBookColor(bookType BookType) color.RGBA {
	switch bookType {
	case BookTypeSkill:
		return color.RGBA{R: 100, G: 150, B: 200, A: 255} // Blue for skill books
	case BookTypeLore:
		return color.RGBA{R: 150, G: 100, B: 200, A: 255} // Purple for lore
	case BookTypeQuest:
		return color.RGBA{R: 200, G: 150, B: 50, A: 255} // Gold for quests
	case BookTypeRecipe:
		return color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green for recipes
	case BookTypeHistory:
		return color.RGBA{R: 150, G: 120, B: 80, A: 255} // Brown for history
	default:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255} // Gray default
	}
}

// NewBookshelfDialogProvider creates a dialog provider for bookshelves.
// This is a placeholder - actual implementation would be in dialog_system.go
func NewBookshelfDialogProvider(bookCount int) DialogProvider {
	// For now, reuse merchant dialog provider with bookshelf context
	// TODO: Create dedicated BookshelfDialogProvider in dialog_system.go
	return NewMerchantDialogProvider(fmt.Sprintf("Bookshelf (%d books)", bookCount))
}
