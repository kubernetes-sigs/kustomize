package doc

import (
	"reflect"
	"testing"
)

func TestFromRelativePath(t *testing.T) {
	type Case struct {
		RelativePath string
		Expected     Document
	}

	testCases := []struct {
		BaseDoc Document
		Cases   []Case
	}{
		{
			BaseDoc: Document{
				RepositoryURL: "example.com/repo",
				FilePath:      "path/to/file/kustomization.yaml",
				DefaultBranch: "master",
			},
			Cases: []Case{
				{
					RelativePath: "../other/file/resource.yaml",
					Expected: Document{
						RepositoryURL: "example.com/repo",
						FilePath:      "path/to/other/file/resource.yaml",
						DefaultBranch: "master",
						User:          "example.com",
					},
				},
				{
					RelativePath: "../file/../../something/../to/other/file/patch.yaml",
					Expected: Document{
						RepositoryURL: "example.com/repo",
						FilePath:      "path/to/other/file/patch.yaml",
						DefaultBranch: "master",
						User:          "example.com",
					},
				},
				{
					RelativePath: "service.yaml",
					Expected: Document{
						RepositoryURL: "example.com/repo",
						FilePath:      "path/to/file/service.yaml",
						DefaultBranch: "master",
						User:          "example.com",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		for _, c := range tc.Cases {
			rd, err := tc.BaseDoc.FromRelativePath(c.RelativePath)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(rd, c.Expected) {
				t.Errorf("document mismatch expected %v, got %v", c.Expected, rd)
			}
		}
	}
}

func TestDocument_RepositoryFullName(t *testing.T) {
	testCases := []struct {
		doc                        Document
		expectedRepositoryFullName string
	}{
		{
			doc: Document{
				RepositoryURL: "https://github.com/user/repo",
			},
			expectedRepositoryFullName: "user/repo",
		},
		{
			doc: Document{
				RepositoryURL: "https://github.com//user/repo////",
			},
			expectedRepositoryFullName: "user/repo",
		},
		{
			doc: Document{
				RepositoryURL: "repo/",
			},
			expectedRepositoryFullName: "repo",
		},
		{
			doc: Document{
				RepositoryURL: "",
			},
			expectedRepositoryFullName: "",
		},
		{
			doc: Document{
				RepositoryURL: "git@github.com:user/repo",
			},
			expectedRepositoryFullName: "user/repo",
		},
	}

	for _, tc := range testCases {
		returnedRepositoryFullName := tc.doc.RepositoryFullName()
		if returnedRepositoryFullName != tc.expectedRepositoryFullName {
			t.Errorf("RepositoryFullName expected %s, got %s",
				tc.expectedRepositoryFullName,
				returnedRepositoryFullName)
		}
	}
}

func TestDocument_UserName(t *testing.T) {
	testCases := []struct {
		repositoryURL    string
		expectedUserName string
	}{
		{
			repositoryURL:    "https://github.com/user/repo",
			expectedUserName: "user",
		},
		{
			repositoryURL:    "https://github.com//user/repo////",
			expectedUserName: "user",
		},
		{
			repositoryURL:    "repo/",
			expectedUserName: "repo",
		},
		{
			repositoryURL:    "",
			expectedUserName: "",
		},
		{
			repositoryURL:    "git@github.com:user/repo",
			expectedUserName: "user",
		},
	}

	for _, tc := range testCases {
		returnedUserName := UserName(tc.repositoryURL)
		if returnedUserName != tc.expectedUserName {
			t.Errorf("UserName expected %s, got %s",
				tc.expectedUserName, returnedUserName)
		}
	}
}
