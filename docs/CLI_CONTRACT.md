# Claude Loop CLI Contract v0.18.0

> This document defines the CLI interface contract for `claude_loop.sh`.
> Go migration must maintain 100% compatibility with this contract.

## Usage

```
claude-loop -p "prompt" (-m max-runs | --max-cost max-cost | --max-duration duration) [options]
claude-loop update
```

---

## CLI Flags (28 flags)

### Required Options (at least one limit required)

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--prompt` | `-p` | string | - | The prompt/goal for Claude Code to work on |
| `--max-runs` | `-m` | int | - | Maximum number of successful iterations (use 0 for unlimited with cost/duration) |
| `--max-cost` | - | float | - | Maximum cost in USD to spend |
| `--max-duration` | - | duration | - | Maximum duration to run (e.g., "2h", "30m", "1h30m", "90s") |

### GitHub Configuration

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--owner` | - | string | auto-detect | GitHub repository owner |
| `--repo` | - | string | auto-detect | GitHub repository name |

### Commit & Branch Management

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--disable-commits` | - | bool | false | Disable automatic commits and PR creation |
| `--disable-branches` | - | bool | false | Commit on current branch without creating branches or PRs |
| `--git-branch-prefix` | - | string | "claude-loop/" | Branch prefix for iterations |
| `--merge-strategy` | - | string | "squash" | PR merge strategy: squash, merge, or rebase |

### Iteration Control

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--completion-signal` | - | string | "CONTINUOUS_CLAUDE_PROJECT_COMPLETE" | Phrase that agents output when project is complete |
| `--completion-threshold` | - | int | 3 | Number of consecutive signals to stop early |
| `--dry-run` | - | bool | false | Simulate execution without making changes |

### Review & CI

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--review-prompt` | `-r` | string | - | Run a reviewer pass after each iteration to validate changes |
| `--disable-ci-retry` | - | bool | false | Disable automatic CI failure retry (enabled by default) |
| `--ci-retry-max` | - | int | 1 | Maximum CI fix attempts per PR |

### Shared State

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--notes-file` | - | string | "SHARED_TASK_NOTES.md" | Shared notes file for iteration context |

### Worktree Support

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--worktree` | - | string | - | Run in a git worktree for parallel execution |
| `--worktree-base-dir` | - | string | "../claude-loop-worktrees" | Base directory for worktrees |
| `--cleanup-worktree` | - | bool | false | Remove worktree after completion |
| `--list-worktrees` | - | bool | false | List all active git worktrees and exit |

### Principles Framework

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--reset-principles` | - | bool | false | Force re-collection of principles |
| `--principles-file` | - | string | ".claude/principles.yaml" | Custom principles file path |
| `--log-decisions` | - | bool | false | Enable decision logging to .claude/principles-decisions.log |

### Update Management

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--help` | `-h` | bool | false | Show help message |
| `--version` | `-v` | bool | false | Show version information |
| `--auto-update` | - | bool | false | Automatically install updates when available |
| `--disable-updates` | - | bool | false | Skip all update checks and prompts |

---

## Commands

| Command | Description |
|---------|-------------|
| `update` | Check for and install the latest version |

---

## Exit Codes

| Code | Name | Description |
|------|------|-------------|
| 0 | SUCCESS | Successful completion |
| 1 | ERROR | General error (invalid args, validation failure, runtime error) |

### Exit Code Details

- `exit 0`: Normal completion, help displayed, version displayed, update successful
- `exit 1`:
  - Missing required arguments (prompt, limits)
  - Invalid argument values (merge strategy, duration format)
  - Missing dependencies (jq, claude, gh)
  - GitHub repository detection failure
  - 3+ consecutive iteration errors
  - CI retry failure
  - Worktree operation failure

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | No | Used by `gh` CLI (auto-managed by `gh auth login`) |
| `ANTHROPIC_API_KEY` | No | Used by Claude CLI (auto-managed by `claude` CLI) |

---

## Duration Format

Supported formats for `--max-duration`:

| Format | Example | Description |
|--------|---------|-------------|
| Hours | `2h` | 2 hours |
| Minutes | `30m` | 30 minutes |
| Seconds | `90s` | 90 seconds |
| Combined | `1h30m` | 1 hour 30 minutes |

---

## Dependencies

| Tool | Purpose | Required |
|------|---------|----------|
| `claude` | Claude Code CLI | Yes |
| `gh` | GitHub CLI | Yes (unless --disable-commits) |
| `git` | Version control | Yes (unless --disable-commits) |

> **Note**: The Go version has native JSON parsing and does not require `jq`.

---

## Configuration Files

### principles.yaml

Location: `.claude/principles.yaml` (or custom path via `--principles-file`)

See [PRINCIPLES_SCHEMA.md](./PRINCIPLES_SCHEMA.md) for full schema.

### Decision Log

Location: `.claude/principles-decisions.log` (when `--log-decisions` enabled)

---

## Flag Forwarding

Unknown flags are forwarded to the `claude` CLI:

```bash
# These flags are forwarded to claude:
claude-loop -p "prompt" -m 5 --model opus
#                            ^^^^^^^^^^^ forwarded
```

---

## Validation Rules

1. **Prompt required**: `-p` or `--prompt` must be provided
2. **Limit required**: At least one of `--max-runs`, `--max-cost`, or `--max-duration`
3. **Zero max-runs**: If `--max-runs 0`, must have `--max-cost` or `--max-duration`
4. **GitHub required**: Unless `--disable-commits`, must have valid GitHub repo (auto-detect or explicit)
5. **Merge strategy**: Must be one of: `squash`, `merge`, `rebase`
6. **Duration format**: Must match pattern: `(\d+h)?(\d+m)?(\d+s)?`

---

## Output Behavior

- **Iteration prefix**: Each Claude output line prefixed with `[Iteration N]`
- **Status updates**: PR check polling shows status changes only
- **Cost tracking**: Cumulative USD displayed after each iteration
- **Completion signal**: Detected and counted per iteration

---

## Internal State

These are internal variables maintained during execution:

| Variable | Type | Description |
|----------|------|-------------|
| `successful_iterations` | int | Count of completed iterations |
| `error_count` | int | Consecutive error counter (reset on success) |
| `completion_signal_count` | int | Consecutive completion signals |
| `total_cost` | float | Accumulated USD cost |
| `start_time` | timestamp | Loop start time |
