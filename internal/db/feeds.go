package db

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
)

func AddUserFeed(userId int, feedTitle, feedUrl, feedTags string) error {
	_, err := DB.Exec(
		`INSERT INTO user_feeds (user_id, feed_url, title, category) VALUES (?, ?, ?, ?)`,
		userId,
		feedUrl,
		feedTitle,
		feedTags,
	)
	if err != nil {
		return fmt.Errorf("failed to add feed url %s to %d feeds: %w", feedUrl, userId, err)
	}

	return nil
}

func RemoveUserFeed(userId int, feedId string) error {
	_, err := DB.Exec(`DELETE FROM user_feeds WHERE id = ? AND user_id = ?`, feedId, userId)
	if err != nil {
		return fmt.Errorf("failed to delete feed id %s for user id %d: %w", feedId, userId, err)
	}

	return nil
}

func UpdateUserFeed(userId int, feedId, feedTitle, feedTags string) error {
	_, err := DB.Exec(
		`UPDATE user_feeds SET title = ?, category = ? WHERE id = ? AND user_id = ?`,
		feedTitle,
		feedTags,
		feedId,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to update feed id %s for user id %d: %w", feedId, userId, err)
	}

	return nil
}

func GetTotalUserFeedItemsCount(userFeeds []string) (int, error) {
	var totalCount int

	placeholders := strings.Repeat(",?", len(userFeeds))[1:]

	query := fmt.Sprintf("SELECT COUNT(*) FROM feeds WHERE feed_url IN (%s)", placeholders)
	args := make([]interface{}, len(userFeeds))

	for i, u := range userFeeds {
		args[i] = u
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to query user feed items count: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&totalCount)
		if err != nil {
			return 0, fmt.Errorf("failed to parse user feed items count: %w", err)
		}
	}

	return totalCount, nil
}

func GetUserFeedItems(userID int, userFeeds []string, perPage, offset int) ([]models.FeedItem, error) {
	var items []models.FeedItem

	placeholders := strings.Repeat(",?", len(userFeeds))[1:]

	argsWithPagination := make([]any, 0, len(userFeeds)+3) // +3 for userID, perpage and offset
	argsWithPagination = append(argsWithPagination, userID)
	for _, u := range userFeeds {
		argsWithPagination = append(argsWithPagination, u)
	}

	argsWithPagination = append(argsWithPagination, perPage, offset)
	query := fmt.Sprintf(`SELECT feeds.title, feeds.link, feeds.date,
		COALESCE(NULLIF(user_feeds.title, ''), feeds.source) AS source,
		feeds.description
		FROM feeds
		JOIN user_feeds ON user_feeds.feed_url = feeds.feed_url AND user_feeds.user_id = ?
		WHERE feeds.feed_url IN (%s) ORDER BY date DESC LIMIT ? OFFSET ?`, placeholders)

	rows, err := DB.Query(query, argsWithPagination...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user feed items: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var item models.FeedItem

		err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description)
		if err != nil {
			slog.Error("failed to scan feed urls", "error", err)

			continue
		}

		item.Date = timeToHumanReadable(item.Date)

		items = append(items, item)
	}

	return items, nil
}

func timeToHumanReadable(t string) string {
	parsedTime, err := time.Parse(time.RFC3339, t)
	if err != nil {
		slog.Error("failed to parse time", "error", err)

		return t
	}

	return parsedTime.Format("2006-01-02 15:04:05")
}

func DeleteUserFeed(userFeedId string) error {
	feedID, err := strconv.Atoi(userFeedId)
	if err != nil {
		return fmt.Errorf("invalid feed ID for deletion: %w", err)

	}

	_, err = DB.Exec(`DELETE FROM user_feeds WHERE id = ?`, feedID)
	if err != nil {
		return fmt.Errorf("failed to delete user feed: %w", err)
	}

	return nil
}
