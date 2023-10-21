package storage

import (
	"time"
)

type Storage interface {
	Report(userId string) (*Report, error)
	SaveReport(userId string, report *Report) error
	Users() ([]FlatUser, error)
	User(userId string) (*User, error)
	SaveUser(user *User) error
}

type User struct {
	Id        int      `json:"id"`
	UserEmail string   `json:"userEmail"`
	UserToken string   `json:"userToken"`
	IsActive  bool     `json:"isActive"`
	Reports   []Report `json:"reports"`
}

type Report struct {
	Id     int         `json:"reportId"`
	UserId int         `json:"reportUserId"`
	Rows   []ReportRow `json:"rows"`
}

type ReportRow struct {
	ReportId  int       `json:"rowReportId"`
	Date      time.Time `json:"date"`
	Task      string    `json:"task"`
	Link      string    `json:"link"`
	TimeSpent float32   `json:"timeSpent"`
}

type FlatUser struct {
	Id          int
	UserEmail   string
	UserToken   string
	IsActive    bool
	ReportId    int
	UserId      int
	ReportRowId int
	Date        time.Time
	Task        string
	Link        string
	TimeSpent   float32
}

type ConvertableUsers struct {
	Users []FlatUser
}

func (cu *ConvertableUsers) Convert() []User {
	var result []User

	flatUserMap := make(map[[2]int][]FlatUser)

	for _, flatUser := range cu.Users {
		k := [2]int{flatUser.Id, flatUser.ReportId}
		flatUserMap[k] = append(flatUserMap[k], flatUser)
	}

	for k, v := range flatUserMap {
		tUser := User{
			Id:        k[0],
			UserEmail: v[0].UserEmail,
			UserToken: v[0].UserToken,
			IsActive:  v[0].IsActive,
		}

		tReport := Report{
			Id:     k[1],
			UserId: k[0],
		}

		for _, fu := range v {
			tRow := ReportRow{
				ReportId:  fu.ReportRowId,
				Date:      fu.Date,
				Task:      fu.Task,
				Link:      fu.Link,
				TimeSpent: fu.TimeSpent,
			}

			tReport.Rows = append(tReport.Rows, tRow)
		}

		tUser.Reports = append(tUser.Reports, tReport)
		result = append(result, tUser)
	}

	return result
}
