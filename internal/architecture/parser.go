package architecture

import (
	"regexp"
	"strings"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Section patterns for parsing Architecture output (supports ## and ### headers).
var (
	componentsPattern    = regexp.MustCompile(`(?i)##[#]?\s*Components\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	dependenciesPattern  = regexp.MustCompile(`(?i)##[#]?\s*Dependencies\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	fileStructurePattern = regexp.MustCompile(`(?i)##[#]?\s*File\s*Structure\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	techDecisionsPattern = regexp.MustCompile(`(?i)##[#]?\s*Technical\s*Decisions\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	titlePattern         = regexp.MustCompile(`(?i)##[#]?\s*(?:Title|Architecture\s*Title)\s*[:\n]\s*(.+)`)
	summaryPattern       = regexp.MustCompile(`(?i)##[#]?\s*(?:Summary|Overview)\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)
	rationalePattern     = regexp.MustCompile(`(?i)##[#]?\s*Rationale\s*\n([\s\S]*?)(?:##[#]?[^#]|$)`)

	// Component parsing patterns
	componentNamePattern  = regexp.MustCompile(`^\s*[-*]\s*\*?\*?([^:*]+)\*?\*?\s*:\s*(.*)$`)
	componentDescPattern  = regexp.MustCompile(`(?i)^\s*[-*]\s*(?:Description|Desc)\s*:\s*(.+)$`)
	componentFilesPattern = regexp.MustCompile(`(?i)^\s*[-*]\s*Files?\s*:\s*(.+)$`)
	listItemPattern       = regexp.MustCompile(`^\s*[-*\d.]+\s*(.+)$`)
)

// Parser extracts structured Architecture data from Claude's raw output.
type Parser struct{}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts Architecture structure from raw output.
func (p *Parser) Parse(rawOutput string) (*Architecture, error) {
	arch := &Architecture{
		Architecture: planner.Architecture{RawOutput: rawOutput},
	}

	// Extract complex sections
	arch.Components = p.extractComponents(rawOutput)
	arch.Dependencies = p.extractListItems(dependenciesPattern, rawOutput)
	arch.FileStructure = p.extractListItems(fileStructurePattern, rawOutput)
	arch.TechDecisions = p.extractListItems(techDecisionsPattern, rawOutput)

	// Extract extended sections
	arch.Title = p.extractSingleLine(titlePattern, rawOutput)
	arch.Summary = p.extractSection(summaryPattern, rawOutput)
	arch.Rationale = p.extractSection(rationalePattern, rawOutput)

	// Validate minimum required fields
	if len(arch.Components) == 0 {
		return nil, ErrParseNoComponents
	}

	return arch, nil
}

// extractComponents parses the Components section into structured Component objects.
// Expected formats:
// - **ComponentName**: Description
//   - Files: file1.go, file2.go
// OR
// - ComponentName: Description
//   - Files: file1.go file2.go
func (p *Parser) extractComponents(text string) []planner.Component {
	matches := componentsPattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil
	}

	section := matches[1]
	lines := strings.Split(section, "\n")

	var components []planner.Component
	var current *planner.Component

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for Files line FIRST (before component name pattern)
		// This prevents "- Files: x.go" from being parsed as a component named "Files"
		if filesMatch := componentFilesPattern.FindStringSubmatch(line); len(filesMatch) > 1 {
			if current != nil {
				filesStr := strings.TrimSpace(filesMatch[1])
				// Split by comma or space
				for _, f := range strings.FieldsFunc(filesStr, func(r rune) bool {
					return r == ',' || r == ' '
				}) {
					f = strings.TrimSpace(f)
					if f != "" {
						current.Files = append(current.Files, f)
					}
				}
			}
			continue
		}

		// Check for Description line (before component name pattern)
		if descMatch := componentDescPattern.FindStringSubmatch(line); len(descMatch) > 1 {
			if current != nil {
				current.Description = strings.TrimSpace(descMatch[1])
			}
			continue
		}

		// Check for component name line (starts with - or * at top level)
		if nameMatch := componentNamePattern.FindStringSubmatch(line); len(nameMatch) > 1 {
			// Only process if it's not an indented sub-item (indented lines have leading spaces)
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " \t"))
			if leadingSpaces < 4 || current == nil {
				if current != nil && current.Name != "" {
					components = append(components, *current)
				}
				name := strings.TrimSpace(nameMatch[1])
				desc := strings.TrimSpace(nameMatch[2])
				current = &planner.Component{
					Name:        name,
					Description: desc,
					Files:       []string{},
				}
				continue
			}
		}
	}

	// Don't forget the last component
	if current != nil && current.Name != "" {
		components = append(components, *current)
	}

	return components
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
