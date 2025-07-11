package http

import (
	"log"
	"net/http"
)

var toLoadFeeds []string

func New(feeds []string) {

	toLoadFeeds = feeds

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(PrepareTemplate(), w, r)
	})
	http.HandleFunc("/refresh", refreshHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
