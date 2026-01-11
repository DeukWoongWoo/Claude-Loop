# Impact Analysis Guide

Systematic approach to understanding what is affected by a code change.

## The Three Questions

Before any modification, answer:

1. **What calls this code?** (Who depends on me?)
2. **What does this code call?** (What do I depend on?)
3. **What data flows through?** (What state is affected?)

---

## Step 1: Direct Impact (Files Being Modified)

List all files that will be directly changed:

```
Direct Changes:
- src/utils/validator.ts (modifying validation logic)
- src/types/user.ts (changing interface)
```

---

## Step 2: Caller Analysis (Upstream Dependencies)

Find everything that uses the code being modified.

### Search Strategies

**For Functions:**
```bash
# Search for function calls
grep -r "functionName(" --include="*.ts"

# Search for imports
grep -r "import.*functionName" --include="*.ts"
```

**For Classes:**
```bash
# Search for instantiation
grep -r "new ClassName" --include="*.ts"

# Search for inheritance
grep -r "extends ClassName" --include="*.ts"

# Search for type usage
grep -r ": ClassName" --include="*.ts"
```

**For Exports:**
```bash
# Find all importers of a module
grep -r "from './module'" --include="*.ts"
grep -r "from '../path/module'" --include="*.ts"
```

### Caller Categories

| Category | Risk | Example |
|----------|------|---------|
| Public API endpoints | High | Routes that call this function |
| Other internal modules | Medium | Services using this utility |
| Tests | Low (but important) | Unit tests covering this code |
| Scripts/tools | Low | Build scripts, migrations |

---

## Step 3: Callee Analysis (Downstream Dependencies)

Understand what the modified code relies on.

### What to Check

- External libraries (version compatibility?)
- Internal utilities (will they still work?)
- Database queries (schema assumptions?)
- API calls (contract still valid?)
- Environment variables (still available?)

### Risk Assessment

Changing code that depends on external systems requires extra caution:

```
High Risk Dependencies:
- Database schemas
- External API contracts
- Authentication services
- Payment processors

Medium Risk Dependencies:
- Internal shared utilities
- Configuration files
- Cache systems

Low Risk Dependencies:
- UI utilities
- Formatting helpers
- Local state
```

---

## Step 4: Data Flow Analysis

Trace how data moves through the affected code.

### Questions to Answer

1. **Input sources:** Where does data come from?
   - User input
   - Database queries
   - API responses
   - Configuration

2. **Transformations:** How is data modified?
   - Validation
   - Normalization
   - Calculation
   - Formatting

3. **Output destinations:** Where does data go?
   - Database writes
   - API requests
   - UI rendering
   - Logs

### Example Data Flow

```
User Form Input
    ↓
[validateEmail()] ← MODIFYING THIS
    ↓
API Request Body
    ↓
Database INSERT
    ↓
Confirmation Email
```

Modifying `validateEmail()` affects:
- What gets sent to API
- What gets stored in DB
- What email addresses receive confirmations

---

## Step 5: Test Coverage Analysis

### Find Relevant Tests

```bash
# Tests that import modified module
grep -r "import.*from.*modified-module" --include="*.test.ts"

# Tests that mention function name
grep -r "describe.*FunctionName\|it.*functionName" --include="*.test.ts"
```

### Test Categories

| Category | Action Needed |
|----------|--------------|
| Tests that will break | Must update with new behavior |
| Tests that should break but don't | Gap in coverage, add tests |
| Tests unaffected | Verify they still pass |

---

## Risk Assessment Matrix

| Factor | Low | Medium | High |
|--------|-----|--------|------|
| **Caller count** | 0-1 | 2-5 | 6+ |
| **Domain** | UI, formatting | Business logic | Auth, payments, data |
| **Test coverage** | High (>80%) | Medium (50-80%) | Low (<50%) |
| **Change scope** | Additive only | Modification | Breaking change |
| **Reversal difficulty** | Easy rollback | Needs migration | Data affected |

### Overall Risk Calculation

- Any **High** factor → High risk overall
- Multiple **Medium** factors → High risk overall
- All **Low** factors → Low risk overall

---

## Impact Report Template

```markdown
## Impact Analysis

### Direct Changes
- [ ] file1.ts: [description of change]
- [ ] file2.ts: [description of change]

### Upstream Impact (Callers)
- [ ] caller1.ts:45 - uses function X
- [ ] caller2.ts:102 - imports type Y
- [ ] api/route.ts:30 - exposes as endpoint

### Downstream Impact (Dependencies)
- [ ] Relies on database table 'users'
- [ ] Calls external API 'auth-service'

### Data Flow
- Input: [source]
- Output: [destination]
- Side effects: [mutations, logs, external calls]

### Test Coverage
- Existing tests: [list]
- Tests needing update: [list]
- New tests needed: [list]

### Risk Assessment
- Risk Level: [High/Medium/Low]
- Risk Factors:
  - [factor 1]
  - [factor 2]

### Verification Plan
- [ ] Run existing tests
- [ ] Manual test [scenario]
- [ ] Check [integration point]
```

---

## Common Blind Spots

### Often Missed Impacts

1. **Error handling paths**
   - Changes might break catch blocks elsewhere
   - Error messages used in UI or logs

2. **Default values**
   - Code relying on old defaults
   - Configuration fallbacks

3. **Implicit contracts**
   - Expected return shapes
   - Side effect assumptions
   - Timing expectations

4. **Build and deploy**
   - Environment variable changes
   - Configuration file updates
   - Migration requirements

5. **Documentation**
   - API docs
   - README examples
   - Comments referencing changed code

### Prevention Checklist

Before finalizing impact analysis:

- [ ] Searched for ALL usages (function name, class name, type name)
- [ ] Checked test files
- [ ] Reviewed imports/exports
- [ ] Considered error handling paths
- [ ] Checked documentation references
- [ ] Identified required migrations or config changes
