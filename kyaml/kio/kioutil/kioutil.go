// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kioutil

import (
	"fmt"
	"sort"
	"strconv"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type AnnotationKey = string

const (
	// IndexAnnotation records the index of a specific resource in a file or input stream.
	IndexAnnotation AnnotationKey = "config.kubernetes.io/index"

	// PathAnnotation records the path to the file the Resource was read from
	PathAnnotation AnnotationKey = "config.kubernetes.io/path"

	// PackageAnnotation records the name of the package the Resource was read from
	PackageAnnotation AnnotationKey = "config.kubernetes.io/package"
)

func GetFileAnnotations(rn *yaml.RNode) (string, string, error) {
	meta, err := rn.GetMeta()
	if err != nil {
		return "", "", err
	}
	path := meta.Annotations[PathAnnotation]
	index := meta.Annotations[IndexAnnotation]
	return path, index, nil
}

// ErrorIfMissingAnnotation validates the provided annotations are present on the given resources
func ErrorIfMissingAnnotation(nodes []*yaml.RNode, keys ...AnnotationKey) error {
	for _, key := range keys {
		for _, node := range nodes {
			val, err := node.Pipe(yaml.GetAnnotation(key))
			if err != nil {
				return err
			}
			if val == nil {
				return fmt.Errorf("missing package annotation %s", key)
			}
		}
	}
	return nil
}

// Map invokes fn for each element in nodes.
func Map(nodes []*yaml.RNode, fn func(*yaml.RNode) (*yaml.RNode, error)) ([]*yaml.RNode, error) {
	var returnNodes []*yaml.RNode
	for i := range nodes {
		n, err := fn(nodes[i])
		if err != nil {
			return nil, err
		}
		if n != nil {
			returnNodes = append(returnNodes, n)
		}
	}
	return returnNodes, nil
}

// SortNodes sorts nodes in place:
// - by PathAnnotation annotation
// - by IndexAnnotation annotation
func SortNodes(nodes []*yaml.RNode) error {
	var err error
	// use stable sort to keep ordering of equal elements
	sort.SliceStable(nodes, func(i, j int) bool {
		if err != nil {
			return false
		}
		var iMeta, jMeta yaml.ResourceMeta
		if iMeta, _ = nodes[i].GetMeta(); err != nil {
			return false
		}
		if jMeta, _ = nodes[j].GetMeta(); err != nil {
			return false
		}

		iValue := iMeta.Annotations[PathAnnotation]
		jValue := jMeta.Annotations[PathAnnotation]
		if iValue != jValue {
			return iValue < jValue
		}

		iValue = iMeta.Annotations[IndexAnnotation]
		jValue = jMeta.Annotations[IndexAnnotation]

		// put resource config without an index first
		if iValue == jValue {
			return false
		}
		if iValue == "" {
			return true
		}
		if jValue == "" {
			return false
		}

		// sort by index
		var iIndex, jIndex int
		iIndex, err = strconv.Atoi(iValue)
		if err != nil {
			err = fmt.Errorf("unable to parse config.kubernetes.io/index %s :%v", iValue, err)
			return false
		}
		jIndex, err = strconv.Atoi(jValue)
		if err != nil {
			err = fmt.Errorf("unable to parse config.kubernetes.io/index %s :%v", jValue, err)
			return false
		}
		if iIndex != jIndex {
			return iValue < jValue
		}

		// elements are equal
		return false
	})
	return err
}
