# Project Tree Markdown

`project-tree-md` is a simple CLI tool that generates a Markdown-formatted tree structure of any directory. Perfect for creating README documentation that shows your project structure.

## Features

- 🌳 **Clean markdown output** - Generates properly formatted tree structures using box-drawing characters
- 📁 **Smart filtering** - Automatically skips common unwanted files and directories (.git, node_modules, .DS_Store, etc.)
- 🔧 **Easy to use** - Simple command-line interface with path flag
- 📋 **Copy-paste ready** - Output is wrapped in markdown code blocks, ready for your README

## Installation

### Build from source (using Make)

```bash
git clone https://github.com/neilsmahajan/project-tree-md.git
cd project-tree-md
make build
```

### Build from source (using Go directly)

```bash
git clone https://github.com/neilsmahajan/project-tree-md.git
cd project-tree-md
go build -o project-tree-md cmd/main.go
```

### Run directly with Go

```bash
go run cmd/main.go -path /path/to/your/project
```

### Install system-wide (optional)

```bash
make install  # Installs to /usr/local/bin
```

## Usage

### Basic usage

```bash
# Generate tree for current directory
./project-tree-md

# Generate tree for specific directory
./project-tree-md -path /path/to/your/project

# Show help
./project-tree-md -help
```

### Make commands

```bash
# Build the project
make build

# Build and run on current directory
make run

# Clean build artifacts
make clean

# Run development tasks (format, vet, test, build)
make dev

# Build for multiple platforms
make build-all

# Show all available make targets
make help
```

### Example output

Running `./project-tree-md` on this project generates:

```
project-tree-md
├── cmd/
│   └── main.go
├── internal/
│   ├── cli/
│   │   └── cli.go
│   └── tree/
│       └── printer.go
├── .gitignore
├── go.mod
├── LICENSE
├── Makefile
└── README.md
```

## What gets filtered out

The tool automatically skips these common files and directories:

- Version control: `.git`, `.github`
- IDE/Editor: `.vscode`, `.idea`
- Dependencies: `node_modules`, `vendor`
- Build artifacts: `target`, `build`, `dist`, `__pycache__`
- System files: `.DS_Store`, `Thumbs.db`
- Environment files: `.env`, `.env.local`
- Most hidden files (except important ones like `.gitignore`, `.dockerignore`, etc.)

## Author

**Neil Mahajan**

- Email: [neilsmahajan@gmail.com](mailto:neilsmahajan@gmail.com)
- Portfolio: [https://neilsmahajan.com/](https://neilsmahajan.com/)

## Contributing

Feel free to open issues or submit pull requests to improve the tool!

## License

MIT License
