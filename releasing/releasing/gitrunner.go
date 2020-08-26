package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type gitRunner struct {
	// Original git repo path, which should be current working directory
	originalGitPath string
	// A temporary path for worktree
	worktreePath string
	// Does this have worktree
	hasWorktree bool
}

func newGitRunner(worktree bool) (gitRunner, error) {
	gr := gitRunner{}
	pwd, err := os.Getwd()
	if err != nil {
		return gr, err
	}
	gr.originalGitPath = pwd
	gr.hasWorktree = worktree
	if worktree {
		err = gr.CreateWorktreeDir()
		if err != nil {
			return gr, err
		}
	}
	return gr, nil
}

func (gr *gitRunner) Close() error {
	if !gr.hasWorktree {
		return nil
	}
	err := gr.DeleteWorktreeDir()
	if err != nil {
		return err
	}
	err = gr.PruneWorktree()
	if err != nil {
		return err
	}
	return nil
}

func (gr *gitRunner) DeleteWorktreeDir() error {
	logDebug("Deleting git worktree dir: %s", gr.worktreePath)
	err := os.RemoveAll(gr.worktreePath)
	if err != nil {
		return err
	}
	logDebug("Deleting done")
	return nil
}

func (gr *gitRunner) WorktreePath() (string, error) {
	if gr.worktreePath == "" {
		return "", fmt.Errorf("empty worktree path")
	}
	return gr.worktreePath, nil
}

func (gr *gitRunner) OriginalGitPath() (string, error) {
	if gr.originalGitPath == "" {
		return "", fmt.Errorf("empty git path")
	}
	return gr.originalGitPath, nil
}

func (gr *gitRunner) CreateWorktreeDir() error {
	// Create temporary directory
	temp, err := ioutil.TempDir("", "kustomize-releases")
	if err != nil {
		return err
	}
	gr.worktreePath = filepath.Join(temp, "sigs.k8s.io/kustomize")
	err = os.MkdirAll(gr.worktreePath, 0700)
	logDebug("Created git worktree dir: %s", gr.worktreePath)
	if err != nil {
		return err
	}
	return nil
}

func (gr *gitRunner) CheckRemoteExistence(remote string) error {
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	logDebug("Checking remote %s in %s", remote, path)
	cmd := exec.Command("git", "remote")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	logDebug("Remotes:\n%s", string(stdoutStderr))

	regString := fmt.Sprintf("(?m)^\\s*%s\\s*$", remote)
	reg := regexp.MustCompile(regString)
	if !reg.MatchString(string(stdoutStderr)) {
		return fmt.Errorf("cannot find remote named %s", remote)
	}
	logDebug("Remote %s exists", remote)
	return nil
}

func (gr *gitRunner) FetchTags(remote string) error {
	logDebug("Fetching latest tags")
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "fetch", "-t", remote, "-f")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	logDebug("Finished fetching")
	return nil
}

func (gr *gitRunner) GetTags() (string, error) {
	path, err := gr.OriginalGitPath()
	if err != nil {
		return "", err
	}
	logDebug("Getting latest tag in repo %s", path)
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	logDebug("Finished getting tags")
	return string(stdoutStderr), nil
}

func (gr *gitRunner) CheckBranchExistence(name string) (bool, error) {
	logDebug("Checking branch %s existence", name)
	path, err := gr.OriginalGitPath()
	if err != nil {
		return false, err
	}
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	branches := strings.Split(string(stdoutStderr), "\n")
	for _, branch := range branches {
		if strings.Trim(branch, " ") == "remotes/"+name {
			return true, nil
		}
	}
	return false, nil
}

func (gr *gitRunner) NewBranch(name string) error {
	logInfo("Creating new branch %s", name)
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	upstreamBranch := "upstream/" + name
	cmd := exec.Command("git", "branch", name, upstreamBranch)
	exist, err := gr.CheckBranchExistence(upstreamBranch)
	if err != nil {
		return err
	}
	if !exist {
		logInfo("Remote branch %s doesn't exist", upstreamBranch)
		cmd = exec.Command("git", "branch", name)
	}
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	return nil
}

func (gr *gitRunner) DeleteBranch(name string) error {
	logDebug("Deleting branch %s", name)
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "branch", "-D", name)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	logDebug("Finished deleting branch")
	return nil
}

func (gr *gitRunner) AddWorktree(branch string) error {
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	tempDir, err := gr.WorktreePath()
	if err != nil {
		return err
	}
	logInfo("Adding worktree %s for branch %s", tempDir, branch)
	cmd := exec.Command("git", "worktree", "add", tempDir, branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	return nil
}

func (gr *gitRunner) PruneWorktree() error {
	path, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	logDebug("Pruning worktree for repo %s", path)
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	logDebug("Finished pruning worktree")
	return nil
}

func (gr *gitRunner) Merge(branch string) error {
	logInfo("Merging %s", branch)
	path, err := gr.WorktreePath()
	if err != nil {
		return err
	}
	logDebug("Working dir: %s", path)
	cmd := exec.Command("git", "merge", branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	return nil
}

func (gr *gitRunner) PushRelease(branch string, mod module) error {
	logInfo("Pushing branch %s", branch)
	path, err := gr.WorktreePath()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "push", "upstream", branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}

	logInfo("Creating tag %s", mod.Tag())
	cmd = exec.Command(
		"git", "tag",
		"-a", mod.Tag(),
		"-m", fmt.Sprintf("Release %s on branch %s", mod.Tag(), branch),
	)
	cmd.Dir = path
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}

	logInfo("Pushing tag %s", mod.Tag())
	cmd = exec.Command("git", "push", "upstream", mod.Tag())
	cmd.Dir = path
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stdoutStderr)
	}
	return nil
}
