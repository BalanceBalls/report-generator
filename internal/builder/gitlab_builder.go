package builder

import (
	"errors"
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

var trackedActions = []string{initCommit, commit, createMergeRequest, acceptMergeRequest}

var ErrNoGitlabActions = errors.New("No gitlab actions to report found for current day")

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

func (gb *GitlabBuilder) Build() (*storage.Report, error) {
	result := storage.Report{
		UserId: gb.userId,
	}

	// Current time with the server's offset
	pointOfReference := time.Now().Add(time.Minute * time.Duration(gb.tzOffset))
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
		return nil, fmt.Errorf("Failed to fetch gitlab data: %w", err)
	}

	filteredEvents, err := gb.filterByTime(events, timeRangeStart, timeRangeEnd)

	if err != nil {
		return nil, err
	}

	branch2events, err := gb.groupByBranches(filteredEvents)

	if err != nil {
		return nil, fmt.Errorf("Failed to group events by branches: %w", err)
	}

	for k, v := range branch2events {
		row := gb.buildRow(k, v)
		result.Rows = append(result.Rows, row)
	}

	return &result, nil
}

func (gb *GitlabBuilder) filterByTime(events []gitlab.Event, start time.Time, end time.Time) ([]gitlab.Event, error) {
	result := []gitlab.Event{}

	for _, e := range events {
		if e.CreatedAt.After(start) && e.CreatedAt.Before(end) {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitlabActions
	}

	return result, nil
}

func (gb *GitlabBuilder) groupByBranches(events []gitlab.Event) (map[string][]gitlab.Event, error) {
	var branch2events map[string][]gitlab.Event

	for _, event := range events {
		if slices.Contains(trackedActions, event.ActionName) {

			if event.TargetType == mergeRequestTarget {
				mr, err := gb.client.MergeRequest(event.ProjectId, event.TargetIid, gb.userToken)

				if err != nil {
					return nil, fmt.Errorf("Could not get MR data: %w", err)
				}

				event.MR = mr
			}

			branch2events[event.PushData.Ref] = append(branch2events[event.PushData.Ref], event)
		}
	}

	return branch2events, nil
}

func (gb *GitlabBuilder) buildRow(branchName string, branchEvents []gitlab.Event) storage.ReportRow {
	var taskName string
	var taskLink string
	var mergeRequest *gitlab.MergeRequest

	for _, event := range branchEvents {
		if event.MR != nil {
			mergeRequest = event.MR
		}
	}

	if mergeRequest != nil {
		taskName = mergeRequest.Title
		taskLink = mergeRequest.WebUrl
	} else {
		taskName = branchEvents[0].PushData.Ref

		links, err := gb.getCommitLinks(branchEvents)
		if err != nil {
			taskLink = "Failed to get commits"
		}
		taskLink = strings.Join(links, ", \n")
	}

	timeSpent := branchEvents[len(branchEvents)-1].CreatedAt.Sub(branchEvents[0].CreatedAt)

	result := storage.ReportRow{
		ReportId: 0,
		Date:     branchEvents[0].CreatedAt,

		// If no MR for a branch, use branchName
		// Otherwise use MR title
		Task: taskName,

		// If no MR for a branch - include links to all commits for today for that branch
		// Otherwise just link to MR
		Link: taskLink,

		// In 4.3h format
		TimeSpent: float32(timeSpent.Minutes() / time.Hour.Minutes()),
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
		return nil, fmt.Errorf("Failed to fetch commit info: %w", err)
	}

	hash := path.Base(commitInfo.WebUrl)
	commitBaseUrl := strings.ReplaceAll(commitInfo.WebUrl, hash, "")

	for _, event := range branchEvents {
		url := commitBaseUrl + event.PushData.CommitTo
		result = append(result, url)
	}

	return result, nil
}
