// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/edit"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/git"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

// Manager manages a git repo.
// All data already loaded and validated, it's ready to go.
type Manager struct {
	// Underlying file system facts.
	dg *DotGitData

	// The remote used for fetching tags, pushing tags,
	// and pushing release branches.
	remoteName misc.TrackedRepo

	// The list of known Go modules in the repo.
	modules misc.LesModules

	// List of globally allowed module replacements.
	allowedReplacements []string
}

func (mgr *Manager) AbsPath() string {
	return mgr.dg.AbsPath()
}

func (mgr *Manager) RepoPath() string {
	return mgr.dg.RepoPath()
}

func (mgr *Manager) FindModule(
	target misc.ModuleShortName) misc.LaModule {
	return mgr.modules.Find(target)
}

func (mgr *Manager) Tidy(doIt bool) error {
	return mgr.modules.Apply(func(m misc.LaModule) error {
		return edit.New(m, doIt).Tidy()
	})
}

func (mgr *Manager) Pin(
	doIt bool, target misc.LaModule, newV semver.SemVer) error {
	return mgr.modules.Apply(func(m misc.LaModule) error {
		if yes, oldVersion := m.DependsOn(target); yes {
			return edit.New(m, doIt).Pin(target, oldVersion, newV)
		}
		return nil
	})
}

func (mgr *Manager) UnPin(
	doIt bool, target misc.LaModule, conditional misc.LaModule) error {
	if conditional == nil {
		conditional = target
	}
	return mgr.modules.Apply(func(m misc.LaModule) error {
		if yes, oldVersion := m.DependsOn(conditional); yes {
			return edit.New(m, doIt).UnPin(target, oldVersion)
		}
		return nil
	})
}

func (mgr *Manager) hasUnPinnedDeps(m misc.LaModule) string {
	if len(m.GetDisallowedReplacements(mgr.allowedReplacements)) > 0 {
		return "yes"
	}
	return ""
}

func (mgr *Manager) List() error {
	// Auto-update local tags
	gr := git.NewQuiet(mgr.AbsPath(), false, false)
	for _, module := range mgr.modules {
		releaseBranch := fmt.Sprintf("release-%s", module.ShortName())
		_, err := gr.GetLatestTag(releaseBranch)
		if err != nil {
			return fmt.Errorf("failed getting latest tags for %s", module)
		}
	}

	fmt.Printf("   src path: %s\n", mgr.dg.SrcPath())
	fmt.Printf("  repo path: %s\n", mgr.RepoPath())
	fmt.Printf("     remote: %s\n", mgr.remoteName)
	format := "%-" +
		strconv.Itoa(mgr.modules.LenLongestName()+2) +
		"s%-11s%-11s%17s  %s\n"
	fmt.Printf(
		format, "MODULE-NAME", "LOCAL", "REMOTE",
		"HAS-UNPINNED-DEPS", "INTRA-REPO-DEPENDENCIES")
	fmt.Printf(
		format, "-----------", "-----", "------",
		"-----------------", "-----------------------")
	return mgr.modules.Apply(func(m misc.LaModule) error {
		fmt.Printf(
			format, m.ShortName(),
			m.VersionLocal().Pretty(),
			m.VersionRemote().Pretty(),
			mgr.hasUnPinnedDeps(m),
			mgr.modules.InternalDeps(m))
		return nil
	})
}

func determineReleaseWorkBranch(v semver.SemVer) string {
	return fmt.Sprintf("release-%s", v.BranchLabel())
}

func determineTag(
	m misc.LaModule, v semver.SemVer) string {
	if m.ShortName() == misc.ModuleAtTop {
		return v.String()
	}
	return string(m.ShortName()) + "/" + v.String()
}

func (mgr *Manager) Debug(_ misc.LaModule, doIt bool, localFlag bool) error {
	gr := git.NewLoud(mgr.AbsPath(), doIt, localFlag)
	return gr.Debug(mgr.remoteName)
}

