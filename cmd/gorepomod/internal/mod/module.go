// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package mod

import (
	"fmt"
	"path/filepath"

	"golang.org/x/mod/modfile"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/utils"
)

// Module is an immutable representation of a Go module.
type Module struct {
	repo      misc.LaRepository
	shortName misc.ModuleShortName
	mf        *modfile.File
	vLocal    semver.SemVer
	vRemote   semver.SemVer
}

func New(
	repo misc.LaRepository,
	shortName misc.ModuleShortName,
	mf *modfile.File,
	vl semver.SemVer,
	vr semver.SemVer) *Module {
	return &Module{
		repo:      repo,
		shortName: shortName,
		mf:        mf,
		vLocal:    vl,
		vRemote:   vr,
	}
}

func (m *Module) GitRepo() misc.LaRepository {
	return m.repo
}

func (m *Module) VersionLocal() semver.SemVer {
	return m.vLocal
}

func (m *Module) VersionRemote() semver.SemVer {
	return m.vRemote
}

func (m *Module) ShortName() misc.ModuleShortName {
	return m.shortName
}

func (m *Module) ImportPath() string {
	return filepath.Join(m.repo.RepoPath(), string(m.ShortName()))
}

func (m *Module) AbsPath() string {
	return filepath.Join(m.repo.AbsPath(), string(m.ShortName()))
}

func (m *Module) DependsOn(target misc.LaModule) (bool, semver.SemVer) {
	for _, r := range m.mf.Require {
		if r.Mod.Path == target.ImportPath() {
			v, err := semver.Parse(r.Mod.Version)
			if err != nil {
				panic(err)
			}
			return true, v
		}
	}
	return false, semver.Zero()
}

func (m *Module) GetReplacements() (result []string) {
	for _, r := range m.mf.Replace {
		result = append(
			result, fmt.Sprintf("%s => %s", r.Old.String(), r.New.String()))
	}
	return
}

func (m *Module) GetDisallowedReplacements(
	allowedReplacements []string) (badReps []string) {
	for _, r := range m.GetReplacements() {
		m := utils.ExtractModule(r)
		if !utils.SliceContains(allowedReplacements, m) {
			badReps = append(badReps, r)
		}
	}
	return badReps
}
