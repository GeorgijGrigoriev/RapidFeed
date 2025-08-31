package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"github.com/gofiber/fiber/v2"
)

var (
	errEmptyUsername = errors.New("username is empty")
	errEmptyRole     = errors.New("role is empty")
)

func jsonError(code int, err string) string {
	body := map[string]any{"status": code, "error": err}

	marshaledBody, marshalError := json.Marshal(body)
	if marshalError != nil {
		panic(marshalError)
	}

	return string(marshaledBody)
}

func jsonResponse(code int, content any) string {
	body := map[string]any{"status": code, "content": content}

	marshaledBody, marshalError := json.Marshal(body)
	if marshalError != nil {
		panic(marshalError)
	}

	return string(marshaledBody)
}

func getUserIdFromCtx(c *fiber.Ctx) (int, error) {
	userId := c.Get("userId")

	fmt.Println(userId)

	id, err := strconv.Atoi(userId)
	if err != nil {
		return 0, fmt.Errorf("failed to parse user id from ctx: %w", err)
	}

	return id, nil
}

func getUsernameFromCtx(c *fiber.Ctx) (string, error) {
	username := c.Get("username")
	if username == "" {
		return "", errEmptyUsername
	}

	return username, nil
}

func getRoleFromCtx(c *fiber.Ctx) (string, error) {
	role := c.Get("role")
	if role == "" {
		return "", errEmptyRole
	}

	return role, nil
}

func getUserInfoFromCtx(c *fiber.Ctx) (*models.User, error) {
	var (
		user models.User
		err  error
	)

	user.ID, err = getUserIdFromCtx(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from ctx: %w", err)
	}

	user.Username, err = getUsernameFromCtx(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from ctx: %w", err)
	}

	user.Role, err = getRoleFromCtx(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from ctx: %w", err)
	}

	return &user, nil
}
