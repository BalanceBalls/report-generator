package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/BalanceBalls/report-generator/internal/logger"
	"github.com/BalanceBalls/report-generator/internal/report"
	"github.com/BalanceBalls/report-generator/internal/storage"
)

type SqliteStorage struct {
	db *sql.DB
}

func New(name string) (*SqliteStorage, error) {
	slog.Info("initializing DB...", "db_name=", name)
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		return nil, fmt.Errorf("could not open database:  %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not access database: %w", err)
	}

	return &SqliteStorage{db: db}, nil
}

func (s *SqliteStorage) Up(ctx context.Context) error {
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

func (s *SqliteStorage) AddUser(ctx context.Context, user report.User) error {
	_, err := s.db.Exec(addUser, user.Id, user.GitlabId, user.UserEmail, user.UserToken, user.TimezoneOffset, user.IsActive)
	if err != nil {
		return fmt.Errorf("could not add new user: %w", err)
	}

	return nil
}

func (s *SqliteStorage) UserExists(ctx context.Context, userId int64) bool {
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

func (s *SqliteStorage) User(ctx context.Context, userId int64) (report.User, error) {
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

func (s *SqliteStorage) UpdateUser(ctx context.Context, user report.User) error {
	_, err := s.db.Exec(updateUser, user.GitlabId, user.UserEmail, user.UserToken, user.TimezoneOffset, user.Id)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return nil
}

func (s *SqliteStorage) RemoveUser(ctx context.Context, userId int64) error {
	_, err := s.db.Exec(removeUser, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteStorage) Users(ctx context.Context) ([]storage.FlatUser, error) {
	rows, err := s.db.QueryContext(ctx, getFullUsers, 10, 0)

	if err != nil {
		return []storage.FlatUser{}, err
	}

	defer rows.Close()

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
