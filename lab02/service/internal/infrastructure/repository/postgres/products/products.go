package products

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virogg/networks-course/service/internal/domain"
	domainerr "github.com/virogg/networks-course/service/internal/domain/errors"
	infraerr "github.com/virogg/networks-course/service/internal/infrastructure/errors"
)

type ProductsPostgresRepository struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewProductsPostgresRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *ProductsPostgresRepository {
	return &ProductsPostgresRepository{
		db:     db,
		getter: getter,
	}
}

func (r *ProductsPostgresRepository) Create(ctx context.Context, product *domain.Product) (int64, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("products").
		Columns("name", "description").
		Values(product.Name, product.Description).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build insert query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	if err := conn.QueryRow(ctx, query, args...).Scan(&product.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, domainerr.ErrProductExists
		}
		return 0, fmt.Errorf("%w: during `create booking`: %w", infraerr.ErrDB, err)
	}

	return product.ID, nil
}

func (r *ProductsPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name", "description", "COALESCE(icon_path, '') AS icon_path").
		From("products").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get by id query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var product domain.Product
	if err := conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Description, &product.IconPath); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: during `get product by id`: %w", infraerr.ErrDB, err)
	}
	return &product, nil
}

func (r *ProductsPostgresRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name", "description", "COALESCE(icon_path, '') AS icon_path").
		From("products").
		OrderBy("id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get all query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: during `get all products`: %w", infraerr.ErrDB, err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.IconPath); err != nil {
			return nil, fmt.Errorf("%w: during `scan product`: %w", infraerr.ErrDB, err)
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: during `get all products`: %w", infraerr.ErrDB, err)
	}

	return products, nil
}

func (r *ProductsPostgresRepository) Update(ctx context.Context, product *domain.Product) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("products").
		Set("name", product.Name).
		Set("description", product.Description).
		Where(sq.Eq{"id": product.ID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerr.ErrProductNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domainerr.ErrProductExists
		}

		return fmt.Errorf("%w: during `update product`: %w", infraerr.ErrDB, err)
	}

	return nil
}

func (r *ProductsPostgresRepository) Delete(ctx context.Context, id int64) (*domain.Product, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Delete("products").
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, name, description, COALESCE(icon_path, '') AS icon_path")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build delete query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	var product domain.Product

	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Description, &product.IconPath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerr.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: during `delete product`: %w", infraerr.ErrDB, err)
	}

	return &product, nil
}

func (r *ProductsPostgresRepository) SetIcon(ctx context.Context, id int64, iconPath string) error {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("products").
		Set("icon_path", iconPath).
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build set icon query: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerr.ErrProductNotFound
		}
		return fmt.Errorf("%w: during `set icon`: %w", infraerr.ErrDB, err)
	}
	return nil
}
