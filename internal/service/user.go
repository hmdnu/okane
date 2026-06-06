package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/hmdnu/okane/internal/dto"
	"github.com/hmdnu/okane/internal/model"
	"github.com/hmdnu/okane/internal/view/page/login"
	"github.com/hmdnu/okane/internal/view/page/register"
	"github.com/hmdnu/okane/internal/view/page/setting"
	"github.com/hmdnu/okane/lib"
)

type UserService struct {
	db *sql.DB
}

func UserServiceInit(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Register(user *dto.RegisterDto, ctx context.Context) error {
	hashedPassword, err := lib.HashPassword(user.Password)

	if err != nil {
		log.Println(err)
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO users (username, password, email) VALUES (?, ?, ?)`,
		user.Username,
		hashedPassword,
		user.Email,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *UserService) Login(dto *dto.LoginDto, ctx context.Context) (uint, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, email, password FROM users WHERE email = ? OR username = ?`, dto.Identifier, dto.Identifier)

	var user model.User

	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		return 0, errors.New("credential incorrect")
	}
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if !lib.CheckPasswordHash(dto.Password, user.Password) {
		return 0, errors.New("credential incorrect")
	}

	return user.ID, nil
}

func (s *UserService) GetUserSetting(ctx context.Context, userID uint) (dto.UserSettingData, error) {
	var data dto.UserSettingData
	err := s.db.QueryRowContext(
		ctx,
		`SELECT username, email, created_at, updated_at
		FROM users
		WHERE id = ?
			AND deleted_at IS NULL`,
		userID,
	).Scan(&data.Username, &data.Email, &data.CreatedAt, &data.UpdatedAt)
	if err != nil {
		log.Println(err)
		return data, err
	}

	salarySetting, err := s.GetSalarySetting(ctx, userID)
	if err != nil {
		return data, err
	}
	data.SalarySetting = salarySetting

	categories, err := s.GetCategories(ctx, userID)
	if err != nil {
		return data, err
	}
	data.Categories = categories

	return data, nil
}

func (s *UserService) GetSalarySetting(ctx context.Context, userID uint) (dto.SalarySettingData, error) {
	setting := dto.SalarySettingData{
		Enabled:    false,
		DayOfMonth: 25,
	}

	var configJSON string
	var enabled int
	err := s.db.QueryRowContext(
		ctx,
		`SELECT config_json, day_of_month, enabled
		FROM scheduled_jobs
		WHERE user_id = ?
			AND kind = 'salary'
			AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 1`,
		userID,
	).Scan(&configJSON, &setting.DayOfMonth, &enabled)
	if err == sql.ErrNoRows {
		if categoryID, err := s.defaultSalaryCategoryID(ctx, userID); err == nil {
			setting.CategoryID = categoryID
		}
		return setting, nil
	}
	if err != nil {
		log.Println(err)
		return setting, err
	}

	var config salaryJobConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Println(err)
		return setting, err
	}

	setting.Enabled = enabled == 1
	setting.Amount = config.Amount
	setting.CategoryID = config.CategoryID
	return setting, nil
}

func (s *UserService) GetCategories(ctx context.Context, userID uint) ([]dto.DashboardCategory, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name
		FROM categories
		WHERE deleted_at IS NULL
			AND (user_id = ? OR user_id IS NULL)
		ORDER BY name ASC`,
		userID,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	categories := make([]dto.DashboardCategory, 0)
	for rows.Next() {
		var category dto.DashboardCategory
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Println(err)
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return categories, nil
}

func (s *UserService) SaveSalarySetting(ctx context.Context, userID uint, setting dto.SalarySettingData) error {
	if setting.CategoryID == 0 {
		return ErrCategoryNotFound
	}

	var foundCategoryID uint
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id
		FROM categories
		WHERE id = ?
			AND deleted_at IS NULL
			AND (user_id = ? OR user_id IS NULL)`,
		setting.CategoryID,
		userID,
	).Scan(&foundCategoryID)
	if err == sql.ErrNoRows {
		return ErrCategoryNotFound
	}
	if err != nil {
		log.Println(err)
		return err
	}

	configJSON, err := json.Marshal(salaryJobConfig{
		Amount:     setting.Amount,
		CategoryID: foundCategoryID,
	})
	if err != nil {
		return err
	}

	enabled := 0
	if setting.Enabled {
		enabled = 1
	}

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE scheduled_jobs
		SET config_json = ?,
			day_of_month = ?,
			enabled = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
			AND kind = 'salary'
			AND deleted_at IS NULL`,
		string(configJSON),
		setting.DayOfMonth,
		enabled,
		userID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 0 {
		return nil
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO scheduled_jobs (user_id, kind, name, config_json, frequency, day_of_month, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID,
		"salary",
		"Salary",
		string(configJSON),
		"monthly",
		setting.DayOfMonth,
		enabled,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *UserService) defaultSalaryCategoryID(ctx context.Context, userID uint) (uint, error) {
	var categoryID uint
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id
		FROM categories
		WHERE deleted_at IS NULL
			AND (user_id = ? OR user_id IS NULL)
			AND name = 'Salary'
		ORDER BY user_id IS NOT NULL ASC
		LIMIT 1`,
		userID,
	).Scan(&categoryID)
	return categoryID, err
}

func (s *UserService) RenderLogin(w http.ResponseWriter, r *http.Request, validationErrors []lib.FormError, successMsg string) {
	err := login.Login(validationErrors, successMsg).Render(r.Context(), w)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *UserService) RenderUserSetting(w http.ResponseWriter, r *http.Request, data dto.UserSettingData) {
	err := setting.UserSetting(data).Render(r.Context(), w)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *UserService) RenderRegister(w http.ResponseWriter, r *http.Request, validationErrors []lib.FormError) {
	err := register.Register(validationErrors).Render(r.Context(), w)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
