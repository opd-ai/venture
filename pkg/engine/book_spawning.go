// Package engine provides book and bookshelf spawning utilities for V4.0 integration.
// This file implements functions to spawn procedurally generated books and bookshelves into the game world.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

var bookSpawnLog = logrus.New()

func init() {
	bookSpawnLog.SetReportCaller(true)
	bookSpawnLog.SetLevel(logrus.DebugLevel)
}

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
	if err := validateBookshelfSpawnInputs(terr, bookshelves, seed); err != nil {
		return 0, err
	}

	rng := rand.New(rand.NewSource(seed + 7000))
	libraryRooms := selectLibraryRooms(terr, rng)

	if len(libraryRooms) == 0 {
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "SpawnBookshelvesInTerrain",
		}).Error("No available rooms for bookshelf spawning")
		return 0, fmt.Errorf("no available rooms for bookshelf spawning")
	}

	return spawnBookshelvesInRooms(world, libraryRooms, bookshelves, rng), nil
}

// validateBookshelfSpawnInputs validates terrain and bookshelf data for spawning.
func validateBookshelfSpawnInputs(terr *terrain.Terrain, bookshelves []BookshelfSpawnData, seed int64) error {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":       "SpawnBookshelvesInTerrain",
		"seed":            seed,
		"num_bookshelves": len(bookshelves),
		"num_rooms":       len(terr.Rooms),
	}).Debug("Starting bookshelf spawning in terrain")

	if len(terr.Rooms) < 2 {
		bookSpawnLog.WithFields(logrus.Fields{
			"operation":  "SpawnBookshelvesInTerrain",
			"num_rooms":  len(terr.Rooms),
			"min_needed": 2,
		}).Error("Insufficient rooms for bookshelf spawning")
		return fmt.Errorf("insufficient rooms for bookshelf spawning (need at least 2, got %d)", len(terr.Rooms))
	}

	if len(bookshelves) == 0 {
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "SpawnBookshelvesInTerrain",
		}).Debug("No bookshelves to spawn, returning early")
		return fmt.Errorf("no bookshelves provided")
	}

	return nil
}

// selectLibraryRooms filters and returns suitable rooms for bookshelf placement.
func selectLibraryRooms(terr *terrain.Terrain, rng *rand.Rand) []*terrain.Room {
	libraryRooms := filterRoomsByArea(terr, 40)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":         "SpawnBookshelvesInTerrain",
		"large_rooms_found": len(libraryRooms),
		"threshold":         40,
	}).Debug("Completed large room filtering")

	if len(libraryRooms) == 0 {
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "SpawnBookshelvesInTerrain",
			"threshold": 25,
		}).Debug("No large rooms found, falling back to medium-sized rooms")
		libraryRooms = filterRoomsByArea(terr, 25)
	}

	rng.Shuffle(len(libraryRooms), func(i, j int) {
		libraryRooms[i], libraryRooms[j] = libraryRooms[j], libraryRooms[i]
	})

	return libraryRooms
}

// filterRoomsByArea returns rooms with area >= minArea, skipping first room.
func filterRoomsByArea(terr *terrain.Terrain, minArea int) []*terrain.Room {
	rooms := make([]*terrain.Room, 0)
	for i := 1; i < len(terr.Rooms); i++ {
		room := terr.Rooms[i]
		area := room.Width * room.Height
		if area >= minArea {
			rooms = append(rooms, room)
			bookSpawnLog.WithFields(logrus.Fields{
				"operation": "filterRoomsByArea",
				"room_idx":  i,
				"area":      area,
				"threshold": minArea,
			}).Debug("Selected room for library placement")
		}
	}
	return rooms
}

