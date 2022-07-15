// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// This package contains utilities for the command kustomize localize, which is
// documented under proposals/localize-command or at
// https://github.com/kubernetes-sigs/kustomize/blob/master/proposals/22-04-localize-command.md
package localizer

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const alpha = "alpha release"
const urlSeparator = "/"
const localizeDir = "localized-files"

var (
	errReference = errors.New("invalid reference")
)

// Writer takes care of writing localized files to their appropriate destination
// in newDir for command kustomize localize. Each Writer is associated with a
// kustomization root, either remote or local, referenced by target and writes
// files that the root directly references.
//
// Writer is meant to work with loader, which performs the path checks.
type Writer struct {
	fSys filesys.FileSystem

	// newDir argument, as defined by kustomize localize
	newDir filesys.ConfirmedDir

	// path to local copy of kustomization root for which Writer writes referenced
	// files
	root filesys.ConfirmedDir

	// directory in newDir that mirrors root
	dstDir filesys.ConfirmedDir
}

func makeAndClean(path string, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	var cleanPath filesys.ConfirmedDir
	var err error

	if err = fSys.MkdirAll(path); err != nil {
		return "", errors.Wrapf(err, "unable to create path")
	}
	if cleanPath, _, err = fSys.CleanedAbs(path); err != nil {
		return "", errors.Wrapf(err, "invalid path: cannot construct clean absolute path")
	}

	return cleanPath, nil
}

// NewWriter is the factory method for Writer.
func NewWriter(target filesys.ConfirmedDir, scope filesys.ConfirmedDir, newDir filesys.ConfirmedDir, fSys filesys.FileSystem) (*Writer, error) {
	inNewDir, _ := filepath.Rel(scope.String(), target.String())
	dst, dstErr := makeAndClean(newDir.Join(inNewDir), fSys)
	if dstErr != nil {
		return nil, errors.Wrapf(dstErr, "unable to write to localize destination directory")
	}

	return &Writer{
		fSys:   fSys,
		newDir: newDir,
		root:   target,
		dstDir: dst,
	}, nil
}

func (wr *Writer) getLocalizePath(u *url.URL) string {
	pathParts := strings.Split(strings.Trim(u.Path, urlSeparator), urlSeparator)
	return filepath.Join(localizeDir, u.Hostname(), filepath.Join(pathParts...))
}

// Write writes content to the newDir location that corresponds to path,
// relative to the kustomization root that this Writer is associated with.
// Write returns a path localized to the newly written file.
//
// Path must be a valid file path following kustomize rules. For instance,
// Load(path) should run without error. Write only throws errors associated with
// writing content.
func (wr *Writer) Write(path string, content []byte) (string, error) {
	var localizedPath string
	switch {
	case loader.HasRemoteFileScheme(path):
		u, _ := url.Parse(path)
		// version always present in raw github urls
		localizedPath = wr.getLocalizePath(u)
	case !filepath.IsAbs(path):
		// avoid symlinks; only write file corresponding to actual location in root
		// avoid path that Load() shows to be in root, but may traverse outside
		// temporarily; for example, ../root/config; problematic for rename and
		// relocation
		locDir, locFile, _ := wr.fSys.CleanedAbs(wr.root.Join(path))
		if locDir.HasPrefix(wr.newDir) { // newDir can be inside scope or target
			return "",
				errors.Wrapf(errReference,
					fmt.Sprintf("root '%s' in localize source references path '%s' in localize destination '%s'",
						wr.root, path, wr.newDir))
		}
		localizedPath, _ = filepath.Rel(wr.root.String(), locDir.Join(locFile))
	case filepath.IsAbs(path):
		return "",
			errors.Wrapf(
				errors.Wrapf(errReference, fmt.Sprintf("root '%s' references absolute path '%s'", wr.root, path)),
				alpha)
	default:
		return "",
			errors.Wrapf(errReference,
				fmt.Sprintf("root '%s' contains unhandled, invalid reference '%s'", wr.root, path))
	}

	dst := wr.dstDir.Join(localizedPath)
	if !wr.fSys.Exists(dst) {
		err := wr.fSys.WriteFile(dst, content)
		if err != nil {
			return "", errors.Wrapf(err, "failed to write file %s", dst)
		}
	}

	return localizedPath, nil
}
