package http

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func New() {
	store = sessions.NewCookieStore([]byte(utils.SecretKey))

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.HandleFunc("/login", LoginHandler)

	if utils.RegisterAllowed {
		http.HandleFunc("/register", RegisterHandler)
	}

	http.Handle("/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("INCOMING", "URL", r.URL.String(),
			"Method", r.Method,
			"IP", r.RemoteAddr,
			"Proto", r.Proto,
			"UserAgent", r.UserAgent())

		if r.URL.Path == "/login" || r.URL.Path == "/register" {
			handler(PrepareTemplate("internal/templates/index.html",
				"internal/templates/navbar.html",
				"internal/templates/base.html"), w, r)

			return
		}

		handler(PrepareTemplate("internal/templates/index.html",
			"internal/templates/navbar.html",
			"internal/templates/base.html"), w, r)
	}))

	http.Handle("/refresh", AuthMiddleware(refreshHandler))
	http.Handle("/settings", AuthMiddleware(userSettingsHandler))
	http.Handle("/admin/users", AdminMiddleware(adminSettingsHandler))
	http.HandleFunc("/logout", LogoutHandler)

	http.HandleFunc("/403", forbiddenHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(RapidFeed.Static)))

	// api section
	http.Handle("/api/users/list", TokenAuthMiddleware(nil))

	slog.Info("Server is now listening", "listen address", utils.Listen)

	log.Fatal(http.ListenAndServe(utils.Listen, nil))
}
