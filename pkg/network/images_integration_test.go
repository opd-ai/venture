package network

import (
	"testing"
	"time"
)

// TestImageUpload_LatencySimulation tests image upload performance at different latency levels.
// Acceptance: 500KB image uploads in <5 seconds at 200ms latency
func TestImageUpload_LatencySimulation(t *testing.T) {
	tests := []struct {
		name           string
		latency        time.Duration
		imageSize      int // approximate image dimensions
		maxUploadTime  time.Duration
		expectSuccess  bool
	}{
		{
			name:          "200ms latency - should succeed within 5s",
			latency:       200 * time.Millisecond,
			imageSize:     400, // ~400x400 PNG ≈ 400-500KB
			maxUploadTime: 5 * time.Second,
			expectSuccess: true,
		},
		{
			name:          "2000ms latency - may take longer but should succeed",
			latency:       2000 * time.Millisecond,
			imageSize:     300, // smaller image for faster test
			maxUploadTime: 10 * time.Second,
			expectSuccess: true,
		},
		{
			name:          "5000ms latency - high latency scenario",
			latency:       5000 * time.Millisecond,
			imageSize:     200, // small image for manageable test time
			maxUploadTime: 15 * time.Second,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewImageManager()

			// Create test image
			data, err := createTestImage(tt.imageSize, tt.imageSize, ImageFormatPNG)
			if err != nil {
				t.Fatalf("Failed to create test image: %v", err)
			}

			t.Logf("Test image size: %d bytes", len(data))

			// Simulate chunked upload with latency
			totalChunks := (len(data) + MaxChunkSize - 1) / MaxChunkSize
			
			startTime := time.Now()

			imageID, err := manager.StartChunkedUpload(1, 0, ImageFormatPNG, len(data), totalChunks)
			if err != nil {
				t.Fatalf("Failed to start chunked upload: %v", err)
			}

			// Simulate latency for each chunk
			for i := 0; i < totalChunks; i++ {
				// Simulate network latency
				time.Sleep(tt.latency)

				startOffset := i * MaxChunkSize
				endOffset := min(startOffset+MaxChunkSize, len(data))
				chunkData := data[startOffset:endOffset]

				chunk := &ImageChunk{
					ImageID:     imageID,
					ChunkIndex:  i,
					TotalChunks: totalChunks,
					Data:        chunkData,
				}

				_, err := manager.ReceiveChunk(chunk)
				if err != nil {
					t.Fatalf("Failed to receive chunk %d: %v", i, err)
				}
			}

			uploadTime := time.Since(startTime)

			t.Logf("Upload completed in %v (max allowed: %v)", uploadTime, tt.maxUploadTime)

			if tt.expectSuccess {
				if uploadTime > tt.maxUploadTime {
					t.Errorf("Upload took %v, exceeds max %v", uploadTime, tt.maxUploadTime)
				}
			}
		})
	}
}

// TestImageValidation_SizeLimit tests that images exceeding 500KB are rejected.
// Acceptance: >500KB images rejected with error message
func TestImageValidation_SizeLimit(t *testing.T) {
	manager := NewImageManager()

	// Create large data exceeding 500KB
	largeData := make([]byte, MaxImageSize+1000)
	for i := range largeData {
		largeData[i] = 0xFF
	}

	req := &ImageUploadRequest{
		SenderID: 1,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     largeData,
	}

	_, err := manager.UploadImage(req)
	if err != ErrImageTooLarge {
		t.Errorf("Expected ErrImageTooLarge for %d byte image, got %v", len(largeData), err)
	}

	t.Logf("Correctly rejected image of %d bytes (limit: %d)", len(largeData), MaxImageSize)
}

