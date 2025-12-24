package cache_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/badjware/gitforgefs/cache"
	"github.com/badjware/gitforgefs/types"
)

type mockRepoGroupSource struct {
	ID   uint64
	Name string
	Path string
}

func (m mockRepoGroupSource) GetGroupID() uint64   { return m.ID }
func (m mockRepoGroupSource) GetGroupName() string { return m.Name }
func (m mockRepoGroupSource) GetGroupPath() string { return m.Path }

type mockBackend struct {
	RootCalls  int
	GroupCalls int

	rootContentValue  map[string]types.RepositoryGroupSource
	groupContentValue types.RepositoryGroupContent
}

func (m *mockBackend) FetchRootGroupContent(ctx context.Context) (map[string]types.RepositoryGroupSource, error) {
	m.RootCalls++
	return m.rootContentValue, nil
}

func (m *mockBackend) FetchGroupContent(ctx context.Context, source types.RepositoryGroupSource) (types.RepositoryGroupContent, error) {
	m.GroupCalls++
	return m.groupContentValue, nil
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestFetchRootGroupContent(t *testing.T) {
	backend := &mockBackend{
		rootContentValue: map[string]types.RepositoryGroupSource{
			"g": mockRepoGroupSource{ID: 1, Name: "g", Path: "g"},
		},
	}

	logger := newLogger()
	c := cache.NewForgeCache(backend, logger)

	ctx := context.Background()

	result1, err := c.FetchRootGroupContent(ctx)
	if err != nil {
		t.Fatalf("first FetchRootGroupContent failed: %v", err)
	}
	if content, ok := result1["g"]; !ok || content.GetGroupID() != 1 {
		t.Fatalf("unexpected root content fetched: %v", result1)
	}

	result2, err := c.FetchRootGroupContent(ctx)
	if err != nil {
		t.Fatalf("second FetchRootGroupContent failed: %v", err)
	}
	if content, ok := result2["g"]; !ok || content.GetGroupID() != 1 {
		t.Fatalf("unexpected root content fetched: %v", result2)
	}

	if backend.RootCalls != 1 {
		t.Fatalf("expected backend.FetchRootGroupContent to be called once, got %d", backend.RootCalls)
	}
}

func TestFetchGroupContent(t *testing.T) {
	backend := &mockBackend{
		groupContentValue: types.RepositoryGroupContent{
			Groups:       map[string]types.RepositoryGroupSource{"group2": mockRepoGroupSource{ID: 2, Name: "group2", Path: "path/group1/group2"}},
			Repositories: map[string]types.RepositorySource{},
		},
	}

	logger := newLogger()
	c := cache.NewForgeCache(backend, logger)

	src := mockRepoGroupSource{ID: 1, Name: "group1", Path: "path/group1"}

	ctx := context.Background()

	result1, err := c.FetchGroupContent(ctx, src)
	if err != nil {
		t.Fatalf("first FetchGroupContent failed: %v", err)
	}
	if content, ok := result1.Groups["group2"]; !ok || content.GetGroupID() != 2 {
		t.Fatalf("unexpected group content fetched: %v", result1)
	}

	result2, err := c.FetchGroupContent(ctx, src)
	if err != nil {
		t.Fatalf("second FetchGroupContent failed: %v", err)
	}
	if content, ok := result2.Groups["group2"]; !ok || content.GetGroupID() != 2 {
		t.Fatalf("unexpected group content fetched: %v", result2)
	}

	if backend.GroupCalls != 1 {
		t.Fatalf("expected backend.FetchGroupContent to be called once, got %d", backend.GroupCalls)
	}
}

func TestInvalidateCache(t *testing.T) {
	backend := &mockBackend{
		groupContentValue: types.RepositoryGroupContent{
			Groups:       map[string]types.RepositoryGroupSource{"group2": mockRepoGroupSource{ID: 2, Name: "group2", Path: "path/group1/group2"}},
			Repositories: map[string]types.RepositorySource{},
		},
	}

	logger := newLogger()
	c := cache.NewForgeCache(backend, logger)

	src := mockRepoGroupSource{ID: 1, Name: "group1", Path: "path/group1"}

	ctx := context.Background()

	if _, err := c.FetchGroupContent(ctx, src); err != nil {
		t.Fatalf("first FetchGroupContent failed: %v", err)
	}

	c.InvalidateCache(src.GetGroupPath())

	if _, err := c.FetchGroupContent(ctx, src); err != nil {
		t.Fatalf("second FetchGroupContent failed: %v", err)
	}

	if backend.GroupCalls != 2 {
		t.Fatalf("expected backend.FetchGroupContent to be called twice, got %d", backend.GroupCalls)
	}
}
