# claude-loop

> Autonomous AI development loop with 4-Layer Principles Framework

Based on [continuous-claude](https://github.com/AnandChowdhary/continuous-claude) by Anand Chowdhary.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [CLI Reference](#cli-reference)
- [Principles Framework](#principles-framework)
- [LLM Council](#llm-council)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Features

- **4-Layer Principles Framework** - 18 configurable principles for autonomous decision-making
- **LLM Council Integration** - AI-powered conflict resolution for principle conflicts
- **Decision Logging** - Track AI decisions with rationale
- **3 Presets** - startup, enterprise, opensource configurations
- **Auto-setup** - Automatic council file installation
- **Parallel Execution** - Git worktree support for concurrent tasks
- **CI Integration** - Automatic PR creation, CI monitoring, and failure retry

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/install.sh | bash
```

This will:
- Detect your platform (macOS, Linux, Windows via WSL/Git Bash)
- Download the appropriate binary
- Verify SHA256 checksum
- Install to `~/.local/bin`
- Guide you through PATH configuration

> **Note**: Windows users need WSL, Git Bash, or MSYS2 to run the install script.

### Go Install

```bash
go install github.com/DeukWoongWoo/claude-loop/cmd/claude-loop@latest
```

### Manual Installation

Download the latest release from [GitHub Releases](https://github.com/DeukWoongWoo/claude-loop/releases):

```bash
# macOS (Apple Silicon)
curl -fsSL https://github.com/DeukWoongWoo/claude-loop/releases/latest/download/claude-loop_darwin_arm64.tar.gz | tar xz
sudo mv claude-loop /usr/local/bin/

# macOS (Intel)
curl -fsSL https://github.com/DeukWoongWoo/claude-loop/releases/latest/download/claude-loop_darwin_amd64.tar.gz | tar xz
sudo mv claude-loop /usr/local/bin/

# Linux (x86_64)
curl -fsSL https://github.com/DeukWoongWoo/claude-loop/releases/latest/download/claude-loop_linux_amd64.tar.gz | tar xz
sudo mv claude-loop /usr/local/bin/

# Linux (ARM64)
curl -fsSL https://github.com/DeukWoongWoo/claude-loop/releases/latest/download/claude-loop_linux_arm64.tar.gz | tar xz
sudo mv claude-loop /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/DeukWoongWoo/claude-loop.git
cd claude-loop
make build
# Binary will be at ./bin/claude-loop
```

## Prerequisites

Before using `claude-loop`, you need:

| Tool | Purpose | Installation |
|------|---------|--------------|
| [Claude Code CLI](https://claude.ai/code) | AI code generation | `claude auth` to authenticate |
| [GitHub CLI](https://cli.github.com) | PR and CI management | `gh auth login` to authenticate |
| Git | Version control | Pre-installed on most systems |

## Quick Start

### How It Works

This tool automates the PR lifecycle using Claude Code for iterative development:

1. Run Claude Code in a loop based on your prompt
2. Commit changes to a new branch and create a PR
3. Wait for CI checks and code reviews to pass
4. Merge the PR and repeat until complete
5. Maintain context via `SHARED_TASK_NOTES.md` between iterations
6. Configure project principles on first run via questionnaire

### Basic Usage

```bash
# Run with iteration limit
claude-loop --prompt "add unit tests until all code is covered" --max-runs 5

# Run with cost budget
claude-loop --prompt "add unit tests" --max-cost 10.00

# Run for a specific duration (time-boxed)
claude-loop --prompt "add documentation" --max-duration 2h

# Combine limits (stops when first limit is reached)
claude-loop --prompt "improve tests" --max-duration 1h --max-cost 5.00
```

## CLI Reference

### Required Options

At least one limit is required:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--prompt` | `-p` | string | The prompt/goal for Claude Code |
| `--max-runs` | `-m` | int | Maximum iterations (0 = unlimited with cost/duration) |
| `--max-cost` | | float | Maximum cost in USD |
| `--max-duration` | | duration | Maximum duration (e.g., `2h`, `30m`, `1h30m`) |

### GitHub Configuration

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--owner` | string | auto-detect | GitHub repository owner |
| `--repo` | string | auto-detect | GitHub repository name |

### Commit & Branch Management

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--disable-commits` | bool | false | Disable automatic commits and PR creation |
| `--disable-branches` | bool | false | Commit on current branch without PRs |
| `--git-branch-prefix` | string | `claude-loop/` | Branch prefix for iterations |
| `--merge-strategy` | string | `squash` | PR merge strategy: squash, merge, rebase |

### Iteration Control

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--completion-signal` | string | `CONTINUOUS_CLAUDE_PROJECT_COMPLETE` | Phrase for project completion |
| `--completion-threshold` | int | 3 | Consecutive signals to stop early |
| `--dry-run` | bool | false | Simulate execution without changes |

### Review & CI

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--review-prompt` | `-r` | string | | Reviewer pass after each iteration |
| `--disable-ci-retry` | | bool | false | Disable automatic CI failure retry |
| `--ci-retry-max` | | int | 1 | Maximum CI fix attempts per PR |

### Shared State

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--notes-file` | string | `SHARED_TASK_NOTES.md` | Shared notes file for context |

### Worktree Support

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--worktree` | string | | Run in a git worktree for parallel execution |
| `--worktree-base-dir` | string | `../claude-loop-worktrees` | Base directory for worktrees |
| `--cleanup-worktree` | bool | false | Remove worktree after completion |
| `--list-worktrees` | bool | false | List all active worktrees and exit |

### Principles Framework

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--reset-principles` | bool | false | Force re-collection of principles |
| `--principles-file` | string | `.claude/principles.yaml` | Custom principles file path |
| `--log-decisions` | bool | false | Enable decision logging |

### Update Management

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--help` | `-h` | bool | Show help message |
| `--version` | `-v` | bool | Show version information |
| `--auto-update` | | bool | Auto-install updates when available |
| `--disable-updates` | | bool | Skip all update checks |

## Principles Framework

### How It Works

On first run, you'll be asked to configure project principles through a 3-step questionnaire:

1. **Select a preset** (startup/enterprise/opensource)
2. **Customize key principles** (optional adjustments)
3. **Generate** `.claude/principles.yaml`

Claude then uses these principles for autonomous decision-making during iterations.

### Presets

| Preset | Use Case | Key Characteristics |
|--------|----------|---------------------|
| **startup** | MVP/Fast validation | Speed > Correctness, Small blast radius |
| **enterprise** | Stability-focused | Security, Compliance, Large teams |
| **opensource** | Community projects | Transparency, Community-first |

### The 18 Principles (4 Layers)

**Layer 0 - Product Principles:**
- Trust Architecture, Curation Model, Scope Philosophy
- Monetization Model, Privacy Posture, UX Philosophy
- Authority Stance, Auditability, Interoperability

**Layer 1 - Development Principles:**
- Speed vs Correctness, Innovation vs Stability, Blast Radius
- Clarity of Intent, Reversibility Priority, Security Posture
- Urgency Tiers, Cost Efficiency

**Layer 2 - Resolution Rules (R1-R10):**
- Automatic conflict resolution between principles

**Layer 3 - Escalation:**
- LLM Council invocation for unresolvable conflicts

See [docs/PRINCIPLES_SCHEMA.md](docs/PRINCIPLES_SCHEMA.md) for the full schema specification.

## LLM Council

When Claude encounters unresolvable principle conflicts, it automatically invokes the LLM Council:

- Analyzes the conflict using the R10 3-step resolution protocol
- Synthesizes a consensus recommendation
- Logs the council decision with rationale

Council files are auto-downloaded on first run from GitHub.

## Configuration

### principles.yaml

Location: `.claude/principles.yaml` (or custom path via `--principles-file`)

```yaml
version: "2.3"
preset: startup
created_at: "2026-01-11"

layer0:
  trust_architecture: 3
  curation_model: 2
  scope_philosophy: 4
  # ... (1-10 scale for each principle)

layer1:
  speed_correctness: 7
  innovation_stability: 6
  blast_radius: 8
  # ... (1-10 scale for each principle)
```

See [docs/PRINCIPLES_SCHEMA.md](docs/PRINCIPLES_SCHEMA.md) for the full schema.

### Decision Log

Location: `.claude/principles-decisions.log` (when `--log-decisions` enabled)

Format:
```
[2026-01-11T10:30:00Z] DECISION: feature_scope
  Conflict: scope_philosophy (4) vs cost_efficiency (6)
  Resolution: R3 - Prioritize cost efficiency for non-critical features
  Rationale: MVP phase, limited budget
```

## Examples

### Branch and Merge Control

```bash
# Use custom branch prefix and merge strategy
claude-loop -p "Feature work" -m 10 --git-branch-prefix "ai/" --merge-strategy merge

# Commit on current branch without PRs
claude-loop -p "Quick fix" -m 3 --disable-branches
```

### Parallel Execution with Worktrees

```bash
# Terminal 1
claude-loop -p "Add unit tests" -m 5 --worktree tests

# Terminal 2 (simultaneously)
claude-loop -p "Add docs" -m 5 --worktree docs

# List all active worktrees
claude-loop --list-worktrees

# Cleanup worktree after completion
claude-loop -p "Task" -m 5 --worktree temp --cleanup-worktree
```

### Reviewer Pass

```bash
# Use a reviewer to validate changes
claude-loop -p "Add new feature" -m 5 -r "Run npm test and npm run lint, fix any failures"
```

### Principles Framework

```bash
# Force re-collection of principles
claude-loop -p "New project" -m 5 --reset-principles

# Use custom principles file
claude-loop -p "Feature work" -m 5 --principles-file custom-principles.yaml

# Enable decision logging
claude-loop -p "Complex task" -m 10 --log-decisions
```

## Troubleshooting

### Common Issues

#### "Claude CLI not found"

```bash
# Install Claude Code CLI
# Visit: https://claude.ai/code
# Then authenticate:
claude auth
```

#### "GitHub CLI not authenticated"

```bash
# Authenticate with GitHub
gh auth login
```

#### "Repository not detected"

```bash
# Make sure you're in a git repository with a remote
git remote -v

# Or specify explicitly
claude-loop -p "task" -m 5 --owner MyUser --repo my-project
```

#### "PATH not configured"

After installation, add to your shell profile:

```bash
# For zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc

# For bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (invalid args, runtime error, CI failure) |

### Getting Help

```bash
# Show help
claude-loop --help

# Show version
claude-loop --version

# Check for updates
claude-loop update
```

## Example Output

```
🔄 (1/1) Starting iteration...
📋 Principles loaded [startup]
   Scope=3 Speed=4 Security=7 Blast=7
🌿 (1/1) Creating branch: claude-loop/iteration-1/2026-01-11-be939873
🤖 (1/1) Running Claude Code...
📝 (1/1) Output: Successfully completed the task...
💰 (1/1) Cost: $0.042
✅ (1/1) Work completed
📝 Decision logged to .claude/principles-decisions.log
💬 (1/1) Committing changes...
📤 (1/1) Pushing branch...
🔨 (1/1) Creating pull request...
✅ (1/1) All PR checks passed
🔀 (1/1) Merging PR #123...
🎉 Done with total cost: $0.042
```

## License

[MIT](./LICENSE)

Originally created by [Anand Chowdhary](https://github.com/AnandChowdhary)
Extended by [DeukWoong Woo](https://github.com/DeukWoongWoo)
