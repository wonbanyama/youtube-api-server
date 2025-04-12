package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	server := GetRouteService(app)
	server.RegisterRoutes()
	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
