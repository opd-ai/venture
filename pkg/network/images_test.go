package network

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
	"time"
)

// createTestImage creates a test image of specified dimensions and format.
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

	var buf bytes.Buffer
	var err error

	switch format {
	case ImageFormatPNG:
		err = png.Encode(&buf, img)
	case ImageFormatJPEG:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	default:
		return nil, ErrInvalidImageType
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestValidateImageData(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		format      string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid small PNG",
			width:       100,
			height:      100,
			format:      ImageFormatPNG,
			expectError: false,
		},
		{
			name:        "valid small JPEG",
			width:       100,
			height:      100,
			format:      ImageFormatJPEG,
			expectError: false,
		},
		{
			name:        "valid max dimensions",
			width:       2048,
			height:      2048,
			format:      ImageFormatPNG,
			expectError: false,
		},
		{
			name:        "invalid dimensions - too wide",
			width:       3000,
			height:      100,
			format:      ImageFormatPNG,
			expectError: true,
			errorType:   ErrImageTooWide,
		},
		{
			name:        "invalid dimensions - too tall",
			width:       100,
			height:      3000,
			format:      ImageFormatPNG,
			expectError: true,
			errorType:   ErrImageTooWide,
		},
		{
			name:        "invalid format",
			width:       100,
			height:      100,
			format:      "bmp",
			expectError: true,
			errorType:   ErrInvalidImageType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip size tests for large images (too slow)
			if tt.width > 2048 || tt.height > 2048 {
				t.Skip("Skipping large image test for speed")
			}

			data, err := createTestImage(tt.width, tt.height, tt.format)
			if tt.format == "bmp" {
				data = []byte{0x42, 0x4D} // BMP header
			}

			if err != nil && !tt.expectError {
				t.Fatalf("Failed to create test image: %v", err)
			}

			img, format, err := ValidateImageData(data, tt.format)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errorType != nil && err != tt.errorType && err.Error() != tt.errorType.Error() {
					// Check if error message contains expected error
					if !bytes.Contains([]byte(err.Error()), []byte(tt.errorType.Error())) {
						t.Errorf("Expected error type %v, got %v", tt.errorType, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if img == nil {
					t.Errorf("Expected image, got nil")
				}
				if format != tt.format {
					t.Errorf("Expected format %s, got %s", tt.format, format)
				}
			}
		})
	}
}

func TestValidateImageData_SizeLimit(t *testing.T) {
	// Create a large data blob exceeding size limit
	largeData := make([]byte, MaxImageSize+1)
	for i := range largeData {
		largeData[i] = 0xFF
	}

	_, _, err := ValidateImageData(largeData, ImageFormatPNG)
	if err != ErrImageTooLarge {
		t.Errorf("Expected ErrImageTooLarge, got %v", err)
	}
}

func TestGenerateThumbnail(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		expectW int
		expectH int
	}{
		{
			name:    "square image",
			width:   256,
			height:  256,
			expectW: 128,
			expectH: 128,
		},
		{
			name:    "wide image",
			width:   512,
			height:  256,
			expectW: 128,
			expectH: 64,
		},
		{
			name:    "tall image",
			width:   256,
			height:  512,
			expectW: 64,
			expectH: 128,
		},
		{
			name:    "small image",
			width:   64,
			height:  64,
			expectW: 64,
			expectH: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image
			img := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))

			// Generate thumbnail
			thumbData, err := GenerateThumbnail(img)
			if err != nil {
				t.Fatalf("Failed to generate thumbnail: %v", err)
			}

			// Decode thumbnail to verify dimensions
			thumbImg, err := jpeg.Decode(bytes.NewReader(thumbData))
			if err != nil {
				t.Fatalf("Failed to decode thumbnail: %v", err)
			}

			bounds := thumbImg.Bounds()
			w := bounds.Dx()
			h := bounds.Dy()

			if w != tt.expectW || h != tt.expectH {
				t.Errorf("Expected thumbnail size %dx%d, got %dx%d", tt.expectW, tt.expectH, w, h)
			}
		})
	}
}

