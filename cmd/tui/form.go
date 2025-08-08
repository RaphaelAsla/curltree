package main

import (
	"fmt"
	"strings"

	"curltree/internal/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type formModel struct {
	inputs     []textinput.Model
	focusIndex int
	width      int
}

func newFormModel() *formModel {
	inputs := make([]textinput.Model, 5)

	// Full Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Your full name"
	inputs[0].CharLimit = 100
	inputs[0].Width = 48
	inputs[0].Focus()
	inputs[0].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[0].TextStyle = lipgloss.NewStyle()

	// Username
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Username (alphanumeric, -, _)"
	inputs[1].CharLimit = 50
	inputs[1].Width = 48
	inputs[1].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[1].TextStyle = lipgloss.NewStyle()

	// About
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Tell us about yourself (optional)"
	inputs[2].CharLimit = 500
	inputs[2].Width = 48
	inputs[2].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[2].TextStyle = lipgloss.NewStyle()

	// Link
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Link name"
	inputs[3].CharLimit = 100
	inputs[3].Width = 23
	inputs[3].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[3].TextStyle = lipgloss.NewStyle()

	// Add URL input
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "https://example.com"
	inputs[4].CharLimit = 500
	inputs[4].Width = 23
	inputs[4].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[4].TextStyle = lipgloss.NewStyle()

	return &formModel{
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (f *formModel) populateFromUser(user *models.User) {
	if len(f.inputs) >= 3 {
		f.inputs[0].SetValue(user.FullName)
		f.inputs[1].SetValue(user.Username)
		f.inputs[2].SetValue(user.About)
	}

	f.clearLinks()
	for _, link := range user.Links {
		f.addLink(link.Name, link.URL)
	}
}

func (f *formModel) Update(msg tea.Msg) {
	for i := range f.inputs {
		if i == f.focusIndex {
			f.inputs[i], _ = f.inputs[i].Update(msg)
		}
	}
}

func (f *formModel) clearLinks() {
	// Keep only the first 3 inputs (fullname, username, about)
	if len(f.inputs) > 3 {
		f.inputs = f.inputs[:3]
		if f.focusIndex >= len(f.inputs) {
			f.focusIndex = len(f.inputs) - 1
		}
	}
}

func (f *formModel) addLink(name, url string) {
	// Add name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Link name"
	nameInput.CharLimit = 100
	nameInput.Width = 23
	nameInput.SetValue(name)
	nameInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	nameInput.TextStyle = lipgloss.NewStyle()
	nameInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	// Add URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "https://example.com"
	urlInput.CharLimit = 500
	urlInput.Width = 23
	urlInput.SetValue(url)
	urlInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	urlInput.TextStyle = lipgloss.NewStyle()
	urlInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	f.inputs = append(f.inputs, nameInput, urlInput)
}

func (f *formModel) deleteCurrentLink() {
	if len(f.inputs) <= 3 {
		return
	}

	linkInputIndex := -1
	for i := 3; i < len(f.inputs); i += 2 {
		if f.focusIndex == i || f.focusIndex == i+1 {
			linkInputIndex = i
			break
		}
	}

	if linkInputIndex != -1 {
		f.inputs = append(f.inputs[:linkInputIndex], f.inputs[linkInputIndex+2:]...)

		if f.focusIndex >= len(f.inputs) {
			f.focusIndex = len(f.inputs) - 1
		}
	}
}

func (f *formModel) nextField() {
	if f.focusIndex < len(f.inputs)-1 {
		f.inputs[f.focusIndex].Blur()
		f.focusIndex++
		f.inputs[f.focusIndex].Focus()
	}
}

func (f *formModel) prevField() {
	if f.focusIndex > 0 {
		f.inputs[f.focusIndex].Blur()
		f.focusIndex--
		f.inputs[f.focusIndex].Focus()
	}
}

func (f *formModel) validate() error {
	if len(f.inputs) < 3 {
		return fmt.Errorf("form not properly initialized")
	}

	// Validate full name (required)
	fullName := strings.TrimSpace(f.inputs[0].Value())
	if fullName == "" {
		return fmt.Errorf("full name is required")
	}

	// Validate username (required)
	username := strings.TrimSpace(f.inputs[1].Value())
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if !isValidUsername(username) {
		return fmt.Errorf("username must be alphanumeric with optional hyphens and underscores")
	}

	// Validate links (pairs of name/URL)
	for i := 3; i < len(f.inputs); i += 2 {
		if i+1 >= len(f.inputs) {
			continue
		}

		name := strings.TrimSpace(f.inputs[i].Value())
		url := strings.TrimSpace(f.inputs[i+1].Value())

		if name != "" || url != "" { // If either is filled, both must be valid
			if name == "" {
				return fmt.Errorf("link name is required")
			}
			if url == "" {
				return fmt.Errorf("link URL is required")
			}
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("URL must start with http:// or https://")
			}
		}
	}
	return nil
}

func (f *formModel) toCreateRequest(sshKey string) *models.CreateUserRequest {
	req := &models.CreateUserRequest{
		SSHPublicKey: sshKey,
		Links:        []models.LinkInput{},
	}

	if len(f.inputs) >= 3 {
		req.FullName = strings.TrimSpace(f.inputs[0].Value())
		req.Username = strings.TrimSpace(f.inputs[1].Value())
		req.About = strings.TrimSpace(f.inputs[2].Value())
	}

	for i := 3; i < len(f.inputs); i += 2 {
		if i+1 < len(f.inputs) {
			name := strings.TrimSpace(f.inputs[i].Value())
			url := strings.TrimSpace(f.inputs[i+1].Value())
			if name != "" && url != "" {
				req.Links = append(req.Links, models.LinkInput{
					Name: name,
					URL:  url,
				})
			}
		}
	}

	return req
}

func (f *formModel) toUpdateRequest() *models.UpdateUserRequest {
	req := &models.UpdateUserRequest{
		Links: []models.LinkInput{},
	}

	if len(f.inputs) >= 3 {
		req.FullName = strings.TrimSpace(f.inputs[0].Value())
		req.Username = strings.TrimSpace(f.inputs[1].Value())
		req.About = strings.TrimSpace(f.inputs[2].Value())
	}

	for i := 3; i < len(f.inputs); i += 2 {
		if i+1 < len(f.inputs) {
			name := strings.TrimSpace(f.inputs[i].Value())
			url := strings.TrimSpace(f.inputs[i+1].Value())
			if name != "" && url != "" {
				req.Links = append(req.Links, models.LinkInput{
					Name: name,
					URL:  url,
				})
			}
		}
	}

	return req
}

func (f *formModel) View() string {
	var content strings.Builder

	labels := []string{"Full Name *", "Username *", "About"}

	// Define styles for boxes and labels with more prominent focus
	focusedBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Width(50)

	normalBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Width(50)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	normalLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Render basic fields (Full Name, Username, About)
	for i := 0; i < 3 && i < len(f.inputs); i++ {
		label := labels[i]

		var labelStyle lipgloss.Style
		var boxStyle lipgloss.Style
		if i == f.focusIndex {
			labelStyle = focusedLabelStyle
			boxStyle = focusedBoxStyle
			f.inputs[i].Focus()
		} else {
			labelStyle = normalLabelStyle
			boxStyle = normalBoxStyle
			f.inputs[i].Blur()
		}

		content.WriteString(labelStyle.Render(label) + "\n")
		content.WriteString(boxStyle.Render(f.inputs[i].View()) + "\n\n")
	}

	// Render links (name and URL side by side)
	for i := 3; i < len(f.inputs); i += 2 {
		if i+1 >= len(f.inputs) {
			break
		}

		linkIndex := (i-3)/2 + 1
		nameLabel := fmt.Sprintf("Link %d Name", linkIndex)
		urlLabel := fmt.Sprintf("Link %d URL", linkIndex)

		// Style for name field
		var nameLabelStyle, nameBoxStyle lipgloss.Style
		if i == f.focusIndex {
			nameLabelStyle = focusedLabelStyle
			nameBoxStyle = focusedBoxStyle.Width(21)
			f.inputs[i].Focus()
		} else {
			nameLabelStyle = normalLabelStyle
			nameBoxStyle = normalBoxStyle.Width(21)
			f.inputs[i].Blur()
		}

		// Style for URL field
		var urlLabelStyle, urlBoxStyle lipgloss.Style
		if i+1 == f.focusIndex {
			urlLabelStyle = focusedLabelStyle
			urlBoxStyle = focusedBoxStyle.Width(25)
			f.inputs[i+1].Focus()
		} else {
			urlLabelStyle = normalLabelStyle
			urlBoxStyle = normalBoxStyle.Width(25)
			f.inputs[i+1].Blur()
		}

		// Render labels on the same line
		labelsRow := lipgloss.JoinHorizontal(lipgloss.Top,
			nameLabelStyle.Width(25).Render(nameLabel),
			lipgloss.NewStyle().Width(2).Render("  "),
			urlLabelStyle.Width(25).Render(urlLabel))
		content.WriteString(labelsRow + "\n")

		// Render input boxes on the same line
		inputsRow := lipgloss.JoinHorizontal(lipgloss.Top,
			nameBoxStyle.Render(f.inputs[i].View()),
			lipgloss.NewStyle().Width(2).Render("  "),
			urlBoxStyle.Render(f.inputs[i+1].View()))
		content.WriteString(inputsRow + "\n\n")
	}

	return content.String()
}

func isValidUsername(username string) bool {
	if len(username) < 1 || len(username) > 50 {
		return false
	}
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}
