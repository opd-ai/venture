package persistence

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"
)

// createTestImage creates a simple test image
func createTestImage(width, height int, fillColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}
	return img
}

func TestNewImageGallery(t *testing.T) {
	gallery := NewImageGallery("player1")

	if gallery.PlayerID != "player1" {
		t.Errorf("Expected PlayerID 'player1', got '%s'", gallery.PlayerID)
	}

	if len(gallery.Images) != 0 {
		t.Errorf("Expected 0 images, got %d", len(gallery.Images))
	}

	if gallery.TotalBytes != 0 {
		t.Errorf("Expected TotalBytes 0, got %d", gallery.TotalBytes)
	}
}

func TestAddImage(t *testing.T) {
	tests := []struct {
		name    string
		img     image.Image
		title   string
		format  ImageFormat
		tags    []string
		wantErr bool
	}{
		{
			name:    "valid PNG",
			img:     createTestImage(64, 64, color.RGBA{255, 0, 0, 255}),
			title:   "Test Image",
			format:  ImageFormatPNG,
			tags:    []string{"test", "red"},
			wantErr: false,
		},
		{
			name:    "valid JPEG",
			img:     createTestImage(128, 128, color.RGBA{0, 255, 0, 255}),
			title:   "JPEG Test",
			format:  ImageFormatJPEG,
			tags:    []string{"test", "green"},
			wantErr: false,
		},
		{
			name:    "nil image",
			img:     nil,
			title:   "Invalid",
			format:  ImageFormatPNG,
			tags:    nil,
			wantErr: true,
		},
		{
			name:    "unsupported format",
			img:     createTestImage(32, 32, color.RGBA{0, 0, 255, 255}),
			title:   "Bad Format",
			format:  ImageFormat("gif"),
			tags:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gallery := NewImageGallery("player1")
			stored, err := gallery.AddImage(tt.img, tt.title, tt.format, tt.tags)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if stored == nil {
					t.Error("Expected stored image, got nil")
					return
				}
				if stored.Title != tt.title {
					t.Errorf("Expected title '%s', got '%s'", tt.title, stored.Title)
				}
				if stored.Format != tt.format {
					t.Errorf("Expected format %s, got %s", tt.format, stored.Format)
				}
				if len(stored.Tags) != len(tt.tags) {
					t.Errorf("Expected %d tags, got %d", len(tt.tags), len(stored.Tags))
				}
				if stored.OwnerID != "player1" {
					t.Errorf("Expected OwnerID 'player1', got '%s'", stored.OwnerID)
				}
			}
		})
	}
}

func TestAddImageDeduplication(t *testing.T) {
	gallery := NewImageGallery("player1")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})

	// Add same image twice
	stored1, err := gallery.AddImage(img, "First", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("First AddImage failed: %v", err)
	}

	stored2, err := gallery.AddImage(img, "Second", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("Second AddImage failed: %v", err)
	}

	// Should return same image
	if stored1.ID != stored2.ID {
		t.Errorf("Expected deduplication, got different IDs: %s vs %s", stored1.ID, stored2.ID)
	}

	// Should only have one image
	if len(gallery.Images) != 1 {
		t.Errorf("Expected 1 image after deduplication, got %d", len(gallery.Images))
	}
}

func TestAddImageLRUEviction(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add MaxImagesPerPlayer + 5 images
	totalImages := MaxImagesPerPlayer + 5
	for i := 0; i < totalImages; i++ {
		// Create unique images (different colors)
		r := uint8(i % 256)
		img := createTestImage(32, 32, color.RGBA{r, 0, 0, 255})
		_, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
		if err != nil {
			t.Fatalf("AddImage failed at iteration %d: %v", i, err)
		}
	}

	// Should have exactly MaxImagesPerPlayer images
	if len(gallery.Images) != MaxImagesPerPlayer {
		t.Errorf("Expected %d images after LRU eviction, got %d", MaxImagesPerPlayer, len(gallery.Images))
	}
}

func TestGetImage(t *testing.T) {
	gallery := NewImageGallery("player1")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	// Get existing image
	retrieved, err := gallery.GetImage(stored.ID)
	if err != nil {
		t.Errorf("GetImage failed: %v", err)
	}
	if retrieved.ID != stored.ID {
		t.Errorf("Expected ID %s, got %s", stored.ID, retrieved.ID)
	}

	// Get non-existent image
	_, err = gallery.GetImage("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent image, got nil")
	}
}

