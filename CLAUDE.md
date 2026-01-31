# Claude Loop

Autonomous AI development loop CLI written in Go.

## Tech Stack

- Go 1.21+
- Cobra v1.8.1 (CLI framework)
- Viper v1.19.0 (config management)
- yaml.v3 (YAML parsing)
- testify v1.9.0 (testing)

## Documentation

| Document | Description |
|----------|-------------|
| [README.md](README.md) | User guide, installation, CLI reference |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributor guidelines, dev setup |
| [docs/CLI_CONTRACT.md](docs/CLI_CONTRACT.md) | CLI interface contract (28 flags) |
| [docs/PRINCIPLES_SCHEMA.md](docs/PRINCIPLES_SCHEMA.md) | principles.yaml schema |
| [docs/FEATURE_MATRIX.md](docs/FEATURE_MATRIX.md) | Implementation status tracking |

## Project Structure

```
cmd/claude-loop/main.go     # Entry point
internal/
  cli/
    root.go                 # Cobra root command, flag registration, main loop execution
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
  claude/
    client.go               # Claude Code subprocess wrapper
    parser.go               # JSON stream parser
    errors.go               # ClaudeError, ParseError types
    *_test.go               # Unit tests
  git/
    branch.go               # Branch operations (create, switch, delete)
    worktree.go             # Git worktree management
    commit.go               # Git staging, commit, push operations
    errors.go               # GitError, BranchError, WorktreeError types
    *_test.go               # Unit tests
  github/
    repo.go                 # Repository info and detection
    pr.go                   # PR creation and management
    checks.go               # CI status monitoring
    ci.go                   # CI failure analysis
    cifix.go                # CI auto-fix orchestration
    workflow.go             # Complete PR workflow manager
    types.go                # Config types (WorkflowConfig, CIFixConfig)
    errors.go               # GitHubError, PRError, CheckError types
    *_test.go               # Unit tests
  prompt/
    builder.go              # Main prompt builder
    templates.go            # Prompt templates
    cifix.go                # CI fix prompt builder
    notes.go                # Notes file loader
    *_test.go               # Unit tests
  reviewer/
    types.go                # Reviewer interface, Config, Result, ClaudeClient
    errors.go               # ReviewerError type
    prompt.go               # Reviewer prompt builder
    reviewer.go             # DefaultReviewer implementation
    *_test.go               # Unit tests (96% coverage)
  council/
    types.go                # Council interface, Config, Result, Decision types
    errors.go               # CouncilError type
    detector.go             # ConflictDetector (regex pattern matching)
    prompt.go               # PromptBuilder (R10 3-step resolution protocol)
    logger.go               # DecisionLogger (.claude/principles-decisions.log)
    council.go              # DefaultCouncil implementation
    *_test.go               # Unit tests (93% coverage)
  prd/
    types.go                # PRD extended type, Config, Generator interface
    errors.go               # PRDError, ValidationError types
    parser.go               # Claude output parsing (regex-based section extraction)
    validator.go            # PRD validation (goals, requirements, success criteria)
    generator.go            # DefaultGenerator implementation
    *_test.go               # Unit tests (100% coverage)
  planner/
    types.go                # Plan, PRD, Architecture, TaskGraph, Task types
    errors.go               # PlannerError type
    prompt.go               # PromptBuilder for planning phases
    runner.go               # PhaseRunner (orchestrates PRD → Architecture → Tasks)
    phase.go                # PlanningPhase interface
    persistence.go          # FilePersistence (atomic writes, YAML serialization)
    *_test.go               # Unit tests
  update/
    types.go                # Core types, options (DownloaderOptions, InstallerOptions)
    errors.go               # UpdateError, VersionError, ChecksumError
    version.go              # Semantic version parsing and comparison
    checker.go              # GitHub release query via `gh` CLI
    downloader.go           # Binary download, tar.gz/zip extraction, checksum
    installer.go            # Binary replacement, restart, rollback
    manager.go              # Update flow orchestration
    *_test.go               # Unit tests (71% coverage)
docs/
  CLI_CONTRACT.md           # CLI interface contract
  FEATURE_MATRIX.md         # Implementation status tracking
  PRINCIPLES_SCHEMA.md      # principles.yaml schema
.github/
  workflows/
    ci.yml                  # CI workflow (test, lint, build)
    release.yml             # Release workflow (goreleaser)
.goreleaser.yaml            # Multi-platform release configuration
install.sh                  # Binary installer script
CONTRIBUTING.md             # Contributor guidelines
test/
  mocks/
    shared_mocks.go         # ConfigurableClaudeClient, SequentialClaudeClient
    mock_command_executor.go # CommandSequence for CLI command mocking
  testutil/
    helpers.go              # Test helper functions (GetFixturePath, WriteFile)
    git_repo.go             # Git repo setup/teardown helpers
  fixtures/
    principles/             # Test fixtures for config loading
      valid.yaml            # Valid principles file
      invalid_syntax.yaml   # YAML syntax error fixture
      invalid_schema.yaml   # Schema validation failure fixture
  integration/
    cli_integration_test.go       # CLI flag parsing integration tests
    executor_integration_test.go  # Executor + Reviewer integration tests
    workflow_integration_test.go  # GitHub workflow integration tests
    config_integration_test.go    # Config loading integration tests
    council_integration_test.go   # Council conflict detection integration tests
  e2e/
    e2e_test.go             # E2E tests (binary build + execution)
  golden/                   # Output snapshots for regression
    help.txt
    version.txt
```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/claude-loop` |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make test-e2e` | Run E2E tests (builds binary first) |
| `make test-all` | Run all tests (unit, integration, E2E, golden) |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run golangci-lint |
| `make golden-test` | Compare output with golden files |
| `make clean` | Remove build artifacts |

## Coding Conventions

- Standard Go project layout (`cmd/`, `internal/`)
- Version injected via ldflags: `-X .../version.Version=$(VERSION)`
- Error handling: return errors, log at top level
- No `panic` in library code
- Tests in same package with `_test.go` suffix

## Testing Strategy

- **Unit tests**: `internal/**/*_test.go` - 95%+ coverage per package
- **Integration tests**: `test/integration/` - Cross-package interaction tests
- **E2E tests**: `test/e2e/` - Full binary execution tests
- **Golden tests**: `test/golden/` - Output snapshot regression tests
- **Mocks**: Custom mock implementations in `test/mocks/` (not testify/mock)
