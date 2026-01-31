# Plan Anti-Patterns

Common mistakes in implementation plans that should be flagged during validation.

---

## 1. Assumption-Based Planning

**Description**: Plan contains claims without code evidence, using speculative language.

### Red Flag Keywords
- "probably", "might", "possibly", "should"
- "I think", "I believe", "I assume"
- "likely", "unlikely", "maybe"

### Example

```markdown
❌ BAD PLAN:
"The cache probably expires too quickly, so we should increase the TTL.
The user service might be calling the deprecated API."

✅ GOOD PLAN:
"Cache TTL is set to 60 seconds (cache.config.ts:15). Increasing to 300 seconds.
User service calls deprecated API at user.service.ts:78 (evidence: import statement at line 3)."
```

### Severity: CRITICAL

### Validation Action
- Add to Assumptions Registry
- Require evidence discovery before proceeding
- Status: NEEDS REVISION

---

## 2. Symptom Patching

**Description**: Plan addresses the observable symptom rather than the root cause.

### Pattern Detection

| Symptom Patch | Root Cause Fix |
|---------------|----------------|
| Add null check | Fix why value is null |
| Add try-catch | Fix what throws exception |
| Add retry logic | Fix why it fails |
| Increase timeout | Fix why it's slow |
| Add default value | Fix why value is missing |

### Example

```markdown
❌ BAD PLAN:
"Add null check before accessing user.email to prevent crash"

// Proposed change:
if (user && user.email) {
  sendEmail(user.email);
}

✅ GOOD PLAN:
"User is null because getUser() returns null when token is expired.
Fix: Add token refresh before getUser() call at auth.ts:45.
The null check is a symptom patch - the real issue is expired token handling."

// Root cause fix:
const user = await refreshTokenAndGetUser(token); // Handles expiry internally
sendEmail(user.email);
```

### Severity: HIGH

### Validation Action
- Apply "5 Whys" technique to find root cause
- Ensure plan addresses root cause, not symptom
- Status: NEEDS REVISION if only symptom addressed

---

## 3. Pattern Violation

**Description**: Plan creates new implementations when existing patterns/utilities should be used.

### Pattern Detection

1. Search for existing utilities before accepting new ones
2. Check if similar functionality exists in:
   - `src/utils/`
   - `src/helpers/`
   - `src/lib/`
   - Third-party dependencies

### Example

```markdown
❌ BAD PLAN:
"Create new formatDate() function in src/components/DatePicker/utils.ts"

// Proposed:
export function formatDate(date: Date): string {
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
}

✅ GOOD PLAN:
"Use existing date-fns library (already in package.json) for date formatting.
Import format from 'date-fns' and use format(date, 'yyyy-MM-dd')."

// Evidence:
// - date-fns in package.json:15 → "date-fns": "^2.30.0"
// - Used elsewhere: src/utils/dateHelpers.ts:3 → import { format } from 'date-fns'
```

### Severity: MEDIUM

### Validation Action
- Flag duplicate utility creation
- Recommend using existing utility
- Status: NEEDS REVISION

---

## 4. Incomplete Impact Analysis

**Description**: Plan modifies code without identifying all affected callers/consumers.

### Pattern Detection

1. Count callers using grep
2. Compare with plan's stated impact
3. Flag any discrepancy

### Example

```markdown
❌ BAD PLAN:
"Modify validateUser() signature to accept additional parameter.
Update call site in auth.controller.ts"

// Grep reveals 5 callers, plan only mentions 1

✅ GOOD PLAN:
"Modify validateUser() signature to accept additional parameter.

Impact Analysis (5 callers found):
1. src/controllers/auth.controller.ts:45 - Update required
2. src/controllers/user.controller.ts:23 - Update required
3. src/services/admin.service.ts:78 - Update required
4. src/middleware/auth.middleware.ts:12 - Update required
5. test/auth.test.ts:34 - Update required

All 5 callers addressed in this plan."
```

### Severity: CRITICAL

### Validation Action
- Run grep to find all callers
- Verify each is addressed in plan
- Status: NEEDS REVISION if any missing

---

## 5. Vague Implementation

**Description**: Plan describes changes in abstract terms without specific, actionable steps.

### Red Flag Phrases
- "improve error handling"
- "refactor the code"
- "clean up the implementation"
- "optimize performance"
- "fix the issue"

### Example

```markdown
❌ BAD PLAN:
"Improve error handling in the user service.
Refactor the authentication flow for better maintainability."

✅ GOOD PLAN:
"Add specific error types in user.service.ts:

1. Line 45: Change `throw new Error('User not found')` to
   `throw new UserNotFoundError(userId)`
2. Line 67: Add try-catch around database call, throw DatabaseError on failure
3. Line 89: Add input validation, throw ValidationError for invalid email format

Create new error types in src/errors/user.errors.ts:
- UserNotFoundError extends AppError
- DatabaseError extends AppError
- ValidationError extends AppError"
```

### Severity: HIGH

