package doc

import (
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

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
	DocumentData  string    `json:"document,omitempty"`
	Kinds         []string  `json:"kinds,omitempty"`
	Identifiers   []string  `json:"identifiers,omitempty"`
	Values        []string  `json:"values,omitempty"`
	FilePath      string    `json:"filePath,omitempty"`
	RepositoryURL string    `json:"repositoryUrl,omitempty"`
	CreationTime  time.Time `json:"creationTime,omitempty"`
}

func (doc *KustomizationDocument) ParseYAML() error {
	doc.Identifiers = make([]string, 0)
	doc.Values = make([]string, 0)

	var kustomization map[string]interface{}
	err := yaml.Unmarshal([]byte(doc.DocumentData), &kustomization)
	if err != nil {
		return fmt.Errorf("unable to parse kustomization file: %s", err)
	}

	type Map struct {
		data   map[string]interface{}
		prefix string
	}

	toVisit := []Map{
		{
			data:   kustomization,
			prefix: "",
		},
	}

	identifierSet := make(map[string]struct{})
	valueSet := make(map[string]struct{})
	for i := 0; i < len(toVisit); i++ {
		visiting := toVisit[i]
		for k, v := range visiting.data {
			identifier := fmt.Sprintf("%s:%s", visiting.prefix,
				strings.Replace(k, ":", "%3A", -1))
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
					esc := strings.Replace(fmt.Sprintf("%v",
						value), ":", "%3A", -1)

					valuePath := fmt.Sprintf("%s=%v",
						identifier, esc)
					valueSet[valuePath] = struct{}{}
				}
			}
			traverseStructure(v)

			identifierSet[identifier] = struct{}{}

		}
	}

	for val := range valueSet {
		doc.Values = append(doc.Values, val)
	}

	for key := range identifierSet {
		doc.Identifiers = append(doc.Identifiers, key)
	}

	return nil
}
