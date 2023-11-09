package builder

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/BalanceBalls/report-generator/internal/clients/gitlab"
	"github.com/BalanceBalls/report-generator/internal/storage"
	"golang.org/x/exp/slices"
)

// Action names
const (
	commit             = "pushed to"
	acceptMergeRequest = "accepted"
	createMergeRequest = "opened"
	initCommit         = "pushed new"
)

// Target types
const (
	mergeRequestTarget = "MergeRequest"
)

var branches2exclude = []string{"main", "master", "develop"}
var trackedActions = []string{initCommit, commit, createMergeRequest, acceptMergeRequest}

type GitlabBuilder struct {
	client    gitlab.GitlabClient
	userId    int
	userToken string
	tzOffset  int
}

func New(client gitlab.GitlabClient, userId int, userToken string, tz int) *GitlabBuilder {
	return &GitlabBuilder{
		client:    client,
		userId:    userId,
		userToken: userToken,
		tzOffset:  tz,
	}
}

func (gb *GitlabBuilder) Build() (storage.Report, error) {
	result := storage.Report{
		UserId: gb.userId,
	}

	// Current time with the server's offset
	pointOfReference := time.Now().UTC().Add(time.Minute * time.Duration(gb.tzOffset)) // .Add(time.Hour * 24 * -10)
	// Get start time of the current day
	timeRangeStart := pointOfReference.Truncate(time.Hour * 24)
	// Set time range to a whole day
	timeRangeEnd := timeRangeStart.Add(time.Hour * 24)

	eventsReq := gitlab.EventsReq{
		Before:    pointOfReference.AddDate(0, 0, 1),
		After:     pointOfReference.AddDate(0, 0, -1),
		UserId:    gb.userId,
		UserToken: gb.userToken,
	}

	events, err := gb.client.Events(eventsReq)

	if err != nil {
		return storage.Report{}, fmt.Errorf("failed to fetch gitlab data: %w", err)
	}

	filteredEvents, err := gb.filterByTime(events, timeRangeStart, timeRangeEnd)
	if err != nil {
		return storage.Report{}, err
	}

	filteredEvents, err = gb.filterByBranches(events)
	if err != nil {
		return storage.Report{}, err
	}

	filteredEvents, err = gb.filterByActions(events)
	if err != nil {
		return storage.Report{}, err
	}

	filteredEvents, err = gb.loadMergeRequests(events)
	if err != nil {
		return storage.Report{}, err
	}

	slices.SortFunc(filteredEvents, func(i, j gitlab.Event) int {
		return i.CreatedAt.Compare(j.CreatedAt)
	})

	branch2events := gb.groupByBranches(filteredEvents)

	if err != nil {
		return storage.Report{}, fmt.Errorf("failed to group events by branches: %w", err)
	}

	var prevTime = filteredEvents[0].CreatedAt
	prevTime = gb.initPrevTime(branch2events, prevTime)

	for k, v := range branch2events {
		row := gb.buildRow(k, v, prevTime)
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

func (gb *GitlabBuilder) initPrevTime(branch2events map[string][]gitlab.Event, defaultValue time.Time) time.Time {
	// Only one row in report
	if len(branch2events) == 1 {
		for _, events := range branch2events{
			fullWorkDay := events[0].CreatedAt.Add(time.Hour * -8) 

			// When a single branch contains MR
			hasMr, _ := gb.tryGetMrForBranch(events)
			if hasMr {
				return fullWorkDay 
			}

			// When a single branch has only one commit
			commitCount := gb.getCommitsCountForBranch(events)
			if commitCount == 1 {
				return fullWorkDay
			}

			// When multiple commits in a single branch 
			return events[0].CreatedAt
		}	
	}

	return defaultValue
}

func (gb *GitlabBuilder) filterByTime(events []gitlab.Event, start time.Time, end time.Time) ([]gitlab.Event, error) {
	result := []gitlab.Event{}

	for _, e := range events {
		if e.CreatedAt.After(start) && e.CreatedAt.Before(end) {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func (gb *GitlabBuilder) filterByActions(events []gitlab.Event) ([]gitlab.Event, error) {
	result := []gitlab.Event{}

	for _, e := range events {
		if slices.Contains(trackedActions, e.ActionName) {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func (gb *GitlabBuilder) filterByBranches(events []gitlab.Event) ([]gitlab.Event, error) {
	result := []gitlab.Event{}

	for _, e := range events {
		if !slices.Contains(branches2exclude, e.PushData.Ref) {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func (gb *GitlabBuilder) loadMergeRequests(events []gitlab.Event) ([]gitlab.Event, error) {
	for _, event := range events {
		if event.TargetType == mergeRequestTarget {
			mr, err := gb.client.MergeRequest(event.ProjectId, event.TargetIid, gb.userToken)

			if err != nil {
				return nil, fmt.Errorf("could not get MR data: %w", err)
			}

			event.MR = mr
		}
	}

	return events, nil
}

func (gb *GitlabBuilder) groupByBranches(events []gitlab.Event) map[string][]gitlab.Event {
	var branch2events = make(map[string][]gitlab.Event)

	for _, event := range events {
		branchName := event.PushData.Ref
		branch2events[branchName] = append(branch2events[branchName], event)
	}

	return branch2events
}

func (gb *GitlabBuilder) buildRow(branchName string, branchEvents []gitlab.Event, prevTime time.Time) storage.ReportRow {
	var taskName string
	var taskLink string

	hasMr, mergeRequest := gb.tryGetMrForBranch(branchEvents)

	if hasMr {
		// If a branch has an MR
		taskName = mergeRequest.Title
		taskLink = mergeRequest.WebUrl
	} else {
		// If no MR for a branch - set branch name as task name
		taskName = branchEvents[0].PushData.Ref

		// Use links to all commits for today as a link to the task links
		links, err := gb.getCommitLinks(branchEvents)
		if err != nil {
			taskLink = "Failed to get commits"
		}
		taskLink = strings.Join(links, "\n")
	}

	// TODO: Cover following cases
	// 1. When it is first event (prevTime is empty)
	// 2. When a branch has only one commit (calculate work hours from prevTime)
	// 3. When it is first event and only event for the day
	timeSpent := branchEvents[len(branchEvents)-1].CreatedAt.Sub(prevTime).Hours()

	result := storage.ReportRow{
		ReportId: 0,
		Date:     branchEvents[0].CreatedAt,

		// If no MR for a branch, use branchName
		// Otherwise use MR title
		Task: taskName,

		// If no MR for a branch - include links to all commits for today for that branch
		// Otherwise just link to MR
		Link:      taskLink,
		TimeSpent: float32(timeSpent),
	}

	return result
}

func (gb *GitlabBuilder) getCommitLinks(branchEvents []gitlab.Event) ([]string, error) {
	var result []string

	commitInfo, err := gb.client.Commit(
		branchEvents[0].ProjectId,
		branchEvents[0].PushData.CommitTo,
		gb.userToken)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit info: %w", err)
	}

	hash := path.Base(commitInfo.WebUrl)
	commitBaseUrl := strings.ReplaceAll(commitInfo.WebUrl, hash, "")

	for _, event := range branchEvents {
		url := commitBaseUrl + event.PushData.CommitTo
		result = append(result, url)
	}

	return result, nil
}

func (gb *GitlabBuilder) tryGetMrForBranch(branchEvents []gitlab.Event) (bool, *gitlab.MergeRequest) {
	for _, event := range branchEvents {
		if event.MR != nil {
			return true, event.MR
		}
	}

	return false, nil
}

func (gb *GitlabBuilder) getCommitsCountForBranch(branchEvents []gitlab.Event) int {
	result := 0
	for _, event := range branchEvents {
		if event.ActionName == initCommit || event.ActionName == commit {
			result++
		}
	}

	return result
}
