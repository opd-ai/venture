package housing

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

// SortField defines fields that blueprints can be sorted by.
type SortField int

const (
	SortByRating    SortField = iota // Sort by rating
	SortByDownloads                  // Sort by download count
	SortByCreated                    // Sort by creation date
	SortByModified                   // Sort by modification date
	SortByName                       // Sort by name (alphabetical)
)

// FilterOptions defines criteria for filtering blueprints.
type FilterOptions struct {
	Author    string   // Filter by author (exact match)
	GenreID   string   // Filter by genre (exact match)
	Tags      []string // Filter by tags (must contain all)
	MinRating float64  // Minimum rating (0.0-5.0)
	MaxRating float64  // Maximum rating (0.0-5.0)
	MinSize   int      // Minimum building size (width or height)
	MaxSize   int      // Maximum building size (width or height)
}

// BlueprintLibrary manages a collection of blueprints with search and filtering.
type BlueprintLibrary struct {
	mu         sync.RWMutex
	blueprints map[string]*Blueprint // ID -> Blueprint
}

// NewBlueprintLibrary creates a new empty blueprint library.
func NewBlueprintLibrary() *BlueprintLibrary {
	return &BlueprintLibrary{
		blueprints: make(map[string]*Blueprint),
	}
}

// Add adds a blueprint to the library.
// If a blueprint with the same ID exists, it is replaced.
func (bl *BlueprintLibrary) Add(bp *Blueprint) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.blueprints[bp.ID] = bp
}

// Get retrieves a blueprint by ID.
// Returns nil if not found.
func (bl *BlueprintLibrary) Get(id string) *Blueprint {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.blueprints[id]
}

// Remove removes a blueprint from the library by ID.
// Returns true if the blueprint was found and removed.
func (bl *BlueprintLibrary) Remove(id string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	if _, exists := bl.blueprints[id]; exists {
		delete(bl.blueprints, id)
		return true
	}
	return false
}

// List returns all blueprints in the library.
func (bl *BlueprintLibrary) List() []*Blueprint {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	result := make([]*Blueprint, 0, len(bl.blueprints))
	for _, bp := range bl.blueprints {
		result = append(result, bp)
	}
	return result
}

// Count returns the number of blueprints in the library.
func (bl *BlueprintLibrary) Count() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.blueprints)
}

// Filter returns blueprints that match the given filter options.
// Empty/zero filter values are ignored.
func (bl *BlueprintLibrary) Filter(opts FilterOptions) []*Blueprint {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	result := make([]*Blueprint, 0)
	for _, bp := range bl.blueprints {
		if !matchesFilter(bp, opts) {
			continue
		}
		result = append(result, bp)
	}
	return result
}

// matchesFilter checks if a blueprint matches the filter options.
func matchesFilter(bp *Blueprint, opts FilterOptions) bool {
	if !matchesAuthorFilter(bp, opts) {
		return false
	}

	if !matchesGenreFilter(bp, opts) {
		return false
	}

	if !matchesTagsFilter(bp, opts) {
		return false
	}

	if !matchesRatingFilter(bp, opts) {
		return false
	}

	if !matchesSizeFilter(bp, opts) {
		return false
	}

	return true
}

// matchesAuthorFilter checks if blueprint matches author filter.
func matchesAuthorFilter(bp *Blueprint, opts FilterOptions) bool {
	if opts.Author != "" && !strings.EqualFold(bp.Author, opts.Author) {
		return false
	}
	return true
}

// matchesGenreFilter checks if blueprint matches genre filter.
func matchesGenreFilter(bp *Blueprint, opts FilterOptions) bool {
	if opts.GenreID != "" && !strings.EqualFold(bp.GenreID, opts.GenreID) {
		return false
	}
	return true
}

// matchesTagsFilter checks if blueprint contains all required tags.
func matchesTagsFilter(bp *Blueprint, opts FilterOptions) bool {
	if len(opts.Tags) > 0 {
		if !containsAllTags(bp.Tags, opts.Tags) {
			return false
		}
	}
	return true
}

