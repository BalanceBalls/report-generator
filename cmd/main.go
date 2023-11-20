package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BalanceBalls/report-generator/internal/builder"
	"github.com/BalanceBalls/report-generator/internal/clients/gitlab"
	htmlgenerator "github.com/BalanceBalls/report-generator/internal/generator/html"
)

func main() {
	// Add telegram client
	// Add viper as config util
	// Add migrator tool
	// Add context to io calls
	// Add logger
	// Add tests

	/*
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
	*/

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	const gitHost = "localhost:4443" //"gitlab.com"
	const gitBasePath = "api/v4"
	const token = "glpat-wEp2SkQS_Yvr7vgDyt7A"
	const userId = 18375700

	gitClient := gitlab.New(gitHost, gitBasePath)
	gBuilder := builder.New(*gitClient, userId, token, 0)

	respch := make(chan builder.BuildResult)
	go gBuilder.Build(ctx, respch)

	select {
	case <-ctx.Done():
		panic(ctx.Err())
	case response := <-respch:
		if response.Err != nil {
			panic(response.Err)
		}

		htmlGen := htmlgenerator.New("reports", "html_report.tmpl")
		if _, hErr := htmlGen.Generate(response.Report); hErr != nil {
			fmt.Println(hErr)
		}
	}
}
