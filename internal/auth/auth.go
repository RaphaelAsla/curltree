package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"curltree/internal/database"
	"curltree/internal/models"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) Middleware() wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			// Let the SSH session continue - public key validation
			// will be handled in the TUI after the handshake is complete
			sh(s)
		}
	}
}

func (a *AuthService) IsUserRegistered(sshKey string) (bool, *models.User, error) {
	user, err := a.db.GetUserBySSHKey(sshKey)
	if err != nil {
		return false, nil, err
	}
	return user != nil, user, nil
}

func formatSSHKey(key ssh.PublicKey) string {
	keyBytes := key.Marshal()
	hash := sha256.Sum256(keyBytes)
	return fmt.Sprintf("%s:%s", key.Type(), hex.EncodeToString(hash[:]))
}

func normalizeSSHKey(keyString string) string {
	parts := strings.Fields(keyString)
	if len(parts) >= 2 {
		return fmt.Sprintf("%s %s", parts[0], parts[1])
	}
	return keyString
}

