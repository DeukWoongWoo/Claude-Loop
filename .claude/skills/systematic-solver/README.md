# Systematic Problem Solver

Systematic problem-solving and code modification framework ensuring root cause analysis and regression prevention.

## Usage

Invoke this skill when debugging or modifying code.

**Trigger phrases:**
- "Fix this bug"
- "Debug this issue"
- "Refactor this code"
- "Optimize performance"
- "Troubleshoot this problem"

## Description

This skill ensures thorough, non-patchwork solutions by following the ARIA workflow. Instead of applying quick fixes to symptoms, it guides you through proper root cause analysis and impact assessment.

**Key capabilities:**
- Complexity assessment to determine workflow depth (simple vs complex)
- Root cause analysis using the 5 Whys technique
- Impact analysis across codebase dependencies
- Regression prevention with tests and similar code review

## Workflow

When invoked, the skill follows the ARIA process:

1. **Assess**: Evaluate complexity - simple fixes proceed immediately, complex issues continue
2. **Root Cause**: Apply 5 Whys technique, trace execution paths, validate with code evidence
3. **Impact**: Analyze dependencies, check callers/callees, assess risk level, review test coverage
4. **Act**: Implement fix with regression prevention, check for similar issues, verify no new problems

**Output**: Implementation summary containing:
- Root cause statement with evidence
- Impact report with risk assessment
- Changes made with explanations
- Regression prevention checklist
- Verification steps

## Prerequisites

No special setup required.
