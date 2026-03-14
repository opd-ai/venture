package persistence

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"sync"
	"time"
)

// StoredImage represents a single image with metadata
type StoredImage struct {
	ID        string      `json:"id"`
	OwnerID   string      `json:"owner_id"`
	Title     string      `json:"title,omitempty"`
	Data      string      `json:"data"` // Base64-encoded image data
	Format    ImageFormat `json:"format"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	SizeBytes int         `json:"size_bytes"`
	Hash      string      `json:"hash"` // SHA256 hash for deduplication
	Timestamp time.Time   `json:"timestamp"`
	Tags      []string    `json:"tags,omitempty"`
}

// ImageGallery manages persistent image storage for a player
type ImageGallery struct {
	PlayerID     string         `json:"player_id"`
	Images       []*StoredImage `json:"images"`
	TotalBytes   int            `json:"total_bytes"`
	mu           sync.RWMutex
	timeProvider TimeProvider `json:"-"` // TimeProvider for deterministic timestamps
}

// NewImageGallery creates a new image gallery manager using real system time.
func NewImageGallery(playerID string) *ImageGallery {
	return NewImageGalleryWithTimeProvider(playerID, DefaultTimeProvider())
}

// NewImageGalleryWithTimeProvider creates a new image gallery manager with a custom TimeProvider.
// Use this constructor in tests to inject a mock TimeProvider for deterministic timestamps.
func NewImageGalleryWithTimeProvider(playerID string, tp TimeProvider) *ImageGallery {
	return &ImageGallery{
		PlayerID:     playerID,
		Images:       make([]*StoredImage, 0, MaxImagesPerPlayer),
		TotalBytes:   0,
		timeProvider: tp,
	}
}

// AddImage adds a new image to the gallery
// encodeImageToBytes encodes an image to bytes in the specified format
func encodeImageToBytes(img image.Image, format ImageFormat) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	switch format {
	case ImageFormatPNG:
		err = png.Encode(&buf, img)
	case ImageFormatJPEG:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return buf.Bytes(), nil
}

// findDuplicateImage checks if an image with the same hash already exists
func (g *ImageGallery) findDuplicateImage(hash string) *StoredImage {
	for _, existing := range g.Images {
		if existing.Hash == hash {
			return existing
		}
	}
	return nil
}

// evictOldestImage removes the oldest image from the gallery
func (g *ImageGallery) evictOldestImage() {
	if len(g.Images) > 0 {
		removed := g.Images[0]
		g.Images = g.Images[1:]
		g.TotalBytes -= removed.SizeBytes
	}
}

// evictImagesToFitSize removes images until there is space for the new image
func (g *ImageGallery) evictImagesToFitSize(newImageSize int) {
	maxTotalBytes := MaxImagesPerPlayer * MaxImageSizeBytes
	for g.TotalBytes+newImageSize > maxTotalBytes && len(g.Images) > 0 {
		g.evictOldestImage()
	}
}

// createStoredImage creates a new StoredImage record using a pre-computed hash
func (g *ImageGallery) createStoredImage(img image.Image, title string, imageData []byte, format ImageFormat, tags []string, hash string) *StoredImage {
	bounds := img.Bounds()
	sizeBytes := len(imageData)
	now := g.timeProvider.Now()

	return &StoredImage{
		ID:        fmt.Sprintf("%s-%d", g.PlayerID, now.UnixNano()),
		OwnerID:   g.PlayerID,
		Title:     title,
		Data:      base64.StdEncoding.EncodeToString(imageData),
		Format:    format,
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		SizeBytes: sizeBytes,
		Hash:      hash,
		Timestamp: now,
		Tags:      tags,
	}
}

// AddImage adds a new image to the gallery.
// The image is encoded to the specified format, deduplicated via SHA256 hash,
// and stored with base64 encoding. If the gallery exceeds MaxImagesPerPlayer,
// the oldest images are evicted (LRU). Duplicate images (same hash) return
// the existing stored image without creating a new entry.
// Returns the stored image or an error if the image is nil, the format is
// unsupported, or the encoded size exceeds MaxImageSizeBytes.
func (g *ImageGallery) AddImage(img image.Image, title string, format ImageFormat, tags []string) (*StoredImage, error) {
	if img == nil {
		return nil, fmt.Errorf("image cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	imageData, err := encodeImageToBytes(img, format)
	if err != nil {
		return nil, err
	}

	sizeBytes := len(imageData)
	if sizeBytes > MaxImageSizeBytes {
		return nil, fmt.Errorf("image size %d bytes exceeds maximum %d bytes", sizeBytes, MaxImageSizeBytes)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(imageData))

	if duplicate := g.findDuplicateImage(hash); duplicate != nil {
		return duplicate, nil
	}

	if len(g.Images) >= MaxImagesPerPlayer {
		g.evictOldestImage()
	}

	g.evictImagesToFitSize(sizeBytes)

	stored := g.createStoredImage(img, title, imageData, format, tags, hash)

	g.Images = append(g.Images, stored)
	g.TotalBytes += sizeBytes

	return stored, nil
}

// GetImage retrieves an image by ID
func (g *ImageGallery) GetImage(id string) (*StoredImage, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, img := range g.Images {
		if img.ID == id {
			return img, nil
		}
	}

	return nil, fmt.Errorf("image not found: %s", id)
}

// GetAllImages returns all images in the gallery
func (g *ImageGallery) GetAllImages() []*StoredImage {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*StoredImage, len(g.Images))
	copy(result, g.Images)
	return result
}

// GetImagesByTag returns images matching the given tag
func (g *ImageGallery) GetImagesByTag(tag string) []*StoredImage {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*StoredImage, 0)
	for _, img := range g.Images {
		for _, t := range img.Tags {
			if t == tag {
				result = append(result, img)
				break
			}
		}
	}

	return result
}

// DeleteImage removes an image from the gallery
func (g *ImageGallery) DeleteImage(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i, img := range g.Images {
		if img.ID == id {
			g.Images = append(g.Images[:i], g.Images[i+1:]...)
			g.TotalBytes -= img.SizeBytes
			return nil
		}
	}

	return fmt.Errorf("image not found: %s", id)
}

// GetImageCount returns the number of images in the gallery
func (g *ImageGallery) GetImageCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Images)
}

// GetTotalSize returns the total size of all images in bytes
func (g *ImageGallery) GetTotalSize() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.TotalBytes
}

// DecodeImage decodes a StoredImage back to an image.Image
func (g *ImageGallery) DecodeImage(stored *StoredImage) (image.Image, error) {
	if stored == nil {
		return nil, fmt.Errorf("stored image cannot be nil")
	}

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(stored.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decode image
	reader := bytes.NewReader(imageData)
	var img image.Image
	switch stored.Format {
	case ImageFormatPNG:
		img, err = png.Decode(reader)
	case ImageFormatJPEG:
		img, err = jpeg.Decode(reader)
	default:
		return nil, fmt.Errorf("unsupported format: %s", stored.Format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// Save serializes the gallery to compressed JSON
func (g *ImageGallery) Save() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Marshal to JSON
	data, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gallery: %w", err)
	}

	// Compress with gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress gallery: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes the gallery from compressed JSON
func (g *ImageGallery) Load(data []byte) error {
	// Decompress gzip
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return fmt.Errorf("failed to decompress gallery: %w", err)
	}

	// Unmarshal JSON
	var temp ImageGallery
	if err := json.Unmarshal(buf.Bytes(), &temp); err != nil {
		return fmt.Errorf("failed to unmarshal gallery: %w", err)
	}

	// Update gallery with loaded data
	g.mu.Lock()
	defer g.mu.Unlock()

	g.PlayerID = temp.PlayerID
	g.Images = temp.Images
	g.TotalBytes = temp.TotalBytes
	// timeProvider is preserved from the original gallery (not serialized)
	// If nil after load, initialize with default
	if g.timeProvider == nil {
		g.timeProvider = DefaultTimeProvider()
	}

	return nil
}

// SetTimeProvider sets the time provider for the gallery.
// This is useful when loading a gallery from JSON and needing to inject a mock time provider.
func (g *ImageGallery) SetTimeProvider(tp TimeProvider) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.timeProvider = tp
}

// Clear removes all images from the gallery
func (g *ImageGallery) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Images = make([]*StoredImage, 0, MaxImagesPerPlayer)
	g.TotalBytes = 0
}

// GetThumbnails returns lightweight metadata for all images (without image data)
func (g *ImageGallery) GetThumbnails() []ImageThumbnail {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]ImageThumbnail, len(g.Images))
	for i, img := range g.Images {
		result[i] = ImageThumbnail{
			ID:        img.ID,
			OwnerID:   img.OwnerID,
			Title:     img.Title,
			Format:    img.Format,
			Width:     img.Width,
			Height:    img.Height,
			SizeBytes: img.SizeBytes,
			Timestamp: img.Timestamp,
			Tags:      img.Tags,
		}
	}

	return result
}
