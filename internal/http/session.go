package http

import (
	"errors"
	"fmt"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
)

// global session store
var sessionStore *session.Store

const defaultSessionExpire = 24 * time.Hour // maybe move to config?

var errNoAuth = errors.New("no auth")

func newSessionStore() *session.Store {
	return session.New(session.Config{
		Expiration:   defaultSessionExpire,
		KeyLookup:    "cookie:session_id",
		KeyGenerator: utils.UUIDv4,
	})
}

func saveSessionInfo(c *fiber.Ctx, userInfo *models.User) error {
	sess, err := sessionStore.Get(c)
	if err != nil {
		return fmt.Errorf("failed to get session store: %w", err)
	}

	sess.Set("userId", userInfo.ID)
	sess.Set("username", userInfo.Username)
	sess.Set("role", userInfo.Role)

	if err = sess.Save(); err != nil {
		return fmt.Errorf("failed to save session info: %w", err)
	}

	return nil
}

func getSessionInfo(c *fiber.Ctx) (*models.User, error) {
	sess, err := sessionStore.Get(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session store: %w", err)
	}

	var user models.User

	userId := sess.Get("userId")
	if userId == nil {
		return nil, errNoAuth
	}

	userIdCasted := userId.(int)
	if userIdCasted == 0 {
		return nil, errNoAuth
	}

	user.ID = userIdCasted

	username := sess.Get("username")
	if username == nil {
		return nil, errNoAuth
	}

	user.Username = username.(string)

	role := sess.Get("role")
	if role == nil {
		return nil, errNoAuth
	}

	user.Role = role.(string)

	return &user, nil
}
