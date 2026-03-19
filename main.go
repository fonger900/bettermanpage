package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().
			BorderStyle(b).
			Padding(0, 1).
			Foreground(Mauve).
			Bold(true)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()

	tocStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(Subtext).
			Padding(0, 1).
			Foreground(Text)

	tocSelectedStyle = lipgloss.NewStyle().
				Foreground(Base).
				Background(Blue).
				Bold(true).
				Padding(0, 1)

	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Pink).
			Padding(1, 2).
			Background(Base).
			Foreground(Text)
)

type tldrMsg struct {
	content string
	err     error
}

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
	showTLDR    bool
	tldrContent string
	tldrLoading bool
	lastKey     string
}

func (m model) Init() tea.Cmd {
	return nil
}

func fetchTLDR(page string) tea.Cmd {
	return func() tea.Msg {
		// Try common first
		url := fmt.Sprintf("https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/common/%s.md", page)
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			// Try linux if common fails
			url = fmt.Sprintf("https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/linux/%s.md", page)
			resp, err = http.Get(url)
			if err != nil || resp.StatusCode != 200 {
				// Try osx
				url = fmt.Sprintf("https://raw.githubusercontent.com/tldr-pages/tldr/main/pages/osx/%s.md", page)
				resp, err = http.Get(url)
				if err != nil || resp.StatusCode != 200 {
					return tldrMsg{err: fmt.Errorf("could not fetch TLDR for %s", page)}
				}
			}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return tldrMsg{err: err}
		}
		return tldrMsg{content: string(body)}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tldrMsg:
		m.tldrLoading = false
		if msg.err != nil {
			m.tldrContent = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(msg.err.Error())
		} else {
			m.tldrContent = highlightTLDR(msg.content)
		}
		return m, nil

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

		if m.showTLDR {
			if msg.String() == "esc" || msg.String() == "e" || msg.String() == "q" {
				m.showTLDR = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.showTOC {
				m.showTOC = false
				m.updateLayout()
				return m, nil
			}
			return m, nil
		case "tab":
			m.showTOC = !m.showTOC
			m.updateLayout()
		case "/":
			m.showSearch = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, nil
		case "e":
			m.showTLDR = true
			m.tldrLoading = true
			m.tldrContent = ""
			return m, fetchTLDR(os.Args[1])
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
			m.viewport.LineUp(1)
		case "down", "j":
			if m.showTOC {
				if m.tocIndex < len(m.sections)-1 {
					m.tocIndex++
				}
				return m, nil
			}
			m.viewport.LineDown(1)
		case "g":
			if m.lastKey == "g" {
				m.viewport.GotoTop()
				m.lastKey = ""
			} else {
				m.lastKey = "g"
			}
			m.updateHighlight()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
		case "d":
			m.viewport.HalfPageDown()
		case "u":
			m.viewport.HalfPageUp()
		case "ctrl+f", "pgdown":
			m.viewport.PageDown()
		case "ctrl+b", "pgup":
			m.viewport.PageUp()
		case "enter":
			if m.showTOC && len(m.sections) > 0 {
				m.viewport.SetYOffset(m.sections[m.tocIndex].Start)
				return m, nil
			}
		}

		if msg.String() != "g" {
			m.lastKey = ""
		}
		m.updateHighlight()

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

func (m *model) updateHighlight() {
	var currentMatchLine int = -1
	if len(m.lastMatches) > 0 {
		currentMatchLine = m.lastMatches[m.matchIndex]
	}
	m.highlighted = highlight(m.content, m.searchTerm, currentMatchLine)
	m.viewport.SetContent(m.highlighted)
}

func (m *model) searchContent() {
	if m.searchTerm == "" {
		m.lastMatches = nil
		m.updateHighlight()
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
	m.updateHighlight()
}

func (m *model) nextMatch() {
	if len(m.lastMatches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex + 1) % len(m.lastMatches)
	m.viewport.SetYOffset(m.lastMatches[m.matchIndex])
	m.updateHighlight()
}

func (m *model) prevMatch() {
	if len(m.lastMatches) == 0 {
		return
	}
	m.matchIndex = (m.matchIndex - 1 + len(m.lastMatches)) % len(m.lastMatches)
	m.viewport.SetYOffset(m.lastMatches[m.matchIndex])
	m.updateHighlight()
}

func (m model) getCurrentSection() string {
	if len(m.sections) == 0 {
		return ""
	}
	y := m.viewport.YOffset
	current := m.sections[0].Name
	for _, s := range m.sections {
		if y >= s.Start {
			current = s.Name
		} else {
			break
		}
	}
	return current
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	mainContent := m.viewport.View()
	if m.showTOC {
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, m.tocView(), mainContent)
	}

	header := m.headerView()
	footer := m.footerView()

	view := fmt.Sprintf("%s\n%s\n%s", header, mainContent, footer)

	if m.showTLDR {
		return m.renderTLDR()
	}

	return view
}

func (m model) renderTLDR() string {
	content := m.tldrContent
	if m.tldrLoading {
		loadingStyle := lipgloss.NewStyle().Foreground(Mauve).Bold(true).Italic(true)
		content = "\n\n  " + loadingStyle.Render("Searching tldr-pages...")
	}

	w := m.width * 80 / 100
	h := m.height * 80 / 100
	if w > 100 {
		w = 100
	}
	if h > 30 {
		h = 30
	}

	title := lipgloss.NewStyle().
		Background(Pink).
		Foreground(Base).
		Bold(true).
		Padding(0, 1).
		Render(" TLDR Cheat Sheet ")

	style := overlayStyle.
		Width(w).
		Height(h).
		Align(lipgloss.Left, lipgloss.Top)

	body := style.Render(content)
	
	// Place the title on top of the border
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, title, body),
	)
}

