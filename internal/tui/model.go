package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kaiser/ledger-cli/internal/client"
	"github.com/kaiser/ledger-cli/internal/reason"
	"github.com/kaiser/ledger-cli/internal/ui"
)

// Model is the main TUI model
type Model struct {
	projects    []ProjectRow
	review      ReviewData
	currentTab  int // 0 = Projects, 1 = Review
	selected    int // currently selected project index
	viewport    viewport.Model
	width       int
	height      int
	ready       bool
	loading     bool
	err         error

	// Touch form state
	touchMode   bool
	touchSlug   string
	touchReason string
	touchField  int // 0 = slug, 1 = reason
	touchMsg    string // success or error message

	// Help modal
	helpMode bool
}

// ProjectRow represents a project in the list
type ProjectRow struct {
	ID          string
	Name        string
	LastTouched time.Time
	Reason      string
	NextAction  string
	StatusDot   string
	DaysAgo     int
}

// ReviewData holds the three review buckets with actual rows
type ReviewData struct {
	Neglected       []ProjectRow
	MissingNext     []ProjectRow
	RecentlyTouched []ProjectRow
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		currentTab: 0,
		selected:   0,
		loading:    true,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return fetchProjectsCmd
}

// fetchProjectsCmd loads projects from Firestore
func fetchProjectsCmd() tea.Msg {
	ctx := context.Background()
	c, err := client.New(ctx)
	if err != nil {
		return errMsg{err}
	}
	defer c.Close()

	projects, err := c.GetAllProjects(ctx)
	if err != nil {
		return errMsg{err}
	}

	var rows []ProjectRow
	for _, p := range projects {
		days := int(time.Since(p.LastTouched).Hours() / 24)
		dot := ui.StatusDot(p.LastTouched)

		rows = append(rows, ProjectRow{
			ID:          p.ID,
			Name:        p.Name,
			LastTouched: p.LastTouched,
			Reason:      p.LastTouchReason,
			NextAction:  p.NextAction,
			StatusDot:   dot,
			DaysAgo:     days,
		})
	}

	// Build review buckets
	var neglected, missing, recent []ProjectRow
	now := time.Now()

	for _, r := range rows {
		days := int(now.Sub(r.LastTouched).Hours() / 24)
		if days >= 10 {
			neglected = append(neglected, r)
		}
		if r.NextAction == "" {
			missing = append(missing, r)
		}
		if days <= 3 {
			recent = append(recent, r)
		}
	}

	return projectsLoadedMsg{
		projects: rows,
		review: ReviewData{
			Neglected:       neglected,
			MissingNext:     missing,
			RecentlyTouched: recent,
		},
	}
}

// Messages
type errMsg struct{ err error }
type projectsLoadedMsg struct {
	projects []ProjectRow
	review   ReviewData
}
type touchResultMsg struct {
	success bool
	message string
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-8)
			m.ready = true
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 8

	case projectsLoadedMsg:
		m.projects = msg.projects
		m.review = msg.review
		m.loading = false
		m.updateViewportContent()

	case errMsg:
		m.err = msg.err
		m.loading = false

	case touchResultMsg:
		m.touchMsg = msg.message
		if msg.success {
			m.touchSlug = ""
			m.touchReason = ""
			m.touchField = 0
			m.touchMode = false
			return m, fetchProjectsCmd
		}

	case tea.KeyMsg:
		if m.helpMode {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.helpMode = false
			}
			return m, nil
		}

		if m.touchMode {
			return m.handleTouchInput(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "right":
			m.currentTab = (m.currentTab + 1) % 2
			m.selected = 0

		case "shift+tab", "left":
			m.currentTab = (m.currentTab - 1 + 2) % 2
			m.selected = 0

		case "up", "k":
			if m.currentTab == 0 && m.selected > 0 {
				m.selected--
				m.ensureSelectedVisible()
			}

		case "down", "j":
			if m.currentTab == 0 && m.selected < len(m.projects)-1 {
				m.selected++
				m.ensureSelectedVisible()
			}

		case "r":
			// Refresh data
			m.loading = true
			return m, fetchProjectsCmd

		case "t":
			m.touchMode = true
			m.touchMsg = ""
			if len(m.projects) > 0 {
				m.touchSlug = m.projects[m.selected].Name
			}
			return m, nil

		case "?":
			m.helpMode = true
			return m, nil
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) handleTouchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.touchMode = false
		m.touchSlug = ""
		m.touchReason = ""
		m.touchField = 0
		m.touchMsg = ""

	case "tab":
		m.touchField = (m.touchField + 1) % 2

	case "enter":
		return m, m.performTouchCmd()

	case "backspace":
		if m.touchField == 0 {
			if len(m.touchSlug) > 0 {
				m.touchSlug = m.touchSlug[:len(m.touchSlug)-1]
			}
		} else {
			if len(m.touchReason) > 0 {
				m.touchReason = m.touchReason[:len(m.touchReason)-1]
			}
		}

	default:
		if len(msg.String()) == 1 {
			if m.touchField == 0 {
				m.touchSlug += msg.String()
			} else {
				m.touchReason += msg.String()
			}
		}
	}

	return m, nil
}

