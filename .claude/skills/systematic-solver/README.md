# Systematic Solver

A framework that ensures thorough, non-patchwork solutions for code modifications and debugging tasks through root cause analysis and impact assessment.

## Usage

Invoke this skill for debugging, bug fixes, or code modifications that could affect other parts of the codebase.

**Trigger phrases:**
- "Fix bug"
- "Debug issue"
- "Solve problem"
- "Refactor code"

**With arguments:**
```
/systematic-solver                     # Analyze current issue
/systematic-solver "error message"     # Investigate specific error
```

## Description

This skill performs systematic problem-solving using the **ARIA Workflow**:

1. **ASSESS** - Determine complexity (Simple → fix immediately, Complex → full analysis)
2. **ROOT CAUSE** - Find the actual cause using "5 Whys" technique with code evidence
3. **IMPACT** - Analyze dependencies, risk level, and test coverage
4. **ACT** - Implement fix with regression prevention

**Core Principle**: Fix the root cause, not the symptom. A null check without fixing why the value is null is a patchwork fix.

## Workflow

When invoked, the skill:

1. **Assess Complexity**: Simple issues get immediate fixes, complex ones get full analysis
2. **Find Root Cause**: Apply "5 Whys" with code evidence for each step
3. **Analyze Impact**: Trace callers/callees, assess risk, check test coverage
4. **Select Approach**: Choose THE best solution based on project context
5. **Act with Prevention**: Implement fix and add regression tests

**Output**: Implementation summary with:
- Root cause statement with evidence
- Impact report (direct/indirect files, tests, risk level)
- Changes made with reasoning
- Regression prevention measures

## Reference Documents

| Document | Purpose |
|----------|---------|
| `references/complexity-assessment.md` | Detailed criteria for complexity tiers |
| `references/root-cause-patterns.md` | Common root cause patterns and examples |
| `references/impact-analysis-guide.md` | Comprehensive impact analysis checklist |
