---
name: claude-md-validator
description: Validate and improve CLAUDE.md files following best practices. Use when user asks to review, validate, check, or improve CLAUDE.md files, or mentions CLAUDE.md quality.
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
- [ ] Total lines <= 100 (token efficiency)
- [ ] Uses declarative bullet points, not narrative paragraphs
- [ ] Has clear section headers

### Content Quality
- [ ] Project-specific information only (not general language/framework knowledge)
- [ ] Tech stack with exact versions
- [ ] Key file locations with clickable links
- [ ] Common commands documented
- [ ] Coding patterns and conventions

### What to Exclude
- [ ] No general documentation available elsewhere
- [ ] No time-sensitive information (dates, deadlines)
- [ ] No redundant information

### Format
- [ ] Consistent terminology throughout
- [ ] Code examples are minimal and focused
- [ ] Tables used for structured data (API routes, commands)

## Output Format

```markdown
## CLAUDE.md Validation Report

### Summary
- Total lines: X (target: <= 100)
- Status: PASS / NEEDS IMPROVEMENT

### Issues Found
1. [Line X] Issue description
   - Suggestion: How to fix

### Recommendations
- Priority improvements listed
```

## Examples

### Good: Declarative Style
```markdown
- Use `'use client'` for interactive components
- No `any` type - use proper types or `unknown`
```

### Bad: Narrative Style
```markdown
When creating components, you should consider whether they need
client-side interactivity. If so, add the 'use client' directive...
```

## Additional Reference

See BEST_PRACTICES.md for detailed criteria and examples.
