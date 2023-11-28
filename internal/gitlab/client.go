package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/BalanceBalls/report-generator/internal/logger"
	"github.com/BalanceBalls/report-generator/internal/report"
)

const tokenHeaderKey = "PRIVATE-TOKEN"

type GitlabClient struct {
	host     string
	basePath string
	client   http.Client
}

type EventsResponse struct {
	Events []Event
	Err    error
}

func NewClient(host string, basePath string) *GitlabClient {
	return &GitlabClient{
		host:     host,
		basePath: basePath,
		client:   http.Client{},
	}
}

func (gc *GitlabClient) Events(ctx context.Context, user report.User, before time.Time, after time.Time) ([]Event, error) {
	logger := logger.GetFromContext(ctx)
	path := path.Join("users", fmt.Sprint(user.GitlabId), "events")
	params := url.Values{}

	if !after.IsZero() {
		params.Add("after", after.String())
	}

	if !before.IsZero() {
		params.Add("before", before.String())
	}

	eventsData, err := gc.doRequest(ctx, user.UserToken, path, params)

	if err != nil {
		logger.ErrorContext(ctx, "request failed", "error", err)
		return nil, fmt.Errorf("Events get request failed: %w", err)
	}

	var resData []Event

	if err = json.Unmarshal(eventsData, &resData); err != nil {
		logger.ErrorContext(ctx, "response parsing failed", "error", err)
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}
	return resData, nil
}

func (gc *GitlabClient) MergeRequest(ctx context.Context, user report.User, projectId int, mrId int) (*MergeRequest, error) {
	logger := logger.GetFromContext(ctx)
	path := path.Join("projects", strconv.Itoa(projectId), "merge_requests", strconv.Itoa(mrId))

	res, err := gc.doRequest(ctx, user.UserToken, path, nil)

	if err != nil {
		logger.ErrorContext(ctx, "request failed", "error", err)
		return nil, fmt.Errorf("MergeRequest get request failed: %w", err)
	}

	var resData MergeRequest
	if err = json.Unmarshal(res, &resData); err != nil {
		logger.ErrorContext(ctx, "response parsing failed", "error", err)
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}

	return &resData, nil
}

func (gc *GitlabClient) Commit(ctx context.Context, user report.User, projectId int, cHash string) (*Commit, error) {
	logger := logger.GetFromContext(ctx)
	path := path.Join("projects", strconv.Itoa(projectId), "repository", "commits", cHash)

	res, err := gc.doRequest(ctx, user.UserToken, path, nil)

	if err != nil {
		logger.ErrorContext(ctx, "request failed", "error", err)
		return nil, fmt.Errorf("Commit get request failed: %w", err)
	}

	var resData Commit
	if err = json.Unmarshal(res, &resData); err != nil {
		logger.ErrorContext(ctx, "response parsing failed", "error", err)
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}

	return &resData, nil
}

func (gc *GitlabClient) doRequest(ctx context.Context, token string, endpointPath string, params url.Values) ([]byte, error) {
	logger := logger.GetFromContext(ctx)
	u := url.URL{
		Scheme: "https",
		Host:   gc.host,
		Path:   path.Join(gc.basePath, endpointPath),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Could not construct request: %w", err)
	}

	req.Header.Set(tokenHeaderKey, token)
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	res, err := gc.client.Do(req)

	if err != nil {
		logger.ErrorContext(ctx, "http request failed", "error", err)
		return nil, fmt.Errorf("Failed to query gitlab api (%q) : %w", endpointPath, err)
	}

	logger.InfoContext(ctx, "http request finished",
		"request_url", res.Request.URL.String(),
		"status_code", res.StatusCode)

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("response status code does not indicate success: %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		logger.ErrorContext(ctx, "response body read failed", "error", err)
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	return resBody, nil
}
