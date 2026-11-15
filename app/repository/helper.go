package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"foundnear/app/model"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", model.ErrDuplicate, pgErr.ConstraintName)
		case "23503":
			return fmt.Errorf("%w: %s", model.ErrForeignKey, pgErr.ConstraintName)
		case "23502", "23514":
			return fmt.Errorf("%w: %s", model.ErrConstraint, pgErr.Message)
		case "P0001":
			return fmt.Errorf("%w: %s", model.ErrBusinessRule, pgErr.Message)
		}
	}
	return err
}

func normalizePaging(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}