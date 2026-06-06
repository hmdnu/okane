package config

import (
	"database/sql"

	"github.com/hmdnu/okane/internal/storage"
)

func StartDb() (*sql.DB, error) {
	db, err := storage.Open()
	if err != nil {
		return nil, err
	}

	if err := seedGlobalCategories(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func seedGlobalCategories(db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO categories (user_id, name)
		SELECT NULL, 'Salary'
		WHERE NOT EXISTS (
			SELECT 1
			FROM categories
			WHERE user_id IS NULL
				AND name = 'Salary'
				AND deleted_at IS NULL
		)`,
	)
	return err
}
