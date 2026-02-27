package network

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/image/draw"
)

// Image format constants
const (
	ImageFormatPNG  = "png"
	ImageFormatJPEG = "jpeg"
	ImageFormatGIF  = "gif"
)

// Image sharing constraints as per Phase 23 spec
const (
	MaxImageSize      = 500 * 1024 // 500KB
	MaxImageDimension = 2048       // 2048x2048 pixels
	ThumbnailSize     = 128        // 128x128 pixels
	ThumbnailQuality  = 75         // JPEG quality
	ImageExpiryTime   = 10 * time.Minute
	ImageRateLimit    = 60 * time.Second // 1 image per 60 seconds
	MaxChunkSize      = 64 * 1024        // 64KB chunks for upload/download
)

var (
	// ErrImageTooLarge is returned when image exceeds size limit
	ErrImageTooLarge = errors.New("image exceeds 500KB size limit")
	// ErrInvalidImageType is returned when image type is not PNG/JPEG/GIF
	ErrInvalidImageType = errors.New("image type must be PNG, JPEG, or GIF")
	// ErrImageTooWide is returned when image dimensions exceed limit
	ErrImageTooWide = errors.New("image dimensions exceed 2048x2048 pixels")
	// ErrRateLimitExceeded is returned when upload rate limit is violated
	ErrRateLimitExceeded = errors.New("image upload rate limit exceeded (1 per 60 seconds)")
	// ErrImageNotFound is returned when requested image doesn't exist
	ErrImageNotFound = errors.New("image not found")
	// ErrImageExpired is returned when image has expired
	ErrImageExpired = errors.New("image has expired")
	// ErrInvalidChunkSequence is returned when chunk sequence is invalid
	ErrInvalidChunkSequence = errors.New("invalid chunk sequence")
)

// generateImageID generates a unique image ID (UUID v4 format).
func generateImageID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		// Fallback to timestamp-based ID
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"error":       err.Error(),
		}).Warn("crypto rand failed, using timestamp fallback for image ID")
		return fmt.Sprintf("img-%d", time.Now().UnixNano())
	}

	// Set version (4) and variant bits per RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return "img-" + hex.EncodeToString(uuid)
}

// ImageMetadata contains metadata about a shared image.
type ImageMetadata struct {
	ImageID    string    // Unique image identifier
	SenderID   uint64    // Player who uploaded the image
	Channel    int       // Chat channel for distribution
	Format     string    // Image format (png, jpeg, gif)
	Width      int       // Image width in pixels
	Height     int       // Image height in pixels
	Size       int       // Image size in bytes
	UploadTime time.Time // When image was uploaded
	ExpiryTime time.Time // When image expires
}

// ImageData represents a complete image with metadata and data.
type ImageData struct {
	Metadata  ImageMetadata
	FullData  []byte // Full image data
	Thumbnail []byte // 128x128 JPEG thumbnail
}

// ImageChunk represents a chunk of image data for chunked transfer.
type ImageChunk struct {
	ImageID      string // Image being transferred
	ChunkIndex   int    // Chunk sequence number (0-based)
	TotalChunks  int    // Total number of chunks
	Data         []byte // Chunk data
	IsResume     bool   // Whether this is a resumed transfer
	LastChunkIdx int    // Last successfully received chunk (for resume)
}

// ImageUploadRequest represents a request to upload an image.
type ImageUploadRequest struct {
	SenderID uint64 // Player uploading
	Channel  int    // Distribution channel
	Format   string // Image format
	Data     []byte // Image data (must be ≤500KB)
}

// ImageDownloadRequest represents a request to download a full image.
type ImageDownloadRequest struct {
	RequesterID uint64 // Player requesting download
	ImageID     string // Image to download
}

// ModerationHook is a callback function for image moderation.
// Returns error if image should be rejected.
type ModerationHook func(metadata *ImageMetadata, data []byte) error

