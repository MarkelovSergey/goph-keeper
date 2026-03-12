# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`goph-keeper` is a password/credential manager with a server-client architecture:
- **Server**: HTTP API backend with PostgreSQL storage, JWT auth, optional TLS
- **Client**: CLI application for interacting with the server

Module: `github.com/MarkelovSergey/goph-keeper`

## Commands

### Build
```bash
make build           # Build both server and client to bin/
make build-server    # bin/goph-keeper-server
make build-client    # bin/goph-keeper
make build-all       # Cross-platform (Linux, macOS, Windows)
```

### Run
```bash
make run-server      # go run ./cmd/server
make run-client      # go run ./cmd/client
```

### Test & Lint
```bash
make test            # Run all tests
make test-cover      # Run tests with coverage report (coverage.out)
make lint            # Run golangci-lint
```

### Database
```bash
make docker-up       # Start PostgreSQL 16 container (port 5432)
make docker-down     # Stop and remove containers
make migrate-up      # Apply migrations (requires DATABASE_DSN env var)
make migrate-down    # Rollback migrations
```

### Docs
```bash
make swag            # Generate Swagger docs from cmd/server/main.go into docs/
```

## Environment Variables

Copy `.env.example` to `.env`. Config loaders search parent directories upward for `.env`.

**Server** (`internal/server/config/config.go`):
- `LISTEN_ADDR` — TCP listen address (default: `:8080`)
- `DATABASE_DSN` — PostgreSQL DSN (required)
- `JWT_SECRET` — HMAC secret for JWT (required)
- `TLS_CERT` / `TLS_KEY` — TLS certificate paths (optional)

**Client** (`internal/client/config/config.go`):
- `SERVER_ADDRESS` — Server base URL (default: `http://localhost:8080`)
- `TLS_INSECURE` — Skip TLS verification (default: false)
- `GOPHKEEPER_CONFIG_DIR` — Config dir (default: `~/.gophkeeper`)

## Architecture

Standard layered Go architecture. The project is a skeleton — most layers are stubs awaiting implementation.

**Server** (`internal/server/`):
```
handler/      HTTP request handlers
middleware/   HTTP middleware (auth, logging, etc.)
service/      Business logic
repository/   Data access (PostgreSQL)
model/        Server-specific domain models
app/          Application wiring/initialization
config/       Configuration (implemented)
```

**Client** (`internal/client/`):
```
cmd/          Cobra CLI command definitions
api/          HTTP client for server communication
app/          Client application logic
crypto/       Cryptographic operations
config/       Configuration (implemented)
```

**Shared**: `internal/model/` — models used by both server and client.

**Entry points**: `cmd/server/main.go`, `cmd/client/main.go`

Binaries receive `version` and `buildDate` via ldflags at build time; both support a `version` subcommand.
