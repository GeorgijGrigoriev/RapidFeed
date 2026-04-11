package mcp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	gomcp "github.com/localrivet/gomcp/server"
)

const maxMCPItems = 1000

type tokenArgs struct {
	Token *string `json:"token" description:"User MCP access token (optional if provided via X-MCP-Token header)"`
}

type limitArgs struct {
	Token *string `json:"token" description:"User MCP access token (optional if provided via X-MCP-Token header)"`
	Limit int     `json:"limit" description:"Max number of items to return" required:"true"`
}

type feedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Date        string `json:"date"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type feedResponse struct {
	Items       []feedItem `json:"items"`
	Count       int        `json:"count"`
	Period      string     `json:"period,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	GeneratedAt string     `json:"generated_at"`
}

func Start(addr string) error {
	srv := gomcp.NewServer("rapidfeed-mcp")

	registerTools(srv)

	transport := newHTTPTransport(addr)
	srv.GetServer().SetTransport(transport)

	return srv.Run()
}

func registerTools(srv gomcp.Server) {
	srv.Tool("feeds_today", "Return all posts for today from the user's feeds.", func(ctx *gomcp.Context, args *tokenArgs) (interface{}, error) {
		_ = ctx
		token := tokenFromArgs(args)
		userID, err := userIDFromToken(token)
		if err != nil {
			return nil, err
		}

		items, err := fetchUserFeedItemsByPeriod(userID, "today")
		if err != nil {
			return nil, err
		}

		return feedResponse{
			Items:       items,
			Count:       len(items),
			Period:      "today",
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil
	})

	srv.Tool("feeds_yesterday", "Return all posts from yesterday from the user's feeds.", func(ctx *gomcp.Context, args *tokenArgs) (interface{}, error) {
		_ = ctx
		token := tokenFromArgs(args)
		userID, err := userIDFromToken(token)
		if err != nil {
			return nil, err
		}

		items, err := fetchUserFeedItemsByPeriod(userID, "yesterday")
		if err != nil {
			return nil, err
		}

		return feedResponse{
			Items:       items,
			Count:       len(items),
			Period:      "yesterday",
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil
	})

	srv.Tool("feeds_latest", "Return the latest N posts from the user's feeds.", func(ctx *gomcp.Context, args *limitArgs) (interface{}, error) {
		_ = ctx
		if args.Limit <= 0 {
			return nil, fmt.Errorf("limit must be a positive integer")
		}
		if args.Limit > maxMCPItems {
			return nil, fmt.Errorf("limit must be <= %d", maxMCPItems)
		}

		token := tokenFromArgs(args)
		userID, err := userIDFromToken(token)
		if err != nil {
			return nil, err
		}

		items, err := fetchUserFeedItemsByLimit(userID, args.Limit)
		if err != nil {
			return nil, err
		}

		return feedResponse{
			Items:       items,
			Count:       len(items),
			Limit:       args.Limit,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil
	})
}

func tokenFromArgs(args interface{}) string {
	switch v := args.(type) {
	case *tokenArgs:
		if v != nil && v.Token != nil {
			return strings.TrimSpace(*v.Token)
		}
	case *limitArgs:
		if v != nil && v.Token != nil {
			return strings.TrimSpace(*v.Token)
		}
	}

	return ""
}

func userIDFromToken(token string) (int, error) {
	if strings.TrimSpace(token) == "" {
		return 0, fmt.Errorf("token is required (use X-MCP-Token header or tool argument)")
	}

	userID, err := db.GetUserIDByToken(token)
	if err != nil {
		if errors.Is(err, db.ErrTokenNotFound) {
			return 0, fmt.Errorf("unauthorized: invalid token")
		}
		return 0, err
	}

	role, err := db.GetUserRole(userID)
	if err != nil {
		return 0, err
	}
	if role == "blocked" {
		return 0, fmt.Errorf("forbidden: user is blocked")
	}

	return userID, nil
}

func fetchUserFeedItemsByPeriod(userID int, period string) (items []feedItem, err error) {
	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		return nil, err
	}

	if len(userFeeds) == 0 {
		return []feedItem{}, nil
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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	items = make([]feedItem, 0)
	for rows.Next() {
		var item feedItem
		if err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func fetchUserFeedItemsByLimit(userID int, limit int) (items []feedItem, err error) {
	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		return nil, err
	}

	if len(userFeeds) == 0 {
		return []feedItem{}, nil
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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	items = make([]feedItem, 0)
	for rows.Next() {
		var item feedItem
		if err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
