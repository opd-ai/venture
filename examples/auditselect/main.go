// Package main implements an automated package selection tool for code audits.
// It analyzes Go package dependencies to select the next package requiring audit
// based on dependency depth (packages with fewer dependencies are prioritized).
package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageInfo holds metadata about a package for audit selection
type PackageInfo struct {
	Path            string   // Relative path from repo root
	Name            string   // Package name
	HasAudit        bool     // Whether AUDIT.md exists
	Dependencies    []string // Internal package dependencies
	DependencyCount int      // Count of internal dependencies (depth metric)
}

func main() {
	repoRoot := flag.String("repo", ".", "Repository root directory")
	pkgDir := flag.String("pkg", "pkg", "Package directory to scan")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	packages, err := findPackages(filepath.Join(*repoRoot, *pkgDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding packages: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Found %d packages\n", len(packages))
	}

	analyzeDependencies(*repoRoot, packages, *verbose)
	unaudited := filterUnauditedPackages(packages)

	if len(unaudited) == 0 {
		fmt.Println("All packages have been audited!")
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "%d packages need auditing\n", len(unaudited))
	}

	sortPackagesByPriority(unaudited)
	selected := unaudited[0]
	outputSelection(selected, *verbose)
}

// analyzeDependencies analyzes dependencies for all packages in place.
func analyzeDependencies(repoRoot string, packages []PackageInfo, verbose bool) {
	for i := range packages {
		deps, err := analyzePackageDependencies(repoRoot, packages[i].Path)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to analyze %s: %v\n", packages[i].Path, err)
			}
			continue
		}
		packages[i].Dependencies = deps
		packages[i].DependencyCount = len(deps)
	}
}

// filterUnauditedPackages returns packages that don't have audit files.
func filterUnauditedPackages(packages []PackageInfo) []PackageInfo {
	unaudited := make([]PackageInfo, 0)
	for _, pkg := range packages {
		if !pkg.HasAudit {
			unaudited = append(unaudited, pkg)
		}
	}
	return unaudited
}

// sortPackagesByPriority sorts packages by dependency depth then path priority.
func sortPackagesByPriority(packages []PackageInfo) {
	sort.Slice(packages, func(i, j int) bool {
		if packages[i].DependencyCount != packages[j].DependencyCount {
			return packages[i].DependencyCount < packages[j].DependencyCount
		}
		return getPriority(packages[i].Path) < getPriority(packages[j].Path)
	})
}

// outputSelection prints the selected package and optional details.
func outputSelection(selected PackageInfo, verbose bool) {
	fmt.Printf("%s\n", selected.Path)
	if verbose {
		fmt.Fprintf(os.Stderr, "Selected: %s (depth: %d dependencies)\n", selected.Path, selected.DependencyCount)
		if len(selected.Dependencies) > 0 {
			fmt.Fprintf(os.Stderr, "Dependencies: %s\n", strings.Join(selected.Dependencies, ", "))
		}
	}
}

// findPackages recursively finds all Go packages in the given directory
func findPackages(root string) ([]PackageInfo, error) {
	var packages []PackageInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and vendor
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
		}

		// Check if this directory contains Go files
		if d.IsDir() {
			hasGo, err := hasGoFiles(path)
			if err != nil {
				return err
			}

			if hasGo {
				// Check for AUDIT.md
				auditPath := filepath.Join(path, "AUDIT.md")
				hasAudit := false
				if _, err := os.Stat(auditPath); err == nil {
					hasAudit = true
				}

				// Get package name
				pkgName, err := getPackageName(path)
				if err != nil {
					pkgName = filepath.Base(path)
				}

				packages = append(packages, PackageInfo{
					Path:     path,
					Name:     pkgName,
					HasAudit: hasAudit,
				})
			}
		}

		return nil
	})

	return packages, err
}

// hasGoFiles checks if a directory contains any .go files (excluding tests)
func hasGoFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true, nil
		}
	}

	return false, nil
}

// getPackageName extracts the package name from Go files in a directory
func getPackageName(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				continue
			}
			return file.Name.Name, nil
		}
	}

	return "", fmt.Errorf("no package name found")
}

// analyzePackageDependencies finds internal package dependencies
func analyzePackageDependencies(repoRoot, pkgPath string) ([]string, error) {
	deps := make(map[string]bool)

	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(pkgPath, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Only count internal dependencies (venture packages)
			if strings.HasPrefix(importPath, "github.com/opd-ai/venture/pkg/") {
				// Extract package path relative to repo
				relPath := strings.TrimPrefix(importPath, "github.com/opd-ai/venture/")
				deps[relPath] = true
			}
		}
	}

	// Convert to sorted slice
	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}
	sort.Strings(result)

	return result, nil
}

// getPriority returns a priority value for package paths
// Lower values = higher priority
func getPriority(path string) int {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Extract package path relative to pkg/
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "pkg" && i+1 < len(parts) {
			path = strings.Join(parts[i+1:], "/")
			break
		}
	}

	// Priority order: engine > procgen > rendering > others
	if path == "engine" || strings.HasPrefix(path, "engine/") {
		return 1
	}
	if path == "procgen" || strings.HasPrefix(path, "procgen/") {
		return 2
	}
	if path == "rendering" || strings.HasPrefix(path, "rendering/") {
		return 3
	}

	// All others have lower priority
	return 4
}
