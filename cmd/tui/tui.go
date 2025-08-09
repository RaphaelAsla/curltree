package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"curltree/internal/database"
	"curltree/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	asciiStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			MarginBottom(1)
)

func getASCIIArt() string {
	return `
 ██████╗██╗   ██╗██████╗ ██╗     ████████╗██████╗ ███████╗███████╗
██╔════╝██║   ██║██╔══██╗██║     ╚══██╔══╝██╔══██╗██╔════╝██╔════╝
██║     ██║   ██║██████╔╝██║        ██║   ██████╔╝█████╗  █████╗  
██║     ██║   ██║██╔══██╗██║        ██║   ██╔══██╗██╔══╝  ██╔══╝  
╚██████╗╚██████╔╝██║  ██║███████╗   ██║   ██║  ██║███████╗███████╗
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚══════╝`
}

func newTUIModel(s ssh.Session, db *database.DB) tea.Model {
	publicKey := s.PublicKey()
	var sshKey string

	if publicKey == nil {
		return &tuiModel{
			session: s,
			db:      db,
			state:   models.StateError,
			err:     fmt.Errorf("No SSH public key found - please ensure you're connecting with a valid SSH key"),
		}
	}

	keyBytes := publicKey.Marshal()
	hash := sha256.Sum256(keyBytes)
	sshKey = fmt.Sprintf("%s:%s", publicKey.Type(), hex.EncodeToString(hash[:]))

	// Debug: Log the SSH key being processed
	fmt.Printf("DEBUG: Processing SSH key: %s\n", sshKey)
	fmt.Printf("DEBUG: Key type: %s, Key length: %d bytes\n", publicKey.Type(), len(keyBytes))

	user, err := db.GetUserBySSHKey(sshKey)
	if err != nil {
		fmt.Printf("DEBUG: Error looking up user by SSH key: %v\n", err)
	} else if user != nil {
		fmt.Printf("DEBUG: Found existing user: %s (%s)\n", user.Username, user.FullName)
	} else {
		fmt.Printf("DEBUG: No existing user found for this SSH key\n")
	}

	var state models.AppState
	if user != nil {
		state = models.StateProfileView
	} else {
		state = models.StateProfileCreate
	}

	return &tuiModel{
		session: s,
		db:      db,
		user:    user,
		sshKey:  sshKey,
		state:   state,
		form:    newFormModel(),
	}
}

type tuiModel struct {
	session ssh.Session
	db      *database.DB
	user    *models.User
	sshKey  string
	state   models.AppState
	form    *formModel
	width   int
	height  int
	message string
	err     error
}

func (m *tuiModel) Init() tea.Cmd {
	return nil
}

func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.state == models.StateProfileView || m.state == models.StateProfileCreate || m.state == models.StateError {
				return m, tea.Quit
			}
		}
		return m.handleKeyPress(msg)

	case profileCreatedMsg:
		m.user = msg.user
		m.state = models.StateProfileView
		m.message = "Profile created successfully!"
		return m, nil

	case profileUpdatedMsg:
		m.user = msg.user
		m.state = models.StateProfileView
		m.message = "Profile updated successfully!"
		return m, nil

	case errorMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

type profileCreatedMsg struct {
	user *models.User
}

type profileUpdatedMsg struct {
	user *models.User
}

type errorMsg struct {
	err error
}

func (m *tuiModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case models.StateProfileView:
		return m.handleProfileViewKeys(msg)
	case models.StateProfileEdit:
		return m.handleProfileEditKeys(msg)
	case models.StateProfileCreate:
		return m.handleProfileCreateKeys(msg)
	case models.StateConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	}
	return m, nil
}

func (m *tuiModel) View() string {
	switch m.state {
	case models.StateLoading:
		return m.loadingView()
	case models.StateError:
		return m.errorView()
	case models.StateProfileView:
		return m.profileView()
	case models.StateProfileEdit:
		return m.editView()
	case models.StateProfileCreate:
		return m.createView()
	case models.StateConfirmDelete:
		return m.confirmDeleteView()
	}
	return ""
}