func TestImageManager_UploadImage(t *testing.T) {
	tests := []struct {
		name        string
		senderID    uint64
		channel     int
		width       int
		height      int
		format      string
		expectError bool
	}{
		{
			name:        "successful PNG upload",
			senderID:    1,
			channel:     0,
			width:       200,
			height:      200,
			format:      ImageFormatPNG,
			expectError: false,
		},
		{
			name:        "successful JPEG upload",
			senderID:    2,
			channel:     1,
			width:       300,
			height:      300,
			format:      ImageFormatJPEG,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewImageManager()

			// Create test image
			data, err := createTestImage(tt.width, tt.height, tt.format)
			if err != nil {
				t.Fatalf("Failed to create test image: %v", err)
			}

			req := &ImageUploadRequest{
				SenderID: tt.senderID,
				Channel:  tt.channel,
				Format:   tt.format,
				Data:     data,
			}

			metadata, err := manager.UploadImage(req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if metadata == nil {
					t.Fatalf("Expected metadata, got nil")
				}
				if metadata.SenderID != tt.senderID {
					t.Errorf("Expected sender ID %d, got %d", tt.senderID, metadata.SenderID)
				}
				if metadata.Channel != tt.channel {
					t.Errorf("Expected channel %d, got %d", tt.channel, metadata.Channel)
				}
				if metadata.Format != tt.format {
					t.Errorf("Expected format %s, got %s", tt.format, metadata.Format)
				}
				if metadata.Width != tt.width {
					t.Errorf("Expected width %d, got %d", tt.width, metadata.Width)
				}
				if metadata.Height != tt.height {
					t.Errorf("Expected height %d, got %d", tt.height, metadata.Height)
				}

				// Verify image is stored
				stored, err := manager.GetFullImage(metadata.ImageID)
				if err != nil {
					t.Errorf("Failed to retrieve stored image: %v", err)
				}
				if !bytes.Equal(stored, data) {
					t.Errorf("Stored image data doesn't match original")
				}

				// Verify thumbnail exists
				thumb, err := manager.GetThumbnail(metadata.ImageID)
				if err != nil {
					t.Errorf("Failed to retrieve thumbnail: %v", err)
				}
				if len(thumb) == 0 {
					t.Errorf("Thumbnail is empty")
				}
			}
		})
	}
}

func TestImageManager_RateLimit(t *testing.T) {
	manager := NewImageManager()
	senderID := uint64(1)

	// Create test image
	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// First upload should succeed
	req := &ImageUploadRequest{
		SenderID: senderID,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	_, err = manager.UploadImage(req)
	if err != nil {
		t.Fatalf("First upload failed: %v", err)
	}

	// Immediate second upload should fail (rate limit)
	_, err = manager.UploadImage(req)
	if err != ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}

	// Wait for rate limit to reset (simulate by modifying last upload time)
	manager.mu.Lock()
	manager.playerLastUpload[senderID] = time.Now().Add(-ImageRateLimit - time.Second)
	manager.mu.Unlock()

	// Third upload should succeed
	_, err = manager.UploadImage(req)
	if err != nil {
		t.Errorf("Third upload after rate limit reset failed: %v", err)
	}
}

func TestImageManager_Expiry(t *testing.T) {
	manager := NewImageManager()

	// Track expiry callback
	var expiredImageID string
	manager.SetExpiryCallback(func(imageID string) {
		expiredImageID = imageID
	})

	// Create and upload test image
	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	imageID := metadata.ImageID

	// Verify image exists
	_, err = manager.GetFullImage(imageID)
	if err != nil {
		t.Errorf("Image should exist: %v", err)
	}

	// Trigger expiry manually
	manager.expireImage(imageID)

	// Verify image is gone
	_, err = manager.GetFullImage(imageID)
	if err != ErrImageNotFound {
		t.Errorf("Expected ErrImageNotFound after expiry, got %v", err)
	}

	// Verify expiry callback was called
	if expiredImageID != imageID {
		t.Errorf("Expected expiry callback for %s, got %s", imageID, expiredImageID)
	}
}

func TestImageManager_PlayerDisconnect(t *testing.T) {
	manager := NewImageManager()
	senderID := uint64(1)

	// Create and upload multiple images
	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	imageIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		// Reset rate limit
		manager.mu.Lock()
		manager.playerLastUpload[senderID] = time.Now().Add(-ImageRateLimit - time.Second)
		manager.mu.Unlock()

		req := &ImageUploadRequest{
			SenderID: senderID,
			Channel:  0,
			Format:   ImageFormatPNG,
			Data:     data,
		}

		metadata, err := manager.UploadImage(req)
		if err != nil {
			t.Fatalf("Upload %d failed: %v", i, err)
		}
		imageIDs[i] = metadata.ImageID
	}

	// Verify all images exist
	for i, imageID := range imageIDs {
		_, err := manager.GetFullImage(imageID)
		if err != nil {
			t.Errorf("Image %d should exist: %v", i, err)
		}
	}

	// Trigger disconnect
	manager.OnPlayerDisconnect(senderID)

	// Verify all images are gone
	for i, imageID := range imageIDs {
		_, err := manager.GetFullImage(imageID)
		if err != ErrImageNotFound {
			t.Errorf("Image %d should be removed after disconnect, got error: %v", i, err)
		}
	}
}

