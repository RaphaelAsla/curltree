package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	SSHPublicKey string    `json:"ssh_public_key" db:"ssh_public_key"`
	FullName     string    `json:"full_name" db:"full_name"`
	Username     string    `json:"username" db:"username"`
	About        string    `json:"about" db:"about"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Links        []Link    `json:"links"`
}

type Link struct {
	ID       string `json:"id" db:"id"`
	UserID   string `json:"user_id" db:"user_id"`
	Name     string `json:"name" db:"name"`
	URL      string `json:"url" db:"url"`
	Position int    `json:"position" db:"position"`
}

type CreateUserRequest struct {
	SSHPublicKey string `json:"ssh_public_key"`
	FullName     string `json:"full_name"`
	Username     string `json:"username"`
	About        string `json:"about"`
	Links        []LinkInput `json:"links"`
}

type UpdateUserRequest struct {
	FullName string      `json:"full_name"`
	Username string      `json:"username"`
	About    string      `json:"about"`
	Links    []LinkInput `json:"links"`
}

type LinkInput struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PublicProfile struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
	About    string `json:"about"`
	Links    []Link `json:"links"`
}