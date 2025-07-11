package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

func handler(tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	rows, err := db.DB.Query("SELECT COUNT(*) FROM feeds")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var totalCount int
	for rows.Next() {
		err := rows.Scan(&totalCount)
		if err != nil {
			log.Fatal(err)
		}
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	rows, err = db.DB.Query("SELECT title, link, date, source, description FROM feeds ORDER BY date DESC LIMIT ? OFFSET ?", perPage, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var items []feeder.FeedItem
	for rows.Next() {
		var item feeder.FeedItem
		err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description)
		if err != nil {
			log.Fatal(err)
		}

		items = append(items, item)
	}

	paginatedItems := feeder.PaginatedFeedItems{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		TotalItems: totalCount,
	}

	session, _ := store.Get(r, "rapid-feed")
	userID, ok := session.Values["user_id"].(int)
	var user feeder.User
	if ok && userID != 0 {
		err := db.DB.QueryRow(`SELECT id, username, role FROM users WHERE id = ?`, userID).Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			log.Println("Error fetching user:", err)
		}
	}

	data := map[string]interface{}{
		"PaginatedItems": paginatedItems,
		"User":           user,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
}
