// Package engine provides the mod browser component for in-game mod discovery.
// ModBrowserComponent tracks available mods from repositories, installed mods,
// and user browsing state including search, filtering, and download progress.
package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ModBrowserComponent tracks mod browsing and installation state.
type ModBrowserComponent struct {
	AvailableMods  []ModListing         // mods from repository
	InstalledMods  map[string]bool      // modID -> installed
	Categories     []string             // available categories
	SearchQuery    string               // current search
	SortBy         ModSortField         // rating, downloads, date, name
	SortDescending bool                 // sort direction
	ActiveCategory string               // filter by category (empty = all)
	Downloads      map[string]*Download // modID -> active download
	LastRefresh    int64                // unix timestamp of last repository fetch
	RefreshPending bool                 // true if refresh in progress

	mu sync.RWMutex
}

// ModListing represents a mod available in the repository.
type ModListing struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Author       string   `json:"author"`
	Description  string   `json:"description"`
	Rating       float64  `json:"rating"`       // 0.0-5.0
	RatingCount  int      `json:"rating_count"` // number of ratings
	Downloads    int      `json:"downloads"`    // total downloads
	Version      string   `json:"version"`      // mod version
	GameVersion  string   `json:"game_version"` // minimum game version
	Categories   []string `json:"categories"`   // mod categories
	Size         int64    `json:"size"`         // size in bytes
	UploadedAt   int64    `json:"uploaded_at"`  // unix timestamp
	UpdatedAt    int64    `json:"updated_at"`   // unix timestamp
	Dependencies []string `json:"dependencies"` // required mod IDs
	Screenshots  []string `json:"screenshots"`  // screenshot URLs
	Featured     bool     `json:"featured"`     // featured in repository
}

// Download represents an active mod download.
type Download struct {
	ModID           string  `json:"mod_id"`
	StartedAt       int64   `json:"started_at"`       // unix timestamp
	Progress        float64 `json:"progress"`         // 0.0-1.0
	TotalBytes      int64   `json:"total_bytes"`      // total size
	DownloadedBytes int64   `json:"downloaded_bytes"` // bytes downloaded
	Status          string  `json:"status"`           // pending, downloading, installing, complete, failed
	Error           string  `json:"error"`            // error message if failed
}

// ModSortField represents sorting options for mod listings.
type ModSortField string

const (
	ModSortByRating    ModSortField = "rating"
	ModSortByDownloads ModSortField = "downloads"
	ModSortByDate      ModSortField = "date"
	ModSortByName      ModSortField = "name"
	ModSortByUpdated   ModSortField = "updated"
)

// NewModBrowserComponent creates a new mod browser component.
func NewModBrowserComponent() *ModBrowserComponent {
	return &ModBrowserComponent{
		AvailableMods:  make([]ModListing, 0),
		InstalledMods:  make(map[string]bool),
		Categories:     defaultModCategories(),
		SortBy:         ModSortByRating,
		SortDescending: true,
		Downloads:      make(map[string]*Download),
	}
}

// defaultModCategories returns the default mod category list.
func defaultModCategories() []string {
	return []string{
		"gameplay",
		"graphics",
		"audio",
		"ui",
		"content",
		"items",
		"quests",
		"npcs",
		"balance",
		"utility",
		"quality-of-life",
		"cheats",
	}
}

// Type returns the component type identifier.
func (m *ModBrowserComponent) Type() string {
	return "mod_browser"
}

// SetAvailableMods updates the list of available mods from repository.
func (m *ModBrowserComponent) SetAvailableMods(mods []ModListing) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AvailableMods = mods
	m.LastRefresh = time.Now().Unix()
	m.RefreshPending = false

	// Extract unique categories from mods
	categorySet := make(map[string]bool)
	for _, cat := range defaultModCategories() {
		categorySet[cat] = true
	}
	for _, mod := range mods {
		for _, cat := range mod.Categories {
			categorySet[cat] = true
		}
	}

	m.Categories = make([]string, 0, len(categorySet))
	for cat := range categorySet {
		m.Categories = append(m.Categories, cat)
	}
	sort.Strings(m.Categories)
}

