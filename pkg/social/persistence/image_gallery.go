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

// MaxImagesPerPlayer is the maximum number of images per player
const MaxImagesPerPlayer = 100

// MaxImageSizeBytes is the maximum size per image (500KB)
const MaxImageSizeBytes = 500 * 1024

// ImageFormat represents supported image formats
type ImageFormat string

const (
	// ImageFormatPNG is the PNG format
	ImageFormatPNG ImageFormat = "png"
	// ImageFormatJPEG is the JPEG format
	ImageFormatJPEG ImageFormat = "jpeg"
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
	PlayerID   string         `json:"player_id"`
	Images     []*StoredImage `json:"images"`
	TotalBytes int            `json:"total_bytes"`
	mu         sync.RWMutex
}

// NewImageGallery creates a new image gallery manager
func NewImageGallery(playerID string) *ImageGallery {
	return &ImageGallery{
		PlayerID:   playerID,
		Images:     make([]*StoredImage, 0, MaxImagesPerPlayer),
		TotalBytes: 0,
	}
}

// AddImage adds a new image to the gallery
func (g *ImageGallery) AddImage(img image.Image, title string, format ImageFormat, tags []string) (*StoredImage, error) {
	if img == nil {
		return nil, fmt.Errorf("image cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Encode image to bytes
	var buf bytes.Buffer
	var err error
	switch format {
	case ImageFormatPNG:
		err = png.Encode(&buf, img)
	case ImageFormatJPEG:
		// Use quality 85 for good compression/quality balance
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	imageData := buf.Bytes()
	sizeBytes := len(imageData)

	// Check size limit
	if sizeBytes > MaxImageSizeBytes {
		return nil, fmt.Errorf("image size %d bytes exceeds maximum %d bytes", sizeBytes, MaxImageSizeBytes)
	}

	// Calculate hash for deduplication
	hash := fmt.Sprintf("%x", sha256.Sum256(imageData))

	// Check for duplicate
	for _, existing := range g.Images {
		if existing.Hash == hash {
			return existing, nil // Return existing image
		}
	}

	// Check if we need to evict (LRU)
	if len(g.Images) >= MaxImagesPerPlayer {
		// Remove oldest image
		removed := g.Images[0]
		g.Images = g.Images[1:]
		g.TotalBytes -= removed.SizeBytes
	}

	// Check if adding this image would exceed total storage limit
	maxTotalBytes := MaxImagesPerPlayer * MaxImageSizeBytes
	if g.TotalBytes+sizeBytes > maxTotalBytes {
		// Need to remove more images
		for g.TotalBytes+sizeBytes > maxTotalBytes && len(g.Images) > 0 {
			removed := g.Images[0]
			g.Images = g.Images[1:]
			g.TotalBytes -= removed.SizeBytes
		}
	}

	// Create stored image
	bounds := img.Bounds()
	stored := &StoredImage{
		ID:        fmt.Sprintf("%s-%d", g.PlayerID, time.Now().UnixNano()),
		OwnerID:   g.PlayerID,
		Title:     title,
		Data:      base64.StdEncoding.EncodeToString(imageData),
		Format:    format,
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		SizeBytes: sizeBytes,
		Hash:      hash,
		Timestamp: time.Now(),
		Tags:      tags,
	}

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

	return nil
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

// ImageThumbnail is lightweight metadata without image data
type ImageThumbnail struct {
	ID        string      `json:"id"`
	OwnerID   string      `json:"owner_id"`
	Title     string      `json:"title,omitempty"`
	Format    ImageFormat `json:"format"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	SizeBytes int         `json:"size_bytes"`
	Timestamp time.Time   `json:"timestamp"`
	Tags      []string    `json:"tags,omitempty"`
}
