package sqlite

import "fmt"

func (s *SqliteStorage) Seed() error {
	usersSeed := `INSERT INTO users (id, user_email, user_token, timezone_offset, is_active) VALUES (?, ?, ?, ?, ?)`

	if _, err := s.db.Exec(usersSeed, 1, "TestEmail", "TestToken", 300, true); err != nil {
		return fmt.Errorf("could not seed user data: %w", err)
	}

	if _, err := s.db.Exec(usersSeed, 2, "qwe@asd.com", "asdfjkdf;la34i", 300, true); err != nil {
		return fmt.Errorf("could not seed user data: %w", err)
	}

	reportsSeed := `INSERT INTO reports (user_id) VALUES (?)`

	if _, reportsErr := s.db.Exec(reportsSeed, 1); reportsErr != nil {
		return fmt.Errorf("could not seed reports data: %w", reportsErr)
	}

	if _, reportsErr := s.db.Exec(reportsSeed, 2); reportsErr != nil {
		return fmt.Errorf("could not seed reports data: %w", reportsErr)
	}

	rowsSeed := `INSERT INTO rows (report_id, date, task, link, time_spent) VALUES (?, ?, ?, ?, ?)`

	if _, rowsErr := s.db.Exec(rowsSeed, 1, "2006-01-02T15:04:05Z", "#Casino/145", "http://git.casino.com/mr/i2ji4314uhiouhi4124l", 4.2); rowsErr != nil {
		return fmt.Errorf("could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 1, "2006-01-02T18:04:05Z", "#Casino/149", "http://git.casino.com/mr/i2ji4314uojhj4pfrufu", 3.8); rowsErr != nil {
		return fmt.Errorf("could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 2, "2009-01-02T15:04:05Z", "#Casino/201", "http://git.casino.com/mr/i2ji4314uhiouhi4124l", 6.2); rowsErr != nil {
		return fmt.Errorf("could not seed rows data: %w", rowsErr)
	}

	if _, rowsErr := s.db.Exec(rowsSeed, 2, "2009-01-02T18:04:05Z", "#Casino/199", "http://git.casino.com/mr/i2ji4314uokjskdjsdiiii", 1.2); rowsErr != nil {
		return fmt.Errorf("could not seed rows data: %w", rowsErr)
	}

	return nil
}