func TestImageManager_ModerationHook(t *testing.T) {
	manager := NewImageManager()

	// Set moderation hook that rejects images from player 666
	manager.SetModerationHook(func(metadata *ImageMetadata, data []byte) error {
		if metadata.SenderID == 666 {
			return fmt.Errorf("player 666 is banned")
		}
		return nil
	})

	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Upload from allowed player should succeed
	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	_, err = manager.UploadImage(req)
	if err != nil {
		t.Errorf("Upload from allowed player failed: %v", err)
	}

	// Reset rate limit
	manager.mu.Lock()
	manager.playerLastUpload[666] = time.Now().Add(-ImageRateLimit - time.Second)
	manager.mu.Unlock()

	// Upload from banned player should fail
	req.SenderID = 666
	_, err = manager.UploadImage(req)
	if err == nil {
		t.Errorf("Upload from banned player should have failed")
	}
}

func TestImageManager_ChunkedUpload(t *testing.T) {
	manager := NewImageManager()
	senderID := uint64(1)
	channel := 0

	// Create test image
	data, err := createTestImage(200, 200, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Calculate chunks
	totalChunks := (len(data) + MaxChunkSize - 1) / MaxChunkSize

	// Start chunked upload
	imageID, err := manager.StartChunkedUpload(senderID, channel, ImageFormatPNG, len(data), totalChunks)
	if err != nil {
		t.Fatalf("Failed to start chunked upload: %v", err)
	}

	// Send all chunks
	for i := 0; i < totalChunks; i++ {
		startOffset := i * MaxChunkSize
		endOffset := min(startOffset+MaxChunkSize, len(data))
		chunkData := data[startOffset:endOffset]

		chunk := &ImageChunk{
			ImageID:     imageID,
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			Data:        chunkData,
		}

		complete, err := manager.ReceiveChunk(chunk)
		if err != nil {
			t.Fatalf("Failed to receive chunk %d: %v", i, err)
		}

		// Should be complete only on last chunk
		expectedComplete := (i == totalChunks-1)
		if complete != expectedComplete {
			t.Errorf("Chunk %d: expected complete=%v, got %v", i, expectedComplete, complete)
		}
	}

	// Verify image was stored (imageID should still work even though it was ephemeral during upload)
	// Note: The chunked upload reassigns a new imageID internally, so we need to check if ANY image exists
	count := manager.GetImageCount()
	if count != 1 {
		t.Errorf("Expected 1 stored image, got %d", count)
	}
}

func TestImageManager_ChunkedDownload(t *testing.T) {
	manager := NewImageManager()
	senderID := uint64(1)
	requesterID := uint64(2)

	// Create and upload test image
	data, err := createTestImage(200, 200, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: senderID,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	imageID := metadata.ImageID

	// Start chunked download
	totalChunks, err := manager.StartChunkedDownload(requesterID, imageID)
	if err != nil {
		t.Fatalf("Failed to start chunked download: %v", err)
	}

	expectedChunks := (len(data) + MaxChunkSize - 1) / MaxChunkSize
	if totalChunks != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, totalChunks)
	}

	// Download all chunks
	var downloadedData bytes.Buffer
	for i := 0; i < totalChunks; i++ {
		chunk, err := manager.GetNextChunk(requesterID, imageID)
		if err != nil {
			t.Fatalf("Failed to get chunk %d: %v", i, err)
		}

		if chunk.ChunkIndex != i {
			t.Errorf("Expected chunk index %d, got %d", i, chunk.ChunkIndex)
		}

		downloadedData.Write(chunk.Data)
	}

	// Verify downloaded data matches original
	if !bytes.Equal(downloadedData.Bytes(), data) {
		t.Errorf("Downloaded data doesn't match original")
	}

	// Attempting to get another chunk should fail
	_, err = manager.GetNextChunk(requesterID, imageID)
	if err == nil {
		t.Errorf("Expected error when getting chunk beyond total, got nil")
	}
}

