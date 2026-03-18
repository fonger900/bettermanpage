package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Foreground(lipgloss.Color("205")).Bold(true)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	tocStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			Padding(0, 1)

	tocSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Bold(true)
)

type model struct {
	content     string
	highlighted string
	ready       bool
	viewport    viewport.Model
	sections    []Section
	showTOC     bool
	tocIndex    int
	searchInput textinput.Model
	showSearch  bool
	searchTerm  string
	lastMatches []int
	matchIndex  int
	width       int
	height      int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showSearch {
			if msg.String() == "enter" {
				m.searchTerm = m.searchInput.Value()
				m.showSearch = false
				m.searchContent()
				return m, nil
			}
			if msg.String() == "esc" {
				m.showSearch = false
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab":
			m.showTOC = !m.showTOC
			m.updateLayout()
		case "/":
			m.showSearch = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, nil
		case "n":
			m.nextMatch()
		case "N":
			m.prevMatch()
		case "up", "k":
			if m.showTOC {
				if m.tocIndex > 0 {
					m.tocIndex--
				}
				return m, nil
			}
		case "down", "j":
			if m.showTOC {
				if m.tocIndex < len(m.sections)-1 {
					m.tocIndex++
				}
				return m, nil
			}
		case "enter":
			if m.showTOC && len(m.sections) > 0 {
				m.viewport.SetYOffset(m.sections[m.tocIndex].Start)
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.updateLayout()
			m.viewport.SetContent(m.highlighted)
			m.ready = true
		} else {
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.updateLayout()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateLayout() {
	if m.showTOC {
		m.viewport.Width = m.width - 26
	} else {
		m.viewport.Width = m.width
	}
}

func (m *model) searchContent() {
	if m.searchTerm == "" {
		m.lastMatches = nil
		return
	}
	m.lastMatches = nil
	lines := strings.Split(m.content, "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(m.searchTerm)) {
			m.lastMatches = append(m.lastMatches, i)
		}
	}
	if len(m.lastMatches) > 0 {
		m.matchIndex = 0
		m.viewport.SetYOffset(m.lastMatches[0])
	}
}

func (m *model) nextMatch() {
	if len(m.lastMatches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex + 1) % len(m.lastMatches)
	m.viewport.SetYOffset(m.lastMatches[m.matchIndex])
}

func (m *model) prevMatch() {
	if len(m.lastMatches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex - 1 + len(m.lastMatches)) % len(m.lastMatches)
	m.viewport.SetYOffset(m.lastMatches[m.matchIndex])
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	mainContent := m.viewport.View()
	if m.showTOC {
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, m.tocView(), mainContent)
	}

	footer := m.footerView()
	if m.showSearch {
		footer = "Search: " + m.searchInput.View()
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), mainContent, footer)
}

func (m model) tocView() string {
	var toc strings.Builder
	for i, section := range m.sections {
		name := section.Name
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		if i == m.tocIndex {
			toc.WriteString(tocSelectedStyle.Render("> " + name))
		} else {
			toc.WriteString("  " + name)
		}
		toc.WriteString("\n")
	}
	return tocStyle.Height(m.viewport.Height).Render(toc.String())
}

func (m model) headerView() string {
	title := titleStyle.Render("BetterManPage")
	info := ""
	if len(os.Args) > 1 {
		info = " [" + os.Args[1] + "]"
	}
	lineCount := m.width - lipgloss.Width(title) - lipgloss.Width(info)
	line := strings.Repeat("─", max(0, lineCount))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, info)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	help := " [Tab: TOC | /: Search | n/N: Nav | q: Quit]"
	lineCount := m.width - lipgloss.Width(info) - lipgloss.Width(help)
	line := strings.Repeat("─", max(0, lineCount))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, help, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getManPage(page string) (string, error) {
	cmd := exec.Command("man", page)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	colCmd := exec.Command("col", "-b")
	colCmd.Stdin = strings.NewReader(string(output))
	finalOutput, err := colCmd.Output()
	if err != nil {
		return string(output), nil
	}

	return string(finalOutput), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bman <page>")
		os.Exit(1)
	}

	page := os.Args[1]
	content, err := getManPage(page)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ti := textinput.New()
	ti.Placeholder = "Search term..."
	ti.CharLimit = 156
	ti.Width = 20

	m := model{
		content:     content,
		highlighted: highlight(content),
		sections:    parseSections(content),
		searchInput: ti,
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
