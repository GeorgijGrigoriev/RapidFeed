package RapidFeed

import (
	"embed"
)

//go:embed internal/templates/*
var HTMLTemplates embed.FS

//go:embed internal/static/*
var Static embed.FS
