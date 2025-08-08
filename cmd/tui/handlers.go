package main

import (
	"fmt"

	"curltree/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *tuiModel) handleProfileViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+e":
		m.state = models.StateProfileEdit
		m.form = newFormModel()
		if m.user != nil {
			m.form.populateFromUser(m.user)
		}
		return m, nil
	case "ctrl+d":
		m.state = models.StateConfirmDelete
		return m, nil
	}
	return m, nil
}

func (m *tuiModel) handleProfileEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = models.StateProfileView
		return m, nil
	case "ctrl+s":
		return m.saveProfile()
	case "tab":
		m.form.nextField()
		return m, nil
	case "shift+tab":
		m.form.prevField()
		return m, nil
	case "ctrl+n":
		m.form.addLink("", "")
		return m, nil
	case "ctrl+d":
		m.form.deleteCurrentLink()
		return m, nil
	default:
		// Pass the message to the form for text input handling
		m.form.Update(msg)
		return m, nil
	}
}

func (m *tuiModel) handleProfileCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, tea.Quit
	case "ctrl+s":
		return m.createProfile()
	case "tab":
		m.form.nextField()
		return m, nil
	case "shift+tab":
		m.form.prevField()
		return m, nil
	case "ctrl+n":
		m.form.addLink("", "")
		return m, nil
	case "ctrl+d":
		m.form.deleteCurrentLink()
		return m, nil
	default:
		// Pass the message to the form for text input handling
		m.form.Update(msg)
		return m, nil
	}
}

func (m *tuiModel) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.deleteProfile()
	case "n", "N", "esc":
		m.state = models.StateProfileView
		return m, nil
	}
	return m, nil
}

func (m *tuiModel) createProfile() (tea.Model, tea.Cmd) {
	if err := m.form.validate(); err != nil {
		return m, func() tea.Msg { return errorMsg{err} }
	}

	req := m.form.toCreateRequest(m.sshKey)

	return m, func() tea.Msg {
		exists, err := m.db.IsUsernameExists(req.Username)
		if err != nil {
			return errorMsg{err}
		}
		if exists {
			return errorMsg{fmt.Errorf("Username '%s' already exists", req.Username)}
		}

		user, err := m.db.CreateUser(req)
		if err != nil {
			return errorMsg{err}
		}

		return profileCreatedMsg{user}
	}
}

func (m *tuiModel) saveProfile() (tea.Model, tea.Cmd) {
	if m.user == nil {
		return m, func() tea.Msg { return errorMsg{fmt.Errorf("No user to save")} }
	}

	if err := m.form.validate(); err != nil {
		return m, func() tea.Msg { return errorMsg{err} }
	}

	req := m.form.toUpdateRequest()
	userID := m.user.ID
	currentUsername := m.user.Username

	return m, func() tea.Msg {
		if req.Username != currentUsername {
			exists, err := m.db.IsUsernameExists(req.Username)
			if err != nil {
				return errorMsg{err}
			}
			if exists {
				return errorMsg{fmt.Errorf("Username '%s' already exists", req.Username)}
			}
		}

		user, err := m.db.UpdateUser(userID, req)
		if err != nil {
			return errorMsg{err}
		}

		return profileUpdatedMsg{user}
	}
}

func (m *tuiModel) deleteProfile() (tea.Model, tea.Cmd) {
	if m.user == nil {
		return m, func() tea.Msg { return errorMsg{fmt.Errorf("No user to delete")} }
	}

	userID := m.user.ID

	return m, func() tea.Msg {
		if err := m.db.DeleteUser(userID); err != nil {
			return errorMsg{err}
		}
		return tea.Quit()
	}
}

