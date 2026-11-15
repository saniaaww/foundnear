package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// UserRepository: kontrak dengan Sania (Register, Login, Profil, JWT).
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
}

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

const userSelect = `
	SELECT u.id, u.role_id, r.name, u.name, u.email, u.password_hash, u.phone, u.created_at, u.updated_at
	FROM users u
	JOIN roles r ON r.id = u.role_id`

func scanUser(s rowScanner, u *model.User) error {
	return s.Scan(&u.ID, &u.RoleID, &u.RoleName, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	query := `
		INSERT INTO users (role_id, name, email, password_hash, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, user.RoleID, user.Name, user.Email, user.PasswordHash, user.Phone).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return mapPgError(err)
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var u model.User
	if err := scanUser(r.db.QueryRow(ctx, userSelect+` WHERE u.email = $1`, email), &u); err != nil {
		return nil, mapPgError(err)
	}
	return &u, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	if err := scanUser(r.db.QueryRow(ctx, userSelect+` WHERE u.id = $1`, id), &u); err != nil {
		return nil, mapPgError(err)
	}
	return &u, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, phone = $2 WHERE id = $3 RETURNING updated_at`,
		user.Name, user.Phone, user.ID).Scan(&user.UpdatedAt)
	return mapPgError(err)
}
