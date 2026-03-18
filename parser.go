package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	optionStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
	commentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	tldrTitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Underline(true)
	tldrDescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	tldrExampleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	tldrCodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))

	searchStyle        = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("255"))
	currentSearchStyle = lipgloss.NewStyle().Background(lipgloss.Color("205")).Foreground(lipgloss.Color("255")).Bold(true)
)

func highlight(content string, searchTerm string, currentMatchLine int) string {
	lines := strings.Split(content, "\n")
	highlighted := make([]string, len(lines))

	// Regex for headers (all caps, starts at col 0, at least 2 chars)
	headerRegex := regexp.MustCompile(`^[A-Z][A-Z\s\(\)0-9_-]+$`)
	// Regex for options (-a, --all, -abc)
	optionRegex := regexp.MustCompile(`(\s|^)(-[a-zA-Z0-9-]|--[a-zA-Z0-9-]+)`)
	// Regex for values/placeholders (<value>, [FILE], VARIABLE_NAME)
	valueRegex := regexp.MustCompile(`(<[^>]+>|\[[^\]]+\]|\b[A-Z_][A-Z_]{2,}\b)`)

	var searchRegex *regexp.Regexp
	if searchTerm != "" {
		searchRegex = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(searchTerm))
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			highlighted[i] = ""
			continue
		}

		var newLine string
		if headerRegex.MatchString(line) {
			newLine = headerStyle.Render(line)
		} else {
			// Apply highlighting to parts of the line
			newLine = line

			newLine = optionRegex.ReplaceAllStringFunc(newLine, func(match string) string {
				prefix := ""
				if strings.HasPrefix(match, " ") {
					prefix = " "
					match = match[1:]
				}
				return prefix + optionStyle.Render(match)
			})

			newLine = valueRegex.ReplaceAllStringFunc(newLine, func(match string) string {
				return valueStyle.Render(match)
			})
		}

		// Apply search highlighting last
		if searchRegex != nil {
			// Split by ANSI sequences to avoid matching inside them
			parts := splitByANSI(newLine)
			for j, part := range parts {
				if !strings.HasPrefix(part, "\x1b[") {
					parts[j] = searchRegex.ReplaceAllStringFunc(part, func(match string) string {
						style := searchStyle
						if i == currentMatchLine {
							style = currentSearchStyle
						}
						return style.Render(match)
					})
				}
			}
			newLine = strings.Join(parts, "")
		}

		highlighted[i] = newLine
	}

	return strings.Join(highlighted, "\n")
}

func splitByANSI(s string) []string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	matches := ansiRegex.FindAllStringIndex(s, -1)
	if matches == nil {
		return []string{s}
	}

	var parts []string
	lastPos := 0
	for _, m := range matches {
		if m[0] > lastPos {
			parts = append(parts, s[lastPos:m[0]])
		}
		parts = append(parts, s[m[0]:m[1]])
		lastPos = m[1]
	}
	if lastPos < len(s) {
		parts = append(parts, s[lastPos:])
	}
	return parts
}

func highlightTLDR(content string) string {
	lines := strings.Split(content, "\n")
	highlighted := make([]string, len(lines))

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			highlighted[i] = tldrTitleStyle.Render(trimmed)
		} else if strings.HasPrefix(trimmed, ">") {
			highlighted[i] = tldrDescStyle.Render(trimmed)
		} else if strings.HasPrefix(trimmed, "-") {
			highlighted[i] = tldrExampleStyle.Render(trimmed)
		} else if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") {
			code := trimmed[1 : len(trimmed)-1]
			highlighted[i] = "  " + tldrCodeStyle.Render(code)
		} else {
			highlighted[i] = trimmed
		}
	}

	return strings.Join(highlighted, "\n")
}

type Section struct {
	Name  string
	Start int
}

func parseSections(content string) []Section {
	lines := strings.Split(content, "\n")
	headerRegex := regexp.MustCompile(`^[A-Z][A-Z\s\(\)0-9_-]+$`)
	var sections []Section

	for i, line := range lines {
		if headerRegex.MatchString(line) {
			name := strings.TrimSpace(line)
			if name != "" {
				sections = append(sections, Section{
					Name:  name,
					Start: i,
				})
			}
		}
	}
	return sections
}
