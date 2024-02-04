// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

const (
	refsTags       = "refs/tags/"
	pathSep        = "/"
	remoteOrigin   = misc.TrackedRepo("origin")
	remoteUpstream = misc.TrackedRepo("upstream")
	mainBranch     = "master"
	indent         = "  "
	doing          = "  [x] "
	faking         = "  [ ] "
)

type safetyLevel int

const (
	// Commands that don't hurt, e.g. checking out an existing branch.
	noHarmDone safetyLevel = iota
	// Commands that write, and could be hard to undo.
	undoPainful
)

type Verbosity int

const (
	Low Verbosity = iota
	High
)

var recognizedRemotes = []misc.TrackedRepo{remoteUpstream, remoteOrigin}

// Runner runs specific git tasks using the git CLI.
type Runner struct {
	// From which directory do we run the commands.
	workDir string
	// Run commands, or merely print commands.
	doIt bool
	// Indicate local execution to adapt with path.
	localFlag bool
	// Run commands, or merely print commands.
	verbosity Verbosity
}

func NewLoud(wd string, doIt bool, localFlag bool) *Runner {
	return newRunner(wd, doIt, High, localFlag)
}

func NewQuiet(wd string, doIt bool, localFlag bool) *Runner {
	return newRunner(wd, doIt, Low, localFlag)
}

func newRunner(wd string, doIt bool, v Verbosity, localFlag bool) *Runner {
	return &Runner{workDir: wd, doIt: doIt, verbosity: v, localFlag: localFlag}
}

func (gr *Runner) comment(f string) {
	if gr.verbosity == Low {
		return
	}
	fmt.Print(indent)
	fmt.Println(f)
}

func (gr *Runner) doing(s string) {
	if gr.verbosity == Low {
		return
	}
	fmt.Print(indent)
	fmt.Print(doing)
	fmt.Println(s)
}

func (gr *Runner) faking(s string) {
	if gr.verbosity == Low {
		return
	}
	fmt.Print(indent)
	fmt.Print(faking)
	fmt.Println(s)
}

func (gr *Runner) run(sl safetyLevel, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = gr.workDir
	if gr.doIt || sl == noHarmDone {
		gr.doing(c.String())
		out, err := c.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"%s out=%q", err.Error(), strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	gr.faking(c.String())
	return "", nil
}

func (gr *Runner) runNoOut(s safetyLevel, args ...string) error {
	_, err := gr.run(s, args...)
	return err
}

// TODO: allow for other remote names.
func (gr *Runner) DetermineRemoteToUse() (misc.TrackedRepo, error) {
	gr.comment("determining remote to use")
	out, err := gr.run(noHarmDone, "remote")
	if err != nil {
		return "", err
	}
	remotes := strings.Split(out, "\n")
	if len(remotes) < 1 {
		return "", fmt.Errorf("need at least one remote")
	}
	for _, n := range recognizedRemotes {
		if contains(remotes, n) {
			return n, nil
		}
	}
	return "", fmt.Errorf(
		"unable to find recognized remote %v", recognizedRemotes)
}

func contains(list []string, item misc.TrackedRepo) bool {
	for _, n := range list {
		if n == string(item) {
			return true
		}
	}
	return false
}

func (gr *Runner) LoadLocalTags() (result misc.VersionMap, err error) {
	gr.comment("loading local tags")
	var out string
	out, err = gr.run(noHarmDone, "tag", "-l")
	if err != nil {
		return nil, err
	}
	result = make(misc.VersionMap)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		n, v, err := parseModuleSpec(l)
		if err != nil {
			// ignore it
			continue
		}
		result[n] = append(result[n], v)
	}
	for _, versions := range result {
		sort.Sort(versions)
	}
	return
}

func (gr *Runner) LoadRemoteTags(
	remote misc.TrackedRepo) (result misc.VersionMap, err error) {
	var out string

	// Update latest tags from upstream
	gr.comment("updating tags from upstream")
	_, err = gr.run(noHarmDone, "fetch", "-t", string(remoteUpstream), string(mainBranch))
	if err != nil {
		// Handle if repo is not a fork
		_, err = gr.run(noHarmDone, "fetch", "-t", string(mainBranch))
		if err != nil {
			_ = fmt.Errorf("failed to fetch tags from %s", string(mainBranch))
		}
	}

	gr.comment("loading remote tags")
	out, err = gr.run(noHarmDone, "ls-remote", "--ref", string(remote))
	if err != nil {
		return nil, err
	}
	result = make(misc.VersionMap)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) < 2 {
			// ignore it
			continue
		}
		if !strings.HasPrefix(fields[1], refsTags) {
			// ignore it
			continue
		}
		path := fields[1][len(refsTags):]
		n, v, err := parseModuleSpec(path)
		if err != nil {
			// ignore it
			continue
		}
		result[n] = append(result[n], v)
	}
	for _, versions := range result {
		sort.Sort(versions)
	}
	return
}

func parseModuleSpec(
	path string) (n misc.ModuleShortName, v semver.SemVer, err error) {
	fields := strings.Split(path, pathSep)
	v, err = semver.Parse(fields[len(fields)-1])
	if err != nil {
		// Silently ignore versions we don't understand.
		return "", v, err
	}
	n = misc.ModuleAtTop
	if len(fields) > 1 {
		n = misc.ModuleShortName(
			strings.Join(fields[:len(fields)-1], pathSep))
	}
	return
}

func (gr *Runner) Debug(remote misc.TrackedRepo) error {
	return nil // gr.CheckoutMainBranch(remote)
}

func (gr *Runner) AssureCleanWorkspace() error {
	gr.comment("assuring a clean workspace")
	out, err := gr.run(noHarmDone, "status")
	if err != nil {
		return err
	}
	if !strings.Contains(out, "nothing to commit, working tree clean") {
		return fmt.Errorf("the workspace isn't clean")
	}
	return nil
}

