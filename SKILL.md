---
name: project-tree-md
description: Draw a directory's structure as a Markdown tree (box-drawing characters in a ``` block) with the project-tree-md CLI, leaving out .git, node_modules, build output, .env and most dot files. Use when writing or updating a README, design doc or PR description that should show a project's layout, or when you need a filtered, structured listing of a directory as JSON.
---

# project-tree-md

`project-tree-md` prints a directory as a Markdown tree. It is read-only and
never prompts.

## Use

For a tree to put in a document, use the Markdown output and paste it as is:

```sh
project-tree-md path/to/project                 # wrapped in ``` lines
project-tree-md --no-fence path/to/project      # without them
project-tree-md --depth 2 path/to/project       # big repos: limit the depth
project-tree-md --exclude '*.snap' --exclude fixtures path/to/project
```

To read the structure yourself, use JSON:

```sh
project-tree-md --json path/to/project
```

Top-level keys: `schemaVersion` (1), `path` (absolute), `root`, `counts`
(`directories`, `files`, not counting the root) and `problems`. Each node has
`name` and `type` (`directory`, `file`, `symlink`, `other`); directories have
`children`, except when `truncated: true` (cut off by `--depth`, not read) or
when they have an `error`. Symlinks have a `target` (or an `error` if the link couldn't be read) and
are not followed (a symlink passed as the directory itself is).

Exit codes: `0` success, `1` the directory doesn't exist or can't be read, or
a folder or symlink inside it couldn't be read (the tree is still printed; see
`problems` or stderr), `2` bad flags. If the command isn't installed, say so;
it installs with `go install github.com/neilsmahajan/project-tree-md@latest`.

## Things to know

- By default it hides `.git`, `.github`, `.vscode`, `.idea`, `node_modules`,
  `vendor`, `target`, `build`, `dist`, `coverage`, `__pycache__`, `.env`,
  `.env.local` and dot files other than `.gitignore`, `.dockerignore`,
  `.editorconfig`, `.npmrc`, `.gitkeep` and `.eslintrc*`/`.prettierrc*`/`.babelrc*`.
  So the tree is not a complete listing. Pass `--all` (usually with
  `--exclude .git --exclude node_modules`) when hidden folders such as
  `.github/workflows` matter.
- `--exclude` matches a glob against each name, not the path.
- On a large repository, start with `--depth 2` or `3` rather than the full
  tree; long trees make poor documentation.

## Don't

- Don't paste trees of private or client projects into public places: folder
  and file names can reveal what a project is.
- Don't hand-edit a pasted tree to add files that don't exist; regenerate it.
