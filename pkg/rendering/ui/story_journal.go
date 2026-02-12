package ui

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"

	"github.com/opd-ai/venture/pkg/engine"
)

// StoryJournalUI represents a story journal interface for viewing discovered fragments.
type StoryJournalUI struct {
	X, Y          int
	Width, Height int
	GenreID       string

	// Navigation state
	SelectedSeriesIndex   int
	SelectedFragmentIndex int
	ViewMode              JournalViewMode // 0=series list, 1=fragment list for series, 2=fragment detail

	// Content
	SeriesList       []SeriesEntry
	VisibleFragments []FragmentEntry

	// Colors
	BackgroundColor color.Color
	TextColor       color.Color
	HighlightColor  color.Color
	CompleteColor   color.Color
	IncompleteColor color.Color
}

// JournalViewMode represents the current view in the journal.
type JournalViewMode int

const (
	ViewSeriesList JournalViewMode = iota
	ViewFragmentList
	ViewFragmentDetail
)

// SeriesEntry represents a story series in the journal.
type SeriesEntry struct {
	SeriesID        string
	SeriesName      string
	FragmentCount   int
	DiscoveredCount int
	IsComplete      bool
}

// FragmentEntry represents a single fragment in the journal.
type FragmentEntry struct {
	SeriesID      string
	SequenceNum   int
	FragmentType  string
	Content       string
	SpritePattern string
	IsDiscovered  bool
}

// NewStoryJournalUI creates a new story journal UI.
func NewStoryJournalUI(x, y, width, height int, genreID string) *StoryJournalUI {
	// Determine colors based on genre
	bgColor := color.RGBA{20, 15, 10, 230}            // Dark parchment
	textColor := color.RGBA{220, 210, 190, 255}       // Light text
	highlightColor := color.RGBA{255, 200, 100, 255}  // Gold highlight
	completeColor := color.RGBA{100, 255, 100, 255}   // Green for complete
	incompleteColor := color.RGBA{150, 150, 150, 255} // Gray for incomplete

	if genreID == "scifi" {
		bgColor = color.RGBA{10, 15, 25, 230}
		textColor = color.RGBA{100, 200, 255, 255}
		highlightColor = color.RGBA{0, 255, 255, 255}
	} else if genreID == "horror" {
		bgColor = color.RGBA{15, 5, 5, 240}
		textColor = color.RGBA{200, 180, 180, 255}
		highlightColor = color.RGBA{200, 50, 50, 255}
	}

	return &StoryJournalUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		GenreID:         genreID,
		ViewMode:        ViewSeriesList,
		BackgroundColor: bgColor,
		TextColor:       textColor,
		HighlightColor:  highlightColor,
		CompleteColor:   completeColor,
		IncompleteColor: incompleteColor,
	}
}

// LoadFromJournal populates the UI from a StoryJournalComponent and world entities.
func (j *StoryJournalUI) LoadFromJournal(journal *engine.StoryJournalComponent, world *engine.World) {
	j.SeriesList = []SeriesEntry{}

	seriesMap := j.buildSeriesMap(journal, world)
	j.markCompletedSeries(journal, seriesMap)
	j.convertAndSortSeries(seriesMap)
}

// buildSeriesMap creates a map of series from world fragments.
func (j *StoryJournalUI) buildSeriesMap(journal *engine.StoryJournalComponent, world *engine.World) map[string]*SeriesEntry {
	seriesMap := make(map[string]*SeriesEntry)
	fragments := world.GetEntitiesWith("storyfragment")

	for _, fragEntity := range fragments {
		storyFrag := j.extractStoryFragment(fragEntity)
		if storyFrag == nil {
			continue
		}

		j.updateSeriesEntry(seriesMap, storyFrag, journal)
	}

	return seriesMap
}

// extractStoryFragment extracts and validates story fragment component.
func (j *StoryJournalUI) extractStoryFragment(fragEntity *engine.Entity) *engine.StoryFragmentComponent {
	fragComp, ok := fragEntity.GetComponent("storyfragment")
	if !ok {
		return nil
	}

	storyFrag, ok := fragComp.(*engine.StoryFragmentComponent)
	if !ok {
		return nil
	}

	return storyFrag
}

// updateSeriesEntry creates or updates series entry with fragment data.
func (j *StoryJournalUI) updateSeriesEntry(seriesMap map[string]*SeriesEntry, storyFrag *engine.StoryFragmentComponent, journal *engine.StoryJournalComponent) {
	series := j.getOrCreateSeries(seriesMap, storyFrag.SeriesID)
	series.FragmentCount++

	if journal.IsDiscovered(storyFrag.SeriesID, storyFrag.SequenceNum) {
		series.DiscoveredCount++
	}
}

// getOrCreateSeries retrieves existing or creates new series entry.
func (j *StoryJournalUI) getOrCreateSeries(seriesMap map[string]*SeriesEntry, seriesID string) *SeriesEntry {
	series, exists := seriesMap[seriesID]
	if !exists {
		series = &SeriesEntry{
			SeriesID:        seriesID,
			SeriesName:      formatSeriesName(seriesID),
			FragmentCount:   0,
			DiscoveredCount: 0,
			IsComplete:      false,
		}
		seriesMap[seriesID] = series
	}
	return series
}

