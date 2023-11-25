package report

import "time"

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

type Channel struct {
	Report Report
	Err error
}

type Result struct {
	Name string
	Data []byte
}
