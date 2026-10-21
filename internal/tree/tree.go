// Package tree reads a directory into a filtered tree and renders it as
// Markdown or JSON.
package tree

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Kind is the type of an entry in the tree.
type Kind string

const (
	KindDirectory Kind = "directory"
	KindFile      Kind = "file"
	KindSymlink   Kind = "symlink"
	KindOther     Kind = "other" // sockets, devices, named pipes
)

// Node is one entry in the tree.
type Node struct {
	Name   string
	Kind   Kind
	Target string // symlink target, as stored in the link
	// Children is nil for anything that isn't a directory, and for directories
	// that were not read (Truncated) or could not be read (Err).
	Children  []*Node
	Truncated bool  // a directory whose contents were cut off by MaxDepth
	Err       error // a directory that could not be read, or a symlink whose target could not be
}

// Options control which entries are read.
type Options struct {
	// MaxDepth limits how many levels below the root are listed. 0 means no
	// limit; 1 lists only the root's own entries.
	MaxDepth int
	// Exclude holds glob patterns (path.Match syntax) matched against each
	// entry's name. Matching entries are left out.
	Exclude []string
	// All turns off the built-in skip list and the hidden-file rule. Exclude
	// still applies.
	All bool
}

// Problem is a directory below the root that could not be read, or a symlink
// whose target could not be. Path is
// relative to the root and uses forward slashes.
type Problem struct {
	Path string
	Err  error
}

// ValidatePattern reports whether pattern is a valid Exclude pattern.
func ValidatePattern(pattern string) error {
	if _, err := path.Match(pattern, ""); err != nil {
		return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
	}
	return nil
}

// Build reads the directory at root. It returns an error only when the root
// itself can't be used; directories below it that can't be read are marked
// on their node and listed in the returned problems.
func Build(root string, opts Options) (*Node, []Problem, error) {
	for _, p := range opts.Exclude {
		if err := ValidatePattern(p); err != nil {
			return nil, nil, err
		}
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("path does not exist: %s", root)
		}
		return nil, nil, err
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("path is not a directory: %s", root)
	}
	node := &Node{Name: rootName(root), Kind: KindDirectory}
	b := builder{opts: opts}
	if err := b.fill(node, root, "", 0); err != nil {
		return nil, nil, err
	}
	return node, b.problems, nil
}

func rootName(root string) string {
	name := filepath.Base(root)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return root
	}
	return name
}

type builder struct {
	opts     Options
	problems []Problem
}

// fill reads dir into node. depth is node's depth below the root (0 for the
// root). Errors reading the root are returned; errors below it are recorded.
func (b *builder) fill(node *Node, dir, rel string, depth int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if depth == 0 {
			return err
		}
		node.Err = err
		b.problems = append(b.problems, Problem{Path: rel, Err: err})
		return nil
	}
	node.Children = []*Node{}
	for _, entry := range entries {
		name := entry.Name()
		if b.skip(name) {
			continue
		}
		child := &Node{Name: name, Kind: kindOf(entry.Type())}
		childPath := filepath.Join(dir, name)
		childRel := path.Join(rel, name)
		switch child.Kind {
		case KindSymlink:
			target, err := os.Readlink(childPath)
			if err != nil {
				child.Err = err
				b.problems = append(b.problems, Problem{Path: childRel, Err: err})
			}
			child.Target = target
		case KindDirectory:
			if b.opts.MaxDepth > 0 && depth+1 >= b.opts.MaxDepth {
				child.Truncated = true
			} else if err := b.fill(child, childPath, childRel, depth+1); err != nil {
				return err
			}
		}
		node.Children = append(node.Children, child)
	}
	sortNodes(node.Children)
	return nil
}

func kindOf(mode os.FileMode) Kind {
	switch {
	case mode.IsDir():
		return KindDirectory
	case mode&os.ModeSymlink != 0:
		return KindSymlink
	case mode.IsRegular():
		return KindFile
	default:
		return KindOther
	}
}

// sortNodes puts directories first, then everything else, each group in
// case-insensitive name order (ties broken by exact name).
func sortNodes(nodes []*Node) {
	slices.SortFunc(nodes, func(a, b *Node) int {
		aDir, bDir := a.Kind == KindDirectory, b.Kind == KindDirectory
		if aDir != bDir {
			if aDir {
				return -1
			}
			return 1
		}
		return cmp.Or(
			strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			strings.Compare(a.Name, b.Name),
		)
	})
}

func (b *builder) skip(name string) bool {
	for _, p := range b.opts.Exclude {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	if b.opts.All {
		return false
	}
	return DefaultSkip(name)
}

// skipNames are left out unless Options.All is set.
var skipNames = map[string]bool{
	".git":          true,
	".github":       true,
	".vscode":       true,
	".idea":         true,
	"node_modules":  true,
	".DS_Store":     true,
	"Thumbs.db":     true,
	".env":          true,
	".env.local":    true,
	"vendor":        true,
	"target":        true,
	"build":         true,
	"dist":          true,
	"__pycache__":   true,
	".pytest_cache": true,
	".coverage":     true,
	"coverage":      true,
	".nyc_output":   true,
}

// allowedDotFiles are hidden files that are kept.
var allowedDotFiles = map[string]bool{
	".dockerignore": true,
	".editorconfig": true,
	".gitignore":    true,
	".gitkeep":      true,
	".npmrc":        true,
}

// allowedDotPrefixes keep config files that come in several forms
// (".eslintrc.cjs", ".prettierrc.yaml").
var allowedDotPrefixes = []string{".babelrc", ".eslintrc", ".prettierrc"}

// DefaultSkip reports whether name is left out by the built-in rules: the
// skip list, then any other hidden name that isn't an allowed dot file.
func DefaultSkip(name string) bool {
	if skipNames[name] {
		return true
	}
	if !strings.HasPrefix(name, ".") || allowedDotFiles[name] {
		return false
	}
	for _, prefix := range allowedDotPrefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return true
}

// Count returns the number of directories and non-directories below n,
// not counting n itself.
func Count(n *Node) (dirs, files int) {
	for _, c := range n.Children {
		if c.Kind == KindDirectory {
			dirs++
			d, f := Count(c)
			dirs += d
			files += f
		} else {
			files++
		}
	}
	return dirs, files
}
