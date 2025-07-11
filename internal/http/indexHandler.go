package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"html/template"
	"log"
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

	totalPages := (totalCount + perPage - 1) / perPage

	rows, err = db.DB.Query("SELECT title, link, date, source FROM feeds ORDER BY date DESC LIMIT ? OFFSET ?", perPage, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var items []feeder.FeedItem
	for rows.Next() {
		var item feeder.FeedItem
		err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source)
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
	}

	err = tmpl.Execute(w, paginatedItems)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
}