// markCompletedSeries marks series that are completed in the journal.
func (j *StoryJournalUI) markCompletedSeries(journal *engine.StoryJournalComponent, seriesMap map[string]*SeriesEntry) {
	for seriesID, series := range seriesMap {
		if journal.CompletedSeries[seriesID] {
			series.IsComplete = true
		}
	}
}

// convertAndSortSeries converts map to sorted slice.
func (j *StoryJournalUI) convertAndSortSeries(seriesMap map[string]*SeriesEntry) {
	for _, series := range seriesMap {
		j.SeriesList = append(j.SeriesList, *series)
	}

	sort.Slice(j.SeriesList, func(i, k int) bool {
		if j.SeriesList[i].IsComplete != j.SeriesList[k].IsComplete {
			return j.SeriesList[i].IsComplete
		}
		return j.SeriesList[i].SeriesName < j.SeriesList[k].SeriesName
	})
}

// LoadFragmentsForSeries loads fragments for the currently selected series.
func (j *StoryJournalUI) LoadFragmentsForSeries(journal *engine.StoryJournalComponent, world *engine.World) {
	if j.SelectedSeriesIndex >= len(j.SeriesList) {
		return
	}

	seriesID := j.SeriesList[j.SelectedSeriesIndex].SeriesID
	j.VisibleFragments = []FragmentEntry{}

	// Get all fragments for this series
	fragments := world.GetEntitiesWith("storyfragment")

	for _, fragEntity := range fragments {
		fragComp, ok := fragEntity.GetComponent("storyfragment")
		if !ok {
			continue
		}

		storyFrag, ok := fragComp.(*engine.StoryFragmentComponent)
		if !ok {
			continue
		}

		if storyFrag.SeriesID != seriesID {
			continue
		}

		// Check if discovered using journal's IsDiscovered method
		isDiscovered := journal.IsDiscovered(seriesID, storyFrag.SequenceNum)

		entry := FragmentEntry{
			SeriesID:      seriesID,
			SequenceNum:   storyFrag.SequenceNum,
			FragmentType:  storyFrag.Fragment.Type.String(),
			Content:       storyFrag.Fragment.Content,
			SpritePattern: storyFrag.Fragment.SpritePattern,
			IsDiscovered:  isDiscovered,
		}

		j.VisibleFragments = append(j.VisibleFragments, entry)
	}

	// Sort by sequence number
	sort.Slice(j.VisibleFragments, func(i, k int) bool {
		return j.VisibleFragments[i].SequenceNum < j.VisibleFragments[k].SequenceNum
	})
}

// NavigateUp moves selection up in the current view.
func (j *StoryJournalUI) NavigateUp() {
	switch j.ViewMode {
	case ViewSeriesList:
		if j.SelectedSeriesIndex > 0 {
			j.SelectedSeriesIndex--
		}
	case ViewFragmentList:
		if j.SelectedFragmentIndex > 0 {
			j.SelectedFragmentIndex--
		}
	}
}

// NavigateDown moves selection down in the current view.
func (j *StoryJournalUI) NavigateDown() {
	switch j.ViewMode {
	case ViewSeriesList:
		if j.SelectedSeriesIndex < len(j.SeriesList)-1 {
			j.SelectedSeriesIndex++
		}
	case ViewFragmentList:
		if j.SelectedFragmentIndex < len(j.VisibleFragments)-1 {
			j.SelectedFragmentIndex++
		}
	}
}

// NavigateSelect confirms selection (Enter key).
func (j *StoryJournalUI) NavigateSelect(journal *engine.StoryJournalComponent, world *engine.World) {
	switch j.ViewMode {
	case ViewSeriesList:
		// Open fragment list for selected series
		j.LoadFragmentsForSeries(journal, world)
		j.ViewMode = ViewFragmentList
		j.SelectedFragmentIndex = 0
	case ViewFragmentList:
		// Open fragment detail
		if j.SelectedFragmentIndex < len(j.VisibleFragments) {
			frag := j.VisibleFragments[j.SelectedFragmentIndex]
			if frag.IsDiscovered {
				j.ViewMode = ViewFragmentDetail
			}
		}
	}
}

// NavigateBack goes back to previous view (Escape key).
func (j *StoryJournalUI) NavigateBack() {
	switch j.ViewMode {
	case ViewFragmentList:
		j.ViewMode = ViewSeriesList
	case ViewFragmentDetail:
		j.ViewMode = ViewFragmentList
	}
}

// Render renders the journal UI to an image.
// Returns an image that can be drawn to the screen.
func (j *StoryJournalUI) Render() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, j.Width, j.Height))

	// Fill background
	j.fillBackground(img)

	// Render based on view mode
	switch j.ViewMode {
	case ViewSeriesList:
		j.renderSeriesList(img)
	case ViewFragmentList:
		j.renderFragmentList(img)
	case ViewFragmentDetail:
		j.renderFragmentDetail(img)
	}

	return img
}

