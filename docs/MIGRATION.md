# Migration Guide: Bash to Go

This document provides guidance for users migrating from the bash version of `claude-loop` to the Go version.

## Overview

The Go version is a complete rewrite of the original 2,600+ line bash script, providing:

- **Better performance**: Compiled binary, no shell startup overhead
- **Type safety**: Compile-time error detection
- **Cross-platform**: Single binary for macOS, Linux, and Windows
- **Easier installation**: No external dependencies (jq, awk)
- **Better testing**: Comprehensive unit, integration, and E2E tests

## Breaking Changes

### None

The Go version maintains **100% CLI compatibility** with the bash version. All flags, options, and behaviors are preserved.

## Installation Changes

### Old (Bash)

```bash
# Download the script
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/claude_loop.sh -o claude-loop
chmod +x claude-loop
sudo mv claude-loop /usr/local/bin/
```

### New (Go Binary)

```bash
# Quick install (recommended)
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/install.sh | bash

# Or with go install
go install github.com/DeukWoongWoo/claude-loop/cmd/claude-loop@latest
```

## Dependency Changes

| Dependency | Bash | Go | Purpose |
|------------|------|-----|---------|
| `claude` | Required | Required | Claude Code CLI |
| `gh` | Required | Required* | GitHub CLI |
| `git` | Required | Required* | Version control |
| `jq` | Required | Not needed | JSON parsing (native in Go) |
| `awk`, `sed` | Required | Not needed | Text processing |
| `curl` | Required | Install only** | HTTP requests |

*Optional with `--disable-commits` flag
**`curl` is only needed for the install script, not at runtime

## Behavior Differences

### Performance

The Go version is faster due to compiled binary execution, native JSON parsing, and no shell startup overhead.

### Error Messages

Error messages may differ slightly in wording, but convey the same information:

```bash
# Bash version
Error: Missing required argument: --prompt

# Go version
Error: required flag(s) "prompt" not set
```

### Update Mechanism

| Version | Mechanism |
|---------|-----------|
| Bash | Downloads and replaces script file |
| Go | Downloads platform-specific binary with checksum verification |

## Flag Compatibility

All 28 CLI flags are 100% compatible between versions. See [CLI_CONTRACT.md](./CLI_CONTRACT.md) for the complete flag reference.

## Configuration Compatibility

### principles.yaml

The YAML schema is identical between versions:

```yaml
version: "2.3"
preset: startup
created_at: "2026-01-11"

layer0:
  trust_architecture: 3
  curation_model: 2
  # ... (same structure)

layer1:
  speed_correctness: 7
  # ... (same structure)
```

### Decision Log

The decision log format is identical:

```
[2026-01-11T10:30:00Z] DECISION: feature_scope
  Conflict: scope_philosophy (4) vs cost_efficiency (6)
  Resolution: R3 - Prioritize cost efficiency
  Rationale: MVP phase
```

## Exit Codes

Exit codes are compatible:

| Code | Bash | Go | Description |
|------|------|-----|-------------|
| 0 | Yes | Yes | Success |
| 1 | Yes | Yes | Error |

## Migration Steps

### 1. Backup Current Installation

```bash
# If you have the bash version installed
which claude-loop
cp $(which claude-loop) ~/claude-loop-bash-backup
```

### 2. Install Go Version

```bash
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/install.sh | bash
```

### 3. Verify Installation

```bash
claude-loop --version
# Should show: claude-loop version X.Y.Z
```

### 4. Test with Dry Run

```bash
claude-loop --prompt "test" --max-runs 1 --dry-run
```

### 5. Verify Installation

```bash
claude-loop --version
claude-loop --help
```

## Feature Matrix

For detailed feature-by-feature comparison, see [FEATURE_MATRIX.md](./FEATURE_MATRIX.md).

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/DeukWoongWoo/claude-loop/issues)
- **Documentation**: [README.md](../README.md)
- **CLI Reference**: [CLI_CONTRACT.md](./CLI_CONTRACT.md)
