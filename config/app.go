package config

import (
	"foundnear/app/repository"
	"foundnear/app/service"
	"foundnear/middleware"
	"foundnear/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(db *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: GetEnv("APP_NAME", "FoundNear"),
	})

	// Security and request middleware
	app.Use(requestid.New())
	app.Use(middleware.Recovery())
	app.Use(cors.New())
	app.Use(helmet.New())
	app.Use(middleware.RequestLogger())

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Initialize services
	jwtSecret := GetEnv("JWT_SECRET", "")
	authService := service.NewAuthService(
		userRepo,
		roleRepo,
		jwtSecret,
	)

	// Register routes
	route.RegisterHealthRoute(app)
	route.RegisterAuthRoute(app, authService, jwtSecret)

	return app
}
