package http

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/ui"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore
var sessionStore *session.Store

func New() {
	// delete after migrate to gofiber
	store = sessions.NewCookieStore([]byte(utils.SecretKey))

	// delete after migrate to gofiber
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	sessionStore = newSessionStore()

	app := fiber.New(fiber.Config{
		Views: initTemplateEngine(),
	})

	// static
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(ui.Static),
		PathPrefix: "static",
		Browse:     false,
	}))

	app.Use(logger.New())

	app.All("/login", loginHandler)

	// protected app routes with check session middleware
	appRouters := app.Group("/", CheckSessionMiddleware())
	appRouters.Get("/", feedsPageHandler)
	appRouters.All("/settings", nil)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("yo")
	})

	log.Fatal(app.Listen(":3000"))

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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(ui.Static)))

	// api section
	http.Handle("/api/users/list", TokenAuthMiddleware(nil))
	http.Handle("/api/feeds/get", TokenAuthMiddleware(getUserFeedsByTimeRange))

	slog.Info("Server is now listening", "listen address", utils.Listen)

	log.Fatal(http.ListenAndServe(utils.Listen, nil))
}
