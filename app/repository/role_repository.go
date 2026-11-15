package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// RoleRepository: akses data role 
type RoleRepository interface {
	GetRoleByID(ctx context.Context, id int64) (*model.Role, error)
	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
}

type postgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &postgresRoleRepository{db: db}
}

func (r *postgresRoleRepository) GetRoleByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(ctx, `SELECT id, name FROM roles WHERE id = $1`, id).Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &role, nil
}

func (r *postgresRoleRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(ctx, `SELECT id, name FROM roles WHERE name = $1`, name).Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &role, nil
}

func (r *postgresRoleRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM roles ORDER BY id`)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	roles := []model.Role{}
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}