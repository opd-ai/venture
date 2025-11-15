package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/opd-ai/venture/pkg/network"
	"github.com/sirupsen/logrus"
)

var (
	width       = flag.Int("width", 256, "Image width in pixels")
	height      = flag.Int("height", 256, "Image height in pixels")
	format      = flag.String("format", "png", "Image format (png, jpeg, gif)")
	uploadCount = flag.Int("count", 1, "Number of images to upload")
	testMode    = flag.String("test", "", "Test mode: upload, thumbnail, chunked, expiry, ratelimit, moderation")
	outputDir   = flag.String("output", "./test_images", "Output directory for generated images")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	// Setup logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Info("Image Sharing System Test Tool")
	logrus.Infof("Test mode: %s", *testMode)

	switch *testMode {
	case "upload":
		testUpload()
	case "thumbnail":
		testThumbnail()
	case "chunked":
		testChunkedTransfer()
	case "expiry":
		testExpiry()
	case "ratelimit":
		testRateLimit()
	case "moderation":
		testModeration()
	case "":
		logrus.Error("No test mode specified. Use -test flag")
		flag.Usage()
		os.Exit(1)
	default:
		logrus.Errorf("Unknown test mode: %s", *testMode)
		flag.Usage()
		os.Exit(1)
	}
}

// createTestImage creates a test image with gradient pattern
func createTestImage(width, height int, format string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Save to buffer
	var buf []byte
	switch format {
	case "png":
		file, err := os.CreateTemp("", "test-*.png")
		if err != nil {
			return nil, err
		}
		defer os.Remove(file.Name())
		if err := png.Encode(file, img); err != nil {
			return nil, err
		}
		file.Close()
		buf, err = os.ReadFile(file.Name())
		if err != nil {
			return nil, err
		}
	case "jpeg":
		file, err := os.CreateTemp("", "test-*.jpg")
		if err != nil {
			return nil, err
		}
		defer os.Remove(file.Name())
		if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
		file.Close()
		buf, err = os.ReadFile(file.Name())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return buf, nil
}

// testUpload tests basic image upload functionality
func testUpload() {
	logrus.Info("Testing image upload...")

	manager := network.NewImageManager()

	// Set thumbnail relay callback
	thumbnailCount := 0
	manager.SetThumbnailRelayCallback(func(metadata *network.ImageMetadata, thumbnail []byte) {
		thumbnailCount++
		logrus.Infof("Thumbnail relayed for image %s (size: %d bytes)", metadata.ImageID, len(thumbnail))
	})

	start := time.Now()

	for i := 0; i < *uploadCount; i++ {
		// Create test image
		data, err := createTestImage(*width, *height, *format)
		if err != nil {
			logrus.Fatalf("Failed to create test image %d: %v", i, err)
		}

		logrus.Infof("Upload %d: Created %s image (%dx%d, %d bytes)", i+1, *format, *width, *height, len(data))

		// Upload image
		req := &network.ImageUploadRequest{
			SenderID: uint64(i + 1),
			Channel:  0,
			Format:   *format,
			Data:     data,
		}

		metadata, err := manager.UploadImage(req)
		if err != nil {
			logrus.Errorf("Upload %d failed: %v", i+1, err)
			continue
		}

		logrus.Infof("Upload %d: Success! ImageID=%s, Size=%d bytes", i+1, metadata.ImageID, metadata.Size)

		// Verify retrieval
		fullImage, err := manager.GetFullImage(metadata.ImageID)
		if err != nil {
			logrus.Errorf("Failed to retrieve full image: %v", err)
			continue
		}

		if len(fullImage) != metadata.Size {
			logrus.Errorf("Retrieved image size mismatch: expected %d, got %d", metadata.Size, len(fullImage))
		}

		thumb, err := manager.GetThumbnail(metadata.ImageID)
		if err != nil {
			logrus.Errorf("Failed to retrieve thumbnail: %v", err)
			continue
		}

		logrus.Infof("Upload %d: Retrieved full image (%d bytes) and thumbnail (%d bytes)", i+1, len(fullImage), len(thumb))
	}

	elapsed := time.Since(start)
	logrus.Infof("Uploaded %d images in %v (avg: %v per image)", *uploadCount, elapsed, elapsed/time.Duration(*uploadCount))
	logrus.Infof("Thumbnails relayed: %d", thumbnailCount)
	logrus.Infof("Images stored: %d", manager.GetImageCount())
}

// testThumbnail tests thumbnail generation
func testThumbnail() {
	logrus.Info("Testing thumbnail generation...")

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logrus.Fatalf("Failed to create output directory: %v", err)
	}

	// Create test image
	data, err := createTestImage(*width, *height, *format)
	if err != nil {
		logrus.Fatalf("Failed to create test image: %v", err)
	}

	logrus.Infof("Created %s image (%dx%d, %d bytes)", *format, *width, *height, len(data))

	// Validate and decode
	img, detectedFormat, err := network.ValidateImageData(data, *format)
	if err != nil {
		logrus.Fatalf("Image validation failed: %v", err)
	}

	logrus.Infof("Image validated successfully (format: %s)", detectedFormat)

	// Generate thumbnail
	start := time.Now()
	thumbnail, err := network.GenerateThumbnail(img)
	if err != nil {
		logrus.Fatalf("Thumbnail generation failed: %v", err)
	}
	elapsed := time.Since(start)

	logrus.Infof("Thumbnail generated in %v (%d bytes)", elapsed, len(thumbnail))

	// Save thumbnail to file
	thumbPath := filepath.Join(*outputDir, fmt.Sprintf("thumbnail_%dx%d.jpg", *width, *height))
	if err := os.WriteFile(thumbPath, thumbnail, 0644); err != nil {
		logrus.Errorf("Failed to save thumbnail: %v", err)
	} else {
		logrus.Infof("Thumbnail saved to: %s", thumbPath)
	}

	// Calculate compression ratio
	compressionRatio := float64(len(data)) / float64(len(thumbnail))
	logrus.Infof("Compression ratio: %.2fx (original: %d bytes, thumbnail: %d bytes)", compressionRatio, len(data), len(thumbnail))
}

