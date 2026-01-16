# claude-loop Execution Flow

> Complete code flow trace of the `claude-loop` CLI execution, documented file by file.

---

## 1. Entry Point

### `cmd/claude-loop/main.go`

```go
func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

- Calls `cli.Execute()`

---

## 2. CLI Execution

### `internal/cli/root.go`

#### 2.1 Execute() Function (line 481-487)

```go
func Execute() error {
    if err := rootCmd.Execute(); err != nil {
        return err
    }
    return nil
}
```

- Executes Cobra's `rootCmd.Execute()`

#### 2.2 rootCmd Definition (line 26-162)

```go
var rootCmd = &cobra.Command{
    PreRunE: func(cmd *cobra.Command, args []string) error {
        // 1. Parse duration
        // 2. Validation
        return globalFlags.Validate()
    },
    Run: runMainLoop,  // ← Actual execution function
}
```

#### 2.3 runMainLoop() Function (line 367-445)

**Execution Order:**

1. Handle `--list-worktrees` (line 370-373)
2. Show help if no prompt (line 376-379)
3. Set up Context + Signal handling (line 381-392)
4. Check for updates (line 395)
5. Load Principles (line 397-413)
6. **Create loopConfig** (line 416-430)
7. **Create Claude Client** (line 433)
8. **Create and run Executor** (line 436-437)
9. Display results (line 444)

**Core Code:**

```go
// line 416-430
loopConfig := ConfigToLoopConfig(globalFlags)
loopConfig.Principles = principles
loopConfig.NeedsPrincipleCollection = needsPrincipleCollection

// line 433
claudeClient := claude.NewClient(nil)

