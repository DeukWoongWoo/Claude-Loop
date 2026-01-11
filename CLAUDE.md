# Claude Loop

Autonomous AI development loop CLI - Go migration from bash (2600+ lines).

## Tech Stack

- Go 1.21+
- Cobra v1.8.1 (CLI framework)
- Viper v1.19.0 (config management)
- yaml.v3 (YAML parsing)
- testify v1.9.0 (testing)

## Project Structure

```
cmd/claude-loop/main.go     # Entry point
internal/
  cli/
    root.go                 # Cobra root command, flag registration
    flags.go                # Flags struct, defaults
    validation.go           # CLI validation logic
    *_test.go               # Unit tests (95%+ coverage)
  version/version.go        # Version info (set via ldflags)
  config/
    types.go                # Preset, Principles, Layer0, Layer1 structs
    defaults.go             # Preset defaults (startup, enterprise, opensource)
    validation.go           # Schema validation (version, date, 1-10 range)
    loader.go               # YAML file loading
    *_test.go               # Unit tests (99% coverage)
  loop/
    types.go                # Config, State, StopReason, ClaudeClient interface
    errors.go               # LoopError, IterationError types
    limits.go               # LimitChecker (cost/time/runs)
    completion.go           # CompletionDetector (signal detection)
    iteration.go            # IterationHandler (single iteration)
    executor.go             # Executor (main loop orchestration)
    *_test.go               # Unit tests (98% coverage)
  claude/                   # Claude CLI wrapper (TBD)
  git/                      # Git/worktree operations (TBD)
  github/                   # PR workflow (TBD)
  update/                   # Auto-update (TBD)
docs/
  CLI_CONTRACT.md           # CLI interface contract
  FEATURE_MATRIX.md         # Bash → Go migration tracking
  PRINCIPLES_SCHEMA.md      # principles.yaml schema
test/golden/                # Output snapshots for regression
```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/claude-loop` |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make golden-test` | Compare output with golden files |
| `make clean` | Remove build artifacts |

## Coding Conventions

- Standard Go project layout (`cmd/`, `internal/`)
- Version injected via ldflags: `-X .../version.Version=$(VERSION)`
- Error handling: return errors, log at top level
- No `panic` in library code
- Tests in same package with `_test.go` suffix

## Key Files

- [CLI Contract](docs/CLI_CONTRACT.md) - 28 flags, exit codes, validation rules
- [Feature Matrix](docs/FEATURE_MATRIX.md) - Migration progress tracking
- [Golden Tests](test/golden/) - help.txt, version.txt snapshots
