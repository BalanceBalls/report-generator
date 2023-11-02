package gitlab

import "time"

// Action names
const (
	commit             = "pushed to"
	acceptMergeRequest = "accepted"
	createMergeRequest = "opened"
	initCommit         = "pushed new"
)

// Target types
const (
	mergeRequest = "MergeRequest"
)

type EventsReq struct {
	Before    time.Time
	After     time.Time
	UserId    int
	UserToken string
}

type Event struct {
	ProjectId   int       `json:"project_id"`
	ActionName  string    `json:"action_name"`
	TargetType  string    `json:"target_type"`
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
}