// spawnBookshelvesInRooms spawns bookshelf entities in selected rooms.
func spawnBookshelvesInRooms(world *World, libraryRooms []*terrain.Room, bookshelves []BookshelfSpawnData, rng *rand.Rand) int {
	spawned := 0
	for i, shelfData := range bookshelves {
		if i >= len(libraryRooms) {
			bookSpawnLog.WithFields(logrus.Fields{
				"operation":         "SpawnBookshelvesInTerrain",
				"shelf_idx":         i,
				"available_rooms":   len(libraryRooms),
				"remaining_shelves": len(bookshelves) - i,
			}).Warn("No more rooms available for remaining bookshelves")
			break
		}

		room := libraryRooms[i]
		spawnX, spawnY := calculateWallSpawnPosition(room, rng, i)
		entity := createBookshelfEntity(world, shelfData, spawnX, spawnY)

		if entity != nil {
			spawned++
			bookSpawnLog.WithFields(logrus.Fields{
				"operation": "SpawnBookshelvesInTerrain",
				"shelf_idx": i,
				"entity_id": entity.ID,
				"spawn_x":   spawnX,
				"spawn_y":   spawnY,
				"num_books": len(shelfData.Books),
				"spawned":   spawned,
			}).Info("Successfully spawned bookshelf entity")
		} else {
			bookSpawnLog.WithFields(logrus.Fields{
				"operation": "SpawnBookshelvesInTerrain",
				"shelf_idx": i,
			}).Error("Failed to create bookshelf entity")
		}
	}

	bookSpawnLog.WithFields(logrus.Fields{
		"operation":       "SpawnBookshelvesInTerrain",
		"total_spawned":   spawned,
		"requested":       len(bookshelves),
		"available_rooms": len(libraryRooms),
	}).Info("Completed bookshelf spawning")

	return spawned
}

// calculateWallSpawnPosition determines spawn coordinates along a room wall.
func calculateWallSpawnPosition(room *terrain.Room, rng *rand.Rand, shelfIdx int) (float64, float64) {
	wall := rng.Intn(4)
	var spawnX, spawnY float64

	switch wall {
	case 0: // Left wall
		spawnX = float64(room.X*32) + 16
		spawnY = float64((room.Y + room.Height/2) * 32)
		logWallPlacement(shelfIdx, "left", spawnX, spawnY)
	case 1: // Top wall
		spawnX = float64((room.X + room.Width/2) * 32)
		spawnY = float64(room.Y*32) + 16
		logWallPlacement(shelfIdx, "top", spawnX, spawnY)
	case 2: // Right wall
		spawnX = float64((room.X+room.Width-1)*32) - 16
		spawnY = float64((room.Y + room.Height/2) * 32)
		logWallPlacement(shelfIdx, "right", spawnX, spawnY)
	case 3: // Bottom wall
		spawnX = float64((room.X + room.Width/2) * 32)
		spawnY = float64((room.Y+room.Height-1)*32) - 16
		logWallPlacement(shelfIdx, "bottom", spawnX, spawnY)
	}

	return spawnX, spawnY
}

// logWallPlacement logs bookshelf placement details.
func logWallPlacement(shelfIdx int, wall string, x, y float64) {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "calculateWallSpawnPosition",
		"shelf_idx": shelfIdx,
		"wall":      wall,
		"spawn_x":   x,
		"spawn_y":   y,
	}).Debug("Positioned bookshelf against wall")
}