// GetFilteredMods returns mods filtered by search query and category.
func (m *ModBrowserComponent) GetFilteredMods() []ModListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []ModListing

	query := strings.ToLower(m.SearchQuery)

	for _, mod := range m.AvailableMods {
		// Filter by category if set
		if m.ActiveCategory != "" {
			categoryMatch := false
			for _, cat := range mod.Categories {
				if cat == m.ActiveCategory {
					categoryMatch = true
					break
				}
			}
			if !categoryMatch {
				continue
			}
		}

		// Filter by search query
		if query != "" {
			nameMatch := strings.Contains(strings.ToLower(mod.Name), query)
			descMatch := strings.Contains(strings.ToLower(mod.Description), query)
			authorMatch := strings.Contains(strings.ToLower(mod.Author), query)
			if !nameMatch && !descMatch && !authorMatch {
				continue
			}
		}

		filtered = append(filtered, mod)
	}

	// Sort filtered results
	m.sortMods(filtered)

	return filtered
}

// sortMods sorts the mod list based on current sort settings.
func (m *ModBrowserComponent) sortMods(mods []ModListing) {
	sort.Slice(mods, func(i, j int) bool {
		var less bool
		switch m.SortBy {
		case ModSortByRating:
			less = mods[i].Rating < mods[j].Rating
		case ModSortByDownloads:
			less = mods[i].Downloads < mods[j].Downloads
		case ModSortByDate:
			less = mods[i].UploadedAt < mods[j].UploadedAt
		case ModSortByUpdated:
			less = mods[i].UpdatedAt < mods[j].UpdatedAt
		case ModSortByName:
			less = strings.ToLower(mods[i].Name) < strings.ToLower(mods[j].Name)
		default:
			less = mods[i].Rating < mods[j].Rating
		}

		if m.SortDescending {
			return !less
		}
		return less
	})
}

// SetSearchQuery updates the search query.
func (m *ModBrowserComponent) SetSearchQuery(query string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SearchQuery = query
}

// SetActiveCategory sets the category filter.
func (m *ModBrowserComponent) SetActiveCategory(category string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveCategory = category
}

// SetSortBy sets the sort field and direction.
func (m *ModBrowserComponent) SetSortBy(field ModSortField, descending bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SortBy = field
	m.SortDescending = descending
}

// GetMod returns a specific mod by ID.
func (m *ModBrowserComponent) GetMod(modID string) (*ModListing, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := range m.AvailableMods {
		if m.AvailableMods[i].ID == modID {
			return &m.AvailableMods[i], true
		}
	}
	return nil, false
}

// IsInstalled checks if a mod is installed.
func (m *ModBrowserComponent) IsInstalled(modID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.InstalledMods[modID]
}

// SetInstalled marks a mod as installed or uninstalled.
func (m *ModBrowserComponent) SetInstalled(modID string, installed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if installed {
		m.InstalledMods[modID] = true
	} else {
		delete(m.InstalledMods, modID)
	}
}

// StartDownload initiates a mod download.
func (m *ModBrowserComponent) StartDownload(modID string, totalBytes int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Downloads[modID]; exists {
		return fmt.Errorf("download already in progress for mod %s", modID)
	}

	m.Downloads[modID] = &Download{
		ModID:      modID,
		StartedAt:  time.Now().Unix(),
		Progress:   0.0,
		TotalBytes: totalBytes,
		Status:     "pending",
	}

	return nil
}

// UpdateDownloadProgress updates the progress of a download.
func (m *ModBrowserComponent) UpdateDownloadProgress(modID string, downloadedBytes int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	download, exists := m.Downloads[modID]
	if !exists {
		return fmt.Errorf("no download in progress for mod %s", modID)
	}

	download.DownloadedBytes = downloadedBytes
	if download.TotalBytes > 0 {
		download.Progress = float64(downloadedBytes) / float64(download.TotalBytes)
	}
	download.Status = "downloading"

	return nil
}

// SetDownloadStatus updates the status of a download.
func (m *ModBrowserComponent) SetDownloadStatus(modID, status, err string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	download, exists := m.Downloads[modID]
	if !exists {
		return fmt.Errorf("no download in progress for mod %s", modID)
	}

	download.Status = status
	download.Error = err

	return nil
}

