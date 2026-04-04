# Contributing to RecallKit

Thank you for your interest in contributing! This document covers everything you need to get started — from setting up the development environment to submitting a polished pull request.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
  - [Go contributors (no Node required)](#go-contributors-no-node-required)
  - [UI contributors (Svelte frontend)](#ui-contributors-svelte-frontend)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Commit Guidelines](#commit-guidelines)
  - [Signed Commits (Required)](#signed-commits-required)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs & Requesting Features](#reporting-bugs--requesting-features)

---

## Code of Conduct

Please be respectful, constructive, and inclusive. We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) in all project spaces.

---

## Getting Started

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/<your-username>/Recall-Kit.git
   cd Recall-Kit
   ```
3. **Add the upstream remote** so you can pull future changes:
   ```bash
   git remote add upstream https://github.com/RecallKit/Recall-Kit.git
   ```

---

## Development Setup

RecallKit has **two contributor personas** with different tooling requirements. The Svelte frontend assets are pre-built and committed to `ui/dist/`, so the Go binary always embeds a working UI without needing Node installed.

> [!NOTE]
> A `Makefile` is provided at the repo root. Run `make help` for a full list of targets.

### Go contributors (no Node required)

If you're working on the core engine, TUI, CLI, or any Go code, you only need Go and Ollama:

| Tool | Version |
|---|---|
| [Go](https://go.dev/dl/) | 1.24+ |
| [Ollama](https://ollama.com/) | Latest |
| [Git](https://git-scm.com/) | 2.34+ (with signing configured) |

```bash
# Download Go module dependencies
go mod download

# Run the TUI (make sure Ollama is running first)
ollama serve &
ollama pull llama3
go run . start

# Run all tests
make test                         # equivalent to: go test ./...

# Build the binary
make build                        # equivalent to: go build -o recallkit .
```

> [!IMPORTANT]
> `make test-race` (the race detector) requires CGO and a native 64-bit GCC on Windows. On Linux/macOS it runs out of the box. For Windows contributors, `make test` (without `-race`) is the standard.

### UI contributors (Svelte frontend)

If you're modifying the web UI, you additionally need Node 20+:

| Tool | Version |
|---|---|
| [Node.js](https://nodejs.org/) | 20+ LTS |
| npm | bundled with Node |

```bash
# Install JS dependencies and build Svelte assets into ui/dist/
make build-ui

# Hot-reload Svelte dev server (no Go binary needed)
make dev-ui

# Build Go binary picking up your fresh ui/dist/ assets
make build
```

> [!IMPORTANT]
> After running `make build-ui`, **commit the contents of `ui/dist/`** as part of your PR. This is the pre-built embed strategy — Go contributors on a fresh clone get a working binary without ever running Node.

```bash
git add ui/dist/
git commit -m "chore(ui): rebuild frontend assets"
```

---

## Project Structure

```
recallkit/
├── cmd/               # CLI entry points (Cobra commands)
├── internal/
│   ├── engine/        # Ollama REST client & streaming
│   ├── db/            # Kùzu graph DB layer
│   └── tui/           # Bubble Tea TUI components
└── main.go
```

Keep business logic inside `internal/` — packages there are intentionally unexported and not importable by external consumers.

---

## Making Changes

1. **Create a branch** from `dev` (not `main`):
   ```bash
   git checkout dev
   git pull upstream dev
   git checkout -b feat/your-feature-name
   ```
   Branch naming conventions:
   - `feat/<description>` — new feature
   - `fix/<description>` — bug fix
   - `chore/<description>` — tooling, deps, docs
   - `test/<description>` — test-only changes

2. **Write tests** for any new behaviour. PRs that reduce test coverage without justification will not be merged.

3. **Run the full test suite** and make sure everything passes:
   ```bash
   go test -race ./...
   ```

4. **Format and vet your code** before committing:
   ```bash
   go fmt ./...
   go vet ./...
   ```

---

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/). Each commit message should be in the form:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

**Types:** `feat`, `fix`, `docs`, `test`, `chore`, `refactor`, `perf`, `ci`

**Examples:**
```
feat(engine): add context-aware prompt injection
fix(tui): resolve panic on empty message history
docs: add CONTRIBUTING.md
test(engine): add unit tests for OllamaClient streaming
```

### Signed Commits (Required)

> [!IMPORTANT]
> **All commits to this repository must be cryptographically signed.** Unsigned commits will be rejected by the branch protection rules on `main` and `dev`.

We support **GPG** and **SSH** signing. Choose one:

#### Option A — GPG signing

1. [Generate a GPG key](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key) if you don't have one.
2. Add it to your GitHub account.
3. Configure Git to use it:
   ```bash
   git config --global user.signingkey <YOUR_KEY_ID>
   git config --global commit.gpgsign true
   ```

#### Option B — SSH signing (simpler, recommended)

1. Use an existing SSH key already added to GitHub, or [generate one](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent).
2. Configure Git:
   ```bash
   git config --global gpg.format ssh
   git config --global user.signingkey ~/.ssh/id_ed25519.pub
   git config --global commit.gpgsign true
   ```

#### Verify a signed commit

```bash
git log --show-signature -1
```

You should see `Good "git" signature` in the output.

---

## Pull Request Process

1. **Push your branch** to your fork:
   ```bash
   git push origin feat/your-feature-name
   ```

2. **Open a Pull Request** targeting the `dev` branch (not `main`). The `main` branch only receives vetted release merges from `dev`.

3. **Fill in the PR template** — describe:
   - What problem this solves
   - How you tested it
   - Any breaking changes

4. **PR checklist before requesting review:**
   - [ ] All commits are signed
   - [ ] `go test -race ./...` passes locally
   - [ ] `go fmt ./...` and `go vet ./...` show no issues
   - [ ] New public functions/types have doc comments
   - [ ] Updated `CHANGELOG.md` if applicable

5. **One approving review** from a maintainer is required before merging.

6. Prefer **squash merges** to keep the `dev` history clean. Maintainers may ask you to squash before merging.

---

## Reporting Bugs & Requesting Features

- **Bugs:** Open an issue using the **Bug Report** template. Include OS, Go version, Ollama version, and reproduction steps.
- **Features:** Open an issue using the **Feature Request** template before writing code — this avoids duplicated effort.
- **Security vulnerabilities:** Do **not** open a public issue. Email the maintainers directly.

---

Thank you for helping make RecallKit better! 🚀
