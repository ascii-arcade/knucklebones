package menu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ascii-arcade/knucklebones/keys"
	"github.com/ascii-arcade/knucklebones/messages"
	"github.com/ascii-arcade/knucklebones/players"
	"github.com/ascii-arcade/knucklebones/screen"
	"github.com/ascii-arcade/knucklebones/utils"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants for better maintainability
const (
	cursorSSHKeyName = iota
	cursorSSHKey
	maxCursorPos = 1

	minSSHKeys       = 1
	maxUsernameLen   = 20
	maxSSHKeyNameLen = 20
	maxSSHKeyLen     = 600
	maxPlayerNameLen = 25
	sshKeyWidth      = 50
	tableHeight      = 7

	cursorColor     = "205"
	errorColor      = "1"
	borderColor     = "240"
	selectedBgColor = "57"
	selectedFgColor = "229"
)

// screenMode represents the current mode of the edit profile screen
type screenMode int

const (
	modeEditProfile screenMode = iota
	modeEditPlayerName
	modeAddSSHKey
	modeManageKeys
	modeMergePlayers
	modeMergeConfirm
	modeAddGithubSSHKey
	modeAddGithubSSHKeyConfirm
)

type editProfileScreen struct {
	model *Model
	style lipgloss.Style

	userNameInput       textinput.Model
	sshKeyNameInput     textinput.Model
	sshKeyInput         textinput.Model
	mergePlayersInput   textinput.Model
	githubUsernameInput textinput.Model

	cursorPos   int
	currentMode screenMode

	sshKeysTable table.Model
}

func (m *Model) newEditProfileScreen() *editProfileScreen {
	s := &editProfileScreen{
		model:       m,
		style:       m.style,
		currentMode: modeEditProfile,
	}

	userNameInput := textinput.New()
	userNameInput.Cursor.Style = m.style.Foreground(lipgloss.Color(cursorColor))
	userNameInput.Placeholder = ""
	userNameInput.CharLimit = maxUsernameLen
	userNameInput.Focus()
	userNameInput.SetValue(m.player.Username)
	s.userNameInput = userNameInput

	sshKeyNameInput := textinput.New()
	sshKeyNameInput.Cursor.Style = m.style.Foreground(lipgloss.Color(cursorColor))
	sshKeyNameInput.CharLimit = maxSSHKeyNameLen
	sshKeyNameInput.Blur()
	s.sshKeyNameInput = sshKeyNameInput

	sshKeyInput := textinput.New()
	sshKeyInput.Cursor.Style = m.style.Foreground(lipgloss.Color(cursorColor))
	sshKeyInput.Placeholder = ""
	sshKeyInput.CharLimit = maxSSHKeyLen
	sshKeyInput.Width = sshKeyWidth
	sshKeyInput.Blur()
	s.sshKeyInput = sshKeyInput

	mergePlayersInput := textinput.New()
	mergePlayersInput.Cursor.Style = m.style.Foreground(lipgloss.Color(cursorColor))
	mergePlayersInput.Placeholder = "Enter player name to merge"
	mergePlayersInput.CharLimit = maxPlayerNameLen
	mergePlayersInput.Blur()
	s.mergePlayersInput = mergePlayersInput

	githubUsernameInput := textinput.New()
	githubUsernameInput.Cursor.Style = m.style.Foreground(lipgloss.Color(cursorColor))
	githubUsernameInput.Placeholder = "Enter your GitHub username"
	githubUsernameInput.CharLimit = maxPlayerNameLen
	githubUsernameInput.Blur()
	s.githubUsernameInput = githubUsernameInput

	s.initializeSSHKeysTable()

	return s
}

func (s *editProfileScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case cursor.BlinkMsg:
		return s.model, nil

	case tea.KeyMsg:
		if keys.PreviousScreen.TriggeredBy(msg.String()) {
			if s.currentMode == modeEditProfile {
				if err := s.model.player.Save(); err != nil {
					s.model.error = err.Error()
					return s.model, nil
				}

				return s.model, func() tea.Msg {
					return messages.SwitchScreenMsg{
						Screen: s.model.newOptionScreen(),
					}
				}
			}

			s.currentMode = modeEditProfile
			return s.model, nil
		}
	}

	switch s.currentMode {
	case modeEditPlayerName:
		return s.handleEditPlayerNameInput(msg)
	case modeAddSSHKey:
		return s.handleAddSSHKeyInput(msg)
	case modeManageKeys:
		return s.handleManageKeysInput(msg)
	case modeMergePlayers:
		return s.handleMergePlayersInput(msg)
	case modeAddGithubSSHKey:
		return s.handleAddGithubSSHKeyInput(msg)
	case modeMergeConfirm, modeAddGithubSSHKeyConfirm:
		return s.handleConfirmInput(msg)
	}

	return s.handleNormalInput(msg)
}

