package auth

import (
	"context"

	"curltree/internal/models"
)

type contextKey string

const (
	userKey   contextKey = "user"
	sshKeyKey contextKey = "ssh_key"
)

func GetUser(ctx context.Context) *models.User {
	user, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetSSHKey(ctx context.Context) string {
	sshKey, ok := ctx.Value(sshKeyKey).(string)
	if !ok {
		return ""
	}
	return sshKey
}