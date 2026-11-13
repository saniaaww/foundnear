package route

import "github.com/gofiber/fiber/v2"

func RegisterHealthRoute(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "FoundNear API is running",
		})
	})
}