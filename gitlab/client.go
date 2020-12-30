package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GitlabFetcher interface {
	GroupFetcher
	UserFetcher
}

type GitlabClientParam struct {
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
		client: client,
	}
	return gitlabClient, nil
}