// ImageManager handles image upload, storage, and distribution.
type ImageManager struct {
	mu                 sync.RWMutex
	images             map[string]*ImageData                  // ImageID → image data
	playerLastUpload   map[uint64]time.Time                   // PlayerID → last upload time
	expiryTimers       map[string]*time.Timer                 // ImageID → expiry timer
	playerDisconnect   map[uint64]map[string]struct{}         // PlayerID → set of image IDs
	uploadInProgress   map[string]*ChunkedUpload              // ImageID → upload state
	downloadInProgress map[string]map[uint64]*ChunkedDownload // ImageID → RequesterID → download state
	moderationHook     ModerationHook
	onThumbnailRelay   func(metadata *ImageMetadata, thumbnail []byte) // Callback to relay thumbnails
	onImageExpiry      func(imageID string)                            // Callback for image expiry
}

// ChunkedUpload tracks the state of a chunked image upload.
type ChunkedUpload struct {
	Metadata     ImageMetadata
	Chunks       [][]byte  // Received chunks
	ReceivedMask []bool    // Which chunks have been received
	TotalChunks  int       // Expected number of chunks
	StartTime    time.Time // When upload started
	LastActivity time.Time // Last chunk received
}

// ChunkedDownload tracks the state of a chunked image download.
type ChunkedDownload struct {
	ImageID      string
	SentChunks   int       // Number of chunks sent
	TotalChunks  int       // Total chunks to send
	LastActivity time.Time // Last chunk sent
}

// NewImageManager creates a new image manager.
func NewImageManager() *ImageManager {
	return &ImageManager{
		images:             make(map[string]*ImageData),
		playerLastUpload:   make(map[uint64]time.Time),
		expiryTimers:       make(map[string]*time.Timer),
		playerDisconnect:   make(map[uint64]map[string]struct{}),
		uploadInProgress:   make(map[string]*ChunkedUpload),
		downloadInProgress: make(map[string]map[uint64]*ChunkedDownload),
	}
}

// SetModerationHook sets a callback for image moderation.
// The hook is called before an image is accepted and stored.
func (im *ImageManager) SetModerationHook(hook ModerationHook) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.moderationHook = hook
}

// SetThumbnailRelayCallback sets a callback for relaying thumbnails to recipients.
func (im *ImageManager) SetThumbnailRelayCallback(callback func(*ImageMetadata, []byte)) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.onThumbnailRelay = callback
}

// SetExpiryCallback sets a callback for image expiry notifications.
func (im *ImageManager) SetExpiryCallback(callback func(string)) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.onImageExpiry = callback
}

// ValidateImageData validates image data against size, type, and dimension constraints.
// Returns the decoded image, format, and any validation error.
func ValidateImageData(data []byte, format string) (image.Image, string, error) {
	// Check size limit
	if len(data) > MaxImageSize {
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"size_bytes":  len(data),
			"max_bytes":   MaxImageSize,
			"format":      format,
		}).Warn("image exceeds size limit")
		return nil, "", ErrImageTooLarge
	}

	// Validate format and decode image
	var img image.Image
	var err error
	var detectedFormat string

	reader := bytes.NewReader(data)

	switch format {
	case ImageFormatPNG:
		img, err = png.Decode(reader)
		detectedFormat = ImageFormatPNG
	case ImageFormatJPEG:
		img, err = jpeg.Decode(reader)
		detectedFormat = ImageFormatJPEG
	case ImageFormatGIF:
		img, err = gif.Decode(reader)
		detectedFormat = ImageFormatGIF
	default:
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"format":      format,
		}).Warn("invalid image format")
		return nil, "", ErrInvalidImageType
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"format":      format,
			"size_bytes":  len(data),
			"error":       err.Error(),
		}).Error("failed to decode image")
		return nil, "", fmt.Errorf("failed to decode %s image: %w", format, err)
	}

	// Check dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > MaxImageDimension || height > MaxImageDimension {
		logrus.WithFields(logrus.Fields{
			"system_name":   "image_manager",
			"width":         width,
			"height":        height,
			"max_dimension": MaxImageDimension,
			"format":        format,
		}).Warn("image dimensions exceed limit")
		return nil, "", ErrImageTooWide
	}

	return img, detectedFormat, nil
}

