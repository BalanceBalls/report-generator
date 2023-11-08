package main

import (
	"fmt"
	"go/build"
	"os"
	"time"

	"github.com/BalanceBalls/report-generator/internal/storage"
	"github.com/BalanceBalls/report-generator/internal/storage/sqlite"
)

func main() {
	// Add telegram client
	// Add gitlab client
	// Add viper as config util
	// Add migrator tool
  // Add context to io calls
  // Add tests

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	fmt.Println(gopath)

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

	fmt.Println(nestedUsers)
/*
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

	loc := time.FixedZone("", int(time.Minute.Seconds() * 60 * 3))
	fmt.Println(loc)
	/* 
	const gitHost = "gitlab.com"
	const gitBasePath = "api/v4"
	const token = "glpat-wEp2SkQS_Yvr7vgDyt7A"
	
	gitClient := gitlab.New(gitHost, gitBasePath)
	eventsReq := gitlab.EventsReq{
		Before:    time.Now().UTC(),
		After:     time.Now().UTC().Add(-(time.Hour * 24 * 4)),
		UserId:    18375700,
		UserToken: token,
	}
	
	events, err := gitClient.Events(eventsReq)
	
	if err != nil {
		fmt.Println(err)
	}

	json, _ := json.Marshal(events)

	fmt.Println(string(json))

	// Get datetime representing beginning of the current date
	t := time.Now().UTC()
	fmt.Println(t.Truncate(time.Hour * 24))


	fmt.Println("----------------------------")
	tp, _ := time.Parse(time.RFC3339 , "2023-10-27T18:07:50.866+03:00")

	fmt.Println(tp)
	*/
}
