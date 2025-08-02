package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neilsmahajan/project-tree-md/internal/tree"
)

func Run() error {
	var (
		pathFlag = flag.String("path", ".", "Path to the directory to print")
		helpFlag = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate a markdown-formatted directory tree structure.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -path ./my-project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s  # uses current directory\n", os.Args[0])
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return nil
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(*pathFlag)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Check if path is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("error accessing path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	if err := tree.PrintMarkdownTree(absPath); err != nil {
		return fmt.Errorf("error printing tree: %w", err)
	}

	return nil
}
