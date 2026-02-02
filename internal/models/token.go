package models

import "time"

type Token struct {
	ID          int       `json:"id"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permissions int       `json:"permissions"`
}
