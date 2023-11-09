package builder

import (
	"errors"

	"github.com/BalanceBalls/report-generator/internal/storage"
)

var ErrNoGitActions = errors.New("no gitlab actions to report found for current day")

type Builder interface {
	Build() (storage.Report, error)
}

