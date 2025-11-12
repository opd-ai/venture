// Package engine provides bookshelf component for Phase 23.2.
// Bookshelves are interactive environmental objects that can contain books.
// Players can interact with them to browse and read books.
package engine

// BookshelfComponent represents an interactive bookshelf entity.
// Bookshelves can be found in dungeons, towns, and libraries.
type BookshelfComponent struct {
	// Books contains entity IDs of books on this shelf
	Books []uint64

	// Capacity is the maximum number of books this shelf can hold (0 = unlimited)
	Capacity int

	// RequiresKey indicates if a key is required to access this bookshelf
	RequiresKey bool

	// KeyItemID is the item ID required to unlock this bookshelf (if RequiresKey is true)
	KeyItemID string

	// IsLocked indicates if the bookshelf is currently locked
	IsLocked bool

	// GenreID for thematic consistency (affects which books spawn)
	GenreID string
}

// Type returns the component type identifier.
func (b BookshelfComponent) Type() string {
	return "bookshelf"
}

// NewBookshelfComponent creates a new bookshelf component.
func NewBookshelfComponent(capacity int, genreID string) *BookshelfComponent {
	return &BookshelfComponent{
		Books:       make([]uint64, 0, capacity),
		Capacity:    capacity,
		RequiresKey: false,
		IsLocked:    false,
		GenreID:     genreID,
	}
}

// AddBook adds a book entity ID to this bookshelf.
// Returns false if the shelf is full.
func (b *BookshelfComponent) AddBook(bookEntityID uint64) bool {
	if b.Capacity > 0 && len(b.Books) >= b.Capacity {
		return false // Shelf is full
	}

	b.Books = append(b.Books, bookEntityID)
	return true
}

// RemoveBook removes a book entity ID from this bookshelf.
// Returns true if the book was found and removed.
func (b *BookshelfComponent) RemoveBook(bookEntityID uint64) bool {
	for i, id := range b.Books {
		if id == bookEntityID {
			// Remove book by replacing with last element and shrinking slice
			b.Books[i] = b.Books[len(b.Books)-1]
			b.Books = b.Books[:len(b.Books)-1]
			return true
		}
	}
	return false
}

// HasBook checks if a book entity ID is on this bookshelf.
func (b *BookshelfComponent) HasBook(bookEntityID uint64) bool {
	for _, id := range b.Books {
		if id == bookEntityID {
			return true
		}
	}
	return false
}

// IsFull returns true if the bookshelf has reached its capacity.
func (b *BookshelfComponent) IsFull() bool {
	return b.Capacity > 0 && len(b.Books) >= b.Capacity
}

// IsEmpty returns true if the bookshelf has no books.
func (b *BookshelfComponent) IsEmpty() bool {
	return len(b.Books) == 0
}

// GetBookCount returns the number of books currently on this shelf.
func (b *BookshelfComponent) GetBookCount() int {
	return len(b.Books)
}
