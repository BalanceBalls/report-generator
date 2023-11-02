package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

const tokenHeaderKey = "PRIVATE-TOKEN"

type GitlabClient struct {
	host      string
	basePath  string
	client    http.Client
}

func New(host string, basePath string) *GitlabClient {
	return &GitlabClient{
		host:     host,
		basePath: basePath,
		client:   http.Client{},
	}
}

func (gc *GitlabClient) Events(req EventsReq) ([]Event, error) {
	path := path.Join("users", strconv.Itoa(req.UserId), "events")

	res, err := gc.doRequest(path, nil, req.UserToken)

	if err != nil {
		return nil, fmt.Errorf("Events request failed: %w", err)
	}
	
	var resData []Event
	
	if err = json.Unmarshal(res, &resData); err != nil {
		return nil, fmt.Errorf("Could not parse response data: %w", err)
	}

	return resData, nil
}

func (gc *GitlabClient) doRequest(endpointPath string, params url.Values, token string) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   gc.host,
		Path:   path.Join(gc.basePath, endpointPath),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set(tokenHeaderKey, token)

	if err != nil {
		return nil, fmt.Errorf("Could not construct request: %w", err)
	}

	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

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