// fillBackground fills the background with a solid color.
func (j *StoryJournalUI) fillBackground(img *image.RGBA) {
	bgCol := j.BackgroundColor.(color.RGBA)
	for y := 0; y < j.Height; y++ {
		for x := 0; x < j.Width; x++ {
			img.Set(x, y, bgCol)
		}
	}
}

// renderSeriesList renders the list of story series.
func (j *StoryJournalUI) renderSeriesList(img *image.RGBA) {
	// Title
	j.drawText(img, 20, 20, "STORY JOURNAL", j.HighlightColor)

	// Series list
	startY := 60
	lineHeight := 30

	for i, series := range j.SeriesList {
		y := startY + i*lineHeight

		// Skip if off screen
		if y > j.Height-40 {
			break
		}

		// Determine color
		col := j.TextColor
		if i == j.SelectedSeriesIndex {
			col = j.HighlightColor
		} else if series.IsComplete {
			col = j.CompleteColor
		} else {
			col = j.IncompleteColor
		}

		// Format text: "Series Name (2/5)"
		text := fmt.Sprintf("%s (%d/%d)", series.SeriesName, series.DiscoveredCount, series.FragmentCount)
		if series.IsComplete {
			text += " ✓"
		}

		j.drawText(img, 40, y, text, col)
	}

	// Instructions
	j.drawText(img, 20, j.Height-30, "↑↓: Navigate  Enter: Select  Esc: Close", j.TextColor)
}

// renderFragmentList renders the list of fragments in a series.
func (j *StoryJournalUI) renderFragmentList(img *image.RGBA) {
	if j.SelectedSeriesIndex >= len(j.SeriesList) {
		return
	}

	series := j.SeriesList[j.SelectedSeriesIndex]

	// Title
	title := fmt.Sprintf("%s - Fragments", series.SeriesName)
	j.drawText(img, 20, 20, title, j.HighlightColor)

	// Fragment list
	startY := 60
	lineHeight := 25

	for i, frag := range j.VisibleFragments {
		y := startY + i*lineHeight

		if y > j.Height-60 {
			break
		}

		// Determine color
		col := j.TextColor
		if i == j.SelectedFragmentIndex {
			col = j.HighlightColor
		} else if !frag.IsDiscovered {
			col = j.IncompleteColor
		}

		// Format text
		text := fmt.Sprintf("%d. ", frag.SequenceNum)
		if frag.IsDiscovered {
			text += fmt.Sprintf("[%s] %s...", frag.FragmentType, truncateText(frag.Content, 40))
		} else {
			text += "??? (Not discovered)"
		}

		j.drawText(img, 40, y, text, col)
	}

	// Instructions
	j.drawText(img, 20, j.Height-30, "↑↓: Navigate  Enter: Read  Esc: Back", j.TextColor)
}

// renderFragmentDetail renders the full content of a fragment.
func (j *StoryJournalUI) renderFragmentDetail(img *image.RGBA) {
	if j.SelectedFragmentIndex >= len(j.VisibleFragments) {
		return
	}

	frag := j.VisibleFragments[j.SelectedFragmentIndex]

	// Title
	title := fmt.Sprintf("%s - Fragment %d", frag.FragmentType, frag.SequenceNum)
	j.drawText(img, 20, 20, title, j.HighlightColor)

	// Content (word-wrapped)
	startY := 60
	lineHeight := 20
	maxCharsPerLine := 60

	lines := wrapText(frag.Content, maxCharsPerLine)

	for i, line := range lines {
		y := startY + i*lineHeight
		if y > j.Height-60 {
			break
		}
		j.drawText(img, 30, y, line, j.TextColor)
	}

	// Instructions
	j.drawText(img, 20, j.Height-30, "Esc: Back", j.TextColor)
}

// drawText draws text at the specified position (simple implementation).
// This is a placeholder - real implementation would use proper font rendering.
func (j *StoryJournalUI) drawText(img *image.RGBA, x, y int, text string, col color.Color) {
	// This is a simple placeholder that just draws a colored rectangle
	// In a real implementation, this would use a font rendering library
	// For now, we'll just mark the area to indicate text would be here
	textCol := col.(color.RGBA)
	for i := 0; i < len(text) && x+i*6 < j.Width; i++ {
		for dy := 0; dy < 8; dy++ {
			for dx := 0; dx < 5; dx++ {
				if x+i*6+dx < j.Width && y+dy < j.Height && y+dy >= 0 {
					img.Set(x+i*6+dx, y+dy, textCol)
				}
			}
		}
	}
}

// formatSeriesName formats a series ID into a readable name.
func formatSeriesName(seriesID string) string {
	// Remove underscores and capitalize words
	parts := strings.Split(seriesID, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// truncateText truncates text to a maximum length with ellipsis.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// wrapText wraps text into lines of maximum character width.
func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	lines := []string{}
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