// GenerateThumbnail creates a 128x128 JPEG thumbnail from an image.
// Images smaller than 128x128 are not upscaled.
func GenerateThumbnail(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling to fit within 128x128 while preserving aspect ratio
	// Don't upscale images that are already smaller
	maxDimension := max(width, height)
	var newWidth, newHeight int

	if maxDimension <= ThumbnailSize {
		// Image is already small enough, don't upscale
		newWidth = width
		newHeight = height
	} else {
		// Scale down to fit within ThumbnailSize
		scale := float64(ThumbnailSize) / float64(maxDimension)
		newWidth = int(float64(width) * scale)
		newHeight = int(float64(height) * scale)
	}

	// Create thumbnail image
	thumbnail := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use high-quality scaling (CatmullRom interpolation)
	draw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), img, bounds, draw.Over, nil)

	// Encode as JPEG with quality 75
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: ThumbnailQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CheckRateLimit checks if a player can upload an image (rate limiting).
// Returns true if upload is allowed, false if rate limit exceeded.
func (im *ImageManager) CheckRateLimit(playerID uint64) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	lastUpload, exists := im.playerLastUpload[playerID]
	if !exists {
		return true // First upload, allowed
	}

	return time.Since(lastUpload) >= ImageRateLimit
}

// UploadImage processes an image upload request.
// Validates the image, generates thumbnail, stores it, and relays thumbnail to recipients.
// Returns ImageMetadata and error.
func (im *ImageManager) UploadImage(req *ImageUploadRequest) (*ImageMetadata, error) {
	// Check rate limit
	if !im.CheckRateLimit(req.SenderID) {
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"playerID":    req.SenderID,
			"channel":     req.Channel,
		}).Warn("image upload rate limit exceeded")
		return nil, ErrRateLimitExceeded
	}

	// Validate image data
	img, format, err := ValidateImageData(req.Data, req.Format)
	if err != nil {
		return nil, err
	}

	// Generate thumbnail
	thumbnail, err := GenerateThumbnail(img)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "image_manager",
			"playerID":    req.SenderID,
			"format":      format,
			"error":       err.Error(),
		}).Error("failed to generate thumbnail")
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Create metadata
	imageID := generateImageID()
	bounds := img.Bounds()
	metadata := ImageMetadata{
		ImageID:    imageID,
		SenderID:   req.SenderID,
		Channel:    req.Channel,
		Format:     format,
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		Size:       len(req.Data),
		UploadTime: time.Now(),
		ExpiryTime: time.Now().Add(ImageExpiryTime),
	}

	// Run moderation hook if set
	im.mu.RLock()
	hook := im.moderationHook
	im.mu.RUnlock()

	if hook != nil {
		if err := hook(&metadata, req.Data); err != nil {
			logrus.WithFields(logrus.Fields{
				"system_name": "image_manager",
				"playerID":    req.SenderID,
				"imageID":     imageID,
				"error":       err.Error(),
			}).Warn("image rejected by moderation")
			return nil, fmt.Errorf("image rejected by moderation: %w", err)
		}
	}

	// Store image
	imageData := &ImageData{
		Metadata:  metadata,
		FullData:  req.Data,
		Thumbnail: thumbnail,
	}

	im.mu.Lock()

	// Store image data
	im.images[imageID] = imageData

	// Update rate limit tracking
	im.playerLastUpload[req.SenderID] = time.Now()

	// Track for disconnect-based expiry
	if im.playerDisconnect[req.SenderID] == nil {
		im.playerDisconnect[req.SenderID] = make(map[string]struct{})
	}
	im.playerDisconnect[req.SenderID][imageID] = struct{}{}

	// Set expiry timer
	timer := time.AfterFunc(ImageExpiryTime, func() {
		im.expireImage(imageID)
	})
	im.expiryTimers[imageID] = timer

	// Get relay callback
	relayCallback := im.onThumbnailRelay

	im.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name":     "image_manager",
		"playerID":        req.SenderID,
		"imageID":         imageID,
		"format":          format,
		"width":           metadata.Width,
		"height":          metadata.Height,
		"size_bytes":      metadata.Size,
		"thumbnail_bytes": len(thumbnail),
		"channel":         req.Channel,
	}).Info("image uploaded successfully")

	// Relay thumbnail to recipients (if callback set)
	if relayCallback != nil {
		relayCallback(&metadata, thumbnail)
	}

	return &metadata, nil
}