func TestGetAllImages(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add 3 images
	for i := 0; i < 3; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 50), 0, 0, 255})
		_, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
		if err != nil {
			t.Fatalf("AddImage failed: %v", err)
		}
	}

	images := gallery.GetAllImages()
	if len(images) != 3 {
		t.Errorf("Expected 3 images, got %d", len(images))
	}
}

func TestGetImagesByTag(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add images with different tags
	img1 := createTestImage(32, 32, color.RGBA{255, 0, 0, 255})
	_, err := gallery.AddImage(img1, "Red", ImageFormatPNG, []string{"red", "test"})
	if err != nil {
		t.Fatalf("AddImage 1 failed: %v", err)
	}

	img2 := createTestImage(32, 32, color.RGBA{0, 255, 0, 255})
	_, err = gallery.AddImage(img2, "Green", ImageFormatPNG, []string{"green", "test"})
	if err != nil {
		t.Fatalf("AddImage 2 failed: %v", err)
	}

	img3 := createTestImage(32, 32, color.RGBA{0, 0, 255, 255})
	_, err = gallery.AddImage(img3, "Blue", ImageFormatPNG, []string{"blue"})
	if err != nil {
		t.Fatalf("AddImage 3 failed: %v", err)
	}

	// Get by tag "test"
	testImages := gallery.GetImagesByTag("test")
	if len(testImages) != 2 {
		t.Errorf("Expected 2 images with tag 'test', got %d", len(testImages))
	}

	// Get by tag "blue"
	blueImages := gallery.GetImagesByTag("blue")
	if len(blueImages) != 1 {
		t.Errorf("Expected 1 image with tag 'blue', got %d", len(blueImages))
	}

	// Get by non-existent tag
	noneImages := gallery.GetImagesByTag("nonexistent")
	if len(noneImages) != 0 {
		t.Errorf("Expected 0 images with tag 'nonexistent', got %d", len(noneImages))
	}
}

func TestDeleteImage(t *testing.T) {
	gallery := NewImageGallery("player1")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	initialSize := gallery.TotalBytes

	// Delete existing image
	err = gallery.DeleteImage(stored.ID)
	if err != nil {
		t.Errorf("DeleteImage failed: %v", err)
	}

	if len(gallery.Images) != 0 {
		t.Errorf("Expected 0 images after delete, got %d", len(gallery.Images))
	}

	if gallery.TotalBytes != 0 {
		t.Errorf("Expected TotalBytes 0 after delete, got %d", gallery.TotalBytes)
	}

	// Delete non-existent image
	err = gallery.DeleteImage("nonexistent")
	if err == nil {
		t.Error("Expected error deleting non-existent image, got nil")
	}

	_ = initialSize // Use variable to avoid unused warning
}

func TestGetImageCount(t *testing.T) {
	gallery := NewImageGallery("player1")

	if gallery.GetImageCount() != 0 {
		t.Errorf("Expected 0 images initially, got %d", gallery.GetImageCount())
	}

	// Add 3 images
	for i := 0; i < 3; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 50), 0, 0, 255})
		_, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
		if err != nil {
			t.Fatalf("AddImage failed: %v", err)
		}
	}

	if gallery.GetImageCount() != 3 {
		t.Errorf("Expected 3 images, got %d", gallery.GetImageCount())
	}
}

func TestGetTotalSize(t *testing.T) {
	gallery := NewImageGallery("player1")

	if gallery.GetTotalSize() != 0 {
		t.Errorf("Expected 0 bytes initially, got %d", gallery.GetTotalSize())
	}

	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	if gallery.GetTotalSize() != stored.SizeBytes {
		t.Errorf("Expected TotalSize %d, got %d", stored.SizeBytes, gallery.GetTotalSize())
	}
}

func TestDecodeImage(t *testing.T) {
	gallery := NewImageGallery("player1")
	originalImg := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})

	stored, err := gallery.AddImage(originalImg, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	// Decode image
	decodedImg, err := gallery.DecodeImage(stored)
	if err != nil {
		t.Fatalf("DecodeImage failed: %v", err)
	}

	// Verify dimensions
	bounds := decodedImg.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Expected dimensions 64x64, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Test nil image
	_, err = gallery.DecodeImage(nil)
	if err == nil {
		t.Error("Expected error decoding nil image, got nil")
	}
}

