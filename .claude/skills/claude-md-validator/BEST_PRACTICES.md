# CLAUDE.md Best Practices Reference

## Core Principles

### 1. Token Efficiency
CLAUDE.md is prepended to every prompt. Bloated files:
- Increase token costs
- Add noise that obscures important instructions
- Reduce Claude's ability to follow directions

**Target: Under 100 lines**

### 2. Write for Claude, Not Humans
Claude already knows general programming. Only include:
- Project-specific patterns
- Non-obvious architectural decisions
- Unexpected behaviors or warnings
- Repository-specific conventions

### 3. Declarative Over Narrative

| Style | Example |
|-------|---------|
| Good | `- Always add dark mode variants` |
| Bad | `When styling components, you should remember to add dark mode variants because...` |

## Required Sections

### Minimal CLAUDE.md Structure
```markdown
# Project Name
One-line description.

## Tech Stack
- Framework: Name version
- Styling: Name version

## Project Structure
[Abbreviated tree]

## Commands
[Essential scripts only]

## Key Patterns
[3-5 critical conventions]

## Common Tasks
[Step-by-step for frequent operations]
```

## Content Guidelines

### Include
- Exact dependency versions
- File path conventions with examples
- API patterns unique to this project
- Environment variable names
- Non-obvious component relationships

### Exclude
- How React/Next.js/TypeScript works
- General best practices (SOLID, DRY, etc.)
- Information in official docs
- Detailed API references (use separate files)
- Complete type definitions

## Hierarchical Organization

```
~/.claude/CLAUDE.md           # Personal global preferences
./CLAUDE.md                    # Project root (commit to git)
./src/tests/CLAUDE.md          # Subdirectory-specific rules
```

Most specific (nested) file takes precedence.

## Anti-Patterns

| Anti-Pattern | Why It's Bad |
|--------------|--------------|
| 500+ lines | Token waste, noise |
| "You should consider..." | Narrative, not declarative |
| Complete TypeScript interfaces | Put in separate file |
| General coding standards | Claude already knows |
| Version history | Not actionable |

## Maintenance

### Add to CLAUDE.md When
- You repeat the same instruction across sessions
- Claude makes the same mistake twice
- A pattern is project-specific and non-obvious

### Remove from CLAUDE.md When
- Information is outdated
- It duplicates official documentation
- Claude consistently follows it without reminder

## Validation Criteria

### Pass Criteria
- [ ] <= 100 lines
- [ ] All bullet points are declarative
- [ ] No general programming knowledge
- [ ] Versions specified for key dependencies
- [ ] File paths use clickable markdown links
- [ ] No narrative paragraphs > 2 sentences

### Improvement Indicators
- Line count > 100: Needs trimming
- "You should" phrases: Convert to imperatives
- Long code blocks: Move to separate reference file
- Detailed explanations: Summarize or remove