func (m *tuiModel) loadingView() string {
	return "Loading..."
}

func (m *tuiModel) errorView() string {
	content := asciiStyle.Render(getASCIIArt()) + "\n\n"
	content += errorStyle.Render("Authentication Error") + "\n\n"

	if m.err != nil {
		content += fmt.Sprintf("Error: %v\n\n", m.err)
	}

	content += "Please ensure you're connecting with a valid SSH key.\n"
	content += "Example: ssh -i ~/.ssh/id_rsa user@curltree.dev\n\n"

	help := helpStyle.Render("ctrl+c: exit")
	return content + help
}

func (m *tuiModel) profileView() string {
	if m.user == nil {
		return errorStyle.Render("No user data available")
	}

	content := asciiStyle.Render(getASCIIArt()) + "\n\n"

	// Use exact same format as curl output
	content += fmt.Sprintf("┌─ %s (@%s)\n", m.user.FullName, m.user.Username)
	content += "│\n"

	// About section
	if m.user.About != "" {
		content += "├─ About:\n"
		content += "│  ├─ "

		// Split about text into words for proper wrapping (same as backend)
		words := strings.Fields(m.user.About)
		currentLine := ""
		linePrefix := "│     "
		maxLineLength := 60

		for i, word := range words {
			testLine := currentLine + word
			if i > 0 {
				testLine = currentLine + " " + word
			}

			if len(testLine) > maxLineLength && currentLine != "" {
				content += fmt.Sprintf("%s\n%s", currentLine, linePrefix)
				currentLine = word
			} else {
				if i > 0 {
					currentLine += " "
				}
				currentLine += word
			}
		}

		if currentLine != "" {
			content += fmt.Sprintf("%s\n", currentLine)
		}
		content += "│\n"
	}

	// Links section
	if len(m.user.Links) > 0 {
		content += "├─ Links\n"
		for i, link := range m.user.Links {
			if i == len(m.user.Links)-1 {
				content += fmt.Sprintf("│  └─ 🔗 %s: %s\n", link.Name, link.URL)
			} else {
				content += fmt.Sprintf("│  ├─ 🔗 %s: %s\n", link.Name, link.URL)
			}
		}
		content += "│\n"
	}

	// Footer
	content += "└─ Powered by curltree.dev\n\n"

	if m.message != "" {
		content += successStyle.Render(m.message) + "\n\n"
		m.message = ""
	}

	if m.err != nil {
		content += errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
		m.err = nil
	}

	help := helpStyle.Render("ctrl+e: edit • ctrl+d: delete • ctrl+c: exit")
	return content + help
}

func (m *tuiModel) editView() string {
	content := asciiStyle.Render(getASCIIArt()) + "\n\n"
	content += titleStyle.Render("Edit Profile") + "\n\n"
	content += m.form.View()

	if m.err != nil {
		content += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		m.err = nil
	}

	help := helpStyle.Render("tab/shift+tab: navigate • ctrl+n: add link • ctrl+d: delete link • ctrl+s: save • esc: cancel")
	return content + "\n\n" + help
}

func (m *tuiModel) createView() string {
	content := asciiStyle.Render(getASCIIArt()) + "\n\n"
	content += "Welcome! Let's create your profile.\n\n"
	content += m.form.View()

	if m.err != nil {
		content += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		m.err = nil
	}

	help := helpStyle.Render("tab/shift+tab: navigate • ctrl+n: add link • ctrl+d: delete link • ctrl+s: create • esc: exit")
	return content + "\n\n" + help
}

func (m *tuiModel) confirmDeleteView() string {
	content := asciiStyle.Render(getASCIIArt()) + "\n\n"
	content += errorStyle.Render("Are you sure you want to delete your profile?") + "\n"
	content += "This action cannot be undone.\n\n"
	content += helpStyle.Render("y: yes, delete • n: no, cancel")
	return content
}
