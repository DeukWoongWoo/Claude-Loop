# Code Flow Tracing Patterns

This document provides patterns and techniques for tracing code execution flow effectively.

---

## 1. Entry Point Discovery Patterns

### 1.1 Web API Entry Points

**Pattern**: HTTP route handlers

**Search strategy**:
```
Grep for: app.get|app.post|router.get|@Get|@Post|handleRequest
File types: *.ts, *.js, routes.*, controller.*
```

**Example trace**:
```
Entry: POST /api/users (routes/users.ts:25)
  └→ createUserHandler(req, res)
       └→ userService.createUser(req.body)
            └→ validateUser(userData)
            └→ db.users.insert(userData)
```

### 1.2 CLI Entry Points

**Pattern**: Command handlers

**Search strategy**:
```
Grep for: program.command|yargs.command|commander|process.argv
File types: cli.*, bin/*, commands/*
```

**Example trace**:
```
Entry: `myapp build` (cli/index.ts:15)
  └→ buildCommand.execute(args)
       └→ buildService.build(config)
            └→ compiler.compile(files)
```

### 1.3 Event Handler Entry Points

**Pattern**: Event listeners and callbacks

**Search strategy**:
```
Grep for: .on\(|.addEventListener|@OnEvent|subscribe
File types: *.ts, *.js
```

**Example trace**:
```
Entry: 'user.created' event (listeners/userEvents.ts:10)
  └→ onUserCreated(event)
       └→ notificationService.sendWelcome(event.user)
       └→ analyticsService.track('user_created')
```

### 1.4 Scheduled Job Entry Points

**Pattern**: Cron jobs and scheduled tasks

**Search strategy**:
```
Grep for: @Cron|schedule|setInterval|cron.schedule
File types: jobs/*, tasks/*, schedulers/*
```

---

## 2. Forward Tracing Techniques

### 2.1 Direct Call Tracing

**Goal**: Follow function calls in sequence

**Method**:
1. Start at entry point function
2. For each function call inside:
   a. Read the called function definition
   b. Note parameters passed
   c. Recursively trace nested calls
3. Stop at leaf functions (no further calls)

**Diagram**:
```
main() ─────────────────────────────────────────────────►
  │
  ├─► validateInput(data) ─────────────────────────────►
  │     │
  │     └─► checkSchema(data) ───────────────────────►
  │     │
  │     └─► sanitize(data) ──────────────────────────►
  │
  └─► processData(data) ───────────────────────────────►
        │
        └─► transform(data) ─────────────────────────►
        │
        └─► save(data) ──────────────────────────────►
```

### 2.2 Async Call Tracing

**Goal**: Track Promise chains and async/await flow

**Method**:
1. Identify async functions (async keyword, returns Promise)
2. Track await points
3. Verify Promise chain continuations (.then, .catch)
4. Check for parallel execution (Promise.all)

**Example trace**:
```
async function processOrder(orderId) {
  const order = await getOrder(orderId)      // await point 1
  const validated = await validate(order)     // await point 2
  const [inventory, payment] = await Promise.all([  // parallel
    checkInventory(order),
    processPayment(order)
  ])
  return await createShipment(order, inventory)  // await point 3
}
```

### 2.3 Callback Tracing

**Goal**: Track callback-based async patterns

**Method**:
1. Identify callback parameters
2. Trace to where callback is invoked
3. Note callback arguments passed

**Example**:
```javascript
// Caller
fetchData(url, (error, data) => {
  if (error) handleError(error)
  else processData(data)
})

// Callee - find where callback is invoked
function fetchData(url, callback) {
  http.get(url, (response) => {
    callback(null, response.body)  // Callback invoked here
  }).on('error', (err) => {
    callback(err, null)             // Error callback here
  })
}
```

---

## 3. Backward Tracing Techniques

### 3.1 Caller Discovery

**Goal**: Find all places where a function is called

