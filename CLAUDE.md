# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run for current platform
go run cmd/rapidfeed/main.go

# Build for current platform
make bin

# Cross-compile for linux/darwin × amd64/arm64
make build

# Run all tests
go test ./...

# Run a single test
go test ./internal/db/... -run TestGetUserInfo

# Run tests with verbose output
go test -v ./...

# Docker build (latest tag)
make docker-latest
```

## Architecture

RapidFeed is a server-rendered RSS reader with no frontend JavaScript. It runs two HTTP servers concurrently:

- **Main app** (`:8080` by default) — Fiber-based web server serving HTML
- **MCP server** (`:8090` by default) — Streamable HTTP MCP endpoint for LLM tool access

### Startup sequence (`cmd/rapidfeed/main.go`)

`init()` loads env config → opens SQLite DB → `RunMigrations(MigrateUp)` → `CreateDefaultAdmin()`. Then `main()` starts `feeder.StartAutoRefresh()` in a goroutine, `mcp.Start()` in another goroutine, and calls `http.New()` which blocks.

### Key packages

| Package | Purpose |
|---|---|
| `internal/db` | All SQLite access via the global `db.DB *sql.DB`. Migrations via `golang-migrate` (embedded SQL files in `migrations/`), and all CRUD live here. |
| `internal/http` | Fiber route registration, handlers, session middleware, template engine setup. |
| `internal/feeder` | Fetches RSS feeds via `gofeed`, deduplicates by `(link, feed_url)`, and runs a per-user auto-refresh loop checking every minute. |
| `internal/mcp` | MCP server exposing `feeds_today`, `feeds_yesterday`, `feeds_latest` tools. Auth via `user_tokens` table. |
| `internal/auth` | Password hashing and token generation. |
| `internal/ui` | Embeds `templates/` and `static/` via `embed.FS`. |
| `internal/utils` | Env var config globals (`Listen`, `MCPListen`, `SecretKey`, etc.) and feed text normalization helpers. |
| `internal/models` | Shared struct types (`User`, `UserFeed`, `UserWithFeeds`, etc.). |

### Database schema

SQLite at `./feeds.db` (configurable via `DB_PATH`). Tables:
- `feeds` — aggregated feed items; deduplicated on `(link, feed_url)`
- `users` — roles: `user`, `admin`, `blocked`
- `user_feeds` — per-user RSS feed subscriptions with optional `category`
- `user_tokens` — one MCP access token per user (upsert semantics)
- `user_refresh_settings` — per-user auto-refresh interval and next update timestamp
- `token_storage` — session tokens for the web UI

Migrations are managed by `golang-migrate` with versioned SQL files in `migrations/`. They run automatically at startup via `RunMigrations(MigrateUp)`. Use `cmd/migrator` or `make migrate-up` / `make migrate-down` to run migrations manually.

### Authentication

Two separate auth systems:
- **Web sessions**: cookie-based via `checkSessionMiddleware` / `adminSessionMiddleware`
- **MCP tokens**: `X-MCP-Token` or `Authorization: Bearer` header, validated against `user_tokens`

### Templates

HTML templates are embedded at compile time from `internal/ui/templates/`. The Fiber template engine is initialized in `internal/http/templateEngine.go` with custom funcs (`add`, `sub`, `seq`, `min`, `max`, `urlquery`).

### Feed text processing

All feed content (titles, descriptions, sources) is run through `utils.StripHTMLAndNormalizeFeedText` before storage. `NormalizeFeedText` is the inner helper used by `StripHTMLAndNormalizeFeedText`.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `LISTEN` | `:8080` | Web server address |
| `MCP_LISTEN` | `:8090` | MCP server address |
| `SECRET_KEY` | `strong-secretkey` | Session secret — change before first run |
| `REGISTRATION_ALLOWED` | `true` | Whether `/register` is exposed |
| `DB_PATH` | `./feeds.db` | SQLite database path |
| `ADMIN_PASSWORD` | *(random, logged once)* | Override default admin password |
