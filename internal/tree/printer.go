package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PrintMarkdownTree prints the directory structure in markdown format
func PrintMarkdownTree(rootPath string) error {
	fmt.Println("```")
	rootName := filepath.Base(rootPath)
	fmt.Println(rootName)
	return printTree(rootPath, "", true)
}

func printTree(currentPath, prefix string, isRoot bool) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	// Filter out unwanted files and directories
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		if shouldSkip(entry.Name()) {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(filteredEntries, func(i, j int) bool {
		if filteredEntries[i].IsDir() && !filteredEntries[j].IsDir() {
			return true
		}
		if !filteredEntries[i].IsDir() && filteredEntries[j].IsDir() {
			return false
		}
		return strings.ToLower(filteredEntries[i].Name()) < strings.ToLower(filteredEntries[j].Name())
	})

	for i, entry := range filteredEntries {
		isLast := i == len(filteredEntries)-1

		if entry.IsDir() {
			// Print directory
			if isLast {
				fmt.Printf("%s└── %s/\n", prefix, entry.Name())
				newPrefix := prefix + "    "
				subPath := filepath.Join(currentPath, entry.Name())
				if err := printTree(subPath, newPrefix, false); err != nil {
					return err
				}
			} else {
				fmt.Printf("%s├── %s/\n", prefix, entry.Name())
				newPrefix := prefix + "│   "
				subPath := filepath.Join(currentPath, entry.Name())
				if err := printTree(subPath, newPrefix, false); err != nil {
					return err
				}
			}
		} else {
			// Print file
			if isLast {
				fmt.Printf("%s└── %s\n", prefix, entry.Name())
			} else {
				fmt.Printf("%s├── %s\n", prefix, entry.Name())
			}
		}
	}

	// Close the markdown code block only at the root level
	if isRoot {
		fmt.Println("```")
	}

	return nil
}

// Initialize skip sets as package-level variables for better performance
var (
	skipSet = map[string]bool{
		".git":          true,
		".github":       true,
		".vscode":       true,
		".idea":         true,
		"node_modules":  true,
		".DS_Store":     true,
		"Thumbs.db":     true,
		".env":          true,
		".env.local":    true,
		"vendor":        true,
		"target":        true,
		"build":         true,
		"dist":          true,
		"__pycache__":   true,
		".pytest_cache": true,
		".coverage":     true,
		"coverage":      true,
		".nyc_output":   true,
	}

	allowedDotFilesSet = map[string]bool{
		".gitignore":       true,
		".gitkeep":         true,
		".dockerignore":    true,
		".editorconfig":    true,
		".eslintrc":        true,
		".eslintrc.js":     true,
		".eslintrc.json":   true,
		".prettierrc":      true,
		".prettierrc.js":   true,
		".prettierrc.json": true,
		".babelrc":         true,
		".babelrc.js":      true,
		".babelrc.json":    true,
		".npmrc":           true,
	}
)

func shouldSkip(name string) bool {
	// O(1) lookup for exact matches
	if skipSet[name] {
		return true
	}

	// Skip hidden files that start with . (except allowed ones)
	if strings.HasPrefix(name, ".") {
		// Check if it's an allowed dot file (exact match or prefix match)
		if allowedDotFilesSet[name] {
			return false
		}

		// Check for prefix matches for config files with extensions
		for allowed := range allowedDotFilesSet {
			if strings.HasPrefix(name, allowed) {
				return false
			}
		}

		return true
	}

	return false
}
