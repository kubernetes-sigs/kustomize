package doc

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

var fileReader = kunstruct.NewKunstructuredFactoryImpl()

// This document is meant to be used at the elasticsearch document type.
// Fields are serialized as-is to elasticsearch, where indices are built
// to facilitate text search queries. Identifiers, Values, FilePath,
// RepositoryURL and DocumentData are meant to be searched for text queries
// directly, while the other fields can either be used as a filter, or as
// additional metadata displayed in the UI.
//
// The fields of the document and their purpose are listed below:
// - DocumentData contains the contents of the kustomization file.
// - Kinds Represents the kubernetes Kinds that are in this file.
// - Identifiers are a list of (partial and full) identifier paths that can be
//   found by users. Each part of a path is delimited by ":" e.g. spec:replicas.
// - Values are a list of identifier paths and their values that can be found by
//   search queries. The path is delimited by ":" and the value follows the "="
//   symbol e.g. spec:replicas=4.
// - FilePath is the path of the file.
// - RepositoryURL is the URL of the source repository.
// - CreationTime is the time at which the file was created.
//
// Representing each Identifier and Value as a flat string representation
// facilitates the use of complex text search features from elasticsearch such
// as fuzzy searching, regex, wildcards, etc.
type KustomizationDocument struct {
	Document
	Kinds       []string `json:"kinds,omitempty"`
	Identifiers []string `json:"identifiers,omitempty"`
	Values      []string `json:"values,omitempty"`
}

type set map[string]struct{}

// Implements the CrawlerDocument interface.
func (doc *KustomizationDocument) GetResources() ([]*Document, error) {
	isResource := true
	for _, suffix := range konfig.RecognizedKustomizationFileNames() {
		if strings.HasSuffix(doc.FilePath, "/"+suffix) {
			isResource = false
		}
	}
	if isResource {
		return []*Document{}, nil
	}

	content := []byte(doc.DocumentData)
	content, err := FixKustomizationPreUnmarshallingNonFatal(content)
	if err != nil {
		return nil, fmt.Errorf("could not fix kustomize file: %v", err)
	}

	var k types.Kustomization
	err = yaml.Unmarshal(content, &k)
	if err != nil {
		return nil, fmt.Errorf(
			"could not parse kustomization: %v", err)
	}
	k.FixKustomizationPostUnmarshalling()

	res := make([]*Document, 0, len(k.Resources))
	for _, r := range k.Resources {
		next, err := doc.Document.FromRelativePath(r)
		if err != nil {
			fmt.Printf("GetResources error: %v\n", err)
			continue
		}
		res = append(res, &next)
	}

	return res, nil
}

func (doc *KustomizationDocument) readBytes() ([]map[string]interface{}, error) {
	data := []byte(doc.DocumentData)

	for _, suffix := range konfig.RecognizedKustomizationFileNames() {
		if !strings.HasSuffix(doc.FilePath, "/"+suffix) {
			continue
		}
		var config map[string]interface{}
		err := yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to parse kustomization: %v", err)
		}
		return []map[string]interface{}{config}, nil
	}

	configs := make([]map[string]interface{}, 0)
	ks, err := fileReader.SliceFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse resource: %v", err)
	}
	for _, k := range ks {
		configs = append(configs, k.Map())
	}

	return configs, nil
}

func (doc *KustomizationDocument) ParseYAML() error {
	doc.Identifiers = make([]string, 0)
	doc.Values = make([]string, 0)
	doc.Kinds = make([]string, 0, 1)

	identifierSet := make(set)
	valueSet := make(set)
	getKind := func(m map[string]interface{}) string {
		const defaultStr = "Kustomization"
		kind, ok := m["kind"]
		if !ok {
			return defaultStr
		}
		if str, ok := kind.(string); ok && str != "" {
			return str
		}
		return defaultStr
	}

	ks, err := doc.readBytes()
	if err != nil {
		return err
	}

	for _, contents := range ks {
		doc.Kinds = append(doc.Kinds, getKind(contents))
		createFlatStructure(identifierSet, valueSet, contents)
	}

	for val := range valueSet {
		doc.Values = append(doc.Values, val)
	}

	for key := range identifierSet {
		doc.Identifiers = append(doc.Identifiers, key)
	}

	return nil
}

func createFlatStructure(identifierSet set, valueSet set, contents map[string]interface{}) {
	type Map struct {
		data   map[string]interface{}
		prefix string
	}

	toVisit := []Map{
		{
			data:   contents,
			prefix: "",
		},
	}

	for i := 0; i < len(toVisit); i++ {
		visiting := toVisit[i]
		for k, v := range visiting.data {
			identifier := fmt.Sprintf("%s:%s", visiting.prefix, k)
			// noop after the first iteration.
			identifier = strings.TrimLeft(identifier, ":")

			// Recursive function traverses structure to find
			// identifiers and values. These later get formatted
			// into doc.Identifiers and doc.Values respectively.
			var traverseStructure func(interface{})
			traverseStructure = func(arg interface{}) {
				switch value := arg.(type) {
				case map[string]interface{}:
					toVisit = append(toVisit, Map{
						data:   value,
						prefix: identifier,
					})
				case []interface{}:
					for _, val := range value {
						traverseStructure(val)
					}
				case interface{}:
					esc := fmt.Sprintf("%v", value)

					valuePath := fmt.Sprintf("%s=%v",
						identifier, esc)
					valueSet[valuePath] = struct{}{}
				}
			}
			traverseStructure(v)

			identifierSet[identifier] = struct{}{}
		}
	}
}
