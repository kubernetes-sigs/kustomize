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
		kinds       []string
		filepath    string
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
				"kind=",
				"namePrefix=dev-",
				"metadata:name=app",
			},
			kinds: []string{
				"Kustomization",
			},
			filepath: "some/path/to/kustomization.yaml",
			yaml: `
namePrefix: dev-
metadata:
  name: app
kind: ""
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
				"kind=Kustomization",
				"replicas:name=n1",
				"replicas:name=n2",
				"replicas:count=3",
				"resource=file1.yaml",
				"resource=file2.yaml",
			},
			kinds: []string{
				"Kustomization",
			},
			filepath: "./kustomization.yaml",
			yaml: `
namePrefix: dev-
# map of map
metadata:
  name: n1
  spec:
    replicas: 3
kind: Kustomization

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
		{
			identifiers: []string{
				"kind",
				"metadata",
				"metadata:name",
			},
			values: []string{
				"kind=Deployment",
				"kind=Service",
				"kind=Custom",
				"metadata:name=app",
				"metadata:name=app-service",
				"metadata:name=app-crd",
			},
			kinds: []string{
				"Deployment",
				"Service",
				"Custom",
			},
			filepath: "resources.yaml",
			yaml: `
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app-service
---
kind: Custom
metadata:
  name: app-crd
`,
		},
	}

	for _, test := range testCases {
		doc := KustomizationDocument{
			Document: Document{
				DocumentData: test.yaml,
				FilePath:     test.filepath,
			},
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
		cmpStrings(doc.Kinds, test.kinds, "kinds")
	}
}

func TestGetResources(t *testing.T) {
	tests := []struct {
		doc       KustomizationDocument
		resources []*Document
	}{
		{
			doc: KustomizationDocument{
				Document: Document{
					RepositoryURL: "sigs.k8s.io/kustomize",
					FilePath:      "some/path/to/kdir/kustomization.yaml",
					DocumentData: `
bases:
- ../base
- ../otherbase

resources:
- file.yaml
- https://github.com/kubernetes-sigs/kustomize/examples/helloWorld?ref=v3.1.0
`},
			},
			resources: []*Document{
				{
					RepositoryURL: "sigs.k8s.io/kustomize",
					FilePath:      "some/path/to/base",
				},
				{
					RepositoryURL: "sigs.k8s.io/kustomize",
					FilePath:      "some/path/to/otherbase",
				},
				{
					RepositoryURL: "sigs.k8s.io/kustomize",
					FilePath:      "some/path/to/kdir/file.yaml",
				},
				{
					RepositoryURL: "https://github.com/kubernetes-sigs/kustomize",
					FilePath:      "examples/helloWorld",
					DefaultBranch: "v3.1.0",
				},
			},
		},
		{
			doc: KustomizationDocument{
				Document: Document{
					RepositoryURL: "https://github.com/some/repo",
					FilePath:      "some/resource.yaml",
					DocumentData: `
bases:
- ../base
- ../overlay

resources:
- https://github.com/kubernetes-sigs/kustomize/examples/helloWorld?ref=v3.1.0
- some/file.yaml
`,
				},
			},
			resources: []*Document{},
		},
	}

	for _, test := range tests {
		res, err := test.doc.GetResources()
		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
			continue
		}
		if len(test.resources) != len(res) {
			t.Errorf("Number of resources does not match.")
			continue
		}
		cmp := func(docs []*Document) func(i, j int) bool {
			return func(i, j int) bool {
				if docs[i].RepositoryURL != docs[j].RepositoryURL {
					return docs[i].RepositoryURL <
						docs[j].RepositoryURL
				}

				if docs[i].FilePath != docs[j].FilePath {
					return docs[i].FilePath <
						docs[j].FilePath
				}

				return docs[i].DefaultBranch < docs[j].DefaultBranch
			}
		}
		sort.Slice(test.resources, cmp(test.resources))
		sort.Slice(res, cmp(res))
		for i, r := range test.resources {
			if !reflect.DeepEqual(res[i], r) {
				t.Errorf("Expected '%+v' to equal '%+v'\n",
					res[i], r)
			}
		}
	}
}
