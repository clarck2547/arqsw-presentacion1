package persistence

import (
	"Taller2/Product/internal/command/domain/entities"
	"Taller2/Product/internal/command/domain/repositories"
	"context"
	"database/sql"
	"fmt"
)

type SQLiteProductRepository struct {
	db *sql.DB
}

func NewSQLiteProductRepository(dsn string) (repositories.ProductRepository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Crear tabla si no existe
	schema := `
	CREATE TABLE IF NOT EXISTS products_command (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		price REAL,
		stock INTEGER
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &SQLiteProductRepository{db: db}, nil
}

func (r *SQLiteProductRepository) Save(ctx context.Context, product *entities.Product) error {
	query := `INSERT INTO products_command (id, name, description, price, stock) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, product.ID, product.Name, product.Description, product.Price, product.Stock)
	return err
}

func (r *SQLiteProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products_command WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *SQLiteProductRepository) Update(ctx context.Context, product *entities.Product) error {
	query := `
		UPDATE products_command 
		SET name = ?, description = ?, price = ?, stock = ?
		WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
