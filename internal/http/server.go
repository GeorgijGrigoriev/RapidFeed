package http

import (
	"log"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/ui"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

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

	app.Get("/login", loginRender)
	app.Post("/login", loginHandler)

	if utils.RegisterAllowed {
		app.Get("/register", registerRender)
		app.Post("/register", registerHandler)
	}

	// protected app routes with check session middleware
	appRoutes := app.Group("/", checkSessionMiddleware())
	appRoutes.Get("/", feedsPageHandler)
	appRoutes.Get("/refresh", refreshHandler)
	appRoutes.Get("/settings", userSettingsRender)
	appRoutes.Get("/logout", logoutHandler)

	internalApiRoutes := app.Group("/internal/api/", checkSessionMiddleware())
	internalApiRoutes.Post("/user/settings/password/change", changePasswordHandler)
	internalApiRoutes.Post("/user/settings/feed/add", nil)
	internalApiRoutes.Post("/user/settings/feed/delete", nil)
	internalApiRoutes.Post("/user/settings/apiToken/add", nil)
	internalApiRoutes.Post("user/settings/apiToken/revoke", nil)

	adminRoutes := app.Group("/admin/", adminSessionMiddleware())
	adminRoutes.Get("/admin/users", nil)

	log.Fatal(app.Listen(utils.Listen))

	//http.Handle("/refresh", AuthMiddleware(refreshHandler))
	//http.Handle("/settings", AuthMiddleware(userSettingsHandler))
	//http.Handle("/admin/users", AdminMiddleware(adminSettingsHandler))
	//
	//http.HandleFunc("/403", forbiddenHandler)
	//
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(ui.Static)))
	//
	//// api section
	//http.Handle("/api/users/list", TokenAuthMiddleware(nil))
	//http.Handle("/api/feeds/get", TokenAuthMiddleware(getUserFeedsByTimeRange))
	//
	//slog.Info("Server is now listening", "listen address", utils.Listen)
	//
	//log.Fatal(http.ListenAndServe(utils.Listen, nil))
}
