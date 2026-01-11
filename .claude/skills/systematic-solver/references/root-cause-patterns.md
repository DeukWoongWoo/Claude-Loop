# Root Cause Analysis Patterns

Common bug categories and their typical root causes vs surface symptoms.

## The 5 Whys Technique

Keep asking "why" until you reach something you can directly fix.

### Example: API Returns 500 Error

```
Symptom: API returns 500 error on user registration

Why 1: Database INSERT throws exception
Why 2: Unique constraint violation on email column
Why 3: Email normalization differs between check and insert
Why 4: Check uses toLowerCase(), insert uses raw input
Root Cause: Inconsistent email normalization

Surface Fix: Catch the exception and return "email exists"
Root Fix: Normalize email consistently at entry point
```

### Example: Test Fails Intermittently

```
Symptom: Test passes locally, fails in CI sometimes

Why 1: Assertion fails on timestamp comparison
Why 2: Expected and actual differ by milliseconds
Why 3: Test creates data, then immediately queries
Why 4: Database and test use different time sources
Root Cause: Race condition between write and read

Surface Fix: Add retry or increase tolerance
Root Fix: Use deterministic time injection in tests
```

---

## Common Bug Categories

### Null/Undefined Reference Errors

| Surface Symptom | Common Root Causes |
|-----------------|-------------------|
| NullPointerException | Uninitialized state, async timing, missing validation |
| "Cannot read property of undefined" | Optional chain missing, wrong data shape |
| Empty array/object errors | Failed fetch, wrong API response format |

**Questions to Ask:**
- Where should this value have been set?
- What code path skipped initialization?
- Is this an async timing issue?

### Data Inconsistency

| Surface Symptom | Common Root Causes |
|-----------------|-------------------|
| Wrong data displayed | Stale cache, race condition, wrong query |
| Data corruption | Concurrent writes, partial updates, no transactions |
| Duplicate records | Missing unique constraints, retry without idempotency |

**Questions to Ask:**
- Where is the single source of truth?
- Are writes atomic and consistent?
- Could concurrent operations cause this?

### Performance Issues

| Surface Symptom | Common Root Causes |
|-----------------|-------------------|
| Slow page load | N+1 queries, missing indexes, large payloads |
| Memory leak | Event listeners not removed, closure references, caching |
| High CPU | Unnecessary re-renders, infinite loops, expensive calculations |

**Questions to Ask:**
- What's the actual bottleneck? (measure, don't guess)
- Is this O(n), O(n²), or worse?
- Are we doing work that could be cached or skipped?

### Authentication/Authorization Bugs

| Surface Symptom | Common Root Causes |
|-----------------|-------------------|
| Access denied unexpectedly | Wrong role check, stale token, permission misconfiguration |
| Unauthorized access | Missing auth check, broken middleware chain |
| Session issues | Cookie settings, token expiry, concurrent login handling |

**Questions to Ask:**
- Where exactly is authorization checked?
- What's the full middleware/filter chain?
- Are all entry points protected?

### Integration Failures

| Surface Symptom | Common Root Causes |
|-----------------|-------------------|
| External API errors | Rate limiting, auth expiry, API version change |
| Timeout errors | Network issues, slow response, missing retry |
| Data format errors | Schema drift, encoding issues, null vs missing |

**Questions to Ask:**
- What changed recently on either side?
- Are we handling all response codes?
- Do we have proper error boundaries?

---

## Surface Fix vs Root Fix Examples

### Example 1: Form Validation

```
Bug: User can submit form with invalid email

Surface Fix:
+ if (!email.includes('@')) return false;

Root Fix:
- Implement proper email validation at form component
- Add server-side validation as backup
- Create reusable validation utility
- Add tests for edge cases
```

### Example 2: Race Condition

```
Bug: Double-click causes duplicate order

Surface Fix:
+ let isSubmitting = false;
+ if (isSubmitting) return;
+ isSubmitting = true;

Root Fix:
- Use idempotency keys for orders
- Implement proper loading states
- Add database unique constraint
- Handle duplicates gracefully on backend
```

### Example 3: Performance

```
Bug: User list page is slow

Surface Fix:
+ Add pagination (limit 20)

Root Fix:
- Profile to find actual bottleneck
- If N+1: Add eager loading
- If indexing: Add database indexes
- If rendering: Virtualize list
- Add pagination as UX improvement
```

---

## Root Cause Validation Checklist

Before concluding root cause analysis:

1. [ ] Can I point to the exact line(s) causing the issue?
2. [ ] Does fixing this prevent the symptom from recurring?
3. [ ] Have I traced back far enough? (could there be a deeper cause?)
4. [ ] Would this explain ALL observed symptoms?
5. [ ] Is there evidence (code, logs, data) supporting each "why"?

---

## Anti-Patterns in Root Cause Analysis

### Stopping Too Early
```
Bad: "The error is a NullPointerException" (that's the symptom)
Good: "The user object is null because the auth middleware didn't run"
Better: "Auth middleware is skipped for this route due to wrong matcher pattern"
```

### Assuming Without Evidence
```
Bad: "It's probably a caching issue"
Good: "Cache TTL is 1 hour, but data changed 30 minutes ago and is stale"
```

### Fixing Symptoms in a Loop
```
Bad: Add null check → new null error → add another null check → ...
Good: Find where the null originates and fix initialization
```

### Blaming External Factors
```
Bad: "The API is broken" (maybe, but verify)
Good: "Our request format changed in commit X, breaking compatibility"
```
