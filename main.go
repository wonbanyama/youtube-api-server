package main

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

// 특정시간 내 영상만 추출
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
		kstTime := v.Item.PublishedAt.In(loc)
		fmt.Printf("%s %d위\n🆕 %s\n📅 %s\n👀 조회수: %d\n🔗 https://www.youtube.com/watch?v=%s\n\n",
			emoji, i+1, v.Item.Title, kstTime.Format("2006-01-02 15:04"), v.ViewCount, v.Item.ID)
	}
}

func FindChannelID(channelName string) string {
	baseURL := "https://www.googleapis.com/youtube/v3/search"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "channel")
	params.Set("q", channelName)
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("채널 검색 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	var result channel.YouTubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("채널 검색 응답 디코딩 실패: %v", err)
	}

	if len(result.Items) > 0 {
		return result.Items[0].Snippet.ChannelId
	}
	log.Fatal("채널 정보를 찾을 수 없습니다.")
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("채널명을 입력하세요 (종료하려면 'exit'): ")
		scanner.Scan()
		channelName := scanner.Text()

		if strings.ToLower(channelName) == "exit" {
			break
		}

		fmt.Print("몇 개의 영상을 가져올까요?: ")
		scanner.Scan()
		countStr := scanner.Text()

		videoCount, err := strconv.Atoi(countStr)
		if err != nil {
			fmt.Println("숫자로 입력해주세요.")
			continue
		}

		fmt.Print("몇시간 이내의 영상을 가져올까요?: ")
		scanner.Scan()
		hourStr := scanner.Text()

		hour, err := strconv.Atoi(hourStr)
		if err != nil {
			fmt.Println("숫자로 입력해주세요.")
			continue
		}

		// 여기에 실제 로직 호출
		channelID := FindChannelID(channelName)
		playlistID := getUploadPlaylistID(channelID)
		recentVideos := getRecentVideoIDsFromPlaylist(playlistID, videoCount, hour)
		getVideoStatsForRecent(recentVideos)

		fmt.Println("--- 완료 ---\n")
	}
}
