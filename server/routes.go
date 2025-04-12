package main

import (
	"github.com/gofiber/fiber/v2"
	"youtube-backend/server/controller"
)

type RouteService struct {
	ServerApp *fiber.App
}

func GetRouteService(app *fiber.App) *RouteService {
	return &RouteService{
		ServerApp: app,
	}
}

func (r *RouteService) RegisterRoutes() {
	r.ServerApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Server!")
	})
	r.ServerApp.Get("/channel", controller.GetChannelID)
	r.ServerApp.Get("/video", controller.GetFilteredVideos)
}
