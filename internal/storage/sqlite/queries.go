package sqlite

const (
	createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
  id								INT PRIMARY KEY,
	gitlab_id         INT,
  user_email 				TEXT,
  user_token 				TEXT,
	timezone_offset 	INT,
  is_active 				BIT
)`

	createReportsTable = `
CREATE TABLE IF NOT EXISTS reports (
  id			INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id	INT,

	FOREIGN KEY(user_id) REFERENCES users(id)
)`

	createRowsTable = `
CREATE TABLE IF NOT EXISTS rows (
  report_id		INT,
  date				TEXT,
  task				TEXT,
  link				TEXT,
  time_spent	REAL,

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
LIMIT ?
OFFSET ?`

	getUserById = `
SELECT 
  id, gitlab_id, user_email, user_token, is_active 
FROM users 
WHERE id = ?
  `

	addUser = `
INSERT INTO users (id, gitlab_id, user_email, user_token, timezone_offset, is_active) 
VALUES (?, ?, ?, ?, ?, ?)
	`

	updateUser = `
UPDATE users SET 
	gitlab_id = ?,
	user_email = ?,
	user_token = ?,
	timezone_offset = ?
WHERE id = ?
	`

	removeUser = `
DELETE FROM users
WHERE id = ?
	`

  checkUserExists = `
SELECT 1 FROM users
WHERE id = ?
	`
)
