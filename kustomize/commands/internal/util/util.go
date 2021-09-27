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

// GlobPatterns accepts a slice of glob strings and returns the set of
// matching file paths. If files are not found, will try load from remote.
func GlobPatternsWithLoader(fSys filesys.FileSystem, ldr ifc.Loader, patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		files, err := fSys.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			loader, err := ldr.New(pattern)
			if err != nil {
				log.Printf("%s has no match", pattern)
			} else {
				result = append(result, pattern)
				if loader != nil {
					loader.Cleanup()
				}
			}
			continue
		}
		result = append(result, files...)
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
		if c == 0 {
			// key is not passed
			return nil, fmt.Errorf("invalid %s: '%s' (%s)", kind, input, "need k:v pair where v may be quoted")
		} else if c < 0 {
			// only key passed
			result[input] = ""
		} else {
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
