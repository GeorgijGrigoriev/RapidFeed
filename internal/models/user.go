package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserFeed struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	FeedURL string `json:"feed_url"`
}

type UserWithFeeds struct {
	User      User
	UserFeeds []UserFeed
}
