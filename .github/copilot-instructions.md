# PicoClaw — Copilot Instructions

Ultra-lightweight personal AI assistant in Go. Runs on $10 hardware, <10MB RAM, boots in <1s.

## Go Conventions

- **Go 1.25.7** — module path: `github.com/sipeed/picoclaw`
- Import grouping: stdlib → 3rd-party → internal (`github.com/sipeed/picoclaw/...`), separated by blank lines
- Errors: return `error` as last value, wrap with `fmt.Errorf("context: %w", err)`, never panic in library code
- Naming: camelCase locals, PascalCase exports, acronyms uppercase (`ID`, `URL`, `HTTP`)
- Use `zerolog` for logging (`log.Info().Str("key", val).Msg("...")`) — no `fmt.Println` or `log.Printf`
- JSON tags on all exported struct fields: `json:"field_name"`
- Context: pass `context.Context` as first parameter when available

## Architecture

### Registration Patterns
- **Providers**: factory pattern via `init()` in each provider package → `RegisterProvider(name, constructor)` in `pkg/providers/factory.go`
- **Channels**: `init()` registration → `RegisterChannel(name, constructor)` in `pkg/channels/registry.go`
- **Tools**: implement `Tool` interface (Name, Description, InputSchema, Execute) in `pkg/tools/`, optional `AsyncExecutor` for streaming
- **MCP**: dynamic tool discovery with TTL-based unlocking via `pkg/mcp/manager.go`

### Message Bus
- Generic pub/sub in `pkg/bus/bus.go` with typed `Publish[T]` / `Subscribe[T]` helpers
- All inter-component communication goes through the bus

### Config
- JSON config loaded by `pkg/config/config.go`
- Environment override: `PICOCLAW_TOOLS_<SECTION>_<KEY>=value`
- Model routing via `model_list` entries with `model_name` / `model` / `api_key` / `api_base`

### Key Files (handle with care)
| File | Purpose |
|------|---------|
| `pkg/agent/loop.go` | Core agent event loop |
| `pkg/providers/protocoltypes/types.go` | Shared LLM protocol types (all providers depend on this) |
| `pkg/providers/factory.go` | Provider registration and routing |
| `pkg/channels/registry.go` | Channel registration |
| `pkg/tools/registry.go` | Tool registration and TTL discovery |
| `pkg/config/config.go` | Configuration loading |
| `pkg/bus/bus.go` | Message bus |
| `cmd/picoclaw/main.go` | CLI entry point (cobra) |

## Build & Test

```bash
make build          # Build for current platform (runs go generate first)
make test           # Run all tests
make check          # deps + fmt + vet + test (CI gate)
make lint           # golangci-lint v2
make fmt            # Format code
make vet            # Static analysis
```

- CI pipeline (`.github/workflows/pr.yml`): go generate → lint → test → govulncheck
- Cross-compile targets: linux/amd64, arm, arm64, mipsle, riscv64, loong64, darwin/arm64, windows/amd64

## Testing

- Framework: `testify` (`assert` for soft checks, `require` for fatal)
- Test files: colocated `*_test.go` in same package
- Mocks: inline structs implementing interfaces (no codegen mock framework)
- Run single package: `go test ./pkg/agent/...`

## Security

- **Exec tool**: 40+ deny regex patterns block dangerous shell commands (rm -rf, format, shutdown, sudo, etc.) — defined in `pkg/tools/exec.go`
- **Workspace sandbox**: `restrict_to_workspace: true` by default — file/exec operations confined to workspace dir
- **Credentials**: support `enc://` prefix for AES-256-GCM encrypted values in config
- Never disable sandbox or weaken exec deny patterns without explicit justification

## Conventions

- Build tag `whatsapp_native` enables whatsmeow (larger binary); default build excludes it
- Skills live in `workspace/skills/` with a `SKILL.md` manifest
- Channel webhook services share a single gateway (`gateway.host:gateway.port`)
- `go generate` embeds `workspace/` into the binary — run `make generate` after modifying workspace files
- PR branches: branch off `main`, one feature per branch
