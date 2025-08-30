package http

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
)

const defaultSessionExpire = 24 * time.Hour // maybe move to config?

func newSessionStore() *session.Store {
	return session.New(session.Config{
		Expiration:   defaultSessionExpire,
		KeyLookup:    "cookie:session_id",
		KeyGenerator: utils.UUIDv4,
	})
}
