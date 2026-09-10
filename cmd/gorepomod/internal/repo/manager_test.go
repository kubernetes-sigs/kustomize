// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"testing"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

type fakeModule struct {
	shortName misc.ModuleShortName
}

func (f fakeModule) ShortName() misc.ModuleShortName { return f.shortName }
func (f fakeModule) ModulePath() string              { return "" }
func (f fakeModule) ImportPath() string              { return "" }
func (f fakeModule) AbsPath() string                 { return "" }
func (f fakeModule) VersionLocal() semver.SemVer     { return semver.Zero() }
func (f fakeModule) VersionRemote() semver.SemVer    { return semver.Zero() }
func (f fakeModule) DependsOn(misc.LaModule) (bool, semver.SemVer) {
	return false, semver.Zero()
}
func (f fakeModule) GetReplacements() []string { return nil }
func (f fakeModule) GetDisallowedReplacements([]string) []string {
	return nil
}

func TestDetermineTag(t *testing.T) {
	tag := determineTag(
		fakeModule{shortName: "api"},
		semver.New(1, 2, 3),
	)
	if tag != "api/v1.2.3" {
		t.Fatalf("tag = %q, want %q", tag, "api/v1.2.3")
	}
}

func TestDetermineReleaseBranchUsesArgument(t *testing.T) {
	branch, err := determineReleaseBranch("release-v5.8.2")
	if err != nil {
		t.Fatalf("determineReleaseBranch() error = %v", err)
	}
	if branch != "release-v5.8.2" {
		t.Fatalf("branch = %q, want %q", branch, "release-v5.8.2")
	}
}

func TestDetermineReleaseBranchRequiresArgument(t *testing.T) {
	_, err := determineReleaseBranch("")
	if err == nil {
		t.Fatal("determineReleaseBranch() error = nil, want non-nil")
	}
}
