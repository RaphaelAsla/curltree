package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"curltree/internal/database"
	"curltree/internal/models"
	"curltree/pkg/utils"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	profile, err := h.db.GetPublicProfile(username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	acceptHeader := r.Header.Get("Accept")
	userAgent := r.Header.Get("User-Agent")

	if strings.Contains(acceptHeader, "application/json") || strings.Contains(userAgent, "curl") == false {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	h.renderPlainText(w, profile)
}

func (h *Handler) renderPlainText(w http.ResponseWriter, profile *models.PublicProfile) {
	// Header with box drawing
	fmt.Fprintf(w, "┌─ %s (@%s)\n", profile.FullName, profile.Username)
	fmt.Fprintf(w, "│\n")
	
	// About section
	if profile.About != "" {
		fmt.Fprintf(w, "├─ About:\n")
		fmt.Fprintf(w, "│  ├─ ")
		
		// Split about text into words for proper wrapping
		words := strings.Fields(profile.About)
		currentLine := ""
		linePrefix := "│     "
		maxLineLength := 60
		
		for i, word := range words {
			testLine := currentLine + word
			if i > 0 {
				testLine = currentLine + " " + word
			}
			
			if len(testLine) > maxLineLength && currentLine != "" {
				fmt.Fprintf(w, "%s\n%s", currentLine, linePrefix)
				currentLine = word
			} else {
				if i > 0 {
					currentLine += " "
				}
				currentLine += word
			}
		}
		
		if currentLine != "" {
			fmt.Fprintf(w, "%s\n", currentLine)
		}
		fmt.Fprintf(w, "│\n")
	}
	
	// Links section
	if len(profile.Links) > 0 {
		fmt.Fprintf(w, "├─ Links\n")
		for i, link := range profile.Links {
			if i == len(profile.Links)-1 {
				fmt.Fprintf(w, "│  └─ 🔗 %s: %s\n", link.Name, link.URL)
			} else {
				fmt.Fprintf(w, "│  ├─ 🔗 %s: %s\n", link.Name, link.URL)
			}
		}
		fmt.Fprintf(w, "│\n")
	}
	
	// Footer
	fmt.Fprintf(w, "└─ Powered by curltree.dev\n")
}

func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validateCreateRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := h.db.IsUsernameExists(req.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	user, err := h.db.CreateUser(&req)
	if err != nil {
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validateUpdateRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.db.UpdateUser(userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteUser(userID); err != nil {
		http.Error(w, "Failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) validateCreateRequest(req *models.CreateUserRequest) error {
	req.SSHPublicKey = utils.SanitizeInput(req.SSHPublicKey)
	req.FullName = utils.SanitizeInput(req.FullName)
	req.Username = utils.SanitizeInput(req.Username)
	req.About = utils.SanitizeInput(req.About)

	if err := utils.ValidateSSHKey(req.SSHPublicKey); err != nil {
		return utils.NewValidationError("ssh_public_key", err.Error())
	}
	if err := utils.ValidateFullName(req.FullName); err != nil {
		return utils.NewValidationError("full_name", err.Error())
	}
	if err := utils.ValidateUsername(req.Username); err != nil {
		return utils.NewValidationError("username", err.Error())
	}
	if err := utils.ValidateAbout(req.About); err != nil {
		return utils.NewValidationError("about", err.Error())
	}
	return h.validateLinks(req.Links)
}

func (h *Handler) validateUpdateRequest(req *models.UpdateUserRequest) error {
	req.FullName = utils.SanitizeInput(req.FullName)
	req.Username = utils.SanitizeInput(req.Username)
	req.About = utils.SanitizeInput(req.About)

	if err := utils.ValidateFullName(req.FullName); err != nil {
		return utils.NewValidationError("full_name", err.Error())
	}
	if err := utils.ValidateUsername(req.Username); err != nil {
		return utils.NewValidationError("username", err.Error())
	}
	if err := utils.ValidateAbout(req.About); err != nil {
		return utils.NewValidationError("about", err.Error())
	}
	return h.validateLinks(req.Links)
}

func (h *Handler) validateLinks(links []models.LinkInput) error {
	for i, link := range links {
		sanitizedName := utils.SanitizeInput(link.Name)
		sanitizedURL := utils.SanitizeInput(link.URL)
		
		if err := utils.ValidateLinkName(sanitizedName); err != nil {
			return utils.NewValidationError(fmt.Sprintf("link[%d].name", i), err.Error())
		}
		if err := utils.ValidateURL(sanitizedURL); err != nil {
			return utils.NewValidationError(fmt.Sprintf("link[%d].url", i), err.Error())
		}
		
		links[i].Name = sanitizedName
		links[i].URL = sanitizedURL
	}
	return nil
}