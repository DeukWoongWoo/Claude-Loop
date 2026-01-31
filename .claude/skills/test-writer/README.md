# Test Writer

Write meaningful tests that catch real bugs, not tests that just increase coverage numbers. Focuses on behavior testing and mutation resistance.

## Usage

Invoke this skill to write tests for new or modified code.

**Trigger phrases:**
- "Write tests"
- "Add tests"
- "Create test cases"

**With arguments:**
```
/test-writer                           # Write tests for recent changes
/test-writer src/utils/auth.ts         # Write tests for specific file
```

## Description

This skill follows a behavior-first testing philosophy:

**The Litmus Test**: "If I introduce a bug into this code, will this test catch it?"

**Core Principles**:
1. **Test Behavior, Not Implementation** - Verify *what* the system does, not *how*
2. **Coverage is a Map, Not a Score** - Use coverage to find untested areas, not as a target
3. **Mock Minimally** - Only mock external boundaries (network, disk, time)
4. **Each Test Must Kill Mutants** - A test that can't detect code changes is worthless

## Workflow

When invoked, the skill:

1. **Analyze Changes**: Identify new/modified behaviors and public interfaces
2. **Validate Targets**: Filter what deserves tests (skip trivial getters, framework code)
3. **Identify Cases**: Derive test cases (happy path, edge cases, error cases)
4. **Build Tests**: Write tests following Arrange-Act-Assert pattern
5. **Evaluate Quality**: Verify mutation resistance and clarity

**Output**: Test files with:
- Behavior-focused test cases
- Descriptive test names
- Minimal mocking at system boundaries
- Independence (no shared state, no order dependencies)

## What NOT to Write

- Assertion-free tests (code runs but nothing verified)
- Implementation-coupled tests (break on refactor)
- Over-mocked tests (testing mocks, not system)
- Happy-path-only tests (bugs live in unhappy paths)

## Reference Documents

| Document | Purpose |
|----------|---------|
| `references/test-quality-checklist.md` | Detailed validation checklist |
| `references/test-anti-patterns.md` | Common anti-patterns with examples |