// expireImage removes an image and notifies via callback.
func (im *ImageManager) expireImage(imageID string) {
	im.mu.Lock()

	// Get image to find sender
	imageData, exists := im.images[imageID]
	if !exists {
		im.mu.Unlock()
		return
	}

	senderID := imageData.Metadata.SenderID

	// Remove from images map
	delete(im.images, imageID)

	// Remove from player disconnect tracking
	if playerImages, ok := im.playerDisconnect[senderID]; ok {
		delete(playerImages, imageID)
		if len(playerImages) == 0 {
			delete(im.playerDisconnect, senderID)
		}
	}

	// Cancel expiry timer
	if timer, ok := im.expiryTimers[imageID]; ok {
		timer.Stop()
		delete(im.expiryTimers, imageID)
	}

	// Remove any in-progress downloads
	delete(im.downloadInProgress, imageID)

	// Get callback
	callback := im.onImageExpiry

	im.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name": "image_manager",
		"imageID":     imageID,
		"playerID":    senderID,
	}).Debug("image expired")

	// Notify via callback
	if callback != nil {
		callback(imageID)
	}
}

// GetThumbnail retrieves the thumbnail for an image.
// Returns thumbnail data and error.
func (im *ImageManager) GetThumbnail(imageID string) ([]byte, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	imageData, exists := im.images[imageID]
	if !exists {
		return nil, ErrImageNotFound
	}

	// Check if expired
	if time.Now().After(imageData.Metadata.ExpiryTime) {
		return nil, ErrImageExpired
	}

	return imageData.Thumbnail, nil
}

// GetFullImage retrieves the full image data.
// Returns image data and error.
func (im *ImageManager) GetFullImage(imageID string) ([]byte, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	imageData, exists := im.images[imageID]
	if !exists {
		return nil, ErrImageNotFound
	}

	// Check if expired
	if time.Now().After(imageData.Metadata.ExpiryTime) {
		return nil, ErrImageExpired
	}

	return imageData.FullData, nil
}

// GetMetadata retrieves metadata for an image.
func (im *ImageManager) GetMetadata(imageID string) (*ImageMetadata, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	imageData, exists := im.images[imageID]
	if !exists {
		return nil, ErrImageNotFound
	}

	// Check if expired
	if time.Now().After(imageData.Metadata.ExpiryTime) {
		return nil, ErrImageExpired
	}

	metadata := imageData.Metadata
	return &metadata, nil
}

