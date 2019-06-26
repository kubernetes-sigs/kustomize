package doc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestFieldLoadSaver(t *testing.T) {

	commonTestCases := []KustomizationDocument{
		{
			identifiers:   []Atom{"namePrefix", "metadata.name", "kind"},
			FilePath:      "some/path/kustomization.yaml",
			RepositoryURL: "https://example.com/kustomize",
			CreationTime:  time.Now(),
			DocumentData: `
namePrefix: dev-
metadata:
  name: app
kind: Deployment
`,
		},
	}

	for _, test := range commonTestCases {
		fields, metadata, err := test.Save()
		if err != nil {
			t.Errorf("Error calling Save(): %s\n", err)
		}
		doc := KustomizationDocument{}
		err = doc.Load(fields, metadata)
		if err != nil {
			t.Errorf("Doc failed to load: %s\n", err)
		}
		if !reflect.DeepEqual(test, doc) {
			t.Errorf("Expected loaded document (%+v) to be equal to (%+v)\n", doc, test)
		}
	}
}

func TestParseYAML(t *testing.T) {
	testCases := []struct {
		identifiers []Atom
		yaml        string
	}{
		{
			identifiers: []Atom{
				"namePrefix",
				"metadata",
				"metadata name",
				"kind",
			},
			yaml: `
namePrefix: dev-
metadata:
  name: app
kind: Deployment
`,
		},
		{
			identifiers: []Atom{
				"namePrefix",
				"metadata",
				"metadata name",
				"metadata spec",
				"metadata spec replicas",
				"kind",
				"replicas",
				"replicas name",
				"replicas count",
				"resource",
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

	atomStrs := func(atoms []Atom) []string {
		strs := make([]string, 0, len(atoms))
		for _, val := range atoms {
			strs = append(strs, fmt.Sprintf("%v", val))
		}
		return strs
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

		docIDs := atomStrs(doc.identifiers)
		expectedIDs := atomStrs(test.identifiers)
		sort.Strings(docIDs)
		sort.Strings(expectedIDs)

		if !reflect.DeepEqual(docIDs, expectedIDs) {
			t.Errorf("Expected loaded document (%v) to be equal to (%v)\n",
				strings.Join(docIDs, ","), strings.Join(expectedIDs, ","))
		}
	}
}
