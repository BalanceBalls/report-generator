package main

import (
  "encoding/json"
  "fmt"

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

  /*
    if err = db.Seed(); err != nil {
      fmt.Println(err)
    }
  */

  users, err := db.Users()

  if err != nil {
    fmt.Println(err)
  }

  convUser := storage.ConvertableUsers{
    Users: users,
  }

  nestedUsers := convUser.Convert()

  for _, nu := range nestedUsers {
    tJson, err := json.Marshal(&nu)

    if err != nil {
      panic(err)
    }

    fmt.Println("New USER")
    fmt.Println(string(tJson))
  }
}
