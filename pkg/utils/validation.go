package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	sshKeyRegex   = regexp.MustCompile(`^ssh-[a-z0-9]+ [A-Za-z0-9+/=]+ ?.*$`)
)

func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(username) < 2 {
		return fmt.Errorf("username must be at least 2 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username cannot be longer than 50 characters")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, hyphens, and underscores")
	}
	if strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") ||
		strings.HasPrefix(username, "_") || strings.HasSuffix(username, "_") {
		return fmt.Errorf("username cannot start or end with hyphens or underscores")
	}
	return nil
}

func ValidateFullName(fullName string) error {
	if strings.TrimSpace(fullName) == "" {
		return fmt.Errorf("full name cannot be empty")
	}
	if len(fullName) > 100 {
		return fmt.Errorf("full name cannot be longer than 100 characters")
	}
	return nil
}

func ValidateAbout(about string) error {
	if len(about) > 500 {
		return fmt.Errorf("about section cannot be longer than 500 characters")
	}
	return nil
}

func ValidateURL(linkURL string) error {
	if linkURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(linkURL) > 500 {
		return fmt.Errorf("URL cannot be longer than 500 characters")
	}
	
	parsedURL, err := url.Parse(linkURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}
	
	return nil
}

func ValidateLinkName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("link name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("link name cannot be longer than 100 characters")
	}
	return nil
}

func ValidateSSHKey(sshKey string) error {
	if sshKey == "" {
		return fmt.Errorf("SSH key cannot be empty")
	}
	
	sshKey = strings.TrimSpace(sshKey)
	if !sshKeyRegex.MatchString(sshKey) {
		return fmt.Errorf("invalid SSH key format")
	}
	
	return nil
}

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.ReplaceAll(input, "\r", "")
	return input
}

func SanitizeHTML(input string) string {
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}