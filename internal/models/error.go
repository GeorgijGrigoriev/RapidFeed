package models

type Error struct {
	Status  string
	Title   string
	Error   error
	Message string
	User    any // kostyl
}
