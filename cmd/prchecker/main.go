// prchecker examines pull requests
//
// - When a PR includes files from multiple modules that we'd rather not
//   modify at the same time (in an effort to  have more self-contained
//   release notes), the script will exit with a non-zero exit code.
//
// Usage:
//
//    go run . \
//      -owner=kubernetes-sigs \
//      -repo=kustomize \
//      -pr=2997 \
//     cmd/config api kustomize kyaml

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

// splitCommitsHelp is a help message displayed when a PR has commits which
// span modules.
const splitCommitsHelp = "\nCommits that include multiple Go modules must be split into multiple commits that each touch no more than one module.\n" +
	"Splitting instructions: https://git-scm.com/docs/git-rebase#_splitting_commits\n"

// Ignore span pattern defines a regular expression pattern that, when matched,
// causes this check to be ignored.
// Pattern expects "ALLOW_MODULE_SPAN" in CAPS on a line by itself.
// Spaces may be included before or after but no other characters with one
// exception. ">" may be used to quote the exclusion.
// Pattern may be provided at any point in the pull request description.
//
// Ex: "ALLOW_MODULE_SPAN", "> ALLOW_MODULE_SPAN", "   ALLOW_MODULE_SPAN   "
const ignoreSpanPattern = "(?m)^>?\\s*ALLOW_MODULE_SPAN\\s*$"

// Changeset represents a set of file modifications associated with a commit
type Changeset struct {
	files []string
	id    string
}

// GitHubRepository represents a pairing of the owner and repository to make
// passing around references to specific projects simpler
type GitHubRepository struct {
	client *github.Client
	owner  *string
	repo   *string
}

// GitHubService is the collection of GitHub API interactions this service consumes
type GitHubService interface {
	GetPullRequest(prId int) (*github.PullRequest, *github.Response, error)
	GetCommit(commitSha string) (*github.RepositoryCommit, *github.Response, error)
	ListCommits(prId int, options *github.ListOptions) ([]*github.RepositoryCommit, *github.Response, error)
}

// GetPullRequest retrieves details about a pull request from GitHub API
func (repository GitHubRepository) GetPullRequest(prId int) (
	*github.PullRequest, *github.Response, error) {
	return repository.client.PullRequests.Get(
		context.Background(),
		*repository.owner,
		*repository.repo,
		prId)
}

// GetCommit retrieves commit details from GitHub API
func (repository GitHubRepository) GetCommit(commitSha string) (
	*github.RepositoryCommit, *github.Response, error) {
	return repository.client.Repositories.GetCommit(context.Background(),
		*repository.owner,
		*repository.repo,
		commitSha)
}

// ListCommits lists commits in a PR via GitHub API
func (repository GitHubRepository) ListCommits(prId int, options *github.ListOptions) (
	[]*github.RepositoryCommit, *github.Response, error) {
	return repository.client.PullRequests.ListCommits(context.Background(),
		*repository.owner,
		*repository.repo,
		prId,
		options)
}

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

	githubRepo := &GitHubRepository{
		client: client,
		owner:  owner,
		repo:   repo,
	}

	// Check if module span is allowed before scanning on commits
	isSpanAllowed, err := ModuleSpanAllowed(githubRepo, *pullrequest)

	if err != nil {
		fmt.Printf("unable to retrieve pull request details: %v", err.Error())
		os.Exit(2)
	}

	if isSpanAllowed {
		fmt.Println("Check not run. Module spanning was allowed in this Pull Request.")
		os.Exit(0)
	}

	isSpanningPull, _, err := PullRequestSpanningPathList(githubRepo, *pullrequest, restrictedPaths)

	if err != nil {
		fmt.Printf("unable to retrieve pull request details: %v", err)
		os.Exit(2)
	}

	// Exit with error if two or more restricted directories where modified
	if isSpanningPull {
		// Provide a suggestion for potential solution if the check fails.
		fmt.Println(splitCommitsHelp)
		os.Exit(1)
	}
}

