package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func run(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = Main(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func sampleDir(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "sample")
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"src/main.go", "README.md"} {
		if err := os.WriteFile(filepath.Join(root, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const sampleTree = "sample\n├── src/\n│   └── main.go\n└── README.md\n"

func TestMarkdownOutput(t *testing.T) {
	dir := sampleDir(t)
	for _, args := range [][]string{
		{dir},
		{"--path", dir},
		{"-path", dir},
	} {
		code, out, errOut := run(args...)
		if code != ExitOK || errOut != "" {
			t.Errorf("%v: code %d, stderr %q", args, code, errOut)
		}
		if want := "```\n" + sampleTree + "```\n"; out != want {
			t.Errorf("%v: got %q, want %q", args, out, want)
		}
	}
}

func TestFlagsAfterDirectory(t *testing.T) {
	dir := sampleDir(t)
	code, out, _ := run(dir, "--no-fence")
	if code != ExitOK || out != sampleTree {
		t.Errorf("code %d, out %q", code, out)
	}
	code, out, _ = run("--no-fence", "--", dir)
	if code != ExitOK || out != sampleTree {
		t.Errorf("after --: code %d, out %q", code, out)
	}
	// "--" as a flag's value doesn't end the options.
	code, out, errOut := run("--exclude", "--", dir, "--no-fence")
	if code != ExitOK || out != sampleTree {
		t.Errorf("--exclude --: code %d, out %q, stderr %q", code, out, errOut)
	}
	code, _, _ = run("--", dir, "--no-fence")
	if code != ExitUsage {
		t.Errorf("flag after --: code %d, want %d", code, ExitUsage)
	}
}

func TestDefaultsToCurrentDirectory(t *testing.T) {
	t.Chdir(sampleDir(t))
	code, out, _ := run("--no-fence")
	if code != ExitOK || out != sampleTree {
		t.Errorf("code %d, out %q", code, out)
	}
}

func TestJSONOutput(t *testing.T) {
	dir := sampleDir(t)
	code, out, errOut := run("--json", dir)
	if code != ExitOK || errOut != "" {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	var doc struct {
		SchemaVersion int            `json:"schemaVersion"`
		Path          string         `json:"path"`
		Counts        map[string]int `json:"counts"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if doc.SchemaVersion != 1 || doc.Path != dir || doc.Counts["directories"] != 1 || doc.Counts["files"] != 2 {
		t.Errorf("unexpected document: %+v", doc)
	}
}

func TestHelpGoesToStdout(t *testing.T) {
	for _, flag := range []string{"-h", "--help", "-help"} {
		code, out, errOut := run(flag)
		if code != ExitOK || errOut != "" || !strings.Contains(out, "Exit codes:") {
			t.Errorf("%s: code %d, stderr %q, stdout %q", flag, code, errOut, out)
		}
	}
}

func TestVersion(t *testing.T) {
	code, out, _ := run("--version")
	if code != ExitOK || !strings.HasPrefix(out, "project-tree-md ") {
		t.Errorf("code %d, out %q", code, out)
	}
}

func TestUsageErrors(t *testing.T) {
	dir := sampleDir(t)
	for _, args := range [][]string{
		{"--nope"},
		{dir, dir},
		{"--path", dir, dir},
		{"--depth", "-1", dir},
		{"--depth", "x", dir},
		{"--exclude", "[", dir},
		{"--json", "--no-fence", dir},
	} {
		code, out, errOut := run(args...)
		if code != ExitUsage || out != "" || !strings.Contains(errOut, "--help") {
			t.Errorf("%v: code %d, stdout %q, stderr %q", args, code, out, errOut)
		}
	}
}

func TestRuntimeErrors(t *testing.T) {
	dir := sampleDir(t)
	for _, args := range [][]string{
		{filepath.Join(dir, "missing")},
		{filepath.Join(dir, "README.md")},
	} {
		code, out, errOut := run(args...)
		if code != ExitError || out != "" || errOut == "" {
			t.Errorf("%v: code %d, stdout %q, stderr %q", args, code, out, errOut)
		}
	}
}

func TestPartialTreeExitsOne(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions work differently on Windows")
	}
	dir := sampleDir(t)
	locked := filepath.Join(dir, "src")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })
	if _, err := os.ReadDir(locked); err == nil {
		t.Skip("running with permission to read everything (root?)")
	}

	code, out, errOut := run("--no-fence", dir)
	if code != ExitError {
		t.Errorf("code %d, want %d", code, ExitError)
	}
	if want := "sample\n├── src/ (unreadable)\n└── README.md\n"; out != want {
		t.Errorf("got %q, want %q", out, want)
	}
	if !strings.Contains(errOut, "warning: couldn't read src") {
		t.Errorf("stderr %q", errOut)
	}

	code, out, _ = run("--json", dir)
	if code != ExitError || !strings.Contains(out, `"problems": [`) || !strings.Contains(out, `"path": "src"`) {
		t.Errorf("json: code %d, out %s", code, out)
	}
}

// The schema file in the repo must describe the version the code writes.
func TestSchemaFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "project-tree-md.v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			SchemaVersion struct {
				Const int `json:"const"`
			} `json:"schemaVersion"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema isn't valid JSON: %v", err)
	}
	if schema.Properties.SchemaVersion.Const != 1 {
		t.Errorf("schema const = %d, want 1", schema.Properties.SchemaVersion.Const)
	}
}
