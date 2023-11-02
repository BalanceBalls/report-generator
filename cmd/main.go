package main

import (
	// "encoding/json"
	"fmt"
	"go/build"
	"os"
	"time"

	"github.com/BalanceBalls/report-generator/internal/clients/gitlab"
	// htmlgenerator "github.com/BalanceBalls/report-generator/internal/generator/html"
	// "github.com/BalanceBalls/report-generator/internal/storage"
	// "github.com/BalanceBalls/report-generator/internal/storage/sqlite"
)

func main() {
	// Add telegram client
	// Add viper as config util

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	fmt.Println(gopath)

/*
	db, err := sqlite.New("test.sqlite")
	if err != nil {
		fmt.Println(err)
	}

	if err = db.Up(); err != nil {
		fmt.Println(err)
	}

	if err = db.Seed(); err != nil {
		fmt.Println(err)
	}

	users, err := db.Users()

	if err != nil {
		fmt.Println(err)
	}

	convUser := storage.ConvertableUsers{
		Users: users,
	}

	nestedUsers := convUser.Convert()

	htmlGen := htmlgenerator.New("reports", "html-report.tmpl")
	for _, nu := range nestedUsers {
		_, err := json.Marshal(&nu)

		if err != nil {
			panic(err)
		}

		if _, hErr := htmlGen.Generate(nu.Reports[0]); hErr != nil {
			fmt.Println(hErr)
		}

		// fmt.Println("New USER")
		// fmt.Println(string(tJson))
	}
*/

	const gitHost = "gitlab.com"
	const gitBasePath = "api/v4"
	const token = ""

	gitClient := gitlab.New(gitHost, gitBasePath)
	eventsReq := gitlab.EventsReq{
		Before:    time.Time{},
		After:     time.Time{},
		UserId:    0,
		UserToken: token,
	}

	events, err := gitClient.Events(eventsReq)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(events)
}
