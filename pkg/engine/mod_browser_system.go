// Package engine provides the mod browser system for in-game mod management.
// ModBrowserSystem manages mod repository interactions, mod installation/uninstallation,
// and integrates with the existing modding.Manager for mod loading.
package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
	"github.com/sirupsen/logrus"
)

// ModBrowserSystem manages mod browsing and installation.
// It processes entities with ModBrowserComponent to handle repository
// fetching, mod installation, and uninstallation.
type ModBrowserSystem struct {
	world             *World
	repository        ModRepository
	installCallback   ModInstallCallback
	uninstallCallback ModUninstallCallback

	mu sync.RWMutex
}

// ModRepository defines the interface for fetching mods from a repository.
// This allows for mock implementations in testing and different backends.
type ModRepository interface {
	// FetchMods retrieves available mods from the repository.
	FetchMods() ([]ModListing, error)

	// DownloadMod downloads a mod by ID. Returns mod data on success.
	DownloadMod(modID string, progressCallback func(downloaded, total int64)) ([]byte, error)

	// GetModDetails retrieves detailed information for a specific mod.
	GetModDetails(modID string) (*ModListing, error)
}

// ModInstallCallback is called when a mod is installed.
type ModInstallCallback func(modID string, modData []byte) error

// ModUninstallCallback is called when a mod is uninstalled.
type ModUninstallCallback func(modID string) error

// NewModBrowserSystem creates a new mod browser system.
func NewModBrowserSystem(world *World) *ModBrowserSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "mod_browser",
	}).Debug("Creating mod browser system")

	return &ModBrowserSystem{
		world: world,
	}
}

// SetRepository sets the mod repository backend.
func (s *ModBrowserSystem) SetRepository(repo ModRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repository = repo
}

// SetInstallCallback sets the callback for mod installation.
func (s *ModBrowserSystem) SetInstallCallback(callback ModInstallCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.installCallback = callback
}

// SetUninstallCallback sets the callback for mod uninstallation.
func (s *ModBrowserSystem) SetUninstallCallback(callback ModUninstallCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uninstallCallback = callback
}

// Update processes mod browser entities and handles pending operations.
func (s *ModBrowserSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("mod_browser") {
			continue
		}

		comp, ok := entity.GetComponent("mod_browser")
		if !ok || comp == nil {
			continue
		}
		browserComp := comp.(*ModBrowserComponent)

		// Process pending refresh
		if browserComp.RefreshPending {
			s.processRefresh(browserComp)
		}

		// Process active downloads
		s.processDownloads(browserComp)
	}
}

// processRefresh handles repository refresh requests.
func (s *ModBrowserSystem) processRefresh(comp *ModBrowserComponent) {
	s.mu.RLock()
	repo := s.repository
	s.mu.RUnlock()

	if repo == nil {
		logrus.Warn("mod browser: no repository configured")
		comp.RefreshPending = false
		return
	}

	mods, err := repo.FetchMods()
	if err != nil {
		logrus.WithError(err).Error("mod browser: failed to fetch mods from repository")
		comp.RefreshPending = false
		return
	}

	comp.SetAvailableMods(mods)
	logrus.WithFields(logrus.Fields{
		"mod_count": len(mods),
	}).Debug("mod browser: refreshed mod listing")
}

// processDownloads handles active downloads.
func (s *ModBrowserSystem) processDownloads(comp *ModBrowserComponent) {
	downloads := comp.GetActiveDownloads()

	for _, download := range downloads {
		if download.Status == "complete" || download.Status == "failed" {
			continue
		}

		// Start downloading if pending
		if download.Status == "pending" {
			go s.downloadMod(comp, download.ModID)
		}
	}
}

// downloadMod handles the async download of a mod.
func (s *ModBrowserSystem) downloadMod(comp *ModBrowserComponent, modID string) {
	defer recovery.RecoverPanicWithLogger("mod_browser", "download mod", func() {
		// Cleanup: mark download as failed on panic
		comp.SetDownloadStatus(modID, "failed", "unexpected error during download")
	})()

	s.mu.RLock()
	repo := s.repository
	installCb := s.installCallback
	s.mu.RUnlock()

	if repo == nil {
		comp.SetDownloadStatus(modID, "failed", "no repository configured")
		return
	}

	// Update status to downloading
	comp.SetDownloadStatus(modID, "downloading", "")

	// Create progress callback
	progressCallback := func(downloaded, total int64) {
		comp.UpdateDownloadProgress(modID, downloaded)
	}

	// Download the mod
	modData, err := repo.DownloadMod(modID, progressCallback)
	if err != nil {
		comp.SetDownloadStatus(modID, "failed", err.Error())
		logrus.WithFields(logrus.Fields{
			"mod_id": modID,
		}).WithError(err).Error("mod browser: failed to download mod")
		return
	}

	// Update to installing status
	comp.SetDownloadStatus(modID, "installing", "")

	// Call install callback if set
	if installCb != nil {
		if err := installCb(modID, modData); err != nil {
			comp.SetDownloadStatus(modID, "failed", err.Error())
			logrus.WithFields(logrus.Fields{
				"mod_id": modID,
			}).WithError(err).Error("mod browser: failed to install mod")
			return
		}
	}

	// Mark as complete
	comp.CompleteDownload(modID)
	logrus.WithFields(logrus.Fields{
		"mod_id": modID,
	}).Info("mod browser: mod installed successfully")
}

