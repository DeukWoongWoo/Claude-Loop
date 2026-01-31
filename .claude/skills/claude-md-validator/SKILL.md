---
name: claude-md-validator
description: CLAUDE.md configuration workflow. Validates, improves, and optimizes CLAUDE.md files against best practices. Use for any CLAUDE.md-related task.
---

# CLAUDE.md Validator

Analyze CLAUDE.md files against best practices and provide actionable improvements.

## Workflow

1. Read the target CLAUDE.md file
2. Analyze against each criterion in the checklist below
3. Report findings with specific line references
4. Suggest concrete improvements

## Validation Checklist

### Structure (Critical)
- [ ] Total lines <= 250 (token efficiency for complex monorepos)
- [ ] Uses declarative bullet points, not narrative paragraphs
- [ ] Has clear section headers
- [ ] Uses `@import` syntax for detailed patterns (e.g., `@documents/claude/patterns/*.md`)

### Content Quality
- [ ] Project-specific information only (not general Python/FastAPI knowledge)
- [ ] Tech stack with exact versions
- [ ] Key file locations with clickable links
- [ ] Common commands documented
- [ ] Coding patterns and conventions (Pydantic, type hints, import order)

### What to Exclude
- [ ] No general documentation available elsewhere
- [ ] No time-sensitive information (dates, deadlines)
- [ ] No redundant information
- [ ] No detailed implementation patterns (move to `documents/claude/patterns/`)

### Format
- [ ] Consistent terminology throughout
- [ ] Code examples are minimal and focused
- [ ] Tables used for structured data (layer responsibilities, commands)

## Output Format

```markdown
## CLAUDE.md Validation Report

### Summary
- Total lines: X (target: <= 250)
- Status: PASS / NEEDS IMPROVEMENT

### Issues Found
1. [Line X] Issue description
   - Suggestion: How to fix

### Recommendations
- Priority improvements listed
```

## Examples

### Good: Declarative Style
```python
# Import order
# 1. Standard library
# 2. Third-party
# 3. Local
```

### Bad: Narrative Style
```markdown
When importing modules, you should consider the order carefully.
First import standard library modules, then third-party packages...
```

### Good: Pattern Reference
```markdown
## Framework Development Patterns

상세 구현 패턴은 필요할 때 참조:

- **API 엔드포인트 추가**: @documents/claude/patterns/api-endpoint.md
- **CLI 명령어 추가**: @documents/claude/patterns/cli-command.md
```

### Bad: Inline Pattern Details
```markdown
## Adding API Endpoint

1. Create router in `alphaverse/server/api/v1/<feature>.py`:
[50+ lines of code examples...]
```

## Additional Reference

See BEST_PRACTICES.md for detailed criteria and examples.
