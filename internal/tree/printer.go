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
		fmt.Printf("%s- %s\n", indent, entry.Name())
		if entry.IsDir() {
			if err := PrintTree(filepath.Join(path, entry.Name()), indent+"  "); err != nil {
				return fmt.Errorf("error reading directory %s: %w", filepath.Join(path, entry.Name()), err)
			}
		}
	}
	return nil
}
