package sqlite

import "fmt"

func (s *SqliteStorage) Seed() error {
	usersSeed := `INSERT INTO users (Id, UserEmail, UserToken, TimezoneOffset, IsActive) VALUES (?, ?, ?, ?, ?)`

	if _, err := s.db.Exec(usersSeed, 1, "TestEmail", "TestToken", 300, true); err != nil {
		return fmt.Errorf("Could not seed user data: %w", err)
	}

	if _, err := s.db.Exec(usersSeed, 2, "qwe@asd.com", "asdfjkdf;la34i", 300, true); err != nil {
		return fmt.Errorf("Could not seed user data: %w", err)
	}

	reportsSeed := `INSERT INTO reports (UserId) VALUES (?)`

	if _, reportsErr := s.db.Exec(reportsSeed, 1); reportsErr != nil {
		return fmt.Errorf("Could not seed reports data: %w", reportsErr)
	}

	if _, reportsErr := s.db.Exec(reportsSeed, 2); reportsErr != nil {
		return fmt.Errorf("Could not seed reports data: %w", reportsErr)
	}

	rowsSeed := `INSERT INTO rows (ReportId, Date, Task, Link, TimeSpent) VALUES (?, ?, ?, ?, ?)`

	if _, rowsErr := s.db.Exec(rowsSeed, 1, "2006-01-02T15:04:05Z", "#Casino/145", "http://git.casino.com/mr/i2ji4314uhiouhi4124l", 4.2); rowsErr != nil {
		return fmt.Errorf("Could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 1, "2006-01-02T18:04:05Z", "#Casino/149", "http://git.casino.com/mr/i2ji4314uojhj4pfrufu", 3.8); rowsErr != nil {
		return fmt.Errorf("Could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 2, "2009-01-02T15:04:05Z", "#Casino/201", "http://git.casino.com/mr/i2ji4314uhiouhi4124l", 6.2); rowsErr != nil {
		return fmt.Errorf("Could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 2, "2009-01-02T18:04:05Z", "#Casino/199", "http://git.casino.com/mr/i2ji4314uokjskdjsdiiii", 1.2); rowsErr != nil {
		return fmt.Errorf("Could not seed rows data: %w", rowsErr)
	}

	return nil
}
