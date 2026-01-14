# Claude Loop Feature Matrix

> Mapping from bash functions to Go packages for migration tracking.

## Legend

| Status | Meaning |
|--------|---------|
| - | Not started |
| P | Planned |
| W | In Progress |
| D | Done |
| T | Tested |

---

## Phase 0: Preparation

| Feature | Bash Function | Lines | Go Package | Status | Notes |
|---------|---------------|-------|------------|--------|-------|
| CLI Contract | N/A | N/A | docs/ | D | This document |
| Golden Tests | N/A | N/A | test/golden/ | D | help.txt, version.txt |

---

## Phase 1: Foundation (DOU-137, DOU-138, DOU-139)

### Project Structure (DOU-137)

| Component | Description | Go Location | Status |
|-----------|-------------|-------------|--------|
| Entry point | Main function | cmd/claude-loop/main.go | - |
| Go module | Module init | go.mod | - |
| Build config | Makefile | Makefile | - |
| CI workflow | GitHub Actions | .github/workflows/ | - |

### CLI Parsing (DOU-138)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Argument parsing | `parse_arguments()` | 687-809 | internal/cli | - |
| Update flags | `parse_update_flags()` | 811-832 | internal/cli | - |
| Argument validation | `validate_arguments()` | 834-921 | internal/cli | - |
| Help display | `show_help()` | 265-384 | internal/cli | - |
| Version display | `show_version()` | 386-388 | internal/cli | - |

### Configuration (DOU-139)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Principles loading | `load_principles()` | 1139-1158 | internal/config | - |
| Principles validation | `ensure_principles()` | 1245-1262 | internal/config | - |
| YAML parsing | N/A (jq) | N/A | internal/config | - |

---

## Phase 2: Core Loop (DOU-140, DOU-141, DOU-142)

### Main Loop (DOU-140)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Main loop | `main_loop()` | 2476-2528 | internal/loop | - |
| Single iteration | `execute_single_iteration()` | 2355-2474 | internal/loop | - |
| Completion summary | `show_completion_summary()` | 2530-2554 | internal/loop | - |
| Entry point | `main()` | 2555-2599 | cmd/claude-loop | - |

### Claude Integration (DOU-141)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Claude execution | `run_claude_iteration()` | 1915-2008 | internal/claude | - |
| Result parsing | `parse_claude_result()` | 2191-2208 | internal/claude | - |
| Stream-JSON parsing | N/A (jq pipe) | N/A | internal/claude | - |
| Cost extraction | N/A (jq) | N/A | internal/claude | - |

### Prompt Builder (DOU-142)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Workflow context | `PROMPT_WORKFLOW_CONTEXT` | 15-23 | internal/prompt | - |
| Commit message | `PROMPT_COMMIT_MESSAGE` | 13 | internal/prompt | - |
| Notes update | `PROMPT_NOTES_*` | 25-40 | internal/prompt | - |
| Reviewer context | `PROMPT_REVIEWER_CONTEXT` | 42-45 | internal/prompt | - |
| CI fix context | `PROMPT_CI_FIX_CONTEXT` | 46-64 | internal/prompt | - |
| Principle collection | `PROMPT_PRINCIPLE_COLLECTION` | 66+ | internal/prompt | - |

---

## Phase 3: Git/GitHub (DOU-143, DOU-144, DOU-145)

### Git Operations (DOU-143)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Branch creation | `create_iteration_branch()` | 1546-1594 | internal/git | - |
| Worktree list | `list_worktrees()` | 1777-1792 | internal/git | - |
| Worktree setup | `setup_worktree()` | 1794-1858 | internal/git | - |
| Worktree cleanup | `cleanup_worktree()` | 1860-1901 | internal/git | - |
| Iteration display | `get_iteration_display()` | 1902-1914 | internal/git | - |

### GitHub PR Workflow (DOU-144)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| PR workflow | `continuous_claude_commit()` | 1595-1729 | internal/github | - |
| Current branch commit | `commit_on_current_branch()` | 1731-1775 | internal/github | - |
| PR checks wait | `wait_for_pr_checks()` | 1263-1456 | internal/github | - |
| PR merge | `merge_pr_and_cleanup()` | 1483-1544 | internal/github | - |
| Failed run ID | `get_failed_run_id()` | 1458-1481 | internal/github | - |
| Repo detection | `detect_github_repo()` | 642-685 | internal/github | - |

### CI Auto-Fix (DOU-145)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| CI fix iteration | `run_ci_fix_iteration()` | 2052-2151 | internal/github | - |
| CI fix recheck | `attempt_ci_fix_and_recheck()` | 2154-2189 | internal/github | - |

---

## Phase 4: Advanced Features (DOU-146 ~ DOU-151)

### Auto Update (DOU-146)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Version check | `get_latest_version()` | 390-404 | internal/update | - |
| Version compare | `compare_versions()` | 406-462 | internal/update | - |
| Script path | `get_script_path()` | 463-469 | internal/update | - |
| Download update | `download_and_install_update()` | 470-527 | internal/update | - |
| Update check | `check_for_updates()` | 528-585 | internal/update | - |
| Update command | `handle_update_command()` | 586-641 | internal/update | - |

### Reviewer Pass (DOU-147)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Reviewer iteration | `run_reviewer_iteration()` | 2010-2051 | internal/reviewer | - |

### LLM Council (DOU-148)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Decision logging | `log_principle_decision()` | 1160-1198 | internal/council | - |
| Conflict detection | `detect_principle_conflict()` | 1200-1208 | internal/council | - |
| Council invocation | `invoke_llm_council()` | 1210-1243 | internal/council | - |
| Council setup | `ensure_council_setup()` | 952-995 | internal/council | - |

### Release Automation (DOU-149)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| goreleaser | N/A | N/A | .goreleaser.yaml | T |
| CI workflow | N/A | N/A | .github/workflows/ci.yml | T |
| Release workflow | N/A | N/A | .github/workflows/release.yml | T |
| Install script | N/A | N/A | install.sh | T |

### Testing (DOU-150)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Integration tests | N/A | N/A | test/integration/ | - |
| E2E tests | N/A | N/A | test/e2e/ | - |
| Golden tests | N/A | N/A | test/golden/ | - |

### Documentation (DOU-151)

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| README | N/A | N/A | README.md | - |
| Migration guide | N/A | N/A | docs/MIGRATION.md | - |
| Contributing | N/A | N/A | CONTRIBUTING.md | - |

---

## Utility Functions

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Duration parsing | `parse_duration()` | 188-236 | internal/util | - |
| Duration format | `format_duration()` | 238-263 | internal/util | - |
| Requirements check | `validate_requirements()` | 923-951 | internal/util | - |

---

## Error Handling

| Feature | Bash Function | Lines | Go Package | Status |
|---------|---------------|-------|------------|--------|
| Iteration error | `handle_iteration_error()` | 2210-2254 | internal/loop | - |
| Iteration success | `handle_iteration_success()` | 2256-2353 | internal/loop | - |

---

## Summary

| Phase | Total Functions | Completed | Progress |
|-------|-----------------|-----------|----------|
| Phase 0 | 2 | 2 | 100% |
| Phase 1 | 8 | 0 | 0% |
| Phase 2 | 9 | 0 | 0% |
| Phase 3 | 11 | 0 | 0% |
| Phase 4 | 18 | 4 | 22% |
| **Total** | **48** | **6** | **13%** |
