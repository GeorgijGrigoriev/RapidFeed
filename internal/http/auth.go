package http

import (
	"fmt"
	"github.com/GeorgijGrigoriev/RapidFeed"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"html/template"
	"log/slog"
	"net/http"
)

var loginTemplate = prepareHTMLTemplate("login")
var registerTemplate = prepareHTMLTemplate("register")

func prepareHTMLTemplate(name string) *template.Template {
	f, err := RapidFeed.HTMLTemplates.ReadFile(fmt.Sprintf("internal/templates/%s.html", name))
	if err != nil {
		panic(err)
	}

	return template.Must(template.New(name).Parse(string(f)))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var userID int
		err := db.DB.QueryRow(`SELECT id FROM users WHERE username = ? AND password = ?`, username, password).Scan(&userID)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		session, err := store.Get(r, "rapid-feed")
		if err != nil {
			slog.Error("old session possibly corrupted, creating new one", "error", err)
		}

		session.Values["user_id"] = userID

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound) // Redirect to the main page after successful login
	} else {

		loginTemplate.Execute(w, nil)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		_, err := db.DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, 'user')`, username, password)
		if err != nil {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		registerTemplate.Execute(w, nil)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "rapid-feed")
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		slog.Error("failed to invalidate session", "error", err)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
