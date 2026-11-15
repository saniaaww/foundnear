package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// ItemReturnRepository: pencatatan pengembalian barang.
type ItemReturnRepository interface {
	CreateReturn(ctx context.Context, ret *model.ItemReturn) error
	GetReturnByID(ctx context.Context, id int64) (*model.ItemReturn, error)
	GetReturnByClaimID(ctx context.Context, claimID int64) (*model.ItemReturn, error)
}

type postgresItemReturnRepository struct {
	db *pgxpool.Pool
}

func NewItemReturnRepository(db *pgxpool.Pool) ItemReturnRepository {
	return &postgresItemReturnRepository{db: db}
}

func (r *postgresItemReturnRepository) CreateReturn(ctx context.Context, ret *model.ItemReturn) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var reportID int64
	err = tx.QueryRow(ctx, `SELECT status, report_id FROM item_claims WHERE id = $1 FOR UPDATE`, ret.ClaimID).
		Scan(&status, &reportID)
	if err != nil {
		return mapPgError(err)
	}
	if status != model.ClaimStatusApproved {
		return fmt.Errorf("%w: claim status is %s, must be approved", model.ErrInvalidState, status)
	}

	var returnDate any
	if !ret.ReturnDate.IsZero() {
		returnDate = ret.ReturnDate
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO item_returns (claim_id, returned_by, received_by, return_date, return_note)
		 VALUES ($1, $2, $3, COALESCE($4::timestamptz, NOW()), $5)
		 RETURNING id, return_date`,
		ret.ClaimID, ret.ReturnedBy, ret.ReceivedBy, returnDate, ret.ReturnNote).
		Scan(&ret.ID, &ret.ReturnDate)
	if err != nil {
		return mapPgError(err)
	}

	if _, err = tx.Exec(ctx, `UPDATE item_reports SET status = $1 WHERE id = $2`,
		model.ReportStatusReturned, reportID); err != nil {
		return mapPgError(err)
	}

	return tx.Commit(ctx)
}

func (r *postgresItemReturnRepository) scanReturn(ctx context.Context, where string, arg any) (*model.ItemReturn, error) {
	var ret model.ItemReturn
	err := r.db.QueryRow(ctx,
		`SELECT id, claim_id, returned_by, received_by, return_date, return_note FROM item_returns WHERE `+where, arg).
		Scan(&ret.ID, &ret.ClaimID, &ret.ReturnedBy, &ret.ReceivedBy, &ret.ReturnDate, &ret.ReturnNote)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &ret, nil
}

func (r *postgresItemReturnRepository) GetReturnByID(ctx context.Context, id int64) (*model.ItemReturn, error) {
	return r.scanReturn(ctx, "id = $1", id)
}

func (r *postgresItemReturnRepository) GetReturnByClaimID(ctx context.Context, claimID int64) (*model.ItemReturn, error) {
	return r.scanReturn(ctx, "claim_id = $1", claimID)
}
