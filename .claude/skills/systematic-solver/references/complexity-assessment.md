# Complexity Assessment Guide

Use this guide to determine whether a fix requires full ARIA analysis or can be applied immediately.

## Decision Matrix

| Factor | Simple (Fix Now) | Complex (Run ARIA) |
|--------|------------------|-------------------|
| **Cause clarity** | Obvious, well-understood | Unclear, requires investigation |
| **Scope** | Single file, single function | Multiple files or modules |
| **Dependencies** | No callers or single caller | Multiple callers, shared code |
| **Risk** | Low impact if wrong | Could break other features |
| **Recurrence** | One-time issue | Recurring or intermittent |
| **Domain** | Presentation, formatting | Business logic, data, security |

## Simple Fixes (Proceed Immediately)

### Typos and Spelling
```
- Variable name typos in error messages
- Comment spelling errors
- Documentation typos
- UI text corrections
```

### Obvious Syntax Errors
```
- Missing semicolons, brackets, parentheses
- Import statement errors
- Clearly wrong operator (= vs ==)
```

### Isolated Style Issues
```
- Formatting inconsistencies
- Adding missing type annotations (when obvious)
- Renaming local variables for clarity
```

### User-Specified Exact Fixes
```
- User provides the exact code change
- Change is localized and clear
- No ambiguity in implementation
```

## Complex Fixes (Run Full ARIA)

### Unknown Root Cause
```
- "It sometimes fails"
- "This started happening recently"
- "Not sure why this is broken"
- Stack trace points to symptom, not cause
```

### Multi-File Impact
```
- Shared utility functions
- Base classes with multiple subclasses
- Configuration that affects multiple features
- Database schema changes
```

### Architectural Concerns
```
- Changes to public APIs
- Modifications to core abstractions
- Performance optimizations
- Security-related code
```

### Recurring Issues
```
- Bug that was "fixed" before
- Intermittent failures
- Race conditions
- Environment-specific issues
```

### High-Stakes Domains
```
- Authentication/authorization
- Payment processing
- Data persistence
- User data handling
```

## Edge Cases

### Seems Simple But Is Complex

| Appears Simple | Actually Complex Because |
|----------------|-------------------------|
| "Just add a null check" | Why is it null? Could indicate deeper issue |
| "Just increase the timeout" | Why is it slow? Masking real problem |
| "Just catch the exception" | What caused it? Will silently hide bugs |
| "Just add a retry" | Why does it fail? Retry may not help |

### Seems Complex But Is Simple

| Appears Complex | Actually Simple Because |
|-----------------|------------------------|
| Large diff | All changes are mechanical (rename, format) |
| Multiple files | Same simple change in each file |
| Unfamiliar code | User provides full context and solution |

## Quick Assessment Checklist

Before starting any fix, ask:

1. [ ] Do I understand WHY this is broken?
2. [ ] Do I know ALL the places this change affects?
3. [ ] Could this fix break something else?
4. [ ] Has this been "fixed" before?
5. [ ] Is this in a high-risk area (auth, data, payments)?

**If ANY answer is "No" or "Unsure" → Run full ARIA workflow**

## Default Behavior

**When in doubt, treat as complex.**

The cost of running unnecessary analysis is low (few minutes).
The cost of patchwork fix is high (bugs return, trust erodes, tech debt grows).
