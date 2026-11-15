package route

import (
	"foundnear/app/service"
	"foundnear/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterClaimRoute(
	app *fiber.App,
	claimService *service.ClaimService,
	jwtSecret string,
) {
	claims := app.Group(
		"/api/v1/claims",
		middleware.RequireAuth(jwtSecret),
	)

	// Pengajuan klaim kepemilikan barang.
	claims.Post("/", claimService.CreateClaim)

	// Melihat klaim milik pengguna.
	claims.Get("/my", claimService.ListMyClaims)

	// Melihat seluruh klaim (admin/petugas).
	claims.Get("/", claimService.ListClaims)

	// Melihat detail klaim.
	claims.Get("/:id", claimService.GetClaimDetail)

	// Membatalkan klaim.
	claims.Put("/:id/cancel", claimService.CancelClaim)
}
