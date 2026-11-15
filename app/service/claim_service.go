package service

import (
	"strings"

	"foundnear/app/model"
	"foundnear/app/repository"

	"github.com/gofiber/fiber/v2"
)

type ClaimService struct {
	claimRepo  repository.ClaimRepository
	reportRepo repository.ReportRepository
}

func NewClaimService(
	claimRepo repository.ClaimRepository,
	reportRepo repository.ReportRepository,
) *ClaimService {
	return &ClaimService{
		claimRepo:  claimRepo,
		reportRepo: reportRepo,
	}
}

// CreateClaim mengajukan klaim kepemilikan barang.
func (s *ClaimService) CreateClaim(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	var req struct {
		ReportID         int64  `json:"report_id"`
		ClaimDescription string `json:"claim_description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	req.ClaimDescription = strings.TrimSpace(req.ClaimDescription)

	if req.ReportID <= 0 || req.ClaimDescription == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Report ID and claim description are required")
	}

	if len(req.ClaimDescription) > 2000 {
		return errorResponse(c, fiber.StatusBadRequest, "Claim description is too long")
	}

	report, err := s.reportRepo.GetReportByID(c.UserContext(), req.ReportID)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve report")
	}

	if report.Status != model.ReportStatusActive {
		return errorResponse(c, fiber.StatusConflict, "This report is not available for claiming")
	}

	if report.UserID == userID {
		return errorResponse(c, fiber.StatusForbidden, "You cannot claim your own report")
	}

	claim := &model.ItemClaim{
		ReportID:         req.ReportID,
		ClaimantID:       userID,
		ClaimDescription: req.ClaimDescription,
		Status:           model.ClaimStatusPending,
	}

	if err := s.claimRepo.CreateClaim(c.UserContext(), claim); err != nil {
		return handleRepositoryError(c, err, "Failed to create claim")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Claim submitted successfully",
		"data":    claim,
	})
}

// ListMyClaims menampilkan klaim milik pengguna.
func (s *ClaimService) ListMyClaims(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	claims, err := s.claimRepo.ListClaimsByClaimant(
		c.UserContext(), userID, limit, offset,
	)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve your claims")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Your claims retrieved successfully",
		"data":    claims,
	})
}

// ListClaims menampilkan seluruh klaim untuk admin atau petugas.
func (s *ClaimService) ListClaims(c *fiber.Ctx) error {
	if !isClaimOfficer(c) {
		return errorResponse(c, fiber.StatusForbidden, "You do not have permission to access claims")
	}

	status := strings.ToLower(strings.TrimSpace(c.Query("status")))

	if status != "" &&
		status != model.ClaimStatusPending &&
		status != model.ClaimStatusApproved &&
		status != model.ClaimStatusRejected &&
		status != model.ClaimStatusCancelled {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid claim status")
	}

	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	claims, err := s.claimRepo.ListClaims(
		c.UserContext(), status, limit, offset,
	)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve claims")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Claims retrieved successfully",
		"data":    claims,
	})
}

// GetClaimDetail menampilkan detail klaim.
func (s *ClaimService) GetClaimDetail(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid claim ID")
	}

	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	claim, err := s.claimRepo.GetClaimByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve claim")
	}

	if claim.ClaimantID != userID && !isClaimOfficer(c) {
		return errorResponse(c, fiber.StatusForbidden, "You cannot access this claim")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Claim retrieved successfully",
		"data":    claim,
	})
}

// CancelClaim membatalkan klaim milik pengguna.
func (s *ClaimService) CancelClaim(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid claim ID")
	}

	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	claim, err := s.claimRepo.GetClaimByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve claim")
	}

	if claim.ClaimantID != userID {
		return errorResponse(c, fiber.StatusForbidden, "You cannot cancel this claim")
	}

	if claim.Status != model.ClaimStatusPending {
		return errorResponse(c, fiber.StatusConflict, "Only pending claims can be cancelled")
	}

	if err := s.claimRepo.UpdateClaimStatus(
		c.UserContext(), id, model.ClaimStatusCancelled,
	); err != nil {
		return handleRepositoryError(c, err, "Failed to cancel claim")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Claim cancelled successfully",
	})
}

// isClaimOfficer memeriksa hak akses admin atau petugas.
func isClaimOfficer(c *fiber.Ctx) bool {
	role, _ := c.Locals("role").(string)

	return strings.EqualFold(role, model.RoleAdmin) ||
		strings.EqualFold(role, model.RoleOfficer)
}
