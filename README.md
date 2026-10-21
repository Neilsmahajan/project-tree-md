# project-tree-md

Draw a project's folder structure as a Markdown tree, ready to paste into a
README, a design doc or a pull request. It leaves out the noise (`.git`,
`node_modules`, build output, `.env` files) so the tree shows the code.

![project-tree-md drawing the tree of a sample project](docs/demo.gif)

<sub>The project in the demo is made up (`docs/demo-project.sh` creates it).</sub>

```
acme-api
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── users.go
│   │   └── users_test.go
│   └── store/
│       └── postgres.go
├── web/
│   ├── src/
│   │   ├── App.tsx
│   │   └── main.tsx
│   └── package.json
├── .gitignore
├── Dockerfile
├── go.mod
└── README.md
```

## Install

With Go 1.24 or later:

```sh
go install github.com/neilsmahajan/project-tree-md@latest
```

The module path is lowercase; use it exactly as written. The older path,
`github.com/neilsmahajan/project-tree-md/cmd@latest`, still works but installs
the command as `cmd`.

Or build from a clone:

```sh
git clone https://github.com/Neilsmahajan/project-tree-md.git
cd project-tree-md
make build      # ./project-tree-md
make install    # copies it to /usr/local/bin
```

## Usage

```sh
project-tree-md                          # the current directory
project-tree-md ./my-project             # another directory
project-tree-md --depth 2                # only two levels deep
project-tree-md --exclude '*.log' --exclude testdata
project-tree-md --no-fence               # without the surrounding ``` lines
project-tree-md --all --exclude .git     # include .github, dot files and the rest
project-tree-md --json                   # the same tree as JSON
project-tree-md --help
```

The output is wrapped in a ```` ``` ```` code block, so you can pipe it straight
into a file or your clipboard:

```sh
project-tree-md | pbcopy                 # macOS
project-tree-md >> README.md
```

Directories come first, then files, each sorted by name ignoring case.
Directories end in `/`. Symlinks inside the tree are shown as `name -> target`
and are not followed (if the directory you pass is itself a symlink, it is). `--exclude` takes a glob that's matched against each file or folder
name, not the whole path.

## What gets left out

Unless you pass `--all`:

- Version control and editors: `.git`, `.github`, `.vscode`, `.idea`
- Dependencies: `node_modules`, `vendor`
- Build and test output: `target`, `build`, `dist`, `coverage`, `__pycache__`,
  `.pytest_cache`, `.coverage`, `.nyc_output`
- System files: `.DS_Store`, `Thumbs.db`
- Environment files: `.env`, `.env.local`
- Every other name starting with a dot, except `.gitignore`, `.gitkeep`,
  `.dockerignore`, `.editorconfig`, `.npmrc` and `.eslintrc*`, `.prettierrc*`,
  `.babelrc*`

## Works with coding agents

project-tree-md is also easy for scripts and coding agents to call, for
example to describe a repository's layout in docs they write:

- `--json` prints the tree in a [versioned schema](schema/project-tree-md.v1.schema.json)
  (`schemaVersion: 1`), with the same filtering as the Markdown, plus counts
  and any folders it couldn't read. Fields can be added within version 1;
  removing or changing one bumps it.
- The tree goes to stdout; errors and warnings go to stderr.
- Exit codes: `0` the tree was printed, `1` the directory doesn't exist or
  can't be read, or a folder or symlink inside it couldn't be read (the rest
  of the tree is still printed, and the entry is marked `(unreadable)`), `2`
  bad flags or arguments.
- It never prompts and has no interactive mode, so it behaves the same in a
  terminal, a pipe or CI.
- [`SKILL.md`](SKILL.md) tells an agent when and how to use it. Copy it to
  `~/.claude/skills/project-tree-md/SKILL.md` (Claude Code) or
  `~/.agents/skills/project-tree-md/SKILL.md` (Codex and others).

## Development

```sh
make test       # go test ./...
make vet
make lint       # staticcheck
make dev        # fmt, vet, lint, test and build
make demo       # re-records docs/demo.gif with vhs
```

CI runs the tests on Linux, macOS and Windows, plus `go vet`, staticcheck and
gitleaks. Issues and pull requests are welcome.

## License

[MIT](LICENSE)
