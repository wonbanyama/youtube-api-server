package channel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"youtube-backend/server/config"
)

type Service struct{}

func GetService() *Service {
	return &Service{}
}

func (s *Service) FindChannelID(channelName string) (string, error) {
	baseURL := "https://www.googleapis.com/youtube/v3/search"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "channel")
	params.Set("q", channelName)
	params.Set("key", config.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Fatalf("채널 검색 요청 실패: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	var result YouTubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("채널 검색 응답 디코딩 실패: %v", err)
	}

	if len(result.Items) > 0 {
		return result.Items[0].Snippet.ChannelId, nil
	}
	return "", fmt.Errorf("채널 정보를 찾을 수 없습니다.")
}
