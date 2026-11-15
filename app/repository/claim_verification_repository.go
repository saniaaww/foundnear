package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// ClaimVerificationRepository: verifikasi klaim oleh petugas.
type ClaimVerificationRepository interface {
	VerifyClaim(ctx context.Context, v *model.ClaimVerification) error
	GetVerificationByClaimID(ctx context.Context, claimID int64) (*model.ClaimVerification, error)
}

type postgresClaimVerificationRepository struct {
	db *pgxpool.Pool
}

func NewClaimVerificationRepository(db *pgxpool.Pool) ClaimVerificationRepository {
	return &postgresClaimVerificationRepository{db: db}
}

func (r *postgresClaimVerificationRepository) VerifyClaim(ctx context.Context, v *model.ClaimVerification) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var reportID int64
	err = tx.QueryRow(ctx, `SELECT status, report_id FROM item_claims WHERE id = $1 FOR UPDATE`, v.ClaimID).
		Scan(&status, &reportID)
	if err != nil {
		return mapPgError(err)
	}
	if status != model.ClaimStatusPending {
		return fmt.Errorf("%w: claim status is %s", model.ErrInvalidState, status)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO claim_verifications (claim_id, officer_id, verification_note, result)
		 VALUES ($1, $2, $3, $4) RETURNING id, verified_at`,
		v.ClaimID, v.OfficerID, v.VerificationNote, v.Result).Scan(&v.ID, &v.VerifiedAt)
	if err != nil {
		return mapPgError(err)
	}

	if _, err = tx.Exec(ctx, `UPDATE item_claims SET status = $1 WHERE id = $2`, v.Result, v.ClaimID); err != nil {
		return mapPgError(err)
	}

	if v.Result == model.VerificationApproved {
		if _, err = tx.Exec(ctx, `UPDATE item_reports SET status = $1 WHERE id = $2`,
			model.ReportStatusMatched, reportID); err != nil {
			return mapPgError(err)
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresClaimVerificationRepository) GetVerificationByClaimID(ctx context.Context, claimID int64) (*model.ClaimVerification, error) {
	var v model.ClaimVerification
	err := r.db.QueryRow(ctx,
		`SELECT id, claim_id, officer_id, verification_note, result, verified_at
		 FROM claim_verifications WHERE claim_id = $1`, claimID).
		Scan(&v.ID, &v.ClaimID, &v.OfficerID, &v.VerificationNote, &v.Result, &v.VerifiedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &v, nil
}