// createBookshelfEntity creates a bookshelf entity with pre-generated books.
func createBookshelfEntity(world *World, shelfData BookshelfSpawnData, x, y float64) *Entity {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":     "createBookshelfEntity",
		"x":             x,
		"y":             y,
		"num_books":     len(shelfData.Books),
		"shelf_size":    shelfData.ShelfSize,
		"collider_size": shelfData.ColliderSize,
	}).Debug("Creating bookshelf entity")

	entity := world.CreateEntity()
	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "createBookshelfEntity",
		"entity_id": entity.ID,
		"x":         x,
		"y":         y,
	}).Debug("Created base entity for bookshelf")

	// Add position
	entity.AddComponent(&PositionComponent{X: x, Y: y})
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"component_type": "position",
		"x":              x,
		"y":              y,
	}).Debug("Added position component")

	// Add bookshelf sprite
	sprite := NewSpriteComponent(float64(shelfData.ShelfSize), float64(shelfData.ShelfSize), shelfData.ShelfColor)
	sprite.Layer = 7 // Below NPCs/vehicles but above ground
	sprite.Visible = true
	entity.AddComponent(sprite)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"component_type": "sprite",
		"layer":          7,
		"size":           shelfData.ShelfSize,
	}).Debug("Added sprite component")

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
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"component_type": "collider",
		"width":          shelfData.ColliderSize,
		"height":         shelfData.ColliderSize,
		"layer":          3,
		"solid":          true,
	}).Debug("Added collider component")

	// Add bookshelf component (marks as container for books)
	bookshelfComp := NewBookshelfComponent(8, "fantasy") // Max 8 books per shelf
	entity.AddComponent(bookshelfComp)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"component_type": "bookshelf",
		"max_books":      8,
		"genre":          "fantasy",
	}).Debug("Added bookshelf component")

	// Spawn individual book entities and add them to the bookshelf
	booksAdded := 0
	for bookIdx, bookComp := range shelfData.Books {
		bookSpawnLog.WithFields(logrus.Fields{
			"operation":  "createBookshelfEntity",
			"entity_id":  entity.ID,
			"book_idx":   bookIdx,
			"book_type":  bookComp.BookType,
			"book_title": bookComp.Title,
		}).Debug("Processing book for bookshelf")

		bookEntity := createBookEntity(world, bookComp, x, y)
		if bookEntity != nil {
			bookshelfComp.AddBook(bookEntity.ID)
			booksAdded++
			bookSpawnLog.WithFields(logrus.Fields{
				"operation":      "createBookshelfEntity",
				"entity_id":      entity.ID,
				"book_idx":       bookIdx,
				"book_entity_id": bookEntity.ID,
				"book_type":      bookComp.BookType,
				"books_added":    booksAdded,
			}).Debug("Added book to bookshelf")
		} else {
			bookSpawnLog.WithFields(logrus.Fields{
				"operation": "createBookshelfEntity",
				"entity_id": entity.ID,
				"book_idx":  bookIdx,
				"book_type": bookComp.BookType,
			}).Error("Failed to create book entity")
		}
	}

	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"books_added":    booksAdded,
		"books_expected": len(shelfData.Books),
	}).Info("Completed book addition to bookshelf")

	// Add dialog component for interaction
	dialogProvider := NewBookshelfDialogProvider(bookshelfComp.GetBookCount())
	dialogComp := NewDialogComponent(dialogProvider)
	entity.AddComponent(dialogComp)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookshelfEntity",
		"entity_id":      entity.ID,
		"component_type": "dialog",
		"book_count":     bookshelfComp.GetBookCount(),
	}).Debug("Added dialog component")

	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "createBookshelfEntity",
		"entity_id": entity.ID,
		"x":         x,
		"y":         y,
		"num_books": booksAdded,
	}).Debug("Bookshelf entity creation complete")

	return entity
}