// TestImageValidation_InvalidTypes tests that non-PNG/JPEG/GIF images are rejected.
// Acceptance: .bmp/.tiff rejected with error message
func TestImageValidation_InvalidTypes(t *testing.T) {
	manager := NewImageManager()

	invalidTypes := []string{"bmp", "tiff", "webp", "svg", "ico"}

	for _, invalidType := range invalidTypes {
		t.Run(invalidType, func(t *testing.T) {
			// Create dummy data (doesn't matter, type check comes first)
			dummyData := []byte{0x00, 0x01, 0x02, 0x03}

			req := &ImageUploadRequest{
				SenderID: 1,
				Channel:  0,
				Format:   invalidType,
				Data:     dummyData,
			}

			_, err := manager.UploadImage(req)
			if err == nil {
				t.Errorf("Expected error for invalid type %s, got nil", invalidType)
			} else {
				t.Logf("Correctly rejected type %s: %v", invalidType, err)
			}
		})
	}
}

// TestImageModeration_HookInvocation tests that moderation hook is called for all uploads.
// Acceptance: OnImageUpload hook invoked for all uploads
func TestImageModeration_HookInvocation(t *testing.T) {
	manager := NewImageManager()

	hookCallCount := 0
	var lastMetadata *ImageMetadata
	var lastData []byte

	// Set moderation hook
	manager.SetModerationHook(func(metadata *ImageMetadata, data []byte) error {
		hookCallCount++
		lastMetadata = metadata
		lastData = data
		return nil
	})

	// Upload multiple images
	for i := 0; i < 3; i++ {
		// Reset rate limit
		manager.mu.Lock()
		manager.playerLastUpload[uint64(i)] = time.Now().Add(-ImageRateLimit - time.Second)
		manager.mu.Unlock()

		data, err := createTestImage(100, 100, ImageFormatPNG)
		if err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}

		req := &ImageUploadRequest{
			SenderID: uint64(i),
			Channel:  0,
			Format:   ImageFormatPNG,
			Data:     data,
		}

		_, err = manager.UploadImage(req)
		if err != nil {
			t.Fatalf("Upload %d failed: %v", i, err)
		}
	}

	if hookCallCount != 3 {
		t.Errorf("Expected hook to be called 3 times, got %d", hookCallCount)
	}

	if lastMetadata == nil {
		t.Errorf("Hook should have received metadata")
	}

	if lastData == nil {
		t.Errorf("Hook should have received image data")
	}

	t.Logf("Moderation hook correctly called %d times", hookCallCount)
}

// TestImageExpiry_TimeAndDisconnect tests that images expire after 10 minutes or disconnect.
// Acceptance: Images deleted after 10 minutes or sender disconnect
func TestImageExpiry_TimeAndDisconnect(t *testing.T) {
	t.Run("disconnect expiry", func(t *testing.T) {
		manager := NewImageManager()

		data, err := createTestImage(100, 100, ImageFormatPNG)
		if err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}

		senderID := uint64(1)
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

		// Verify image exists
		_, err = manager.GetFullImage(imageID)
		if err != nil {
			t.Errorf("Image should exist before disconnect: %v", err)
		}

		// Trigger disconnect
		manager.OnPlayerDisconnect(senderID)

		// Verify image is gone
		_, err = manager.GetFullImage(imageID)
		if err != ErrImageNotFound {
			t.Errorf("Image should be removed after disconnect, got error: %v", err)
		}

		t.Logf("Image correctly expired on sender disconnect")
	})

	t.Run("time expiry", func(t *testing.T) {
		manager := NewImageManager()

		expiredImageID := ""
		manager.SetExpiryCallback(func(imageID string) {
			expiredImageID = imageID
		})

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

		// Manually expire the image (simulates timer firing)
		manager.expireImage(imageID)

		// Verify image is gone
		_, err = manager.GetFullImage(imageID)
		if err != ErrImageNotFound {
			t.Errorf("Image should be removed after expiry, got error: %v", err)
		}

		if expiredImageID != imageID {
			t.Errorf("Expected expiry callback for %s, got %s", imageID, expiredImageID)
		}

		t.Logf("Image correctly expired after timeout")
	})
}

