package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	"github.com/BalanceBalls/report-generator/internal/logger"
	"github.com/BalanceBalls/report-generator/internal/report"
	"github.com/BalanceBalls/report-generator/internal/storage"
)

type PostgresStorage struct {
	db *sql.DB
}

func New(connectionString string) (*PostgresStorage, error) {
	slog.Info("initializing Postgres DB...")
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		return nil, fmt.Errorf("could not open database:  %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not access database: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Up(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, createUsersTable)
	if err != nil {
		return fmt.Errorf("could not create table users: %w", err)
	}

	_, err = s.db.Exec(createReportsTable)
	if err != nil {
		return fmt.Errorf("could not create table reports: %w", err)
	}

	_, err = s.db.Exec(createRowsTable)
	if err != nil {
		return fmt.Errorf("could not create table reports: %w", err)
	}

	return nil
}

func (s *PostgresStorage) AddUser(ctx context.Context, user report.User) error {
	_, err := s.db.ExecContext(ctx, addUser,
		user.Id, user.GitlabId, user.UserEmail, user.UserToken, user.TimezoneOffset, user.IsActive)
	if err != nil {
		return fmt.Errorf("could not add new user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) UserExists(ctx context.Context, userId int64) bool {
	logger := logger.GetFromContext(ctx)
	q, err := s.db.Prepare(checkUserExists)

	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		return false
	}

	var exists int
	err = q.QueryRowContext(ctx, userId).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}

		logger.ErrorContext(ctx, err.Error())
	}

	return true
}

func (s *PostgresStorage) User(ctx context.Context, userId int64) (report.User, error) {
	q, err := s.db.Prepare(getUserById)

	if err != nil {
		return report.User{}, fmt.Errorf("failed to build query: %w", err)
	}

	user := report.User{}
	err = q.QueryRowContext(ctx, userId).Scan(&user.Id, &user.GitlabId, &user.UserEmail, &user.UserToken, &user.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return report.User{}, storage.ErrUserNotFound
		}

		return report.User{}, fmt.Errorf("failed to fetch row: %w", err)
	}

	return user, nil
}

func (s *PostgresStorage) UpdateUser(ctx context.Context, user report.User) error {
	_, err := s.db.ExecContext(ctx, updateUser, user.GitlabId, user.UserEmail, user.UserToken, user.TimezoneOffset, user.Id)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) RemoveUser(ctx context.Context, userId int64) error {
	_, err := s.db.ExecContext(ctx, removeUser, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) SaveReport(ctx context.Context, report report.Report, userId int64) error {
	lastInsertId := 0
	err := s.db.QueryRowContext(ctx, addReport, userId).Scan(&lastInsertId)

	if err != nil {
		return err
	}

	columnsCnt := 5
	values := make([]interface{}, 0, len(report.Rows)*columnsCnt)
	query := addRows
	for i, reportRow := range report.Rows {
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d),",
			columnsCnt*i+1, columnsCnt*i+2, columnsCnt*i+3, columnsCnt*i+4, columnsCnt*i+5)
	
		values = append(values, lastInsertId, reportRow.Date, reportRow.Task, reportRow.Link, reportRow.TimeSpent)
	}

	// Trim comma at the end
	query = query[:len(query)-1]

	_, err = s.db.Exec(query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) Users(ctx context.Context) ([]storage.FlatUser, error) {
	rows, err := s.db.QueryContext(ctx, getFullUsers, 10, 0)

	if err != nil {
		return []storage.FlatUser{}, err
	}

	defer func() {
		err = rows.Close()
	}()

	result := []storage.FlatUser{}

	for rows.Next() {
		tFlatUser := storage.FlatUser{}

		var rawDate string
		err := rows.Scan(
			&tFlatUser.Id, &tFlatUser.GitlabId, &tFlatUser.UserEmail, &tFlatUser.UserToken, &tFlatUser.IsActive,
			&tFlatUser.ReportId, &tFlatUser.UserId,
			&tFlatUser.ReportRowId, &rawDate, &tFlatUser.Task, &tFlatUser.Link, &tFlatUser.TimeSpent)

		if err != nil {
			return []storage.FlatUser{}, err
		}

		rowDateParsed, dateErr := time.Parse(time.RFC3339, rawDate)
		if dateErr != nil {
			return []storage.FlatUser{}, err
		}

		tFlatUser.Date = rowDateParsed
		result = append(result, tFlatUser)
	}

	return result, nil
}
