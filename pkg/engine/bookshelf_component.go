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

// Serialize converts BookshelfComponent to bytes for network transmission.
// Format: bookCount(4) + books(8*N) + capacity(4) + requiresKey(1) + isLocked(1) + genreIDLen(4) + genreID(N)
func (b *BookshelfComponent) Serialize() []byte {
	bookCount := len(b.Books)
	genreLen := len(b.GenreID)
	size := 14 + (bookCount * 8) + genreLen

	buf := make([]byte, size)
	offset := 0

	// Book count (4 bytes)
	writeInt32(buf[offset:], int32(bookCount))
	offset += 4

	// Book IDs (8 bytes each)
	for _, bookID := range b.Books {
		writeUint64(buf[offset:], bookID)
		offset += 8
	}

	// Capacity (4 bytes)
	writeInt32(buf[offset:], int32(b.Capacity))
	offset += 4

	// RequiresKey (1 byte)
	writeBool(buf[offset:], b.RequiresKey)
	offset++

	// IsLocked (1 byte)
	writeBool(buf[offset:], b.IsLocked)
	offset++

	// GenreID (4 bytes length + data)
	offset += writeString(buf[offset:], b.GenreID)

	return buf
}

// Deserialize restores BookshelfComponent from bytes.
func (b *BookshelfComponent) Deserialize(data []byte) error {
	if len(data) < 14 {
		return ErrInvalidComponentData
	}

	offset := 0

	// Book count
	bookCount := int(readInt32(data[offset:]))
	offset += 4

	// Book IDs
	b.Books = make([]uint64, bookCount)
	for i := 0; i < bookCount; i++ {
		b.Books[i] = readUint64(data[offset:])
		offset += 8
	}

	// Capacity
	b.Capacity = int(readInt32(data[offset:]))
	offset += 4

	// RequiresKey
	b.RequiresKey = readBool(data[offset:])
	offset++

	// IsLocked
	b.IsLocked = readBool(data[offset:])
	offset++

	// GenreID
	genreID, consumed := readString(data[offset:])
	b.GenreID = genreID
	offset += consumed

	return nil
}