func TestImageGallerySaveLoad(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add test images
	for i := 0; i < 3; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 50), 0, 0, 255})
		tags := []string{fmt.Sprintf("tag%d", i)}
		_, err := gallery.AddImage(img, fmt.Sprintf("Image %d", i), ImageFormatPNG, tags)
		if err != nil {
			t.Fatalf("AddImage %d failed: %v", i, err)
		}
	}

	// Save gallery
	data, err := gallery.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new gallery and load
	newGallery := NewImageGallery("player2") // Different player ID initially
	err = newGallery.Load(data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded data
	if newGallery.PlayerID != "player1" {
		t.Errorf("Expected PlayerID 'player1', got '%s'", newGallery.PlayerID)
	}

	if len(newGallery.Images) != 3 {
		t.Errorf("Expected 3 images, got %d", len(newGallery.Images))
	}

	if newGallery.TotalBytes != gallery.TotalBytes {
		t.Errorf("Expected TotalBytes %d, got %d", gallery.TotalBytes, newGallery.TotalBytes)
	}

	// Verify individual images
	for i := 0; i < 3; i++ {
		if newGallery.Images[i].Title != gallery.Images[i].Title {
			t.Errorf("Image %d: Expected title '%s', got '%s'", i, gallery.Images[i].Title, newGallery.Images[i].Title)
		}
		if newGallery.Images[i].Hash != gallery.Images[i].Hash {
			t.Errorf("Image %d: Hash mismatch", i)
		}
	}
}

func TestImageGalleryClear(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add images
	for i := 0; i < 5; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 50), 0, 0, 255})
		_, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
		if err != nil {
			t.Fatalf("AddImage failed: %v", err)
		}
	}

	if len(gallery.Images) != 5 {
		t.Fatalf("Expected 5 images, got %d", len(gallery.Images))
	}

	// Clear gallery
	gallery.Clear()

	if len(gallery.Images) != 0 {
		t.Errorf("Expected 0 images after Clear, got %d", len(gallery.Images))
	}

	if gallery.TotalBytes != 0 {
		t.Errorf("Expected TotalBytes 0 after Clear, got %d", gallery.TotalBytes)
	}
}

func TestGetThumbnails(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add images
	for i := 0; i < 3; i++ {
		img := createTestImage(32+i*16, 32+i*16, color.RGBA{uint8(i * 50), 0, 0, 255})
		tags := []string{fmt.Sprintf("tag%d", i)}
		_, err := gallery.AddImage(img, fmt.Sprintf("Image %d", i), ImageFormatPNG, tags)
		if err != nil {
			t.Fatalf("AddImage %d failed: %v", i, err)
		}
	}

	thumbnails := gallery.GetThumbnails()

	if len(thumbnails) != 3 {
		t.Errorf("Expected 3 thumbnails, got %d", len(thumbnails))
	}

	for i, thumb := range thumbnails {
		if thumb.Title != fmt.Sprintf("Image %d", i) {
			t.Errorf("Thumbnail %d: Expected title 'Image %d', got '%s'", i, i, thumb.Title)
		}
		if thumb.Width != 32+i*16 {
			t.Errorf("Thumbnail %d: Expected width %d, got %d", i, 32+i*16, thumb.Width)
		}
		if len(thumb.Tags) != 1 {
			t.Errorf("Thumbnail %d: Expected 1 tag, got %d", i, len(thumb.Tags))
		}
	}
}

func TestImageGalleryConcurrency(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Add images concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			img := createTestImage(32, 32, color.RGBA{uint8(idx * 20), 0, 0, 255})
			_, err := gallery.AddImage(img, fmt.Sprintf("Image %d", idx), ImageFormatPNG, nil)
			if err != nil {
				t.Errorf("Concurrent AddImage failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 unique images
	if gallery.GetImageCount() != 10 {
		t.Errorf("Expected 10 images after concurrent adds, got %d", gallery.GetImageCount())
	}
}

// Benchmarks

func BenchmarkImageGalleryAddImage(b *testing.B) {
	gallery := NewImageGallery("player1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create unique images by varying color
		uniqueImg := createTestImage(64, 64, color.RGBA{uint8(i % 256), 0, 0, 255})
		_, err := gallery.AddImage(uniqueImg, "Test", ImageFormatPNG, nil)
		if err != nil {
			b.Fatalf("AddImage failed: %v", err)
		}
		if i%100 == 99 {
			gallery.Clear() // Prevent unlimited growth
		}
	}
}

func BenchmarkImageGalleryGetImage(b *testing.B) {
	gallery := NewImageGallery("player1")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, _ := gallery.AddImage(img, "Test", ImageFormatPNG, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gallery.GetImage(stored.ID)
		if err != nil {
			b.Fatalf("GetImage failed: %v", err)
		}
	}
}

func BenchmarkImageGalleryGetAllImages(b *testing.B) {
	gallery := NewImageGallery("player1")

	// Add 100 images
	for i := 0; i < 100; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i), 0, 0, 255})
		_, _ = gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gallery.GetAllImages()
	}
}

