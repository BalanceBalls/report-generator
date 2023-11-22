package storage

import (
	"context"
	"time"
)

type Storage interface {
	Users(ctx context.Context) ([]FlatUser, error)
	User(ctx context.Context, userId int64) (*User, error)
	AddUser(ctx context.Context, user User) error
	UserExists(ctx context.Context, userId int64) bool
	UpdateUser(ctx context.Context, user User) error
	RemoveUser(ctx context.Context, userId int64) error
} 

type User struct {
	Id        int64  `json:"id"`
	UserEmail string `json:"userEmail"`
	UserToken string `json:"userToken"`
	IsActive  bool   `json:"isActive"`

	// Minutes from UTC
	TimezoneOffset int      `json:"timezoneOffset"`
	Reports        []Report `json:"reports"`
}

type Report struct {
	Id     int64       `json:"reportId"`
	UserId int64       `json:"reportUserId"`
	Rows   []ReportRow `json:"rows"`
}

type ReportRow struct {
	ReportId  int64     `json:"rowReportId"`
	Date      time.Time `json:"date"`
	Task      string    `json:"task"`
	Link      string    `json:"link"`
	TimeSpent float32   `json:"timeSpent"`
}

type FlatUser struct {
	Id             int64
	UserEmail      string
	UserToken      string
	TimezoneOffset int
	IsActive       bool
	ReportId       int64
	UserId         int64
	ReportRowId    int64
	Date           time.Time
	Task           string
	Link           string
	TimeSpent      float32
}

type ConvertableUsers struct {
	Users []FlatUser
}

func (cu *ConvertableUsers) Convert() []User {
	var result []User

	flatUserMap := make(map[[2]int64][]FlatUser)

	for _, flatUser := range cu.Users {
		k := [2]int64{flatUser.Id, flatUser.ReportId}
		flatUserMap[k] = append(flatUserMap[k], flatUser)
	}

	for k, v := range flatUserMap {
		tUser := User{
			Id:             k[0],
			UserEmail:      v[0].UserEmail,
			UserToken:      v[0].UserToken,
			TimezoneOffset: v[0].TimezoneOffset,
			IsActive:       v[0].IsActive,
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
