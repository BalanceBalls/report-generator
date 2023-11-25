package main

import (
	"context"

	"github.com/BalanceBalls/report-generator/internal/bot"
)

func main() {
	// Add telegram client
	// Add viper as config util
	// Add context to io calls
	// Add logger
	// Add tests

	/*
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				gopath = build.Default.GOPATH
			}
			fmt.Println(gopath)
	*/

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot := bot.New("")
	bot.Serve(ctx)
}
