package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"foundnear/app/model"
)

// CategoryRepository: data kategori barang.
type CategoryRepository interface {
	CreateCategory(ctx context.Context, c *model.Category) error
	GetCategoryByID(ctx context.Context, id int64) (*model.Category, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	UpdateCategory(ctx context.Context, c *model.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

type postgresCategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &postgresCategoryRepository{db: db}
}

func (r *postgresCategoryRepository) CreateCategory(ctx context.Context, c *model.Category) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id`,
		c.Name, c.Description).Scan(&c.ID)
	return mapPgError(err)
}

func (r *postgresCategoryRepository) GetCategoryByID(ctx context.Context, id int64) (*model.Category, error) {
	var c model.Category
	err := r.db.QueryRow(ctx, `SELECT id, name, description FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Description)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &c, nil
}

func (r *postgresCategoryRepository) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description FROM categories ORDER BY name`)
	if err != nil {
		return nil, mapPgError(err)
	}
	defer rows.Close()

	list := []model.Category{}
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *postgresCategoryRepository) UpdateCategory(ctx context.Context, c *model.Category) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE categories SET name = $1, description = $2 WHERE id = $3`,
		c.Name, c.Description, c.ID)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *postgresCategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
