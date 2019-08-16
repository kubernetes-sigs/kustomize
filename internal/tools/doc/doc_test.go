package doc

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParseYAML(t *testing.T) {
	testCases := []struct {
		identifiers []string
		values      []string
		yaml        string
	}{
		{
			identifiers: []string{
				"namePrefix",
				"metadata",
				"metadata:name",
				"kind",
			},
			values: []string{
				"namePrefix=dev-",
				"metadata:name=app",
				"kind=Deployment",
			},
			yaml: `
namePrefix: dev-
metadata:
  name: app
kind: Deployment
`,
		},
		{
			identifiers: []string{
				"namePrefix",
				"metadata",
				"metadata:name",
				"metadata:spec",
				"metadata:spec:replicas",
				"kind",
				"replicas",
				"replicas:name",
				"replicas:count",
				"resource",
			},
			values: []string{
				"namePrefix=dev-",
				"metadata:name=n1",
				"metadata:spec:replicas=3",
				"kind=Deployment",
				"replicas:name=n1",
				"replicas:name=n2",
				"replicas:count=3",
				"resource=file1.yaml",
				"resource=file2.yaml",
			},
			yaml: `
namePrefix: dev-
# map of map
metadata:
  name: n1
  spec:
    replicas: 3
kind: Deployment

#list of map
replicas:
- name: n1
  count: 3
- name: n2
  count: 3

# list
resource:
- file1.yaml
- file2.yaml
`,
		},
	}

	for _, test := range testCases {
		doc := KustomizationDocument{
			DocumentData: test.yaml,
			FilePath:     "example/path/kustomization.yaml",
		}

		err := doc.ParseYAML()
		if err != nil {
			t.Errorf("Document error error: %s", err)
		}

		cmpStrings := func(got, expected []string, label string) {
			sort.Strings(got)
			sort.Strings(expected)

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Expected %s (%v) to be equal to (%v)\n",
					label,
					strings.Join(got, ","),
					strings.Join(expected, ","))
			}

		}

		cmpStrings(doc.Identifiers, test.identifiers, "identifiers")
		cmpStrings(doc.Values, test.values, "values")
	}
}
