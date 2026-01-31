---
name: plan-validator
description: Implementation plan validation workflow. Validates, reviews, and ensures plans are evidence-based with proper root cause analysis and code consistency. Use for any plan validation-related task.
---

# Plan Validator

Systematically validate implementation plans before execution using the TFACS framework. Ensures plans are grounded in codebase reality, follow existing patterns, and address root causes rather than symptoms.

## Core Principle: Evidence-First

**CRITICAL**: Never validate based on assumptions. Every finding must have code evidence.

```
❌ WRONG: "The plan looks reasonable"

✅ CORRECT: "Plan references validateUser() at user.service.ts:45"
   - Evidence: Function exists at user.service.ts:45-67
   - Signature matches: validateUser(email: string): Promise<User>
   - 5 callers found via grep, plan addresses 5/5
```

**Every validation finding MUST include:**
1. **File:Line** - Exact location verified
2. **Code snippet** - The actual code examined
3. **Status** - VERIFIED / UNVERIFIED / MISSING

---

## TFACS Framework

Execute these phases sequentially. Evidence collection (T) MUST come first.

```
Phase 0: TIER Assessment
    ↓
[Simple] → T + F only
[Standard] → Full TFACS
[Complex] → TFACS + extended impact
    ↓
Phase 1: TRACEABILITY - Collect evidence for all claims
    ↓
Phase 2: FOUNDATION - Verify plan matches codebase reality
    ↓
Phase 3: ARCHITECTURE - Check pattern consistency
    ↓
Phase 4: COMPLETENESS - Ensure nothing is missed
    ↓
Phase 5: SUSTAINABILITY - Assess root cause + security + rollback
    ↓
Output: Validation Report
```

---

## Phase 0: Tier Assessment

**Goal**: Determine validation depth based on plan complexity.

| Tier | Criteria | Validation Scope |
|------|----------|------------------|
| **Simple** | 1-2 files, additive only | T + F only |
| **Standard** | 3-5 files, modifications | Full TFACS |
| **Complex** | 6+ files, architectural changes | TFACS + extended impact |

**Complexity indicators for Complex tier:**
- Database schema changes
- Public API modifications
- Authentication/authorization changes
- Core module refactoring
- Dependency updates

**When in doubt, treat as Standard.**

---

## Phase 1: TRACEABILITY (Evidence Collection)

**Goal**: Collect file:line evidence for EVERY claim in the plan.

### Process

1. **Extract all claims from the plan**
   - File references (paths, line numbers)
   - Function/class references
   - Behavioral assumptions ("X calls Y", "Z returns null when...")

2. **For each claim, gather evidence**
   ```
   Claim: "Modify validateUser() in user.service.ts"

   Evidence gathering:
   - Glob: user.service.ts exists? → YES
   - Grep: validateUser function exists? → YES, line 45-67
   - Read: Verify function signature → validateUser(email: string)
   ```

3. **Categorize each claim**

   | Status | Meaning |
   |--------|---------|
   | VERIFIED | Code evidence confirms claim |
   | UNVERIFIED | Claim made without evidence (assumption) |
   | CONTRADICTED | Evidence contradicts claim |
   | MISSING | Referenced code does not exist |

4. **Build Assumptions Registry**
   - List all UNVERIFIED claims
   - Specify what discovery is needed

### Output
```
Evidence Summary:
- Total claims: 15
- Verified: 12
- Unverified: 2 (assumptions requiring discovery)
- Contradicted: 1 (CRITICAL)
- Missing: 0
```

---

## Phase 2: FOUNDATION (Reality Check)

**Goal**: Verify the plan matches current codebase state.

### Verification Checklist

1. **File Existence**
   - Do all referenced files exist?
   - Are paths correct (no typos)?

2. **Code Existence**
   - Do referenced functions/classes exist?
   - Are signatures accurate?
   - Are line numbers current?

3. **State Accuracy**
   - Does plan account for recent changes? (check git status)
   - Are dependencies at expected versions?

4. **Type Compatibility**
   - Do proposed changes maintain type safety?
   - Are return types handled correctly?

### Red Flags

| Issue | Example | Severity |
|-------|---------|----------|
| Phantom file | Plan references `src/utils/helper.ts` that doesn't exist | CRITICAL |
| Wrong line number | Plan says "line 45" but function is at line 78 | HIGH |
| Outdated signature | Plan uses old function signature | HIGH |
| Git conflict | Modified file has uncommitted changes | MEDIUM |

---

## Phase 3: ARCHITECTURE (Pattern Alignment)

