package persistence

import (
	"Taller2/Product/internal/query/domain/models"
	"Taller2/Product/internal/query/domain/repositories"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type SQLiteProductReadRepository struct {
	db *sql.DB
}

func NewSQLiteProductReadRepository(dsn string) (repositories.ProductReadRepository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Crear la tabla si no existe
	createTable := `
	CREATE TABLE IF NOT EXISTS products_query (
		id TEXT PRIMARY KEY,
		name TEXT,
		description TEXT,
		price REAL,
		stock INTEGER
	);`
	if _, err := db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	return &SQLiteProductReadRepository{db: db}, nil
}

func (r *SQLiteProductReadRepository) Save(ctx context.Context, product *models.ProductRead) error {
	query := `INSERT INTO products_query (id, name, description, price, stock) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, product.ID, product.Name, product.Description, product.Price, product.Stock)
	return err
}

func (r *SQLiteProductReadRepository) ListAll(ctx context.Context) ([]*models.ProductRead, error) {
	query := `SELECT id, name, description, price, stock FROM products_query`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.ProductRead
	for rows.Next() {
		var product models.ProductRead
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, nil
}

func (r *SQLiteProductReadRepository) FindByID(ctx context.Context, id string) (*models.ProductRead, error) {
	query := `SELECT id, name, price, stock FROM products_query WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var product models.ProductRead
	if err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *SQLiteProductReadRepository) ListWithPagination(ctx context.Context, limit, offset int) ([]*models.ProductRead, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM products_query`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	query := `
		SELECT id, name, description, price, stock 
		FROM products_query 
		ORDER BY name ASC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	var products []*models.ProductRead
	for rows.Next() {
		var p models.ProductRead
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock); err != nil {
			return nil, 0, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, &p)
	}

	return products, total, nil
}

func (r *SQLiteProductReadRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM products_query WHERE id = ?", id)
	return err
}

func (r *SQLiteProductReadRepository) Update(ctx context.Context, product *models.ProductRead) error {
	query := `
		UPDATE products_query 
		SET name = ?, description = ?, price = ?, stock = ? 
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ID)

	return err
}