// line 436-437
executor := loop.NewExecutor(loopConfig, claudeClient)
result, err := executor.Run(ctx)
```

---

## 3. Config Conversion

### `internal/cli/root.go` - ConfigToLoopConfig() (line 313-327)

```go
func ConfigToLoopConfig(f *Flags) *loop.Config {
    return &loop.Config{
        Prompt:               f.Prompt,
        MaxRuns:              f.MaxRuns,
        MaxCost:              f.MaxCost,
        MaxDuration:          f.MaxDuration,
        CompletionSignal:     f.CompletionSignal,
        CompletionThreshold:  f.CompletionThreshold,
        MaxConsecutiveErrors: 3,
        DryRun:               f.DryRun,
        NotesFile:            f.NotesFile,  // ← Default: "SHARED_TASK_NOTES.md"
        ReviewPrompt:         f.ReviewPrompt,
        LogDecisions:         f.LogDecisions,
    }
}
```

### Flag Defaults (`internal/cli/flags.go` line 54-76)

```go
func DefaultFlags() *Flags {
    return &Flags{
        GitBranchPrefix:     "claude-loop/",
        MergeStrategy:       "squash",
        CompletionSignal:    "CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
        CompletionThreshold: 3,
        CIRetryMax:          1,
        NotesFile:           "SHARED_TASK_NOTES.md",  // ← Default value
        WorktreeBaseDir:     "../claude-loop-worktrees",
        PrinciplesFile:      ".claude/principles.yaml",
    }
}
```

---

## 4. Executor Execution

### `internal/loop/executor.go`

#### 4.1 NewExecutor() (line 60-87)

```go
func NewExecutor(config *Config, client ClaudeClient) *Executor {
    e := &Executor{
        config:             config,
        limitChecker:       NewLimitChecker(config),
        completionDetector: NewCompletionDetector(config),
        iterationHandler:   NewIterationHandler(config, client),  // ← Iteration handler
    }

    // Initialize Reviewer (if config.ReviewPrompt is set)
    // Initialize Council (if config.Principles is set)

    return e
}
```

#### 4.2 Run() - Main Loop (line 91-182)

```go
func (e *Executor) Run(ctx context.Context) (*LoopResult, error) {
    state := NewState()  // Initialize state

    for {  // ← Infinite loop
        // 1. Check context cancellation
        select {
        case <-ctx.Done():
            return &LoopResult{StopReason: StopReasonContextCancelled}, nil
        default:
        }

        // 2. Check limits (max_runs, max_cost, max_duration)
        if result := e.limitChecker.Check(state); result.LimitReached {
            return &LoopResult{StopReason: result.Reason}, nil
        }

        // 3. Check completion threshold
        if result := e.completionDetector.CheckThreshold(state); result.LimitReached {
            return &LoopResult{StopReason: result.Reason}, nil
        }

        // 4. Execute iteration ★
        iterResult, err := e.iterationHandler.Execute(ctx, state)

        // 5. Handle errors
        if err != nil {
            shouldContinue := e.iterationHandler.HandleError(state, err)
            if !shouldContinue {
                return &LoopResult{StopReason: StopReasonConsecutiveErrors}, nil
            }
            continue
        }

        // 6. Handle Council (skipped in dry-run)
        if e.council != nil && !e.config.DryRun {
            e.handleCouncil(ctx, state, iterResult.Output)
        }

        // 7. Run Reviewer (skipped in dry-run)
        if e.reviewer != nil && !e.config.DryRun {
            e.runReviewerPass(ctx, state)
        }

        // 8. Progress callback
        if e.config.OnProgress != nil {
            e.config.OnProgress(state)
        }

        // 9. Check limits again (cost may have changed)
        // 10. Check completion threshold again
    }
}
```

---

## 5. Single Iteration Execution

### `internal/loop/iteration.go`

#### 5.1 Execute() (line 41-84)

```go
func (ih *IterationHandler) Execute(ctx context.Context, state *State) (*IterationResult, error) {
    state.TotalIterations++

    // Simulate if dry-run mode
    if ih.config.DryRun {
        return ih.executeDryRun(state)
    }

    // Build prompt ★
    buildCtx := prompt.BuildContext{
        UserPrompt:               ih.config.Prompt,
        Principles:               ih.config.Principles,
        NeedsPrincipleCollection: ih.config.NeedsPrincipleCollection && state.SuccessfulIterations == 0,
        CompletionSignal:         ih.config.CompletionSignal,
        NotesFile:                ih.config.NotesFile,  // ← "SHARED_TASK_NOTES.md"
        Iteration:                state.TotalIterations,
    }

    buildResult, err := ih.promptBuilder.Build(buildCtx)
    if err != nil {
        return nil, err
    }

    // Execute Claude ★
    result, err := ih.client.Execute(ctx, buildResult.Prompt)
    if err != nil {
        return nil, err
    }

    // Update state
    ih.updateStateOnSuccess(state, result)

    return result, nil
}
```

---

## 6. Prompt Building

### `internal/prompt/builder.go`

#### 6.1 Build() (line 45-122)

Prompt composition order:

```
1. [Conditional] Principle Collection prompt (first run + no principles.yaml)
2. [Conditional] Decision Principles (if principles are loaded)
3. Workflow Context (includes completion signal)
4. User Prompt
5. [Conditional] Notes content from previous iteration ★
6. Notes writing instructions (UPDATE or CREATE)
7. Notes guidelines
```

**Notes Loading Code (line 85-101):**

```go
// 5. Notes from Previous Iteration (if file exists)
notesContent, notesExists, err := b.notesLoader.Load(ctx.NotesFile)
if err != nil {
    return nil, fmt.Errorf("failed to load notes: %w", err)
}