// testChunkedTransfer tests chunked upload and download
func testChunkedTransfer() {
	logrus.Info("Testing chunked transfer...")

	manager := network.NewImageManager()

	// Create large test image
	largeWidth := 1024
	largeHeight := 1024
	data, err := createTestImage(largeWidth, largeHeight, "png")
	if err != nil {
		logrus.Fatalf("Failed to create test image: %v", err)
	}

	logrus.Infof("Created large test image (%dx%d, %d bytes)", largeWidth, largeHeight, len(data))

	// Calculate chunks
	totalChunks := (len(data) + network.MaxChunkSize - 1) / network.MaxChunkSize
	logrus.Infof("Image will be split into %d chunks (%d bytes each)", totalChunks, network.MaxChunkSize)

	// Start chunked upload
	senderID := uint64(1)
	channel := 0
	imageID, err := manager.StartChunkedUpload(senderID, channel, "png", len(data), totalChunks)
	if err != nil {
		logrus.Fatalf("Failed to start chunked upload: %v", err)
	}

	logrus.Infof("Chunked upload started: ImageID=%s", imageID)

	// Send chunks
	start := time.Now()
	for i := 0; i < totalChunks; i++ {
		startOffset := i * network.MaxChunkSize
		endOffset := startOffset + network.MaxChunkSize
		if endOffset > len(data) {
			endOffset = len(data)
		}
		chunkData := data[startOffset:endOffset]

		chunk := &network.ImageChunk{
			ImageID:     imageID,
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			Data:        chunkData,
		}

		complete, err := manager.ReceiveChunk(chunk)
		if err != nil {
			logrus.Fatalf("Failed to receive chunk %d: %v", i, err)
		}

		logrus.Debugf("Chunk %d/%d received (%d bytes)", i+1, totalChunks, len(chunkData))

		if complete {
			elapsed := time.Since(start)
			logrus.Infof("Chunked upload complete! Total time: %v", elapsed)
			logrus.Infof("Upload speed: %.2f MB/s", float64(len(data))/(1024*1024)/elapsed.Seconds())
		}
	}

	// Verify image count
	imageCount := manager.GetImageCount()
	logrus.Infof("Images in manager: %d", imageCount)
}

