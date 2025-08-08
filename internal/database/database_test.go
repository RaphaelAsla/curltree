package database

import (
	"testing"

	"curltree/internal/models"
)

func setupTestDB(t *testing.T) *DB {
	tmpFile := t.TempDir() + "/test.db"
	db, err := NewSQLiteDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	req := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links: []models.LinkInput{
			{Name: "Website", URL: "https://example.com"},
			{Name: "GitHub", URL: "https://github.com/testuser"},
		},
	}

	user, err := db.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.FullName != req.FullName {
		t.Errorf("Expected FullName %s, got %s", req.FullName, user.FullName)
	}
	if user.Username != req.Username {
		t.Errorf("Expected Username %s, got %s", req.Username, user.Username)
	}
	if len(user.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(user.Links))
	}

	retrievedUser, err := db.GetUserBySSHKey(req.SSHPublicKey)
	if err != nil {
		t.Fatalf("GetUserBySSHKey failed: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("Expected user, got nil")
	}
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
	}

	retrievedUser, err = db.GetUserByUsername(req.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("Expected user, got nil")
	}
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	createReq := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links: []models.LinkInput{
			{Name: "Website", URL: "https://example.com"},
		},
	}

	user, err := db.CreateUser(createReq)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	updateReq := &models.UpdateUserRequest{
		FullName: "Updated User",
		Username: "updateduser",
		About:    "Updated about",
		Links: []models.LinkInput{
			{Name: "New Website", URL: "https://newexample.com"},
			{Name: "GitHub", URL: "https://github.com/updateduser"},
		},
	}

	updatedUser, err := db.UpdateUser(user.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if updatedUser.FullName != updateReq.FullName {
		t.Errorf("Expected FullName %s, got %s", updateReq.FullName, updatedUser.FullName)
	}
	if updatedUser.Username != updateReq.Username {
		t.Errorf("Expected Username %s, got %s", updateReq.Username, updatedUser.Username)
	}
	if len(updatedUser.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(updatedUser.Links))
	}
}

func TestUsernameExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	exists, err := db.IsUsernameExists("nonexistent")
	if err != nil {
		t.Fatalf("IsUsernameExists failed: %v", err)
	}
	if exists {
		t.Error("Expected username to not exist")
	}

	req := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links:        []models.LinkInput{},
	}

	_, err = db.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	exists, err = db.IsUsernameExists("testuser")
	if err != nil {
		t.Fatalf("IsUsernameExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected username to exist")
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	req := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links:        []models.LinkInput{},
	}

	user, err := db.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = db.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	retrievedUser, err := db.GetUserBySSHKey(req.SSHPublicKey)
	if err != nil {
		t.Fatalf("GetUserBySSHKey failed: %v", err)
	}
	if retrievedUser != nil {
		t.Error("Expected user to be deleted")
	}
}

func TestGetPublicProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	req := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links: []models.LinkInput{
			{Name: "Website", URL: "https://example.com"},
		},
	}

	_, err := db.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	profile, err := db.GetPublicProfile("testuser")
	if err != nil {
		t.Fatalf("GetPublicProfile failed: %v", err)
	}
	if profile == nil {
		t.Fatal("Expected profile, got nil")
	}
	if profile.FullName != req.FullName {
		t.Errorf("Expected FullName %s, got %s", req.FullName, profile.FullName)
	}
	if len(profile.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(profile.Links))
	}

	profile, err = db.GetPublicProfile("nonexistent")
	if err != nil {
		t.Fatalf("GetPublicProfile failed: %v", err)
	}
	if profile != nil {
		t.Error("Expected profile to be nil for nonexistent user")
	}
}