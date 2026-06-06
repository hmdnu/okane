package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type ScheduledJobService struct {
	db *sql.DB
}

type salaryJobConfig struct {
	Amount     float64 `json:"amount"`
	CategoryID uint    `json:"categoryId"`
}

func ScheduledJobServiceInit(db *sql.DB) *ScheduledJobService {
	return &ScheduledJobService{db: db}
}

func StartScheduledJobRunner(db *sql.DB) {
	service := ScheduledJobServiceInit(db)
	go service.Run(context.Background())
}

func (s *ScheduledJobService) Run(ctx context.Context) {
	s.runDueJobs(ctx, time.Now())

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.runDueJobs(ctx, now)
		}
	}
}

func (s *ScheduledJobService) runDueJobs(ctx context.Context, now time.Time) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, user_id, config_json, day_of_month, COALESCE(last_run_period, '')
		FROM scheduled_jobs
		WHERE kind = 'salary'
			AND frequency = 'monthly'
			AND enabled = 1
			AND deleted_at IS NULL`,
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var userID uint
		var configJSON string
		var dayOfMonth int
		var lastRunPeriod string

		if err := rows.Scan(&id, &userID, &configJSON, &dayOfMonth, &lastRunPeriod); err != nil {
			log.Println(err)
			continue
		}

		if !salaryJobDue(now, dayOfMonth, lastRunPeriod) {
			continue
		}

		var config salaryJobConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			log.Println(err)
			continue
		}

		if err := s.createSalaryTransaction(ctx, id, userID, now.Format("2006-01"), config); err != nil {
			log.Println(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
	}
}

func salaryJobDue(now time.Time, dayOfMonth int, lastRunPeriod string) bool {
	period := now.Format("2006-01")
	if lastRunPeriod == period {
		return false
	}

	runDay := dayOfMonth
	if lastDay := daysInMonth(now); runDay > lastDay {
		runDay = lastDay
	}

	return now.Day() == runDay
}

func daysInMonth(value time.Time) int {
	return time.Date(value.Year(), value.Month()+1, 0, 0, 0, 0, 0, value.Location()).Day()
}

func (s *ScheduledJobService) createSalaryTransaction(ctx context.Context, jobID uint, userID uint, period string, config salaryJobConfig) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`UPDATE scheduled_jobs
		SET last_run_period = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
			AND enabled = 1
			AND deleted_at IS NULL
			AND (last_run_period IS NULL OR last_run_period != ?)`,
		period,
		jobID,
		period,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return nil
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transactions (user_id, category_id, name, amount, type, note)
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID,
		config.CategoryID,
		"Salary",
		config.Amount,
		"income",
		"Auto salary",
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