// matchesRatingFilter checks if blueprint rating is within specified range.
func matchesRatingFilter(bp *Blueprint, opts FilterOptions) bool {
	rating := bp.GetRating()
	if opts.MinRating > 0 && rating < opts.MinRating {
		return false
	}
	if opts.MaxRating > 0 && rating > opts.MaxRating {
		return false
	}
	return true
}

// matchesSizeFilter checks if blueprint size is within specified range.
func matchesSizeFilter(bp *Blueprint, opts FilterOptions) bool {
	if bp.BuildingDef == nil {
		return true
	}

	maxDim := bp.BuildingDef.Width
	if bp.BuildingDef.Height > maxDim {
		maxDim = bp.BuildingDef.Height
	}

	if opts.MinSize > 0 && maxDim < opts.MinSize {
		return false
	}
	if opts.MaxSize > 0 && maxDim > opts.MaxSize {
		return false
	}

	return true
}

// containsAllTags checks if the blueprint tags contain all required tags.
func containsAllTags(blueprintTags, requiredTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range blueprintTags {
		tagSet[strings.ToLower(tag)] = true
	}

	for _, required := range requiredTags {
		if !tagSet[strings.ToLower(required)] {
			return false
		}
	}
	return true
}

// Sort sorts a slice of blueprints by the specified field.
// If descending is true, sorts in descending order.
func (bl *BlueprintLibrary) Sort(blueprints []*Blueprint, field SortField, descending bool) {
	sort.Slice(blueprints, func(i, j int) bool {
		var less bool
		switch field {
		case SortByRating:
			less = blueprints[i].GetRating() < blueprints[j].GetRating()
		case SortByDownloads:
			less = blueprints[i].GetDownloads() < blueprints[j].GetDownloads()
		case SortByCreated:
			less = blueprints[i].CreatedAt.Before(blueprints[j].CreatedAt)
		case SortByModified:
			less = blueprints[i].ModifiedAt.Before(blueprints[j].ModifiedAt)
		case SortByName:
			less = strings.ToLower(blueprints[i].Name) < strings.ToLower(blueprints[j].Name)
		default:
			less = blueprints[i].CreatedAt.Before(blueprints[j].CreatedAt)
		}

		if descending {
			return !less
		}
		return less
	})
}

// Export exports a blueprint to a gzip-compressed JSON file.
func (bp *Blueprint) Export(filepath string) (err error) {
	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close blueprint file: %w", closeErr)
		}
	}()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to flush gzip writer: %w", closeErr)
		}
	}()

	// Encode as JSON
	encoder := json.NewEncoder(gzipWriter)
	encoder.SetIndent("", "  ")
	if encErr := encoder.Encode(bp); encErr != nil {
		return fmt.Errorf("failed to encode blueprint: %w", encErr)
	}
	return nil
}

// ImportBlueprint imports a blueprint from a gzip-compressed JSON file.
func ImportBlueprint(filepath string) (*Blueprint, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Decode JSON
	var bp Blueprint
	decoder := json.NewDecoder(gzipReader)
	if err := decoder.Decode(&bp); err != nil {
		return nil, fmt.Errorf("failed to decode blueprint: %w", err)
	}

	return &bp, nil
}

// IncrementDownloads increments the download count for a blueprint.
// This is typically called when a player imports/uses a blueprint.
// Thread-safe for concurrent access.
func (bp *Blueprint) IncrementDownloads() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.downloads++
}

// AddRating adds a new rating to the blueprint and updates the average.
// rating should be between 0.0 and 5.0.
// Thread-safe for concurrent access.
func (bp *Blueprint) AddRating(rating float64) error {
	if rating < 0.0 || rating > 5.0 {
		return fmt.Errorf("rating must be between 0.0 and 5.0, got %.2f", rating)
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Calculate new average rating
	totalRating := bp.rating * float64(bp.ratingCount)
	bp.ratingCount++
	bp.rating = (totalRating + rating) / float64(bp.ratingCount)

	return nil
}
