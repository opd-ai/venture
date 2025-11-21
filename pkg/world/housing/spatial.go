package housing

// SpatialGrid provides efficient spatial queries for plot placement.
// Uses a uniform grid with cell-based indexing.
type SpatialGrid struct {
	cellSize float64
	cells    map[int]map[int][]*Plot // [x][y] -> plots
}

// NewSpatialGrid creates a new spatial grid with the given cell size.
func NewSpatialGrid(cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize: cellSize,
		cells:    make(map[int]map[int][]*Plot),
	}
}

// Insert adds a plot to the spatial grid.
func (sg *SpatialGrid) Insert(plot *Plot) {
	min, max := plot.Bounds()

	// Find all cells that this plot touches
	minCellX, minCellY := sg.getCell(min)
	maxCellX, maxCellY := sg.getCell(max)

	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			sg.insertIntoCell(x, y, plot)
		}
	}
}

// Remove removes a plot from the spatial grid.
func (sg *SpatialGrid) Remove(plot *Plot) {
	min, max := plot.Bounds()

	minCellX, minCellY := sg.getCell(min)
	maxCellX, maxCellY := sg.getCell(max)

	for x := minCellX; x <= maxCellX; x++ {
		for y := minCellY; y <= maxCellY; y++ {
			sg.removeFromCell(x, y, plot)
		}
	}
}

// Query returns all plots that potentially intersect the given area.
func (sg *SpatialGrid) Query(min, max Vector2) []*Plot {
	minCellX, minCellY := sg.getCell(min)
	maxCellX, maxCellY := sg.getCell(max)

	seen := make(map[string]bool)
	var results []*Plot

	for x := minCellX; x <= maxCellX; x++ {
		if col, ok := sg.cells[x]; ok {
			for y := minCellY; y <= maxCellY; y++ {
				if plots, ok := col[y]; ok {
					for _, plot := range plots {
						if !seen[plot.ID] {
							seen[plot.ID] = true
							results = append(results, plot)
						}
					}
				}
			}
		}
	}

	return results
}

// getCell returns the cell coordinates for a world position.
func (sg *SpatialGrid) getCell(pos Vector2) (int, int) {
	x := int(pos.X / sg.cellSize)
	y := int(pos.Y / sg.cellSize)
	if pos.X < 0 {
		x--
	}
	if pos.Y < 0 {
		y--
	}
	return x, y
}

// insertIntoCell adds a plot to a specific cell.
func (sg *SpatialGrid) insertIntoCell(x, y int, plot *Plot) {
	if sg.cells[x] == nil {
		sg.cells[x] = make(map[int][]*Plot)
	}
	sg.cells[x][y] = append(sg.cells[x][y], plot)
}

// removeFromCell removes a plot from a specific cell.
func (sg *SpatialGrid) removeFromCell(x, y int, plot *Plot) {
	if col, ok := sg.cells[x]; ok {
		if plots, ok := col[y]; ok {
			for i, p := range plots {
				if p.ID == plot.ID {
					sg.cells[x][y] = append(plots[:i], plots[i+1:]...)
					return
				}
			}
		}
	}
}

// Update removes a plot from its old position and re-inserts it.
// This is more efficient than Remove followed by Insert when the plot may have moved.
func (sg *SpatialGrid) Update(plot *Plot) {
	sg.Remove(plot)
	sg.Insert(plot)
}
