// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"log"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// DefaultNamespace is the default namespace name in Kubernetes.
const DefaultNamespace = "default"

// GlobPatterns accepts a slice of glob strings and returns the set of
// matching file paths.
func GlobPatterns(fSys filesys.FileSystem, patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		files, err := fSys.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			log.Printf("%s has no match", pattern)
			continue
		}
		result = append(result, files...)
	}
	return result, nil
}

// GlobPatternsWithLoader accepts a slice of glob strings and returns the set of matching file paths.
// If validation is skipped, then it will return the patterns as provided.
// Otherwise, It will try to load the files from the filesystem.
// If files are not found in the filesystem, it will try to load from remote.
// It returns an error if validation is not skipped and there are no matching files or it can't load from remote.
func GlobPatternsWithLoader(fSys filesys.FileSystem, ldr ifc.Loader, patterns []string, skipValidation bool) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		if skipValidation {
			result = append(result, pattern)
			continue
		}

		files, err := fSys.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error checking the filesystem: %w", err)
		}

		if len(files) != 0 {
			result = append(result, files...)
			continue
		}

		loader, err := ldr.New(pattern)
		if err != nil {
			return nil, fmt.Errorf("%s has no match: %w", pattern, err)
		}

		result = append(result, pattern)
		if loader != nil {
			if err = loader.Cleanup(); err != nil {
				return nil, fmt.Errorf("error cleaning up loader: %w", err)
			}
		}
	}
	return result, nil
}

// ConvertToMap converts a string in the form of `key:value,key:value,...` into a map.
func ConvertToMap(input string, kind string) (map[string]string, error) {
	result := make(map[string]string)
	if input == "" {
		return result, nil
	}
	inputs := strings.Split(input, ",")
	return ConvertSliceToMap(inputs, kind)
}

// ConvertSliceToMap converts a slice of strings in the form of
// `key:value` into a map.
func ConvertSliceToMap(inputs []string, kind string) (map[string]string, error) {
	result := make(map[string]string)
	for _, input := range inputs {
		c := strings.Index(input, ":")
		switch {
		case c == 0:
			// key is not passed
			return nil, fmt.Errorf("invalid %s: '%s' (%s)", kind, input, "need k:v pair where v may be quoted")
		case c < 0:
			// only key passed
			result[input] = ""
		default:
			// both key and value passed
			key := input[:c]
			value := trimQuotes(input[c+1:])
			result[key] = value
		}
	}
	return result, nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// NamespaceEqual checks if two namespaces are the same. It considers the empty namespace and the default namespace to
// be the same. As such, when one namespace is the empty string ('""') and the other namespace is "default", this function
// will return true.
func NamespaceEqual(namespace string, otherNamespace string) bool {
	if "" == namespace {
		namespace = DefaultNamespace
	}

	if "" == otherNamespace {
		otherNamespace = DefaultNamespace
	}

	return namespace == otherNamespace
}
