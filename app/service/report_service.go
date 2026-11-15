package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"foundnear/app/model"
	"foundnear/app/repository"

	"github.com/gofiber/fiber/v2"
)

type ReportService struct {
	reportRepo   repository.ReportRepository
	categoryRepo repository.CategoryRepository
}

func NewReportService(
	reportRepo repository.ReportRepository,
	categoryRepo repository.CategoryRepository,
) *ReportService {
	return &ReportService{
		reportRepo:   reportRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateReport membuat laporan barang hilang atau ditemukan.
func (s *ReportService) CreateReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	var req model.CreateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	req.ReportType = strings.ToLower(strings.TrimSpace(req.ReportType))
	req.ItemName = strings.TrimSpace(req.ItemName)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)

	if req.ReportType != model.ReportTypeLost &&
		req.ReportType != model.ReportTypeFound {
		return errorResponse(c, fiber.StatusBadRequest, "Report type must be lost or found")
	}

	if req.CategoryID <= 0 || req.ItemName == "" || req.Location == "" ||
		req.IncidentDate == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Required fields are missing")
	}

	if len(req.ItemName) > 150 || len(req.Location) > 255 {
		return errorResponse(c, fiber.StatusBadRequest, "Item name or location is too long")
	}

	incidentDate, err := time.Parse("2006-01-02", req.IncidentDate)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Incident date must use YYYY-MM-DD format")
	}

	if _, err := s.categoryRepo.GetCategoryByID(c.UserContext(), req.CategoryID); err != nil {
		return handleRepositoryError(c, err, "Failed to validate category")
	}

	report := &model.ItemReport{
		UserID:       userID,
		CategoryID:   req.CategoryID,
		ReportType:   req.ReportType,
		ItemName:     req.ItemName,
		Description:  req.Description,
		PrivateInfo:  cleanOptionalString(req.PrivateInfo),
		Location:     req.Location,
		IncidentDate: incidentDate,
		ImageURL:     cleanOptionalString(req.ImageURL),
		Status:       model.ReportStatusActive,
	}

	if err := s.reportRepo.CreateReport(c.UserContext(), report); err != nil {
		return handleRepositoryError(c, err, "Failed to create report")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Report created successfully",
		"data":    report,
	})
}

// ListPublicReports menampilkan daftar laporan publik.
func (s *ReportService) ListPublicReports(c *fiber.Ctx) error {
	filter := model.ReportFilter{
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		ReportType: strings.ToLower(strings.TrimSpace(c.Query("report_type"))),
		Status:     strings.ToLower(strings.TrimSpace(c.Query("status"))),
		Location:   strings.TrimSpace(c.Query("location")),
		Limit:      c.QueryInt("limit", 10),
		Offset:     c.QueryInt("offset", 0),
	}

	if filter.ReportType != "" &&
		filter.ReportType != model.ReportTypeLost &&
		filter.ReportType != model.ReportTypeFound {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report type")
	}

	if filter.Status != "" &&
		filter.Status != model.ReportStatusActive &&
		filter.Status != model.ReportStatusMatched &&
		filter.Status != model.ReportStatusReturned &&
		filter.Status != model.ReportStatusClosed {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report status")
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil || id <= 0 {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid category ID")
		}
		filter.CategoryID = &id
	}

	reports, total, err := s.reportRepo.ListPublicReports(c.UserContext(), filter)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve reports")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Reports retrieved successfully",
		"data":    reports,
		"meta": fiber.Map{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// GetPublicReport menampilkan detail laporan tanpa private information.
func (s *ReportService) GetPublicReport(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report ID")
	}

	report, err := s.reportRepo.GetPublicReportByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve report")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Report retrieved successfully",
		"data":    report,
	})
}

// ListMyReports menampilkan laporan milik pengguna yang sedang login.
func (s *ReportService) ListMyReports(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	reports, err := s.reportRepo.ListReportsByUser(
		c.UserContext(), userID, limit, offset,
	)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve your reports")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Your reports retrieved successfully",
		"data":    reports,
	})
}

// UpdateReport mengubah laporan milik pengguna atau laporan oleh admin.
func (s *ReportService) UpdateReport(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report ID")
	}

	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid user authentication")
	}

	report, err := s.reportRepo.GetReportByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve report")
	}

	if !canManageReport(c, report.UserID) {
		return errorResponse(c, fiber.StatusForbidden, "You cannot update this report")
	}

	var req model.UpdateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	req.ItemName = strings.TrimSpace(req.ItemName)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)

	if req.CategoryID <= 0 || req.ItemName == "" ||
		req.Location == "" || req.IncidentDate == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Required fields are missing")
	}

	if len(req.ItemName) > 150 || len(req.Location) > 255 {
		return errorResponse(c, fiber.StatusBadRequest, "Item name or location is too long")
	}

	incidentDate, err := time.Parse("2006-01-02", req.IncidentDate)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Incident date must use YYYY-MM-DD format")
	}

	if _, err := s.categoryRepo.GetCategoryByID(c.UserContext(), req.CategoryID); err != nil {
		return handleRepositoryError(c, err, "Failed to validate category")
	}

	report.CategoryID = req.CategoryID
	report.ItemName = req.ItemName
	report.Description = req.Description
	report.PrivateInfo = cleanOptionalString(req.PrivateInfo)
	report.Location = req.Location
	report.IncidentDate = incidentDate
	report.ImageURL = cleanOptionalString(req.ImageURL)

	if err := s.reportRepo.UpdateReport(c.UserContext(), report); err != nil {
		return handleRepositoryError(c, err, "Failed to update report")
	}

	updated, err := s.reportRepo.GetReportByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Report updated but failed to retrieve it")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Report updated successfully",
		"data":    updated,
	})
}

// DeleteReport menghapus laporan milik pengguna atau laporan oleh admin.
func (s *ReportService) DeleteReport(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid report ID")
	}

	report, err := s.reportRepo.GetReportByID(c.UserContext(), id)
	if err != nil {
		return handleRepositoryError(c, err, "Failed to retrieve report")
	}

	if !canManageReport(c, report.UserID) {
		return errorResponse(c, fiber.StatusForbidden, "You cannot delete this report")
	}

	if err := s.reportRepo.DeleteReport(c.UserContext(), id); err != nil {
		return handleRepositoryError(c, err, "Failed to delete report")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Report deleted successfully",
	})
}

func parseID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}

func canManageReport(c *fiber.Ctx, ownerID int64) bool {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return false
	}

	role, _ := c.Locals("role").(string)
	return userID == ownerID || role == model.RoleAdmin
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func errorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

func handleRepositoryError(c *fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return errorResponse(c, fiber.StatusNotFound, "Data not found")
	case errors.Is(err, model.ErrDuplicate):
		return errorResponse(c, fiber.StatusConflict, "Data already exists")
	case errors.Is(err, model.ErrForeignKey):
		return errorResponse(c, fiber.StatusBadRequest, "Invalid reference or data still in use")
	case errors.Is(err, model.ErrConstraint):
		return errorResponse(c, fiber.StatusBadRequest, "Data violates a constraint")
	case errors.Is(err, model.ErrBusinessRule),
		errors.Is(err, model.ErrInvalidState):
		return errorResponse(c, fiber.StatusConflict, "Operation violates a business rule")
	default:
		return errorResponse(c, fiber.StatusInternalServerError, fallback)
	}
}