// TestImageRateLimit_TwoWithin60s tests rate limiting.
// Acceptance: Uploading 2 images within 60s triggers rejection
func TestImageRateLimit_TwoWithin60s(t *testing.T) {
	manager := NewImageManager()
	senderID := uint64(1)

	data, err := createTestImage(100, 100, ImageFormatPNG)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	req := &ImageUploadRequest{
		SenderID: senderID,
		Channel:  0,
		Format:   ImageFormatPNG,
		Data:     data,
	}

	// First upload
	_, err = manager.UploadImage(req)
	if err != nil {
		t.Fatalf("First upload should succeed: %v", err)
	}

	t.Logf("First upload succeeded")

	// Second upload within 60s (should fail)
	_, err = manager.UploadImage(req)
	if err != ErrRateLimitExceeded {
		t.Errorf("Second upload within 60s should be rate limited, got: %v", err)
	}

	t.Logf("Second upload correctly rate limited: %v", err)

	// Third upload after rate limit period (should succeed)
	manager.mu.Lock()
	manager.playerLastUpload[senderID] = time.Now().Add(-ImageRateLimit - time.Second)
	manager.mu.Unlock()

	_, err = manager.UploadImage(req)
	if err != nil {
		t.Errorf("Upload after rate limit period should succeed: %v", err)
	}

	t.Logf("Third upload after rate limit period succeeded")
}

// TestManualAccept_ThumbnailAutoDownload tests manual accept workflow.
// Acceptance: Full image not downloaded until user clicks thumbnail
// Note: This is a workflow test showing the pattern. Actual UI integration happens in client code.
func TestManualAccept_ThumbnailAutoDownload(t *testing.T) {
	manager := NewImageManager()

	// Track thumbnail relays
	thumbnailRelayed := false
	var relayedThumbnail []byte

	manager.SetThumbnailRelayCallback(func(metadata *ImageMetadata, thumbnail []byte) {
		thumbnailRelayed = true
		relayedThumbnail = thumbnail
	})

	// Upload image
	data, err := createTestImage(200, 200, ImageFormatPNG)
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

	// Verify thumbnail was relayed automatically
	if !thumbnailRelayed {
		t.Errorf("Thumbnail should be relayed automatically on upload")
	}

	if len(relayedThumbnail) == 0 {
		t.Errorf("Relayed thumbnail should not be empty")
	}

	t.Logf("Thumbnail auto-downloaded: %d bytes", len(relayedThumbnail))

	// Simulate user clicking thumbnail to request full image
	// In real implementation, this would be a UI event triggering download request
	requesterID := uint64(2) // Different player

	totalChunks, err := manager.StartChunkedDownload(requesterID, metadata.ImageID)
	if err != nil {
		t.Fatalf("Failed to start full image download: %v", err)
	}

	t.Logf("Full image download initiated on user request: %d chunks", totalChunks)

	// Download first chunk as proof of manual acceptance
	chunk, err := manager.GetNextChunk(requesterID, metadata.ImageID)
	if err != nil {
		t.Fatalf("Failed to get first chunk: %v", err)
	}

	if len(chunk.Data) == 0 {
		t.Errorf("Chunk data should not be empty")
	}

	t.Logf("Manual accept workflow: thumbnail auto-downloaded, full image downloaded on request")
}

// Benchmarks

func BenchmarkImageUpload_WithLatency(b *testing.B) {
	manager := NewImageManager()
	data, err := createTestImage(300, 300, ImageFormatPNG)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	totalChunks := (len(data) + MaxChunkSize - 1) / MaxChunkSize
	latency := 200 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imageID, _ := manager.StartChunkedUpload(uint64(i), 0, ImageFormatPNG, len(data), totalChunks)

		for j := 0; j < totalChunks; j++ {
			time.Sleep(latency) // Simulate network latency

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
