package http

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("RAPID_FEED_SESSION_KEY")))
var toLoadFeeds []string

func New(feeds []string) {
	toLoadFeeds = feeds

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode, // Критично для мобильных!
	}

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)

	// Wrap all other routes with AuthMiddleware
	http.Handle("/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("INCOMING", "URL", r.URL.String(),
			"Method", r.Method,
			"IP", r.RemoteAddr,
			"Proto", r.Proto,
			"UserAgent", r.UserAgent())

		if r.URL.Path == "/login" || r.URL.Path == "/register" {
			handler(PrepareTemplate("internal/templates/index.html"), w, r)
			return
		}

		handler(PrepareTemplate("internal/templates/index.html"), w, r)
	}))
	http.Handle("/refresh", AuthMiddleware(refreshHandler))
	http.Handle("/settings", AuthMiddleware((&UserSettingsHandler{}).ServeHTTP))
	http.Handle("/admin/users", AdminMiddleware((&AdminUsersHandler{}).ServeHTTP))
	http.HandleFunc("/logout", LogoutHandler)

	listen := flag.String("listen", ":8080", "listen host:port")

	flag.Parse()

	log.Fatal(http.ListenAndServe(*listen, nil))
}
