package gitlab

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/BalanceBalls/report-generator/internal/report"
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
	client GitlabClient
}

func NewReportBuilder(client GitlabClient) *GitlabBuilder {
	return &GitlabBuilder{
		client: client,
	}
}

func (gb *GitlabBuilder) Build(ctx context.Context, respch chan report.Channel) {
	result := report.Report{
		UserId: ctx.Value(userCtxKey).(int64),
	}

	// Current time with the server's offset
	pointOfReference := time.Date(2023, 10, 05, 03, 34, 58, 651387237, time.UTC) //time.Now().UTC().Add(time.Minute * time.Duration(gb.tzOffset))
	// Get start time of the current day
	timeRangeStart := pointOfReference.Truncate(time.Hour * 24)
	// Set time range to a whole day
	timeRangeEnd := timeRangeStart.Add(time.Hour * 24)

	before := pointOfReference.AddDate(0, 0, 1)
	after := pointOfReference.AddDate(0, 0, -1)

	events, err := gb.client.Events(ctx, before, after)
	handleErr(err, respch)

	var filteredEvents []Event
	filteredEvents, err = filterByTime(events, timeRangeStart, timeRangeEnd)
	handleErr(err, respch)

	filteredEvents, err = filterByBranches(filteredEvents)
	handleErr(err, respch)

	filteredEvents, err = filterByActions(filteredEvents)
	handleErr(err, respch)

	filteredEvents, err = gb.loadMergeRequests(ctx, filteredEvents)
	handleErr(err, respch)

	sortEvents(filteredEvents)
	branch2events := groupByBranches(filteredEvents)

	// Get branches ordered by first event time
	orderedBranches := sortBranches(branch2events)

	var prevTime = filteredEvents[0].CreatedAt
	prevTime = initPrevTime(branch2events, prevTime)

	for _, branchName := range orderedBranches {
		events := branch2events[branchName]
		row := gb.buildRow(ctx, branchName, events, prevTime)
		result.Rows = append(result.Rows, row)

		// Use time of the last event for the branch
		// as a backup staring point for the next branch
		prevTime = events[len(events)-1].CreatedAt
	}

	respch <- report.Channel{
		Report: result,
		Err:    nil,
	}
}

func (gb *GitlabBuilder) loadMergeRequests(ctx context.Context, events []Event) ([]Event, error) {
	loaded := map[int]MergeRequest{}
	for i, event := range events {

		tmr, isLoaded := loaded[event.TargetIid]
		if isLoaded {
			events[i].MR = &tmr
			continue
		}

		if event.TargetType == mergeRequestTarget {
			mr, err := gb.client.MergeRequest(ctx, event.ProjectId, event.TargetIid)

			if err != nil {
				return nil, fmt.Errorf("could not get MR data: %w", err)
			}

			events[i].MR = mr
			loaded[event.TargetIid] = *mr
		}
	}

	return events, nil
}

func (gb *GitlabBuilder) buildRow(
	ctx context.Context, branchName string, branchEvents []Event, prevTime time.Time,
) report.ReportRow {

	var taskName string
	var taskLink string
	var actionLinks []string

	fmt.Println("Building row for:", branchName)
	hasMr, mergeRequest := tryGetMrForBranch(branchEvents)

	if hasMr {
		// If a branch has an MR
		taskName = mergeRequest.IssueUrl
		mrLinks := getMergeRequestLinks(branchEvents)
		actionLinks = append(actionLinks, mrLinks...)
	} else {
		// If no MR for a branch - set branch name as task name
		taskName = branchEvents[0].PushData.Ref
	}

	commitLinks, err := gb.getCommitLinks(ctx, branchEvents)
	if err != nil {
		taskLink = "Failed to get commits"
	} else {
		actionLinks = append(actionLinks, commitLinks...)
	}

	taskLink = strings.Join(actionLinks, " \n ")
	timeSpent := getHoursSpentOnBranch(prevTime, branchEvents)

	result := report.ReportRow{
		ReportId: 0,
		Date:     branchEvents[0].CreatedAt,

		// If no MR for a branch, use branchName
		// Otherwise use a link to an issue
		Task:      taskName,
		Link:      taskLink,
		TimeSpent: float32(timeSpent),
	}

	return result
}

func (gb *GitlabBuilder) getCommitLinks(ctx context.Context, branchEvents []Event) ([]string, error) {
	var result []string

	var firstCommit Event

	for _, event := range branchEvents {
		if event.MR == nil {
			firstCommit = event
			break
		}
	}

	// Get info about any single commit in order to
	// acquire base commit URL which will be used for other commits
	commitInfo, err := gb.client.Commit(
		ctx,
		firstCommit.ProjectId,
		firstCommit.PushData.CommitTo)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit info: %w", err)
	}

	// Construct base commit url by removing
	// commit hash from the endpoint path
	hash := path.Base(commitInfo.WebUrl)
	commitBaseUrl := strings.ReplaceAll(commitInfo.WebUrl, hash, "")

	for _, event := range branchEvents {
		if event.MR == nil {
			url := commitBaseUrl + event.PushData.CommitTo
			result = append(result, url)
		}
	}

	return result, nil
}

