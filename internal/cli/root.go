// Package cli provides the command-line interface for claude-loop.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/architecture"
	"github.com/DeukWoongWoo/claude-loop/internal/claude"
	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/DeukWoongWoo/claude-loop/internal/decomposer"
	"github.com/DeukWoongWoo/claude-loop/internal/git"
	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"github.com/DeukWoongWoo/claude-loop/internal/prd"
	"github.com/DeukWoongWoo/claude-loop/internal/principles"
	"github.com/DeukWoongWoo/claude-loop/internal/update"
	"github.com/DeukWoongWoo/claude-loop/internal/version"
	"github.com/spf13/cobra"
)

// maxDurationStr holds the raw duration string for parsing.
var maxDurationStr string

// ConsoleStreamHandler implements claude.ToolStreamHandler for real-time output.
type ConsoleStreamHandler struct {
	formatter *ToolFormatter
}

// NewConsoleStreamHandler creates a new handler with formatter.
func NewConsoleStreamHandler() *ConsoleStreamHandler {
	return &ConsoleStreamHandler{formatter: NewToolFormatter()}
}

// OnText prints text to stdout in real-time.
func (h *ConsoleStreamHandler) OnText(text string) {
	fmt.Print(text)
}

// OnToolUse prints tool invocation to stdout.
func (h *ConsoleStreamHandler) OnToolUse(name string, input string) {
	formatted := h.formatter.FormatToolUse(name, input)
	fmt.Printf("\n[Tool: %s] %s\n", name, formatted)
}

// OnToolResult prints tool result to stdout.
func (h *ConsoleStreamHandler) OnToolResult(content string, isError bool) {
	prefix := "[Result]"
	if isError {
		prefix = "[Error]"
	}
	formatted := h.formatter.FormatToolResult(content)
	fmt.Printf("%s %s\n", prefix, formatted)
}