**Goal**: Ensure plan follows existing codebase patterns.

### Verification Checklist

1. **Pattern Conformance**
   ```
   Search: How are similar problems solved?
   - Grep for similar implementations
   - Check existing utilities
   - Review established conventions
   ```

2. **Code Style Consistency**
   - Naming conventions (camelCase, snake_case, etc.)
   - File organization patterns
   - Import/export style

3. **Reuse vs Reinvent**
   - Does plan reuse existing utilities?
   - Or does it create duplicate functionality?

4. **Module Boundaries**
   - Does plan respect existing module structure?
   - Are dependencies flowing in correct direction?

### Red Flags

| Issue | Example | Severity |
|-------|---------|----------|
| Pattern violation | Creating new validator when `src/utils/validators/` exists | MEDIUM |
| Style inconsistency | Using `getUserData()` when codebase uses `get_user_data()` | MEDIUM |
| Duplicate utility | Creating `formatDate()` when `date-fns` is already used | MEDIUM |

---

## Phase 4: COMPLETENESS (Gap Analysis)

**Goal**: Ensure nothing is missed in the plan.

### Verification Checklist

1. **Impact Scope**
   ```
   For each modified function:
   - Grep for all callers (upstream impact)
   - Read to find all callees (downstream impact)
   - Verify ALL are addressed in plan
   ```

2. **Edge Cases**
   - Null/undefined handling
   - Error paths
   - Boundary conditions
   - Empty collections

3. **Test Coverage**
   - Are new tests planned for new behavior?
   - Are existing test updates identified?
   - Is test file location correct?

4. **Documentation**
   - API documentation updates needed?
   - README changes required?

### Red Flags

| Issue | Example | Severity |
|-------|---------|----------|
| Missing callers | 5 callers exist, plan addresses 2 | CRITICAL |
| No error handling | New code path has no try-catch | HIGH |
| Missing tests | New feature without test plan | HIGH |

---

## Phase 5: SUSTAINABILITY (Long-term Assessment)

**Goal**: Assess root cause resolution, security, and reversibility.

### Root Cause Analysis

| Aspect | Good | Poor |
|--------|------|------|
| Problem addressed | Fixes WHY value is null | Adds null check |
| Recurrence prevention | Adds validation at source | Patches at symptom |
| Test coverage | Tests prevent regression | No regression tests |

### Security Checklist

- [ ] Input validation for user data
- [ ] No secrets/credentials in code
- [ ] AuthZ/AuthN properly checked
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)

### Rollback Assessment

For Standard/Complex tiers:
- [ ] Can changes be reverted easily?
- [ ] Is there a migration strategy?
- [ ] Are feature flags needed?
- [ ] What's the blast radius if it fails?

### Red Flags

| Issue | Example | Severity |
|-------|---------|----------|
| Symptom patching | Adding null check without fixing root cause | HIGH |
| No rollback plan | Database migration without down migration | HIGH |
| Security gap | User input directly in SQL query | CRITICAL |

---

## Output Format

Use the template from `@references/report-template.md`

### Status Definitions

| Status | Criteria |
|--------|----------|
| **PASSED** | All critical checks pass, no unverified assumptions |
| **PASSED WITH NOTES** | Passes but has minor concerns to address |
| **NEEDS REVISION** | Has issues that must be fixed before implementation |
| **REJECTED** | Fundamental problems, plan needs complete rework |

---

## Anti-Patterns

See `@references/anti-patterns.md` for detailed examples.

**Quick list:**
1. Assumption-based planning ("probably", "might", "should")
2. Symptom patching (null checks without fixing root cause)
3. Pattern violations (new utility when existing one works)
4. Incomplete impact analysis (missing callers)
5. Vague implementation ("improve error handling")
6. No rollback strategy (high-risk changes without recovery plan)

---

## Quick Reference

```
TFACS Framework:
T - TRACEABILITY: Collect evidence FIRST (file:line for every claim)
F - FOUNDATION: Does plan match codebase reality?
A - ARCHITECTURE: Does it follow existing patterns?
C - COMPLETENESS: Is anything missing? (callers, tests, edge cases)
S - SUSTAINABILITY: Root cause? Security? Rollback?

Tier Assessment:
Simple (1-2 files, additive) → T + F only
Standard (3-5 files, modifications) → Full TFACS
Complex (6+ files, architectural) → TFACS + extended

Core Rule: NEVER GUESS
- Every claim needs file:line evidence
- Every assumption goes in Assumptions Registry
- Evidence contradicting plan = CRITICAL issue
```
