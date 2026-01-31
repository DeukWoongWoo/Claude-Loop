---
name: test-writer
description: Test development workflow. Analyzes, designs, writes, and validates meaningful tests that catch real bugs through behavior testing and mutation resistance. Use for any test writing-related task.
---

# Test Writer

Write meaningful tests that catch real bugs, not tests that just increase coverage numbers.

## Core Philosophy

**The Litmus Test**: "If I introduce a bug into this code, will this test catch it?"

If the answer is no, the test provides false confidence and should not be written.

### Key Principles

1. **Test Behavior, Not Implementation** - Tests should verify *what* the system does, not *how* it does it. Tests coupled to implementation break during refactoring even when behavior is unchanged.

2. **Coverage is a Map, Not a Score** - Use coverage to find untested risky areas, not as a target to hit. Goodhart's Law: "When a measure becomes a target, it ceases to be a good measure."

3. **Mock Minimally** - Kent Beck: "I mock almost nothing." Only mock external boundaries (network, disk, time). Over-mocking means you're testing mocks, not the system.

4. **Each Test Must Kill Mutants** - A test that can't detect code changes is worthless. Verify results: `assert result == expected`, not just `assert function_runs()`.

---

## Workflow

### Phase 0: Analyze Changes

Understand what code changed and what behaviors were added or modified.

- Read the changed code
- Identify new/modified behaviors
- Note the public interfaces affected

### Phase 1: Validate Test Targets

Not all code deserves tests. Filter ruthlessly.

**Worth Testing:**
- Business logic and domain rules
- Public API behavior
- State transitions with side effects
- Error handling paths
- Boundary conditions

**Skip:**
- Simple getters/setters (no logic)
- Framework/library code
- Private methods (test through public API)
- Third-party dependency behavior
- Pure boilerplate

### Phase 2: Identify Test Cases

For each test target, derive cases:

| Category | Description | Priority |
|----------|-------------|----------|
| Happy Path | Normal successful operation | High |
| Edge Cases | Boundaries, empty inputs, limits | High |
| Error Cases | Invalid inputs, failures, exceptions | High |
| Integration | Interaction with dependencies | Medium |

**Ask**: What could go wrong? What has broken before?

### Phase 3: Build Tests

Write tests following these patterns:

**Structure**: Arrange-Act-Assert
```
// Arrange: Set up preconditions
// Act: Execute the behavior
// Assert: Verify the outcome
```

**Naming**: Describe the behavior being tested
```
// Good: "returns_empty_list_when_no_items_match_filter"
// Bad: "test1" or "testGetItems"
```

**Independence**: Each test must run in isolation
- No shared mutable state
- No order dependencies
- No reliance on external systems without mocking

**Minimal Mocking**: Only mock at system boundaries
- Network calls
- File system
- Time/dates
- External services

### Phase 4: Evaluate Quality

Before finalizing, verify each test:

| Check | Question |
|-------|----------|
| Mutation | If I change this line of code, will this test fail? |
| Brittleness | Will this test break if I refactor without changing behavior? |
| Clarity | Can someone understand what's being tested without reading the code? |
| Speed | Does this test run fast enough to not discourage running it? |

See `references/test-quality-checklist.md` for detailed validation.

---

## What NOT to Write

### Anti-patterns to Avoid

1. **Assertion-Free Tests** - Code runs but nothing is verified
2. **Implementation Coupling** - Tests break on refactor
3. **Over-Mocking** - So many mocks the real system isn't tested
4. **Happy Path Only** - Bugs live in the unhappy paths
5. **Flaky Tests** - Intermittent failures destroy trust

See `references/test-anti-patterns.md` for detailed examples.

### Tests to Explicitly Skip

```
// SKIP: Tests trivial getter
test('getName returns name') { expect(user.getName()).toBe('Alice'); }

// SKIP: Tests framework behavior
test('component renders') { render(<Button />); expect(screen.getByRole('button')).toBeInTheDocument(); }

// SKIP: Tests implementation detail
test('calls internal helper') { expect(spy).toHaveBeenCalledWith(internalArgs); }
```

---

## Examples

### Good Test (Behavior)
```typescript
test('user cannot withdraw more than balance', () => {
  const account = new Account(100);
  expect(() => account.withdraw(150)).toThrow('Insufficient funds');
  expect(account.getBalance()).toBe(100); // unchanged
});
```

### Bad Test (Implementation)
```typescript
// BAD: Tests internal call, not behavior
test('finalize calls calculateTax', () => {
  const spy = jest.spyOn(taxService, 'calculateTax');
  order.finalize();
  expect(spy).toHaveBeenCalledWith(100, 'US', 0.08);
});

// GOOD: Tests observable behavior
test('finalized order includes correct tax', () => {
  const order = createOrder({ subtotal: 100, region: 'US' });
  order.finalize();
  expect(order.taxAmount).toBe(8);
});
```

---

## Quick Reference

**4 Questions Before Writing Any Test:**

1. "If I introduce a bug here, will this test catch it?"
2. "Am I testing behavior or implementation?"
3. "If this test fails, is it a real problem?"
4. "Are there too many mocks?"

**Remember**: A test that can't fail for the right reasons is worse than no test - it provides false confidence.
