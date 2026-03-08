# floxybot

A standalone CLI tool for getting advice, collaborating, and automating tasks with [Flox](https://flox.dev). Powered by Claude, with local RAG over Flox documentation.

## What it does

**floxybot** is a terminal-native AI assistant that knows Flox. It ships with a pre-indexed snapshot of the Flox docs and blog ("canon"), uses Voyage AI for semantic search and reranking, and talks to Claude for answers.

Three modes:

- **Chat** — Tabbed TUI for interactive Q&A. Ask questions about Flox, get streaming markdown answers backed by RAG-retrieved documentation.
- **Co-Pilot** — Agent automation. Describe a task ("set up a Python dev environment"), and Claude executes it via tool-use against the Flox MCP server.
- **Context** — Detects your current Flox environment (active env, installed packages, manifest, env vars) and uses it to give context-aware answers.

## Quick start

```bash
# Clone and enter the dev environment
git clone https://github.com/8BitTacoSupreme/floxybot.git
cd floxybot
flox activate

# Build
make build

# Set API keys
export ANTHROPIC_API_KEY="sk-..."
export VOYAGE_API_KEY="pa-..."

# Download the canon snapshot
./floxybot canon update

# Launch the TUI
./floxybot
```

## Commands

| Command | Description |
|---------|-------------|
| `floxybot` | Launch the tabbed TUI (default) |
| `floxybot chat` | Same as above — opens Chat tab |
| `floxybot ask "question"` | Single-shot Q&A, prints to stdout |
| `floxybot agent "task"` | Run a task via Claude tool-use loop |
| `floxybot context` | Print detected Flox environment info |
| `floxybot canon update` | Download latest canon snapshot |
| `floxybot version` | Print version |

### Global flags

```
--api-key string   Anthropic API key (or ANTHROPIC_API_KEY env var)
--model string     Claude model (default: claude-sonnet-4-20250514)
-v, --verbose      Verbose output
```

### Agent mode

The `agent` command starts a Claude tool-use loop that drives `flox-mcp-server` over stdio. It can search packages, install them, edit manifests, and manage environments.

```bash
# Auto-approve all tool calls
./floxybot agent --yes "search for python packages"

# Interactive confirmation for destructive ops (default)
./floxybot agent "set up a Go dev environment"
```

Requires `flox-mcp-server` in PATH (included in the Flox dev environment).

## Configuration

Config file: `~/.config/floxybot/config.toml`

```toml
anthropic_api_key = "sk-..."
voyage_api_key = "pa-..."
model = "claude-sonnet-4-20250514"
backend_url = "https://floxybot.example.com"
```

Environment variables override the config file:

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | Claude API access |
| `VOYAGE_API_KEY` | Voyage AI embeddings + reranking |
| `FLOXYBOT_MODEL` | Claude model override |
| `FLOXYBOT_CANON_DIR` | Canon snapshot directory |
| `FLOXYBOT_BACKEND_URL` | Backend server URL |

## Architecture

```
┌────────────────────────────────────────────────────────┐
│                   floxybot binary                      │
│                                                        │
│  ┌─[Chat]──[Co-Pilot]──[Context]────────────────────┐  │
│  │  Tabbed TUI (bubbletea + glamour markdown)       │  │
│  └──────────────┬──────────────────┬────────────────┘  │
│                 │                  │                    │
│  ┌──────────────▼───────┐  ┌──────▼──────────────┐     │
│  │  RAG Pipeline        │  │  MCP Client         │     │
│  │  Voyage embed query  │  │  (stdio JSON-RPC)   │     │
│  │  → chromem-go search │  └──────┬──────────────┘     │
│  │  → Voyage rerank     │         │ subprocess         │
│  │  → Claude generate   │  ┌──────▼──────────────┐     │
│  └──────────┬───────────┘  │  flox-mcp-server    │     │
│             │              └─────────────────────┘     │
│  ┌──────────▼───────────┐  ┌─────────────────────┐     │
│  │  Canon Snapshot      │  │  Flox Context       │     │
│  │  (gob, ~3MB)         │  │  (env, manifest,    │     │
│  └──────────────────────┘  │   packages)          │     │
│                            └─────────────────────┘     │
└───────────────────────────┬────────────────────────────┘
                            │ HTTPS
                   ┌────────▼─────────┐
                   │  Linode Backend   │
                   │  /canon/latest    │
                   │  /feedback        │
                   └──────────────────┘
```

### RAG pipeline

```
User query
  → Voyage embed (voyage-3-lite, 512 dims)
  → chromem-go vector search (top 20 candidates)
  → Voyage rerank (voyage-rerank-2, top 5)
  → Inject ranked context into Claude system prompt
  → Claude generates streaming response
```

### Canon (documentation index)

The canon is a pre-built snapshot of flox.dev/docs and flox.dev/blog, stored as a gob file. It contains chunked text with URLs and titles for citation.

- Built offline with `go run ./canon/build/` (converts JSON chunks to gob)
- Hosted on the backend at `GET /canon/latest`
- Downloaded by `floxybot canon update` with ETag caching
- ~50 pages, ~10K chunks, ~3MB

### Project layout

```
cmd/floxybot/          Entry point
internal/
  cli/                 Cobra commands
  tui/                 Bubbletea TUI (tabs, chat, copilot, context)
  claude/              Claude API client (streaming, tool use)
  voyage/              Voyage AI (embeddings, reranking)
  rag/                 RAG pipeline (chunker, vector store, retriever)
  canon/               Canon management (scraper, snapshot, updater)
  mcp/                 MCP subprocess client (JSON-RPC 2.0)
  agent/               Tool-use loop with safety checks
  floxctx/             Flox environment detection
  config/              XDG config loader
  feedback/            Anonymous feedback client
server/                Linode backend (canon hosting + feedback API)
canon/build/           Offline canon builder
deploy/                Linode provisioning scripts
embedded/              System prompt template
```

## Development

```bash
# Enter Flox dev environment (installs Go, flox-mcp-server, etc.)
flox activate

# Build
make build

# Run tests
make test

# Lint
make vet
make lint   # requires staticcheck
```

### Key dependencies

| Library | Purpose |
|---------|---------|
| [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) | Claude API (chat, streaming, tool use) |
| [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [glamour](https://github.com/charmbracelet/glamour) | Markdown rendering |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [cobra](https://github.com/spf13/cobra) | CLI framework |
| [chromem-go](https://github.com/philippgille/chromem-go) | In-memory vector DB (pure Go) |
| [BurntSushi/toml](https://github.com/BurntSushi/toml) | TOML config parsing |

## Backend

A minimal Go HTTP server on Linode serving two endpoints:

- `GET /canon/latest` — Latest canon snapshot (gob file, ETag support)
- `POST /feedback` — Anonymous vote submission (stored in PostgreSQL)

See `deploy/` for provisioning scripts.
