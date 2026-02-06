package http

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const userSettingsTemplate = "templates/user_settings"

func userSettingsRender(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	userFeeds, err := db.GetUserFeeds(userInfo.ID)
	if err != nil {
		log.Error("failed to get user feeds: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	userToken := ""
	token, err := db.GetUserToken(userInfo.ID)
	if err != nil {
		if !errors.Is(err, db.ErrTokenNotFound) {
			log.Error("failed to get user token: ", err)
			return c.Render(errorTemplate, defaultInternalErrorMap(nil))
		}
	} else {
		userToken = token
	}

	refreshInterval, err := db.GetUserRefreshInterval(userInfo.ID)
	if err != nil {
		log.Error("failed to get user refresh interval: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	lastUpdate, err := db.GetLastUpdateTS(userInfo.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("failed to get last update ts for %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	nextUpdate, err := db.GetNextUpdateTS(userInfo.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("failed to get next update ts for %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	luStr := lastUpdate.Format(time.DateTime)
	nuStr := nextUpdate.Format(time.DateTime)

	if lastUpdate.IsZero() {
		luStr = "Not performed yet"
	}

	if nextUpdate.IsZero() {
		nuStr = "Will be performed soon"
	}

	return c.Render(userSettingsTemplate, fiber.Map{
		"UserFeeds":       userFeeds,
		"User":            userInfo,
		"UserToken":       userToken,
		"Title":           "RapidFeed - Settings",
		"RefreshInterval": refreshInterval,
		"LastUpdate":      luStr,
		"NextUpdate":      nuStr,
	})
}

func changePasswordHandler(c *fiber.Ctx) error {
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")

	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	hash, err := db.GetUserHash(userInfo.Username)
	if err != nil {
		log.Error("failed to get user hash: ", err)
	}

	err = auth.CheckPassword(hash, currentPassword)
	if err != nil {
		log.Error("wrong current password")
		//TODO: add alert on settings page like in login page, to clearly show where user was wrong
		return c.Redirect("/settings", http.StatusConflict)
	}

	err = db.ChangeUserPassword(userInfo.ID, newPassword)
	if err != nil {
		log.Error("failed to change user password: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings#change-password", http.StatusFound)
}

func addFeedHandler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	feedUrl := c.FormValue("feed_url")
	feedTitle := strings.TrimSpace(c.FormValue("feed_title"))
	feedTags := normalizeTags(c.FormValue("feed_tags"))

	feeds, err := db.GetUserFeeds(userInfo.ID)
	if err != nil {
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	for _, feed := range feeds {
		if feed.FeedURL == feedUrl {
			log.Warnf("feed %s already exists in %s feeds", feedUrl, userInfo.Username)

			return c.Redirect("/settings#manage-feeds", http.StatusFound)
		}
	}

	if feedTitle == "" {
		feedTitle = feeder.ExtractSourceFromURL(feedUrl)
	}

	err = db.AddUserFeed(userInfo.ID, feedTitle, feedUrl, feedTags)
	if err != nil {
		log.Errorf("failed to add %s to %s feeds: %v", feedUrl, userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	feedUrls, err := db.GetUserFeedUrls(userInfo.ID)
	if err != nil {
		log.Errorf("failed to get new %s feeds list: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	feeder.FetchAndSaveFeeds(feedUrls)

	return c.Redirect("/settings#manage-feeds", http.StatusFound)
}

func normalizeTags(rawTags string) string {
	parts := strings.Split(rawTags, ",")
	seen := make(map[string]struct{})
	cleaned := make([]string, 0, len(parts))

	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}

		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		cleaned = append(cleaned, tag)
	}

	return strings.Join(cleaned, ", ")
}

func removeFeedHandler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	feedId := c.FormValue("feed_id")

	err = db.RemoveUserFeed(userInfo.ID, feedId)
	if err != nil {
		log.Error("failed to remove user %s feed by id: ", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings#manage-feeds", http.StatusFound)
}

func autorefreshIntervalChangeHadler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	intervalStr := c.FormValue("refresh_interval")

	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 0 {
		log.Errorf("failed to parse autorefresh interval, username %s, err %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	err = db.SetUserRefreshInterval(userInfo.ID, interval)
	if err != nil {
		log.Errorf("failed to set %s refresh interval: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings#autorefresh", http.StatusFound)
}

func addUserTokenHandler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	token, err := auth.GenerateToken(32)
	if err != nil {
		log.Error("failed to generate token: ", err)
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	if err := db.UpsertUserToken(userInfo.ID, token); err != nil {
		log.Error("failed to store user token: ", err)
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings#api-token", http.StatusFound)
}

func revokeUserTokenHandler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	if err := db.DeleteUserToken(userInfo.ID); err != nil {
		log.Error("failed to delete user token: ", err)
		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings#api-token", http.StatusFound)
}
