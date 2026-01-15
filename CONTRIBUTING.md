# Contributing to Claude Loop

Thank you for your interest in contributing to Claude Loop! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Commit Message Format](#commit-message-format)

## Development Setup

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Build and run |
| Git | Any | Version control |
| Make | Any | Build automation |
| [golangci-lint](https://golangci-lint.run/usage/install/) | Any | Linting |

### Getting Started

```bash
# Clone the repository
git clone https://github.com/DeukWoongWoo/claude-loop.git
cd claude-loop

# Download dependencies
make deps

# Build the binary
make build

# Run tests
make test
```

### Build Commands

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

## Project Structure

```
cmd/claude-loop/main.go     # Entry point
internal/
  cli/                      # CLI framework (Cobra)
  config/                   # Configuration and principles loading
  loop/                     # Main execution loop
  claude/                   # Claude Code subprocess wrapper
  git/                      # Git operations
  github/                   # GitHub PR and CI integration
  prompt/                   # Prompt building
  reviewer/                 # Reviewer pass implementation
  council/                  # LLM Council for conflict resolution
  update/                   # Auto-update mechanism
  version/                  # Version information
docs/                       # Documentation
test/
  mocks/                    # Custom mock implementations
  testutil/                 # Test helper functions
  fixtures/                 # Test data files
  integration/              # Integration tests
  e2e/                      # End-to-end tests
  golden/                   # Golden file snapshots
```

## Code Style

### General Guidelines

- Follow standard Go project layout (`cmd/`, `internal/`)
- Use `gofmt` for formatting (automatic with most editors)
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Keep functions focused and small
- Use meaningful variable and function names

### Error Handling

- Return errors from functions; don't panic in library code
- Log errors at the top level (main or command handlers)
- Use custom error types for specific error categories (see `**/errors.go` files)

```go
// Good
func DoSomething() error {
    if err := operation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    return nil
}

// Bad - Don't panic in library code
func DoSomething() {
    if err := operation(); err != nil {
        panic(err)  // Don't do this
    }
}
```

### Version Injection

Version information is injected via ldflags at build time:

```go
// internal/version/version.go
var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)
```

## Testing

### Requirements

| Requirement | Details |
|-------------|---------|
| Coverage | 95%+ per package |
| Location | Same package with `_test.go` suffix |
| Mocks | Custom implementations in `test/mocks/` (not testify/mock) |

### Test Types

| Type | Location | Purpose |
|------|----------|---------|
| Unit | `internal/**/*_test.go` | Test individual functions/components |
| Integration | `test/integration/` | Test cross-package interactions |
| E2E | `test/e2e/` | Test full binary execution |
| Golden | `test/golden/` | Snapshot regression tests |

### Running Tests

```bash
# Run unit tests only
make test

# Run integration tests
make test-integration

# Run E2E tests (builds binary first)
make test-e2e

# Run all tests
make test-all

# Run with coverage report
make test-coverage
```

### Writing Tests

```go
func TestFeature(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"

    // Act
    result := Feature(input)

    // Assert
    if result != expected {
        t.Errorf("Feature(%q) = %q, want %q", input, result, expected)
    }
}
```

For table-driven tests:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"normal input", "hello", "HELLO"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Feature(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Using Mocks

Custom mocks are located in `test/mocks/`. Example usage:

```go
import "github.com/DeukWoongWoo/claude-loop/test/mocks"

func TestWithMock(t *testing.T) {
    mockClient := &mocks.ConfigurableClaudeClient{
        RunFunc: func(ctx context.Context, prompt string) (*claude.Result, error) {
            return &claude.Result{Output: "mocked output"}, nil
        },
    }

    // Use mockClient in your test
}
```

## Pull Request Process

### Before Submitting

1. **Run all tests**: `make test-all`
2. **Run linter**: `make lint`
3. **Update documentation** if needed
4. **Add tests** for new functionality

### PR Guidelines

- Create a feature branch from `main`
- Keep PRs focused on a single feature or fix
- Include a clear description of changes
- Reference related issues (e.g., "Fixes #123")
- Ensure CI passes before requesting review

### Branch Naming

```
feature/description    # New features
fix/description        # Bug fixes
docs/description       # Documentation updates
refactor/description   # Code refactoring
test/description       # Test additions/improvements
```

## Commit Message Format

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Code style changes (formatting, etc.) |
| `refactor` | Code refactoring |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks |

### Examples

```
feat(cli): add --dry-run flag for simulation mode

fix(github): handle rate limiting in PR creation

docs: update installation instructions

test(loop): add integration tests for executor

refactor(config): simplify principles loading logic
```

### Co-Authored Commits

When using AI assistance, include co-author attribution:

```
feat(cli): add new feature

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Questions?

If you have questions or need help:

1. Check existing [issues](https://github.com/DeukWoongWoo/claude-loop/issues)
2. Open a new issue with your question
3. Review the [documentation](docs/)

Thank you for contributing!