var rootCmd = &cobra.Command{
	Use:   "claude-loop",
	Short: "Autonomous AI development loop with 4-Layer Principles Framework",
	Long: `Claude Loop - Autonomous AI development loop with 4-Layer Principles Framework

USAGE:
    claude-loop -p "prompt" (-m max-runs | --max-cost max-cost | --max-duration duration) [--owner owner] [--repo repo] [options]
    claude-loop update

REQUIRED OPTIONS:
    -p, --prompt <text>           The prompt/goal for Claude Code to work on
    -m, --max-runs <number>       Maximum number of successful iterations (use 0 for unlimited with --max-cost or --max-duration)
    --max-cost <dollars>          Maximum cost in USD to spend (alternative to --max-runs)
    --max-duration <duration>     Maximum duration to run (e.g., "2h", "30m", "1h30m") (alternative to --max-runs)

OPTIONAL FLAGS:
    -h, --help                    Show this help message
    -v, --version                 Show version information
    --owner <owner>               GitHub repository owner (auto-detected from git remote if not provided)
    --repo <repo>                 GitHub repository name (auto-detected from git remote if not provided)
    --disable-commits             Disable automatic commits and PR creation
    --disable-branches            Commit on current branch without creating branches or PRs
    --auto-update                 Automatically install updates when available
    --disable-updates             Skip all update checks and prompts
    --git-branch-prefix <prefix>  Branch prefix for iterations (default: "claude-loop/")
    --merge-strategy <strategy>   PR merge strategy: squash, merge, or rebase (default: "squash")
    --notes-file <file>           Shared notes file for iteration context (default: "SHARED_TASK_NOTES.md")
    --worktree <name>             Run in a git worktree for parallel execution (creates if needed)
    --worktree-base-dir <path>    Base directory for worktrees (default: "../claude-loop-worktrees")
    --cleanup-worktree            Remove worktree after completion
    --list-worktrees              List all active git worktrees and exit
    --dry-run                     Simulate execution without making changes
    --completion-signal <phrase>  Phrase that agents output when project is complete (default: "CONTINUOUS_CLAUDE_PROJECT_COMPLETE")
    --completion-threshold <num>  Number of consecutive signals to stop early (default: 3)
    -r, --review-prompt <text>    Run a reviewer pass after each iteration to validate changes
                                  (e.g., run build/lint/tests and fix any issues)
    --disable-ci-retry            Disable automatic CI failure retry (enabled by default)
    --ci-retry-max <number>       Maximum CI fix attempts per PR (default: 1)
    --reset-principles            Force re-collection of principles
    --principles-file <path>      Custom principles file path (default: ".claude/principles.yaml")
    --log-decisions               Enable decision logging to .claude/principles-decisions.log
    --verbose                     Show detailed iteration summaries
    --stream                      Stream Claude output in real-time
    --plan                        Enable planning mode (PRD → Architecture → Tasks)
    --plan-only                   Generate plan without execution (implies --plan)
    --resume <plan-id>            Resume from saved plan ID

COMMANDS:
    update                        Check for and install the latest version

EXAMPLES:
    # Run 5 iterations to fix bugs
    claude-loop -p "Fix all linter errors" -m 5 --owner myuser --repo myproject

    # Run with cost limit
    claude-loop -p "Add tests" --max-cost 10.00 --owner myuser --repo myproject

    # Run for a maximum duration (time-boxed)
    claude-loop -p "Add documentation" --max-duration 2h --owner myuser --repo myproject

    # Run for 30 minutes
    claude-loop -p "Refactor module" --max-duration 30m --owner myuser --repo myproject

    # Run without commits (testing mode)
    claude-loop -p "Refactor code" -m 3 --disable-commits

    # Run with commits on current branch (no branches or PRs)
    claude-loop -p "Quick fixes" -m 3 --disable-branches

    # Use custom branch prefix and merge strategy
    claude-loop -p "Feature work" -m 10 --owner myuser --repo myproject \
        --git-branch-prefix "ai/" --merge-strategy merge

    # Combine duration and cost limits (whichever comes first)
    claude-loop -p "Add tests" --max-duration 1h30m --max-cost 5.00 --owner myuser --repo myproject

    # Run in a worktree for parallel execution
    claude-loop -p "Add unit tests" -m 5 --owner myuser --repo myproject --worktree instance-1

    # Run multiple instances in parallel (in different terminals)
    claude-loop -p "Task A" -m 5 --owner myuser --repo myproject --worktree task-a
    claude-loop -p "Task B" -m 5 --owner myuser --repo myproject --worktree task-b

    # List all active worktrees
    claude-loop --list-worktrees

    # Clean up worktree after completion
    claude-loop -p "Quick fix" -m 1 --owner myuser --repo myproject \
        --worktree temp --cleanup-worktree

    # Use completion signal to stop early when project is done
    claude-loop -p "Add unit tests to all files" -m 50 --owner myuser --repo myproject \
        --completion-threshold 3

    # Use a reviewer to validate and fix changes after each iteration
    claude-loop -p "Add new feature" -m 5 --owner myuser --repo myproject \
        -r "Run npm test and npm run lint, fix any failures"

    # Allow up to 2 CI fix attempts per PR (default is 1)
    claude-loop -p "Add tests" -m 5 --owner myuser --repo myproject --ci-retry-max 2

    # Disable automatic CI failure retry
    claude-loop -p "Add tests" -m 5 --owner myuser --repo myproject --disable-ci-retry

    # Run with custom principles file
    claude-loop -p "Feature work" -m 5 --principles-file custom-principles.yaml

    # Force re-collection of principles
    claude-loop -p "New project" -m 5 --reset-principles

    # Check for and install updates
    claude-loop update

REQUIREMENTS:
    - Claude Code CLI (https://claude.ai/code)
    - GitHub CLI (gh) - authenticated with 'gh auth login'
    - jq - JSON parsing utility
    - Git repository (unless --disable-commits is used)

NOTE:
    claude-loop automatically checks for updates at startup. You can press 'N' to skip the update.

For more information, visit: https://github.com/DeukWoongWoo/claude-loop`,
	Version: version.Version,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Parse duration string to time.Duration
		if err := parseDuration(); err != nil {
			return err
		}
		// Skip validation for --list-worktrees (standalone action)
		if globalFlags.ListWorktrees {
			return nil
		}
		// Skip validation if no flags provided (will show help)
		if globalFlags.Prompt == "" && globalFlags.MaxRuns == 0 && globalFlags.MaxCost == 0 && maxDurationStr == "" {
			return nil
		}
		// Validate flags
		return globalFlags.Validate()
	},
	Run: runMainLoop,
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install the latest version",
	Long:  `Check for and install the latest version of claude-loop.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		manager := update.NewManager(&update.ManagerOptions{
			AutoUpdate: true, // Force update when explicitly running update command
			OnProgress: func(status string) {
				fmt.Println(status)
			},
			CheckerOptions: update.DefaultCheckerOptions(version.Version),
		})

		result, err := manager.Update(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result.Message)

		if result.Updated && result.RestartRequired {
			fmt.Println("Restarting...")
			if err := manager.Restart(ctx, os.Args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Restart failed: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// checkForUpdatesAtStartup checks for updates at startup.
func checkForUpdatesAtStartup(ctx context.Context, flags *Flags) {
	if flags.DisableUpdates {
		return
	}

	manager := update.NewManager(&update.ManagerOptions{
		AutoUpdate: flags.AutoUpdate,
		OnPrompt: func(current, latest string) bool {
			fmt.Printf("New version %s available (current: %s). Update? [Y/n]: ", latest, current)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return false
			}
			response = strings.TrimSpace(strings.ToLower(response))
			return response == "" || response == "y" || response == "yes"
		},
		OnProgress: func(status string) {
			fmt.Println(status)
		},
		CheckerOptions: update.DefaultCheckerOptions(version.Version),
	})

	result, err := manager.CheckAndUpdate(ctx)
	if err != nil {
		// Non-fatal at startup, just warn
		fmt.Fprintf(os.Stderr, "Warning: update check failed: %v\n", err)
		return
	}

	if result.Updated && result.RestartRequired {
		fmt.Println("Update installed. Please restart claude-loop.")
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)
	configureCommand(rootCmd)
	registerFlags(rootCmd)
}

// configureCommand sets version and help templates on a command.
func configureCommand(cmd *cobra.Command) {
	cmd.SetVersionTemplate("claude-loop version {{.Version}}\n")
	cmd.SetHelpTemplate(`{{.Long}}