### Validation Action
- Require specific file:line changes
- Each change should be independently verifiable
- Status: NEEDS REVISION until specific

---

## 6. Missing Rollback Strategy

**Description**: High-risk changes without recovery plan if something goes wrong.

### High-Risk Changes Requiring Rollback Plan
- Database schema modifications
- API breaking changes
- Authentication/authorization changes
- Data migrations
- Third-party integration changes

### Example

```markdown
❌ BAD PLAN:
"Add new column 'preferences' to users table.
Update User model to include preferences field."

// No rollback mentioned for database change

✅ GOOD PLAN:
"Add new column 'preferences' to users table.

Migration: migrations/20240129_add_user_preferences.ts
- UP: ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}'
- DOWN: ALTER TABLE users DROP COLUMN preferences

Rollback Strategy:
1. If issues detected: Run down migration within 1 hour
2. Data is additive only - no existing data modified
3. Feature flag 'user_preferences_enabled' controls access
4. Blast radius: LOW - new column with default, no breaking changes"
```

### Severity: HIGH (for high-risk changes)

### Validation Action
- Identify high-risk changes
- Require explicit rollback strategy
- Status: NEEDS REVISION if missing

---

## 7. Phantom File Reference

**Description**: Plan references files, functions, or line numbers that don't exist or are incorrect.

### Pattern Detection
- Verify every file path with Glob
- Verify every function with Grep
- Check line numbers are within file length

### Example

```markdown
❌ BAD PLAN:
"Modify the helper function at src/utils/stringHelper.ts:45"

// File doesn't exist (actual: src/utils/string-helpers.ts)
// Or line 45 doesn't exist (file only has 30 lines)

✅ GOOD PLAN:
"Modify formatString() at src/utils/string-helpers.ts:23

Evidence:
- File exists: ls src/utils/string-helpers.ts ✓
- Function exists at line 23: grep -n 'formatString' src/utils/string-helpers.ts ✓
- File has 45 lines: wc -l src/utils/string-helpers.ts ✓"
```

### Severity: CRITICAL

### Validation Action
- Verify every file reference exists
- Verify line numbers are valid
- Status: REJECTED if critical references are phantom

---

## 8. Security Oversight

**Description**: Plan introduces security vulnerabilities or ignores security best practices.

### Common Security Anti-Patterns

| Anti-Pattern | Example | Fix |
|--------------|---------|-----|
| SQL injection | `query("SELECT * FROM users WHERE id = " + userId)` | Use parameterized queries |
| XSS | `innerHTML = userInput` | Use textContent or sanitize |
| Hardcoded secrets | `const API_KEY = "abc123"` | Use environment variables |
| Missing auth check | New endpoint without auth middleware | Add authentication |
| Path traversal | `readFile(userProvidedPath)` | Validate and sanitize path |

### Example

```markdown
❌ BAD PLAN:
"Add new search endpoint that queries users by name:
app.get('/search', (req, res) => {
  const result = db.query(`SELECT * FROM users WHERE name LIKE '%${req.query.name}%'`);
  res.json(result);
});"

✅ GOOD PLAN:
"Add new search endpoint with proper security:

1. Add authentication middleware (auth.middleware.ts)
2. Use parameterized query to prevent SQL injection
3. Validate and sanitize input
4. Rate limit the endpoint

app.get('/search',
  authMiddleware,
  rateLimitMiddleware,
  validateSearchInput,
  async (req, res) => {
    const result = await db.query(
      'SELECT * FROM users WHERE name LIKE $1',
      [`%${sanitize(req.query.name)}%`]
    );
    res.json(result);
  }
);"
```

### Severity: CRITICAL

### Validation Action
- Run security checklist for any code handling user input
- Check for auth middleware on new endpoints
- Status: REJECTED if security vulnerability detected

---

## Anti-Pattern Summary Table

| # | Anti-Pattern | Severity | Key Indicator |
|---|--------------|----------|---------------|
| 1 | Assumption-Based Planning | CRITICAL | "probably", "might", "should" |
| 2 | Symptom Patching | HIGH | Null checks, try-catch wrappers |
| 3 | Pattern Violation | MEDIUM | New utility when existing available |
| 4 | Incomplete Impact Analysis | CRITICAL | Callers not fully enumerated |
| 5 | Vague Implementation | HIGH | "improve", "refactor", "fix" |
| 6 | Missing Rollback Strategy | HIGH | DB/API changes without recovery |
| 7 | Phantom File Reference | CRITICAL | Non-existent files or wrong lines |
| 8 | Security Oversight | CRITICAL | User input without validation |

---

## Quick Validation Rules

1. **Every claim needs file:line evidence**
2. **No speculative language** ("probably", "might", "should")
3. **All callers must be enumerated**
4. **Specific changes, not vague descriptions**
5. **Root cause, not symptom**
6. **Use existing patterns/utilities**
7. **Rollback plan for high-risk changes**
8. **Security check for user input handling**