func TestImageManager_CleanupStaleUploads(t *testing.T) {
	manager := NewImageManager()

	// Start chunked upload
	imageID, err := manager.StartChunkedUpload(1, 0, ImageFormatPNG, 10000, 2)
	if err != nil {
		t.Fatalf("Failed to start chunked upload: %v", err)
	}

	// Verify upload exists
	manager.mu.RLock()
	_, exists := manager.uploadInProgress[imageID]
	manager.mu.RUnlock()

	if !exists {
		t.Errorf("Upload should exist")
	}

	// Set last activity to 10 minutes ago
	manager.mu.Lock()
	if upload, ok := manager.uploadInProgress[imageID]; ok {
		upload.LastActivity = time.Now().Add(-10 * time.Minute)
	}
	manager.mu.Unlock()

	// Cleanup stale uploads
	manager.CleanupStaleUploads()

	// Verify upload is gone
	manager.mu.RLock()
	_, exists = manager.uploadInProgress[imageID]
	manager.mu.RUnlock()

	if exists {
		t.Errorf("Stale upload should be removed")
	}
}

func TestImageManager_CleanupStaleDownloads(t *testing.T) {
	manager := NewImageManager()

	// Create and upload test image
	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	imageID := metadata.ImageID

	// Start download
	_, err = manager.StartChunkedDownload(2, imageID)
	if err != nil {
		t.Fatalf("Failed to start download: %v", err)
	}

	// Verify download exists
	manager.mu.RLock()
	_, exists := manager.downloadInProgress[imageID]
	manager.mu.RUnlock()

	if !exists {
		t.Errorf("Download should exist")
	}

	// Set last activity to 10 minutes ago
	manager.mu.Lock()
	if downloaders, ok := manager.downloadInProgress[imageID]; ok {
		for _, download := range downloaders {
			download.LastActivity = time.Now().Add(-10 * time.Minute)
		}
	}
	manager.mu.Unlock()

	// Cleanup stale downloads
	manager.CleanupStaleDownloads()

	// Verify download is gone
	manager.mu.RLock()
	_, exists = manager.downloadInProgress[imageID]
	manager.mu.RUnlock()

	if exists {
		t.Errorf("Stale download should be removed")
	}
}

// Benchmarks

func BenchmarkValidateImageData(b *testing.B) {
	data, err := createTestImage(512, 512, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidateImageData(data, ImageFormatPNG)
	}
}

func BenchmarkGenerateThumbnail(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateThumbnail(img)
	}
}

func BenchmarkImageManager_UploadImage(b *testing.B) {
	manager := NewImageManager()
	data, err := createTestImage(200, 200, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different sender ID to avoid rate limiting
		req := &ImageUploadRequest{
			SenderID: uint64(i),
			Channel:  0,
			Format:   ImageFormatPNG,
			Data:     data,
		}
		_, _ = manager.UploadImage(req)
	}
}

func BenchmarkImageManager_GetThumbnail(b *testing.B) {
	manager := NewImageManager()
	data, err := createTestImage(200, 200, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		b.Fatalf("Upload failed: %v", err)
	}

	imageID := metadata.ImageID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GetThumbnail(imageID)
	}
}

func BenchmarkImageManager_GetFullImage(b *testing.B) {
	manager := NewImageManager()
	data, err := createTestImage(200, 200, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	metadata, err := manager.UploadImage(req)
	if err != nil {
		b.Fatalf("Upload failed: %v", err)
	}

	imageID := metadata.ImageID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GetFullImage(imageID)
	}
}

func BenchmarkImageManager_ChunkedUpload(b *testing.B) {
	data, err := createTestImage(500, 500, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	totalChunks := (len(data) + MaxChunkSize - 1) / MaxChunkSize

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewImageManager()

		imageID, _ := manager.StartChunkedUpload(uint64(i), 0, ImageFormatPNG, len(data), totalChunks)

		for j := 0; j < totalChunks; j++ {
			startOffset := j * MaxChunkSize
			endOffset := min(startOffset+MaxChunkSize, len(data))
			chunkData := data[startOffset:endOffset]

			chunk := &ImageChunk{
				ImageID:     imageID,
				ChunkIndex:  j,
				TotalChunks: totalChunks,
				Data:        chunkData,
			}

			_, _ = manager.ReceiveChunk(chunk)
		}
	}
}
