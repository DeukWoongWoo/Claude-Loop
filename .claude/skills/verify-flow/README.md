# Verify Flow

Systematically verify code changes by tracing execution flow from entry points through the entire call chain. Ensures parameters are correctly passed, return values properly handled, and error cases covered.

## Usage

Invoke this skill to verify code changes, validate plan implementations, or trace code execution paths.

**Trigger phrases:**
- "Verify flow"
- "Verify code flow"
- "Trace code path"
- "Check implementation"

**With arguments:**
```
/verify-flow                           # Verify uncommitted changes
/verify-flow src/api/users.ts          # Verify specific files
/verify-flow --plan                    # Verify current plan file
```

## Description

This skill performs evidence-based code flow verification using the **VERT Framework**:

1. **VALIDATE** - Determine verification scope (modified files, plan)
2. **ENTRY** - Discover entry points (API, CLI, event handlers)
3. **ROUTE** - Trace code flow (parameters, returns, errors)
4. **TRACE** - Classify issues (CRITICAL/WARNING/INFO) with code evidence
5. **REPORT** - Generate actionable report for Claude Code

**Core Principle**: Every finding must have code evidence (file:line, snippet, trigger condition). No guessing allowed.

## Workflow

When invoked, the skill:

1. **Validate Scope**: Identify verification target and changed functions
2. **Discover Entry Points**: Find where code execution begins (reverse trace)
3. **Trace Routes**: Follow call chain, verify parameters, returns, errors
4. **Verify Issues**: Classify by severity with code evidence
5. **Generate Report**: Output Action Items table for Claude Code processing

**Output**: Structured verification report with:
- Entry point analysis and flow paths
- Parameter flow verification tables
- Issues classified by priority (CRITICAL → WARNING → INFO)
- Action Items table (File, Line, Issue, Fix)
- Verification checklist for post-fix confirmation

## Report Format

The report is optimized for Claude Code to process automatically:

```markdown
## Action Items (For Claude Code Processing)

### CRITICAL (Immediate Fix Required)
| # | File | Line | Issue | Fix |
|---|------|------|-------|-----|
| 1 | file.ts | 30 | Missing null check | Add `if (!data) return;` |
```

## Reference Documents

| Document | Purpose |
|----------|---------|
| `references/verification-rules.md` | Parameter, return, error verification rules |
| `references/trace-patterns.md` | Entry point discovery and tracing techniques |
| `references/report-template.md` | Complete report template with examples |
