package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"curltree/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL embed.FS

type DB struct {
	conn *sqlx.DB
}

func NewSQLiteDB(dbPath string) (*DB, error) {
	conn, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign keys
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	schema, err := schemaSQL.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.conn.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema applied successfully")
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetUserBySSHKey(sshPublicKey string) (*models.User, error) {
	var user models.User
	err := db.conn.Get(&user, `
		SELECT id, ssh_public_key, full_name, username, about, created_at, updated_at 
		FROM users 
		WHERE ssh_public_key = ?`, sshPublicKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by SSH key: %w", err)
	}

	links, err := db.GetUserLinks(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}
	user.Links = links

	return &user, nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.conn.Get(&user, `
		SELECT id, ssh_public_key, full_name, username, about, created_at, updated_at 
		FROM users 
		WHERE username = ?`, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	links, err := db.GetUserLinks(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}
	user.Links = links

	return &user, nil
}

func (db *DB) GetPublicProfile(username string) (*models.PublicProfile, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &models.PublicProfile{
		FullName: user.FullName,
		Username: user.Username,
		About:    user.About,
		Links:    user.Links,
	}, nil
}

func (db *DB) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID string
	err = tx.Get(&userID, `
		INSERT INTO users (ssh_public_key, full_name, username, about) 
		VALUES (?, ?, ?, ?) 
		RETURNING id`,
		req.SSHPublicKey, req.FullName, req.Username, req.About)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := db.updateUserLinks(tx, userID, req.Links); err != nil {
		return nil, fmt.Errorf("failed to create user links: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetUserBySSHKey(req.SSHPublicKey)
}

func (db *DB) UpdateUser(userID string, req *models.UpdateUserRequest) (*models.User, error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE users 
		SET full_name = ?, username = ?, about = ?
		WHERE id = ?`,
		req.FullName, req.Username, req.About, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err := db.updateUserLinks(tx, userID, req.Links); err != nil {
		return nil, fmt.Errorf("failed to update user links: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var sshKey string
	err = db.conn.Get(&sshKey, "SELECT ssh_public_key FROM users WHERE id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key: %w", err)
	}

	return db.GetUserBySSHKey(sshKey)
}

func (db *DB) DeleteUser(userID string) error {
	_, err := db.conn.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (db *DB) IsUsernameExists(username string) (bool, error) {
	var count int
	err := db.conn.Get(&count, "SELECT COUNT(*) FROM users WHERE username = ?", username)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return count > 0, nil
}

func (db *DB) GetUserLinks(userID string) ([]models.Link, error) {
	var links []models.Link
	err := db.conn.Select(&links, `
		SELECT id, user_id, name, url, position 
		FROM links 
		WHERE user_id = ? 
		ORDER BY position`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}
	return links, nil
}

func (db *DB) updateUserLinks(tx *sqlx.Tx, userID string, linkInputs []models.LinkInput) error {
	_, err := tx.Exec("DELETE FROM links WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing links: %w", err)
	}

	for i, link := range linkInputs {
		_, err := tx.Exec(`
			INSERT INTO links (user_id, name, url, position) 
			VALUES (?, ?, ?, ?)`,
			userID, link.Name, link.URL, i)
		if err != nil {
			return fmt.Errorf("failed to insert link: %w", err)
		}
	}

	return nil
}