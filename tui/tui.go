package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	data "github.com/titusdmoore/cli-newsfeed/data"
)

type StateModel struct {
	// Model Fields for list view
	Cursor    int
	Items     []data.Item
	paginator paginator.Model

	// Model Fields for Item View
	SelectedItem *data.Item
	viewport     viewport.Model
	ready        bool

	// This is just a debug field
	Message string
}
type bestItemsMsg []uint32
type itemsMsg []data.Item
type itemSelectedMessage data.Item
type requestErrMsg struct{ err error }

func (e requestErrMsg) Error() string { return e.err.Error() }

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

// Command that handles the initial fetch of the best rated items
func fetchBestItemsCmd() tea.Msg {
	var ids []uint32
	err := data.FetchResource("beststories.json", &ids)
	if err != nil {
		return requestErrMsg{err}
	}

	return bestItemsMsg(ids)
}

// fetchItems is a command function wrapper. Commands in Bubbletea cannot take args
// so we return a function that is the actual command with the arguments scoped.
// Docs here: https://github.com/charmbracelet/bubbletea/tree/main/tutorials/commands/#one-more-thing-about-commands
func fetchItems(item_ids []uint32) tea.Cmd {
	return func() tea.Msg {
		items, err := data.FetchItems(item_ids)
		if err != nil {
			return requestErrMsg{err}
		}

		return itemsMsg(items)
	}
}

// This function just runs when we initialize the program, if we want to set data, use Init()
func GenerateInitialStateModel() StateModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 20
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	return StateModel{
		Cursor:    0,
		Items:     make([]data.Item, 0),
		Message:   "Running Application...\n\n",
		paginator: p,
	}
}

// Init() runs automatically after the program is initialized.
// Fetches API for best itesm which returns an array of Item Ids
func (m StateModel) Init() tea.Cmd {
	return fetchBestItemsCmd
}

// Update handles events in the program. The entire function is a switch on event types.
func (m StateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor += 1
			}

			return m, nil
		case "k":
			if m.Cursor > 0 {
				m.Cursor -= 1
			}

			return m, nil
		case "enter":
			m.SelectedItem = &m.Items[m.Cursor]

			// Once we select an item we need to query the window size which initializes the viewport
			return m, tea.WindowSize()
		case "esc":
			m.SelectedItem = nil

			return m, nil
		}
	case tea.WindowSizeMsg:
		// We don't really care (as of v1.0) about resizing unless we have a SelectedItem
		if m.SelectedItem == nil {
			return m, nil
		}

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.SelectedItem.Text)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	case itemsMsg:
		m.Items = msg
		m.paginator.SetTotalPages(len(m.Items))
		return m, nil
	case bestItemsMsg:
		m.Message = "Fetching best stories"
		return m, fetchItems(msg)
	case requestErrMsg:
		m.Message = msg.Error()
		return m, nil
	case itemSelectedMessage:
	}

	m.paginator, cmd = m.paginator.Update(msg)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m StateModel) headerView() string {
	title := titleStyle.Render(m.SelectedItem.Title)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m StateModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func basicRender(m StateModel) string {
	if len(m.Items) > 0 {
		var builder strings.Builder

		start, end := m.paginator.GetSliceBounds(len(m.Items))
		for i, item := range m.Items[start:end] {
			if i == m.Cursor {
				builder.WriteString("> ")
			} else {
				builder.WriteString("  ")
			}

			builder.WriteString(item.Title + "\n")
		}
		builder.WriteString("  " + m.paginator.View())
		builder.WriteString("\n\n  h/l ←/→ page • q: quit\n")

		return builder.String()
	}

	return m.Message
}

func itemRender(m StateModel) string {
	if !m.ready {
		return "\n Initializing reader..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

// View is the render function of the program
// TODO: Implent actual rendering logic
func (m StateModel) View() string {
	if m.SelectedItem != nil {
		return itemRender(m)
	}

	// Currently just set to basic render to enable the ability to include a second view for single articles in the future
	return basicRender(m)
}
