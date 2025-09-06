package console

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OpenFunc is a callback function type for opening a new model
type OpenFunc func(prev tea.Model) (tea.Model, tea.Cmd)

type ListItem struct {
	Name   string
	OpenFn OpenFunc
}

type List struct {
	Items  []ListItem
	cursor int
	// Prev is used for backward navigation
	Prev tea.Model

	// loadFn is an asynchronous worker function that loads the items.
	loadFn func(context.Context) ([]ListItem, error)

	// err is used to display error messages
	err    error
	cancel context.CancelFunc

	cursorColor lipgloss.Style
}

func (l List) Init() tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	l.cancel = cancel

	return l.reload(ctx)
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataMsg:
		if msg.err != nil {
			l.err = msg.err
			return l, nil
		}
		if items, ok := msg.data.([]ListItem); ok {
			l.Items = items
			l.cursor = 0
			return l, nil
		}
		// TODO: handle other types of data

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if l.cancel != nil {
				l.cancel()
			}
			ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
			l.cancel = cancel
			return l, l.reload(ctx)

		case "up", "k":
			l.up()
			return l, nil

		case "down", "j":
			l.down()
			return l, nil

		case "right", "enter":
			return l.open()
		case "left", "q", "esc":
			return l.goPrev()
		}
	}
	return l, nil
}

type dataMsg struct {
	data interface{}
	err  error
}

func (l List) reload(ctx context.Context) tea.Cmd {
	if l.loadFn != nil {
		return func() tea.Msg {
			items, err := l.loadFn(ctx)
			return dataMsg{
				data: items,
				err:  err,
			}
		}
	}
	return nil
}

func (l List) current() (ListItem, bool) {
	if len(l.Items) > 0 {
		return l.Items[l.cursor], true
	}
	return ListItem{}, false
}

func (l *List) down() (int, bool) {
	if l.cursor < len(l.Items)-1 {
		l.cursor++
		return l.cursor, true
	}

	return l.cursor, false
}

func (l *List) up() (int, bool) {
	if l.cursor > 0 {
		l.cursor--
		return l.cursor, true
	}

	return l.cursor, false
}

func (l List) open() (tea.Model, tea.Cmd) {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	if item, ok := l.current(); ok {
		if item.OpenFn != nil {
			return item.OpenFn(l)
		}
	}
	return l, nil
}

func (l List) goPrev() (tea.Model, tea.Cmd) {
	if l.Prev != nil {
		return l.Prev, nil
	}
	return l, tea.Quit
}

func (l List) View() string {
	if len(l.Items) == 0 {
		return "No items found."
	}

	var s string
	for i, item := range l.Items {
		var text string
		if i == l.cursor {
			text = l.cursorColor.Render(item.Name)
		} else {
			text = item.Name
		}

		s += text + "\n"
	}
	s += fmt.Sprintf(
		"\n[%d/%d] (↑/k ↓/j to move, q to quit)\n",
		l.cursor+1, len(l.Items),
	)

	return s
}

func NewList(prev tea.Model, loadFn func(context.Context) ([]ListItem, error)) (tea.Model, tea.Cmd) {
	l := List{
		loadFn:      loadFn,
		Items:       nil,
		Prev:        prev,
		cursor:      0,
		cursorColor: lipgloss.NewStyle().Background(lipgloss.Color("2")),
		err:         nil,
	}
	return l, l.Init()
}
