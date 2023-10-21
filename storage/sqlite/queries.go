package sqlite

const (
  createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
  Id INT PRIMARY KEY,
  UserEmail TEXT,
  UserToken TEXT,
  IsActive BIT
  )`

  createReportsTable = `
CREATE TABLE IF NOT EXISTS reports (
  Id INT PRIMARY KEY,
  UserId INT
  )`

  createRowsTable = `
CREATE TABLE IF NOT EXISTS rows (
  ReportId INT,
  Date TEXT,
  Task TEXT,
  Link TEXT,
  TimeSpent REAL
  )`

  getFullUsers = `
SELECT 
  u.Id, u.UserEmail, u.UserToken, u.IsActive,
  r.Id, r.UserId, 
  ro.ReportId, ro.Date, ro.Task, ro.Link, ro.TimeSpent
FROM users u 
  INNER JOIN reports r on r.UserId = u.Id
  INNER JOIN rows ro on ro.ReportId = r.Id
LIMIT ?
OFFSET ?`

  getUserById = `
SELECT 
  Id, UserEmail, UserToken, IsActive 
FROM users 
WHERE Id = ?
  `
)
