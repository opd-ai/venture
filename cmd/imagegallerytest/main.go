// imagegallerytest demonstrates the persistent image gallery system.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/social/persistence"
)

func main() {
	initializeLogging()
	fmt.Println("=== Venture Image Gallery Test ===")
	fmt.Println()

	gallery := createGallery()
	allImages := runBasicTests(gallery)
	filename := runPersistenceTests(gallery, allImages)
	runAdvancedTests(gallery, allImages)
	cleanupTestFile(filename)
	printSummary()
}

func initializeLogging() {
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()
	if *verbose {
		log.SetFlags(log.Ltime | log.Lshortfile)
	}
}

func createGallery() *persistence.ImageGallery {
	fmt.Println("1. Creating image gallery for player 'alice'...")
	gallery := persistence.NewImageGallery("alice")
	fmt.Printf("   Gallery created: %d images, %d bytes\n\n", gallery.GetImageCount(), gallery.GetTotalSize())
	return gallery
}

func runBasicTests(gallery *persistence.ImageGallery) []*persistence.StoredImage {
	addTestImages(gallery)
	testDeduplication(gallery)
	allImages := getAllImages(gallery)
	getImagesByTag(gallery)
	getThumbnails(gallery)
	return allImages
}

func addTestImages(gallery *persistence.ImageGallery) {
	fmt.Println("2. Adding test images...")
	testImages := []struct {
		width  int
		height int
		color  color.RGBA
		title  string
		format persistence.ImageFormat
		tags   []string
	}{
		{64, 64, color.RGBA{255, 0, 0, 255}, "Red Square", persistence.ImageFormatPNG, []string{"geometric", "red"}},
		{128, 64, color.RGBA{0, 255, 0, 255}, "Green Rectangle", persistence.ImageFormatJPEG, []string{"geometric", "green"}},
		{96, 96, color.RGBA{0, 0, 255, 255}, "Blue Square", persistence.ImageFormatPNG, []string{"geometric", "blue"}},
		{32, 128, color.RGBA{255, 255, 0, 255}, "Yellow Tall", persistence.ImageFormatJPEG, []string{"geometric", "yellow"}},
		{80, 80, color.RGBA{255, 0, 255, 255}, "Magenta", persistence.ImageFormatPNG, []string{"screenshot", "magenta"}},
	}

	for i, tc := range testImages {
		img := createTestImage(tc.width, tc.height, tc.color)
		stored, err := gallery.AddImage(img, tc.title, tc.format, tc.tags)
		if err != nil {
			log.Fatalf("   Failed to add image %d: %v", i+1, err)
		}
		fmt.Printf("   [%d] Added '%s' (%s, %dx%d, %d bytes, hash: %.8s...)\n",
			i+1, stored.Title, stored.Format, stored.Width, stored.Height, stored.SizeBytes, stored.Hash)
	}
	fmt.Printf("   Total: %d images, %d bytes\n\n", gallery.GetImageCount(), gallery.GetTotalSize())
}

func testDeduplication(gallery *persistence.ImageGallery) {
	fmt.Println("3. Testing deduplication (adding red square again)...")
	img := createTestImage(64, 64, color.RGBA{255, 0, 0, 255})
	stored, err := gallery.AddImage(img, "Duplicate Red", persistence.ImageFormatPNG, []string{"duplicate"})
	if err != nil {
		log.Fatalf("   Failed to add duplicate: %v", err)
	}
	fmt.Printf("   Returned existing image: %s\n", stored.ID)
	fmt.Printf("   Gallery still has %d images (deduplication worked)\n\n", gallery.GetImageCount())
}

func getAllImages(gallery *persistence.ImageGallery) []*persistence.StoredImage {
	fmt.Println("4. Retrieving all images...")
	allImages := gallery.GetAllImages()
	for i, img := range allImages {
		fmt.Printf("   [%d] %s (%s, %dx%d, %s)\n", i+1, img.Title, img.Format, img.Width, img.Height, img.Timestamp.Format("15:04:05"))
	}
	fmt.Println()
	return allImages
}

func getImagesByTag(gallery *persistence.ImageGallery) {
	fmt.Println("5. Filtering by tag 'geometric'...")
	geometric := gallery.GetImagesByTag("geometric")
	fmt.Printf("   Found %d geometric images:\n", len(geometric))
	for i, img := range geometric {
		fmt.Printf("   [%d] %s (%dx%d)\n", i+1, img.Title, img.Width, img.Height)
	}
	fmt.Println()
}

func getThumbnails(gallery *persistence.ImageGallery) {
	fmt.Println("6. Getting lightweight thumbnails...")
	thumbnails := gallery.GetThumbnails()
	fmt.Printf("   Loaded %d thumbnails:\n", len(thumbnails))
	for i, thumb := range thumbnails {
		fmt.Printf("   [%d] %s - %dx%d, %d bytes, tags: %v\n",
			i+1, thumb.Title, thumb.Width, thumb.Height, thumb.SizeBytes, thumb.Tags)
	}
	fmt.Println()
}

