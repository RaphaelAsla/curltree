package utils

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "testuser", false},
		{"valid with numbers", "test123", false},
		{"valid with hyphen", "test-user", false},
		{"valid with underscore", "test_user", false},
		{"empty username", "", true},
		{"too short", "a", true},
		{"too long", string(make([]byte, 51)), true},
		{"starts with hyphen", "-testuser", true},
		{"ends with hyphen", "testuser-", true},
		{"starts with underscore", "_testuser", true},
		{"ends with underscore", "testuser_", true},
		{"invalid characters", "test@user", true},
		{"spaces", "test user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/path", false},
		{"valid with query", "https://example.com/path?query=1", false},
		{"empty url", "", true},
		{"no scheme", "example.com", true},
		{"invalid scheme", "ftp://example.com", true},
		{"no host", "https://", true},
		{"too long", "https://" + string(make([]byte, 500)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFullName(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		wantErr  bool
	}{
		{"valid name", "John Doe", false},
		{"single name", "John", false},
		{"empty name", "", true},
		{"only spaces", "   ", true},
		{"too long", string(make([]byte, 101)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFullName(tt.fullName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFullName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLinkName(t *testing.T) {
	tests := []struct {
		name     string
		linkName string
		wantErr  bool
	}{
		{"valid name", "My Website", false},
		{"empty name", "", true},
		{"only spaces", "   ", true},
		{"too long", string(make([]byte, 101)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLinkName(tt.linkName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLinkName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal input", "hello world", "hello world"},
		{"with spaces", "  hello world  ", "hello world"},
		{"with null bytes", "hello\x00world", "helloworld"},
		{"with carriage return", "hello\rworld", "helloworld"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal input", "hello world", "hello world"},
		{"with HTML tags", "hello <script>alert('xss')</script>", "hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"with ampersand", "hello & world", "hello &amp; world"},
		{"with quotes", `hello "world"`, "hello &quot;world&quot;"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}