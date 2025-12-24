package types

import (
	"context"
)

type GitForge interface {
	FetchRootGroupContent(ctx context.Context) (map[string]GroupSource, error)
	FetchGroupContent(ctx context.Context, source GroupSource) (GroupContent, error)
}

type GroupSource interface {
	GetGroupID() uint64
	GetGroupPath() string
}

type RepositorySource interface {
	GetRepositoryID() uint64
	GetRepositoryName() string
	GetCloneURL() string
	GetDefaultBranch() string
}

type GroupContent struct {
	Groups       map[string]GroupSource
	Repositories map[string]RepositorySource
}
