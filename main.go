// Command project-tree-md draws a directory as a Markdown tree, ready to
// paste into a README.
//
//	go install github.com/neilsmahajan/project-tree-md@latest
package main

import (
	"os"

	"github.com/neilsmahajan/project-tree-md/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