func (s *editProfileScreen) handleEditPlayerNameInput(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Submit.TriggeredBy(msg.String()) {
			s.model.player.UpdateUsername(s.userNameInput.Value())
			s.userNameInput.SetValue("")
			s.currentMode = modeEditProfile
		}
	}

	var cmd tea.Cmd
	s.userNameInput, cmd = s.userNameInput.Update(msg)
	return s.model, cmd
}

func (s *editProfileScreen) WithModel(model any) screen.Screen {
	s.model = model.(*Model)
	return s
}

func (s *editProfileScreen) initializeSSHKeysTable() {
	columns := []table.Column{
		{Title: "", Width: 10},
		{Title: "Public SSH Keys", Width: 40},
	}

	rows := s.buildSSHKeyRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = s.style.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = s.style.
		Foreground(lipgloss.Color(selectedFgColor)).
		Background(lipgloss.Color(selectedBgColor)).
		Bold(false)
	t.SetStyles(tableStyle)

	s.sshKeysTable = t
}

func (s *editProfileScreen) buildSSHKeyRows() []table.Row {
	rows := make([]table.Row, 0, len(s.model.player.SshPubKeys))
	for name, key := range s.model.player.SshPubKeys {
		rows = append(rows, table.Row{name, key})
	}
	if len(rows) == 0 {
		rows = append(rows, table.Row{"-", "No SSH Keys Added"})
	}
	return rows
}

func (s *editProfileScreen) refreshSSHKeysTable() {
	rows := s.buildSSHKeyRows()
	s.sshKeysTable.SetRows(rows)
}

func (s *editProfileScreen) updateCursorPosition() {
	switch {
	case s.cursorPos > maxCursorPos:
		s.cursorPos = cursorSSHKeyName
	case s.cursorPos < cursorSSHKeyName:
		s.cursorPos = maxCursorPos
	}
	s.setInputFocus()
}

func (s *editProfileScreen) setInputFocus() {
	s.sshKeyNameInput.Blur()
	s.sshKeyInput.Blur()

	switch s.cursorPos {
	case cursorSSHKeyName:
		s.sshKeyNameInput.Focus()
	case cursorSSHKey:
		s.sshKeyInput.Focus()
	}
}

func (s *editProfileScreen) clearSSHKeyInputs() {
	s.sshKeyNameInput.SetValue("")
	s.sshKeyNameInput.Blur()
	s.sshKeyInput.SetValue("")
	s.sshKeyInput.Blur()
	s.userNameInput.Focus()
	s.cursorPos = cursorSSHKeyName
}

func (s *editProfileScreen) validateAndSaveSSHKey() error {
	if s.sshKeyNameInput.Value() == "" || s.sshKeyInput.Value() == "" {
		return &ValidationError{s.model.lang().Get("error", "edit_profile", "name_and_key_required")}
	}

	if !utils.ValidPublicKey(s.sshKeyInput.Value()) {
		return &ValidationError{s.model.lang().Get("error", "edit_profile", "invalid_key")}
	}

	if _, exists := s.model.player.SshPubKeys[s.sshKeyNameInput.Value()]; exists {
		return &ValidationError{s.model.lang().Get("error", "edit_profile", "key_name_exists")}
	}

	key := strings.Join(strings.Split(s.sshKeyInput.Value(), " ")[:2], " ")
	if _, found := players.Get(key); found {
		return &ValidationError{s.model.lang().Get("error", "edit_profile", "key_in_use")}
	}

	s.model.player.SshPubKeys[s.sshKeyNameInput.Value()] = s.sshKeyInput.Value()
	return s.model.player.Save()
}

func (s *editProfileScreen) parsePlayerName(playerName string) (username, discriminator string, err error) {
	playerName = strings.TrimSpace(playerName)
	if playerName == "" {
		return "", "", &ValidationError{s.model.lang().Get("error", "edit_profile", "player_name_required")}
	}

	playerNameSplit := strings.SplitN(playerName, "#", 2)
	if len(playerNameSplit) != 2 {
		return "", "", &ValidationError{s.model.lang().Get("error", "edit_profile", "player_name_format_required")}
	}

	return playerNameSplit[0], playerNameSplit[1], nil
}

