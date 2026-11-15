package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// ClaimRepository: pengajuan klaim kepemilikan.
type ClaimRepository interface {
	CreateClaim(ctx context.Context, claim *model.ItemClaim) error
	GetClaimByID(ctx context.Context, id int64) (*model.ItemClaim, error)
	ListClaimsByReportID(ctx context.Context, reportID int64) ([]model.ItemClaim, error)
	ListClaimsByClaimant(ctx context.Context, claimantID int64, limit, offset int) ([]model.ItemClaim, error)
	ListClaims(ctx context.Context, status string, limit, offset int) ([]model.ItemClaim, error)
	UpdateClaimStatus(ctx context.Context, id int64, status string) error
}

type postgresClaimRepository struct {
	db *pgxpool.Pool
}

func NewClaimRepository(db *pgxpool.Pool) ClaimRepository {
	return &postgresClaimRepository{db: db}
}

const claimColumns = `id, report_id, claimant_id, claim_description, status, created_at, updated_at`

func scanClaim(s rowScanner, c *model.ItemClaim) error {
	return s.Scan(&c.ID, &c.ReportID, &c.ClaimantID, &c.ClaimDescription, &c.Status, &c.CreatedAt, &c.UpdatedAt)
}

func (r *postgresClaimRepository) CreateClaim(ctx context.Context, c *model.ItemClaim) error {
	if c.Status == "" {
		c.Status = model.ClaimStatusPending
	}
	err := r.db.QueryRow(ctx,
		`INSERT INTO item_claims (report_id, claimant_id, claim_description, status)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		c.ReportID, c.ClaimantID, c.ClaimDescription, c.Status).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return mapPgError(err)
}

func (r *postgresClaimRepository) GetClaimByID(ctx context.Context, id int64) (*model.ItemClaim, error) {
	var c model.ItemClaim
	if err := scanClaim(r.db.QueryRow(ctx, `SELECT `+claimColumns+` FROM item_claims WHERE id = $1`, id), &c); err != nil {
		return nil, mapPgError(err)
	}
	return &c, nil
}

func (r *postgresClaimRepository) collectClaims(ctx context.Context, query string, args ...any) ([]model.ItemClaim, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	list := []model.ItemClaim{}
	for rows.Next() {
		var c model.ItemClaim
		if err := scanClaim(rows, &c); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *postgresClaimRepository) ListClaimsByReportID(ctx context.Context, reportID int64) ([]model.ItemClaim, error) {
	return r.collectClaims(ctx,
		`SELECT `+claimColumns+` FROM item_claims WHERE report_id = $1 ORDER BY created_at DESC, id DESC`, reportID)
}

func (r *postgresClaimRepository) ListClaimsByClaimant(ctx context.Context, claimantID int64, limit, offset int) ([]model.ItemClaim, error) {
	limit, offset = normalizePaging(limit, offset)
	return r.collectClaims(ctx,
		`SELECT `+claimColumns+` FROM item_claims WHERE claimant_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
		claimantID, limit, offset)
}

func (r *postgresClaimRepository) ListClaims(ctx context.Context, status string, limit, offset int) ([]model.ItemClaim, error) {
	limit, offset = normalizePaging(limit, offset)
	if status == "" {
		return r.collectClaims(ctx,
			`SELECT `+claimColumns+` FROM item_claims ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
	}
	return r.collectClaims(ctx,
		`SELECT `+claimColumns+` FROM item_claims WHERE status = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
		status, limit, offset)
}

func (r *postgresClaimRepository) UpdateClaimStatus(ctx context.Context, id int64, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE item_claims SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
