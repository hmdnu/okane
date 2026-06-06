package service

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/hmdnu/okane/internal/dto"
	"github.com/hmdnu/okane/internal/model"
	"github.com/hmdnu/okane/internal/view/page/category"
)

type CategoryService struct {
	db *sql.DB
}

func CategoryServiceInit(db *sql.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) GetCategoryManagement(ctx context.Context, userID uint) (dto.CategoryManagementData, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, user_id IS NULL, created_at
		FROM categories
		WHERE deleted_at IS NULL
			AND (user_id = ? OR user_id IS NULL)
		ORDER BY user_id IS NULL ASC, name ASC`,
		userID,
	)
	if err != nil {
		log.Println(err)
		return dto.CategoryManagementData{}, err
	}
	defer rows.Close()

	data := dto.CategoryManagementData{
		Categories: make([]dto.CategoryItem, 0),
	}

	for rows.Next() {
		var category dto.CategoryItem
		if err := rows.Scan(&category.ID, &category.Name, &category.IsGlobal, &category.CreatedAt); err != nil {
			log.Println(err)
			return dto.CategoryManagementData{}, err
		}

		data.Categories = append(data.Categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return dto.CategoryManagementData{}, err
	}

	return data, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, category model.CategoryDto) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO categories (user_id, name)
		VALUES (?, ?)`,
		category.UserID,
		category.Name,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id uint, userID uint, name string) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE categories
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
			AND user_id = ?
			AND deleted_at IS NULL`,
		name,
		id,
		userID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uint, userID uint) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE categories
		SET deleted_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
			AND user_id = ?
			AND deleted_at IS NULL`,
		id,
		userID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *CategoryService) RenderCategoryManagement(w http.ResponseWriter, r *http.Request, data dto.CategoryManagementData) {
	err := category.CategoryManagement(data).Render(r.Context(), w)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
