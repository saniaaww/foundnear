package config

import (
	"github.com/gofiber/fiber/v2"
	"foundnear/route"
)

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: GetEnv("APP_NAME", "FoundNear"),
	})

	route.RegisterHealthRoute(app)

	return app
}