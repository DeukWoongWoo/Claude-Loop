# Test Quality Checklist

Use this checklist to validate tests before finalizing.

## Pre-Flight Checks

Before writing any test, answer these questions:

| Question | Required Answer |
|----------|-----------------|
| Does this code have logic worth testing? | Yes |
| Is this testing behavior, not implementation? | Yes |
| Will this test catch real bugs? | Yes |
| Is there existing test coverage for this? | No (or insufficient) |

If any answer is wrong, reconsider whether the test is needed.

---

## The 4 Core Questions

For every test, verify:

### 1. Mutation Resistance
> "If I introduce a bug into this code, will this test fail?"

**Pass**: Test would fail if:
- Off-by-one error introduced
- Condition flipped (< to >)
- Return value changed
- Exception not thrown when should be

**Fail**: Test passes regardless of code correctness

### 2. Behavior Focus
> "Am I testing behavior or implementation?"

**Pass**: Test verifies:
- Observable outputs
- State changes visible through public API
- Side effects (emails sent, files written)

**Fail**: Test verifies:
- Internal method calls
- Private state
- Specific SQL queries or API calls
- Order of internal operations

### 3. Failure Meaning
> "If this test fails, is it a real problem?"

**Pass**: Failure indicates:
- A bug in business logic
- A regression in expected behavior
- A broken contract

**Fail**: Failure indicates:
- Refactoring (behavior unchanged)
- Implementation detail changed
- Test setup issue

### 4. Mock Reasonableness
> "Are there too many mocks?"

**Pass**: Mocks only at:
- System boundaries (network, disk, time)
- External services
- Non-deterministic sources

**Fail**: Mocks for:
- Internal classes
- Value objects
- Domain logic
- More than 2-3 mocks per test

---

## Quality Attributes (FIRST)

### Fast
- [ ] Test runs in milliseconds, not seconds
- [ ] No unnecessary I/O operations
- [ ] No sleeps or arbitrary waits

### Independent
- [ ] Test can run in any order
- [ ] No shared mutable state with other tests
- [ ] No reliance on test execution sequence

### Repeatable
- [ ] Same result every time
- [ ] No flakiness from timing/concurrency
- [ ] No external dependencies without mocking

### Self-Validating
- [ ] Clear pass/fail (no manual inspection needed)
- [ ] Meaningful assertions present
- [ ] Failure message explains what went wrong

### Timely
- [ ] Written with or before the code
- [ ] Tests edge cases discovered during implementation

---

## Test Case Coverage

For each test target, verify coverage of:

### Input Categories
- [ ] **Valid inputs** - Normal expected values
- [ ] **Boundary values** - Min, max, limits
- [ ] **Empty/null values** - Empty strings, null, undefined
- [ ] **Invalid inputs** - Wrong types, out of range

### Execution Paths
- [ ] **Happy path** - Normal successful flow
- [ ] **Error paths** - Exception handling
- [ ] **Edge cases** - Unusual but valid scenarios

### State Transitions
- [ ] **Initial state** - Before operation
- [ ] **Final state** - After operation
- [ ] **Intermediate states** - If observable

---

## Assertion Quality

### Good Assertions
```typescript
// Specific and meaningful
expect(account.balance).toBe(150);
expect(() => withdraw(200)).toThrow('Insufficient funds');
expect(user.isActive).toBe(true);
```

### Bad Assertions
```typescript
// Too vague
expect(result).toBeDefined();
expect(result).toBeTruthy();
expect(array.length).toBeGreaterThan(0);

// Testing implementation
expect(mockService.save).toHaveBeenCalled();
expect(spy).toHaveBeenCalledWith(internalArgs);
```

---

## Naming Quality

### Good Test Names
- Describe the behavior being tested
- Include the condition and expected outcome
- Readable as a specification

```
✓ returns_empty_list_when_no_items_match_filter
✓ throws_insufficient_funds_when_balance_below_withdrawal
✓ sends_confirmation_email_after_successful_registration
```

### Bad Test Names
```
✗ test1
✗ testGetUser
✗ should_work
✗ handles_edge_case
```

---

## Final Checklist

Before committing tests:

- [ ] All 4 core questions answered positively
- [ ] FIRST attributes satisfied
- [ ] No anti-patterns present (see test-anti-patterns.md)
- [ ] Test names describe behavior
- [ ] Assertions are specific and meaningful
- [ ] Edge cases and error paths covered
- [ ] Tests run fast and reliably
- [ ] No flaky behavior observed

---

## Red Flags

Stop and reconsider if you see:

| Red Flag | Problem |
|----------|---------|
| More than 3 mocks | Likely testing implementation |
| Test file longer than source file | Over-testing |
| Testing private methods | Implementation coupling |
| No assertions | Assertion-free test |
| Shared test state | Dependent tests |
| Sleep/wait calls | Flaky test risk |
| Testing getters/setters | Testing trivial code |
