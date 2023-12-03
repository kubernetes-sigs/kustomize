// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func DirExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func SliceToSet(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, x := range slice {
		if _, ok := result[x]; ok {
			log.Fatalf("programmer error - repeated value: %s", x)
		} else {
			result[x] = true
		}
	}
	return result
}

func ExtractModule(m string) string {
	k := strings.Index(m, " => ")
	if k < 0 {
		return m
	}
	return m[:k]
}

func SliceContains(slice []string, target string) bool {
	for _, x := range slice {
		if x == target {
			return true
		}
	}
	return false
}

// Receive git remote url as input and produce string containing git repository name
// e.g. kustomize, myrepo/kustomize
func ParseGitRepositoryPath(urlString string) string {
	var d []string = getURLStringArray(urlString)
	protocol := d[0]

	var repoPath string

	// TODO(antoooks): Confirm if we should handle other formats not commonly supported by Github
	switch protocol {
	// ssh protocol, e.g. git@github.com:path/repo.git
	case "git":
		repoPath = strings.Join(d[2:len(d)-1], "/") + "/" + d[3][:len(d[3])-4]
	// https protocol, e.g. https://github.com/path/repo.git
	case "https":
		repoPath = strings.Join(d[2:], "/")
		repoPath = repoPath[:len(repoPath)-4]
	// unsupported format
	default:
		_ = fmt.Errorf("protocol format is not supported: %s", protocol)
		return ""
	}

	return d[1] + "/" + repoPath
}

// Extract array of string from urlString
func getURLStringArray(urlString string) []string {
	// Supported git regex based on URI allowed regex as defined under RFC3986
	const rfc3986 = `[A-Za-z0-9][A-Za-z0-9+.-]*`
	re := regexp.MustCompile(rfc3986)
	var u []string = re.FindAllString(urlString, -1)
	return u
}
