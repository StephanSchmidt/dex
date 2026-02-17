package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().PaddingLeft(2)

	cursorStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#00B4D8", Dark: "#00D4FF"}).
			PaddingLeft(1)

	slideNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#626262"})

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#626262"})

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF5F5F"}).
			Bold(true).PaddingLeft(2)
)

type UICmd struct {
	Dir string `arg:"" optional:"" default:"" help:"Presentation directory or file path."`
}

func (c *UICmd) Run() error {
	d, file, err := readDeck(c.Dir)
	if err != nil {
		return err
	}

	m := newUIModel(d.Slides, d.Metadata, file)
	p := tea.NewProgram(m, tea.WithOutput(stdout))
	final, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := final.(uiModel); ok && fm.err != nil {
		return fm.err
	}
	return nil
}

type uiModel struct {
	slides   []slide
	titles   []string
	file     string
	metadata string
	cursor   int
	selected int // -1 = normal mode
	width    int // terminal width from WindowSizeMsg
	height   int // terminal height from WindowSizeMsg
	offset   int // index of first visible slide
	err      error
}

func newUIModel(slides []slide, metadata, file string) uiModel {
	m := uiModel{
		slides:   slides,
		file:     file,
		metadata: metadata,
		cursor:   0,
		selected: -1,
	}
	m.rebuildTitles()
	return m
}

func (m *uiModel) rebuildTitles() {
	m.titles = make([]string, len(m.slides))
	for i, s := range m.slides {
		m.titles[i] = displayTitle(s)
	}
}

func (m *uiModel) writeback() {
	m.err = writeDeck(m.file, deck{Metadata: m.metadata, Slides: m.slides})
	m.rebuildTitles()
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wm.Width
		m.height = wm.Height
		m.clampOffset()
		return m, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.selected >= 0 {
		return m.updateSelected(km)
	}
	return m.updateNormal(km)
}

func (m uiModel) updateNormal(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.slides)-1 {
			m.cursor++
		}

	case "K", "shift+up":
		if m.cursor > 0 {
			m.slides[m.cursor], m.slides[m.cursor-1] = m.slides[m.cursor-1], m.slides[m.cursor]
			m.cursor--
			m.writeback()
		}

	case "J", "shift+down":
		if m.cursor < len(m.slides)-1 {
			m.slides[m.cursor], m.slides[m.cursor+1] = m.slides[m.cursor+1], m.slides[m.cursor]
			m.cursor++
			m.writeback()
		}

	case "d", "x":
		if len(m.slides) <= 1 {
			m.err = fmt.Errorf("cannot delete the only slide")
			return m, nil
		}
		m.slides = append(m.slides[:m.cursor], m.slides[m.cursor+1:]...)
		if m.cursor >= len(m.slides) {
			m.cursor = len(m.slides) - 1
		}
		m.writeback()

	case "enter", " ":
		m.selected = m.cursor
	}

	m.clampOffset()
	return m, nil
}

func (m uiModel) updateSelected(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "esc":
		m.selected = -1

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.slides)-1 {
			m.cursor++
		}

	case "m":
		src := m.selected
		dst := m.cursor
		if src != dst {
			s := m.slides[src]
			m.slides = append(m.slides[:src], m.slides[src+1:]...)
			// After removal, clamp insert index to valid range.
			ins := min(dst, len(m.slides))
			result := make([]slide, 0, len(m.slides)+1)
			result = append(result, m.slides[:ins]...)
			result = append(result, s)
			result = append(result, m.slides[ins:]...)
			m.slides = result
			m.cursor = dst
			m.writeback()
		}
		m.selected = -1

	case "c":
		src := m.selected
		dst := m.cursor
		copied := m.slides[src]
		result := make([]slide, 0, len(m.slides)+1)
		result = append(result, m.slides[:dst]...)
		result = append(result, copied)
		result = append(result, m.slides[dst:]...)
		m.slides = result
		m.cursor = dst
		m.selected = -1
		m.writeback()

	case "s":
		src := m.selected
		dst := m.cursor
		if src != dst {
			m.slides[src], m.slides[dst] = m.slides[dst], m.slides[src]
			m.writeback()
		}
		m.selected = -1
	}

	m.clampOffset()
	return m, nil
}

func (m *uiModel) visibleLines() int {
	if m.height == 0 {
		return len(m.slides)
	}
	chrome := 6 // appStyle padding (2) + title bar (1) + blank (1) + blank (1) + help (1)
	if m.err != nil {
		chrome += 2 // blank + error line
	}
	return max(m.height-chrome, 1)
}

func (m *uiModel) clampOffset() {
	visible := m.visibleLines()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func (m uiModel) View() string {
	var b strings.Builder

	contentWidth := m.width - appStyle.GetHorizontalPadding()
	if contentWidth <= 0 {
		contentWidth = 80
	}

	title := titleBarStyle.Width(contentWidth).Render("Slides")
	b.WriteString(title)
	b.WriteString("\n\n")

	visible := m.visibleLines()
	end := min(m.offset+visible, len(m.titles))
	for i := m.offset; i < end; i++ {
		num := slideNumStyle.Render(fmt.Sprintf("%2d", i+1))
		line := fmt.Sprintf("%s  %s", num, m.titles[i])

		switch {
		case m.cursor == i:
			b.WriteString(cursorStyle.Render(line))
		case m.selected == i:
			b.WriteString(selectedStyle.Render(line))
		default:
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.selected >= 0 {
		b.WriteString(helpStyle.Render("m move  c copy  s swap  esc cancel"))
	} else {
		b.WriteString(helpStyle.Render("enter select  j/k move  J/K nudge  d delete  q quit"))
	}

	return appStyle.Render(b.String())
}
