package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Foreground(Pink).Bold(true)
	optionStyle  = lipgloss.NewStyle().Foreground(Green)
	valueStyle   = lipgloss.NewStyle().Foreground(Peach).Italic(true)
	commentStyle = lipgloss.NewStyle().Foreground(Subtext)

	tldrTitleStyle   = lipgloss.NewStyle().Foreground(Pink).Bold(true).Underline(true)
	tldrDescStyle    = lipgloss.NewStyle().Foreground(Subtext)
	tldrExampleStyle = lipgloss.NewStyle().Foreground(Green)
	tldrCodeStyle    = lipgloss.NewStyle().Foreground(Sapphire)

	searchStyle        = lipgloss.NewStyle().Background(Overlay).Foreground(Sky)
	currentSearchStyle = lipgloss.NewStyle().Background(Pink).Foreground(Base).Bold(true)

	// Built-in roff styles
	boldStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff"))
	underlineStyle = lipgloss.NewStyle().Underline(true).Foreground(Sky)

	// Advanced styles
	pathStyle    = lipgloss.NewStyle().Foreground(Sapphire).Underline(true)
	urlStyle     = lipgloss.NewStyle().Foreground(Sky).Underline(true)
	exampleStyle = lipgloss.NewStyle().Foreground(Lavender)
	stringStyle  = lipgloss.NewStyle().Foreground(Green).Italic(true)
)

type Cell struct {
	Char      rune
	Bold      bool
	Underline bool
	Style     lipgloss.Style // Custom style override
}

func highlight(rawContent string, searchTerm string, currentMatchLine int) string {
	lines := strings.Split(rawContent, "\n")
	highlightedLines := make([]string, len(lines))

	// Pre-compile regexes for performance
	optionRegex := regexp.MustCompile(`(\s|^)(-[a-zA-Z0-9-]|--[a-zA-Z0-9-]+)`)
	pathRegex := regexp.MustCompile(`\B/(?:[a-zA-Z0-9._-]+/?)*`)
	urlRegex := regexp.MustCompile(`https?://[^\s)\]]+`)
	valueRegex := regexp.MustCompile(`(<[^>]+>|\[[^\]]+\]|\b[A-Z_][A-Z_]{2,}\b)`)
	stringRegex := regexp.MustCompile(`"([^"]+)"|'([^']+)'`)

	var searchRegex *regexp.Regexp
	if searchTerm != "" {
		searchRegex = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(searchTerm))
	}

	for i, rawLine := range lines {
		cells := parseFormattedLine(rawLine)
		plain := cellsToPlain(cells)

		// 1. Identify Sections
		isHeader := isHeaderLine(plain)

		// 2. Multi-pass semantic highlighting on plain text
		applyRegexStyle(cells, pathRegex, pathStyle)
		applyRegexStyle(cells, urlRegex, urlStyle)
		applyRegexStyle(cells, optionRegex, optionStyle)
		applyRegexStyle(cells, valueRegex, valueStyle)
		applyRegexStyle(cells, stringRegex, stringStyle)

		// 3. Search matches
		var searchMatches [][]int
		if searchRegex != nil {
			searchMatches = searchRegex.FindAllStringIndex(plain, -1)
		}

		// Build the final styled line
		var b strings.Builder
		for j, cell := range cells {
			style := cell.Style

			// Apply troff styles if not already overridden by semantic style
			if cell.Bold {
				style = style.Bold(true)
			}
			if cell.Underline {
				style = style.Underline(true)
			}

			if isHeader {
				style = headerStyle
			}

			// Apply Search styling (last override)
			for _, m := range searchMatches {
				if j >= m[0] && j < m[1] {
					if i == currentMatchLine {
						style = currentSearchStyle
					} else {
						style = searchStyle
					}
				}
			}

			b.WriteString(style.Render(string(cell.Char)))
		}
		highlightedLines[i] = b.String()
	}

	return strings.Join(highlightedLines, "\n")
}

func applyRegexStyle(cells []Cell, re *regexp.Regexp, style lipgloss.Style) {
	plain := cellsToPlain(cells)
	matches := re.FindAllStringIndex(plain, -1)

	// Create a map from byte index to rune index
	byteToRune := make(map[int]int)
	byteIdx := 0
	for runeIdx, r := range []rune(plain) {
		byteToRune[byteIdx] = runeIdx
		byteIdx += len(string(r))
	}
	byteToRune[byteIdx] = len(cells) // End boundary

	for _, m := range matches {
		startRune, startOk := byteToRune[m[0]]
		endRune, endOk := byteToRune[m[1]]

		if startOk && endOk {
			for i := startRune; i < endRune; i++ {
				// Special handling for patterns that might have a leading space (like options)
				if i == startRune && plain[i] == ' ' {
					continue
				}
				if i < len(cells) {
					cells[i].Style = style
				}
			}
		}
	}
}

func parseFormattedLine(line string) []Cell {
	var cells []Cell
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		if i+2 < len(runes) && runes[i+1] == '\b' {
			char := runes[i+2]
			if runes[i] == '_' {
				cells = append(cells, Cell{Char: char, Underline: true})
			} else if runes[i] == char {
				cells = append(cells, Cell{Char: char, Bold: true})
			} else {
				cells = append(cells, Cell{Char: char})
			}
			i += 2
			continue
		}
		if runes[i] == '\b' {
			continue
		}
		cells = append(cells, Cell{Char: runes[i]})
	}
	return cells
}

func cellsToPlain(cells []Cell) string {
	var b strings.Builder
	for _, c := range cells {
		b.WriteRune(c.Char)
	}
	return b.String()
}

func isHeaderLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	headerRegex := regexp.MustCompile(`^[A-Z][A-Z\s\(\)0-9_-]+$`)
	return headerRegex.MatchString(line)
}

type Section struct {
	Name  string
	Start int
}

func parseSections(rawContent string) []Section {
	lines := strings.Split(rawContent, "\n")
	var sections []Section

	for i, rawLine := range lines {
		cells := parseFormattedLine(rawLine)
		plain := cellsToPlain(cells)
		if isHeaderLine(plain) {
			name := strings.TrimSpace(plain)
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
