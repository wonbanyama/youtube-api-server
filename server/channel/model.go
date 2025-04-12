package channel

import "time"

type YouTubeResponse struct {
	Items []struct {
		Snippet struct {
			ChannelId string `json:"channelId"`
			Title     string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

type BasicResponse struct {
	Items []struct {
		ContentDetails struct {
			RelatedPlaylists struct {
				Uploads string `json:"uploads"`
			} `json:"relatedPlaylists"`
		} `json:"contentDetails"`
	} `json:"items"`
}

type PlaylistResponse struct {
	Items []struct {
		Snippet struct {
			Title       string    `json:"title"`
			PublishedAt time.Time `json:"publishedAt"`
			ResourceId  struct {
				VideoId string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
	} `json:"items"`
}
