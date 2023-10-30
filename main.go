package main

import (
	"encoding/json"
	"fmt"

	htmlgenerator "github.com/BalanceBalls/report-generator/generator/html"
	"github.com/BalanceBalls/report-generator/storage"
	"github.com/BalanceBalls/report-generator/storage/sqlite"
)

func main() {
	// Add telegram client
	// Add gitlab client
	// Add viper as config util

	db, err := sqlite.New("test.sqlite")
	if err != nil {
		fmt.Println(err)
	}

	if err = db.Up(); err != nil {
		fmt.Println(err)
	}

	 // if err = db.Seed(); err != nil {
		//  fmt.Println(err)
	 // }

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
}