func sortEvents(events []Event) []Event {
	slices.SortFunc(events, func(i, j Event) int {
		return i.CreatedAt.Compare(j.CreatedAt)
	})

	return events
}

// Rerutns a slice of strings which represents
// an ordered by time array of branch names
func sortBranches(branch2events map[string][]Event) []string {
	result := make([]string, 0, len(branch2events))
	time2branch := make(map[time.Time]string, len(branch2events))
	sortedTime := make([]time.Time, 0, len(branch2events))

	for k, v := range branch2events {
		tempEvents := sortEvents(v)
		minTime := tempEvents[0].CreatedAt

		sortedTime = append(sortedTime, minTime)
		time2branch[minTime] = k
	}

	slices.SortFunc(sortedTime, func(i, j time.Time) int {
		return i.Compare(j)
	})

	for _, t := range sortedTime {
		result = append(result, time2branch[t])
	}

	return result
}

// Get a time point from which to calculate
// working hours for different cases
func initPrevTime(branch2events map[string][]Event, defaultValue time.Time) time.Time {
	// Only one row in report
	if len(branch2events) == 1 {
		for _, events := range branch2events {
			// A point in time which will be 8 hrs prior to a git action by user
			fullWorkDay := events[0].CreatedAt.Add(time.Hour * -8)

			// When a single branch contains MR
			hasMr, _ := tryGetMrForBranch(events)
			if hasMr {
				return fullWorkDay
			}

			// When a single branch has only one commit
			commitCount := getCommitsCountForBranch(events)
			if commitCount == 1 {
				return fullWorkDay
			}

			// When multiple commits in a single branch
			return events[0].CreatedAt
		}
	}

	return defaultValue
}

func filterByTime(events []Event, start time.Time, end time.Time) ([]Event, error) {
	result := []Event{}

	for _, event := range events {
		if event.CreatedAt.After(start) && event.CreatedAt.Before(end) {
			result = append(result, event)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func filterByActions(events []Event) ([]Event, error) {
	result := []Event{}

	for _, event := range events {
		if slices.Contains(trackedActions, event.ActionName) {
			result = append(result, event)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func filterByBranches(events []Event) ([]Event, error) {
	result := []Event{}

	for _, event := range events {
		branchName := event.PushData.Ref
		if branchName == "" || !slices.Contains(branches2exclude, branchName) {
			result = append(result, event)
		}
	}

	if len(result) == 0 {
		return nil, ErrNoGitActions
	}

	return result, nil
}

func groupByBranches(events []Event) map[string][]Event {
	var branch2events = make(map[string][]Event)

	for _, event := range events {
		var branchName string
		if event.MR != nil {
			// Name of the branch being merged into other branch
			branchName = event.MR.SourceBranch
		} else {
			branchName = event.PushData.Ref
		}
		branch2events[branchName] = append(branch2events[branchName], event)
	}

	return branch2events
}

// Calculates hours spent working in a branch
func getHoursSpentOnBranch(prevTime time.Time, events []Event) float64 {
	var timeSpent float64

	// If last branch instersects time of current
	if prevTime.After(events[0].CreatedAt) {
		// If intersects partially, calculate delta
		if prevTime.Before(events[len(events)-1].CreatedAt) {
			return events[len(events)-1].CreatedAt.Sub(prevTime).Hours()
		} else {
			return 0
		}
	}

	for i := 1; i < len(events); i++ {
		if events[i-1].CreatedAt.Before(events[i].CreatedAt) {
			timeSpent += events[i].CreatedAt.Sub(events[i-1].CreatedAt).Hours()
		}
	}

	return timeSpent
}

func getMergeRequestLinks(branchEvents []Event) []string {
	var result []string

	for _, event := range branchEvents {
		if event.MR != nil {
			link := event.MR.WebUrl
			if !slices.Contains(result, link) {
				result = append(result, link)
			}
		}
	}

	return result
}

func tryGetMrForBranch(branchEvents []Event) (bool, *MergeRequest) {
	for _, event := range branchEvents {
		if event.MR != nil {
			return true, event.MR
		}
	}

	return false, nil
}

func getCommitsCountForBranch(branchEvents []Event) int {
	result := 0
	for _, event := range branchEvents {
		if event.ActionName == initCommit || event.ActionName == commit {
			result++
		}
	}

	return result
}

func handleErr(err error, respch chan report.Channel) {
	if err != nil {
		log.Print(err)

		respch <- report.Channel{
			Report: report.Report{},
			Err:    err,
		}
	}
}