func (m model) tocView() string {
	var toc strings.Builder
	unselectedStyle := lipgloss.NewStyle().Foreground(Subtext)
	for i, section := range m.sections {
		name := section.Name
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		if i == m.tocIndex {
			toc.WriteString(tocSelectedStyle.Render(name))
		} else {
			toc.WriteString("  " + unselectedStyle.Render(name))
		}
		toc.WriteString("\n")
	}
	return tocStyle.Height(m.viewport.Height).Render(toc.String())
}

func (m model) headerView() string {
	title := titleStyle.Render("BetterManPage")
	page := ""
	if len(os.Args) > 1 {
		page = os.Args[1]
	}

	section := m.getCurrentSection()
	breadcrumb := ""
	if page != "" {
		pStyle := lipgloss.NewStyle().Foreground(Pink).Bold(true)
		sStyle := lipgloss.NewStyle().Foreground(Peach)
		sepStyle := lipgloss.NewStyle().Foreground(Subtext)
		breadcrumb = fmt.Sprintf(" %s %s %s", pStyle.Render(page), sepStyle.Render(">"), sStyle.Render(section))
	}

	lineCount := m.width - lipgloss.Width(title) - lipgloss.Width(breadcrumb)
	line := lipgloss.NewStyle().Foreground(Subtext).Render(strings.Repeat("─", max(0, lineCount)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, breadcrumb)
}

func (m model) footerView() string {
	percentStyle := lipgloss.NewStyle().Foreground(Mauve).Bold(true)
	info := infoStyle.Render(percentStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)))

	var status string
	statusStyle := lipgloss.NewStyle().Foreground(Green)
	highlightStatusStyle := lipgloss.NewStyle().Foreground(Pink).Bold(true)

	if m.showSearch {
		status = " Search: " + m.searchInput.View()
	} else if len(m.lastMatches) > 0 {
		status = fmt.Sprintf(" Search: [%s/%s] %s ",
			highlightStatusStyle.Render(fmt.Sprint(m.matchIndex+1)),
			highlightStatusStyle.Render(fmt.Sprint(len(m.lastMatches))),
			statusStyle.Render(m.searchTerm))
	} else {
		helpStyle := lipgloss.NewStyle().Foreground(Subtext)
		status = helpStyle.Render(" [Tab: TOC | /: Search | e: TLDR | q: Quit]")
	}

	lineCount := m.width - lipgloss.Width(info) - lipgloss.Width(status)
	line := lipgloss.NewStyle().Foreground(Subtext).Render(strings.Repeat("─", max(0, lineCount)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, status, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func getManPage(page string) (string, error) {
	width := getTerminalWidth()

	// Set environment variables for consistent output
	// MANPAGER=cat ensures man doesn't open its own pager
	// MANWIDTH ensures the output fits the terminal width
	// GROFF_NO_SGR=1 forces the use of backspaces for bold/underline instead of ANSI
	env := os.Environ()
	env = append(env, "MANPAGER=cat")
	env = append(env, fmt.Sprintf("MANWIDTH=%d", width))
	env = append(env, "GROFF_NO_SGR=1")
	env = append(env, "LANG=en_US.UTF-8")
	env = append(env, "LC_ALL=en_US.UTF-8")

	cmd := exec.Command("man", page)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback for mandoc
		cmd = exec.Command("man", "-Tutf8", "-O", fmt.Sprintf("width=%d", width), page)
		output2, err2 := cmd.CombinedOutput()
		if err2 == nil {
			output = output2
		} else {
			return "", fmt.Errorf("man %s failed: %v", page, err)
		}
	}

	return string(output), nil
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
		highlighted: highlight(content, "", -1),
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