`)
}

// registerFlags registers all CLI flags on the given command.
// This centralizes flag registration to avoid duplication between init() and NewRootCmd().
func registerFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	f := globalFlags

	// Required options
	flags.StringVarP(&f.Prompt, "prompt", "p", "", "The prompt/goal for Claude Code to work on")
	flags.IntVarP(&f.MaxRuns, "max-runs", "m", 0, "Maximum number of successful iterations")
	flags.Float64Var(&f.MaxCost, "max-cost", 0, "Maximum cost in USD to spend")
	flags.StringVar(&maxDurationStr, "max-duration", "", "Maximum duration to run (e.g., \"2h\", \"30m\")")

	// GitHub configuration
	flags.StringVar(&f.Owner, "owner", "", "GitHub repository owner")
	flags.StringVar(&f.Repo, "repo", "", "GitHub repository name")

	// Commit & Branch management
	flags.BoolVar(&f.DisableCommits, "disable-commits", false, "Disable automatic commits and PR creation")
	flags.BoolVar(&f.DisableBranches, "disable-branches", false, "Commit on current branch without creating branches or PRs")
	flags.StringVar(&f.GitBranchPrefix, "git-branch-prefix", "claude-loop/", "Branch prefix for iterations")
	flags.StringVar(&f.MergeStrategy, "merge-strategy", "squash", "PR merge strategy: squash, merge, or rebase")

	// Iteration control
	flags.StringVar(&f.CompletionSignal, "completion-signal", "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", "Phrase that agents output when project is complete")
	flags.IntVar(&f.CompletionThreshold, "completion-threshold", 3, "Number of consecutive signals to stop early")
	flags.BoolVar(&f.DryRun, "dry-run", false, "Simulate execution without making changes")

	// Review & CI
	flags.StringVarP(&f.ReviewPrompt, "review-prompt", "r", "", "Run a reviewer pass after each iteration")
	flags.BoolVar(&f.DisableCIRetry, "disable-ci-retry", false, "Disable automatic CI failure retry")
	flags.IntVar(&f.CIRetryMax, "ci-retry-max", 1, "Maximum CI fix attempts per PR")

	// Shared state
	flags.StringVar(&f.NotesFile, "notes-file", "SHARED_TASK_NOTES.md", "Shared notes file for iteration context")

	// Worktree support
	flags.StringVar(&f.Worktree, "worktree", "", "Run in a git worktree for parallel execution")
	flags.StringVar(&f.WorktreeBaseDir, "worktree-base-dir", "../claude-loop-worktrees", "Base directory for worktrees")
	flags.BoolVar(&f.CleanupWorktree, "cleanup-worktree", false, "Remove worktree after completion")
	flags.BoolVar(&f.ListWorktrees, "list-worktrees", false, "List all active git worktrees and exit")

	// Principles framework
	flags.BoolVar(&f.ResetPrinciples, "reset-principles", false, "Force re-collection of principles")
	flags.StringVar(&f.PrinciplesFile, "principles-file", ".claude/principles.yaml", "Custom principles file path")
	flags.BoolVar(&f.LogDecisions, "log-decisions", false, "Enable decision logging")

	// Output control
	flags.BoolVar(&f.Verbose, "verbose", false, "Show detailed iteration summaries")
	flags.BoolVar(&f.Stream, "stream", false, "Stream Claude output in real-time")

	// Update management
	flags.BoolVar(&f.AutoUpdate, "auto-update", false, "Automatically install updates when available")
	flags.BoolVar(&f.DisableUpdates, "disable-updates", false, "Skip all update checks and prompts")

	// Planning mode
	flags.BoolVar(&f.Plan, "plan", false, "Enable planning mode (PRD → Architecture → Tasks)")
	flags.BoolVar(&f.PlanOnly, "plan-only", false, "Generate plan without execution (implies --plan)")
	flags.StringVar(&f.Resume, "resume", "", "Resume from saved plan ID")
}

// parseDuration parses the max-duration flag if provided.
func parseDuration() error {
	if maxDurationStr != "" {
		d, err := time.ParseDuration(maxDurationStr)
		if err != nil {
			return fmt.Errorf("invalid duration format %q: %w", maxDurationStr, err)
		}
		globalFlags.MaxDuration = d
	}
	return nil
}

// ConfigToLoopConfig creates a loop.Config from CLI Flags.
// Note: Principles must be set separately after loading.
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
		NotesFile:            f.NotesFile,
		ReviewPrompt:         f.ReviewPrompt,
		LogDecisions:         f.LogDecisions,
	}
}

// handleListWorktrees handles the --list-worktrees flag.
func handleListWorktrees() {
	ctx := context.Background()
	wm := git.NewWorktreeManager(nil)

	worktrees, err := wm.List(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(wm.FormatList(worktrees))
}

// runPlanningMode executes the planning workflow (PRD → Architecture → Tasks).
func runPlanningMode(ctx context.Context, flags *Flags) error {
	// Create Claude client with optional streaming
	var clientOpts *claude.ClientOptions
	if flags.Stream {
		clientOpts = &claude.ClientOptions{StreamHandler: NewConsoleStreamHandler()}
	}
	claudeClient := claude.NewClient(clientOpts)

	// claude.Client implements loop.ClaudeClient, wrap with planner adapter
	adapter := planner.NewClaudeClientAdapter(claudeClient)

	// Create Phase implementations
	prdGenerator := prd.NewGenerator(nil, adapter)
	archGenerator := architecture.NewGenerator(nil, adapter)
	taskDecomposer := decomposer.NewDecomposer(nil, adapter)

	phases := []planner.PlanningPhase{
		prd.NewPhase(prdGenerator),
		architecture.NewPhase(archGenerator),
		decomposer.NewPhase(taskDecomposer),
	}

	// Create PhaseRunner
	plannerConfig := planner.DefaultConfig()
	runner := planner.NewPhaseRunner(plannerConfig, nil, phases...)

	// Resume mode: load and continue existing plan
	if flags.Resume != "" {
		planPath := runner.Persistence().DefaultPlanPath(flags.Resume)
		fmt.Printf("Resuming plan from: %s\n", planPath)

		result, err := runner.ResumePlan(ctx, planPath)
		if err != nil {
			return fmt.Errorf("failed to resume plan: %w", err)
		}

		displayPlanResult(result)

		if flags.PlanOnly {
			return nil
		}

		// TODO: Execute tasks from result.Plan.TaskGraph
		fmt.Println("Task execution not yet implemented")
		return nil
	}

	// Create new Plan with timestamp-based ID
	planID := fmt.Sprintf("plan-%d", time.Now().UnixNano())
	plan := planner.NewPlan(planID, flags.Prompt)

	fmt.Printf("Starting planning mode (ID: %s)\n", planID)
	fmt.Printf("Prompt: %s\n\n", flags.Prompt)

	// Run planning phases
	result, err := runner.Run(ctx, plan)
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	displayPlanResult(result)

	// If --plan-only, stop here
	if flags.PlanOnly {
		fmt.Printf("\nPlan saved. Resume with: claude-loop --resume %s\n", planID)
		return nil
	}

	// TODO: Execute tasks from plan.TaskGraph
	fmt.Println("\nTask execution not yet implemented")
	return nil
}

// displayPlanResult displays the planning result summary.
func displayPlanResult(result *planner.RunResult) {
	fmt.Println("\n=== Planning Complete ===")
	if result.Plan != nil {
		fmt.Printf("Completed phases: %v\n", result.Plan.CompletedPhases)
	}
	fmt.Printf("Total cost: $%.4f\n", result.TotalCost)
	fmt.Printf("Duration: %s\n", result.TotalDuration.Round(time.Second))

	if result.Plan != nil {
		if result.Plan.PRD != nil {
			fmt.Printf("PRD goals: %d\n", len(result.Plan.PRD.Goals))
		}
		if result.Plan.Architecture != nil {
			fmt.Printf("Architecture components: %d\n", len(result.Plan.Architecture.Components))
		}
		if result.Plan.TaskGraph != nil {
			fmt.Printf("Tasks: %d\n", len(result.Plan.TaskGraph.Tasks))
		}
	}
}

// displayLoopResult displays the final loop result.
func displayLoopResult(result *loop.LoopResult) {
	state := result.State

	fmt.Println("\n=== Loop Complete ===")
	fmt.Printf("Stop reason: %s\n", result.StopReason)
	fmt.Printf("Successful iterations: %d\n", state.SuccessfulIterations)
	fmt.Printf("Total iterations: %d\n", state.TotalIterations)
	fmt.Printf("Total cost: $%.4f\n", state.TotalCost)
	fmt.Printf("Duration: %s\n", state.Elapsed().Round(time.Second))

	if state.ReviewerCost > 0 {
		fmt.Printf("Reviewer cost: $%.4f\n", state.ReviewerCost)
	}
	if state.CouncilInvocations > 0 {
		fmt.Printf("Council invocations: %d (cost: $%.4f)\n",
			state.CouncilInvocations, state.CouncilCost)
	}

	if result.LastError != nil {
		fmt.Printf("Last error: %v\n", result.LastError)
	}
}

// runMainLoop executes the main loop. Shared by rootCmd and NewRootCmd.
func runMainLoop(cmd *cobra.Command, args []string) {
	// Handle --list-worktrees (standalone action)
	if globalFlags.ListWorktrees {
		handleListWorktrees()
		return
	}

	// If no arguments and no prompt provided (and not resuming), show help
	if len(args) == 0 && globalFlags.Prompt == "" && globalFlags.Resume == "" {
		_ = cmd.Help()
		return
	}

	// Create cancellable context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal, stopping...")
		cancel()
	}()

	// Check for updates at startup
	checkForUpdatesAtStartup(ctx, globalFlags)

	// Check for planning mode
	if globalFlags.Plan || globalFlags.PlanOnly || globalFlags.Resume != "" {
		if err := runPlanningMode(ctx, globalFlags); err != nil {
			fmt.Fprintf(os.Stderr, "Planning failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load or collect principles
	loadedPrinciples, err := loadOrCollectPrinciples(ctx, globalFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create loop config from flags
	loopConfig := ConfigToLoopConfig(globalFlags)
	loopConfig.Principles = loadedPrinciples

	// Track previous cost for per-iteration cost calculation in verbose mode
	var previousCost float64
	loopConfig.OnProgress = func(state *loop.State) {
		maxRunsStr := "unlimited"
		if globalFlags.MaxRuns > 0 {
			maxRunsStr = fmt.Sprintf("%d", globalFlags.MaxRuns)
		}

		if globalFlags.Verbose {
			// Detailed output - differentiate success vs failure
			status := "Complete"
			if state.ErrorCount > 0 {
				status = "Failed"
			}
			fmt.Printf("\n--- Iteration %d/%s %s ---\n",
				state.TotalIterations, maxRunsStr, status)
			fmt.Printf("Cost: $%.4f (Total: $%.4f)\n",
				state.TotalCost-previousCost, state.TotalCost)
			fmt.Printf("Elapsed: %s\n", state.Elapsed().Round(time.Second))
			if state.CompletionSignalCount > 0 {
				fmt.Printf("Completion signals: %d/%d\n",
					state.CompletionSignalCount, globalFlags.CompletionThreshold)
			}
			if state.ErrorCount > 0 {
				fmt.Printf("Consecutive errors: %d\n", state.ErrorCount)
			}
			fmt.Println()
		} else {
			// Default minimal output
			fmt.Printf("[%d/%s] Cost: $%.4f | Elapsed: %s\n",
				state.SuccessfulIterations,
				maxRunsStr,
				state.TotalCost,
				state.Elapsed().Round(time.Second),
			)
		}
		previousCost = state.TotalCost
	}

	// Create Claude client for main loop
	var clientOpts *claude.ClientOptions
	if globalFlags.Stream {
		clientOpts = &claude.ClientOptions{StreamHandler: NewConsoleStreamHandler()}
	}
	claudeClient := claude.NewClient(clientOpts)

	// Create and run Executor
	executor := loop.NewExecutor(loopConfig, claudeClient)
	result, err := executor.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loop failed: %v\n", err)
		os.Exit(1)
	}

	// Display result
	displayLoopResult(result)
}

// newRootCmdWithRunner creates a new root command with the specified run function.
func newRootCmdWithRunner(runner func(cmd *cobra.Command, args []string)) *cobra.Command {
	ResetFlags()
	maxDurationStr = ""

	cmd := &cobra.Command{
		Use:     "claude-loop",
		Short:   "Autonomous AI development loop with 4-Layer Principles Framework",
		Long:    rootCmd.Long,
		Version: rootCmd.Version,
		PreRunE: rootCmd.PreRunE,
		Run:     runner,
	}

	configureCommand(cmd)
	registerFlags(cmd)

	return cmd
}

// NewRootCmd creates a new root command for testing.
// This allows tests to work with a fresh command instance.
func NewRootCmd() *cobra.Command {
	return newRootCmdWithRunner(runMainLoop)
}

// NewRootCmdForFlagParsing creates a root command that only parses flags without executing.
// Use this for unit/integration tests that need to test flag parsing behavior
// without actually running the main loop.
func NewRootCmdForFlagParsing() *cobra.Command {
	return newRootCmdWithRunner(func(cmd *cobra.Command, args []string) {})
}

// loadOrCollectPrinciples loads existing principles or collects them interactively.
func loadOrCollectPrinciples(ctx context.Context, flags *Flags) (*config.Principles, error) {
	collector := principles.NewCollector(flags.PrinciplesFile)

	if !collector.NeedsCollection(flags.ResetPrinciples) {
		return config.LoadFromFile(flags.PrinciplesFile)
	}

	// In dry-run mode, use defaults instead of interactive collection
	if flags.DryRun {
		fmt.Println("Principles file not found. Using default principles for dry-run mode.")
		return config.DefaultPrinciples(config.PresetStartup), nil
	}

	// Run interactive collection
	fmt.Println("Principles file not found. Starting interactive collection...")
	fmt.Println("Please answer the following questions to configure project principles.")
	fmt.Println()

	if err := collector.Collect(ctx); err != nil {
		return nil, fmt.Errorf("collecting principles: %w", err)
	}

	fmt.Println()
	fmt.Println("Principles collected successfully. Continuing with main loop...")
	fmt.Println()

	return config.LoadFromFile(flags.PrinciplesFile)
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
