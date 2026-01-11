---
name: systematic-solver
description: Systematic problem-solving and code modification framework. Use this skill when debugging, fixing bugs, resolving errors, troubleshooting, refactoring, optimizing performance, modifying code, investigating issues, diagnosing problems, patching code, or correcting behavior. Ensures root cause analysis, impact assessment, and regression prevention instead of patchwork fixes.
---

# Systematic Problem Solver

A framework that ensures thorough, non-patchwork solutions for code modifications and debugging tasks.

## When This Skill Activates

- User reports a bug or error
- User requests debugging help
- User asks to fix something
- User requests refactoring
- User reports performance issues
- Any code modification that could affect other parts of the codebase

## The ARIA Workflow

```
Phase 0: ASSESS complexity
    ↓
[Simple] → Fix immediately
[Complex] → Continue to Phase 1
    ↓
Phase 1: ROOT CAUSE analysis
    ↓
Phase 2: IMPACT analysis
    ↓
Phase 3: Best APPROACH selection
    ↓
Phase 4: ACT with regression prevention
```

---

## Phase 0: Complexity Assessment

**Before any modification, assess complexity to determine workflow depth.**

### Simple (Fix Immediately)
- Typos, spelling errors
- Single-line obvious fixes
- Clear syntax errors
- Well-isolated changes with no dependencies
- User explicitly provides the exact fix

### Complex (Run Full ARIA)
- Cause is unclear or requires investigation
- Multiple files may be affected
- Structural or architectural changes
- Performance issues
- Recurring or intermittent bugs
- Changes to shared utilities, APIs, or core modules

**When in doubt, treat as complex.** See `references/complexity-assessment.md` for detailed criteria.

---

## Phase 1: Root Cause Analysis

**Goal: Find the actual cause, not the surface symptom.**

### Process

1. **Identify the Symptom**
   - What is the observable problem?
   - Error message, unexpected behavior, performance degradation?

2. **Apply the "5 Whys" Technique**
   ```
   Symptom: API returns 500 error
   Why 1: Database query throws exception
   Why 2: Query references non-existent column
   Why 3: Migration was not run after schema change
   Why 4: Deployment script skipped migration step
   Root Cause: Incomplete deployment process
   ```

3. **Validate with Code Evidence**
   - Read relevant source files
   - Trace execution paths
   - Check logs, stack traces
   - Each "why" must be supported by actual code/data

4. **Distinguish Surface vs Root**

   | Surface Fix (Avoid) | Root Cause Fix (Prefer) |
   |---------------------|------------------------|
   | Add null check | Fix why value is null |
   | Catch and ignore exception | Fix what causes exception |
   | Add retry logic | Fix why it fails initially |
   | Increase timeout | Fix why it's slow |

See `references/root-cause-patterns.md` for common patterns.

### Output
```
Root Cause Statement:
- Symptom: [observable problem]
- Root Cause: [actual underlying issue]
- Evidence: [code references, stack traces]
```

---

## Phase 2: Impact Analysis

**Goal: Understand what else is affected by the problem AND the proposed fix.**

### Process

1. **Trace Dependencies**
   - What calls this code? (callers/consumers)
   - What does this code call? (callees/dependencies)
   - What data flows through here?

2. **Search for Usages**
   ```
   - Grep for function/class name across codebase
   - Check imports and exports
   - Review test files that cover this code
   ```

3. **Assess Risk Level**

   | Risk | Criteria |
   |------|----------|
   | High | Public APIs, shared utilities, database schemas, authentication/security |
   | Medium | Internal modules with multiple callers, configuration changes |
   | Low | Isolated code with single caller, local variables, private functions |

4. **Check Test Coverage**
   - Existing tests that cover affected code
   - Tests that may need updates
   - Missing test coverage

See `references/impact-analysis-guide.md` for detailed checklist.

### Output
```
Impact Report:
- Direct: [files/functions being modified]
- Indirect: [files/functions that depend on modified code]
- Tests: [relevant test files]
- Risk Level: [High/Medium/Low]
- Risk Factors: [specific concerns]
```

---

## Phase 3: Best Approach Selection

**Goal: Select THE best solution based on project context, not list alternatives.**

### Process

1. **Analyze Project Context**
   - How are similar problems solved in this codebase?
   - What patterns and conventions exist?
   - What dependencies are already available?

2. **Generate Candidate Approaches** (Internal)
   - Quick fix: Minimal change to address symptom
   - Proper fix: Addresses root cause directly
   - Structural fix: Improves architecture to prevent recurrence

3. **Evaluate Against Criteria**

   | Criterion | Weight | Question |
   |-----------|--------|----------|
   | Solves root cause | HIGH | Does it fix the actual problem, not just symptom? |
   | Matches project patterns | HIGH | Is it consistent with existing code style? |
   | Minimizes blast radius | MEDIUM | Does it limit changes to affected areas? |
   | Testable | HIGH | Can we verify the fix works? |
   | Maintainable | MEDIUM | Will future developers understand it? |

4. **Select and Justify**
   - Choose ONE approach
   - Explain WHY this is best for THIS project
   - Do NOT present multiple options unless explicitly asked

### Output
```
Recommended Approach:
- Approach: [description]
- Reasoning: [why this is best for this project context]
- Trade-offs acknowledged: [what we're accepting]
```

---

## Phase 4: Act with Regression Prevention

**Goal: Implement the fix AND prevent recurrence.**

### Process

1. **Execute the Fix**
   - Apply the selected approach
   - Follow project coding standards
   - Keep changes focused and minimal

2. **Add Regression Prevention**
   - Write tests for the fixed behavior
   - Add assertions/validations where appropriate
   - Update documentation if needed

3. **Check for Similar Issues**
   - Are there similar code patterns that might have the same bug?
   - Should the fix be applied elsewhere?

4. **Verify No New Problems**
   - Run existing tests
   - Check that indirect dependencies still work
   - Validate the original issue is resolved

### Output
```
Implementation Summary:

## Root Cause
[From Phase 1]

## Impact
[From Phase 2]

## Changes Made
- [file]: [what changed and why]

## Regression Prevention
- [x] Tests added/updated
- [x] Similar code reviewed
- [x] Existing tests pass

## Verification
- [How to verify the fix works]
```

---

## Anti-Patterns to Avoid

1. **Patchwork Fixing**
   - Fixing only the immediate symptom
   - Adding workarounds without understanding cause
   - "It works now" without knowing why

2. **Tunnel Vision**
   - Modifying code without checking who uses it
   - Ignoring test failures as "unrelated"
   - Assuming changes are isolated

3. **Alternative Paralysis**
   - Presenting 5 options without recommendation
   - Asking user to choose technical approach
   - Avoiding commitment to a solution

4. **Regression Blindness**
   - Not adding tests for fixed behavior
   - Ignoring similar code that might have same issue
   - Not running existing tests

---

## Quick Reference

```
ARIA Workflow:
A - Assess complexity (simple → fix, complex → continue)
R - Root cause analysis (5 Whys, code evidence)
I - Impact analysis (dependencies, risk, tests)
A - Act with best approach + regression prevention
```
