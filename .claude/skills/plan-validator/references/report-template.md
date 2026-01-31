# Plan Validation Report Template

Use this template to generate validation reports.

---

## Full Report Template

```markdown
# Plan Validation Report

## Summary

- **Status**: [PASSED | PASSED WITH NOTES | NEEDS REVISION | REJECTED]
- **Plan**: [Brief description of what the plan does]
- **Tier**: [Simple | Standard | Complex]
- **Validation Date**: [YYYY-MM-DD]

## Evidence Summary

| Metric | Count |
|--------|-------|
| Total claims | [N] |
| Verified | [N] |
| Unverified (assumptions) | [N] |
| Contradicted | [N] |
| Missing references | [N] |

---

## Phase Scores (TFACS)

| Phase | Score | Status | Critical Issues |
|-------|-------|--------|-----------------|
| T - Traceability | [X/10] | [PASS/FAIL] | [count] |
| F - Foundation | [X/10] | [PASS/FAIL] | [count] |
| A - Architecture | [X/10] | [PASS/FAIL] | [count] |
| C - Completeness | [X/10] | [PASS/FAIL] | [count] |
| S - Sustainability | [X/10] | [PASS/FAIL] | [count] |

---

## Validation Details

### Phase 1: Traceability

**Evidence Collection Results**

| Claim | Source | Status | Evidence |
|-------|--------|--------|----------|
| [claim description] | [plan section] | VERIFIED | [file:line] |
| [claim description] | [plan section] | UNVERIFIED | [none - assumption] |
| [claim description] | [plan section] | CONTRADICTED | [file:line shows different] |

**Issues Found**

#### [CRITICAL] [Issue Title]
- **Claim**: [What the plan states]
- **Evidence**: [What was actually found]
- **Location**: [file:line]
- **Action Required**: [What needs to change]

---

### Phase 2: Foundation

**Verification Results**

| Check | Status | Details |
|-------|--------|---------|
| Files exist | [PASS/FAIL] | [N/M files verified] |
| Functions exist | [PASS/FAIL] | [N/M functions verified] |
| Line numbers accurate | [PASS/FAIL] | [N/M line refs verified] |
| Git state clean | [PASS/FAIL] | [uncommitted changes?] |

**Issues Found**

#### [HIGH] [Issue Title]
- **Reference**: [What plan references]
- **Reality**: [What actually exists]
- **Action Required**: [What needs to change]

---

### Phase 3: Architecture

**Pattern Analysis**

| Pattern | Codebase Standard | Plan Approach | Match |
|---------|-------------------|---------------|-------|
| Naming | [convention used] | [plan uses] | [YES/NO] |
| Error handling | [pattern used] | [plan uses] | [YES/NO] |
| Utilities | [existing utils] | [plan uses] | [YES/NO] |

**Issues Found**

#### [MEDIUM] [Issue Title]
- **Existing Pattern**: [file:line showing pattern]
- **Plan Approach**: [what plan proposes]
- **Recommendation**: [how to align]

---

### Phase 4: Completeness

**Impact Analysis**

| Modified Code | Callers Found | Callers Addressed | Gap |
|---------------|---------------|-------------------|-----|
| [function name] | [N] | [M] | [N-M missing] |

**Missing Callers Detail**

| # | Caller | File:Line | Status |
|---|--------|-----------|--------|
| 1 | [caller name] | [file:line] | MISSING from plan |

**Test Coverage**

| Aspect | Status |
|--------|--------|
| New tests planned | [YES/NO] |
| Existing tests updated | [YES/NO] |
| Test file locations correct | [YES/NO] |

**Issues Found**

#### [CRITICAL] [Issue Title]
- **Gap**: [What is missing]
- **Evidence**: [Grep results showing missing items]
- **Action Required**: [What needs to be added]

---

### Phase 5: Sustainability

**Root Cause Assessment**

| Question | Answer |
|----------|--------|
| Symptom | [observable problem] |
| Root cause identified | [YES/NO] |
| Plan fixes root cause | [YES/NO] |
| Regression prevention | [tests planned?] |

**Security Checklist**

- [ ] Input validation addressed
- [ ] No hardcoded secrets
- [ ] Auth/AuthZ proper
- [ ] SQL injection prevented
- [ ] XSS prevented

**Rollback Assessment**

| Aspect | Status |
|--------|--------|
| Git revertible | [YES/NO] |
| Database rollback | [YES/NO/N/A] |
| Feature flag | [YES/NO/N/A] |
| Blast radius | [LOW/MEDIUM/HIGH] |

**Issues Found**

#### [HIGH] [Issue Title]
- **Issue**: [What's wrong]
- **Risk**: [What could go wrong]
- **Required**: [What needs to be added]

---

## Assumptions Registry

| # | Assumption | Plan Location | Required Discovery | Priority |
|---|------------|---------------|-------------------|----------|
| 1 | [assumption] | [section] | [action needed] | [HIGH/MED/LOW] |

---

## Action Items

### CRITICAL (Must Fix)

| # | Phase | Issue | Required Action |
|---|-------|-------|-----------------|
| 1 | [T/F/A/C/S] | [issue] | [action] |

### HIGH (Should Fix)

| # | Phase | Issue | Recommended Action |
|---|-------|-------|-------------------|
| 1 | [T/F/A/C/S] | [issue] | [action] |

### MEDIUM (Consider)

| # | Phase | Issue | Suggestion |
|---|-------|-------|------------|
| 1 | [T/F/A/C/S] | [issue] | [suggestion] |

---

## Verification Checklist (Post-Revision)

After addressing the action items, verify:

- [ ] All CRITICAL issues resolved
- [ ] All HIGH issues addressed
- [ ] Assumptions discovered and documented
- [ ] Evidence added for all claims
- [ ] Impact scope complete
- [ ] Tests planned for new behavior
- [ ] Security concerns addressed
- [ ] Rollback strategy defined (if high-risk)
```

