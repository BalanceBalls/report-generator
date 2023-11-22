package gitlab

import (
	"time"
)

type EventsReq struct {
	Before    time.Time
	After     time.Time
	UserId    int64
	UserToken string
}

type Event struct {
	ProjectId   int       `json:"project_id"`
	ActionName  string    `json:"action_name"`
	TargetType  string    `json:"target_type"`
	TargetIid   int       `json:"target_iid"`   // ID of MR
	TargetTitle string    `json:"target_title"` // Merge request title
	CreatedAt   time.Time `json:"created_at"`

	PushData struct {
		Action      string `json:"action"`
		RefType     string `json:"ref_type"`
		CommitFrom  string `json:"commit_from"`  // previous commit hash
		CommitTo    string `json:"commit_to"`    // current commit hash
		Ref         string `json:"ref"`          // branch name
		CommitTitle string `json:"commit_title"` // commit message
	} `json:"push_data"`

	MR *MergeRequest `json:"-"`
}

type MergeRequest struct {
	Title        string `json:"title"`
	Id           int    `json:"iid"`
	State        string `json:"state"`
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
	WebUrl       string `json:"web_url"`
	IssueUrl     string `json:"target_title"`
}

type Commit struct {
	Id      string `json:"id"`
	ShortId string `json:"short_id"`
	WebUrl  string `json:"web_url"`
	Title   string `json:"title"`
}