func BenchmarkImageGallerySave(b *testing.B) {
	gallery := NewImageGallery("player1")

	// Add 50 images
	for i := 0; i < 50; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 5), 0, 0, 255})
		_, _ = gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gallery.Save()
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkImageGalleryLoad(b *testing.B) {
	gallery := NewImageGallery("player1")

	// Add 50 images
	for i := 0; i < 50; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i * 5), 0, 0, 255})
		_, _ = gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	}

	data, _ := gallery.Save()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newGallery := NewImageGallery("player1")
		err := newGallery.Load(data)
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

func BenchmarkImageGalleryDecodeImage(b *testing.B) {
	gallery := NewImageGallery("player1")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, _ := gallery.AddImage(img, "Test", ImageFormatPNG, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gallery.DecodeImage(stored)
		if err != nil {
			b.Fatalf("DecodeImage failed: %v", err)
		}
	}
}

func BenchmarkImageGalleryGetThumbnails(b *testing.B) {
	gallery := NewImageGallery("player1")

	// Add 100 images
	for i := 0; i < 100; i++ {
		img := createTestImage(32, 32, color.RGBA{uint8(i), 0, 0, 255})
		_, _ = gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gallery.GetThumbnails()
	}
}

// MockTimeProvider implements TimeProvider for deterministic testing
type MockTimeProvider struct {
	fixedTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.fixedTime
}

func TestImageGalleryDeterministicTimestamps(t *testing.T) {
	// Create a fixed time for deterministic testing
	fixedTime := time.Date(2026, 2, 13, 12, 0, 0, 123456789, time.UTC)
	mockTP := &MockTimeProvider{fixedTime: fixedTime}

	gallery := NewImageGalleryWithTimeProvider("player1", mockTP)
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})

	stored, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	// Verify ID includes the fixed timestamp's UnixNano
	expectedID := fmt.Sprintf("player1-%d", fixedTime.UnixNano())
	if stored.ID != expectedID {
		t.Errorf("Expected deterministic ID '%s', got '%s'", expectedID, stored.ID)
	}

	// Verify timestamp matches fixed time
	if !stored.Timestamp.Equal(fixedTime) {
		t.Errorf("Expected deterministic timestamp %v, got %v", fixedTime, stored.Timestamp)
	}
}

func TestImageGalleryTimeProviderAfterLoad(t *testing.T) {
	// Create gallery with mock time provider
	fixedTime := time.Date(2026, 2, 13, 12, 0, 0, 123456789, time.UTC)
	mockTP := &MockTimeProvider{fixedTime: fixedTime}

	gallery := NewImageGalleryWithTimeProvider("player1", mockTP)
	img := createTestImage(32, 32, color.RGBA{255, 0, 0, 255})
	_, err := gallery.AddImage(img, "Test1", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	// Save gallery
	data, err := gallery.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new gallery with different time provider
	newTime := time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC)
	newMockTP := &MockTimeProvider{fixedTime: newTime}
	newGallery := NewImageGalleryWithTimeProvider("player2", newMockTP)
	err = newGallery.Load(data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Add new image - should use the new time provider
	newImg := createTestImage(32, 32, color.RGBA{0, 255, 0, 255})
	stored, err := newGallery.AddImage(newImg, "Test2", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage after load failed: %v", err)
	}

	// Verify new image uses the new time provider
	expectedID := fmt.Sprintf("player1-%d", newTime.UnixNano())
	if stored.ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, stored.ID)
	}

	if !stored.Timestamp.Equal(newTime) {
		t.Errorf("Expected timestamp %v, got %v", newTime, stored.Timestamp)
	}
}

func TestSetTimeProvider(t *testing.T) {
	gallery := NewImageGallery("player1")

	// Set a mock time provider
	fixedTime := time.Date(2026, 2, 13, 15, 0, 0, 0, time.UTC)
	mockTP := &MockTimeProvider{fixedTime: fixedTime}
	gallery.SetTimeProvider(mockTP)

	img := createTestImage(32, 32, color.RGBA{0, 0, 255, 255})
	stored, err := gallery.AddImage(img, "Test", ImageFormatPNG, nil)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	// Verify timestamp uses the mock time provider
	if !stored.Timestamp.Equal(fixedTime) {
		t.Errorf("Expected timestamp %v, got %v", fixedTime, stored.Timestamp)
	}
}
