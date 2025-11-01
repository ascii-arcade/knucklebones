package menu

import (
	"fmt"
	"strings"

	"github.com/ascii-arcade/knucklebones/config"
	"github.com/ascii-arcade/knucklebones/games"
	"github.com/ascii-arcade/knucklebones/keys"
	"github.com/ascii-arcade/knucklebones/language"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionScreen struct {
	model *Model
	style lipgloss.Style
}

func (m *Model) newOptionScreen() *optionScreen {
	return &optionScreen{
		model: m,
		style: m.style,
	}
}

func (s *optionScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *optionScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case tea.KeyMsg:
		switch {
		case keys.MenuEnglish.TriggeredBy(msg.String()):
			s.model.player.LanguagePreference = language.LanguageEN
		case keys.MenuSpanish.TriggeredBy(msg.String()):
			s.model.player.LanguagePreference = language.LanguageES
		case keys.MenuEditProfile.TriggeredBy(msg.String()):
			if s.model.player.Visitor {
				return s.model, nil
			}
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newEditProfileScreen(),
				}
			}
		case keys.MenuStartNewGame.TriggeredBy(msg.String()):
			newGame := games.New()
			if err := s.model.joinGame(newGame.Code); err != nil {
				s.model.setError(err.Error())
				return s.model, nil
			}

			return s.model, func() tea.Msg { return messages.SwitchToBoardMsg{Game: newGame} }
		case keys.MenuJoinGame.TriggeredBy(msg.String()):
			return s.model, func() tea.Msg {
				return messages.SwitchScreenMsg{
					Screen: s.model.newJoinScreen(),
				}
			}
		}
	}

	return s.model, nil
}

func (s *optionScreen) View() string {
	var content strings.Builder
	content.WriteString(s.model.lang().Get("menu", "welcome") + "\n\n")
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_create"), keys.MenuStartNewGame.String(s.style)) + "\n")
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_join"), keys.MenuJoinGame.String(s.style)) + "\n")
	if !s.model.player.Visitor {
		content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "press_to_edit_profile"), keys.MenuEditProfile.String(s.style)) + "\n")
	}
	content.WriteString("\n\n")

	if s.model.lang() == language.Languages[language.LanguageEN] {
		content.WriteString(fmt.Sprintf(language.Languages[language.LanguageES].Get("menu", "choose_language"), keys.MenuSpanish.String(s.style)))
	} else if s.model.lang() == language.Languages[language.LanguageES] {
		content.WriteString(fmt.Sprintf(language.Languages[language.LanguageEN].Get("menu", "choose_language"), keys.MenuEnglish.String(s.style)))
	}

	content.WriteString("\n\n")
	content.WriteString(s.style.Foreground(lipgloss.Color("#555555")).Render(config.Version))

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))

	return style.Render(content.String())
}
