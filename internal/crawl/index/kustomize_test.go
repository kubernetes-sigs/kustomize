package index

import (
	"reflect"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	testCases := []struct {
		query  string
		result map[string]interface{}
	}{
		{
			query:  "    \t\n\r",
			result: map[string]interface{}{"size": 0},
		},
		{
			query: "\tidentifier1 identifier2\nidentifier3\r",
			result: map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []map[string]interface{}{
							multiMatch("identifier1"),
							multiMatch("identifier2"),
							multiMatch("identifier3"),
						},
					},
				},
			},
		},
		{
			query: "kind=Kustomization",
			result: map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []map[string]interface{}{
							{
								"term": map[string]interface{}{
									"kinds.keyword": "Kustomization",
								},
							},
						},
					},
				},
			},
		},
		{
			query: "kind=Kustomization identifier2",
			result: map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []map[string]interface{}{
							{
								"term": map[string]interface{}{
									"kinds.keyword": "Kustomization",
								},
							},
							multiMatch("identifier2"),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		result := BuildQuery(tc.query)
		if !reflect.DeepEqual(tc.result, result) {
			t.Errorf("Expected %#v to match %#v", result, tc.result)
		}
	}
}
