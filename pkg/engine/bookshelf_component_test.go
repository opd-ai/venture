package engine

import (
	"testing"
)

func TestNewBookshelfComponent(t *testing.T) {
	bookshelf := NewBookshelfComponent(10, "fantasy")

	if bookshelf == nil {
		t.Fatal("NewBookshelfComponent returned nil")
	}
	if bookshelf.Capacity != 10 {
		t.Errorf("expected capacity 10, got %d", bookshelf.Capacity)
	}
	if bookshelf.GenreID != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", bookshelf.GenreID)
	}
	if bookshelf.IsLocked {
		t.Error("bookshelf should not be locked by default")
	}
	if bookshelf.RequiresKey {
		t.Error("bookshelf should not require key by default")
	}
	if len(bookshelf.Books) != 0 {
		t.Errorf("expected empty bookshelf, got %d books", len(bookshelf.Books))
	}
}

func TestBookshelfType(t *testing.T) {
	bookshelf := NewBookshelfComponent(5, "scifi")
	if bookshelf.Type() != "bookshelf" {
		t.Errorf("expected type bookshelf, got %s", bookshelf.Type())
	}
}

func TestAddBook(t *testing.T) {
	bookshelf := NewBookshelfComponent(3, "fantasy")

	// Add books
	if !bookshelf.AddBook(1) {
		t.Error("failed to add first book")
	}
	if !bookshelf.AddBook(2) {
		t.Error("failed to add second book")
	}
	if !bookshelf.AddBook(3) {
		t.Error("failed to add third book")
	}

	if bookshelf.GetBookCount() != 3 {
		t.Errorf("expected 3 books, got %d", bookshelf.GetBookCount())
	}

	// Try to add beyond capacity
	if bookshelf.AddBook(4) {
		t.Error("should not be able to add book beyond capacity")
	}
}

func TestAddBookUnlimited(t *testing.T) {
	bookshelf := NewBookshelfComponent(0, "fantasy") // 0 = unlimited

	// Add many books
	for i := uint64(1); i <= 100; i++ {
		if !bookshelf.AddBook(i) {
			t.Errorf("failed to add book %d to unlimited bookshelf", i)
		}
	}

	if bookshelf.GetBookCount() != 100 {
		t.Errorf("expected 100 books, got %d", bookshelf.GetBookCount())
	}
}

func TestRemoveBook(t *testing.T) {
	bookshelf := NewBookshelfComponent(5, "fantasy")

	// Add books
	bookshelf.AddBook(1)
	bookshelf.AddBook(2)
	bookshelf.AddBook(3)

	// Remove a book
	if !bookshelf.RemoveBook(2) {
		t.Error("failed to remove existing book")
	}
	if bookshelf.GetBookCount() != 2 {
		t.Errorf("expected 2 books after removal, got %d", bookshelf.GetBookCount())
	}

	// Try to remove non-existent book
	if bookshelf.RemoveBook(99) {
		t.Error("should not be able to remove non-existent book")
	}

	// Verify remaining books
	if !bookshelf.HasBook(1) {
		t.Error("book 1 should still be on shelf")
	}
	if bookshelf.HasBook(2) {
		t.Error("book 2 should not be on shelf after removal")
	}
	if !bookshelf.HasBook(3) {
		t.Error("book 3 should still be on shelf")
	}
}

func TestHasBook(t *testing.T) {
	bookshelf := NewBookshelfComponent(5, "fantasy")

	bookshelf.AddBook(10)
	bookshelf.AddBook(20)

	if !bookshelf.HasBook(10) {
		t.Error("should have book 10")
	}
	if !bookshelf.HasBook(20) {
		t.Error("should have book 20")
	}
	if bookshelf.HasBook(30) {
		t.Error("should not have book 30")
	}
}

func TestIsFull(t *testing.T) {
	bookshelf := NewBookshelfComponent(2, "fantasy")

	if bookshelf.IsFull() {
		t.Error("empty bookshelf should not be full")
	}

	bookshelf.AddBook(1)
	if bookshelf.IsFull() {
		t.Error("bookshelf with 1/2 books should not be full")
	}

	bookshelf.AddBook(2)
	if !bookshelf.IsFull() {
		t.Error("bookshelf with 2/2 books should be full")
	}

	// Test unlimited capacity
	unlimited := NewBookshelfComponent(0, "fantasy")
	for i := uint64(1); i <= 100; i++ {
		unlimited.AddBook(i)
	}
	if unlimited.IsFull() {
		t.Error("unlimited bookshelf should never be full")
	}
}

func TestIsEmpty(t *testing.T) {
	bookshelf := NewBookshelfComponent(5, "fantasy")

	if !bookshelf.IsEmpty() {
		t.Error("new bookshelf should be empty")
	}

	bookshelf.AddBook(1)
	if bookshelf.IsEmpty() {
		t.Error("bookshelf with books should not be empty")
	}

	bookshelf.RemoveBook(1)
	if !bookshelf.IsEmpty() {
		t.Error("bookshelf should be empty after removing all books")
	}
}

func TestBookshelfLocking(t *testing.T) {
	bookshelf := NewBookshelfComponent(5, "fantasy")

	// Lock the bookshelf
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "golden_key"

	if !bookshelf.IsLocked {
		t.Error("bookshelf should be locked")
	}
	if !bookshelf.RequiresKey {
		t.Error("bookshelf should require key")
	}
	if bookshelf.KeyItemID != "golden_key" {
		t.Errorf("expected key ID golden_key, got %s", bookshelf.KeyItemID)
	}
}

func TestGetBookCount(t *testing.T) {
	bookshelf := NewBookshelfComponent(10, "fantasy")

	if bookshelf.GetBookCount() != 0 {
		t.Errorf("expected 0 books, got %d", bookshelf.GetBookCount())
	}

	bookshelf.AddBook(1)
	bookshelf.AddBook(2)
	bookshelf.AddBook(3)

	if bookshelf.GetBookCount() != 3 {
		t.Errorf("expected 3 books, got %d", bookshelf.GetBookCount())
	}

	bookshelf.RemoveBook(2)

	if bookshelf.GetBookCount() != 2 {
		t.Errorf("expected 2 books after removal, got %d", bookshelf.GetBookCount())
	}
}
