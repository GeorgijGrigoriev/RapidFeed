package db

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in‑memory sqlite: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test db: %v", err)
		}
	})

	DB = db

	schema := `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            password TEXT,
            role TEXT NOT NULL
        );`
	if _, err := DB.Exec(schema); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
}

func TestGetUserInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		setupTestDB(t)

		_, err := DB.Exec(
			`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`,
			"alice", "secret", "admin")
		if err != nil {
			t.Fatalf("failed to insert test user: %v", err)
		}

		user, err := GetUserInfoById(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.ID != 1 || user.Username != "alice" || user.Role != "admin" {
			t.Fatalf("unexpected user data: %+v", user)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		setupTestDB(t)

		_, err := GetUserInfoById(42)
		if err == nil {
			t.Fatalf("expected error for missing user, got nil")
		}

		if !strings.Contains(err.Error(), "failed to get user info") {
			t.Fatalf("error does not contain expected prefix: %v", err)
		}
	})
}
