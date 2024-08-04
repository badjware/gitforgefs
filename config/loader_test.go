package config_test

import (
	"reflect"
	"testing"

	"github.com/badjware/gitlabfs/config"
)

func TestLoadConfig(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected *config.Config
	}{
		"LoadConfig": {
			input: "config.test.yaml",
			expected: &config.Config{
				FS: config.FSConfig{
					Mountpoint:   "/tmp/gitlabfs/test/mnt/gitlab",
					MountOptions: "nodev",
					Platform:     "gitlab",
				},
				Gitlab: config.GitlabClientConfig{
					URL:                     "https://example.com",
					Token:                   "12345",
					PullMethod:              "ssh",
					GroupIDs:                []int{123},
					UserIDs:                 []int{456},
					ArchivedProjectHandling: "hide",
					IncludeCurrentUser:      true,
				},
				Github: config.GithubClientConfig{
					Token:                "12345",
					PullMethod:           "http",
					OrgNames:             []string{"test-org"},
					UserNames:            []string{"test-user"},
					ArchivedRepoHandling: "hide",
					IncludeCurrentUser:   true,
				},
				Git: config.GitClientConfig{
					CloneLocation:    "/tmp/gitlabfs/test/cache/gitlab",
					Remote:           "origin",
					OnClone:          "clone",
					AutoPull:         false,
					Depth:            0,
					QueueSize:        100,
					QueueWorkerCount: 1,
				}},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := config.LoadConfig(test.input)
			expected := test.expected
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("LoadConfig(%v) returned %v; expected %v; error: %v", test.input, got, expected, err)
			}
		})
	}
}

func TestMakeGitConfig(t *testing.T) {
	tests := map[string]struct {
		input    *config.Config
		expected *config.GitClientConfig
	}{
		"ValidConfig": {
			input: &config.Config{
				FS: config.FSConfig{
					Platform: "gitlab",
				},
				Git: config.GitClientConfig{
					CloneLocation:    "/tmp",
					Remote:           "origin",
					OnClone:          "init",
					AutoPull:         false,
					Depth:            0,
					QueueSize:        200,
					QueueWorkerCount: 5,
				},
			},
			expected: &config.GitClientConfig{
				CloneLocation:    "/tmp",
				Remote:           "origin",
				OnClone:          "init",
				AutoPull:         false,
				Depth:            0,
				QueueSize:        200,
				QueueWorkerCount: 5,
			},
		},
		"InvalidOnClone": {
			input: &config.Config{
				FS: config.FSConfig{
					Platform: "gitlab",
				},
				Git: config.GitClientConfig{
					CloneLocation:    "/tmp",
					Remote:           "origin",
					OnClone:          "invalid",
					AutoPull:         false,
					Depth:            0,
					QueueSize:        200,
					QueueWorkerCount: 5,
				},
			},
			expected: nil,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := config.MakeGitConfig(test.input)
			expected := test.expected
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("MakeGitConfig(%v) returned %v; expected %v; error %v", test.input, got, expected, err)
			}
		})
	}
}

func TestMakeGitlabConfig(t *testing.T) {
	tests := map[string]struct {
		input    *config.Config
		expected *config.GitlabClientConfig
	}{
		"ValidConfig": {
			input: &config.Config{
				FS: config.FSConfig{
					Platform: "gitlab",
				},
				Gitlab: config.GitlabClientConfig{
					URL:                     "https://gitlab.com",
					PullMethod:              "http",
					Token:                   "",
					GroupIDs:                []int{9970},
					UserIDs:                 []int{},
					ArchivedProjectHandling: "hide",
					IncludeCurrentUser:      true,
				},
			},
			expected: &config.GitlabClientConfig{
				URL:                     "https://gitlab.com",
				PullMethod:              "http",
				Token:                   "",
				GroupIDs:                []int{9970},
				UserIDs:                 []int{},
				ArchivedProjectHandling: "hide",
				IncludeCurrentUser:      true,
			},
		},
		"InvalidPullMethod": {
			input: &config.Config{
				FS: config.FSConfig{
					Platform: "gitlab",
				},
				Gitlab: config.GitlabClientConfig{
					URL:                     "https://gitlab.com",
					PullMethod:              "invalid",
					Token:                   "",
					GroupIDs:                []int{9970},
					UserIDs:                 []int{},
					ArchivedProjectHandling: "hide",
					IncludeCurrentUser:      true,
				},
			},
			expected: nil,
		},
		"InvalidArchiveHandling": {
			input: &config.Config{
				Gitlab: config.GitlabClientConfig{
					URL:                     "https://gitlab.com",
					PullMethod:              "http",
					Token:                   "",
					GroupIDs:                []int{9970},
					UserIDs:                 []int{},
					IncludeCurrentUser:      true,
					ArchivedProjectHandling: "invalid",
				},
			},
			expected: nil,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := config.MakeGitlabConfig(test.input)
			expected := test.expected
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("MakeGitlabConfig(%v) returned %v; expected %v; error: %v", test.input, got, expected, err)
			}
		})
	}
}
