package storage

import (
	"time"

	"github.com/BalanceBalls/report-generator/internal/report"
)

type FlatUser struct {
	Id             int64
	GitlabId       int
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

func (cu *ConvertableUsers) Convert() []report.User {
	var result []report.User

	flatUserMap := make(map[[2]int64][]FlatUser)

	for _, flatUser := range cu.Users {
		k := [2]int64{flatUser.Id, flatUser.ReportId}
		flatUserMap[k] = append(flatUserMap[k], flatUser)
	}

	for k, v := range flatUserMap {
		tUser := report.User{
			Id:             k[0],
			GitlabId:       v[0].GitlabId,
			UserEmail:      v[0].UserEmail,
			UserToken:      v[0].UserToken,
			TimezoneOffset: v[0].TimezoneOffset,
			IsActive:       v[0].IsActive,
		}

		tReport := report.Report{
			Id:     k[1],
			UserId: k[0],
		}

		for _, fu := range v {
			tRow := report.ReportRow{
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
