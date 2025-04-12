package video

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"youtube-backend/server/channel"
	"youtube-backend/server/config"
)

type Service struct{}

func GetService() *Service {
	return &Service{}
}

func (s *Service) GetVideoStatsForRecent(channelID, count, hour string) ([]WithViews, error) {
	countInt, err := strconv.Atoi(count)
	if err != nil {
		return nil, err
	}
	hourInt, err := strconv.Atoi(hour)
	if err != nil {
		return nil, err
	}

	playlistID := getUploadPlaylistID(channelID)
	recentVideos := getRecentVideoIDsFromPlaylist(playlistID, countInt, hourInt)
	return getVideoStatsForRecent(recentVideos)
}

func getUploadPlaylistID(channelID string) string {
	baseURL := "https://www.googleapis.com/youtube/v3/channels"
	params := url.Values{}
	params.Set("part", "contentDetails")
	params.Set("id", channelID)
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("채널 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	var result channel.BasicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("채널 응답 디코딩 실패: %v", err)
	}

	if len(result.Items) > 0 {
		return result.Items[0].ContentDetails.RelatedPlaylists.Uploads
	}
	log.Fatal("채널 정보를 찾을 수 없습니다.")
	return ""
}

func getRecentVideoIDsFromPlaylist(playlistID string, max int, hour int) []Item {
	baseURL := "https://www.googleapis.com/youtube/v3/playlistItems"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("playlistId", playlistID)
	params.Set("maxResults", fmt.Sprintf("%d", max))
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("영상 목록 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	var result channel.PlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("영상 목록 디코딩 실패: %v", err)
	}

	var recentVideos []Item
	now := time.Now()
	for _, item := range result.Items {
		if now.Sub(item.Snippet.PublishedAt) <= time.Duration(hour)*time.Hour {
			recentVideos = append(recentVideos, Item{
				ID:          item.Snippet.ResourceId.VideoId,
				Title:       item.Snippet.Title,
				PublishedAt: item.Snippet.PublishedAt,
			})
		}
	}

	return recentVideos
}

func getVideoStatsForRecent(videos []Item) ([]WithViews, error) {
	var videosWithViews []WithViews
	if len(videos) == 0 {
		return nil, fmt.Errorf("📭 해당시간 내 영상이 없습니다.")
	}

	var ids []string
	for _, v := range videos {
		ids = append(ids, v.ID)
	}

	baseURL := "https://www.googleapis.com/youtube/v3/videos"
	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("id", strings.Join(ids, ","))
	params.Set("key", config.APIKey)

	resp, apiErr := http.Get(baseURL + "?" + params.Encode())
	if apiErr != nil {
		log.Fatalf("영상 정보 요청 실패: %v", apiErr)
		return nil, apiErr
	}
	defer resp.Body.Close()

	var result StatsResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		log.Fatalf("영상 정보 디코딩 실패: %v", decodeErr)
		return nil, decodeErr
	}
	loc, _ := time.LoadLocation("Asia/Seoul")

	// 📦 조회수와 함께 묶기
	for _, item := range result.Items {
		viewCount, err := strconv.Atoi(item.Statistics.ViewCount)
		if err != nil {
			viewCount = 0
		}
		videosWithViews = append(videosWithViews, WithViews{
			Item: Item{
				ID:          item.Id,
				Title:       item.Snippet.Title,
				PublishedAt: item.Snippet.PublishedAt,
			},
			ViewCount: viewCount,
			UploadAt:  item.Snippet.PublishedAt.In(loc),
		})
	}

	// 📊 조회수 내림차순 정렬
	sort.Slice(videosWithViews, func(i, j int) bool {
		return videosWithViews[i].ViewCount > videosWithViews[j].ViewCount
	})

	for i := range videosWithViews {
		videosWithViews[i].Rank = i + 1
	}
	return videosWithViews, nil
}
