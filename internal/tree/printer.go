package tree

import (
	"fmt"
	"os"
	"path/filepath"
)

func PrintTree(currentPath, indent string) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".github", ".vscode":
				continue
			}
			fmt.Println(indent + entry.Name() + "/")
			subPath := filepath.Join(currentPath, entry.Name())
			if err := PrintTree(subPath, indent+"-"); err != nil {
				return err
			}
		} else {
			switch entry.Name() {
			case ".DS_Store":
				continue
			}
			fmt.Println(indent + entry.Name())
		}
	}
	return nil
}
