
package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	appmiddleware "foundnear/middleware"
	"foundnear/route"
)

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: GetEnv("APP_NAME", "FoundNear"),
	})

	// Security and request middleware
	app.Use(requestid.New())
	app.Use(appmiddleware.Recovery())
	app.Use(cors.New())
	app.Use(helmet.New())
	app.Use(appmiddleware.RequestLogger())

	// Register routes
	route.RegisterHealthRoute(app)

	return app
}
