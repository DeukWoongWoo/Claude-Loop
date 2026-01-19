// Package principles provides principle collection functionality for the Claude Loop.
package principles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeukWoongWoo/claude-loop/internal/prompt"
)

// InteractiveClient defines the interface for interactive Claude execution.
type InteractiveClient interface {
	ExecuteInteractive(ctx context.Context, prompt string) error
}

// Collector handles the collection of project principles via interactive prompts.
type Collector struct {
	client         InteractiveClient
	principlesPath string
}

// NewCollector creates a new Collector.
func NewCollector(client InteractiveClient, principlesPath string) *Collector {
	return &Collector{
		client:         client,
		principlesPath: principlesPath,
	}
}

// NeedsCollection checks whether principles need to be collected.
// Returns true if the principles file doesn't exist or forceReset is true.
func (c *Collector) NeedsCollection(forceReset bool) bool {
	if forceReset {
		return true
	}
	_, err := os.Stat(c.principlesPath)
	return os.IsNotExist(err)
}

// Collect runs an interactive Claude session to collect principles from the user.
// The principles are saved to the configured principlesPath.
func (c *Collector) Collect(ctx context.Context) error {
	// Ensure the parent directory exists
	dir := filepath.Dir(c.principlesPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &CollectorError{
				Message: "failed to create principles directory",
				Err:     err,
			}
		}
	}

	// Build the collection prompt with the target path substituted
	collectionPrompt := buildCollectionPrompt(c.principlesPath)

	// Execute interactive session
	if err := c.client.ExecuteInteractive(ctx, collectionPrompt); err != nil {
		return &CollectorError{
			Message: "interactive principles collection failed",
			Err:     err,
		}
	}

	// Verify the file was created
	if _, err := os.Stat(c.principlesPath); os.IsNotExist(err) {
		return &CollectorError{
			Message: "principles file was not created after collection",
		}
	}

	return nil
}

// buildCollectionPrompt returns the prompt for principle collection with the path substituted.
func buildCollectionPrompt(principlesPath string) string {
	return strings.ReplaceAll(
		prompt.TemplatePrincipleCollection,
		prompt.PlaceholderPrinciplesFile,
		principlesPath,
	)
}

// CollectorError represents an error during principle collection.
type CollectorError struct {
	Message string
	Err     error
}

func (e *CollectorError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CollectorError) Unwrap() error {
	return e.Err
}
