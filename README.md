# gopk

**gopk** is a personal package registry and CLI launcher for Go modules.

It helps you remember, alias, and install the Go packages you already use—without searching, retyping long import paths, or relying on IDE suggestions.

gopk complements the Go toolchain: it helps you *recall* packages, while `go get` handles dependency resolution.

---

## Why gopk exists

Most Go developers repeatedly use the same set of packages:

- a logger  
- a web framework  
- a CLI helper  
- a config library  

But Go module paths are long, easy to forget, and often ambiguous. IDEs can install packages once you choose them, but they don’t remember your personal preferences across projects or machines.

gopk acts as **personal memory** for your Go dependencies.

This is **not** a discovery tool or public registry.  
If a package is in gopk, you have already used and trusted it before.

---

## Core ideas

- **Aliases over memory**  
  Map short names to full module paths.
```

zap  → go.uber.org/zap
gin  → github.com/gin-gonic/gin

````

- **Fast recall, not discovery**  
Name + module path is enough to recognize a package you already know.

- **Explicit side effects**  
Adding a package does not modify your project unless you explicitly ask it to.

- **Project-aware installs**  
Installation is always tied to a Go module (`go.mod`), never global.

---

## Installation

```bash
go install github.com/lewvy/gopk@latest
````

---

## Usage

### Add a package to your registry

```bash
gopk add go.uber.org/zap
```

By default, this only stores the package for later use.

You can provide a custom alias:

```bash
gopk add github.com/gin-gonic/gin --name gin
```

Semantic import versions (`/v2`, `/v3`, …) are ignored when inferring aliases unless you explicitly provide one.

---

### Add and install immediately

```bash
gopk add go.uber.org/zap --install
```

This will:

1. Save the package in your gopk registry
2. Run `go get` in the current Go module

---

### Install saved packages into a project

```bash
gopk get zap
gopk get zap gin
```

`get` resolves aliases and runs `go get` for each package.

This command:

* Requires an existing `go.mod`
* Does not modify the gopk registry

---

### List saved packages

```bash
gopk list
```

Displays all saved aliases and their module paths.

---

## Storage & configuration

gopk stores its data locally using SQLite.

* **Data directory**: `~/.local/share/gopk/`
* **Config directory**: `~/.config/gopk/`

Aliases are enforced as **globally unique identifiers** to guarantee deterministic resolution.

---

## Design philosophy

* Instant startup and offline-first
* No hidden network calls
* No implicit project mutations
* Clear separation between remembering and installing
* Domain-aware handling of Go modules

If something feels surprising, it is probably a bug.

---

## Roadmap

* [ ] Interactive TUI (Bubble Tea)
* [ ] Import scanner (`go.mod` → gopk)
* [ ] Manual cross-device sync (GitHub Gist)
* [ ] Optional metadata enrichment (explicit, cached)

---

## Non-goals

* Public package discovery
* Ranking or recommendations
* Replacing `go get`
* Acting as a package manager

---

## License

MIT