// StartChunkedUpload initiates a chunked upload for large images.
// Returns the upload state or error.
func (im *ImageManager) StartChunkedUpload(senderID uint64, channel int, format string, totalSize, totalChunks int) (string, error) {
	// Check rate limit
	if !im.CheckRateLimit(senderID) {
		return "", ErrRateLimitExceeded
	}

	// Validate size
	if totalSize > MaxImageSize {
		return "", ErrImageTooLarge
	}

	// Validate format
	if format != ImageFormatPNG && format != ImageFormatJPEG && format != ImageFormatGIF {
		return "", ErrInvalidImageType
	}

	imageID := generateImageID()

	metadata := ImageMetadata{
		ImageID:    imageID,
		SenderID:   senderID,
		Channel:    channel,
		Format:     format,
		Size:       totalSize,
		UploadTime: time.Now(),
		ExpiryTime: time.Now().Add(ImageExpiryTime), // Will be updated on completion
	}

	upload := &ChunkedUpload{
		Metadata:     metadata,
		Chunks:       make([][]byte, totalChunks),
		ReceivedMask: make([]bool, totalChunks),
		TotalChunks:  totalChunks,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}

	im.mu.Lock()
	im.uploadInProgress[imageID] = upload
	im.mu.Unlock()

	return imageID, nil
}

// ReceiveChunk processes a received chunk during chunked upload.
// Returns true if upload is complete, false otherwise.
func (im *ImageManager) ReceiveChunk(chunk *ImageChunk) (bool, error) {
	im.mu.Lock()
	upload, exists := im.uploadInProgress[chunk.ImageID]
	if !exists {
		im.mu.Unlock()
		return false, ErrImageNotFound
	}

	// Validate chunk index
	if chunk.ChunkIndex < 0 || chunk.ChunkIndex >= upload.TotalChunks {
		im.mu.Unlock()
		return false, ErrInvalidChunkSequence
	}

	// Store chunk
	upload.Chunks[chunk.ChunkIndex] = chunk.Data
	upload.ReceivedMask[chunk.ChunkIndex] = true
	upload.LastActivity = time.Now()

	// Check if all chunks received
	allReceived := true
	for _, received := range upload.ReceivedMask {
		if !received {
			allReceived = false
			break
		}
	}

	if !allReceived {
		im.mu.Unlock()
		return false, nil
	}

	// All chunks received, assemble image
	var fullData bytes.Buffer
	for _, chunkData := range upload.Chunks {
		fullData.Write(chunkData)
	}

	imageData := fullData.Bytes()
	metadata := upload.Metadata

	// Remove from in-progress
	delete(im.uploadInProgress, chunk.ImageID)

	im.mu.Unlock()

	// Complete upload (validate, generate thumbnail, store)
	req := &ImageUploadRequest{
		SenderID: metadata.SenderID,
		Channel:  metadata.Channel,
		Format:   metadata.Format,
		Data:     imageData,
	}

	// Note: This will check rate limit again, but we already checked at start
	// We update the last upload time to before the rate limit window to avoid double-counting
	im.mu.Lock()
	im.playerLastUpload[metadata.SenderID] = time.Now().Add(-ImageRateLimit - time.Second)
	im.mu.Unlock()

	_, err := im.UploadImage(req)
	if err != nil {
		return false, fmt.Errorf("failed to complete chunked upload: %w", err)
	}

	return true, nil
}

// StartChunkedDownload initiates a chunked download for an image.
// Returns the total number of chunks or error.
func (im *ImageManager) StartChunkedDownload(requesterID uint64, imageID string) (int, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	imageData, exists := im.images[imageID]
	if !exists {
		return 0, ErrImageNotFound
	}

	// Check if expired
	if time.Now().After(imageData.Metadata.ExpiryTime) {
		return 0, ErrImageExpired
	}

	// Calculate number of chunks
	totalChunks := (len(imageData.FullData) + MaxChunkSize - 1) / MaxChunkSize

	// Track download state
	if im.downloadInProgress[imageID] == nil {
		im.downloadInProgress[imageID] = make(map[uint64]*ChunkedDownload)
	}

	im.downloadInProgress[imageID][requesterID] = &ChunkedDownload{
		ImageID:      imageID,
		SentChunks:   0,
		TotalChunks:  totalChunks,
		LastActivity: time.Now(),
	}

	return totalChunks, nil
}

