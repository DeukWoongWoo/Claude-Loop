# CLAUDE.md Best Practices Reference

## Core Principles

### 1. Token Efficiency
CLAUDE.md is prepended to every prompt. Bloated files:
- Increase token costs
- Add noise that obscures important instructions
- Reduce Claude's ability to follow directions

**Target: Under 250 lines** (for complex monorepos like Alphaverse)

### 2. Write for Claude, Not Humans
Claude already knows general programming. Only include:
- Project-specific patterns
- Non-obvious architectural decisions
- Unexpected behaviors or warnings
- Repository-specific conventions

### 3. Declarative Over Narrative

| Style | Example |
|-------|---------|
| Good | `- Type hints required, use Python 3.10+ union syntax (X \| Y)` |
| Bad | `When writing functions, you should remember to add type hints because...` |

### 4. Reference Over Inline

| Approach | When to Use |
|----------|-------------|
| `@import` reference | Detailed patterns, step-by-step guides |
| Inline in CLAUDE.md | Quick rules, conventions, key locations |

## Required Sections

### Minimal CLAUDE.md Structure (Alphaverse)
```markdown
# Project Name
One-line description.

## Quick Reference
- Common commands
- Environment variables

## Architecture Overview
- Directory structure (abbreviated)
- Layer responsibilities table

## Code Conventions
- Style rules
- Naming conventions
- Import order

## Framework Development Patterns
- @import references to detailed patterns

## Project Guidelines
- Placement rules
- Database access patterns

## Key Files
- Entry points
- Configuration files

## References
- @import to usage guides and patterns
```

## Content Guidelines

### Include
- Exact dependency versions (Python 3.11, FastAPI 0.104.0)
- File path conventions with examples
- Pydantic schema patterns unique to this project
- Environment variable names
- Layer responsibilities and placement rules
- Database client usage (Snowflake, PostgreSQL, DuckDB)

### Exclude
- How Python/FastAPI/Pydantic works
- General best practices (SOLID, DRY, etc.)
- Information in official docs
- Detailed implementation patterns (use `@documents/claude/patterns/`)
- Complete Pydantic model definitions

## Hierarchical Organization

```
~/.claude/CLAUDE.md                    # Personal global preferences
./CLAUDE.md                            # Project root (commit to git)
./documents/claude/
  usage-guide.md                       # Scenario/asset developer guide
  patterns/
    api-endpoint.md                    # API endpoint pattern
    cli-command.md                     # CLI command pattern
    core-schema.md                     # Core schema pattern
```

Use `@import` syntax to reference:
```markdown
- **API 엔드포인트 추가**: @documents/claude/patterns/api-endpoint.md
```

## Anti-Patterns

| Anti-Pattern | Why It's Bad | Solution |
|--------------|--------------|----------|
| 500+ lines | Token waste, noise | Split into `@import` references |
| "You should consider..." | Narrative, not declarative | Convert to imperative bullets |
| Complete Pydantic models | Put in separate file | Reference schema files |
| General Python knowledge | Claude already knows | Remove |
| Inline step-by-step patterns | Too verbose | Move to `documents/claude/patterns/` |

## Maintenance

### Add to CLAUDE.md When
- You repeat the same instruction across sessions
- Claude makes the same mistake twice
- A pattern is project-specific and non-obvious

### Remove from CLAUDE.md When
- Information is outdated
- It duplicates official documentation
- Claude consistently follows it without reminder

### Move to Pattern Files When
- Content exceeds 20 lines of code examples
- It's a step-by-step implementation guide
- Multiple files need to be created/modified

## Validation Criteria

### Pass Criteria
- [ ] <= 250 lines
- [ ] All bullet points are declarative
- [ ] No general programming knowledge
- [ ] Versions specified for key dependencies
- [ ] File paths use clickable markdown links
- [ ] No narrative paragraphs > 2 sentences
- [ ] Detailed patterns use `@import` references

### Improvement Indicators
- Line count > 250: Move patterns to separate files
- "You should" phrases: Convert to imperatives
- Long code blocks (>10 lines): Move to pattern file
- Detailed explanations: Summarize or remove
- Step-by-step guides: Use `@import` reference
