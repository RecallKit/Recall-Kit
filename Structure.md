# RecallKit — Implementation Guide

This document outlines the technical architecture, project structure, and phased implementation plan for RecallKit, a local-first context engineering tool for LLMs. 

## 1. Architecture Overview

RecallKit operates as a single, statically compiled binary with zero external service dependencies (other than the local Ollama runtime). It features a dual-interface design sharing a unified core engine.

### Tech Stack
* **Core Language:** Go (Golang) - Chosen for its lightweight concurrency, fast startup times, and single-binary compilation.
* **CLI Framework:** Cobra - Manages routing and command execution (`start`, `inject`, `ui`).
* **Terminal UI (TUI):** Bubble Tea - Handles the interactive terminal chat experience.
* **Database (Context Graph):** Kùzu - An embedded, in-process graph database using the Cypher query language to map conversational nodes and edges without requiring a background server.
* **Web UI Framework:** Svelte - Compiled to static assets and served via Go's `go:embed`.
* **LLM Runtime:** Ollama (Local REST API).

---

## 2. Project Structure

The repository follows standard Go project layout conventions to strictly isolate the business logic from the interfaces.

```text
recallkit/
├── cmd/                    # CLI entry points (Cobra commands)
│   ├── root.go             # Base 'recallkit' command
│   ├── start.go            # 'recallkit start' (TUI entry)
│   ├── ui.go               # 'recallkit ui' (Web server entry)
│   └── start_test.go       
├── internal/               # Private application code
│   ├── engine/             # Core LLM logic
│   │   ├── ollama.go       # Ollama REST client and streaming parser
│   │   └── context.go      # Context composition and injection logic
│   ├── db/                 # Database layer
│   │   ├── kuzu.go         # Kùzu connection management
│   │   └── queries.go      # Cypher queries for node/edge retrieval
│   └── tui/                # Bubble Tea UI components
│       ├── model.go        # TUI state definition
│       ├── update.go       # Keystroke and message event handling
│       └── view.go         # Terminal rendering logic (markdown parsing)
├── ui/                     # Svelte frontend source code
│   ├── src/                
│   ├── package.json        
│   └── svelte.config.js    
├── main.go                 # Main application bootstrapper
├── go.mod
└── go.sum