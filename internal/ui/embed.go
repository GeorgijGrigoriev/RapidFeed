package ui

import (
	"embed"
)

//go:embed templates/*
var HTMLTemplates embed.FS

//go:embed static/*
var Static embed.FS
