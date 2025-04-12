package controller

import (
	"github.com/gofiber/fiber/v2"
	"youtube-backend/server/channel"
)

func GetChannelID(c *fiber.Ctx) error {
	var result fiber.Map

	svc := channel.GetService()
	channelName := c.Query("channelName")
	channelID, err := svc.FindChannelID(channelName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "채널 정보를 찾을 수 없습니다.",
		})
	}

	result = fiber.Map{
		"channelID": channelID,
	}

	return c.JSON(result)
}
