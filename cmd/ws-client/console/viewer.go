package console

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Content struct {
	Id     int
	Sender int
	Time   time.Time
	Text   string
}

type Viewer struct {
	Title        string
	vp           viewport.Model
	ta           textarea.Model
	prev         tea.Model
	focus        bool
	contents     []Content
	err          error
	loadFn       func(context.Context) ([]Content, error)
	cancel       context.CancelFunc
	initiallized bool
}

func newViewer() *Viewer {
	vp := viewport.New(0, 0)
	ta := textarea.New()

	v := Viewer{
		vp: vp,
		ta: ta,
	}
	return &v
}

func NewViewer(prev tea.Model, contents []Content) (*Viewer, tea.Cmd) {
	v := newViewer()
	v.prev = prev
	v.contents = contents

	return v, v.Init()
}

func NewViewerWithLoader(prev tea.Model, loadFn func(context.Context) ([]Content, error)) (*Viewer, tea.Cmd) {
	v := newViewer()
	v.loadFn = loadFn

	return v, v.Init()
}

func (v *Viewer) Init() tea.Cmd {
	if !v.initiallized {
		v.initiallized = true
		return v.reload(context.Background())
	}
	return nil
}

func (v *Viewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var vpCmd tea.Cmd
	var tiCmd tea.Cmd

	switch msg := msg.(type) {
	case dataMsg:
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		switch data := msg.data.(type) {
		case []Content:
			v.contents = data
			var s string
			for _, c := range v.contents {
				s += fmt.Sprintf("%s: %s", c.Time.Format(time.RFC3339), c.Text)
			}
			v.vp.SetContent(s)
			v.vp.GotoBottom()
			return v, nil
		}

	case tea.KeyMsg:
		if v.focus {
			switch msg.Type {
			case tea.KeyCtrlC:
				v.ta.Blur()
				v.focus = false
			}
		} else {
			switch msg.String() {
			case "r":
				ctx, cancel := context.WithCancel(context.Background())
				v.cancel = cancel
				return v, v.reload(ctx)

			case "i":
				v.ta.Focus()
				v.focus = true
				return v, nil

			case "q", "left":
				return v.prev, nil

			}
		}
	}

	v.vp, vpCmd = v.vp.Update(msg)
	v.ta, tiCmd = v.ta.Update(msg)

	return v, tea.Batch(vpCmd, tiCmd)
}

func (v *Viewer) View() string {
	if v.err != nil {
		return fmt.Sprintf("Error: %v", v.err)
	}
	if len(v.contents) != 0 {
	} else {
		v.ta.SetValue("No messages to display.")
	}
	return fmt.Sprintf("%s\n%s",
		v.vp.View(),
		v.ta.View(),
	)
}

func (v *Viewer) reload(ctx context.Context) tea.Cmd {
	if v.loadFn != nil {
		return func() tea.Msg {
			contents, err := v.loadFn(ctx)
			return dataMsg{
				data: contents,
				err:  err,
			}
		}
	}
	return nil
}

func OpenViewer(prev tea.Model, loadFn func(context.Context) ([]Content, error)) (tea.Model, tea.Cmd) {
	loader := func(ctx context.Context) (tea.Model, tea.Cmd) {
		contents, err := loadFn(ctx)
		if err != nil {
			// TODO: Handle error
			return nil, nil
		}
		return NewViewer(prev, contents)
	}
	return NewLoading(prev, loader)
}