func (s *editProfileScreen) performPlayerMerge(username, discriminator string) error {
	mergePlayer, found := players.GetByName(username, discriminator)
	if !found {
		return &ValidationError{s.model.lang().Get("error", "player_not_found")}
	}

	if mergePlayer.Id == s.model.player.Id {
		return &ValidationError{s.model.lang().Get("error", "edit_profile", "same_player_merge")}
	}

	err := players.Merge(s.model.player, mergePlayer)
	if err != nil {
		return err
	}

	players.RemovePlayer(mergePlayer)
	return nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (s *editProfileScreen) handleDeleteSSHKey() (any, tea.Cmd) {
	if len(s.model.player.SshPubKeys) <= minSSHKeys {
		s.model.error = s.model.lang().Get("error", "edit_profile", "min_ssh_keys_required")
		return s.model, nil
	}

	keyName := s.sshKeysTable.SelectedRow()[0]
	delete(s.model.player.SshPubKeys, keyName)

	if err := s.model.player.Save(); err != nil {
		s.model.error = err.Error()
		return s.model, nil
	}

	s.refreshSSHKeysTable()
	_, cmd := s.sshKeysTable.Update(tea.KeyMsg{Type: tea.KeyDelete})
	return s.model, cmd
}

func (s *editProfileScreen) handleManageKeysInput(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Up.TriggeredBy(msg.String()) {
			s.sshKeysTable.SetCursor(s.sshKeysTable.Cursor() - 1)
			return s.model, nil
		}

		if keys.Down.TriggeredBy(msg.String()) {
			s.sshKeysTable.SetCursor(s.sshKeysTable.Cursor() + 1)
			return s.model, nil
		}

		switch msg.String() {
		case "d":
			return s.handleDeleteSSHKey()
		}
	}

	s.sshKeysTable, cmd = s.sshKeysTable.Update(msg)
	return s.model, cmd
}

func (s *editProfileScreen) handleAddSSHKeyInput(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Up.TriggeredBy(msg.String()) {
			s.cursorPos--
			s.updateCursorPosition()
			return s.model, nil
		}

		if keys.Down.TriggeredBy(msg.String()) {
			s.cursorPos++
			s.updateCursorPosition()
			return s.model, nil
		}

		if keys.Submit.TriggeredBy(msg.String()) {
			err := s.validateAndSaveSSHKey()
			if err != nil {
				s.model.error = err.Error()
				return s.model, nil
			}
			s.clearSSHKeyInputs()
			s.refreshSSHKeysTable()

			s.currentMode = modeEditProfile
		}
	}

	switch s.cursorPos {
	case cursorSSHKeyName:
		s.sshKeyNameInput, cmd = s.sshKeyNameInput.Update(msg)
	case cursorSSHKey:
		s.sshKeyInput, cmd = s.sshKeyInput.Update(msg)
	}

	return s.model, cmd
}

func (s *editProfileScreen) handleAddGithubSSHKeyInput(msg tea.Msg) (any, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Submit.TriggeredBy(msg.String()) {
			s.currentMode = modeAddGithubSSHKeyConfirm
			return s.model, nil
		}
	}

	s.githubUsernameInput, cmd = s.githubUsernameInput.Update(msg)
	return s.model, cmd
}

func (s *editProfileScreen) handleNormalInput(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			s.userNameInput.Focus()
			s.currentMode = modeEditPlayerName
		case "a":
			s.sshKeyNameInput.Focus()
			s.currentMode = modeAddSSHKey
		case "k":
			s.currentMode = modeManageKeys
		case "g":
			s.currentMode = modeAddGithubSSHKey
			s.githubUsernameInput.Focus()
		}
	}

	s.model.error = ""
	return s.model, nil
}

func (s *editProfileScreen) renderEditProfileView() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "edit_profile", "set_player_name")+"\n", "n"))
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "edit_profile", "add_ssh_key")+"\n", "a"))
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "edit_profile", "add_ssh_key_github")+"\n", "g"))
	content.WriteString(fmt.Sprintf(s.model.lang().Get("menu", "edit_profile", "manage_ssh_keys"), "k"))

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}