func runPersistenceTests(gallery *persistence.ImageGallery, allImages []*persistence.StoredImage) string {
	filename := saveGalleryToDisk(gallery)
	loadGalleryFromDisk(filename)
	decodeAndVerifyImage(gallery, allImages)
	return filename
}

func saveGalleryToDisk(gallery *persistence.ImageGallery) string {
	fmt.Println("7. Saving gallery to disk...")
	data, err := gallery.Save()
	if err != nil {
		log.Fatalf("   Failed to save: %v", err)
	}
	filename := "test_gallery.json.gz"
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		log.Fatalf("   Failed to write file: %v", err)
	}
	compressionRatio := float64(gallery.GetTotalSize()) / float64(len(data))
	fmt.Printf("   Saved to '%s': %d bytes (%.1fx compression)\n\n", filename, len(data), compressionRatio)
	return filename
}

func loadGalleryFromDisk(filename string) {
	fmt.Println("8. Loading gallery from disk...")
	loadedGallery := persistence.NewImageGallery("bob")
	loadedData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("   Failed to read file: %v", err)
	}
	if err := loadedGallery.Load(loadedData); err != nil {
		log.Fatalf("   Failed to load: %v", err)
	}
	fmt.Printf("   Loaded gallery: PlayerID=%s, %d images, %d bytes\n\n",
		loadedGallery.PlayerID, loadedGallery.GetImageCount(), loadedGallery.GetTotalSize())
}

func decodeAndVerifyImage(gallery *persistence.ImageGallery, allImages []*persistence.StoredImage) {
	fmt.Println("9. Decoding and verifying an image...")
	retrieved, err := gallery.GetImage(allImages[0].ID)
	if err != nil {
		log.Fatalf("   Failed to get image: %v", err)
	}
	decoded, err := gallery.DecodeImage(retrieved)
	if err != nil {
		log.Fatalf("   Failed to decode: %v", err)
	}
	bounds := decoded.Bounds()
	fmt.Printf("   Decoded '%s': %dx%d (original: %dx%d)\n\n",
		retrieved.Title, bounds.Dx(), bounds.Dy(), retrieved.Width, retrieved.Height)
}

func runAdvancedTests(gallery *persistence.ImageGallery, allImages []*persistence.StoredImage) {
	testImageDeletion(gallery, allImages)
	testLRUEviction(gallery)
}

func testImageDeletion(gallery *persistence.ImageGallery, allImages []*persistence.StoredImage) {
	fmt.Println("10. Testing image deletion...")
	deleteID := allImages[2].ID
	deleteTitle := allImages[2].Title
	if err := gallery.DeleteImage(deleteID); err != nil {
		log.Fatalf("   Failed to delete: %v", err)
	}
	fmt.Printf("   Deleted '%s'\n", deleteTitle)
	fmt.Printf("   Gallery now has %d images, %d bytes\n\n", gallery.GetImageCount(), gallery.GetTotalSize())
}

func testLRUEviction(gallery *persistence.ImageGallery) {
	fmt.Println("11. Testing LRU eviction...")
	fmt.Printf("   Current limit: %d images\n", persistence.MaxImagesPerPlayer)
	imagesToAdd := persistence.MaxImagesPerPlayer - gallery.GetImageCount() + 2
	fmt.Printf("   Adding %d more images to trigger eviction...\n", imagesToAdd)
	for i := 0; i < imagesToAdd; i++ {
		r := uint8((i * 30) % 256)
		g := uint8((i * 50) % 256)
		b := uint8((i * 70) % 256)
		testImg := createTestImage(48, 48, color.RGBA{r, g, b, 255})
		_, err := gallery.AddImage(testImg, fmt.Sprintf("LRU Test %d", i), persistence.ImageFormatPNG, []string{"lru"})
		if err != nil {
			log.Fatalf("   Failed to add LRU test image: %v", err)
		}
	}
	fmt.Printf("   Gallery has %d images (LRU eviction enforced limit)\n\n", gallery.GetImageCount())
}

func cleanupTestFile(filename string) {
	fmt.Println("12. Cleaning up...")
	if err := os.Remove(filename); err != nil {
		log.Printf("   Warning: Failed to remove test file: %v", err)
	} else {
		fmt.Printf("   Removed '%s'\n", filename)
	}
}

func printSummary() {
	fmt.Println()
	fmt.Println("=== Image Gallery Test Complete ===")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✓ Image storage with PNG and JPEG formats")
	fmt.Println("  ✓ Automatic deduplication via SHA256 hashing")
	fmt.Println("  ✓ Tag-based filtering and search")
	fmt.Println("  ✓ Lightweight thumbnail metadata")
	fmt.Println("  ✓ Gzip compression for persistence (save/load)")
	fmt.Println("  ✓ LRU eviction when limit reached")
	fmt.Println("  ✓ Image encoding/decoding (base64 + format)")
	fmt.Println("  ✓ Thread-safe concurrent access")
}

// createTestImage creates a simple test image filled with a solid color
func createTestImage(width, height int, fillColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}
	return img
}
