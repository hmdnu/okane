package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/hmdnu/okane/internal/dto"
	"github.com/hmdnu/okane/internal/model"
	"github.com/hmdnu/okane/internal/view/page/dashboard"
)

type TransactionService struct {
	db *sql.DB
}

var ErrCategoryNotFound = errors.New("category not found")

func TransactionServiceInit(db *sql.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) GetDashboard(ctx context.Context, userID uint, filter dto.DashboardFilter) (dto.DashboardData, error) {
	data := dto.DashboardData{Filter: filter}

	var err error
	if filter.IsActive {
		selectedDate := time.Date(filter.Year, time.Month(filter.Month), filter.Day, 0, 0, 0, 0, time.Local)
		monthStart := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.Local)
		selectedDateEnd := selectedDate.AddDate(0, 0, 1)

		err = s.db.QueryRowContext(
			ctx,
			`SELECT
				COALESCE(SUM(CASE WHEN type IN ('income', 'initial_balance') THEN amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
			FROM transactions
			WHERE user_id = ?
				AND deleted_at IS NULL
				AND created_at >= ?
				AND created_at < ?`,
			userID,
			monthStart,
			selectedDateEnd,
		).Scan(&data.Summary.Income, &data.Summary.Expense)
	} else {
		err = s.db.QueryRowContext(
			ctx,
			`SELECT
				COALESCE(SUM(CASE WHEN type IN ('income', 'initial_balance') THEN amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
			FROM transactions
			WHERE user_id = ?
				AND deleted_at IS NULL`,
			userID,
		).Scan(&data.Summary.Income, &data.Summary.Expense)
	}
	if err != nil {
		log.Println(err)
		return data, err
	}

	data.Summary.Balance = data.Summary.Income - data.Summary.Expense

	averageSpending, err := s.getMonthlyAverageSpending(ctx, userID, filter)
	if err != nil {
		return data, err
	}
	data.Summary.AverageSpending = averageSpending

	rows, err := s.getDashboardTransactions(ctx, userID, filter)
	if err != nil {
		log.Println(err)
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction dto.DashboardTransaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Name,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Note,
			&transaction.CategoryName,
			&transaction.CreatedAt,
		); err != nil {
			log.Println(err)
			return data, err
		}

		data.Transactions = append(data.Transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return data, err
	}

	categorySpendings, err := s.getDashboardCategorySpendings(ctx, userID, filter)
	if err != nil {
		return data, err
	}
	data.CategorySpendings = categorySpendings

	categories, err := s.GetCategories(ctx, userID)
	if err != nil {
		return data, err
	}
	data.Categories = categories

	return data, nil
}

func (s *TransactionService) getDashboardTransactions(ctx context.Context, userID uint, filter dto.DashboardFilter) (*sql.Rows, error) {
	if !filter.IsActive {
		return s.db.QueryContext(
			ctx,
			`SELECT
				t.id,
				t.name,
				t.amount,
				t.type,
				COALESCE(t.note, ''),
				COALESCE(c.name, ''),
				t.created_at
			FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.user_id = ?
				AND t.deleted_at IS NULL
			ORDER BY t.created_at DESC`,
			userID,
		)
	}

	selectedDate := time.Date(filter.Year, time.Month(filter.Month), filter.Day, 0, 0, 0, 0, time.Local)
	selectedDateEnd := selectedDate.AddDate(0, 0, 1)

	return s.db.QueryContext(
		ctx,
		`SELECT
			t.id,
			t.name,
			t.amount,
			t.type,
			COALESCE(t.note, ''),
			COALESCE(c.name, ''),
			t.created_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = ?
			AND t.deleted_at IS NULL
			AND t.created_at >= ?
			AND t.created_at < ?
		ORDER BY t.created_at DESC`,
		userID,
		selectedDate,
		selectedDateEnd,
	)
}

