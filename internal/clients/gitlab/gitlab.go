package gitlab

import (
	"net/http"
)

type GitlabClient struct {
	host      string
	basePath  string
	userToken string
	userId    int
	client    http.Client
}

func New(host string, basePath string) *GitlabClient {
	return &GitlabClient{
		host:     host,
		basePath: basePath,
		client:   http.Client{},
	}
}

func (gc *GitlabClient) Events(req EventsReq) error {
	return nil
}
