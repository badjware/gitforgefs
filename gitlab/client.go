package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

const (
	PullMethodHTTP = "http"
	PullMethodSSH  = "ssh"
)

type GitlabFetcher interface {
	GroupFetcher
	UserFetcher
}

type GitlabClientParam struct {
	PullMethod string
}

type gitlabClient struct {
	GitlabClientParam
	client *gitlab.Client
}

func NewClient(gitlabUrl string, gitlabToken string, p GitlabClientParam) (*gitlabClient, error) {
	client, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %v", err)
	}

	gitlabClient := &gitlabClient{
		GitlabClientParam: p,
		client:            client,
	}
	return gitlabClient, nil
}
