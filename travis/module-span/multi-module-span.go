// multi-module-span script:
// A script which can detect when a pull request would make changes which
// span a series of restricted directories (ex: multiple go modules or projects)
//
// When a pull request includes files which span two modules the script will
// exit with a non-zero exit code.
//
// Running:
// go run multi-module-span.go -owner=kubernetes-sigs -repo=kustomize -pr=2997 cmd/config  api/ kustomize/ kyaml/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

func main() {
	owner := flag.String("owner", "", "the github repository owner name")
	repo := flag.String("repo", "", "the github repository name")
	pullrequest := flag.Int("pr", -1, "the pull request number")
	flag.Parse()
	// Treat all following arguments as restricted directories
	restrictedPaths := flag.Args()

	// Short circuit check if restricted paths is less than 2
	// Conflicts won't exist in this scenario so we don't need to call the API
	if len(restrictedPaths) <= 1 {
		fmt.Println("Check not run. Add at least two restricted paths and run again.")
		os.Exit(0)
	}

	client := github.NewClient(nil)

	files, err := ListAllPullRequestFiles(client, owner, pullrequest, repo)

	if err != nil {
		fmt.Println("Unable to retrieve pull request details:", err.Error())
		os.Exit(2)
	}

	contributedRestrictedPaths := CountRestrictedPathUses(files, restrictedPaths)
	modifiedRestrictedDirectories := CountModifiedRestrictedDirectories(contributedRestrictedPaths)

	// Exit with error if two or more restricted directories where modified
	if modifiedRestrictedDirectories > 1 {
		fmt.Println("Modifications to multiple restricted directories occurred.")
		os.Exit(1)
	}
}

// ListAllPullRequestFiles retrieves as many files as possible for the
// target pull request.
//
// Note: GitHub API limits ListFiles to a maximum of 3000 files. Very large
// changes which exceed this limit may pass this check even if they
// do contain spanning changes.
// see: https://developer.github.com/v3/pulls/#list-pull-requests-files
func ListAllPullRequestFiles(client *github.Client, owner *string, pullrequest *int, repo *string) ([]*github.CommitFile, error) {
	// foundFiles across all pages from github api
	var foundFiles []*github.CommitFile
	// GitHub returns the first 30 files by default, increase this value
	// Note: Page 1 is the first page of results. Page 0 is an end of list mark.
	// Github only returns (max) 100 results per page and PR's may exceed this
	// so loop until all pages have been enumerated.
	options := &github.ListOptions{PerPage: 100, Page: 1}
	for options.Page != 0 {
		files, response, err := client.PullRequests.ListFiles(context.Background(), *owner, *repo, *pullrequest, options)

		// If an error has occurred while querying api exit early, report error
		if err != nil {
			return nil, err
		}
		foundFiles = append(foundFiles, files...)
		// setup next page to continue loop
		options = &github.ListOptions{PerPage: 100, Page: response.NextPage}
	}
	return foundFiles, nil
}

// CountModifiedRestrictedDirectories Accepts a map of paths and the number of
// occurances and returns the count of the paths which had a non-zero value.
func CountModifiedRestrictedDirectories(contributedRestrictedPaths map[string]int) int {
	modifiedRestrictedDirectories := 0
	for _, occurance := range contributedRestrictedPaths {
		if occurance != 0 {
			modifiedRestrictedDirectories++
		}
	}
	return modifiedRestrictedDirectories
}

// CountRestrictedPathUses Constructs a map that contains the number of
// references keyed to each restricted path. This provides details about how
// many files in the list are associated with each restricted path.
func CountRestrictedPathUses(files []*github.CommitFile, restrictedPaths []string) map[string]int {
	contributedRestrictedPaths := make(map[string]int)
	for _, path := range restrictedPaths {
		contributedRestrictedPaths[path] = 0
	}

	for _, file := range files {
		for path := range contributedRestrictedPaths {
			if strings.HasPrefix(*file.Filename, path) {
				contributedRestrictedPaths[path]++
			}
		}
	}
	return contributedRestrictedPaths
}
