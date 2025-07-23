package utils

import (
	"os"
	"strings"
)

var (
	Listen          string
	SecretKey       string
	RegisterAllowed bool
)

func GetStringEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func GetBoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return strings.ToLower(value) == "true"
	}

	return fallback
}