// RefreshRepository triggers a refresh of available mods.
func (s *ModBrowserSystem) RefreshRepository(comp *ModBrowserComponent) {
	if comp == nil {
		return
	}
	comp.RefreshPending = true
}

// InstallMod initiates the installation of a mod.
func (s *ModBrowserSystem) InstallMod(comp *ModBrowserComponent, modID string) error {
	if comp == nil {
		return fmt.Errorf("mod browser component is nil")
	}

	// Check if already installed
	if comp.IsInstalled(modID) {
		return fmt.Errorf("mod %s is already installed", modID)
	}

	// Check if download is already in progress
	if download, exists := comp.GetDownload(modID); exists && download.Status != "failed" {
		return fmt.Errorf("download already in progress for mod %s", modID)
	}

	// Get mod details for size
	mod, found := comp.GetMod(modID)
	if !found {
		return fmt.Errorf("mod %s not found in repository", modID)
	}

	// Check dependencies
	if missing, ok := comp.CheckDependencies(modID); !ok {
		return fmt.Errorf("missing dependencies: %v", missing)
	}

	// Start download
	return comp.StartDownload(modID, mod.Size)
}

// UninstallMod removes an installed mod.
func (s *ModBrowserSystem) UninstallMod(comp *ModBrowserComponent, modID string) error {
	if comp == nil {
		return fmt.Errorf("mod browser component is nil")
	}

	// Check if installed
	if !comp.IsInstalled(modID) {
		return fmt.Errorf("mod %s is not installed", modID)
	}

	// Check for dependents
	dependents := s.findDependents(comp, modID)
	if len(dependents) > 0 {
		return fmt.Errorf("cannot uninstall: mods %v depend on %s", dependents, modID)
	}

	s.mu.RLock()
	uninstallCb := s.uninstallCallback
	s.mu.RUnlock()

	// Call uninstall callback if set
	if uninstallCb != nil {
		if err := uninstallCb(modID); err != nil {
			return fmt.Errorf("failed to uninstall mod: %w", err)
		}
	}

	// Mark as uninstalled
	comp.SetInstalled(modID, false)

	logrus.WithFields(logrus.Fields{
		"mod_id": modID,
	}).Info("mod browser: mod uninstalled")

	return nil
}

// findDependents returns IDs of installed mods that depend on the given mod.
func (s *ModBrowserSystem) findDependents(comp *ModBrowserComponent, modID string) []string {
	var dependents []string

	installed := comp.GetInstalledMods()
	for _, installedID := range installed {
		mod, found := comp.GetMod(installedID)
		if !found {
			continue
		}

		for _, depID := range mod.Dependencies {
			if depID == modID {
				dependents = append(dependents, installedID)
				break
			}
		}
	}

	return dependents
}

// GetModsByCategory returns mods filtered by category.
func (s *ModBrowserSystem) GetModsByCategory(comp *ModBrowserComponent, category string) []ModListing {
	if comp == nil {
		return nil
	}

	comp.SetActiveCategory(category)
	return comp.GetFilteredMods()
}

// SearchMods searches for mods by name or description.
func (s *ModBrowserSystem) SearchMods(comp *ModBrowserComponent, query string) []ModListing {
	if comp == nil {
		return nil
	}

	comp.SetSearchQuery(query)
	return comp.GetFilteredMods()
}

