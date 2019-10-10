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
					},
				},
				{
					RelativePath: "../file/../../something/../to/other/file/patch.yaml",
					Expected: Document{
						RepositoryURL: "example.com/repo",
						FilePath:      "path/to/other/file/patch.yaml",
						DefaultBranch: "master",
					},
				},
				{
					RelativePath: "service.yaml",
					Expected: Document{
						RepositoryURL: "example.com/repo",
						FilePath:      "path/to/file/service.yaml",
						DefaultBranch: "master",
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
