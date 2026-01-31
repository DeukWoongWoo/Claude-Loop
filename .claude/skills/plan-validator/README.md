# Plan Validator

Systematically validate implementation plans before execution using the TFACS framework. Ensures plans are grounded in codebase reality, follow existing patterns, and address root causes rather than symptoms.

## Usage

Invoke this skill to validate implementation plans before coding begins.

**Trigger phrases:**
- "Validate plan"
- "Review plan"
- "Check plan"

**With arguments:**
```
/plan-validator                        # Validate current plan file
/plan-validator path/to/plan.md        # Validate specific plan file
```

## Description

This skill performs evidence-based plan validation using the **TFACS Framework**:

1. **TRACEABILITY** - Collect file:line evidence for every claim in the plan
2. **FOUNDATION** - Verify plan matches current codebase state
3. **ARCHITECTURE** - Check pattern consistency with existing code
4. **COMPLETENESS** - Ensure nothing is missed (callers, tests, edge cases)
5. **SUSTAINABILITY** - Assess root cause resolution, security, rollback

**Core Principle**: Never validate based on assumptions. Every finding must have code evidence.

## Workflow

When invoked, the skill:

1. **Assess Tier**: Determine validation depth (Simple/Standard/Complex)
2. **Collect Evidence**: Verify all file/function references exist
3. **Check Foundation**: Confirm plan matches codebase reality
4. **Verify Architecture**: Ensure pattern conformance
5. **Analyze Completeness**: Find missing callers, tests, edge cases
6. **Evaluate Sustainability**: Check root cause, security, rollback

**Output**: Validation report with status:
- PASSED - All critical checks pass
- PASSED WITH NOTES - Minor concerns to address
- NEEDS REVISION - Issues must be fixed before implementation
- REJECTED - Fundamental problems, plan needs rework

## Reference Documents

| Document | Purpose |
|----------|---------|
| `references/report-template.md` | Complete validation report template |
| `references/anti-patterns.md` | Common planning anti-patterns to avoid |
