package console

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Loader func(ctx context.Context) (tea.Model, tea.Cmd)

type Loading struct {
	loadingText string
	isLoading   bool
	pos         int
	spinner     []rune
	loader      Loader
	prev        tea.Model
	cancel      context.CancelFunc
}

func NewLoading(prev tea.Model, loader Loader) (*Loading, tea.Cmd) {
	l := &Loading{
		loadingText: "Loading...",
		isLoading:   true,
		pos:         0,
		spinner:     []rune{'|', '/', '-', '\\'},
		loader:      loader,
		prev:        prev,
	}
	return l, l.Init()
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (l *Loading) Init() tea.Cmd {
	// start spinner + async loader
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	l.cancel = cancel

	return tea.Batch(
		tick(),
		l.load(ctx),
	)
}

type loadMsg struct {
	model tea.Model
	cmd   tea.Cmd
}

type tickMsg struct{}

func (l *Loading) load(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		model, cmd := l.loader(ctx)
		return loadMsg{
			model: model,
			cmd:   cmd,
		}
	}
}

func (l *Loading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadMsg:
		// loading finished
		l.isLoading = false
		if l.cancel != nil {
			l.cancel()
			l.cancel = nil
		}
		return msg.model, msg.cmd

	case tickMsg:
		if l.isLoading {
			l.pos = (l.pos + 1) % len(l.spinner)
			return l, tick()
		}
		return l, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// restart loading
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if l.cancel != nil {
				l.cancel()
			}
			l.cancel = cancel
			l.isLoading = true
			return l, tea.Batch(tick(), l.load(ctx))
		}
	}

	return l, nil
}

func (l *Loading) View() string {
	if l.isLoading {
		return l.loadingText + " " + string(l.spinner[l.pos])
	}
	return "done"
}

func Clear() {
	fmt.Printf("\033[H\033[2J")
}