**Method**:
1. Grep for function name followed by `(`
2. Filter out: definitions, test files, comments
3. Read each call site for context

**Search command**:
```
Grep: functionName\(
Exclude: "function functionName", "// functionName", test files
```

### 3.2 Data Origin Tracing

**Goal**: Trace where a value comes from

**Method**:
1. Start at variable usage
2. Trace back through assignments
3. Continue to function parameters
4. Follow to caller's arguments
5. Repeat until origin found

**Example trace (backward)**:
```
user.email.toLowerCase()      // Where does user come from?
  ↑
const user = await getUser(id)  // From getUser return value
  ↑
function getUser(id) { return db.users.findById(id) }  // From database
```

---

## 4. Branch Path Analysis

### 4.1 Conditional Branch Tracing

**Goal**: Verify all branches are traced

**Method**:
1. Identify conditionals (if, switch, ternary)
2. Trace each branch path
3. Note early returns
4. Document unreachable code

**Template**:
```
Condition at line 25: if (user.isAdmin)
  ├─ TRUE path: lines 26-35
  │   └─► adminHandler(user)
  │
  └─ FALSE path: lines 37-45
      └─► userHandler(user)

Condition at line 40: if (!data) return null  // Early return
  ├─ TRUE path: return null (terminates)
  │
  └─ FALSE path: continues to line 42
```

### 4.2 Exception Path Tracing

**Goal**: Track error propagation paths

**Method**:
1. Identify throw statements
2. Trace to catch blocks
3. Verify error handling at each level
4. Check for unhandled exceptions

**Template**:
```
throw new ValidationError()  (validator.ts:25)
  │
  ├─► Caught at: service.ts:50 try-catch
  │     └─► Re-thrown as: ServiceError
  │
  └─► Caught at: controller.ts:30 try-catch
        └─► Converted to: HTTP 400 response
```

---

## 5. Data Flow Analysis

### 5.1 Value Transformation Tracking

**Goal**: Track how data changes through the flow

**Method**:
1. Start with input data shape
2. Note each transformation
3. Document shape at each step
4. Verify final shape matches expectation

**Template**:
```
Input: { name: string, email: string }

Step 1: validate(input)
  Output: { name: string, email: string } // unchanged

Step 2: transform(input)
  Output: {
    name: string,
    email: string,
    createdAt: Date,     // added
    id: string           // added
  }

Step 3: sanitize(input)
  Output: {
    name: string,        // trimmed
    email: string,       // lowercased
    createdAt: Date,
    id: string
  }
```

### 5.2 Type Narrowing Tracking

**Goal**: Track TypeScript type narrowing

**Method**:
1. Note initial type
2. Track type guards and narrowing
3. Verify narrowed type at usage point

**Example**:
```typescript
function process(input: string | null) {  // Type: string | null
  if (!input) return                        // Type narrowed: null handled

  const trimmed = input.trim()              // Type: string (safe)
  return trimmed.toLowerCase()
}
```

---

## 6. Verification Points

At each step in the trace, verify:

| Point | Verification | Tool |
|-------|-------------|------|
| Function call | Parameters match signature | Read both |
| Return value | Captured and handled | Read caller |
| Error | Caught or propagated | Check try-catch |
| Null | Checked before use | Check guard |
| Type | Compatible with expectation | Compare types |
| Async | Properly awaited | Check await |

---

## 7. Trace Documentation Format

For each traced path, document:

```markdown
## Trace: [Entry Point Name]

**Entry**: `functionName(params)` at file.ts:line

**Flow**:
1. `step1()` at file1.ts:10
   - Input: description
   - Output: description
   - Status: ✅ / ⚠️ / ❌

2. `step2()` at file2.ts:20
   - Input: description
   - Output: description
   - Status: ✅ / ⚠️ / ❌

**Issues Found**:
- [SEVERITY] Description at location
  - Evidence: code snippet
  - Fix: suggested change
```