func (m Model) performTouchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		c, err := client.New(ctx)
		if err != nil {
			return touchResultMsg{success: false, message: "Auth failed: " + err.Error()}
		}
		defer c.Close()

		finalReason := m.touchReason
		if preset, ok := reason.Presets[m.touchReason]; ok {
			finalReason = preset
		}
		if finalReason == "" {
			finalReason = "none"
		}

		err = c.TouchProject(ctx, m.touchSlug, finalReason)
		if err != nil {
			return touchResultMsg{success: false, message: "Touch failed: " + err.Error()}
		}

		return touchResultMsg{
			success: true,
			message: fmt.Sprintf("Touched %s (%s)", m.touchSlug, finalReason),
		}
	}
}

func (m *Model) ensureSelectedVisible() {
	if len(m.projects) == 0 {
		return
	}

	top := m.viewport.YOffset
	bottom := top + m.viewport.Height

	if m.selected < top {
		m.viewport.YOffset = m.selected
	} else if m.selected >= bottom {
		m.viewport.YOffset = m.selected - m.viewport.Height + 1
	}
}

func (m *Model) updateViewportContent() {
	if len(m.projects) == 0 {
		m.viewport.SetContent("No projects found.")
		return
	}

	var lines []string
	for i, p := range m.projects {
		line := p.StatusDot + " " + p.Name + "  (" + fmt.Sprintf("%dd", p.DaysAgo) + " ago)"
		if i == m.selected {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Render(line)
		}
		lines = append(lines, line)
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// View implements tea.Model
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit"
	}
	if m.loading {
		return "🪄 Loading Ledger TUI...\n\nConnecting to Firestore..."
	}
	if !m.ready {
		return "Initializing..."
	}

	if m.helpMode {
		return m.renderHelpModal()
	}

	if m.touchMode {
		return m.renderTouchForm()
	}

	var s string

	switch m.currentTab {
	case 0:
		s = m.renderProjectsView()
	case 1:
		s = m.renderReviewView()
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(" q: quit • tab: switch view • ↑/↓ or j/k: navigate • r: refresh • t: touch • ?: help")

	return s + "\n\n" + footer
}

func (m Model) renderHelpModal() string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201")).Render("🪄 LEDGER TUI — HELP")

	helpText := `
Keyboard Shortcuts

  ↑ / k          Move selection up
  ↓ / j          Move selection down
  Tab / →        Next tab (Projects ↔ Review)
  Shift+Tab / ←  Previous tab

  t              Quick Touch (open touch form)
  r              Refresh project data from Firestore
  ?              Show this help
  q / Esc        Quit / Close modal

Touch Form (when open)

  Tab            Switch between Slug / Reason
  Enter          Submit touch
  Esc            Cancel touch form

Review Tab

  Shows three buckets + live sparkline of activity.
`

	return fmt.Sprintf("%s\n%s\n\n(Press ? or Esc to close)", header, helpText)
}

func (m Model) renderTouchForm() string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201")).Render("🪄 QUICK TOUCH")

	slugField := "Slug:   " + m.touchSlug
	reasonField := "Reason: " + m.touchReason

	if m.touchField == 0 {
		slugField = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(slugField)
	} else {
		reasonField = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(reasonField)
	}

	msg := ""
	if m.touchMsg != "" {
		msg = "\n" + m.touchMsg
	}

	return fmt.Sprintf("%s\n\n%s\n%s\n\n(tab: switch field • enter: touch • esc: cancel)%s",
		header, slugField, reasonField, msg)
}

func (m Model) renderProjectsView() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	header := headerStyle.Render("═══ PROJECTS ═══")

	if len(m.projects) == 0 {
		return header + "\n\nNo projects found."
	}

	m.updateViewportContent()
	return header + "\n\n" + m.viewport.View()
}

func (m Model) renderReviewView() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	header := headerStyle.Render("═══ WEEKLY REVIEW ═══")

	neglected := len(m.review.Neglected)
	missing := len(m.review.MissingNext)
	recent := len(m.review.RecentlyTouched)

	sparkValues := []int{neglected, missing, recent}
	spark := ui.Sparkline(sparkValues)

	return header + "\n\n" +
		"🔴 Neglected (≥10d):      " + fmt.Sprintf("%d", neglected) + "\n" +
		"⚠️  Missing Next Action:   " + fmt.Sprintf("%d", missing) + "\n" +
		"🟢 Recently Touched (≤3d): " + fmt.Sprintf("%d", recent) + "\n\n" +
		"Sparkline: " + spark + "  (Neglected → Recent)"
}
