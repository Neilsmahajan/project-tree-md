package main

import "github.com/neilsmahajan/project-tree-md/internal/cli"

func main() {
	if err := cli.Run(); err != nil {
		panic(err)
	}
}