---

## Abbreviated Report Template (Simple Tier)

For Simple tier validations, use this abbreviated format:

```markdown
# Plan Validation Report (Quick)

**Status**: [PASSED | NEEDS REVISION]
**Plan**: [Brief description]
**Tier**: Simple

## Traceability Check

| Claim | Evidence | Status |
|-------|----------|--------|
| [claim] | [file:line] | [VERIFIED/UNVERIFIED] |

## Foundation Check

| File | Exists | Status |
|------|--------|--------|
| [path] | [YES/NO] | [OK/ISSUE] |

## Issues (if any)

[List issues or "None found"]

## Action Items

[List actions or "None - PASSED"]
```

---

## Status Definitions

### PASSED

All checks pass:
- All claims have evidence
- All references verified
- No critical or high issues
- No unverified assumptions

### PASSED WITH NOTES

Passes but has minor concerns:
- All critical checks pass
- Some medium-severity suggestions
- Minor pattern deviations (documented)
- Non-blocking recommendations

### NEEDS REVISION

Has issues that must be fixed:
- Unverified assumptions present
- Missing impact coverage
- Pattern violations
- No test plan for new behavior
- High-severity issues

### REJECTED

Fundamental problems:
- Critical references don't exist
- Plan based on incorrect assumptions
- Security vulnerabilities
- No root cause addressed
- Multiple critical failures

---

## Severity Guidelines

| Severity | Criteria | Action |
|----------|----------|--------|
| CRITICAL | Blocks implementation, will cause failure | Must fix before any implementation |
| HIGH | Significant risk or gap | Should fix before implementation |
| MEDIUM | Improvement opportunity, best practice | Consider fixing, document if not |
| LOW | Minor suggestion | Optional, at discretion |

---

## Example: Completed Report

```markdown
# Plan Validation Report

## Summary

- **Status**: NEEDS REVISION
- **Plan**: Add OAuth2 authentication to user service
- **Tier**: Standard
- **Validation Date**: 2026-01-29

## Evidence Summary

| Metric | Count |
|--------|-------|
| Total claims | 12 |
| Verified | 9 |
| Unverified (assumptions) | 2 |
| Contradicted | 1 |
| Missing references | 0 |

---

## Phase Scores (TFACS)

| Phase | Score | Status | Critical Issues |
|-------|-------|--------|-----------------|
| T - Traceability | 7/10 | PASS | 0 |
| F - Foundation | 9/10 | PASS | 0 |
| A - Architecture | 8/10 | PASS | 0 |
| C - Completeness | 5/10 | FAIL | 1 |
| S - Sustainability | 6/10 | PASS | 0 |

---

## Validation Details

### Phase 4: Completeness

**Impact Analysis**

| Modified Code | Callers Found | Callers Addressed | Gap |
|---------------|---------------|-------------------|-----|
| authenticateUser() | 5 | 3 | 2 missing |

**Missing Callers Detail**

| # | Caller | File:Line | Status |
|---|--------|-----------|--------|
| 1 | AdminController | src/api/admin.ts:78 | MISSING from plan |
| 2 | SyncWorker | src/workers/sync.ts:23 | MISSING from plan |

**Issues Found**

#### [CRITICAL] Incomplete Impact Analysis
- **Gap**: 2 of 5 callers not addressed
- **Evidence**:
  ```bash
  grep -rn "authenticateUser(" src/
  # Shows 5 results, plan only addresses 3
  ```
- **Action Required**: Add impact analysis for admin.ts and sync.ts

---

## Assumptions Registry

| # | Assumption | Plan Location | Required Discovery | Priority |
|---|------------|---------------|-------------------|----------|
| 1 | "OAuth tokens expire in 1 hour" | Step 3 | Read auth.config.ts | HIGH |
| 2 | "Refresh tokens stored in Redis" | Step 4 | Verify Redis config | MEDIUM |

---

## Action Items

### CRITICAL (Must Fix)

| # | Phase | Issue | Required Action |
|---|-------|-------|-----------------|
| 1 | C | 2 callers missing | Add admin.ts:78 and sync.ts:23 to plan |

### HIGH (Should Fix)

| # | Phase | Issue | Recommended Action |
|---|-------|-------|-------------------|
| 1 | T | OAuth token expiry unverified | Read auth.config.ts, verify TTL |
| 2 | S | No rollback plan | Add OAuth disable feature flag |

### MEDIUM (Consider)

| # | Phase | Issue | Suggestion |
|---|-------|-------|------------|
| 1 | A | New OAuth client creation | Consider using existing oauth-client at utils/oauth.ts |
```
