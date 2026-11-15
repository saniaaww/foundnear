package route

import (
	"foundnear/app/service"
	"foundnear/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoute(
	app *fiber.App,
	authService *service.AuthService,
	jwtSecret string,
) {
	auth := app.Group("/api/v1/auth")

	auth.Post("/register", authService.Register)
	auth.Post("/login", authService.Login)

	auth.Get(
		"/me",
		middleware.RequireAuth(jwtSecret),
		authService.Me,
	)
}
