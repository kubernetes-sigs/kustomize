package remove

import (
	"testing"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/edit/remove_test"
)

func TestRemoveTransformer(t *testing.T) {
	testCases := []remove_test.Case{
		{
			Description: "remove transformers",
			Given: remove_test.Given{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"transformer1.yaml"},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"transformer2.yaml",
					"transformer3.yaml",
				},
				Deleted: []string{
					"transformer1.yaml",
				},
			},
		},
		{
			Description: "remove transformer with pattern",
			Given: remove_test.Given{
				Items: []string{
					"foo/transformer1.yaml",
					"foo/transformer2.yaml",
					"foo/transformer3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{"foo/transformer*.yaml"},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/transformer1.yaml",
					"foo/transformer2.yaml",
					"foo/transformer3.yaml",
				},
			},
		},
		{
			Description: "nothing found to remove",
			Given: remove_test.Given{
				Items: []string{
					"transformer1.yaml",
					"transformer2.yaml",
					"transformer3.yaml",
				},
				RemoveArgs: []string{"foo"},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"transformer2.yaml",
					"transformer3.yaml",
					"transformer1.yaml",
				},
			},
		},
		{
			Description: "no arguments",
			Given:       remove_test.Given{},
			Expected: remove_test.Expected{
				Err: errors.New("must specify a transformer file"),
			},
		},
		{
			Description: "remove with multiple pattern arguments",
			Given: remove_test.Given{
				Items: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"transformer3.yaml",
					"do/not/deleteme/please.yaml",
				},
				RemoveArgs: []string{
					"foo/*.*",
					"bar/*.*",
					"tra*.yaml",
				},
			},
			Expected: remove_test.Expected{
				Items: []string{
					"do/not/deleteme/please.yaml",
				},
				Deleted: []string{
					"foo/foo.yaml",
					"bar/bar.yaml",
					"transformer3.yaml",
				},
			},
		},
	}

	remove_test.ExecuteTestCases(t, testCases, "transformers", newCmdRemoveTransformer)
}
