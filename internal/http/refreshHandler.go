package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"net/http"
)

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := checkSession(r)
	if err != nil {
		internalServerErrorHandler(w, r, err)

		return
	}

	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		internalServerErrorHandler(w, r, err)
	}

	feeder.FetchAndSaveFeeds(userFeeds)

	http.Redirect(w, r, "/", http.StatusFound)
}
