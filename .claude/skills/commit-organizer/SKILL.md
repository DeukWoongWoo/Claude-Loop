---
name: commit-organizer
description: Git commit workflow. Analyzes, organizes, groups, and documents git changes into logical commits with MR descriptions. Use for any commit organization-related task.
---

# Commit Organizer

Analyze git changes, organize into logical commits, review code quality, and generate MR description.

## Workflow

1. Analyze current git status and diff
2. Read all changed/new files to understand changes
3. Group changes into logical commits
4. Review code quality against project patterns
5. Generate commit commands and MR description

## Analysis Steps

### Step 1: Gather Git Information

```bash
git status --porcelain
git diff --stat HEAD
git diff HEAD -- <modified files>
```

### Step 2: Read Changed Files

- Read all new files completely
- Review diffs for modified files
- Check test files for coverage

### Step 3: Group Changes Logically

Common grouping patterns (in commit order):
1. **Dependencies** - package.json, pyproject.toml, lock files
2. **Infrastructure** - DB clients, config, utilities
3. **Core Features** - new APIs, services, schemas
4. **Documentation** - README, CLAUDE.md, pattern docs

### Step 4: Code Quality Review

Check against project patterns:
- [ ] Follows API endpoint pattern (@documents/claude/patterns/api-endpoint.md)
- [ ] Proper error handling with standard format
- [ ] Type hints and docstrings
- [ ] Tests included for new features
- [ ] No security vulnerabilities (SQL injection, etc.)

## Output Format

### Commit Commands

```markdown
### Commit 1: <short description>
```bash
git add <files> && git commit -m "<type>: <message>"
```

### Commit 2: <short description>
```bash
git add <files> && git commit -m "<type>: <message>"
```
```

### MR Description

```markdown
## Summary

- Bullet point summary of all changes

## Changes

### <Category>
- Detailed change descriptions

## Test Plan

- [ ] Test items with checkboxes
```

## Commit Message Format

**CRITICAL: Use single-line commits only. Do NOT add body or trailers.**

Use conventional commits with `-m "message"` format:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

**Rules:**
- One line only, under 72 characters
- Use simple `-m "message"` format, NOT HEREDOC
- Do NOT add commit body
- Do NOT add "Generated with Claude Code" trailer
- Do NOT add "Co-Authored-By" trailer

## Examples

### Correct vs Wrong Commit Format

**CORRECT - single line only:**
```bash
git add src/auth.ts && git commit -m "feat: Add user authentication"
```

**WRONG - do not use HEREDOC or multi-line:**
```bash
# DO NOT DO THIS
git commit -m "$(cat <<'EOF'
feat: Add user authentication

Some description here...

Generated with Claude Code

Co-Authored-By: ...
EOF
)"
```

### Good Commit Grouping

```markdown
### Commit 1: Dependencies
git add pyproject.toml uv.lock && git commit -m "feat: Add sqlglot and openai-agents dependencies"

### Commit 2: Infrastructure
git add alphaverse/server/db/postgresql_client.py && git commit -m "fix: Improve connection pool error handling"

### Commit 3: Feature
git add alphaverse/server/api/v1/*.py alphaverse/server/services/*.py tests/ && git commit -m "feat: Add Query Agent API"

### Commit 4: Documentation
git add CLAUDE.md documents/ && git commit -m "docs: Update API patterns documentation"
```

### Good MR Description

```markdown
## Summary

- Add Query Agent API for natural language to SQL conversion
- Improve PostgreSQL connection handling
- Update documentation patterns

## Changes

### New Features
- **Query Agent API** (`/api/v1/agents/query`)
  - Natural language to SQL using OpenAI
  - SQL validation with sqlglot

### Infrastructure
- PostgreSQL client: TCP keepalive, dead connection handling

## Test Plan

- [ ] Run `pytest tests/`
- [ ] Test API endpoints manually
- [ ] Verify SQL injection prevention
```

## Quality Checklist

Before finalizing:
- [ ] Commits are in logical dependency order
- [ ] Each commit is atomic and can be reverted independently
- [ ] No commit includes unrelated changes
- [ ] Commit messages follow conventional format
- [ ] MR description covers all changes
- [ ] Test plan is actionable
