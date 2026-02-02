package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
)

const maxMCPItems = 1000

type mcpFeedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Date        string `json:"date"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type mcpFeedsResponse struct {
	Items       []mcpFeedItem `json:"items"`
	Count       int           `json:"count"`
	Period      string        `json:"period,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	GeneratedAt string        `json:"generated_at"`
}

type mcpErrorResponse struct {
	Error string `json:"error"`
}

func mcpFeedsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	period := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("period")))
	limitStr := strings.TrimSpace(r.URL.Query().Get("limit"))
	if period != "" && limitStr != "" {
		writeJSONError(w, http.StatusBadRequest, "use either period or limit, not both")
		return
	}

	if period == "" && limitStr == "" {
		writeJSONError(w, http.StatusBadRequest, "period or limit is required")
		return
	}

	var (
		items []mcpFeedItem
		err   error
		resp  mcpFeedsResponse
	)

	switch period {
	case "today", "yesterday":
		items, err = fetchUserFeedItemsByPeriod(userID, period)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load feeds")
			return
		}
		resp = mcpFeedsResponse{
			Items:       items,
			Count:       len(items),
			Period:      period,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}
	case "":
		limit, parseErr := strconv.Atoi(limitStr)
		if parseErr != nil || limit <= 0 {
			writeJSONError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if limit > maxMCPItems {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("limit must be <= %d", maxMCPItems))
			return
		}

		items, err = fetchUserFeedItemsByLimit(userID, limit)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load feeds")
			return
		}
		resp = mcpFeedsResponse{
			Items:       items,
			Count:       len(items),
			Limit:       limit,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}
	default:
		writeJSONError(w, http.StatusBadRequest, "invalid period: use today or yesterday")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func fetchUserFeedItemsByPeriod(userID int, period string) ([]mcpFeedItem, error) {
	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		return nil, err
	}

	if len(userFeeds) == 0 {
		return []mcpFeedItem{}, nil
	}

	start, end, err := dayRange(period)
	if err != nil {
		return nil, err
	}

	placeholders := strings.Repeat(",?", len(userFeeds))[1:]
	query := fmt.Sprintf(`SELECT title, link, date, source, description FROM feeds
        WHERE feed_url IN (%s) AND datetime(date) >= datetime(?) AND datetime(date) < datetime(?)
        ORDER BY datetime(date) DESC`, placeholders)

	args := make([]any, 0, len(userFeeds)+2)
	for _, u := range userFeeds {
		args = append(args, u)
	}
	args = append(args, start.Format(time.RFC3339), end.Format(time.RFC3339))

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []mcpFeedItem
	for rows.Next() {
		var item mcpFeedItem
		if err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func fetchUserFeedItemsByLimit(userID int, limit int) ([]mcpFeedItem, error) {
	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		return nil, err
	}

	if len(userFeeds) == 0 {
		return []mcpFeedItem{}, nil
	}

	placeholders := strings.Repeat(",?", len(userFeeds))[1:]
	query := fmt.Sprintf(`SELECT title, link, date, source, description FROM feeds
        WHERE feed_url IN (%s)
        ORDER BY datetime(date) DESC LIMIT ?`, placeholders)

	args := make([]any, 0, len(userFeeds)+1)
	for _, u := range userFeeds {
		args = append(args, u)
	}
	args = append(args, limit)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []mcpFeedItem
	for rows.Next() {
		var item mcpFeedItem
		if err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func dayRange(period string) (time.Time, time.Time, error) {
	now := time.Now().In(time.Local)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch period {
	case "today":
		return startOfToday, startOfToday.AddDate(0, 0, 1), nil
	case "yesterday":
		start := startOfToday.AddDate(0, 0, -1)
		return start, startOfToday, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, mcpErrorResponse{Error: message})
}
