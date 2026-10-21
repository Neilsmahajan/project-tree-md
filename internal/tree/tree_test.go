package tree

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// makeTree creates files (and their parent directories) under a new temp
// directory. Names ending in "/" are created as empty directories.
func makeTree(t *testing.T, paths ...string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "my-app")
	for _, p := range append([]string{""}, paths...) {
		full := filepath.Join(root, filepath.FromSlash(p))
		if p == "" || strings.HasSuffix(p, "/") {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func render(t *testing.T, root string, opts Options, fence bool) string {
	t.Helper()
	node, problems, err := Build(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, node, fence); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestMarkdown(t *testing.T) {
	root := makeTree(t,
		"src/app.go", "src/util/strings.go", "README.md", "go.mod",
		"docs/", ".gitignore", "Makefile",
	)
	got := render(t, root, Options{}, true)
	want := "```\n" + `my-app
├── docs/
├── src/
│   ├── util/
│   │   └── strings.go
│   └── app.go
├── .gitignore
├── go.mod
├── Makefile
└── README.md
` + "```\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestNoFence(t *testing.T) {
	root := makeTree(t, "a.txt")
	got := render(t, root, Options{}, false)
	want := "my-app\n└── a.txt\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSortIsStableAcrossCase(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("needs a case-sensitive file system")
	}
	root := makeTree(t, "b", "a", "B", "A")
	got := render(t, root, Options{}, false)
	want := "my-app\n├── A\n├── a\n├── B\n└── b\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultSkips(t *testing.T) {
	root := makeTree(t,
		".git/HEAD", "node_modules/x/index.js", ".github/workflows/ci.yml",
		".env", ".env.production", ".DS_Store", ".secret",
		".eslintrc.cjs", ".prettierrc", ".gitignore", "build/out", "keep.txt",
	)
	got := render(t, root, Options{}, false)
	want := "my-app\n├── .eslintrc.cjs\n├── .gitignore\n├── .prettierrc\n└── keep.txt\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAllAndExclude(t *testing.T) {
	root := makeTree(t, ".github/workflows/ci.yml", "node_modules/x.js", "app.log", "main.go")
	got := render(t, root, Options{All: true, Exclude: []string{"node_modules", "*.log"}}, false)
	want := "my-app\n├── .github/\n│   └── workflows/\n│       └── ci.yml\n└── main.go\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDepth(t *testing.T) {
	root := makeTree(t, "a/b/c/deep.txt", "a/top.txt", "z.txt")
	got := render(t, root, Options{MaxDepth: 2}, false)
	want := "my-app\n├── a/\n│   ├── b/\n│   └── top.txt\n└── z.txt\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	node, _, err := Build(root, Options{MaxDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	b := node.Children[0].Children[0]
	if b.Name != "b" || !b.Truncated || b.Children != nil {
		t.Errorf("b = %+v, want truncated with no children", b)
	}
}

func TestSymlink(t *testing.T) {
	root := makeTree(t, "real/file.txt")
	if err := os.Symlink("real", filepath.Join(root, "link")); err != nil {
		t.Skipf("can't create symlinks here: %v", err)
	}
	got := render(t, root, Options{}, false)
	want := "my-app\n├── real/\n│   └── file.txt\n└── link -> real\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestControlCharactersAreQuoted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't allow these names")
	}
	root := makeTree(t)
	if err := os.WriteFile(filepath.Join(root, "two\nlines"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got := render(t, root, Options{}, false)
	want := "my-app\n└── \"two\\nlines\"\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUnreadableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions work differently on Windows")
	}
	root := makeTree(t, "locked/secret.txt", "open/file.txt")
	locked := filepath.Join(root, "locked")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })
	if _, err := os.ReadDir(locked); err == nil {
		t.Skip("running with permission to read everything (root?)")
	}

	node, problems, err := Build(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || problems[0].Path != "locked" {
		t.Fatalf("problems = %v, want one for locked", problems)
	}
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, node, false); err != nil {
		t.Fatal(err)
	}
	want := "my-app\n├── locked/ (unreadable)\n└── open/\n    └── file.txt\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestBuildErrors(t *testing.T) {
	root := makeTree(t, "file.txt")
	if _, _, err := Build(filepath.Join(root, "missing"), Options{}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("missing dir: err = %v", err)
	}
	if _, _, err := Build(filepath.Join(root, "file.txt"), Options{}); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("file: err = %v", err)
	}
	if _, _, err := Build(root, Options{Exclude: []string{"["}}); err == nil {
		t.Error("bad pattern: want an error")
	}
}