func (s *TransactionService) getMonthlyAverageSpending(ctx context.Context, userID uint, filter dto.DashboardFilter) (float64, error) {
	now := time.Now()
	monthYear := now.Year()
	month := now.Month()

	if filter.IsActive {
		monthYear = filter.Year
		month = time.Month(filter.Month)
	}

	monthStart := time.Date(monthYear, month, 1, 0, 0, 0, 0, time.Local)
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	var average sql.NullFloat64
	err := s.db.QueryRowContext(
		ctx,
		`SELECT AVG(amount)
		FROM transactions
		WHERE user_id = ?
			AND deleted_at IS NULL
			AND type = 'expense'
			AND created_at >= ?
			AND created_at < ?`,
		userID,
		monthStart,
		nextMonthStart,
	).Scan(&average)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if !average.Valid {
		return 0, nil
	}

	return average.Float64, nil
}

func (s *TransactionService) getDashboardCategorySpendings(ctx context.Context, userID uint, filter dto.DashboardFilter) ([]dto.DashboardCategorySpending, error) {
	var rows *sql.Rows
	var err error

	if filter.IsActive {
		selectedDate := time.Date(filter.Year, time.Month(filter.Month), filter.Day, 0, 0, 0, 0, time.Local)
		selectedDateEnd := selectedDate.AddDate(0, 0, 1)

		rows, err = s.db.QueryContext(
			ctx,
			`SELECT COALESCE(c.name, 'No category'), COALESCE(SUM(t.amount), 0)
			FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.user_id = ?
				AND t.deleted_at IS NULL
				AND t.type = 'expense'
				AND t.created_at >= ?
				AND t.created_at < ?
			GROUP BY COALESCE(c.name, 'No category')
			ORDER BY SUM(t.amount) DESC`,
			userID,
			selectedDate,
			selectedDateEnd,
		)
	} else {
		rows, err = s.db.QueryContext(
			ctx,
			`SELECT COALESCE(c.name, 'No category'), COALESCE(SUM(t.amount), 0)
			FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.user_id = ?
				AND t.deleted_at IS NULL
				AND t.type = 'expense'
			GROUP BY COALESCE(c.name, 'No category')
			ORDER BY SUM(t.amount) DESC`,
			userID,
		)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	categorySpendings := make([]dto.DashboardCategorySpending, 0)
	for rows.Next() {
		var spending dto.DashboardCategorySpending
		if err := rows.Scan(&spending.CategoryName, &spending.Total); err != nil {
			log.Println(err)
			return nil, err
		}
		categorySpendings = append(categorySpendings, spending)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return categorySpendings, nil
}

func (s *TransactionService) GetCategories(ctx context.Context, userID uint) ([]dto.DashboardCategory, error) {
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

func (s *TransactionService) CreateTransaction(ctx context.Context, transaction model.TransactionDto) error {
	var categoryID any
	if transaction.CategoryID != nil {
		var foundCategoryID uint
		err := s.db.QueryRowContext(
			ctx,
			`SELECT id
			FROM categories
			WHERE id = ?
				AND deleted_at IS NULL
				AND (user_id = ? OR user_id IS NULL)`,
			*transaction.CategoryID,
			transaction.UserID,
		).Scan(&foundCategoryID)
		if err == sql.ErrNoRows {
			return ErrCategoryNotFound
		}
		if err != nil {
			log.Println(err)
			return err
		}

		categoryID = foundCategoryID
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO transactions (user_id, category_id, name, amount, type, note)
		VALUES (?, ?, ?, ?, ?, ?)`,
		transaction.UserID,
		categoryID,
		transaction.Name,
		transaction.Amount,
		transaction.Type,
		transaction.Note,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, id uint, userID uint) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE transactions
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

func (s *TransactionService) RenderDashboard(w http.ResponseWriter, r *http.Request, data dto.DashboardData) {
	err := dashboard.Dashboard(data).Render(r.Context(), w)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
