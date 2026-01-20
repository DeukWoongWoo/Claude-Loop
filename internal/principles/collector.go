// Package principles provides principle collection functionality for the Claude Loop.
package principles

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
)

// Collector handles the collection of project principles via interactive prompts.
type Collector struct {
	principlesPath string
	reader         io.Reader
}

// NewCollector creates a new Collector.
func NewCollector(principlesPath string) *Collector {
	return &Collector{
		principlesPath: principlesPath,
		reader:         os.Stdin,
	}
}

// NewCollectorWithReader creates a new Collector with a custom reader (for testing).
func NewCollectorWithReader(principlesPath string, reader io.Reader) *Collector {
	return &Collector{
		principlesPath: principlesPath,
		reader:         reader,
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

// Collect runs an interactive CLI session to collect principles from the user.
// The principles are saved to the configured principlesPath.
func (c *Collector) Collect(ctx context.Context) error {
	reader := bufio.NewReader(c.reader)

	// Step 1: Ask project type
	preset, err := c.askProjectType(ctx, reader)
	if err != nil {
		return err
	}

	// Step 2: Load defaults and ask follow-up questions
	principles := config.DefaultPrinciples(preset)
	if err := c.askFollowUpQuestions(ctx, reader, preset, principles); err != nil {
		return err
	}

	// Step 3: Set timestamp and save
	principles.CreatedAt = time.Now().Format("2006-01-02")
	if err := config.SaveToFile(c.principlesPath, principles); err != nil {
		return &CollectorError{
			Message: "failed to save principles file",
			Err:     err,
		}
	}

	return nil
}

// askProjectType prompts the user to select a project type.
func (c *Collector) askProjectType(ctx context.Context, reader *bufio.Reader) (config.Preset, error) {
	select {
	case <-ctx.Done():
		return "", &CollectorError{Message: "cancelled", Err: ctx.Err()}
	default:
	}

	fmt.Println("Select project type:")
	fmt.Println("  1) Startup/MVP - Fast validation, focus on core features")
	fmt.Println("  2) Enterprise - Stability first, thorough testing")
	fmt.Println("  3) Open Source - Community contributions, API stability")
	fmt.Print("> ")

	input, _ := reader.ReadString('\n')
	switch strings.TrimSpace(input) {
	case "1":
		return config.PresetStartup, nil
	case "2":
		return config.PresetEnterprise, nil
	case "3":
		return config.PresetOpenSource, nil
	default:
		return config.PresetStartup, nil
	}
}

// askFollowUpQuestions asks preset-specific follow-up questions.
func (c *Collector) askFollowUpQuestions(ctx context.Context, reader *bufio.Reader, preset config.Preset, p *config.Principles) error {
	fmt.Println()
	var err error
	switch preset {
	case config.PresetStartup:
		p.Layer0.ScopePhilosophy, err = c.askNumber(ctx, reader, "MVP scope (1=minimal, 10=expansive)", 3)
		if err != nil {
			return err
		}
		p.Layer1.SpeedCorrectness, err = c.askNumber(ctx, reader, "Speed vs Quality (1=speed, 10=quality)", 4)
		if err != nil {
			return err
		}
	case config.PresetEnterprise:
		p.Layer1.BlastRadius, err = c.askNumber(ctx, reader, "Change size - how large changes can be (1=large sweeping changes, 10=small incremental changes)", 9)
		if err != nil {
			return err
		}
		p.Layer1.InnovationStability, err = c.askNumber(ctx, reader, "Tech choice - technology preference (1=new/experimental tech, 10=proven/stable tech)", 8)
		if err != nil {
			return err
		}
	case config.PresetOpenSource:
		p.Layer0.CurationModel, err = c.askNumber(ctx, reader, "Contributions (1=open, 10=verified)", 5)
		if err != nil {
			return err
		}
		p.Layer0.UXPhilosophy, err = c.askNumber(ctx, reader, "UX (1=easy, 10=powerful)", 5)
		if err != nil {
			return err
		}
	}
	return nil
}

// askNumber prompts for a number input with a default value.
func (c *Collector) askNumber(ctx context.Context, reader *bufio.Reader, prompt string, defaultVal int) (int, error) {
	select {
	case <-ctx.Done():
		return 0, &CollectorError{Message: "cancelled", Err: ctx.Err()}
	default:
	}

	fmt.Printf("%s [%d]: ", prompt, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	if n, err := strconv.Atoi(input); err == nil && n >= 1 && n <= 10 {
		return n, nil
	}
	return defaultVal, nil
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
