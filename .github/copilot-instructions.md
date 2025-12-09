# Firewall Utility - AI Coding Agent Instructions

## Project Overview
This is a firewall utility project located in `d:\projects\utils\firewall`.

**Status**: Functional implementation complete. Go 1.22; module path is `github.com/vhPedroGitHub/firewall`.

- Language: Go (golang)
- Backend: Sqlite
- Platforms: Windows + Linux
- CLI framework: Cobra
- GUI framework: Wails
- Build the GUI using Wails v2 (requires Wails runtime installed)
- Guideline: run tests for any change when they exist.

## Current Layout

- `go.mod`: module github.com/vhPedroGitHub/firewall with all dependencies.
- `README.md`: goals, architecture, command examples, monitoring documentation, and implementation status.
- `cmd/cli`: Cobra-based CLI entry with root, rules, profiles, monitor, and version subcommands; fully wired to sqlite stores.
- `cmd/gui/main.go`: Wails GUI entry with AppService binding including monitor controls; embedded frontend in `frontend/dist/index.html`.
- `internal/app`: shared service layer wrapping rule store and platform dispatcher for CLI/GUI reuse.
- `internal/rules`: rule model with validation, sqlite schema/CRUD store with tests.
- `internal/profiles`: profile model, validation, and sqlite store for configuration management.
- `internal/platform`: dispatcher with OS-specific adapters using netsh (Windows) and iptables (Linux) behind build tags.
- `internal/notify`: OS-dispatched notifications with user choice return using PowerShell MessageBox (Windows) and zenity (Linux).
- `internal/logging`: structured JSON-based event logging with file backend.
- `internal/stats`: in-memory traffic stats collection with filtering capabilities.
- `internal/config`: JSON configuration file support for DB paths, logging, default profiles, and GUI preferences.
- `internal/monitor`: connection monitoring infrastructure with Windows (netstat-based) and Linux (/proc/net-based) implementations; DefaultHandler for rule checking and user prompts; auto-rule creation from user decisions.

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
- CLI monitor commands:
  - Start: `go run ./cmd/cli monitor start`
  - Stop: `go run ./cmd/cli monitor stop`
  - Status: `go run ./cmd/cli monitor status`
- Lint/format: `gofmt -w ./...`
- Tests: `go test ./...` (includes rule validation, store CRUD, and CLI integration tests)
- Dependencies: run `go mod tidy` after adding imports; main deps are cobra, sqlite3, and wails/v2.

## Architecture & Patterns

- Keep OS-specific code behind build tags in `internal/platform/{windows,linux}`; stubs exist for non-host builds.
- Core rule/profile models live in `internal/rules` and `internal/profiles`; validation enforces required fields. Stores use sqlite schema with upsert/delete/query helpers.
- CLI/GUI share logic via `internal/app.Service` and direct store access; prefer adding features to shared layers.
- Platform adapters dispatch by OS: Windows uses `netsh advfirewall firewall` commands, Linux uses `iptables` with comment-based tracking.
- Notifications dispatch by OS: Windows uses PowerShell MessageBox returning user choice, Linux uses zenity dialogs.
- Logging writes JSON events line-by-line to file; stats kept in memory with query/filter API.
- Monitoring uses polling: Windows via netstat, Linux via /proc/net; production would use WFP/netfilter hooks. DefaultHandler checks rules, prompts users for unknown connections, and saves decisions as auto-generated rules.

## App Requirements (from owner)

- GUI + CLI for firewall rule management. ✓ Implemented
- List all processes attempting network connections. ✓ Implemented via active processes tracking in monitor tab
- Windows + Linux compatibility; secure and efficient handling of network connections. ✓ Implemented via platform adapters
- Prompt allow/deny when an app connects without a prior rule. ✓ Implemented with optional toggle to enable/disable prompts
- Profiles, import/export of configurations. ✓ Implemented with sqlite store and JSON export/import
- Logging of firewall events; stats with filtering for traffic data. ✓ Implemented with file-based JSON logging and in-memory stats
- Intuitive UI; powerful CLI covering all functionality. ✓ Wails GUI + Cobra CLI both operational with monitor controls

## Next Agent Steps

Implementa en el monitor una mejor visualizacion de los datos de los procesos activos y eventos de conexion, sigue esta estructura:
PID, Nombre de la aplicacion, Direccion local, Puerto local, Direccion remota, Puerto remoto, Estado, Hora del evento
actualiza automáticamente cada 5 segundos. 
mostrar en tiempo real el trafico de red (entrante y saliente) por proceso.
Agregar tests unitarios/integracion faltantes para el monitor y handler.
la parte de estadisitica debes completarla y que funcione bien