package board

import (
	"fmt"

	"github.com/ascii-arcade/knucklebones/colors"
	"github.com/ascii-arcade/knucklebones/keys"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyScreen struct {
	model *Model
	style lipgloss.Style
}

func (m *Model) newLobbyScreen() *lobbyScreen {
	return &lobbyScreen{
		model: m,
		style: m.style,
	}
}

func (s *lobbyScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *lobbyScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case tea.KeyMsg:
		if keys.LobbyStartGame.TriggeredBy(msg.String()) {
			if s.model.game.GetPlayerData(s.model.player).IsHost {
				if err := s.model.game.Begin(); err != nil {
					s.model.error = s.model.lang().Get("error", err.Error())
					return s.model, nil
				}
			}
		}

	case messages.RefreshBoard:
		if s.model.game.InProgress {
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newTableScreen(),
				}
			}
		}
	}

	return s.model, nil
}

func (s *lobbyScreen) View() string {
	style := s.style.Width(s.model.width / 2)

	footer := s.model.lang().Get("board", "waiting_for_start")
	if s.model.game.GetPlayerData(s.model.player).IsHost {
		footer = fmt.Sprintf(s.model.lang().Get("board", "press_to_start"), keys.LobbyStartGame.String(s.style))

		if err := s.model.game.IsPlayerCountOk(); err != nil {
			errorMessage := s.model.lang().Get("error", err.Error())
			footer = s.style.Foreground(colors.Error).Render(errorMessage)
		}
	}
	footer += "\n"
	footer += fmt.Sprintf(s.model.lang().Get("global", "quit"), keys.ExitApplication.String(s.style))

	header := s.model.game.Code
	playerList := s.style.Render(s.playerList())

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		style.Align(lipgloss.Center).MarginBottom(2).Render(header),
		style.Render(playerList),
		style.Render(footer),
	)

	return s.style.Width(s.model.width).Height(s.model.height).Render(
		lipgloss.Place(
			s.model.width,
			s.model.height,
			lipgloss.Center,
			lipgloss.Center,
			s.style.
				Padding(2, 2).
				BorderStyle(lipgloss.NormalBorder()).
				Render(content),
		),
	)
}

func (s *lobbyScreen) playerList() string {
	playerList := ""
	for _, p := range s.model.game.GetPlayers() {
		playerList += "* " + p.Name
		if p.Name == s.model.player.Name {
			playerList += fmt.Sprintf(" (%s)", s.model.lang().Get("board", "player_list_you"))
		}
		if s.model.game.GetPlayerData(s.model.player).IsHost {
			playerList += fmt.Sprintf(" (%s)", s.model.lang().Get("board", "player_list_host"))
		}
		playerList += "\n"
	}
	return playerList
}
