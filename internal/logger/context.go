package logger

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

func AddToContext(ctx context.Context, ctxLogger *slog.Logger) context.Context {
		return context.WithValue(ctx, loggerKey{}, ctxLogger)
}

func GetFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		slog.WarnContext(ctx, "could not extract logger from context")
		logger = slog.Default()
	}

	return logger
}