// GetNextChunk retrieves the next chunk for a chunked download.
// Returns the chunk or error.
func (im *ImageManager) GetNextChunk(requesterID uint64, imageID string) (*ImageChunk, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Get download state
	downloaders, exists := im.downloadInProgress[imageID]
	if !exists {
		return nil, ErrImageNotFound
	}

	download, exists := downloaders[requesterID]
	if !exists {
		return nil, fmt.Errorf("no download in progress for requester %d", requesterID)
	}

	// Get image data
	imageData, exists := im.images[imageID]
	if !exists {
		return nil, ErrImageNotFound
	}

	// Check if all chunks sent
	if download.SentChunks >= download.TotalChunks {
		// Clean up download state
		delete(downloaders, requesterID)
		if len(downloaders) == 0 {
			delete(im.downloadInProgress, imageID)
		}
		return nil, fmt.Errorf("all chunks already sent")
	}

	// Calculate chunk boundaries
	chunkIndex := download.SentChunks
	startOffset := chunkIndex * MaxChunkSize
	endOffset := min(startOffset+MaxChunkSize, len(imageData.FullData))
	chunkData := imageData.FullData[startOffset:endOffset]

	// Create chunk
	chunk := &ImageChunk{
		ImageID:     imageID,
		ChunkIndex:  chunkIndex,
		TotalChunks: download.TotalChunks,
		Data:        chunkData,
	}

	// Update download state
	download.SentChunks++
	download.LastActivity = time.Now()

	// If this was the last chunk, clean up
	if download.SentChunks >= download.TotalChunks {
		delete(downloaders, requesterID)
		if len(downloaders) == 0 {
			delete(im.downloadInProgress, imageID)
		}
	}

	return chunk, nil
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// OnPlayerDisconnect handles player disconnect, expiring all their images.
func (im *ImageManager) OnPlayerDisconnect(playerID uint64) {
	im.mu.Lock()

	// Get all images uploaded by this player
	imageIDs := make([]string, 0)
	if playerImages, ok := im.playerDisconnect[playerID]; ok {
		for imageID := range playerImages {
			imageIDs = append(imageIDs, imageID)
		}
	}

	im.mu.Unlock()

	// Expire each image
	for _, imageID := range imageIDs {
		im.expireImage(imageID)
	}
}

// CleanupStaleUploads removes stale in-progress uploads (inactive for >5 minutes).
func (im *ImageManager) CleanupStaleUploads() {
	im.mu.Lock()
	defer im.mu.Unlock()

	staleTimeout := 5 * time.Minute
	now := time.Now()

	staleIDs := make([]string, 0)
	for imageID, upload := range im.uploadInProgress {
		if now.Sub(upload.LastActivity) > staleTimeout {
			staleIDs = append(staleIDs, imageID)
		}
	}

	// Remove stale uploads
	for _, imageID := range staleIDs {
		delete(im.uploadInProgress, imageID)
	}
}

// CleanupStaleDownloads removes stale in-progress downloads (inactive for >5 minutes).
func (im *ImageManager) CleanupStaleDownloads() {
	im.mu.Lock()
	defer im.mu.Unlock()

	staleTimeout := 5 * time.Minute
	now := time.Now()

	for imageID, downloaders := range im.downloadInProgress {
		staleRequesterIDs := make([]uint64, 0)

		for requesterID, download := range downloaders {
			if now.Sub(download.LastActivity) > staleTimeout {
				staleRequesterIDs = append(staleRequesterIDs, requesterID)
			}
		}

		// Remove stale downloaders
		for _, requesterID := range staleRequesterIDs {
			delete(downloaders, requesterID)
		}

		// Clean up empty downloader map
		if len(downloaders) == 0 {
			delete(im.downloadInProgress, imageID)
		}
	}
}

// GetImageCount returns the number of images currently stored.
func (im *ImageManager) GetImageCount() int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return len(im.images)
}
