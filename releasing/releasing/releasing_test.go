package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestList(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf(err.Error())
	}
	remote := "upstream"
	// Check remotes
	checkRemoteExistence(pwd, remote)
	// Fetch latest tags from remote
	fetchTags(pwd, remote)
	for _, mod := range modules {
		v := getModuleCurrentVersion(mod, pwd)
		valid, err := regexp.MatchString("^v(\\d+\\.){2}\\d+$", v)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !valid {
			t.Errorf("Returned version %s is not valid", v)
		}
	}
}

func TestRelease(t *testing.T) {
	prepareGit()
	modName := "api"
	versionType := "patch"
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf(err.Error())
	}
	remote := "upstream"
	// Check remotes
	checkRemoteExistence(pwd, remote)
	// Fetch latest tags from remote
	fetchTags(pwd, remote)
	mod := module{
		name: modName,
		path: pwd,
	}
	mod.UpdateCurrentVersion()

	oldVersion := mod.version.String()
	mod.version.Bump(versionType)
	newVersion := mod.version.String()
	logInfo("Bumping version: %s => %s", oldVersion, newVersion)

	// Create branch
	branch := fmt.Sprintf("release-%s-v%d.%d", mod.name, mod.version.major, mod.version.minor)
	newBranch(pwd, branch)

	addWorktree(pwd, tempDir, branch)

	merge(tempDir, "upstream/master")
	// Update module path
	mod.path = tempDir

	// Clean
	cleanGit()
	pruneWorktree(pwd)
	deleteBranch(pwd, branch)
}
