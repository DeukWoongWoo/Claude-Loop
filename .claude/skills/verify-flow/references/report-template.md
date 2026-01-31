# Report Template

Use this exact template for verification reports. The format is optimized for Claude Code to parse and execute fixes automatically.

---

## Template

```markdown
# Code Flow Verification Report

## Summary

- **Status**: [STATUS]
- **Entry Points Analyzed**: [N]
- **Issues Found**: Critical: [X], Warning: [Y], Info: [Z]

### Verification Scope
- **Target**: [What was verified - files, plan, or changes]
- **Changed Functions**: [List of functions verified]

---

## Entry Point Analysis

### Entry Point 1: `[functionName]` ([file:line])

**Type**: [API Endpoint / CLI Command / Event Handler / Job]

**Flow Path**:
1. `[function1](args)` → [file.ts:line]
2. `[function2](args)` → [file.ts:line]
3. `[function3](args)` → [file.ts:line]

**Parameter Flow**:
| Step | Caller | Callee | Expected | Actual | Status |
|------|--------|--------|----------|--------|--------|
| 1→2  | func1  | func2  | string   | string | ✅     |
| 2→3  | func2  | func3  | object   | object | ✅     |

**Return Value Handling**:
| Function | Returns | Handled | Status |
|----------|---------|---------|--------|
| func1    | Promise<T> | await | ✅   |
| func2    | boolean | checked | ✅   |

**Issues Found**:

#### [CRITICAL] [Brief description]
- **Location**: `[file.ts:line]`
- **Code**:
  ```typescript
  [actual code snippet]
  ```
- **Trigger**: [When this issue occurs]
- **Call Chain**: [How execution reaches here]
- **Evidence**: [Why this is a problem, reference to code]
- **Fix**: [Specific change needed]

#### [WARNING] [Brief description]
- **Location**: `[file.ts:line]`
- **Code**:
  ```typescript
  [actual code snippet]
  ```
- **Trigger**: [When this issue occurs]
- **Suggestion**: [Recommended improvement]

---

### Entry Point 2: `[functionName]` ([file:line])

[Same structure as Entry Point 1]

---

## Action Items

This section is structured for Claude Code to process automatically.

### CRITICAL (Immediate Fix Required)

| # | File | Line | Issue | Fix |
|---|------|------|-------|-----|
| 1 | [file.ts] | [line] | [Concise issue description] | [Specific fix instruction] |
| 2 | [file.ts] | [line] | [Concise issue description] | [Specific fix instruction] |

### WARNING (Review Required)

| # | File | Line | Issue | Suggestion |
|---|------|------|-------|------------|
| 1 | [file.ts] | [line] | [Issue] | [Suggestion] |

### INFO (For Reference)

- `[file.ts:line]`: [Description]

---

## Verification Checklist

After applying fixes, verify:

- [ ] All CRITICAL issues addressed
- [ ] All WARNING issues reviewed
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Entry points re-verified
- [ ] No new issues introduced

---

## Additional Recommendations

1. **[Category]**: [Recommendation description]
2. **[Category]**: [Recommendation description]
```

---

## Status Values

| Status | Symbol | Meaning |
|--------|--------|---------|
| PASSED | ✅ | No issues found |
| WARNINGS | ⚠️ | Non-critical issues found |
| FAILED | ❌ | Critical issues found |

---

## Severity Markers

Use these exact markers for consistency:

- `[CRITICAL]` - Must fix before proceeding
- `[WARNING]` - Should review and consider fixing
- `[INFO]` - Informational, no action required

---

## Action Items Guidelines

The Action Items table must be:

1. **Precise**: Exact file path and line number
2. **Actionable**: Specific fix instruction, not vague suggestion
3. **Ordered**: CRITICAL first, then WARNING, then INFO
4. **Complete**: Each issue has a corresponding action

**Good Action Item**:
```
| 1 | src/service.ts | 45 | Missing null check for `user` | Add `if (!user) return null;` before line 45 |
```

**Bad Action Item**:
```
| 1 | service.ts | somewhere | Might have issues | Consider checking |
```

---

## Example Report

```markdown
# Code Flow Verification Report

## Summary

- **Status**: ❌ FAILED
- **Entry Points Analyzed**: 2
- **Issues Found**: Critical: 1, Warning: 2, Info: 1

### Verification Scope
- **Target**: Uncommitted changes
- **Changed Functions**: createUser, validateEmail

---

## Entry Point Analysis

### Entry Point 1: `createUserHandler` (src/routes/users.ts:25)

**Type**: API Endpoint (POST /api/users)

**Flow Path**:
1. `createUserHandler(req, res)` → src/routes/users.ts:25
2. `userService.createUser(req.body)` → src/services/user.ts:42
3. `validateEmail(email)` → src/utils/validation.ts:15

**Parameter Flow**:
| Step | Caller | Callee | Expected | Actual | Status |
|------|--------|--------|----------|--------|--------|
| 1→2  | handler | createUser | UserDTO | any | ⚠️ |
| 2→3  | createUser | validateEmail | string | string | ✅ |

**Issues Found**:

#### [CRITICAL] Null pointer dereference
- **Location**: `src/utils/validation.ts:18`
- **Code**:
  ```typescript
  return email.toLowerCase().includes('@')
  ```
- **Trigger**: When `email` is undefined (from missing request body field)
- **Call Chain**: createUserHandler → createUser → validateEmail
- **Evidence**: No null check at validation.ts:15-17
- **Fix**: Add null check: `if (!email) return false;`

#### [WARNING] Loose typing
- **Location**: `src/routes/users.ts:27`
- **Code**:
  ```typescript
  userService.createUser(req.body)  // body is any
  ```
- **Suggestion**: Type req.body as UserDTO

---

## Action Items

### CRITICAL (Immediate Fix Required)

| # | File | Line | Issue | Fix |
|---|------|------|-------|-----|
| 1 | src/utils/validation.ts | 15 | Missing null check for `email` | Add `if (!email) return false;` at line 15 |

### WARNING (Review Required)

| # | File | Line | Issue | Suggestion |
|---|------|------|-------|------------|
| 1 | src/routes/users.ts | 27 | `req.body` is untyped | Add type assertion: `req.body as UserDTO` |
| 2 | src/services/user.ts | 45 | No error handling for db operation | Wrap in try-catch |

### INFO (For Reference)

- `src/services/user.ts:52`: Unused variable `temp`

---

## Verification Checklist

After applying fixes, verify:

- [ ] All CRITICAL issues addressed
- [ ] All WARNING issues reviewed
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Entry points re-verified
- [ ] No new issues introduced
```
