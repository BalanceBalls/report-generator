package bot

import (
	"context"

	"github.com/BalanceBalls/report-generator/internal/report"
)

type Storage interface {
	User(ctx context.Context, userId int64) (report.User, error)
	AddUser(ctx context.Context, user report.User) error
	UserExists(ctx context.Context, userId int64) bool
	UpdateUser(ctx context.Context, user report.User) error
	RemoveUser(ctx context.Context, userId int64) error
	Up(ctx context.Context) error
}

type Builder interface {
	Build(ctx context.Context, user report.User, respch chan report.Channel)
}

type Generator interface {
	Generate(report report.Report) (report.Result, error)
}
