// Package cli implements the project-tree-md command line.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/neilsmahajan/project-tree-md/internal/tree"
)

// Exit codes.
const (
	ExitOK    = 0 // the tree was printed
	ExitError = 1 // the path can't be read, or some directory below it couldn't be (the tree is still printed)
	ExitUsage = 2 // bad flags or arguments
)

const name = "project-tree-md"

// version is set at build time with
// -ldflags "-X github.com/neilsmahajan/project-tree-md/internal/cli.version=v0.1.0".
// Without it, the module version recorded by `go install` is used.
var version string

// Version returns the version of this build.
func Version() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

const usage = `project-tree-md draws a directory as a Markdown tree, ready to paste into a README.

Usage:
  project-tree-md [options] [directory]

The directory defaults to the current one.

Options:
  --json             print the tree as JSON (schema version 1) instead of Markdown
  --depth N          list at most N levels below the directory (0, the default, means no limit)
  --exclude PATTERN  leave out names matching PATTERN, a glob such as '*.log' (repeatable)
  --all              don't apply the built-in skip list or hide dot files
  --no-fence         leave out the surrounding ` + "```" + ` lines
  --path DIR         the directory, as a flag (same as the argument)
  --version          print the version
  -h, --help         show this help

Skipped by default: .git, .github, .vscode, .idea, node_modules, vendor, target,
build, dist, coverage, __pycache__, .pytest_cache, .coverage, .nyc_output,
.DS_Store, Thumbs.db, .env, .env.local, and other dot files except .gitignore,
.gitkeep, .dockerignore, .editorconfig, .npmrc and .eslintrc*, .prettierrc*,
.babelrc*. Symlinks inside the tree are listed but not followed; if the
directory you give is itself a symlink, it is followed.

Output: the tree goes to stdout; errors and warnings go to stderr. It never
prompts, so it behaves the same in a pipe or a script.

Exit codes:
  0  the tree was printed
  1  the directory doesn't exist or can't be read, or a directory or symlink
     inside it couldn't be read (the rest of the tree is still printed)
  2  bad flags or arguments

Examples:
  project-tree-md                       # the current directory
  project-tree-md ./my-project
  project-tree-md --depth 2 --exclude '*.log' .
  project-tree-md --json . | jq '.counts'
`

// valueFlags take a value, which may be a separate argument ("--depth 2").
var valueFlags = map[string]bool{"path": true, "depth": true, "exclude": true}

// splitAtTerminator splits args at the first "--" that ends the options
// (not one given as a flag's value, as in "--exclude --").
func splitAtTerminator(args []string) (flags, after []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return args[:i], args[i+1:]
		}
		if !strings.HasPrefix(a, "-") || a == "-" || strings.Contains(a, "=") {
			continue
		}
		if valueFlags[strings.TrimLeft(a, "-")] {
			i++ // skip the value
		}
	}
	return args, nil
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

// Main runs the command with args (without the program name) and returns the
// exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	var (
		pathFlag    = fs.String("path", "", "")
		jsonFlag    = fs.Bool("json", false, "")
		depthFlag   = fs.Int("depth", 0, "")
		allFlag     = fs.Bool("all", false, "")
		noFenceFlag = fs.Bool("no-fence", false, "")
		versionFlag = fs.Bool("version", false, "")
		excludes    multiFlag
	)
	fs.Var(&excludes, "exclude", "")

	usageError := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "%s: %s\nRun '%s --help' for usage.\n", name, fmt.Sprintf(format, a...), name)
		return ExitUsage
	}

	// The flag package stops at the first argument that isn't a flag, so
	// keep parsing after it: `project-tree-md dir --json` works too. Anything
	// after a "--" that isn't a flag's value is an argument.
	flagArgs, afterDashes := splitAtTerminator(args)
	var positional []string
	rest := flagArgs
	for {
		if err := fs.Parse(rest); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				fmt.Fprint(stdout, usage)
				return ExitOK
			}
			return usageError("%v", err)
		}
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		rest = rest[1:]
	}
	positional = append(positional, afterDashes...)

	if *versionFlag {
		fmt.Fprintf(stdout, "%s %s\n", name, Version())
		return ExitOK
	}

	dir := "."
	switch {
	case len(positional) > 1:
		return usageError("expected one directory, got %d: %s", len(positional), strings.Join(positional, " "))
	case len(positional) == 1 && *pathFlag != "":
		return usageError("give the directory either as an argument or with --path, not both")
	case len(positional) == 1:
		dir = positional[0]
	case *pathFlag != "":
		dir = *pathFlag
	}
	if *depthFlag < 0 {
		return usageError("--depth must be 0 or more, got %d", *depthFlag)
	}
	if *jsonFlag && *noFenceFlag {
		return usageError("--no-fence only applies to Markdown output, not --json")
	}
	for _, p := range excludes {
		if err := tree.ValidatePattern(p); err != nil {
			return usageError("%v", err)
		}
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(stderr, "%s: resolving %s: %v\n", name, dir, err)
		return ExitError
	}
	root, problems, err := tree.Build(absPath, tree.Options{
		MaxDepth: *depthFlag,
		Exclude:  excludes,
		All:      *allFlag,
	})
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return ExitError
	}

	if *jsonFlag {
		err = tree.WriteJSON(stdout, tree.NewDocument(absPath, root, problems))
	} else {
		err = tree.WriteMarkdown(stdout, root, !*noFenceFlag)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: writing output: %v\n", name, err)
		return ExitError
	}

	for _, p := range problems {
		fmt.Fprintf(stderr, "%s: warning: couldn't read %s: %v\n", name, p.Path, p.Err)
	}
	if len(problems) > 0 {
		return ExitError
	}
	return ExitOK
}
