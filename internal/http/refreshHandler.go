package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"net/http"
)

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	feeder.FetchAndSaveFeeds(toLoadFeeds)
	http.Redirect(w, r, "/", http.StatusFound)
}
