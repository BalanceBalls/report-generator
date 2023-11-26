package main

import (
	"context"
	"log"

	"github.com/BalanceBalls/report-generator/internal/bot"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func main() {
	// Add logger
	// Add tests

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}
	cfg := bot.Config{}

	err = env.Parse(&cfg)
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}

	bot := bot.New(&cfg)
	bot.Serve(ctx)
}