func TestJSON(t *testing.T) {
	root := makeTree(t, "src/main.go", "docs/", "README.md")
	if err := os.Symlink("README.md", filepath.Join(root, "link")); err != nil {
		t.Logf("no symlink: %v", err)
	}
	node, problems, err := Build(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, NewDocument(root, node, problems)); err != nil {
		t.Fatal(err)
	}

	var doc struct {
		SchemaVersion int    `json:"schemaVersion"`
		Path          string `json:"path"`
		Root          struct {
			Name     string            `json:"name"`
			Type     string            `json:"type"`
			Children []json.RawMessage `json:"children"`
		} `json:"root"`
		Counts   map[string]int `json:"counts"`
		Problems []any          `json:"problems"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if doc.SchemaVersion != 1 || doc.Path != root || doc.Root.Name != "my-app" || doc.Root.Type != "directory" {
		t.Errorf("unexpected header: %+v", doc)
	}
	if doc.Problems == nil || len(doc.Problems) != 0 {
		t.Errorf("problems = %v, want an empty array", doc.Problems)
	}
	if doc.Counts["directories"] != 2 {
		t.Errorf("counts = %v, want 2 directories", doc.Counts)
	}
	// An empty directory has "children": [], a file has no children key, and
	// a symlink has its target.
	byName := map[string]map[string]json.RawMessage{}
	for _, raw := range doc.Root.Children {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		var name string
		json.Unmarshal(fields["name"], &name)
		byName[name] = fields
	}
	if got := string(byName["docs"]["children"]); got != "[]" {
		t.Errorf(`docs children = %q, want "[]"`, got)
	}
	if _, ok := byName["README.md"]["children"]; ok {
		t.Error("a file has a children key")
	}
	if link, ok := byName["link"]; ok {
		if string(link["type"]) != `"symlink"` || string(link["target"]) != `"README.md"` {
			t.Errorf("link = %v", link)
		}
	}
}

func TestDotFileExceptionsAreExact(t *testing.T) {
	for name, skip := range map[string]bool{
		".gitignore": false, ".npmrc": false, ".eslintrc.cjs": false, ".prettierrc.yaml": false,
		".babelrc.json": false, ".npmrc.backup": true, ".gitignore-private": true,
		".editorconfig-secrets": true, ".env.production": true,
	} {
		if got := DefaultSkip(name); got != skip {
			t.Errorf("DefaultSkip(%q) = %v, want %v", name, got, skip)
		}
	}
}

func TestFenceGrowsForBackticks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't allow backticks in some file systems")
	}
	root := makeTree(t, "```")
	got := render(t, root, Options{}, true)
	want := "````\nmy-app\n└── ```\n````\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUnreadableSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions work differently on Windows")
	}
	root := makeTree(t, "dir/")
	dir := filepath.Join(root, "dir")
	if err := os.Symlink("somewhere", filepath.Join(dir, "link")); err != nil {
		t.Skipf("can't create symlinks here: %v", err)
	}
	// Readable but not searchable: the names can be listed, the link can't be read.
	if err := os.Chmod(dir, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	if _, err := os.Readlink(filepath.Join(dir, "link")); err == nil {
		t.Skip("running with permission to read everything (root?)")
	}
	if _, err := os.ReadDir(dir); err != nil {
		t.Skipf("can't list a directory without search permission here: %v", err)
	}

	node, problems, err := Build(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || problems[0].Path != "dir/link" {
		t.Fatalf("problems = %v, want one for dir/link", problems)
	}
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, node, false); err != nil {
		t.Fatal(err)
	}
	if want := "my-app\n└── dir/\n    └── link -> (unreadable)\n"; buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
	link := toJSON(node.Children[0].Children[0])
	if link.Type != KindSymlink || link.Target != "" || link.Error == "" {
		t.Errorf("JSON node = %+v, want a symlink with an error and no target", link)
	}
}

func TestSymlinkRootIsFollowed(t *testing.T) {
	root := makeTree(t, "src/main.go")
	link := filepath.Join(filepath.Dir(root), "shortcut")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("can't create symlinks here: %v", err)
	}
	got := render(t, link, Options{}, false)
	want := "shortcut\n└── src/\n    └── main.go\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
