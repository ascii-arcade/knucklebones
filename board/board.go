package board

import (
	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/games"
	"github.com/ascii-arcade/knucklebones/keys"
	"github.com/ascii-arcade/knucklebones/language"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int
	screen screen.Screen
	style  lipgloss.Style

	error string

	player *players.Player
	game   *games.Game
}

func NewModel(width, height int, style lipgloss.Style, player *players.Player) Model {
	m := Model{
		width:  width,
		height: height,
		style:  style,
		player: player,
	}

	m.screen = m.newLobbyScreen()
	return m
}

func (m *Model) SetGame(game *games.Game) {
	m.game = game
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForRefreshSignal(m.player.UpdateChan()),
		tea.WindowSize(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.player.SignalActivity()

		if keys.ExitApplication.TriggeredBy(msg.String()) {
			m.game.RemovePlayer(m.player)
			return m, tea.Quit
		}

	case messages.SwitchScreenMsg:
		m.screen = msg.Screen.WithModel(&m)
		return m, nil

	case messages.RefreshBoard:
		cmds = append(cmds, waitForRefreshSignal(m.player.UpdateChan()))
	}

	activeScreenModel, cmd := m.activeScreen().Update(msg)
	cmds = append(cmds, cmd)
	return activeScreenModel.(*Model), tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width < config.MinimumWidth {
		return m.lang().Get("error", "window_too_narrow")
	}
	if m.height < config.MinimumHeight {
		return m.lang().Get("error", "window_too_short")
	}

	return m.activeScreen().View()
}

func (m *Model) activeScreen() screen.Screen {
	return m.screen.WithModel(m)
}

func waitForRefreshSignal(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return messages.RefreshBoard(<-ch)
	}
}

func (m *Model) lang() *language.Language {
	return language.Languages[m.player.LanguagePreference]
}
