package tree

import (
	"fmt"
	"os"
	"path/filepath"
)

func PrintTree(path, indent string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".github", ".vscode":
				// Skip these directories
				continue
			}

			fmt.Println(indent + entry.Name() + "/")
			subPath := filepath.Join(path, entry.Name())
			if err := PrintTree(subPath, indent+"  "); err != nil {
				return err
			}
		} else {
			switch entry.Name() {
			case ".DS_Store":
				// Skip .DS_Store files
				continue
			}
			fmt.Println(indent + entry.Name())
		}
	}
	return nil
}
