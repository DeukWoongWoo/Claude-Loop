# Test Anti-Patterns

Patterns that make tests harmful rather than helpful. Avoid these.

## Critical Anti-Patterns (Ranked by Harm)

### 1. Flaky Tests (MOST HARMFUL)

Tests that fail intermittently without code changes.

**Why it's harmful:**
- Destroys team confidence in the test suite
- Teams start ignoring failures ("probably flaky")
- Red builds become meaningless

**Signs:**
- Test passes locally, fails in CI (or vice versa)
- Test fails on retry without code changes
- Test depends on timing, network, or external state

**Causes:**
- Shared mutable state between tests
- Timing dependencies (sleeps, race conditions)
- External service dependencies
- Order-dependent tests

**Fix:**
- Isolate tests completely
- Mock external dependencies
- Use deterministic data
- Never rely on sleep/timing

---

### 2. Over-Mocking (Mockery)

So many mocks that you're testing mocks, not the system.

**Example (BAD):**
```typescript
test('processes order', () => {
  const mockDb = mock(Database);
  const mockPayment = mock(PaymentService);
  const mockEmail = mock(EmailService);
  const mockInventory = mock(InventoryService);
  const mockLogger = mock(Logger);

  // What are we even testing here?
  // The mocks return what we told them to return
});
```

**Why it's harmful:**
- Tests pass but production fails
- Refactoring breaks tests even when behavior unchanged
- False confidence in test coverage

**Fix:**
- Use real implementations where possible
- Use fakes (in-memory implementations) over mocks
- Only mock at system boundaries (network, disk, time)

---

### 3. Happy Path Only

Never testing edge cases, boundaries, or error conditions.

**Example (INCOMPLETE):**
```typescript
// Only tests success case
test('calculates discount', () => {
  expect(calculateDiscount(100, 10)).toBe(90);
});

// Missing tests:
// - What if amount is 0?
// - What if discount > 100%?
// - What if inputs are negative?
// - What if inputs are not numbers?
```

**Why it's harmful:**
- Bugs live in the unhappy paths
- Production errors that tests don't catch
- False sense of security

**Fix:**
- Always test boundaries and edge cases
- Test error conditions and exception paths
- Ask "what could go wrong?"

---

### 4. Testing Implementation Details (Inspector)

Tests that know too much about internal implementation.

**Example (BAD):**
```typescript
test('saves user', () => {
  const user = new User('Alice');
  user.save();

  // Testing internal state/implementation
  expect(user._isDirty).toBe(false);
  expect(user._lastSaved).toBeDefined();
  expect(mockDb.query).toHaveBeenCalledWith(
    'INSERT INTO users (name) VALUES (?)', ['Alice']
  );
});
```

**Why it's harmful:**
- Breaks on any refactoring
- Tests become maintenance burden
- Creates fear of improving code

**Fix:**
```typescript
// Test observable behavior instead
test('saved user can be retrieved', () => {
  const user = new User('Alice');
  user.save();

  const retrieved = User.findByName('Alice');
  expect(retrieved.name).toBe('Alice');
});
```

---

### 5. Assertion-Free Tests

Tests that run code but don't verify anything meaningful.

**Example (WORTHLESS):**
```typescript
test('processes data', () => {
  const processor = new DataProcessor();
  processor.process(data);
  // No assertions! What did we test?
});

test('renders component', () => {
  render(<MyComponent />);
  // Just verifies it doesn't throw
});
```

**Why it's harmful:**
- Can't detect bugs (no assertions)
- Inflates coverage metrics meaninglessly
- Provides false confidence

**Fix:**
- Every test must assert on observable outcomes
- Ask "what should be different after this runs?"

---

### 6. Chained/Dependent Tests

Tests that depend on other tests running first.

**Example (BAD):**
```typescript
let sharedUser;

test('creates user', () => {
  sharedUser = createUser('Alice');
  expect(sharedUser).toBeDefined();
});

test('updates user', () => {
  // Depends on previous test!
  sharedUser.name = 'Bob';
  sharedUser.save();
  expect(sharedUser.name).toBe('Bob');
});
```

**Why it's harmful:**
- Can't run tests in isolation
- Parallel execution breaks
- One failure cascades to many

**Fix:**
- Each test sets up its own data
- No shared mutable state
- Tests must pass when run alone

---

### 7. Giant Tests

Single tests with many assertions testing multiple things.

**Example (BAD):**
```typescript
test('user workflow', () => {
  // Registration
  const user = register('alice@test.com', 'password');
  expect(user).toBeDefined();
  expect(user.email).toBe('alice@test.com');

  // Login
  const session = login('alice@test.com', 'password');
  expect(session.token).toBeDefined();

  // Profile update
  user.updateProfile({ name: 'Alice' });
  expect(user.name).toBe('Alice');

  // ... 50 more assertions
});
```

**Why it's harmful:**
- Hard to identify what failed
- Slow to run
- Difficult to maintain

**Fix:**
- One logical assertion per test
- Separate tests for each behavior
- Use descriptive test names

---

### 8. Testing Trivial Code

Testing code that has no logic.

**Example (WORTHLESS):**
```typescript
class User {
  constructor(public name: string) {}
  getName() { return this.name; }
}

// Don't test this!
test('getName returns name', () => {
  const user = new User('Alice');
  expect(user.getName()).toBe('Alice');
});
```

**Why it's harmful:**
- Wastes time writing and maintaining
- Adds to test suite without value
- Inflates coverage metrics meaninglessly

**Fix:**
- Only test code that has logic
- Skip pure data accessors
- Focus on business rules

---

## Summary Table

| Anti-Pattern | Detection | Fix |
|--------------|-----------|-----|
| Flaky Tests | Intermittent failures | Isolate, mock externals |
| Over-Mocking | 3+ mocks per test | Use fakes, mock boundaries only |
| Happy Path Only | No edge/error tests | Add boundary tests |
| Implementation Testing | Tests private/internal | Test public behavior |
| Assertion-Free | No expect/assert | Add meaningful assertions |
| Chained Tests | Shared mutable state | Isolate each test |
| Giant Tests | Many assertions | Split into focused tests |
| Testing Trivial | No logic in code | Don't test |
