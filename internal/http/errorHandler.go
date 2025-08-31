package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
)

const errorTemplate = "templates/error"

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
