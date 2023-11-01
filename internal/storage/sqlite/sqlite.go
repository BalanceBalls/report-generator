package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/BalanceBalls/report-generator/internal/storage"
)

type SqliteStorage struct {
	db *sql.DB
}

func New(name string) (*SqliteStorage, error) {
	db, err := sql.Open("sqlite3", name)

	if err != nil {
		return nil, fmt.Errorf("Could not open database:  %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Could not access database: %w", err)
	}

	return &SqliteStorage{db: db}, nil
}

func (s *SqliteStorage) Up() error {
	_, err := s.db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("Could not create table users: %w", err)
	}

	_, err = s.db.Exec(createReportsTable)
	if err != nil {
		return fmt.Errorf("Could not create table reports: %w", err)
	}

	_, err = s.db.Exec(createRowsTable)
	if err != nil {
		return fmt.Errorf("Could not create table reports: %w", err)
	}

	return nil
}

func (s *SqliteStorage) User(userId int) (*storage.User, error) {
	q, err := s.db.Prepare(getUserById)

	if err != nil {
		return nil, fmt.Errorf("Failed to build query: %w", err)
	}

	user := &storage.User{}
	err = q.QueryRow(userId).Scan(&user.Id, &user.UserEmail, &user.UserToken, &user.IsActive)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch row: %w", err)
	}

	return user, nil
}

func (s *SqliteStorage) Users() ([]storage.FlatUser, error) {
	rows, err := s.db.Query(getFullUsers, 10, 0)

	if err != nil {
		return []storage.FlatUser{}, fmt.Errorf("Failed to fetch rows: %w", err)
	}

	defer rows.Close()

	result := []storage.FlatUser{}

	for rows.Next() {
		tFlatUser := storage.FlatUser{}

		var rawDate string
		err := rows.Scan(
			&tFlatUser.Id, &tFlatUser.UserEmail, &tFlatUser.UserToken, &tFlatUser.IsActive,
			&tFlatUser.ReportId, &tFlatUser.UserId,
			&tFlatUser.ReportRowId, &rawDate, &tFlatUser.Task, &tFlatUser.Link, &tFlatUser.TimeSpent)

		if err != nil {
			return []storage.FlatUser{}, fmt.Errorf("Error scanning rows: %w", err)
		}

		rowDateParsed, dateErr := time.Parse(time.RFC3339, rawDate)
		if dateErr != nil {
			return []storage.FlatUser{}, fmt.Errorf("Error parsing date: %w", dateErr)
		}

		tFlatUser.Date = rowDateParsed
		result = append(result, tFlatUser)
	}

	return result, nil
}