// testExpiry tests image expiry functionality
func testExpiry() {
	logrus.Info("Testing image expiry...")

	manager := network.NewImageManager()

	// Set expiry callback
	expiryCount := 0
	manager.SetExpiryCallback(func(imageID string) {
		expiryCount++
		logrus.Infof("Image expired: %s (total expired: %d)", imageID, expiryCount)
	})

	// Upload multiple images
	imageIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		data, err := createTestImage(100, 100, "png")
		if err != nil {
			logrus.Fatalf("Failed to create test image: %v", err)
		}

		req := &network.ImageUploadRequest{
			SenderID: uint64(i + 1),
			Channel:  0,
			Format:   "png",
			Data:     data,
		}

		metadata, err := manager.UploadImage(req)
		if err != nil {
			logrus.Fatalf("Upload failed: %v", err)
		}

		imageIDs[i] = metadata.ImageID
		logrus.Infof("Uploaded image %d: %s (expires at %s)", i+1, metadata.ImageID, metadata.ExpiryTime.Format(time.RFC3339))

		// Wait between uploads to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	logrus.Infof("All images uploaded. Images in manager: %d", manager.GetImageCount())

	// Simulate player disconnect to trigger expiry
	logrus.Info("Waiting 2 seconds before triggering player disconnect...")
	time.Sleep(2 * time.Second)

	logrus.Infof("Simulating player 1 disconnect (should expire their image)")
	manager.OnPlayerDisconnect(1)

	time.Sleep(100 * time.Millisecond)

	logrus.Infof("Images remaining in manager: %d", manager.GetImageCount())
	logrus.Infof("Total images expired: %d", expiryCount)

	// Verify remaining images are still accessible
	for i := 1; i < 3; i++ {
		_, err := manager.GetFullImage(imageIDs[i])
		if err != nil {
			logrus.Errorf("Image %d (%s) should still be accessible: %v", i+1, imageIDs[i], err)
		} else {
			logrus.Infof("Image %d (%s) is still accessible", i+1, imageIDs[i])
		}
	}
}

// testRateLimit tests upload rate limiting
func testRateLimit() {
	logrus.Info("Testing rate limiting...")

	manager := network.NewImageManager()
	senderID := uint64(1)

	// Create test image
	data, err := createTestImage(100, 100, "png")
	if err != nil {
		logrus.Fatalf("Failed to create test image: %v", err)
	}

	req := &network.ImageUploadRequest{
		SenderID: senderID,
		Channel:  0,
		Format:   "png",
		Data:     data,
	}

	// First upload should succeed
	logrus.Info("Attempt 1: First upload (should succeed)")
	metadata, err := manager.UploadImage(req)
	if err != nil {
		logrus.Errorf("First upload failed: %v", err)
	} else {
		logrus.Infof("First upload succeeded: %s", metadata.ImageID)
	}

	// Immediate second upload should fail
	logrus.Info("Attempt 2: Immediate second upload (should fail - rate limit)")
	_, err = manager.UploadImage(req)
	if err == network.ErrRateLimitExceeded {
		logrus.Info("Second upload correctly blocked by rate limit")
	} else {
		logrus.Errorf("Second upload should have failed with rate limit error, got: %v", err)
	}

	// Wait and try again
	logrus.Info("Waiting 2 seconds...")
	time.Sleep(2 * time.Second)

	logrus.Info("Attempt 3: After 2 seconds (should still fail - 60s limit)")
	_, err = manager.UploadImage(req)
	if err == network.ErrRateLimitExceeded {
		logrus.Info("Third upload correctly blocked by rate limit")
	} else {
		logrus.Errorf("Third upload should have failed with rate limit error, got: %v", err)
	}

	logrus.Info("Rate limit test complete (full 60s test would take too long)")
	logrus.Infof("Note: Rate limit is %v per player", network.ImageRateLimit)
}

// testModeration tests moderation hook functionality
func testModeration() {
	logrus.Info("Testing moderation hooks...")

	manager := network.NewImageManager()

	// Set moderation hook that rejects images over 50KB
	manager.SetModerationHook(func(metadata *network.ImageMetadata, data []byte) error {
		logrus.Infof("Moderation hook called for image %s (size: %d bytes)", metadata.ImageID, len(data))

		if len(data) > 50*1024 {
			logrus.Warn("Image rejected by moderation (too large)")
			return fmt.Errorf("image too large for moderation test")
		}

		logrus.Info("Image approved by moderation")
		return nil
	})

	// Test 1: Small image (should pass)
	logrus.Info("Test 1: Uploading small image (should pass moderation)")
	smallData, err := createTestImage(100, 100, "png")
	if err != nil {
		logrus.Fatalf("Failed to create small image: %v", err)
	}

	req := &network.ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   "png",
		Data:     smallData,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		logrus.Errorf("Small image upload failed: %v", err)
	} else {
		logrus.Infof("Small image upload succeeded: %s", metadata.ImageID)
	}

	// Test 2: Large image (should fail)
	logrus.Info("Test 2: Uploading large image (should fail moderation)")
	largeData, err := createTestImage(500, 500, "png")
	if err != nil {
		logrus.Fatalf("Failed to create large image: %v", err)
	}

	req.SenderID = 2 // Different sender to avoid rate limit
	req.Data = largeData

	_, err = manager.UploadImage(req)
	if err != nil {
		logrus.Infof("Large image correctly rejected by moderation: %v", err)
	} else {
		logrus.Error("Large image should have been rejected by moderation")
	}

	logrus.Infof("Moderation test complete. Images in manager: %d", manager.GetImageCount())
}
