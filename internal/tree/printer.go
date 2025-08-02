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

func shouldSkip(name string) bool {
	skipDirs := []string{
		".git", ".github", ".vscode", ".idea", "node_modules",
		".DS_Store", "Thumbs.db", ".env", ".env.local",
		"vendor", "target", "build", "dist", "__pycache__",
		".pytest_cache", ".coverage", "coverage", ".nyc_output",
	}

	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}

	// Skip hidden files that start with . (except .gitignore, .gitkeep, etc.)
	if strings.HasPrefix(name, ".") {
		allowedDotFiles := []string{
			".gitignore", ".gitkeep", ".dockerignore", ".editorconfig",
			".eslintrc", ".prettierrc", ".babelrc", ".npmrc",
		}
		for _, allowed := range allowedDotFiles {
			if strings.HasPrefix(name, allowed) {
				return false
			}
		}
		return true
	}

	return false
}
