# gopk (Go Picker)

gopk is a personal bookmark manager and TUI launcher for your favorite Go packages.

Stop Googling "what was that full import path for viper?" every time you start a new project. gopk allows you to alias, organize, and fuzzy-find your commonly used Go modules, syncing them across all your development machines.

Note: This tool complements the standard Go toolchain. It helps you find packages, while go get handles the actual dependency resolution.

## Why gopk?

Brain-Friendly Aliases: Map gin to github.com/gin-gonic/gin. Never type the full URL again.

Interactive TUI: Built with Bubble Tea. Press Space to multi-select packages and Enter to install them all at once.

Distributed Workflow: (Coming Soon) Sync your favorite package list via GitHub Gists so your laptop and desktop are always in sync.

Visual Feedback: See descriptions and metadata before you install.

## Installation

go install [github.com/yourusername/gopk@latest](https://github.com/yourusername/gopk@latest)


🛠️ Usage

1. Interactive Mode (The Magic ✨)

Simply run the command without arguments to open the TUI:

```bash
gopk
```

Type to fuzzy search your saved packages.

Space to toggle selection (select multiple tools!).

Enter to run go get on all selected packages.

2. CLI Mode

Add a package to your registry:

``` bash
# Auto-detects alias from URL, or you can specify one
gopk add [https://github.com/spf13/cobra](https://github.com/spf13/cobra)
# Saved as "cobra" -> "[github.com/spf13/cobra](https://github.com/spf13/cobra)"
```


Install a specific package by alias:

gopk get cobra
# Executes: go get [github.com/spf13/cobra](https://github.com/spf13/cobra)

List all saved packages:

```bash
gopk list
```


## Configuration

Your package registry is stored in a simple JSON format, making it easy to back up or version control.

Location: ~/.config/gopk/packages.json

[
  {
    "alias": "gin",
    "url": "[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)",
    "description": "HTTP web framework written in Go"
  },
  {
    "alias": "bubbletea",
    "url": "[github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)",
    "description": "A powerful TUI framework"
  }
]

## Roadmap

[x] Basic CLI (Add/Remove/List)

[ ] Bubble Tea TUI implementation

[ ] Automatic description scraping from GitHub

[ ] Distributed Sync: Gist/S3 integration to sync config across devices

[ ] import scanner: Scan existing go.mod files to populate your list

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.
