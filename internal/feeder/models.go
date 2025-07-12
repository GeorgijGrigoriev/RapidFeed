package feeder

type FeedItem struct {
	Title       string
	Link        string
	Date        string
	Source      string
	Description string
}

type PaginatedFeedItems struct {
	Items      []FeedItem
	Page       int
	PerPage    int
	TotalPages int
	TotalItems int
}