// CompleteDownload marks a download as complete and removes it.
func (m *ModBrowserComponent) CompleteDownload(modID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if download, exists := m.Downloads[modID]; exists {
		download.Status = "complete"
		download.Progress = 1.0
	}

	// Mark as installed
	m.InstalledMods[modID] = true
}

// CancelDownload cancels and removes a download.
func (m *ModBrowserComponent) CancelDownload(modID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Downloads, modID)
}

// GetDownload returns the download state for a mod.
func (m *ModBrowserComponent) GetDownload(modID string) (*Download, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	download, exists := m.Downloads[modID]
	return download, exists
}

// GetActiveDownloads returns all active downloads.
func (m *ModBrowserComponent) GetActiveDownloads() []*Download {
	m.mu.RLock()
	defer m.mu.RUnlock()

	downloads := make([]*Download, 0, len(m.Downloads))
	for _, d := range m.Downloads {
		downloads = append(downloads, d)
	}
	return downloads
}

// GetInstalledModCount returns the number of installed mods.
func (m *ModBrowserComponent) GetInstalledModCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.InstalledMods)
}

// GetAvailableModCount returns the number of available mods.
func (m *ModBrowserComponent) GetAvailableModCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.AvailableMods)
}

// GetFeaturedMods returns featured mods from the repository.
func (m *ModBrowserComponent) GetFeaturedMods() []ModListing {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var featured []ModListing
	for _, mod := range m.AvailableMods {
		if mod.Featured {
			featured = append(featured, mod)
		}
	}
	return featured
}

// GetInstalledMods returns all installed mod IDs.
func (m *ModBrowserComponent) GetInstalledMods() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mods := make([]string, 0, len(m.InstalledMods))
	for modID := range m.InstalledMods {
		mods = append(mods, modID)
	}
	sort.Strings(mods)
	return mods
}

// CheckDependencies checks if all dependencies for a mod are installed.
func (m *ModBrowserComponent) CheckDependencies(modID string) (missing []string, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mod, found := m.getMod(modID)
	if !found {
		return nil, false
	}

	for _, depID := range mod.Dependencies {
		if !m.InstalledMods[depID] {
			missing = append(missing, depID)
		}
	}

	return missing, len(missing) == 0
}

// getMod is an internal helper to find a mod by ID (no locking).
func (m *ModBrowserComponent) getMod(modID string) (*ModListing, bool) {
	for i := range m.AvailableMods {
		if m.AvailableMods[i].ID == modID {
			return &m.AvailableMods[i], true
		}
	}
	return nil, false
}

// ModBrowserData represents serializable mod browser data.
type ModBrowserData struct {
	InstalledMods  []string     `json:"installed_mods"`
	SearchQuery    string       `json:"search_query"`
	SortBy         ModSortField `json:"sort_by"`
	SortDescending bool         `json:"sort_descending"`
	ActiveCategory string       `json:"active_category"`
	LastRefresh    int64        `json:"last_refresh"`
}

// Serialize converts the component to JSON bytes.
func (m *ModBrowserComponent) Serialize() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	installedMods := make([]string, 0, len(m.InstalledMods))
	for modID := range m.InstalledMods {
		installedMods = append(installedMods, modID)
	}
	sort.Strings(installedMods)

	data := ModBrowserData{
		InstalledMods:  installedMods,
		SearchQuery:    m.SearchQuery,
		SortBy:         m.SortBy,
		SortDescending: m.SortDescending,
		ActiveCategory: m.ActiveCategory,
		LastRefresh:    m.LastRefresh,
	}

	return json.Marshal(data)
}

// Deserialize loads the component from JSON bytes.
func (m *ModBrowserComponent) Deserialize(data []byte) error {
	var browserData ModBrowserData
	if err := json.Unmarshal(data, &browserData); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.InstalledMods = make(map[string]bool)
	for _, modID := range browserData.InstalledMods {
		m.InstalledMods[modID] = true
	}

	m.SearchQuery = browserData.SearchQuery
	m.SortBy = browserData.SortBy
	m.SortDescending = browserData.SortDescending
	m.ActiveCategory = browserData.ActiveCategory
	m.LastRefresh = browserData.LastRefresh

	return nil
}
