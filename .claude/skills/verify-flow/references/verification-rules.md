# Verification Rules

This document defines the rules for verifying code flow. Each rule includes what to check, how to verify, and examples of issues.

---

## 1. Parameter Passing Rules

### 1.1 Type Consistency

**Rule**: Parameters passed to a function must match the expected types.

**How to verify**:
1. Read the function signature (callee)
2. Read the call site (caller)
3. Compare argument types with parameter types

**Evidence required**:
```
Caller: processUser(userData)     // userData: unknown
Callee: function processUser(user: User) { ... }
Issue: Type mismatch - unknown passed where User expected
```

**Common issues**:
- `any` or `unknown` passed to typed parameter
- `undefined` passed to non-optional parameter
- Object with missing required properties
- Array passed where single item expected

### 1.2 Null/Undefined Safety

**Rule**: Nullable values must be checked before use.

**How to verify**:
1. Trace the value origin
2. Check if null/undefined is possible
3. Verify guard clause exists before usage

**Evidence required**:
```
Line 10: const user = await getUser(id)  // Returns User | null
Line 15: user.name.toLowerCase()          // No null check
Issue: Potential null dereference at line 15
```

**Common issues**:
- Optional parameter used without check
- Async result used without null check
- Array access without bounds check
- Object property access without existence check

### 1.3 Required vs Optional

**Rule**: Required parameters must always be provided with valid values.

**How to verify**:
1. Check function signature for required params
2. Trace all call sites
3. Verify each call provides the required value

**Evidence required**:
```
Signature: function save(data: Data, options: Options)
Call site: save(data)  // Missing second argument
Issue: Required parameter 'options' not provided
```

---

## 2. Return Value Rules

### 2.1 Return Value Handling

**Rule**: Return values that indicate state must be handled.

**How to verify**:
1. Check if function returns meaningful value
2. Verify caller captures and uses return value
3. Check for ignored error indicators

**Evidence required**:
```
Line 20: validateInput(data)  // Returns boolean, result ignored
Line 25: processData(data)    // Processes invalid data
Issue: Validation result not checked before processing
```

**Return types that MUST be handled**:
- Boolean (success/failure)
- Error/Result types
- Nullable values
- Promises (must be awaited or returned)

### 2.2 Promise Handling

**Rule**: All Promises must be properly awaited or returned.

**How to verify**:
1. Identify async functions
2. Check if Promise is awaited or returned
3. Verify error handling with try-catch or .catch()

**Common issues**:
- Missing `await` keyword
- Unhandled Promise rejection
- Fire-and-forget async calls without error handling

---

## 3. Error Handling Rules

### 3.1 Error Propagation

**Rule**: Errors must be either handled or propagated correctly.

**How to verify**:
1. Identify operations that can throw
2. Check for try-catch or error callback
3. Verify error is logged, handled, or re-thrown

**Evidence required**:
```
Line 30: const result = JSON.parse(input)  // Can throw
// No try-catch wrapper
Issue: Uncaught exception possible at line 30
```

**Operations that commonly throw**:
- JSON.parse/stringify
- File I/O
- Network requests
- Database operations
- User input parsing

### 3.2 Error Type Preservation

**Rule**: When re-throwing, error context should be preserved.

**How to verify**:
1. Find catch blocks
2. Check if original error is included when re-throwing
3. Verify error message includes context

**Good pattern**:
```typescript
try {
  await operation()
} catch (error) {
  throw new ServiceError('Operation failed', { cause: error })
}
```

---

## 4. State Mutation Rules

### 4.1 Immutability Expectations

**Rule**: Values expected to be immutable should not be mutated.

**How to verify**:
1. Identify parameters marked as readonly or const
2. Check if function modifies the input
3. Verify mutations are intentional

**Evidence required**:
```
Signature: function process(items: readonly Item[])
Line 15: items.push(newItem)  // Mutates readonly array
Issue: Mutation of readonly parameter
```

### 4.2 Side Effect Documentation

**Rule**: Functions with side effects should be clearly identified.

**Side effects to check**:
- Global state modification
- File system operations
- Network calls
- Database writes
- Console output
- Event emission

---

## 5. Branching Logic Rules

### 5.1 All Paths Return

**Rule**: All code paths must return a value or throw.

**How to verify**:
1. Identify all branches (if/else, switch, early returns)
2. Verify each path has explicit return
3. Check for missing default case

**Evidence required**:
```
function getStatus(code: number): string {
  if (code === 200) return 'OK'
  if (code === 404) return 'Not Found'
  // Missing return for other codes
}
Issue: Not all code paths return a value
```

### 5.2 Exhaustive Matching

**Rule**: Switch/match statements should handle all cases.

**How to verify**:
1. Identify discriminated unions or enums
2. Check switch statement cases
3. Verify default case handles unknown values appropriately

---

## 6. Severity Classification Matrix

| Category | CRITICAL | WARNING | INFO |
|----------|----------|---------|------|
| Type | Type mismatch causing runtime error | Loose typing (`any`) | Type could be more specific |
| Null | Null dereference | Optional without check | Redundant null check |
| Return | Ignored error return | Unused return value | Return could be void |
| Error | Uncaught exception | Catch without handle | Generic catch |
| State | Unexpected mutation | Hidden side effect | Could be pure function |
| Branch | Missing return path | Missing default case | Redundant branch |

---

## 7. Evidence Requirements Checklist

For each issue, include:

- [ ] **File path and line number**
- [ ] **Exact code snippet** (copy from source)
- [ ] **Trigger condition** (when does this fail?)
- [ ] **Call chain** (how does execution reach here?)
- [ ] **Impact assessment** (what breaks?)
- [ ] **Fix suggestion** (specific change needed)
