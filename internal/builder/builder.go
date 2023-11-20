package builder

import (
	"context"
	"errors"
)

var ErrNoGitActions = errors.New("no gitlab actions to report found for current day")

type Builder interface {
	Build(ctx context.Context, respch chan BuildResult)
}
