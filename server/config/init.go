package config

import (
	"github.com/joho/godotenv"
	"os"
)

var APIKey string // 전역 변수 선언

func init() {
	_ = godotenv.Load()

	// 전역 변수 초기화
	APIKey = os.Getenv("YOUTUBE_API_KEY")
}
