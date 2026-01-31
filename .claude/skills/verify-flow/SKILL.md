---
name: verify-flow
description: Code flow verification workflow. Traces, verifies, validates, and documents code execution paths from entry points to implementation. Use for any code flow verification-related task.
---

# Code Flow Verifier

Systematically verify code changes by tracing execution flow from entry points through the entire call chain. This skill ensures all parameters are correctly passed, return values are properly handled, and error cases are covered.

## Core Principle: Evidence-Based Verification

**CRITICAL**: Never guess or assume. Every finding must have code evidence.

```
❌ WRONG: "This function might possibly return null"

✅ CORRECT: "This function returns null"
   - Evidence: `service.ts:45` - `return null` statement exists
   - Trigger: When `if (!data)` condition at service.ts:42
```

**Every verification finding MUST include:**
1. **File:Line** - Exact location
2. **Code snippet** - The actual problematic code
3. **Trigger condition** - When the issue occurs
4. **Call chain** - How execution reaches this point

---

## VERT Framework

Execute these phases sequentially. Do not skip phases.

### Phase 0: VALIDATE Scope

**Goal**: Determine what to verify and set verification boundaries.

1. **Identify verification target**:
   - If verifying code changes: Use `git diff` or provided file list
   - If verifying a plan: Read the plan file to extract affected files/functions
   - If user specifies: Use the provided scope

2. **List changed functions/methods**:
   ```
   Use Grep to find function definitions in changed files
   Pattern: "function |const .* = |class |def |async function"
   ```

3. **Assess complexity**:
   - **Simple** (1-2 functions, single file): Proceed to quick verification
   - **Complex** (3+ functions, multiple files): Execute full VERT

**Output**: List of functions to verify with their file locations

---

### Phase 1: ENTRY Point Discovery

**Goal**: Find where code execution begins for each changed function.

1. **For each changed function**, use Grep to find callers:
   ```
   Search for: functionName(
   Exclude: function definition itself, test files, comments
   ```

2. **Trace upward** until you find an entry point:
   - External API endpoints (routes, handlers)
   - CLI commands
   - Event listeners
   - Public exported functions
   - Scheduled jobs / cron

3. **Document each entry point**:
   ```
   Entry Point: handleUserRequest
   File: src/api/routes.ts:45
   Type: API Endpoint (POST /users)
   Leads to: createUser() → validateInput() → saveToDb()
   ```

**Output**: Entry point list with call paths

---

### Phase 2: ROUTE Tracing

**Goal**: Trace execution from each entry point through the changed code.

For each entry point, perform forward tracing:

1. **Read the entry point function** using Read tool

2. **At each function call**, verify:

   | Check | How to Verify | Evidence Required |
   |-------|---------------|-------------------|
   | Parameter types | Compare caller args with callee params | Both signatures |
   | Parameter values | Check if values can be undefined/null | Null check presence |
   | Return handling | Verify caller uses return value correctly | Assignment or check |
   | Error handling | Confirm try-catch or error propagation | Error handling code |
   | Side effects | Identify mutations, I/O, global state | State changes |

3. **For branching logic**, trace ALL paths:
   - if/else branches
   - switch cases
   - early returns
   - error conditions

4. **Document the flow**:
   ```
   Step 1: handleRequest(req) → routes.ts:45
     - req.body passed to validateInput()
     - ✅ Type: object → object (correct)

   Step 2: validateInput(data) → validator.ts:12
     - data can be undefined if req.body is empty
     - ⚠️ No undefined check before accessing data.email
   ```

**Reference**: @references/trace-patterns.md for common patterns

---

### Phase 3: TRACE Verification

**Goal**: Classify and document all discovered issues.

1. **Classify each issue**:

   | Severity | Criteria | Examples |
   |----------|----------|----------|
   | CRITICAL | Will cause runtime error | Null dereference, type mismatch, missing import |
   | WARNING | May cause logic error | Unhandled edge case, missing validation |
   | INFO | Improvement opportunity | Unused variable, redundant check |

2. **For each issue, document**:
   ```
   [CRITICAL] Null pointer dereference
   - Location: validator.ts:15
   - Code: `data.email.toLowerCase()`
   - Trigger: When `data` is undefined (from empty request body)
   - Call chain: handleRequest → validateInput → (here)
   - Evidence: No null check at validator.ts:12-14
   ```

3. **Assess impact scope**:
   - Which entry points are affected?
   - Are there other callers?
   - Can this break existing functionality?

**Reference**: @references/verification-rules.md for verification criteria

---

### Phase 4: REPORT Generation

**Goal**: Generate actionable report for Claude Code to process.

Use the exact format from @references/report-template.md

**Key sections**:

1. **Summary**: Overall status (PASSED/WARNINGS/FAILED), counts

2. **Entry Point Analysis**: Flow path and issues per entry point

3. **Action Items**:
   - Table format with File, Line, Issue, Fix
   - Ordered by priority (CRITICAL first)

4. **Verification Checklist**: Post-fix verification items

**Important**: The Action Items section must be precise enough for Claude Code to:
- Navigate to the exact file and line
- Understand what needs to change
- Make the fix without additional investigation

---

## Invocation Patterns

### Verify recent changes
```
/verify-flow
```
Verifies uncommitted changes or most recent commit.

### Verify specific files
```
/verify-flow src/api/users.ts src/services/auth.ts
```
Verifies only the specified files.

### Verify a plan
```
/verify-flow --plan
```
Reads the current plan file and verifies the proposed changes.

---

## Quick Reference

**Tools to use**:
- `Grep`: Find function usages, callers
- `Read`: Read function implementations
- `Glob`: Find related files
- `Bash(git diff)`: Get changed files

**Never**:
- Guess about behavior without reading code
- Skip entry point discovery
- Report issues without code evidence
- Assume types without verification
