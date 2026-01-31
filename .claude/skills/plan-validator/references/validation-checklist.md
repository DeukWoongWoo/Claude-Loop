# TFACS Validation Checklist

Detailed checklist for each phase of the TFACS framework.

---

## Phase 1: TRACEABILITY Checklist

### Claim Extraction

For each section of the plan, extract:

- [ ] File paths mentioned
- [ ] Function/class names mentioned
- [ ] Line numbers referenced
- [ ] Behavioral assumptions made
- [ ] Dependencies mentioned

### Evidence Collection

For each extracted claim:

| Claim Type | Verification Method | Tool |
|------------|---------------------|------|
| File exists | Check file presence | Glob |
| Function exists | Search for definition | Grep |
| Signature correct | Read function | Read |
| Behavior accurate | Trace execution | Read + Grep |
| Dependency version | Check package.json | Read |

### Evidence Documentation Format

```markdown
## Claim: [description]
- **Source**: [plan section/line]
- **Status**: VERIFIED | UNVERIFIED | CONTRADICTED | MISSING
- **Evidence**: [file:line or "no evidence found"]
- **Details**: [code snippet or explanation]
```

### Assumptions Registry Format

```markdown
| # | Assumption | Plan Location | Required Discovery |
|---|------------|---------------|-------------------|
| 1 | "Cache TTL is 5 minutes" | Step 3 | Read cache.config.ts |
| 2 | "User model has email field" | Step 1 | Verify User type definition |
```

---

## Phase 2: FOUNDATION Checklist

### File Verification

- [ ] All referenced files exist
- [ ] File paths are correct (no typos, correct case)
- [ ] Files are in expected directories

```bash
# Verification command pattern
ls -la [referenced-file-path]
```

### Code Verification

- [ ] All referenced functions/classes exist
- [ ] Signatures match plan description
- [ ] Line numbers are accurate (within 5 lines tolerance)

```bash
# Verification command pattern
grep -n "function functionName" [file]
grep -n "class ClassName" [file]
```

### State Verification

- [ ] Check git status for uncommitted changes in plan files
- [ ] Verify no pending migrations affect plan
- [ ] Check for recent commits that might affect plan

```bash
# Verification commands
git status [referenced-files]
git log --oneline -5 [referenced-files]
```

### Type Compatibility

- [ ] Proposed changes maintain type safety
- [ ] Return types are handled by callers
- [ ] Parameter types match at call sites

---

## Phase 3: ARCHITECTURE Checklist

### Pattern Discovery

1. **Find similar implementations**
   ```bash
   # Search for similar patterns
   grep -r "similar-pattern" src/
   ```

2. **Check existing utilities**
   ```bash
   # Look in common utility locations
   ls src/utils/ src/helpers/ src/lib/
   ```

3. **Review conventions**
   - Naming conventions used in codebase
   - File organization patterns
   - Import/export styles

### Pattern Conformance

| Check | Question | How to Verify |
|-------|----------|---------------|
| Naming | Does plan follow existing naming? | Compare with similar files |
| Structure | Does plan follow file structure? | Check similar components |
| Error handling | Does plan match error patterns? | Find try-catch examples |
| Logging | Does plan use existing logger? | Search for logger usage |

### Reuse Assessment

- [ ] Plan uses existing utilities where available
- [ ] No duplicate functionality created
- [ ] Follows DRY principle

```markdown
## Reuse Check
- Existing utility: [utility name] at [file:line]
- Plan approach: [uses existing | creates new]
- Assessment: [OK | NEEDS REVISION - should use existing]
```

---

## Phase 4: COMPLETENESS Checklist

### Impact Scope

1. **Find all callers (upstream)**
   ```bash
   grep -r "functionName(" src/ --include="*.ts" --include="*.js"
   ```

2. **Find all callees (downstream)**
   - Read function implementation
   - List all functions it calls
   - Verify those are addressed

