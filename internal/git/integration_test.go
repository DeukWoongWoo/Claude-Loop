// +build integration

package git

import (
	"context"
	"fmt"
	"testing"
)

// Run with: go test -tags=integration ./internal/git -v -run TestIntegration
func TestIntegration_RealGit(t *testing.T) {
	ctx := context.Background()

	t.Run("Repository GetInfo", func(t *testing.T) {
		repo := NewRepository(nil)
		info, err := repo.GetInfo(ctx)
		if err != nil {
			t.Fatalf("GetInfo error: %v", err)
		}
		fmt.Printf("Repository Info:\n")
		fmt.Printf("  Root: %s\n", info.RootPath)
		fmt.Printf("  Branch: %s\n", info.CurrentBranch)
		fmt.Printf("  Remote: %s\n", info.RemoteURL)
		fmt.Printf("  Clean: %v\n", info.IsClean)
	})

	t.Run("BranchManager GenerateName", func(t *testing.T) {
		bm := NewBranchManager(nil)
		name, err := bm.GenerateBranchName("")
		if err != nil {
			t.Fatalf("GenerateBranchName error: %v", err)
		}
		fmt.Printf("Generated branch name: %s\n", name)
	})

	t.Run("WorktreeManager List", func(t *testing.T) {
		wm := NewWorktreeManager(nil)
		worktrees, err := wm.List(ctx)
		if err != nil {
			t.Fatalf("List worktrees error: %v", err)
		}
		fmt.Printf("%s", wm.FormatList(worktrees))
	})
}
