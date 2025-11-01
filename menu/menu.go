package menu

import (
	"errors"
	"time"

	"github.com/ascii-arcade/knucklebones/colors"
	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/games"
	"github.com/ascii-arcade/knucklebones/keys"
	"github.com/ascii-arcade/knucklebones/language"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/screen"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `++------------------------------------------------------------------------------++
++------------------------------------------------------------------------------++
||                                                                              ||
||                                                                              ||
||      _    ____   ____ ___ ___        _    ____   ____    _    ____  _____    ||
||     / \  / ___| / ___|_ _|_ _|      / \  |  _ \ / ___|  / \  |  _ \| ____|   ||
||    / _ \ \___ \| |    | | | |_____ / _ \ | |_) | |     / _ \ | | | |  _|     ||
||   / ___ \ ___) | |___ | | | |_____/ ___ \|  _ <| |___ / ___ \| |_| | |___    ||
||  /_/   \_\____/ \____|___|___|   /_/   \_\_| \_\\____/_/   \_\____/|_____|   ||
||                                                                              ||
||                                                                              ||
||                                                                              ||
++------------------------------------------------------------------------------++
++------------------------------------------------------------------------------++`

type doneMsg struct{}

type Model struct {
	width  int
	height int
	screen screen.Screen
	style  lipgloss.Style

	error         string
	gameCodeInput textinput.Model

	player *players.Player
}

func NewModel(width, height int, style lipgloss.Style, player *players.Player) Model {
	ti := textinput.New()
	ti.Width = 9
	ti.CharLimit = 7

	m := Model{
		width:  width,
		height: height,
		style:  style,

		gameCodeInput: ti,
		player:        player,
	}

	m.screen = m.newSplashScreen()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return doneMsg{}
		}),
		tea.WindowSize(),
		textinput.Blink,
	)
}

func (m *Model) lang() *language.Language {
	return language.Languages[m.player.LanguagePreference]
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchScreenMsg:
		m.screen = msg.Screen.WithModel(&m)
		return m, nil

	case tea.KeyMsg:
		if keys.ExitApplication.TriggeredBy(msg.String()) {
			return m, tea.Quit
		}
	}

	screenModel, cmd := m.screen.Update(msg)
	return screenModel.(*Model), cmd
}

func (m Model) View() string {
	if m.width < config.MinimumWidth {
		return m.lang().Get("error", "window_too_narrow")
	}
	if m.height < config.MinimumHeight {
		return m.lang().Get("error", "window_too_short")
	}

	style := m.style.Width(m.width).Height(m.height)
	paneStyle := m.style.Width(m.width).PaddingTop(1)

	panes := lipgloss.JoinVertical(
		lipgloss.Center,
		paneStyle.Align(lipgloss.Center, lipgloss.Bottom).Foreground(colors.Logo).Height(m.height/2).Render(logo),
		paneStyle.Align(lipgloss.Center, lipgloss.Top).Render(m.screen.View()),
	)

	return style.Render(panes)
}

func (m *Model) setError(err string) {
	m.error = err
}

func (m *Model) clearError() {
	m.error = ""
}

func (m *Model) joinGame(code string) error {
	game, err := games.GetOpenGame(code)
	if err != nil && !(errors.Is(err, games.ErrGameInProgress) && game.HasPlayer(m.player)) {
		return err
	}

	if err := game.AddPlayer(m.player); err != nil {
		return err
	}
	return nil
}
