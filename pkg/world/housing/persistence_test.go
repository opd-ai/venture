package housing

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type closeErrorWriteCloser struct {
	io.Writer
	closeErr error
}

func (c *closeErrorWriteCloser) Close() error {
	return c.closeErr
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "housing_test.json.gz")

	// Create manager with plots
	m1 := NewManager()
	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)

	m1.PlacePlot(plot1)
	m1.PlacePlot(plot2)

	// Save
	err := m1.Save(filename)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}

	// Load into new manager
	m2 := NewManager()
	err = m2.Load(filename)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify plots loaded
	if m2.PlotCount() != 2 {
		t.Errorf("PlotCount() after load = %v, want 2", m2.PlotCount())
	}

	// Verify plot data
	loadedPlot1, ok := m2.GetPlot(plot1.ID)
	if !ok {
		t.Error("Plot1 not found after load")
	} else {
		if loadedPlot1.OwnerID != plot1.OwnerID {
			t.Errorf("Loaded plot OwnerID = %v, want %v", loadedPlot1.OwnerID, plot1.OwnerID)
		}
		if loadedPlot1.Position != plot1.Position {
			t.Errorf("Loaded plot Position = %v, want %v", loadedPlot1.Position, plot1.Position)
		}
		if loadedPlot1.Size != plot1.Size {
			t.Errorf("Loaded plot Size = %v, want %v", loadedPlot1.Size, plot1.Size)
		}
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	m := NewManager()
	err := m.Load("nonexistent_file.json.gz")
	if err == nil {
		t.Error("Load() with nonexistent file should return error")
	}
}

func TestSaveInvalidPath(t *testing.T) {
	m := NewManager()
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	m.PlacePlot(plot)

	// Try to save to invalid path (root directory on Unix)
	err := m.Save("/invalid/path/housing.json.gz")
	if err == nil {
		t.Error("Save() with invalid path should return error")
	}
}

func TestSavePlayerData(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "player1_housing.json.gz")

	m := NewManager()
	plot1 := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot2 := NewPlot("player1", Vector2{X: 200, Y: 200}, SizeSmall)
	plot3 := NewPlot("player2", Vector2{X: 300, Y: 300}, SizeLarge)

	m.PlacePlot(plot1)
	m.PlacePlot(plot2)
	m.PlacePlot(plot3)

	// Save only player1's data
	err := m.SavePlayerData("player1", filename)
	if err != nil {
		t.Fatalf("SavePlayerData() error = %v, want nil", err)
	}

	// Load into new manager
	m2 := NewManager()
	err = m2.Load(filename)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Should only have player1's plots
	if m2.PlotCount() != 2 {
		t.Errorf("PlotCount() after load = %v, want 2", m2.PlotCount())
	}

	// Verify player2's plot is not loaded
	_, ok := m2.GetPlot(plot3.ID)
	if ok {
		t.Error("Player2's plot should not be loaded from player1's save")
	}
}

func TestSaveEmptyManager(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "empty_housing.json.gz")

	m := NewManager()

	// Save empty manager
	err := m.Save(filename)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Load into new manager
	m2 := NewManager()
	err = m2.Load(filename)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if m2.PlotCount() != 0 {
		t.Errorf("PlotCount() after loading empty save = %v, want 0", m2.PlotCount())
	}
}

func TestSavePreservesPermissions(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "permissions_test.json.gz")

	m1 := NewManager()
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	plot.Permissions.SetPermission("player2", PermissionFriend)
	plot.Permissions.DefaultLevel = PermissionVisit

	m1.PlacePlot(plot)
	m1.Save(filename)

	// Load and verify permissions preserved
	m2 := NewManager()
	m2.Load(filename)

	loadedPlot, ok := m2.GetPlot(plot.ID)
	if !ok {
		t.Fatal("Plot not found after load")
	}

	if loadedPlot.Permissions.GetPermission("player2") != PermissionFriend {
		t.Error("Player2 permission not preserved")
	}

	if loadedPlot.Permissions.DefaultLevel != PermissionVisit {
		t.Error("Default permission level not preserved")
	}
}

func BenchmarkSave(b *testing.B) {
	tempDir := b.TempDir()

	// Create manager with plots
	m := NewManager()
	for i := 0; i < 100; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		m.PlacePlot(plot)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, "bench_save.json.gz")
		m.Save(filename)
	}
}

func BenchmarkLoad(b *testing.B) {
	tempDir := b.TempDir()
	filename := filepath.Join(tempDir, "bench_load.json.gz")

	// Create and save manager with plots
	m1 := NewManager()
	for i := 0; i < 100; i++ {
		plot := NewPlot("player1", Vector2{X: float64(i * 50), Y: float64(i * 50)}, SizeMedium)
		m1.PlacePlot(plot)
	}
	m1.Save(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m2 := NewManager()
		m2.Load(filename)
	}
}

// TestSaveCloseErrors verifies that close errors are logged properly.
// This is a smoke test to ensure the close error handling code paths are reachable.
func TestSaveCloseErrors(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "close_errors_test.json.gz")

	m := NewManager()
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	m.PlacePlot(plot)

	origCreateSaveFile := createSaveFile
	createSaveFile = func(string) (io.WriteCloser, error) {
		return &closeErrorWriteCloser{
			Writer:   io.Discard,
			closeErr: errors.New("injected close error"),
		}, nil
	}
	defer func() { createSaveFile = origCreateSaveFile }()

	err := m.Save(filename)
	if err == nil {
		t.Fatal("Save() error = nil, want close error")
	}
	if !strings.Contains(err.Error(), "failed to close file") {
		t.Fatalf("Save() error = %v, want close error context", err)
	}
}

// TestSavePlayerDataCloseErrors verifies close error handling for SavePlayerData.
func TestSavePlayerDataCloseErrors(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "player_close_errors_test.json.gz")

	m := NewManager()
	plot := NewPlot("player1", Vector2{X: 100, Y: 100}, SizeMedium)
	m.PlacePlot(plot)

	origCreateSaveFile := createSaveFile
	createSaveFile = func(string) (io.WriteCloser, error) {
		return &closeErrorWriteCloser{
			Writer:   io.Discard,
			closeErr: errors.New("injected close error"),
		}, nil
	}
	defer func() { createSaveFile = origCreateSaveFile }()

	err := m.SavePlayerData("player1", filename)
	if err == nil {
		t.Fatal("SavePlayerData() error = nil, want close error")
	}
	if !strings.Contains(err.Error(), "failed to close file") {
		t.Fatalf("SavePlayerData() error = %v, want close error context", err)
	}
}
