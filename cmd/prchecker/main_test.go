package main

import (
	"testing"

	"github.com/google/go-github/github"
)

func TestStringAllowsModuleSpan(t *testing.T) {
	var tests = []struct {
		name          string
		body          string
		matchExpected bool
	}{
		{
			"exception not included",
			"foo",
			false,
		},
		{
			"exception mentioned in sentence does not exempt span check",
			"don't ALLOW_MODULE_SPAN",
			false,
		},
		{
			"PR body is just exception",
			"ALLOW_MODULE_SPAN",
			true,
		},
		{
			"support markdown quoting exception",
			"> ALLOW_MODULE_SPAN",
			true,
		},
		{
			"support whitespace padding",
			"\t ALLOW_MODULE_SPAN\t ",
			true,
		},
		{
			"module span exemption allowed at start of string",
			"ALLOW_MODULE_SPAN\nat start of file",
			true,
		},
		{
			"module span exemption allowed at end of string",
			"at end of file\nALLOW_MODULE_SPAN",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringAllowsModuleSpan(tt.body)

			if err != nil {
				t.Errorf(err.Error())
			}

			if result != tt.matchExpected {
				t.Errorf("got %t, want %t", result, tt.matchExpected)
			}
		})
	}
}

func TestIsModuleSpanAllowed(t *testing.T) {
	var tests = []struct {
		name          string
		body          string
		matchExpected bool
	}{
		{
			"module spanning not allowed",
			"don't ALLOW_MODULE_SPAN",
			false,
		},
		{
			"module spanning allowed",
			"ALLOW_MODULE_SPAN",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := FakeGitHubService{
				pullRequest: &github.PullRequest{
					Body: &tt.body,
				},
			}

			result, _ := ModuleSpanAllowed(&client, 1)

			if result != tt.matchExpected {
				t.Errorf("got %t, want %t", result, tt.matchExpected)
			}
		})
	}
}

func TestGetCommitChanges(t *testing.T) {
	client := FakeGitHubService{
		commit: &github.RepositoryCommit{
			SHA:   github.String("abc123"),
			Files: []github.CommitFile{{Filename: github.String("foo")}, {Filename: github.String("bar")}},
		},
	}

	result, _ := GetCommitChanges(&client, "abc123")

	if result.id != "abc123" {
		t.Errorf("got %v, want %v", result.id, "abc123")
	}

	if len(result.files) != 2 {
		t.Errorf("got %v, want %v", len(result.files), "2")
	}

	expectedFiles := []string{"foo", "bar"}
	if !StringSliceAreEqual(result.files, expectedFiles) {
		t.Errorf("got %v, want %v", result.files, expectedFiles)
	}
}

func TestGetPullRequestCommits(t *testing.T) {
	client := FakeGitHubService{
		commit: &github.RepositoryCommit{
			SHA:   github.String("abc123"),
			Files: []github.CommitFile{{Filename: github.String("foo")}, {Filename: github.String("bar")}},
		},
		commitList: []*github.RepositoryCommit{
			{
				SHA: github.String("abc123"),
			},
			{
				SHA: github.String("abc123"),
			},
		},
	}

	result, _ := GetPullRequestCommits(&client, 42)

	if len(result) != 2 {
		t.Errorf("got %v, want %v", len(result), 2)
	}

	expectedFiles := []string{"foo", "bar"}
	if !StringSliceAreEqual(result[0].files, expectedFiles) {
		t.Errorf("[%d] got %v, want %v", 0, result[0].files, expectedFiles)
	}
	if !StringSliceAreEqual(result[1].files, expectedFiles) {
		t.Errorf("[%d] got %v, want %v", 1, result[1].files, expectedFiles)
	}
}

func TestContstructingChangeset(t *testing.T) {
	var tests = []struct {
		name     string
		sha      string
		files    []github.CommitFile
		expected Changeset
	}{
		{
			"construct from single file",
			"abc123",
			[]github.CommitFile{{Filename: github.String("foo")}},
			Changeset{id: "abc123", files: []string{"foo"}},
		},
		{
			"construct from multiple files",
			"1234",
			[]github.CommitFile{{Filename: github.String("foo")}, {Filename: github.String("bar")}},
			Changeset{id: "1234", files: []string{"foo", "bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := &github.RepositoryCommit{SHA: github.String(tt.sha), Files: tt.files}
			result := ConstructChangeset(commit)

			if tt.expected.id != result.id {
				t.Errorf("got %v, want %v", result.id, tt.expected.id)
			}

			if !StringSliceAreEqual(tt.expected.files, result.files) {
				t.Errorf("got %v, want %v", result.files, tt.expected.files)
			}
		})
	}
}

func TestIsChangesetSpanning(t *testing.T) {
	var tests = []struct {
		name      string
		changeset []string
		files     []string
		expected  bool
	}{
		{
			"matching sets of 1 element do not span",
			[]string{"1"},
			[]string{"1"},
			false,
		},
		{
			"subdirectories do not match top level directories",
			[]string{"a/1"},
			[]string{"1"},
			false,
		},
		{
			"distinct sets do not span",
			[]string{"1", "2"},
			[]string{"a", "b"},
			false,
		},
		{
			"single matching path does not span",
			[]string{"1", "a"},
			[]string{"1", "2"},
			false,
		},
		{
			"path and subdirectory of same restriction do not span",
			[]string{"1", "1/a"},
			[]string{"1", "2", "3"},
			false,
		},
		{
			"matching sets span",
			[]string{"1", "2"},
			[]string{"1", "2"},
			true,
		},
		{
			"superset of restricted paths spans",
			[]string{"1", "2", "3"},
			[]string{"1", "2"},
			true,
		},
		{
			"subset of restricted paths spans",
			[]string{"1", "3"},
			[]string{"1", "2", "3"},
			true,
		},
		{
			"subdirectories of restricted paths span",
			[]string{"1/a", "3/b"},
			[]string{"1", "2", "3"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			changeset := &Changeset{files: tt.changeset}

			result := changeset.isSpanningPaths(tt.files)

			if result != tt.expected {
				t.Errorf("got %t, want %t", result, tt.expected)
			}
		})
	}
}

type FakeGitHubService struct {
	pullRequest *github.PullRequest
	commit      *github.RepositoryCommit
	commitList  []*github.RepositoryCommit
}

func (repository FakeGitHubService) GetPullRequest(prId int) (*github.PullRequest, *github.Response, error) {
	return repository.pullRequest, nil, nil
}

func (repository FakeGitHubService) GetCommit(commitSha string) (*github.RepositoryCommit, *github.Response, error) {
	return repository.commit, nil, nil

}

func (repository FakeGitHubService) ListCommits(prId int, options *github.ListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	return repository.commitList, &github.Response{NextPage: 0}, nil
}

func StringSliceAreEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i, elem := range left {
		if elem != right[i] {
			return false
		}
	}

	return true
}
