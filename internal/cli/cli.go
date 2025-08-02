package cli

import (
	"flag"
	"fmt"

	"github.com/neilsmahajan/project-tree-md/internal/tree"
)

func Run() error {
	// Placeholder for CLI logic
	fmt.Println("Running CLI...")
	path := flag.String("path", ".", "Path to the directory to print")
	flag.Parse()
	fmt.Println("# Project Structure")
	if err := tree.PrintTree(*path, ""); err != nil {
		return fmt.Errorf("error printing tree: %w", err)
	}
	return nil
}
