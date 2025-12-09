# Firewall Utility - AI Coding Agent Instructions

## Project Overview
This is a firewall utility project located in `d:\projects\utils\firewall`.

**Status**: Functional implementation complete. Go 1.22; module placeholder is `module firewall` (update when repo path is known).

- Language: Go (golang)
- Backend: Sqlite
- Platforms: Windows + Linux
- CLI framework: Cobra
- GUI framework: Wails
- Guideline: run tests for any change when they exist.

## Current Layout

- `go.mod`: module placeholder; update to real import path when repo is established.
- `README.md`: goals, architecture, command examples, and implementation status.
- `cmd/cli`: Cobra-based CLI entry with root, rules, profiles, and version subcommands; fully wired to sqlite stores.
- `cmd/gui/main.go`: Wails GUI entry with AppService binding; includes embedded frontend in `frontend/dist/index.html`.
- `internal/app`: shared service layer wrapping rule store and platform dispatcher for CLI/GUI reuse.
- `internal/rules`: rule model with validation, sqlite schema/CRUD store with tests.
- `internal/profiles`: profile model, validation, and sqlite store for configuration management.
- `internal/platform`: dispatcher with OS-specific adapters using netsh (Windows) and iptables (Linux) behind build tags.
- `internal/notify`: OS-dispatched notifications using PowerShell MessageBox (Windows) and zenity (Linux).
- `internal/logging`: structured JSON-based event logging with file backend.
- `internal/stats`: in-memory traffic stats collection with filtering capabilities.
- `internal/config`: JSON configuration file support for DB paths, logging, default profiles, and GUI preferences.

## Development Workflow

- Build/run CLI: `go run ./cmd/cli --help`
- Build/run GUI: `go run ./cmd/gui` (requires Wails runtime; opens desktop window)
- Configuration: Copy `firewall.json.example` to `firewall.json` and customize settings
- CLI rule commands:
  - Add: `go run ./cmd/cli rules add --name web --app "C:/Program Files/App/app.exe" --action allow --protocol tcp --direction outbound --ports 80,443`
  - List: `go run ./cmd/cli rules list`
  - Remove: `go run ./cmd/cli rules remove --name web`
- CLI profile commands:
  - Create: `go run ./cmd/cli profiles create --name work --description "Work profile"`
  - List: `go run ./cmd/cli profiles list`
  - Activate: `go run ./cmd/cli profiles activate --name work`
  - Export: `go run ./cmd/cli profiles export --name work --file work.json`
  - Import: `go run ./cmd/cli profiles import --file work.json`
- Lint/format: `gofmt -w ./...`
- Tests: `go test ./...` (includes rule validation, store CRUD, and CLI integration tests)
- Dependencies: run `go mod tidy` after adding imports; main deps are cobra, sqlite3, and wails/v2.

## Architecture & Patterns

- Keep OS-specific code behind build tags in `internal/platform/{windows,linux}`; stubs exist for non-host builds.
- Core rule/profile models live in `internal/rules` and `internal/profiles`; validation enforces required fields. Stores use sqlite schema with upsert/delete/query helpers.
- CLI/GUI share logic via `internal/app.Service` and direct store access; prefer adding features to shared layers.
- Platform adapters dispatch by OS: Windows uses `netsh advfirewall firewall` commands, Linux uses `iptables` with comment-based tracking.
- Notifications dispatch by OS: Windows uses PowerShell MessageBox, Linux uses zenity dialogs.
- Logging writes JSON events line-by-line to file; stats kept in memory with query/filter API.

## App Requirements (from owner)

- GUI + CLI for firewall rule management. ✓ Implemented
- Windows + Linux compatibility; secure and efficient handling of network connections. ✓ Implemented via platform adapters
- Prompt allow/deny when an app connects without a prior rule. ✓ Notification infrastructure ready
- Profiles, import/export of configurations. ✓ Implemented with sqlite store and JSON export/import
- Logging of firewall events; stats with filtering for traffic data. ✓ Implemented with file-based JSON logging and in-memory stats
- Intuitive UI; powerful CLI covering all functionality. ✓ Wails GUI + Cobra CLI both operational

## Next Agent Steps

- Implement active connection monitoring to trigger allow/deny prompts (requires privileged network hooks: netfilter/Windows Filtering Platform).
- Update module path from `module firewall` to actual repository path when established.