package video

import "time"

type Item struct {
	ID          string
	Title       string
	PublishedAt time.Time
}

type WithViews struct {
	Item      Item
	ViewCount int
	Rank      int
	UploadAt  time.Time
}

type StatsResponse struct {
	Items []struct {
		Id      string `json:"id"`
		Snippet struct {
			Title       string    `json:"title"`
			PublishedAt time.Time `json:"publishedAt"`
		} `json:"snippet"`
		Statistics struct {
			ViewCount string `json:"viewCount"`
			LikeCount string `json:"likeCount,omitempty"`
			// 추가 필드 가능: CommentCount, FavoriteCount 등
		} `json:"statistics"`
	} `json:"items"`
}

type RecentVideosResponse struct {
	WithViews []WithViews `json:"with_views"`
	Rank      int         `json:"rank"`
}
