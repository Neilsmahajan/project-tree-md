# Changelog

## 0.1.0

First tagged release.

- The directory can be given as an argument (`project-tree-md ./my-project`);
  `-path` still works.
- `--depth`, `--exclude` (repeatable glob), `--all` and `--no-fence`.
- `--json` with a versioned schema (`schema/project-tree-md.v1.schema.json`).
- Documented exit codes (0, 1, 2), errors and warnings on stderr, and
  `--version`. Errors no longer print a Go panic.
- A folder (or symlink) that can't be read is marked `(unreadable)` and the
  rest of the tree is still printed, with exit code 1, instead of stopping
  halfway with an unclosed code block.
- Only the documented dot files are kept: `.npmrc.backup` or
  `.gitignore-private` are now hidden like other dot files.
- Symlinks are shown as `name -> target`; names with control characters are
  quoted; the code fence grows if a name contains backticks; sorting is the
  same on every run.
- `go install github.com/neilsmahajan/project-tree-md@latest` installs the
  command as `project-tree-md`. The old `/cmd` path still works.
- Tests, CI (Linux, macOS, Windows; go vet, staticcheck, gitleaks), a demo GIF
  and `SKILL.md` for coding agents.
