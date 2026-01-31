package prd

import (
	"regexp"
	"strings"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Section patterns for parsing PRD output (supports ## and ### headers).
var (
	goalsPattern           = regexp.MustCompile(`(?i)##[#]?\s*Goals\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	requirementsPattern    = regexp.MustCompile(`(?i)##[#]?\s*Requirements\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	constraintsPattern     = regexp.MustCompile(`(?i)##[#]?\s*Constraints\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	successCriteriaPattern = regexp.MustCompile(`(?i)##[#]?\s*Success\s*Criteria\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	titlePattern           = regexp.MustCompile(`(?i)##[#]?\s*(?:Title|PRD Title)\s*[:\n]\s*(.+)`)
	summaryPattern         = regexp.MustCompile(`(?i)##[#]?\s*(?:Summary|Overview)\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	outOfScopePattern      = regexp.MustCompile(`(?i)##[#]?\s*(?:Out\s*of\s*Scope|Non-Goals)\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	listItemPattern        = regexp.MustCompile(`^\s*[-*\d.]+\s*(.+)$`)
)

// Parser extracts structured PRD data from Claude's raw output.
type Parser struct{}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts PRD structure from raw output.
func (p *Parser) Parse(rawOutput string) (*PRD, error) {
	prd := &PRD{
		PRD: planner.PRD{RawOutput: rawOutput},
	}

	// Extract core sections
	prd.Goals = p.extractListItems(goalsPattern, rawOutput)
	prd.Requirements = p.extractListItems(requirementsPattern, rawOutput)
	prd.Constraints = p.extractListItems(constraintsPattern, rawOutput)
	prd.SuccessCriteria = p.extractListItems(successCriteriaPattern, rawOutput)

	// Extract extended sections
	prd.Title = p.extractSingleLine(titlePattern, rawOutput)
	prd.Summary = p.extractSection(summaryPattern, rawOutput)
	prd.OutOfScope = p.extractListItems(outOfScopePattern, rawOutput)

	// Validate minimum required fields
	if len(prd.Goals) == 0 {
		return nil, ErrParseNoGoals
	}
	if len(prd.Requirements) == 0 {
		return nil, ErrParseNoRequirements
	}

	return prd, nil
}

// extractListItems extracts list items from a section matched by pattern.
func (p *Parser) extractListItems(pattern *regexp.Regexp, text string) []string {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil
	}

	var items []string
	for _, line := range strings.Split(matches[1], "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if itemMatch := listItemPattern.FindStringSubmatch(line); len(itemMatch) > 1 {
			if item := strings.TrimSpace(itemMatch[1]); item != "" {
				items = append(items, item)
			}
		}
	}
	return items
}

// extractSingleLine extracts a single line value matched by pattern.
func (p *Parser) extractSingleLine(pattern *regexp.Regexp, text string) string {
	if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractSection extracts a full section as text matched by pattern.
func (p *Parser) extractSection(pattern *regexp.Regexp, text string) string {
	if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