// createBookEntity creates a book entity from a BookComponent.
func createBookEntity(world *World, bookComp *BookComponent, shelfX, shelfY float64) *Entity {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":  "createBookEntity",
		"book_type":  bookComp.BookType,
		"book_title": bookComp.Title,
		"shelf_x":    shelfX,
		"shelf_y":    shelfY,
	}).Debug("Creating book entity")

	entity := world.CreateEntity()
	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "createBookEntity",
		"entity_id": entity.ID,
		"book_type": bookComp.BookType,
	}).Debug("Created base entity for book")

	// Add position (same as bookshelf initially)
	entity.AddComponent(&PositionComponent{X: shelfX, Y: shelfY})
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookEntity",
		"entity_id":      entity.ID,
		"component_type": "position",
		"x":              shelfX,
		"y":              shelfY,
	}).Debug("Added position component")

	// Add the book component
	entity.AddComponent(bookComp)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookEntity",
		"entity_id":      entity.ID,
		"component_type": "book",
		"book_type":      bookComp.BookType,
		"title":          bookComp.Title,
	}).Debug("Added book component")

	// Add sprite (book appearance)
	bookColor := getBookColor(bookComp.BookType)
	sprite := NewSpriteComponent(16, 16, bookColor) // Books are small
	sprite.Layer = 11                               // Above everything when picked up
	sprite.Visible = false                          // Start invisible (on shelf)
	entity.AddComponent(sprite)
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":      "createBookEntity",
		"entity_id":      entity.ID,
		"component_type": "sprite",
		"layer":          11,
		"size":           16,
		"visible":        false,
		"color_r":        bookColor.R,
		"color_g":        bookColor.G,
		"color_b":        bookColor.B,
	}).Debug("Added sprite component")

	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "createBookEntity",
		"entity_id": entity.ID,
		"book_type": bookComp.BookType,
		"title":     bookComp.Title,
	}).Debug("Book entity creation complete")

	return entity
}

// getBookColor returns the sprite color for a book type.
func getBookColor(bookType BookType) color.RGBA {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation": "getBookColor",
		"book_type": bookType,
	}).Debug("Determining book color based on type")

	var bookColor color.RGBA
	switch bookType {
	case BookTypeSkill:
		bookColor = color.RGBA{R: 100, G: 150, B: 200, A: 255} // Blue for skill books
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "blue",
		}).Debug("Assigned blue color for skill book")
	case BookTypeLore:
		bookColor = color.RGBA{R: 150, G: 100, B: 200, A: 255} // Purple for lore
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "purple",
		}).Debug("Assigned purple color for lore book")
	case BookTypeQuest:
		bookColor = color.RGBA{R: 200, G: 150, B: 50, A: 255} // Gold for quests
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "gold",
		}).Debug("Assigned gold color for quest book")
	case BookTypeRecipe:
		bookColor = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green for recipes
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "green",
		}).Debug("Assigned green color for recipe book")
	case BookTypeHistory:
		bookColor = color.RGBA{R: 150, G: 120, B: 80, A: 255} // Brown for history
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "brown",
		}).Debug("Assigned brown color for history book")
	default:
		bookColor = color.RGBA{R: 150, G: 150, B: 150, A: 255} // Gray default
		bookSpawnLog.WithFields(logrus.Fields{
			"operation": "getBookColor",
			"book_type": bookType,
			"color":     "gray",
		}).Warn("Unknown book type, assigned default gray color")
	}

	return bookColor
}

// NewBookshelfDialogProvider creates a dialog provider for bookshelves.
// INTEGRATION FIX [Category B]: Bookshelf Dialog Provider
// Gap: Bookshelves reuse merchant dialog instead of dedicated bookshelf dialog
// Fix: Create BookshelfDialogProvider in dialog_system.go with book-specific responses
// Roadmap: ROADMAP_V4.md Phase 23.2 - Lore Integration (bookshelf interaction)
// Temporary: Reuses MerchantDialogProvider with bookshelf context until dedicated provider added
func NewBookshelfDialogProvider(bookCount int) DialogProvider {
	bookSpawnLog.WithFields(logrus.Fields{
		"operation":  "NewBookshelfDialogProvider",
		"book_count": bookCount,
	}).Debug("Creating dialog provider for bookshelf")

	// For now, reuse merchant dialog provider with bookshelf context
	provider := NewMerchantDialogProvider(fmt.Sprintf("Bookshelf (%d books)", bookCount))

	bookSpawnLog.WithFields(logrus.Fields{
		"operation":  "NewBookshelfDialogProvider",
		"book_count": bookCount,
		"provider":   "MerchantDialogProvider",
	}).Debug("Created bookshelf dialog provider (using merchant dialog as temporary solution)")

	return provider
}
