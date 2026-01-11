# claude-loop

> Autonomous AI development loop with 4-Layer Principles Framework

Based on [continuous-claude](https://github.com/AnandChowdhary/continuous-claude) by Anand Chowdhary.

## What's New

This project extends the original continuous-claude with:

- **4-Layer Principles Framework** - 18 configurable principles for autonomous decision-making
- **LLM Council Integration** - Multi-LLM conflict resolution (Claude, Gemini, Codex)
- **Decision Logging** - Track AI decisions with rationale
- **3 Presets** - startup, enterprise, opensource configurations
- **Auto-setup** - Automatic council file installation

## How it Works

Using Claude Code to drive iterative development, this script fully automates the PR lifecycle from code changes through to merged commits:

- Claude Code runs in a loop based on your prompt
- All changes are committed to a new branch
- A new pull request is created
- It waits for all required PR checks and code reviews to complete
- Once checks pass and reviews are approved, the PR is merged
- This process repeats until your task is complete
- A `SHARED_TASK_NOTES.md` file maintains continuity by passing context between iterations
- **NEW**: On first run, configure project principles through a 3-step questionnaire
- **NEW**: Claude uses these principles for autonomous decision-making during iterations

## Quick Start

### Installation

```bash
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/install.sh | bash
```

This will:
- Install `claude-loop` to `~/.local/bin`
- Check for required dependencies
- Guide you through adding it to your PATH if needed

### Manual Installation

```bash
# Download the script
curl -fsSL https://raw.githubusercontent.com/DeukWoongWoo/claude-loop/main/claude_loop.sh -o claude-loop

# Make it executable
chmod +x claude-loop

# Move to a directory in your PATH
sudo mv claude-loop /usr/local/bin/
```

### Prerequisites

Before using `claude-loop`, you need:

1. **[Claude Code CLI](https://claude.ai/code)** - Authenticate with `claude auth`
2. **[GitHub CLI](https://cli.github.com)** - Authenticate with `gh auth login`
3. **jq** - Install with `brew install jq` (macOS) or `apt-get install jq` (Linux)

### Usage

```bash
# Run with your prompt, max runs, and GitHub repo (owner and repo auto-detected from git remote)
claude-loop --prompt "add unit tests until all code is covered" --max-runs 5

# Or explicitly specify the owner and repo
claude-loop --prompt "add unit tests until all code is covered" --max-runs 5 --owner MyUser --repo my-project

# Or run with a cost budget instead
claude-loop --prompt "add unit tests until all code is covered" --max-cost 10.00

# Or run for a specific duration (time-boxed bursts)
claude-loop --prompt "add unit tests until all code is covered" --max-duration 2h
```

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

## LLM Council

When Claude encounters unresolvable principle conflicts, it automatically invokes the LLM Council:

- Queries Claude, Gemini, and Codex for opinions
- Synthesizes a consensus recommendation
- Logs the council decision

Council files are auto-downloaded on first run from GitHub.

## Flags

### Required (one of these)

- `-p, --prompt <text>`: Task prompt for Claude Code
- `-m, --max-runs <number>`: Maximum number of successful iterations (use 0 for unlimited)
- `--max-cost <dollars>`: Maximum cost in USD to spend
- `--max-duration <duration>`: Maximum duration to run (e.g., `2h`, `30m`, `1h30m`)

### Optional

| Flag | Description | Default |
|------|-------------|---------|
| `--owner <owner>` | GitHub repository owner | auto-detected |
| `--repo <repo>` | GitHub repository name | auto-detected |
| `--disable-commits` | Disable automatic commits and PR creation | false |
| `--disable-branches` | Commit on current branch without PRs | false |
| `--git-branch-prefix <prefix>` | Branch prefix for iterations | `claude-loop/` |
| `--merge-strategy <strategy>` | PR merge strategy: squash, merge, rebase | `squash` |
| `--notes-file <file>` | Shared notes file | `SHARED_TASK_NOTES.md` |
| `--worktree <name>` | Run in a git worktree for parallel execution | - |
| `--worktree-base-dir <path>` | Base directory for worktrees | `../claude-loop-worktrees` |
| `--cleanup-worktree` | Remove worktree after completion | false |
| `--list-worktrees` | List all active git worktrees and exit | - |
| `--dry-run` | Simulate execution without making changes | false |
| `--completion-signal <phrase>` | Phrase for project completion | `CONTINUOUS_CLAUDE_PROJECT_COMPLETE` |
| `--completion-threshold <num>` | Consecutive signals to stop early | 3 |
| `-r, --review-prompt <text>` | Reviewer pass after each iteration | - |
| `--disable-ci-retry` | Disable automatic CI failure retry | false |
| `--ci-retry-max <number>` | Maximum CI fix attempts per PR | 1 |

### Principles Framework Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--reset-principles` | Force re-collection of principles | false |
| `--principles-file <path>` | Custom principles file path | `.claude/principles.yaml` |
| `--log-decisions` | Enable decision logging | false |

## Examples

```bash
# Run 5 iterations to fix bugs
claude-loop -p "Fix all linter errors" -m 5

# Run with cost limit
claude-loop -p "Add tests" --max-cost 10.00

# Run for 2 hours (time-boxed)
claude-loop -p "Add documentation" --max-duration 2h

# Combine duration and cost limits
claude-loop -p "Improve tests" --max-duration 1h --max-cost 5.00

# Test without commits
claude-loop -p "Refactor code" -m 3 --disable-commits

# Use custom branch prefix and merge strategy
claude-loop -p "Feature work" -m 10 --git-branch-prefix "ai/" --merge-strategy merge

# Run in parallel with worktrees
claude-loop -p "Add unit tests" -m 5 --worktree task-a
claude-loop -p "Add docs" -m 5 --worktree task-b

# Use a reviewer to validate changes
claude-loop -p "Add new feature" -m 5 -r "Run npm test and npm run lint, fix any failures"

# Force re-collection of principles
claude-loop -p "New project" -m 5 --reset-principles

# Use custom principles file
claude-loop -p "Feature work" -m 5 --principles-file custom-principles.yaml

# Enable decision logging
claude-loop -p "Complex task" -m 10 --log-decisions
```

### Running in Parallel

Use git worktrees to run multiple instances simultaneously:

```bash
# Terminal 1
claude-loop -p "Add unit tests" -m 5 --worktree tests

# Terminal 2 (simultaneously)
claude-loop -p "Add docs" -m 5 --worktree docs
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
