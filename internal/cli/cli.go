package cli

import (
	"flag"
	"fmt"
	"path"

	"github.com/neilsmahajan/project-tree-md/internal/tree"
)

func Run() error {
	currentPath := flag.String("path", ".", "Path to the directory to print")
	flag.Parse()
	lastPath := path.Base(*currentPath)
	fmt.Println(lastPath + "\\")
	if err := tree.PrintTree(*currentPath, "|-"); err != nil {
		return fmt.Errorf("error printing tree: %w", err)
	}
	return nil
}
