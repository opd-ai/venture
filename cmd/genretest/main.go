package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/opd-ai/venture/pkg/procgen/genre"
)

var (
	listAll  = flag.Bool("list", false, "List all available genres")
	genreID  = flag.String("genre", "", "Show details for a specific genre")
	showAll  = flag.Bool("all", false, "Show detailed information for all genres")
	validate = flag.String("validate", "", "Validate a genre ID")
)

func main() {
	flag.Parse()

	registry := genre.DefaultRegistry()

	// List all genres
	if *listAll {
		listGenres(registry)
		return
	}

	// Show details for a specific genre
	if *genreID != "" {
		showGenreDetails(registry, *genreID)
		return
	}

	// Show all genres with details
	if *showAll {
		showAllGenres(registry)
		return
	}

	// Validate a genre ID
	if *validate != "" {
		validateGenre(registry, *validate)
		return
	}

	// Default: show usage
	flag.Usage()
	fmt.Println("\nExamples:")
	fmt.Println("  genretest -list                    # List all genres")
	fmt.Println("  genretest -genre fantasy           # Show fantasy genre details")
	fmt.Println("  genretest -all                     # Show all genre details")
	fmt.Println("  genretest -validate fantasy        # Validate a genre ID")
}

func listGenres(registry *genre.Registry) {
	genres := registry.All()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTHEMES")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	for _, g := range genres {
		themes := strings.Join(g.Themes, ", ")
		if len(themes) > 50 {
			themes = themes[:47] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", g.ID, g.Name, themes)
	}

	w.Flush()
	fmt.Printf("\nTotal genres: %d\n", registry.Count())
}

func showGenreDetails(registry *genre.Registry, id string) {
	g, err := registry.Get(id)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Genre Details")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("ID:             %s\n", g.ID)
	fmt.Printf("Name:           %s\n", g.Name)
	fmt.Printf("Description:    %s\n", g.Description)
	fmt.Println()
	fmt.Printf("Themes:         %s\n", strings.Join(g.Themes, ", "))
	fmt.Println()
	fmt.Println("Color Palette:")
	fmt.Printf("  Primary:      %s\n", g.PrimaryColor)
	fmt.Printf("  Secondary:    %s\n", g.SecondaryColor)
	fmt.Printf("  Accent:       %s\n", g.AccentColor)
	fmt.Println()
	fmt.Println("Name Prefixes:")
	fmt.Printf("  Entity:       %s\n", g.EntityPrefix)
	fmt.Printf("  Item:         %s\n", g.ItemPrefix)
	fmt.Printf("  Location:     %s\n", g.LocationPrefix)
}

func showAllGenres(registry *genre.Registry) {
	genres := registry.All()

	for i, g := range genres {
		if i > 0 {
			fmt.Println()
		}
		showGenreDetailsInline(g)
	}

	fmt.Printf("\nTotal genres: %d\n", len(genres))
}

func showGenreDetailsInline(g *genre.Genre) {
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%s (%s)\n", g.Name, g.ID)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Description: %s\n", g.Description)
	fmt.Printf("Themes:      %s\n", strings.Join(g.Themes, ", "))
	fmt.Printf("Colors:      %s, %s, %s\n", g.PrimaryColor, g.SecondaryColor, g.AccentColor)
	fmt.Printf("Prefixes:    Entity='%s', Item='%s', Location='%s'\n",
		g.EntityPrefix, g.ItemPrefix, g.LocationPrefix)
}

func validateGenre(registry *genre.Registry, id string) {
	if registry.Has(id) {
		fmt.Printf("✓ Genre '%s' is valid\n", id)
		g, _ := registry.Get(id)
		fmt.Printf("  Name: %s\n", g.Name)
	} else {
		fmt.Printf("✗ Genre '%s' is not found\n", id)
		fmt.Println("\nAvailable genres:")
		for _, id := range registry.IDs() {
			fmt.Printf("  - %s\n", id)
		}
		os.Exit(1)
	}
}
