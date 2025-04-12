package controller

import (
	"github.com/gofiber/fiber/v2"
	"youtube-backend/server/video"
)

func GetFilteredVideos(c *fiber.Ctx) error {
	var result fiber.Map

	svc := video.GetService()
	channelID := c.Query("channelID")
	count := c.Query("count")
	hour := c.Query("hour")
	videos, err := svc.GetVideoStatsForRecent(channelID, count, hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "영상 정보를 찾을 수 없습니다.",
		})
	}

	result = fiber.Map{
		"videos": videos,
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
