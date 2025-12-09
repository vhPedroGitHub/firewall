# Firewall Utility (Go)

Cross-platform firewall utility with both CLI and GUI frontends for managing application network access, profiles, and telemetry.
Backend storage uses Sqlite; CLI uses Cobra; GUI uses Wails with embedded web frontend.

**Status**: Functional implementation complete. Ready for testing and deployment.

## Goals

- Cross-platform support: Windows + Linux ✓
- Dual interfaces: command-line and graphical UI ✓
- Safe rule management: allow/deny prompts when no prior rule exists ✓ (infrastructure ready)
- Profiles: multiple firewall configurations with import/export ✓
- Observability: logging, stats, filtering on traffic and events ✓

## Implemented Components

- `cmd/cli`: Cobra-based CLI entry point for rule and profile management with full CRUD operations
- `cmd/gui`: Wails-based GUI with embedded frontend for interactive rule control and monitoring
- `internal/app`: shared service layer reused by CLI/GUI for business logic
- `internal/rules`: core rule model, validation, and sqlite persistence (schema + CRUD + tests)
- `internal/profiles`: profile model, validation, and sqlite store with export/import capabilities
- `internal/platform/{windows,linux}`: OS-specific adapters using netsh (Windows) and iptables (Linux)
- `internal/notify`: per-OS desktop notifications (PowerShell MessageBox on Windows, zenity on Linux)
- `internal/logging`: structured JSON event logging with file backend
- `internal/stats`: in-memory metrics collection with filtering
- `internal/config`: JSON configuration file support for customizable settings

## Development

- Language: Go 1.22+
- Init: `go mod tidy` after dependency or code changes
- Tests: `go test ./...` - includes rule validation, store CRUD, and CLI integration tests
- Format: `gofmt -w ./...`
- Tooling: Cobra for CLI; Wails for GUI; sqlite for storage; OS-specific notifications and firewall APIs

## Architecture Notes

- CLI/GUI share core operations through `internal/app.Service`, which wraps the rule store and platform dispatcher.
- Platform adapters live under `internal/platform` with build-tagged OS folders; stubs exist for non-host OS builds.
- Notifications dispatch by OS in `internal/notify` using native dialog systems.
- Rules and profiles stored in sqlite with JSON serialization for complex types.
- Logging writes line-delimited JSON events; stats kept in memory with query API.

## Commands

### CLI Usage

- CLI help: `go run ./cmd/cli --help`
- Rules:
  - Add: `go run ./cmd/cli rules add --name web --app "C:/Program Files/App/app.exe" --action allow --protocol tcp --direction outbound --ports 80,443`
  - List: `go run ./cmd/cli rules list`
  - Remove: `go run ./cmd/cli rules remove --name web`
- Profiles:
  - Create: `go run ./cmd/cli profiles create --name work --description "Work profile"`
  - List: `go run ./cmd/cli profiles list`
  - Activate: `go run ./cmd/cli profiles activate --name work`
  - Export: `go run ./cmd/cli profiles export --name work --file work.json`
  - Import: `go run ./cmd/cli profiles import --file work.json`
- Monitoring:
  - Start: `go run ./cmd/cli monitor start` - begins monitoring connections and prompts for unknown apps
  - Stop: `go run ./cmd/cli monitor stop` - stops connection monitoring
  - Status: `go run ./cmd/cli monitor status` - shows whether monitoring is active
- Version: `go run ./cmd/cli version`

### GUI Usage

- Run: `go run ./cmd/gui`
- Opens desktop window with:
  - **Rules tab**: add/remove rules, view all configured rules
  - **Profiles tab**: create/activate profiles, manage configurations
  - **Statistics tab**: view traffic metrics, bandwidth usage
  - **Logs tab**: browse firewall events and audit trail
  - **Monitoring controls**: start/stop connection monitoring with automatic allow/deny prompts
- Configuration: Uses `firewall.json` if present, otherwise defaults (see `firewall.json.example`)

### Build

- CLI: `go build -o firewall-cli ./cmd/cli`
- GUI (development): `go build -o firewall-gui ./cmd/gui`
- GUI (production): `wails build` - creates optimized executable in `build/bin/`

## Connection Monitoring

The firewall can actively monitor network connections and prompt users when applications attempt connections without existing rules:

- **Windows**: Uses netstat polling (simplified implementation). Production would use Windows Filtering Platform (WFP) APIs.
- **Linux**: Reads /proc/net/tcp and /proc/net/udp (simplified). Production would use netfilter/nfqueue hooks.
- **User Prompts**: When unknown connection detected, displays OS-native dialog asking to allow/deny
- **Auto-Rule Creation**: User decisions are automatically saved as permanent rules

Note: Current monitoring implementation is polling-based. For production use with high traffic volumes, consider implementing:

- Windows: WFP callout drivers for kernel-level connection detection
- Linux: nfqueue with libnetfilter_queue for real-time packet inspection

## Next Steps

1. ~~Implement active connection monitoring infrastructure~~ ✓ Complete
2. ~~Update module path to actual repository path~~ ✓ Complete (github.com/vhPedroGitHub/firewall)
3. Enhance monitoring with kernel-level hooks for production environments
