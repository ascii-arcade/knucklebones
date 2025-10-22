package app

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/ascii-arcade/knucklebones/board"
	"github.com/ascii-arcade/knucklebones/dice"
	"github.com/ascii-arcade/knucklebones/menu"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/players"
)

type Model struct {
	active tea.Model
	menu   menu.Model
	board  board.Model
}

func (m Model) Init() tea.Cmd {
	return m.active.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchToBoardMsg:
		m.board.SetGame(msg.Game)
		m.active = m.board
		initcmd := m.board.Init()
		return m, initcmd

	case messages.SwitchToMenuMsg:
		m.active = m.menu
		initcmd := m.menu.Init()
		return m, initcmd
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.active.View()
}

func TeaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := sess.Pty()
	renderer := bubbletea.MakeRenderer(sess)
	style := renderer.NewStyle()

	dice.Setup(style)

	player, ok := sess.Context().Value("PLAYER").(*players.Player)
	if !ok {
		slog.Warn("That's weird")
		sess.Close()
		return nil, nil
	}

	m := Model{
		board: board.NewModel(pty.Window.Width, pty.Window.Height, style, player),
		menu:  menu.NewModel(pty.Window.Width, pty.Window.Height, style, player),
	}
	m.active = m.menu

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
