package main

// 필요한 패키지 임포트
import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"youtube-backend/server/channel"
	"youtube-backend/server/config"
	"youtube-backend/server/video"
)

// VideoStatsResponse는 YouTube API로부터 받아오는 비디오 통계 정보를 담는 구조체입니다.
type VideoStatsResponse struct {
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

// getVideoStats는 주어진 비디오 ID 목록에 대한 통계 정보를 가져와 출력합니다.
func getVideoStats(videoIDs []string) {
	baseURL := "https://www.googleapis.com/youtube/v3/videos"
	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("id", strings.Join(videoIDs, ","))
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("영상 정보 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	var result VideoStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("영상 정보 디코딩 실패: %v", err)
	}

	for _, item := range result.Items {
		fmt.Printf("🎬 %s\n   🔗 https://www.youtube.com/watch?v=%s\n   👀 조회수: %s\n\n",
			item.Snippet.Title, item.Id, item.Statistics.ViewCount)
	}
}

// getUploadPlaylistID는 채널 ID를 받아 해당 채널의 업로드 재생목록 ID를 반환합니다.
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

// getVideoIDsFromPlaylist는 재생목록 ID와 최대 개수를 받아 비디오 ID 목록을 반환합니다.
func getVideoIDsFromPlaylist(playlistID string, max int) []string {
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

	var ids []string
	for _, item := range result.Items {
		ids = append(ids, item.Snippet.ResourceId.VideoId)
	}
	return ids
}

// getRecentVideoIDsFromPlaylist는 특정 시간 내의 영상만 추출하여 반환합니다.
// hour 파라미터는 현재 시간으로부터 몇 시간 이내의 영상을 가져올지 지정합니다.
func getRecentVideoIDsFromPlaylist(playlistID string, max int, hour int) []video.Item {
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

	var recentVideos []video.Item
	now := time.Now()
	for _, item := range result.Items {
		if now.Sub(item.Snippet.PublishedAt) <= time.Duration(hour)*time.Hour {
			recentVideos = append(recentVideos, video.Item{
				ID:          item.Snippet.ResourceId.VideoId,
				Title:       item.Snippet.Title,
				PublishedAt: item.Snippet.PublishedAt,
			})
		}
	}

	return recentVideos
}

// getVideoStatsForRecent는 최근 업로드된 영상들의 통계 정보를 가져와 조회수 순으로 정렬하여 출력합니다.
func getVideoStatsForRecent(videos []video.Item) {
	if len(videos) == 0 {
		fmt.Println("📭 해당시간 내 영상이 없습니다.")
		return
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

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("영상 정보 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	var result VideoStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("영상 정보 디코딩 실패: %v", err)
	}
	loc, _ := time.LoadLocation("Asia/Seoul")

	// 📦 조회수와 함께 묶기
	var videosWithViews []video.WithViews
	for _, item := range result.Items {
		viewCount, err := strconv.Atoi(item.Statistics.ViewCount)
		if err != nil {
			viewCount = 0
		}
		videosWithViews = append(videosWithViews, video.WithViews{
			Item: video.Item{
				ID:          item.Id,
				Title:       item.Snippet.Title,
				PublishedAt: item.Snippet.PublishedAt,
			},
			ViewCount: viewCount,
		})
	}

	// 📊 조회수 내림차순 정렬
	sort.Slice(videosWithViews, func(i, j int) bool {
		return videosWithViews[i].ViewCount > videosWithViews[j].ViewCount
	})

	// 🏅 출력 + 순위 + 이모지
	rankEmojis := []string{"🥇", "🥈", "🥉"}
	for i, v := range videosWithViews {
		emoji := "⭐️"
		if i < len(rankEmojis) {
			emoji = rankEmojis[i]
		}
		publishedTime := v.PublishedAt.In(loc).Format("2006-01-02 15:04:05")
		fmt.Printf("%s %s\n   🔗 https://www.youtube.com/watch?v=%s\n   👀 조회수: %d\n   🕒 업로드: %s\n\n",
			emoji, v.Title, v.ID, v.ViewCount, publishedTime)
	}
}

// FindChannelID는 채널 이름으로 채널 ID를 검색하여 반환합니다.
func FindChannelID(channelName string) string {
	baseURL := "https://www.googleapis.com/youtube/v3/search"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "channel")
	params.Set("q", channelName)
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("채널 검색 실패: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Id struct {
				ChannelId string `json:"channelId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("채널 검색 결과 디코딩 실패: %v", err)
	}

	if len(result.Items) > 0 {
		return result.Items[0].Id.ChannelId
	}
	return ""
}

// main 함수는 프로그램의 진입점으로, 사용자 입력을 받아 YouTube 채널의 영상 정보를 분석합니다.
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("채널 이름을 입력하세요: ")
	scanner.Scan()
	channelName := scanner.Text()

	channelID := FindChannelID(channelName)
	if channelID == "" {
		log.Fatal("채널을 찾을 수 없습니다.")
	}

	playlistID := getUploadPlaylistID(channelID)
	recentVideos := getRecentVideoIDsFromPlaylist(playlistID, 50, 24) // 최근 24시간
	getVideoStatsForRecent(recentVideos)
}