// ConstructChangeset creates a changeset from a GitHub Commit object
func ConstructChangeset(commit *github.RepositoryCommit) *Changeset {
	id := commit.SHA
	fileset := []string{}

	for _, file := range commit.Files {
		fileset = append(fileset, *file.Filename)
	}

	return &Changeset{
		files: fileset,
		id:    *id,
	}
}

// StringAllowsModuleSpan tests if a string matches the allow span regex
func StringAllowsModuleSpan(body string) (bool, error) {
	return regexp.MatchString(ignoreSpanPattern, body)
}

// ModuleSpanAllowed tests a Pull Requests description for a regular
// expression. If the expression matches then spanning modules are allowed.
func ModuleSpanAllowed(repository GitHubService, pullId int) (bool, error) {

	// Note: There are multiple ways to pull a github commit object
	// we want a RepositoryCommit.
	pullRequest, _, err := repository.GetPullRequest(pullId)

	if err != nil {
		return false, err
	}

	return StringAllowsModuleSpan(*pullRequest.Body)
}

// GetCommitChanges looks up a github commit by SHA and returns a Changeset
// containing the modified files in the specified commit.
func GetCommitChanges(repository GitHubService, commitSha string) (*Changeset, error) {
	commit, _, err := repository.GetCommit(commitSha)

	if err != nil {
		return nil, err
	}

	return ConstructChangeset(commit), nil
}

// GetPullRequestCommits constructs a list of all commits and the associated
// files changes in the given pull request
func GetPullRequestCommits(repository GitHubService, pullrequest int) ([]*Changeset, error) {

	// foundFiles across all pages from github api
	var collectedCommits []*github.RepositoryCommit
	// Github only returns a limited set of commits per request and PR's may
	// exceed this so loop until all pages have been enumerated.
	options := &github.ListOptions{Page: 1}
	for options.Page != 0 {
		commits, response, err := repository.ListCommits(pullrequest, options)

		// If an error has occurred while querying api exit early, report error
		if err != nil {
			return nil, err
		}
		collectedCommits = append(collectedCommits, commits...)
		// setup next page to continue loop
		options = &github.ListOptions{Page: response.NextPage}
	}

	var changesetResults []*Changeset
	for _, commit := range collectedCommits {
		// The repository commits from list commits are not hydrated
		// We will need to retrieve the complete object:
		changeset, err := GetCommitChanges(repository, *commit.SHA)
		if err != nil {
			return nil, err
		}

		changesetResults = append(changesetResults, changeset)
	}
	return changesetResults, nil
}

// PullRequestSpanningPathList tests if a pull request spans multiple
// directory paths
func PullRequestSpanningPathList(repository GitHubService, pullrequest int, paths []string) (bool, []*Changeset, error) {
	// Create a buffer for commits
	changesets, err := GetPullRequestCommits(repository, pullrequest)

	if err != nil {
		return false, nil, err
	}

	spanningChangesExist := false
	for _, changeset := range changesets {
		if changeset.isSpanningPaths(paths) {
			// When detecting the first spanning changeset print a prefix message
			if !spanningChangesExist {
				fmt.Printf("Spanning changesets detected in the following commits:\n\n")
			}

			fmt.Printf("\t* %s\n", changeset.id)
			// In order provide a full list of outstanding commits, do not shortcircuit this check
			spanningChangesExist = true
		}
	}

	return spanningChangesExist, changesets, nil
}

// isSpanningPaths tests if a changeset is spanning
// multiple directory paths.
func (changeset *Changeset) isSpanningPaths(paths []string) bool {
	matchedPath := ""

	for _, file := range changeset.files {
		for _, path := range paths {
			if strings.HasPrefix(file, path) {
				// If a different path has already matched then the changeset spans multiple restricted paths
				if matchedPath != "" && matchedPath != path {
					return true
				}
				matchedPath = path
			}
		}
	}

	return false
}
