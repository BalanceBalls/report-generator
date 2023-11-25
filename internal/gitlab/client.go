package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
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

func (gc *GitlabClient) Events(ctx context.Context, before time.Time, after time.Time) ([]Event, error) {
	ctxUserId, ok := ctx.Value(userCtxKey).(int64)
	if !ok {
		return nil, ErrNoUserInCtx
	}

	path := path.Join("users", fmt.Sprint(ctxUserId), "events")
	params := url.Values{}

	if !after.IsZero() {
		params.Add("after", after.String())
	}

	if !before.IsZero() {
		params.Add("before", before.String())
	}

	eventsData, err := gc.doRequest(ctx, path, params)

	if err != nil {
		return nil, fmt.Errorf("Events get request failed: %w", err)
	}

	var resData []Event

	if err = json.Unmarshal(eventsData, &resData); err != nil {
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}
	return resData, nil
}

func (gc *GitlabClient) MergeRequest(ctx context.Context, projectId int, mrId int) (*MergeRequest, error) {
	path := path.Join("projects", strconv.Itoa(projectId), "merge_requests", strconv.Itoa(mrId))

	res, err := gc.doRequest(ctx, path, nil)

	if err != nil {
		return nil, fmt.Errorf("MergeRequest get request failed: %w", err)
	}

	var resData MergeRequest
	if err = json.Unmarshal(res, &resData); err != nil {
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}

	return &resData, nil
}

func (gc *GitlabClient) Commit(ctx context.Context, projectId int, cHash string) (*Commit, error) {
	path := path.Join("projects", strconv.Itoa(projectId), "repository", "commits", cHash)

	res, err := gc.doRequest(ctx, path, nil)

	if err != nil {
		return nil, fmt.Errorf("Commit get request failed: %w", err)
	}

	var resData Commit
	if err = json.Unmarshal(res, &resData); err != nil {
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}

	return &resData, nil
}

func (gc *GitlabClient) doRequest(ctx context.Context, endpointPath string, params url.Values) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   gc.host,
		Path:   path.Join(gc.basePath, endpointPath),
	}

	ctxToken, ok := ctx.Value(tokenCtxKey).(string)
	if !ok {
		return nil, ErrNoTokenInCtx
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set(tokenHeaderKey, ctxToken)

	if err != nil {
		return nil, fmt.Errorf("Could not construct request: %w", err)
	}

	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	log.Print(req.URL.String())
	res, err := gc.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to query gitlab api (%q) : %w", endpointPath, err)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	return resBody, nil
}
