package postgres

const (
	createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
  id                INTEGER PRIMARY KEY,
  gitlab_id         INTEGER,
  user_email        TEXT,
  user_token        TEXT,
  timezone_offset   INTEGER,
  is_active         BOOLEAN
)`

	createReportsTable = `
CREATE TABLE IF NOT EXISTS reports (
  id      SERIAL PRIMARY KEY,
  user_id INTEGER,

  FOREIGN KEY(user_id) REFERENCES users(id)
)`

	createRowsTable = `
CREATE TABLE IF NOT EXISTS rows (
  report_id   INTEGER,
  date        TEXT,
  task        TEXT,
  link        TEXT,
  time_spent  REAL,

  FOREIGN KEY(report_id) REFERENCES reports(id)
)`

	getFullUsers = `
SELECT 
  u.id, u.gitlab_id, u.user_email, u.user_token, u.is_active,
  r.id, r.user_id, 
  ro.report_id, ro.date, ro.task, ro.link, ro.time_spent
FROM users u 
  INNER JOIN reports r on r.user_id = u.id
  INNER JOIN rows ro on ro.report_id = r.id
LIMIT $1
OFFSET $2`

	getUserById = `
SELECT 
  id, gitlab_id, user_email, user_token, is_active 
FROM users 
WHERE id = $1
  `

	addUser = `
INSERT INTO users (id, gitlab_id, user_email, user_token, timezone_offset, is_active) 
VALUES ($1, $2, $3, $4, $5, $6)
	`

	updateUser = `
UPDATE users SET 
	gitlab_id = $1,
	user_email = $2,
	user_token = $3,
	timezone_offset = $4
WHERE id = $5
	`

	removeUser = `
DELETE FROM users
WHERE id = $1
	`

  checkUserExists = `
SELECT 1 FROM users
WHERE id = $1
	`
)
