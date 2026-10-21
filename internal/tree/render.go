package tree

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// WriteMarkdown writes the tree with box-drawing characters. With fence set,
// it is wrapped in a code block, ready to paste into a README. The fence is
// ``` unless a name contains a run of backticks that long.
func WriteMarkdown(w io.Writer, root *Node, fence bool) error {
	var body strings.Builder
	body.WriteString(displayName(root.Name) + "\n")
	writeChildren(&body, root, "")
	out := body.String()
	if fence {
		f := strings.Repeat("`", max(3, longestBacktickRun(out)+1))
		out = f + "\n" + out + f + "\n"
	}
	_, err := io.WriteString(w, out)
	return err
}

func longestBacktickRun(s string) int {
	longest, run := 0, 0
	for _, r := range s {
		if r == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	return longest
}

func writeChildren(sb *strings.Builder, n *Node, prefix string) {
	for i, c := range n.Children {
		branch, indent := "├── ", "│   "
		if i == len(n.Children)-1 {
			branch, indent = "└── ", "    "
		}
		sb.WriteString(prefix + branch + label(c) + "\n")
		writeChildren(sb, c, prefix+indent)
	}
}

func label(n *Node) string {
	s := displayName(n.Name)
	switch n.Kind {
	case KindDirectory:
		s += "/"
		if n.Err != nil {
			s += " (unreadable)"
		}
	case KindSymlink:
		if n.Err != nil {
			s += " -> (unreadable)"
		} else {
			s += " -> " + displayName(n.Target)
		}
	}
	return s
}

// displayName quotes names that contain control characters (a newline in a
// file name would otherwise break the tree).
func displayName(name string) string {
	if strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return strconv.Quote(name)
	}
	return name
}

// SchemaVersion is the version of the JSON document written by WriteJSON.
// Fields may be added within a version; removing or changing one bumps it.
const SchemaVersion = 1

// Document is the JSON output. See schema/project-tree-md.v1.schema.json.
type Document struct {
	SchemaVersion int           `json:"schemaVersion"`
	Path          string        `json:"path"`
	Root          jsonNode      `json:"root"`
	Counts        jsonCounts    `json:"counts"`
	Problems      []jsonProblem `json:"problems"`
}

type jsonNode struct {
	Name      string      `json:"name"`
	Type      Kind        `json:"type"`
	Target    string      `json:"target,omitempty"`
	Children  *[]jsonNode `json:"children,omitempty"`
	Truncated bool        `json:"truncated,omitempty"`
	Error     string      `json:"error,omitempty"`
}

type jsonCounts struct {
	Directories int `json:"directories"`
	Files       int `json:"files"`
}

type jsonProblem struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// NewDocument builds the JSON document for a tree read from absPath.
func NewDocument(absPath string, root *Node, problems []Problem) Document {
	dirs, files := Count(root)
	doc := Document{
		SchemaVersion: SchemaVersion,
		Path:          absPath,
		Root:          toJSON(root),
		Counts:        jsonCounts{Directories: dirs, Files: files},
		Problems:      []jsonProblem{},
	}
	for _, p := range problems {
		doc.Problems = append(doc.Problems, jsonProblem{Path: p.Path, Message: p.Err.Error()})
	}
	return doc
}

func toJSON(n *Node) jsonNode {
	j := jsonNode{Name: n.Name, Type: n.Kind, Target: n.Target, Truncated: n.Truncated}
	if n.Err != nil {
		j.Error = n.Err.Error()
	}
	if n.Children != nil {
		children := make([]jsonNode, 0, len(n.Children))
		for _, c := range n.Children {
			children = append(children, toJSON(c))
		}
		j.Children = &children
	}
	return j
}

// WriteJSON writes doc as indented JSON followed by a newline.
func WriteJSON(w io.Writer, doc Document) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}
