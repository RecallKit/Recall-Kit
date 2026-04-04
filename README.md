# RecallKit

**RecallKit** is a local-first context engineering tool for LLMs. It maintains a persistent, queryable graph of your conversation history so that every session with your AI assistant is grounded in what you've already discussed — no cloud, no subscriptions, no data leaving your machine.

---

## Features

- 🧠 **Persistent Context Graph** — embeds conversation nodes and relationships into a local [Kùzu](https://kuzudb.com/) graph database using Cypher queries
- 💬 **Interactive TUI** — a full terminal chat interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- 🌐 **Web UI** — a Svelte-compiled frontend served directly from the binary via `go:embed`
- ⚡ **Local LLM** — streams responses from [Ollama](https://ollama.com/) running on your machine
- 📦 **Zero dependencies** — single statically compiled binary, no background services required (beyond Ollama)

---

## Prerequisites

| Requirement | Version |
|---|---|
| [Go](https://go.dev/dl/) | 1.24+ |
| [Ollama](https://ollama.com/) | Latest |
| A pulled Ollama model | e.g. `ollama pull llama3` |

---

## Installation

### From source

```bash
git clone https://github.com/RecallKit/Recall-Kit.git
cd Recall-Kit
go build -o recallkit .
```

### Run directly

```bash
go run . start
```

---

## Usage

```
recallkit <command>
```

| Command | Description |
|---|---|
| `recallkit start` | Launch the interactive terminal UI (TUI) |
| `recallkit ui` | Start the web UI server on `localhost:8001` |

### Quick start

```bash
# Make sure Ollama is running with a model available
ollama serve &
ollama pull llama3

# Start a session
./recallkit start
```

---

## Architecture

RecallKit is a single binary with a dual-interface design — both the TUI and web UI share the same core engine.

```
recallkit/
├── cmd/               # CLI entry points (Cobra commands)
│   ├── root.go        # Base command
│   ├── start.go       # 'recallkit start' → TUI
│   └── ui/            # 'recallkit ui' → Web server
├── internal/
│   ├── engine/        # Ollama REST client & streaming parser
│   ├── db/            # Kùzu graph DB connection & Cypher queries
│   └── tui/           # Bubble Tea UI model, update loop, and view
└── main.go
```

**Tech stack:**

- **Language:** Go — fast startup, single-binary builds, lightweight concurrency
- **CLI:** [Cobra](https://github.com/spf13/cobra)
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Graph DB:** [Kùzu](https://kuzudb.com/) (embedded, no server needed)
- **Web UI:** [Svelte](https://svelte.dev/) (compiled to static assets, embedded in binary)
- **LLM Runtime:** [Ollama](https://ollama.com/) (local REST API)

---

## Development

```bash
# Run tests
go test ./...

# Run with live reload (requires air)
air

# Build binary
go build -o recallkit .
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, coding conventions, and the pull request process.

---

## License

[MIT](LICENSE) © 2026 RecallKit
