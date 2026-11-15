package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// ReportRepository: laporan barang hilang/ditemukan (lost & found).
type ReportRepository interface {
	CreateReport(ctx context.Context, report *model.ItemReport) error
	GetReportByID(ctx context.Context, id int64) (*model.ItemReport, error)
	GetPublicReportByID(ctx context.Context, id int64) (*model.PublicItemReport, error)
	ListPublicReports(ctx context.Context, filter model.ReportFilter) ([]model.PublicItemReport, int64, error)
	ListReportsByUser(ctx context.Context, userID int64, limit, offset int) ([]model.ItemReport, error)
	UpdateReport(ctx context.Context, report *model.ItemReport) error
	UpdateReportStatus(ctx context.Context, id int64, status string) error
	DeleteReport(ctx context.Context, id int64) error
}

type postgresReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) ReportRepository {
	return &postgresReportRepository{db: db}
}

const reportColumns = `id, user_id, category_id, report_type, item_name, description, private_info,
	location, incident_date, image_url, status, created_at, updated_at`

func scanReport(s rowScanner, r *model.ItemReport) error {
	return s.Scan(&r.ID, &r.UserID, &r.CategoryID, &r.ReportType, &r.ItemName, &r.Description,
		&r.PrivateInfo, &r.Location, &r.IncidentDate, &r.ImageURL, &r.Status, &r.CreatedAt, &r.UpdatedAt)
}

const publicReportSelect = `
	SELECT r.id, r.category_id, c.name, r.report_type, r.item_name, r.description,
	       r.location, r.incident_date, r.image_url, r.status, r.created_at
	FROM item_reports r
	JOIN categories c ON c.id = r.category_id`

func scanPublicReport(s rowScanner, p *model.PublicItemReport) error {
	return s.Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.ReportType, &p.ItemName, &p.Description,
		&p.Location, &p.IncidentDate, &p.ImageURL, &p.Status, &p.CreatedAt)
}

func (r *postgresReportRepository) CreateReport(ctx context.Context, rp *model.ItemReport) error {
	if rp.Status == "" {
		rp.Status = model.ReportStatusActive
	}
	query := `
		INSERT INTO item_reports
			(user_id, category_id, report_type, item_name, description, private_info, location, incident_date, image_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, rp.UserID, rp.CategoryID, rp.ReportType, rp.ItemName, rp.Description,
		rp.PrivateInfo, rp.Location, rp.IncidentDate, rp.ImageURL, rp.Status).
		Scan(&rp.ID, &rp.CreatedAt, &rp.UpdatedAt)
	return mapPgError(err)
}

func (r *postgresReportRepository) GetReportByID(ctx context.Context, id int64) (*model.ItemReport, error) {
	var rp model.ItemReport
	if err := scanReport(r.db.QueryRow(ctx, `SELECT `+reportColumns+` FROM item_reports WHERE id = $1`, id), &rp); err != nil {
		return nil, mapPgError(err)
	}
	return &rp, nil
}

func (r *postgresReportRepository) GetPublicReportByID(ctx context.Context, id int64) (*model.PublicItemReport, error) {
	var p model.PublicItemReport
	if err := scanPublicReport(r.db.QueryRow(ctx, publicReportSelect+` WHERE r.id = $1`, id), &p); err != nil {
		return nil, mapPgError(err)
	}
	return &p, nil
}

func buildReportWhere(f model.ReportFilter) (string, []any) {
	conds := []string{}
	args := []any{}
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.Keyword != "" {
		add("(r.item_name ILIKE $%[1]d OR r.description ILIKE $%[1]d)", "%"+f.Keyword+"%")
	}
	if f.ReportType != "" {
		add("r.report_type = $%[1]d", f.ReportType)
	}
	if f.Status != "" {
		add("r.status = $%[1]d", f.Status)
	}
	if f.CategoryID != nil {
		add("r.category_id = $%[1]d", *f.CategoryID)
	}
	if f.Location != "" {
		add("r.location ILIKE $%[1]d", "%"+f.Location+"%")
	}
	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func (r *postgresReportRepository) ListPublicReports(ctx context.Context, f model.ReportFilter) ([]model.PublicItemReport, int64, error) {
	where, args := buildReportWhere(f)

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM item_reports r`+where, args...).Scan(&total); err != nil {
		return nil, 0, mapPgError(err)
	}

	limit, offset := normalizePaging(f.Limit, f.Offset)
	args = append(args, limit, offset)
	query := publicReportSelect + where +
		fmt.Sprintf(` ORDER BY r.created_at DESC, r.id DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, mapPgError(err)
	}
	defer rows.Close()

	list := []model.PublicItemReport{}
	for rows.Next() {
		var p model.PublicItemReport
		if err := scanPublicReport(rows, &p); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

func (r *postgresReportRepository) ListReportsByUser(ctx context.Context, userID int64, limit, offset int) ([]model.ItemReport, error) {
	limit, offset = normalizePaging(limit, offset)
	rows, err := r.db.Query(ctx,
		`SELECT `+reportColumns+` FROM item_reports WHERE user_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	list := []model.ItemReport{}
	for rows.Next() {
		var rp model.ItemReport
		if err := scanReport(rows, &rp); err != nil {
			return nil, err
		}
		list = append(list, rp)
	}
	return list, rows.Err()
}

func (r *postgresReportRepository) UpdateReport(ctx context.Context, rp *model.ItemReport) error {
	query := `
		UPDATE item_reports
		SET category_id = $1, item_name = $2, description = $3, private_info = $4,
		    location = $5, incident_date = $6, image_url = $7
		WHERE id = $8
		RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, rp.CategoryID, rp.ItemName, rp.Description, rp.PrivateInfo,
		rp.Location, rp.IncidentDate, rp.ImageURL, rp.ID).Scan(&rp.UpdatedAt)
	return mapPgError(err)
}

func (r *postgresReportRepository) UpdateReportStatus(ctx context.Context, id int64, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE item_reports SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *postgresReportRepository) DeleteReport(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM item_reports WHERE id = $1`, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}