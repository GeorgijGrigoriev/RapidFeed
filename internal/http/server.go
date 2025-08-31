package http

import (
	"log"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/ui"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func New() {
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
	internalApiRoutes.Post("/user/settings/feed/add", addFeedHandler)
	internalApiRoutes.Post("/user/settings/feed/remove", removeFeedHandler)
	internalApiRoutes.Post("/user/settings/autorefresh/set", autorefreshIntervalChangeHadler)
	internalApiRoutes.Post("/user/settings/apiToken/add", nil)
	internalApiRoutes.Post("user/settings/apiToken/revoke", nil)

	adminRoutes := app.Group("/admin/", adminSessionMiddleware())
	adminRoutes.Get("/users", adminSettingsRender)

	adminApiRoutes := app.Group("/internal/api/admin/", adminSessionMiddleware())
	adminApiRoutes.Post("/user/add", addUserHandler)
	adminApiRoutes.Post("/user/block", blockUserHandler)
	adminApiRoutes.Post("/user/unblock", unblockUserHandler)
	adminApiRoutes.Post("/user/feed/remove", removeUserFeedHandler)

	log.Fatal(app.Listen(utils.Listen))
}