// DetermineNextVersion calculates the next version to use
// when releasing multiple modules at once.
// It returns the highest version among the targets.
func (mgr *Manager) DetermineNextVersion(
	targets []misc.LaModule, bump semver.SvBump) (semver.SemVer, error) {
	if len(targets) == 0 {
		return semver.Zero(), fmt.Errorf("no targets specified")
	}

	latestVer := semver.Zero()
	for _, target := range targets {
		newVersion := target.VersionLocal().Bump(bump)

		if err := mgr.verifyMajorVersion(target.ModulePath(), newVersion); err != nil {
			return semver.Zero(), err
		}

		fmt.Fprintf(
			os.Stderr,
			"Releasing %s, stepping from %s to %s\n",
			target.ShortName(), target.VersionLocal(), newVersion)

		if latestVer.LessThan(newVersion) {
			latestVer = newVersion
		}
	}

	return latestVer, nil
}

func (mgr *Manager) verifyMajorVersion(modulePath string, nextVersion semver.SemVer) error {
	nextMajor := nextVersion.Major()

	// For v0 and v1, /vN suffix isn't necessary.
	if nextMajor == 0 || nextMajor == 1 {
		return nil
	}

	suffix := fmt.Sprintf("/v%d", nextMajor)

	if !strings.HasSuffix(modulePath, suffix) {
		return fmt.Errorf("module %s should end with %s", modulePath, suffix)
	}

	return nil
}

// DetermineNextVersion calculates the next version to use
// when releasing multiple modules at once.
// It returns the highest version among the targets.
func (mgr *Manager) PushToReleaseWorkBranch(
	version semver.SemVer, doIt bool, localFlag bool) error {
	branch := determineReleaseWorkBranch(version)

	fmt.Printf(
		"Pushing to release work branch %q for version %s\n",
		branch, version)

	gr := git.NewLoud(mgr.AbsPath(), doIt, localFlag)

	if err := gr.AssureCleanWorkspace(); err != nil {
		return err
	}
	if err := gr.FetchRemote(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.CheckoutMainBranch(); err != nil {
		return err
	}
	if err := gr.MergeFromRemoteMain(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.AssureCleanWorkspace(); err != nil {
		return err
	}
	if err := gr.CheckoutReleaseBranch(mgr.remoteName, branch); err != nil {
		return err
	}
	if err := gr.MergeFromRemoteMain(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.PushBranchToRemote(mgr.remoteName, branch); err != nil {
		return err
	}

	return nil
}

// Release supports a gitlab flow style release process.
//
// * All development happens in the branch named "master".
// * Each minor release gets its own branch.
func (mgr *Manager) Release(
	target misc.LaModule, newVersion semver.SemVer, doIt bool, localFlag bool) error {
	if reps := target.GetDisallowedReplacements(
		mgr.allowedReplacements); len(reps) > 0 {
		return fmt.Errorf(
			"to release %q, first pin these replacements: %v",
			target.ShortName(), reps)
	}

	if newVersion.Equals(target.VersionRemote()) {
		return fmt.Errorf(
			"version %s already exists on remote - delete it first", newVersion)
	}
	if newVersion.LessThan(target.VersionRemote()) {
		fmt.Printf(
			"version %s is less than the most recent remote version (%s)",
			newVersion, target.VersionRemote())
	}

	gr := git.NewLoud(mgr.AbsPath(), doIt, localFlag)

	relBranch, err := gr.GetCurrentBranch()
	if err != nil {
		return err
	}

	expectedBranch := determineReleaseWorkBranch(newVersion)
	if relBranch != expectedBranch {
		return fmt.Errorf(
			"current branch is %q, want %q",
			relBranch, expectedBranch)
	}

	relTag := determineTag(target, newVersion)

	fmt.Printf(
		"Releasing %s, stepping from %s to %s\n",
		target.ShortName(), target.VersionLocal(), newVersion)

	if err := gr.AssureCleanWorkspace(); err != nil {
		return err
	}
	if err := gr.FetchRemote(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.MergeFromRemoteBranch(mgr.remoteName, relBranch); err != nil {
		return err
	}
	if err := gr.CreateLocalReleaseTag(relTag, relBranch); err != nil {
		return err
	}
	if err := gr.PushTagToRemote(mgr.remoteName, relTag); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) UnRelease(target misc.LaModule, doIt bool, localFlag bool) error {
	fmt.Printf(
		"Unreleasing %s/%s\n",
		target.ShortName(), target.VersionRemote())

	tag := determineTag(target, target.VersionRemote())

	gr := git.NewLoud(mgr.AbsPath(), doIt, localFlag)

	if err := gr.DeleteTagFromRemote(mgr.remoteName, tag); err != nil {
		return err
	}
	if err := gr.DeleteLocalTag(tag); err != nil {
		return err
	}
	return nil
}