if notesExists && notesContent != "" {
    notesHeader := strings.ReplaceAll(
        TemplateNotesContext,
        PlaceholderNotesFile,
        ctx.NotesFile,
    )
    sb.WriteString(notesHeader)
    sb.WriteString(notesContent)  // ← Insert notes content
    sb.WriteString("\n\n")
    result.NotesIncluded = true
}
```

### `internal/prompt/notes.go`

#### 6.2 Load() (line 26-40)

```go
func (l *FileNotesLoader) Load(path string) (string, bool, error) {
    if path == "" {
        return "", false, nil  // Ignore if path is empty
    }

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return "", false, nil  // Return false if file doesn't exist
        }
        return "", false, err
    }

    return string(data), true, nil  // Return file content
}
```

---

## 7. Claude CLI Execution

### `internal/claude/client.go`

#### 7.1 Execute() (line 86-166)

```go
func (c *Client) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
    // Build command: claude -p "prompt" --dangerously-skip-permissions --output-format stream-json --verbose
    args := []string{"-p", prompt}
    args = append(args, c.opts.AdditionalFlags...)

    cmd := c.opts.Executor.CommandContext(ctx, c.opts.ClaudePath, args...)

    // Execute and parse output...

    return &loop.IterationResult{
        Output:   parsed.Output,
        Cost:     parsed.TotalCostUSD,
        Duration: duration,
    }, nil
}
```

**Actual Command Executed:**

```bash
claude -p "<built_prompt>" --dangerously-skip-permissions --output-format stream-json --verbose
```

---

## 8. Iteration Continuity Mechanism

### Core Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      Iteration 1                            │
├─────────────────────────────────────────────────────────────┤
│ 1. promptBuilder.Build() called                             │
│ 2. NotesLoader.Load("SHARED_TASK_NOTES.md")                │
│    → File not found: exists=false                           │
│ 3. Build prompt (without Notes)                             │
│ 4. Execute Claude                                           │
│ 5. Claude creates/modifies SHARED_TASK_NOTES.md ★          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Iteration 2                            │
├─────────────────────────────────────────────────────────────┤
│ 1. promptBuilder.Build() called                             │
│ 2. NotesLoader.Load("SHARED_TASK_NOTES.md")                │
│    → File found: exists=true, content="previous work..."   │
│ 3. Build prompt (with Notes included) ★                    │
│ 4. Execute Claude (aware of previous context)               │
│ 5. Claude updates SHARED_TASK_NOTES.md                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
                            ...
```

### Why Continuity is Guaranteed

1. **Default Value**: `NotesFile` defaults to `"SHARED_TASK_NOTES.md"` (flags.go:68)
2. **File Loaded Every Iteration**: `notesLoader.Load()` called at builder.go:86
3. **Content Included in Prompt**: Notes content inserted at builder.go:97-99
4. **Claude Instructed to Update**: Templates at templates.go:115-119

### Notes Templates Included in Prompt

```go
// templates.go:137-141
const TemplateNotesContext = `## CONTEXT FROM PREVIOUS ITERATION

The following is from ` + "`NOTES_FILE_PLACEHOLDER`" + `, maintained by previous iterations to provide context:

`

// templates.go:115-116
const TemplateNotesUpdateExisting = "Update the `NOTES_FILE_PLACEHOLDER` file with relevant context for the next iteration..."

// templates.go:119
const TemplateNotesCreateNew = "Create a `NOTES_FILE_PLACEHOLDER` file with relevant context and instructions for the next iteration."
```

---

## 9. Termination Conditions

| Condition | Code Location | Description |
|-----------|---------------|-------------|
| `max_runs` reached | `executor.go:108` | `-m` flag value |
| `max_cost` reached | `executor.go:108` | `--max-cost` flag value |
| `max_duration` reached | `executor.go:108` | `--max-duration` flag value |
| `completion_threshold` reached | `executor.go:116` | Consecutive completion signal detection count |
| 3 consecutive errors | `executor.go:134` | `MaxConsecutiveErrors: 3` |
| Context cancelled (Ctrl+C) | `executor.go:97` | Signal handling |

---

## 10. Conclusion

**Even without specifying `--notes-file`, the default value `SHARED_TASK_NOTES.md` is used, automatically ensuring iteration continuity.**

Every iteration:
1. Reads `SHARED_TASK_NOTES.md` and includes it in the prompt
2. Instructs Claude to update the file after work
3. Next iteration reads the updated content

This allows all iterations to **work continuously toward a single goal**.
