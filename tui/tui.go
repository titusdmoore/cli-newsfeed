package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	data "github.com/titusdmoore/cli-newsfeed/data"
)

type StateModel struct {
	Cursor int
	Items  []data.Item

	// This is just a debug field
	Message string
}
type bestItemsMsg []uint32
type itemsMsg []data.Item
type requestErrMsg struct{ err error }

func (e requestErrMsg) Error() string { return e.err.Error() }

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
	return StateModel{
		Cursor:  0,
		Items:   make([]data.Item, 0),
		Message: "Running Application...\n\n",
	}
}

// Init() runs automatically after the program is initialized.
// Fetches API for best itesm which returns an array of Item Ids
func (m StateModel) Init() tea.Cmd {
	return fetchBestItemsCmd
}

// Update handles events in the program. The entire function is a switch on event types.
func (m StateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case itemsMsg:
		m.Items = msg
		return m, nil
	case bestItemsMsg:
		m.Message = "Fetching best stories"
		return m, fetchItems(msg)
	case requestErrMsg:
		m.Message = msg.Error()
		return m, nil
	}

	return m, nil
}

// View is the render function of the program
// TODO: Implent actual rendering logic
func (m StateModel) View() string {
	if len(m.Items) > 0 {
		s := ""
		for i, item := range m.Items {
			if i > 50 {
				break
			}

			s = s + item.Title + "\n"
		}

		return s
	}

	return m.Message
}
