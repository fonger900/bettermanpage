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
)

func highlight(content string) string {
	lines := strings.Split(content, "\n")
	highlighted := make([]string, len(lines))

	// Regex for headers (all caps, starts at col 0, at least 2 chars)
	headerRegex := regexp.MustCompile(`^[A-Z][A-Z\s\(\)0-9_-]+$`)
	// Regex for options (-a, --all, -abc)
	optionRegex := regexp.MustCompile(`(\s|^)(-[a-zA-Z0-9-]|--[a-zA-Z0-9-]+)`)
	// Regex for values/placeholders (<value>, [FILE], VARIABLE_NAME)
	valueRegex := regexp.MustCompile(`(<[^>]+>|\[[^\]]+\]|\b[A-Z_][A-Z_]{2,}\b)`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			highlighted[i] = ""
			continue
		}

		if headerRegex.MatchString(line) {
			highlighted[i] = headerStyle.Render(line)
			continue
		}

		// Apply highlighting to parts of the line
		newLine := line

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

		highlighted[i] = newLine
	}

	return strings.Join(highlighted, "\n")
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
