package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/BalanceBalls/report-generator/internal/bot"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func main() {
	// Add tests
	// log backups
	// db backups

	logFile, fileErr := os.Create("bot.log")
	if fileErr != nil {
		panic(fileErr)
	}
	defer func() {
		err := logFile.Close()
		panic(err)
	}()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.InfoContext(ctx, "bot starting...")
	err := godotenv.Load(".env")
	if err != nil {
		slog.ErrorContext(ctx, "error loading .env file: %s", err)
		panic(err)
	}
	cfg := bot.Config{}

	err = env.Parse(&cfg)
	if err != nil {
		slog.ErrorContext(ctx, "unable to parse ennvironment variables: %e", err)
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			slog.ErrorContext(ctx, "panic occurred", "reason", err)
			panic(err)
		}
	}()

	bot := bot.New(&cfg)
	bot.Serve(ctx)
}