3. **Impact Documentation**
   ```markdown
   ## Impact Analysis: [function name]

   ### Callers (upstream)
   | File | Line | Status in Plan |
   |------|------|----------------|
   | src/api/user.ts | 45 | Addressed |
   | src/api/admin.ts | 78 | MISSING |

   ### Callees (downstream)
   | Function | File | Impact |
   |----------|------|--------|
   | validateEmail | validators.ts | No change needed |
   ```

### Edge Case Coverage

- [ ] Null/undefined inputs handled
- [ ] Empty arrays/objects handled
- [ ] Error conditions covered
- [ ] Boundary values considered
- [ ] Concurrent access scenarios (if applicable)

### Test Coverage

- [ ] New tests planned for new functionality
- [ ] Existing tests updated for changed behavior
- [ ] Test file locations correct
- [ ] Test naming follows conventions

```markdown
## Test Plan Assessment
- New tests needed: [count]
- Existing tests to update: [count]
- Test locations verified: YES | NO
```

### Documentation Coverage

- [ ] API documentation updates identified
- [ ] README changes noted (if needed)
- [ ] Code comments planned for complex logic
- [ ] Type definitions updated

---

## Phase 5: SUSTAINABILITY Checklist

### Root Cause Assessment

| Question | Answer |
|----------|--------|
| What is the symptom? | [description] |
| What is the root cause? | [description] |
| Does plan fix root cause? | YES | NO - only fixes symptom |
| Will issue recur? | YES | NO |

### Root Cause vs Symptom Examples

| Symptom Fix (AVOID) | Root Cause Fix (PREFER) |
|---------------------|------------------------|
| Add null check | Fix why value is null |
| Catch and ignore exception | Fix what causes exception |
| Add retry logic | Fix why it fails initially |
| Increase timeout | Fix why it's slow |
| Add try-catch wrapper | Handle error at source |

### Security Checklist

#### Input Validation
- [ ] User inputs validated before use
- [ ] File paths sanitized
- [ ] SQL parameters escaped/parameterized
- [ ] HTML output encoded

#### Authentication/Authorization
- [ ] AuthN checks in place for protected routes
- [ ] AuthZ checks for resource access
- [ ] Token validation implemented

#### Secrets Management
- [ ] No hardcoded credentials
- [ ] Secrets from environment/vault
- [ ] No secrets in logs

#### Common Vulnerabilities
- [ ] SQL injection prevented (parameterized queries)
- [ ] XSS prevented (output encoding)
- [ ] CSRF tokens used (if applicable)
- [ ] Path traversal prevented

### Rollback Assessment

| Aspect | Status |
|--------|--------|
| Can changes be reverted with git? | YES | NO |
| Database migration has down script? | YES | NO | N/A |
| Feature flag available? | YES | NO | N/A |
| Blast radius if failure? | LOW | MEDIUM | HIGH |

### Rollback Plan Format

```markdown
## Rollback Strategy

### Immediate Rollback
1. Revert commit: `git revert [commit-hash]`
2. Redeploy previous version

### Database Rollback (if applicable)
1. Run down migration: `npm run migrate:down`
2. Verify data integrity

### Feature Flag (if applicable)
1. Disable flag: [flag-name]
2. Monitor for issues
```

---

## Validation Summary Template

```markdown
# Validation Summary

## Tier: [Simple | Standard | Complex]

## Phase Results

| Phase | Status | Critical Issues |
|-------|--------|-----------------|
| T - Traceability | PASS/FAIL | [count] |
| F - Foundation | PASS/FAIL | [count] |
| A - Architecture | PASS/FAIL | [count] |
| C - Completeness | PASS/FAIL | [count] |
| S - Sustainability | PASS/FAIL | [count] |

## Overall Status: [PASSED | PASSED WITH NOTES | NEEDS REVISION | REJECTED]

## Critical Action Items
1. [item]
2. [item]
```
