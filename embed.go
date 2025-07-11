package RapidFeed

import (
	"embed"
)

//go:embed internal/templates/*
var HTMLTemplates embed.FS
