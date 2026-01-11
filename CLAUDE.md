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
  cli/root.go               # Cobra root command (28 flags)
  version/version.go        # Version info (set via ldflags)
  config/                   # Principles YAML loading (TBD)
  loop/                     # Main loop engine (TBD)
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

## Migration Status

Phase 1 (Current): Project initialization
- See [DOU-137](https://linear.app/doublew/issue/DOU-137) for current task
- Full issue list in Linear project: Claude Loop Go Migration
