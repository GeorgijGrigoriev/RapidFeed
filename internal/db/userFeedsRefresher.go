package db

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"
)

const defaultRefreshInterval = 60 // in minutes

// GetUserRefreshInterval - returns user refresh feeds interval, if no interval set - return default
func GetUserRefreshInterval(userID int) (int, error) {
	var interval int

	err := DB.QueryRow("SELECT interval_minutes FROM user_refresh_settings WHERE user_id = ?", userID).Scan(&interval)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("no value is set for update interval, so returning default", "user id", userID)

			return defaultRefreshInterval, nil
		}

		slog.Error("failed to get user refresh interval", "error", err)

		return 0, err
	}

	return interval, nil
}

// SetUserRefreshInterval - sets the refresh interval for a user
func SetUserRefreshInterval(userID int, intervalMinutes int) error {
	_, err := DB.Exec(`
              INSERT OR REPLACE INTO user_refresh_settings (user_id, interval_minutes)
              VALUES (?, ?)`,
		userID, intervalMinutes)
	if err != nil {
		slog.Error("failed to set user refresh interval", "error", err)

		return err
	}

	return nil
}

// SetLastUpdateTS - sets when last fetch is performed
func SetLastUpdateTS(userID int, interval int) error {
	now := time.Now()
	next := now.Add(time.Duration(interval) * time.Minute).Unix()

	_, err := DB.Exec(`INSERT OR REPLACE INTO user_refresh_settings 
	(user_id, last_update_ts, next_update_ts, interval_minutes) VALUES (?, ?, ?, ?)`,
		userID, now.Unix(), next, interval)
	if err != nil {
		slog.Error("failed to update last_update_ts in user_refresh_settings", "error", err)

		return err
	}

	return nil
}

// GetLastUpdateTS - return ts when last fetch is performed
func GetNextUpdateTS(userID int) (time.Time, error) {
	var timestampStr sql.NullString

	query := "SELECT next_update_ts FROM user_refresh_settings WHERE user_id = ?"

	err := DB.QueryRow(query, userID).Scan(&timestampStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	if !timestampStr.Valid {
		return time.Time{}, nil
	}

	ts, err := strconv.ParseInt(timestampStr.String, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	timestamp := time.Unix(ts, 0)

	return timestamp, nil
}

// GetLastUpdateTS - returns ts when last update was performed
func GetLastUpdateTS(userID int) (time.Time, error) {
	var timestampStr sql.NullString

	query := "SELECT last_update_ts FROM user_refresh_settings WHERE user_id = ?"

	err := DB.QueryRow(query, userID).Scan(&timestampStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	if !timestampStr.Valid {
		return time.Time{}, nil
	}

	ts, err := strconv.ParseInt(timestampStr.String, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	timestamp := time.Unix(ts, 0)

	return timestamp, nil
}
