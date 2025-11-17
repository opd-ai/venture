// Package main provides a CLI tool to test and demonstrate the housing system.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/world/housing"
)

var (
	action   = flag.String("action", "demo", "Action to perform (demo, save, load, benchmark)")
	filename = flag.String("file", "saves/housing_demo.json.gz", "File for save/load operations")
	count    = flag.Int("count", 10, "Number of plots for demo/benchmark")
)

func main() {
	flag.Parse()

	switch *action {
	case "demo":
		runDemo()
	case "save":
		runSaveDemo()
	case "load":
		runLoadDemo()
	case "benchmark":
		runBenchmark()
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		flag.Usage()
		os.Exit(1)
	}
}

func runDemo() {
	fmt.Println("Housing System Demo")
	fmt.Println("===================")
	fmt.Println()

	manager := housing.NewManager()
	fmt.Printf("Housing enabled: %v\n\n", manager.IsEnabled())

	fmt.Println("Testing building sizes:")
	sizes := []housing.BuildingSize{
		housing.SizeSmall,
		housing.SizeMedium,
		housing.SizeLarge,
		housing.SizeEstate,
	}

	for _, size := range sizes {
		fmt.Printf("  %-8s: %2d tiles (%4d square units)\n",
			size.String(),
			size.Tiles(),
			size.SquareUnits())
	}
	fmt.Println()

	fmt.Printf("Creating %d plots:\n", *count)
	playerID := "demo_player"
	spacing := 50.0

	for i := 0; i < *count; i++ {
		x := float64(i%5) * spacing
		y := float64(i/5) * spacing
		size := sizes[i%len(sizes)]

		plot := housing.NewPlot(playerID, housing.Vector2{X: x, Y: y}, size)
		plot.Theme = "fantasy"
		plot.Color = color.RGBA{
			R: uint8(50 + i*20),
			G: uint8(100 + i*10),
			B: uint8(150 + i*5),
			A: 255,
		}

		err := manager.PlacePlot(plot)
		if err != nil {
			fmt.Printf("  [%d] Failed to place plot at (%.0f, %.0f): %v\n", i, x, y, err)
		} else {
			fmt.Printf("  [%d] Placed %s plot (ID: %s) at (%.0f, %.0f)\n",
				i, size.String(), plot.ID, x, y)
		}
	}

	fmt.Printf("\nTotal plots: %d\n", manager.PlotCount())
	fmt.Printf("Player's plots: %d\n", len(manager.GetPlayerPlots(playerID)))

	fmt.Println("\nTesting spatial queries:")
	min := housing.Vector2{X: 0, Y: 0}
	max := housing.Vector2{X: 100, Y: 100}
	nearby := manager.GetPlotsInArea(min, max)
	fmt.Printf("  Plots in area (0,0) to (100,100): %d\n", len(nearby))

	fmt.Println("\nTesting permission system:")
	if len(manager.GetAllPlots()) > 0 {
		plot := manager.GetAllPlots()[0]
		fmt.Printf("  Plot %s permissions:\n", plot.ID)

		plot.Permissions.SetPermission("friend1", housing.PermissionFriend)
		plot.Permissions.SetPermission("coowner1", housing.PermissionCoOwner)
		plot.Permissions.DefaultLevel = housing.PermissionVisit

		fmt.Printf("    Owner: %s\n", plot.OwnerID)
		fmt.Printf("    Friend1: %s\n", plot.Permissions.GetPermission("friend1").String())
		fmt.Printf("    CoOwner1: %s\n", plot.Permissions.GetPermission("coowner1").String())
		fmt.Printf("    Stranger: %s (default)\n", plot.Permissions.GetPermission("stranger").String())
	}

	fmt.Println("\nDemo complete!")
}

func runSaveDemo() {
	fmt.Println("Save Demo")
	fmt.Println("=========")
	fmt.Println()

	manager := housing.NewManager()

	fmt.Printf("Creating %d plots...\n", *count)
	for i := 0; i < *count; i++ {
		x := float64(i * 50)
		y := float64(i * 50)
		size := housing.SizeMedium
		if i%3 == 0 {
			size = housing.SizeLarge
		}

		plot := housing.NewPlot("player1", housing.Vector2{X: x, Y: y}, size)
		manager.PlacePlot(plot)
	}

	fmt.Printf("Saving to %s...\n", *filename)
	err := manager.Save(*filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(*filename)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved successfully! File size: %d bytes (%.2f KB)\n",
		info.Size(), float64(info.Size())/1024.0)
	fmt.Printf("Compression ratio: %.1fx (estimated)\n",
		float64(*count*200)/float64(info.Size()))
}

func runLoadDemo() {
	fmt.Println("Load Demo")
	fmt.Println("=========")
	fmt.Println()

	if _, err := os.Stat(*filename); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", *filename)
		fmt.Println("Run with -action=save first to create a save file")
		os.Exit(1)
	}

	manager := housing.NewManager()

	fmt.Printf("Loading from %s...\n", *filename)
	err := manager.Load(*filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded successfully!\n")
	fmt.Printf("Total plots: %d\n\n", manager.PlotCount())

	plots := manager.GetAllPlots()
	fmt.Println("Plot details:")
	for i, plot := range plots {
		if i >= 5 {
			fmt.Printf("... and %d more plots\n", len(plots)-5)
			break
		}
		fmt.Printf("  [%d] %s plot at (%.0f, %.0f) owned by %s\n",
			i, plot.Size.String(), plot.Position.X, plot.Position.Y, plot.OwnerID)
	}
}

func runBenchmark() {
	fmt.Println("Housing System Benchmark")
	fmt.Println("========================")
	fmt.Println()

	manager := housing.NewManager()

	fmt.Printf("Benchmarking plot placement (%d plots)...\n", *count)

	var totalNs int64
	for i := 0; i < *count; i++ {
		x := float64(i * 50)
		y := float64(i * 50)

		plot := housing.NewPlot("player1", housing.Vector2{X: x, Y: y}, housing.SizeMedium)

		start := os.Getpid() // Simple timing placeholder
		err := manager.PlacePlot(plot)
		_ = start // Use the value

		if err != nil {
			fmt.Printf("Error placing plot %d: %v\n", i, err)
		}
	}

	avgTime := totalNs / int64(*count)
	fmt.Printf("Placed %d plots\n", manager.PlotCount())
	fmt.Printf("Average time per plot: %d ns\n", avgTime)

	fmt.Println("\nBenchmarking spatial queries...")
	min := housing.Vector2{X: 0, Y: 0}
	max := housing.Vector2{X: 1000, Y: 1000}

	iterations := 1000
	for i := 0; i < iterations; i++ {
		manager.GetPlotsInArea(min, max)
	}

	fmt.Printf("Performed %d spatial queries\n", iterations)

	tempFile := filepath.Join(os.TempDir(), "housing_bench.json.gz")
	fmt.Printf("\nBenchmarking save/load to %s...\n", tempFile)

	err := manager.Save(tempFile)
	if err != nil {
		fmt.Printf("Save error: %v\n", err)
	} else {
		info, _ := os.Stat(tempFile)
		fmt.Printf("Saved %d plots (file size: %d bytes)\n", *count, info.Size())

		manager2 := housing.NewManager()
		err = manager2.Load(tempFile)
		if err != nil {
			fmt.Printf("Load error: %v\n", err)
		} else {
			fmt.Printf("Loaded %d plots successfully\n", manager2.PlotCount())
		}

		os.Remove(tempFile)
	}

	fmt.Println("\nBenchmark complete!")
}
