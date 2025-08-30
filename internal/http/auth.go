package http

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/ui"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
)

var loginTemplate = prepareHTMLTemplate("login")
var registerTemplate = prepareHTMLTemplate("register")

func prepareHTMLTemplate(name string) *template.Template {
	f, _ := ui.HTMLTemplates.ReadFile(fmt.Sprintf("internal/templates/%s.html", name))

	return template.Must(template.New(name).Parse(string(f)))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			slog.Error("username or password is empty")

			internalServerErrorHandler(w, r, fmt.Errorf("username or pasword is empty but required"))

			return
		}

		blocked, err := db.CheckUserBlocked(username)
		if err != nil {
			slog.Error("username blocked check failed", "username", username, "error", err)

			invalidCredentialsHandler(w, r)

			return
		}

		if blocked {
			forbiddenHandler(w, r)

			return
		}

		storedHash, err := db.GetUserHash(username)
		if err != nil {
			slog.Error("username not found", "error", err)

			invalidCredentialsHandler(w, r)

			return
		}

		err = auth.CheckPassword(storedHash, password)
		if err != nil {
			slog.Error("invalid password", "error", err)

			invalidCredentialsHandler(w, r)

			return
		}

		var userID int
		err = db.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&userID)
		if err != nil {

			invalidCredentialsHandler(w, r)

			return
		}

		session, err := store.Get(r, utils.SecretKey)
		if err != nil {
			slog.Error("old session possibly corrupted, creating new one", "error", err)
		}

		session.Values["user_id"] = userID

		err = session.Save(r, w)
		if err != nil {
			slog.Error("session save error", "error", err)

			internalServerErrorHandler(w, r, fmt.Errorf("failed to save session"))

			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		data := map[string]interface{}{
			"RegisterAllowed": utils.RegisterAllowed,
		}

		loginTemplate.Execute(w, data)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			slog.Error("username or password is empty")

			internalServerErrorHandler(w, r, fmt.Errorf("username or pasword is empty but required"))

			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			slog.Error("hashing error", "error", err)

			internalServerErrorHandler(w, r, fmt.Errorf("failed to create user"))

			return
		}

		_, err = db.DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, 'user')`, username, hash)
		if err != nil {
			slog.Error("new user failed", "error", err)

			internalServerErrorHandler(w, r, fmt.Errorf("failed to create user"))

			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		registerTemplate.Execute(w, nil)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, utils.SecretKey)
	session.Options.MaxAge = -1

	err := session.Save(r, w)
	if err != nil {
		slog.Error("failed to invalidate session", "error", err)

		internalServerErrorHandler(w, r, fmt.Errorf("failed to invalidate session"))

		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}
