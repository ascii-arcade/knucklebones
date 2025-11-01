package keys

import (
	"slices"

	"github.com/charmbracelet/lipgloss"
)

type Keys []string

var (
	MenuJoinGame     = Keys{"j"}
	MenuStartNewGame = Keys{"n"}
	MenuEnglish      = Keys{"1"}
	MenuSpanish      = Keys{"2"}
	MenuEditProfile  = Keys{"p"}

	Back    = Keys{"q"}
	Confirm = Keys{"y", "Y", "enter"}

	PreviousScreen = Keys{"esc"}
	Submit         = Keys{"enter"}
	Up             = Keys{"up", "shift+tab"}
	Down           = Keys{"down", "tab"}

	ExitApplication = Keys{"ctrl+c"}
	LobbyStartGame  = Keys{"s"}

	ActionRoll         = Keys{"r"}
	ActionRestart      = Keys{"r"}
	ActionSelectColumn = Keys{"1", "2", "3"}
)

func (k Keys) TriggeredBy(msg string) bool {
	return slices.Contains(k, msg)
}

func (k Keys) String(style lipgloss.Style) string {
	return k.IndexedString(0, style)
}

func (k Keys) IndexedString(index int, style lipgloss.Style) string {
	if len(k) == 0 {
		return ""
	}
	return style.Bold(true).Italic(true).Render("'" + k[index] + "'")
}
