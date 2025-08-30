package http

import (
	"log/slog"
	"strings"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
)

func internalServerError(err error) string {
	buf := new(strings.Builder)

	tmpl := PrepareTemplate("internal/templates/error.html",
		"internal/templates/base.html",
		"internal/templates/navbar.html")

	data := errorPage{
		Status:  "500",
		Title:   "Internal Server Error",
		Error:   err,
		Message: "Server can't process your request.",
		User:    nil,
	}

	execErr := tmpl.ExecuteTemplate(buf, "base", data)
	if execErr != nil {
		slog.Error("can't execute template", "error", execErr)
	}

	return buf.String()
}

func defaultForbiddenMap() models.Error {
	return models.Error{
		Title:   "Forbidden",
		Status:  "403",
		Error:   nil,
		Message: "Not allowed",
		User:    nil,
	}
}

func defaultInternalErrorMap(err error) models.Error {
	return models.Error{
		Status:  "500",
		Title:   "Internal Server Error",
		Error:   err,
		Message: "Server can't process your request.",
		User:    nil,
	}
}

func defaultAuthRequiredMap() models.Error {
	return models.Error{
		Status:  "401",
		Title:   "Invalid Credentials",
		Error:   nil,
		Message: "Bad username or password, please try again",
		User:    nil,
	}
}
