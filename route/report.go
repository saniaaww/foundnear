package route

import (
	"foundnear/app/service"
	"foundnear/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterReportRoute(
	app *fiber.App,
	reportService *service.ReportService,
	jwtSecret string,
) {
	reports := app.Group("/api/v1/reports")

	// Endpoint publik.
	reports.Get("/", reportService.ListPublicReports)

	// Endpoint yang membutuhkan autentikasi.
	auth := reports.Group("", middleware.RequireAuth(jwtSecret))

	auth.Post("/", reportService.CreateReport)
	auth.Get("/my", reportService.ListMyReports)
	auth.Put("/:id", reportService.UpdateReport)
	auth.Delete("/:id", reportService.DeleteReport)

	// Detail publik diletakkan setelah endpoint /my.
	reports.Get("/:id", reportService.GetPublicReport)
}