func (gr *Runner) AssureOnMainBranch() error {
	gr.comment("assuring main branch checked out")
	out, err := gr.run(noHarmDone, "status")
	if err != nil {
		return err
	}
	if !strings.Contains(out, "On branch "+mainBranch) {
		return fmt.Errorf("expected to be on branch %q", mainBranch)
	}
	return nil
}

// CheckoutMainBranch does that.
func (gr *Runner) CheckoutMainBranch() error {
	gr.comment("checking out main branch")
	fullBranchSpec := fmt.Sprintf("%s/%s", remoteOrigin, mainBranch)
	return gr.runNoOut(noHarmDone, "checkout", fullBranchSpec)
}

// FetchRemote does that.
func (gr *Runner) FetchRemote(remote misc.TrackedRepo) error {
	gr.comment("fetching remote")
	err := gr.runNoOut(noHarmDone, "fetch", string(remote))
	if err != nil {
		// If current repo is fork
		err = gr.runNoOut(noHarmDone, "fetch", string(remoteUpstream))
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}

// MergeFromRemoteMain does a fast forward only merge with main branch.
func (gr *Runner) MergeFromRemoteMain(remote misc.TrackedRepo) error {
	remo := strings.Join(
		[]string{string(remote), mainBranch}, pathSep)
	gr.comment("merging from remote")
	return gr.runNoOut(undoPainful, "merge", "--ff-only", remo)
}

// CheckoutReleaseBranch attempts to checkout or create a branch.
// If it's on the remote already, fail if we cannot check it out locally.
func (gr *Runner) CheckoutReleaseBranch(
	remote misc.TrackedRepo, branch string) error {
	yes, err := gr.doesRemoteBranchExist(remote, branch)
	if err != nil {
		return err
	}
	if yes {
		gr.comment("checking out branch")
		if out, err := gr.run(noHarmDone, "checkout", branch); err != nil {
			fmt.Printf("error with checkout: %q", err.Error())
			fmt.Printf("out: %q", out)
			return fmt.Errorf(
				"branch %q exists on remote %q, but isn't present locally",
				branch, string(remote))
		}
		return nil
	}
	gr.comment("creating branch")
	// The branch doesn't exist remotely.  Create or reset it locally.
	out, err := gr.run(noHarmDone, "checkout", "-B", branch)
	if err != nil {
		return err
	}
	// Expected strings: "Switched to a new branch" or "Switched to and reset branch"
	if !strings.Contains(out, "Switched to") {
		return fmt.Errorf("unexpected branch creation output: %q", out)
	}
	return nil
}

func (gr *Runner) doesRemoteBranchExist(
	remote misc.TrackedRepo, branch string) (bool, error) {
	gr.comment("looking for branch on remote")
	out, err := gr.run(noHarmDone, "branch", "-r")
	if err != nil {
		return false, err
	}
	lookFor := strings.Join([]string{string(remote), branch}, pathSep)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == lookFor {
			return true, nil
		}
	}
	return false, nil
}

func (gr *Runner) PushBranchToRemote(
	remote misc.TrackedRepo, branch string) error {
	gr.comment("pushing branch to remote")
	return gr.runNoOut(undoPainful, "push", "-f", string(remote), branch)
}

func (gr *Runner) CreateLocalReleaseTag(tag, branch string) error {
	msg := fmt.Sprintf("\"Release %s on branch %s\"", tag, branch)
	gr.comment("creating local release tag")
	return gr.runNoOut(
		undoPainful,
		"tag", "-a",
		"-m", msg,
		tag)
}

func (gr *Runner) DeleteLocalTag(tag string) error {
	gr.comment("deleting local tag")
	return gr.runNoOut(undoPainful, "tag", "--delete", tag)
}

func (gr *Runner) PushTagToRemote(
	remote misc.TrackedRepo, tag string) error {
	gr.comment("pushing tag to remote")
	return gr.runNoOut(undoPainful, "push", string(remote), tag)
}

func (gr *Runner) DeleteTagFromRemote(
	remote misc.TrackedRepo, tag string) error {
	gr.comment("deleting tags from remote")
	return gr.runNoOut(undoPainful, "push", string(remote), ":"+refsTags+tag)
}

func (gr *Runner) GetLatestTag(releaseBranch string) (string, error) {
	var latestTag string
	// Assuming release branch has this format: release-path/to/module-vX.Y.Z
	// and each release branch maintains tags, extract version from latest `releaseBranch`
	gr.comment("extract version from latest release branch")
	filteredBranchList, err := gr.run(noHarmDone, "branch", "-a", "--list", "*"+releaseBranch+"*", "--sort=-committerdate")
	if len(filteredBranchList) < 1 {
		_ = fmt.Errorf("latest tag not found for %s", releaseBranch)
		return "", err
	}
	newestBranch := strings.Split(strings.ReplaceAll(filteredBranchList, "\r\n", "\n"), "\n")
	split := strings.Split(newestBranch[0], "-")
	latestTag = split[len(split)-1]
	if err != nil {
		_ = fmt.Errorf("error getting latest tag for %s", releaseBranch)
	}

	return latestTag, nil
}

func (gr *Runner) GetMainBranch() string {
	return string(mainBranch)
}

func (gr *Runner) GetCurrentVersion() string {
	currentBranchName, err := gr.run(noHarmDone, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		_ = fmt.Errorf("error getting current version")
	}
	// Assuming release branch has this format: release-path/to/module-vX.Y.Z
	splitBranchName := strings.Split(currentBranchName, "-")
	return splitBranchName[len(splitBranchName)-1]
}