func (s *editProfileScreen) View() string {
	var out string

	switch s.currentMode {
	case modeEditPlayerName:
		out = s.renderEditPlayerNameView()
	case modeAddSSHKey:
		out = s.renderAddSSHKeyView()
	case modeManageKeys:
		out = s.renderManageKeysView()
	case modeMergePlayers:
		out = s.renderMergePlayersView()
	case modeMergeConfirm:
		out = s.renderMergeConfirmView()
	case modeAddGithubSSHKey:
		out = s.renderAddGithubSSHKeyView()
	case modeAddGithubSSHKeyConfirm:
		out = s.renderMergeConfirmView()
	default:
		out = s.renderEditProfileView()
	}

	out += "\n\n" + fmt.Sprintf(s.model.lang().Get("global", "go_back"), keys.PreviousScreen.String(s.style))

	if s.model.error != "" {
		out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(s.model.error)
	}
	return out
}

func (s *editProfileScreen) renderAddGithubSSHKeyView() string {
	var content strings.Builder

	content.WriteString("Github Username | " + s.githubUsernameInput.View() + "\n")

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}

func (s *editProfileScreen) renderAddSSHKeyView() string {
	var content strings.Builder

	keyName := s.model.lang().Get("menu", "edit_profile", "key_name")
	publicKey := s.model.lang().Get("menu", "edit_profile", "public_key")
	largest := max(len(keyName), len(publicKey))
	smallest := min(len(keyName), len(publicKey))
	padding := largest - smallest
	if padding > 0 {
		keyName += strings.Repeat(" ", padding)
	} else {
		publicKey += strings.Repeat(" ", -padding)
	}

	content.WriteString(keyName + " | " + s.sshKeyNameInput.View())
	content.WriteString("\n")
	content.WriteString(publicKey + " | " + s.sshKeyInput.View())

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}

func (s *editProfileScreen) handleConfirmInput(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Confirm.TriggeredBy(msg.String()) {
			switch s.currentMode {
			case modeMergeConfirm:
				username, discriminator, err := s.parsePlayerName(s.mergePlayersInput.Value())
				if err != nil {
					s.model.error = err.Error()
					return s.model, nil
				}

				err = s.performPlayerMerge(username, discriminator)
				if err != nil {
					s.model.error = err.Error()
					return s.model, nil
				}

				s.currentMode = modeEditProfile
			case modeAddGithubSSHKeyConfirm:
				githubUsername := s.githubUsernameInput.Value()

				publicKeys, err := utils.GetPublicSSHKeys(githubUsername)
				if err != nil {
					s.model.error = err.Error()
					s.currentMode = modeAddGithubSSHKey
					return s.model, nil
				}

				for i, key := range publicKeys {
					s.model.player.AddPubKey("from-github-"+strconv.Itoa(i), key)
				}

				s.refreshSSHKeysTable()
			}
		}
	}

	s.currentMode = modeEditProfile
	return s.model, nil
}

func (s *editProfileScreen) handleMergePlayersInput(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keys.Submit.TriggeredBy(msg.String()) {
			s.currentMode = modeMergeConfirm
			return s.model, nil
		}
	}

	var cmd tea.Cmd
	s.mergePlayersInput, cmd = s.mergePlayersInput.Update(msg)
	return s.model, cmd
}

func (s *editProfileScreen) renderManageKeysView() string {
	var content strings.Builder

	content.WriteString(s.sshKeysTable.View())
	if s.model.error != "" {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(s.model.error))
		s.model.error = ""
	}

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}

func (s *editProfileScreen) renderMergeConfirmView() string {
	style := lipgloss.NewStyle().Align(lipgloss.Center).Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(borderColor)).Width(50)
	content := strings.Builder{}

	switch s.currentMode {
	case modeAddGithubSSHKeyConfirm:
		content.WriteString(fmt.Sprintf("Are you sure you want to add all public SSH keys from the GitHub user '%s' to your account?\n\n", s.githubUsernameInput.Value()))
	case modeMergeConfirm:
		content.WriteString("Are you sure you want to merge the player accounts? This action cannot be undone.\n\n")
	}

	content.WriteString("Press 'y' to confirm or 'n' to cancel.")
	return style.Render(content.String())
}

func (s *editProfileScreen) renderMergePlayersView() string {
	var content strings.Builder

	content.WriteString("Merge player accounts")
	if s.model.error != "" {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(s.model.error))
		s.model.error = ""
	}
	content.WriteString("\n\n")
	content.WriteString("Enter player name to merge | ")
	content.WriteString(s.mergePlayersInput.View())
	content.WriteString("\n\nPress 'esc' to return to profile editing")

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}

func (s *editProfileScreen) renderEditPlayerNameView() string {
	var content strings.Builder

	content.WriteString("Player Name | " + s.userNameInput.View())

	style := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Left).Width(lipgloss.Width(content.String()))
	return style.Render(content.String())
}