// GetRecommendedMods returns recommended mods based on installed mods and ratings.
func (s *ModBrowserSystem) GetRecommendedMods(comp *ModBrowserComponent, limit int) []ModListing {
	if comp == nil {
		return nil
	}

	// Get installed mod categories
	installedCategories := make(map[string]int)
	for _, installedID := range comp.GetInstalledMods() {
		if mod, found := comp.GetMod(installedID); found {
			for _, cat := range mod.Categories {
				installedCategories[cat]++
			}
		}
	}

	// Score available mods
	type scoredMod struct {
		mod   ModListing
		score float64
	}

	var scored []scoredMod
	for _, mod := range comp.AvailableMods {
		// Skip installed mods
		if comp.IsInstalled(mod.ID) {
			continue
		}

		// Calculate score based on rating and category match
		score := mod.Rating

		// Boost score for matching categories
		for _, cat := range mod.Categories {
			if count, ok := installedCategories[cat]; ok {
				score += float64(count) * 0.5
			}
		}

		// Boost featured mods
		if mod.Featured {
			score += 1.0
		}

		scored = append(scored, scoredMod{mod: mod, score: score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top N
	result := make([]ModListing, 0, limit)
	for i := 0; i < len(scored) && i < limit; i++ {
		result = append(result, scored[i].mod)
	}

	return result
}

// CheckGameVersionCompatibility checks if a mod is compatible with the game version.
func (s *ModBrowserSystem) CheckGameVersionCompatibility(mod *ModListing, gameVersion string) bool {
	if mod == nil || mod.GameVersion == "" {
		return true // No version requirement
	}

	return compareVersions(mod.GameVersion, gameVersion) <= 0
}

// compareVersions compares two semantic version strings.
// Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2.
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// InMemoryModRepository provides a simple in-memory mod repository for testing.
type InMemoryModRepository struct {
	mods    []ModListing
	modData map[string][]byte
	mu      sync.RWMutex
}

// NewInMemoryModRepository creates a new in-memory mod repository.
func NewInMemoryModRepository() *InMemoryModRepository {
	return &InMemoryModRepository{
		mods:    make([]ModListing, 0),
		modData: make(map[string][]byte),
	}
}

// AddMod adds a mod to the repository.
func (r *InMemoryModRepository) AddMod(mod ModListing, data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mods = append(r.mods, mod)
	if data != nil {
		r.modData[mod.ID] = data
	}
}

// FetchMods implements ModRepository.
func (r *InMemoryModRepository) FetchMods() ([]ModListing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ModListing, len(r.mods))
	copy(result, r.mods)
	return result, nil
}

// DownloadMod implements ModRepository.
func (r *InMemoryModRepository) DownloadMod(modID string, progressCallback func(downloaded, total int64)) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.modData[modID]
	if !exists {
		return nil, fmt.Errorf("mod %s not found", modID)
	}

	// Simulate progress
	if progressCallback != nil {
		total := int64(len(data))
		step := total / 10
		if step < 1 {
			step = 1
		}
		for i := int64(0); i <= total; i += step {
			progressCallback(i, total)
			time.Sleep(time.Millisecond) // Simulate network delay
		}
		progressCallback(total, total)
	}

	return data, nil
}

// GetModDetails implements ModRepository.
func (r *InMemoryModRepository) GetModDetails(modID string) (*ModListing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := range r.mods {
		if r.mods[i].ID == modID {
			return &r.mods[i], nil
		}
	}

	return nil, fmt.Errorf("mod %s not found", modID)
}

// GenerateModListing generates a deterministic mod listing for testing.
func GenerateModListing(seed int64) ModListing {
	// Use seed for deterministic data
	id := fmt.Sprintf("mod_%d", seed)

	categories := []string{"gameplay", "graphics", "audio", "ui", "content"}
	catIndex := int(seed % int64(len(categories)))

	return ModListing{
		ID:          id,
		Name:        fmt.Sprintf("Test Mod %d", seed),
		Author:      fmt.Sprintf("Author%d", seed%10),
		Description: fmt.Sprintf("This is test mod %d with seed-based content", seed),
		Rating:      float64(seed%5) + float64(seed%10)/10.0,
		RatingCount: int(seed % 1000),
		Downloads:   int(seed * 100),
		Version:     fmt.Sprintf("%d.%d.%d", seed%10, seed%5, seed%3),
		GameVersion: "10.0.0",
		Categories:  []string{categories[catIndex]},
		Size:        seed * 1024,
		UploadedAt:  seed * 1000,
		UpdatedAt:   seed * 1500,
		Featured:    seed%7 == 0,
	}
}

// ModBrowserState represents the serializable state of the mod browser.
type ModBrowserState struct {
	InstalledMods []string `json:"installed_mods"`
	LastRefresh   int64    `json:"last_refresh"`
}

// SerializeModBrowserState serializes the mod browser component to JSON.
func SerializeModBrowserState(comp *ModBrowserComponent) ([]byte, error) {
	if comp == nil {
		return nil, fmt.Errorf("component is nil")
	}

	state := ModBrowserState{
		InstalledMods: comp.GetInstalledMods(),
		LastRefresh:   comp.LastRefresh,
	}

	return json.Marshal(state)
}

// DeserializeModBrowserState deserializes JSON to mod browser component.
func DeserializeModBrowserState(comp *ModBrowserComponent, data []byte) error {
	if comp == nil {
		return fmt.Errorf("component is nil")
	}

	var state ModBrowserState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	for _, modID := range state.InstalledMods {
		comp.SetInstalled(modID, true)
	}
	comp.LastRefresh = state.LastRefresh

	return nil
}
