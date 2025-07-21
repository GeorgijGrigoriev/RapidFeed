package http

import (
	"log/slog"
	"net/http"
)

type errorPage struct {
	Status  string
	Title   string
	Error   error
	Message string
	User    any // kostyl
}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := PrepareTemplate("internal/templates/error.html",
		"internal/templates/base.html",
		"internal/templates/navbar.html")

	data := errorPage{
		Title:   "Forbidden",
		Status:  "403",
		Error:   nil,
		Message: "Not allowed",
		User:    nil,
	}

	execErr := tmpl.ExecuteTemplate(w, "base", data)
	if execErr != nil {
		slog.Error("can't execute template", "error", execErr)
	}
}

func internalServerErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
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

	execErr := tmpl.ExecuteTemplate(w, "base", data)
	if execErr != nil {
		slog.Error("can't execute template", "error", execErr)
	}
}

func invalidCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := PrepareTemplate("internal/templates/error.html",
		"internal/templates/base.html",
		"internal/templates/navbar.html")

	data := errorPage{
		Status:  "401",
		Title:   "Invalid Credentials",
		Error:   nil,
		Message: "Bad username or password, please try again",
		User:    nil,
	}

	execErr := tmpl.ExecuteTemplate(w, "base", data)
	if execErr != nil {
		slog.Error("can't execute template", "error", execErr)
	}
}
