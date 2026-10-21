// This copy of the command keeps the original install path working
// (go install github.com/neilsmahajan/project-tree-md/cmd@latest installs it
// as "cmd"). Prefer the module root:
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
